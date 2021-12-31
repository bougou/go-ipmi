package ipmi

// 35.6 Set Sensor Hysteresis Command
type SetSensorHysteresisRequest struct {
	SensorNumber                     uint8
	PositiveGoingThresholdHysteresis uint8
	NegativeGoingThresholdHysteresis uint8
}

type SetSensorHysteresisResponse struct {
}

func (req *SetSensorHysteresisRequest) Command() Command {
	return CommandSetSensorHysteresis
}

func (req *SetSensorHysteresisRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.SensorNumber, out, 0)
	packUint8(0xff, out, 1)
	packUint8(req.PositiveGoingThresholdHysteresis, out, 2)
	packUint8(req.NegativeGoingThresholdHysteresis, out, 3)

	return out
}

func (res *SetSensorHysteresisResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetSensorHysteresisResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetSensorHysteresisResponse) Format() string {
	return ""
}

// This command provides a mechanism for setting the hysteresis values associated
// with the thresholds of a sensor that has threshold based event generation.
func (c *Client) SetSensorHysteresis(sensorNumber uint8, positiveHysteresis uint8, negativeHysteresis uint8) (response *SetSensorHysteresisResponse, err error) {
	request := &SetSensorHysteresisRequest{
		SensorNumber:                     sensorNumber,
		PositiveGoingThresholdHysteresis: positiveHysteresis,
		NegativeGoingThresholdHysteresis: negativeHysteresis,
	}
	response = &SetSensorHysteresisResponse{}
	err = c.Exchange(request, response)
	return
}
