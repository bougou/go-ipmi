package ipmi

// 22.7 Send Message Command
type SendMessageRequest struct {
	Tracked bool

	Encrypted bool

	Authenticated bool

	ChannelNumber uint8

	Data []byte
}

type SendMessageResponse struct {
	NetFn

	LUN

	Command

	CompletionCode

	// 	Data []byte
}

func (r *SendMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Invalid Session Handle. The session handle does not match up with any currently active sessions for this channel.",
		0x81: "Lost Arbitration",
		0x82: "Bus Error",
		0x83: "NAK on Write",
	}
}

func (c *Client) SendMessage() {
}
