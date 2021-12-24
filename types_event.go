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

// 43.13 Device Type Codes
// DeviceType codes are used to identify different types of devices on
// an IPMB, PCI Management Bus, or Private Management Bus connection
// to an IPMI management controller
type DeviceTypeWithModifier uint16

// IPMB/I2C Device Type Codes
// EEPROM，或写作E2PROM，全称电子式可擦除可编程只读存储器 （英语：Electrically-Erasable Programmable Read-Only Memory），是一种可以通过电子方式多次复写的半导体存储设备。
var deviceTypeMap = map[DeviceTypeWithModifier]string{
	0x00:   "Reserved",
	0x01:   "Reserved",
	0x02:   "DS1624 temperature sensor",
	0x03:   "DS1621 temperature sensor",
	0x04:   "LM75 Temperature Sensor",
	0x05:   "Heceta ASIC",
	0x06:   "Reserved",
	0x07:   "Reserved",
	0x08:   "EEPROM, 24C01",
	0x09:   "EEPROM, 24C02",
	0x0a:   "EEPROM, 24C04",
	0x0b:   "EEPROM, 24C08",
	0x0c:   "EEPROM, 24C16",
	0x0d:   "EEPROM, 24C17",
	0x0e:   "EEPROM, 24C32",
	0x0f:   "EEPROM, 24C64",
	0x0010: "IPMI FRU Inventory",
	0x0110: "DIMM Memory ID",
	0x0210: "IPMI FRU Inventory",
	0x0310: "System Processor Cartridge FRU",
	0x11:   "Reserved",
	0x12:   "Reserved",
	0x13:   "Reserved",
	0x14:   "PCF 8570 256 byte RAM",
	0x15:   "PCF 8573 clock/calendar",
	0x16:   "PCF 8574A I/O Port",
	0x17:   "PCF 8583 clock/calendar",
	0x18:   "PCF 8593 clock/calendar",
	0x19:   "Clock calendar",
	0x1a:   "PCF 8591 A/D, D/A Converter",
	0x1b:   "I/O Port",
	0x1c:   "A/D Converter",
	0x1d:   "D/A Converter",
	0x1e:   "A/D, D/A Converter",
	0x1f:   "LCD Controller/Driver",
	0x20:   "Core Logic (Chip set) Device",
	0x21:   "LMC6874 Intelligent Battery controller",
	0x22:   "Intelligent Batter controller",
	0x23:   "Combo Management ASIC",
	0x24:   "Maxim 1617 Temperature Sensor",
	0xbf:   "Other/Unspecified",
}
