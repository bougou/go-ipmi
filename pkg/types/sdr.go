package types

import (
	"fmt"
	"iter"

	"github.com/bougou/go-ipmi/pkg/render"
)

const SDRRecordHeaderSize int = 5

// SDRCommandSetVersion is the SDR command-set version returned by Get SDR
// Repository Info (§33.9) and stored in SDR record headers (§43).
const SDRCommandSetVersion uint8 = 0x51

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

	recordStr := "" +
		fmt.Sprintf("Record ID:             : %#02x\n", sdr.RecordHeader.RecordID) +
		fmt.Sprintf("Record Type:           : %s\n", sdr.RecordHeader.RecordType) +
		fmt.Sprintf("SDR Version:           : %#02x\n", sdr.RecordHeader.SDRVersion) +
		fmt.Sprintf("Record Length:         : %d\n", sdr.RecordHeader.RecordLength)

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

// Not all SDRs have a sensor number.
// Only Full/Compact/EventOnly SDRs have a sensor number.
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
// ParseSDRHeader parses the fixed 5-byte SDR record header (v2.0§43).
func ParseSDRHeader(data []byte) (*SDRHeader, error) {
	if len(data) < SDRRecordHeaderSize {
		return nil, ErrNotEnoughDataWith("sdr record header size", len(data), SDRRecordHeaderSize)
	}

	sdrHeader := &SDRHeader{}
	sdrHeader.RecordID, _, _ = UnpackUint16L(data, 0)
	sdrHeader.SDRVersion, _, _ = UnpackUint8(data, 2)
	recordType, _, _ := UnpackUint8(data, 3)
	sdrHeader.RecordType = SDRRecordType(recordType)
	sdrHeader.RecordLength, _, _ = UnpackUint8(data, 4)
	return sdrHeader, nil
}

