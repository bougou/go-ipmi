package ipmi

import (
	"fmt"
	"strings"
)

// section 43
type SDRRecordType uint8

const (
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
	SDRRecordTypeOEM                               SDRRecordType = 0xc0

	SDRRecordHeaderSize                         int = 5
	SDRFullSensorMinSize                        int = 48 // plus the ID String Bytes (optional 16 bytes maximum)
	SDRCompactSensorMinSize                     int = 32 // plus the ID String Bytes (optional 16 bytes maximum)
	SDREventOnlyMinSize                         int = 17
	SDREntityAssociationSize                    int = 16
	SDRDeviceRelativeEntityAssociationSize          = 32
	SDRGenericLocatorMinSize                        = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	SDRFRUDeviceLocatorMinSize                      = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	SDRManagementControllerDeviceLocatorMinSize     = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	SDRManagementControllerConfirmationSize         = 32
	SDRBMCMessageChannelInfoSize                    = 16
	SDROEMMinSize                                   = 8
	SDROEMMaxSize                                   = 64 // OEM defined records are limited to a maximum of 64 bytes, including the header
)

// 43.6 SDR Type 0Ah:0Fh - Reserved Records
// This range and all other unspecified SDR Type values are reserved.
var sdrRecordTypeMap = map[SDRRecordType]string{
	0x01: "Full",
	0x02: "Compact",
	0x03: "Event",
	0x08: "Entity Assoc",
	0x09: "Device Entity Assoc",
	0x10: "Generic Device Loc",
	0x11: "FRU Device Loc",
	0x12: "MC Device Loc", // MC: Management Controller
	0x13: "MC Confirmation",
	0x14: "BMC Msg Channel Info",
	0xc0: "OEM",
}

func (sdrRecordType SDRRecordType) String() string {
	s, ok := sdrRecordTypeMap[sdrRecordType]
	if !ok {
		return "Reserved"
	}
	return s
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

type SDRHeader struct {
	RecordID     uint16
	SDRVersion   uint8         // The version number of the SDR specification.
	RecordType   SDRRecordType // A number representing the type of the record. E.g. 01h = 8-bit Sensor with Thresholds.
	RecordLength uint8         // Number of bytes of data following the Record Length field.
}

// 43. Sensor Data Record Formats
type SDR struct {
	// NextRecordID should be filled by ParseSDR function.
	NextRecordID uint16

	RecordHeader *SDRHeader

	Full                        *SDRFull
	Compact                     *SDRCompact
	EventOnly                   *SDREventOnly
	EntityAssociation           *SDREntityAssociation
	DeviceRelative              *SDRDeviceRelative
	GenericDeviceLocator        *SDRGenericDeviceLocator
	FRUDeviceLocator            *SDRFRUDeviceLocator
	MgmtControllerDeviceLocator *SDRMgmtControllerDeviceLocator
	MgmtControllerConfirmation  *SDRMgmtControllerConfirmation
	BMCChannelInfo              *SDRBMCChannelInfo
	OEM                         *SDROEM
	Reserved                    *SDRReserved
}

func (sdr *SDR) SensorNumber() uint8 {
	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return sdr.Full.SensorNumber
	case SDRRecordTypeCompactSensor:
		return sdr.Compact.SensorNumber
	case SDRRecordTypeEventOnly:
		return sdr.EventOnly.SensorNumber
	case SDRRecordTypeEntityAssociation:
		return 0
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		return 0
	case SDRRecordTypeGenericLocator:
		return 0
	case SDRRecordTypeFRUDeviceLocator:
		return 0
	case SDRRecordTypeManagementControllerDeviceLocator:
		return 0
	case SDRRecordTypeManagementControllerConfirmation:
		return 0
	case SDRRecordTypeOEM:
		return 0
	default:
		return 0
	}
}

func (sdr *SDR) SensorName() string {
	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return string(sdr.Full.IDStringBytes)
	case SDRRecordTypeCompactSensor:
		return string(sdr.Compact.IDStringBytes)
	case SDRRecordTypeEventOnly:
		return string(sdr.EventOnly.IDStringBytes)
	case SDRRecordTypeEntityAssociation:
		return "N/A"
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		return "N/A"
	case SDRRecordTypeGenericLocator:
		return string(sdr.GenericDeviceLocator.DeviceIDString)
	case SDRRecordTypeFRUDeviceLocator:
		return string(sdr.FRUDeviceLocator.DeviceIDBytes)
	case SDRRecordTypeManagementControllerDeviceLocator:
		return string(sdr.MgmtControllerDeviceLocator.DeviceIDBytes)
	case SDRRecordTypeManagementControllerConfirmation:
		return "N/A"
	case SDRRecordTypeOEM:
		return "N/A"
	default:
		return "N/A"
	}
}

func (sdr *SDR) GeneratorID() uint16 {
	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return sdr.Full.GeneratorID
	case SDRRecordTypeCompactSensor:
		return sdr.Compact.GeneratorID
	case SDRRecordTypeEventOnly:
		return sdr.EventOnly.GeneratorID
	case SDRRecordTypeEntityAssociation:
		return 0
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		return 0
	case SDRRecordTypeGenericLocator:
		return 0
	case SDRRecordTypeFRUDeviceLocator:
		return 0
	case SDRRecordTypeManagementControllerDeviceLocator:
		return 0
	case SDRRecordTypeManagementControllerConfirmation:
		return 0
	case SDRRecordTypeOEM:
		return 0
	default:
		return 0
	}
}

// 43.1 SDRFull Type 01h, Full Sensor Record
type SDRFull struct {
	//
	// Record KEY
	//

	// The Record 'Key' Fields are a set of fields that together are unique amongst instances of a given record type.
	// The Record Key bytes shall be contiguous and follow the Record Header.
	// The number of bytes that make up the Record Key field may vary according to record type.

	GeneratorID  uint16
	SensorNumber uint8

	//
	// RECORD BODY
	//

	SensorEntityID         uint8
	SensorEntityInstance   uint8
	SensorInitialization   uint8
	SensorCapabilitites    uint8
	SensorType             uint8
	SensorEventReadingType EventReadingType

	AssertionEventLowerThresholdReadingMask   AssertionEventLowerThresholdReadingMask
	DeassertionEventUpperThresholdReadingMask DeassertionEventUpperThresholdReadingMask
	DiscreteSettableReadableMask              DiscreteSettableReadableMask

	SensorUnits1 uint8
	SensorUnits2 uint8 // Base Unit [7:0] - Units Type code: See Table 43-, Sensor Unit Type Codes
	SensorUnits3 uint8 // Modifier Unit [7:0] - Units Type code, 00h if unused

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

	Linearization uint8
	M             uint8
	MTolerance    uint8
	B             uint8
	BAccuracy     uint8

	RexpBexp uint8

	AnalogCharacteristicsFlags uint8
	NormalMinSpecifiedFlag     bool
	NormalMaxSpecifiedFlag     bool
	NominalReadingSpecified    bool

	NominalReading uint8
	NormalMaximum  uint8
	NormalMinimum  uint8

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

	IDStringTypeLength TypeLength
	IDStringBytes      []byte
}

