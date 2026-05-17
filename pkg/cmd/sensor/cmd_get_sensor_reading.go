package sensor

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
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
	ActiveStates ipmi.Mask_DiscreteEvent

	OptionalData1 uint8
	OptionalData2 uint8
}

func (req *GetSensorReadingRequest) Command() ipmi.Command {
	return ipmi.CommandGetSensorReading
}

func (req *GetSensorReadingRequest) Pack() []byte {
	out := make([]byte, 1)
	ipmi.PackUint8(req.SensorNumber, out, 0)
	return out
}

func (res *GetSensorReadingResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.Reading, _, _ = ipmi.UnpackUint8(msg, 0)

	b1, _, _ := ipmi.UnpackUint8(msg, 1)
	res.EventMessagesDisabled = !ipmi.IsBit7Set(b1)
	res.SensorScanningDisabled = !ipmi.IsBit6Set(b1)
	res.ReadingUnavailable = ipmi.IsBit5Set(b1)

	if len(msg) >= 3 {
		b2, _, _ := ipmi.UnpackUint8(msg, 2)
		// For threshold-based sensors, Present threshold comparison status
		res.Above_UNR = ipmi.IsBit5Set(b2)
		res.Above_UCR = ipmi.IsBit4Set(b2)
		res.Above_UNC = ipmi.IsBit3Set(b2)
		res.Below_LNR = ipmi.IsBit2Set(b2)
		res.Below_LCR = ipmi.IsBit1Set(b2)
		res.Below_LNC = ipmi.IsBit0Set(b2)
		// For discrete reading sensors
		res.ActiveStates.State_7 = ipmi.IsBit7Set(b2)
		res.ActiveStates.State_6 = ipmi.IsBit6Set(b2)
		res.ActiveStates.State_5 = ipmi.IsBit5Set(b2)
		res.ActiveStates.State_4 = ipmi.IsBit4Set(b2)
		res.ActiveStates.State_3 = ipmi.IsBit3Set(b2)
		res.ActiveStates.State_2 = ipmi.IsBit2Set(b2)
		res.ActiveStates.State_1 = ipmi.IsBit1Set(b2)
		res.ActiveStates.State_0 = ipmi.IsBit0Set(b2)

		res.OptionalData1 = b2
	}

	// For discrete reading sensors only. (Optional)
	if len(msg) >= 4 {
		b3, _, _ := ipmi.UnpackUint8(msg, 3)
		// For discrete reading sensors
		res.ActiveStates.State_14 = ipmi.IsBit6Set(b3)
		res.ActiveStates.State_13 = ipmi.IsBit5Set(b3)
		res.ActiveStates.State_12 = ipmi.IsBit4Set(b3)
		res.ActiveStates.State_11 = ipmi.IsBit3Set(b3)
		res.ActiveStates.State_10 = ipmi.IsBit2Set(b3)
		res.ActiveStates.State_9 = ipmi.IsBit1Set(b3)
		res.ActiveStates.State_8 = ipmi.IsBit0Set(b3)

		res.OptionalData2 = b3
	}

	return nil
}

func (r *GetSensorReadingResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (r *GetSensorReadingResponse) ThresholdStatus() ipmi.SensorThresholdStatus {
	if r.Above_UCR {
		return ipmi.SensorThresholdStatus_UCR
	}
	if r.Above_UNC {
		return ipmi.SensorThresholdStatus_UCR
	}
	if r.Above_UNR {
		return ipmi.SensorThresholdStatus_UCR
	}
	if r.Below_LCR {
		return ipmi.SensorThresholdStatus_UCR
	}
	if r.Below_LNC {
		return ipmi.SensorThresholdStatus_UCR
	}
	if r.Below_LNR {
		return ipmi.SensorThresholdStatus_UCR
	}
	return ipmi.SensorThresholdStatus_OK
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
