package ipmi

import (
	"context"
)

type SetDiscoveredRequest struct {
}

type SetDiscoveredResponse struct {
}

func (req *SetDiscoveredRequest) Command() Command {
	return CommandSetDiscovered
}

func (req *SetDiscoveredRequest) Pack() []byte {
	return []byte{}
}

func (res *SetDiscoveredResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetDiscoveredResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDiscoveredResponse) Format() string {
	return ""
}

func (c *Client) SetDiscovered(ctx context.Context) (response *SetDiscoveredResponse, err error) {
	request := &SetDiscoveredRequest{}
	response = &SetDiscoveredResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
