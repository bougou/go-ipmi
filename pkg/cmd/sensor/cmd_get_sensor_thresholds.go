package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 35.9 Get Sensor Thresholds Command
type GetSensorThresholdsRequest struct {
	SensorNumber uint8
}

type GetSensorThresholdsResponse struct {
	// Readable thresholds mask
	UNR_Readable bool
	UCR_Readable bool
	UNC_Readable bool
	LNR_Readable bool
	LCR_Readable bool
	LNC_Readable bool

	// Threshold value
	LNC_Raw uint8
	LCR_Raw uint8
	LNR_Raw uint8
	UNC_Raw uint8
	UCR_Raw uint8
	UNR_Raw uint8
}

func (req *GetSensorThresholdsRequest) Command() types.Command {
	return types.CommandGetSensorThresholds
}

func (req *GetSensorThresholdsRequest) Pack() []byte {
	out := make([]byte, 1)
	types.PackUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorThresholdsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 7)
	}
	b, _, _ := types.UnpackUint8(msg, 0)
	res.UNR_Readable = types.IsBit5Set(b)
	res.UCR_Readable = types.IsBit4Set(b)
	res.UNC_Readable = types.IsBit3Set(b)
	res.LNR_Readable = types.IsBit2Set(b)
	res.LCR_Readable = types.IsBit1Set(b)
	res.LNC_Readable = types.IsBit0Set(b)

	res.LNC_Raw, _, _ = types.UnpackUint8(msg, 1)
	res.LCR_Raw, _, _ = types.UnpackUint8(msg, 2)
	res.LNR_Raw, _, _ = types.UnpackUint8(msg, 3)
	res.UNC_Raw, _, _ = types.UnpackUint8(msg, 4)
	res.UCR_Raw, _, _ = types.UnpackUint8(msg, 5)
	res.UNR_Raw, _, _ = types.UnpackUint8(msg, 6)

	return nil
}

func (r *GetSensorThresholdsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorThresholdsResponse) Format() string {
	return "" +
		fmt.Sprintf("UNR Readable : %v%s\n", res.UNR_Readable, types.FormatBool(res.UNR_Readable, fmt.Sprintf(", raw: %#02x", res.UNR_Raw), "")) +
		fmt.Sprintf("UCR Readable : %v%s\n", res.UCR_Readable, types.FormatBool(res.UCR_Readable, fmt.Sprintf(", raw: %#02x", res.UCR_Raw), "")) +
		fmt.Sprintf("UNC Readable : %v%s\n", res.UNC_Readable, types.FormatBool(res.UNC_Readable, fmt.Sprintf(", raw: %#02x", res.UNC_Raw), "")) +
		fmt.Sprintf("LNR Readable : %v%s\n", res.LNR_Readable, types.FormatBool(res.LNR_Readable, fmt.Sprintf(", raw: %#02x", res.LNR_Raw), "")) +
		fmt.Sprintf("LCR Readable : %v%s\n", res.LCR_Readable, types.FormatBool(res.LCR_Readable, fmt.Sprintf(", raw: %#02x", res.LCR_Raw), "")) +
		fmt.Sprintf("LNC Readable : %v%s\n", res.LNC_Readable, types.FormatBool(res.LNC_Readable, fmt.Sprintf(", raw: %#02x", res.LNC_Raw), ""))
}

// This command retrieves the threshold for the given sensor.
