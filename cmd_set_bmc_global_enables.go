package ipmi

import (
	"context"
	"fmt"
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

func (req *SetBMCGlobalEnablesRequest) Command() Command {
	return CommandSetBMCGlobalEnables
}

func (req *SetBMCGlobalEnablesRequest) Pack() []byte {
	var b uint8 = 0

	if req.EnableOEM2 {
		b = setBit7(b)
	}
	if req.EnableOEM1 {
		b = setBit6(b)
	}
	if req.EnableOEM0 {
		b = setBit5(b)
	}
	if req.EnableSystemEventLogging {
		b = setBit3(b)
	}
	if req.EnableEventMessageBuffer {
		b = setBit2(b)
	}
	if req.EnableEventMessageBufferFullInterrupt {
		b = setBit1(b)
	}
	if req.EnableReceiveMessageQueueInterrupt {
		b = setBit0(b)
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

func (c *Client) SetBMCGlobalEnables(ctx context.Context, enableSystemEventLogging bool, enableEventMessageBuffer bool, enableEventMessageBufferFullInterrupt bool, enableReceiveMessageQueueInterrupt bool) (response *SetBMCGlobalEnablesResponse, err error) {
	getRes, err := c.GetBMCGlobalEnables(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetBMCGlobalEnables failed, err: %s", err)
	}

	request := &SetBMCGlobalEnablesRequest{
		EnableOEM2: getRes.OEM2Enabled,
		EnableOEM1: getRes.OEM1Enabled,
		EnableOEM0: getRes.OEM0Enabled,

		EnableSystemEventLogging:              enableSystemEventLogging,
		EnableEventMessageBuffer:              enableEventMessageBuffer,
		EnableEventMessageBufferFullInterrupt: enableEventMessageBufferFullInterrupt,
		EnableReceiveMessageQueueInterrupt:    enableReceiveMessageQueueInterrupt,
	}
	response = &SetBMCGlobalEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
