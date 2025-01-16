package ipmi

import (
	"context"
)

type GetEventCountRequest struct {
}

type GetEventCountResponse struct {
	EventCount uint8
}

func (req *GetEventCountRequest) Command() Command {
	return CommandGetEventCount
}

func (req *GetEventCountRequest) Pack() []byte {
	return []byte{}
}

func (res *GetEventCountResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.EventCount = msg[0]
	return nil
}

func (res *GetEventCountResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetEventCountResponse) Format() string {
	return ""
}

func (c *Client) GetEventCount(ctx context.Context, data []byte) (response *GetEventCountResponse, err error) {
	request := &GetEventCountRequest{}
	response = &GetEventCountResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
