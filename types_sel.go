package ipmi

import (
	"bytes"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
)

// 32. SEL Record Formats
type SEL struct {
	// SEL Record IDs 0000h and FFFFh are reserved for functional use and are not legal ID values.
	// Record IDs are handles. They are not required to be sequential or consecutive.
	// Applications should not assume that SEL Record IDs will follow any particular numeric ordering.
	RecordID   uint16
	RecordType SELRecordType

	Standard          *SELStandard
	OEMTimestamped    *SELOEMTimestamped
	OEMNonTimestamped *SELOEMNonTimestamped
}

func (sel *SEL) Pack() []byte {
	msg := make([]byte, 16)
	packUint16L(sel.RecordID, msg, 0)
	packUint8(uint8(sel.RecordType), msg, 2)

	packUint32L(uint32(sel.Standard.Timestamp.Unix()), msg, 3)

	packUint16L(uint16(sel.Standard.GeneratorID), msg, 7)

	packUint8(sel.Standard.EvMRev, msg, 9)
	packUint8(uint8(sel.Standard.SensorType), msg, 10)
	packUint8(uint8(sel.Standard.SensorNumber), msg, 11)

	var eventType = uint8(sel.Standard.EventReadingType)
	if sel.Standard.EventDir {
		eventType = eventType | 0x80
	}
	packUint8(eventType, msg, 12)

	packUint8(sel.Standard.EventData.EventData1, msg, 13)
	packUint8(sel.Standard.EventData.EventData2, msg, 14)
	packUint8(sel.Standard.EventData.EventData3, msg, 15)

	return msg
}

func ParseSEL(msg []byte) (*SEL, error) {
	if len(msg) != 16 {
		return nil, fmt.Errorf("SEL Record msg should be 16 bytes in length")
	}

	recordID, _, _ := unpackUint16L(msg, 0)
	recordType, _, _ := unpackUint8(msg, 2)
	sel := &SEL{
		RecordID:   recordID,
		RecordType: SELRecordType(recordType),
	}

	recordTypeRange := sel.RecordType.Range()
	switch recordTypeRange {
	case SELRecordTypeRangeStandard:
		if err := parseSELDefault(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELDefault failed, err: %s", err)
		}
	case SELRecordTypeRangeTimestampedOEM:
		if err := parseSELOEMTimestamped(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELOEMTimestamped failed, err: %s", err)
		}
	case SELRecordTypeRangeNonTimestampedOEM:
		if err := parseSELOEMNonTimestamped(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELOEMNonTimestamped failed, err: %s", err)
		}
	}
	return sel, nil
}

// 32.2 OEM SEL Record - Type C0h-DFh
type SELOEMTimestamped struct {
	Timestamp      time.Time // Time when event was logged. uint32 LS byte first.
	ManufacturerID uint32    // only 3 bytes
	OEMDefined     []byte    // 6 bytes
}

type SELOEMNonTimestamped struct {
	OEM []byte // 13 bytes
}

// 32.1 SEL Standard Event Records
type SELStandard struct {
	Timestamp    time.Time    // Time when event was logged. uint32 LS byte first.
	GeneratorID  GeneratorID  // RqSA & LUN if event was generated from IPMB. Software ID if event was generatedfrom system software.
	EvMRev       uint8        // Event Message Revision (format version)
	SensorType   SensorType   // Sensor Type Code for sensor that generated the event
	SensorNumber SensorNumber // Number of sensor that generated the event

	EventDir         EventDir         // Event Direction. [7] -0b = Assertion event. 1b = Deassertion event.
	EventReadingType EventReadingType // Type of trigger for the event. [6:0] - Event Type Code

	// 29.7 Event Data Field Formats
	//
	// The sensor class determines the corresponding Event Data format.
	// The sensor class can be extracted from EventReadingType.
	EventData EventData
}

// EventString return string description of the event.
func (sel *SELStandard) EventString() string {
	return sel.EventReadingType.EventString(sel.SensorType, sel.SensorNumber, sel.EventData)
}

