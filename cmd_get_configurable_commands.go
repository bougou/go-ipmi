package ipmi

// 21.5 Get Configurable Commands Command
type GetConfigurableCommandsRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            NetFn
	LUN              uint8

	CodeForNetFn2C uint8
	OEM_IANA       uint32 // 3 bytes only
}

type GetConfigurableCommandsResponse struct {
	// Todo
	CommandSupportMask []byte
}

func (req *GetConfigurableCommandsRequest) Command() Command {
	return CommandGetConfigurableCommands
}

func (req *GetConfigurableCommandsRequest) Pack() []byte {
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

func (res *GetConfigurableCommandsResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShort
	}

	res.CommandSupportMask, _, _ = unpackBytes(msg, 0, 16)
	return nil
}

func (*GetConfigurableCommandsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetConfigurableCommandsResponse) Format() string {
	// Todo
	return ""
}

func (c *Client) GetConfigurableCommands(channelNumber uint8, commandRangeMask CommandRangeMask, netFn NetFn, lun uint8, code uint8, oemIANA uint32) (response *GetConfigurableCommandsResponse, err error) {
	request := &GetConfigurableCommandsRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEM_IANA:         oemIANA,
	}
	response = &GetConfigurableCommandsResponse{}
	err = c.Exchange(request, response)
	return
}
