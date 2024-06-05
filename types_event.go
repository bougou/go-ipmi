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

// Event direction
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
	EventReadingTypeUnspecified            EventReadingType = 0x00
	EventReadingTypeThreshold              EventReadingType = 0x01
	EventReadingTypeTransitionState        EventReadingType = 0x02
	EventReadingTypeState                  EventReadingType = 0x03
	EventReadingTypePredicitiveFailure     EventReadingType = 0x04
	EventReadingTypeLimit                  EventReadingType = 0x05
	EventReadingTypePeformance             EventReadingType = 0x06
	EventReadingTypeTransitionSeverity     EventReadingType = 0x07
	EventReadingTypeDevicePresent          EventReadingType = 0x08
	EventReadingTypeDeviceEnabled          EventReadingType = 0x09
	EventReadingTypeTransitionAvailability EventReadingType = 0x0a
	EventReadingTypeRedundancy             EventReadingType = 0x0b
	EventReadingTypeACPIPowerState         EventReadingType = 0x0c
	EventReadingTypeSensorSpecific         EventReadingType = 0x6f
	EventReadingTypeOEMMin                 EventReadingType = 0x70
	EventReadingTypeOEMMax                 EventReadingType = 0x7f
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
func (typ EventReadingType) EventString(sensorType SensorType, sensorNumber SensorNumber, eventData EventData) string {
	event := typ.Event(sensorType, sensorNumber, eventData)

	if event == nil {
		return ""
	}
	return event.EventName
}

// EventSeverity return the severity for the event.
// Todo, refactor
func (typ EventReadingType) EventSeverity(sensorType SensorType, sensorNumber SensorNumber, eventData EventData, eventDir EventDir) EventSeverity {
	event := typ.Event(sensorType, sensorNumber, eventData)

	if event == nil {
		return EventSeverityInfo
	}

	switch typ {
	case EventReadingTypeUnspecified:
		return EventSeverityInfo

	case EventReadingTypeThreshold:
		if eventDir {
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
		if eventDir {
			return event.AssertionSeverity
		}
		return event.DeassertionSeverity

	default:
		if typ >= 0x02 && typ <= 0x0c {
			if eventDir {
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
func (typ EventReadingType) Event(sensorType SensorType, sensorNumber SensorNumber, eventData EventData) *Event {
	offset := eventData.EventReadingOffset()

	switch typ {
	case EventReadingTypeUnspecified:
		return nil
	case EventReadingTypeThreshold:
		return genericEvent(typ, offset)
	case EventReadingTypeSensorSpecific:
		return sensorSpecificEvent(sensorType, offset)
	default:
		if typ >= 0x02 && typ <= 0x0c {
			return genericEvent(typ, offset)
		} else if typ >= EventReadingTypeOEMMin && typ <= EventReadingTypeOEMMax {
			return oemEvent(sensorType, sensorNumber, offset)
		} else {
			return nil
		}
	}
}

func genericEvent(typ EventReadingType, offset uint8) *Event {
	e, ok := GenericEvents[typ]
	if !ok {
		return nil
	}
	event, ok := e[offset]
	if !ok {
		return nil
	}
	return &event
}

func oemEvent(sensorType SensorType, sensorNumber SensorNumber, offset uint8) *Event {
	return nil
}

func sensorSpecificEvent(sensorType SensorType, offset uint8) *Event {
	e, ok := SensorSpecificEvents[sensorType]
	if !ok {
		return nil
	}
	event, ok := e[offset]
	if !ok {
		return nil
	}
	return &event
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

	ED2 map[uint8]string
	ED3 map[uint8]string
}

type SensorEvent struct {
	SensorClass SensorClass

	ThresholdType SensorThresholdType
	Assert        bool // true -> assertion events; false -> deassertion events
	High          bool // true -> going high; false -> going low

	State uint8 // state 0-14 (total 15 possible states)
}

func (e SensorEvent) String() string {
	switch e.SensorClass {
	case SensorClassThreshold:
		out := e.ThresholdType.Abbr()
		if e.High {
			out += "+"
		} else {
			out += "-"
		}
		return out
	case SensorClassDiscrete:
		return fmt.Sprintf("state%d", e.State)
	}
	return ""
}

type SensorEvents []SensorEvent

func (events SensorEvents) Strings() []string {
	out := make([]string, 0)
	return out
}
func (events SensorEvents) FilterAssert() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.Assert {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterDeassert() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if !event.Assert {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterThreshold() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.SensorClass == SensorClassThreshold {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterDiscrete() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.SensorClass == SensorClassDiscrete {
			out = append(out, event)
		}
	}
	return out
}

var (
	SensorEvent_UNC_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UNC_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LNR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LNR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LCR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LCR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LNC_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LNC_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        true,
		High:          false,
	}

	SensorEvent_UNR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UNR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_UCR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UCR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_State_14_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       14,
	}

	SensorEvent_State_13_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       13,
	}

	SensorEvent_State_12_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       12,
	}

	SensorEvent_State_11_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       11,
	}

	SensorEvent_State_10_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       10,
	}

	SensorEvent_State_9_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       9,
	}

	SensorEvent_State_8_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       8,
	}

	SensorEvent_State_7_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       7,
	}

	SensorEvent_State_6_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       6,
	}

	SensorEvent_State_5_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       5,
	}

	SensorEvent_State_4_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       4,
	}

	SensorEvent_State_3_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       3,
	}

	SensorEvent_State_2_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       2,
	}

	SensorEvent_State_1_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       1,
	}

	SensorEvent_State_0_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       0,
	}

	SensorEvent_UNC_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        false,
		High:          true,
	}

	// Dessaert Events

	SensorEvent_UNC_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_LCR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LCR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_LNC_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNC_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        false,
		High:          false,
	}

	SensorEvent_UNR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_UNR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_UCR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_UCR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_State_14_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       14,
	}

	SensorEvent_State_13_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       13,
	}

	SensorEvent_State_12_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       12,
	}

	SensorEvent_State_11_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       11,
	}

	SensorEvent_State_10_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       10,
	}

	SensorEvent_State_9_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       9,
	}

	SensorEvent_State_8_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       8,
	}

	SensorEvent_State_7_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       7,
	}

	SensorEvent_State_6_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       6,
	}

	SensorEvent_State_5_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       5,
	}

	SensorEvent_State_4_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       4,
	}

	SensorEvent_State_3_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       3,
	}

	SensorEvent_State_2_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       2,
	}

	SensorEvent_State_1_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       1,
	}

	SensorEvent_State_0_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       0,
	}
)

