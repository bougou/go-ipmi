package ipmi

import (
	"fmt"
)

// 35.14 Get Sensor Reading Command
//
// Retrieve a raw sensor reading (current reading) from ipmb
type GetSensorReadingRequest struct {
	SensorNumber uint8
}

type GetSensorReadingResponse struct {
	AnalogReading uint8 // reading byte. Ignore on read if sensor does not return an numeric (analog) reading

	EventMessagesDisabled bool
	SensorScaningDisabled bool
	ReadingUnavailable    bool // Software should use this bit to avoid getting an incorrect status while the first sensor update is in progress.

	// The following fields are optionally, they are only meaningful when reading is valid.

	Above_UNR_Threshold bool // at or above
	Above_UCR_Threshold bool // at or above
	Above_UNC_Threshold bool // at or above
	Below_LNR_Threshold bool // at or below
	Below_LCR_Threshold bool // at or below
	Below_LNC_Threshold bool // at or below

	State7Asserted bool
	State6Asserted bool
	State5Asserted bool
	State4Asserted bool
	State3Asserted bool
	State2Asserted bool
	State1Asserted bool
	State0Asserted bool

	State14Asserted bool
	State13Asserted bool
	State12Asserted bool
	State11Asserted bool
	State10Asserted bool
	State9Asserted  bool
	State8Asserted  bool

	optionalData1 uint8
	optionalData2 uint8
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
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}
	res.AnalogReading, _, _ = unpackUint8(msg, 0)

	b1, _, _ := unpackUint8(msg, 1)
	res.EventMessagesDisabled = !isBit7Set(b1)
	res.SensorScaningDisabled = !isBit6Set(b1)
	res.ReadingUnavailable = isBit5Set(b1)

	if len(msg) >= 3 {
		b2, _, _ := unpackUint8(msg, 2)
		// For threshold-based sensors, Present threshold comparison status
		res.Above_UNR_Threshold = isBit5Set(b2)
		res.Above_UCR_Threshold = isBit4Set(b2)
		res.Above_UNC_Threshold = isBit3Set(b2)
		res.Below_LNR_Threshold = isBit2Set(b2)
		res.Below_LCR_Threshold = isBit1Set(b2)
		res.Below_LNC_Threshold = isBit0Set(b2)
		// For discrete reading sensors
		res.State7Asserted = isBit7Set(b2)
		res.State6Asserted = isBit6Set(b2)
		res.State5Asserted = isBit5Set(b2)
		res.State4Asserted = isBit4Set(b2)
		res.State3Asserted = isBit3Set(b2)
		res.State2Asserted = isBit2Set(b2)
		res.State1Asserted = isBit1Set(b2)
		res.State0Asserted = isBit0Set(b2)

		res.optionalData1 = b2
	}

	// For discrete reading sensors only. (Optional)
	if len(msg) >= 4 {
		b3, _, _ := unpackUint8(msg, 3)
		// For discrete reading sensors
		res.State14Asserted = isBit6Set(b3)
		res.State13Asserted = isBit5Set(b3)
		res.State12Asserted = isBit4Set(b3)
		res.State11Asserted = isBit3Set(b3)
		res.State10Asserted = isBit2Set(b3)
		res.State9Asserted = isBit1Set(b3)
		res.State8Asserted = isBit0Set(b3)

		res.optionalData2 = b3
	}

	return nil
}

func (r *GetSensorReadingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (r *GetSensorReadingResponse) ThresholdStatus() SensorThresholdStatus {
	if r.Above_UCR_Threshold {
		return SensorThresholdStatus_UCR
	}
	if r.Above_UNC_Threshold {
		return SensorThresholdStatus_UCR
	}
	if r.Above_UNR_Threshold {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LCR_Threshold {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LNC_Threshold {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LNR_Threshold {
		return SensorThresholdStatus_UCR
	}
	return SensorThresholdStatus_OK
}

func (res *GetSensorReadingResponse) Format() string {
	return fmt.Sprintf(`
Sensor Reading         : %d
Event Message Disabled : %v
Scanning Disabled      : %v
Reading Unavailable    : %v
Threshold Status       : %s
`,
		res.AnalogReading,
		res.EventMessagesDisabled,
		res.SensorScaningDisabled,
		res.ReadingUnavailable,
		res.ThresholdStatus(),
	)
}

func (c *Client) GetSensorReading(sensorNumber uint8) (response *GetSensorReadingResponse, err error) {
	request := &GetSensorReadingRequest{
		SensorNumber: sensorNumber,
	}
	response = &GetSensorReadingResponse{}
	err = c.Exchange(request, response)
	return
}
