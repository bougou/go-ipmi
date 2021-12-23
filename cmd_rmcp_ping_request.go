package ipmi

import "fmt"

// RmcpPingRequest
// 13.2.3 RMCP/ASF Presence Ping Message
type RmcpPingRequest struct {
	// empty
}

type RmcpPingResponse struct {
	// If no OEM-specific capabilities exist, this field contains the ASF IANA (4542) and the OEM-defined field is set to all zeroes (00000000h). Otherwise, this field contains the OEM's IANA Enterprise Number and the OEM-defined field contains the OEM-specific capabilities.
	OEMIANA uint32

	// Not used for IPMI.
	// This field can contain OEM-defined values; the definition of these values is left to the manufacturer identified by the preceding IANA Enterprise number.
	OEMDefined uint32

	// 81h for IPMI
	// [7] 1b = IPMI Supported
	// [6:4] Reserved
	// [3:0] 0001b = ASF Version 1.0
	SupportedEntities uint8
	IPMISupported     bool

	// [7] Set to 1b if RMCP security extensions are supported[1]
	// [6] Reserved for future definition by ASF specification. Set to 0b.
	// [5] Set to 1b if DMTF DASH is supported
	// [4:0] Reserved for future definition by ASF specification, set to 00000b
	SupportedInteractions           uint8
	RMCPSecurityExtensionsSupported bool
	DMTFDashSupported               bool

	// Reserved for future definition by ASF specification,
	// set to 00 00 00 00 00 00h, six bytes
	Reserverd []byte
}

func (req *RmcpPingRequest) Pack() []byte {
	return nil
}

func (req *RmcpPingRequest) Command() Command {
	return CommandNone
}

func (res *RmcpPingResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShort
	}
	res.OEMIANA, _, _ = unpackUint32L(msg, 0)
	res.OEMDefined, _, _ = unpackUint32L(msg, 4)
	res.SupportedEntities, _, _ = unpackUint8(msg, 8)
	res.IPMISupported = res.SupportedEntities&0x80 == 0x80
	res.SupportedInteractions, _, _ = unpackUint8(msg, 9)
	res.RMCPSecurityExtensionsSupported = res.SupportedInteractions&0x80 == 0x80
	res.DMTFDashSupported = res.SupportedInteractions&0x20 == 0x20
	res.Reserverd, _, _ = unpackBytes(msg, 10, 6)
	return nil
}

func (r *RmcpPingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *RmcpPingResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) RmcpPing() (response *RmcpPingResponse, err error) {
	request := &RmcpPingRequest{}
	response = &RmcpPingResponse{}
	err = c.Exchange(request, response)

	return
}
