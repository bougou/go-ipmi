package protocol

import "encoding/binary"

// BuildASFPresencePong builds an RMCP/ASF Presence Pong response for the
// given tag byte. The Pong advertises IPMI support per the ASF 2.0 spec.
func BuildASFPresencePong(tag byte) []byte {
	resp := make([]byte, 28)
	// RMCP header
	resp[0] = 0x06 // version
	resp[1] = 0x00 // reserved
	resp[2] = 0xFF // seq (no ACK)
	resp[3] = 0x06 // class ASF
	// ASF header: IANA + type + tag + reserved + data-length
	binary.BigEndian.PutUint32(resp[4:8], 4542) // ASF IANA
	resp[8] = 0x40                              // Presence Pong
	resp[9] = tag
	resp[10] = 0x00 // reserved
	resp[11] = 16   // data length
	// Pong data (16 bytes)
	binary.BigEndian.PutUint32(resp[12:16], 4542) // IANA
	binary.BigEndian.PutUint32(resp[16:20], 0)    // OEM IANA
	resp[20] = 0x81                               // supported entities: IPMI
	resp[21] = 0x00                               // supported interactions
	// bytes 22-27: reserved zeros
	return resp
}
