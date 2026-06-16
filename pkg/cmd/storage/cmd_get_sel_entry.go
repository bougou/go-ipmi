package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 31.5 Get SEL Entry Command
type GetSELEntryRequest struct {
	// LS Byte first. Only required for partial Get. Use 0000h otherwise.
	ReservationID uint16

	// SEL Record ID, LS Byte first.
	//  0000h = GET FIRST ENTRY
	//  FFFFh = GET LAST ENTRY
	RecordID uint16

	// Offset into record
	Offset uint8

	// FFh means read entire record.
	ReadBytes uint8
}

type GetSELEntryResponse struct {
	NextRecordID uint16
	Data         []byte // Record Data, 16 bytes for entire record, at least 1 byte
}

func (req *GetSELEntryRequest) Command() ipmi.Command {
	return ipmi.CommandGetSELEntry
}

func (req *GetSELEntryRequest) Pack() []byte {
	var msg = make([]byte, 6)
	ipmi.PackUint16L(req.ReservationID, msg, 0)
	ipmi.PackUint16L(req.RecordID, msg, 2)
	ipmi.PackUint8(req.Offset, msg, 4)
	ipmi.PackUint8(req.ReadBytes, msg, 5)
	return msg
}

func (res *GetSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.NextRecordID, _, _ = ipmi.UnpackUint16L(msg, 0)
	res.Data, _, _ = ipmi.UnpackBytesMost(msg, 2, 16)
	return nil
}

func (*GetSELEntryResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *GetSELEntryResponse) Format() string {
	return fmt.Sprintf("%v", res)
}
