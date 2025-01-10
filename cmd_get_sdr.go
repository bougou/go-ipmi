package ipmi

import (
	"context"
	"fmt"
)

// 33.12 Get SDR Command
type GetSDRRequest struct {
	ReservationID uint16 // LS Byte first
	RecordID      uint16 // LS Byte first
	ReadOffset    uint8  // Offset into record
	ReadBytes     uint8  // FFh means read entire record
}

type GetSDRResponse struct {
	NextRecordID uint16
	RecordData   []byte
}

func (req *GetSDRRequest) Pack() []byte {
	msg := make([]byte, 6)
	packUint16L(req.ReservationID, msg, 0)
	packUint16L(req.RecordID, msg, 2)
	packUint8(req.ReadOffset, msg, 4)
	packUint8(req.ReadBytes, msg, 5)
	return msg
}

func (req *GetSDRRequest) Command() Command {
	return CommandGetSDR
}

func (res *GetSDRResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.NextRecordID, _, _ = unpackUint16L(msg, 0)
	res.RecordData, _, _ = unpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSDRResponse) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *GetSDRResponse) Format() string {
	return fmt.Sprintf("%v", res)
}

// GetSDR returns raw SDR record.
func (c *Client) GetSDR(ctx context.Context, recordID uint16) (response *GetSDRResponse, err error) {
	request := &GetSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &GetSDRResponse{}
	err = c.Exchange(ctx, request, response)

	// try read partial data if err (ResponseError and CompletionCode) indicate
	// reading full data (0xff) exceeds the maximum transfer length for the interface
	if resErr, ok := err.(*ResponseError); ok {
		if resErr.CompletionCode() == CompletionCodeCannotReturnRequestedDataBytes {
			return c.getSDR(ctx, recordID)
		}
	}

	return
}

func (c *Client) GetSDREnhanced(ctx context.Context, recordID uint16) (*SDR, error) {
	res, err := c.GetSDR(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %s", recordID, err)
	}

	sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
	if err != nil {
		return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
	}

	if err := c.enhanceSDR(ctx, sdr); err != nil {
		return sdr, fmt.Errorf("enhanceSDR failed, err: %w", err)
	}

	return sdr, nil
}

// getSDR return SDR in a partial read way.
func (c *Client) getSDR(ctx context.Context, recordID uint16) (response *GetSDRResponse, err error) {
	var data []byte
	// the actual data length of the SDR can only be determined after the first GetSDR request/response.
	dataLength := uint8(0)

	reservationID := uint16(0)
	readBytes := uint8(16)
	readTotal := uint8(0)
	readOffset := uint8(0)

	for {
		request := &GetSDRRequest{
			ReservationID: reservationID,
			RecordID:      recordID,
			ReadOffset:    readOffset,
			ReadBytes:     readBytes,
		}
		response = &GetSDRResponse{}
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

		rsp, err := c.ReserveSDRRepo(ctx)
		if err == nil {
			reservationID = rsp.ReservationID
		} else {
			reservationID = 0
		}
	}

	return &GetSDRResponse{
		NextRecordID: response.NextRecordID,
		RecordData:   data,
	}, nil
}

func (c *Client) GetSDRBySensorID(ctx context.Context, sensorNumber uint8) (*SDR, error) {
	if SensorNumber(sensorNumber) == SensorNumberReserved {
		return nil, fmt.Errorf("not valid sensorNumber, %#0x is reserved", sensorNumber)
	}

	var recordID uint16 = 0
	for {
		res, err := c.GetSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %s", recordID, err)
		}
		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
		}
		if uint8(sdr.SensorNumber()) != sensorNumber {
			recordID = sdr.NextRecordID
			if recordID == 0xffff {
				break
			}
			continue
		}

		if err := c.enhanceSDR(ctx, sdr); err != nil {
			return sdr, fmt.Errorf("enhanceSDR failed, err: %w", err)
		}
		return sdr, nil
	}

	return nil, fmt.Errorf("not found SDR for sensor id (%#0x)", sensorNumber)
}

func (c *Client) GetSDRBySensorName(ctx context.Context, sensorName string) (*SDR, error) {
	var recordID uint16 = 0
	for {
		res, err := c.GetSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %s", recordID, err)
		}
		sdr, err := ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
		}

		if sdr.SensorName() != sensorName {
			recordID = sdr.NextRecordID
			if recordID == 0xffff {
				break
			}
			continue
		}

		if err := c.enhanceSDR(ctx, sdr); err != nil {
			return sdr, fmt.Errorf("enhanceSDR failed, err: %w", err)
		}
		return sdr, nil
	}

	return nil, fmt.Errorf("not found SDR for sensor name (%s)", sensorName)
}

// GetSDRs fetches the SDR records with the specified RecordTypes.
// The parameter is a slice of SDRRecordType used as filter.
// Empty means to get all SDR records.
func (c *Client) GetSDRs(ctx context.Context, recordTypes ...SDRRecordType) ([]*SDR, error) {
	var recordID uint16 = 0
	var out = make([]*SDR, 0)
	for {
		sdr, err := c.GetSDREnhanced(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR for recordID (%#0x) failed, err: %s", recordID, err)
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

		recordID = sdr.NextRecordID
		if recordID == 0xffff {
			break
		}
	}

	return out, nil
}

// GetSDRsMap returns all Full/Compact SDRs grouped by GeneratorID and SensorNumber.
// The sensor name can only be got from SDR record.
// So use this method to construct a map from which you can get sensor name.
func (c *Client) GetSDRsMap(ctx context.Context) (SDRMapBySensorNumber, error) {
	var out = make(map[GeneratorID]map[SensorNumber]*SDR)

	var recordID uint16 = 0
	for {
		sdr, err := c.GetSDREnhanced(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR for recordID (%#0x) failed, err: %s", recordID, err)
		}

		var generatorID GeneratorID
		var sensorNumber SensorNumber

		recordType := sdr.RecordHeader.RecordType
		switch recordType {
		case SDRRecordTypeFullSensor:
			generatorID = sdr.Full.GeneratorID
			sensorNumber = sdr.Full.SensorNumber
		case SDRRecordTypeCompactSensor:
			generatorID = sdr.Compact.GeneratorID
			sensorNumber = sdr.Compact.SensorNumber
		}

		if recordType == SDRRecordTypeFullSensor || recordType == SDRRecordTypeCompactSensor {
			if _, ok := out[generatorID]; !ok {
				out[generatorID] = make(map[SensorNumber]*SDR)
			}
			out[generatorID][sensorNumber] = sdr
		}

		recordID = sdr.NextRecordID
		if recordID == 0xffff {
			break
		}
	}

	return out, nil
}
