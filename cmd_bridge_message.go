package ipmi

import (
	"context"
)

type BridgeMessageRequest struct {
	// Data to be bridged to next bus.
	Data []byte
}

type BridgeMessageResponse struct {
}

func (req *BridgeMessageRequest) Command() Command {
	return CommandBridgeMessage
}

func (req *BridgeMessageRequest) Pack() []byte {
	out := make([]byte, len(req.Data))
	packBytes(req.Data, out, 0)
	return out
}

func (res *BridgeMessageResponse) Unpack(msg []byte) error {
	return nil
}

func (res *BridgeMessageResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *BridgeMessageResponse) Format() string {
	return ""
}

func (c *Client) BridgeMessage(ctx context.Context, data []byte) (response *BridgeMessageResponse, err error) {
	request := &BridgeMessageRequest{
		Data: data,
	}
	response = &BridgeMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
