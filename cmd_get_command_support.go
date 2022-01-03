package ipmi

// 21.3 Get Command Support Command
type GetCommandSupportRequest struct {
	ChannelNumber uint8

	CommandRangeMask CommandRangeMask
	NetFn            NetFn
	LUN              uint8

	CodeForNetFn2C uint8
	OEM_IANA       uint32 // 3 bytes only
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

func (req *GetCommandSupportRequest) Command() Command {
	return CommandGetCommandSupport
}

func (req *GetCommandSupportRequest) Pack() []byte {
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

func (res *GetCommandSupportResponse) Unpack(msg []byte) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShort
	}

	res.CommandSupportMask, _, _ = unpackBytes(msg, 0, 16)
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

func (c *Client) GetCommandSupport(channelNumber uint8, commandRangeMask CommandRangeMask, netFn NetFn, lun uint8, code uint8, oemIANA uint32) (response *GetCommandSupportResponse, err error) {
	request := &GetCommandSupportRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEM_IANA:         oemIANA,
	}
	response = &GetCommandSupportResponse{}
	err = c.Exchange(request, response)
	return
}
