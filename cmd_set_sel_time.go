package ipmi

import (
	"context"
	"fmt"
	"time"
)

// 31.11 Set SEL Time Command
type SetSELTimeRequest struct {
	Time time.Time
}

type SetSELTimeResponse struct {
}

func (req *SetSELTimeRequest) Pack() []byte {
	var out = make([]byte, 4)
	packUint32L(uint32(req.Time.Unix()), out, 0)
	return out
}

func (req *SetSELTimeRequest) Command() Command {
	return CommandSetSELTime
}

func (res *SetSELTimeResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSELTimeResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *SetSELTimeResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

func (c *Client) SetSELTime(ctx context.Context, t time.Time) (response *SetSELTimeResponse, err error) {
	request := &SetSELTimeRequest{
		Time: t,
	}
	response = &SetSELTimeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
