package ipmi

import "fmt"

const (
	SDRRecordHeaderSize int = 5
	SDRFullSensorSize   int = 48 // the ID String Bytes is optional 16 bytes, maximum.
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

type SDRHeader struct {
	RecordID     uint16
	SDRVersion   uint8         // The version number of the SDR specification.
	RecordType   SDRRecordType // A number representing the type of the record. E.g. 01h = 8-bit Sensor with Thresholds.
	RecordLength uint8         // Number of bytes of data following the Record Length field.
}

// 43. Sensor Data Record Formats
type SDR struct {
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

// 43.1 SDRFull Type 01h, Full Sensor Record
type SDRFull struct {
	//
	// Record Header
	//

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
	IDString   []byte
}

// 43.2 SDR Type 02h, Compact Sensor Record
type SDRCompact struct {
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

	PositiveGoingThresholdHysteresisValue uint8
	NegativeGoingThresholdHysteresisValue uint8

	TypeLength TypeLength
	IDString   []byte
}

// 43.3 SDR Type 03h, Event-Only Record
type SDREventOnly struct {
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

	SensorEntityID        uint8
	SensorEntityInstance  uint8
	SensorType            uint8
	SensorEventType       uint8
	SensorDirection       uint8
	EntityInstanceSharing uint8

	TypeLength TypeLength
	IDString   []byte
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
	IDString   []byte
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
	IDString   []byte
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
	IDString   []byte
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
	IDString   []byte
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
	IDString   []byte
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
	ManufacturerID    uint8
	ProductID         uint8
	DeviceGUID        uint8
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

	ManufacturerID uint8
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

func ParseSDR(data []byte) (*SDR, error) {
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

	}

	return sdr, nil
}

func parseSDRFullSensor(data []byte, sdr *SDR) error {
	if len(data) < SDRFullSensorSize {
		return fmt.Errorf("sdr full sensor data must be longer than %d", SDRFullSensorSize)
	}

	fullSDR := &SDRFull{}
	fullSDR.SensorOwnerID, _, _ = unpackUint8(data, 5)
	fullSDR.SensorOwnerLUN, _, _ = unpackUint8(data, 6)
	fullSDR.SensorNumber, _, _ = unpackUint8(data, 7)
	fullSDR.SensorEntityID, _, _ = unpackUint8(data, 8)
	fullSDR.SensorEntityInstance, _, _ = unpackUint8(data, 9)
	fullSDR.SensorInitialization, _, _ = unpackUint8(data, 10)
	fullSDR.SensorCapabilitites, _, _ = unpackUint8(data, 11)
	fullSDR.SensorType, _, _ = unpackUint8(data, 12)
	fullSDR.SensorEventType, _, _ = unpackUint8(data, 13)

	fullSDR.SensorUnits1, _, _ = unpackUint8(data, 20)
	fullSDR.SensorUnits1, _, _ = unpackUint8(data, 21)
	fullSDR.SensorUnits3, _, _ = unpackUint8(data, 22)

	fullSDR.Linearization, _, _ = unpackUint8(data, 23)

	fullSDR.NominalReading, _, _ = unpackUint8(data, 31)

	fullSDR.NormalMaximum, _, _ = unpackUint8(data, 32)
	fullSDR.NormalMinimum, _, _ = unpackUint8(data, 33)
	fullSDR.SensorMaximumReading, _, _ = unpackUint8(data, 34)
	fullSDR.SensorMinimumReading, _, _ = unpackUint8(data, 35)

	fullSDR.UpperNonRecoverableThreshold, _, _ = unpackUint8(data, 36)
	fullSDR.UpperCriticalThreshold, _, _ = unpackUint8(data, 37)
	fullSDR.UpperNonCriticalThreshold, _, _ = unpackUint8(data, 38)

	fullSDR.LowerNonRecoverableThreshold, _, _ = unpackUint8(data, 39)
	fullSDR.LowerCriticalThreshold, _, _ = unpackUint8(data, 40)
	fullSDR.LowerNonCriticalThreshold, _, _ = unpackUint8(data, 41)

	fullSDR.PositiveGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 42)
	fullSDR.NegativeGoingThresholdHysteresisValue, _, _ = unpackUint8(data, 43)

	typeLength, _, _ := unpackUint8(data, 47)
	fullSDR.TypeLength = TypeLength(typeLength)

	idStrLen := int(fullSDR.TypeLength.Length())

	if len(data) < 48+idStrLen {
		return fmt.Errorf("sdr full sensor data must be longer than %d", 48+idStrLen)
	}
	fullSDR.IDString, _, _ = unpackBytes(data, 48, idStrLen)

	sdr.Full = fullSDR
	return nil
}

// the first byte of data should be starting from Record Key fields
func parseSDRCompactSensor(data []byte, sdr *SDR) error {
	return nil
}
