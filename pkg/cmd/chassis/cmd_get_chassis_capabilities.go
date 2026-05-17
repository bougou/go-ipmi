package chassis

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 28.1 Get Chassis Capabilities Command
)

type GetChassisCapabilitiesRequest struct {
	// empty
}

type GetChassisCapabilitiesResponse struct {
	ProvidePowerInterlock      bool
	ProvideDiagnosticInterrupt bool
	ProvideFrontPanelLockout   bool
	ProvideIntrusionSensor     bool

	// Chassis FRU Device
	FRUDeviceAddress uint8

	SDRDeviceAddress uint8

	SELDeviceAddress uint8

	SystemManagementDeviceAddress uint8

	//  If this field is not provided, the address is assumed to be the BMC address (20h).
	BridgeDeviceAddress uint8
}

func (req *GetChassisCapabilitiesRequest) Pack() []byte {
	return []byte{}
}

func (req *GetChassisCapabilitiesRequest) Command() ipmi.Command {
	return ipmi.CommandGetChassisCapabilities
}

func (res *GetChassisCapabilitiesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChassisCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	b1, _, _ := ipmi.UnpackUint8(msg, 0)
	res.ProvidePowerInterlock = ipmi.IsBit3Set(b1)
	res.ProvideDiagnosticInterrupt = ipmi.IsBit2Set(b1)
	res.ProvideFrontPanelLockout = ipmi.IsBit1Set(b1)
	res.ProvideIntrusionSensor = ipmi.IsBit0Set(b1)

	res.FRUDeviceAddress, _, _ = ipmi.UnpackUint8(msg, 1)
	res.SDRDeviceAddress, _, _ = ipmi.UnpackUint8(msg, 2)
	res.SELDeviceAddress, _, _ = ipmi.UnpackUint8(msg, 3)
	res.SystemManagementDeviceAddress, _, _ = ipmi.UnpackUint8(msg, 4)

	if len(msg) == 6 {
		res.BridgeDeviceAddress, _, _ = ipmi.UnpackUint8(msg, 5)
	}
	return nil
}

func (res *GetChassisCapabilitiesResponse) Format() string {
	return "todo"
}
