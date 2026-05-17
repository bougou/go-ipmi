package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *ClearMessageFlagsRequest) Command() ipmi.Command {
	return ipmi.CommandClearMessageFlags
}

func (req *ClearMessageFlagsRequest) Pack() []byte {
	var b uint8 = 0
	if req.ClearOEM2 {
		b = ipmi.SetBit7(b)
	}
	if req.ClearOEM1 {
		b = ipmi.SetBit6(b)
	}
	if req.ClearOEM0 {
		b = ipmi.SetBit5(b)
	}
	if req.ClearWatchdogPreTimeoutInterruptFlag {
		b = ipmi.SetBit3(b)
	}
	if req.ClearEventMessageBuffer {
		b = ipmi.SetBit1(b)
	}
	if req.ClearReceiveMessageQueue {
		b = ipmi.SetBit0(b)
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
