package ipmi

type BootInfoAcknowledgeBy uint8

const (
	BootInfoAcknowledgeByBIOSPOST           BootInfoAcknowledgeBy = 1 << 0
	BootInfoAcknowledgeByOSLoader           BootInfoAcknowledgeBy = 1 << 1
	BootInfoAcknowledgeByOSServicePartition BootInfoAcknowledgeBy = 1 << 2
	BootInfoAcknowledgeBySMS                BootInfoAcknowledgeBy = 1 << 3
	BootInfoAcknowledgeByOEM                BootInfoAcknowledgeBy = 1 << 4
)

type BIOSVerbosity uint8 // only 2 bits, occupied 0-3

const (
	BIOSVerbosityDefault BIOSVerbosity = 0
	BIOSVerbosityQuiet   BIOSVerbosity = 1
	BIOSVerbosityVerbose BIOSVerbosity = 2
)

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

const (
	ConsoleRedirectionControl_Default ConsoleRedirectionControl = 0
	ConsoleRedirectionControl_Skip    ConsoleRedirectionControl = 1
	ConsoleRedirectionControl_Enable  ConsoleRedirectionControl = 2
)

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
