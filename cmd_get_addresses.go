package ipmi

import (
	"context"
)

type GetAddressesRequest struct {
}

type GetAddressesResponse struct {
	BridgeAddresses []byte
}

func (req *GetAddressesRequest) Command() Command {
	return CommandGetAddresses
}

func (req *GetAddressesRequest) Pack() []byte {
	return []byte{}
}

func (res *GetAddressesResponse) Unpack(msg []byte) error {
	if len(msg) > 0 {
		res.BridgeAddresses, _, _ = unpackBytes(msg, 0, len(msg))
	}
	return nil
}

func (res *GetAddressesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetAddressesResponse) Format() string {
	return ""
}

func (c *Client) GetAddresses(ctx context.Context) (response *GetAddressesResponse, err error) {
	request := &GetAddressesRequest{}
	response = &GetAddressesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
