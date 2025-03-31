package ipmi

import "context"

const (
	IPMIRequesterSequenceMax uint8 = 0x3f // RequesterSequence only occupy 6 bits
)

// 13.8 IPMI LAN Message Format
type IPMIRequest struct {
	// SlaveAddress or SoftwareID
	// Responder's Slave Address. 1 byte. LS bit is 0 for Slave Addresses and 1 for Software IDs. Upper 7-bits hold Slave Address or Software ID, respectively. This byte is always 20h when the BMC is the responder.
	ResponderAddr uint8 // SlaveAddress or SoftwareID

	// The lower 2-bits of the netFn byte identify the logical unit number, which provides further sub-addressing within the target node.
	NetFn        NetFn // (even) / rsLUN
	ResponderLUN uint8 // lower 2 bits

	// Checksum1 is filled by calling ComputeChecksum method
	Checksum1 uint8

	// SlaveAddress or SoftwareID
	// Requester's Address. 1 byte. LS bit is 0 for Slave Addresses and 1 for Software IDs. Upper 7-bits hold Slave Address or Software ID, respectively. This byte is always 20h when the BMC is the requester.
	RequesterAddr uint8 // rqSA

	RequesterSequence uint8 // rqSeq, occupies the highest 6 bits, (so should left shit 2 bits)
	RequesterLUN      uint8 // rqLUN, occupies the lowest 2 bits

	Command uint8 // Command ID

	// Command Request Body Data defined by each command.
	CommandData []byte // optional, 0 or more

	// Checksum2 is filled by calling ComputeChecksum method
	Checksum2 uint8
}

// IPMIResponse represent IPMI PayloadType msg response
type IPMIResponse struct {
	// Requester's Address. 1 byte. LS bit is 0 for Slave Addresses and 1 for Software IDs. Upper 7-bits hold Slave Address or Software ID, respectively. This byte is always 20h when the BMC is the requester.
	RequesterAddr uint8 // SlaveAddress or SoftwareID

	// Network Function code
	// The lower 2-bits of the netFn byte identify the logical unit number, which provides further sub-addressing within the target node.
	NetFn NetFn // (odd) higher 6 bits
	// Requester's LUN
	RequestLUN uint8 // lower 2 bits

	// 8-bit checksum algorithm: Initialize checksum to 0. For each byte, checksum = (checksum + byte) modulo 256. Then checksum = - checksum. When the checksum and the bytes are added together, modulo 256, the result should be 0.
	Checksum1 uint8

	// Responder's Slave Address. 1 byte. LS bit is 0 for Slave Addresses and 1 for Software IDs. Upper 7-bits hold Slave Address or Software ID, respectively. This byte is always 20h when the BMC is the responder.
	ResponderAddr uint8 // SlaveAddress or SoftwareID

	// Sequence number. This field is used to verify that a response is for a particular instance of a request. Refer to [IPMB] for additional information on use and operation of the Seq field.
	RequesterSequence uint8 // higher 6 bits
	ResponderLUN      uint8 // lower 2 bits

	Command uint8

	// Completion code returned in the response to indicated success/failure status of the request.
	CompletionCode uint8

	// Response Data
	Data []byte // optional

	Checksum2 uint8
}

func (req *IPMIRequest) Pack() []byte {
	msgLen := 6 + len(req.CommandData) + 1
	msg := make([]byte, msgLen)

	packUint8(req.ResponderAddr, msg, 0)

	netFn := uint8(req.NetFn) << 2
	resLun := req.ResponderLUN & 0x03
	packUint8(netFn|resLun, msg, 1)

	packUint8(req.Checksum1, msg, 2)
	packUint8(req.RequesterAddr, msg, 3)

	var seq uint8 = req.RequesterSequence << 2
	reqLun := req.RequesterLUN & 0x03
	packUint8(seq|reqLun, msg, 4)

	packUint8(uint8(req.Command), msg, 5)

	if req.CommandData != nil {
		packBytes(req.CommandData, msg, 6)
	}

	packUint8(req.Checksum2, msg, msgLen-1)
	return msg
}

func (req *IPMIRequest) ComputeChecksum() {
	// 8-bit checksum algorithm: Initialize checksum to 0. For each byte, checksum = (checksum + byte) modulo 256. Then checksum = - checksum. When the checksum and the bytes are added together, modulo 256, the result should be 0.
	//
	// the position end is not included
	var checksumFn = func(msg []byte, start int, end int) uint8 {
		c := 0
		for i := start; i < end; i++ {
			c = (c + int(msg[i])) % 256
		}
		return -uint8(c)
	}

	tempData := req.Pack()

	cs1Start, cs1End := 0, 2
	req.Checksum1 = checksumFn(tempData, cs1Start, cs1End)

	cs2Start, cs2End := 3, len(tempData)-1
	req.Checksum2 = checksumFn(tempData, cs2Start, cs2End)
}

func (res *IPMIResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShortWith(len(msg), 7)
	}

	res.RequesterAddr, _, _ = unpackUint8(msg, 0)

	b, _, _ := unpackUint8(msg, 1)
	res.NetFn = NetFn(b >> 2) // the most 6 bit
	res.RequestLUN = b & 0x03 // the least 2 bit

	res.Checksum1, _, _ = unpackUint8(msg, 2)
	res.ResponderAddr, _, _ = unpackUint8(msg, 3)

	b4, _, _ := unpackUint8(msg, 4)
	res.RequesterSequence = b4 >> 2
	res.ResponderLUN = b4 & 0x03

	res.Command, _, _ = unpackUint8(msg, 5)
	res.CompletionCode, _, _ = unpackUint8(msg, 6)

	if len(msg) == 7 {
		// Response with only Completion Code
		return nil
	}

	dataLen := len(msg) - 7 - 1
	res.Data, _, _ = unpackBytes(msg, 7, dataLen)
	res.Checksum2, _, _ = unpackUint8(msg, len(msg)-1)

	return nil
}

// BuildIPMIRequest creates IPMIRequest for a Command Request.
// It also fills the Checksum1 and Checksum2 fields of IPMIRequest.
func (c *Client) BuildIPMIRequest(ctx context.Context, reqCmd Request) (*IPMIRequest, error) {
	c.lock()
	defer c.unlock()

	ipmiReq := &IPMIRequest{
		ResponderAddr: c.responderAddr,

		NetFn:        reqCmd.Command().NetFn,
		ResponderLUN: c.responderLUN,

		RequesterAddr: c.requesterAddr,

		RequesterSequence: c.session.ipmiSeq,
		RequesterLUN:      c.requesterLUN,

		Command:     reqCmd.Command().ID,
		CommandData: reqCmd.Pack(),
	}

	c.session.ipmiSeq += 1
	if c.session.ipmiSeq > IPMIRequesterSequenceMax {
		c.session.ipmiSeq = 1
	}

	ipmiReq.ComputeChecksum()

	return ipmiReq, nil
}

// AllCC returns all possible completion codes for the specified response.
// i.e.:
//
//	the generic completion codes for all ipmi cmd response
//	+
//	the specific completion codes for specified cmd response.
func AllCC(response Response) map[uint8]string {
	out := map[uint8]string{}
	for k, v := range CC {
		out[k] = v
	}
	for k, v := range response.CompletionCodes() {
		out[k] = v
	}
	return out
}

// StrCC return the description of ccode for the specified response.
// The available completion codes set consists of general completion codes (CC) for all
// commands response and specific completion codes for this response.
func StrCC(response Response, ccode uint8) string {
	s, ok := AllCC(response)[ccode]
	if ok {
		return s
	}
	return "unknown completion code"
}
