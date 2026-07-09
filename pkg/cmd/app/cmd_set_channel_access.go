package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.22 Set Channel Access Command
)

type SetChannelAccessRequest struct {
	ChannelNumber uint8

	// [7:6] - 00b = don't set or change Channel Access
	//         01b = set non-volatile Channel Access according to bits [5:0]
	//         10b = set volatile (active) setting of Channel Access according to bit [5:0]
	//         11b = reserved
	AccessOption         uint8
	DisablePEFAlerting   bool
	DisablePerMsgAuth    bool
	DisableUserLevelAuth bool
	AccessMode           types.ChannelAccessMode

	PrivilegeOption   uint8
	MaxPrivilegeLevel uint8
}

type SetChannelAccessResponse struct {
}

func (req *SetChannelAccessRequest) Pack() []byte {
	out := make([]byte, 3)

	types.PackUint8(req.ChannelNumber, out, 0)

	var b = req.AccessOption << 6
	if req.DisablePEFAlerting {
		b = types.SetBit5(b)
	}
	if req.DisablePerMsgAuth {
		b = types.SetBit4(b)
	}
	if req.DisableUserLevelAuth {
		b = types.SetBit3(b)
	}
	b |= uint8(req.AccessMode) & 0x07
	types.PackUint8(b, out, 1)

	var b2 = req.PrivilegeOption << 6
	b2 |= req.MaxPrivilegeLevel & 0x3f
	types.PackUint8(b2, out, 2)

	return out
}

func (req *SetChannelAccessRequest) Command() types.Command {
	return types.CommandSetChannelAccess
}

func (res *SetChannelAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x82: "set not supported on selected channel (e.g. channel is session-less.)",
		0x83: "access mode not supported",
	}
}

func (res *SetChannelAccessResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetChannelAccessResponse) Format() string {
	return ""
}
