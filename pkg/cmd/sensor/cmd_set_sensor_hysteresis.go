package sensor

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 35.6 Set Sensor Hysteresis Command
)

type SetSensorHysteresisRequest struct {
	SensorNumber       uint8
	PositiveHysteresis uint8
	NegativeHysteresis uint8
}

type SetSensorHysteresisResponse struct {
}

func (req *SetSensorHysteresisRequest) Command() ipmi.Command {
	return ipmi.CommandSetSensorHysteresis
}

func (req *SetSensorHysteresisRequest) Pack() []byte {
	out := make([]byte, 4)
	ipmi.PackUint8(req.SensorNumber, out, 0)
	ipmi.PackUint8(0xff, out, 1) // reserved for future "hysteresis mask" definition. Write as FFh
	ipmi.PackUint8(req.PositiveHysteresis, out, 2)
	ipmi.PackUint8(req.NegativeHysteresis, out, 3)

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
