package ipmi

// 35.13 Get Sensor Event Status Command
type GetSensorEventStatusRequest struct {
	SensorNumber uint8
}

type GetSensorEventStatusResponse struct {
	AllEventMessagesDisabled bool
	AllSensorScaningDisabled bool
	ReadingStateUnavailable  bool

	AssertionEvent_UNC_GoHigh bool
	AssertionEvent_UNC_GoLow  bool
	AssertionEvent_LNR_GoHigh bool
	AssertionEvent_LNR_GoLow  bool
	AssertionEvent_LC_GoHigh  bool
	AssertionEvent_LC_GoLow   bool
	AssertionEvent_LNC_GoHigh bool
	AssertionEvent_LNC_GoLow  bool
	AssertionEventState7      bool
	AssertionEventState6      bool
	AssertionEventState5      bool
	AssertionEventState4      bool
	AssertionEventState3      bool
	AssertionEventState2      bool
	AssertionEventState1      bool
	AssertionEventState0      bool

	AssertionEvent_UNR_GoHigh bool
	AssertionEvent_UNR_GoLow  bool
	AssertionEvent_UC_GoHigh  bool
	AssertionEvent_UC_GoLow   bool
	AssertionEventState14     bool
	AssertionEventState13     bool
	AssertionEventState12     bool
	AssertionEventState11     bool
	AssertionEventState10     bool
	AssertionEventState9      bool
	AssertionEventState8      bool

	DeassertionEvent_UNC_GoHigh bool
	DeassertionEvent_UNC_GoLow  bool
	DeassertionEvent_LNR_GoHigh bool
	DeassertionEvent_LNR_GoLow  bool
	DeassertionEvent_LC_GoHigh  bool
	DeassertionEvent_LC_GoLow   bool
	DeassertionEvent_LNC_GoHigh bool
	DeassertionEvent_LNC_GoLow  bool
	DeassertionEventState7      bool
	DeassertionEventState6      bool
	DeassertionEventState5      bool
	DeassertionEventState4      bool
	DeassertionEventState3      bool
	DeassertionEventState2      bool
	DeassertionEventState1      bool
	DeassertionEventState0      bool

	DeassertionEvent_UNR_GoHigh bool
	DeassertionEvent_UNR_GoLow  bool
	DeassertionEvent_UC_GoHigh  bool
	DeassertionEvent_UC_GoLow   bool
	DeassertionEventState14     bool
	DeassertionEventState13     bool
	DeassertionEventState12     bool
	DeassertionEventState11     bool
	DeassertionEventState10     bool
	DeassertionEventState9      bool
	DeassertionEventState8      bool
}

func (req *GetSensorEventStatusRequest) Command() Command {
	return CommandGetSensorEventStatus
}

func (req *GetSensorEventStatusRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(0xff, out, 1)
	return out
}

func (res *GetSensorEventStatusResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	b1, _, _ := unpackUint8(msg, 0)
	res.AllEventMessagesDisabled = !isBit7Set(b1)
	res.AllSensorScaningDisabled = !isBit6Set(b1)
	res.ReadingStateUnavailable = isBit5Set(b1)

	if len(msg) >= 2 {
		b2, _, _ := unpackUint8(msg, 1)
		res.AssertionEvent_UNC_GoHigh = isBit7Set(b2)
		res.AssertionEvent_UNC_GoLow = isBit6Set(b2)
		res.AssertionEvent_LNR_GoHigh = isBit5Set(b2)
		res.AssertionEvent_LNR_GoLow = isBit4Set(b2)
		res.AssertionEvent_LC_GoHigh = isBit3Set(b2)
		res.AssertionEvent_LC_GoLow = isBit2Set(b2)
		res.AssertionEvent_LNC_GoHigh = isBit1Set(b2)
		res.AssertionEvent_LNC_GoLow = isBit0Set(b2)
		res.AssertionEventState7 = isBit7Set(b2)
		res.AssertionEventState6 = isBit6Set(b2)
		res.AssertionEventState5 = isBit5Set(b2)
		res.AssertionEventState4 = isBit4Set(b2)
		res.AssertionEventState3 = isBit3Set(b2)
		res.AssertionEventState2 = isBit2Set(b2)
		res.AssertionEventState1 = isBit1Set(b2)
		res.AssertionEventState0 = isBit0Set(b2)
	}

	if len(msg) >= 3 {
		b3, _, _ := unpackUint8(msg, 2)
		res.AssertionEvent_UNR_GoHigh = isBit3Set(b3)
		res.AssertionEvent_UNR_GoLow = isBit2Set(b3)
		res.AssertionEvent_UC_GoHigh = isBit1Set(b3)
		res.AssertionEvent_UC_GoLow = isBit0Set(b3)
		res.AssertionEventState14 = isBit6Set(b3)
		res.AssertionEventState13 = isBit5Set(b3)
		res.AssertionEventState12 = isBit4Set(b3)
		res.AssertionEventState11 = isBit3Set(b3)
		res.AssertionEventState10 = isBit2Set(b3)
		res.AssertionEventState9 = isBit1Set(b3)
		res.AssertionEventState8 = isBit0Set(b3)
	}

	if len(msg) >= 4 {
		b4, _, _ := unpackUint8(msg, 3)
		res.DeassertionEvent_UNC_GoHigh = isBit7Set(b4)
		res.DeassertionEvent_UNC_GoLow = isBit6Set(b4)
		res.DeassertionEvent_LNR_GoHigh = isBit5Set(b4)
		res.DeassertionEvent_LNR_GoLow = isBit4Set(b4)
		res.DeassertionEvent_LC_GoHigh = isBit3Set(b4)
		res.DeassertionEvent_LC_GoLow = isBit2Set(b4)
		res.DeassertionEvent_LNC_GoHigh = isBit1Set(b4)
		res.DeassertionEvent_LNC_GoLow = isBit0Set(b4)
		res.DeassertionEventState7 = isBit7Set(b4)
		res.DeassertionEventState6 = isBit6Set(b4)
		res.DeassertionEventState5 = isBit5Set(b4)
		res.DeassertionEventState4 = isBit4Set(b4)
		res.DeassertionEventState3 = isBit3Set(b4)
		res.DeassertionEventState2 = isBit2Set(b4)
		res.DeassertionEventState1 = isBit1Set(b4)
		res.DeassertionEventState0 = isBit0Set(b4)
	}

	if len(msg) >= 5 {
		b5, _, _ := unpackUint8(msg, 4)
		res.DeassertionEvent_UNR_GoHigh = isBit3Set(b5)
		res.DeassertionEvent_UNR_GoLow = isBit2Set(b5)
		res.DeassertionEvent_UC_GoHigh = isBit1Set(b5)
		res.DeassertionEvent_UC_GoLow = isBit0Set(b5)
		res.DeassertionEventState14 = isBit6Set(b5)
		res.DeassertionEventState13 = isBit5Set(b5)
		res.DeassertionEventState12 = isBit4Set(b5)
		res.DeassertionEventState11 = isBit3Set(b5)
		res.DeassertionEventState10 = isBit2Set(b5)
		res.DeassertionEventState9 = isBit1Set(b5)
		res.DeassertionEventState8 = isBit0Set(b5)
	}

	return nil
}

func (r *GetSensorEventStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorEventStatusResponse) Format() string {
	return ""
}

func (c *Client) GetSensorEventStatus(sensorNumber uint8, positiveHysteresis uint8, negativeHysteresis uint8) (response *GetSensorEventStatusResponse, err error) {
	request := &GetSensorEventStatusRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorEventStatusResponse{}
	err = c.Exchange(request, response)
	return
}
