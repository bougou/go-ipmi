package ipmi

import (
	"context"
)

type SetChassisDeviceIDRequest struct {
	DeviceID uint8
}

type SetChassisDeviceIDResponse struct {
}

func (req *SetChassisDeviceIDRequest) Command() Command {
	return CommandSetChassisDeviceID
}

func (req *SetChassisDeviceIDRequest) Pack() []byte {
	return []byte{req.DeviceID}
}

func (res *SetChassisDeviceIDResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetChassisDeviceIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetChassisDeviceIDResponse) Format() string {
	return ""
}

func (c *Client) SetChassisDeviceID(ctx context.Context, deviceID uint8) (response *SetChassisDeviceIDResponse, err error) {
	request := &SetChassisDeviceIDRequest{
		DeviceID: deviceID,
	}
	response = &SetChassisDeviceIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
