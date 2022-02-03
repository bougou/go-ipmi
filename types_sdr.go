package ipmi

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

// 43. Sensor Data Record Formats
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

	recordStr := fmt.Sprintf(`RecordID:              : %#02x
RecordType:            : %s
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
// Todo, logic is not clear.
func (sdr *SDR) HasAnalogReading() bool {
	// Only Full sensors can return analog values.
	// Compact sensors can't return analog values
	if sdr.RecordHeader.RecordType != SDRRecordTypeFullSensor {
		return false
	}

	if sdr.Full == nil {
		return false
	}

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

	if sdr.Full.SensorEventReadingType.IsThreshold() {
		// for threshold sensors
		return true
	}

	// for non-threshold sensors

	if !sdr.Full.SensorUnit.IsAnalog() {
		return false
	}

	// for non-threshold sensors, but the analog data format indicates analog.
	// this rarely exists, except HP.
	// Todo

	return false
}

// ParseSDR parses raw SDR record data to SDR struct.
// This function is normally used after GetSDRResponse or GetDeviceSDRResponse to
// interpret the raw SDR record data in the response.
func ParseSDR(data []byte, nextRecordID uint16) (*SDR, error) {
	const SDRRecordHeaderSize int = 5

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

// FormatSDRs returns a table formatted string for print.
func FormatSDRs(records []*SDR) string {
	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)

	headers := []string{
		"RecordID",
		"RecordType",
		"GeneratorID",
		"SensorNumber",
		"SensorName",
		"Entity",
		"SensorType",
		"EventReadingType",
		"ReadingUnit",
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
		switch recordType {
		case SDRRecordTypeFullSensor:
			generatorID = sdr.Full.GeneratorID
			sensorUnit = sdr.Full.SensorUnit
			entityID = sdr.Full.SensorEntityID
			entityInstance = sdr.Full.SensorEntityInstance
			sensorType = sdr.Full.SensorType
			eventReadingType = sdr.Full.SensorEventReadingType

		case SDRRecordTypeCompactSensor:
			generatorID = sdr.Compact.GeneratorID
			sensorUnit = sdr.Compact.SensorUnit
			entityID = sdr.Compact.SensorEntityID
			entityInstance = sdr.Compact.SensorEntityInstance
			sensorType = sdr.Compact.SensorType
			eventReadingType = sdr.Compact.SensorEventReadingType

		default:
			// ignored the SDR
			continue
		}

		table.Append([]string{
			fmt.Sprintf("%#02x", recordID),
			fmt.Sprintf("%s (%#02x)", recordType.String(), uint8(recordType)),
			fmt.Sprintf("%#04x", generatorID),
			fmt.Sprintf("%#02x", sdr.SensorNumber()),
			sdr.SensorName(),
			canonicalEntityString(entityID, entityInstance),
			sensorType.String(),
			eventReadingType.String(),
			sensorUnit.String(),
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

type Mask_Discrete struct {
	// Assertion Event Mask for non-threshold based sensors

	Event_14_Assert bool
	Event_13_Assert bool
	Event_12_Assert bool
	Event_11_Assert bool
	Event_10_Assert bool
	Event_9_Assert  bool
	Event_8_Assert  bool
	Event_7_Assert  bool
	Event_6_Assert  bool
	Event_5_Assert  bool
	Event_4_Assert  bool
	Event_3_Assert  bool
	Event_2_Assert  bool
	Event_1_Assert  bool
	Event_0_Assert  bool
	// Deassertion Event Mask for non-threshold based sensors

	Event_14_Deassert bool
	Event_13_Deassert bool
	Event_12_Deassert bool
	Event_11_Deassert bool
	Event_10_Deassert bool
	Event_9_Deassert  bool
	Event_8_Deassert  bool
	Event_7_Deassert  bool
	Event_6_Deassert  bool
	Event_5_Deassert  bool
	Event_4_Deassert  bool
	Event_3_Deassert  bool
	Event_2_Deassert  bool
	Event_1_Deassert  bool
	Event_0_Deassert  bool

	// Reading Mask for non-threshold based sensors
	Reading_14_Supported bool
	Reading_13_Supported bool
	Reading_12_Supported bool
	Reading_11_Supported bool
	Reading_10_Supported bool
	Reading_9_Supported  bool
	Reading_8_Supported  bool
	Reading_7_Supported  bool
	Reading_6_Supported  bool
	Reading_5_Supported  bool
	Reading_4_Supported  bool
	Reading_3_Supported  bool
	Reading_2_Supported  bool
	Reading_1_Supported  bool
	Reading_0_Supported  bool
}

// Mask holds
//  - Assertion Event Mask / Lower Threshold Reading Mask
//  - Deassertion Event Mask / Upper Threshold Reading Mask
//  - Discrete Reading Mask / Settable Threshold Mask, Readable Threshold Mask
//
// Used in Full and Compact SDR
type Mask struct {
	Threshold Mask_Thresholds
	Discrete  Mask_Discrete
}

func (mask *Mask) ParseAssertLower(b uint16) {
	lsb := uint8(b & 0x00ff) // Least Significant Byte
	msb := uint8(b >> 8)     // Most Significant Byte

	mask.Discrete.Event_14_Assert = isBit6Set(lsb)
	mask.Discrete.Event_13_Assert = isBit5Set(lsb)
	mask.Discrete.Event_12_Assert = isBit4Set(lsb)
	mask.Discrete.Event_11_Assert = isBit3Set(lsb)
	mask.Discrete.Event_10_Assert = isBit2Set(lsb)
	mask.Discrete.Event_9_Assert = isBit1Set(lsb)
	mask.Discrete.Event_8_Assert = isBit0Set(lsb)
	mask.Discrete.Event_7_Assert = isBit7Set(msb)
	mask.Discrete.Event_6_Assert = isBit6Set(msb)
	mask.Discrete.Event_5_Assert = isBit5Set(msb)
	mask.Discrete.Event_4_Assert = isBit4Set(msb)
	mask.Discrete.Event_3_Assert = isBit3Set(msb)
	mask.Discrete.Event_2_Assert = isBit2Set(msb)
	mask.Discrete.Event_1_Assert = isBit1Set(msb)
	mask.Discrete.Event_0_Assert = isBit0Set(msb)

	mask.Threshold.LNR.StatusReturned = isBit6Set(lsb)
	mask.Threshold.LCR.StatusReturned = isBit5Set(lsb)
	mask.Threshold.LNC.StatusReturned = isBit4Set(lsb)
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

	mask.Discrete.Event_14_Deassert = isBit6Set(lsb)
	mask.Discrete.Event_13_Deassert = isBit5Set(lsb)
	mask.Discrete.Event_12_Deassert = isBit4Set(lsb)
	mask.Discrete.Event_11_Deassert = isBit3Set(lsb)
	mask.Discrete.Event_10_Deassert = isBit2Set(lsb)
	mask.Discrete.Event_9_Deassert = isBit1Set(lsb)
	mask.Discrete.Event_8_Deassert = isBit0Set(lsb)
	mask.Discrete.Event_7_Deassert = isBit7Set(msb)
	mask.Discrete.Event_6_Deassert = isBit6Set(msb)
	mask.Discrete.Event_5_Deassert = isBit5Set(msb)
	mask.Discrete.Event_4_Deassert = isBit4Set(msb)
	mask.Discrete.Event_3_Deassert = isBit3Set(msb)
	mask.Discrete.Event_2_Deassert = isBit2Set(msb)
	mask.Discrete.Event_1_Deassert = isBit1Set(msb)
	mask.Discrete.Event_0_Deassert = isBit0Set(msb)

	mask.Threshold.UNR.StatusReturned = isBit6Set(lsb)
	mask.Threshold.UCR.StatusReturned = isBit5Set(lsb)
	mask.Threshold.UNC.StatusReturned = isBit4Set(lsb)

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

	mask.Discrete.Reading_14_Supported = isBit6Set(lsb)
	mask.Discrete.Reading_13_Supported = isBit5Set(lsb)
	mask.Discrete.Reading_12_Supported = isBit4Set(lsb)
	mask.Discrete.Reading_11_Supported = isBit3Set(lsb)
	mask.Discrete.Reading_10_Supported = isBit2Set(lsb)
	mask.Discrete.Reading_9_Supported = isBit1Set(lsb)
	mask.Discrete.Reading_8_Supported = isBit0Set(lsb)
	mask.Discrete.Reading_7_Supported = isBit7Set(msb)
	mask.Discrete.Reading_6_Supported = isBit6Set(msb)
	mask.Discrete.Reading_5_Supported = isBit5Set(msb)
	mask.Discrete.Reading_4_Supported = isBit4Set(msb)
	mask.Discrete.Reading_3_Supported = isBit3Set(msb)
	mask.Discrete.Reading_2_Supported = isBit2Set(msb)
	mask.Discrete.Reading_1_Supported = isBit1Set(msb)
	mask.Discrete.Reading_0_Supported = isBit0Set(msb)

	mask.Threshold.UNR.Settable = isBit5Set(lsb)
	mask.Threshold.UCR.Settable = isBit4Set(lsb)
	mask.Threshold.UNC.Settable = isBit3Set(lsb)
	mask.Threshold.LNR.Settable = isBit2Set(lsb)
	mask.Threshold.LCR.Settable = isBit1Set(lsb)
	mask.Threshold.LNC.Settable = isBit0Set(lsb)
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

	// Desseartion Events
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

// SensorCapbilitites represent the capabilities of the sensor.
// SDRs of Full/Compact record type has this field.
type SensorCapabilitites struct {
	// [7] - 1b = IgnoreWithEntity sensor if Entity is not present or disabled. 0b = don't ignore sensor
	IgnoreWithEntity bool

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
	// using this bit to report settable sensors is optional. I.e. it is
	// ok to report a settable sensor as 'not settable' in the
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
	// Reports how this sensor comes up on device power up and hardware/cold reset.
	// The Initialization Agent does not use this bit. This bit solely reports to software
	// how the sensor comes prior to being initialized by the Initialization Agent.

	// 0b = event generation disabled, 1b = event generation enabled
	EventGenerationEnabled bool
	// 0b = sensor scanning disabled, 1b = sensor scanning enabled
	SensorScanningEnabled bool
}
