package ipmi

import "context"

// 35.17 Set Sensor Reading And Event Status Command
type SetSensorReadingAndEventStatusRequest struct {
	SensorNumber uint8

	EventDataBytesOperation  uint8
	AssertionBitsOperation   uint8
	DeassertionBitsOperation uint8
	SensorReadingOperation   uint8

	SensorReading uint8

	SensorEventFlag

	EventData1 uint8
	EventData2 uint8
	EventData3 uint8
}

type SetSensorReadingAndEventStatusResponse struct {
	// empty
}

func (req *SetSensorReadingAndEventStatusRequest) Command() Command {
	return CommandSetSensorReadingAndEventStatus
}

func (req *SetSensorReadingAndEventStatusRequest) Pack() []byte {
	out := make([]byte, 9)
	packUint8(req.SensorNumber, out, 0)

	var operation uint8
	operation |= uint8(req.EventDataBytesOperation) << 6
	operation |= (uint8(req.AssertionBitsOperation) & 0x3f) << 4
	operation |= (uint8(req.DeassertionBitsOperation) & 0x0f) << 2
	operation |= uint8(req.SensorReadingOperation) & 0x03
	packUint8(operation, out, 1)

	packUint8(req.SensorReading, out, 2)

	// Todo determine sensor is threshold based or discrete

	return out
}

func (res *SetSensorReadingAndEventStatusResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetSensorReadingAndEventStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Attempt to change reading or set or clear status bits that are not settable via this command",
		0x81: "Attempted to set Event Data Bytes, but setting Event Data Bytes is not supported for this sensor.",
	}
}

func (res *SetSensorReadingAndEventStatusResponse) Format() string {
	return ""
}

func (c *Client) SetSensorReadingAndEventStatus(ctx context.Context, request *SetSensorReadingAndEventStatusRequest) (response *SetSensorReadingAndEventStatusResponse, err error) {
	response = &SetSensorReadingAndEventStatusResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
