package chassis

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 28.2 Get Chassis Status Command
type GetChassisStatusRequest struct {
	// empty
}

type GetChassisStatusResponse struct {
	// Current Power State
	PowerRestorePolicy PowerRestorePolicy
	PowerControlFault  bool // Controller attempted to turn system power on or off, but system did not enter desired state.
	PowerFault         bool // fault detected in main power subsystem
	InterLock          bool // chassis is presently shut down because a chassis	panel interlock switch is active
	PowerOverload      bool // system shutdown because of power overload condition.
	PowerIsOn          bool // 系统电源：上电

	// Last Power Event
	LastPowerOnByCommand                   bool
	LastPowerDownByPowerFault              bool
	LastPowerDownByPowerInterlockActivated bool
	LastPowerDownByPowerOverload           bool
	ACFailed                               bool

	// Last Power Event

	// Misc. Chassis State
	ChassisIdentifySupported bool
	ChassisIdentifyState     ChassisIdentifyState
	CollingFanFault          bool
	DriveFault               bool
	FrontPanelLockoutActive  bool // (power off and reset via chassis push-buttons disabled. 前面板锁定)
	ChassisIntrusionActive   bool // 机箱入侵:（机箱盖被打开）

	// Front Panel Button Capabilities and disable/enable status (Optional)
	SleepButtonDisableAllowed      bool
	DiagnosticButtonDisableAllowed bool
	ResetButtonDisableAllowed      bool
	PoweroffButtonDisableAllowed   bool
	SleepButtonDisabled            bool
	DiagnosticButtonDisabled       bool
	ResetButtonDisabled            bool
	PoweroffButtonDisabled         bool
}

type ChassisIdentifyState uint8

const (
	ChassisIdentifyStateOff          ChassisIdentifyState = 0
	ChassisIdentifyStateTemporaryOn  ChassisIdentifyState = 1
	ChassisIdentifyStateIndefiniteOn ChassisIdentifyState = 2
)

func (c ChassisIdentifyState) String() string {
	m := map[ChassisIdentifyState]string{
		0: "Off",
		1: "Temporary (timed) On",
		2: "Indefinite On",
	}
	s, ok := m[c]
	if ok {
		return s
	}
	return "reserved"
}

// PowerRestorePolicy
// 通电开机策略
type PowerRestorePolicy uint8

const (
	PowerRestorePolicyAlwaysOff PowerRestorePolicy = 0 // 保持下电（关机）
	PowerRestorePolicyPrevious  PowerRestorePolicy = 1 // 与之前保持一致（恢复断电前状态）
	PowerRestorePolicyAlwaysOn  PowerRestorePolicy = 2 // 保持上电（开机）
)

var SupportedPowerRestorePolicies = []string{
	"always-off", "always-on", "previous",
}

func (p PowerRestorePolicy) String() string {
	m := map[PowerRestorePolicy]string{
		0: "always-off", // chassis stays powered off after AC/mains returns
		1: "previous",   // after AC returns, power is restored to the state that was in effect when AC/mains was lost
		2: "always-on",  // chassis always powers up after AC/mains returns
	}
	s, ok := m[p]
	if ok {
		return s
	}
	return "unknown"
}

func (req *GetChassisStatusRequest) Pack() []byte {
	return []byte{}
}

// Pack serialises the response per the bit layout in §28.2 Table 28-3, the
// inverse of [Unpack]. Byte 3 (front-panel button disables) is always emitted
// per spec: "Return as 00h if the panel button disable function is not supported."
func (res *GetChassisStatusResponse) Pack() []byte {
	var b0 uint8
	b0 |= (uint8(res.PowerRestorePolicy) & 0x07) << 5
	if res.PowerControlFault {
		b0 = types.SetBit4(b0)
	}
	if res.PowerFault {
		b0 = types.SetBit3(b0)
	}
	if res.InterLock {
		b0 = types.SetBit2(b0)
	}
	if res.PowerOverload {
		b0 = types.SetBit1(b0)
	}
	if res.PowerIsOn {
		b0 = types.SetBit0(b0)
	}

	var b1 uint8
	if res.LastPowerOnByCommand {
		b1 = types.SetBit4(b1)
	}
	if res.LastPowerDownByPowerFault {
		b1 = types.SetBit3(b1)
	}
	if res.LastPowerDownByPowerInterlockActivated {
		b1 = types.SetBit2(b1)
	}
	if res.LastPowerDownByPowerOverload {
		b1 = types.SetBit1(b1)
	}
	if res.ACFailed {
		b1 = types.SetBit0(b1)
	}

	var b2 uint8
	if res.ChassisIdentifySupported {
		b2 = types.SetBit6(b2)
	}
	b2 |= (uint8(res.ChassisIdentifyState) & 0x03) << 4
	if res.CollingFanFault {
		b2 = types.SetBit3(b2)
	}
	if res.DriveFault {
		b2 = types.SetBit2(b2)
	}
	if res.FrontPanelLockoutActive {
		b2 = types.SetBit1(b2)
	}
	if res.ChassisIntrusionActive {
		b2 = types.SetBit0(b2)
	}

	var b3 uint8
	if res.SleepButtonDisableAllowed {
		b3 = types.SetBit7(b3)
	}
	if res.DiagnosticButtonDisableAllowed {
		b3 = types.SetBit6(b3)
	}
	if res.ResetButtonDisableAllowed {
		b3 = types.SetBit5(b3)
	}
	if res.PoweroffButtonDisableAllowed {
		b3 = types.SetBit4(b3)
	}
	if res.SleepButtonDisabled {
		b3 = types.SetBit3(b3)
	}
	if res.DiagnosticButtonDisabled {
		b3 = types.SetBit2(b3)
	}
	if res.ResetButtonDisabled {
		b3 = types.SetBit1(b3)
	}
	if res.PoweroffButtonDisabled {
		b3 = types.SetBit0(b3)
	}

	return []byte{b0, b1, b2, b3}
}

