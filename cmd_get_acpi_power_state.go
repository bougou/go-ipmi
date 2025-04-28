package ipmi

import (
	"context"
	"fmt"
)

// 20.7 Get ACPI Power State Command
type GetACPIPowerStateRequest struct {
	// empty
}

type GetACPIPowerStateResponse struct {
	SystemPowerState SystemPowerState
	DevicePowerState DevicePowerState
}

func (req *GetACPIPowerStateRequest) Pack() []byte {
	return nil
}

func (req *GetACPIPowerStateRequest) Command() Command {
	return CommandGetACPIPowerState
}

func (res *GetACPIPowerStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetACPIPowerStateResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	b1, _, _ := unpackUint8(msg, 0)
	res.SystemPowerState = SystemPowerState(b1 & 0x7f)
	b2, _, _ := unpackUint8(msg, 1)
	res.DevicePowerState = DevicePowerState(b2 & 0x7f)
	return nil
}

func (res *GetACPIPowerStateResponse) Format() string {
	return "" +
		fmt.Sprintf("ACPI System Power State: %s\n", res.SystemPowerState) +
		fmt.Sprintf("ACPI Device Power State: %s\n", res.DevicePowerState)
}

// This command is provided to allow system software to tell a controller the present ACPI power state of the system.
func (c *Client) GetACPIPowerState(ctx context.Context) (response *GetACPIPowerStateResponse, err error) {
	request := &GetACPIPowerStateRequest{}
	response = &GetACPIPowerStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
