package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
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
	ActiveStates types.Mask_DiscreteEvent

	OptionalData1 uint8
	OptionalData2 uint8
}

func (req *GetSensorReadingRequest) Command() types.Command {
	return types.CommandGetSensorReading
}

func (req *GetSensorReadingRequest) Pack() []byte {
	out := make([]byte, 1)
	types.PackUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorReadingResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.Reading, _, _ = types.UnpackUint8(msg, 0)

	b1, _, _ := types.UnpackUint8(msg, 1)
	res.EventMessagesDisabled = !types.IsBit7Set(b1)
	res.SensorScanningDisabled = !types.IsBit6Set(b1)
	res.ReadingUnavailable = types.IsBit5Set(b1)

	if len(msg) >= 3 {
		b2, _, _ := types.UnpackUint8(msg, 2)
		// For threshold-based sensors, Present threshold comparison status
		res.Above_UNR = types.IsBit5Set(b2)
		res.Above_UCR = types.IsBit4Set(b2)
		res.Above_UNC = types.IsBit3Set(b2)
		res.Below_LNR = types.IsBit2Set(b2)
		res.Below_LCR = types.IsBit1Set(b2)
		res.Below_LNC = types.IsBit0Set(b2)
		// For discrete reading sensors
		res.ActiveStates.State_7 = types.IsBit7Set(b2)
		res.ActiveStates.State_6 = types.IsBit6Set(b2)
		res.ActiveStates.State_5 = types.IsBit5Set(b2)
		res.ActiveStates.State_4 = types.IsBit4Set(b2)
		res.ActiveStates.State_3 = types.IsBit3Set(b2)
		res.ActiveStates.State_2 = types.IsBit2Set(b2)
		res.ActiveStates.State_1 = types.IsBit1Set(b2)
		res.ActiveStates.State_0 = types.IsBit0Set(b2)

		res.OptionalData1 = b2
	}

	// For discrete reading sensors only. (Optional)
	if len(msg) >= 4 {
		b3, _, _ := types.UnpackUint8(msg, 3)
		// For discrete reading sensors
		res.ActiveStates.State_14 = types.IsBit6Set(b3)
		res.ActiveStates.State_13 = types.IsBit5Set(b3)
		res.ActiveStates.State_12 = types.IsBit4Set(b3)
		res.ActiveStates.State_11 = types.IsBit3Set(b3)
		res.ActiveStates.State_10 = types.IsBit2Set(b3)
		res.ActiveStates.State_9 = types.IsBit1Set(b3)
		res.ActiveStates.State_8 = types.IsBit0Set(b3)

		res.OptionalData2 = b3
	}

	return nil
}

func (r *GetSensorReadingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (r *GetSensorReadingResponse) ThresholdStatus() types.SensorThresholdStatus {
	if r.Above_UCR {
		return types.SensorThresholdStatus_UCR
	}
	if r.Above_UNC {
		return types.SensorThresholdStatus_UCR
	}
	if r.Above_UNR {
		return types.SensorThresholdStatus_UCR
	}
	if r.Below_LCR {
		return types.SensorThresholdStatus_UCR
	}
	if r.Below_LNC {
		return types.SensorThresholdStatus_UCR
	}
	if r.Below_LNR {
		return types.SensorThresholdStatus_UCR
	}
	return types.SensorThresholdStatus_OK
}

func (res *GetSensorReadingResponse) Format() string {
	return "" +
		fmt.Sprintf("Sensor Reading         : %d\n", res.Reading) +
		fmt.Sprintf("Event Message Disabled : %v\n", res.EventMessagesDisabled) +
		fmt.Sprintf("Scanning Disabled      : %v\n", res.SensorScanningDisabled) +
		fmt.Sprintf("Reading Unavailable    : %v\n", res.ReadingUnavailable) +
		fmt.Sprintf("Threshold Status       : %s\n", res.ThresholdStatus()) +
		fmt.Sprintf("Discrete Events        : %v\n", res.ActiveStates.TrueEvents())
}
