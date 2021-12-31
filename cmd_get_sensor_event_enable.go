package ipmi

// 35.11 Get Sensor Event Enable Command
type GetSensorEventEnableRequest struct {
	SensorNumber uint8
}

type GetSensorEventEnableResponse struct {
	AllEventMessagesDisabled bool
	AllSensorScaningDisabled bool

	AssertionEvent_UNC_GoHighEnabled bool
	AssertionEvent_UNC_GoLowEnabled  bool
	AssertionEvent_LNR_GoHighEnabled bool
	AssertionEvent_LNR_GoLowEnabled  bool
	AssertionEvent_LC_GoHighEnabled  bool
	AssertionEvent_LC_GoLowEnabled   bool
	AssertionEvent_LNC_GoHighEnabled bool
	AssertionEvent_LNC_GoLowEnabled  bool
	AssertionEventState7Enabled      bool
	AssertionEventState6Enabled      bool
	AssertionEventState5Enabled      bool
	AssertionEventState4Enabled      bool
	AssertionEventState3Enabled      bool
	AssertionEventState2Enabled      bool
	AssertionEventState1Enabled      bool
	AssertionEventState0Enabled      bool

	AssertionEvent_UNR_GoHighEnabled bool
	AssertionEvent_UNR_GoLowEnabled  bool
	AssertionEvent_UC_GoHighEnabled  bool
	AssertionEvent_UC_GoLowEnabled   bool
	AssertionEventState14Enabled     bool
	AssertionEventState13Enabled     bool
	AssertionEventState12Enabled     bool
	AssertionEventState11Enabled     bool
	AssertionEventState10Enabled     bool
	AssertionEventState9Enabled      bool
	AssertionEventState8Enabled      bool

	DeassertionEvent_UNC_GoHighEnabled bool
	DeassertionEvent_UNC_GoLowEnabled  bool
	DeassertionEvent_LNR_GoHighEnabled bool
	DeassertionEvent_LNR_GoLowEnabled  bool
	DeassertionEvent_LC_GoHighEnabled  bool
	DeassertionEvent_LC_GoLowEnabled   bool
	DeassertionEvent_LNC_GoHighEnabled bool
	DeassertionEvent_LNC_GoLowEnabled  bool
	DeassertionEventState7Enabled      bool
	DeassertionEventState6Enabled      bool
	DeassertionEventState5Enabled      bool
	DeassertionEventState4Enabled      bool
	DeassertionEventState3Enabled      bool
	DeassertionEventState2Enabled      bool
	DeassertionEventState1Enabled      bool
	DeassertionEventState0Enabled      bool

	DeassertionEvent_UNR_GoHighEnabled bool
	DeassertionEvent_UNR_GoLowEnabled  bool
	DeassertionEvent_UC_GoHighEnabled  bool
	DeassertionEvent_UC_GoLowEnabled   bool
	DeassertionEventState14Enabled     bool
	DeassertionEventState13Enabled     bool
	DeassertionEventState12Enabled     bool
	DeassertionEventState11Enabled     bool
	DeassertionEventState10Enabled     bool
	DeassertionEventState9Enabled      bool
	DeassertionEventState8Enabled      bool
}

func (req *GetSensorEventEnableRequest) Command() Command {
	return CommandGetSensorEventEnable
}

