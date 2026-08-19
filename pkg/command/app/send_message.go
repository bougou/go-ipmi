package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.7 Send Message Command
)

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

func (req SendMessageRequest) Command() types.Command {
	return types.CommandSendMessage
}

func (req *SendMessageRequest) Pack() []byte {
	out := make([]byte, 1+len(req.MessageData))

	var b uint8 = req.ChannelNumber
	if req.Authenticated {
		b = types.SetBit4(b)
	}
	if req.Encrypted {
		b = types.SetBit5(b)
	}
	b |= (req.TrackMask << 6)

	types.PackUint8(b, out, 0)
	types.PackBytes(req.MessageData, out, 1)

	return out
}

func (res *SendMessageResponse) Unpack(msg []byte) error {
	res.Data, _, _ = types.UnpackBytes(msg, 0, len(msg))
	return nil
}

func (res *SendMessageResponse) Format() string {
	return ""
}
