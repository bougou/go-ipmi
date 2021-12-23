package ipmi

import "fmt"

// 31.6 Add SEL Entry Command
type AddSELEntryRequest struct {
	SEL *SEL
}

type AddSELEntryResponse struct {
	CompletionCode uint8
	RecordID       uint16 // Record ID for added record, LS Byte first
}

func (req *AddSELEntryRequest) Command() Command {
	return CommandAddSELEntry
}

func (req *AddSELEntryRequest) Pack() []byte {
	return req.SEL.Pack()
}

func (res *AddSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShort
	}
	res.CompletionCode, _, _ = unpackUint8(msg, 0)
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
	return fmt.Sprintf("%v", res)
}

func (c *Client) AddSELEntry(sel *SEL) (response *AddSELEntryResponse, err error) {
	request := &AddSELEntryRequest{
		SEL: sel,
	}
	response = &AddSELEntryResponse{}
	err = c.Exchange(request, response)
	return
}
