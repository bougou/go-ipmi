package ipmi

import (
	"context"
	"fmt"
	"time"
)

// 30.6 Get Last Processed Event ID Command
type GetLastProcessedEventIdRequest struct {
	// empty
}

type GetLastProcessedEventIdResponse struct {
	MostRecentAdditionTime             time.Time
	LastRecordID                       uint16 // Record ID for last record in SEL. Returns FFFFh if SEL is empty.
	LastSoftwareProcessedEventRecordID uint16
	LastBMCProcessedEventRecordID      uint16 // Returns 0000h when event has been processed but could not be logged because the SEL is full or logging has been disabled.
}

func (req *GetLastProcessedEventIdRequest) Command() Command {
	return CommandGetLastProcessedEventId
}

func (req *GetLastProcessedEventIdRequest) Pack() []byte {
	return []byte{}
}

func (res *GetLastProcessedEventIdResponse) Unpack(msg []byte) error {
	if len(msg) < 10 {
		return ErrUnpackedDataTooShort
	}

	ts, _, _ := unpackUint32L(msg, 0)
	res.MostRecentAdditionTime = parseTimestamp(ts)
	res.LastRecordID, _, _ = unpackUint16L(msg, 4)
	res.LastSoftwareProcessedEventRecordID, _, _ = unpackUint16(msg, 6)
	res.LastBMCProcessedEventRecordID, _, _ = unpackUint16(msg, 8)
	return nil
}

func (r *GetLastProcessedEventIdResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *GetLastProcessedEventIdResponse) Format() string {
	return fmt.Sprintf(`
MostRecentAdditionTime             : %s
LastRecordID                       : %#04x (%d)
LastSoftwareProcessedEventRecordID : %#04x (%d)
LastBMCProcessedEventRecordID      : %#04x (%d)
`,
		res.MostRecentAdditionTime.String(),
		res.LastRecordID, res.LastRecordID,
		res.LastSoftwareProcessedEventRecordID, res.LastSoftwareProcessedEventRecordID,
		res.LastBMCProcessedEventRecordID, res.LastBMCProcessedEventRecordID,
	)
}

func (c *Client) GetLastProcessedEventId(ctx context.Context) (response *GetLastProcessedEventIdResponse, err error) {
	request := &GetLastProcessedEventIdRequest{}
	response = &GetLastProcessedEventIdResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
