package ipmi

import "fmt"

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

func (sdr *SDR) String() string {
	if sdr.RecordHeader == nil {
		return ""
	}

	// RecordID, RecordType, SensorName, SensorNumber
	format := "%#02x | %#02x | %-16s | %#02x"
	recordID := sdr.RecordHeader.RecordID
	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.Full.SensorName), sdr.Full.SensorNumber)
	case SDRRecordTypeCompactSensor:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.Compact.SensorName), sdr.Compact.SensorNumber)
	case SDRRecordTypeEventOnly:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.EventOnly.SensorName), sdr.EventOnly.SensorNumber)
	case SDRRecordTypeEntityAssociation:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.EntityAssociation.SensorName), sdr.EntityAssociation.ContainerEntityID)
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.DeviceRelative.SensorName), sdr.DeviceRelative.ContainerEntityID)
	case SDRRecordTypeGenericLocator:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.GenericDeviceLocator.SensorName), sdr.GenericDeviceLocator.EntityID)
	case SDRRecordTypeFRUDeviceLocator:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.FRUDeviceLocator.SensorName), sdr.FRUDeviceLocator.FRUEntityID)
	case SDRRecordTypeManagementControllerDeviceLocator:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.MgmtControllerDeviceLocator.SensorName), sdr.MgmtControllerDeviceLocator.EntityID)
	case SDRRecordTypeManagementControllerConfirmation:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.MgmtControllerConfirmation.ManufacturerID), sdr.MgmtControllerConfirmation.ManufacturerID)
	case SDRRecordTypeBMCMessageChannelInfo:
		return fmt.Sprintf(format, recordID, uint8(recordType), string(sdr.BMCChannelInfo.Channel0), sdr.BMCChannelInfo.Channel0)
	case SDRRecordTypeOEM:
		return "OEM"
	default:
		return ""
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

	// Todo
	AssertionEventMask           uint16
	DeassertionEventMask         uint16
	DiscreteSettableReadableMask uint16

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

	// ===== Full Sensor ONLY
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
	// ===== Full Sensor ONLY

	PositiveGoingThresholdHysteresisValue uint8
	NegativeGoingThresholdHysteresisValue uint8

	TypeLength TypeLength
	SensorName []byte
}

// 43.2 SDR Type 02h, Compact Sensor Record
type SDRCompact struct {
	//
	// Record KEY
	//

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

	// Todo
	AssertionEventMask           uint16
	DeassertionEventMask         uint16
	DiscreteSettableReadableMask uint16

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

	TypeLength TypeLength
	SensorName []byte
}

// 43.3 SDR Type 03h, Event-Only Record
type SDREventOnly struct {
	//
	// Record KEY
	//

	// [7:1] - 7-bit I2C Slave Address, or 7-bit system software ID[2]
	// [0] - 0b = ID is IPMB Slave Address, 1b = system software ID
	SensorOwnerID uint8

	SensorOwnerLUN uint8
	SensorNumber   uint8 // Unique number identifying the sensor behind a given slave address and LUN. Code FFh reserved.

	//
	// RECORD BODY
	//

	SensorEntityID        uint8
	SensorEntityInstance  uint8
	SensorType            uint8
	SensorEventType       uint8
	SensorDirection       uint8
	EntityInstanceSharing uint8

	TypeLength TypeLength
	SensorName []byte
}

// 43.4 SDR Type 08h - Entity Association Record
type SDREntityAssociation struct {
	//
	// Record KEY
	//

	ContainerEntityID          uint8
	ContainerEntityInstance    uint8
	Flags                      uint8
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

	TypeLength TypeLength
	SensorName []byte
}

// 43.5 SDR Type 09h - Device-relative Entity Association Record
type SDRDeviceRelative struct {
	//
	// Record KEY
	//

	ContainerEntityID               uint8
	ContainerEntityInstance         uint8
	ContainerEntityDeviceAddress    uint8
	ContainerEntityDeviceChannel    uint8
	Flags                           uint8
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

	TypeLength TypeLength
	SensorName []byte
}

// 43.6 SDR Type 0Ah:0Fh - Reserved Records
type SDRReserved struct {
}

// 43.7 SDR Type 10h - Generic Device Locator Record
type SDRGenericDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceAccessAddress uint8
	DeviceSlaveAddress  uint8
	AccessLUNBusID      uint8

	//
	// RECORD BODY
	//

	AddressSpan        uint8
	DeviceType         uint8
	DeviceTypeModifier uint8
	EntityID           uint8
	EntityInstance     uint8

	TypeLength TypeLength
	SensorName []byte
}

