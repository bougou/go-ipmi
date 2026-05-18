package types

// SOLPayloadPacket is the IPMI 2.0 SOL payload data format.
type SOLPayloadPacket struct {
	SequenceNumber         uint8
	AckedSequenceNumber    uint8
	AcceptedCharacterCount uint8

	// ControlByte is written to wire byte 3 for outbound console operation bits.
	// For inbound packets, Unpack stores the full BMC status byte.
	ControlByte uint8

	// NACK sets bit 6 in byte 3 when packing and is populated from bit 6 when unpacking.
	NACK bool

	CharacterData []byte
}

// Pack encodes the SOL payload packet.
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

// Unpack decodes the SOL payload packet.
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

// SOLPayloadRequest is an RMCP+ session request with payload type SOL.
type SOLPayloadRequest struct {
	SOLPayloadPacket
}

// Command returns the pseudo-command metadata for SOL payload packets.
func (r *SOLPayloadRequest) Command() Command {
	return CommandSOLPayload
}

// Pack encodes the SOL payload request.
func (r *SOLPayloadRequest) Pack() []byte {
	return r.SOLPayloadPacket.Pack()
}

// SOLPayloadResponse is the BMC reply to a SOL payload packet.
type SOLPayloadResponse struct {
	SOLPayloadPacket
}

// Unpack decodes the SOL payload response.
func (r *SOLPayloadResponse) Unpack(data []byte) error {
	return r.SOLPayloadPacket.Unpack(data)
}

// CompletionCodes returns command-specific completion codes.
func (r *SOLPayloadResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

// Format returns a human-readable representation.
func (r *SOLPayloadResponse) Format() string {
	return ""
}
