package ipmi

import "fmt"

type SensorThreshold struct {
	LowerNonCritical    int
	LowerCritical       int
	LowerNonRecoverable int
	UpperNonCritical    int
	UpperCritical       int
	UpperNonRecoverable int
}

type SensorClass string

// Sensors are classified according to the type of readings the provide and/or
// the type of events they generate.
// A sensor can return either an analog or discrete readings. Sensor events can
// be discrete or threshold-based.

const (
	SensorClassNotApplicable   = SensorClass("n/a") // 不适用的
	SensorClassThreshold       = SensorClass("threshold")
	SensorClassDiscrete        = SensorClass("discrete")         // multiple states possible
	SensorClassDigitalDiscrete = SensorClass("digital-discrete") // A digital sensor is not really a unique class, but a term commonly used to refer to special case of a discrete sensor that only has two possible states
	SensorClassOEM             = SensorClass("oem")              // Special case of discrete where the meaning of the states (offsets) are OEM defined.
)

type SensorGeneratorID uint16

const (
	SensorGeneratorBMC              = SensorGeneratorID(0x0020)
	SensorGeneratorBIOSPOST         = SensorGeneratorID(0x0001)
	SensorGeneratorBISOSMIHandler   = SensorGeneratorID(0x0033)
	SensorGeneratorIntelNMFirmware  = SensorGeneratorID(0x002C) // Node Manager
	SensorGeneratorIntelMEFirmware  = SensorGeneratorID(0x602C) // Management Engine
	SensorGeneratorMicrosoftOS      = SensorGeneratorID(0x0041)
	SensorGeneratorLinuxKernelPanic = SensorGeneratorID(0x0021)
)

type SensorOwnerID struct {
	SlaveAddress uint8
	SoftwareID   uint8
	Flag         bool // Flag 0 means SlaveAddress, flag 1 means SoftwareID
}

// 43.14 Entity IDs
// The Entity ID field is used for identifying the physical entity that a sensor or device is associated with.
// If multiple sensors refer to the same entity, they will have the same Entity ID field value.
// For example, if a voltage sensor and a temperature sensor are both for a "Power Supply 1" entity,
// the Entity ID in their sensor data records would both be 10 (0Ah), per the Entity ID table.
type SensorEntityID uint8

func (s SensorEntityID) String() string {
	out, ok := sensorEntityIDMap[s]
	if ok {
		return out
	}
	if s >= 0x80 && s <= 0xaf {
		// These IDs are system specific and can be assigned by the chassis provider.
		return fmt.Sprintf("Chassis-specific Entities (#%#02x)", s)
	}
	if s >= 0xb0 && s <= 0xcf {
		// These IDs are system specific and can be assigned by the Board-set provider
		return fmt.Sprintf("Board-set specific Entities (#%#02x)", s)
	}
	if s >= 0xd0 && s <= 0xff {
		// These IDs are system specific and can be assigned by the system integrator, or OEM.
		return fmt.Sprintf("OEM System Integrator defined (#%#02x)", s)
	}
	return fmt.Sprintf("reserved (#%#02x)", s)
}

