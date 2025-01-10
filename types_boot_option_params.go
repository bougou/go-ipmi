package ipmi

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type BootOptionParamSelector uint8 //  only 7 bits occupied, 0-127

const (
	BootOptionParamSelector_SetInProgress            BootOptionParamSelector = 0x00
	BootOptionParamSelector_ServicePartitionSelector BootOptionParamSelector = 0x01
	BootOptionParamSelector_ServicePartitionScan     BootOptionParamSelector = 0x02
	BootOptionParamSelector_BMCBootFlagValidBitClear BootOptionParamSelector = 0x03
	BootOptionParamSelector_BootInfoAcknowledge      BootOptionParamSelector = 0x04
	BootOptionParamSelector_BootFlags                BootOptionParamSelector = 0x05
	BootOptionParamSelector_BootInitiatorInfo        BootOptionParamSelector = 0x06
	BootOptionParamSelector_BootInitiatorMailbox     BootOptionParamSelector = 0x07

	// OEM Parameters, 96:127
)

func (bop BootOptionParamSelector) String() string {
	m := map[BootOptionParamSelector]string{
		BootOptionParamSelector_SetInProgress:            "Set In Progress",
		BootOptionParamSelector_ServicePartitionSelector: "Service Partition Selector",
		BootOptionParamSelector_ServicePartitionScan:     "Service Partition Scan",
		BootOptionParamSelector_BMCBootFlagValidBitClear: "BMC Boot Flag Valid bit Clearing",
		BootOptionParamSelector_BootInfoAcknowledge:      "Boot Info Acknowledge",
		BootOptionParamSelector_BootFlags:                "Boot Flags",
		BootOptionParamSelector_BootInitiatorInfo:        "Boot Initiator Info",
		BootOptionParamSelector_BootInitiatorMailbox:     "Boot Initiator Mailbox",
	}

	s, ok := m[bop]
	if ok {
		return s
	}

	return "Unknown"
}

type BootOptionParameter interface {
	BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8)
	Parameter
}

var (
	_ BootOptionParameter = (*BootOptionParam_SetInProgress)(nil)
	_ BootOptionParameter = (*BootOptionParam_ServicePartitionSelector)(nil)
	_ BootOptionParameter = (*BootOptionParam_ServicePartitionScan)(nil)
	_ BootOptionParameter = (*BootOptionParam_BMCBootFlagValidBitClear)(nil)
	_ BootOptionParameter = (*BootOptionParam_BootInfoAcknowledge)(nil)
	_ BootOptionParameter = (*BootOptionParam_BootFlags)(nil)
	_ BootOptionParameter = (*BootOptionParam_BootInitiatorInfo)(nil)
	_ BootOptionParameter = (*BootOptionParam_BootInitiatorMailbox)(nil)
)

// Table 28-14, Boot Option Parameters
type BootOptions struct {
	SetInProgress            *BootOptionParam_SetInProgress
	ServicePartitionSelector *BootOptionParam_ServicePartitionSelector
	ServicePartitionScan     *BootOptionParam_ServicePartitionScan
	BMCBootFlagValidBitClear *BootOptionParam_BMCBootFlagValidBitClear
	BootInfoAcknowledge      *BootOptionParam_BootInfoAcknowledge
	BootFlags                *BootOptionParam_BootFlags
	BootInitiatorInfo        *BootOptionParam_BootInitiatorInfo
	BootInitiatorMailbox     *BootOptionParam_BootInitiatorMailbox
}

func (bootOptions *BootOptions) Format() string {
	format := func(param BootOptionParameter) string {
		paramSelector, _, _ := param.BootOptionParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%02d] %-24s : %s", paramSelector, paramSelector.String(), content)
	}

	out := ""

	if bootOptions.SetInProgress != nil {
		out += format(bootOptions.SetInProgress)
	}

	if bootOptions.ServicePartitionSelector != nil {
		out += format(bootOptions.ServicePartitionSelector)
	}

	if bootOptions.ServicePartitionScan != nil {
		out += format(bootOptions.ServicePartitionScan)
	}

	if bootOptions.BMCBootFlagValidBitClear != nil {
		out += format(bootOptions.BMCBootFlagValidBitClear)
	}

	if bootOptions.BootInfoAcknowledge != nil {
		out += format(bootOptions.BootInfoAcknowledge)
	}

	if bootOptions.BootFlags != nil {
		out += format(bootOptions.BootFlags)
	}

	if bootOptions.BootInitiatorInfo != nil {
		out += format(bootOptions.BootInitiatorInfo)
	}

	if bootOptions.BootInitiatorMailbox != nil {
		out += format(bootOptions.BootInitiatorMailbox)
	}

	return out
}

