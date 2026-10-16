package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/bougou/go-ipmi/pkg/protocol"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	defaultSOLTransmitRetryCount    = 3
	defaultSOLTransmitRetryInterval = 100 * time.Millisecond
	solSequenceMask                 = 0x0f
	// Sequence numbers wrap from 15 to 1, so 0 is not part of the cycle and the
	// cycle holds 15 values. A packet more than half the cycle ahead of the last
	// one received is read as trailing it rather than skipping past it.
	solSequenceSpace         = 15
	solSequenceHalfSpace     = solSequenceSpace / 2
	solStatusDeactivating    = 1 << 4 // IPMI v2.0 Table 15-2, status bit 4
	solStatusTransmitOverrun = 1 << 3 // IPMI v2.0 Table 15-2, status bit 3
)

// SOLStream exchanges console data over an already active SOL payload. Input is transmitted
// unchanged, unlike [Client.SOLActivate], this method does not interpret terminal escape sequences
// or process signals.
//
// The exchange model follows §15.3/§15.11: the BMC pushes console output
// asynchronously; the client sends packets only to carry input (non-zero
// sequence) or to acknowledge output (ACK-only, sequence 0h, no reply
// expected). No packets are sent while the session is idle.
func (c *Client) SOLStream(ctx context.Context, in io.Reader, out io.Writer) error {
	return c.runSOLStream(ctx, in, out, nil)
}

type solConsoleInput struct {
	value byte
	err   error
}

type solStream struct {
	client        *Client
	input         io.Reader
	output        io.Writer
	interrupt     <-chan os.Signal
	maxInputChars int

	localSequenceNumber  uint8
	remoteSequenceNumber uint8
	acceptedCharCount    uint8
	// ackPending records that accepted console output has not been
	// acknowledged to the BMC yet. Set on the run loop by processOutput,
	// cleared by sendAck or by a data packet carrying the same ACK state.
	ackPending bool

	// outputCh receives every inbound SOL packet in wire order from the
	// receive loop, both asynchronous pushes and responses to our packets.
	// Deep enough that the loop never blocks behind output processing; on
	// overflow packets drop and the BMC's retry retransmits them (§15.11).
	outputCh chan *types.SOLPayloadResponse
}

// nextSOLSequenceNumber returns the next SOL sequence number, wrapping from 15 to 1.
func nextSOLSequenceNumber(sequenceNumber uint8) uint8 {
	sequenceNumber = (sequenceNumber + 1) & solSequenceMask
	if sequenceNumber == 0 {
		return 1
	}

	return sequenceNumber
}

// solSequenceDistance returns how many steps separate sequenceNumber from
// previous, counting forward through the wrapping sequence space. Both must be
// valid data sequence numbers (1 to 15). A result above solSequenceHalfSpace
// means sequenceNumber trails previous.
func solSequenceDistance(sequenceNumber, previous uint8) uint8 {
	return uint8((int(sequenceNumber) - int(previous) + solSequenceSpace) % solSequenceSpace)
}

// processOutput validates one inbound SOL packet and writes its console data
// to the stream's output, tracking what the BMC must be acknowledged for.
// It runs only on the stream's run loop.
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
	case sequenceNumber == nextSOLSequenceNumber(s.remoteSequenceNumber):
	case solSequenceDistance(sequenceNumber, s.remoteSequenceNumber) > solSequenceHalfSpace:
		// A retransmission of an already-acknowledged packet (the ACK was
		// lost, or the BMC's retry raced our ACK). §15.11 requires
		// acknowledging every non-zero packet: re-ACK with the current
		// state — a Table 15-3 "Completion ACK" stops the retry engine.
		s.ackPending = true
		return nil
	default:
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
	s.ackPending = true

	return nil
}

