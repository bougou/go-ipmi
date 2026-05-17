// Package protocol provides IPMI wire-format framing shared by both client and
// server: IPMB checksums, LAN message assembly/disassembly, and RMCP+
// packet building. Neither the client nor the server needs to duplicate this
// logic; they each import this package and carry their session-specific state
// separately.
package protocol

// Checksum computes the two's-complement checksum over data, as required by
// the IPMB and IPMI LAN message formats (section 13.8 of the IPMI 2.0 spec).
func Checksum(data []byte) uint8 {
	var sum uint8
	for _, b := range data {
		sum += b
	}
	return -sum
}

// BMCAddr is the IPMI BMC slave address on the IPMB bus.
const BMCAddr = uint8(0x20)

// RemoteConsoleAddr is the conventional remote console requester address.
const RemoteConsoleAddr = uint8(0x20)

// ParseIPMIRequest parses a raw IPMI LAN request message (post-RMCP,
// post-session header) and returns its NetFn, command code, data body, and
// requester sequence number.
//
// The IPMI LAN message format (Table 13-8) is:
//
//	[0]  rsSA  (BMC = 0x20)
//	[1]  netFn/rsLUN  (netFn in bits [7:2])
//	[2]  checksum1
//	[3]  rqSA
//	[4]  rqSeq/rqLUN  (seq in bits [7:2])
//	[5]  cmd
//	[6..n-1] data
//	[n]  checksum2
func ParseIPMIRequest(msg []byte) (netFn, cmd uint8, data []byte, seq uint8, ok bool) {
	if len(msg) < 7 {
		return 0, 0, nil, 0, false
	}
	netFn = (msg[1] >> 2) & 0x3F
	cmd = msg[5]
	seq = (msg[4] >> 2) & 0x3F
	end := len(msg) - 1 // strip trailing checksum
	if end <= 6 {
		data = nil
	} else {
		data = msg[6:end]
	}
	return netFn, cmd, data, seq, true
}

// BuildIPMIResponse constructs a raw IPMI LAN response message.
// reqNetFn is the *request* NetFn; the response uses reqNetFn|1.
func BuildIPMIResponse(reqNetFn, cmd, seq, cc uint8, data []byte) []byte {
	rspNetFn := (reqNetFn | 1) << 2
	msg := make([]byte, 8+len(data))
	msg[0] = RemoteConsoleAddr // rqSA
	msg[1] = rspNetFn
	msg[2] = Checksum(msg[0:2])
	msg[3] = BMCAddr  // rsSA
	msg[4] = seq << 2 // rqSeq, rqLUN = 0
	msg[5] = cmd
	msg[6] = cc
	copy(msg[7:], data)
	msg[7+len(data)] = Checksum(msg[3:])
	return msg
}