type BootOptionParam_SetInProgress struct {
	Value SetInProgress
}

func (p *BootOptionParam_SetInProgress) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_SetInProgress, 0, 0
}

func (p *BootOptionParam_SetInProgress) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.Value = SetInProgress(paramData[0])
	return nil
}

func (p *BootOptionParam_SetInProgress) Pack() []byte {
	return []byte{uint8(p.Value)}
}

func (p *BootOptionParam_SetInProgress) Format() string {
	return p.Value.String()
}

// This value is used to select which service partition BIOS should boot using.
type BootOptionParam_ServicePartitionSelector struct {
	Selector uint8
}

func (p *BootOptionParam_ServicePartitionSelector) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_ServicePartitionSelector, 0, 0
}

func (p *BootOptionParam_ServicePartitionSelector) Pack() []byte {
	return []byte{p.Selector}
}

func (p *BootOptionParam_ServicePartitionSelector) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}
	p.Selector = paramData[0]
	return nil
}

func (p *BootOptionParam_ServicePartitionSelector) Format() string {
	switch p.Selector {
	case 0:
		return "unspecified"
	default:
		return fmt.Sprintf("%#02x", p)
	}
}

type BootOptionParam_ServicePartitionScan struct {
	// data 1 [7:2] - reserved
	//  - [1] - 1b = Request BIOS to scan for specified service partition.
	//               BIOS clears this bit after the requested scan has been performed.
	//  - [0] - 1b = Service Partition discovered.
	//               BIOS sets this bit to indicate it has discovered the specified service partition.
	//               The bit retains the value from the last scan.
	//               Therefore, to get up-to-date status of the discovery state, a scan may need to be requested.
	RequestBIOSScan            bool
	ServicePartitionDiscovered bool
}

func (p *BootOptionParam_ServicePartitionScan) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_ServicePartitionScan, 0, 0
}

func (p *BootOptionParam_ServicePartitionScan) Pack() []byte {
	var b uint8

	if p.RequestBIOSScan {
		b = setBit1(b)
	}

	if p.ServicePartitionDiscovered {
		b = setBit0(b)
	}

	return []byte{b}
}

func (p *BootOptionParam_ServicePartitionScan) Unpack(paramData []byte) error {
	if len(paramData) != 1 {
		return fmt.Errorf("the parameter data length must be 1 byte")
	}

	p.RequestBIOSScan = isBit1Set(paramData[0])
	p.ServicePartitionDiscovered = isBit0Set(paramData[0])
	return nil
}

func (p BootOptionParam_ServicePartitionScan) Format() string {
	s := "\n"

	if p.RequestBIOSScan {
		s += "   - Request BIOS to scan\n"
	}
	if p.ServicePartitionDiscovered {
		s += "   - Service Partition Discoverd\n"
	}

	if s == "\n" {
		s += "    No flag set\n"
	}

	return s
}

type BootOptionParam_BMCBootFlagValidBitClear struct {
	DontClearOnResetPEFOrPowerCyclePEF      bool // corresponding to restart cause: 0x08, 0x09
	DontClearOnCommandReceivedTimeout       bool // corresponding to restart cause: 0x01
	DontClearOnWatchdogTimeout              bool // corresponding to restart cause: 0x04
	DontClearOnResetPushButtonOrSoftReset   bool // corresponding to restart cause: 0x02, 0x0a
	DontClearOnPowerUpPushButtonOrWakeEvent bool // corresponding to restart cause: 0x03, 0x0b
}

func (p *BootOptionParam_BMCBootFlagValidBitClear) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_BMCBootFlagValidBitClear, 0, 0
}

func (p *BootOptionParam_BMCBootFlagValidBitClear) Format() string {
	s := "\n"

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
		s += "   - Don't clear valid bit on power up via power push button or wake event\n"
	}

	// When any flag was set, then at least one of the above condition will be true, thus 's' would not be empty.
	if s == "\n" {
		s += "    No flag set\n"
	}

	return s
}

