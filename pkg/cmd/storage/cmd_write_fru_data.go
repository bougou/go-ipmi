package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 34.3 Write FRU Data Command
type WriteFRUDataRequest struct {
	FRUDeviceID uint8
	WriteOffset uint16
	WriteData   []byte
}

type WriteFRUDataResponse struct {
	CountWritten uint8
}

func (req *WriteFRUDataRequest) Command() ipmi.Command {
	return ipmi.CommandWriteFRUData
}

func (req *WriteFRUDataRequest) Pack() []byte {
	out := make([]byte, 3+len(req.WriteData))
	ipmi.PackUint8(req.FRUDeviceID, out, 0)
	ipmi.PackUint16L(req.WriteOffset, out, 1)
	if len(req.WriteData) > 0 {
		ipmi.PackBytes(req.WriteData, out, 3)
	}
	return out
}

func (res *WriteFRUDataResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.CountWritten, _, _ = ipmi.UnpackUint8(msg, 0)
	return nil
}

func (r *WriteFRUDataResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "write-protected offset",
		0x81: "FRU device busy",
	}
}

func (res *WriteFRUDataResponse) Format() string {
	return fmt.Sprintf("Count written : %d", res.CountWritten)
}

// The command writes the specified byte or word to the FRU Inventory Info area. This is a low level direct interface to a non-volatile storage area. This means that the interface does not interpret or check any semantics or formatting for the data being written.