func parseSDRFullSensor(data []byte, sdr *SDR) error {
	minSize := SDRFullSensorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (full sensor) data must be longer than %d", minSize)
	}

	s := &SDRFull{}
	sdr.Full = s

	s.GeneratorID, _, _ = unpackUint16L(data, 5)
	s.SensorNumber, _, _ = unpackUint8(data, 7)

	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorInitialization, _, _ = unpackUint8(data, 10)
	s.SensorCapabilitites, _, _ = unpackUint8(data, 11)
	s.SensorType, _, _ = unpackUint8(data, 12)

	eventReadingType, _, _ := unpackUint8(data, 13)
	s.SensorEventReadingType = EventReadingType(eventReadingType)

	b1516, _, _ := unpackUint16(data, 14)
	s.AssertionEventLowerThresholdReadingMask = parseAssertionEventLowerThresholdReadingMask(b1516)

	b1718, _, _ := unpackUint16(data, 16)
	s.DeassertionEventUpperThresholdReadingMask = parseDeassertionEventUpperThresholdReadingMask(b1718)

	b1920, _, _ := unpackUint16(data, 18)
	s.DiscreteSettableReadableMask = parseDiscreteSettableReadableMask(b1920)

	s.SensorUnits1, _, _ = unpackUint8(data, 20)
	s.SensorUnits1, _, _ = unpackUint8(data, 21)
	s.SensorUnits3, _, _ = unpackUint8(data, 22)

	s.Linearization, _, _ = unpackUint8(data, 23)

	s.NominalReading, _, _ = unpackUint8(data, 31)

	s.NormalMaximum, _, _ = unpackUint8(data, 32)
	s.NormalMinimum, _, _ = unpackUint8(data, 33)
	s.SensorMaximumReading, _, _ = unpackUint8(data, 34)
	s.SensorMinimumReading, _, _ = unpackUint8(data, 35)

	s.UpperNonRecoverableThreshold, _, _ = unpackUint8(data, 36)
	s.UpperCriticalThreshold, _, _ = unpackUint8(data, 37)
	s.UpperNonCriticalThreshold, _, _ = unpackUint8(data, 38)

	s.LowerNonRecoverableThreshold, _, _ = unpackUint8(data, 39)
	s.LowerCriticalThreshold, _, _ = unpackUint8(data, 40)
	s.LowerNonCriticalThreshold, _, _ = unpackUint8(data, 41)

	s.PositiveGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 42)
	s.NegativeGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 43)

	typeLength, _, _ := unpackUint8(data, 47)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())

	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (full sensor) data must be longer than %d", minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)

	return nil
}

// 43.2 SDR Type 02h, Compact Sensor Record
type SDRCompact struct {
	//
	// Record KEY
	//

	GeneratorID  uint16
	SensorNumber uint8
	//
	// RECORD BODY
	//

	SensorEntityID         uint8
	SensorEntityInstance   uint8
	SensorInitialization   uint8
	SensorCapabilitites    uint8
	SensorType             uint8
	SensorEventReadingType uint8

	AssertionEventLowerThresholdReadingMask   AssertionEventLowerThresholdReadingMask
	DeassertionEventUpperThresholdReadingMask DeassertionEventUpperThresholdReadingMask
	DiscreteSettableReadableMask              DiscreteSettableReadableMask

	SensorUnits1 uint8
	SensorUnits2 uint8 // Base Unit [7:0] - Units Type code: See Table 43-, Sensor Unit Type Codes
	SensorUnits3 uint8 // Modifier Unit [7:0] - Units Type code, 00h if unused

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

	PositiveGoingThresholdHysteresisValue uint8
	NegativeGoingThresholdHysteresisValue uint8

	IDStringTypeLength TypeLength // Sensor ID String Type/Length Code
	IDStringBytes      []byte     // Sensor ID String bytes.
}

func parseSDRCompactSensor(data []byte, sdr *SDR) error {
	minSize := SDRCompactSensorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (compact sensor) data must be longer than %d", minSize)
	}

	s := &SDRCompact{}
	sdr.Compact = s

	s.GeneratorID, _, _ = unpackUint16L(data, 5)
	s.SensorNumber, _, _ = unpackUint8(data, 7)

	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorInitialization, _, _ = unpackUint8(data, 10)
	s.SensorCapabilitites, _, _ = unpackUint8(data, 11)
	s.SensorType, _, _ = unpackUint8(data, 12)
	s.SensorEventReadingType, _, _ = unpackUint8(data, 13)

	b1516, _, _ := unpackUint16(data, 14)
	s.AssertionEventLowerThresholdReadingMask = parseAssertionEventLowerThresholdReadingMask(b1516)

	b1718, _, _ := unpackUint16(data, 16)
	s.DeassertionEventUpperThresholdReadingMask = parseDeassertionEventUpperThresholdReadingMask(b1718)

	b1920, _, _ := unpackUint16(data, 18)
	s.DiscreteSettableReadableMask = parseDiscreteSettableReadableMask(b1920)

	s.SensorUnits1, _, _ = unpackUint8(data, 20)
	s.SensorUnits1, _, _ = unpackUint8(data, 21)
	s.SensorUnits3, _, _ = unpackUint8(data, 22)

	s.PositiveGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 25)
	s.NegativeGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 26)

	typeLength, _, _ := unpackUint8(data, 31)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (compact sensor) data must be longer than %d", minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.3 SDR Type 03h, Event-Only Record
