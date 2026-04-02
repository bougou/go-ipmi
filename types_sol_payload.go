package ipmi

import "context"

// 15.9 SOL Payload Data Format
type SOLPayloadPacket struct {
	SequenceNumber         uint8
	AckedSequenceNumber    uint8
	AcceptedCharacterCount uint8

	// ControlByte is written to wire byte [3] for outbound (console operation bits). For inbound, Unpack sets
	// this to the full status byte from the BMC.
	ControlByte uint8

	// NACK on Pack ORs 0x40 into byte [3] (console NACK). On Unpack, set from BMC response (0x40).
	NACK bool

	CharacterData []byte
}

func (p *SOLPayloadPacket) Pack() []byte {
	out := make([]byte, 4+len(p.CharacterData))
	out[0] = p.SequenceNumber & 0x0f
	out[1] = p.AckedSequenceNumber & 0x0f
	out[2] = p.AcceptedCharacterCount

	b3 := p.ControlByte
	if p.NACK {
		b3 |= 0x40
	}
	out[3] = b3
	copy(out[4:], p.CharacterData)
	return out
}

func (p *SOLPayloadPacket) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	p.SequenceNumber = msg[0] & 0x0f
	p.AckedSequenceNumber = msg[1] & 0x0f
	p.AcceptedCharacterCount = msg[2]

	p.ControlByte = msg[3]
	p.NACK = msg[3]&0x40 != 0

	if len(msg) > 4 {
		p.CharacterData = append([]byte{}, msg[4:]...)
	} else {
		p.CharacterData = nil
	}
	return nil
}

// SOLPayloadRequest is an RMCP+ session request with PayloadType SOL (not an IPMI command frame).
type SOLPayloadRequest struct {
	SOLPayloadPacket
}

func (r *SOLPayloadRequest) Command() Command {
	return CommandSOLPayload
}

func (r *SOLPayloadRequest) Pack() []byte {
	return r.SOLPayloadPacket.Pack()
}

// SOLPayloadResponse is the BMC reply to a SOL payload packet.
type SOLPayloadResponse struct {
	SOLPayloadPacket
}

func (r *SOLPayloadResponse) Unpack(data []byte) error {
	return r.SOLPayloadPacket.Unpack(data)
}

func (r *SOLPayloadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (r *SOLPayloadResponse) Format() string {
	return ""
}

// SOLPayload exchanges a single SOL payload packet over an active RMCP+ session.
func (c *Client) SOLPayload(ctx context.Context, request *SOLPayloadRequest) (response *SOLPayloadResponse, err error) {
	response = &SOLPayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
