package ipmi

import (
	"context"
	"fmt"
)

// 28.12 Set System Boot Options Command
type SetSystemBootOptionsRequest struct {
	// Parameter valid
	//  - 1b = mark parameter invalid / locked
	//  - 0b = mark parameter valid / unlocked
	MarkParameterInvalid bool
	// [6:0] - boot option parameter selector
	ParamSelector BootOptionParamSelector

	ParamData []byte
}

// Table 28-14, Boot Option Parameters

type SetSystemBootOptionsResponse struct {
}

func (req *SetSystemBootOptionsRequest) Pack() []byte {

	out := make([]byte, 1+len(req.ParamData))

	b := uint8(req.ParamSelector)
	if req.MarkParameterInvalid {
		b = setBit7(b)
	} else {
		b = clearBit7(b)
	}
	packUint8(b, out, 0)

	packBytes(req.ParamData, out, 1)

	return out
}

func (req *SetSystemBootOptionsRequest) Command() Command {
	return CommandSetSystemBootOptions
}

func (res *SetSystemBootOptionsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state. (This completion code provides a way to recognize that another party has already 'claimed' the parameters)",
		0x82: "attempt to write read-only parameter",
	}
}

func (res *SetSystemBootOptionsResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemBootOptionsResponse) Format() string {
	return ""
}

// This command is used to set parameters that direct the system boot following a system power up or reset.
// The boot flags only apply for one system restart. It is the responsibility of the system BIOS
// to read these settings from the BMC and then clear the boot flags
func (c *Client) SetSystemBootOptions(ctx context.Context, request *SetSystemBootOptionsRequest) (response *SetSystemBootOptionsResponse, err error) {
	response = &SetSystemBootOptionsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSystemBootOptionsFor(ctx context.Context, param BootOptionParameter) error {
	paramSelector, _, _ := param.BootOptionParameter()
	paramData := param.Pack()

	request := &SetSystemBootOptionsRequest{
		MarkParameterInvalid: false,
		ParamSelector:        paramSelector,
		ParamData:            paramData,
	}

	if _, err := c.SetSystemBootOptions(ctx, request); err != nil {
		return fmt.Errorf("SetSystemBootOptions failed, err: %s", err)
	}

	return nil
}
