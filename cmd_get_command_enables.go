package ipmi

import "context"

// 21.8 Get Command Enables Command
type GetCommandEnablesRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            NetFn
	LUN              uint8

	CodeForNetFn2C uint8
	OEM_IANA       uint32 // 3 bytes only
}

type GetCommandEnablesResponse struct {
	// Todo
	CommandEnableMask []byte
}

func (req *GetCommandEnablesRequest) Command() Command {
	return CommandGetCommandEnables
}

func (req *GetCommandEnablesRequest) Pack() []byte {
	out := make([]byte, 6)
	packUint8(req.ChannelNumber, out, 0)

	netfn := uint8(req.NetFn) & (uint8(req.CommandRangeMask) << 6)
	packUint8(netfn, out, 1)

	packUint8(req.LUN&0x03, out, 2)

	if uint8(req.NetFn) == 0x2c {
		packUint8(req.CodeForNetFn2C, out, 3)
		return out[0:4]
	}

	if uint8(req.NetFn) == 0x2e {
		packUint24L(req.OEM_IANA, out, 3)
		return out[0:6]
	}

	return out[0:3]
}

func (res *GetCommandEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	res.CommandEnableMask, _, _ = unpackBytes(msg, 0, 16)
	return nil
}

func (*GetCommandEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetCommandEnablesResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetCommandEnables(ctx context.Context, channelNumber uint8, commandRangeMask CommandRangeMask, netFn NetFn, lun uint8, code uint8, oemIANA uint32) (response *GetCommandEnablesResponse, err error) {
	request := &GetCommandEnablesRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEM_IANA:         oemIANA,
	}
	response = &GetCommandEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
