package ipmi

import (
	"context"
	"fmt"
)

// 35.5 Get Sensor Reading Factors Command
type GetSensorReadingFactorsRequest struct {
	SensorNumber uint8
	Reading      uint8
}

type GetSensorReadingFactorsResponse struct {
	NextReading uint8

	ReadingFactors
}

func (req *GetSensorReadingFactorsRequest) Command() Command {
	return CommandGetSensorReadingFactors
}

func (req *GetSensorReadingFactorsRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(req.Reading, out, 1)
	return out
}

func (res *GetSensorReadingFactorsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShortWith(len(msg), 7)
	}

	res.NextReading, _, _ = unpackUint8(msg, 0)

	b1, _, _ := unpackUint8(msg, 1)
	b2, _, _ := unpackUint8(msg, 2)

	m := uint16(b2&0xc0)<<2 | uint16(b1)
	res.M = int16(twosComplement(uint32(m), 10))

	res.Tolerance = b2 & 0x3f

	b3, _, _ := unpackUint8(msg, 3)
	b4, _, _ := unpackUint8(msg, 4)
	b5, _, _ := unpackUint8(msg, 5)

	b := uint16(b4&0xc0)<<2 | uint16(b3)
	res.B = int16(twosComplement(uint32(b), 10))

	res.Accuracy = uint16(b5&0xf0)<<2 | uint16(b4&0x3f)
	res.Accuracy_Exp = (b5 & 0x0c) >> 2

	b6, _, _ := unpackUint8(msg, 6)

	rExp := uint8((b6 & 0xf0) >> 4)
	res.R_Exp = int8(twosComplement(uint32(rExp), 4))

	bExp := uint8(b6 & 0x0f)
	res.B_Exp = int8(twosComplement(uint32(bExp), 4))

	return nil
}

func (r *GetSensorReadingFactorsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorReadingFactorsResponse) Format() string {
	return fmt.Sprintf(`M: %d
B: %d
B_Exp (K1): %d
R_Exp (K2): %d
Tolerance: %d
Accuracy: %d
AccuracyExp: %d`,
		res.M,
		res.B,
		res.B_Exp,
		res.R_Exp,
		res.Tolerance,
		res.Accuracy,
		res.Accuracy_Exp,
	)
}

// This command returns the Sensor Reading Factors fields for the specified reading value on the specified sensor.
func (c *Client) GetSensorReadingFactors(ctx context.Context, sensorNumber uint8, reading uint8) (response *GetSensorReadingFactorsResponse, err error) {
	request := &GetSensorReadingFactorsRequest{
		SensorNumber: sensorNumber,
		Reading:      reading,
	}
	response = &GetSensorReadingFactorsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
