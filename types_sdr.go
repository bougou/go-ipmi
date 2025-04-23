package ipmi

import (
	"bytes"
	"context"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

const SDRRecordHeaderSize int = 5

// 43. Sensor Data Record Formats
// SDRRecordType is a number representing the type of the record.
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
)

func (sdrRecordType SDRRecordType) String() string {
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

	s, ok := sdrRecordTypeMap[sdrRecordType]
	if !ok {
		return "Reserved"
	}
	return s
}

type SDRHeader struct {
	RecordID     uint16
	SDRVersion   uint8         // The version number of the SDR specification.
	RecordType   SDRRecordType // A number representing the type of the record. E.g. 01h = 8-bit Sensor with Thresholds.
	RecordLength uint8         // Number of bytes of data following the Record Length field.
}

// 43. Sensor Data Record Formats
type SDR struct {
	// NextRecordID should be filled by ParseSDR.
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

	recordStr := fmt.Sprintf(`
Record ID:             : %#02x
Record Type:           : %s
SDR Version:           : %#02x
Record Length:         : %d
`,
		sdr.RecordHeader.RecordID,
		sdr.RecordHeader.RecordType,
		sdr.RecordHeader.SDRVersion,
		sdr.RecordHeader.RecordLength,
	)

	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return recordStr + sdr.Full.String()
	case SDRRecordTypeCompactSensor:
		return recordStr + sdr.Compact.String()
	case SDRRecordTypeEventOnly:
		return recordStr + sdr.EventOnly.String()
	case SDRRecordTypeEntityAssociation:
		return recordStr
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		return recordStr
	case SDRRecordTypeGenericLocator:
		return recordStr
	case SDRRecordTypeFRUDeviceLocator:
		return recordStr
	case SDRRecordTypeManagementControllerDeviceLocator:
		return recordStr
	case SDRRecordTypeManagementControllerConfirmation:
		return recordStr
	case SDRRecordTypeOEM:
		return recordStr
	default:
		return recordStr
	}
}

func (sdr *SDR) SensorNumber() SensorNumber {
	recordType := sdr.RecordHeader.RecordType
	switch recordType {
	case SDRRecordTypeFullSensor:
		return sdr.Full.SensorNumber
	case SDRRecordTypeCompactSensor:
		return sdr.Compact.SensorNumber
	case SDRRecordTypeEventOnly:
		return sdr.EventOnly.SensorNumber
	}
	return SensorNumberReserved
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
	}
	return ""
}

// Determine if sensor has an analog reading
func (sdr *SDR) HasAnalogReading() bool {

	// Only Full sensors can return analog values, Compact sensors can't return analog values.
	// But not all Full sensors return analog values.

	if sdr.RecordHeader.RecordType != SDRRecordTypeFullSensor {
		return false
	}

	if sdr.Full == nil {
		return false
	}

	return sdr.Full.HasAnalogReading()
}

// ParseSDR parses raw SDR record data to SDR struct.
// This function is normally used after getting GetSDRResponse or GetDeviceSDRResponse to
// interpret the raw SDR record data in the response.
func ParseSDR(data []byte, nextRecordID uint16) (*SDR, error) {
	sdrHeader := &SDRHeader{}
	if len(data) < SDRRecordHeaderSize {
		return nil, ErrNotEnoughDataWith("sdr record header size", len(data), SDRRecordHeaderSize)
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
			return nil, fmt.Errorf("parseSDRFullSensor failed, err: %w", err)
		}
	case SDRRecordTypeCompactSensor:
		if err := parseSDRCompactSensor(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRCompactSensor failed, err: %w", err)
		}
	case SDRRecordTypeEventOnly:
		if err := parseSDREventOnly(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDREventOnly failed, err: %w", err)
		}
	case SDRRecordTypeEntityAssociation:
		if err := parseSDREntityAssociation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDREntityAssociation failed, err: %w", err)
		}
	case SDRRecordTypeDeviceRelativeEntityAssociation:
		if err := parseSDRDeviceRelativeEntityAssociation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRDeviceRelativeEntityAssociation failed, err: %w", err)
		}
	case SDRRecordTypeGenericLocator:
		if err := parseSDRGenericLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRGenericLocator failed, err: %w", err)
		}
	case SDRRecordTypeFRUDeviceLocator:
		if err := parseSDRFRUDeviceLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRFRUDeviceLocator failed, err: %w", err)
		}
	case SDRRecordTypeManagementControllerDeviceLocator:
		if err := parseSDRManagementControllerDeviceLocator(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRManagementControllerDeviceLocator failed, err: %w", err)
		}
	case SDRRecordTypeManagementControllerConfirmation:
		if err := parseSDRManagementControllerConfirmation(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRManagementControllerConfirmation failed, err: %w", err)
		}
	case SDRRecordTypeBMCMessageChannelInfo:
		if err := parseSDRBMCMessageChannelInfo(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDRBMCMessageChannelInfo failed, err: %w", err)
		}
	case SDRRecordTypeOEM:
		if err := parseSDROEM(data, sdr); err != nil {
			return nil, fmt.Errorf("parseSDROEM failed, err: %w", err)
		}
	}

	return sdr, nil
}

