package ipmi

import (
	"bytes"
	"fmt"
	"math"

	"github.com/olekukonko/tablewriter"
)

// 42.1
// Sensors are classified according to the type of readings they provide and/or the type of events they generate.
//
// Three sensor classes: threshold, discrete, oem
// (oem is a special case of discrete)
//
// A sensor can return either an analog or discrete readings. Sensor events can be discrete or threshold-based.
// Valid sensorclass string values are:
// "N/A", "threshold", "discrete", "oem"
type SensorClass string

const (
	SensorClassNotApplicable SensorClass = "N/A" // 不适用的

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
	return "unknown"
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
	SensorTypePowerSupply                  SensorType = 0x08
	SensorTypePowerUnit                    SensorType = 0x09
	SensorTypeCollingDevice                SensorType = 0x0a
	SensorTypeOtherUnitsbased              SensorType = 0x0b
	SensorTypeMemory                       SensorType = 0x0c
	SensorTypeDriveSlot                    SensorType = 0x0d
	SensorTypePostMemoryResize             SensorType = 0x0e
	SensorTypeSystemFirmwareProgress       SensorType = 0x0f
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
	SensorTypePlatformAlert                SensorType = 0x24
	SensorTypeEntityPresence               SensorType = 0x25
	SensorTypeMonitorASIC                  SensorType = 0x26
	SensorTypeLAN                          SensorType = 0x27
	SensorTypeManagementSubsystemHealth    SensorType = 0x28
	SensorTypeBattery                      SensorType = 0x29
	SensorTypeSessionAudit                 SensorType = 0x2a
	SensorTypeVersionChange                SensorType = 0x2b
	SensorTypeFRUState                     SensorType = 0x2c

	// Reserved: 0x2D - 0xBF
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
type SensorUnitType uint8

const (
	SensorUnitType_Unspecified        SensorUnitType = 0  // unspecified
	SensorUnitType_DegreesC           SensorUnitType = 1  // degrees C, Celsius, 摄氏度 ℃
	SensorUnitType_DegreesF           SensorUnitType = 2  // degrees F, Fahrenheit, 华氏度
	SensorUnitType_DegreesK           SensorUnitType = 3  // degrees K, Kelvins, 开尔文
	SensorUnitType_Volts              SensorUnitType = 4  // Volts, 伏特（电压单位）
	SensorUnitType_Amps               SensorUnitType = 5  // Amps, 安培数
	SensorUnitType_Watts              SensorUnitType = 6  // Watts, 瓦特（功率单位）
	SensorUnitType_Joules             SensorUnitType = 7  // Joules, 焦耳
	SensorUnitType_Coulombs           SensorUnitType = 8  // Coulombs, 库伦
	SensorUnitType_VA                 SensorUnitType = 9  // VA, 伏安
	SensorUnitType_Nits               SensorUnitType = 10 // Nits, 尼特（光度单位）
	SensorUnitType_Lumen              SensorUnitType = 11 // lumen, 流明（光通量单位）
	SensorUnitType_Lux                SensorUnitType = 12 // lux, 勒克斯（照明单位）
	SensorUnitType_Candela            SensorUnitType = 13 // Candela, 坎, 坎德拉（发光强度单位）
	SensorUnitType_KPa                SensorUnitType = 14 // kPa kilopascal, 千帕, 千帕斯卡
	SensorUnitType_PSI                SensorUnitType = 15 // PSI
	SensorUnitType_Newton             SensorUnitType = 16 // Newton, 牛顿（力的单位）
	SensorUnitType_CFM                SensorUnitType = 17 // CFM, 风量, cubic feet per minute (cu ft/min)
	SensorUnitType_RPM                SensorUnitType = 18 // RPM, 每分钟转数, Revolutions per minute, is the number of turns in one minute
	SensorUnitType_Hz                 SensorUnitType = 19 // Hz, 赫兹
	SensorUnitType_MicroSecond        SensorUnitType = 20 // microsecond, 微秒
	SensorUnitType_MilliSecond        SensorUnitType = 21 // millisecond, 毫秒
	SensorUnitType_Second             SensorUnitType = 22 // second, 秒
	SensorUnitType_Minute             SensorUnitType = 23 // minute, 分
	SensorUnitType_Hour               SensorUnitType = 24 // hour, 时
	SensorUnitType_Day                SensorUnitType = 25 // day, 日
	SensorUnitType_Week               SensorUnitType = 26 // week, 周
	SensorUnitType_Mil                SensorUnitType = 27 // mil, 毫升；密耳（千分之一寸）
	SensorUnitType_Inches             SensorUnitType = 28 // inches, 英寸（inch的复数）
	SensorUnitType_Fleet              SensorUnitType = 29 // feet
	SensorUnitType_CuIn               SensorUnitType = 30 // cu in, 立方英寸（cubic inch）
	SensorUnitType_CuFleet            SensorUnitType = 31 // cu feet
	SensorUnitType_MM                 SensorUnitType = 32 // mm, 毫米（millimeter）
	SensorUnitType_CM                 SensorUnitType = 33 // cm, 厘米（centimeter）
	SensorUnitType_M                  SensorUnitType = 34 // m, 米
	SensorUnitType_CuCM               SensorUnitType = 35 // cu cm
	SensorUnitType_Cum                SensorUnitType = 36 // cum
	SensorUnitType_Liters             SensorUnitType = 37 // liters, 公升（容量单位）
	SensorUnitType_FluidOunce         SensorUnitType = 38 // fluid ounce, 液盎司（液体容量单位, 等于 fluidounce）
	SensorUnitType_Radians            SensorUnitType = 39 // radians, 弧度（radian的复数）
	SensorUnitType_vSteradians        SensorUnitType = 40 // steradians, 球面度, 立体弧度（立体角国际单位制, 等于 sterad）
	SensorUnitType_Revolutions        SensorUnitType = 41 // revolutions, 转数（revolution的复数形式）
	SensorUnitType_Cycles             SensorUnitType = 42 // cycles, 周期, 圈
	SensorUnitType_Gravities          SensorUnitType = 43 // gravities, 重力
	SensorUnitType_Ounce              SensorUnitType = 44 // ounce, 盎司
	SensorUnitType_Pound              SensorUnitType = 45 // pound, 英镑
	SensorUnitType_FootPound          SensorUnitType = 46 // ft-lb, 英尺-磅（foot pound）
	SensorUnitType_OzIn               SensorUnitType = 47 // oz-in, 扭力；盎司英寸
	SensorUnitType_Gauss              SensorUnitType = 48 // gauss, 高斯（磁感应或磁场的单位）
	SensorUnitType_Gilberts           SensorUnitType = 49 // gilberts, 吉伯（磁通量的单位）
	SensorUnitType_Henry              SensorUnitType = 50 // henry, 亨利（电感单位）
	SensorUnitType_MilliHenry         SensorUnitType = 51 // millihenry, 毫亨（利）（电感单位）
	SensorUnitType_Farad              SensorUnitType = 52 // farad, 法拉（电容单位）
	SensorUnitType_MicroFarad         SensorUnitType = 53 // microfarad, 微法拉（电容量的实用单位）
	SensorUnitType_Ohms               SensorUnitType = 54 // ohms, 欧姆（Ohm） ：电阻的量度单位, 欧姆值越大, 电阻越大
	SensorUnitType_Siemens            SensorUnitType = 55 // siemens, 西门子, 电导单位
	SensorUnitType_Mole               SensorUnitType = 56 // mole, 摩尔 [化学] 克分子（等于mole）
	SensorUnitType_Becquerel          SensorUnitType = 57 // becquerel, 贝可（放射性活度单位）
	SensorUnitType_PPM                SensorUnitType = 58 // PPM (parts/million), 百万分率, 百万分之…（parts per million）
	SensorUnitType_Reserved           SensorUnitType = 59 // reserved
	SensorUnitType_Decibels           SensorUnitType = 60 // Decibels, 分贝（声音强度单位, decibel的复数）
	SensorUnitType_DbA                SensorUnitType = 61 // DbA, dBA is often used to specify the loudness of the fan used to cool the microprocessor and associated components. Typical dBA ratings are in the neighborhood of 25 dBA, representing 25 A-weighted decibels above the threshold of hearing. This is approximately the loudness of a person whispering in a quiet room.
	SensorUnitType_DbC                SensorUnitType = 62 // DbC
	SensorUnitType_Gray               SensorUnitType = 63 // gray, 核吸收剂量(Gy)
	SensorUnitType_Sievert            SensorUnitType = 64 // sievert, 希沃特（辐射效果单位, 简称希）
	SensorUnitType_ColorTempDegK      SensorUnitType = 65 // color temp deg K, 色温
	SensorUnitType_Bit                SensorUnitType = 66 // bit, 比特（二进位制信息单位）
	SensorUnitType_Kilobit            SensorUnitType = 67 // kilobit, 千比特
	SensorUnitType_Megabit            SensorUnitType = 68 // megabit, 兆比特
	SensorUnitType_Gigabit            SensorUnitType = 69 // gigabit, 吉比特
	SensorUnitType_Byte               SensorUnitType = 70 // byte, 字节
	SensorUnitType_Kilobyte           SensorUnitType = 71 // kilobyte, 千字节
	SensorUnitType_Megabyte           SensorUnitType = 72 // megabyte, 兆字节
	SensorUnitType_Gigabyte           SensorUnitType = 73 // gigabyte, 吉字节
	SensorUnitType_Word               SensorUnitType = 74 // word (data), 字
	SensorUnitType_DWord              SensorUnitType = 75 // dword, 双字
	SensorUnitType_QWord              SensorUnitType = 76 // qword, 四字
	SensorUnitType_Line               SensorUnitType = 77 // line (re. mem. line)
	SensorUnitType_Hit                SensorUnitType = 78 // hit, 命中
	SensorUnitType_Miss               SensorUnitType = 79 // miss, 未击中, 未命中
	SensorUnitType_Retry              SensorUnitType = 80 // retry, 重试（次数）
	SensorUnitType_Reset              SensorUnitType = 81 // reset, 重置（次数）
	SensorUnitType_Overrun            SensorUnitType = 82 // overrun) / overflow 满载, 溢出（次数）
	SensorUnitType_Underrun           SensorUnitType = 83 // underrun 欠载
	SensorUnitType_Collision          SensorUnitType = 84 // collision, 冲突
	SensorUnitType_Packet             SensorUnitType = 85 // packets, 包, 数据包
	SensorUnitType_Message            SensorUnitType = 86 // messages, 消息
	SensorUnitType_Characters         SensorUnitType = 87 // characters, 字符
	SensorUnitType_Error              SensorUnitType = 88 // error, 错误
	SensorUnitType_CorrectableError   SensorUnitType = 89 // correctable error 可校正错误
	SensorUnitType_UncorrectableError SensorUnitType = 90 // uncorrectable error 不可校正错误
	SensorUnitType_FatalError         SensorUnitType = 91 // fatal error, 致命错误, 不可恢复的错误
	SensorUnitType_Grams              SensorUnitType = 92 // grams, 克（gram的复数形式）
)

func (u SensorUnitType) String() string {
	s, ok := sensorUnitMap[u]
	if ok {
		return s
	}
	return ""
}

var sensorUnitMap = map[SensorUnitType]string{
	0:  "unspecified",
	1:  "degrees C",
	2:  "degrees F",
	3:  "degrees K",
	4:  "Volts",
	5:  "Amps",
	6:  "Watts",
	7:  "Joules",
	8:  "Coulombs",
	9:  "VA",
	10: "Nits",
	11: "lumen",
	12: "lux",
	13: "Candela",
	14: "kPa",
	15: "PSI",
	16: "Newton",
	17: "CFM",
	18: "RPM",
	19: "Hz",
	20: "microsecond",
	21: "millisecond",
	22: "second",
	23: "minute",
	24: "hour",
	25: "day",
	26: "week",
	27: "mil",
	28: "inches",
	29: "feet",
	30: "cu in",
	31: "cu feet",
	32: "mm",
	33: "cm",
	34: "m",
	35: "cu cm",
	36: "cu m",
	37: "liters",
	38: "fluid ounce",
	39: "radians",
	40: "steradians",
	41: "revolutions",
	42: "cycles",
	43: "gravities",
	44: "ounce",
	45: "pound",
	46: "ft-lb",
	47: "oz-in",
	48: "gauss",
	49: "gilberts",
	50: "henry",
	51: "millihenry",
	52: "farad",
	53: "microfarad",
	54: "ohms",
	55: "siemens",
	56: "mole",
	57: "becquerel",
	58: "PPM",
	59: "reserved",
	60: "Decibels",
	61: "DbA",
	62: "DbC",
	63: "gray",
	64: "sievert",
	65: "color temp deg K",
	66: "bit",
	67: "kilobit",
	68: "megabit",
	69: "gigabit",
	70: "byte",
	71: "kilobyte",
	72: "megabyte",
	73: "gigabyte",
	74: "word",
	75: "dword",
	76: "qword",
	77: "line",
	78: "hit",
	79: "miss",
	80: "retry",
	81: "reset",
	82: "overflow",
	83: "underrun",
	84: "collision",
	85: "packets",
	86: "messages",
	87: "characters",
	88: "error",
	89: "correctable error",
	90: "uncorrectable error",
	91: "fatal error",
	92: "grams",
}

// SensorThresholdType are enums for types of threshold
type SensorThresholdType string

const (
	SensorThresholdType_LNC SensorThresholdType = "LowerNonCritical"
	SensorThresholdType_LCR SensorThresholdType = "LowerCritical"
	SensorThresholdType_LNR SensorThresholdType = "LowerNonRecoverable"
	SensorThresholdType_UNC SensorThresholdType = "UpperNonCritical"
	SensorThresholdType_UCR SensorThresholdType = "UpperCritical"
	SensorThresholdType_UNR SensorThresholdType = "UpperNonRecoverable"
)

func (sensorThresholdType SensorThresholdType) Abbr() string {
	m := map[SensorThresholdType]string{
		SensorThresholdType_LNC: "lnc",
		SensorThresholdType_LCR: "lcr",
		SensorThresholdType_LNR: "lnr",
		SensorThresholdType_UNC: "unc",
		SensorThresholdType_UCR: "ucr",
		SensorThresholdType_UNR: "unr",
	}
	s, ok := m[sensorThresholdType]
	if ok {
		return s
	}
	return ""
}

type SensorThresholdTypes []SensorThresholdType

func (types SensorThresholdTypes) Strings() []string {
	out := []string{}
	for _, v := range types {
		out = append(out, v.Abbr())
	}
	return out
}

// SensorThresholdStatus are enums for threshold status of sensor.
//
// ....UNR status (NonRecoverable)
// -----------------UNR threshold
// ....UCR status (Critical)
// -----------------UCR threshold
// ....UNC status (NonCritical)
// -----------------UNC threshold
// ....OK status (OK)
// -----------------LNC threshold
// ....LNC status (NonCritical)
// -----------------LCR threshold
// ....LCR status (Critical)
// -----------------LNR threshold
// ....LNR status (NonRecoverable)
type SensorThresholdStatus string

const (
	SensorThresholdStatus_OK  = "ok"
	SensorThresholdStatus_LNC = "lnc"
	SensorThresholdStatus_LCR = "lcr"
	SensorThresholdStatus_LNR = "lnr"
	SensorThresholdStatus_UNC = "unc"
	SensorThresholdStatus_UCR = "ucr"
	SensorThresholdStatus_UNR = "unr"
)

type SensorStatus string

const (
	// SensorStatusOK means okay (the sensor is present and operating correctly)
	SensorStatusOK = "OK"

	// SensorStatusNoSensor means no sensor (corresponding reading will say disabled or Not Readable)
	SensorStatusNoSensor = "N/A"

	// SensorStatusNonCritical means non-critical error (lower or upper)
	SensorStatusNonCritical = "NC"

	// SensorStatusCritical means critical error (lower or upper)
	SensorStatusCritical = "CR"

	// SensorStatusNonRecoverable means non-recoverable error (lower or upper)
	SensorStatusNonRecoverable = "NR"
)

// SensorThreshold holds all values and attributes of a specified threshold type.
type SensorThreshold struct {
	// type of threshold
	Type SensorThresholdType
	Mask Mask_Threshold
	// threshold raw reading value before conversion
	Raw uint8
}

// LinearizationFunc is linearization function used in "Sensor Reading Conversion Formula"
// 线性化函数
type LinearizationFunc uint8

const (
	LinearizationFunc_Linear LinearizationFunc = 0x00
	LinearizationFunc_LN     LinearizationFunc = 0x01
	LinearizationFunc_LOG10  LinearizationFunc = 0x02
	LinearizationFunc_LOG2   LinearizationFunc = 0x03
	LinearizationFunc_E      LinearizationFunc = 0x04
	LinearizationFunc_EXP10  LinearizationFunc = 0x05
	LinearizationFunc_EXP2   LinearizationFunc = 0x06
	LinearizationFunc_1X     LinearizationFunc = 0x07
	LinearizationFunc_SQR    LinearizationFunc = 0x08
	LinearizationFunc_CUBE   LinearizationFunc = 0x09
	LinearizationFunc_SQRT   LinearizationFunc = 0x0a
	LinearizationFunc_CBRT   LinearizationFunc = 0x0b

	// 70h = non-linear.
	// 71h-7Fh = non-linear OEM
	LinearizationFunc_NonLinear LinearizationFunc = 0x70
)

func (l LinearizationFunc) IsNonLinear() bool {
	if uint8(l) >= 0x70 && uint8(l) <= 0x7f {
		return true
	}
	return false
}

func (l LinearizationFunc) String() string {
	m := map[LinearizationFunc]string{
		0x00: "linear",
		0x01: "ln",
		0x02: "log10",
		0x03: "log2",
		0x04: "e",
		0x05: "exp10",
		0x06: "exp2",
		0x07: "1/x",
		0x08: "sqr(x)",  // 平方 sqr(3) = 9
		0x09: "cube(x)", // 立方 cube(3) = 27
		0x0a: "sqrt(x)", // 平方根 sqrt(9) = 3
		0x0b: "cbrt(x)", // 立方根 cbrt(27) = 3
	}
	s, ok := m[l]
	if ok {
		return fmt.Sprintf("%s(%d)", s, uint8(l))
	}
	return fmt.Sprintf("(%d)", uint8(l))
}

// Apply applies linearization func (itself) to the input value and returns the result.
func (l LinearizationFunc) Apply(x float64) float64 {
	switch l {
	case LinearizationFunc_LN:
		return math.Log(float64(x))
	case LinearizationFunc_LOG10:
		return math.Log10(float64(x))
	case LinearizationFunc_LOG2:
		return math.Log2(float64(x))
	case LinearizationFunc_E:
		return math.Pow(math.E, float64(x))
	case LinearizationFunc_EXP10:
		return math.Pow10(int(x))
	case LinearizationFunc_EXP2:
		return math.Exp2(float64(x))
	case LinearizationFunc_1X:
		return math.Pow(float64(x), -1)
	case LinearizationFunc_SQR:
		return math.Pow(float64(x), 2.0)
	case LinearizationFunc_CUBE:
		return math.Pow(float64(x), 3.0)
	case LinearizationFunc_SQRT:
		return math.Sqrt(float64(x))
	case LinearizationFunc_CBRT:
		return math.Cbrt(float64(x))
	case LinearizationFunc_Linear:
		// `linear means y=f(x)=x`, nothing to do
	default:
		// other values mean sensor is non-linear, also no linearization function is applied. (see 36.2 third paragraph)
	}
	return x
}

type SensorUnit struct {
	AnalogDataFormat SensorAnalogUnitFormat
	RateUnit         SensorRateUnit
	ModifierRelation SensorModifierRelation
	Percentage       bool // Percentage 0b = no, 1b = yes

	BaseUnit     SensorUnitType
	ModifierUnit SensorUnitType
}

func (unit SensorUnit) String() string {
	if !unit.IsAnalog() {
		return "discrete"
	}
	// return unit.BaseUnit.String()

	var percentageStr string
	if unit.Percentage {
		percentageStr = "% "
	}

	switch unit.ModifierRelation {

	case SensorModifierRelation_Div:
		return fmt.Sprintf("%s%s/%s", percentageStr, unit.BaseUnit, unit.ModifierUnit)

	case SensorModifierRelation_Mul:
		return fmt.Sprintf("%s%s*%s", percentageStr, unit.BaseUnit, unit.ModifierUnit)

	// SensorModifierRelation_None:
	default:
		if unit.BaseUnit == SensorUnitType_Unspecified && unit.Percentage {
			return "percent"
		}
		return fmt.Sprintf("%s%s", percentageStr, unit.BaseUnit)
	}
}

func (unit SensorUnit) IsAnalog() bool {
	return unit.AnalogDataFormat != SensorAnalogUnitFormat_NotAnalog
}

type SensorAnalogUnitFormat uint8

const (
	SensorAnalogUnitFormat_Unsigned     SensorAnalogUnitFormat = 0 // unsigned
	SensorAnalogUnitFormat_1sComplement SensorAnalogUnitFormat = 1 // 1's complement (signed)
	SensorAnalogUnitFormat_2sComplement SensorAnalogUnitFormat = 2 // 2's complement (signed)
	SensorAnalogUnitFormat_NotAnalog    SensorAnalogUnitFormat = 3 // does not return analog (numeric) reading
)

func (format SensorAnalogUnitFormat) String() string {
	m := map[SensorAnalogUnitFormat]string{
		0: "unsigned",
		1: "1s comp",
		2: "2s comp",
		3: "not analog",
	}
	s, ok := m[format]
	if ok {
		return s
	}
	return "unknown"
}

type SensorRateUnit uint8

const (
	SensorRateUnit_None        SensorRateUnit = 0
	SensorRateUnit_PerMicroSec SensorRateUnit = 1
	SensorRateUnit_PerMilliSec SensorRateUnit = 2
	SensorRateUnit_PerSec      SensorRateUnit = 3
	SensorRateUnit_PerMin      SensorRateUnit = 4
	SensorRateUnit_PerHour     SensorRateUnit = 5
	SensorRateUnit_PerDay      SensorRateUnit = 6
	SensorRateUnit_Reserved    SensorRateUnit = 7
)

func (unit SensorRateUnit) String() string {
	m := map[SensorRateUnit]string{
		0: "none",
		1: "per µS",
		2: "per ms",
		3: "per s",
		4: "per minute",
		5: "per hour",
		6: "per day",
		7: "reserved",
	}
	s, ok := m[unit]
	if ok {
		return s
	}
	return ""
}

type SensorModifierRelation uint8

const (
	SensorModifierRelation_None SensorModifierRelation = 0
	SensorModifierRelation_Div  SensorModifierRelation = 1 // Basic Unit / Modifier Unit
	SensorModifierRelation_Mul  SensorModifierRelation = 2 // Basic Unit * Modifier Unit
)

func (unit SensorModifierRelation) String() string {
	m := map[SensorModifierRelation]string{
		0: "none",
		1: "div",
		2: "mul",
		3: "reserved",
	}
	s, ok := m[unit]
	if ok {
		return s
	}
	return ""
}

// SensorEventMessageControl indicates whether this sensor generates Event Messages,
// and if so, what type of Event Message control is offered.
type SensorEventMessageControl uint8

const (
	// per threshold/discrete-state event enable/disable control (implies
	// that entire sensor and global disable are also supported)
	SensorEventMessageControl_PerThresholdState SensorEventMessageControl = 0
	// entire sensor only (implies that global disable is also supported)
	SensorEventMessageControl_EntireSensorOnly SensorEventMessageControl = 1
	// global disable only
	SensorEventMessageControl_GlobalDisableOnly SensorEventMessageControl = 2
	// no events from sensor
	SensorEventMessageControl_NoEvents SensorEventMessageControl = 3
)

func (a SensorEventMessageControl) String() string {
	switch a {
	case 0:
		return "Per-threshold"
	case 1:
		return "Entire Sensor Only"
	case 2:
		return "Global Disable Only"
	case 3:
		return "No Events From Sensor"
	default:
		return ""
	}

}

// SensorThresholdAccess represents the access mode for the threshold value of the sensor.
type SensorThresholdAccess uint8

const (
	// no thresholds.
	SensorThresholdAccess_No SensorThresholdAccess = 0
	// thresholds are readable, per Reading Mask
	SensorThresholdAccess_Readable SensorThresholdAccess = 1
	// thresholds are readable and settable, per Reading Mask and Settable Threshold Mask, respectively.
	SensorThresholdAccess_ReadableSettable SensorThresholdAccess = 2
	// Fixed, unreadable, thresholds.
	// Which thresholds are supported is reflected by the Reading Mask.
	// The threshold value fields report the values that are hard-coded in the sensor.
	SensorThresholdAccess_Fixed SensorThresholdAccess = 3
)

func (a SensorThresholdAccess) String() string {
	switch a {
	case 0:
		return "No"
	case 1:
		return "Readable"
	case 2:
		return "ReadableSettable"
	case 3:
		return "FixedUnreadable"
	default:
		return ""
	}
}

// SensorHysteresisAccess represents the access mode for the hysteresis value of the sensor.
type SensorHysteresisAccess uint8

const (
	// No hysteresis, or hysteresis built-in but not specified
	SensorHysteresisAccess_No SensorHysteresisAccess = 0
	// hysteresis is readable.
	SensorHysteresisAccess_Readable SensorHysteresisAccess = 1
	// hysteresis is readable and settable.
	SensorHysteresisAccess_ReadableSettable SensorHysteresisAccess = 2
	// Fixed, unreadable, hysteresis. Hysteresis fields values implemented in the sensor.
	SensorHysteresisAccess_Fixed SensorHysteresisAccess = 3
)

func (a SensorHysteresisAccess) String() string {
	switch a {
	case 0:
		return "No"
	case 1:
		return "Readable"
	case 2:
		return "ReadableSettable"
	case 3:
		return "FixedUnreadable"
	default:
		return ""
	}
}

// ReadingFactors is used in "Sensor Reading Conversion Formula"
// Only Full SDR defines reading factors.
// see: 36.3 Sensor Reading Conversion Formula
type ReadingFactors struct {
	M int16 // 10 bits used

	// in +/- ½ raw counts
	Tolerance uint8 // 6 bits used

	B int16 // 10 bits used

	// Unsigned, 10-bit Basic Sensor Accuracy in 1/100 percent scaled up by unsigned Accuracy exponent.
	Accuracy uint16 // 10 bits, unsigned

	Accuracy_Exp uint8 // 2 bits, unsigned

	// [7:4] - R (result) exponent 4 bits, 2's complement, signed
	// [3:0] - B exponent 4 bits, 2's complement, signed
	R_Exp int8 // 4 bits, signed, also called K2
	B_Exp int8 // 4 bits, signed, also called K1
}

func (f ReadingFactors) String() string {
	return fmt.Sprintf("M: (%d), T: (%d), B: (%d), A: (%d), A_Exp: (%d), R_Exp: (%d), B_Exp: (%d)",
		f.M, f.Tolerance, f.B, f.Accuracy, f.Accuracy_Exp, f.R_Exp, f.B_Exp)
}

// The raw analog data is unpacked as an unsigned integer.
// But whether it is a positive number (>0) or negative number (<0) is determined
// by the "analog data format" field (SensorUnit.AnalogDataFormat)
func AnalogValue(raw uint8, format SensorAnalogUnitFormat) int32 {
	switch format {
	case SensorAnalogUnitFormat_NotAnalog:
		return int32(raw)

	case SensorAnalogUnitFormat_Unsigned:
		return int32(raw)

	case SensorAnalogUnitFormat_1sComplement:
		return int32(onesComplement(uint32(raw), 8))

	case SensorAnalogUnitFormat_2sComplement:
		return int32(twosComplement(uint32(raw), 8))
	}

	return int32(raw)
}

// ConvertReading converts raw sensor reading or raw sensor threshold value to real value in the desired units for the sensor.
//
// see: 36.3 Sensor Reading Conversion Formula
//
//	INPUT: raw (unsigned)
//	  -- APPLY: analogDataFormat
//	    --> GOT: analog (signed)
//	      -- APPLY: factors/linearization
//	        --> GOT: converted (float64)
func ConvertReading(raw uint8, analogDataFormat SensorAnalogUnitFormat, factors ReadingFactors, linearizationFunc LinearizationFunc) float64 {
	// y = L[(Mx + (B * 10^B_Exp) ) * 10^R_Exp ] units

	analog := AnalogValue(raw, analogDataFormat)

	x := float64(analog)

	M := float64(factors.M)
	B := float64(factors.B)
	Bexp := math.Pow(10, float64(factors.B_Exp))
	Rexp := math.Pow(10, float64(factors.R_Exp))

	y := (M*x + B*Bexp) * Rexp

	return linearizationFunc.Apply(y)
}

// ConvertSensorHysteresis converts raw sensor hysteresis value to real value in the desired units for the sensor.
//
// see: 36.3 Sensor Reading Conversion Formula
func ConvertSensorHysteresis(raw uint8, analogDataFormat SensorAnalogUnitFormat, factors ReadingFactors, linearizationFunc LinearizationFunc) float64 {
	// y = L[(Mx + (B * 10^B_Exp) ) * 10^R_Exp ] units

	analog := AnalogValue(raw, analogDataFormat)

	x := float64(analog)

	M := float64(factors.M)
	B := float64(factors.B)
	Bexp := math.Pow(10, float64(factors.B_Exp))
	Rexp := math.Pow(10, float64(factors.R_Exp))

	y := (M*x + B*Bexp) * Rexp

	return linearizationFunc.Apply(y)
}

// ConvertSensorTolerance converts raw sensor tolerance value to real value in the desired units for the sensor.
//
// see: 36.4.1 Tolerance
func ConvertSensorTolerance(raw uint8, analogDataFormat SensorAnalogUnitFormat, factors ReadingFactors, linearizationFunc LinearizationFunc) float64 {
	// y = L[Mx/2 * 10^R_Exp ] units.

	analog := AnalogValue(raw, analogDataFormat)

	x := float64(analog)

	M := float64(factors.M)
	Rexp := math.Pow(10, float64(factors.R_Exp))

	y := (M * x / 2) * Rexp

	return linearizationFunc.Apply(y)
}

// Sensor holds all attribute of a sensor.
type Sensor struct {
	Number uint8
	Name   string

	SDRRecordType    SDRRecordType
	HasAnalogReading bool

	SensorType           SensorType
	EventReadingType     EventReadingType
	SensorUnit           SensorUnit
	SensorInitialization SensorInitialization
	SensorCapabilities   SensorCapabilities

	EntityID       EntityID
	EntityInstance EntityInstance

	scanningDisabled bool // update by GetSensorReading
	readingAvailable bool // update by GetSensorReading

	// Raw reading value before conversion
	Raw uint8
	// reading value after conversion
	Value float64

	Threshold struct {
		Mask Mask_Thresholds

		// Threshold Status, updated by GetSensorReadingResponse.ThresholdStatus()
		ThresholdStatus SensorThresholdStatus

		// Only Full SDR
		LinearizationFunc LinearizationFunc

		ReadingFactors

		LNC_Raw uint8
		LCR_Raw uint8
		LNR_Raw uint8
		UNC_Raw uint8
		UCR_Raw uint8
		UNR_Raw uint8

		LNC float64
		LCR float64
		LNR float64
		UNC float64
		UCR float64
		UNR float64

		PositiveHysteresisRaw uint8
		NegativeHysteresisRaw uint8

		PositiveHysteresis float64
		NegativeHysteresis float64
	}

	Discrete struct {
		Mask         Mask_Discrete
		ActiveStates Mask_DiscreteEvent

		// filled by GetSensorReadingResponse
		optionalData1 uint8
		optionalData2 uint8
	}

	OccurredEvents []SensorEvent
}

func (s *Sensor) String() string {
	sensorReadingStr := fmt.Sprintf("%d", s.Raw)
	sensorValueStr := fmt.Sprintf("%.3f %s", s.Value, s.SensorUnit)
	if s.scanningDisabled {
		sensorReadingStr = "Unable to read sensor: Device Not Present"
		sensorValueStr = "Unable to read sensor: Device Not Present"
	}

	return fmt.Sprintf(
		fmt.Sprintf("Sensor ID              : %s (%#02x)\n", s.Name, s.Number) +
			fmt.Sprintf(" Entity ID            : %d.%d (%s)\n", uint8(s.EntityID), uint8(s.EntityInstance), s.EntityID) +
			fmt.Sprintf(" Sensor Type          : %s (%#02x) (%s)\n", s.SensorType.String(), uint8(s.SensorType), string(s.EventReadingType.SensorClass())) +
			fmt.Sprintf(" Sensor Number        : %#02x\n", s.Number) +
			fmt.Sprintf(" Sensor Name          : %s\n", s.Name) +
			fmt.Sprintf(" Sensor Reading (raw) : %s\n", sensorReadingStr) +
			fmt.Sprintf(" Sensor Value         : %s\n", sensorValueStr) +
			fmt.Sprintf(" Sensor Status        : %s\n", s.Status()),
	)
}

// FormatSensors return a string of table printed for sensors
func FormatSensors(extended bool, sensors ...*Sensor) string {

	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)

	headers := []string{
		"SDRType",
		"SensorNumber",
		"SensorName",
		"SensorType",
		"Reading",
		"Unit",
		"Status",
		"LNR",
		"LCR",
		"LNC",
		"UNC",
		"UCR",
		"UNR",
	}

	if extended {
		headers = append(headers, []string{
			"EventReadingType",
			"AnalogDataFormat",
			"ReadV",
			"ScanD",
			"ReadU",
			"HasAR",
			"DiscreteEvents",
		}...)
	}

	table.SetHeader(headers)
	table.SetFooter(headers)

	for _, sensor := range sensors {
		content := []string{
			sensor.SDRRecordType.String(),
			fmt.Sprintf("%#02x", sensor.Number),
			sensor.Name,
			fmt.Sprintf("%s (%#02x)", sensor.SensorType.String(), uint8(sensor.SensorType)),
			sensor.ReadingStr(),
			sensor.SensorUnit.String(),
			sensor.Status(),
			sensor.ThresholdStr(SensorThresholdType_LNR),
			sensor.ThresholdStr(SensorThresholdType_LCR),
			sensor.ThresholdStr(SensorThresholdType_LNC),
			sensor.ThresholdStr(SensorThresholdType_UNC),
			sensor.ThresholdStr(SensorThresholdType_UCR),
			sensor.ThresholdStr(SensorThresholdType_UNR),
		}

		if extended {
			content = append(content, []string{
				fmt.Sprintf("%s (%#02x)", sensor.EventReadingType.String(), uint8(sensor.EventReadingType)),
				sensor.SensorUnit.AnalogDataFormat.String(),
				fmt.Sprintf("%v", sensor.IsReadingValid()),
				fmt.Sprintf("%v", sensor.scanningDisabled),
				fmt.Sprintf("%v", !sensor.readingAvailable),
				fmt.Sprintf("%v", sensor.HasAnalogReading),
			}...)

			if sensor.IsThreshold() {
				content = append(content, "N/A")
			} else {
				content = append(content, fmt.Sprintf("%v", sensor.Discrete.ActiveStates.TrueEvents()))
			}
		}

		table.Append(content)
	}

	table.Render()

	return buf.String()
}

