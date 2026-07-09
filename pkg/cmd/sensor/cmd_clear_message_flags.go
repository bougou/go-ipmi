package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 22.3 Clear Message Flags Command
)

type ClearMessageFlagsRequest struct {
	ClearOEM2                            bool
	ClearOEM1                            bool
	ClearOEM0                            bool
	ClearWatchdogPreTimeoutInterruptFlag bool
	ClearEventMessageBuffer              bool
	ClearReceiveMessageQueue             bool
}

type ClearMessageFlagsResponse struct {
}

func (req *ClearMessageFlagsRequest) Command() types.Command {
	return types.CommandClearMessageFlags
}

func (req *ClearMessageFlagsRequest) Pack() []byte {
	var b uint8 = 0
	if req.ClearOEM2 {
		b = types.SetBit7(b)
	}
	if req.ClearOEM1 {
		b = types.SetBit6(b)
	}
	if req.ClearOEM0 {
		b = types.SetBit5(b)
	}
	if req.ClearWatchdogPreTimeoutInterruptFlag {
		b = types.SetBit3(b)
	}
	if req.ClearEventMessageBuffer {
		b = types.SetBit1(b)
	}
	if req.ClearReceiveMessageQueue {
		b = types.SetBit0(b)
	}

	return []byte{b}
}

func (res *ClearMessageFlagsResponse) Unpack(msg []byte) error {
	return nil
}

func (*ClearMessageFlagsResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *ClearMessageFlagsResponse) Format() string {
	// Todo
	return ""
}
