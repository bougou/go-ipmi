package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 35.7 Get Sensor Hysteresis Command
type GetSensorHysteresisRequest struct {
	SensorNumber uint8
}

type GetSensorHysteresisResponse struct {
	PositiveRaw uint8
	NegativeRaw uint8
}

func (req *GetSensorHysteresisRequest) Command() types.Command {
	return types.CommandGetSensorHysteresis
}

func (req *GetSensorHysteresisRequest) Pack() []byte {
	out := make([]byte, 2)
	types.PackUint8(req.SensorNumber, out, 0)
	types.PackUint8(0xff, out, 1) // reserved for future "hysteresis mask" definition. Write as "FFh"
	return out
}

func (res *GetSensorHysteresisResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.PositiveRaw, _, _ = types.UnpackUint8(msg, 0)
	res.NegativeRaw, _, _ = types.UnpackUint8(msg, 1)
	return nil
}

func (r *GetSensorHysteresisResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorHysteresisResponse) Format() string {
	return "" +
		fmt.Sprintf("Positive Hysteresis : %d\n", res.PositiveRaw) +
		fmt.Sprintf("Negative Hysteresis : %d\n", res.NegativeRaw)
}

// This command retrieves the present hysteresis values for the specified sensor.
// If the sensor hysteresis values are "fixed", then the hysteresis values can be obtained from the SDR for the sensor.
