package ipmi

// 35.7 Get Sensor Hysteresis Command
type GetSensorHysteresisRequest struct {
	SensorNumber uint8
}

type GetSensorHysteresisResponse struct {
	PositiveGoingThresholdHysteresis uint8
	NegativeGoingThresholdHysteresis uint8
}

func (req *GetSensorHysteresisRequest) Command() Command {
	return CommandGetSensorHysteresis
}

func (req *GetSensorHysteresisRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(0xff, out, 1)
	return out
}

func (res *GetSensorHysteresisResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	res.PositiveGoingThresholdHysteresis, _, _ = unpackUint8(msg, 0)
	res.NegativeGoingThresholdHysteresis, _, _ = unpackUint8(msg, 1)
	return nil
}

func (r *GetSensorHysteresisResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorHysteresisResponse) Format() string {
	return ""
}

// This command retrieves the present hysteresis values for the specified sensor.
// If the sensor hysteresis values are "fixed", then the hysteresis values can be obtained from the SDR for the sensor.
func (c *Client) GetSensorHysteresis(sensorNumber uint8, positiveHysteresis uint8, negativeHysteresis uint8) (response *GetSensorHysteresisResponse, err error) {
	request := &GetSensorHysteresisRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorHysteresisResponse{}
	err = c.Exchange(request, response)
	return
}
