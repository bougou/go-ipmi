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
	solSequenceMask                 = 0x0f
	solStatusDeactivating           = 1 << 4 // IPMI v2.0 Table 15-2, status bit 4
	solStatusTransmitOverrun        = 1 << 3 // IPMI v2.0 Table 15-2, status bit 3
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

// solOutputStream tracks the state of the SOL output stream.
type solOutputStream struct {
	writer            io.Writer
	sequenceNumber    uint8
	acceptedCharCount uint8
}

// nextSOLSequenceNumber returns the next SOL sequence number, wrapping from 15 to 1.
func nextSOLSequenceNumber(sequenceNumber uint8) uint8 {
	sequenceNumber = (sequenceNumber + 1) & solSequenceMask
	if sequenceNumber == 0 {
		return 1
	}

	return sequenceNumber
}

func (s *solOutputStream) process(response *types.SOLPayloadResponse) error {
	if response == nil {
		return fmt.Errorf("nil SOL payload response")
	}

	// These status bits mean the stream cannot deliver complete console output.
	if response.ControlByte&solStatusDeactivating != 0 {
		return fmt.Errorf("BMC reported SOL deactivation")
	}

	if response.ControlByte&solStatusTransmitOverrun != 0 {
		return fmt.Errorf("BMC reported SOL transmit overrun, console output was lost")
	}

	sequenceNumber := response.SequenceNumber & solSequenceMask
	data := response.CharacterData
	// Sequence number zero acknowledges console input and must not carry BMC output.
	if sequenceNumber == 0 {
		if len(data) != 0 {
			return fmt.Errorf("BMC sent %d SOL data bytes with ACK-only sequence 0", len(data))
		}
		return nil
	}

	// AcceptedCharacterCount is one byte, so larger packets cannot be acknowledged.
	if len(data) > 255 {
		return fmt.Errorf("BMC sent %d SOL data bytes, acknowledgement count is limited to 255", len(data))
	}

	writeOffset := 0
	switch {
	// No BMC output sequence has been received yet.
	case s.sequenceNumber == 0:
	case sequenceNumber == s.sequenceNumber:
		// A retransmission may append data. Write only bytes not previously delivered.
		if len(data) < int(s.acceptedCharCount) {
			return fmt.Errorf("BMC shortened retried SOL sequence %d from %d to %d bytes",
				sequenceNumber, s.acceptedCharCount, len(data))
		}
		writeOffset = int(s.acceptedCharCount)
	case sequenceNumber != nextSOLSequenceNumber(s.sequenceNumber):
		// A gap means BMC output was lost before it reached the writer.
		return fmt.Errorf("unexpected BMC SOL sequence %d after %d", sequenceNumber, s.sequenceNumber)
	}

	dataToWrite := data[writeOffset:]
	if len(dataToWrite) > 0 {
		n, err := s.writer.Write(dataToWrite)
		if err != nil {
			return fmt.Errorf("write SOL output: %w", err)
		}

		if n != len(dataToWrite) {
			return fmt.Errorf("write SOL output: %w", io.ErrShortWrite)
		}
	}
	s.sequenceNumber = sequenceNumber
	s.acceptedCharCount = uint8(len(data))

	return nil
}

func readSOLInput(ctx context.Context, in io.Reader) <-chan solConsoleInput {
	events := make(chan solConsoleInput, 256)
	go func() {
		reader := bufio.NewReader(in)
		for {
			value, err := reader.ReadByte()
			if ctx.Err() != nil {
				return
			}
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
	inputCtx, cancelInput := context.WithCancel(ctx)
	defer cancelInput()

	input := readSOLInput(inputCtx, in)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var localSeq uint8 = 1
	output := solOutputStream{writer: out}

	sendPacket := func(chars []byte) (*types.SOLPayloadResponse, error) {
		req := &types.SOLPayloadRequest{
			SOLPayloadPacket: types.SOLPayloadPacket{
				SequenceNumber:         localSeq,
				AckedSequenceNumber:    output.sequenceNumber,
				AcceptedCharacterCount: output.acceptedCharCount,
				CharacterData:          chars,
			},
		}
		res, err := c.SOLPayload(ctx, req)
		if err != nil {
			return nil, err
		}
		localSeq = nextSOLSequenceNumber(localSeq)

		if err := output.process(res); err != nil {
			return nil, err
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
