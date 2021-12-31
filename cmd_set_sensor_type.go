package ipmi

// 35.15 Set Sensor Type Command
type SetSensorTypeRequest struct {
	SensorNumber     uint8
	SensorType       SensorType
	EventReadingType EventReadingType
}

type SetSensorTypeResponse struct {
	// empty
}

func (req *SetSensorTypeRequest) Command() Command {
	return CommandSetSensorType
}

func (req *SetSensorTypeRequest) Pack() []byte {
	out := make([]byte, 3)
	packUint8(req.SensorNumber, out, 0)
	packUint8(uint8(req.SensorType), out, 1)
	packUint8(uint8(req.EventReadingType), out, 2)
	return out
}

func (res *SetSensorTypeResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetSensorTypeResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetSensorTypeResponse) Format() string {
	return ""
}

func (c *Client) SetSensorType(sensorNumber uint8, sensorType SensorType, eventReadingType EventReadingType) (response *SetSensorTypeResponse, err error) {
	request := &SetSensorTypeRequest{
		SensorNumber:     sensorNumber,
		SensorType:       sensorType,
		EventReadingType: eventReadingType,
	}
	response = &SetSensorTypeResponse{}
	err = c.Exchange(request, response)
	return
}
