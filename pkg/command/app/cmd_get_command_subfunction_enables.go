package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

type GetCommandSubfunctionEnablesRequest struct {
	ChannelNumber uint8

	NetFn types.NetFn
	LUN   uint8
	Cmd   uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type GetCommandSubfunctionEnablesResponse struct {
	SubfunctionEnables []bool
}

func (req *GetCommandSubfunctionEnablesRequest) Command() types.Command {
	return types.CommandGetCommandSubfunctionEnables
}

func (req *GetCommandSubfunctionEnablesRequest) Pack() []byte {
	out := make([]byte, 7)
	types.PackUint8(req.ChannelNumber, out, 0)

	types.PackUint8(uint8(req.NetFn)&0x3f, out, 1)
	types.PackUint8(req.LUN&0x03, out, 2)
	types.PackUint8(req.Cmd, out, 3)

	if uint8(req.NetFn) == 0x2c {
		types.PackUint8(req.CodeForNetFn2C, out, 4)
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		types.PackUint24L(req.OEMIANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetCommandSubfunctionEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	if len(msg) > 4 && len(msg) < 8 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	var enables []bool
	var bytesLen int
	if len(msg) == 4 {
		enables = make([]bool, 32)
		bytesLen = 4
	} else if len(msg) >= 8 {
		enables = make([]bool, 64)
		bytesLen = 8
	}

	for i := 0; i < bytesLen; i++ {
		b := msg[i]
		enables[i*8+0] = types.IsBit0Set(b)
		enables[i*8+1] = types.IsBit1Set(b)
		enables[i*8+2] = types.IsBit2Set(b)
		enables[i*8+3] = types.IsBit3Set(b)
		enables[i*8+4] = types.IsBit4Set(b)
		enables[i*8+5] = types.IsBit5Set(b)
		enables[i*8+6] = types.IsBit6Set(b)
		enables[i*8+7] = types.IsBit7Set(b)
	}

	res.SubfunctionEnables = enables
	return nil
}

func (*GetCommandSubfunctionEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetCommandSubfunctionEnablesResponse) Format() string {
	// Todo
	return ""
}
