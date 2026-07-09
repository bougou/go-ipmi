package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 35.5 Get Sensor Reading Factors Command
type GetSensorReadingFactorsRequest struct {
	SensorNumber uint8
	Reading      uint8
}

type GetSensorReadingFactorsResponse struct {
	NextReading uint8
	types.ReadingFactors
}

func (req *GetSensorReadingFactorsRequest) Command() types.Command {
	return types.CommandGetSensorReadingFactors
}

func (req *GetSensorReadingFactorsRequest) Pack() []byte {
	out := make([]byte, 2)
	types.PackUint8(req.SensorNumber, out, 0)
	types.PackUint8(req.Reading, out, 1)
	return out
}

func (res *GetSensorReadingFactorsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 7)
	}

	res.NextReading, _, _ = types.UnpackUint8(msg, 0)

	b1, _, _ := types.UnpackUint8(msg, 1)
	b2, _, _ := types.UnpackUint8(msg, 2)

	m := uint16(b2&0xc0)<<2 | uint16(b1)
	res.M = int16(types.TwoSComplement(uint32(m), 10))

	res.Tolerance = b2 & 0x3f

	b3, _, _ := types.UnpackUint8(msg, 3)
	b4, _, _ := types.UnpackUint8(msg, 4)
	b5, _, _ := types.UnpackUint8(msg, 5)

	b := uint16(b4&0xc0)<<2 | uint16(b3)
	res.B = int16(types.TwoSComplement(uint32(b), 10))

	res.Accuracy = uint16(b5&0xf0)<<2 | uint16(b4&0x3f)
	res.Accuracy_Exp = (b5 & 0x0c) >> 2

	b6, _, _ := types.UnpackUint8(msg, 6)

	rExp := uint8((b6 & 0xf0) >> 4)
	res.R_Exp = int8(types.TwoSComplement(uint32(rExp), 4))

	bExp := uint8(b6 & 0x0f)
	res.B_Exp = int8(types.TwoSComplement(uint32(bExp), 4))

	return nil
}

func (r *GetSensorReadingFactorsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorReadingFactorsResponse) Format() string {
	return "" +
		fmt.Sprintf("M           : %d\n", res.M) +
		fmt.Sprintf("B           : %d\n", res.B) +
		fmt.Sprintf("B_Exp (K1)  : %d\n", res.B_Exp) +
		fmt.Sprintf("R_Exp (K2)  : %d\n", res.R_Exp) +
		fmt.Sprintf("Tolerance   : %d\n", res.Tolerance) +
		fmt.Sprintf("Accuracy    : %d\n", res.Accuracy) +
		fmt.Sprintf("AccuracyExp : %d\n", res.Accuracy_Exp)
}

// This command returns the Sensor Reading Factors fields for the specified reading value on the specified sensor.
