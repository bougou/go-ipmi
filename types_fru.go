package ipmi

import "time"

type FRU struct {
	CommandHeader   *FRUCommonHeader
	InternalUseArea *FRUInternalUseArea
	ChassisInfoArea *FRUChassisInfoArea
	BoardInfoArea   *FRUBoardInfoArea
	ProductInfoArea *FRUProductInfoArea
	MultiRecords    []FRUMultiRecord
}

type FRUCommonHeader struct {
	FormatVersion             uint8
	InternalUseAreaOffset     uint8
	ChassisInfoAreaOffset     uint8
	BoardInfoAreaOffset       uint8
	ProductInfoAreaOffset     uint8
	MultiRecordInfoAreaOffset uint8
	Checksum                  uint8
}

func (s FRUCommonHeader) Pack() []byte {
	out := make([]byte, 0)
	packUint8(s.FormatVersion, out, 0)
	packUint8(s.InternalUseAreaOffset, out, 1)
	packUint8(s.ChassisInfoAreaOffset, out, 2)
	packUint8(s.BoardInfoAreaOffset, out, 3)
	packUint8(s.ProductInfoAreaOffset, out, 4)
	packUint8(s.MultiRecordInfoAreaOffset, out, 5)
	packUint8(s.Checksum, out, 7)

	return out
}

type FRUInternalUseArea struct {
	FormatVersion uint8
	Data          []byte
}

type FRUChassisInfoArea struct {
	FormatVersion               uint8
	Length                      uint8
	ChassisType                 ChassisType
	ChassisPartNumberTypeLength uint8
	ChassisPartNumber           []byte
	Custom                      uint8
	Checksum                    uint8
}

type ChassisType uint8

// SMBIOS Specification: Table 17 - System Enclosure or Chassis Types
var ChassisTypeMaps = map[uint8]string{
	0x00: "",

	0x01: "Other",
	0x02: "Unknown",
	0x03: "Desktop",
	0x04: "Low Profile Desktop",
	0x05: "Pizza Box",
	0x06: "Mini Tower",
	0x07: "Tower",
	0x08: "Portable",
	0x09: "Laptop",
	0x0a: "Notebook",
	0x0b: "Hand Held",
	0x0c: "Docking Station",
	0x0d: "All in One",
	0x0e: "Sub Notebook",
	0x0f: "Space-saving",
	0x10: "Lunch Box",
	0x11: "Main Server Chassis",
	0x12: "Expansion Chassis",
	0x13: "SubChassis",
	0x14: "Bus Expansion Chassis",
	0x15: "Peripheral Chassis",
	0x16: "RAID Chassis",
	0x17: "Rack Mount Chassis",
	0x18: "Sealed-case PC",
	0x19: "Multi-system chassis",
	0x1a: "Compact PCI",
	0x1b: "Advanced TCA",
	0x1c: "Blade",
	0x1d: "Blade Enclosure",
	0x1e: "Tablet",
	0x1f: "Convertible",
	0x20: "Detachable",
	0x21: "IoT Gateway",
	0x22: "Embedded PC",
	0x23: "Mini PC",
	0x24: "Stick PC",
}

type ChassisState uint8

// SMBIOS Specification: Table 18 - System Enclosure or Chassis States
var ChassisStateMap = map[uint8]string{
	0x01: "Other",
	0x02: "Unknown",
	0x03: "Safe",
	0x04: "Warning",
	0x05: "Critical",
	0x06: "Non-recoverable",
}

type ChassisSecurityStatus uint8

// SMBIOS Specification: // Table 19 - System Enclosure or Chassis Security Status field
var ChassisSecurityStatusMap = map[uint8]string{
	0x01: "Other",
	0x02: "Unknown",
	0x03: "None",
	0x04: "External interface locked out",
	0x05: "External interface enabled",
}

type FRUBoardInfoArea struct {
	FormatVersion               uint8
	Length                      uint8
	LanguageCode                uint8
	MfgDateTime                 time.Time
	BoardManufacturerTypeLength uint8
	BoardManufacturer           []byte
	BoardProductNameTypeLength  uint8
	BoardProductName            []byte
	BoardSerialNumberTypeLength uint8
	BoardSerialNumber           []byte
	BoardPartNumberTypeLength   uint8
	BoardPartNumber             []byte
	FRUFileIDTypeLength         uint8
	FRUFileID                   []byte
	Custom                      uint8
	Checksum                    uint8
}

type BoardType uint8

var BoardTypeMap = map[uint8]string{
	0x01: "Unknown",
	0x02: "Other",
	0x03: "Server Blade",
	0x04: "Connectivity Switch",
	0x05: "System Management Module",
	0x06: "Processor Module",
	0x07: "I/O Module",
	0x08: "Memory Module",
	0x09: "Daughter board",
	0x0a: "Motherboard",
	0x0b: "Processor/Memory Module",
	0x0c: "Processor/IO Module",
	0x0d: "Interconnect board",
}

