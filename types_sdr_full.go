package ipmi

import (
	"fmt"
	"strings"
)

// 43.1 SDRFull Type 01h, Full Sensor Record
//
// The Full Sensor Record can be used to describe any type of sensor.
type SDRFull struct {
	//
	// Record KEY
	//

	// The Record 'Key' Fields are a set of fields that together are unique amongst instances of a given record type.
	// The Record Key bytes shall be contiguous and follow the Record Header.
	// The number of bytes that make up the Record Key field may vary according to record type.

	GeneratorID  GeneratorID
	SensorNumber SensorNumber

	//
	// RECORD BODY
	//

	// Indicates the physical entity that the sensor is monitoring or is otherwise
	// associated with the sensor.
	SensorEntityID       EntityID
	SensorEntityInstance EntityInstance

	// For example, if this bit is set, and the Entity ID is "Processor",
	// the container entity would be considered to represent a logical "Processor Group" rather than a physical processor.
	//
	// This bit is typically used in conjunction with an Entity Association full.
	//
	//  0b = treat entity as a physical entity per Entity ID table
	//  1b = treat entity as a logical container entity.
	SensorEntityIsLogical bool

	SensorInitialization SensorInitialization

	SensorCapabilities SensorCapabilities

	SensorType             SensorType
	SensorEventReadingType EventReadingType

	Mask Mask

	SensorUnit SensorUnit

	// Note, SensorValue is not stored in SDR intrinsically, this field is set by `enhanceSDR`
	// It is fetched by IPMI command GetSensorReading and aligned/converted to SensorUnit based.
	SensorValue float64

	// Note, SensorStatus is not stored in SDR intrinsically, this field is set by `enhanceSDR`
	SensorStatus string

	EntityInstanceSharing uint8

	// see: 36.3 Sensor Reading Conversion Formula
	// y = L[(Mx + (B * 10^B_Exp) ) * 10^R_Exp ] units

	// LinearizationFunc is the Linearization func. (L of the Sensor Reading Conversion Formula)
	//
	// [6:0] - enum (linear, ln, log10, log2, e, exp10, exp2, 1/x, sqr(x), cube(x), sqrt(x),
	// cube-1 (x) )
	// - 70h = non-linear.
	// 71h-7Fh = non-linear, OEM defined.
	LinearizationFunc LinearizationFunc

	ReadingFactors

	// Sensor Direction. Indicates whether the sensor is monitoring an input or
	// output relative to the given Entity. E.g. if the sensor is monitoring a
	// current, this can be used to specify whether it is an input voltage or an
	// output voltage.
	// 00b = unspecified / not applicable
	// 01b = input
	// 10b = output
	// 11b = reserved
	SensorDirection uint8

	// Analog Flags

	NominalReadingSpecified bool
	NormalMaxSpecified      bool
	NormalMinSpecified      bool

	// 额定值, 标称值
	// Given as a raw value. Must be converted to units-based value using the y=Mx+B
	// formula. 1's or 2's complement signed or unsigned per flag bits in Sensor Units 1
	//
	// Only meaningful when NominalReadingSpecified is true
	NominalReadingRaw uint8

	// 最大正常值
	// Only meaningful when NormalMaxSpecified is true
	NormalMaxRaw uint8

	// 最小正常值
	// Only meaningful when NormalMinSpecified is true
	NormalMinRaw uint8

	// Given as a raw value. Must be converted to units-based value based using the
	// y=Mx+B formula. Signed or unsigned per "signed" bit in sensor flags. Normally
	// "FFh" for an 8-bit unsigned sensor, but can be a lesser value if the sensor has a
	// restricted range. If max. reading cannot be pre-specified this value should be set
	// to max value, based on data format, (e.g. FFh for an unsigned sensor, 7Fh for 2"s
	// complement, etc.)
	SensorMaxReadingRaw uint8

	// Given as a raw value. Must be converted to units-based value using the "y=Mx+B"
	// formula. Signed or unsigned per "signed" bit in sensor flags. If min. reading
	// cannot be pre-specified this value should be set to min value, based on data
	// format, (e.g. 00h for an unsigned sensor, 80h for 2"s complement, etc.)
	SensorMinReadingRaw uint8

	// Given as raw value.
	UNR_Raw uint8
	UCR_Raw uint8
	UNC_Raw uint8

	LNR_Raw uint8
	LCR_Raw uint8
	LNC_Raw uint8

	// Positive hysteresis is defined as the unsigned number of counts that are
	// subtracted from the raw threshold values to create the "re-arm" point for all
	// positive-going thresholds on the sensor. 0 indicates that there is no hysteresis on
	// positive-going thresholds for this sensor. Hysteresis values are given as raw
	// counts. That is, to find the degree of hysteresis in units, the value must be
	// converted using the "y=Mx+B" formula.
	//
	// 正向迟滞量
	PositiveHysteresisRaw uint8

	// Negative hysteresis is defined as the unsigned number of counts that are added
	// to the raw threshold value to create the "re-arm" point for all negative-going
	// thresholds on the sensor. 0 indicates that there is no hysteresis on negative-going
	// thresholds for this sensor.
	//
	// 负向迟滞量
	NegativeHysteresisRaw uint8

	IDStringTypeLength TypeLength
	IDStringBytes      []byte
}

