package ipmi

import "fmt"

// 31.6.1 SEL Record Type Ranges
type EventRecordType uint8
type EventRecordTypeRange string

const (
	// Range reserved for standard SEL Record Types.
	// As of this writing, only type 02h is defined.
	// Records are automatically timestamped unless otherwise indicated
	// 00h - BFh
	EventRecordTypeRangeStandard EventRecordTypeRange = "standard"

	// 32.2 OEM SEL Record - Type C0h-DFh
	// Range reserved for timestamped OEM SEL records.
	// These records are automatically timestamped by the SEL Device
	// C0h - DFh
	EventRecordTypeRangeTimestampedOEM EventRecordTypeRange = "timestamped OEM"

	// 32.3 OEM SEL Record - Type E0h-FFh
	// Range reserved for non-timestamped OEM SEL records.
	// The SEL Device does not automatically timestamp these records.
	// The four bytes passed in the byte locations for the timestamp will be directly entered into the SEL.
	// E0h - FFh
	EventRecordTypeRangeNonTimestampedOEM EventRecordTypeRange = "non-timestamped OEM"
)

func (typ EventRecordType) Range() EventRecordTypeRange {
	t := uint8(typ)
	if t >= 0x00 && t <= 0xbf {
		return EventRecordTypeRangeStandard
	}

	if t >= 0xc0 && t <= 0xdf {
		return EventRecordTypeRangeTimestampedOEM
	}

	// t >= 0xe0 && t <= 0xff
	return EventRecordTypeRangeNonTimestampedOEM
}

func (typ EventRecordType) String() string {
	return string(typ.Range())
}

type EventDir bool

const (
	EventDirDeassertion = true
	EventDirAssertion   = false
)

func (d EventDir) String() string {
	if d {
		return "Deassertion"
	}
	return "Assertion"
}

// 29.7 Event Data Field Formats
type EventData struct {
	EventData1 uint8
	EventData2 uint8
	EventData3 uint8
}

// 29.7
// Event Data 1
// [3:0] - Offset from Event/Reading Code for threshold event.
func (ed *EventData) EventReadingOffset() uint8 {
	return ed.EventData1 & 0x0f
}

func (ed *EventData) String() string {
	return fmt.Sprintf("%02x%02x%02x", ed.EventData1, ed.EventData2, ed.EventData3)
}

// 41.2 Event/Reading Type Code
// 42.1 Event/Reading Type Codes
type EventReadingType uint8

func (typ EventReadingType) String() string {
	var c string
	switch typ {
	case 0x00:
		c = "Unspecified"
	case 0x01:
		c = "Threshold"
	case 0x6f:
		c = "Sensor Specific"
	default:
		if typ >= 0x02 && typ <= 0x0c {
			c = "Generic"
		} else if typ >= 0x70 && typ <= 0x7f {
			c = "OEM"
		} else {
			c = "Reserved"
		}
	}
	return c
}

// EventString returns description of the event
func (typ EventReadingType) EventString(sensorType SensorType, sensorNumber SensorNumber, eventData EventData) string {
	offset := eventData.EventReadingOffset()

	var s string
	switch typ {
	case 0x00:
		s = "Unspecified"
	case 0x01:
		s = genericEventString(typ, offset)
	case 0x6f:
		s = sensorSpecificEventString(sensorType, offset)
	default:
		if typ >= 0x02 && typ <= 0x0c {
			s = genericEventString(typ, offset)
		} else if typ >= 0x70 && typ <= 0x7f {
			s = oemEventString(sensorType, sensorNumber, offset)
		} else {
			s = "Reserved"
		}
	}
	return s
}

func genericEventString(typ EventReadingType, offset uint8) string {
	e, ok := GenericEvents[uint8(typ)]
	if !ok {
		return ""
	}
	event, ok := e[offset]
	if !ok {
		return ""
	}
	return event.EventName
}

func oemEventString(sensorType SensorType, sensorNumber SensorNumber, offset uint8) string {
	var s string
	return s
}

func sensorSpecificEventString(sensorType SensorType, offset uint8) string {
	e, ok := SensorSpecificEvents[uint8(sensorType)]
	if !ok {
		return ""
	}
	event, ok := e[offset]
	if !ok {
		return ""
	}
	return event
}

type EventSeverity string

