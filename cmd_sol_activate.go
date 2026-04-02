package ipmi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

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
	OnActivated func(payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse)
	// OnDeactivated is called once after the session loop ends.
	OnDeactivated func(payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse)
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

// SOLActivate activates the SOL payload and runs an interactive session: it reads bytes from in,
// sends SOL payload packets to the BMC, and writes inbound serial data to out. The session ends
// when the context is cancelled, SIGINT is received, or the escape sequence "~." is read at the
// beginning of a line (same convention as ipmitool).
//
// On exit it sends Deactivate Payload for the SOL instance. Input is line-buffered (byte-wise from bufio).
//
// Requires a lanplus (IPMI v2.0) session; see [Client.Connect].
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
			opts.OnActivated = func(payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse) {
				defaultOnActivated(terminalConfig, payloadInstance, in, out, res)
			}
		}
		if useDefaultDeactivated {
			opts.OnDeactivated = func(payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse) {
				defaultOnDeactivated(terminalConfig, payloadInstance, in, out, res)
			}
		}
	}

	activatePayloadResponse, err := c.ActivatePayload(ctx, &ActivatePayloadRequest{
		PayloadType:     PayloadTypeSOL,
		PayloadInstance: payloadInstance,
	})
	if err != nil {
		if respErr, ok := isResponseError(err); ok && respErr.CompletionCode() == CompletionCode(0x80) {
			return errors.New("SOL payload already active on another session")
		}
		return err
	}
	if opts.OnActivated != nil {
		opts.OnActivated(payloadInstance, in, out, activatePayloadResponse)
	}

	defer func() {
		_, _ = c.DeactivatePayload(ctx, &DeactivatePayloadRequest{
			PayloadType:     PayloadTypeSOL,
			PayloadInstance: payloadInstance,
		})

		if opts.OnDeactivated != nil {
			opts.OnDeactivated(payloadInstance, in, out, activatePayloadResponse)
		}
	}()

	inputCh := make(chan byte, 256)
	errCh := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(in)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				errCh <- err
				return
			}
			inputCh <- b
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	var localSeq uint8 = 1
	var remoteSeq uint8
	var pendingAckCount uint8

	atLineStart := true
	pendingEscape := false

	sendPacket := func(chars []byte) error {
		req := &SOLPayloadRequest{
			SOLPayloadPacket: SOLPayloadPacket{
				SequenceNumber:         localSeq,
				AckedSequenceNumber:    remoteSeq,
				AcceptedCharacterCount: pendingAckCount,
				CharacterData:          chars,
			},
		}
		res, err := c.SOLPayload(ctx, req)
		if err != nil {
			return err
		}
		localSeq++
		if localSeq > 0x0f {
			localSeq = 1
		}
		pendingAckCount = 0

		remoteSeq = res.SequenceNumber & 0x0f
		pendingAckCount = uint8(len(res.CharacterData))

		if len(res.CharacterData) > 0 {
			if _, err := out.Write(res.CharacterData); err != nil {
				return err
			}
		}
		return nil
	}

	if err := sendPacket(nil); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-sigCh:
			return nil

		case err := <-errCh:
			if err == io.EOF {
				return nil
			}
			return err

		case b := <-inputCh:
			if pendingEscape {
				pendingEscape = false
				if b == '.' {
					return nil
				}
				if err := sendPacket([]byte{'~'}); err != nil {
					return err
				}
			}

			if atLineStart && b == '~' {
				pendingEscape = true
				continue
			}

			if err := sendPacket([]byte{b}); err != nil {
				return err
			}
			atLineStart = (b == '\r' || b == '\n')

		case <-ticker.C:
			if err := sendPacket(nil); err != nil {
				return err
			}
		}
	}
}

type solTerminalConfig struct {
	enableTTYRaw   func() error
	restoreTTY     func() error
	ttyInteractive bool
	rawModeEnabled bool
}

func defaultOnActivated(terminalConfig *solTerminalConfig, payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse) {
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

func defaultOnDeactivated(terminalConfig *solTerminalConfig, payloadInstance uint8, in io.Reader, out io.Writer, res *ActivatePayloadResponse) {
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
	return f, (info.Mode() & os.ModeCharDevice) != 0
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
