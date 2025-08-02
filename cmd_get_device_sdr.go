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

// The Get Device SDR command allows SDR information for sensors for a Sensor Device
// (typically implemented in a satellite management controller) to be returned.
//
// The Get Device SDR Command can return any type of SDR, not just Types 01h and 02h.
// This is an optional command for Static Sensor Devices, and mandatory for Dynamic Sensor Devices.
// The format and action of this command is similar to that for the Get SDR command
// for SDR Repository Devices.
//
// Sensor Devices that support the Get Device SDR command return SDR Records that
// match the SDR Repository formats.
func (c *Client) GetDeviceSDR(ctx context.Context, recordID uint16) (response *GetDeviceSDRResponse, err error) {
	request := &GetDeviceSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &GetDeviceSDRResponse{}
	err = c.Exchange(ctx, request, response)

	if respErr, ok := isResponseError(err); ok {
		if respErr.CompletionCode() == CompletionCodeCannotReturnRequestedDataBytes {
			return c.getDeviceSDR(ctx, recordID)
		}
	}

	return
}

// getDeviceSDR reads the Device SDR record in partial read way.
func (c *Client) getDeviceSDR(ctx context.Context, recordID uint16) (response *GetDeviceSDRResponse, err error) {

	var data []byte
	// the actual data length of the SDR can only be determined after the first GetSDR request/response.
	dataLength := uint8(0)

	reservationID := uint16(0)
	readBytes := uint8(16)
	readTotal := uint8(0)
	readOffset := uint8(0)

	for {
		request := &GetDeviceSDRRequest{
			ReservationID: reservationID,
			RecordID:      recordID,
			ReadOffset:    readOffset,
			ReadBytes:     readBytes,
		}
		response = &GetDeviceSDRResponse{}
		if err = c.Exchange(ctx, request, response); err != nil {
			return
		}

		// determine the total data length by parsing the SDR Header part
		if readOffset == 0 {
			if len(response.RecordData) < SDRRecordHeaderSize {
				return nil, fmt.Errorf("too short record data for SDR header (%d/%d)", len(response.RecordData), SDRRecordHeaderSize)
			}
			dataLength = response.RecordData[4] + uint8(SDRRecordHeaderSize)
			data = make([]byte, dataLength)
		}

		copy(data[readOffset:readOffset+readBytes], response.RecordData[:])

		readOffset += uint8(len(response.RecordData))
		readTotal += uint8(len(response.RecordData))

		if readTotal >= dataLength {
			break
		}

		if readOffset+readBytes > dataLength {
			// decrease the readBytes for the last read.
			readBytes = dataLength - readOffset
		}

		rsp, err := c.ReserveDeviceSDRRepo(ctx)
		if err == nil {
			reservationID = rsp.ReservationID
		} else {
			reservationID = 0
		}
	}

	return &GetDeviceSDRResponse{
		NextRecordID: response.NextRecordID,
		RecordData:   data,
	}, nil
}

func (c *Client) GetDeviceSDRBySensorID(ctx context.Context, sensorNumber uint8) (*SDR, error) {

	var recordID uint16 = 0
	for {
		res, err := c.GetDeviceSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR for recordID (%#0x) failed, err: %w", recordID, err)
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
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR for recordID (%#0x) failed, err: %w", recordID, err)
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