type SDREventOnly struct {
	//
	// Record KEY
	//

	GeneratorID  uint16
	SensorNumber uint8 // Unique number identifying the sensor behind a given slave address and LUN. Code FFh reserved.

	//
	// RECORD BODY
	//

	SensorEntityID         uint8
	SensorEntityInstance   uint8
	SensorType             uint8
	SensorEventReadingType uint8
	SensorDirection        uint8
	EntityInstanceSharing  uint8

	IDStringTypeLength TypeLength
	IDStringBytes      []byte
}

func parseSDREventOnly(data []byte, sdr *SDR) error {
	minSize := SDREventOnlyMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize)
	}

	s := &SDREventOnly{}
	sdr.EventOnly = s

	s.GeneratorID, _, _ = unpackUint16L(data, 5)
	s.SensorNumber, _, _ = unpackUint8(data, 7)

	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorType, _, _ = unpackUint8(data, 10)
	s.SensorEventReadingType, _, _ = unpackUint8(data, 11)

	typeLength, _, _ := unpackUint8(data, 16)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.4 SDR Type 08h - Entity Association Record
type SDREntityAssociation struct {
	//
	// Record KEY
	//

	ContainerEntityID       uint8
	ContainerEntityInstance uint8

	// [7] - 0b = contained entities specified as list
	//       1b = contained entities specified as range
	ContainedEntitiesAsRange bool
	// [6] - Record Link
	//       0b = no linked Entity Association records
	//       1b = linked Entity Association records exist
	LinkedEntityAssiactionExist bool
	// [5] - 0b = Container entity and contained entities can be assumed absent
	//            if presence sensor for container entity cannot be accessed.
	//            This value is also used if the entity does not have a presence sensor.
	//       1b = Presence sensor should always be accessible. Software should consider
	//            it an error if the presence sensor associated with the container entity
	//            is not accessible. If a presence sensor is accessible, then the
	//            presence sensor can still report that the container entity is absent.
	PresenceSensorAlwaysAccessible bool

	ContaineredEntity1ID       uint8
	ContaineredEntity1Instance uint8

	//
	// RECORD BODY
	//

	ContaineredEntity2ID       uint8
	ContaineredEntity2Instance uint8
	ContaineredEntity3ID       uint8
	ContaineredEntity3Instance uint8
	ContaineredEntity4ID       uint8
	ContaineredEntity4Instance uint8
}

func parseSDREntityAssociation(data []byte, sdr *SDR) error {
	size := SDREntityAssociationSize
	if len(data) < size {
		return fmt.Errorf("sdr (entity association) data must be longer than %d", size)
	}

	s := &SDREntityAssociation{}
	sdr.EntityAssociation = s

	s.ContainerEntityID, _, _ = unpackUint8(data, 5)
	s.ContainerEntityInstance, _, _ = unpackUint8(data, 6)

	flag, _, _ := unpackUint8(data, 7)
	s.ContainedEntitiesAsRange = isBit7Set(flag)
	s.LinkedEntityAssiactionExist = isBit6Set(flag)
	s.PresenceSensorAlwaysAccessible = isBit5Set(flag)

	s.ContaineredEntity1ID, _, _ = unpackUint8(data, 8)
	s.ContaineredEntity1Instance, _, _ = unpackUint8(data, 9)
	s.ContaineredEntity2ID, _, _ = unpackUint8(data, 10)
	s.ContaineredEntity2Instance, _, _ = unpackUint8(data, 11)
	s.ContaineredEntity3ID, _, _ = unpackUint8(data, 12)
	s.ContaineredEntity3Instance, _, _ = unpackUint8(data, 13)
	s.ContaineredEntity4ID, _, _ = unpackUint8(data, 14)
	s.ContaineredEntity4Instance, _, _ = unpackUint8(data, 15)

	return nil
}

// 43.5 SDR Type 09h - Device-relative Entity Association Record
type SDRDeviceRelative struct {
	//
	// Record KEY
	//

	ContainerEntityID            uint8
	ContainerEntityInstance      uint8
	ContainerEntityDeviceAddress uint8
	ContainerEntityDeviceChannel uint8

	// [7] - 0b = contained entities specified as list
	//       1b = contained entities specified as range
	ContainedEntitiesAsRange bool
	// [6] - Record Link
	//       0b = no linked Entity Association records
	//       1b = linked Entity Association records exist
	LinkedEntityAssiactionExist bool
	// [5] - 0b = Container entity and contained entities can be assumed absent
	//            if presence sensor for container entity cannot be accessed.
	//            This value is also used if the entity does not have a presence sensor.
	//       1b = Presence sensor should always be accessible. Software should consider
	//            it an error if the presence sensor associated with the container entity
	//            is not accessible. If a presence sensor is accessible, then the
	//            presence sensor can still report that the container entity is absent.
	PresenceSensorAlwaysAccessible bool

	ContaineredEntity1DeviceAddress uint8
	ContaineredEntity1DeviceChannel uint8
	ContaineredEntity1ID            uint8
	ContaineredEntity1Instance      uint8

	//
	// RECORD BODY
	//

	ContaineredEntity2DeviceAddress uint8
	ContaineredEntity2DeviceChannel uint8
	ContaineredEntity2ID            uint8
	ContaineredEntity2Instance      uint8

	ContaineredEntity3DeviceAddress uint8
	ContaineredEntity3DeviceChannel uint8
	ContaineredEntity3ID            uint8
	ContaineredEntity3Instance      uint8

	ContaineredEntity4DeviceAddress uint8
	ContaineredEntity4DeviceChannel uint8
	ContaineredEntity4ID            uint8
	ContaineredEntity4Instance      uint8
}