const (
	EventSeverityInfo     EventSeverity = "Info"
	EventSeverityOK       EventSeverity = "OK"
	EventSeverityWarning  EventSeverity = "Warning"
	EventSeverityCritical EventSeverity = "Critical"
	EventSeverityDegraded EventSeverity = "Degraded"
	EventSeverityNonFatal EventSeverity = "Non-fatal"
)

type Event struct {
	EventName           string
	EventDesc           string
	AssertionSeverity   EventSeverity
	DeassertionSeverity EventSeverity
	ED2                 map[uint8]string
	ED3                 map[uint8]string
}

// Table 42-2, Generic Event/Reading Type Codes
// Including Genereic threshold-based events (0x01)
// and Generic discrete-based events (0x02 - 0x0c)
// EventReadingType, Offset
var GenericEvents = map[uint8]map[uint8]Event{
	// THRESHOLD BASED STATES
	0x01: {
		0x00: {
			EventName: "Lower Non-critical - going low",
		},
		0x01: {
			EventName: "Lower Non-critical - going high",
		},
		0x02: {
			EventName: "Lower Critical - going low",
		},
		0x03: {
			EventName: "Lower Critical - going high",
		},
		0x04: {
			EventName: "Lower Non-recoverable - going low",
		},
		0x5: {
			EventName: "Lower Non-recoverable - going high",
		},
		0x06: {
			EventName: "Upper Non-critical - going low",
		},
		0x07: {
			EventName: "Upper Non-critical - going high",
		},
		0x08: {
			EventName: "Upper Critical - going low",
		},
		0x09: {
			EventName: "Upper Critical - going high",
		},
		0x0a: {
			EventName: "Upper Non-recoverable - going low",
		},
		0x0b: {
			EventName: "Upper Non-recoverable - going high",
		},
	},
	0x02: {
		0x00: {
			EventName: "Transition to Idle",
		},
		0x01: {
			EventName: "Transition to Active",
		},
		0x02: {
			EventName: "Transition to Busy",
		},
	},
	0x03: {
		0x00: {
			EventName: "State Deasserted",
		},
		0x01: {
			EventName: "State Asserted",
		},
	},
	0x04: {
		0x00: {
			EventName: "Predictive Failure deasserted",
		},
		0x01: {
			EventName: "Predictive Failure asserted",
		},
	},
	0x05: {
		0x00: {
			EventName: "Limit Not Exceeded",
		},
		0x01: {
			EventName: "Limit Exceeded",
		},
	},

	0x06: {
		0x00: {
			EventName: "Performance Met",
		},
		0x01: {
			EventName: "Performance Lags",
		},
	},
	0x07: {
		0x00: {
			EventName: "transition to OK",
		},
		0x01: {
			EventName: "transition to Non-Critical from OK",
		},
		0x02: {
			EventName: "transition to Critical from less severe",
		},
		0x03: {
			EventName: "transition to Non-recoverable from less severe",
		},
		0x04: {
			EventName: "transition to Non-Critical from more severe",
		},
		0x05: {
			EventName: "transition to Critical from Non-recoverable",
		},
		0x06: {
			EventName: "transition to Non-recoverable",
		},
		0x07: {
			EventName: "Monitor",
		},
		0x08: {
			EventName: "Informational",
		},
	},
	0x08: {
		0x00: {
			EventName: "Device Removed / Device Absent",
		},
		0x01: {
			EventName: "Device Inserted / Device Present",
		},
	},
	0x09: {
		0x00: {
			EventName: "Device Disabled",
		},
		0x01: {
			EventName: "Device Enabled",
		},
	},
	0x0a: {
		0x00: {
			EventName: "transition to Running",
		},
		0x01: {
			EventName: "transition to In Test",
		},
		0x02: {
			EventName: "transition to Power Off",
		},
		0x03: {
			EventName: "transition to On Line",
		},
		0x04: {
			EventName: "transition to Off Line",
		},
		0x05: {
			EventName: "transition to Off Duty",
		},
		0x06: {
			EventName: "transition to Degraded",
		},
		0x07: {
			EventName: "transition to Power Save",
		},
		0x08: {
			EventName: "Install Error",
		},
	},
	0x0b: {
		0x00: {
			EventName: "Fully Redundant",
		},
		0x01: {
			EventName: "Redundancy Lost",
		},
		0x02: {
			EventName: "Redundancy Degraded",
		},
		0x03: {
			EventName: "Non-redundant (Sufficient Resources from Redundant)",
		},
		0x04: {
			EventName: "Non-redundant (Sufficient Resources from Insufficient Resources)",
		},
		0x05: {
			EventName: "Non-redundant (Insufficient Resources)",
		},
		0x06: {
			EventName: "Redundancy Degraded from Fully Redundant",
		},
		0x07: {
			EventName: "Redundancy Degraded from Non-redundant",
		},
	},
	0x0c: {
		0x00: {
			EventName: "D0 Power State",
		},
		0x01: {
			EventName: "D1 Power State",
		},
		0x02: {
			EventName: "D2 Power State",
		},
		0x03: {
			EventName: "D3 Power State",
		},
	},
}

