package ipmi

import "context"

// 21.9 Set Configurable Command Sub-function Enables Command
type SetCommandSubfunctionEnablesRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            NetFn
	LUN              uint8
	Cmd              uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only

	SubfunctionEnables []bool
}

type SetCommandSubfunctionEnablesResponse struct {
}

func (req *SetCommandSubfunctionEnablesRequest) Command() Command {
	return CommandSetCommandSubfunctionEnables
}

func (req *SetCommandSubfunctionEnablesRequest) Pack() []byte {
	var maxLength int = 15
	out := make([]byte, maxLength)

	out[0] = req.ChannelNumber
	out[1] = uint8(req.NetFn) & (uint8(req.CommandRangeMask) << 6)
	out[2] = req.LUN & 0x03
	out[3] = req.Cmd

	var startIndexOfEnables int = 4

	if uint8(req.NetFn) == 0x2c {
		out[4] = req.CodeForNetFn2C
		startIndexOfEnables = 5
	}

	if uint8(req.NetFn) == 0x2e {
		packUint24L(req.OEMIANA, out, 4)
		startIndexOfEnables = 7
	}

	enables := req.SubfunctionEnables
	if len(req.SubfunctionEnables) > 64 {
		enables = enables[:64]
	}
	// Every 8 elements from the enables slice pack into 1 byte.
	enableBytesLength := len(enables)/8 + 1

	for i := 0; i < enableBytesLength; i++ {
		var b uint8
		b = setOrClearBit0(b, req.SubfunctionEnables[i*8+0])
		b = setOrClearBit1(b, req.SubfunctionEnables[i*8+1])
		b = setOrClearBit2(b, req.SubfunctionEnables[i*8+2])
		b = setOrClearBit3(b, req.SubfunctionEnables[i*8+3])
		b = setOrClearBit4(b, req.SubfunctionEnables[i*8+4])
		b = setOrClearBit5(b, req.SubfunctionEnables[i*8+5])
		b = setOrClearBit6(b, req.SubfunctionEnables[i*8+6])
		b = setOrClearBit7(b, req.SubfunctionEnables[i*8+7])

		out[startIndexOfEnables+i] = b
	}

	return out[0 : startIndexOfEnables+enableBytesLength]

}

func (res *SetCommandSubfunctionEnablesResponse) Unpack(msg []byte) error {
	return nil
}

func (*SetCommandSubfunctionEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{
		0x80: "attempt to enable an unsupported or un-configurable sub-function.",
	}
}

func (res *SetCommandSubfunctionEnablesResponse) Format() string {
	return ""
}

func (c *Client) SetCommandSubfunctionEnables(ctx context.Context, request *SetCommandSubfunctionEnablesRequest) (response *SetCommandSubfunctionEnablesResponse, err error) {
	response = &SetCommandSubfunctionEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
