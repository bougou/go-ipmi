package ipmi

import (
	"fmt"
	"time"
)

// type SELRecordType uint8

// const (
// 	SELRecordTypeDefault = SELRecordType(0x02)

// 	// These records are automatically timestamped by the SEL Device.
// 	SELRecordTypeOEMTimestampedL = SELRecordType(0xC0)
// 	SELRecordTypeOEMTimestampedH = SELRecordType(0xDF)

// 	// The SEL Device does not automatically timestamp these records.
// 	// The four bytes passed in the byte locations for the timestamp will be directly entered into the SEL.
// 	SELRecordTypeOEMNonTimestampedL = SELRecordType(0xE0)
// 	SELRecordTypeOEMNonTimestampedH = SELRecordType(0xFF)
// )

type SEL struct {
	// SEL Record IDs 0000h and FFFFh are reserved for functional use and are not legal ID values.
	// Record IDs are handles. They are not required to be sequential or consecutive.
	// Applications should not assume that SEL Record IDs will follow any particular numeric ordering.
	RecordID   uint16
	RecordType EventRecordType

	Default           *SELDefault
	OEMTimestamped    *SELOEMTimestamped
	OEMNonTimestamped *SELOEMNonTimestamped
}

func (sel *SEL) StringHeader() string {
	formatValues := []formatValue{
		fv("%-6s", "ID"),
		fv("%-15s", "RecordType"),
		fv("%-6s", "EvmRev"),
		fv("%-29s", "Timestamp"),
		fv("%-6s", "GID"),
		fv("%-7s", "Sensor#"),
		fv("%-11s", "SensorClass"),
		fv("%-27s", "SensorType"),
		fv("%-6s", "STCode"),
		fv("%-6s", "ERType"),
		fv("%-15s", "ERCategory"),
		fv("%-6s", "Offset"),
		fv("%-64s", "EventDescription"),
		fv("%-11s", "EventDir"),
		fv("%-10s", "EventData1"),
		fv("%-10s", "EventData2"),
		fv("%-10s", "EventData3"),
	}
	return formatValuesTable(formatValues)
}

func (sel *SEL) Format() string {
	formatValues := []formatValue{
		fv("%-6s", fmt.Sprintf("%#04x", sel.RecordID)),
		fv("%-15s", sel.RecordType),
	}

	if sel.RecordType == EventRecordTypeSystemEvent {
		formatValues = append(formatValues, sel.Default.Format()...)
	}
	if isEventRecordTypeOEMTimestamped(sel.RecordType) {
		formatValues = append(formatValues, sel.OEMTimestamped.Format()...)
	}
	if isEventRecordTypeOEMNonTimestamped(sel.RecordType) {
		formatValues = append(formatValues, sel.OEMNonTimestamped.Format()...)
	}
	return formatValuesTable(formatValues)
}

func (sel *SEL) Pack() []byte {
	msg := make([]byte, 16)
	packUint16L(sel.RecordID, msg, 0)
	packUint8(uint8(sel.RecordType), msg, 2)

	packUint32L(uint32(sel.Default.Timestamp.Unix()), msg, 3)

	packUint16L(uint16(sel.Default.GeneratorID), msg, 7)

	packUint8(sel.Default.EvMRev, msg, 9)
	packUint8(uint8(sel.Default.SensorType), msg, 10)
	packUint8(uint8(sel.Default.SensorNumber), msg, 11)

	var eventType = uint8(sel.Default.EventReadingType)
	if sel.Default.EventDir {
		eventType = eventType | 0x80
	}
	packUint8(eventType, msg, 12)

	packUint8(sel.Default.EventData1, msg, 13)
	packUint8(sel.Default.EventData2, msg, 14)
	packUint8(sel.Default.EventData3, msg, 15)
	return msg
}