// Format SDRs of FRU record type
func FormatSDRs_FRU(records []*SDR) string {
	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)

	headers := []string{
		"RecordID",
		"RecordType",
		"DeviceAccessAddr",
		"FRUDeviceID",
		"IsLogicFRU",
		"AccessLUN",
		"PrivateBusID",
		"ChannelNumber",
		"DeviceType",
		"Modifier",
		"FRUEntityID",
		"FRUEntityInstance",
		"TypeLength",
		"DeviceName",
	}
	table.SetHeader(headers)
	table.SetFooter(headers)

	for _, sdr := range records {
		if sdr == nil || sdr.RecordHeader == nil {
			continue
		}

		recordID := sdr.RecordHeader.RecordID
		recordType := sdr.RecordHeader.RecordType

		switch recordType {
		case SDRRecordTypeFRUDeviceLocator:
			sdrFRU := sdr.FRUDeviceLocator
			table.Append([]string{
				fmt.Sprintf("%#02x", recordID),
				fmt.Sprintf("%s (%#02x)", recordType.String(), uint8(recordType)),
				fmt.Sprintf("%#02x", sdrFRU.DeviceAccessAddress),
				fmt.Sprintf("%#02x", sdrFRU.FRUDeviceID_SlaveAddress),
				fmt.Sprintf("%v", sdrFRU.IsLogicalFRUDevice),
				fmt.Sprintf("%#02x", sdrFRU.AccessLUN),
				fmt.Sprintf("%#02x", sdrFRU.PrivateBusID),
				fmt.Sprintf("%#02x", sdrFRU.ChannelNumber),
				fmt.Sprintf("%s (%#02x)", sdrFRU.DeviceType.String(), uint8(sdrFRU.DeviceType)),
				fmt.Sprintf("%#02x", sdrFRU.DeviceTypeModifier),
				fmt.Sprintf("%#02x", sdrFRU.FRUEntityID),
				fmt.Sprintf("%#02x", sdrFRU.FRUEntityInstance),
				sdrFRU.DeviceIDTypeLength.String(),
				string(sdrFRU.DeviceIDBytes),
			})
		default:
		}

	}

	table.Render()
	return buf.String()

}

// FormatSDRs returns a table formatted string for print.
func FormatSDRs(records []*SDR) string {
	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)

	headers := []string{
		"RecordID",
		"RecordType",
		"GeneratorID",
		"SensorNumber",
		"SensorName",
		"Entity",
		"SensorType",
		"EventReadingType",
		"SensorValue",
		"SensorUnit",
		"SensorStatus",
	}
	table.SetHeader(headers)
	table.SetFooter(headers)

	for _, sdr := range records {
		if sdr == nil || sdr.RecordHeader == nil {
			continue
		}

		recordID := sdr.RecordHeader.RecordID
		recordType := sdr.RecordHeader.RecordType

		var generatorID GeneratorID
		var sensorUnit SensorUnit
		var entityID EntityID
		var entityInstance EntityInstance
		var sensorType SensorType
		var eventReadingType EventReadingType
		var sensorValue float64
		var sensorStatus string

		switch recordType {
		case SDRRecordTypeFullSensor:
			generatorID = sdr.Full.GeneratorID
			sensorUnit = sdr.Full.SensorUnit
			entityID = sdr.Full.SensorEntityID
			entityInstance = sdr.Full.SensorEntityInstance
			sensorType = sdr.Full.SensorType
			eventReadingType = sdr.Full.SensorEventReadingType
			sensorValue = sdr.Full.SensorValue
			sensorStatus = sdr.Full.SensorStatus

		case SDRRecordTypeCompactSensor:
			generatorID = sdr.Compact.GeneratorID
			sensorUnit = sdr.Compact.SensorUnit
			entityID = sdr.Compact.SensorEntityID
			entityInstance = sdr.Compact.SensorEntityInstance
			sensorType = sdr.Compact.SensorType
			eventReadingType = sdr.Compact.SensorEventReadingType
			sensorValue = sdr.Compact.SensorValue
			sensorStatus = sdr.Compact.SensorStatus

		default:
		}

		table.Append([]string{
			fmt.Sprintf("%#02x", recordID),
			fmt.Sprintf("%s (%#02x)", recordType.String(), uint8(recordType)),
			fmt.Sprintf("%#04x", uint16(generatorID)),
			fmt.Sprintf("%#02x", sdr.SensorNumber()),
			sdr.SensorName(),
			canonicalEntityString(entityID, entityInstance),
			sensorType.String(),
			eventReadingType.String(),
			fmt.Sprintf("%#.2f", sensorValue),
			sensorUnit.String(),
			sensorStatus,
		})
	}

	table.Render()
	return buf.String()
}

