package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.6 Get Message Command
)

type GetMessageRequest struct {
	// empty
}

type GetMessageResponse struct {
	ChannelNumber uint8
	MessageData   []byte
}

func (req *GetMessageRequest) Command() types.Command {
	return types.CommandGetMessage
}

func (req *GetMessageRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "data not available (queue / buffer empty)",
	}
}

func (res *GetMessageResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ChannelNumber, _, _ = types.UnpackUint8(msg, 0)
	res.MessageData, _, _ = types.UnpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetMessageResponse) Format() string {
	return ""
}
