package storage

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 31.6 Add SEL Entry Command
type AddSELEntryRequest struct {
	SEL *types.SEL
}

type AddSELEntryResponse struct {
	RecordID uint16 // Record ID for added record, LS Byte first
}

func (req *AddSELEntryRequest) Command() types.Command {
	return types.CommandAddSELEntry
}

func (req *AddSELEntryRequest) Pack() []byte {
	return req.SEL.Pack()
}

func (res *AddSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.RecordID, _, _ = types.UnpackUint16L(msg, 0)
	return nil
}

func (res *AddSELEntryResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "operation not supported for this Record Type",
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *AddSELEntryResponse) Format() string {
	return fmt.Sprintf("Record ID : %d (%#02x)", res.RecordID, res.RecordID)
}