// Mask_Threshold holds masks for a specific threshold type.
type Mask_Threshold struct {
	StatusReturned bool // Indicates whether this threshold comparison status is returned via the Get Sensor Reading command.
	Settable       bool
	Readable       bool
	High_Assert    bool
	Low_Assert     bool
	High_Deassert  bool
	Low_Deassert   bool
}

// Mask_Thresholds holds masks for all threshold types.
type Mask_Thresholds struct {
	LNR Mask_Threshold
	LCR Mask_Threshold
	LNC Mask_Threshold
	UNR Mask_Threshold
	UCR Mask_Threshold
	UNC Mask_Threshold
}

func (mask *Mask_Thresholds) IsThresholdReadable(thresholdType SensorThresholdType) bool {
	switch thresholdType {
	case SensorThresholdType_LCR:
		return mask.LCR.Readable
	case SensorThresholdType_LNR:
		return mask.LNR.Readable
	case SensorThresholdType_LNC:
		return mask.LNC.Readable
	case SensorThresholdType_UCR:
		return mask.UCR.Readable
	case SensorThresholdType_UNC:
		return mask.UNC.Readable
	case SensorThresholdType_UNR:
		return mask.UNR.Readable
	}
	return false
}

type Mask_DiscreteEvent struct {
	State_0  bool
	State_1  bool
	State_2  bool
	State_3  bool
	State_4  bool
	State_5  bool
	State_6  bool
	State_7  bool
	State_8  bool
	State_9  bool
	State_10 bool
	State_11 bool
	State_12 bool
	State_13 bool
	State_14 bool
}

func (mask Mask_DiscreteEvent) TrueEvents() []uint8 {
	events := []uint8{}

	if mask.State_0 {
		events = append(events, 0)
	}
	if mask.State_1 {
		events = append(events, 1)
	}
	if mask.State_2 {
		events = append(events, 2)
	}
	if mask.State_3 {
		events = append(events, 3)
	}
	if mask.State_4 {
		events = append(events, 4)
	}
	if mask.State_5 {
		events = append(events, 5)
	}
	if mask.State_6 {
		events = append(events, 6)
	}
	if mask.State_7 {
		events = append(events, 7)
	}
	if mask.State_8 {
		events = append(events, 8)
	}
	if mask.State_9 {
		events = append(events, 9)
	}
	if mask.State_10 {
		events = append(events, 10)
	}
	if mask.State_11 {
		events = append(events, 11)
	}
	if mask.State_12 {
		events = append(events, 12)
	}
	if mask.State_13 {
		events = append(events, 13)
	}
	if mask.State_14 {
		events = append(events, 14)
	}
	return events
}

type Mask_Discrete struct {
	// Assertion Event Mask for non-threshold based sensors, true means assertion event can be generated for this state
	Assert Mask_DiscreteEvent
	// Deassertion Event Mask for non-threshold based sensors, true means deassertion event can be generated for this state
	Deassert Mask_DiscreteEvent
	// Reading Mask for non-threshold based sensors, true means discrete state can be returned by this sensor
	Reading Mask_DiscreteEvent
}