func parseSELDefault(msg []byte, sel *SEL) error {
	var s = &SELStandard{}
	sel.Standard = s

	ts, _, _ := unpackUint32L(msg, 3)
	s.Timestamp = parseTimestamp(ts)

	gid, _, _ := unpackUint16L(msg, 7)
	s.GeneratorID = GeneratorID(gid)

	s.EvMRev, _, _ = unpackUint8(msg, 9)

	sensorType, _, _ := unpackUint8(msg, 10)
	s.SensorType = SensorType(sensorType)

	sensorNumber, _, _ := unpackUint8(msg, 11)
	s.SensorNumber = SensorNumber(sensorNumber)

	b, _, _ := unpackUint8(msg, 12)
	s.EventDir = EventDir(isBit7Set(b))
	s.EventReadingType = EventReadingType(b & 0x7f) // clear bit 7

	s.EventData.EventData1, _, _ = unpackUint8(msg, 13)
	s.EventData.EventData2, _, _ = unpackUint8(msg, 14)
	s.EventData.EventData3, _, _ = unpackUint8(msg, 15)

	return nil
}

func parseSELOEMTimestamped(msg []byte, sel *SEL) error {
	var s = &SELOEMTimestamped{}
	sel.OEMTimestamped = s

	ts, _, _ := unpackUint32L(msg, 3)
	s.Timestamp = parseTimestamp(ts)

	id, _, _ := unpackUint24L(msg, 7)
	s.ManufacturerID = id

	s.OEMDefined, _, _ = unpackBytes(msg, 10, 6)
	return nil
}

func parseSELOEMNonTimestamped(msg []byte, sel *SEL) error {
	var s = &SELOEMNonTimestamped{}
	sel.OEMNonTimestamped = s

	s.OEM, _, _ = unpackBytes(msg, 3, 13)
	return nil
}

// FormatSELs print sel records in table format.
// The second sdrMap is optional. If the sdrMap is not nil,
// it will also print sensor number, entity id and instance, and asserted discrete states.
// sdrMap can be get by client GetSDRsMap method.
func FormatSELs(records []*SEL, sdrMap SDRMapBySensorNumber) string {
	var elistMode bool
	if sdrMap != nil {
		elistMode = true
	}

	var buf = new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)

	headers := []string{
		"ID",
		"RecordType",
		"EvmRev",
		"Timestamp",
		"GID",
		"SensorNumber",
		"SensorTypeCode",
		"SensorType",
		"EventReadingType",
		"EventReadingType",
		"EventDescription",
		"EventDir",
		"EventData",
	}
	if elistMode {
		headers = append(headers, "SensorName")
	}

	table.SetHeader(headers)
	table.SetFooter(headers)

	for _, sel := range records {
		recordTypeRange := sel.RecordType.Range()

		switch recordTypeRange {
		case SELRecordTypeRangeStandard:
			s := sel.Standard

			content := []string{
				fmt.Sprintf("%#04x", sel.RecordID),
				sel.RecordType.String(),
				fmt.Sprintf("%#02x", s.EvMRev),
				fmt.Sprintf("%v", s.Timestamp),
				fmt.Sprintf("%#04x", s.GeneratorID),
				fmt.Sprintf("%#02x", s.SensorNumber),
				fmt.Sprintf("%#02x", uint8(s.SensorType)),
				s.SensorType.String(),
				fmt.Sprintf("%#02x", uint8(s.EventReadingType)),
				s.EventReadingType.String(),
				s.EventString(),
				s.EventDir.String(),
				s.EventData.String(),
			}

			if elistMode {
				var sensorName string
				sdr, ok := sdrMap[s.GeneratorID][s.SensorNumber]
				if !ok {
					sensorName = fmt.Sprintf("N/A %#04x, %#02x", s.GeneratorID, s.SensorNumber)
				} else {
					sensorName = sdr.SensorName()
				}
				content = append(content, sensorName)
			}

			table.Append(content)

		case SELRecordTypeRangeTimestampedOEM:
		case SELRecordTypeRangeNonTimestampedOEM:
		}
	}

	table.Render()

	return buf.String()
}
