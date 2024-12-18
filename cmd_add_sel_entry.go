package ipmi

import (
	"context"
	"fmt"
)

// 31.6 Add SEL Entry Command
type AddSELEntryRequest struct {
	SEL *SEL
}

type AddSELEntryResponse struct {
	RecordID uint16 // Record ID for added record, LS Byte first
}

func (req *AddSELEntryRequest) Command() Command {
	return CommandAddSELEntry
}

func (req *AddSELEntryRequest) Pack() []byte {
	return req.SEL.Pack()
}

func (res *AddSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.RecordID, _, _ = unpackUint16L(msg, 0)
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

func (c *Client) AddSELEntry(ctx context.Context, sel *SEL) (response *AddSELEntryResponse, err error) {
	request := &AddSELEntryRequest{
		SEL: sel,
	}
	response = &AddSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