func (p *BootOptionParam_BMCBootFlagValidBitClear) Pack() []byte {
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

func (p *BootOptionParam_BMCBootFlagValidBitClear) Unpack(parameterData []byte) error {
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

type BootOptionParam_BootInfoAcknowledge struct {
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

func (p *BootOptionParam_BootInfoAcknowledge) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_BootInfoAcknowledge, 0, 0
}

func (p *BootOptionParam_BootInfoAcknowledge) Format() string {
	s := "\n"

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
		s += "   - BIOS/POST has handled boot info\n"
	}

	if s == "\n" {
		s += "    No flag set\n"
	}

	return fmt.Sprint(s)
}

func (p *BootOptionParam_BootInfoAcknowledge) Pack() []byte {
	var out = make([]byte, 2)

	var b uint8 = 0x00
	var b1 uint8 = 0xe0 // bit 7,6,5 is reserved, write s 1b

	if p.ByOEM {
		b = setBit4(b)
		b1 = setBit4(b1)
	}
	if p.BySMS {
		b = setBit3(b)
		b1 = setBit3(b1)
	}
	if p.ByOSServicePartition {
		b = setBit2(b)
		b1 = setBit2(b1)
	}
	if p.ByOSLoader {
		b = setBit1(b)
		b1 = setBit1(b1)
	}
	if p.ByBIOSPOST {
		b = setBit0(b)
		b1 = setBit0(b1)
	}
	packUint8(b, out, 0)
	packUint8(b1, out, 1)
	return out
}

func (p *BootOptionParam_BootInfoAcknowledge) Unpack(parameterData []byte) error {
	if len(parameterData) != 2 {
		return fmt.Errorf("the parameter data length must be 2 bytes")
	}

	b, _, _ := unpackUint8(parameterData, 1)
	p.ByOEM = isBit4Set(b)
	p.BySMS = isBit3Set(b)
	p.ByOSServicePartition = isBit2Set(b)
	p.ByOSLoader = isBit1Set(b)
	p.ByBIOSPOST = isBit0Set(b)
	return nil
}

type BootOptionParam_BootFlags struct {
	// 1b = boot flags valid.
	// The bit should be set to indicate that valid flag data is present.
	// This bit may be automatically cleared based on the boot flag valid bit clearing parameter, above.
	BootFlagsValid bool
	// 0b = options apply to next boot only.
	// 1b = options requested to be persistent for all future boots (i.e. requests BIOS to change its boot settings)
	Persist bool
	// 0b = "PC compatible" boot (legacy)
	// 1b = Extensible Firmware Interface Boot (EFI)
	BIOSBootType BIOSBootType

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

func (p *BootOptionParam_BootFlags) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_BootFlags, 0, 0
}

func (p *BootOptionParam_BootFlags) Pack() []byte {
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

func (p *BootOptionParam_BootFlags) Unpack(parameterData []byte) error {
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

func (p *BootOptionParam_BootFlags) Format() string {
	s := "\n"

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

	s += fmt.Sprintf("   - BIOS Mux Control Override : %s\n", p.BIOSMuxControl.String())

	return s
}

func (bootFlags *BootOptionParam_BootFlags) OptionsHelp() string {
	supportedOptions := []struct {
		name string
		help string
	}{
		{"help", "print help message"},
		{"valid", "Boot flags valid"},
		{"persistent", "Changes are persistent for all future boots"},
		{"efiboot", "Extensible Firmware Interface Boot (EFI)"},
		{"clear-cmos", "CMOS clear"},
		{"lockkbd", "Lock Keyboard"},
		{"screenblank", "Screen Blank"},
		{"lockoutreset", "Lock out Resetbuttons"},
		{"lockout_power", "Lock out (power off/sleep request) via Power Button"},
		{"verbose=default", "Request quiet BIOS display"},
		{"verbose=no", "Request quiet BIOS display"},
		{"verbose=yes", "Request verbose BIOS display"},
		{"force_pet", "Force progress event traps"},
		{"upw_bypass", "User password bypass"},
		{"lockout_sleep", "Log Out Sleep Button"},
		{"cons_redirect=default", "Console redirection occurs per BIOS configuration setting"},
		{"cons_redirect=skip", "Suppress (skip) console redirection if enabled"},
		{"cons_redirect=enable", "Suppress (skip) console redirection if enabled"},
	}

	var buf bytes.Buffer
	buf.WriteString("Legal options settings are:\n")
	for _, o := range supportedOptions {
		buf.WriteString(fmt.Sprintf("  %-22s : %s\n", o.name, o.help))
	}

	return buf.String()
}

func (bootFlags *BootOptionParam_BootFlags) ParseFromOptionsStr(optionsStr string) error {
	options := strings.Split(optionsStr, ",")
	return bootFlags.ParseFromOptions(options)
}

func (bootFlags *BootOptionParam_BootFlags) ParseFromOptions(options []string) error {
	for _, option := range options {
		switch option {
		case "valid":
			bootFlags.BootFlagsValid = true
		case "persistent":
			bootFlags.Persist = true
		case "efiboot":
			bootFlags.BIOSBootType = BIOSBootTypeEFI
		case "clear-cmos":
			bootFlags.CMOSClear = true
		case "lockkbd":
			bootFlags.LockKeyboard = true
		case "screenblank":
			bootFlags.ScreenBlank = true
		case "lockoutreset":
			bootFlags.LockoutResetButton = true
		case "lockout_power":
			bootFlags.LockoutPowerOff = true
		case "verbose=default":
			bootFlags.BIOSVerbosity = BIOSVerbosityDefault
		case "verbose=no":
			bootFlags.BIOSVerbosity = BIOSVerbosityQuiet
		case "verbose=yes":
			bootFlags.BIOSVerbosity = BIOSVerbosityVerbose
		case "force_pet":
			bootFlags.ForceProgressEventTraps = true
		case "upw_bypass":
			bootFlags.BypassUserPassword = true
		case "lockout_sleep":
			bootFlags.LockoutSleepButton = true
		case "cons_redirect=default":
			bootFlags.ConsoleRedirectionControl = ConsoleRedirectionControl_Default
		case "cons_redirect=skip":
			bootFlags.ConsoleRedirectionControl = ConsoleRedirectionControl_Skip
		case "cons_redirect=enable":
			bootFlags.ConsoleRedirectionControl = ConsoleRedirectionControl_Enable
		default:
			return fmt.Errorf("unsupported boot flag option, supported options: \n%s", bootFlags.OptionsHelp())
		}
	}

	return nil

}

type BootOptionParam_BootInitiatorInfo struct {
	ChannelNumber     uint8
	SessionID         uint32
	BootInfoTimestamp time.Time
}

func (p *BootOptionParam_BootInitiatorInfo) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_BootInitiatorInfo, 0, 0
}

func (p *BootOptionParam_BootInitiatorInfo) Format() string {
	return fmt.Sprintf(`
    Channel Number : %d
    Session Id     : %d
    Timestamp      : %s`, p.ChannelNumber, p.SessionID, p.BootInfoTimestamp)
}

func (p *BootOptionParam_BootInitiatorInfo) Pack() []byte {
	out := make([]byte, 9)
	packUint8(p.ChannelNumber, out, 0)
	packUint32L(p.SessionID, out, 1)

	ts := uint32(p.BootInfoTimestamp.Unix())
	packUint32L(ts, out, 5)
	return out
}

func (p *BootOptionParam_BootInitiatorInfo) Unpack(parameterData []byte) error {
	if len(parameterData) != 9 {
		return fmt.Errorf("the parameter data length must be 9 bytes")
	}

	p.ChannelNumber, _, _ = unpackUint8(parameterData, 0)
	p.SessionID, _, _ = unpackUint32L(parameterData, 1)

	ts, _, _ := unpackUint32L(parameterData, 5)
	p.BootInfoTimestamp = parseTimestamp(ts)
	return nil
}

type BootOptionParam_BootInitiatorMailbox struct {
	SetSelector uint8
	BlockData   []byte
}

func (p *BootOptionParam_BootInitiatorMailbox) BootOptionParameter() (paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) {
	return BootOptionParamSelector_BootInitiatorMailbox, 0, 0
}

func (p *BootOptionParam_BootInitiatorMailbox) Format() string {
	return fmt.Sprintf(`
    Selector   : %d
    Block Data : %02x
`, p.SetSelector, p.BlockData)
}

func (p *BootOptionParam_BootInitiatorMailbox) Pack() []byte {
	out := make([]byte, 1+len(p.BlockData))
	packUint8(p.SetSelector, out, 0)
	packBytes(p.BlockData, out, 1)
	return out
}

func (p *BootOptionParam_BootInitiatorMailbox) Unpack(parameterData []byte) error {
	if len(parameterData) < 1 {
		return fmt.Errorf("the parameter data length must be at least 1 bytes")
	}

	p.SetSelector, _, _ = unpackUint8(parameterData, 0)
	p.BlockData, _, _ = unpackBytes(parameterData, 1, len(parameterData)-1)
	return nil
}
