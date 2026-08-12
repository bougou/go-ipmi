package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bougou/go-ipmi/pkg/types"
)

// SOLStreamOptions configures [Client.SOLStream].
type SOLStreamOptions struct {
	// PollInterval is how often to send an empty SOL packet to poll for inbound data.
	// Zero selects a default (100ms).
	PollInterval time.Duration
}

const (
	defaultSOLTransmitRetryCount    = 3
	defaultSOLTransmitRetryInterval = 100 * time.Millisecond
)

// SOLStream exchanges console data over an already active SOL payload. Input is transmitted
// unchanged, unlike [Client.SOLActivate], this method does not interpret terminal escape sequences
// or process signals.
func (c *Client) SOLStream(ctx context.Context, in io.Reader, out io.Writer, opts *SOLStreamOptions) error {
	pollInterval := defaultPollInterval
	if opts != nil && opts.PollInterval > 0 {
		pollInterval = opts.PollInterval
	}

	return c.runSOLStream(ctx, in, out, pollInterval, nil)
}

type solConsoleInput struct {
	value byte
	err   error
}

func readSOLInput(ctx context.Context, in io.Reader) <-chan solConsoleInput {
	events := make(chan solConsoleInput, 256)
	go func() {
		reader := bufio.NewReader(in)
		for {
			value, err := reader.ReadByte()
			event := solConsoleInput{value: value, err: err}
			select {
			case events <- event:
			case <-ctx.Done():
				return
			}

			if err != nil {
				return
			}
		}
	}()

	return events
}

func (c *Client) runSOLStream(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	pollInterval time.Duration,
	interrupt <-chan os.Signal,
) error {
	input := readSOLInput(ctx, in)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var localSeq uint8 = 1
	var remoteSeq uint8
	var pendingAckCount uint8

	sendPacket := func(chars []byte) (*types.SOLPayloadResponse, error) {
		req := &types.SOLPayloadRequest{
			SOLPayloadPacket: types.SOLPayloadPacket{
				SequenceNumber:         localSeq,
				AckedSequenceNumber:    remoteSeq,
				AcceptedCharacterCount: pendingAckCount,
				CharacterData:          chars,
			},
		}
		res, err := c.SOLPayload(ctx, req)
		if err != nil {
			return nil, err
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
				return nil, err
			}
		}
		return res, nil
	}

	sendData := func(chars []byte) error {
		pending := chars
		noProgressRetries := 0
		for len(pending) > 0 {
			// Retry only the data the BMC did not accept, using a fresh SOL sequence.
			res, err := sendPacket(pending)
			if err != nil {
				return err
			}

			// The accepted count applies only to the data in this request.
			accepted := int(res.AcceptedCharacterCount)
			if accepted > len(pending) {
				return fmt.Errorf("BMC acknowledged %d characters for a %d-byte SOL packet", accepted, len(pending))
			}

			if accepted > 0 {
				// Any progress resets the consecutive no-progress retry budget.
				noProgressRetries = 0
			} else {
				noProgressRetries++
			}

			// Remove the accepted data from the pending buffer.
			pending = pending[accepted:]
			if len(pending) == 0 {
				return nil
			}

			// If the BMC did not accept any data, retry up to a limit.
			if noProgressRetries > defaultSOLTransmitRetryCount {
				return fmt.Errorf("BMC did not accept %d SOL data byte(s) after %d retries",
					len(pending), defaultSOLTransmitRetryCount)
			}

			// Wait before retrying.
			timer := time.NewTimer(defaultSOLTransmitRetryInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		return nil
	}

	if _, err := sendPacket(nil); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-interrupt:
			return nil
		case consoleInput := <-input:
			if consoleInput.err != nil {
				if consoleInput.err == io.EOF {
					return nil
				}
				return consoleInput.err
			}

			if err := sendData([]byte{consoleInput.value}); err != nil {
				return err
			}
		case <-ticker.C:
			if _, err := sendPacket(nil); err != nil {
				return err
			}
		}
	}
}