// For non-threshold-based sensors, Mask holds:
//   - Assertion Event Mask
//   - Deassertion Event Mask
//   - Discrete Reading Mask
//
// For threshold-based sensors, Mask holds:
//   - Lower Threshold Reading Mask
//   - Upper Threshold Reading Mask
//   - Settable Threshold Mask, Readable Threshold Mask
//
// Used in Full and Compact SDR
type Mask struct {
	Threshold Mask_Thresholds
	Discrete  Mask_Discrete
}

// ParseAssertLower fill:
//   - Assertion Event Mask
//   - Lower Threshold Reading Mask
//   - Threshold Assertion Event Mask
func (mask *Mask) ParseAssertLower(b uint16) {
	lsb := uint8(b & 0x00ff) // Least Significant Byte
	msb := uint8(b >> 8)     // Most Significant Byte

	// Assertion Event Mask
	mask.Discrete.Assert.State_14 = isBit6Set(lsb)
	mask.Discrete.Assert.State_13 = isBit5Set(lsb)
	mask.Discrete.Assert.State_12 = isBit4Set(lsb)
	mask.Discrete.Assert.State_11 = isBit3Set(lsb)
	mask.Discrete.Assert.State_10 = isBit2Set(lsb)
	mask.Discrete.Assert.State_9 = isBit1Set(lsb)
	mask.Discrete.Assert.State_8 = isBit0Set(lsb)
	mask.Discrete.Assert.State_7 = isBit7Set(msb)
	mask.Discrete.Assert.State_6 = isBit6Set(msb)
	mask.Discrete.Assert.State_5 = isBit5Set(msb)
	mask.Discrete.Assert.State_4 = isBit4Set(msb)
	mask.Discrete.Assert.State_3 = isBit3Set(msb)
	mask.Discrete.Assert.State_2 = isBit2Set(msb)
	mask.Discrete.Assert.State_1 = isBit1Set(msb)
	mask.Discrete.Assert.State_0 = isBit0Set(msb)

	// Lower Threshold Reading Mask
	// Indicates which lower threshold comparison status is returned via the Get Sensor Reading command
	mask.Threshold.LNR.StatusReturned = isBit6Set(lsb)
	mask.Threshold.LCR.StatusReturned = isBit5Set(lsb)
	mask.Threshold.LNC.StatusReturned = isBit4Set(lsb)

	// Threshold Assertion Event Mask
	mask.Threshold.UNR.High_Assert = isBit3Set(lsb)
	mask.Threshold.UNR.Low_Assert = isBit2Set(lsb)
	mask.Threshold.UCR.High_Assert = isBit1Set(lsb)
	mask.Threshold.UCR.Low_Assert = isBit0Set(lsb)
	mask.Threshold.UNC.High_Assert = isBit7Set(msb)
	mask.Threshold.UNC.Low_Assert = isBit6Set(msb)
	mask.Threshold.LNR.High_Assert = isBit5Set(msb)
	mask.Threshold.LNR.Low_Assert = isBit4Set(msb)
	mask.Threshold.LCR.High_Assert = isBit3Set(msb)
	mask.Threshold.LCR.Low_Assert = isBit2Set(msb)
	mask.Threshold.LNC.High_Assert = isBit1Set(msb)
	mask.Threshold.LNC.Low_Assert = isBit0Set(msb)

}

