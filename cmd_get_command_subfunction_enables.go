package ipmi

import "context"

type GetCommandSubfunctionEnablesRequest struct {
	ChannelNumber uint8

	NetFn NetFn
	LUN   uint8
	Cmd   uint8

	CodeForNetFn2C uint8
	OEMIANA        uint32 // 3 bytes only
}

type GetCommandSubfunctionEnablesResponse struct {
	SubfunctionEnables []bool
}

func (req *GetCommandSubfunctionEnablesRequest) Command() Command {
	return CommandGetCommandSubfunctionEnables
}

func (req *GetCommandSubfunctionEnablesRequest) Pack() []byte {
	out := make([]byte, 7)
	packUint8(req.ChannelNumber, out, 0)

	packUint8(uint8(req.NetFn)&0x3f, out, 1)
	packUint8(req.LUN&0x03, out, 2)
	packUint8(req.Cmd, out, 3)

	if uint8(req.NetFn) == 0x2c {
		packUint8(req.CodeForNetFn2C, out, 4)
		return out[0:5]
	}

	if uint8(req.NetFn) == 0x2e {
		packUint24L(req.OEMIANA, out, 4)
		return out[0:7]
	}

	return out[0:4]
}

func (res *GetCommandSubfunctionEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	if len(msg) > 4 && len(msg) < 8 {
		return ErrUnpackedDataTooShortWith(len(msg), 8)
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
		res.SubfunctionEnables[i*8+0] = isBit0Set(b)
		res.SubfunctionEnables[i*8+1] = isBit1Set(b)
		res.SubfunctionEnables[i*8+2] = isBit2Set(b)
		res.SubfunctionEnables[i*8+3] = isBit3Set(b)
		res.SubfunctionEnables[i*8+4] = isBit4Set(b)
		res.SubfunctionEnables[i*8+5] = isBit5Set(b)
		res.SubfunctionEnables[i*8+6] = isBit6Set(b)
		res.SubfunctionEnables[i*8+7] = isBit7Set(b)
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

func (c *Client) GetCommandSubfunctionEnables(ctx context.Context, request *GetCommandSubfunctionEnablesRequest) (response *GetCommandSubfunctionEnablesResponse, err error) {
	response = &GetCommandSubfunctionEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
