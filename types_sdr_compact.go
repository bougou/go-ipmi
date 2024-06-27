package ipmi

import (
	"fmt"
	"strings"
)

// 43.2 SDR Type 02h, Compact Sensor Record
type SDRCompact struct {
	//
	// Record KEY
	//

	GeneratorID  GeneratorID
	SensorNumber SensorNumber

	//
	// RECORD BODY
	//

	SensorEntityID       EntityID
	SensorEntityInstance EntityInstance
	// 0b = treat entity as a physical entity per Entity ID table
	// 1b = treat entity as a logical container entity. For example, if this bit is set,
	// and the Entity ID is "Processor", the container entity would be considered
	// to represent a logical "Processor Group" rather than a physical processor.
	// This bit is typically used in conjunction with an Entity Association record.
	SensorEntityIsLogical bool

	SensorInitialization SensorInitialization

	SensorCapabilities SensorCapabilities

	SensorType             SensorType
	SensorEventReadingType EventReadingType

	Mask Mask

	SensorUnit SensorUnit

	// SensorValue is not stored in SDR intrinsically, this field is set by `enhanceSDR`
	// It is fetched by IPMI command GetSensorReading and aligned/converted to SensorUnit based.
	SensorValue float64

	// SensorStatus is not stored in SDR intrinsically, this field is set by `enhanceSDR`
	SensorStatus string

	// Sensor Direction. Indicates whether the sensor is monitoring an input or
	// output relative to the given Entity. E.g. if the sensor is monitoring a
	// current, this can be used to specify whether it is an input voltage or an
	// output voltage.
	// 00b = unspecified / not applicable
	// 01b = input
	// 10b = output
	// 11b = reserved
	SensorDirection uint8

	EntityInstanceSharing uint8

	// Positive hysteresis is defined as the unsigned number of counts that are
	// subtracted from the raw threshold values to create the "re-arm" point for all
	// positive-going thresholds on the sensor. 0 indicates that there is no hysteresis on
	// positive-going thresholds for this sensor. Hysteresis values are given as raw
	// counts. That is, to find the degree of hysteresis in units, the value must be
	// converted using the "y=Mx+B" formula.
	//
	// compact SDR can have pos/neg hysteresis, but they cannot be analog!
	PositiveHysteresisRaw uint8

	// Negative hysteresis is defined as the unsigned number of counts that are added
	// to the raw threshold value to create the "re-arm" point for all negative-going
	// thresholds on the sensor. 0 indicates that there is no hysteresis on negative-going
	// thresholds for this sensor.
	//
	// compact SDR can have pos/neg hysteresis, but they cannot be analog!
	NegativeHysteresisRaw uint8

	IDStringTypeLength TypeLength // Sensor ID String Type/Length Code
	IDStringBytes      []byte     // Sensor ID String bytes.
}

func (compact *SDRCompact) String() string {

	return fmt.Sprintf(`Sensor ID              : %s (%#02x)
Generator             : %d
Entity ID             : %d.%d (%s)
Sensor Type (%s) : %s (%#02x)
Sensor Reading        : %.3f %s
Sensor Status         : %s
Sensor Initialization :
  Settable            : %v
  Scanning            : %v
  Events              : %v
  Hysteresis          : %v
  Sensor Type         : %v
  Default State:
    Event Generation  : %s
    Scanning          : %s
Sensor Capabilities   :
  Auto Re-arm         : %s
  Hysteresis Support  : %s
  Threshold Access    : %s
  Ev Message Control  : %s
Mask                  :
  Readable Thresholds : %s
  Settable Thresholds : %s
  Threshold Read Mask : %s
  Assertions Enabled  : %s
  Deassertions Enabled: %s
Positive Hysteresis   : %#02x
Negative Hysteresis   : %#02x`,
		string(compact.IDStringBytes), compact.SensorNumber,
		compact.GeneratorID,
		uint8(compact.SensorEntityID), uint8(compact.SensorEntityInstance), compact.SensorEntityID.String(),
		compact.SensorEventReadingType.SensorClass(), compact.SensorType.String(), uint8(compact.SensorType),
		compact.SensorValue, compact.SensorUnit,
		compact.SensorStatus,
		compact.SensorInitialization.Settable,
		compact.SensorInitialization.InitScanning,
		compact.SensorInitialization.InitEvents,
		compact.SensorInitialization.InitHysteresis,
		compact.SensorInitialization.InitSensorType,
		formatBool(compact.SensorInitialization.EventGenerationEnabled, "enabled", "disabled"),
		formatBool(compact.SensorInitialization.SensorScanningEnabled, "enabled", "disabled"),
		formatBool(compact.SensorCapabilities.AutoRearm, "yes(auto)", "no(manual)"),
		compact.SensorCapabilities.HysteresisAccess.String(),
		compact.SensorCapabilities.ThresholdAccess,
		compact.SensorCapabilities.EventMessageControl,
		strings.Join(compact.Mask.ReadableThresholds().Strings(), " "),
		strings.Join(compact.Mask.SettableThresholds().Strings(), " "),
		strings.Join(compact.Mask.StatusReturnedThresholds().Strings(), " "),
		strings.Join(compact.Mask.SupportedThresholdEvents().FilterAssert().Strings(), " "),
		strings.Join(compact.Mask.SupportedThresholdEvents().FilterDeassert().Strings(), " "),
		compact.PositiveHysteresisRaw,
		compact.NegativeHysteresisRaw,
	)

	//  Assertions Enabled    : Critical Interrupt
	//                          [PCI PERR]
	//                          [PCI SERR]
	//                          [Bus Correctable error]
	//                          [Bus Uncorrectable error]
	//                          [Bus Fatal Error]
	//  Deassertions Enabled  : Critical Interrupt
	//                          [PCI PERR]
	//                          [PCI SERR]
	//                          [Bus Correctable error]
	//                          [Bus Uncorrectable error]
	//                          [Bus Fatal Error]
	//  OEM                   : 0
}

func (record *SDRCompact) PositiveHysteresis() (raw uint8, valid bool) {
	raw = record.PositiveHysteresisRaw
	if raw == 0x00 || raw == 0xff {
		valid = false
	} else {
		valid = true
	}
	return
}

func (record *SDRCompact) NegativeHysteresis() (raw uint8, valid bool) {
	raw = record.NegativeHysteresisRaw
	if raw == 0x00 || raw == 0xff {
		valid = false
	} else {
		valid = true
	}
	return
}

func parseSDRCompactSensor(data []byte, sdr *SDR) error {
	const SDRCompactSensorMinSize int = 32 // plus the ID String Bytes (optional 16 bytes maximum)

	minSize := SDRCompactSensorMinSize
	if len(data) < minSize {
		return ErrNotEnoughDataWith("sdr (compact sensor) min size", len(data), minSize)
	}

	s := &SDRCompact{}
	sdr.Compact = s

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
		IgnoreWithEntity:    isBit7Set(b11),
		AutoRearm:           isBit6Set(b11),
		HysteresisAccess:    SensorHysteresisAccess((b11 & 0x3f) >> 4),
		ThresholdAccess:     SensorThresholdAccess((b11 & 0x0f) >> 2),
		EventMessageControl: SensorEventMessageControl(b11 & 0x03),
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

	s.PositiveHysteresisRaw, _, _ = unpackUint8(data, 25)
	s.NegativeHysteresisRaw, _, _ = unpackUint8(data, 26)

	typeLength, _, _ := unpackUint8(data, 31)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return ErrNotEnoughDataWith("sdr (compact sensor)", len(data), minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}
