package ipmi

// 35.13 Get Sensor Event Status Command
type SetSensorReadingAndEventStatusRequest struct {
	SensorNumber uint8

	EventDataBytesOperation  uint8
	AssertionBitsOperation   uint8
	DeassertionBitsOperation uint8
	SensorReadingOperation   uint8

	SensorReading uint8

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

	EventData1 uint8
	EventData2 uint8
	EventData3 uint8
}

type SetSensorReadingAndEventStatusResponse struct {
	// empty
}

func (req *SetSensorReadingAndEventStatusRequest) Command() Command {
	return CommandSetSensorReadingAndEventStatus
}

func (req *SetSensorReadingAndEventStatusRequest) Pack() []byte {
	out := make([]byte, 9)
	packUint8(req.SensorNumber, out, 0)

	var operation uint8
	operation |= uint8(req.EventDataBytesOperation) << 6
	operation |= (uint8(req.AssertionBitsOperation) & 0x3f) << 4
	operation |= (uint8(req.DeassertionBitsOperation) & 0x0f) << 2
	operation |= uint8(req.SensorReadingOperation) & 0x03
	packUint8(operation, out, 1)

	packUint8(req.SensorReading, out, 2)

	// Todo determine sensor is threshold based or discrete

	return out
}

func (res *SetSensorReadingAndEventStatusResponse) Unpack(msg []byte) error {
	return nil
}

func (r *SetSensorReadingAndEventStatusResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "Attempt to change reading or set or clear status bits that are not settable via this command",
		0x81: "Attempted to set Event Data Bytes, but setting Event Data Bytes is not supported for this sensor.",
	}
}

func (res *SetSensorReadingAndEventStatusResponse) Format() string {
	return ""
}

func (c *Client) SetSensorReadingAndEventStatus(request *SetSensorReadingAndEventStatusRequest) (response *SetSensorReadingAndEventStatusResponse, err error) {
	response = &SetSensorReadingAndEventStatusResponse{}
	err = c.Exchange(request, response)
	return
}