func (mask *Mask) ParseDeassertUpper(b uint16) {
	lsb := uint8(b & 0x00ff) // Least Significant Byte
	msb := uint8(b >> 8)     // Most Significant Byte

	// Deassertion Event Mask
	mask.Discrete.Deassert.State_14 = isBit6Set(lsb)
	mask.Discrete.Deassert.State_13 = isBit5Set(lsb)
	mask.Discrete.Deassert.State_12 = isBit4Set(lsb)
	mask.Discrete.Deassert.State_11 = isBit3Set(lsb)
	mask.Discrete.Deassert.State_10 = isBit2Set(lsb)
	mask.Discrete.Deassert.State_9 = isBit1Set(lsb)
	mask.Discrete.Deassert.State_8 = isBit0Set(lsb)
	mask.Discrete.Deassert.State_7 = isBit7Set(msb)
	mask.Discrete.Deassert.State_6 = isBit6Set(msb)
	mask.Discrete.Deassert.State_5 = isBit5Set(msb)
	mask.Discrete.Deassert.State_4 = isBit4Set(msb)
	mask.Discrete.Deassert.State_3 = isBit3Set(msb)
	mask.Discrete.Deassert.State_2 = isBit2Set(msb)
	mask.Discrete.Deassert.State_1 = isBit1Set(msb)
	mask.Discrete.Deassert.State_0 = isBit0Set(msb)

	// Upper Threshold Reading Mask
	// Indicates which upper threshold comparison status is returned via the Get Sensor Reading command.
	mask.Threshold.UNR.StatusReturned = isBit6Set(lsb)
	mask.Threshold.UCR.StatusReturned = isBit5Set(lsb)
	mask.Threshold.UNC.StatusReturned = isBit4Set(lsb)

	// Threshold Deassertion Event Mask
	mask.Threshold.UNR.High_Deassert = isBit3Set(lsb)
	mask.Threshold.UNR.Low_Deassert = isBit2Set(lsb)
	mask.Threshold.UCR.High_Deassert = isBit1Set(lsb)
	mask.Threshold.UCR.Low_Deassert = isBit0Set(lsb)
	mask.Threshold.UNC.High_Deassert = isBit7Set(msb)
	mask.Threshold.UNC.Low_Deassert = isBit6Set(msb)
	mask.Threshold.LNR.High_Deassert = isBit5Set(msb)
	mask.Threshold.LNR.Low_Deassert = isBit4Set(msb)
	mask.Threshold.LCR.High_Deassert = isBit3Set(msb)
	mask.Threshold.LCR.Low_Deassert = isBit2Set(msb)
	mask.Threshold.LNC.High_Deassert = isBit1Set(msb)
	mask.Threshold.LNC.Low_Deassert = isBit0Set(msb)

}

func (mask *Mask) ParseReading(b uint16) {
	lsb := uint8(b & 0x0000ffff) // Least Significant Byte
	msb := uint8(b >> 8)         // Most Significant Byte

	// Reading Mask (for non-threshold based sensors)
	// Indicates what discrete readings can be returned by this sensor.
	mask.Discrete.Reading.State_14 = isBit6Set(lsb)
	mask.Discrete.Reading.State_13 = isBit5Set(lsb)
	mask.Discrete.Reading.State_12 = isBit4Set(lsb)
	mask.Discrete.Reading.State_11 = isBit3Set(lsb)
	mask.Discrete.Reading.State_10 = isBit2Set(lsb)
	mask.Discrete.Reading.State_9 = isBit1Set(lsb)
	mask.Discrete.Reading.State_8 = isBit0Set(lsb)
	mask.Discrete.Reading.State_7 = isBit7Set(msb)
	mask.Discrete.Reading.State_6 = isBit6Set(msb)
	mask.Discrete.Reading.State_5 = isBit5Set(msb)
	mask.Discrete.Reading.State_4 = isBit4Set(msb)
	mask.Discrete.Reading.State_3 = isBit3Set(msb)
	mask.Discrete.Reading.State_2 = isBit2Set(msb)
	mask.Discrete.Reading.State_1 = isBit1Set(msb)
	mask.Discrete.Reading.State_0 = isBit0Set(msb)

	// Settable Threshold Mask (for threshold-based sensors)
	// Indicates which thresholds are settable via the Set Sensor Thresholds.
	mask.Threshold.UNR.Settable = isBit5Set(lsb)
	mask.Threshold.UCR.Settable = isBit4Set(lsb)
	mask.Threshold.UNC.Settable = isBit3Set(lsb)
	mask.Threshold.LNR.Settable = isBit2Set(lsb)
	mask.Threshold.LCR.Settable = isBit1Set(lsb)
	mask.Threshold.LNC.Settable = isBit0Set(lsb)

	// Readable Threshold Mask (for threshold-based sensors)
	// Indicates which thresholds are readable via the Get Sensor Thresholds command.
	mask.Threshold.UNR.Readable = isBit5Set(msb)
	mask.Threshold.UCR.Readable = isBit4Set(msb)
	mask.Threshold.UNC.Readable = isBit3Set(msb)
	mask.Threshold.LNR.Readable = isBit2Set(msb)
	mask.Threshold.LCR.Readable = isBit1Set(msb)
	mask.Threshold.LNC.Readable = isBit0Set(msb)
}

