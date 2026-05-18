package protocol

import "encoding/binary"

// RMCP+ payload type constants (IPMI 2.0 spec Table 13-16).
const (
	PayloadIPMI                = uint8(0x00)
	PayloadSOL                 = uint8(0x01)
	PayloadOEM                 = uint8(0x02)
	PayloadOpenSessionRequest  = uint8(0x10)
	PayloadOpenSessionResponse = uint8(0x11)
	PayloadRAKPMessage1        = uint8(0x12)
	PayloadRAKPMessage2        = uint8(0x13)
	PayloadRAKPMessage3        = uint8(0x14)
	PayloadRAKPMessage4        = uint8(0x15)
)

// PayloadEncryptedFlag is ORed into the payload-type byte when the payload is
// AES-CBC-128 encrypted (bit 7).
const PayloadEncryptedFlag = uint8(0x80)

// PayloadAuthenticatedFlag is ORed into the payload-type byte when an HMAC
// integrity trailer is appended (bit 6).
const PayloadAuthenticatedFlag = uint8(0x40)

// BuildRMCPPlusPacket assembles a complete RMCP+ (IPMI 2.0) wire packet.
//
//	RMCP header  (4 bytes): version=0x06, reserved, seq=0xFF, class=0x07
//	Session20 header (12 bytes): authType=0x06, payloadType|flags,
//	  reserved×2, sessionID(LE), seqNum(LE), payloadLen(LE)
//	Payload (len bytes)
func BuildRMCPPlusPacket(payloadType, payloadFlags uint8, sessionID, seq uint32, payload []byte) []byte {
	pkt := make([]byte, 4+14+len(payload))
	// RMCP header
	pkt[0] = 0x06 // version
	pkt[1] = 0x00 // reserved
	pkt[2] = 0xFF // sequence: no ACK requested
	pkt[3] = 0x07 // class IPMI
	// Session20 header
	pkt[4] = 0x06                       // AuthType = RMCPPlus
	pkt[5] = payloadType | payloadFlags // payload type + encrypted/authenticated flags
	pkt[6] = 0x00                       // reserved
	pkt[7] = 0x00                       // reserved
	binary.LittleEndian.PutUint32(pkt[8:12], sessionID)
	binary.LittleEndian.PutUint32(pkt[12:16], seq)
	binary.LittleEndian.PutUint16(pkt[16:18], uint16(len(payload)))
	copy(pkt[18:], payload)
	return pkt
}

// ParseRMCPPlusHeader extracts the session-layer fields from a raw RMCP+
// packet starting at offset 4 (after the 4-byte RMCP header).
//
// Returns sessionID, seqNum, payloadType (without flag bits), flags byte,
// payload slice, and ok=false if the buffer is too short.
func ParseRMCPPlusHeader(pkt []byte) (sessionID, seqNum uint32, payloadType, flags uint8, payload []byte, ok bool) {
	// pkt is the full packet including RMCP header
	if len(pkt) < 18 {
		return 0, 0, 0, 0, nil, false
	}
	hdr := pkt[4:] // session header starts after 4-byte RMCP header
	flags = hdr[1]
	payloadType = flags & 0x3F
	sessionID = binary.LittleEndian.Uint32(hdr[4:8])
	seqNum = binary.LittleEndian.Uint32(hdr[8:12])
	payloadLen := binary.LittleEndian.Uint16(hdr[10:12])
	if len(hdr) < 12+int(payloadLen) {
		return 0, 0, 0, 0, nil, false
	}
	payload = hdr[12 : 12+payloadLen]
	return sessionID, seqNum, payloadType, flags, payload, true
}
