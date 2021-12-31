package ipmi

// 35.3 Get Device SDR Command
type GetDeviceSDRRequest struct {
	ReservationID uint16
	RecordID      uint16
	ReadOffset    uint8
	ReadBytes     uint8 // FFh means read entire record
}

type GetDeviceSDRResponse struct {
	NexRecordID uint16
	Data        []byte
}

func (req *GetDeviceSDRRequest) Command() Command {
	return CommandGetDeviceSDR
}

func (req *GetDeviceSDRRequest) Pack() []byte {
	out := make([]byte, 6)
	packUint16L(req.ReservationID, out, 0)
	packUint16L(req.RecordID, out, 2)
	packUint8(req.ReadOffset, out, 4)
	packUint8(req.ReadBytes, out, 5)
	return out
}

func (res *GetDeviceSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShort
	}

	res.NexRecordID, _, _ = unpackUint16L(msg, 0)
	res.Data, _, _ = unpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (r *GetDeviceSDRResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "record changed",
	}
}

func (res *GetDeviceSDRResponse) Format() string {
	return ""
}

// This command returns general information about the collection of sensors in a Dynamic Sensor Device.
func (c *Client) GetDeviceSDR(recordID uint16) (response *GetDeviceSDRResponse, err error) {
	request := &GetDeviceSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &GetDeviceSDRResponse{}
	err = c.Exchange(request, response)
	return
}
