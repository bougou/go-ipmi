package ipmi

import (
	"context"
	"fmt"
)

type BridgeRequestRequest struct {
	// Data to be bridged to next bus.
	Data []byte
}

type BridgeRequestResponse struct {
	// Response from bridge request.
	Data []byte
}

func (req *BridgeRequestRequest) Command() Command {
	return CommandBridgeRequest
}

func (req *BridgeRequestRequest) Pack() []byte {
	out := make([]byte, len(req.Data))
	packBytes(req.Data, out, 0)
	return out
}

func (res *BridgeRequestResponse) Unpack(msg []byte) error {
	res.Data, _, _ = unpackBytes(msg, 0, len(msg))
	return nil
}

func (res *BridgeRequestResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *BridgeRequestResponse) Format() string {
	return fmt.Sprintf("%#2x", res.Data)
}

func (c *Client) BridgeRequest(ctx context.Context, data []byte) (response *BridgeRequestResponse, err error) {
	request := &BridgeRequestRequest{
		Data: data,
	}
	response = &BridgeRequestResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
