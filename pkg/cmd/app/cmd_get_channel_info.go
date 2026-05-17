package app

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 22.24 Get Channel Info Command
type GetChannelInfoRequest struct {
	ChannelNumber uint8
}

type GetChannelInfoResponse struct {
	ActualChannelNumber uint8
	ChannelMedium       ipmi.ChannelMedium   // Channel Medium Type Numbers
	ChannelProtocol     ipmi.ChannelProtocol // Channel Protocol Type Numbers
	SessionSupport      uint8
	ActiveSessionCount  uint8
	VendorID            uint32 // (IANA Enterprise Number) for OEM/Organization that specified the Channel Protocol.

	// Auxiliary Channel Info
	Auxiliary []byte // Auxiliary Channel Info Raw Data, 2 bytes

	// For Channel = Fh (System Interface)
	SMSInterruptType                InterruptType
	EventMessageBufferInterruptType InterruptType
}

type InterruptType uint8

func (typ InterruptType) String() string {
	if typ <= 0x0f {
		return fmt.Sprintf("IRQ %d", uint8(typ))
	}
	if typ >= 0x10 && typ <= 0x13 {
		return fmt.Sprintf("PCI %X", uint8(typ))
	}
	if typ == 0x14 {
		return "SMI"
	}
	if typ == 0x15 {
		return "SCI"
	}
	if typ >= 20 && typ <= 0x5f {
		return fmt.Sprintf("system interrupt %d", uint8(typ-32))
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
	return []byte{req.ChannelNumber}
}

func (req *GetChannelInfoRequest) Command() ipmi.Command {
	return ipmi.CommandGetChannelInfo
}

func (res *GetChannelInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChannelInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 9 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 9)
	}
	res.ActualChannelNumber, _, _ = ipmi.UnpackUint8(msg, 0)

	cm, _, _ := ipmi.UnpackUint8(msg, 1)
	res.ChannelMedium = ipmi.ChannelMedium(cm)

	cp, _, _ := ipmi.UnpackUint8(msg, 2)
	res.ChannelProtocol = ipmi.ChannelProtocol(cp)

	s, _, _ := ipmi.UnpackUint8(msg, 3)
	res.SessionSupport = s >> 6
	res.ActiveSessionCount = s & 0x3f

	res.VendorID, _, _ = ipmi.UnpackUint24L(msg, 4)
	res.Auxiliary, _, _ = ipmi.UnpackBytes(msg, 7, 2)

	res.SMSInterruptType = InterruptType(res.Auxiliary[0])
	res.EventMessageBufferInterruptType = InterruptType(res.Auxiliary[1])

	return nil
}

func (res *GetChannelInfoResponse) Format() string {
	return "" +
		fmt.Sprintf("Channel %#02x info      :\n", res.ActualChannelNumber) +
		fmt.Sprintf("  Channel Medium Type   : %s\n", res.ChannelMedium) +
		fmt.Sprintf("  Channel Protocol Type : %s\n", res.ChannelProtocol) +
		fmt.Sprintf("  Session Support       : %d\n", res.SessionSupport) +
		fmt.Sprintf("  Active Session Count  : %d\n", res.ActiveSessionCount) +
		fmt.Sprintf("  Protocol Vendor ID    : %d\n", res.VendorID)
}
