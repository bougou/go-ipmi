package storage

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 31.8 Delete SEL Entry Command
type DeleteSELEntryRequest struct {
	ReservationID uint16
	RecordID      uint16
}

type DeleteSELEntryResponse struct {
	RecordID uint16
}

func (req *DeleteSELEntryRequest) Command() ipmi.Command {
	return ipmi.CommandDeleteSELEntry
}

func (req *DeleteSELEntryRequest) Pack() []byte {
	out := make([]byte, 4)
	ipmi.PackUint16L(req.ReservationID, out, 0)
	ipmi.PackUint16L(req.RecordID, out, 2)
	return out
}

func (res *DeleteSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.RecordID, _, _ = ipmi.UnpackUint16L(msg, 0)
	return nil
}

func (res *DeleteSELEntryResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "operation not supported for this Record Type",
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *DeleteSELEntryResponse) Format() string {
	return fmt.Sprintf("Record ID : %d (%#02x)", res.RecordID, res.RecordID)
}
