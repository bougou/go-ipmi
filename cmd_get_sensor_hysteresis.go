package ipmi

import (
	"context"
	"fmt"
)

// 35.7 Get Sensor Hysteresis Command
type GetSensorHysteresisRequest struct {
	SensorNumber uint8
}

type GetSensorHysteresisResponse struct {
	PositiveRaw uint8
	NegativeRaw uint8
}

func (req *GetSensorHysteresisRequest) Command() Command {
	return CommandGetSensorHysteresis
}

func (req *GetSensorHysteresisRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(0xff, out, 1) // reserved for future "hysteresis mask" definition. Write as "FFh"
	return out
}

func (res *GetSensorHysteresisResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.PositiveRaw, _, _ = unpackUint8(msg, 0)
	res.NegativeRaw, _, _ = unpackUint8(msg, 1)
	return nil
}

func (r *GetSensorHysteresisResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorHysteresisResponse) Format() string {
	return fmt.Sprintf(`Positive Hysteresis : %d
Negative Hysteresis : %d`,
		res.PositiveRaw,
		res.NegativeRaw,
	)
}

// This command retrieves the present hysteresis values for the specified sensor.
// If the sensor hysteresis values are "fixed", then the hysteresis values can be obtained from the SDR for the sensor.
func (c *Client) GetSensorHysteresis(ctx context.Context, sensorNumber uint8) (response *GetSensorHysteresisResponse, err error) {
	request := &GetSensorHysteresisRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorHysteresisResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
