package ipmi

import (
	"context"
	"fmt"
)

// 31.8 Delete SEL Entry Command
type DeleteSELEntryRequest struct {
	ReservationID uint16
	RecordID      uint16
}

type DeleteSELEntryResponse struct {
	RecordID uint16
}

func (req *DeleteSELEntryRequest) Command() Command {
	return CommandDeleteSELEntry
}

func (req *DeleteSELEntryRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint16L(req.ReservationID, out, 0)
	packUint16L(req.RecordID, out, 2)
	return out
}

func (res *DeleteSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.RecordID, _, _ = unpackUint16L(msg, 0)
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

func (c *Client) DeleteSELEntry(ctx context.Context, recordID uint16, reservationID uint16) (response *DeleteSELEntryResponse, err error) {
	request := &DeleteSELEntryRequest{
		ReservationID: reservationID,
		RecordID:      recordID,
	}
	response = &DeleteSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
