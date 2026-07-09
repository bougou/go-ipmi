package storage

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// 35.3 Get Device SDR Command
type GetDeviceSDRRequest struct {
	ReservationID uint16
	RecordID      uint16
	ReadOffset    uint8
	ReadBytes     uint8 // FFh means read entire record
}

type GetDeviceSDRResponse struct {
	NextRecordID uint16
	RecordData   []byte
}

func (req *GetDeviceSDRRequest) Command() types.Command {
	return types.CommandGetDeviceSDR
}

func (req *GetDeviceSDRRequest) Pack() []byte {
	out := make([]byte, 6)
	types.PackUint16L(req.ReservationID, out, 0)
	types.PackUint16L(req.RecordID, out, 2)
	types.PackUint8(req.ReadOffset, out, 4)
	types.PackUint8(req.ReadBytes, out, 5)
	return out
}

func (res *GetDeviceSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.NextRecordID, _, _ = types.UnpackUint16L(msg, 0)
	res.RecordData, _, _ = types.UnpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (r *GetDeviceSDRResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "record changed",
	}
}

func (res *GetDeviceSDRResponse) Format() string {
	return ""
}