func (req *GetChassisStatusRequest) Command() types.Command {
	return types.CommandGetChassisStatus
}

func (res *GetChassisStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChassisStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	b1, _, _ := types.UnpackUint8(msg, 0)
	// Power restore policy occupies bits [7:5] per §28.2 Table 28-3.
	b := (b1 & 0xE0) >> 5
	res.PowerRestorePolicy = PowerRestorePolicy(b)
	res.PowerControlFault = types.IsBit4Set(b1)
	res.PowerFault = types.IsBit3Set(b1)
	res.InterLock = types.IsBit2Set(b1)
	res.PowerOverload = types.IsBit1Set(b1)
	res.PowerIsOn = types.IsBit0Set(b1)

	b2, _, _ := types.UnpackUint8(msg, 1)
	res.LastPowerOnByCommand = types.IsBit4Set(b2)
	res.LastPowerDownByPowerFault = types.IsBit3Set(b2)
	res.LastPowerDownByPowerInterlockActivated = types.IsBit2Set(b2)
	res.LastPowerDownByPowerOverload = types.IsBit1Set(b2)
	res.ACFailed = types.IsBit0Set(b2)

	b3, _, _ := types.UnpackUint8(msg, 2)
	res.ChassisIdentifySupported = types.IsBit6Set(b3)
	res.ChassisIdentifyState = ChassisIdentifyState((b3 & 0x30) >> 4)
	res.CollingFanFault = types.IsBit3Set(b3)
	res.DriveFault = types.IsBit2Set(b3)
	res.FrontPanelLockoutActive = types.IsBit1Set(b3)
	res.ChassisIntrusionActive = types.IsBit0Set(b3)

	if len(msg) == 4 {
		b4, _, _ := types.UnpackUint8(msg, 3)
		res.SleepButtonDisableAllowed = types.IsBit7Set(b4)
		res.DiagnosticButtonDisableAllowed = types.IsBit6Set(b4)
		res.ResetButtonDisableAllowed = types.IsBit5Set(b4)
		res.PoweroffButtonDisableAllowed = types.IsBit4Set(b4)
		res.SleepButtonDisabled = types.IsBit3Set(b4)
		res.DiagnosticButtonDisabled = types.IsBit2Set(b4)
		res.ResetButtonDisabled = types.IsBit1Set(b4)
		res.PoweroffButtonDisabled = types.IsBit0Set(b4)
	}
	return nil
}

func (res *GetChassisStatusResponse) Format() string {
	return "" +
		fmt.Sprintf("System Power         : %s\n", types.FormatBool(res.PowerIsOn, "on", "off")) +
		fmt.Sprintf("Power Overload       : %v\n", res.PowerOverload) +
		fmt.Sprintf("Power Interlock      : %s\n", types.FormatBool(res.InterLock, "active", "inactive")) +
		fmt.Sprintf("Main Power Fault     : %v\n", res.PowerFault) +
		fmt.Sprintf("Power Control Fault  : %v\n", res.PowerControlFault) +
		fmt.Sprintf("Power Restore Policy : %s\n", res.PowerRestorePolicy.String()) +
		fmt.Sprintf("Last Power Event     : %s\n", types.FormatBool(res.ChassisIntrusionActive, "active", "inactive")) +
		fmt.Sprintf("Chassis Intrusion    : %s\n", types.FormatBool(res.ChassisIntrusionActive, "active", "inactive")) +
		fmt.Sprintf("Front-Panel Lockout  : %s\n", types.FormatBool(res.FrontPanelLockoutActive, "active", "inactive")) +
		fmt.Sprintf("Drive Fault          : %v\n", res.DriveFault) +
		fmt.Sprintf("Cooling/Fan Fault    : %v\n", res.CollingFanFault) +
		fmt.Sprintf("Sleep Button Disable : %s\n", types.FormatBool(res.SleepButtonDisableAllowed, "allowed", "disallowed")) +
		fmt.Sprintf("Diag Button Disable  : %s\n", types.FormatBool(res.DiagnosticButtonDisableAllowed, "allowed", "disallowed")) +
		fmt.Sprintf("Reset Button Disable : %s\n", types.FormatBool(res.ResetButtonDisableAllowed, "allowed", "disallowed")) +
		fmt.Sprintf("Power Button Disable : %s\n", types.FormatBool(res.PoweroffButtonDisableAllowed, "allowed", "disallowed")) +
		fmt.Sprintf("Sleep Button Disabled: %v\n", res.SleepButtonDisabled) +
		fmt.Sprintf("Diag Button Disabled : %v\n", res.DiagnosticButtonDisabled) +
		fmt.Sprintf("Reset Button Disabled: %v\n", res.ResetButtonDisabled) +
		fmt.Sprintf("Power Button Disabled: %v\n", res.PoweroffButtonDisabled)
}