func parseSDRDeviceRelativeEntityAssociation(data []byte, sdr *SDR) error {
	size := SDRDeviceRelativeEntityAssociationSize
	if len(data) < size {
		return fmt.Errorf("sdr (device-relative entity association) data must be longer than %d", size)
	}

	s := &SDRDeviceRelative{}
	sdr.DeviceRelative = s

	s.ContainerEntityID, _, _ = unpackUint8(data, 5)
	s.ContainerEntityInstance, _, _ = unpackUint8(data, 6)
	s.ContainerEntityDeviceAddress, _, _ = unpackUint8(data, 7)
	s.ContainerEntityDeviceChannel, _, _ = unpackUint8(data, 8)

	flag, _, _ := unpackUint8(data, 9)
	s.ContainedEntitiesAsRange = isBit7Set(flag)
	s.LinkedEntityAssiactionExist = isBit6Set(flag)
	s.PresenceSensorAlwaysAccessible = isBit5Set(flag)

	s.ContaineredEntity1DeviceAddress, _, _ = unpackUint8(data, 10)
	s.ContaineredEntity1DeviceChannel, _, _ = unpackUint8(data, 11)
	s.ContaineredEntity1ID, _, _ = unpackUint8(data, 12)
	s.ContaineredEntity1Instance, _, _ = unpackUint8(data, 13)

	s.ContaineredEntity2DeviceAddress, _, _ = unpackUint8(data, 14)
	s.ContaineredEntity2DeviceChannel, _, _ = unpackUint8(data, 15)
	s.ContaineredEntity2ID, _, _ = unpackUint8(data, 16)
	s.ContaineredEntity2Instance, _, _ = unpackUint8(data, 17)

	s.ContaineredEntity3DeviceAddress, _, _ = unpackUint8(data, 18)
	s.ContaineredEntity3DeviceChannel, _, _ = unpackUint8(data, 19)
	s.ContaineredEntity3ID, _, _ = unpackUint8(data, 20)
	s.ContaineredEntity3Instance, _, _ = unpackUint8(data, 21)

	s.ContaineredEntity4DeviceAddress, _, _ = unpackUint8(data, 22)
	s.ContaineredEntity4DeviceChannel, _, _ = unpackUint8(data, 23)
	s.ContaineredEntity4ID, _, _ = unpackUint8(data, 24)
	s.ContaineredEntity4Instance, _, _ = unpackUint8(data, 25)

	unpackBytes(data, 26, 6) // last 6 bytes reserved
	return nil
}

// 43.7 SDR Type 10h - Generic Device Locator Record
// This record is used to store the location and type information for devices
// on the IPMB or management controller private busses that are neither
// IPMI FRU devices nor IPMI management controllers.
//
// These devices can either be common non-intelligent I2C devices, special management ASICs, or proprietary controllers.
//
// IPMI FRU Devices and Management Controllers are located via the FRU Device Locator
// and Management Controller Device Locator records described in following sections.
type SDRGenericDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceAccessAddress uint8 // Slave address of management controller used to access device. 0000000b if device is directly on IPMB
	DeviceSlaveAddress  uint8
	ChannelNumber       uint8 // Channel number for management controller used to access device
	AccessLUN           uint8 // LUN for Master Write-Read command. 00b if device is non-intelligent device directly on IPMB.
	PrivateBusID        uint8 // Private bus ID if bus = Private. 000b if device directly on IPMB

	//
	// RECORD BODY
	//

	AddressSpan        uint8
	DeviceType         uint8
	DeviceTypeModifier uint8
	EntityID           uint8
	EntityInstance     uint8

	DeviceIDTypeLength TypeLength
	DeviceIDString     []byte // Short ID string for the device
}

func parseSDRGenericLocator(data []byte, sdr *SDR) error {
	minSize := SDRGenericLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize)
	}

	s := &SDRGenericDeviceLocator{}
	sdr.GenericDeviceLocator = s

	s.DeviceAccessAddress, _, _ = unpackUint8(data, 5)

	b, _, _ := unpackUint8(data, 6)
	s.DeviceSlaveAddress = b >> 1

	c, _, _ := unpackUint8(data, 7)
	s.ChannelNumber = ((b & 0x01) << 4) | (c >> 5)
	s.AccessLUN = (c & 0x1f) >> 3
	s.PrivateBusID = (c & 0x07)

	s.AddressSpan, _, _ = unpackUint8(data, 8)
	s.DeviceType, _, _ = unpackUint8(data, 10)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	idStrLen := int(s.DeviceIDTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize+idStrLen)
	}
	s.DeviceIDString, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.8 SDR Type 11h - FRU Device Locator Record
// 38. Accessing FRU Devices
type SDRFRUDeviceLocator struct {
	//
	// Record KEY
	//

	// Slave address of controller used to access device. 0000000b if device is directly on IPMB.
	// This field indicates whether the device is on a private bus or not.
	DeviceAccessAddress uint8

	FRUDeviceID        uint8 // For LOGICAL FRU DEVICE
	DeviceSlaveAddress uint8 // For non-intelligent FRU device

	IsLogicalFRUDevice bool
	AccessLUN          uint8
	PrivateBusID       uint8

	ChannelNumber uint8

	//
	// RECORD BODY
	//

	DeviceType         uint8
	DeviceTypeModifier uint8
	FRUEntityID        uint8
	FRUEntityInstance  uint8

	DeviceIDTypeLength TypeLength
	DeviceIDBytes      []byte // Short ID string for the FRU Device
}

func parseSDRFRUDeviceLocator(data []byte, sdr *SDR) error {
	minSize := SDRFRUDeviceLocatorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize)
	}

	s := &SDRFRUDeviceLocator{}
	sdr.FRUDeviceLocator = s

	s.DeviceAccessAddress, _, _ = unpackUint8(data, 5)

	b7, _, _ := unpackUint8(data, 6)
	s.FRUDeviceID = b7
	s.DeviceSlaveAddress = b7 >> 1

	b8, _, _ := unpackUint8(data, 7)
	s.IsLogicalFRUDevice = isBit7Set(b8)
	s.AccessLUN = (b8 & 0x1f) >> 3
	s.PrivateBusID = b8 & 0x07

	b9, _, _ := unpackUint8(data, 8)
	s.ChannelNumber = b9 >> 4

	s.DeviceType, _, _ = unpackUint8(data, 10)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.FRUEntityID, _, _ = unpackUint8(data, 12)
	s.FRUEntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	idStrLen := int(s.DeviceIDTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize+idStrLen)
	}
	s.DeviceIDBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.9 SDR Type 12h - Management Controller Device Locator Record
type SDRMgmtControllerDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8 // 7-bit I2C Slave Address[1] of device on channel
	ChannelNumber      uint8

	//
	// RECORD BODY
	//

	ACPISystemPowerStateNotificationRequired bool
	ACPIDevicePowerStateNotificationRequired bool
	ControllerLogsInitializationAgentErrors  bool
	LogInitializationAgentErrors             bool

	DeviceCap_ChassisDevice      bool // device functions as chassis device
	DeviceCap_Bridge             bool // Controller responds to Bridge NetFn command
	DeviceCap_IPMBEventGenerator bool // device generates event messages on IPMB
	DeviceCap_IPMBEventReceiver  bool // device accepts event messages from IPMB
	DeviceCap_FRUInventoryDevice bool // accepts FRU commands to FRU Device #0 at LUN 00b
	DeviceCap_SELDevice          bool // provides interface to SEL
	DeviceCap_SDRRepoDevice      bool // For BMC, indicates BMC provides interface to	1b = SDR Repository. For other controller, indicates controller accepts Device SDR commands
	DeviceCap_SensorDevice       bool // device accepts sensor commands

	EntityID       uint8
	EntityInstance uint8

	DeviceIDTypeLength TypeLength
	DeviceIDBytes      []byte
}

