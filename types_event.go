package ipmi

import "fmt"

// 31.6.1 SEL Record Type Ranges
type SELRecordType uint8
type SELRecordTypeRange string

const (
	// Range reserved for standard SEL Record Types.
	// As of this writing, only type 02h is defined.
	// Records are automatically timestamped unless otherwise indicated
	// 00h - BFh
	SELRecordTypeRangeStandard SELRecordTypeRange = "standard"

	// 32.2 OEM SEL Record - Type C0h-DFh
	// Range reserved for timestamped OEM SEL records.
	// These records are automatically timestamped by the SEL Device
	// C0h - DFh
	SELRecordTypeRangeTimestampedOEM SELRecordTypeRange = "timestamped OEM"

	// 32.3 OEM SEL Record - Type E0h-FFh
	// Range reserved for non-timestamped OEM SEL records.
	// The SEL Device does not automatically timestamp these records.
	// The four bytes passed in the byte locations for the timestamp will be directly entered into the SEL.
	// E0h - FFh
	SELRecordTypeRangeNonTimestampedOEM SELRecordTypeRange = "non-timestamped OEM"
)

// The SELRecordType can be categorized into 3 ranges according to the SELRecordType value.
//   - 00h - BFh -> standard
//   - C0h - DFh -> timestamped OEM
//   - E0h - FFh -> none-timestamped OEM
func (typ SELRecordType) Range() SELRecordTypeRange {
	t := uint8(typ)
	if t <= 0xbf {
		return SELRecordTypeRangeStandard
	}

	if t >= 0xc0 && t <= 0xdf {
		return SELRecordTypeRangeTimestampedOEM
	}

	// t >= 0xe0 && t <= 0xff
	return SELRecordTypeRangeNonTimestampedOEM
}

func (typ SELRecordType) String() string {
	return string(typ.Range())
}

// Event direction, true for deassertion, false for assertion.
//
// see: 32.1 SEL Event Records Table (Byte 13)
type EventDir bool

