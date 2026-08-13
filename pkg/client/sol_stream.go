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
	defaultSOLTransmitRetryCount        = 3
	defaultSOLTransmitRetryInterval     = 100 * time.Millisecond
	defaultSOLAcknowledgementRetryCount = 3
	solSequenceMask                     = 0x0f
	solStatusDeactivating               = 1 << 4 // IPMI v2.0 Table 15-2, status bit 4
	solStatusTransmitOverrun            = 1 << 3 // IPMI v2.0 Table 15-2, status bit 3
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

type solStream struct {
	client       *Client
	input        io.Reader
	output       io.Writer
	pollInterval time.Duration
	interrupt    <-chan os.Signal

	localSequenceNumber  uint8
	remoteSequenceNumber uint8
	acceptedCharCount    uint8
}

// nextSOLSequenceNumber returns the next SOL sequence number, wrapping from 15 to 1.
func nextSOLSequenceNumber(sequenceNumber uint8) uint8 {
	sequenceNumber = (sequenceNumber + 1) & solSequenceMask
	if sequenceNumber == 0 {
		return 1
	}

	return sequenceNumber
}

func (s *solStream) processOutput(response *types.SOLPayloadResponse) error {
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
	case s.remoteSequenceNumber == 0:
	case sequenceNumber == s.remoteSequenceNumber:
		// A retransmission may append data. Write only bytes not previously delivered.
		if len(data) < int(s.acceptedCharCount) {
			return fmt.Errorf("BMC shortened retried SOL sequence %d from %d to %d bytes",
				sequenceNumber, s.acceptedCharCount, len(data))
		}
		writeOffset = int(s.acceptedCharCount)
	case sequenceNumber != nextSOLSequenceNumber(s.remoteSequenceNumber):
		// A gap means BMC output was lost before it reached the writer.
		return fmt.Errorf("unexpected BMC SOL sequence %d after %d", sequenceNumber, s.remoteSequenceNumber)
	}

	dataToWrite := data[writeOffset:]
	if len(dataToWrite) > 0 {
		n, err := s.output.Write(dataToWrite)
		if err != nil {
			return fmt.Errorf("write SOL output: %w", err)
		}

		if n != len(dataToWrite) {
			return fmt.Errorf("write SOL output: %w", io.ErrShortWrite)
		}
	}
	s.remoteSequenceNumber = sequenceNumber
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
	stream := &solStream{
		client:              c,
		input:               in,
		output:              out,
		pollInterval:        pollInterval,
		interrupt:           interrupt,
		localSequenceNumber: 1,
	}
	return stream.run(ctx)
}

func (s *solStream) exchangePacket(ctx context.Context, chars []byte) (*types.SOLPayloadResponse, error) {
	req := &types.SOLPayloadRequest{
		SOLPayloadPacket: types.SOLPayloadPacket{
			SequenceNumber:         s.localSequenceNumber,
			AckedSequenceNumber:    s.remoteSequenceNumber,
			AcceptedCharacterCount: s.acceptedCharCount,
			CharacterData:          chars,
		},
	}
	res, err := s.client.SOLPayload(ctx, req)
	if err != nil {
		return nil, err
	}
	s.localSequenceNumber = nextSOLSequenceNumber(s.localSequenceNumber)

	if err := s.processOutput(res); err != nil {
		return nil, err
	}
	return res, nil
}

// acknowledgeOutput immediately acknowledges and drains BMC console output.
func (s *solStream) acknowledgeOutput(ctx context.Context, response *types.SOLPayloadResponse) error {
	repeatedResponses := 0
	for response.SequenceNumber&solSequenceMask != 0 {
		previousSequenceNumber := s.remoteSequenceNumber
		previousAcceptedCharCount := s.acceptedCharCount

		// A nonzero empty packet carries the acknowledgement and requires a response from the BMC.
		var err error
		response, err = s.exchangePacket(ctx, nil)
		if err != nil {
			return err
		}
		if response.SequenceNumber&solSequenceMask == 0 {
			return nil
		}

		if s.remoteSequenceNumber == previousSequenceNumber &&
			s.acceptedCharCount == previousAcceptedCharCount {
			repeatedResponses++
			if repeatedResponses > defaultSOLAcknowledgementRetryCount {
				return fmt.Errorf("BMC repeated SOL sequence %d after %d acknowledgement retries",
					s.remoteSequenceNumber, defaultSOLAcknowledgementRetryCount)
			}
		} else {
			repeatedResponses = 0
		}
	}
	return nil
}

func (s *solStream) sendPacket(ctx context.Context, chars []byte) (*types.SOLPayloadResponse, error) {
	response, err := s.exchangePacket(ctx, chars)
	if err != nil {
		return nil, err
	}
	if err := s.acknowledgeOutput(ctx, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (s *solStream) sendData(ctx context.Context, chars []byte) error {
	pending := chars
	noProgressRetries := 0
	for len(pending) > 0 {
		// Retry only the data the BMC did not accept, using a fresh SOL sequence.
		res, err := s.sendPacket(ctx, pending)
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

func (s *solStream) run(ctx context.Context) error {
	inputCtx, cancelInput := context.WithCancel(ctx)
	defer cancelInput()

	input := readSOLInput(inputCtx, s.input)
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	if _, err := s.sendPacket(ctx, nil); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.interrupt:
			return nil
		case consoleInput := <-input:
			if consoleInput.err != nil {
				if consoleInput.err == io.EOF {
					return nil
				}
				return consoleInput.err
			}

			if err := s.sendData(ctx, []byte{consoleInput.value}); err != nil {
				return err
			}
		case <-ticker.C:
			if _, err := s.sendPacket(ctx, nil); err != nil {
				return err
			}
		}
	}
}
