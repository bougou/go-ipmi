package ipmi

type GetChassisCapabilitiesRequest struct {
	// no request data
}

func (req *GetChassisCapabilitiesRequest) Pack() []byte {
	return nil
}

type GetChassisCapabilitiesResponse struct {
	CompletionCode

	// Capabilities Flags
	// [7:4] - reserved
	// [3] - 1b = provides power interlock (IPMI 1.5)
	// [2] - 1b = provides Diagnostic Interrupt (FP NMI) (IPMI 1.5)
	// [1] - 1b = Provides Front Panel Lockout (this indicates that the chassis
	// has capabilities to lock out external power control and reset
	// button or front panel interfaces and/or detect tampering with
	// those interfaces)
	// [0] - 1b = Chassis provides intrusion (physical security) sensor
	ProvidePowerInterlock      bool
	ProvideDiagnosticInterrupt bool
	ProvideFrontPanelLockout   bool
	ProvideIntrusionSensor     bool

	// Chassis FRU Info Device Address.
	// Note: all IPMB addresses used in this command are have the 7-bit I2C slave
	// address as the most-significant 7-bits and the least significant bit set to 0b.
	// 00h = unspecified
	FRUDeviceAddress uint8

	// Chassis SDR Device Address
	SDRDeviceAddress uint8

	// Chassis SEL Device Address
	SELDeviceAddress uint8

	// Chassis System Management Device Address
	SystemManagementDeviceAddress uint8

	// Chassis Bridge Device Address. Reports location of the ICMB bridge
	// function. If this field is not provided, the address is assumed to be the BMC
	// address (20h). Implementing this field is required when the Get Chassis
	// Capabilities command is implemented by a BMC, and whenever the Chassis
	// Bridge function is implemented at an address other than 20h.
	BridgeDeviceAddress uint8
}
