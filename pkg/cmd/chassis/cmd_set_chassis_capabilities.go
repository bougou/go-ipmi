package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 28.7 Set Chassis Capabilities Command
)

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
		b = types.SetBit1(b)
	}
	if req.ProvideIntrusionSensor {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 0)
	types.PackUint8(req.FRUDeviceAddress, out, 1)
	types.PackUint8(req.SDRDeviceAddress, out, 2)
	types.PackUint8(req.SELDeviceAddress, out, 3)
	types.PackUint8(req.SystemManagementDeviceAddress, out, 4)
	return out
}

func (req *SetChassisCapabilitiesRequest) Command() types.Command {
	return types.CommandSetChassisCapabilities
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