// registerSOLStream activates the stream and hands socket reads to the
// receive loop that dispatches inbound SOL packets to it.
func (c *Client) registerSOLStream(ctx context.Context, stream *solStream) error {
	if c.udpClient == nil {
		return fmt.Errorf("SOL stream requires a LAN connection")
	}
	c.lock()
	if c.activeSOLStream != nil {
		c.unlock()
		return fmt.Errorf("SOL stream is already active")
	}
	c.activeSOLStream = stream
	c.unlock()
	// Never hold c.l across the exchange gate (taken by StartRecvLoop):
	// in-flight exchanges acquire c.l from inside the gate.
	if err := c.udpClient.StartRecvLoop(ctx, c.dispatchInboundDatagram); err != nil {
		c.lock()
		c.activeSOLStream = nil
		c.unlock()
		return fmt.Errorf("start SOL receive loop: %w", err)
	}
	return nil
}

func (c *Client) unregisterSOLStream() {
	c.udpClient.StopRecvLoop()
	c.lock()
	c.activeSOLStream = nil
	c.unlock()
}

// dispatchInboundDatagram routes one inbound datagram to the active SOL
// stream. The receive loop calls it for every datagram, including ones a
// waiting exchange claimed, so that piggybacked console data (§15.11) is
// processed in wire order. Non-SOL datagrams are the rule here (every
// command response passes through) and return early and silently.
func (c *Client) dispatchInboundDatagram(recv []byte) {
	// RMCP version and auth type stay explicit: ParseRMCPPlusHeader will
	// read a v1.5 header as RMCP+.
	if len(recv) < 16 || recv[0] != types.RmcpVersion || recv[4] != byte(types.AuthTypeRMCPPlus) {
		return
	}
	_, _, payloadType, _, _, ok := protocol.ParseRMCPPlusHeader(recv)
	if !ok || types.PayloadType(payloadType) != types.PayloadTypeSOL {
		return
	}
	sol, err := c.parseInboundSOLPacket(recv)
	if err != nil {
		c.DebugfYellow("drop recv: SOL dispatch parse failed: %s\n", err)
		return
	}
	c.lock()
	stream := c.activeSOLStream
	c.unlock()
	if stream == nil {
		c.DebugfYellow("drop recv: SOL packet with no active stream\n")
		return
	}
	stream.enqueueOutput(&types.SOLPayloadResponse{SOLPayloadPacket: *sol})
}

// parseInboundSOLPacket extracts the SOL payload packet from an inbound
// RMCP+ datagram. It shares the matcher's parse path so both directions of
// the receive loop classify packets identically.
func (c *Client) parseInboundSOLPacket(recv []byte) (*types.SOLPayloadPacket, error) {
	rmcp := &types.Rmcp{}
	if err := rmcp.Unpack(recv); err != nil {
		return nil, fmt.Errorf("rmcp unpack failed: %w", err)
	}
	if rmcp.Session20 == nil {
		return nil, fmt.Errorf("not an RMCP+ session packet")
	}
	hdr := rmcp.Session20.SessionHeader20
	if hdr.PayloadType != types.PayloadTypeSOL {
		return nil, fmt.Errorf("payload type %s is not SOL", hdr.PayloadType)
	}
	payload := rmcp.Session20.SessionPayload
	if hdr.PayloadEncrypted {
		// decryptPayload reads session keys; take the session lock.
		c.lock()
		d, err := c.decryptPayload(payload)
		c.unlock()
		if err != nil {
			return nil, fmt.Errorf("decrypt SOL payload failed: %w", err)
		}
		payload = d
	}
	var sol types.SOLPayloadPacket
	if err := sol.Unpack(payload); err != nil {
		return nil, fmt.Errorf("SOL payload unpack failed: %w", err)
	}
	return &sol, nil
}

