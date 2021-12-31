package ipmi

// 35.8 Set Sensor Thresholds Command
type SetSensorThresholdsRequest struct {
	SensorNumber uint8

	// Set Threshold flag
	SetUNR bool
	SetUC  bool
	SetUNC bool
	SetLNR bool
	SetLC  bool
	SetLNC bool

	// Threshold value
	LNC uint8
	LC  uint8
	LNR uint8
	UNC uint8
	UC  uint8
	UNR uint8
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
	if req.SetUC {
		b = setBit4(b)
	}
	if req.SetUNC {
		b = setBit3(b)
	}
	if req.SetLNR {
		b = setBit2(b)
	}
	if req.SetLC {
		b = setBit1(b)
	}
	if req.SetLNC {
		b = setBit0(b)
	}
	packUint8(b, out, 1)

	packUint8(req.LNC, out, 2)
	packUint8(req.LC, out, 3)
	packUint8(req.LNR, out, 4)
	packUint8(req.UNC, out, 5)
	packUint8(req.UC, out, 6)
	packUint8(req.UNR, out, 7)
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
