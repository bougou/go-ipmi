package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *GetMessageFlagsRequest) Command() ipmi.Command {
	return ipmi.CommandGetMessageFlags
}

func (req *GetMessageFlagsRequest) Pack() []byte {
	return []byte{}
}

func (res *GetMessageFlagsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := ipmi.UnpackUint8(msg, 0)
	res.OEM2Available = ipmi.IsBit7Set(b)
	res.OEM1Available = ipmi.IsBit6Set(b)
	res.OEM0Available = ipmi.IsBit5Set(b)
	res.WatchdogPreTimeoutInterruptOccurred = ipmi.IsBit3Set(b)
	res.EventMessageBufferFull = ipmi.IsBit1Set(b)
	res.ReceiveMessageQueueAvailable = ipmi.IsBit0Set(b)
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
