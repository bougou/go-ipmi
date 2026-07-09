package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.1 Set BMC Global Enables Command
type SetBMCGlobalEnablesRequest struct {
	// Generic system mgmt. software must do a "read-modify-write" using the Get BMC Global Enables and Set BMC Global Enables to avoid altering EnableOEM_X field.
	EnableOEM2 bool
	EnableOEM1 bool
	EnableOEM0 bool

	EnableSystemEventLogging              bool
	EnableEventMessageBuffer              bool
	EnableEventMessageBufferFullInterrupt bool
	EnableReceiveMessageQueueInterrupt    bool
}

type SetBMCGlobalEnablesResponse struct {
	// empty
}

func (req *SetBMCGlobalEnablesRequest) Command() types.Command {
	return types.CommandSetBMCGlobalEnables
}

func (req *SetBMCGlobalEnablesRequest) Pack() []byte {
	var b uint8 = 0

	if req.EnableOEM2 {
		b = types.SetBit7(b)
	}
	if req.EnableOEM1 {
		b = types.SetBit6(b)
	}
	if req.EnableOEM0 {
		b = types.SetBit5(b)
	}
	if req.EnableSystemEventLogging {
		b = types.SetBit3(b)
	}
	if req.EnableEventMessageBuffer {
		b = types.SetBit2(b)
	}
	if req.EnableEventMessageBufferFullInterrupt {
		b = types.SetBit1(b)
	}
	if req.EnableReceiveMessageQueueInterrupt {
		b = types.SetBit0(b)
	}

	return []byte{b}
}

func (res *SetBMCGlobalEnablesResponse) Unpack(msg []byte) error {
	return nil
}

func (*SetBMCGlobalEnablesResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *SetBMCGlobalEnablesResponse) Format() string {
	// Todo
	return ""
}
