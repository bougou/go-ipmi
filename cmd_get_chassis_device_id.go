package ipmi

import (
	"context"
)

type GetChassisDeviceIDRequest struct {
}

type GetChassisDeviceIDResponse struct {
	DeviceID uint8
}

func (req *GetChassisDeviceIDRequest) Command() Command {
	return CommandGetChassisDeviceID
}

func (req *GetChassisDeviceIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetChassisDeviceIDResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.DeviceID = msg[0]
	return nil
}

func (res *GetChassisDeviceIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetChassisDeviceIDResponse) Format() string {
	return ""
}

func (c *Client) GetChassisDeviceID(ctx context.Context) (response *GetChassisDeviceIDResponse, err error) {
	request := &GetChassisDeviceIDRequest{}
	response = &GetChassisDeviceIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