// ConvertReading converts raw sensor reading or raw sensor threshold value to real value in the desired units for the sensor.
func (full *SDRFull) ConvertReading(raw uint8) float64 {
	if full.HasAnalogReading() {
		return ConvertReading(raw, full.SensorUnit.AnalogDataFormat, full.ReadingFactors, full.LinearizationFunc)
	}
	return float64(raw)
}

// ConvertSensorHysteresis converts raw sensor hysteresis value to real value in the desired units for the sensor.
func (full *SDRFull) ConvertSensorHysteresis(raw uint8) float64 {
	if full.HasAnalogReading() {
		return ConvertSensorHysteresis(raw, full.SensorUnit.AnalogDataFormat, full.ReadingFactors, full.LinearizationFunc)
	}
	return float64(raw)
}

// ConvertSensorTolerance converts raw sensor tolerance value to real value in the desired units for the sensor.
func (full *SDRFull) ConvertSensorTolerance(raw uint8) float64 {
	if full.HasAnalogReading() {
		return ConvertSensorTolerance(raw, full.SensorUnit.AnalogDataFormat, full.ReadingFactors, full.LinearizationFunc)
	}
	return float64(raw)
}

func (full *SDRFull) ReadingStr(raw uint8, valid bool) string {
	if !valid {
		return "unspecified"
	}

	if !full.SensorUnit.IsAnalog() {
		return fmt.Sprintf("%#02x", raw)
	}

	value := full.ConvertReading(raw)
	return fmt.Sprintf("%#02x/%.3f", raw, value)
}

func (full *SDRFull) ReadingMaxStr() string {
	maxRaw := full.SensorMaxReadingRaw
	analogFormat := full.SensorUnit.AnalogDataFormat
	if (analogFormat == SensorAnalogUnitFormat_Unsigned && maxRaw == 0xff) ||
		(analogFormat == SensorAnalogUnitFormat_1sComplement && maxRaw == 0x00) ||
		(analogFormat == SensorAnalogUnitFormat_2sComplement && maxRaw == 0x7f) ||
		(full.SensorUnit.IsAnalog() && full.ConvertReading(maxRaw) == 0.0) {
		return "unspecified"
	}
	return full.ReadingStr(maxRaw, true)
}

func (full *SDRFull) ReadingMinStr() string {
	minRaw := full.SensorMinReadingRaw
	analogFormat := full.SensorUnit.AnalogDataFormat
	if (analogFormat == SensorAnalogUnitFormat_Unsigned && minRaw == 0x00) ||
		(analogFormat == SensorAnalogUnitFormat_1sComplement && minRaw == 0xff) ||
		(analogFormat == SensorAnalogUnitFormat_2sComplement && minRaw == 0x80) ||
		(full.SensorUnit.IsAnalog() && full.ConvertReading(minRaw) == 0.0) {
		return "unspecified"
	}
	return full.ReadingStr(minRaw, true)
}

// ThresholdValueStr formats a threshold value for specified threshold type.
// If the threshold is not readable, return "not readable".
func (full *SDRFull) ThresholdValueStr(thresholdType SensorThresholdType) string {
	thresholdAttr := full.SensorThreshold(thresholdType)
	return full.ReadingStr(thresholdAttr.Raw, thresholdAttr.Mask.Readable)
}

func (full *SDRFull) HysteresisStr(raw uint8) string {
	if !full.SensorUnit.IsAnalog() {
		if raw == 0x00 || raw == 0xff {
			return "unspecified"
		}
		return fmt.Sprintf("%#02x", raw)
	}

	// analog sensor
	value := full.ConvertSensorHysteresis(raw)
	if raw == 0x00 || raw == 0xff || value == 0.0 {
		return "unspecified"
	}
	return fmt.Sprintf("%#02x/%.3f", raw, value)
}

