package ipmi

import (
	"context"
	"fmt"
)

// 35.3 Get Device SDR Command
type GetDeviceSDRRequest struct {
	ReservationID uint16
	RecordID      uint16
	ReadOffset    uint8
	ReadBytes     uint8 // FFh means read entire record
}

type GetDeviceSDRResponse struct {
	NextRecordID uint16
	RecordData   []byte
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
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.NextRecordID, _, _ = unpackUint16L(msg, 0)
	res.RecordData, _, _ = unpackBytes(msg, 2, len(msg)-2)
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
func (c *Client) GetDeviceSDR(ctx context.Context, recordID uint16) (response *GetDeviceSDRResponse, err error) {
	request := &GetDeviceSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &GetDeviceSDRResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDeviceSDRBySensorID(ctx context.Context, sensorNumber uint8) (*SDR, error) {
	if SensorNumber(sensorNumber) == SensorNumberReserved {
		return nil, fmt.Errorf("not valid sensorNumber, %#0x is reserved", sensorNumber)
	}

	var recordID uint16 = 0
	for {
		res, err := c.GetDeviceSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}

		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}
		if uint8(sdr.SensorNumber()) == sensorNumber {
			return sdr, nil
		}

		recordID = res.NextRecordID
		if recordID == 0xffff {
			break
		}
	}

	return nil, fmt.Errorf("not found SDR for sensor id (%#0x)", sensorNumber)
}

func (c *Client) GetDeviceSDRs(ctx context.Context, recordTypes ...SDRRecordType) ([]*SDR, error) {
	var out = make([]*SDR, 0)
	var recordID uint16 = 0
	for {
		res, err := c.GetDeviceSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}

		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}

		if len(recordTypes) == 0 {
			out = append(out, sdr)
		} else {
			for _, v := range recordTypes {
				if sdr.RecordHeader.RecordType == v {
					out = append(out, sdr)
					break
				}
			}
		}

		recordID = res.NextRecordID
		if recordID == 0xffff {
			break
		}
	}
	return out, nil
}
