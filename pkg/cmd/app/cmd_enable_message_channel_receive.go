package app

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 22.5 Enable Message Channel Receive Command
)

type EnableMessageChannelReceiveRequest struct {
	ChannelNumber uint8

	// [7:2] - reserved
	// [1:0] - 00b = disable channel
	//         01b = enable channel
	//         10b = get channel enable/disable state
	//         11b = reserved
	ChannelState uint8
}

type EnableMessageChannelReceiveResponse struct {
	ChannelNumber uint8

	ChannelEnabled bool
}

func (req *EnableMessageChannelReceiveRequest) Command() ipmi.Command {
	return ipmi.CommandEnableMessageChannelReceive
}

func (req *EnableMessageChannelReceiveRequest) Pack() []byte {
	out := make([]byte, 2)
	ipmi.PackUint8(req.ChannelNumber, out, 0)
	ipmi.PackUint8(req.ChannelState, out, 1)
	return out
}

func (res *EnableMessageChannelReceiveResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ChannelNumber, _, _ = ipmi.UnpackUint8(msg, 0)

	b, _, _ := ipmi.UnpackUint8(msg, 1)
	res.ChannelEnabled = ipmi.IsBit0Set(b)
	return nil
}

func (*EnableMessageChannelReceiveResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *EnableMessageChannelReceiveResponse) Format() string {
	// Todo
	return ""
}