// StatusReturnedThresholds returns all supported thresholds comparison status
// via the Get Sensor Reading command.
func (mask *Mask) StatusReturnedThresholds() SensorThresholdTypes {
	out := make([]SensorThresholdType, 0)
	if mask.Threshold.UNC.StatusReturned {
		out = append(out, SensorThresholdType_UNC)
	}
	if mask.Threshold.UCR.StatusReturned {
		out = append(out, SensorThresholdType_UCR)
	}
	if mask.Threshold.UNR.StatusReturned {
		out = append(out, SensorThresholdType_UNR)
	}
	if mask.Threshold.LNC.StatusReturned {
		out = append(out, SensorThresholdType_LNC)
	}
	if mask.Threshold.LCR.StatusReturned {
		out = append(out, SensorThresholdType_LCR)
	}
	if mask.Threshold.LNR.StatusReturned {
		out = append(out, SensorThresholdType_LNR)
	}
	return out
}

// ReadableThresholds returns all readable thresholds for the sensor.
func (mask *Mask) ReadableThresholds() SensorThresholdTypes {
	out := make([]SensorThresholdType, 0)
	if mask.Threshold.UNC.Readable {
		out = append(out, SensorThresholdType_UNC)
	}
	if mask.Threshold.UCR.Readable {
		out = append(out, SensorThresholdType_UCR)
	}
	if mask.Threshold.UNR.Readable {
		out = append(out, SensorThresholdType_UNR)
	}
	if mask.Threshold.LNC.Readable {
		out = append(out, SensorThresholdType_LNC)
	}
	if mask.Threshold.LCR.Readable {
		out = append(out, SensorThresholdType_LCR)
	}
	if mask.Threshold.LNR.Readable {
		out = append(out, SensorThresholdType_LNR)
	}
	return out
}

func (mask *Mask) SettableThresholds() SensorThresholdTypes {
	out := make([]SensorThresholdType, 0)
	if mask.Threshold.UNC.Settable {
		out = append(out, SensorThresholdType_UNC)
	}
	if mask.Threshold.UCR.Settable {
		out = append(out, SensorThresholdType_UCR)
	}
	if mask.Threshold.UNR.Settable {
		out = append(out, SensorThresholdType_UNR)
	}
	if mask.Threshold.LNC.Settable {
		out = append(out, SensorThresholdType_LNC)
	}
	if mask.Threshold.LCR.Settable {
		out = append(out, SensorThresholdType_LCR)
	}
	if mask.Threshold.LNR.Settable {
		out = append(out, SensorThresholdType_LNR)
	}
	return out
}

func (mask *Mask) SupportedThresholdEvents() SensorEvents {
	out := make([]SensorEvent, 0)

	// Assertion Events

	if mask.Threshold.UNC.High_Assert {
		out = append(out, SensorEvent_UNC_High_Assert)
	}
	if mask.Threshold.UNC.Low_Assert {
		out = append(out, SensorEvent_UNC_Low_Assert)
	}

	if mask.Threshold.UCR.High_Assert {
		out = append(out, SensorEvent_UCR_High_Assert)
	}
	if mask.Threshold.UCR.Low_Assert {
		out = append(out, SensorEvent_UCR_Low_Assert)
	}

	if mask.Threshold.UNR.High_Assert {
		out = append(out, SensorEvent_UNR_High_Assert)
	}
	if mask.Threshold.UNR.Low_Assert {
		out = append(out, SensorEvent_UNR_Low_Assert)
	}

	if mask.Threshold.LNC.High_Assert {
		out = append(out, SensorEvent_LNC_High_Assert)
	}
	if mask.Threshold.LNC.Low_Assert {
		out = append(out, SensorEvent_LNC_Low_Assert)
	}

	if mask.Threshold.LCR.High_Assert {
		out = append(out, SensorEvent_LCR_High_Assert)
	}
	if mask.Threshold.LCR.Low_Assert {
		out = append(out, SensorEvent_LCR_Low_Assert)
	}

	if mask.Threshold.LNR.High_Assert {
		out = append(out, SensorEvent_LNR_High_Assert)
	}
	if mask.Threshold.LNR.Low_Assert {
		out = append(out, SensorEvent_LNR_Low_Assert)
	}

	// Deassertion Events
	if mask.Threshold.UNC.High_Deassert {
		out = append(out, SensorEvent_UNC_High_Deassert)
	}
	if mask.Threshold.UNC.Low_Deassert {
		out = append(out, SensorEvent_UNC_Low_Deassert)
	}

	if mask.Threshold.UCR.High_Deassert {
		out = append(out, SensorEvent_UCR_High_Deassert)
	}
	if mask.Threshold.UCR.Low_Deassert {
		out = append(out, SensorEvent_UCR_Low_Deassert)
	}

	if mask.Threshold.UNR.High_Deassert {
		out = append(out, SensorEvent_UNR_High_Deassert)
	}
	if mask.Threshold.UNR.Low_Deassert {
		out = append(out, SensorEvent_UNR_Low_Deassert)
	}

	if mask.Threshold.LNC.High_Deassert {
		out = append(out, SensorEvent_LNC_High_Deassert)
	}
	if mask.Threshold.LNC.Low_Deassert {
		out = append(out, SensorEvent_LNC_Low_Deassert)
	}

	if mask.Threshold.LCR.High_Deassert {
		out = append(out, SensorEvent_LCR_High_Deassert)
	}
	if mask.Threshold.LCR.Low_Deassert {
		out = append(out, SensorEvent_LCR_Low_Deassert)
	}

	if mask.Threshold.LNR.High_Deassert {
		out = append(out, SensorEvent_LNR_High_Deassert)
	}
	if mask.Threshold.LNR.Low_Deassert {
		out = append(out, SensorEvent_LNR_Low_Deassert)
	}

	return out
}