func ParseSEL(msg []byte) (*SEL, error) {
	if len(msg) != 16 {
		return nil, fmt.Errorf("SEL Record msg should be 16 bytes in length")
	}

	sel := &SEL{}
	sel.RecordID, _, _ = unpackUint16L(msg, 0)

	recordType, _, _ := unpackUint8(msg, 2)
	sel.RecordType = EventRecordType(recordType)

	if sel.RecordType == EventRecordTypeSystemEvent {
		if err := parseSELDefault(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELDefault failed, err: %s", err)
		}
	} else if isEventRecordTypeOEMTimestamped(sel.RecordType) {
		if err := parseSELOEMTimestamped(msg, sel); err != nil {
			return nil, fmt.Errorf("parseSELOEMTimestamped failed, err: %s", err)
		}
	} else if isEventRecordTypeOEMNonTimestamped(sel.RecordType) {
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

// 32.1 SELDefault Event Records
// Each SELDefault record is 16 bytes in length.
type SELDefault struct {
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
	EventData1 uint8
	EventData2 uint8
	EventData3 uint8
}

// 29.7
// Event Data 1
// [3:0] - Offset from Event/Reading Code for threshold event.
func (sel *SELDefault) EventReadingOffset() uint8 {
	return sel.EventData1 & 0x0f
}

// EventString return string description of the event.
func (sel *SELDefault) EventString() string {
	offset := sel.EventReadingOffset()
	return sel.EventReadingType.EventString(sel.SensorType, sel.GeneratorID, sel.SensorNumber, offset)
}

// 37 Timestamp Format
func parseTimestamp(timestamp uint32) time.Time {
	return time.Unix(int64(timestamp), 0)
}

func (s *SELDefault) Format() []formatValue {
	formatValues := []formatValue{
		fv("%-6s", fmt.Sprintf("%#02x", s.EvMRev)),
		fv("%-29s", s.Timestamp),
		fv("%-6s", fmt.Sprintf("%#04x", s.GeneratorID)),
		fv("%-7s", fmt.Sprintf("%#02x", s.SensorNumber)),
		fv("%-11s", s.EventReadingType.SensorClass()),
		fv("%-27s", s.SensorType),
		fv("%-6s", fmt.Sprintf("%#02x", uint8(s.SensorType))),
		fv("%-6s", fmt.Sprintf("%#02x", uint8(s.EventReadingType))),
		fv("%-15s", s.EventReadingType.Category()),
		fv("%-6s", fmt.Sprintf("%#02x", s.EventReadingOffset())),
		fv("%-64s", s.EventString()),
		fv("%-11s", s.EventDir),
		fv("%-10s", fmt.Sprintf("%02x", s.EventData1)),
		fv("%-10s", fmt.Sprintf("%02x", s.EventData2)),
		fv("%-10s", fmt.Sprintf("%02x", s.EventData3)),
	}
	return formatValues
}

func (s *SELOEMTimestamped) Format() []formatValue {
	formatValues := []formatValue{
		fv("%-6s", ""),
		fv("%-29s", s.Timestamp),
		fv("%-6s", ""),
		fv("%-7s", ""),
		fv("%-11s", ""),
		fv("%-27s", ""),
		fv("%-6s", ""),
		fv("%-6s", ""),
		fv("%-15s", ""),
		fv("%-6s", ""),
		fv("%-64s", fmt.Sprintf("%v", s.OEMDefined)),
		fv("%-11s", ""),
		fv("%-10s", ""),
		fv("%-10s", ""),
		fv("%-10s", ""),
	}
	return formatValues
}

func (s *SELOEMNonTimestamped) Format() []formatValue {
	formatValues := []formatValue{
		fv("%-6s", ""),
		fv("%-29s", ""),
		fv("%-6s", ""),
		fv("%-7s", ""),
		fv("%-11s", ""),
		fv("%-27s", ""),
		fv("%-6s", ""),
		fv("%-6s", ""),
		fv("%-15s", ""),
		fv("%-6s", ""),
		fv("%-64s", fmt.Sprintf("%v", s.OEM)),
		fv("%-11s", ""),
		fv("%-10s", ""),
		fv("%-10s", ""),
		fv("%-10s", ""),
	}
	return formatValues
}

func parseSELDefault(msg []byte, sel *SEL) error {
	var s = &SELDefault{}
	sel.Default = s

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
	s.EventDir = b&0x80 == 0x80
	s.EventReadingType = EventReadingType(b & 0x7f) // clear bit 7

	s.EventData1, _, _ = unpackUint8(msg, 13)
	s.EventData2, _, _ = unpackUint8(msg, 14)
	s.EventData3, _, _ = unpackUint8(msg, 15)

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