const (
	EventDirDeassertion EventDir = true
	EventDirAssertion   EventDir = false
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

// 29.7 Event Data Field Formats
// Event Data 1
// [3:0] -
// for threshold sensors: Offset from Event/Reading Code for threshold event.
// for discrete sensors: Offset from Event/Reading Code for discrete event state (corresponding 15 possible discrete events)
func (ed *EventData) EventReadingOffset() uint8 {
	return ed.EventData1 & 0x0f
}

func (ed *EventData) String() string {
	return fmt.Sprintf("%02x%02x%02x", ed.EventData1, ed.EventData2, ed.EventData3)
}

// 41.2 Event/Reading Type Code
// 42.1 Event/Reading Type Codes
type EventReadingType uint8

const (
	// Unspecified
	EventReadingTypeUnspecified EventReadingType = 0x00

	// Threshold
	EventReadingTypeThreshold EventReadingType = 0x01

	// Generic
	EventReadingTypeTransitionState        EventReadingType = 0x02
	EventReadingTypeState                  EventReadingType = 0x03
	EventReadingTypePredictiveFailure      EventReadingType = 0x04
	EventReadingTypeLimit                  EventReadingType = 0x05
	EventReadingTypePerformance            EventReadingType = 0x06
	EventReadingTypeTransitionSeverity     EventReadingType = 0x07
	EventReadingTypeDevicePresent          EventReadingType = 0x08
	EventReadingTypeDeviceEnabled          EventReadingType = 0x09
	EventReadingTypeTransitionAvailability EventReadingType = 0x0a
	EventReadingTypeRedundancy             EventReadingType = 0x0b
	EventReadingTypeACPIPowerState         EventReadingType = 0x0c
	EventReadingTypeSensorSpecific         EventReadingType = 0x6f

	// OEM
	EventReadingTypeOEMMin EventReadingType = 0x70
	EventReadingTypeOEMMax EventReadingType = 0x7f

	// Reserved
)

func (typ EventReadingType) String() string {
	var c string
	switch typ {
	case EventReadingTypeUnspecified:
		c = "Unspecified"
	case EventReadingTypeThreshold:
		c = "Threshold"
	case EventReadingTypeSensorSpecific:
		c = "Sensor Specific"
	default:
		if typ >= 0x02 && typ <= 0x0c {
			c = "Generic"
		} else if typ >= EventReadingTypeOEMMin && typ <= EventReadingTypeOEMMax {
			c = "OEM"
		} else {
			c = "Reserved"
		}
	}
	return c
}

func (typ EventReadingType) SensorClass() SensorClass {
	if typ == EventReadingTypeThreshold {
		return SensorClassThreshold
	}
	return SensorClassDiscrete
}

func (typ EventReadingType) IsThreshold() bool {
	return typ == EventReadingTypeThreshold
}

// EventString returns description of the event
func (typ EventReadingType) EventString(sensorType SensorType, eventData EventData) string {
	event := typ.Event(sensorType, eventData)

	if event == nil {
		return ""
	}
	return event.EventName
}

// EventSeverity return the severity for the event.
// Todo, refactor
func (typ EventReadingType) EventSeverity(sensorType SensorType, eventData EventData, eventDir EventDir) EventSeverity {
	event := typ.Event(sensorType, eventData)

	if event == nil {
		return EventSeverityInfo
	}

	switch typ {
	case EventReadingTypeUnspecified:
		return EventSeverityInfo

	case EventReadingTypeThreshold:
		if !eventDir {
			if v, ok := event.AssertionSeverityMap[sensorType]; ok {
				return v
			}
			if v, ok := event.AssertionSeverityMap[SensorTypeReserved]; ok {
				return v
			}
			return EventSeverityInfo
		} else {
			if v, ok := event.DeassertionSeverityMap[sensorType]; ok {
				return v
			}
			if v, ok := event.DeassertionSeverityMap[SensorTypeReserved]; ok {
				return v
			}
			return EventSeverityInfo
		}

	case EventReadingTypeSensorSpecific:
		if !eventDir {
			return event.AssertionSeverity
		}
		return event.DeassertionSeverity

	default:
		if typ >= 0x02 && typ <= 0x0c {
			if !eventDir {
				if v, ok := event.AssertionSeverityMap[sensorType]; ok {
					return v
				}
				if v, ok := event.AssertionSeverityMap[SensorTypeReserved]; ok {
					return v
				}
				return EventSeverityInfo
			} else {
				if v, ok := event.DeassertionSeverityMap[sensorType]; ok {
					return v
				}
				if v, ok := event.DeassertionSeverityMap[SensorTypeReserved]; ok {
					return v
				}
				return EventSeverityInfo
			}

		} else if typ >= EventReadingTypeOEMMin && typ <= EventReadingTypeOEMMax {
			return EventSeverityInfo
		} else {
			return EventSeverityInfo
		}
	}

}

// Event return the predefined Event description struct.
func (typ EventReadingType) Event(sensorType SensorType, eventData EventData) *Event {
	offset := eventData.EventReadingOffset()
	return typ.EventForOffset(sensorType, offset)
}

func (typ EventReadingType) EventForOffset(sensorType SensorType, eventOffset uint8) *Event {
	switch typ {
	case EventReadingTypeUnspecified:
		return nil
	case EventReadingTypeThreshold:
		return genericEvent(typ, eventOffset)
	case EventReadingTypeSensorSpecific:
		return sensorSpecificEvent(sensorType, eventOffset)
	default:
		if typ >= 0x02 && typ <= 0x0c {
			return genericEvent(typ, eventOffset)
		} else if typ >= EventReadingTypeOEMMin && typ <= EventReadingTypeOEMMax {
			return oemEvent(sensorType, eventOffset)
		} else {
			return nil
		}
	}
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
	EventName string
	EventDesc string

	// for generic event, different sensor type may means different severity
	AssertionSeverityMap   map[SensorType]EventSeverity
	DeassertionSeverityMap map[SensorType]EventSeverity

	// for sensor specific event, severity is certain.
	AssertionSeverity   EventSeverity
	DeassertionSeverity EventSeverity

	ED2 map[uint8]string // EventData2
	ED3 map[uint8]string // EventData3
}