// SensorType, Offset
var SensorSpecificEvents = map[uint8]map[uint8]string{
	0x00: {},
	0x01: {},
	0x02: {},
	0x03: {},
	0x04: {},
	// Physical Security (Chassis Intrusion)
	0x05: {
		0x00: "General Chassis Intrusion",
		0x01: "Drive Bay intrusion",
		0x02: "I/O Card area intrusion",
		0x03: "Processor area intrusion",
		0x04: "LAN Leash Lost (system is unplugged from LAN)",
		0x05: "Unauthorized dock",
		0x06: "FAN area intrusion",
	},
	// Platform Security Violation Attempt
	0x06: {
		0x00: "Secure Mode (Front Panel Lockout) Violation attempt",
		0x01: "Pre-boot Password Violation - user password",
		0x02: "Pre-boot Password Violation attempt - setup password",
		0x03: "Pre-boot Password Violation - network boot password",
		0x04: "Other pre-boot Password Violation",
		0x05: "Out-of-band Access Password Violation",
	},
	// Processor
	0x07: {
		0x00: "IERR",
		0x01: "Thermal Trip",
		0x02: "FRB1/BIST failure",
		0x03: "FRB2/Hang in POST failure",
		0x04: "FRB3/Processor Startup/Initialization failure (CPU didn't start)",
		0x05: "Configuration Error",
		0x06: "SM BIOS 'Uncorrectable CPU-complex Error'",
		0x07: "Processor Presence detected",
		0x08: "Processor disabled",
		0x09: "Terminator Presence Detected",
		0x0a: "Processor Automatically Throttled",
		0x0b: "Machine Check Exception (Uncorrectable)",
		0x0c: "Correctable Machine Check Error",
	},
	// Power Supply
	0x08: {
		0x00: "Presence detected",
		0x01: "Power Supply Failure detected",
		0x02: "Predictive Failure",
		0x03: "Power Supply input lost (AC/DC)",
		0x04: "Power Supply input lost or out-of-range",
		0x05: "Power Supply input out-of-range, but present",
		0x06: "Configuration error",
		0x07: "Power Supply Inactive (in standby state)",
	},
	// Power Unit
	0x09: {
		0x00: "Power Off / Power Dow",
		0x01: "Power Cycle",
		0x02: "240VA Power Down",
		0x03: "Interlock Power Down",
		0x04: "AC lost / Power input lost ",
		0x05: "Soft Power Control Failure",
		0x06: "Power Unit Failure detected",
		0x07: "Predictive Failure",
	},
	// Cooling Device
	0x0a: {},
	// Other Units-based Sensor (per units given in SDR)
	0x0b: {},
	// Memory
	0x0c: {
		0x00: "Correctable ECC / other correctable memory error",
		0x01: "Uncorrectable ECC / other uncorrectable memory error",
		0x02: "Parity",
		0x03: "Memory Scrub Failed (stuck bit)",
		0x04: "Memory Device Disabled",
		0x05: "Correctable ECC / other correctable memory error logging limit reached",
		0x06: "Presence detected",
		0x07: "Configuration error",
		0x08: "Spare",
		0x09: "Memory Automatically Throttled",
		0x0a: "Critical Overtemperature",
	},
	// Drive Slot (Bay)
	0x0d: {
		0x00: "Drive Presence",
		0x01: "Drive Fault",
		0x02: "Predictive Failure",
		0x03: "Hot Spare",
		0x04: "Consistency Check / Parity Check in progress",
		0x05: "In Critical Array",
		0x06: "In Failed Array",
		0x07: "Rebuild/Remap in progress",
		0x08: "Rebuild/Remap Aborted (was not completed normally)",
	},
	// POST Memory Resize
	0x0e: {},
	// System Firmware Progress (formerly POST Error)
	0x0f: {
		0x00: "System Firmware Error (POST Error)",
		0x01: "System Firmware Hang",
		0x02: "System Firmware Progress",
	},
	// Event Logging Disabled
	0x10: {
		0x00: "Correctable Memory Error Logging Disabled",
		0x01: "Event 'Type' Logging Disabled",
		0x02: "Log Area Reset/Cleared",
		0x03: "All Event Logging Disabled",
		0x04: "SEL Full",
		0x05: "SEL Almost Full",
		0x06: "Correctable Machine Check Error Logging Disabled",
	},
	// Watchdog 1
	0x11: {
		0x00: "BIOS Watchdog Reset",
		0x01: "OS Watchdog Reset",
		0x02: "OS Watchdog Shut Down",
		0x03: "OS Watchdog Power Down",
		0x04: "OS Watchdog Power Cycle",
		0x05: "OS Watchdog NMI / Diagnostic Interrupt",
		0x06: "OS Watchdog Expired, status only",
		0x07: "OS Watchdog pre-timeout Interrupt, non-NMI",
	},
	// System Event
	0x12: {
		0x00: "System Reconfigured",
		0x01: "OEM System Boot Event",
		0x02: "Undetermined system hardware failure",
		0x03: "Entry added to Auxiliary Log",
		0x04: "PEF Action",
		0x05: "Timestamp Clock Synch",
	},
	// Critical Interrupt
	0x13: {
		0x00: "Front Panel NMI / Diagnostic Interrupt",
		0x01: "Bus Timeout",
		0x02: "I/O channel check NMI",
		0x03: "Software NMI",
		0x04: "PCI PERR",
		0x05: "PCI SERR",
		0x06: "EISA Fail Safe Timeout",
		0x07: "Bus Correctable Error",
		0x08: "Bus Uncorrectable Error",
		0x09: "Fatal NMI",
		0x0a: "Bus Fatal Error",
		0x0b: "Bus Degraded",
	},
	// Button / Switch
	0x14: {
		0x00: "Power Button pressed",
		0x01: "Sleep Button pressed",
		0x02: "Reset Button pressed",
		0x03: "FRU latch open",
		0x04: "FRU service request button",
	},
	// Module / Boar
	0x15: {},
	// Microcontroller / Coprocessor
	0x16: {},
	// Add-in Card
	0x17: {},
	// Chassis
	0x18: {},
	// Chip Set
	0x19: {
		0x00: "Soft Power Control Failure",
		0x01: "Thermal Trip",
	},
	// Other FRU
	0x1a: {},
	// Cable / Interconnect
	0x1b: {
		0x00: "Cable/Interconnect is connected",
		0x01: "Configuration Error - Incorrect cable connected / Incorrect interconnection",
	},
	// Terminator
	0x1c: {},
	// System Boot / Restart Initiated
	0x1d: {
		0x00: "Initiated by power up",
		0x01: "Initiated by hard reset",
		0x02: "Initiated by warm reset",
		0x03: "User requested PXE boot",
		0x04: "Automatic boot to diagnostic",
		0x05: "OS / run-time software initiated hard reset",
		0x06: "OS / run-time software initiated warm reset",
		0x07: "System Restart ",
	},
	// Boot Error
	0x1e: {
		0x00: "No bootable media",
		0x01: "Non-bootable diskette left in drive",
		0x02: "PXE Server not found",
		0x03: "Invalid boot sector",
		0x04: "Timeout waiting for user selection of boot source",
	},
	// Base OS Boot / Installation Status
	0x1f: {
		0x00: "A: boot completed",
		0x01: "C: boot completed",
		0x02: "PXE boot completed",
		0x03: "Diagnostic boot completed",
		0x04: "CD-ROM boot completed",
		0x05: "ROM boot completed",
		0x06: "boot completed - boot device not specified",
		0x07: "Base OS/Hypervisor Installation started",
		0x08: "Base OS/Hypervisor Installation completed",
		0x09: "Base OS/Hypervisor Installation aborted",
		0x0a: "Base OS/Hypervisor Installation failed",
	},
	// OS Stop / Shutdown
	0x20: {
		0x00: "Critical stop during OS load / initialization",
		0x01: "Run-time Critical Stop",
		0x02: "OS Graceful Stop",
		0x03: "OS Graceful Shutdown",
		0x04: "Soft Shutdown initiated by PEF",
		0x05: "Agent Not Responding",
	},
	// Slot / Connector
	0x21: {
		0x00: "Fault Status asserted",
		0x01: "Identify Status asserted",
		0x02: "Slot / Connector Device installed/attached",
		0x03: "Slot / Connector Ready for Device Installation",
		0x04: "Slot/Connector Ready for Device Removal",
		0x05: "Slot Power is Off",
		0x06: "Slot / Connector Device Removal Request",
		0x07: "Interlock asserted",
		0x08: "Slot is Disabled",
		0x09: "Slot holds spare device",
	},
	// System ACPI Power State
	0x22: {
		0x00: "S0 / G0 (working)",
		0x01: "S1 (sleeping with system h/w & processor context maintained)",
		0x02: "S2 (sleeping, processor context lost)",
		0x03: "S3 (sleeping, processor & h/w context lost, memory retained)",
		0x04: "S4 (non-volatile sleep / suspend-to disk)",
		0x05: "S5 / G2 (soft-off)",
		0x06: "S4 / S5 soft-off, particular S4 / S5 state cannot be determined",
		0x07: "G3 / Mechanical Off",
		0x08: "Sleeping in an S1, S2, or S3 states",
		0x09: "G1 sleeping",
		0x0a: "S5 entered by override",
		0x0b: "Legacy ON state",
		0x0c: "Legacy OFF state",
		0x0e: "Unknown",
	},
	// Watchdog 2
	0x23: {
		0x00: "Timer expired, status only (no action, no interrupt)",
		0x01: "Hard Reset",
		0x02: "Power Down",
		0x03: "Power Cycle",
		0x08: "Timer interrupt",
	},
	// Platform Alert
	0x24: {
		0x00: "platform generated page",
		0x01: "platform generated LAN alert",
		0x02: "Platform Event Trap generated",
		0x03: "platform generated SNMP trap",
	},
	// Entity Presence
	0x25: {
		0x00: "Entity Present",
		0x01: "Entity Absent",
		0x02: "Entity Disable",
	},
	// Monitor ASIC / IC
	0x26: {},
	// LAN
	0x27: {
		0x00: "LAN Heartbeat Lost",
		0x01: "LAN Heartbeat",
	},
	// Management Subsystem Health
	0x28: {
		0x00: "sensor access degraded or unavailable",
		0x01: "controller access degraded or unavailable",
		0x02: "management controller off-line",
		0x03: "management controller unavailable",
		0x04: "Sensor failure",
		0x05: "FRU failure",
	},
	// Battery
	0x29: {
		0x00: "FRU failure",
		0x01: "battery failed",
		0x02: "battery presence detected",
	},
	// Session Audit
	0x2a: {
		0x00: "Session Activated",
		0x01: "Session Deactivated",
		0x02: "Invalid Username or Password",
		0x03: "Invalid password disable",
	},
	// Version Change
	0x2b: {
		0x00: "Hardware change detected with associated Entity",
		0x01: "Firmware or software change detected with associated Entity",
		0x02: "Hardware incompatibility detected with associated Entity",
		0x03: "Firmware or software incompatibility detected with associated Entity",
		0x04: "Entity is of an invalid or unsupported hardware version",
		0x05: "Entity contains an invalid or unsupported firmware or software version",
		0x06: "Hardware Change detected with associated Entity was successfu",
		0x07: "Software or F/W Change detected with associated Entity was successful",
	},
	// FRU State
	0x2c: {
		0x00: "FRU Not Installed",
		0x01: "FRU Inactive (in standby or 'hot spare' state)",
		0x02: "FRU Activation Requested",
		0x03: "FRU Activation In Progress",
		0x04: "FRU Active",
		0x05: "FRU Deactivation Requested",
		0x06: "FRU Deactivation In Progress",
		0x07: "FRU Communication Lost",
	},
}
