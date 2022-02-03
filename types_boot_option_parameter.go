package ipmi

import (
	"fmt"
	"time"
)

// Table 28-14, Boot Option Parameters
type BootOptionParameter struct {
	SetInProgressState       *BOP_SetInProgressState
	ServicePartitionSelector *BOP_ServicePartitionSelector
	ServicePartitionScan     *BOP_ServicePartitionScan
	BMCBootFlagValidBitClear *BOP_BMCBootFlagValidBitClear
	BootInfoAcknowledge      *BOP_BootInfoAcknowledge
	BootFlags                *BOP_BootFlags
	BootInitiatorInfo        *BOP_BootInitiatorInfo
	BootInitiatorMailbox     *BOP_BootInitiatorMailbox
}

type BootOptionParameterSelector uint8 //  only 7 bits occupied, 0-127

const (
	BOPS_SetInProgressState       BootOptionParameterSelector = 0x00
	BOPS_ServicePartitionSelector BootOptionParameterSelector = 0x01
	BOPS_ServicePartitionScan     BootOptionParameterSelector = 0x02
	BOPS_BMCBootFlagValidBitClear BootOptionParameterSelector = 0x03
	BOPS_BootInfoAcknowledge      BootOptionParameterSelector = 0x04
	BOPS_BootFlags                BootOptionParameterSelector = 0x05
	BOPS_BootInitiatorInfo        BootOptionParameterSelector = 0x06
	BOPS_BootInitiatorMailbox     BootOptionParameterSelector = 0x07

	// OEM Parameters, 96:127
)

func (bop *BootOptionParameter) Format(paramSelecotr BootOptionParameterSelector) string {
	switch paramSelecotr {
	case BOPS_SetInProgressState:
		return fmt.Sprintf(" Set In Progress : %s", bop.SetInProgressState.Format())
	case BOPS_ServicePartitionSelector:
		return fmt.Sprintf(" Service Partition Selector : %s", bop.ServicePartitionSelector.Format())
	case BOPS_ServicePartitionScan:
		return fmt.Sprintf(" Service Partition Scan :\n%s", bop.ServicePartitionScan.Format())
	case BOPS_BMCBootFlagValidBitClear:
		return fmt.Sprintf(" BMC boot flag valid bit clearing :\n%s", bop.BMCBootFlagValidBitClear.Format())
	case BOPS_BootInfoAcknowledge:
		return fmt.Sprintf(" Boot Info Acknowledge :\n%s", bop.BootInfoAcknowledge.Format())
	case BOPS_BootFlags:
		return fmt.Sprintf(" Boot Flags :\n%s", bop.BootFlags.Format())
	case BOPS_BootInitiatorInfo:
		return fmt.Sprintf(" Boot Initiator Info :\n%s", bop.BootInitiatorInfo.Format())
	case BOPS_BootInitiatorMailbox:
		return bop.BootInitiatorMailbox.Format()
	}
	return ""
}

func (bop *BootOptionParameter) Pack(paramSelecotr BootOptionParameterSelector) []byte {
	switch paramSelecotr {
	case BOPS_SetInProgressState:
		return bop.SetInProgressState.Pack()
	case BOPS_ServicePartitionSelector:
		return bop.ServicePartitionSelector.Pack()
	case BOPS_ServicePartitionScan:
		return bop.ServicePartitionScan.Pack()
	case BOPS_BMCBootFlagValidBitClear:
		return bop.BMCBootFlagValidBitClear.Pack()
	case BOPS_BootInfoAcknowledge:
		return bop.BootInfoAcknowledge.Pack()
	case BOPS_BootFlags:
		return bop.BootFlags.Pack()
	case BOPS_BootInitiatorInfo:
		return bop.BootInitiatorInfo.Pack()
	case BOPS_BootInitiatorMailbox:
		return bop.BootInitiatorMailbox.Pack()
	}
	return nil
}

