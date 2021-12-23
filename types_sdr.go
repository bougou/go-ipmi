package ipmi

// 43. Sensor Data Record Formats
type SDR struct {
	//
	// Record Header
	//

	RecordID     uint16
	SDRVersion   uint8         // The version number of the SDR specification.
	RecordType   SDRRecordType // A number representing the type of the record. E.g. 01h = 8-bit Sensor with Thresholds.
	RecordLength uint8         // Number of bytes of data following the Record Length field.

	//
	// Record KEY FIELDS
	//

	// The Record 'Key' Fields are a set of fields that together are unique amongst instances of a given record type.
	// The Record Key bytes shall be contiguous and follow the Record Header.
	// The number of bytes that make up the Record Key field may vary according to record type.

	// [7:1] - 7-bit I2C Slave Address, or 7-bit system software ID[2]
	// [0] - 0b = ID is IPMB Slave Address, 1b = system software ID
	SensorOwnerID uint8

	SensorOwnerLUN uint8
	SensorNumber   uint8 // Unique number identifying the sensor behind a given slave address and LUN. Code FFh reserved.

	//
	// RECORD BODY
	//

	SensorEntityID       uint8
	SensorEntityInstance uint8
	SensorInitialization uint8
	SensorCapabilitites  uint8
	SensorType           uint8
	SensorEventType      uint8

	SensorUnits1  uint8
	SensorUnits2  uint8
	SensorUnits3  uint8
	Linearization uint8

	// Sensor Direction. Indicates whether the sensor is monitoring an input or
	// output relative to the given Entity. E.g. if the sensor is monitoring a
	// current, this can be used to specify whether it is an input voltage or an
	// output voltage.
	// 00b = unspecified / not applicable
	// 01b = input
	// 10b = output
	// 11b = reserved
	SensorDirection uint8

	NormalMinSpecifiedFlag  bool
	NormalMaxSpecifiedFlag  bool
	NominalReadingSpecified bool
	NominalReading          uint8
	NormalMaximum           uint8
	NormalMinimum           uint8

	SensorMaximumReading uint8
	SensorMinimumReading uint8

	UpperNonRecoverableThreshold uint8
	UpperCriticalThreshold       uint8
	UpperNonCriticalThreshold    uint8

	LowerNonRecoverableThreshold uint8
	LowerCriticalThreshold       uint8
	LowerNonCriticalThreshold    uint8

	PositiveGoingThresholdHysteresisValue uint8
	NegativeGoingThresholdHysteresisValue uint8

	IDLength uint8

	IDString string
}

type SDRRecordType uint8

const (
	// section 43

	SDRRecordTypeFullSensor                        SDRRecordType = 0x01
	SDRRecordTypeCompactSensor                     SDRRecordType = 0x02
	SDRRecordTypeEventOnly                         SDRRecordType = 0x03
	SDRRecordTypeEntityAssociation                 SDRRecordType = 0x08
	SDRRecordTypeDeviceRelativeEntityAssociation   SDRRecordType = 0x09
	SDRRecordTypeGenericLocator                    SDRRecordType = 0x10
	SDRRecordTypeFRUDeviceLocator                  SDRRecordType = 0x11
	SDRRecordTypeManagementControllerDeviceLocator SDRRecordType = 0x12
	SDRRecordTypeManagementControllerConfirmation  SDRRecordType = 0x13
	SDRRecordTypeBMCMessageChannelInfo             SDRRecordType = 0x14
	SDRRecordTypeOEM                               SDRRecordType = 0xC0
)

func (s SDRRecordType) String() string {
	return map[SDRRecordType]string{
		0x01: "Full Sensor Record",
		0x02: "Compact Sensor Record",
		0x03: "Event Only Record",
		0x08: "Entity Association Record",
		0x09: "Device-relative Entity Association Record",
		0x10: "Generic Device Locator Record",
		0x11: "FRU Device Locator Record",
		0x12: "Management Controller Device Locator Record",
		0x13: "Management Controller Confirmation Record",
		0x14: "BMC Message Channel Info Record",
		0xc0: "OEM Record",
	}[s]
}

const (
	// SensorStatusOK means okay (the sensor is present and operating correctly)
	SensorStatusOK = "ok"

	// SensorStatusNoSensor means no sensor (corresponding reading will say disabled or Not Readable)
	SensorStatusNoSensor = "ns"

	// SensorStatusNonCritical means non-critical error regarding the sensor
	SensorStatusNonCritical = "nc"

	// SensorStatusCritical means critical error regarding the sensor
	SensorStatusCritical = "cr"

	// SensorStatusNonRecoverable means non-recoverable error regarding the sensor
	SensorStatusNonRecoverable = "nr"
)