func parseSDRManagementControllerDeviceLocator(data []byte, sdr *SDR) error {
	minSize := SDRManagementControllerDeviceLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerDeviceLocator{}
	sdr.MgmtControllerDeviceLocator = s

	b6, _, _ := unpackUint8(data, 5)
	s.DeviceSlaveAddress = b6 >> 1

	b7, _, _ := unpackUint8(data, 6)
	s.ChannelNumber = b7

	b8, _, _ := unpackUint8(data, 7)
	s.ACPISystemPowerStateNotificationRequired = isBit7Set(b8)
	s.ACPIDevicePowerStateNotificationRequired = isBit6Set(b8)
	s.ControllerLogsInitializationAgentErrors = isBit3Set(b8)
	s.LogInitializationAgentErrors = isBit2Set(b8)

	b9, _, _ := unpackUint8(data, 8)
	s.DeviceCap_ChassisDevice = isBit7Set(b9)
	s.DeviceCap_Bridge = isBit6Set(b9)
	s.DeviceCap_IPMBEventGenerator = isBit5Set(b9)
	s.DeviceCap_IPMBEventReceiver = isBit4Set(b9)
	s.DeviceCap_FRUInventoryDevice = isBit3Set(b9)
	s.DeviceCap_SELDevice = isBit2Set(b9)
	s.DeviceCap_SDRRepoDevice = isBit1Set(b9)
	s.DeviceCap_SensorDevice = isBit0Set(b9)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	idStrLen := int(s.DeviceIDTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize+idStrLen)
	}
	s.DeviceIDBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.10 SDR Type 13h - Management Controller Confirmation Record
type SDRMgmtControllerConfirmation struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8 // 7-bit I2C Slave Address[1] of device on IPMB.
	DeviceID           uint8
	ChannelNumber      uint8
	DeviceRevision     uint8

	//
	// RECORD BODY
	//

	FirmwareMajorRevision uint8 // [6:0] - Major Firmware Revision, binary encoded.
	FirmwareMinorRevision uint8 // Minor Firmware Revision. BCD encoded.

	// IPMI Version from Get Device ID command. Holds IPMI Command Specification
	// Version. BCD encoded. 00h = reserved. Bits 7:4 hold the Least Significant digit of the
	// revision, while bits 3:0 hold the Most Significant bits. E.g. a value of 01h indicates
	// revision 1.0
	MajorIPMIVersion uint8
	MinorIPMIVersion uint8

	ManufacturerID uint32 // 3 bytes only
	ProductID      uint16
	DeviceGUID     []byte // 16 bytes
}

func parseSDRManagementControllerConfirmation(data []byte, sdr *SDR) error {
	minSize := SDRManagementControllerConfirmationSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller confirmation) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerConfirmation{}
	sdr.MgmtControllerConfirmation = s

	b6, _, _ := unpackUint8(data, 5)
	s.DeviceSlaveAddress = b6 >> 1

	s.DeviceID, _, _ = unpackUint8(data, 6)

	b8, _, _ := unpackUint8(data, 7)
	s.ChannelNumber = b8 >> 4
	s.DeviceRevision = b8 & 0x0f

	b9, _, _ := unpackUint8(data, 8)
	s.FirmwareMajorRevision = b9 & 0x7f

	s.FirmwareMinorRevision, _, _ = unpackUint8(data, 9)

	ipmiVersionBCD, _, _ := unpackUint8(data, 10)
	s.MajorIPMIVersion = ipmiVersionBCD & 0x0f
	s.MinorIPMIVersion = ipmiVersionBCD >> 4

	s.ManufacturerID, _, _ = unpackUint24L(data, 11)
	s.ProductID, _, _ = unpackUint16L(data, 14)
	s.DeviceGUID, _, _ = unpackBytes(data, 16, 16)
	return nil
}

// 43.11 SDR Type 14h - BMC Message Channel Info Record
type SDRBMCChannelInfo struct {
	//
	// NO Record KEY
	//

	//
	// RECORD BODY
	//

	Channel0 ChannelInfo
	Channel1 ChannelInfo
	Channel2 ChannelInfo
	Channel3 ChannelInfo
	Channel4 ChannelInfo
	Channel5 ChannelInfo
	Channel6 ChannelInfo
	Channel7 ChannelInfo

	MessagingInterruptType uint8

	EventMessageBufferInterruptType uint8
}

type ChannelInfo struct {
	TransmitSupported bool // false means  receive message queue access only
	MessageReceiveLUN uint8
	ChannelProtocol   uint8
}

func parseChannelInfo(b uint8) ChannelInfo {
	return ChannelInfo{
		TransmitSupported: isBit7Set(b),
		MessageReceiveLUN: (b & 0x7f) >> 4,
		ChannelProtocol:   b & 0x0f,
	}
}

func parseSDRBMCMessageChannelInfo(data []byte, sdr *SDR) error {
	minSize := SDRBMCMessageChannelInfoSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (bmc message channel info) data must be longer than %d", minSize)
	}

	s := &SDRBMCChannelInfo{}
	sdr.BMCChannelInfo = s

	s.Channel0 = parseChannelInfo(data[5])
	s.Channel1 = parseChannelInfo(data[6])
	s.Channel2 = parseChannelInfo(data[7])
	s.Channel3 = parseChannelInfo(data[8])
	s.Channel4 = parseChannelInfo(data[9])
	s.Channel5 = parseChannelInfo(data[10])
	s.Channel6 = parseChannelInfo(data[11])
	s.Channel7 = parseChannelInfo(data[12])

	s.MessagingInterruptType, _, _ = unpackUint8(data, 13)
	s.EventMessageBufferInterruptType, _, _ = unpackUint8(data, 14)
	return nil
}

