package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/bougou/go-ipmi/pkg/command/transport"
	"github.com/bougou/go-ipmi/pkg/types"
	"golang.org/x/term"
)

// SOLActivateOptions configures [Client.SOLActivate].
type SOLActivateOptions struct {
	// PayloadInstance is the SOL payload instance (1–0x3f). Zero means use default 1.
	PayloadInstance uint8
	// PollInterval is how often to send an empty SOL packet to poll for inbound data.
	// Zero selects a default (100ms).
	PollInterval time.Duration
	// OnActivated is called once after Activate Payload (SOL) succeeds and before the session loop.
	OnActivated func(payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse)
	// OnDeactivated is called once after the session loop ends.
	OnDeactivated func(payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse)
}

const (
	defaultPayloadInstance = uint8(1)
	defaultPollInterval    = 100 * time.Millisecond
)

func solActivatePollInterval(opts *SOLActivateOptions) time.Duration {
	if opts == nil || opts.PollInterval <= 0 {
		return defaultPollInterval
	}
	return opts.PollInterval
}

func solActivatePayloadInstance(opts *SOLActivateOptions) uint8 {
	if opts == nil || opts.PayloadInstance == 0 {
		return defaultPayloadInstance
	}
	return opts.PayloadInstance
}

// SOLPayload exchanges a single SOL payload packet over an active RMCP+ session.
func (c *Client) SOLPayload(ctx context.Context, request *types.SOLPayloadRequest) (response *types.SOLPayloadResponse, err error) {
	response = &types.SOLPayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// SOLActivate activates the SOL payload and runs an interactive session.
func (c *Client) SOLActivate(ctx context.Context, in io.Reader, out io.Writer, opts *SOLActivateOptions) error {
	if c.Interface != InterfaceLanplus {
		return fmt.Errorf("SOL activate requires IPMI v2.0 (RMCP+); use InterfaceLanplus (-I lanplus)")
	}

	if opts == nil {
		opts = &SOLActivateOptions{}
	}

	payloadInstance := solActivatePayloadInstance(opts)
	pollEvery := solActivatePollInterval(opts)

	useDefaultActivated := opts.OnActivated == nil
	useDefaultDeactivated := opts.OnDeactivated == nil

	var terminalConfig *solTerminalConfig

	if useDefaultActivated || useDefaultDeactivated {
		config, err := determineTerminalConfig(in)
		if err != nil {
			return err
		}
		terminalConfig = config

		if useDefaultActivated {
			opts.OnActivated = func(payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse) {
				defaultOnActivated(terminalConfig, payloadInstance, in, out, res)
			}
		}
		if useDefaultDeactivated {
			opts.OnDeactivated = func(payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse) {
				defaultOnDeactivated(terminalConfig, payloadInstance, in, out, res)
			}
		}
	}

	activatePayloadResponse, err := c.ActivatePayload(ctx, &transport.ActivatePayloadRequest{
		PayloadType:     types.PayloadTypeSOL,
		PayloadInstance: payloadInstance,
	})
	if err != nil {
		return err
	}
	if opts.OnActivated != nil {
		opts.OnActivated(payloadInstance, in, out, activatePayloadResponse)
	}

	defer func() {
		_, _ = c.DeactivatePayload(ctx, &transport.DeactivatePayloadRequest{
			PayloadType:     types.PayloadTypeSOL,
			PayloadInstance: payloadInstance,
		})

		if opts.OnDeactivated != nil {
			opts.OnDeactivated(payloadInstance, in, out, activatePayloadResponse)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	return c.runSOLStream(ctx, newSOLEscapeReader(in), out, pollEvery, sigCh)
}

type solEscapeReader struct {
	reader *bufio.Reader

	atLineStart bool
}

func newSOLEscapeReader(in io.Reader) *solEscapeReader {
	return &solEscapeReader{reader: bufio.NewReader(in), atLineStart: true}
}

func (r *solEscapeReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	value, err := r.reader.ReadByte()
	if err != nil {
		return 0, err
	}

	if r.atLineStart && value == '~' {
		next, err := r.reader.Peek(1)
		if err != nil {
			return 0, err
		}

		if next[0] == '.' {
			_, _ = r.reader.Discard(1)
			return 0, io.EOF
		}
	} else {
		r.atLineStart = value == '\r' || value == '\n'
	}

	p[0] = value

	return 1, nil
}

type solTerminalConfig struct {
	enableTTYRaw   func() error
	restoreTTY     func() error
	ttyInteractive bool
	rawModeEnabled bool
}

func defaultOnActivated(terminalConfig *solTerminalConfig, payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse) {
	_, _ = fmt.Fprintf(out, "SOL payload activated (instance: %d)\n", payloadInstance)
	_, _ = fmt.Fprintf(out, "Inbound payload size : %d bytes\n", res.InboundPayloadSize)
	_, _ = fmt.Fprintf(out, "Outbound payload size: %d bytes\n", res.OutboundPayloadSize)
	_, _ = fmt.Fprintf(out, "Payload UDP port     : %d\n", res.PayloadUDPPort)
	_, _ = fmt.Fprintf(out, "Payload VLAN ID      : %d\n", res.PayloadVLANID)

	if terminalConfig.ttyInteractive {
		_, _ = io.WriteString(out, "Connected. Use ~. to terminate.\n")
		if terminalConfig.enableTTYRaw != nil {
			if enableErr := terminalConfig.enableTTYRaw(); enableErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: failed to switch terminal to raw mode: %v\n", enableErr)
			} else {
				terminalConfig.rawModeEnabled = true
			}
		}
	} else {
		_, _ = io.WriteString(out, "Connected. Use ~. to terminate (line-buffered mode).\n")
	}
}

func defaultOnDeactivated(terminalConfig *solTerminalConfig, payloadInstance uint8, in io.Reader, out io.Writer, res *transport.ActivatePayloadResponse) {
	hadRawMode := terminalConfig.rawModeEnabled
	if terminalConfig.rawModeEnabled && terminalConfig.restoreTTY != nil {
		if restoreErr := terminalConfig.restoreTTY(); restoreErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: failed to restore terminal mode: %v\n", restoreErr)
		}
		terminalConfig.rawModeEnabled = false
	}

	if hadRawMode {
		_, _ = io.WriteString(out, "\r\nSOL payload deactivated.\n")
	} else {
		_, _ = io.WriteString(out, "SOL payload deactivated.\n")
	}
}

func isTTYReader(in io.Reader) (file *os.File, ok bool) {
	f, ok := in.(*os.File)
	if !ok {
		return nil, false
	}
	info, err := f.Stat()
	if err != nil {
		return nil, false
	}
	return f, info.Mode()&os.ModeCharDevice != 0
}

func determineTerminalConfig(in io.Reader) (*solTerminalConfig, error) {
	var (
		enableTTYRaw   func() error
		restoreTTY     func() error
		ttyInteractive bool
	)

	if inFile, ok := isTTYReader(in); ok {
		ttyInteractive = true

		fd := int(inFile.Fd())
		originalState, err := term.GetState(fd)
		if err != nil {
			return nil, fmt.Errorf("failed to read terminal state: %w", err)
		}
		enableTTYRaw = func() error {
			if _, makeErr := term.MakeRaw(fd); makeErr != nil {
				return fmt.Errorf("failed to switch input stream to raw mode: %w", makeErr)
			}
			return nil
		}
		restoreTTY = func() error {
			return term.Restore(fd, originalState)
		}
	}

	return &solTerminalConfig{
		enableTTYRaw:   enableTTYRaw,
		restoreTTY:     restoreTTY,
		ttyInteractive: ttyInteractive,
	}, nil
}
