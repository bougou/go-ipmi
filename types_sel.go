package ipmi

import (
	"fmt"
	"time"
)

type SELRecordType uint8

const (
	SELRecordTypeDefault = SELRecordType(0x02)

	// These records are automatically timestamped by the SEL Device.
	SELRecordTypeOEMTimestampedL = SELRecordType(0xC0)
	SELRecordTypeOEMTimestampedH = SELRecordType(0xDF)

	// The SEL Device does not automatically timestamp these records.
	// The four bytes passed in the byte locations for the timestamp will be directly entered into the SEL.
	SELRecordTypeOEMNonTimestampedL = SELRecordType(0xE0)
	SELRecordTypeOEMNonTimestampedH = SELRecordType(0xFF)
)

// 32.1 SEL Event Records
// Each SEL record is 16 bytes in length.
type SEL struct {
	// SEL Record IDs 0000h and FFFFh are reserved for functional use and are not legal ID values.
	// Record IDs are handles. They are not required to be sequential or consecutive.
	// Applications should not assume that SEL Record IDs will follow any particular numeric ordering.
	RecordID     uint16
	RecordType   EventRecordType
	Timestamp    time.Time    // Time when event was logged. uint32 LS byte first.
	GeneratorID  uint16       // RqSA & LUN if event was generated from IPMB. Software ID if event was generatedfrom system software.
	EvMRev       uint8        // Event Message format version
	SensorType   SensorType   // Sensor Type Code for sensor that generated the event
	SensorNumber SensorNumber // Number of sensor that generated the event

	EventDir  EventDir  // Event Direction. [7] -0b = Assertion event. 1b = Deassertion event.
	EventType EventType // Type of trigger for the event. [6:0] -Event Type Code

	EventData1 uint8
	EventData2 uint8
	EventData3 uint8
}

// 37 Timestamp Format
func parseTimestamp(timestamp uint32) time.Time {
	return time.Unix(int64(timestamp), 0)
}

func (sel *SEL) StringHeader() string {
	return fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s",
		"RecordID",
		"RecordType",
		"Timestamp",
		"GeneratorID",
		"EvmRev",
		"SensorType #SensorNumber",
		"EventDir",
		"EventType",
		"EventData1",
		"EventData2",
		"EventData3",
	)

}

func (sel *SEL) String() string {
	return fmt.Sprintf("%0x | %s | %s | %02x | %02x | %s #%#02x | %s | %02x | %02x | %02x | %02x",
		sel.RecordID,
		sel.RecordType,
		sel.Timestamp,
		sel.GeneratorID,
		sel.EvMRev,
		sel.SensorType,
		sel.SensorNumber,
		sel.EventDir,
		sel.EventType,
		sel.EventData1,
		sel.EventData2,
		sel.EventData3,
	)
}

func (sel *SEL) Pack() []byte {
	msg := make([]byte, 16)
	packUint16L(sel.RecordID, msg, 0)
	packUint8(uint8(sel.RecordType), msg, 2)
	packUint32L(uint32(sel.Timestamp.Unix()), msg, 3)
	packUint16L(sel.GeneratorID, msg, 7)
	packUint8(sel.EvMRev, msg, 9)
	packUint8(uint8(sel.SensorType), msg, 10)
	packUint8(uint8(sel.SensorNumber), msg, 11)

	var eventType = uint8(sel.EventType)
	if sel.EventDir {
		eventType = eventType | 0x80
	}
	packUint8(eventType, msg, 12)

	packUint8(sel.EventData1, msg, 13)
	packUint8(sel.EventData2, msg, 14)
	packUint8(sel.EventData3, msg, 15)
	return msg
}

func unpackSEL(msg []byte) (*SEL, error) {
	if len(msg) != 16 {
		return nil, fmt.Errorf("SEL Record msg should be 16 bytes in length")
	}

	sel := &SEL{}
	sel.RecordID, _, _ = unpackUint16L(msg, 0)

	recordType, _, _ := unpackUint8(msg, 2)
	sel.RecordType = EventRecordType(recordType)

	ts, _, _ := unpackUint32L(msg, 3)
	sel.Timestamp = parseTimestamp(ts)

	sel.GeneratorID, _, _ = unpackUint16L(msg, 7)
	sel.EvMRev, _, _ = unpackUint8(msg, 9)

	sensorType, _, _ := unpackUint8(msg, 10)
	sel.SensorType = SensorType(sensorType)

	sensorNumber, _, _ := unpackUint8(msg, 11)
	sel.SensorNumber = SensorNumber(sensorNumber)

	b, _, _ := unpackUint8(msg, 12)
	sel.EventDir = b&0x80 == 0x80
	sel.EventType = EventType(b & 0x7f) // clear bit 7

	sel.EventData1, _, _ = unpackUint8(msg, 13)
	sel.EventData2, _, _ = unpackUint8(msg, 14)
	sel.EventData3, _, _ = unpackUint8(msg, 15)
	return sel, nil
}

type SELRecordDefault struct {
}

type SELRecordOEMTimestamped struct {
	RecordID       uint16        `json:"record_id"`
	RecordType     SELRecordType `json:"record_type"` // 0xC0 - 0xDF
	Timestamp      time.Time     `json:"timestamp"`
	ManufacturerID uint16        `json:"manufacturer_id"`
}

type SELRecordOEMNonTimestamped struct {
	RecordID   uint16        `json:"record_id"`
	RecordType SELRecordType `json:"record_type"` // 0xE0 - 0xFF
}
