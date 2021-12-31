package ipmi

// 35.8 Set Sensor Thresholds Command
type GetSensorThresholdsRequest struct {
	SensorNumber uint8
}

type GetSensorThresholdsResponse struct {
	// Readable thresholds flag
	ReadableUNR bool
	ReadableUC  bool
	ReadableUNC bool
	ReadableLNR bool
	ReadableLC  bool
	ReadableLNC bool

	// Threshold value
	LNC uint8
	LC  uint8
	LNR uint8
	UNC uint8
	UC  uint8
	UNR uint8
}

func (req *GetSensorThresholdsRequest) Command() Command {
	return CommandGetSensorThresholds
}

func (req *GetSensorThresholdsRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(0xff, out, 1)
	return out
}

func (res *GetSensorThresholdsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShort
	}
	b, _, _ := unpackUint8(msg, 0)
	res.ReadableUNR = isBit5Set(b)
	res.ReadableUC = isBit5Set(b)
	res.ReadableUNC = isBit5Set(b)
	res.ReadableLNR = isBit5Set(b)
	res.ReadableLC = isBit5Set(b)
	res.ReadableLNC = isBit5Set(b)

	res.LNC, _, _ = unpackUint8(msg, 1)
	res.LC, _, _ = unpackUint8(msg, 2)
	res.LNR, _, _ = unpackUint8(msg, 3)
	res.UNC, _, _ = unpackUint8(msg, 4)
	res.UC, _, _ = unpackUint8(msg, 5)
	res.UNR, _, _ = unpackUint8(msg, 6)

	return nil
}

func (r *GetSensorThresholdsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorThresholdsResponse) Format() string {
	return ""
}

// This command retrieves the threshold for the given sensor.
func (c *Client) GetSensorThresholds(request *GetSensorThresholdsRequest) (response *GetSensorThresholdsResponse, err error) {
	response = &GetSensorThresholdsResponse{}
	err = c.Exchange(request, response)
	return
}