// section 43.14
var sensorEntityIDMap = map[SensorEntityID]string{
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
	0x0A: "power supply",                    // DMI refers to this as a 'power unit', but it's used to represent a power supply
	0x0B: "add-in card",                     // 附加卡
	0x0C: "front panel board",               // 前面板
	0x0D: "back panel board",                // 背板
	0x0E: "power system board",              // 电源系统板
	0x0F: "drive backplane",                 // 驱动器背板
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
	0x1A: "Disk Drive bay",
	0x1B: "Peripheral bay",
	0x1C: "Device bay",
	0x1D: "fan / cooling device",
	0x1E: "cooling unit / cooling domain",
	0x1F: "cable / interconnect",
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
	0x2A: "Connectivity switch",
	0x2B: "Processor/memory module",
	0x2C: "I/O module",
	0x2D: "Processor / IO module",
	0x2E: "Management Controller Firmware",
	0x2F: "IPMI Channel",
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

// the "owner" of the sensor.

// The combination of Sensor Owner ID and Sensor Number uniquely identify a sensor in the system.

// the Sensor Data Record and SEL information must contain information to identify the "owner" of the sensor.
// For management controllers, a Slave Address and LUN identify the owner of a sensor on the IPMB.
// For system software, a Software ID identifies the owner of a sensor.
// These fields are used in Event Messages, where events from management controllers or the IPMB are identified by an eight-bit field where the upper 7-bits are the Slave Address or System Software ID.
// The least significant bit is a 0 if the value represents a Slave Address and a 1 if the value represents a System Software ID.

// see: Intel System Event Log (SEL) Troubleshooting Guide Rev 3.4 September 2019 section 3.1
type SensorNumber uint8

var SensorNumberNameMap = map[SensorNumber]string{
	0x01: "Power Unit Status",
	0x02: "Power Unit Redundancy",
	0x03: "IPMI Watchdog",
	0x04: "Physical Security",
	0x05: "FP Interrupt",
	0x06: "SMI Timeout",
	0x07: "System Event Log",
	0x08: "System Event",
	0x09: "Button Sensor",
	0x0A: "BMC Watchdog",
	0x0B: "Voltage Regulator Watchdog",
	0x0C: "Fan Redundancy",
	0x0D: "SSB Thermal Trip",
	0x0E: "IO Module Presence",
	0x0F: "SAS Module Presense",
	0x10: "BMC Fireware Health",
	0x11: "System Airflow",
	0x12: "Firmware Update Status",
	0x13: "IO Module2 Presence",
	0x14: "Baseboard Temperature 5",
	0x15: "Baseboard Temperature 6",
	0x16: "IO Module2 Temperature",
	0x17: "PCI Riser 3 Temperature",
	0x18: "PCI Riser 4 Temperature",
	0x19: "",
	0x1A: "Fireware Security",
	0x20: "Baseboard Temperature 1",
	0x21: "Front Panel Temperature",
	0x22: "SSB Tempertature",
	0x23: "Baseboard Temperature 2",
	0x24: "Baseboard Temperature 3",
	0x25: "Baseboard Temperature 4",
	0x26: "IO Module Temperature",
	0x27: "PCI Riser 1 Temperature",
	0x28: "IO Riser Temperature",
	0x29: "Hot-Swap Back Plane 1 Temperature",
	0x2A: "Hot-Swap Back Plane 2 Temperature",
	0x2B: "Hot-Swap Back Plane 3 Temperature",
	0x2C: "PCI Riser 2 Temperature",
	0x2D: "SAS Module Temperature",
	0x2E: "Exit Air Temperature",
	0x2F: "Network Interface Controller Temperature",
}

// 41.1 Sensor Type Code
type SensorType uint8

func (c SensorType) String() string {
	return sensorTypeMap[c]
}

const (
	SensorTypeReserved                     SensorType = 0x00
	SensorTypeTemperature                  SensorType = 0x01 // 温度传感器
	SensorTypeVoltage                      SensorType = 0x02 // 电压传感器
	SensorTypeCurrent                      SensorType = 0x03 // 电流传感器
	SensorTypeFan                          SensorType = 0x04 // 风扇传感器
	SensorTypePhysicalSecurity             SensorType = 0x05 // Chassis Intrusion
	SensorTypePlatformSecurity             SensorType = 0x06
	SensorTypeProcessor                    SensorType = 0x07
	SensorTypePowserSupply                 SensorType = 0x08
	SensorTypePowerUnit                    SensorType = 0x09
	SensorTypeCollingDevice                SensorType = 0x0A
	SensorTypeOtherUnitsbased              SensorType = 0x0B
	SensorTypeMemory                       SensorType = 0x0C
	SensorTypeDriveSlot                    SensorType = 0x0D
	SensorTypePostMemoryResize             SensorType = 0x0E
	SensorTypeSystemFirmware               SensorType = 0x0F
	SensorTypeEventLoggingDisabled         SensorType = 0x10
	SensorTypeWatchdog1                    SensorType = 0x11
	SensorTypeSystemEvent                  SensorType = 0x12
	SensorTypeCriticalInterrupt            SensorType = 0x13
	SensorTypeButtonSwitch                 SensorType = 0x14
	SensorTypeModuleBoard                  SensorType = 0x15
	SensorTypeMicrocontrollerCoprocessor   SensorType = 0x16
	SensorTypeAddinCard                    SensorType = 0x17
	SensorTypeChassis                      SensorType = 0x18
	SensorTypeChipSet                      SensorType = 0x19
	SensorTypeOtherFRU                     SensorType = 0x1A
	SensorTypeCableInterconnect            SensorType = 0x1B
	SensorTypeTerminator                   SensorType = 0x1C
	SensorTypeSystemBootRestartInitiated   SensorType = 0x1D
	SensorTypeBootError                    SensorType = 0x1E
	SensorTypeBaseOSBootInstallationStatus SensorType = 0x1F
	SensorTypeOSStopShutdown               SensorType = 0x20
	SensorTypeSlotConnector                SensorType = 0x21
	SensorTypeSystemACPIPowerState         SensorType = 0x22
	SensorTypeWatchdog2                    SensorType = 0x23
	SensorTypePlatormAlert                 SensorType = 0x24
	SensorTypeEntityPresence               SensorType = 0x25
	SensorTypeMonitorASIC                  SensorType = 0x26
	SensorTypeLAN                          SensorType = 0x27
	SensorTypeManagementSubsystemHealth    SensorType = 0x28
	SensorTypeBattery                      SensorType = 0x29
	SensorTypeSessionAudit                 SensorType = 0x2A
	SensorTypeVersionChange                SensorType = 0x2B
	SensorTypeFRUState                     SensorType = 0x2C

	// Reserverd: 0x2D - 0xBF
	// OEM Reserved: 0xC0 - 0xFF
)

var sensorTypeMap = map[SensorType]string{
	SensorTypeReserved:                     "Reserved",
	SensorTypeTemperature:                  "Temperature",
	SensorTypeVoltage:                      "Voltage",
	SensorTypeCurrent:                      "Current",
	SensorTypeFan:                          "Fan",
	SensorTypePhysicalSecurity:             "Physical Security",
	SensorTypePlatformSecurity:             "Platform Security",
	SensorTypeProcessor:                    "Processor",
	SensorTypePowserSupply:                 "Power Supply",
	SensorTypePowerUnit:                    "Power Unit",
	SensorTypeCollingDevice:                "Cooling Device",
	SensorTypeOtherUnitsbased:              "Other",
	SensorTypeMemory:                       "Memory",
	SensorTypeDriveSlot:                    "Drive Slot / Bay",
	SensorTypePostMemoryResize:             "POST Memory Resize",
	SensorTypeSystemFirmware:               "System Firmwares",
	SensorTypeEventLoggingDisabled:         "Event Logging Disabled",
	SensorTypeWatchdog1:                    "Watchdog1",
	SensorTypeSystemEvent:                  "System Event",
	SensorTypeCriticalInterrupt:            "Critical Interrupt",
	SensorTypeButtonSwitch:                 "Button",
	SensorTypeModuleBoard:                  "Module / Board",
	SensorTypeMicrocontrollerCoprocessor:   "Microcontroller/Coprocessor",
	SensorTypeAddinCard:                    "Add-in Card",
	SensorTypeChassis:                      "Chassis",
	SensorTypeChipSet:                      "Chip Set",
	SensorTypeOtherFRU:                     "Other FRU",
	SensorTypeCableInterconnect:            "Cable / Interconnect",
	SensorTypeTerminator:                   "Terminator",
	SensorTypeSystemBootRestartInitiated:   "System Boot Initiated",
	SensorTypeBootError:                    "Boot Error",
	SensorTypeBaseOSBootInstallationStatus: "OS Boot",
	SensorTypeOSStopShutdown:               "OS Critical Stop",
	SensorTypeSlotConnector:                "Slot / Connector",
	SensorTypeSystemACPIPowerState:         "System ACPI Power State",
	SensorTypeWatchdog2:                    "Watchdog2",
	SensorTypePlatormAlert:                 "Platform Alert",
	SensorTypeEntityPresence:               "Entity Presence",
	SensorTypeMonitorASIC:                  "Monitor ASIC",
	SensorTypeLAN:                          "LAN",
	SensorTypeManagementSubsystemHealth:    "Management Subsys Health",
	SensorTypeBattery:                      "Battery",
	SensorTypeSessionAudit:                 "Session Audit",
	SensorTypeVersionChange:                "Version Change",
	SensorTypeFRUState:                     "FRU State",
}

type SensorUnitTypeCode uint8

// section 43.17
const (
	SensorUnitUnspecified        = SensorUnitTypeCode(0)  // unspecified
	SensorUnitDegressC           = SensorUnitTypeCode(1)  // degrees C, Celsius, 摄氏度 ℃
	SensorUnitDegreesF           = SensorUnitTypeCode(2)  // degrees F，Fahrenheit, 华氏度
	SensorUnitDegreesK           = SensorUnitTypeCode(3)  // degrees K，Kelvins, 开尔文
	SensorUnitVolts              = SensorUnitTypeCode(4)  // Volts, 伏特（电压单位）
	SensorUnitAmps               = SensorUnitTypeCode(5)  // Amps, 安培数
	SensorUnitWatts              = SensorUnitTypeCode(6)  // Watts, 瓦特（功率单位）
	SensorUnitJoules             = SensorUnitTypeCode(7)  // Joules, 焦耳
	SensorUnitCoulombs           = SensorUnitTypeCode(8)  // Coulombs, 库伦
	SensorUnitVA                 = SensorUnitTypeCode(9)  // VA, 伏安
	SensorUnitNits               = SensorUnitTypeCode(10) // Nits, 尼特（光度单位）
	SensorUnitLumen              = SensorUnitTypeCode(11) // lumen, 流明（光通量单位）
	SensorUnitLux                = SensorUnitTypeCode(12) // lux, 勒克斯（照明单位）
	SensorUnitCandela            = SensorUnitTypeCode(13) // Candela, 坎，坎德拉（发光强度单位）
	SensorUnitKPa                = SensorUnitTypeCode(14) // kPa kilopascal, 千帕, 千帕斯卡
	SensorUnitPSI                = SensorUnitTypeCode(15) // PSI
	SensorUnitNewton             = SensorUnitTypeCode(16) // Newton, 牛顿（力的单位）
	SensorUnitCFM                = SensorUnitTypeCode(17) // CFM, 风量, cubic feet per minute (cu ft/min)
	SensorUnitRPM                = SensorUnitTypeCode(18) // RPM, 每分钟转数，Revolutions per minute, is the number of turns in one minute
	SensorUnitHz                 = SensorUnitTypeCode(19) // Hz, 赫兹
	SensorUnitMicroSecond        = SensorUnitTypeCode(20) // microsecond， 微秒
	SensorUnitMilliSecond        = SensorUnitTypeCode(21) // millisecond， 毫秒
	SensorUnitSecond             = SensorUnitTypeCode(22) // second，秒
	SensorUnitMinute             = SensorUnitTypeCode(23) // minute， 分
	SensorUnitHour               = SensorUnitTypeCode(24) // hour，时
	SensorUnitDay                = SensorUnitTypeCode(25) // day，日
	SensorUnitWeek               = SensorUnitTypeCode(26) // week，周
	SensorUnitMil                = SensorUnitTypeCode(27) // mil, 毫升；密耳（千分之一寸）
	SensorUnitInches             = SensorUnitTypeCode(28) // inches, 英寸（inch的复数）
	SensorUnitFleet              = SensorUnitTypeCode(29) // feet
	SensorUnitCuIn               = SensorUnitTypeCode(30) // cu in, 立方英寸（cubic inch）
	SensorUnitCuFleet            = SensorUnitTypeCode(31) // cu feet
	SensorUnitMM                 = SensorUnitTypeCode(32) // mm, 毫米（millimeter）
	SensorUnitCM                 = SensorUnitTypeCode(33) // cm, 厘米（centimeter）
	SensorUnitM                  = SensorUnitTypeCode(34) // m, 米
	SensorUnitCuCM               = SensorUnitTypeCode(35) // cu cm
	SensorUnitCum                = SensorUnitTypeCode(36) // cum
	SensorUnitLiters             = SensorUnitTypeCode(37) // liters, 公升（容量单位）
	SensorUnitFluidOunce         = SensorUnitTypeCode(38) // fluid ounce, 液盎司（液体容量单位，等于 fluidounce）
	SensorUnitRadians            = SensorUnitTypeCode(39) // radians, 弧度（radian的复数）
	SensorUnitvSteradians        = SensorUnitTypeCode(40) // steradians, 球面度，立体弧度（立体角国际单位制，等于 sterad）
	SensorUnitRevolutions        = SensorUnitTypeCode(41) // revolutions, 转数（revolution的复数形式）
	SensorUnitCycles             = SensorUnitTypeCode(42) // cycles， 周期，圈
	SensorUnitGravities          = SensorUnitTypeCode(43) // gravities， 重力
	SensorUnitOunce              = SensorUnitTypeCode(44) // ounce， 盎司
	SensorUnitPound              = SensorUnitTypeCode(45) // pound, 英镑
	SensorUnitFootPound          = SensorUnitTypeCode(46) // ft-lb， 英尺-磅（foot pound）
	SensorUnitOzIn               = SensorUnitTypeCode(47) // oz-in， 扭力；盎司英寸
	SensorUnitGauss              = SensorUnitTypeCode(48) // gauss， 高斯（磁感应或磁场的单位）
	SensorUnitGilberts           = SensorUnitTypeCode(49) // gilberts， 吉伯（磁通量的单位）
	SensorUnitHenry              = SensorUnitTypeCode(50) // henry, 亨利（电感单位）
	SensorUnitMilliHenry         = SensorUnitTypeCode(51) // millihenry, 毫亨（利）（电感单位）
	SensorUnitFarad              = SensorUnitTypeCode(52) // farad, 法拉（电容单位）
	SensorUnitMicroFarad         = SensorUnitTypeCode(53) // microfarad, 微法拉（电容量的实用单位）
	SensorUnitOhms               = SensorUnitTypeCode(54) // ohms， 欧姆（Ohm） ：电阻的量度单位，欧姆值越大，电阻越大
	SensorUnitSiemens            = SensorUnitTypeCode(55) // siemens， 西门子， 电导单位
	SensorUnitMole               = SensorUnitTypeCode(56) // mole， 摩尔 [化学] 克分子（等于mole）
	SensorUnitBecquerel          = SensorUnitTypeCode(57) // becquerel， 贝可（放射性活度单位）
	SensorUnitPPM                = SensorUnitTypeCode(58) // PPM (parts/million)， 百万分率，百万分之…（parts per million）
	SensorUnitReserved           = SensorUnitTypeCode(59) // reserved
	SensorUnitDecibels           = SensorUnitTypeCode(60) // Decibels，分贝（声音强度单位，decibel的复数）
	SensorUnitDbA                = SensorUnitTypeCode(61) // DbA, dBA is often used to specify the loudness of the fan used to cool the microprocessor and associated components. Typical dBA ratings are in the neighborhood of 25 dBA, representing 25 A-weighted decibels above the threshold of hearing. This is approximately the loudness of a person whispering in a quiet room.
	SensorUnitDbC                = SensorUnitTypeCode(62) // DbC
	SensorUnitGray               = SensorUnitTypeCode(63) // gray，核吸收剂量(Gy)
	SensorUnitSevert             = SensorUnitTypeCode(64) // sievert, 希沃特（辐射效果单位，简称希）
	SensorUnitColorTempDegK      = SensorUnitTypeCode(65) // color temp deg K, 色温
	SensorUnitBit                = SensorUnitTypeCode(66) // bit, 比特（二进位制信息单位）
	SensorUnitKilobit            = SensorUnitTypeCode(67) // kilobit, 千比特
	SensorUnitMegabit            = SensorUnitTypeCode(68) // megabit, 兆比特
	SensorUnitGigabit            = SensorUnitTypeCode(69) // gigabit，吉比特
	SensorUnitByte               = SensorUnitTypeCode(70) // byte， 字节
	SensorUnitKilobyte           = SensorUnitTypeCode(71) // kilobyte，千字节
	SensorUnitMegabyte           = SensorUnitTypeCode(72) // megabyte，兆字节
	SensorUnitGigabyte           = SensorUnitTypeCode(73) // gigabyte，吉字节
	SensorUnitWord               = SensorUnitTypeCode(74) // word (data)，字
	SensorUnitDWord              = SensorUnitTypeCode(75) // dword， 双字
	SensorUnitQWord              = SensorUnitTypeCode(76) // qword， 四字
	SensorUnitLine               = SensorUnitTypeCode(77) // line (re. mem. line)
	SensorUnitHit                = SensorUnitTypeCode(78) // hit, 命中
	SensorUnitMiss               = SensorUnitTypeCode(79) // miss， 未击中， 未命中
	SensorUnitRetry              = SensorUnitTypeCode(80) // retry, 重试（次数）
	SensorUnitReset              = SensorUnitTypeCode(81) // reset，重置（次数）
	SensorUnitOverrun            = SensorUnitTypeCode(82) // overrun) / overflow 满载，溢出（次数）
	SensorUnitUnderrun           = SensorUnitTypeCode(83) // underrun 欠载
	SensorUnitCollision          = SensorUnitTypeCode(84) // collision, 冲突
	SensorUnitPacket             = SensorUnitTypeCode(85) // packets, 包, 数据包
	SensorUnitMessage            = SensorUnitTypeCode(86) // messages, 消息
	SensorUnitCharacters         = SensorUnitTypeCode(87) // characters，字符
	SensorUnitError              = SensorUnitTypeCode(88) // error， 错误
	SensorUnitCorrectableError   = SensorUnitTypeCode(89) // correctable error 可校正错误
	SensorUnitUncorrectableError = SensorUnitTypeCode(90) // uncorrectable error 不可校正错误
	SensorUnitFatalError         = SensorUnitTypeCode(91) // fatal error， 致命错误，不可恢复的错误
	SensorUnitGrams              = SensorUnitTypeCode(92) // grams, 克（gram的复数形式）
)
