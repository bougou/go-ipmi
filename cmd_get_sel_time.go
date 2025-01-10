package ipmi

import (
	"context"
	"fmt"
	"time"
)

// 31.10 Get SEL Time Command
type GetSELTimeRequest struct {
	// empty
}

type GetSELTimeResponse struct {
	// Present Timestamp clock reading
	Time time.Time
}

func (req *GetSELTimeRequest) Pack() []byte {
	return []byte{}
}

func (req *GetSELTimeRequest) Command() Command {
	return CommandGetSELTime
}

func (res *GetSELTimeResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	t, _, _ := unpackUint32L(msg, 0)
	res.Time = parseTimestamp(t)
	return nil
}

func (res *GetSELTimeResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSELTimeResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) GetSELTime(ctx context.Context) (response *GetSELTimeResponse, err error) {
	request := &GetSELTimeRequest{}
	response = &GetSELTimeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
