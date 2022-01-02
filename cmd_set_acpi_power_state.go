package ipmi

// 20.6 Set ACPI Power State Command
type SetACPIPowerStateRequest struct {
	SetSystemPowerState bool // false means don't change system power state
	SystemPowerState    SystemPowerState
	SetDevicePowerState bool // false means don't change device power state
	DevicePowerState    DevicePowerState
}

type SetACPIPowerStateResponse struct {
	// empty
}

type SystemPowerState uint8

const (
	SystemPowerStateS0G0       uint8 = 0x00
	SystemPowerStateS1         uint8 = 0x01
	SystemPowerStateS2         uint8 = 0x02
	SystemPowerStateS3         uint8 = 0x03
	SystemPowerStateS4         uint8 = 0x04
	SystemPowerStateS5G2       uint8 = 0x05
	SystemPowerStateS4S5       uint8 = 0x06
	SystemPowerStateG3         uint8 = 0x07
	SystemPowerStateSleeping   uint8 = 0x08
	SystemPowerStateG1Sleeping uint8 = 0x09
	SystemPowerStateOverride   uint8 = 0x0a
	SystemPowerStateLegacyOn   uint8 = 0x20
	SystemPowerStateLegacyOff  uint8 = 0x21
	SystemPowerStateUnknown    uint8 = 0x2a
	SystemPowerStateNoChange   uint8 = 0x7f
)

func (s SystemPowerState) String() string {
	m := map[SystemPowerState]string{
		0x00: "S0G0, working",
		0x01: "S1, hardware context maintained, typically equates to processor/chip set clocks stopped",
		0x02: "S2, typically equates to stopped clocks with processor/cache context lost",
		0x03: "s3，typically equates to suspend-to-RAM",
		0x04: "S4, typically equates to suspend-to-disk",
		0x05: "S5/G2, soft off",
		0x06: "S4/S5, sent when message source canno differentiate between S4 and S5",
		0x07: "G2, mechanical off",
		0x08: "sleeping, sleeping - cannot differentiate between S1-S3",
		0x09: "G1 sleeping, sleeping - cannot differentiate between S1-S4",
		0x0a: "override, S5 entered by override",
		0x20: "Legacy On, Legacy On (indicates On for system that don't support ACPI or have ACPI capabilities disabled)",
		0x21: "Legacy Soft-Off",
		0x2a: "Unknown, system power state unknown",
		0x7f: "No Chagne",
	}
	o, ok := m[s]
	if ok {
		return o
	}
	return ""
}

type DevicePowerState uint8

const (
	DevicePowerStateD0       uint8 = 0x00
	DevicePowerStateD1       uint8 = 0x01
	DevicePowerStateD2       uint8 = 0x02
	DevicePowerStateD3       uint8 = 0x03
	DevicePowerStateUnknown  uint8 = 0x2a
	DevicePowerStateNoChange uint8 = 0x7f
)

func (s DevicePowerState) String() string {
	m := map[DevicePowerState]string{
		0x00: "D0",
		0x01: "D1",
		0x02: "D2",
		0x03: "D2",
		0x2a: "Unknown",
		0x7f: "No Change",
	}
	o, ok := m[s]
	if ok {
		return o
	}
	return ""
}

func (req *SetACPIPowerStateRequest) Pack() []byte {
	out := make([]byte, 2)

	var b1 = uint8(req.SystemPowerState)
	if req.SetSystemPowerState {
		b1 |= 0x80
	}
	packUint8(b1, out, 0)

	var b2 = uint8(req.DevicePowerState)
	if req.SetDevicePowerState {
		b2 |= 0x80
	}
	packUint8(b2, out, 1)

	return out
}

func (req *SetACPIPowerStateRequest) Command() Command {
	return CommandSetACPIPowerState
}

func (res *SetACPIPowerStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetACPIPowerStateResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetACPIPowerStateResponse) Format() string {
	return ""
}

// This command is provided to allow system software to tell a controller the present ACPI power state of the system.
func (c *Client) SetACPIPowerState(request *SetACPIPowerStateRequest) (err error) {
	response := &SetACPIPowerStateResponse{}
	err = c.Exchange(request, response)
	return
}
