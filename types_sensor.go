package ipmi

type SensorThreshold struct {
	LowerNonCritical    int
	LowerCritical       int
	LowerNonRecoverable int
	UpperNonCritical    int
	UpperCritical       int
	UpperNonRecoverable int
}

// 42.1
// Sensors are classified according to the type of readings the provide and/or the type of events they generate.
//
// Three sensor classes: threshold, discrete, oem
//
// A sensor can return either an analog or discrete readings. Sensor events can be discrete or threshold-based.
type SensorClass string

const (
	SensorClassNotApplicable SensorClass = "n/a" // 不适用的

	SensorClassThreshold SensorClass = "threshold"

	// 离散 multiple states possible
	// Discrete sensors can contain up to 15 possible states.
	// It is possible for a discrete sensor to have more than one state active at a time
	SensorClassDiscrete SensorClass = "discrete"

	// A digital sensor is not really a unique class, but a term commonly used to refer to
	// special case of a discrete sensor that only has two possible states
	// SensorClassDigitalDiscrete SensorClass = "digital-discrete"

	// Special case of discrete where the meaning of the states (offsets) are OEM defined.
	SensorClassOEM SensorClass = "oem"
)

// 41.1 Sensor Type Code
// 42.2 Sensor Type Codes and Data
type SensorType uint8

func (c SensorType) String() string {
	s, ok := sensorTypeMap[c]
	if ok {
		return s
	}
	return "unkown"
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
	SensorTypeCollingDevice                SensorType = 0x0a
	SensorTypeOtherUnitsbased              SensorType = 0x0b
	SensorTypeMemory                       SensorType = 0x0c
	SensorTypeDriveSlot                    SensorType = 0x0d
	SensorTypePostMemoryResize             SensorType = 0x0e
	SensorTypeSystemFirmware               SensorType = 0x0f
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
	SensorTypeOtherFRU                     SensorType = 0x1a
	SensorTypeCableInterconnect            SensorType = 0x1b
	SensorTypeTerminator                   SensorType = 0x1c
	SensorTypeSystemBootRestartInitiated   SensorType = 0x1d
	SensorTypeBootError                    SensorType = 0x1e
	SensorTypeBaseOSBootInstallationStatus SensorType = 0x1f
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
	SensorTypeSessionAudit                 SensorType = 0x2a
	SensorTypeVersionChange                SensorType = 0x2b
	SensorTypeFRUState                     SensorType = 0x2c

	// Reserverd: 0x2D - 0xBF
	// OEM Reserved: 0xC0 - 0xFF
)

var sensorTypeMap = map[SensorType]string{
	0x00: "Reserved",
	0x01: "Temperature",
	0x02: "Voltage",
	0x03: "Current",
	0x04: "Fan",
	0x05: "Physical Security",
	0x06: "Platform Security",
	0x07: "Processor",
	0x08: "Power Supply",
	0x09: "Power Unit",
	0x0a: "Cooling Device",
	0x0b: "Other",
	0x0c: "Memory",
	0x0d: "Drive Slot / Bay",
	0x0e: "POST Memory Resize",
	0x0f: "System Firmwares",
	0x10: "Event Logging Disabled",
	0x11: "Watchdog1",
	0x12: "System Event",
	0x13: "Critical Interrupt",
	0x14: "Button",
	0x15: "Module / Board",
	0x16: "Microcontroller/Coprocessor",
	0x17: "Add-in Card",
	0x18: "Chassis",
	0x19: "Chip Set",
	0x1a: "Other FRU",
	0x1b: "Cable / Interconnect",
	0x1c: "Terminator",
	0x1d: "System Boot Initiated",
	0x1e: "Boot Error",
	0x1f: "OS Boot",
	0x20: "OS Critical Stop",
	0x21: "Slot / Connector",
	0x22: "System ACPI Power State",
	0x23: "Watchdog2",
	0x24: "Platform Alert",
	0x25: "Entity Presence",
	0x26: "Monitor ASIC",
	0x27: "LAN",
	0x28: "Management Subsys Health",
	0x29: "Battery",
	0x2a: "Session Audit",
	0x2b: "Version Change",
	0x2c: "FRU State",
}

// 43.17 Sensor Unit Type Codes
type SensorUnit uint8

