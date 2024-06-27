package ipmi

// 22.22 Set Channel Access Command
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
	AccessMode           ChannelAccessMode

	PrivilegeOption   uint8
	MaxPrivilegeLevel uint8
}

type SetChannelAccessResponse struct {
}

func (req *SetChannelAccessRequest) Pack() []byte {
	out := make([]byte, 3)

	packUint8(req.ChannelNumber, out, 0)

	var b = req.AccessOption << 6
	if req.DisablePEFAlerting {
		b = setBit5(b)
	}
	if req.DisablePerMsgAuth {
		b = setBit4(b)
	}
	if req.DisableUserLevelAuth {
		b = setBit3(b)
	}
	b |= uint8(req.AccessMode) & 0x07
	packUint8(b, out, 1)

	var b2 = req.PrivilegeOption << 6
	b2 |= req.MaxPrivilegeLevel & 0x3f
	packUint8(b2, out, 2)

	return out
}

func (req *SetChannelAccessRequest) Command() Command {
	return CommandSetChannelAccess
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

func (c *Client) SetChannelAccess(request *SetChannelAccessRequest) (response *SetChannelAccessResponse, err error) {
	response = &SetChannelAccessResponse{}
	err = c.Exchange(request, response)
	return
}
