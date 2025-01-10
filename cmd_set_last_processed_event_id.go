package ipmi

import "context"

// 30.5 Set Last Processed Event ID Command
type SetLastProcessedEventIdRequest struct {
	// 0b = set Record ID for last record processed by software.
	// 1b = set Record ID for last record processed by BMC.
	ByBMC    bool
	RecordID uint16
}

type SetLastProcessedEventIdResponse struct {
	// empty
}

func (req *SetLastProcessedEventIdRequest) Command() Command {
	return CommandSetLastProcessedEventId
}

func (req *SetLastProcessedEventIdRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	var b0 uint8 = 0x0
	if req.ByBMC {
		b0 = 1
	}
	packUint8(b0, out, 0)
	packUint16L(req.RecordID, out, 1)

	return out
}

func (res *SetLastProcessedEventIdResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetLastProcessedEventIdResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x81: "cannot execute command, SEL erase in progress",
	}
}

func (res *SetLastProcessedEventIdResponse) Format() string {
	return ""
}

func (c *Client) SetLastProcessedEventId(ctx context.Context, recordID uint16, byBMC bool) (response *SetLastProcessedEventIdResponse, err error) {
	request := &SetLastProcessedEventIdRequest{
		ByBMC:    byBMC,
		RecordID: recordID,
	}
	response = &SetLastProcessedEventIdResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