func ParseBootOptionParameterData(paramSelecotr BootOptionParameterSelector, paramData []byte) (*BootOptionParameter, error) {
	bop := &BootOptionParameter{}

	var err error
	switch paramSelecotr {
	case BOPS_SetInProgressState:
		var tmp uint8
		p := (*BOP_SetInProgressState)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.SetInProgressState = p

	case BOPS_ServicePartitionSelector:
		var tmp uint8
		p := (*BOP_ServicePartitionSelector)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.ServicePartitionSelector = p

	case BOPS_ServicePartitionScan:
		var tmp uint8
		p := (*BOP_ServicePartitionScan)(&tmp)
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.ServicePartitionScan = p

	case BOPS_BMCBootFlagValidBitClear:
		p := &BOP_BMCBootFlagValidBitClear{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.BMCBootFlagValidBitClear = p

	case BOPS_BootInfoAcknowledge:
		p := &BOP_BootInfoAcknowledge{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.BootInfoAcknowledge = p

	case BOPS_BootFlags:
		p := &BOP_BootFlags{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.BootFlags = p

	case BOPS_BootInitiatorInfo:
		p := &BOP_BootInitiatorInfo{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.BootInitiatorInfo = p

	case BOPS_BootInitiatorMailbox:
		p := &BOP_BootInitiatorMailbox{}
		err = p.Unpack(paramData)
		if err != nil {
			break
		}
		bop.BootInitiatorMailbox = p

	}

	if err != nil {
		return nil, fmt.Errorf("unpack paramData for paramSelector (%d) failed, err: %s", paramSelecotr, err)
	}
	return bop, nil
}

type BOP_SetInProgressState uint8

func (p *BOP_SetInProgressState) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = BOP_SetInProgressState(paramData[0])
	return nil
}

func (p *BOP_SetInProgressState) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p BOP_SetInProgressState) Format() string {
	switch p {
	case 0:
		return "set complete"
	case 1:
		return "set in progess"
	case 2:
		return "commit write"
	}
	return ""
}

// This value is used to select which service partition BIOS should boot using.
type BOP_ServicePartitionSelector uint8

func (p *BOP_ServicePartitionSelector) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *BOP_ServicePartitionSelector) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = BOP_ServicePartitionSelector(paramData[0])
	return nil
}

func (p BOP_ServicePartitionSelector) Format() string {
	switch p {
	case 0:
		return "unspecfied"
	default:
		return fmt.Sprintf("%#02x", p)
	}
}

type BOP_ServicePartitionScan uint8

func (p *BOP_ServicePartitionScan) Pack() []byte {
	return []byte{uint8(*p)}
}

func (p *BOP_ServicePartitionScan) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	*p = BOP_ServicePartitionScan(paramData[0])
	return nil
}

func (p BOP_ServicePartitionScan) Format() string {
	var s string
	if isBit1Set(uint8(p)) {
		s += "   - Request BIOS to casn\n"
	}
	if isBit0Set(uint8(p)) {
		s += "   - Service Partition Discoverd"
	}
	if s == "" {
		return "     No flag set"
	}
	return s
}

type BOP_BMCBootFlagValidBitClear struct {
	DontClearOnResetPEFOrPowerCyclePEF      bool // corresponding to restart cause: 0x08, 0x09
	DontClearOnCommandReceivedTimeout       bool // corresponding to restart cause: 0x01
	DontClearOnWatchdogTimeout              bool // corresponding to restart cause: 0x04
	DontClearOnResetPushButtonOrSoftReset   bool // corresponding to restart cause: 0x02, 0x0a
	DontClearOnPowerUpPushButtonOrWakeEvent bool // corresponding to restart cause: 0x03, 0x0b
}

