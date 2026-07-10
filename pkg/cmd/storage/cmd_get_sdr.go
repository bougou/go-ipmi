package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 33.12 Get SDR Command
type GetSDRRequest struct {
	ReservationID uint16 // LS Byte first
	RecordID      uint16 // LS Byte first
	ReadOffset    uint8  // Offset into record
	ReadBytes     uint8  // FFh means read entire record
}

type GetSDRResponse struct {
	NextRecordID uint16
	RecordData   []byte
}

func (req *GetSDRRequest) Pack() []byte {
	msg := make([]byte, 6)
	types.PackUint16L(req.ReservationID, msg, 0)
	types.PackUint16L(req.RecordID, msg, 2)
	types.PackUint8(req.ReadOffset, msg, 4)
	types.PackUint8(req.ReadBytes, msg, 5)
	return msg
}

func (req *GetSDRRequest) Unpack(msg []byte) error {
	if len(msg) < 6 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 6)
	}
	req.ReservationID, _, _ = types.UnpackUint16L(msg, 0)
	req.RecordID, _, _ = types.UnpackUint16L(msg, 2)
	req.ReadOffset, _, _ = types.UnpackUint8(msg, 4)
	req.ReadBytes, _, _ = types.UnpackUint8(msg, 5)
	return nil
}

func (req *GetSDRRequest) Command() types.Command {
	return types.CommandGetSDR
}

func (res *GetSDRResponse) Pack() []byte {
	out := make([]byte, 2+len(res.RecordData))
	types.PackUint16L(res.NextRecordID, out, 0)
	if len(res.RecordData) > 0 {
		types.PackBytes(res.RecordData, out, 2)
	}
	return out
}

func (res *GetSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.NextRecordID, _, _ = types.UnpackUint16L(msg, 0)
	res.RecordData, _, _ = types.UnpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSDRResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSDRResponse) Format() string {
	return fmt.Sprintf("%v", res)
}
