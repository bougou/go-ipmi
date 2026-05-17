package app

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 21.6 Get Configurable Command Sub-functions Command
type GetConfigurableCommandSubfunctionsRequest struct {
	ChannelNumber uint8

	NetFn ipmi.NetFn
	LUN   uint8

	Cmd uint8

	CodeForNetFn2C uint8  // For Network Function = 2Ch
	OEMIANA        uint32 // For Network Function = 2Eh
}

type GetConfigurableCommandSubfunctionsResponse struct {
	// the index corresponds to sub-function number
	// index 0 -> sub-function 0
	// index 1 -> sub-function 1
	SubfunctionsSupport []bool
}

func (req *GetConfigurableCommandSubfunctionsRequest) Command() ipmi.Command {
	return ipmi.CommandGetConfigurableCommandSubfunctions
}

func (req *GetConfigurableCommandSubfunctionsRequest) Pack() []byte {
	out := make([]byte, 7)

	out[0] = req.ChannelNumber
	out[1] = byte(req.NetFn)
	out[2] = req.LUN & 0x03
	out[3] = req.Cmd

	if uint8(req.NetFn) == 0x2c {
		out[4] = req.CodeForNetFn2C
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		ipmi.PackUint24L(req.OEMIANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetConfigurableCommandSubfunctionsResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	supports := make([]bool, 64)

	for i := 0; i < 4; i++ {
		b := msg[i]
		res.SubfunctionsSupport[i*8+0] = ipmi.IsBit0Set(b)
		res.SubfunctionsSupport[i*8+1] = ipmi.IsBit1Set(b)
		res.SubfunctionsSupport[i*8+2] = ipmi.IsBit2Set(b)
		res.SubfunctionsSupport[i*8+3] = ipmi.IsBit3Set(b)
		res.SubfunctionsSupport[i*8+4] = ipmi.IsBit4Set(b)
		res.SubfunctionsSupport[i*8+5] = ipmi.IsBit5Set(b)
		res.SubfunctionsSupport[i*8+6] = ipmi.IsBit6Set(b)
		res.SubfunctionsSupport[i*8+7] = ipmi.IsBit7Set(b)
	}

	if len(msg) == 4 {
		res.SubfunctionsSupport = supports[0:32]
		return nil
	}

	if len(msg) > 4 && len(msg) < 8 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 8)
	}

	for i := 4; i < 8; i++ {
		b := msg[i]
		res.SubfunctionsSupport[i*8+0] = ipmi.IsBit0Set(b)
		res.SubfunctionsSupport[i*8+1] = ipmi.IsBit1Set(b)
		res.SubfunctionsSupport[i*8+2] = ipmi.IsBit2Set(b)
		res.SubfunctionsSupport[i*8+3] = ipmi.IsBit3Set(b)
		res.SubfunctionsSupport[i*8+4] = ipmi.IsBit4Set(b)
		res.SubfunctionsSupport[i*8+5] = ipmi.IsBit5Set(b)
		res.SubfunctionsSupport[i*8+6] = ipmi.IsBit6Set(b)
		res.SubfunctionsSupport[i*8+7] = ipmi.IsBit7Set(b)
	}

	res.SubfunctionsSupport = supports[0:64]
	return nil
}

func (*GetConfigurableCommandSubfunctionsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetConfigurableCommandSubfunctionsResponse) Format() string {
	out := ""
	for k, v := range res.SubfunctionsSupport {
		out += fmt.Sprintf("sub-function %2d : %s\n", k, ipmi.FormatBool(v, "supported", "unsupported"))
	}
	return out
}
