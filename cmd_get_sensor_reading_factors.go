package ipmi

// 35.5 Get Sensor Reading Factors Command
type GetSensorReadingFactorsRequest struct {
	SensorNumber uint8
	ReadingByte  uint8
}

type GetSensorReadingFactorsResponse struct {
	NextReading uint8

	M uint16 // 10 bits used

	// in +/- ½ raw counts
	Tolerance uint8  // 6 bits used
	B         uint16 // 10 bits used

	// Unsigned, 10-bit Basic Sensor Accuracy in 1/100 percent scaled up by unsigned Accuracy exponent.
	Accuracy         uint16 // 10 bits, unsigned
	AccuracyExponent uint8  // 2 bits, unsigned
	RExponent        int8   // 4 bits, signed
	BExponent        int8   // 4 bits, signed

}

func (req *GetSensorReadingFactorsRequest) Command() Command {
	return CommandGetSensorReadingFactors
}

func (req *GetSensorReadingFactorsRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint8(req.SensorNumber, out, 0)
	packUint8(req.ReadingByte, out, 1)
	return out
}

func (res *GetSensorReadingFactorsResponse) Unpack(msg []byte) error {
	if len(msg) < 7 {
		return ErrUnpackedDataTooShort
	}

	res.NextReading, _, _ = unpackUint8(msg, 0)

	b2, _, _ := unpackUint8(msg, 1)
	b3, _, _ := unpackUint8(msg, 2)

	m := uint16(b3)
	m = m >> 6
	m = m << 8
	m |= uint16(b2)
	res.M = m
	res.Tolerance = b3 & 0x3f // clear highest 2 bits

	b4, _, _ := unpackUint8(msg, 3)
	b5, _, _ := unpackUint8(msg, 4)
	b6, _, _ := unpackUint8(msg, 5)

	b := uint16(b5)
	b = b >> 6
	b = b << 8
	b |= uint16(b4)
	res.B = b

	a := uint16(b6)
	a = a >> 4
	a = a << 6
	a |= (uint16(b5 & 0x3f))
	res.Accuracy = a
	res.AccuracyExponent = (b6 & 0x0f) >> 2

	b7, _, _ := unpackUint8(msg, 6)
	res.RExponent = int8(b7 >> 4)
	res.BExponent = int8(b7 & 0x0f)

	return nil
}

func (r *GetSensorReadingFactorsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSensorReadingFactorsResponse) Format() string {
	return ""
}

// This command returns the Sensor Reading Factors fields for the specified reading value on the specified sensor.
func (c *Client) GetSensorReadingFactors(sensorNumber uint8, readingByte uint8) (response *GetSensorReadingFactorsResponse, err error) {
	request := &GetSensorReadingFactorsRequest{
		SensorNumber: sensorNumber,
		ReadingByte:  readingByte,
	}
	response = &GetSensorReadingFactorsResponse{}
	err = c.Exchange(request, response)
	return
}
