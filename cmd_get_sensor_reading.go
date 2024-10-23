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
	// reading byte. Ignore on read if sensor does not return an numeric (analog) reading
	Reading uint8

	EventMessagesDisabled bool

	// see 16.5 System Software use of Sensor Scanning bits & Entity Info
	//
	// System software must ignore any sensor that has the sensor scanning bit disabled - if system software didn't disable the sensor.
	SensorScanningDisabled bool

	ReadingUnavailable bool // Software should use this bit to avoid getting an incorrect status while the first sensor update is in progress.

	// The following fields are optionally, they are only meaningful when reading is valid.

	Above_UNR bool // at or above UNR threshold
	Above_UCR bool // at or above UCR threshold
	Above_UNC bool // at or above UNC threshold
	Below_LNR bool // at or below LNR threshold
	Below_LCR bool // at or below LCR threshold
	Below_LNC bool // at or below LNC threshold

	// see 42.1
	// (Sensor Classes: Discrete)
	// It is possible for a discrete sensor to have more than one state active at a time.
	ActiveStates Mask_DiscreteEvent

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
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.Reading, _, _ = unpackUint8(msg, 0)

	b1, _, _ := unpackUint8(msg, 1)
	res.EventMessagesDisabled = !isBit7Set(b1)
	res.SensorScanningDisabled = !isBit6Set(b1)
	res.ReadingUnavailable = isBit5Set(b1)

	if len(msg) >= 3 {
		b2, _, _ := unpackUint8(msg, 2)
		// For threshold-based sensors, Present threshold comparison status
		res.Above_UNR = isBit5Set(b2)
		res.Above_UCR = isBit4Set(b2)
		res.Above_UNC = isBit3Set(b2)
		res.Below_LNR = isBit2Set(b2)
		res.Below_LCR = isBit1Set(b2)
		res.Below_LNC = isBit0Set(b2)
		// For discrete reading sensors
		res.ActiveStates.State_7 = isBit7Set(b2)
		res.ActiveStates.State_6 = isBit6Set(b2)
		res.ActiveStates.State_5 = isBit5Set(b2)
		res.ActiveStates.State_4 = isBit4Set(b2)
		res.ActiveStates.State_3 = isBit3Set(b2)
		res.ActiveStates.State_2 = isBit2Set(b2)
		res.ActiveStates.State_1 = isBit1Set(b2)
		res.ActiveStates.State_0 = isBit0Set(b2)

		res.optionalData1 = b2
	}

	// For discrete reading sensors only. (Optional)
	if len(msg) >= 4 {
		b3, _, _ := unpackUint8(msg, 3)
		// For discrete reading sensors
		res.ActiveStates.State_14 = isBit6Set(b3)
		res.ActiveStates.State_13 = isBit5Set(b3)
		res.ActiveStates.State_12 = isBit4Set(b3)
		res.ActiveStates.State_11 = isBit3Set(b3)
		res.ActiveStates.State_10 = isBit2Set(b3)
		res.ActiveStates.State_9 = isBit1Set(b3)
		res.ActiveStates.State_8 = isBit0Set(b3)

		res.optionalData2 = b3
	}

	return nil
}

func (r *GetSensorReadingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (r *GetSensorReadingResponse) ThresholdStatus() SensorThresholdStatus {
	if r.Above_UCR {
		return SensorThresholdStatus_UCR
	}
	if r.Above_UNC {
		return SensorThresholdStatus_UCR
	}
	if r.Above_UNR {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LCR {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LNC {
		return SensorThresholdStatus_UCR
	}
	if r.Below_LNR {
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
Discrete Events        : %v
`,
		res.Reading,
		res.EventMessagesDisabled,
		res.SensorScanningDisabled,
		res.ReadingUnavailable,
		res.ThresholdStatus(),
		res.ActiveStates.TrueEvents(),
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