// SensorEventFlag holds a struct with fields indicating the specified sensor event is set or not.
// SensorEventFlag was embeded in Sensor related commands.
type SensorEventFlag struct {
	SensorEvent_UNC_High_Assert bool
	SensorEvent_UNC_Low_Assert  bool
	SensorEvent_LNR_High_Assert bool
	SensorEvent_LNR_Low_Assert  bool
	SensorEvent_LCR_High_Assert bool
	SensorEvent_LCR_Low_Assert  bool
	SensorEvent_LNC_High_Assert bool
	SensorEvent_LNC_Low_Assert  bool
	SensorEvent_State_7_Assert  bool
	SensorEvent_State_6_Assert  bool
	SensorEvent_State_5_Assert  bool
	SensorEvent_State_4_Assert  bool
	SensorEvent_State_3_Assert  bool
	SensorEvent_State_2_Assert  bool
	SensorEvent_State_1_Assert  bool
	SensorEvent_State_0_Assert  bool

	SensorEvent_UNR_High_Assert bool
	SensorEvent_UNR_Low_Assert  bool
	SensorEvent_UCR_High_Assert bool
	SensorEvent_UCR_Low_Assert  bool
	SensorEvent_State_14_Assert bool
	SensorEvent_State_13_Assert bool
	SensorEvent_State_12_Assert bool
	SensorEvent_State_11_Assert bool
	SensorEvent_State_10_Assert bool
	SensorEvent_State_9_Assert  bool
	SensorEvent_State_8_Assert  bool

	SensorEvent_UNC_High_Deassert bool
	SensorEvent_UNC_Low_Deassert  bool
	SensorEvent_LNR_High_Deassert bool
	SensorEvent_LNR_Low_Deassert  bool
	SensorEvent_LCR_High_Deassert bool
	SensorEvent_LCR_Low_Deassert  bool
	SensorEvent_LNC_High_Deassert bool
	SensorEvent_LNC_Low_Deassert  bool
	SensorEvent_State_7_Deassert  bool
	SensorEvent_State_6_Deassert  bool
	SensorEvent_State_5_Deassert  bool
	SensorEvent_State_4_Deassert  bool
	SensorEvent_State_3_Deassert  bool
	SensorEvent_State_2_Deassert  bool
	SensorEvent_State_1_Deassert  bool
	SensorEvent_State_0_Deassert  bool

	SensorEvent_UNR_High_Deassert bool
	SensorEvent_UNR_Low_Deassert  bool
	SensorEvent_UCR_High_Deassert bool
	SensorEvent_UCR_Low_Deassert  bool
	SensorEvent_State_14_Deassert bool
	SensorEvent_State_13_Deassert bool
	SensorEvent_State_12_Deassert bool
	SensorEvent_State_11_Deassert bool
	SensorEvent_State_10_Deassert bool
	SensorEvent_State_9_Deassert  bool
	SensorEvent_State_8_Deassert  bool
}

