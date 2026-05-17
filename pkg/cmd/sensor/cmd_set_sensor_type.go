package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 35.15 Set Sensor Type Command
)

type SetSensorTypeRequest struct {
	SensorNumber     uint8
	SensorType       ipmi.SensorType
	EventReadingType ipmi.EventReadingType
}

type SetSensorTypeResponse struct {
	// empty
}

func (req *SetSensorTypeRequest) Command() ipmi.Command {
	return ipmi.CommandSetSensorType
}

func (req *SetSensorTypeRequest) Pack() []byte {
	out := make([]byte, 3)
	ipmi.PackUint8(req.SensorNumber, out, 0)
	ipmi.PackUint8(uint8(req.SensorType), out, 1)
	ipmi.PackUint8(uint8(req.EventReadingType), out, 2)
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
