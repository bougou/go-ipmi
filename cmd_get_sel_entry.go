package ipmi

import (
	"context"
	"fmt"
)

// 31.5 Get SEL Entry Command
type GetSELEntryRequest struct {
	// LS Byte first. Only required for partial Get. Use 0000h otherwise.
	ReservationID uint16

	// SEL Record ID, LS Byte first.
	//  0000h = GET FIRST ENTRY
	//  FFFFh = GET LAST ENTRY
	RecordID uint16

	// Offset into record
	Offset uint8

	// FFh means read entire record.
	ReadBytes uint8
}

type GetSELEntryResponse struct {
	NextRecordID uint16
	Data         []byte // Record Data, 16 bytes for entire record, at least 1 byte
}

func (req *GetSELEntryRequest) Command() Command {
	return CommandGetSELEntry
}

func (req *GetSELEntryRequest) Pack() []byte {
	var msg = make([]byte, 6)
	packUint16L(req.ReservationID, msg, 0)
	packUint16L(req.RecordID, msg, 2)
	packUint8(req.Offset, msg, 4)
	packUint8(req.ReadBytes, msg, 5)
	return msg
}

func (res *GetSELEntryResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.NextRecordID, _, _ = unpackUint16L(msg, 0)
	res.Data, _, _ = unpackBytesMost(msg, 2, 16)
	return nil
}

func (*GetSELEntryResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *GetSELEntryResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// The reservationID is only required for partial Get, use 0000h otherwise.
func (c *Client) GetSELEntry(ctx context.Context, reservationID uint16, recordID uint16) (response *GetSELEntryResponse, err error) {
	if _, err := c.GetSELInfo(ctx); err != nil {
		return nil, fmt.Errorf("GetSELInfo failed, err: %s", err)
	}

	request := &GetSELEntryRequest{
		ReservationID: reservationID,
		RecordID:      recordID,
		Offset:        0,
		ReadBytes:     0xff,
	}
	response = &GetSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetSELEntries return all SEL records starting from the specified recordID.
// Pass 0 means retrieve all SEL entries starting from the first record.
func (c *Client) GetSELEntries(ctx context.Context, startRecordID uint16) ([]*SEL, error) {
	// Todo
	// Notice, this extra GetSELInfo call is used to make sure the GetSELEntry works properly.
	// On Huawei TaiShan 200 (Model 2280), the NextRecordID (0xffff) in GetSELEntryResponse is NOT right occasionally.
	// $ ipmitool -I lanplus -H x.x.x.x -U xxx -P xxx raw 0x0a 0x43 0x00 0x00 0x01 0x00 0x00 0xff -v
	// RAW REQ (channel=0x0 netfn=0xa lun=0x0 cmd=0x43 data_len=6)
	// RAW REQUEST (6 bytes)
	// 00 00 01 00 00 ff
	// RAW RSP (18 bytes)
	// ff ff 01 00 02 6d 8e 91 5f 20 00 04 10 79 6f 02
	// ff ff
	//
	// This extra GetSELInfo can avoid it. (I don't known why!)
	if _, err := c.GetSELInfo(ctx); err != nil {
		return nil, fmt.Errorf("GetSELInfo failed, err: %s", err)
	}

	var out = make([]*SEL, 0)
	var recordID uint16 = startRecordID
	for {
		selEntry, err := c.GetSELEntry(ctx, 0, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSELEntry failed, err: %s", err)
		}
		c.DebugBytes("sel entry record data", selEntry.Data, 16)

		sel, err := ParseSEL(selEntry.Data)
		if err != nil {
			return nil, fmt.Errorf("unpackSEL record failed, err: %s", err)
		}
		out = append(out, sel)

		recordID = selEntry.NextRecordID
		if recordID == 0xffff {
			break
		}
	}

	return out, nil
}