// SensorThreshold return SensorThreshold for a specified threshold type.
func (full *SDRFull) SensorThreshold(thresholdType SensorThresholdType) SensorThreshold {
	switch thresholdType {
	case SensorThresholdType_LNR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.LNR,
			Raw:  full.LNR_Raw,
		}

	case SensorThresholdType_LCR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.LCR,
			Raw:  full.LCR_Raw,
		}

	case SensorThresholdType_LNC:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.LNC,
			Raw:  full.LNC_Raw,
		}

	case SensorThresholdType_UNC:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.UNC,
			Raw:  full.UNC_Raw,
		}

	case SensorThresholdType_UCR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.UCR,
			Raw:  full.UCR_Raw,
		}

	case SensorThresholdType_UNR:
		return SensorThreshold{
			Type: thresholdType,
			Mask: full.Mask.Threshold.UNR,
			Raw:  full.UNR_Raw,
		}
	}

	return SensorThreshold{
		Type: thresholdType,
	}
}

func (full *SDRFull) String() string {
	// For pure SDR record, there's no reading for a sensor, unless you use
	// GetSensorReading command to fetch it.
	return "" +
		fmt.Sprintf("Sensor ID              : %s (%#02x)\n", full.IDStringBytes, full.SensorNumber) +
		fmt.Sprintf("Generator ID           : %#04x (%s)\n", uint16(full.GeneratorID), full.GeneratorID.String()) +
		fmt.Sprintf("Entity ID              : %d.%d (%s)\n", uint8(full.SensorEntityID), uint8(full.SensorEntityInstance), full.SensorEntityID.String()) +
		fmt.Sprintf("Sensor Type            : %s (%#02x) (%s)\n", full.SensorType.String(), uint8(full.SensorType), full.SensorEventReadingType.SensorClass()) +
		fmt.Sprintf("Sensor Reading         : %.4f (+/- %d) %s\n", full.SensorValue, full.ReadingFactors.Tolerance, full.SensorUnit) +
		fmt.Sprintf("Sensor Status          : %s\n", full.SensorStatus) +
		fmt.Sprintf("Sensor Initialization  :%s", "\n") +
		fmt.Sprintf("  Settable             : %v\n", full.SensorInitialization.Settable) +
		fmt.Sprintf("  Scanning             : %v\n", full.SensorInitialization.InitScanning) +
		fmt.Sprintf("  Events               : %v\n", full.SensorInitialization.InitEvents) +
		fmt.Sprintf("  Hysteresis           : %v\n", full.SensorInitialization.InitHysteresis) +
		fmt.Sprintf("  Sensor Type          : %v\n", full.SensorInitialization.InitSensorType) +
		fmt.Sprintf("Default State          :%s", "\n") +
		fmt.Sprintf("  Event Generation     : %s\n", formatBool(full.SensorInitialization.EventGenerationEnabled, "enabled", "disabled")) +
		fmt.Sprintf("  Scanning             : %s\n", formatBool(full.SensorInitialization.SensorScanningEnabled, "enabled", "disabled")) +
		fmt.Sprintf("Sensor Capabilities    :%s", "\n") +
		fmt.Sprintf("  Auto Re-arm          : %s\n", formatBool(full.SensorCapabilities.AutoRearm, "yes(auto)", "no(manual)")) +
		fmt.Sprintf("  Hysteresis Support   : %s\n", full.SensorCapabilities.HysteresisAccess.String()) +
		fmt.Sprintf("  Threshold Access     : %s\n", full.SensorCapabilities.ThresholdAccess) +
		fmt.Sprintf("  Ev Message Control   : %s\n", full.SensorCapabilities.EventMessageControl) +
		fmt.Sprintf("Mask                   :%s", "\n") +
		fmt.Sprintf("  Readable Thresholds  : %s\n", strings.Join(full.Mask.ReadableThresholds().Strings(), " ")) +
		fmt.Sprintf("  Settable Thresholds  : %s\n", strings.Join(full.Mask.SettableThresholds().Strings(), " ")) +
		fmt.Sprintf("  Threshold Read Mask  : %s\n", strings.Join(full.Mask.StatusReturnedThresholds().Strings(), " ")) +
		fmt.Sprintf("  Assertions Enabled   : %s\n", strings.Join(full.Mask.SupportedThresholdEvents().FilterAssert().Strings(), " ")) +
		fmt.Sprintf("  Deassertions Enabled : %s\n", strings.Join(full.Mask.SupportedThresholdEvents().FilterDeassert().Strings(), " ")) +
		fmt.Sprintf("Nominal Reading        : %s\n", full.ReadingStr(full.NominalReadingRaw, full.NominalReadingSpecified)) +
		fmt.Sprintf("Normal Minimum         : %s\n", full.ReadingStr(full.NormalMinRaw, full.NormalMinSpecified)) +
		fmt.Sprintf("Normal Maximum         : %s\n", full.ReadingStr(full.NormalMaxRaw, full.NormalMaxSpecified)) +
		fmt.Sprintf("Lower Non-Recoverable  : %s\n", full.ThresholdValueStr(SensorThresholdType_LNR)) +
		fmt.Sprintf("Lower Critical         : %s\n", full.ThresholdValueStr(SensorThresholdType_LCR)) +
		fmt.Sprintf("Lower Non-Critical     : %s\n", full.ThresholdValueStr(SensorThresholdType_LNC)) +
		fmt.Sprintf("Upper Non-Critical     : %s\n", full.ThresholdValueStr(SensorThresholdType_UNC)) +
		fmt.Sprintf("Upper Critical         : %s\n", full.ThresholdValueStr(SensorThresholdType_UCR)) +
		fmt.Sprintf("Upper Non-Recoverable  : %s\n", full.ThresholdValueStr(SensorThresholdType_UNR)) +
		fmt.Sprintf("Positive Hysteresis    : %s\n", full.HysteresisStr(full.PositiveHysteresisRaw)) +
		fmt.Sprintf("Negative Hysteresis    : %s\n", full.HysteresisStr(full.NegativeHysteresisRaw)) +
		fmt.Sprintf("Minimum sensor range   : %s\n", full.ReadingMinStr()) +
		fmt.Sprintf("Maximum sensor range   : %s\n", full.ReadingMaxStr()) +
		fmt.Sprintf("SensorDirection        : %d\n", full.SensorDirection) +
		fmt.Sprintf("LinearizationFunc      : %s\n", full.LinearizationFunc) +
		fmt.Sprintf("Reading Factors        : %s\n", full.ReadingFactors)
}

