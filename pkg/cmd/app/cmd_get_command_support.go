package app

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 21.3 Get Command Support Command
)

type GetCommandSupportRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            ipmi.NetFn
	LUN              uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type GetCommandSupportResponse struct {
	// Todo
	CommandSupportMask []byte
}

type CommandRangeMask uint8

const (
	CommandRangeMask007F uint8 = 0x00
	CommandRangeMask80FF uint8 = 0x01
)

func (req *GetCommandSupportRequest) Command() ipmi.Command {
	return ipmi.CommandGetCommandSupport
}

func (req *GetCommandSupportRequest) Pack() []byte {
	out := make([]byte, 6)
	ipmi.PackUint8(req.ChannelNumber, out, 0)

	netfn := uint8(req.NetFn) & (uint8(req.CommandRangeMask) << 6)
	ipmi.PackUint8(netfn, out, 1)

	ipmi.PackUint8(req.LUN&0x03, out, 2)

	if uint8(req.NetFn) == 0x2c {
		ipmi.PackUint8(req.CodeForNetFn2C, out, 3)
		return out[0:4]
	}

	if uint8(req.NetFn) == 0x2e {
		ipmi.PackUint24L(req.OEMIANA, out, 3)
		return out[0:6]
	}

	return out[0:3]
}

func (res *GetCommandSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	res.CommandSupportMask, _, _ = ipmi.UnpackBytes(msg, 0, 16)
	return nil
}

func (*GetCommandSupportResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetCommandSupportResponse) Format() string {
	// Todo
	return ""
}