// 43.8 SDR Type 11h - FRU Device Locator Record
type SDRFRUDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceAccessAddress uint8
	DeviceSlaveAddress  uint8
	AccessLUNBusID      uint8
	ChannelNumber       uint8

	//
	// RECORD BODY
	//

	DeviceType         uint8
	DeviceTypeModifier uint8
	FRUEntityID        uint8
	FRUEntityInstance  uint8

	TypeLength TypeLength
	SensorName []byte
}

// 43.9 SDR Type 12h - Management Controller Device Locator Record
type SDRMgmtControllerDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8
	ChannelNumber      uint8

	//
	// RECORD BODY
	//

	PowerStateNotification uint8
	DeviceCapabilities     uint8
	EntityID               uint8
	EntityInstance         uint8

	TypeLength TypeLength
	SensorName []byte
}

// 43.10 SDR Type 13h - Management Controller Confirmation Record
type SDRMgmtControllerConfirmation struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8
	DeviceID           uint8
	ChannelNumber      uint8

	//
	// RECORD BODY
	//

	FirmwareRevision1 uint8
	FirmwareRevision2 uint8
	IPMIVersion       uint8
	ManufacturerID    uint32 // 3 bytes only
	ProductID         uint16
	DeviceGUID        []byte // 16 bytes
}

// 43.11 SDR Type 14h - BMC Message Channel Info Record
type SDRBMCChannelInfo struct {
	//
	// NO Record KEY
	//

	//
	// RECORD BODY
	//

	Channel0 uint8
	Channel1 uint8
	Channel2 uint8
	Channel3 uint8
	Channel4 uint8
	Channel5 uint8
	Channel6 uint8
	Channel7 uint8

	MessagingInterruptType uint8

	EventMessageBufferInterruptType uint8
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
	l := uint8(tl) & 0x3f               // the lowest 5 bits

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

func parseSDRFullSensor(data []byte, sdr *SDR) error {
	minSize := SDRFullSensorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (full sensor) data must be longer than %d", minSize)
	}

	s := &SDRFull{}
	sdr.Full = s

	s.SensorOwnerID, _, _ = unpackUint8(data, 5)
	s.SensorOwnerLUN, _, _ = unpackUint8(data, 6)
	s.SensorNumber, _, _ = unpackUint8(data, 7)
	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorInitialization, _, _ = unpackUint8(data, 10)
	s.SensorCapabilitites, _, _ = unpackUint8(data, 11)
	s.SensorType, _, _ = unpackUint8(data, 12)
	s.SensorEventType, _, _ = unpackUint8(data, 13)

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
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())

	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (full sensor) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)

	return nil
}

func parseSDRCompactSensor(data []byte, sdr *SDR) error {
	minSize := SDRCompactSensorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (compact sensor) data must be longer than %d", minSize)
	}

	s := &SDRCompact{}
	sdr.Compact = s

	s.SensorOwnerID, _, _ = unpackUint8(data, 5)
	s.SensorOwnerLUN, _, _ = unpackUint8(data, 6)
	s.SensorNumber, _, _ = unpackUint8(data, 7)
	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorInitialization, _, _ = unpackUint8(data, 10)
	s.SensorCapabilitites, _, _ = unpackUint8(data, 11)
	s.SensorType, _, _ = unpackUint8(data, 12)
	s.SensorEventType, _, _ = unpackUint8(data, 13)

	s.SensorUnits1, _, _ = unpackUint8(data, 20)
	s.SensorUnits1, _, _ = unpackUint8(data, 21)
	s.SensorUnits3, _, _ = unpackUint8(data, 22)

	s.PositiveGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 25)
	s.NegativeGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 26)

	typeLength, _, _ := unpackUint8(data, 31)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (compact sensor) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

