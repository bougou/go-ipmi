package ipmi

import "fmt"

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

func (req *GetSensorThresholdsRequest) Command() Command {
	return CommandGetSensorThresholds
}

func (req *GetSensorThresholdsRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorThresholdsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShort
	}
	b, _, _ := unpackUint8(msg, 0)
	res.UNR_Readable = isBit5Set(b)
	res.UCR_Readable = isBit4Set(b)
	res.UNC_Readable = isBit3Set(b)
	res.LNR_Readable = isBit2Set(b)
	res.LCR_Readable = isBit1Set(b)
	res.LNC_Readable = isBit0Set(b)

	res.LNC_Raw, _, _ = unpackUint8(msg, 1)
	res.LCR_Raw, _, _ = unpackUint8(msg, 2)
	res.LNR_Raw, _, _ = unpackUint8(msg, 3)
	res.UNC_Raw, _, _ = unpackUint8(msg, 4)
	res.UCR_Raw, _, _ = unpackUint8(msg, 5)
	res.UNR_Raw, _, _ = unpackUint8(msg, 6)

	return nil
}

func (r *GetSensorThresholdsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorThresholdsResponse) Format() string {
	return fmt.Sprintf(`UNR Readable : %v%s
UCR Readable : %v%s
UNC Readable : %v%s
LNR Readable : %v%s
LCR Readable : %v%s
LNC Readable : %v%s`,
		res.UNR_Readable, formatBool(res.UNR_Readable, fmt.Sprintf(", raw: %#02x", res.UNR_Raw), ""),
		res.UCR_Readable, formatBool(res.UCR_Readable, fmt.Sprintf(", raw: %#02x", res.UCR_Raw), ""),
		res.UNC_Readable, formatBool(res.UNC_Readable, fmt.Sprintf(", raw: %#02x", res.UNC_Raw), ""),
		res.LNR_Readable, formatBool(res.LNR_Readable, fmt.Sprintf(", raw: %#02x", res.LNR_Raw), ""),
		res.LCR_Readable, formatBool(res.LCR_Readable, fmt.Sprintf(", raw: %#02x", res.LCR_Raw), ""),
		res.LNC_Readable, formatBool(res.LNC_Readable, fmt.Sprintf(", raw: %#02x", res.LNC_Raw), ""),
	)
}

// This command retrieves the threshold for the given sensor.
func (c *Client) GetSensorThresholds(sensorNumber uint8) (response *GetSensorThresholdsResponse, err error) {
	request := &GetSensorThresholdsRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorThresholdsResponse{}
	err = c.Exchange(request, response)
	return
}
