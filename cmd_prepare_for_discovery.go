package ipmi

import (
	"context"
)

type PrepareForDiscoveryRequest struct {
}

type PrepareForDiscoveryResponse struct {
}

func (req *PrepareForDiscoveryRequest) Command() Command {
	return CommandPrepareForDiscovery
}

func (req *PrepareForDiscoveryRequest) Pack() []byte {
	return []byte{}
}

func (res *PrepareForDiscoveryResponse) Unpack(msg []byte) error {
	return nil
}

func (res *PrepareForDiscoveryResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *PrepareForDiscoveryResponse) Format() string {
	return ""
}

func (c *Client) PrepareForDiscovery(ctx context.Context) (response *PrepareForDiscoveryResponse, err error) {
	request := &PrepareForDiscoveryRequest{}
	response = &PrepareForDiscoveryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
