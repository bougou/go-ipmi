package ipmi

// 35.8 Set Sensor Thresholds Command
type SetSensorThresholdsRequest struct {
	SensorNumber uint8

	// Set Threshold flag
	SetUNR bool
	SetUCR bool
	SetUNC bool
	SetLNR bool
	SetLCR bool
	SetLNC bool

	// Threshold value
	LNC_Raw uint8
	LCR_Raw uint8
	LNR_Raw uint8
	UNC_Raw uint8
	UCR_Raw uint8
	UNR_Raw uint8
}

type SetSensorThresholdsResponse struct {
	// empty
}

func (req *SetSensorThresholdsRequest) Command() Command {
	return CommandSetSensorThresholds
}

func (req *SetSensorThresholdsRequest) Pack() []byte {
	out := make([]byte, 8)
	packUint8(req.SensorNumber, out, 0)

	var b uint8
	if req.SetUNR {
		b = setBit5(b)
	}
	if req.SetUCR {
		b = setBit4(b)
	}
	if req.SetUNC {
		b = setBit3(b)
	}
	if req.SetLNR {
		b = setBit2(b)
	}
	if req.SetLCR {
		b = setBit1(b)
	}
	if req.SetLNC {
		b = setBit0(b)
	}
	packUint8(b, out, 1)

	packUint8(req.LNC_Raw, out, 2)
	packUint8(req.LCR_Raw, out, 3)
	packUint8(req.LNR_Raw, out, 4)
	packUint8(req.UNC_Raw, out, 5)
	packUint8(req.UCR_Raw, out, 6)
	packUint8(req.UNR_Raw, out, 7)
	return out
}

func (res *SetSensorThresholdsResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetSensorThresholdsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetSensorThresholdsResponse) Format() string {
	return ""
}

// This command provides a mechanism for setting the hysteresis values associated
// with the thresholds of a sensor that has threshold based event generation.
func (c *Client) SetSensorThresholds(request *SetSensorThresholdsRequest) (response *SetSensorThresholdsResponse, err error) {
	response = &SetSensorThresholdsResponse{}
	err = c.Exchange(request, response)
	return
}
