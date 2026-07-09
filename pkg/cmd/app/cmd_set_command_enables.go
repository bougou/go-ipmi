package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 21.7 Set Command Enables Command
)

type SetCommandEnablesRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            types.NetFn
	LUN              uint8

	// if CommandRangeMask == CommandRangeMask007F
	CommandsMaskBytes [16]byte

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type SetCommandEnablesResponse struct {
}

func (req *SetCommandEnablesRequest) Command() types.Command {
	return types.CommandSetCommandEnables
}

func (req *SetCommandEnablesRequest) Pack() []byte {
	out := make([]byte, 22)

	out[0] = req.ChannelNumber
	out[1] = (uint8(req.NetFn) & 0x3f) | (uint8(req.CommandRangeMask) << 6)
	out[2] = req.LUN & 0x03
	types.PackBytes(req.CommandsMaskBytes[:], out, 3)

	if uint8(req.NetFn) == 0x2c {
		types.PackUint8(req.CodeForNetFn2C, out, 19)
		return out[0:20]
	}

	if uint8(req.NetFn) == 0x2e {
		types.PackUint24L(req.OEMIANA, out, 19)
		return out[0:22]
	}

	return out[0:20]
}

func (res *SetCommandEnablesResponse) Unpack(msg []byte) error {
	return nil
}

func (*SetCommandEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{
		0x80: "attempt to enable an unsupported or un-configurable command",
	}
}

func (res *SetCommandEnablesResponse) Format() string {
	return ""
}
