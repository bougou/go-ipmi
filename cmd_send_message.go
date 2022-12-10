package ipmi

// 22.7 Send Message Command
type SendMessageRequest struct {
	// [7:6] 00b = No tracking
	// 01b = Track Request.
	// 10b = Send Raw. (optional)
	// 11b = reserved
	TrackMask uint8

	Encrypted bool

	Authenticated bool

	ChannelNumber uint8

	// Todo
	MessageData []byte
}

type SendMessageResponse struct {
	// This data will only be present when using the Send Message command to
	// originate requests from IPMB or PCI Management Bus to other channels
	// such as LAN or serial/modem. It is not present in the response to a
	// Send Message command delivered via the System Interface.
	Data []byte
}

func (req SendMessageRequest) Command() Command {
	return CommandSendMessage
}

func (req *SendMessageRequest) Pack() []byte {
	out := make([]byte, 1+len(req.MessageData))

	var b uint8 = req.ChannelNumber
	if req.Authenticated {
		b = setBit4(b)
	}
	if req.Encrypted {
		b = setBit5(b)
	}
	b |= (req.TrackMask << 6)

	packUint8(b, out, 0)
	packBytes(req.MessageData, out, 1)

	return out
}

func (res *SendMessageResponse) Unpack(msg []byte) error {
	res.Data, _, _ = unpackBytes(msg, 0, len(msg))
	return nil
}

func (*SendMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Invalid Session Handle. The Session handle does not match up with any currently active sessions for this channel.",
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
	}
}

func (res *SendMessageResponse) Format() string {
	return ""
}

func (c *Client) SendMessage(channelNumber uint8, authenticated bool, encrypted bool, trackMask uint8, data []byte) (response *SendMessageResponse, err error) {
	request := &SendMessageRequest{
		ChannelNumber: channelNumber,
		Authenticated: authenticated,
		Encrypted:     encrypted,
		TrackMask:     trackMask,
		MessageData:   data,
	}
	response = &SendMessageResponse{}
	err = c.Exchange(request, response)
	return
}
