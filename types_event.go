package ipmi

type EventRecordType uint8

const (
	EventRecordTypeSystemEvent EventRecordType = 0x02

	// 32.2 OEM SEL Record - Type C0h-DFh
	// C0h-DFh = OEM timestamped

	// 32.3 OEM SEL Record - Type E0h-FFh
	// E0h-FFh = OEM non-timestamped
)

func (t EventRecordType) String() string {
	if t == EventRecordTypeSystemEvent {
		return "system event"
	}
	if t >= 0xc0 && t <= 0xdf {
		return "OEM timestamped"
	}
	if t >= 0xe0 && t <= 0xff {
		return "OEM non-timestamped"
	}
	return "unknown"
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

// 41.2 Event/Reading Type Code
type EventType uint8

const (
	EventTypeUnspecified    EventType = 0x00
	EventTypeThreshold      EventType = 0x01
	EventTypeGenericLow     EventType = 0x02
	EventTypeGenericHigh    EventType = 0x0C
	EventTypeSensorSpecific EventType = 0x6F
	EventTypeOEMLow         EventType = 0x70
	EventTypeOEMHigh        EventType = 0x7F
)

func (eventType EventType) Category() EventTypeCategory {
	var c EventTypeCategory

	switch eventType {
	case EventTypeUnspecified:
		c = EventTypeCategoryUnspecified
	case EventTypeThreshold:
		c = EventTypeCategoryThreshold
	case EventTypeSensorSpecific:
		c = EventTypeCategorySensorSpecific
	default:
		if eventType >= EventTypeGenericLow && eventType <= EventTypeGenericHigh {
			c = EventTypeCategoryGeneric
		} else if eventType >= EventTypeOEMLow && eventType <= EventTypeOEMHigh {
			c = EventTypeCategoryOEM
		} else {
			c = EventTypeCategoryReserved
		}
	}

	return c
}

func (eventType EventType) SensorClass() SensorClass {
	var c SensorClass

	switch eventType {
	case EventTypeUnspecified:
		c = SensorClassNotApplicable
	case EventTypeThreshold:
		c = SensorClassThreshold
	case EventTypeSensorSpecific:
		c = SensorClassDiscrete
	default:
		if eventType >= EventTypeGenericLow && eventType <= EventTypeGenericHigh {
			c = SensorClassDiscrete
		} else if eventType >= EventTypeOEMLow && eventType <= EventTypeOEMHigh {
			c = SensorClassOEM
		} else {
			c = SensorClassNotApplicable
		}
	}

	return c
}

type EventTypeCategory string

const (
	EventTypeCategoryUnspecified    EventTypeCategory = "Unspecified"
	EventTypeCategoryThreshold      EventTypeCategory = "Threshold" // Discrete
	EventTypeCategoryGeneric        EventTypeCategory = "Generic"
	EventTypeCategorySensorSpecific EventTypeCategory = "Sensor Specific"
	EventTypeCategoryOEM            EventTypeCategory = "OEM"
	EventTypeCategoryReserved       EventTypeCategory = "Reserved"
)
