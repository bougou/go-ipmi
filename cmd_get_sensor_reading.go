package ipmi

// 35.14 Get Sensor Reading Command
type GetSensorReadingRequest struct {
	SensorNumber uint8
}

type GetSensorReadingResponse struct {
	SensorReading uint8 // reading byte. Ignore on read if sensor does not return an numeric (analog) reading

	AllEventMessagesDisabled bool
	SensorScaningDisabled    bool
	ReadingStateUnavailable  bool // Software should use this bit to avoid getting an incorrect status while the first sensor update is in progress.

	Above_UNR_Threshold bool // at or above
	Above_UC_Threshold  bool // at or above
	Above_UNC_Threshold bool // at or above
	Below_LNR_Threshold bool // at or below
	Below_LC_Threshold  bool // at or below
	Below_LNC_Threshold bool // at or below
	State7Asserted      bool
	State6Asserted      bool
	State5Asserted      bool
	State4Asserted      bool
	State3Asserted      bool
	State2Asserted      bool
	State1Asserted      bool
	State0Asserted      bool

	State14Asserted bool
	State13Asserted bool
	State12Asserted bool
	State11Asserted bool
	State10Asserted bool
	State9Asserted  bool
	State8Asserted  bool
}

func (req *GetSensorReadingRequest) Command() Command {
	return CommandGetSensorReading
}

func (req *GetSensorReadingRequest) Pack() []byte {
	out := make([]byte, 1)
	packUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorReadingResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShort
	}
	res.SensorReading, _, _ = unpackUint8(msg, 0)

	b2, _, _ := unpackUint8(msg, 1)
	res.AllEventMessagesDisabled = !isBit7Set(b2)
	res.SensorScaningDisabled = !isBit6Set(b2)
	res.ReadingStateUnavailable = isBit5Set(b2)

	b3, _, _ := unpackUint8(msg, 2)
	// For threshold-based sensors, Present threshold comparison status
	res.Above_UNR_Threshold = isBit5Set(b3)
	res.Above_UC_Threshold = isBit4Set(b3)
	res.Above_UNC_Threshold = isBit3Set(b3)
	res.Below_LNR_Threshold = isBit2Set(b3)
	res.Below_LC_Threshold = isBit1Set(b3)
	res.Below_LNC_Threshold = isBit0Set(b3)
	// For discrete reading sensors
	res.State7Asserted = isBit7Set(b3)
	res.State6Asserted = isBit6Set(b3)
	res.State5Asserted = isBit5Set(b3)
	res.State4Asserted = isBit4Set(b3)
	res.State3Asserted = isBit3Set(b3)
	res.State2Asserted = isBit2Set(b3)
	res.State1Asserted = isBit1Set(b3)
	res.State0Asserted = isBit0Set(b3)

	b4, _, _ := unpackUint8(msg, 3)
	// For discrete reading sensors
	res.State14Asserted = isBit6Set(b4)
	res.State13Asserted = isBit5Set(b4)
	res.State12Asserted = isBit4Set(b4)
	res.State11Asserted = isBit3Set(b4)
	res.State10Asserted = isBit2Set(b4)
	res.State9Asserted = isBit1Set(b4)
	res.State8Asserted = isBit0Set(b4)
	return nil
}

func (r *GetSensorReadingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSensorReadingResponse) Format() string {

	return ""
}

func (c *Client) GetSensorReading(sensorNumber uint8) (response *GetSensorReadingResponse, err error) {
	request := &GetSensorReadingRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorReadingResponse{}
	err = c.Exchange(request, response)
	return
}
