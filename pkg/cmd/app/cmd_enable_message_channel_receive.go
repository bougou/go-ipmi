package app

import (
	"github.com/bougou/go-ipmi/pkg/types"
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

func (req *EnableMessageChannelReceiveRequest) Command() types.Command {
	return types.CommandEnableMessageChannelReceive
}

func (req *EnableMessageChannelReceiveRequest) Pack() []byte {
	out := make([]byte, 2)
	types.PackUint8(req.ChannelNumber, out, 0)
	types.PackUint8(req.ChannelState, out, 1)
	return out
}

func (res *EnableMessageChannelReceiveResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.ChannelNumber, _, _ = types.UnpackUint8(msg, 0)

	b, _, _ := types.UnpackUint8(msg, 1)
	res.ChannelEnabled = types.IsBit0Set(b)
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
