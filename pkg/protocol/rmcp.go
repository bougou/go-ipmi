package protocol

import "encoding/binary"

// BuildRMCPPlusPacket assembles a complete RMCP+ (IPMI 2.0) wire packet.
//
//	RMCP header  (4 bytes): version=0x06, reserved, seq=0xFF, class=0x07
//	Session20 header (12 bytes): authType=0x06, payloadType|flags,
//	  sessionID(LE), seqNum(LE), payloadLen(LE)
//	Payload (len bytes)
//
// The payloadType and flags parameters are raw bytes; callers typically pass
// uint8(types.PayloadTypeIPMI) | types.PayloadFlagAuthenticated.
func BuildRMCPPlusPacket(payloadType, payloadFlags uint8, sessionID, seq uint32, payload []byte) []byte {
	pkt := make([]byte, 4+12+len(payload))
	// RMCP header
	pkt[0] = 0x06 // version
	pkt[1] = 0x00 // reserved
	pkt[2] = 0xFF // sequence: no ACK requested
	pkt[3] = 0x07 // class IPMI
	// Session20 header
	pkt[4] = 0x06                       // AuthType = RMCPPlus
	pkt[5] = payloadType | payloadFlags // payload type + encrypted/authenticated flags
	binary.LittleEndian.PutUint32(pkt[6:10], sessionID)
	binary.LittleEndian.PutUint32(pkt[10:14], seq)
	binary.LittleEndian.PutUint16(pkt[14:16], uint16(len(payload)))
	copy(pkt[16:], payload)
	return pkt
}

// ParseRMCPPlusHeader extracts the session-layer fields from a raw RMCP+
// packet starting at offset 4 (after the 4-byte RMCP header).
//
// Returns sessionID, seqNum, payloadType (without flag bits), flags byte,
// payload slice, and ok=false if the buffer is too short.
func ParseRMCPPlusHeader(pkt []byte) (sessionID, seqNum uint32, payloadType, flags uint8, payload []byte, ok bool) {
	// pkt is the full packet including RMCP header
	if len(pkt) < 16 {
		return 0, 0, 0, 0, nil, false
	}
	hdr := pkt[4:] // session header starts after 4-byte RMCP header
	flags = hdr[1]
	payloadType = flags & 0x3F
	sessionID = binary.LittleEndian.Uint32(hdr[2:6])
	seqNum = binary.LittleEndian.Uint32(hdr[6:10])
	payloadLen := binary.LittleEndian.Uint16(hdr[10:12])
	if len(hdr) < 12+int(payloadLen) {
		return 0, 0, 0, 0, nil, false
	}
	payload = hdr[12 : 12+payloadLen]
	return sessionID, seqNum, payloadType, flags, payload, true
}