func parseSDRFullSensor(data []byte, sdr *SDR) error {
	const SDRFullSensorMinSize int = 48 // plus the ID String Bytes (optional 16 bytes maximum)

	minSize := SDRFullSensorMinSize
	if len(data) < minSize {
		return ErrNotEnoughDataWith("sdr (full sensor) min size", len(data), minSize)
	}

	s := &SDRFull{}
	sdr.Full = s

	generatorID, _, _ := unpackUint16L(data, 5)
	s.GeneratorID = GeneratorID(generatorID)

	sensorNumber, _, _ := unpackUint8(data, 7)
	s.SensorNumber = SensorNumber(sensorNumber)

	b8, _, _ := unpackUint8(data, 8)
	s.SensorEntityID = EntityID(b8)

	b9, _, _ := unpackUint8(data, 9)
	s.SensorEntityInstance = EntityInstance(b9 & 0x7f)
	s.SensorEntityIsLogical = isBit7Set(b9)

	b10, _, _ := unpackUint8(data, 10)
	s.SensorInitialization = SensorInitialization{
		Settable:               isBit7Set(b10),
		InitScanning:           isBit6Set(b10),
		InitEvents:             isBit5Set(b10),
		InitThresholds:         isBit4Set(b10),
		InitHysteresis:         isBit3Set(b10),
		InitSensorType:         isBit2Set(b10),
		EventGenerationEnabled: isBit1Set(b10),
		SensorScanningEnabled:  isBit0Set(b10),
	}

	b11, _, _ := unpackUint8(data, 11)
	s.SensorCapabilities = SensorCapabilities{
		IgnoreSensorIfNoEntity: isBit7Set(b11),
		AutoRearm:              isBit6Set(b11),
		HysteresisAccess:       SensorHysteresisAccess((b11 & 0x3f) >> 4),
		ThresholdAccess:        SensorThresholdAccess((b11 & 0x0f) >> 2),
		EventMessageControl:    SensorEventMessageControl(b11 & 0x03),
	}

	sensorType, _, _ := unpackUint8(data, 12)
	s.SensorType = SensorType(sensorType)

	eventReadingType, _, _ := unpackUint8(data, 13)
	s.SensorEventReadingType = EventReadingType(eventReadingType)

	mask := Mask{}
	b14, _, _ := unpackUint16(data, 14)
	b16, _, _ := unpackUint16(data, 16)
	b18, _, _ := unpackUint16(data, 18)
	mask.ParseAssertLower(b14)
	mask.ParseDeassertUpper(b16)
	mask.ParseReading(b18)
	s.Mask = mask

	b20, _, _ := unpackUint8(data, 20)
	b21, _, _ := unpackUint8(data, 21)
	b22, _, _ := unpackUint8(data, 22)
	s.SensorUnit = SensorUnit{
		AnalogDataFormat: SensorAnalogUnitFormat((b20 & 0xc0) >> 6),
		RateUnit:         SensorRateUnit((b20 & 0x38) >> 4),
		ModifierRelation: SensorModifierRelation((b20 & 0x06) >> 2),
		Percentage:       isBit0Set(b20),
		BaseUnit:         SensorUnitType(b21),
		ModifierUnit:     SensorUnitType(b22),
	}

	b23, _, _ := unpackUint8(data, 23)
	s.LinearizationFunc = LinearizationFunc(b23)

	b24, _, _ := unpackUint8(data, 24)
	b25, _, _ := unpackUint8(data, 25)

	m := uint16(b25&0xc0)<<2 | uint16(b24)
	s.M = int16(twosComplement(uint32(m), 10))

	s.Tolerance = b25 & 0x3f

	b26, _, _ := unpackUint8(data, 26)
	b27, _, _ := unpackUint8(data, 27)
	b28, _, _ := unpackUint8(data, 28)

	b := uint16(b27&0xc0)<<2 | uint16(b26)
	s.B = int16(twosComplement(uint32(b), 10))

	s.Accuracy = uint16(b28&0xf0)<<2 | uint16(b27&0x3f)
	s.Accuracy_Exp = (b28 & 0x0c) >> 2
	s.SensorDirection = b28 & 0x03

	b29, _, _ := unpackUint8(data, 29)
	rExp := uint8((b29 & 0xf0) >> 4)
	s.R_Exp = int8(twosComplement(uint32(rExp), 4))

	bExp := uint8(b29 & 0x0f)
	s.B_Exp = int8(twosComplement(uint32(bExp), 4))

	b30, _, _ := unpackUint8(data, 30)
	s.NormalMinSpecified = isBit2Set(b30)
	s.NormalMaxSpecified = isBit1Set(b30)
	s.NominalReadingSpecified = isBit0Set(b30)

	s.NominalReadingRaw, _, _ = unpackUint8(data, 31)

	s.NormalMaxRaw, _, _ = unpackUint8(data, 32)
	s.NormalMinRaw, _, _ = unpackUint8(data, 33)
	s.SensorMaxReadingRaw, _, _ = unpackUint8(data, 34)
	s.SensorMinReadingRaw, _, _ = unpackUint8(data, 35)

	s.UNR_Raw, _, _ = unpackUint8(data, 36)
	s.UCR_Raw, _, _ = unpackUint8(data, 37)
	s.UNC_Raw, _, _ = unpackUint8(data, 38)

	s.LNR_Raw, _, _ = unpackUint8(data, 39)
	s.LCR_Raw, _, _ = unpackUint8(data, 40)
	s.LNC_Raw, _, _ = unpackUint8(data, 41)

	s.PositiveHysteresisRaw, _, _ = unpackUint8(data, 42)
	s.NegativeHysteresisRaw, _, _ = unpackUint8(data, 43)

	typeLength, _, _ := unpackUint8(data, 47)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())

	if len(data) < minSize+idStrLen {
		return ErrNotEnoughDataWith("sdr (full sensor)", len(data), minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)

	return nil
}

func (full *SDRFull) HasAnalogReading() bool {
	// Todo, logic is not clear.
	/*
	 * Per the IPMI Specification:
	 *	Only Full Threshold sensors are identified as providing
	 *	analog readings.
	 *
	 * But... HP didn't interpret this as meaning that "Only Threshold
	 *        Sensors" can provide analog readings.  So, HP packed analog
	 *        readings into some of their non-Threshold Sensor.   There is
	 *	  nothing that explicitly prohibits this in the spec, so if
	 *	  an Analog reading is available in a Non-Threshold sensor and
	 *	  there are units specified for identifying the reading then
	 *	  we do an analog conversion even though the sensor is
	 *	  non-Threshold.   To be safe, we provide this extension for
	 *	  HP.
	 *
	 */

	if full.SensorEventReadingType.IsThreshold() {
		// for threshold sensors
		return true
	}

	if full.SensorUnit.IsAnalog() {
		return true
	}

	// for non-threshold sensors, but the analog data format indicates analog.
	// this rarely exists, except HP.
	// Todo

	return false
}