type FRUProductInfoArea struct {
	FormatVersion                 uint8
	Length                        uint8
	LanguageCode                  uint8
	MfgDateTime                   time.Time
	ManufacturerTypeLength        uint8
	Manufacturer                  []byte
	ProductNameTypeLength         uint8
	ProductName                   []byte
	ProductPartModelTypeLength    uint8
	ProductPartModel              []byte
	ProductVersionTypeLength      uint8
	ProductVersion                []byte
	ProductSerialNumberTypeLength uint8
	ProductSerialNumber           []byte
	AssetTag                      uint8
	AssertTag                     []byte
	FRUFileIDTypeLength           uint8
	FRUFileID                     []byte
	Custom                        uint8
	Checksum                      uint8
}

type FRUMultiRecord struct {
	RecordTypeID MultiRecordType

	EndOfList           bool
	RecordFormatVersion uint8

	RecordLength   uint8
	RecordChecksum uint8
	HeaderChecksum uint8

	RecordData []byte
}

type MultiRecordType uint8

var MultiRecordTypeMap = map[uint8]string{
	0x00: "Power Supply",
	0x01: "DC Output",
	0x02: "DC Load",
	0x03: "Management Access",
	0x04: "Base Compatibility",
	0x05: "Extended Compatibility",
	0x06: "ASF Fixed SMBus Device",
	0x07: "ASF Legacy-Device Alerts",
	0x08: "ASF Remote Control",
	0x09: "Extended DC Output",
	0x0a: "Extended DC Load",
}

type RecordTypePowerSupply struct {
	// This field allows for Power Supplies with capacities from 0 to 4095 watts.
	OverallCapacity uint16
	// The highest instantaneous VA value that this supply draws during operation (other than during Inrush). In integer units. FFFFh if not specified.
	PeakVA uint16
	// Maximum inrush of current, in Amps, into the power supply. FFh if not specified.
	InrushCurrent uint8 // 涌入电流
	// Number of milliseconds before power supply loading enters non-startup operating range. Set to 0 if no inrush current specified.
	InrushIntervalMilliSecond uint8
	// This specifies the low end of acceptable voltage into the power supply. The units are 10mV.
	LowEndInputVoltageRange1 uint16
	// This specifies the high end of acceptable voltage into the power supply. The units are 10mV.
	HighEndInputVoltageRange1 uint16
	// This specifies the low end of acceptable voltage into the power supply. This field would be used if the power supply did not support autoswitch. Range 1 would define the 110V range, while range 2 would be used for 220V. The units are 10mV.
	LowEndInputVoltageRange2 uint16
	// This specifies the high end of acceptable voltage into the power supply. This field would be used if the power supply did not support autoswitch. Range 1 would define the 110V range, while range 2 would be used for 220V. The units are 10mV.
	HighEndInputVoltageRange2 uint16
	// This specifies the low end of acceptable frequency range into the power supply. Use 00h if supply accepts a DC input.
	LowEndInputFrequencyRange uint8
	// This specifies the high end of acceptable frequency range into the power supply. Use 00h for both Low End and High End frequency range if supply only takes a DC input.
	HighEndInputFrequencyRange uint8
	// Minimum number of milliseconds the power supply can hold up POWERGOOD (and maintain valid DC output) after input power is lost.
	InputDropoutToleranceMilliSecond uint8

	TachometerPulses bool

	HotSwapSuppot         bool
	Autoswitch            bool
	PowerFactorCorrection bool
	PredictiveFailSupport bool

	// the number of seconds peak wattage can be sustained (0-15 seconds)
	PeakWattageHoldupSecond uint8
	// the peak wattage the power supply can produce during this time period
	PeakCapacity uint16

	CombinedWattageVoltage1 uint8 // bit 7:4 - Voltage 1
	CombinedWattageVoltage2 uint8 // bit 3:0 - Voltage 2
	// 0000b (0) 12V
	// 0001b (1) -12V
	// 0010b (2) 5V
	// 0011b (3) 3.3V

	TotalCombinedWattage uint16

	// This field serves two purposes.
	// It clarifies what type of predictive fail the power supply supports
	// (pass/fail signal or the tachometer output of the power supply fan)
	// and indicates the predictive failing point for tach outputs.
	// This field should be written as zero and ignored if the
	// predictive failure pin of the power supply is not supported.
	//
	// 	0x00 Predictive fail pin indicates pass/fail
	//  0x01 - 0xFF Lower threshold to indicate predictive failure (Rotations per second)
	PredictiveFailTachometerLowerThreshold uint8 // RPS
}

// FRU: 18.2 DC Output (Record Type 0x01)
type RecordTypeDCOutput struct {
}

// FRU: 18.2a Extended DC Output (Record Type 0x09)
type RecordTypeExtenedDCOutput struct {
}

// FRU: 18.3 DC Load (Record Type 0x02)
type RecordTypeDCLoad struct {
}

// FRU: 18.3a Extended DC Load (Record Type 0x0A)
type RecordTypeExtendedDCLoad struct {
}

// FRU: 18.4 Management Access Record (Record Type 0x03)
type RecordTypeManagementAccess struct {
}

// FRU: 18.5 Base Compatibility Record (Record Type 0x04)
type RecordTypeBaseCompatibility struct {
}

// FRU: 18.6 Extended Compatibility Record (Record Type 0x05)
type RecordTypeExtendedCompatiblityRecord struct {
}

// FRU: 18.7 OEM Record (Record Types 0xC0-0xFF)
type RecordTypeOEM struct {
}
