package ipmi

import (
	"context"
)

type DeviceBridgeRequestRequest struct {
	// Data to be bridged to selected device. Note: this command is only accepted as an ICMB broadcast.
	Data []byte
}

type DeviceBridgeRequestResponse struct {
}

func (req *DeviceBridgeRequestRequest) Command() Command {
	return CommandDeviceBridgeRequest
}

func (req *DeviceBridgeRequestRequest) Pack() []byte {
	out := make([]byte, len(req.Data))
	packBytes(req.Data, out, 0)
	return out
}

func (res *DeviceBridgeRequestResponse) Unpack(msg []byte) error {
	return nil
}

func (res *DeviceBridgeRequestResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *DeviceBridgeRequestResponse) Format() string {
	return ""
}

func (c *Client) DeviceBridgeRequest(ctx context.Context, data []byte) (response *DeviceBridgeRequestResponse, err error) {
	request := &DeviceBridgeRequestRequest{
		Data: data,
	}
	response = &DeviceBridgeRequestResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
