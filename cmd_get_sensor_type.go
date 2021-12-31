package ipmi

import (
	"fmt"
)

// GetSensorTypeRequest (31.2) command returns the number of entries in the SEL.
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
		return ErrUnpackedDataTooShort
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
Event/Reading Type          : %#02x,
Event/Reading Type Category : %s`,
		res.SensorType,
		res.EventReadingType,
		res.EventReadingType.Category(),
	)
}

func (c *Client) GetSensorType() (response *GetSensorTypeResponse, err error) {
	request := &GetSensorTypeRequest{}
	response = &GetSensorTypeResponse{}
	err = c.Exchange(request, response)
	return
}
