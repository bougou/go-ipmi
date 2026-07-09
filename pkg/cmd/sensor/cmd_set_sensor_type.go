package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 35.15 Set Sensor Type Command
)

type SetSensorTypeRequest struct {
	SensorNumber     uint8
	SensorType       types.SensorType
	EventReadingType types.EventReadingType
}

type SetSensorTypeResponse struct {
	// empty
}

func (req *SetSensorTypeRequest) Command() types.Command {
	return types.CommandSetSensorType
}

func (req *SetSensorTypeRequest) Pack() []byte {
	out := make([]byte, 3)
	types.PackUint8(req.SensorNumber, out, 0)
	types.PackUint8(uint8(req.SensorType), out, 1)
	types.PackUint8(uint8(req.EventReadingType), out, 2)
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
