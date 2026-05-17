package app

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

type GetCommandSubfunctionEnablesRequest struct {
	ChannelNumber uint8

	NetFn ipmi.NetFn
	LUN   uint8
	Cmd   uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type GetCommandSubfunctionEnablesResponse struct {
	SubfunctionEnables []bool
}

func (req *GetCommandSubfunctionEnablesRequest) Command() ipmi.Command {
	return ipmi.CommandGetCommandSubfunctionEnables
}

func (req *GetCommandSubfunctionEnablesRequest) Pack() []byte {
	out := make([]byte, 7)
	ipmi.PackUint8(req.ChannelNumber, out, 0)

	ipmi.PackUint8(uint8(req.NetFn)&0x3f, out, 1)
	ipmi.PackUint8(req.LUN&0x03, out, 2)
	ipmi.PackUint8(req.Cmd, out, 3)

	if uint8(req.NetFn) == 0x2c {
		ipmi.PackUint8(req.CodeForNetFn2C, out, 4)
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		ipmi.PackUint24L(req.OEMIANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetCommandSubfunctionEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	if len(msg) > 4 && len(msg) < 8 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 8)
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
		res.SubfunctionEnables[i*8+0] = ipmi.IsBit0Set(b)
		res.SubfunctionEnables[i*8+1] = ipmi.IsBit1Set(b)
		res.SubfunctionEnables[i*8+2] = ipmi.IsBit2Set(b)
		res.SubfunctionEnables[i*8+3] = ipmi.IsBit3Set(b)
		res.SubfunctionEnables[i*8+4] = ipmi.IsBit4Set(b)
		res.SubfunctionEnables[i*8+5] = ipmi.IsBit5Set(b)
		res.SubfunctionEnables[i*8+6] = ipmi.IsBit6Set(b)
		res.SubfunctionEnables[i*8+7] = ipmi.IsBit7Set(b)
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