func parseSDREventOnly(data []byte, sdr *SDR) error {
	minSize := SDREventOnlyMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize)
	}

	s := &SDREventOnly{}
	sdr.EventOnly = s

	s.SensorOwnerID, _, _ = unpackUint8(data, 5)
	s.SensorOwnerLUN, _, _ = unpackUint8(data, 6)
	s.SensorNumber, _, _ = unpackUint8(data, 7)
	s.SensorEntityID, _, _ = unpackUint8(data, 8)
	s.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	s.SensorType, _, _ = unpackUint8(data, 10)
	s.SensorEventType, _, _ = unpackUint8(data, 11)

	typeLength, _, _ := unpackUint8(data, 16)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
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
	s.Flags, _, _ = unpackUint8(data, 7)
	s.ContaineredEntity1ID, _, _ = unpackUint8(data, 8)
	s.ContaineredEntity1Instance, _, _ = unpackUint8(data, 9)

	s.ContaineredEntity2ID, _, _ = unpackUint8(data, 10)
	s.ContaineredEntity2Instance, _, _ = unpackUint8(data, 11)
	s.ContaineredEntity3ID, _, _ = unpackUint8(data, 12)
	s.ContaineredEntity3Instance, _, _ = unpackUint8(data, 13)
	s.ContaineredEntity4ID, _, _ = unpackUint8(data, 14)
	s.ContaineredEntity4Instance, _, _ = unpackUint8(data, 15)

	typeLength, _, _ := unpackUint8(data, 16)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < size+idStrLen {
		return fmt.Errorf("sdr (entity association) data must be longer than %d", size+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, size, idStrLen)
	return nil
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

	s.Flags, _, _ = unpackUint8(data, 9)

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

func parseSDRGenericLocator(data []byte, sdr *SDR) error {
	minSize := SDRGenericLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize)
	}

	s := &SDRGenericDeviceLocator{}
	sdr.GenericDeviceLocator = s

	s.DeviceAccessAddress, _, _ = unpackUint8(data, 5)
	s.DeviceSlaveAddress, _, _ = unpackUint8(data, 6)
	s.AccessLUNBusID, _, _ = unpackUint8(data, 7)

	s.AddressSpan, _, _ = unpackUint8(data, 8)
	s.DeviceType, _, _ = unpackUint8(data, 10)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

func parseSDRFRUDeviceLocator(data []byte, sdr *SDR) error {
	minSize := SDRFRUDeviceLocatorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize)
	}

	s := &SDRFRUDeviceLocator{}
	sdr.FRUDeviceLocator = s

	s.DeviceAccessAddress, _, _ = unpackUint8(data, 5)
	s.DeviceSlaveAddress, _, _ = unpackUint8(data, 6)
	s.AccessLUNBusID, _, _ = unpackUint8(data, 7)
	s.ChannelNumber, _, _ = unpackUint8(data, 8)

	s.DeviceType, _, _ = unpackUint8(data, 10)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.FRUEntityID, _, _ = unpackUint8(data, 12)
	s.FRUEntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

func parseSDRManagementControllerDeviceLocator(data []byte, sdr *SDR) error {
	minSize := SDRManagementControllerDeviceLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerDeviceLocator{}
	sdr.MgmtControllerDeviceLocator = s

	s.DeviceSlaveAddress, _, _ = unpackUint8(data, 5)
	s.ChannelNumber, _, _ = unpackUint8(data, 6)

	s.PowerStateNotification, _, _ = unpackUint8(data, 7)
	s.DeviceCapabilities, _, _ = unpackUint8(data, 8)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.TypeLength = TypeLength(typeLength)

	idStrLen := int(s.TypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize+idStrLen)
	}
	s.SensorName, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

func parseSDRManagementControllerConfirmation(data []byte, sdr *SDR) error {
	minSize := SDRManagementControllerConfirmationSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller confirmation) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerConfirmation{}
	sdr.MgmtControllerConfirmation = s

	s.DeviceSlaveAddress, _, _ = unpackUint8(data, 5)
	s.DeviceID, _, _ = unpackUint8(data, 6)
	s.ChannelNumber, _, _ = unpackUint8(data, 7)

	s.FirmwareRevision1, _, _ = unpackUint8(data, 8)
	s.FirmwareRevision2, _, _ = unpackUint8(data, 9)
	s.IPMIVersion, _, _ = unpackUint8(data, 10)
	s.ManufacturerID, _, _ = unpackUint24L(data, 11)
	s.ProductID, _, _ = unpackUint16L(data, 14)
	s.DeviceGUID, _, _ = unpackBytes(data, 16, 16)
	return nil
}

func parseSDRBMCMessageChannelInfo(data []byte, sdr *SDR) error {
	minSize := SDRBMCMessageChannelInfoSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (bmc message channel info) data must be longer than %d", minSize)
	}

	s := &SDRBMCChannelInfo{}
	sdr.BMCChannelInfo = s

	s.Channel0, _, _ = unpackUint8(data, 5)
	s.Channel1, _, _ = unpackUint8(data, 6)
	s.Channel2, _, _ = unpackUint8(data, 7)
	s.Channel3, _, _ = unpackUint8(data, 8)
	s.Channel4, _, _ = unpackUint8(data, 9)
	s.Channel5, _, _ = unpackUint8(data, 10)
	s.Channel6, _, _ = unpackUint8(data, 11)
	s.Channel7, _, _ = unpackUint8(data, 12)

	s.MessagingInterruptType, _, _ = unpackUint8(data, 13)
	s.EventMessageBufferInterruptType, _, _ = unpackUint8(data, 14)
	return nil
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
