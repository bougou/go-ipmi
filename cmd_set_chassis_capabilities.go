package ipmi

import "context"

// 28.7 Set Chassis Capabilities Command
type SetChassisCapabilitiesRequest struct {
	ProvideFrontPanelLockout bool
	ProvideIntrusionSensor   bool

	FRUDeviceAddress uint8

	SDRDeviceAddress uint8

	SELDeviceAddress uint8

	SystemManagementDeviceAddress uint8

	BridgeDeviceAddress uint8
}

type SetChassisCapabilitiesResponse struct {
}

func (req *SetChassisCapabilitiesRequest) Pack() []byte {
	out := make([]byte, 5)

	var b uint8 = 0
	if req.ProvideFrontPanelLockout {
		b = setBit1(b)
	}
	if req.ProvideIntrusionSensor {
		b = setBit0(b)
	}
	packUint8(b, out, 0)
	packUint8(req.FRUDeviceAddress, out, 1)
	packUint8(req.SDRDeviceAddress, out, 2)
	packUint8(req.SELDeviceAddress, out, 3)
	packUint8(req.SystemManagementDeviceAddress, out, 4)
	return out
}

func (req *SetChassisCapabilitiesRequest) Command() Command {
	return CommandSetChassisCapabilities
}

func (res *SetChassisCapabilitiesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetChassisCapabilitiesResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetChassisCapabilitiesResponse) Format() string {
	return ""
}

func (c *Client) SetChassisCapabilities(ctx context.Context, request *SetChassisCapabilitiesRequest) (response *SetChassisCapabilitiesResponse, err error) {
	response = &SetChassisCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