const (
	SensorUnitUnspecified        SensorUnit = 0  // unspecified
	SensorUnitDegressC           SensorUnit = 1  // degrees C, Celsius, 摄氏度 ℃
	SensorUnitDegreesF           SensorUnit = 2  // degrees F，Fahrenheit, 华氏度
	SensorUnitDegreesK           SensorUnit = 3  // degrees K，Kelvins, 开尔文
	SensorUnitVolts              SensorUnit = 4  // Volts, 伏特（电压单位）
	SensorUnitAmps               SensorUnit = 5  // Amps, 安培数
	SensorUnitWatts              SensorUnit = 6  // Watts, 瓦特（功率单位）
	SensorUnitJoules             SensorUnit = 7  // Joules, 焦耳
	SensorUnitCoulombs           SensorUnit = 8  // Coulombs, 库伦
	SensorUnitVA                 SensorUnit = 9  // VA, 伏安
	SensorUnitNits               SensorUnit = 10 // Nits, 尼特（光度单位）
	SensorUnitLumen              SensorUnit = 11 // lumen, 流明（光通量单位）
	SensorUnitLux                SensorUnit = 12 // lux, 勒克斯（照明单位）
	SensorUnitCandela            SensorUnit = 13 // Candela, 坎，坎德拉（发光强度单位）
	SensorUnitKPa                SensorUnit = 14 // kPa kilopascal, 千帕, 千帕斯卡
	SensorUnitPSI                SensorUnit = 15 // PSI
	SensorUnitNewton             SensorUnit = 16 // Newton, 牛顿（力的单位）
	SensorUnitCFM                SensorUnit = 17 // CFM, 风量, cubic feet per minute (cu ft/min)
	SensorUnitRPM                SensorUnit = 18 // RPM, 每分钟转数，Revolutions per minute, is the number of turns in one minute
	SensorUnitHz                 SensorUnit = 19 // Hz, 赫兹
	SensorUnitMicroSecond        SensorUnit = 20 // microsecond， 微秒
	SensorUnitMilliSecond        SensorUnit = 21 // millisecond， 毫秒
	SensorUnitSecond             SensorUnit = 22 // second，秒
	SensorUnitMinute             SensorUnit = 23 // minute， 分
	SensorUnitHour               SensorUnit = 24 // hour，时
	SensorUnitDay                SensorUnit = 25 // day，日
	SensorUnitWeek               SensorUnit = 26 // week，周
	SensorUnitMil                SensorUnit = 27 // mil, 毫升；密耳（千分之一寸）
	SensorUnitInches             SensorUnit = 28 // inches, 英寸（inch的复数）
	SensorUnitFleet              SensorUnit = 29 // feet
	SensorUnitCuIn               SensorUnit = 30 // cu in, 立方英寸（cubic inch）
	SensorUnitCuFleet            SensorUnit = 31 // cu feet
	SensorUnitMM                 SensorUnit = 32 // mm, 毫米（millimeter）
	SensorUnitCM                 SensorUnit = 33 // cm, 厘米（centimeter）
	SensorUnitM                  SensorUnit = 34 // m, 米
	SensorUnitCuCM               SensorUnit = 35 // cu cm
	SensorUnitCum                SensorUnit = 36 // cum
	SensorUnitLiters             SensorUnit = 37 // liters, 公升（容量单位）
	SensorUnitFluidOunce         SensorUnit = 38 // fluid ounce, 液盎司（液体容量单位，等于 fluidounce）
	SensorUnitRadians            SensorUnit = 39 // radians, 弧度（radian的复数）
	SensorUnitvSteradians        SensorUnit = 40 // steradians, 球面度，立体弧度（立体角国际单位制，等于 sterad）
	SensorUnitRevolutions        SensorUnit = 41 // revolutions, 转数（revolution的复数形式）
	SensorUnitCycles             SensorUnit = 42 // cycles， 周期，圈
	SensorUnitGravities          SensorUnit = 43 // gravities， 重力
	SensorUnitOunce              SensorUnit = 44 // ounce， 盎司
	SensorUnitPound              SensorUnit = 45 // pound, 英镑
	SensorUnitFootPound          SensorUnit = 46 // ft-lb， 英尺-磅（foot pound）
	SensorUnitOzIn               SensorUnit = 47 // oz-in， 扭力；盎司英寸
	SensorUnitGauss              SensorUnit = 48 // gauss， 高斯（磁感应或磁场的单位）
	SensorUnitGilberts           SensorUnit = 49 // gilberts， 吉伯（磁通量的单位）
	SensorUnitHenry              SensorUnit = 50 // henry, 亨利（电感单位）
	SensorUnitMilliHenry         SensorUnit = 51 // millihenry, 毫亨（利）（电感单位）
	SensorUnitFarad              SensorUnit = 52 // farad, 法拉（电容单位）
	SensorUnitMicroFarad         SensorUnit = 53 // microfarad, 微法拉（电容量的实用单位）
	SensorUnitOhms               SensorUnit = 54 // ohms， 欧姆（Ohm） ：电阻的量度单位，欧姆值越大，电阻越大
	SensorUnitSiemens            SensorUnit = 55 // siemens， 西门子， 电导单位
	SensorUnitMole               SensorUnit = 56 // mole， 摩尔 [化学] 克分子（等于mole）
	SensorUnitBecquerel          SensorUnit = 57 // becquerel， 贝可（放射性活度单位）
	SensorUnitPPM                SensorUnit = 58 // PPM (parts/million)， 百万分率，百万分之…（parts per million）
	SensorUnitReserved           SensorUnit = 59 // reserved
	SensorUnitDecibels           SensorUnit = 60 // Decibels，分贝（声音强度单位，decibel的复数）
	SensorUnitDbA                SensorUnit = 61 // DbA, dBA is often used to specify the loudness of the fan used to cool the microprocessor and associated components. Typical dBA ratings are in the neighborhood of 25 dBA, representing 25 A-weighted decibels above the threshold of hearing. This is approximately the loudness of a person whispering in a quiet room.
	SensorUnitDbC                SensorUnit = 62 // DbC
	SensorUnitGray               SensorUnit = 63 // gray，核吸收剂量(Gy)
	SensorUnitSevert             SensorUnit = 64 // sievert, 希沃特（辐射效果单位，简称希）
	SensorUnitColorTempDegK      SensorUnit = 65 // color temp deg K, 色温
	SensorUnitBit                SensorUnit = 66 // bit, 比特（二进位制信息单位）
	SensorUnitKilobit            SensorUnit = 67 // kilobit, 千比特
	SensorUnitMegabit            SensorUnit = 68 // megabit, 兆比特
	SensorUnitGigabit            SensorUnit = 69 // gigabit，吉比特
	SensorUnitByte               SensorUnit = 70 // byte， 字节
	SensorUnitKilobyte           SensorUnit = 71 // kilobyte，千字节
	SensorUnitMegabyte           SensorUnit = 72 // megabyte，兆字节
	SensorUnitGigabyte           SensorUnit = 73 // gigabyte，吉字节
	SensorUnitWord               SensorUnit = 74 // word (data)，字
	SensorUnitDWord              SensorUnit = 75 // dword， 双字
	SensorUnitQWord              SensorUnit = 76 // qword， 四字
	SensorUnitLine               SensorUnit = 77 // line (re. mem. line)
	SensorUnitHit                SensorUnit = 78 // hit, 命中
	SensorUnitMiss               SensorUnit = 79 // miss， 未击中， 未命中
	SensorUnitRetry              SensorUnit = 80 // retry, 重试（次数）
	SensorUnitReset              SensorUnit = 81 // reset，重置（次数）
	SensorUnitOverrun            SensorUnit = 82 // overrun) / overflow 满载，溢出（次数）
	SensorUnitUnderrun           SensorUnit = 83 // underrun 欠载
	SensorUnitCollision          SensorUnit = 84 // collision, 冲突
	SensorUnitPacket             SensorUnit = 85 // packets, 包, 数据包
	SensorUnitMessage            SensorUnit = 86 // messages, 消息
	SensorUnitCharacters         SensorUnit = 87 // characters，字符
	SensorUnitError              SensorUnit = 88 // error， 错误
	SensorUnitCorrectableError   SensorUnit = 89 // correctable error 可校正错误
	SensorUnitUncorrectableError SensorUnit = 90 // uncorrectable error 不可校正错误
	SensorUnitFatalError         SensorUnit = 91 // fatal error， 致命错误，不可恢复的错误
	SensorUnitGrams              SensorUnit = 92 // grams, 克（gram的复数形式）
)