// 43.12 SDR Type C0h - OEM Record
type SDROEM struct {
	//
	// NO Record KEY
	//

	//
	// RECORD BODY
	//

	ManufacturerID uint32 // 3 bytes only
	OEMData        []byte
}

func parseSDROEM(data []byte, sdr *SDR) error {
	minSize := SDROEMMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (bmc message channel info) data must be longer than %d", minSize)
	}

	s := &SDROEM{}
	sdr.OEM = s

	s.ManufacturerID, _, _ = unpackUint24L(data, 5)
	s.OEMData, _, _ = unpackBytesMost(data, 8, SDROEMMaxSize-8)
	return nil
}

// 43.6 SDR Type 0Ah:0Fh - Reserved Records
type SDRReserved struct {
}

// 43.15 Type/Length Byte Format
//
//  7:6 00 = Unicode
//      01 = BCD plus (see below)
//      10 = 6-bit ASCII, packed
//      11 = 8-bit ASCII + Latin 1.
//          At least two bytes of data must be present when this type is used.
//          Therefore, the length (number of data bytes) will be >1 if data is present,
//          0 if data is not present. A length of 1 is reserved.
//  5 reserved.
//  4:0 length of following data, in characters.
//      00000b indicates 'none following'.
//      11111b = reserved.
type TypeLength uint8

func (tl TypeLength) Type() string {
	typecode := (uint8(tl) & 0xc0) >> 6 // the highest 2 bits
	var s string
	switch typecode {
	case 0:
		s = "Unspecified"
	case 1:
		s = "BCD plus"
	case 2:
		s = "6-bit ASCII"
	case 3:
		s = "8-bit ASCII"
	}

	return s
}

func (tl TypeLength) Length() uint8 {
	typecode := (uint8(tl) & 0xc0) >> 6 // the highest 2 bits
	l := uint8(tl) & 0x3f               // the lowest 6 bits

	var size uint8
	switch typecode {
	case 0: /* 00b: binary/unspecified */
	case 1: /* 01b: BCD plus */
		/* hex dump or BCD -> 2x length */
		size = (l * 2)
	case 2: /* 10b: 6-bit ASCII packed */
		/* 4 chars per group of 1-3 bytes, round up to 4 bytes boundary */
		size = (l/3 + 1) * 4
	case 3: /* 11b: 8-bit ASCII + Latin 1 */
		/* no length adjustment */
		size = l
	}

	return size
}

type AssertionEventLowerThresholdReadingMask struct {
	// for non-threshold based sensors
	// Assertion Event Mask
	AssertionEvent14Supported bool
	AssertionEvent13Supported bool
	AssertionEvent12Supported bool
	AssertionEvent11Supported bool
	AssertionEvent10Supported bool
	AssertionEvent9Supported  bool
	AssertionEvent8Supported  bool
	AssertionEvent7Supported  bool
	AssertionEvent6Supported  bool
	AssertionEvent5Supported  bool
	AssertionEvent4Supported  bool
	AssertionEvent3Supported  bool
	AssertionEvent2Supported  bool
	AssertionEvent1Supported  bool
	AssertionEvent0Supported  bool

	// for threshold-based sensors
	// Lower Threshold Reading Mask
	// Indicates which lower threshold comparison status is returned via the Get Sensor Reading command.
	LowerNonRecoverableThresholdComparsion bool // LNR
	LowerCriticalThresholdComparsion       bool // LC
	LowerNonCriticalThresholdComparsion    bool // LNC
	// Threshold Assertion Event Mask
	Assertion_UNR_GoHigh_Supported bool
	Assertion_UNR_GoLow_Supported  bool
	Assertion_UC_GoHigh_Supported  bool
	Assertion_UC_GoLow_Supported   bool
	Assertion_UNC_GoHigh_Supported bool
	Assertion_UNC_GoLow_Supported  bool

	Assertion_LNR_GoHigh_Supported bool
	Assertion_LNR_GoLow_Supported  bool
	Assertion_LC_GoHigh_Supported  bool
	Assertion_LC_GoLow_Supported   bool
	Assertion_LNC_GoHigh_Supported bool
	Assertion_LNC_GoLow_Supported  bool
}

func parseAssertionEventLowerThresholdReadingMask(b uint16) AssertionEventLowerThresholdReadingMask {
	lsb := uint8(b & 0x0000ffff) // Least Significant Byte
	msb := uint8(b >> 8)         // Most Significant Byte

	return AssertionEventLowerThresholdReadingMask{
		AssertionEvent14Supported: isBit6Set(lsb),
		AssertionEvent13Supported: isBit5Set(lsb),
		AssertionEvent12Supported: isBit4Set(lsb),
		AssertionEvent11Supported: isBit3Set(lsb),
		AssertionEvent10Supported: isBit2Set(lsb),
		AssertionEvent9Supported:  isBit1Set(lsb),
		AssertionEvent8Supported:  isBit0Set(lsb),
		AssertionEvent7Supported:  isBit7Set(msb),
		AssertionEvent6Supported:  isBit6Set(msb),
		AssertionEvent5Supported:  isBit5Set(msb),
		AssertionEvent4Supported:  isBit4Set(msb),
		AssertionEvent3Supported:  isBit3Set(msb),
		AssertionEvent2Supported:  isBit2Set(msb),
		AssertionEvent1Supported:  isBit1Set(msb),
		AssertionEvent0Supported:  isBit0Set(msb),

		LowerNonRecoverableThresholdComparsion: isBit6Set(lsb),
		LowerCriticalThresholdComparsion:       isBit5Set(lsb),
		LowerNonCriticalThresholdComparsion:    isBit4Set(lsb),
		Assertion_UNR_GoHigh_Supported:         isBit3Set(lsb),
		Assertion_UNR_GoLow_Supported:          isBit2Set(lsb),
		Assertion_UC_GoHigh_Supported:          isBit1Set(lsb),
		Assertion_UC_GoLow_Supported:           isBit0Set(lsb),
		Assertion_UNC_GoHigh_Supported:         isBit7Set(msb),
		Assertion_UNC_GoLow_Supported:          isBit6Set(msb),
		Assertion_LNR_GoHigh_Supported:         isBit5Set(msb),
		Assertion_LNR_GoLow_Supported:          isBit4Set(msb),
		Assertion_LC_GoHigh_Supported:          isBit3Set(msb),
		Assertion_LC_GoLow_Supported:           isBit2Set(msb),
		Assertion_LNC_GoHigh_Supported:         isBit1Set(msb),
		Assertion_LNC_GoLow_Supported:          isBit0Set(msb),
	}
}