func (s *solStream) enqueueOutput(response *types.SOLPayloadResponse) {
	select {
	case s.outputCh <- response:
	default:
		// Backlog overrun: drop. The BMC retransmits unacknowledged output
		// per its retry policy (§15.11), so a transient stall self-heals.
	}
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
	interrupt <-chan os.Signal,
) error {
	if c.Interface != InterfaceLanplus {
		return fmt.Errorf("SOL stream requires IPMI v2.0 (RMCP+); use InterfaceLanplus (-I lanplus)")
	}
	// Table 24-2 includes the four-byte SOL header in the inbound payload size.
	// Table 15-2 limits the acknowledged character count to 255. Keep one-byte
	// input when activation did not provide a usable size.
	c.lock()
	maxInputChars := max(1, min(int(c.session.v20.solInboundPayloadSize)-4, 255))
	c.unlock()
	stream := &solStream{
		client:              c,
		input:               in,
		output:              out,
		interrupt:           interrupt,
		maxInputChars:       maxInputChars,
		localSequenceNumber: 1,
		outputCh:            make(chan *types.SOLPayloadResponse, 256),
	}
	if err := c.registerSOLStream(ctx, stream); err != nil {
		return err
	}
	defer c.unregisterSOLStream()

	return stream.run(ctx)
}

// exchangePacket sends one data-carrying SOL packet and waits for the BMC's
// acknowledgement (§15.11: every non-zero-sequence console packet is ACKed).
// The request also carries the current acknowledgement state for console
// output, so output received so far is ACKed without a separate packet.
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
	// The packet carried the acknowledgement for output received so far.
	s.ackPending = false
	return res, nil
}

// sendAck acknowledges accepted console output with an ACK-only packet
// (sequence 0h), which the BMC does not answer (§15.11).
func (s *solStream) sendAck(ctx context.Context) error {
	_, err := s.client.SOLPayload(ctx, &types.SOLPayloadRequest{
		SOLPayloadPacket: types.SOLPayloadPacket{
			SequenceNumber:         0,
			AckedSequenceNumber:    s.remoteSequenceNumber,
			AcceptedCharacterCount: s.acceptedCharCount,
		},
	})
	if err != nil {
		return fmt.Errorf("send SOL ACK failed: %w", err)
	}
	return nil
}

// takeOutput writes SOL packets already queued by the receive loop, in the
// order they arrived. Non-blocking: the caller has just finished an exchange,
// and that exchange does not return until its own response has been queued.
func (s *solStream) takeOutput() error {
	for {
		select {
		case response := <-s.outputCh:
			if err := s.processOutput(response); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (s *solStream) sendData(ctx context.Context, chars []byte) error {
	pending := chars
	noProgressRetries := 0
	for len(pending) > 0 {
		// Retry only the data the BMC did not accept, using a fresh SOL sequence.
		res, err := s.exchangePacket(ctx, pending)
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.client.lanClosed:
			return net.ErrClosed
		case <-s.interrupt:
			return nil
		case response := <-s.outputCh:
			if err := s.processOutput(response); err != nil {
				return err
			}
		case consoleInput := <-input:
			chars := make([]byte, 0, s.maxInputChars)
		collectInput:
			for consoleInput.err == nil {
				chars = append(chars, consoleInput.value)
				if len(chars) == s.maxInputChars {
					break
				}
				// Batch queued input without waiting for another keystroke.
				select {
				case consoleInput = <-input:
				default:
					break collectInput
				}
			}
			// Deliver queued data before reporting EOF or another reader error.
			if err := s.sendData(ctx, chars); err != nil {
				return err
			}
			if consoleInput.err != nil {
				// The keystroke's reply may carry console output. It is
				// queued on outputCh before sendData returns; leaving it
				// there drops the last screen update.
				if err := s.takeOutput(); err != nil {
					return err
				}
				if s.ackPending {
					if err := s.sendAck(ctx); err != nil {
						return err
					}
					s.ackPending = false
				}
				if consoleInput.err == io.EOF {
					return nil
				}
				return consoleInput.err
			}
		}
		if s.ackPending {
			if err := s.sendAck(ctx); err != nil {
				return err
			}
			s.ackPending = false
		}
	}
}