func (p *BOP_BMCBootFlagValidBitClear) Format() string {

	var s string
	if p.DontClearOnResetPEFOrPowerCyclePEF {
		s += "   - Don't clear valid bit on reset/power cycle cause by PEF\n"
	}
	if p.DontClearOnCommandReceivedTimeout {
		s += "   - Don't automatically clear boot flag valid bit on timeout\n"
	}
	if p.DontClearOnWatchdogTimeout {
		s += "   - Don't clear valid bit on reset/power cycle cause by watchdog\n"
	}
	if p.DontClearOnResetPushButtonOrSoftReset {
		s += "   - Don't clear valid bit on push button reset // soft reset\n"
	}
	if p.DontClearOnPowerUpPushButtonOrWakeEvent {
		s += "   - Don't clear valid bit on power up via power push button or wake event"
	}

	// When any flag was set, then at least one of the above condition will be true, thus 's' would not be empty.
	if s == "" {
		return "     No flag set"
	}

	return s
}

func (p *BOP_BMCBootFlagValidBitClear) Pack() []byte {
	var b uint8

	if p.DontClearOnResetPEFOrPowerCyclePEF {
		b = setBit4(b)
	}
	if p.DontClearOnCommandReceivedTimeout {
		b = setBit3(b)
	}
	if p.DontClearOnWatchdogTimeout {
		b = setBit2(b)
	}
	if p.DontClearOnResetPushButtonOrSoftReset {
		b = setBit1(b)
	}
	if p.DontClearOnPowerUpPushButtonOrWakeEvent {
		b = setBit0(b)
	}
	return []byte{b}
}

