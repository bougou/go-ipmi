package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.4 Get Message Flags Command
)

type GetMessageFlagsRequest struct {
	// empty
}

type GetMessageFlagsResponse struct {
	OEM2Available                       bool
	OEM1Available                       bool
	OEM0Available                       bool
	WatchdogPreTimeoutInterruptOccurred bool
	EventMessageBufferFull              bool
	ReceiveMessageQueueAvailable        bool // One or more messages ready for reading from Receive Message Queue
}

func (req *GetMessageFlagsRequest) Command() types.Command {
	return types.CommandGetMessageFlags
}

func (req *GetMessageFlagsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageFlagsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := types.UnpackUint8(msg, 0)
	res.OEM2Available = types.IsBit7Set(b)
	res.OEM1Available = types.IsBit6Set(b)
	res.OEM0Available = types.IsBit5Set(b)
	res.WatchdogPreTimeoutInterruptOccurred = types.IsBit3Set(b)
	res.EventMessageBufferFull = types.IsBit1Set(b)
	res.ReceiveMessageQueueAvailable = types.IsBit0Set(b)
	return nil
}

func (*GetMessageFlagsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetMessageFlagsResponse) Format() string {
	// Todo
	return ""
}