// IsThreshold returns whether the sensor is threshold sensor class or not.
func (sensor *Sensor) IsThreshold() bool {
	return sensor.EventReadingType.IsThreshold()
}

func (sensor *Sensor) IsReadingValid() bool {
	return sensor.readingAvailable
}

func (sensor *Sensor) IsThresholdAndReadingValid() bool {
	return sensor.IsThreshold() && sensor.IsReadingValid()
}

func (sensor *Sensor) IsThresholdReadable(thresholdType SensorThresholdType) bool {
	if !sensor.IsThreshold() {
		return false
	}

	mask := sensor.Threshold.Mask
	return mask.IsThresholdReadable(thresholdType)
}

// ConvertReading converts raw discrete-sensor reading or raw threshold-sensor value to real value in the desired units for the sensor.
//
// This function can also be applied on raw threshold setting (UNR,UCR,NNC,LNC,LCR,LNR) values.
func (sensor *Sensor) ConvertReading(raw uint8) float64 {
	if sensor.HasAnalogReading {
		return ConvertReading(raw, sensor.SensorUnit.AnalogDataFormat, sensor.Threshold.ReadingFactors, sensor.Threshold.LinearizationFunc)
	}
	return float64(raw)
}

func (sensor *Sensor) ConvertSensorHysteresis(raw uint8) float64 {
	if sensor.HasAnalogReading {
		return ConvertSensorHysteresis(raw, sensor.SensorUnit.AnalogDataFormat, sensor.Threshold.ReadingFactors, sensor.Threshold.LinearizationFunc)
	}
	return float64(raw)
}

