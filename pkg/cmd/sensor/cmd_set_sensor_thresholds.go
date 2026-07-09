package sensor

import (
	"github.com/bougou/go-ipmi/pkg/types"
	// 35.8 Set Sensor Thresholds Command
)

type SetSensorThresholdsRequest struct {
	SensorNumber uint8

	// Set Threshold flag
	SetUNR bool
	SetUCR bool
	SetUNC bool
	SetLNR bool
	SetLCR bool
	SetLNC bool

	// Threshold value
	LNC_Raw uint8
	LCR_Raw uint8
	LNR_Raw uint8
	UNC_Raw uint8
	UCR_Raw uint8
	UNR_Raw uint8
}

type SetSensorThresholdsResponse struct {
	// empty
}

func (req *SetSensorThresholdsRequest) Command() types.Command {
	return types.CommandSetSensorThresholds
}

func (req *SetSensorThresholdsRequest) Pack() []byte {
	out := make([]byte, 8)
	types.PackUint8(req.SensorNumber, out, 0)

	var b uint8
	if req.SetUNR {
		b = types.SetBit5(b)
	}
	if req.SetUCR {
		b = types.SetBit4(b)
	}
	if req.SetUNC {
		b = types.SetBit3(b)
	}
	if req.SetLNR {
		b = types.SetBit2(b)
	}
	if req.SetLCR {
		b = types.SetBit1(b)
	}
	if req.SetLNC {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 1)

	types.PackUint8(req.LNC_Raw, out, 2)
	types.PackUint8(req.LCR_Raw, out, 3)
	types.PackUint8(req.LNR_Raw, out, 4)
	types.PackUint8(req.UNC_Raw, out, 5)
	types.PackUint8(req.UCR_Raw, out, 6)
	types.PackUint8(req.UNR_Raw, out, 7)
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

// SetSensorThresholds is to set the specified threshold for the given sensor.
// Note that the application issuing this command is responsible for ensuring that
// thresholds for a sensor are set in the proper order (e.g. that
// the upper critical threshold is set higher than the upper non-critical threshold)
//
//	Upper Non Recoverable area
//	-----------------UNR threshold
//	Upper Critical area
//	-----------------UCR threshold
//	Upper Non Critical area
//	-----------------UNC threshold
//	OK area
//	-----------------LNC threshold
//	Lower Non Critical area
//	-----------------LCR threshold
//	Lower Critical area
//	-----------------LNR threshold
//	Lower NonRecoverable area