type DeassertionEventUpperThresholdReadingMask struct {
	// for non-threshold based sensors
	// Assertion Event Mask
	DeassertionEvent14Supported bool
	DeassertionEvent13Supported bool
	DeassertionEvent12Supported bool
	DeassertionEvent11Supported bool
	DeassertionEvent10Supported bool
	DeassertionEvent9Supported  bool
	DeassertionEvent8Supported  bool
	DeassertionEvent7Supported  bool
	DeassertionEvent6Supported  bool
	DeassertionEvent5Supported  bool
	DeassertionEvent4Supported  bool
	DeassertionEvent3Supported  bool
	DeassertionEvent2Supported  bool
	DeassertionEvent1Supported  bool
	DeassertionEvent0Supported  bool

	// for threshold-based sensors
	// Lower Threshold Reading Mask
	// Indicates which upper threshold comparison status is returned via the Get Sensor Reading command.
	UpperNonRecoverableThresholdComparsion bool // UNR
	UpperCriticalThresholdComparsion       bool // UC
	UpperNonCriticalThresholdComparsion    bool // UNC
	// Threshold Assertion Event Mask
	Deassertion_UNR_GoHigh_Supported bool
	Deassertion_UNR_GoLow_Supported  bool
	Deassertion_UC_GoHigh_Supported  bool
	Deassertion_UC_GoLow_Supported   bool
	Deassertion_UNC_GoHigh_Supported bool
	Deassertion_UNC_GoLow_Supported  bool

	Deassertion_LNR_GoHigh_Supported bool
	Deassertion_LNR_GoLow_Supported  bool
	Deassertion_LC_GoHigh_Supported  bool
	Deassertion_LC_GoLow_Supported   bool
	Deassertion_LNC_GoHigh_Supported bool
	Deassertion_LNC_GoLow_Supported  bool
}

func parseDeassertionEventUpperThresholdReadingMask(b uint16) DeassertionEventUpperThresholdReadingMask {
	lsb := uint8(b & 0x0000ffff) // Least Significant Byte
	msb := uint8(b >> 8)         // Most Significant Byte
	return DeassertionEventUpperThresholdReadingMask{
		DeassertionEvent14Supported: isBit6Set(lsb),
		DeassertionEvent13Supported: isBit5Set(lsb),
		DeassertionEvent12Supported: isBit4Set(lsb),
		DeassertionEvent11Supported: isBit3Set(lsb),
		DeassertionEvent10Supported: isBit2Set(lsb),
		DeassertionEvent9Supported:  isBit1Set(lsb),
		DeassertionEvent8Supported:  isBit0Set(lsb),
		DeassertionEvent7Supported:  isBit7Set(msb),
		DeassertionEvent6Supported:  isBit6Set(msb),
		DeassertionEvent5Supported:  isBit5Set(msb),
		DeassertionEvent4Supported:  isBit4Set(msb),
		DeassertionEvent3Supported:  isBit3Set(msb),
		DeassertionEvent2Supported:  isBit2Set(msb),
		DeassertionEvent1Supported:  isBit1Set(msb),
		DeassertionEvent0Supported:  isBit0Set(msb),

		UpperNonRecoverableThresholdComparsion: isBit6Set(lsb),
		UpperCriticalThresholdComparsion:       isBit5Set(lsb),
		UpperNonCriticalThresholdComparsion:    isBit4Set(lsb),
		Deassertion_UNR_GoHigh_Supported:       isBit3Set(lsb),
		Deassertion_UNR_GoLow_Supported:        isBit2Set(lsb),
		Deassertion_UC_GoHigh_Supported:        isBit1Set(lsb),
		Deassertion_UC_GoLow_Supported:         isBit0Set(lsb),
		Deassertion_UNC_GoHigh_Supported:       isBit7Set(msb),
		Deassertion_UNC_GoLow_Supported:        isBit6Set(msb),
		Deassertion_LNR_GoHigh_Supported:       isBit5Set(msb),
		Deassertion_LNR_GoLow_Supported:        isBit4Set(msb),
		Deassertion_LC_GoHigh_Supported:        isBit3Set(msb),
		Deassertion_LC_GoLow_Supported:         isBit2Set(msb),
		Deassertion_LNC_GoHigh_Supported:       isBit1Set(msb),
		Deassertion_LNC_GoLow_Supported:        isBit0Set(msb),
	}
}

type DiscreteSettableReadableMask struct {
	// Reading Mask (for non-threshold based sensors)
	DiscreteReading14Supported bool
	DiscreteReading13Supported bool
	DiscreteReading12Supported bool
	DiscreteReading11Supported bool
	DiscreteReading10Supported bool
	DiscreteReading9Supported  bool
	DiscreteReading8Supported  bool
	DiscreteReading7Supported  bool
	DiscreteReading6Supported  bool
	DiscreteReading5Supported  bool
	DiscreteReading4Supported  bool
	DiscreteReading3Supported  bool
	DiscreteReading2Supported  bool
	DiscreteReading1Supported  bool
	DiscreteReading0Supported  bool

	// Settable Threshold Mask (for threshold-based sensors)
	UpperNonRecoverableThresholdSettable bool // UNR
	UpperCriticalThresholdSettable       bool // UC
	UpperNonCriticalThresholdSettable    bool // UNC
	LowerNonRecoverableThresholdSettable bool // LNR
	LowerCriticalThresholdSettable       bool // LC
	LowerNonCriticalThresholdSettable    bool // LNC
	// Readable Threshold Mask (for threshold-based sensors)
	UpperNonRecoverableThresholdReadable bool // UNR
	UpperCriticalThresholdReadable       bool // UC
	UpperNonCriticalThresholdReadable    bool // UNC
	LowerNonRecoverableThresholdReadable bool // LNR
	LowerCriticalThresholdReadable       bool // LC
	LowerNonCriticalThresholdReadable    bool // LNC
}

