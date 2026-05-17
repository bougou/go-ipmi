package storage

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
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

func (req *GetDeviceSDRRequest) Command() ipmi.Command {
	return ipmi.CommandGetDeviceSDR
}

func (req *GetDeviceSDRRequest) Pack() []byte {
	out := make([]byte, 6)
	ipmi.PackUint16L(req.ReservationID, out, 0)
	ipmi.PackUint16L(req.RecordID, out, 2)
	ipmi.PackUint8(req.ReadOffset, out, 4)
	ipmi.PackUint8(req.ReadBytes, out, 5)
	return out
}

func (res *GetDeviceSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.NextRecordID, _, _ = ipmi.UnpackUint16L(msg, 0)
	res.RecordData, _, _ = ipmi.UnpackBytes(msg, 2, len(msg)-2)
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

// The Get Device SDR command allows SDR information for sensors for a Sensor Device
// (typically implemented in a satellite management controller) to be returned.
//
// The Get Device SDR Command can return any type of SDR, not just Types 01h and 02h.
// This is an optional command for Static Sensor Devices, and mandatory for Dynamic Sensor Devices.
// The format and action of this command is similar to that for the Get SDR command
// for SDR Repository Devices.
//
// Sensor Devices that support the Get Device SDR command return SDR Records that
// match the SDR Repository formats.

// getDeviceSDR reads the Device SDR record in partial read way.

// the actual data length of the SDR can only be determined after the first GetSDR request/response.

// determine the total data length by parsing the SDR Header part

// decrease the readBytes for the last read.