func (p *BOP_BMCBootFlagValidBitClear) Unpack(parameterData []byte) error {
	if len(parameterData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	b := parameterData[0]
	p.DontClearOnResetPEFOrPowerCyclePEF = isBit4Set(b)
	p.DontClearOnCommandReceivedTimeout = isBit3Set(b)
	p.DontClearOnWatchdogTimeout = isBit2Set(b)
	p.DontClearOnResetPushButtonOrSoftReset = isBit1Set(b)
	p.DontClearOnPowerUpPushButtonOrWakeEvent = isBit0Set(b)
	return nil
}

type BOP_BootInfoAcknowledge struct {
	WriteMask uint8

	// The boot initiator should typically write FFh to this parameter prior to initiating the boot.
	// The boot initiator may write 0 s if it wants to intentionally direct a given party to ignore the
	// boot info.
	// This field is automatically initialized to 00h when the management controller if first powered up or reset.
	ByOEM                bool
	BySMS                bool
	ByOSServicePartition bool
	ByOSLoader           bool
	ByBIOSPOST           bool
}

func (p *BOP_BootInfoAcknowledge) Format() string {
	var s string
	if p.ByOEM {
		s += "   - OEM has handled boot info\n"
	}
	if p.BySMS {
		s += "   - SMS has handled boot info\n"
	}
	if p.ByOSServicePartition {
		s += "   - OS // service partition has handled boot info\n"
	}
	if p.ByOSLoader {
		s += "   - OS Loader has handled boot info\n"
	}
	if p.ByBIOSPOST {
		s += "   - BIOS/POST has handled boot info"
	}

	if s == "" {
		return "     No flag set\n"
	}
	return fmt.Sprint(s)
}

func (p *BOP_BootInfoAcknowledge) Pack() []byte {
	var out = make([]byte, 2)
	packUint8(p.WriteMask, out, 0)

	var b uint8
	if p.ByOEM {
		b = setBit4(b)
	}
	if p.BySMS {
		b = setBit4(b)
	}
	if p.ByOSServicePartition {
		b = setBit4(b)
	}
	if p.ByOSLoader {
		b = setBit4(b)
	}
	if p.ByBIOSPOST {
		b = setBit4(b)
	}
	packUint8(b, out, 1)
	return out
}

func (p *BOP_BootInfoAcknowledge) Unpack(parameterData []byte) error {
	if len(parameterData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 bytes")
	}

	p.WriteMask, _, _ = unpackUint8(parameterData, 0)

	b, _, _ := unpackUint8(parameterData, 1)
	p.ByOEM = isBit4Set(b)
	p.BySMS = isBit3Set(b)
	p.ByOSServicePartition = isBit2Set(b)
	p.ByOSLoader = isBit1Set(b)
	p.ByBIOSPOST = isBit0Set(b)
	return nil
}

type BOP_BootFlags struct {
	BootFlagsValid bool
	Persist        bool // or else applied to next boot only
	BIOSBootType   BIOSBootType

	CMOSClear          bool
	LockKeyboard       bool
	BootDeviceSelector BootDeviceSelector // 4 bits
	ScreenBlank        bool
	LockoutResetButton bool

	LockoutPowerOff           bool
	BIOSVerbosity             BIOSVerbosity
	ForceProgressEventTraps   bool
	BypassUserPassword        bool
	LockoutSleepButton        bool
	ConsoleRedirectionControl ConsoleRedirectionControl // only 2 bits

	BIOSSharedModeOverride bool
	BIOSMuxControl         BIOSMuxControl // only 3 bits

	DeviceInstanceSelector uint8 // only 5 bits
}

func (p *BOP_BootFlags) Pack() []byte {
	out := make([]byte, 5)

	var b1 uint8
	if p.BootFlagsValid {
		b1 = setBit7(b1)
	}
	if p.Persist {
		b1 = setBit6(b1)
	}
	if p.BIOSBootType {
		b1 = setBit5(b1)
	}
	packUint8(b1, out, 0)

	var b2 = uint8(p.BootDeviceSelector) << 2
	if p.CMOSClear {
		b2 = setBit7(b2)
	}
	if p.LockKeyboard {
		b2 = setBit6(b2)
	}
	if p.ScreenBlank {
		b2 = setBit1(b2)
	}
	if p.LockoutResetButton {
		b2 = setBit0(b2)
	}
	packUint8(b2, out, 1)

	var b3 = uint8(p.BIOSVerbosity) << 5
	if p.LockoutPowerOff {
		b3 = setBit7(b3)
	}
	if p.ForceProgressEventTraps {
		b3 = setBit4(b3)
	}
	if p.BypassUserPassword {
		b3 = setBit3(b3)
	}
	if p.LockoutResetButton {
		b3 = setBit2(b3)
	}
	b3 |= uint8(p.ConsoleRedirectionControl)
	packUint8(b3, out, 2)

	var b4 uint8
	if p.BIOSSharedModeOverride {
		b4 = setBit3(b4)
	}
	b4 |= uint8(p.BIOSMuxControl)
	packUint8(b4, out, 3)

	var b5 = uint8(p.DeviceInstanceSelector)
	packUint8(b5, out, 4)

	return out
}

func (p *BOP_BootFlags) Unpack(parameterData []byte) error {
	if len(parameterData) != 5 {
		return fmt.Errorf("the parameter data length must be 5 bytes")
	}

	b1, _, _ := unpackUint8(parameterData, 0)
	p.BootFlagsValid = isBit7Set(b1)
	p.Persist = isBit6Set(b1)
	p.BIOSBootType = BIOSBootType(isBit5Set(b1))

	b2, _, _ := unpackUint8(parameterData, 1)
	p.CMOSClear = isBit7Set(b2)
	p.LockKeyboard = isBit6Set(b2)
	p.BootDeviceSelector = BootDeviceSelector((b2 & 0x3f) >> 2) // bit 5,4,3,2
	p.ScreenBlank = isBit1Set(b2)
	p.LockoutResetButton = isBit0Set(b2)

	b3, _, _ := unpackUint8(parameterData, 2)
	p.LockoutPowerOff = isBit7Set(b3)
	p.BIOSVerbosity = BIOSVerbosity((b3 & 0x7f) >> 5)
	p.ForceProgressEventTraps = isBit4Set(b3)
	p.BypassUserPassword = isBit3Set(b3)
	p.LockoutResetButton = isBit2Set(b3)
	p.ConsoleRedirectionControl = ConsoleRedirectionControl(b3 & 0x03)

	b4, _, _ := unpackUint8(parameterData, 3)
	p.BIOSSharedModeOverride = isBit3Set(b4)
	p.BIOSMuxControl = BIOSMuxControl(b4 & 0x07)

	b5, _, _ := unpackUint8(parameterData, 4)
	p.DeviceInstanceSelector = b5 & 0x1f

	return nil
}

func (p *BOP_BootFlags) Format() string {
	var s string
	s += fmt.Sprintf("   - Boot Flag %s\n", formatBool(p.BootFlagsValid, "Valid", "Invalid"))
	s += fmt.Sprintf("   - Options apply to %s\n", formatBool(p.Persist, "all future boots", "only next boot"))
	s += fmt.Sprintf("   - %s\n", p.BIOSBootType.String())

	if p.CMOSClear {
		s += "   - CMOS Clear\n"
	}
	if p.LockKeyboard {
		s += "   - Lock Keyboard\n"
	}
	s += fmt.Sprintf("   - Boot Device Selector : %s\n", p.BootDeviceSelector.String())
	if p.ScreenBlank {
		s += "   - Screen blank\n"
	}
	if p.LockoutResetButton {
		s += "   - Lock out Reset buttons\n"
	}

	if p.LockoutPowerOff {
		s += "   - Lock out (power off/sleep request) via Power Button\n"
	}
	s += fmt.Sprintf("   - BIOS verbosity : %s\n", p.BIOSVerbosity.String())
	if p.ForceProgressEventTraps {
		s += "   - Force progress event traps\n"
	}
	if p.BypassUserPassword {
		s += "   - User password bypass\n"
	}
	if p.LockoutSleepButton {
		s += "   - Lock Out Sleep Button\n"
	}

	s += fmt.Sprintf("   - Console Redirection control : %s\n", p.ConsoleRedirectionControl.String())
	if p.BIOSSharedModeOverride {
		s += "   - BIOS Shared Mode Override\n"
	}
	s += fmt.Sprintf("   - BIOS Mux Control Override : %s", p.BIOSMuxControl.String())
	return s
}

type BIOSVerbosity uint8 // only 2 bits, occupied 0-3

func (v BIOSVerbosity) String() string {
	switch v {
	case 0:
		return "System Default"
	case 1:
		return "Request Quiet Display"
	case 2:
		return "Request Verbose Display"
	default:
		return "Flag error"
	}
}

type BIOSBootType bool

const (
	BIOSBootTypeLegacy BIOSBootType = false // PC compatible boot (legacy)
	BIOSBootTypeEFI    BIOSBootType = true  // Extensible Firmware Interface Boot (EFI)
)

func (t BIOSBootType) String() string {
	if t {
		return "BIOS EFI boot"
	}
	return "BIOS PC Compatible (legacy) boot"
}

type BootDeviceSelector uint8 // only 4 bits occupied

const (
	BootDeviceSelectorNoOverride               BootDeviceSelector = 0x00
	BootDeviceSelectorForcePXE                 BootDeviceSelector = 0x01
	BootDeviceSelectorForceHardDrive           BootDeviceSelector = 0x02
	BootDeviceSelectorForceHardDriveSafe       BootDeviceSelector = 0x03
	BootDeviceSelectorForceDiagnosticPartition BootDeviceSelector = 0x04
	BootDeviceSelectorForceCDROM               BootDeviceSelector = 0x05
	BootDeviceSelectorForceBIOSSetup           BootDeviceSelector = 0x06
	BootDeviceSelectorForceRemoteFloppy        BootDeviceSelector = 0x07
	BootDeviceSelectorForceRemoteCDROM         BootDeviceSelector = 0x08
	BootDeviceSelectorForceRemoteMedia         BootDeviceSelector = 0x09
	BootDeviceSelectorForceRemoteHardDrive     BootDeviceSelector = 0x0b
	BootDeviceSelectorForceFloppy              BootDeviceSelector = 0x0f
)

func (s BootDeviceSelector) String() string {
	switch s {
	case 0x00:
		return "No override"
	case 0x01:
		return "Force PXE"
	case 0x02:
		return "Force Boot from default Hard-Drive"
	case 0x03:
		return "Force Boot from default Hard-Drive, request Safe-Mode"
	case 0x04:
		return "Force Boot from Diagnostic Partition"
	case 0x05:
		return "Force Boot from CD/DVD"
	case 0x06:
		return "Force Boot into BIOS Setup"
	case 0x07:
		return "Force Boot from remotely connected Floppy/primary removable media"
	case 0x08:
		return "Force Boot from remotely connected CD/DVD"
	case 0x09:
		return "Force Boot from primary remote media"
	case 0x0b:
		return "Force Boot from remotely connected Hard-Drive"
	case 0x0f:
		return "Force Boot from Floppy/primary removable media"
	default:
		return "Flag error"
	}
}

type ConsoleRedirectionControl uint8

func (c ConsoleRedirectionControl) String() string {
	switch c {
	case 0:
		return "Console redirection occurs per BIOS configuration setting (default)"
	case 1:
		return "Suppress (skip) console redirection if enabled"
	case 2:
		return "Request console redirection be enabled"
	default:
		return "Flag error"
	}
}

type BIOSMuxControl uint8

func (b BIOSMuxControl) String() string {
	switch b {
	case 0:
		return "BIOS uses recommended setting of the mux at the end of POST"
	case 1:
		return "Requests BIOS to force mux to BMC at conclusion of POST/start of OS boot"
	case 2:
		return "Requests BIOS to force mux to system at conclusion of POST/start of OS boot"
	default:
		return "Flag error"
	}
}

type BOP_BootInitiatorInfo struct {
	ChannelNumber     uint8
	SessionID         uint32
	BootInfoTimestamp time.Time
}

func (p *BOP_BootInitiatorInfo) Format() string {
	return fmt.Sprintf(`     Channel Number : %d
     Session Id     : %d
     Timestamp      : %s`, p.ChannelNumber, p.SessionID, p.BootInfoTimestamp)
}

func (p *BOP_BootInitiatorInfo) Pack() []byte {
	out := make([]byte, 9)
	packUint8(p.ChannelNumber, out, 0)
	packUint32L(p.SessionID, out, 1)

	ts := uint32(p.BootInfoTimestamp.Unix())
	packUint32L(ts, out, 5)
	return out
}

func (p *BOP_BootInitiatorInfo) Unpack(parameterData []byte) error {
	if len(parameterData) != 9 {
		return fmt.Errorf("the parameter data length must be 9 bytes")
	}

	p.ChannelNumber, _, _ = unpackUint8(parameterData, 0)
	p.SessionID, _, _ = unpackUint32L(parameterData, 1)

	ts, _, _ := unpackUint32L(parameterData, 5)
	p.BootInfoTimestamp = parseTimestamp(ts)
	return nil
}

type BOP_BootInitiatorMailbox struct {
	SetSelector uint8
	BlockData   []byte
}

func (p *BOP_BootInitiatorMailbox) Format() string {
	// Todo
	return ""
}

func (p *BOP_BootInitiatorMailbox) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData, out, 1)
	return out
}

func (p *BOP_BootInitiatorMailbox) Unpack(parameterData []byte) error {
	if len(parameterData) < 1 {
		return fmt.Errorf("the parameter data length must be at least 1 bytes")
	}

	p.SetSelector, _, _ = unpackUint8(parameterData, 0)
	p.BlockData, _, _ = unpackBytes(parameterData, 1, len(parameterData)-1)
	return nil
}
