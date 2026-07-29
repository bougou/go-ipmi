package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// GetSensorTypeRequest (see [IPMI specification v2.0], section 35.16)
type GetSensorTypeRequest struct {
	SensorNumber uint8
}

type GetSensorTypeResponse struct {
	SensorType       types.SensorType
	EventReadingType types.EventReadingType
}

func (req *GetSensorTypeRequest) Command() types.Command {
	return types.CommandGetSensorType
}

func (req *GetSensorTypeRequest) Pack() []byte {
	out := make([]byte, 1)
	types.PackUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorTypeResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	b1, _, _ := types.UnpackUint8(msg, 0)
	res.SensorType = types.SensorType(b1)
	b2, _, _ := types.UnpackUint8(msg, 1)
	res.EventReadingType = types.EventReadingType(b2)
	return nil
}

func (r *GetSensorTypeResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorTypeResponse) Format() string {
	return "" +
		fmt.Sprintf("Sensor Type        : %s\n", res.SensorType) +
		fmt.Sprintf("Event/Reading Type : %#02x (%s)\n", uint8(res.EventReadingType), res.EventReadingType.String())
}