// TrueEvents returns a slice of SensorEvent those are set to true in the SensorEventFlag.
func (flag *SensorEventFlag) TrueEvents() []SensorEvent {
	out := make([]SensorEvent, 0)
	if flag.SensorEvent_UNC_High_Assert {
		out = append(out, SensorEvent_UNC_High_Assert)
	}
	if flag.SensorEvent_UNC_Low_Assert {
		out = append(out, SensorEvent_UNC_Low_Deassert)
	}
	if flag.SensorEvent_LNR_High_Assert {
		out = append(out, SensorEvent_LNR_High_Assert)
	}
	if flag.SensorEvent_LNR_Low_Assert {
		out = append(out, SensorEvent_LNR_Low_Assert)
	}
	if flag.SensorEvent_LCR_High_Assert {
		out = append(out, SensorEvent_LCR_High_Assert)
	}
	if flag.SensorEvent_LCR_Low_Assert {
		out = append(out, SensorEvent_LCR_Low_Assert)
	}
	if flag.SensorEvent_LNC_High_Assert {
		out = append(out, SensorEvent_LNC_High_Assert)
	}
	if flag.SensorEvent_LNC_Low_Assert {
		out = append(out, SensorEvent_LNC_Low_Assert)
	}
	if flag.SensorEvent_State_7_Assert {
		out = append(out, SensorEvent_State_7_Assert)
	}
	if flag.SensorEvent_State_6_Assert {
		out = append(out, SensorEvent_State_6_Assert)
	}
	if flag.SensorEvent_State_5_Assert {
		out = append(out, SensorEvent_State_5_Assert)
	}
	if flag.SensorEvent_State_4_Assert {
		out = append(out, SensorEvent_State_4_Assert)
	}
	if flag.SensorEvent_State_3_Assert {
		out = append(out, SensorEvent_State_3_Assert)
	}
	if flag.SensorEvent_State_2_Assert {
		out = append(out, SensorEvent_State_2_Assert)
	}
	if flag.SensorEvent_State_1_Assert {
		out = append(out, SensorEvent_State_1_Assert)
	}
	if flag.SensorEvent_State_0_Assert {
		out = append(out, SensorEvent_State_0_Assert)
	}
	if flag.SensorEvent_UNR_High_Assert {
		out = append(out, SensorEvent_UNR_High_Assert)
	}
	if flag.SensorEvent_UNR_Low_Assert {
		out = append(out, SensorEvent_UNR_Low_Assert)
	}
	if flag.SensorEvent_UCR_High_Assert {
		out = append(out, SensorEvent_UCR_High_Assert)
	}
	if flag.SensorEvent_UCR_Low_Assert {
		out = append(out, SensorEvent_UCR_Low_Assert)
	}
	if flag.SensorEvent_State_14_Assert {
		out = append(out, SensorEvent_State_14_Assert)
	}
	if flag.SensorEvent_State_13_Assert {
		out = append(out, SensorEvent_State_13_Assert)
	}
	if flag.SensorEvent_State_12_Assert {
		out = append(out, SensorEvent_State_12_Assert)
	}
	if flag.SensorEvent_State_11_Assert {
		out = append(out, SensorEvent_State_11_Assert)
	}
	if flag.SensorEvent_State_10_Assert {
		out = append(out, SensorEvent_State_10_Assert)
	}
	if flag.SensorEvent_State_9_Assert {
		out = append(out, SensorEvent_State_9_Assert)
	}
	if flag.SensorEvent_State_8_Assert {
		out = append(out, SensorEvent_State_8_Assert)
	}
	if flag.SensorEvent_UNC_High_Deassert {
		out = append(out, SensorEvent_UNC_High_Deassert)
	}
	if flag.SensorEvent_UNC_Low_Deassert {
		out = append(out, SensorEvent_UNC_Low_Deassert)
	}
	if flag.SensorEvent_LNR_High_Deassert {
		out = append(out, SensorEvent_LNR_High_Deassert)
	}
	if flag.SensorEvent_LNR_Low_Deassert {
		out = append(out, SensorEvent_LNR_Low_Deassert)
	}
	if flag.SensorEvent_LCR_High_Deassert {
		out = append(out, SensorEvent_LCR_High_Deassert)
	}
	if flag.SensorEvent_LCR_Low_Deassert {
		out = append(out, SensorEvent_LCR_Low_Deassert)
	}
	if flag.SensorEvent_LNC_High_Deassert {
		out = append(out, SensorEvent_LNC_High_Deassert)
	}
	if flag.SensorEvent_LNC_Low_Deassert {
		out = append(out, SensorEvent_LNC_Low_Deassert)
	}
	if flag.SensorEvent_State_7_Deassert {
		out = append(out, SensorEvent_State_7_Deassert)
	}
	if flag.SensorEvent_State_6_Deassert {
		out = append(out, SensorEvent_State_6_Deassert)
	}
	if flag.SensorEvent_State_5_Deassert {
		out = append(out, SensorEvent_State_5_Deassert)
	}
	if flag.SensorEvent_State_4_Deassert {
		out = append(out, SensorEvent_State_4_Deassert)
	}
	if flag.SensorEvent_State_3_Deassert {
		out = append(out, SensorEvent_State_3_Deassert)
	}
	if flag.SensorEvent_State_2_Deassert {
		out = append(out, SensorEvent_State_2_Deassert)
	}
	if flag.SensorEvent_State_1_Deassert {
		out = append(out, SensorEvent_State_1_Deassert)
	}
	if flag.SensorEvent_State_0_Deassert {
		out = append(out, SensorEvent_State_0_Deassert)
	}
	if flag.SensorEvent_UNR_High_Deassert {
		out = append(out, SensorEvent_UNR_High_Deassert)
	}
	if flag.SensorEvent_UNR_Low_Deassert {
		out = append(out, SensorEvent_UNR_Low_Deassert)
	}
	if flag.SensorEvent_UCR_High_Deassert {
		out = append(out, SensorEvent_UCR_High_Deassert)
	}
	if flag.SensorEvent_UCR_Low_Deassert {
		out = append(out, SensorEvent_UCR_Low_Deassert)
	}
	if flag.SensorEvent_State_14_Deassert {
		out = append(out, SensorEvent_State_14_Deassert)
	}
	if flag.SensorEvent_State_13_Deassert {
		out = append(out, SensorEvent_State_13_Deassert)
	}
	if flag.SensorEvent_State_12_Deassert {
		out = append(out, SensorEvent_State_12_Deassert)
	}
	if flag.SensorEvent_State_11_Deassert {
		out = append(out, SensorEvent_State_11_Deassert)
	}
	if flag.SensorEvent_State_10_Deassert {
		out = append(out, SensorEvent_State_10_Deassert)
	}
	if flag.SensorEvent_State_9_Deassert {
		out = append(out, SensorEvent_State_9_Deassert)
	}
	if flag.SensorEvent_State_8_Deassert {
		out = append(out, SensorEvent_State_8_Deassert)
	}
	return out
}