func parseDiscreteSettableReadableMask(b uint16) DiscreteSettableReadableMask {
	lsb := uint8(b & 0x0000ffff) // Least Significant Byte
	msb := uint8(b >> 8)         // Most Significant Byte
	return DiscreteSettableReadableMask{
		// Reading Mask (for non-threshold based sensors)
		DiscreteReading14Supported: isBit6Set(lsb),
		DiscreteReading13Supported: isBit5Set(lsb),
		DiscreteReading12Supported: isBit4Set(lsb),
		DiscreteReading11Supported: isBit3Set(lsb),
		DiscreteReading10Supported: isBit2Set(lsb),
		DiscreteReading9Supported:  isBit1Set(lsb),
		DiscreteReading8Supported:  isBit0Set(lsb),
		DiscreteReading7Supported:  isBit7Set(msb),
		DiscreteReading6Supported:  isBit6Set(msb),
		DiscreteReading5Supported:  isBit5Set(msb),
		DiscreteReading4Supported:  isBit4Set(msb),
		DiscreteReading3Supported:  isBit3Set(msb),
		DiscreteReading2Supported:  isBit2Set(msb),
		DiscreteReading1Supported:  isBit1Set(msb),
		DiscreteReading0Supported:  isBit0Set(msb),

		// Settable Threshold Mask (for threshold-based sensors)
		UpperNonRecoverableThresholdSettable: isBit5Set(lsb),
		UpperCriticalThresholdSettable:       isBit4Set(lsb),
		UpperNonCriticalThresholdSettable:    isBit3Set(lsb),
		LowerNonRecoverableThresholdSettable: isBit2Set(lsb),
		LowerCriticalThresholdSettable:       isBit1Set(lsb),
		LowerNonCriticalThresholdSettable:    isBit0Set(lsb),
		// Readable Threshold Mask (for threshold-based sensors)
		UpperNonRecoverableThresholdReadable: isBit5Set(msb),
		UpperCriticalThresholdReadable:       isBit4Set(msb),
		UpperNonCriticalThresholdReadable:    isBit3Set(msb),
		LowerNonRecoverableThresholdReadable: isBit2Set(msb),
		LowerCriticalThresholdReadable:       isBit1Set(msb),
		LowerNonCriticalThresholdReadable:    isBit0Set(msb),
	}
}

func ParseSDR(data []byte, nextRecordID uint16) (*SDR, error) {
	sdrHeader := &SDRHeader{}
	if len(data) < SDRRecordHeaderSize {
		return nil, fmt.Errorf("sdr data must be longer than %d", SDRRecordHeaderSize)
	}

	sdrHeader.RecordID, _, _ = unpackUint16L(data, 0)
	sdrHeader.SDRVersion, _, _ = unpackUint8(data, 2)
	recordType, _, _ := unpackUint8(data, 3)
	sdrHeader.RecordType = SDRRecordType(recordType)
	sdrHeader.RecordLength, _, _ = unpackUint8(data, 4)

	sdr := &SDR{
		RecordHeader: sdrHeader,
		NextRecordID: nextRecordID,
	}

	switch sdrHeader.RecordType {
	case SDRRecordTypeFullSensor:
		if err := parseSDRFullSensor(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRFullSensor failed, err: %s", err)
		}
	case SDRRecordTypeCompactSensor:
		if err := parseSDRCompactSensor(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRCompactSensor failed, err: %s", err)
		}
	case SDRRecordTypeEventOnly:
		if err := parseSDREventOnly(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDREventOnly failed, err: %s", err)
		}
	case SDRRecordTypeEntityAssociation:
		if err := parseSDREntityAssociation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDREntityAssociation failed, err: %s", err)
		}
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		if err := parseSDRDeviceRelativeEntityAssociation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRDeviceRelativeEntityAssociation failed, err: %s", err)
		}
	case SDRRecordTypeGenericLocator:
		if err := parseSDRGenericLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRGenericLocator failed, err: %s", err)
		}
	case SDRRecordTypeFRUDeviceLocator:
		if err := parseSDRFRUDeviceLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRFRUDeviceLocator failed, err: %s", err)
		}
	case SDRRecordTypeManagementControllerDeviceLocator:
		if err := parseSDRManagementControllerDeviceLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRManagementControllerDeviceLocator failed, err: %s", err)
		}
	case SDRRecordTypeManagementControllerConfirmation:
		if err := parseSDRManagementControllerConfirmation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRManagementControllerConfirmation failed, err: %s", err)
		}
	case SDRRecordTypeBMCMessageChannelInfo:
		if err := parseSDRBMCMessageChannelInfo(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRBMCMessageChannelInfo failed, err: %s", err)
		}
	case SDRRecordTypeOEM:
		if err := parseSDROEM(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDROEM failed, err: %s", err)
		}
	}

	return sdr, nil
}

func FormatSDRs(records []*SDR) string {
	var lines []string

	headers := []formatValue{
		fv("%-8s", "RecordID"),
		fv("%-10s", "RecordType"),
		fv("%-20s", "RecordTypeStr"),
		fv("%-11s", "GeneratorID"),
		fv("%-7s", "Sensor#"),
		fv("%-16s", "SensorName"),
	}

	lines = append(lines, formatValuesTable(headers))

	for _, sdr := range records {
		if sdr.RecordHeader == nil {
			return ""
		}

		recordID := sdr.RecordHeader.RecordID
		recordType := sdr.RecordHeader.RecordType
		switch recordType {
		case SDRRecordTypeFullSensor:
		case SDRRecordTypeCompactSensor:
		case SDRRecordTypeEventOnly:
		case SDRRecordTypeDeviceRelativeEntityAssociation:
		case SDRRecordTypeGenericLocator:
		case SDRRecordTypeFRUDeviceLocator:
		case SDRRecordTypeManagementControllerDeviceLocator:
		case SDRRecordTypeManagementControllerConfirmation:
		case SDRRecordTypeBMCMessageChannelInfo:
		case SDRRecordTypeOEM:
		default:
		}

		var generatorIDStr string
		if sdr.GeneratorID() == 0 {
			generatorIDStr = "N/A"
		} else {
			generatorIDStr = fmt.Sprintf("%#04x", sdr.GeneratorID())
		}

		content := []formatValue{
			fv("%-8s", fmt.Sprintf("%#02x", recordID)),
			fv("%-10s", fmt.Sprintf("%#02x", uint8(recordType))),
			fv("%-20s", recordType),
			fv("%-11s", generatorIDStr),
			fv("%-7s", fmt.Sprintf("%#02x", sdr.SensorNumber())),
			fv("%-16s", sdr.SensorName()),
		}
		lines = append(lines, formatValuesTable(content))
	}

	lines = append(lines, formatValuesTable(headers))

	return strings.Join(lines, "\n")
}
