package ipmi

import (
	"fmt"
	"time"
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

	switch sel.RecordType.Range() {
	case SELRecordTypeRangeStandard:
		if sel.Standard != nil {
			msg = append(msg[0:3], sel.Standard.Pack()...)
		}

	case SELRecordTypeRangeTimestampedOEM:
		if sel.OEMTimestamped != nil {
			msg = append(msg[0:3], sel.OEMTimestamped.Pack()...)
		}

	case SELRecordTypeRangeNonTimestampedOEM:
		if sel.OEMNonTimestamped != nil {
			msg = append(msg[0:3], sel.OEMNonTimestamped.Pack()...)
		}
	}

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
			return nil, fmt.Errorf("parseSELDefault failed, err: %w", err)
		}
	case SELRecordTypeRangeTimestampedOEM:
		if err := parseSELOEMTimestamped(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELOEMTimestamped failed, err: %w", err)
		}
	case SELRecordTypeRangeNonTimestampedOEM:
		if err := parseSELOEMNonTimestamped(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELOEMNonTimestamped failed, err: %w", err)
		}
	}
	return sel, nil
}

// 32.2 OEM SEL Record - Type C0h-DFh
type SELOEMTimestamped struct {
	Timestamp      time.Time // Time when event was logged. uint32 LS byte first.
	ManufacturerID uint32    // only 3 bytes
	OEMDefined     [6]byte
}

func (oemTimestamped *SELOEMTimestamped) Pack() []byte {
	var msg = make([]byte, 13)
	packUint32L(uint32(oemTimestamped.Timestamp.Unix()), msg, 0)
	packUint24L(oemTimestamped.ManufacturerID, msg, 4)
	packBytes(oemTimestamped.OEMDefined[:], msg, 7)
	return msg
}

type SELOEMNonTimestamped struct {
	OEM [13]byte
}

func (oemNonTimestamped *SELOEMNonTimestamped) Pack() []byte {
	return oemNonTimestamped.OEM[:]
}

// 32.1 SEL Standard Event Records
type SELStandard struct {
	Timestamp    time.Time    // Time when event was logged. uint32 LS byte first.
	GeneratorID  GeneratorID  // RqSA & LUN if event was generated from IPMB. Software ID if event was generated from system software.
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

func (standard *SELStandard) Pack() []byte {
	var msg = make([]byte, 13)

	packUint32L(uint32(standard.Timestamp.Unix()), msg, 0)

	packUint16L(uint16(standard.GeneratorID), msg, 4)

	packUint8(standard.EvMRev, msg, 6)
	packUint8(uint8(standard.SensorType), msg, 7)
	packUint8(uint8(standard.SensorNumber), msg, 8)

	var eventType = uint8(standard.EventReadingType)
	if standard.EventDir {
		eventType = eventType | 0x80
	}
	packUint8(eventType, msg, 9)

	packUint8(standard.EventData.EventData1, msg, 10)
	packUint8(standard.EventData.EventData2, msg, 11)
	packUint8(standard.EventData.EventData3, msg, 12)

	return msg
}

// EventString return string description of the event.
func (sel *SELStandard) EventString() string {
	return sel.EventReadingType.EventString(sel.SensorType, sel.EventData)
}

func (sel *SELStandard) EventSeverity() EventSeverity {
	return sel.EventReadingType.EventSeverity(sel.SensorType, sel.EventData, sel.EventDir)
}

func parseSELDefault(msg []byte, sel *SEL) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShortWith(len(msg), 16)
	}

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
	if len(msg) < 16 {
		return ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	var s = &SELOEMTimestamped{}
	sel.OEMTimestamped = s

	ts, _, _ := unpackUint32L(msg, 3)
	s.Timestamp = parseTimestamp(ts)

	id, _, _ := unpackUint24L(msg, 7)
	s.ManufacturerID = id

	s.OEMDefined = [6]byte{}
	b, _, _ := unpackBytes(msg, 10, 6)
	for i := 0; i < 6; i++ {
		s.OEMDefined[i] = b[i]
	}

	return nil
}

func parseSELOEMNonTimestamped(msg []byte, sel *SEL) error {
	if len(msg) < 16 {
		return ErrUnpackedDataTooShortWith(len(msg), 16)
	}

	var s = &SELOEMNonTimestamped{}
	sel.OEMNonTimestamped = s

	s.OEM = [13]byte{}
	b, _, _ := unpackBytes(msg, 3, 13)
	for i := 0; i < 6; i++ {
		s.OEM[i] = b[i]
	}

	return nil
}

// FormatSELs print sel records in table format.
// The second sdrMap is optional. If the sdrMap is not nil,
// it will also print sensor number, entity id and instance, and asserted discrete states.
// The sdrMap can be fetched by GetSDRsMap method.
func FormatSELs(records []*SEL, sdrMap SDRMapBySensorNumber) string {
	var elistMode bool // extend list
	if sdrMap != nil {
		elistMode = true
	}

	rows := make([]map[string]string, 0)

	for _, sel := range records {
		recordTypeRange := sel.RecordType.Range()

		switch recordTypeRange {
		case SELRecordTypeRangeStandard:
			s := sel.Standard

			row := map[string]string{
				"ID":               fmt.Sprintf("%#04x", sel.RecordID),
				"RecordType":       sel.RecordType.String(),
				"EvmRev":           fmt.Sprintf("%#02x", s.EvMRev),
				"Timestamp":        fmt.Sprintf("%v", s.Timestamp),
				"GID":              fmt.Sprintf("%#04x", uint16(s.GeneratorID)),
				"SensorNumber":     fmt.Sprintf("%#02x", s.SensorNumber),
				"SensorTypeCode":   fmt.Sprintf("%#02x", uint8(s.SensorType)),
				"SensorType":       s.SensorType.String(),
				"EventReadingType": fmt.Sprintf("%#02x (%s)", uint8(s.EventReadingType), s.EventReadingType.String()),
				"EventDescription": s.EventString(),
				"EventDirection":   s.EventDir.String(),
				"EventSeverity":    string(s.EventSeverity()),
				"EventData":        s.EventData.String(),
			}

			if elistMode {
				var sensorName string
				sdr, ok := sdrMap[s.GeneratorID][s.SensorNumber]
				if !ok {
					sensorName = fmt.Sprintf("N/A %#04x, %#02x", uint16(s.GeneratorID), s.SensorNumber)
				} else {
					sensorName = sdr.SensorName()
				}
				row["SensorName"] = sensorName
			}

			rows = append(rows, row)

		case SELRecordTypeRangeTimestampedOEM:
		case SELRecordTypeRangeNonTimestampedOEM:
		}
	}

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
		"EventDescription",
		"EventDirection",
		"EventSeverity",
		"EventData",
	}
	if elistMode {
		headers = append(headers, "SensorName")
	}

	return RenderTable(headers, rows)
}
