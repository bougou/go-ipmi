package ipmi

import "fmt"

// 22.24 Get Channel Info Command
type GetChannelInfoRequest struct {
	ChannnelNumber uint8
}

type GetChannelInfoResponse struct {
	ActualChannelNumber uint8
	ChannelMedium       ChannelMedium   // Channel Medium Type Numbers
	ChannelProtocol     ChannelProtocol // Channel Protocol Type Numbers
	SessionSupport      uint8
	ActiveSessionCount  uint8
	VendorID            uint32 // (IANA Enterprise Number) for OEM/Organization that specified the Channel Protocol.

	// Auxiliray Channel Info
	Auxiliray []byte // Auxiliray Channel Info Raw Data, 2 bytes

	// For Channel = Fh (System Interface)
	SMSInterruptType                InterruptType
	EventMessageBufferInterruptType InterruptType
}

type InterruptType uint8

func (typ InterruptType) String() string {
	if typ >= 0x00 && typ <= 0x0f {
		return fmt.Sprintf("IRQ %d", typ)
	}
	if typ >= 0x10 && typ <= 0x13 {
		return fmt.Sprintf("PCI %X", typ)
	}
	if typ == 0x14 {
		return "SMI"
	}
	if typ == 0x15 {
		return "SCI"
	}
	if typ >= 20 && typ <= 0x5f {
		return fmt.Sprintf("system interrupt %d", typ-32)
	}
	if typ == 0x60 {
		return "assigned by ACPI / Plug in Play BIOS"
	}
	if typ == 0xff {
		return "no interrupt"
	}
	return "reserved"
}

func (req *GetChannelInfoRequest) Pack() []byte {
	return []byte{req.ChannnelNumber}
}

func (req *GetChannelInfoRequest) Command() Command {
	return CommandGetChannelInfo
}

func (res *GetChannelInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChannelInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return ErrUnpackedDataTooShort
	}
	res.ActualChannelNumber, _, _ = unpackUint8(msg, 0)

	cm, _, _ := unpackUint8(msg, 1)
	res.ChannelMedium = ChannelMedium(cm)

	cp, _, _ := unpackUint8(msg, 2)
	res.ChannelProtocol = ChannelProtocol(cp)

	s, _, _ := unpackUint8(msg, 3)
	res.SessionSupport = s >> 6
	res.ActiveSessionCount = s & 0x3f

	res.VendorID, _, _ = unpackUint24L(msg, 4)
	res.Auxiliray, _, _ = unpackBytes(msg, 7, 2)

	res.SMSInterruptType = InterruptType(res.Auxiliray[0])
	res.EventMessageBufferInterruptType = InterruptType(res.Auxiliray[1])

	return nil
}

func (res *GetChannelInfoResponse) Format() string {
	return fmt.Sprintf(`Channel %#02x info:
  Channel Medium Type   : %s
  Channel Protocol Type : %s
  Session Support       : %d
  Active Session Count  : %d
  Protocol Vendor ID    : %d`,
		res.ActualChannelNumber,
		res.ChannelMedium,
		res.ChannelProtocol,
		res.SessionSupport,
		res.ActiveSessionCount,
		res.VendorID,
	)
}

func (c *Client) GetChannelInfo(channelNumber uint8) (response *GetChannelInfoResponse, err error) {
	request := &GetChannelInfoRequest{
		ChannnelNumber: channelNumber,
	}
	response = &GetChannelInfoResponse{}
	err = c.Exchange(request, response)
	return
}