// SensorCapabilities represent the capabilities of the sensor.
// SDRs of Full/Compact record type has this field.
type SensorCapabilities struct {
	// [7] - 1b = ignore sensor if Entity is not present or disabled. 0b = don't ignore sensor
	IgnoreSensorIfNoEntity bool

	// Sensor Auto Re-arm Support
	// Indicates whether the sensor requires manual rearming, or automatically rearms
	// itself when the event clears. 'manual' implies that the get sensor event status and
	// rearm sensor events commands are supported
	// [6] - 0b = no (manual), 1b = yes (auto)
	AutoRearm bool

	HysteresisAccess SensorHysteresisAccess
	ThresholdAccess  SensorThresholdAccess

	EventMessageControl SensorEventMessageControl
}

// SDRs of Full/Compact record type has this field.
type SensorInitialization struct {
	// 1b = Sensor is settable (Support the Set Sensor Reading And Event Status command)
	// 0b = Sensor is not settable
	//
	// using this bit to report settable sensors is optional.
	// I.e. it is ok to report a settable sensor as 'not settable' in the
	// SDR if it is desired to not report this capability to s/w
	Settable bool

	// 1b = enable scanning
	//
	// this bit=1 implies that the sensor
	// accepts the 'enable/disable scanning' bit in the Set
	// Sensor Event Enable command.
	InitScanning bool

	// 1b = enable events (per Sensor Event Message Control
	// Support bits in Sensor Capabilities field, and per
	// the Event Mask fields, below).
	InitEvents bool

	// 1b = initialize sensor thresholds (per settable threshold mask below).
	InitThresholds bool

	// 1b = initialize sensor hysteresis (per Sensor Hysteresis
	// Support bits in the Sensor Capabilities field, below).
	InitHysteresis bool

	// 1b = initialize Sensor Type and Event / Reading Type code
	InitSensorType bool

	// Sensor Default (power up) State
	//
	// Reports how this sensor comes up on device power up and hardware/cold reset.
	// The Initialization Agent does not use this bit. This bit solely reports to software
	// how the sensor comes prior to being initialized by the Initialization Agent.

	// 0b = event generation disabled, 1b = event generation enabled
	EventGenerationEnabled bool
	// 0b = sensor scanning disabled, 1b = sensor scanning enabled
	SensorScanningEnabled bool
}

// enhanceSDR will fill extra data for SDR
func (c *Client) enhanceSDR(ctx context.Context, sdr *SDR) error {
	if sdr == nil {
		return nil
	}

	if sdr.RecordHeader.RecordType != SDRRecordTypeFullSensor &&
		sdr.RecordHeader.RecordType != SDRRecordTypeCompactSensor {
		return nil
	}

	sensor, err := c.sdrToSensor(ctx, sdr)
	if err != nil {
		return fmt.Errorf("sdrToSensor failed, err: %w", err)
	}

	switch sdr.RecordHeader.RecordType {
	case SDRRecordTypeFullSensor:
		sdr.Full.SensorValue = sensor.Value
		sdr.Full.SensorStatus = sensor.Status()

	case SDRRecordTypeCompactSensor:
		sdr.Compact.SensorValue = sensor.Value
		sdr.Compact.SensorStatus = sensor.Status()
	}

	return nil
}
