package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 22.2 Get BMC Global Enables Command
)

type GetBMCGlobalEnablesRequest struct {
	// empty
}

type GetBMCGlobalEnablesResponse struct {
	OEM2Enabled bool
	OEM1Enabled bool
	OEM0Enabled bool

	SystemEventLoggingEnabled              bool
	EventMessageBufferEnabled              bool
	EventMessageBufferFullInterruptEnabled bool
	ReceiveMessageQueueInterruptEnabled    bool
}

func (req *GetBMCGlobalEnablesRequest) Command() ipmi.Command {
	return ipmi.CommandGetBMCGlobalEnables
}

func (req *GetBMCGlobalEnablesRequest) Pack() []byte {
	return []byte{}
}

func (res *GetBMCGlobalEnablesResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	b, _, _ := ipmi.UnpackUint8(msg, 0)
	res.OEM2Enabled = ipmi.IsBit7Set(b)
	res.OEM1Enabled = ipmi.IsBit6Set(b)
	res.OEM0Enabled = ipmi.IsBit5Set(b)
	res.SystemEventLoggingEnabled = ipmi.IsBit3Set(b)
	res.EventMessageBufferEnabled = ipmi.IsBit2Set(b)
	res.EventMessageBufferFullInterruptEnabled = ipmi.IsBit1Set(b)
	res.ReceiveMessageQueueInterruptEnabled = ipmi.IsBit0Set(b)
	return nil
}

func (*GetBMCGlobalEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetBMCGlobalEnablesResponse) Format() string {
	// Todo
	return ""
}
