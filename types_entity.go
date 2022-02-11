package ipmi

import "fmt"

// 43.14 Entity IDs
// 39. Using Entity IDs
// EntityID can be seen as Entity Type
//
// 1. An Entity ID is a standardized numeric code that is used in SDRs to identify
// the types of physical entities or FRUs in the system.
//
// 2. The Entity ID is associated with an Entity Instance value that is used to
// indicate the particular instance of an entity
//
// 3. The SDR for a sensor includes Entity ID and Entity Instance fields that
// identify the entity associated with the sensor.
type EntityID uint8

func (e EntityID) String() string {
	// 43.14 Entity IDs
	var entityIDMap = map[EntityID]string{
		0x00: "unspecified",
		0x01: "other",
		0x02: "unspecified",
		0x03: "processor",
		0x04: "disk or disk bay", // 磁盘托架
		0x05: "peripheral bay",   // 外围托架
		0x06: "system management module",
		0x07: "system board",
		0x08: "memory module",
		0x09: "processor module",
		0x0a: "power supply",                    // DMI refers to this as a 'power unit', but it's used to represent a power supply
		0x0b: "add-in card",                     // 附加卡
		0x0c: "front panel board",               // 前面板
		0x0d: "back panel board",                // 背板
		0x0e: "power system board",              // 电源系统板
		0x0f: "drive backplane",                 // 驱动器背板
		0x10: "system internal expansion board", //
		0x11: "other system board",
		0x12: "processor board",
		0x13: "power unit / power domain",
		0x14: "power module / DC-to-DC converter",
		0x15: "power management / power distribution board",
		0x16: "chassis back panel board", // 机箱后面板
		0x17: "system chassis",           //
		0x18: "sub-chassis",
		0x19: "other chassis board",
		0x1a: "Disk Drive bay",
		0x1b: "Peripheral bay",
		0x1c: "Device bay",
		0x1d: "fan / cooling device",
		0x1e: "cooling unit / cooling domain",
		0x1f: "cable / interconnect",
		0x20: "memory device",
		0x21: "System Management Softeware",
		0x22: "System Firmware", // eg BIOS/EFI
		0x23: "Operating System",
		0x24: "system bus",
		0x25: "Group",
		0x26: "Remote (Out of Band) Management Communication Device",
		0x27: "External Environment",
		0x28: "battery",
		0x29: "Processing blade",
		0x2a: "Connectivity switch",
		0x2b: "Processor/memory module",
		0x2c: "I/O module",
		0x2d: "Processor / IO module",
		0x2e: "Management Controller Firmware",
		0x2f: "IPMI Channel",
		0x30: "PCI Bus",
		0x31: "PCI Express Bus",
		0x32: "SCSI Bus (parallel)",
		0x33: "SATA / SAS bus",
		0x34: "Processor / front-side bus",
		0x35: "Real Time Clock (RTC)",
		0x36: "System Firmware", // reserved. This value was previously a duplicate of 22h (System Firmware).
		0x37: "air inlet",
		0x38: "System Firmware", // reserved. This value was previously a duplicate of 22h (System Firmware).
		0x40: "air inlet",       // This Entity ID value is equivalent to Entity ID 37h. It is provided for interoperability with the DCMI 1.0 specifications.
		0x41: "processor",       // This Entity ID value is equivalent to Entity ID 03h (processor). It is provided for interoperability with the DCMI 1.0 specifications.
		0x42: "system board",    // This Entity ID value is equivalent to Entity ID 07h (system board). It is provided for interoperability with the DCMI 1.0 specifications.
	}

	out, ok := entityIDMap[e]
	if ok {
		return out
	}

	if e >= 0x90 && e <= 0xaf {
		// These IDs are system specific and can be assigned by the chassis provider.
		return fmt.Sprintf("Chassis-specific Entities (#%#02x)", uint8(e))
	}
	if e >= 0xb0 && e <= 0xcf {
		// These IDs are system specific and can be assigned by the Board-set provider
		return fmt.Sprintf("Board-set specific Entities (#%#02x)", uint8(e))
	}
	if e >= 0xd0 && e <= 0xff {
		// These IDs are system specific and can be assigned by the system integrator, or OEM.
		return fmt.Sprintf("OEM System Integrator defined (#%#02x)", uint8(e))
	}

	return fmt.Sprintf("reserved (#%#02x)", uint8(e))
}

