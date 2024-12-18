package ipmi

import (
	"context"
	"fmt"
)

// GetSensorTypeRequest (31.2)
type GetSensorTypeRequest struct {
	SensorNumber uint8
}

type GetSensorTypeResponse struct {
	SensorType       SensorType
	EventReadingType EventReadingType
}

func (req *GetSensorTypeRequest) Command() Command {
	return CommandGetSensorType
}

func (req *GetSensorTypeRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorTypeResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	b1, _, _ := unpackUint8(msg, 0)
	res.SensorType = SensorType(b1)
	b2, _, _ := unpackUint8(msg, 1)
	res.EventReadingType = EventReadingType(b2)
	return nil
}

func (r *GetSensorTypeResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorTypeResponse) Format() string {
	return fmt.Sprintf(`Sensor Type                 : %s
Event/Reading Type          : %#02x (%s)`,
		res.SensorType,
		uint8(res.EventReadingType), res.EventReadingType,
	)
}

func (c *Client) GetSensorType(ctx context.Context, sensorNumber uint8) (response *GetSensorTypeResponse, err error) {
	request := &GetSensorTypeRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorTypeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