func ParseSDR(data []byte, nextRecordID uint16) (*SDR, error) {
	sdrHeader, err := ParseSDRHeader(data)
	if err != nil {
		return nil, err
	}

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

func sdrFRUToRow(sdr *SDR, options ...any) map[string]string {
	row := map[string]string{}
	if sdr == nil || sdr.RecordHeader == nil {
		return row
	}

	recordID := sdr.RecordHeader.RecordID
	recordType := sdr.RecordHeader.RecordType

	switch recordType {
	case SDRRecordTypeFRUDeviceLocator:
		sdrFRU := sdr.FRUDeviceLocator

		row = map[string]string{
			"RecordID":          fmt.Sprintf("%#02x", recordID),
			"RecordType":        fmt.Sprintf("%s (%#02x)", recordType.String(), uint8(recordType)),
			"DeviceAccessAddr":  fmt.Sprintf("%#02x", sdrFRU.DeviceAccessAddress),
			"FRUDeviceID":       fmt.Sprintf("%#02x", sdrFRU.FRUDeviceID_SlaveAddress),
			"IsLogicFRU":        fmt.Sprintf("%v", sdrFRU.IsLogicalFRUDevice),
			"AccessLUN":         fmt.Sprintf("%#02x", sdrFRU.AccessLUN),
			"PrivateBusID":      fmt.Sprintf("%#02x", sdrFRU.PrivateBusID),
			"ChannelNumber":     fmt.Sprintf("%#02x", sdrFRU.ChannelNumber),
			"DeviceType":        fmt.Sprintf("%s (%#02x)", sdrFRU.DeviceType.String(), uint8(sdrFRU.DeviceType)),
			"Modifier":          fmt.Sprintf("%#02x", sdrFRU.DeviceTypeModifier),
			"FRUEntityID":       fmt.Sprintf("%#02x", sdrFRU.FRUEntityID),
			"FRUEntityInstance": fmt.Sprintf("%#02x", sdrFRU.FRUEntityInstance),
			"TypeLength":        sdrFRU.DeviceIDTypeLength.String(),
			"DeviceName":        string(sdrFRU.DeviceIDBytes),
		}
	}

	return row
}

// Format SDRs of FRU record type
func FormatSDRs_FRU(records []*SDR) string {
	return FormatSDRs_FRUWith(defaultRenderer, records)
}

func FormatSDRs_FRUWith(r render.Renderer, records []*SDR) string {
	rows := make([]map[string]string, 0)

	for _, sdr := range records {
		row := sdrFRUToRow(sdr)
		rows = append(rows, row)
	}

	columns := render.Columns(
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
	)

	return r.Render(columns, rows)
}

func FormatSDRsStream_FRU(seq iter.Seq[*Result[SDR]]) error {
	return FormatSDRsStream_FRUWith(defaultRenderer, seq)
}

func FormatSDRsStream_FRUWith(r render.Renderer, seq iter.Seq[*Result[SDR]]) error {
	columns := []render.Column{
		{Name: "RecordID"},
		{Name: "RecordType", Width: 20},
		{Name: "DeviceAccessAddr"},
		{Name: "FRUDeviceID"},
		{Name: "IsLogicFRU"},
		{Name: "AccessLUN"},
		{Name: "PrivateBusID"},
		{Name: "ChannelNumber"},
		{Name: "DeviceType", Width: 20},
		{Name: "Modifier"},
		{Name: "FRUEntityID"},
		{Name: "FRUEntityInstance"},
		{Name: "TypeLength", Width: 56},
		{Name: "DeviceName"},
	}

	return formatStream(r, seq, columns, sdrFRUToRow)
}

func sdrToRow(sdr *SDR, options ...any) map[string]string {

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

	return map[string]string{
		"RecordID":         fmt.Sprintf("%#02x", recordID),
		"RecordType":       fmt.Sprintf("%s (%#02x)", recordType.String(), uint8(recordType)),
		"GeneratorID":      fmt.Sprintf("%#04x", uint16(generatorID)),
		"SensorNumber":     fmt.Sprintf("%#02x", sdr.SensorNumber()),
		"SensorName":       sdr.SensorName(),
		"Entity":           canonicalEntityString(entityID, entityInstance),
		"SensorType":       sensorType.String(),
		"EventReadingType": eventReadingType.String(),
		"SensorValue":      fmt.Sprintf("%#.2f", sensorValue),
		"SensorUnit":       sensorUnit.String(),
		"SensorStatus":     sensorStatus,
	}

}

// FormatSDRs returns a table formatted string for print.
func FormatSDRs(records []*SDR) string {
	return FormatSDRsWith(defaultRenderer, records)
}

func FormatSDRsWith(r render.Renderer, records []*SDR) string {
	rows := make([]map[string]string, 0)

	for _, sdr := range records {
		if sdr == nil || sdr.RecordHeader == nil {
			continue
		}

		row := sdrToRow(sdr)
		rows = append(rows, row)
	}

	columns := render.Columns(
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
	)

	return r.Render(columns, rows)
}

func FormatSDRsStream(seq iter.Seq[*Result[SDR]]) error {
	return FormatSDRsStreamWith(defaultRenderer, seq)
}

func FormatSDRsStreamWith(r render.Renderer, seq iter.Seq[*Result[SDR]]) error {
	columns := []render.Column{
		{Name: "RecordID"},
		{Name: "RecordType", Width: 16},
		{Name: "GeneratorID"},
		{Name: "SensorNumber"},
		{Name: "SensorName", Width: 16},
		{Name: "Entity", Width: 24},
		{Name: "SensorType", Width: 18},
		{Name: "EventReadingType"},
		{Name: "SensorValue"},
		{Name: "SensorUnit", Width: 14},
		{Name: "SensorStatus"},
	}

	return formatStream(r, seq, columns, sdrToRow)
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
// Used in Full and Compact SDR. Each of the three mask fields is a 16-bit
// little-endian value: the record's first byte holds bits [7:0] and its second
// byte bits [15:8]. The Parse methods take the value as read with
// UnpackUint16L, and the bit numbers they use are those of the spec (v2.0
// Table 43-1, bytes 15-20).
type Mask struct {
	Threshold Mask_Thresholds
	Discrete  Mask_Discrete
}

// parseDiscreteEventMask decodes bits [14:0] of a mask field, one bit per
// event offset or discrete state.
func parseDiscreteEventMask(b uint16) Mask_DiscreteEvent {
	bit := func(n uint) bool { return b&(1<<n) != 0 }

	return Mask_DiscreteEvent{
		State_0:  bit(0),
		State_1:  bit(1),
		State_2:  bit(2),
		State_3:  bit(3),
		State_4:  bit(4),
		State_5:  bit(5),
		State_6:  bit(6),
		State_7:  bit(7),
		State_8:  bit(8),
		State_9:  bit(9),
		State_10: bit(10),
		State_11: bit(11),
		State_12: bit(12),
		State_13: bit(13),
		State_14: bit(14),
	}
}

// ParseAssertLower fill:
//   - Assertion Event Mask
//   - Lower Threshold Reading Mask
//   - Threshold Assertion Event Mask
func (mask *Mask) ParseAssertLower(b uint16) {
	bit := func(n uint) bool { return b&(1<<n) != 0 }

	// Assertion Event Mask
	mask.Discrete.Assert = parseDiscreteEventMask(b)

	// Lower Threshold Reading Mask
	// Indicates which lower threshold comparison status is returned via the Get Sensor Reading command
	mask.Threshold.LNR.StatusReturned = bit(14)
	mask.Threshold.LCR.StatusReturned = bit(13)
	mask.Threshold.LNC.StatusReturned = bit(12)

	// Threshold Assertion Event Mask
	mask.Threshold.UNR.High_Assert = bit(11)
	mask.Threshold.UNR.Low_Assert = bit(10)
	mask.Threshold.UCR.High_Assert = bit(9)
	mask.Threshold.UCR.Low_Assert = bit(8)
	mask.Threshold.UNC.High_Assert = bit(7)
	mask.Threshold.UNC.Low_Assert = bit(6)
	mask.Threshold.LNR.High_Assert = bit(5)
	mask.Threshold.LNR.Low_Assert = bit(4)
	mask.Threshold.LCR.High_Assert = bit(3)
	mask.Threshold.LCR.Low_Assert = bit(2)
	mask.Threshold.LNC.High_Assert = bit(1)
	mask.Threshold.LNC.Low_Assert = bit(0)
}

func (mask *Mask) ParseDeassertUpper(b uint16) {
	bit := func(n uint) bool { return b&(1<<n) != 0 }

	// Deassertion Event Mask
	mask.Discrete.Deassert = parseDiscreteEventMask(b)

	// Upper Threshold Reading Mask
	// Indicates which upper threshold comparison status is returned via the Get Sensor Reading command.
	mask.Threshold.UNR.StatusReturned = bit(14)
	mask.Threshold.UCR.StatusReturned = bit(13)
	mask.Threshold.UNC.StatusReturned = bit(12)

	// Threshold Deassertion Event Mask
	mask.Threshold.UNR.High_Deassert = bit(11)
	mask.Threshold.UNR.Low_Deassert = bit(10)
	mask.Threshold.UCR.High_Deassert = bit(9)
	mask.Threshold.UCR.Low_Deassert = bit(8)
	mask.Threshold.UNC.High_Deassert = bit(7)
	mask.Threshold.UNC.Low_Deassert = bit(6)
	mask.Threshold.LNR.High_Deassert = bit(5)
	mask.Threshold.LNR.Low_Deassert = bit(4)
	mask.Threshold.LCR.High_Deassert = bit(3)
	mask.Threshold.LCR.Low_Deassert = bit(2)
	mask.Threshold.LNC.High_Deassert = bit(1)
	mask.Threshold.LNC.Low_Deassert = bit(0)
}

func (mask *Mask) ParseReading(b uint16) {
	bit := func(n uint) bool { return b&(1<<n) != 0 }

	// Reading Mask (for non-threshold based sensors)
	// Indicates what discrete readings can be returned by this sensor.
	mask.Discrete.Reading = parseDiscreteEventMask(b)

	// Settable Threshold Mask (for threshold-based sensors)
	// Indicates which thresholds are settable via the Set Sensor Thresholds.
	mask.Threshold.UNR.Settable = bit(13)
	mask.Threshold.UCR.Settable = bit(12)
	mask.Threshold.UNC.Settable = bit(11)
	mask.Threshold.LNR.Settable = bit(10)
	mask.Threshold.LCR.Settable = bit(9)
	mask.Threshold.LNC.Settable = bit(8)

	// Readable Threshold Mask (for threshold-based sensors)
	// Indicates which thresholds are readable via the Get Sensor Thresholds command.
	mask.Threshold.UNR.Readable = bit(5)
	mask.Threshold.UCR.Readable = bit(4)
	mask.Threshold.UNC.Readable = bit(3)
	mask.Threshold.LNR.Readable = bit(2)
	mask.Threshold.LCR.Readable = bit(1)
	mask.Threshold.LNC.Readable = bit(0)
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