// see: 39.1 System- and Device-relative Entity Instance Values
//
// Entity Instance values in the system-relative range are required to be unique for all entities with the same Entity ID in the system.
//
// Device-relative Entity Instance values are only required to be unique among all entities that have the same Entity ID within a given device (management controller).
//
// For example, management controller A and B could both have FAN entities that have an Entity Instance value of 60h.
//
// EntityInstance only occupy 7 bits, range is 0x00 ~ 0x7f
type EntityInstance uint8

// 39.1 System- and Device-relative Entity Instance Values
func isEntityInstanceSystemRelative(e EntityInstance) bool {
	return e >= 0x00 && e <= 0x5f
}

func isEntityInstanceDeviceRelative(e EntityInstance) bool {
	return e >= 0x60 && e <= 0x7f
}

func (e EntityInstance) Type() string {
	if isEntityInstanceSystemRelative(e) {
		return "system-relative"
	}
	if isEntityInstanceDeviceRelative(e) {
		return "device-relative"
	}
	return "'"
}

func canonicalEntityString(entityID EntityID, entityInstance EntityInstance) string {
	if isEntityInstanceSystemRelative(entityInstance) {
		return fmt.Sprintf("System, %s, %d", entityID.String(), entityInstance)
	}
	if isEntityInstanceDeviceRelative(entityInstance) {
		return fmt.Sprintf("Controller 1, %s, %d", entityID.String(), entityInstance)
	}
	return "Unkown"
}

// 43.13 Device Type Codes
// DeviceType codes are used to identify different types of devices on
// an IPMB, PCI Management Bus, or Private Management Bus connection
// to an IPMI management controller
type DeviceType uint16

func (d DeviceType) String() string {
	// IPMB/I2C Device Type Codes
	// EEPROM，或写作E2PROM，全称电子式可擦除可编程只读存储器 （英语：Electrically-Erasable Programmable Read-Only Memory），是一种可以通过电子方式多次复写的半导体存储设备。
	var deviceTypeMap = map[DeviceType]string{
		0x00: "Reserved",
		0x01: "Reserved",
		0x02: "DS1624 temperature sensor",
		0x03: "DS1621 temperature sensor",
		0x04: "LM75 Temperature Sensor",
		0x05: "Heceta ASIC",
		0x06: "Reserved",
		0x07: "Reserved",

		// modifier codes for deviceTye 0x08 ~ 0x0f
		// 00h = unspecified
		// 01h = DIMM Memory ID
		// 02h = IPMI FRU Inventory
		// 03h = System Processor Cartridge FRU / PIROM
		// (processor information ROM)
		// all other = reserved
		0x08: "EEPROM, 24C01",
		0x09: "EEPROM, 24C02",
		0x0a: "EEPROM, 24C04",
		0x0b: "EEPROM, 24C08",
		0x0c: "EEPROM, 24C16",
		0x0d: "EEPROM, 24C17",
		0x0e: "EEPROM, 24C32",
		0x0f: "EEPROM, 24C64",

		// modifier codes for deviceType 0x10
		// 00h = IPMI FRU Inventory [1]
		// 01h = DIMM Memory ID
		// 02h = IPMI FRU Inventory[1]
		// 03h = System Processor Cartridge FRU / PIROM
		// (processor information ROM)
		// all other = reserved
		// FFh = unspecified
		0x10: "FRU Inventory Device behind management controller", // (accessed using Read/Write FRU commands at LUN other than 00b)

		0x11: "Reserved",
		0x12: "Reserved",
		0x13: "Reserved",
		0x14: "PCF 8570 256 byte RAM",
		0x15: "PCF 8573 clock/calendar",
		0x16: "PCF 8574A I/O Port",
		0x17: "PCF 8583 clock/calendar",
		0x18: "PCF 8593 clock/calendar",
		0x19: "Clock calendar",
		0x1a: "PCF 8591 A/D, D/A Converter",
		0x1b: "I/O Port",
		0x1c: "A/D Converter",
		0x1d: "D/A Converter",
		0x1e: "A/D, D/A Converter",
		0x1f: "LCD Controller/Driver",
		0x20: "Core Logic (Chip set) Device",
		0x21: "LMC6874 Intelligent Battery controller",
		0x22: "Intelligent Batter controller",
		0x23: "Combo Management ASIC",
		0x24: "Maxim 1617 Temperature Sensor",
		0xbf: "Other/Unspecified",
	}

	s, ok := deviceTypeMap[d]
	if ok {
		return s
	}
	return ""
}