func (sensor *Sensor) ConvertSensorTolerance(raw uint8) float64 {
	if sensor.HasAnalogReading {
		return ConvertSensorTolerance(raw, sensor.SensorUnit.AnalogDataFormat, sensor.Threshold.ReadingFactors, sensor.Threshold.LinearizationFunc)
	}
	return float64(raw)
}

// SensorThreshold return SensorThreshold for a specified threshold type.
func (sensor *Sensor) SensorThreshold(thresholdType SensorThresholdType) SensorThreshold {
	switch thresholdType {
	case SensorThresholdType_LNR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.LNR,
			Raw:  sensor.Threshold.LNR_Raw,
		}

	case SensorThresholdType_LCR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.LCR,
			Raw:  sensor.Threshold.LCR_Raw,
		}

	case SensorThresholdType_LNC:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.LNC,
			Raw:  sensor.Threshold.LNC_Raw,
		}

	case SensorThresholdType_UNC:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.UNC,
			Raw:  sensor.Threshold.UNC_Raw,
		}

	case SensorThresholdType_UCR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.UCR,
			Raw:  sensor.Threshold.UCR_Raw,
		}

	case SensorThresholdType_UNR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: sensor.Threshold.Mask.UNR,
			Raw:  sensor.Threshold.UNR_Raw,
		}
	}

	return SensorThreshold{
		Type: thresholdType,
	}
}
func (sensor *Sensor) Status() string {
	if sensor.scanningDisabled {
		return "N/A"
	}

	if !sensor.IsReadingValid() {
		return "N/A"
	}

	if sensor.IsThreshold() {
		return string(sensor.Threshold.ThresholdStatus)
	}

	return fmt.Sprintf("0x%02x%02x", sensor.Discrete.optionalData1, sensor.Discrete.optionalData2)
}

