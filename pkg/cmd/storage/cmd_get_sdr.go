package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
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
	ipmi.PackUint16L(req.ReservationID, msg, 0)
	ipmi.PackUint16L(req.RecordID, msg, 2)
	ipmi.PackUint8(req.ReadOffset, msg, 4)
	ipmi.PackUint8(req.ReadBytes, msg, 5)
	return msg
}

func (req *GetSDRRequest) Command() ipmi.Command {
	return ipmi.CommandGetSDR
}

func (res *GetSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.NextRecordID, _, _ = ipmi.UnpackUint16L(msg, 0)
	res.RecordData, _, _ = ipmi.UnpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSDRResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSDRResponse) Format() string {
	return fmt.Sprintf("%v", res)
}