func (req *GetSensorEventEnableRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorEventEnableResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}
	b1, _, _ := unpackUint8(msg, 0)
	res.AllEventMessagesDisabled = !isBit7Set(b1)
	res.AllSensorScaningDisabled = !isBit6Set(b1)

	if len(msg) >= 2 {
		b2, _, _ := unpackUint8(msg, 1)
		res.AssertionEvent_UNC_GoHighEnabled = isBit7Set(b2)
		res.AssertionEvent_UNC_GoLowEnabled = isBit6Set(b2)
		res.AssertionEvent_LNR_GoHighEnabled = isBit5Set(b2)
		res.AssertionEvent_LNR_GoLowEnabled = isBit4Set(b2)
		res.AssertionEvent_LC_GoHighEnabled = isBit3Set(b2)
		res.AssertionEvent_LC_GoLowEnabled = isBit2Set(b2)
		res.AssertionEvent_LNC_GoHighEnabled = isBit1Set(b2)
		res.AssertionEvent_LNC_GoLowEnabled = isBit0Set(b2)
		res.AssertionEventState7Enabled = isBit7Set(b2)
		res.AssertionEventState6Enabled = isBit6Set(b2)
		res.AssertionEventState5Enabled = isBit5Set(b2)
		res.AssertionEventState4Enabled = isBit4Set(b2)
		res.AssertionEventState3Enabled = isBit3Set(b2)
		res.AssertionEventState2Enabled = isBit2Set(b2)
		res.AssertionEventState1Enabled = isBit1Set(b2)
		res.AssertionEventState0Enabled = isBit0Set(b2)
	}

	if len(msg) >= 3 {
		b3, _, _ := unpackUint8(msg, 2)
		res.AssertionEvent_UNR_GoHighEnabled = isBit3Set(b3)
		res.AssertionEvent_UNR_GoLowEnabled = isBit2Set(b3)
		res.AssertionEvent_UC_GoHighEnabled = isBit1Set(b3)
		res.AssertionEvent_UC_GoLowEnabled = isBit0Set(b3)
		res.AssertionEventState14Enabled = isBit6Set(b3)
		res.AssertionEventState13Enabled = isBit5Set(b3)
		res.AssertionEventState12Enabled = isBit4Set(b3)
		res.AssertionEventState11Enabled = isBit3Set(b3)
		res.AssertionEventState10Enabled = isBit2Set(b3)
		res.AssertionEventState9Enabled = isBit1Set(b3)
		res.AssertionEventState8Enabled = isBit0Set(b3)
	}

	if len(msg) >= 4 {
		b4, _, _ := unpackUint8(msg, 3)
		res.DeassertionEvent_UNC_GoHighEnabled = isBit7Set(b4)
		res.DeassertionEvent_UNC_GoLowEnabled = isBit6Set(b4)
		res.DeassertionEvent_LNR_GoHighEnabled = isBit5Set(b4)
		res.DeassertionEvent_LNR_GoLowEnabled = isBit4Set(b4)
		res.DeassertionEvent_LC_GoHighEnabled = isBit3Set(b4)
		res.DeassertionEvent_LC_GoLowEnabled = isBit2Set(b4)
		res.DeassertionEvent_LNC_GoHighEnabled = isBit1Set(b4)
		res.DeassertionEvent_LNC_GoLowEnabled = isBit0Set(b4)
		res.DeassertionEventState7Enabled = isBit7Set(b4)
		res.DeassertionEventState6Enabled = isBit6Set(b4)
		res.DeassertionEventState5Enabled = isBit5Set(b4)
		res.DeassertionEventState4Enabled = isBit4Set(b4)
		res.DeassertionEventState3Enabled = isBit3Set(b4)
		res.DeassertionEventState2Enabled = isBit2Set(b4)
		res.DeassertionEventState1Enabled = isBit1Set(b4)
		res.DeassertionEventState0Enabled = isBit0Set(b4)
	}

	if len(msg) >= 5 {
		b5, _, _ := unpackUint8(msg, 4)
		res.DeassertionEvent_UNR_GoHighEnabled = isBit3Set(b5)
		res.DeassertionEvent_UNR_GoLowEnabled = isBit2Set(b5)
		res.DeassertionEvent_UC_GoHighEnabled = isBit1Set(b5)
		res.DeassertionEvent_UC_GoLowEnabled = isBit0Set(b5)
		res.DeassertionEventState14Enabled = isBit6Set(b5)
		res.DeassertionEventState13Enabled = isBit5Set(b5)
		res.DeassertionEventState12Enabled = isBit4Set(b5)
		res.DeassertionEventState11Enabled = isBit3Set(b5)
		res.DeassertionEventState10Enabled = isBit2Set(b5)
		res.DeassertionEventState9Enabled = isBit1Set(b5)
		res.DeassertionEventState8Enabled = isBit0Set(b5)
	}

	return nil
}

func (r *GetSensorEventEnableResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorEventEnableResponse) Format() string {
	return ""
}

func (c *Client) GetSensorEventEnable(sensorNumber uint8) (response *GetSensorEventEnableResponse, err error) {
	request := &GetSensorEventEnableRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorEventEnableResponse{}
	err = c.Exchange(request, response)
	return
}