func (sensor *Sensor) ReadingStr() string {
	if sensor.scanningDisabled {
		return "N/A"
	}

	if !sensor.IsReadingValid() {
		return "N/A"
	}

	if sensor.IsThreshold() {
		return fmt.Sprintf("%.3f", sensor.Value)
	}

	return fmt.Sprintf("%d", sensor.Raw)
}

func (sensor *Sensor) ThresholdStr(thresholdType SensorThresholdType) string {
	if !sensor.IsThresholdReadable(thresholdType) {
		return "N/A"
	}

	var value float64
	switch thresholdType {
	case SensorThresholdType_LCR:
		value = sensor.Threshold.LCR
	case SensorThresholdType_LNR:
		value = sensor.Threshold.LNR
	case SensorThresholdType_LNC:
		value = sensor.Threshold.LNC
	case SensorThresholdType_UCR:
		value = sensor.Threshold.UCR
	case SensorThresholdType_UNC:
		value = sensor.Threshold.UNC
	case SensorThresholdType_UNR:
		value = sensor.Threshold.UNR
	}

	return fmt.Sprintf("%.3f", value)
}

func (sensor *Sensor) HysteresisStr(raw uint8) string {
	switch sensor.SDRRecordType {
	case SDRRecordTypeFullSensor:
		if !sensor.SensorUnit.IsAnalog() {
			if raw == 0x00 || raw == 0xff {
				return "unspecified"
			}
			return fmt.Sprintf("%#02x", raw)
		}

		// analog sensor
		value := sensor.ConvertSensorHysteresis(raw)
		if raw == 0x00 || raw == 0xff || value == 0.0 {
			return "unspecified"
		}
		return fmt.Sprintf("%#02x/%.3f", raw, value)

	case SDRRecordTypeCompactSensor:
		if raw == 0x00 || raw == 0xff {
			return "unspecified"
		}
		return fmt.Sprintf("%#02x", raw)
	}

	return ""
}
