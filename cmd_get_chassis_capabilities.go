package ipmi

// 28.1 Get Chassis Capabilities Command
type GetChassisCapabilitiesRequest struct {
	// no request data
}

type GetChassisCapabilitiesResponse struct {
	ProvidePowerInterlock      bool
	ProvideDiagnosticInterrupt bool
	ProvideFrontPanelLockout   bool
	ProvideIntrusionSensor     bool

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

func (req *GetChassisCapabilitiesRequest) Command() Command {
	return CommandGetChassisCapabilities
}

func (res *GetChassisCapabilitiesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChassisCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShort
	}

	b1, _, _ := unpackUint8(msg, 0)
	res.ProvidePowerInterlock = isBit3Set(b1)
	res.ProvideDiagnosticInterrupt = isBit2Set(b1)
	res.ProvideFrontPanelLockout = isBit1Set(b1)
	res.ProvideIntrusionSensor = isBit0Set(b1)

	res.FRUDeviceAddress, _, _ = unpackUint8(msg, 1)
	res.SDRDeviceAddress, _, _ = unpackUint8(msg, 2)
	res.SELDeviceAddress, _, _ = unpackUint8(msg, 3)
	res.SystemManagementDeviceAddress, _, _ = unpackUint8(msg, 4)

	if len(msg) == 6 {
		res.BridgeDeviceAddress, _, _ = unpackUint8(msg, 5)
	}
	return nil
}

func (res *GetChassisCapabilitiesResponse) Format() string {
	return ""
}

func (c *Client) GetChassisCapabilities() (response *GetChassisCapabilitiesResponse, err error) {
	request := &GetChassisCapabilitiesRequest{}
	response = &GetChassisCapabilitiesResponse{}
	err = c.Exchange(request, response)
	return
}
