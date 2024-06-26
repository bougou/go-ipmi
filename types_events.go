package ipmi

// Table 42-2, Generic Event/Reading Type Codes
// Including Genereic threshold-based events (0x01)
// and Generic discrete-based events (0x02 - 0x0c)
// EventReadingType, Offset
//
// The severity is copied from
// freeipmi/libfreeipmi/interpret/ipmi-interpret-config-sel.c
var GenericEvents = map[EventReadingType]map[uint8]Event{
	EventReadingTypeThreshold: {
		0x00: {
			EventName: "Lower Non-critical - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Lower Non-critical - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x02: {
			EventName: "Lower Critical - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x03: {
			EventName: "Lower Critical - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x04: {
			EventName: "Lower Non-recoverable - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x5: {
			EventName: "Lower Non-recoverable - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x06: {
			EventName: "Upper Non-critical - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x07: {
			EventName: "Upper Non-critical - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x08: {
			EventName: "Upper Critical - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x09: {
			EventName: "Upper Critical - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x0a: {
			EventName: "Upper Non-recoverable - going low",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x0b: {
			EventName: "Upper Non-recoverable - going high",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
	},
	EventReadingTypeTransitionState: {
		0x00: {
			EventName: "Transition to Idle",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Transition to Active",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
		},
		0x02: {
			EventName: "Transition to Busy",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeSystemEvent: EventSeverityInfo,
			},
		},
	},
	EventReadingTypeState: {
		0x00: {
			EventName: "State Deasserted",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:               EventSeverityInfo,
				SensorTypeSystemEvent:            EventSeverityInfo,
				SensorTypeButtonSwitch:           EventSeverityInfo,
				SensorTypeModuleBoard:            EventSeverityInfo,
				SensorTypeBootError:              EventSeverityInfo,
				SensorTypeOSStopShutdown:         EventSeverityInfo,
				SensorTypePlatformAlert:          EventSeverityInfo,
				SensorTypeTemperature:            EventSeverityInfo,
				SensorTypeVoltage:                EventSeverityInfo,
				SensorTypeFan:                    EventSeverityInfo,
				SensorTypeProcessor:              EventSeverityInfo,
				SensorTypePowserSupply:           EventSeverityInfo,
				SensorTypePowerUnit:              EventSeverityInfo,
				SensorTypeMemory:                 EventSeverityInfo,
				SensorTypeDriveSlot:              EventSeverityWarning,
				SensorTypePostMemoryResize:       EventSeverityInfo,
				SensorTypeSystemFirmwareProgress: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:               EventSeverityInfo,
				SensorTypeSystemEvent:            EventSeverityInfo,
				SensorTypeButtonSwitch:           EventSeverityInfo,
				SensorTypeModuleBoard:            EventSeverityInfo,
				SensorTypeBootError:              EventSeverityInfo,
				SensorTypeOSStopShutdown:         EventSeverityInfo,
				SensorTypePlatformAlert:          EventSeverityInfo,
				SensorTypeTemperature:            EventSeverityInfo,
				SensorTypeVoltage:                EventSeverityInfo,
				SensorTypeFan:                    EventSeverityInfo,
				SensorTypeProcessor:              EventSeverityInfo,
				SensorTypePowserSupply:           EventSeverityInfo,
				SensorTypePowerUnit:              EventSeverityInfo,
				SensorTypeMemory:                 EventSeverityInfo,
				SensorTypeDriveSlot:              EventSeverityWarning,
				SensorTypePostMemoryResize:       EventSeverityInfo,
				SensorTypeSystemFirmwareProgress: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "State Asserted",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:               EventSeverityInfo,
				SensorTypeSystemEvent:            EventSeverityWarning,
				SensorTypeButtonSwitch:           EventSeverityInfo,
				SensorTypeModuleBoard:            EventSeverityCritical,
				SensorTypeBootError:              EventSeverityCritical,
				SensorTypeOSStopShutdown:         EventSeverityCritical,
				SensorTypePlatformAlert:          EventSeverityCritical,
				SensorTypeTemperature:            EventSeverityWarning,
				SensorTypeVoltage:                EventSeverityWarning,
				SensorTypeFan:                    EventSeverityWarning,
				SensorTypeProcessor:              EventSeverityCritical,
				SensorTypePowserSupply:           EventSeverityWarning,
				SensorTypePowerUnit:              EventSeverityWarning,
				SensorTypeMemory:                 EventSeverityCritical,
				SensorTypeDriveSlot:              EventSeverityInfo,
				SensorTypePostMemoryResize:       EventSeverityWarning,
				SensorTypeSystemFirmwareProgress: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:               EventSeverityInfo,
				SensorTypeSystemEvent:            EventSeverityWarning,
				SensorTypeButtonSwitch:           EventSeverityInfo,
				SensorTypeModuleBoard:            EventSeverityCritical,
				SensorTypeBootError:              EventSeverityCritical,
				SensorTypeOSStopShutdown:         EventSeverityCritical,
				SensorTypePlatformAlert:          EventSeverityCritical,
				SensorTypeTemperature:            EventSeverityWarning,
				SensorTypeVoltage:                EventSeverityWarning,
				SensorTypeFan:                    EventSeverityWarning,
				SensorTypeProcessor:              EventSeverityCritical,
				SensorTypePowserSupply:           EventSeverityWarning,
				SensorTypePowerUnit:              EventSeverityWarning,
				SensorTypeMemory:                 EventSeverityCritical,
				SensorTypeDriveSlot:              EventSeverityInfo,
				SensorTypePostMemoryResize:       EventSeverityWarning,
				SensorTypeSystemFirmwareProgress: EventSeverityWarning,
			},
		},
	},
	EventReadingTypePredicitiveFailure: {
		0x00: {
			EventName: "Predictive Failure deasserted",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:  EventSeverityInfo,
				SensorTypeDriveSlot: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:  EventSeverityInfo,
				SensorTypeDriveSlot: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Predictive Failure asserted",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:  EventSeverityCritical,
				SensorTypeDriveSlot: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:  EventSeverityCritical,
				SensorTypeDriveSlot: EventSeverityCritical,
			},
		},
	},
	EventReadingTypeLimit: {
		0x00: {
			EventName: "Limit Not Exceeded",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeTemperature: EventSeverityInfo,
				SensorTypeVoltage:     EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityInfo,
				SensorTypeTemperature: EventSeverityInfo,
				SensorTypeVoltage:     EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Limit Exceeded",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityCritical,
				SensorTypeTemperature: EventSeverityCritical,
				SensorTypeVoltage:     EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved:    EventSeverityCritical,
				SensorTypeTemperature: EventSeverityCritical,
				SensorTypeVoltage:     EventSeverityCritical,
			},
		},
	},
	EventReadingTypePeformance: {
		0x00: {
			EventName: "Performance Met",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
				SensorTypeVoltage:  EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
				SensorTypeVoltage:  EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Performance Lags",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
				SensorTypeVoltage:  EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
				SensorTypeVoltage:  EventSeverityCritical,
			},
		},
	},
	EventReadingTypeTransitionSeverity: {
		0x00: {
			EventName: "transition to OK",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "transition to Non-Critical from OK",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x02: {
			EventName: "transition to Critical from less severe",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x03: {
			EventName: "transition to Non-recoverable from less severe",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x04: {
			EventName: "transition to Non-Critical from more severe",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x05: {
			EventName: "transition to Critical from Non-recoverable",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x06: {
			EventName: "transition to Non-recoverable",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x07: {
			EventName: "Monitor",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x08: {
			EventName: "Informational",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
	},
	EventReadingTypeDevicePresent: {
		0x00: {
			EventName: "Device Removed / Device Absent",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x01: {
			EventName: "Device Inserted / Device Present",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
	},
	EventReadingTypeDeviceEnabled: {
		0x00: {
			EventName: "Device Disabled",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x01: {
			EventName: "Device Enabled",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
	},
	EventReadingTypeTransitionAvailability: {
		0x00: {
			EventName: "transition to Running",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "transition to In Test",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x02: {
			EventName: "transition to Power Off",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x03: {
			EventName: "transition to On Line",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x04: {
			EventName: "transition to Off Line",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x05: {
			EventName: "transition to Off Duty",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x06: {
			EventName: "transition to Degraded",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
		0x07: {
			EventName: "transition to Power Save",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x08: {
			EventName: "Install Error",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
		},
	},
	EventReadingTypeRedundancy: {
		0x00: {
			EventName: "Fully Redundant",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "Redundancy Lost",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x02: {
			EventName: "Redundancy Degraded",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x03: {
			EventName: "Non-redundant (Sufficient Resources from Redundant)",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x04: {
			EventName: "Non-redundant (Sufficient Resources from Insufficient Resources)",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x05: {
			EventName: "Non-redundant (Insufficient Resources)",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityCritical,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x06: {
			EventName: "Redundancy Degraded from Fully Redundant",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
		0x07: {
			EventName: "Redundancy Degraded from Non-redundant",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityWarning,
			},
		},
	},
	EventReadingTypeACPIPowerState: {
		0x00: {
			EventName: "D0 Power State",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x01: {
			EventName: "D1 Power State",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x02: {
			EventName: "D2 Power State",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
		0x03: {
			EventName: "D3 Power State",
			AssertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
			DeassertionSeverityMap: map[SensorType]EventSeverity{
				SensorTypeReserved: EventSeverityInfo,
			},
		},
	},
}

// 42.2 Sensor Type Codes and Data
// Sensor Specific Events
// SensorType, Offset
var SensorSpecificEvents = map[SensorType]map[uint8]Event{
	SensorTypeReserved:    {},
	SensorTypeTemperature: {},
	SensorTypeVoltage:     {},
	SensorTypeCurrent:     {},
	SensorTypeFan:         {},
	SensorTypePhysicalSecurity: {
		0x00: {
			EventName:           "General Chassis Intrusion",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Drive Bay intrusion",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "I/O Card area intrusion",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Processor area intrusion",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "LAN Leash Lost (system is unplugged from LAN)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "Unauthorized dock",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "FAN area intrusion",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypePlatformSecurity: {
		0x00: {
			EventName:           "Secure Mode (Front Panel Lockout) Violation attempt",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Pre-boot Password Violation - user password",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Pre-boot Password Violation attempt - setup password",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Pre-boot Password Violation - network boot password",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "Other pre-boot Password Violation",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "Out-of-band Access Password Violation",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeProcessor: {
		0x00: {
			EventName:           "IERR",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Thermal Trip",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "FRB1/BIST failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "FRB2/Hang in POST failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "FRB3/Processor Startup/Initialization failure (CPU didn't start)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "Configuration Error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "SM BIOS 'Uncorrectable CPU-complex Error'",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x07: {
			EventName:           "Processor Presence detected",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x08: {
			EventName:           "Processor disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x09: {
			EventName:           "Terminator Presence Detected",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x0a: {
			EventName:           "Processor Automatically Throttled",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x0b: {
			EventName:           "Machine Check Exception (Uncorrectable)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x0c: {
			EventName:           "Correctable Machine Check Error",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypePowserSupply: {
		0x00: {
			EventName:           "Presence detected",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Power Supply Failure detected",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Predictive Failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Power Supply input lost (AC/DC)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "Power Supply input lost or out-of-range",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x05: {
			EventName:           "Power Supply input out-of-range, but present",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x06: {
			EventName:           "Configuration error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x07: {
			EventName:           "Power Supply Inactive (in standby state)",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypePowerUnit: {
		0x00: {
			EventName:           "Power Off / Power Dow",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Power Cycle",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "240VA Power Down",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "Interlock Power Down",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "AC lost / Power input lost ",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x05: {
			EventName:           "Soft Power Control Failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "Power Unit Failure detected",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x07: {
			EventName:           "Predictive Failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeCollingDevice: {},
	// Other Units-based Sensor (per units given in SDR)
	SensorTypeOtherUnitsbased: {},
	SensorTypeMemory: {
		0x00: {
			EventName:           "Correctable ECC / other correctable memory error",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "Uncorrectable ECC / other uncorrectable memory error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Parity",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Memory Scrub Failed (stuck bit)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "Memory Device Disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "Correctable ECC / other correctable memory error logging limit reached",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x06: {
			EventName:           "Presence detected",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x07: {
			EventName:           "Configuration error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x08: {
			EventName:           "Spare",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x09: {
			EventName:           "Memory Automatically Throttled",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x0a: {
			EventName:           "Critical Overtemperature",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeDriveSlot: {
		0x00: {
			EventName:           "Drive Presence",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Drive Fault",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Predictive Failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Hot Spare",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "Consistency Check / Parity Check in progress",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "In Critical Array",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "In Failed Array",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x07: {
			EventName:           "Rebuild/Remap in progress",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x08: {
			EventName:           "Rebuild/Remap Aborted (was not completed normally)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypePostMemoryResize: {},
	SensorTypeSystemFirmwareProgress: {
		0x00: {
			EventName:           "System Firmware Error (POST Error)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "System Firmware Hang",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "System Firmware Progress",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeEventLoggingDisabled: {
		0x00: {
			EventName:           "Correctable Memory Error Logging Disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Event 'Type' Logging Disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Log Area Reset/Cleared",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "All Event Logging Disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "SEL Full",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "SEL Almost Full",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x06: {
			EventName:           "Correctable Machine Check Error Logging Disabled",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeWatchdog1: {
		0x00: {
			EventName: "BIOS Watchdog Reset",
		},
		0x01: {
			EventName: "OS Watchdog Reset",
		},
		0x02: {
			EventName: "OS Watchdog Shut Down",
		},
		0x03: {
			EventName: "OS Watchdog Power Down",
		},
		0x04: {
			EventName: "OS Watchdog Power Cycle",
		},
		0x05: {
			EventName: "OS Watchdog NMI / Diagnostic Interrupt",
		},
		0x06: {
			EventName: "OS Watchdog Expired, status only",
		},
		0x07: {
			EventName: "OS Watchdog pre-timeout Interrupt, non-NMI",
		},
	},
	SensorTypeSystemEvent: {
		0x00: {
			EventName:           "System Reconfigured",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "OEM System Boot Event",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "Undetermined system hardware failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Entry added to Auxiliary Log",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "PEF Action",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "Timestamp Clock Synch",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypeCriticalInterrupt: {
		0x00: {
			EventName:           "Front Panel NMI / Diagnostic Interrupt",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Bus Timeout",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "I/O channel check NMI",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Software NMI",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "PCI PERR",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "PCI SERR",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "EISA Fail Safe Timeout",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x07: {
			EventName:           "Bus Correctable Error",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x08: {
			EventName:           "Bus Uncorrectable Error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x09: {
			EventName:           "Fatal NMI",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x0a: {
			EventName:           "Bus Fatal Error",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x0b: {
			EventName:           "Bus Degraded",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypeButtonSwitch: {
		0x00: {
			EventName:           "Power Button pressed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Sleep Button pressed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "Reset Button pressed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "FRU latch open",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "FRU service request button",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypeModuleBoard:                {},
	SensorTypeMicrocontrollerCoprocessor: {},
	SensorTypeAddinCard:                  {},
	SensorTypeChassis:                    {},
	SensorTypeChipSet: {
		0x00: {
			EventName:           "Soft Power Control Failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Thermal Trip",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeOtherFRU: {},
	SensorTypeCableInterconnect: {
		0x00: {
			EventName:           "Cable/Interconnect is connected",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Configuration Error - Incorrect cable connected / Incorrect interconnection",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeTerminator: {},
	SensorTypeSystemBootRestartInitiated: {
		0x00: {
			EventName:           "Initiated by power up",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Initiated by hard reset",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "Initiated by warm reset",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "User requested PXE boot",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "Automatic boot to diagnostic",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "OS / run-time software initiated hard reset",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x06: {
			EventName:           "OS / run-time software initiated warm reset",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x07: {
			EventName:           "System Restart",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeBootError: {
		0x00: {
			EventName:           "No bootable media",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Non-bootable diskette left in drive",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "PXE Server not found",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Invalid boot sector",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "Timeout waiting for user selection of boot source",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypeBaseOSBootInstallationStatus: {
		0x00: {
			EventName:           "A: boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "C: boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "PXE boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "Diagnostic boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "CD-ROM boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "ROM boot completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x06: {
			EventName:           "boot completed - boot device not specified",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x07: {
			EventName:           "Base OS/Hypervisor Installation started",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x08: {
			EventName:           "Base OS/Hypervisor Installation completed",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x09: {
			EventName:           "Base OS/Hypervisor Installation aborted",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x0a: {
			EventName:           "Base OS/Hypervisor Installation failed",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeOSStopShutdown: {
		0x00: {
			EventName:           "Critical stop during OS load / initialization",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Run-time Critical Stop",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "OS Graceful Stop",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x03: {
			EventName:           "OS Graceful Shutdown",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "Soft Shutdown initiated by PEF",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x05: {
			EventName:           "Agent Not Responding",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeSlotConnector: {
		0x00: {
			EventName:           "Fault Status asserted",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "Identify Status asserted",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x02: {
			EventName:           "Slot / Connector Device installed/attached",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "Slot / Connector Ready for Device Installation",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "Slot/Connector Ready for Device Removal",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "Slot Power is Off",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x06: {
			EventName:           "Slot / Connector Device Removal Request",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x07: {
			EventName:           "Interlock asserted",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x08: {
			EventName:           "Slot is Disabled",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x09: {
			EventName:           "Slot holds spare device",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeSystemACPIPowerState: {
		0x00: {
			EventName:           "S0 / G0 (working)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "S1 (sleeping with system h/w & processor context maintained)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "S2 (sleeping, processor context lost)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "S3 (sleeping, processor & h/w context lost, memory retained)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x04: {
			EventName:           "S4 (non-volatile sleep / suspend-to disk)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "S5 / G2 (soft-off)",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x06: {
			EventName:           "S4 / S5 soft-off, particular S4 / S5 state cannot be determined",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x07: {
			EventName:           "G3 / Mechanical Off",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x08: {
			EventName:           "Sleeping in an S1, S2, or S3 states",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x09: {
			EventName:           "G1 sleeping",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x0a: {
			EventName:           "S5 entered by override",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x0b: {
			EventName:           "Legacy ON state",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x0c: {
			EventName:           "Legacy OFF state",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x0e: {
			EventName:           "Unknown",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeWatchdog2: {
		0x00: {
			EventName:           "Timer expired, status only (no action, no interrupt)",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "Hard Reset",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Power Down",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Power Cycle",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x08: {
			EventName:           "Timer interrupt",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
	},
	SensorTypePlatformAlert: {
		0x00: {
			EventName:           "platform generated page",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "platform generated LAN alert",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "Platform Event Trap generated",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x03: {
			EventName:           "platform generated SNMP trap",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeEntityPresence: {
		0x00: {
			EventName:           "Entity Present",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Entity Absent",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "Entity Disable",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeMonitorASIC: {},
	SensorTypeLAN: {
		0x00: {
			EventName:           "LAN Heartbeat Lost",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "LAN Heartbeat",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeManagementSubsystemHealth: {
		0x00: {
			EventName:           "sensor access degraded or unavailable",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "controller access degraded or unavailable",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "management controller off-line",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "management controller unavailable",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "Sensor failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "FRU failure",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeBattery: {
		0x00: {
			EventName:           "battery low (predictive failure)",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "battery failed",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "battery presence detected",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeSessionAudit: {
		0x00: {
			EventName:           "Session Activated",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x01: {
			EventName:           "Session Deactivated",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x02: {
			EventName:           "Invalid Username or Password",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x03: {
			EventName:           "Invalid password disable",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
	SensorTypeVersionChange: {
		0x00: {
			EventName:           "Hardware change detected with associated Entity",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x01: {
			EventName:           "Firmware or software change detected with associated Entity",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x02: {
			EventName:           "Hardware incompatibility detected with associated Entity",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x03: {
			EventName:           "Firmware or software incompatibility detected with associated Entity",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x04: {
			EventName:           "Entity is of an invalid or unsupported hardware version",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x05: {
			EventName:           "Entity contains an invalid or unsupported firmware or software version",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x06: {
			EventName:           "Hardware Change detected with associated Entity was successfu",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x07: {
			EventName:           "Software or F/W Change detected with associated Entity was successful",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
	},
	SensorTypeFRUState: {
		0x00: {
			EventName:           "FRU Not Installed",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x01: {
			EventName:           "FRU Inactive (in standby or 'hot spare' state)",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
		0x02: {
			EventName:           "FRU Activation Requested",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x03: {
			EventName:           "FRU Activation In Progress",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x04: {
			EventName:           "FRU Active",
			AssertionSeverity:   EventSeverityInfo,
			DeassertionSeverity: EventSeverityInfo,
		},
		0x05: {
			EventName:           "FRU Deactivation Requested",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x06: {
			EventName:           "FRU Deactivation In Progress",
			AssertionSeverity:   EventSeverityWarning,
			DeassertionSeverity: EventSeverityWarning,
		},
		0x07: {
			EventName:           "FRU Communication Lost",
			AssertionSeverity:   EventSeverityCritical,
			DeassertionSeverity: EventSeverityCritical,
		},
	},
}
