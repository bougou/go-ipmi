package client

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) ReserveSDRRepo(ctx context.Context) (response *storage.ReserveSDRRepoResponse, err error) {
	request := &storage.ReserveSDRRepoRequest{}
	response = &storage.ReserveSDRRepoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command returns general information about the collection of sensors in a Dynamic Sensor Device.
func (c *Client) GetDeviceSDRInfo(ctx context.Context, getSDRCount bool) (response *storage.GetDeviceSDRInfoResponse, err error) {
	request := &storage.GetDeviceSDRInfoRequest{
		GetSDRCount: getSDRCount,
	}
	response = &storage.GetDeviceSDRInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) DeleteSELEntry(ctx context.Context, recordID uint16, reservationID uint16) (response *storage.DeleteSELEntryResponse, err error) {
	request := &storage.DeleteSELEntryRequest{
		ReservationID: reservationID,
		RecordID:      recordID,
	}
	response = &storage.DeleteSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
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
func (c *Client) GetDeviceSDR(ctx context.Context, recordID uint16) (response *storage.GetDeviceSDRResponse, err error) {
	request := &storage.GetDeviceSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &storage.GetDeviceSDRResponse{}
	err = c.Exchange(ctx, request, response)

	if respErr, ok := types.IsResponseError(err); ok {
		if respErr.CompletionCode() == types.CompletionCodeCannotReturnRequestedDataBytes {
			return c.getDeviceSDR(ctx, recordID)
		}
	}

	return
}

// getDeviceSDR reads the Device SDR record in partial read way.
func (c *Client) getDeviceSDR(ctx context.Context, recordID uint16) (response *storage.GetDeviceSDRResponse, err error) {

	var data []byte

	dataLength := uint8(0)

	reservationID := uint16(0)
	readBytes := uint8(16)
	readTotal := uint8(0)
	readOffset := uint8(0)

	for {
		request := &storage.GetDeviceSDRRequest{
			ReservationID: reservationID,
			RecordID:      recordID,
			ReadOffset:    readOffset,
			ReadBytes:     readBytes,
		}
		response = &storage.GetDeviceSDRResponse{}
		if err = c.Exchange(ctx, request, response); err != nil {
			return
		}

		if readOffset == 0 {
			if len(response.RecordData) < types.SDRRecordHeaderSize {
				return nil, fmt.Errorf("too short record data for SDR header (%d/%d)", len(response.RecordData), types.SDRRecordHeaderSize)
			}
			dataLength = response.RecordData[4] + uint8(types.SDRRecordHeaderSize)
			data = make([]byte, dataLength)
		}

		copy(data[readOffset:readOffset+readBytes], response.RecordData[:])

		readOffset += uint8(len(response.RecordData))
		readTotal += uint8(len(response.RecordData))

		if readTotal >= dataLength {
			break
		}

		if readOffset+readBytes > dataLength {

			readBytes = dataLength - readOffset
		}

		rsp, err := c.ReserveDeviceSDRRepo(ctx)
		if err == nil {
			reservationID = rsp.ReservationID
		} else {
			reservationID = 0
		}
	}

	return &storage.GetDeviceSDRResponse{
		NextRecordID: response.NextRecordID,
		RecordData:   data,
	}, nil
}

func (c *Client) GetDeviceSDRBySensorID(ctx context.Context, sensorNumber uint8) (*types.SDR, error) {

	var recordID uint16 = 0
	for {
		res, err := c.GetDeviceSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		sdr, err := types.ParseSDR(res.RecordData, res.NextRecordID)
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

func (c *Client) GetDeviceSDRs(ctx context.Context, recordTypes ...types.SDRRecordType) ([]*types.SDR, error) {
	var out = make([]*types.SDR, 0)
	var recordID uint16 = 0
	for {
		res, err := c.GetDeviceSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetDeviceSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		sdr, err := types.ParseSDR(res.RecordData, res.NextRecordID)
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

// The command returns the specified data from the FRU Inventory Info area.
func (c *Client) ReadFRUData(ctx context.Context, fruDeviceID uint8, readOffset uint16, readCount uint8) (response *storage.ReadFRUDataResponse, err error) {
	request := &storage.ReadFRUDataRequest{
		FRUDeviceID: fruDeviceID,
		ReadOffset:  readOffset,
		ReadCount:   readCount,
	}
	response = &storage.ReadFRUDataResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// readFRUDataByLength reads FRU Data in loop until reaches the specified data length
func (c *Client) readFRUDataByLength(ctx context.Context, deviceID uint8, offset uint16, length uint16) ([]byte, error) {
	var data []byte
	c.Debugf("Read FRU Data by Length, offset: (%d), length: (%d)\n", offset, length)

	for {
		if length <= 0 {
			break
		}

		res, err := c.tryReadFRUData(ctx, deviceID, offset, length)
		if err != nil {
			return nil, fmt.Errorf("tryReadFRUData failed, err: %w", err)
		}
		c.Debug("", res.Format())
		data = append(data, res.Data...)

		length -= uint16(res.CountReturned)
		c.Debugf("left length: %d\n", length)

		offset += uint16(res.CountReturned)
	}

	return data, nil
}

// tryReadFRUData will try to read FRU data with a read count which starts with
// the minimal number of the specified length and the hard-coded 32, if the
// ReadFRUData failed, it try another request with a decreased read count.
func (c *Client) tryReadFRUData(ctx context.Context, deviceID uint8, readOffset uint16, length uint16) (response *storage.ReadFRUDataResponse, err error) {
	var readCount uint8 = 32
	if length <= uint16(readCount) {
		readCount = uint8(length)
	}

	for {
		if readCount <= 0 {
			return nil, fmt.Errorf("nothing to read")
		}

		c.Debugf("Try Read FRU Data, offset: (%d), count: (%d)\n", readOffset, readCount)
		res, err := c.ReadFRUData(ctx, deviceID, readOffset, readCount)
		if err == nil {
			return res, nil
		}

		if respErr, ok := types.IsResponseError(err); ok {
			cc := respErr.CompletionCode()
			if storage.ReadFRUDataLength2Big(cc) {
				readCount -= 1
				continue
			}
		}

		return nil, fmt.Errorf("ReadFRUData failed, err: %w", err)
	}
}

// The reservationID is only required for partial Get, use 0000h otherwise.
func (c *Client) GetSELEntry(ctx context.Context, reservationID uint16, recordID uint16) (response *storage.GetSELEntryResponse, err error) {
	if _, err := c.GetSELInfo(ctx); err != nil {
		return nil, fmt.Errorf("GetSELInfo failed, err: %w", err)
	}

	request := &storage.GetSELEntryRequest{
		ReservationID: reservationID,
		RecordID:      recordID,
		Offset:        0,
		ReadBytes:     0xff,
	}
	response = &storage.GetSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetSELEntries return all SEL records starting from the specified recordID.
// Pass 0 means retrieve all SEL entries starting from the first record.
func (c *Client) GetSELEntries(ctx context.Context, startRecordID uint16) ([]*types.SEL, error) {
	if _, err := c.GetSELInfo(ctx); err != nil {
		return nil, fmt.Errorf("GetSELInfo failed, err: %w", err)
	}

	var out = make([]*types.SEL, 0)
	var recordID uint16 = startRecordID
	for {
		selEntry, err := c.GetSELEntry(ctx, 0, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSELEntry failed, err: %w", err)
		}
		c.DebugBytes("sel entry record data", selEntry.Data, 16)

		sel, err := types.ParseSEL(selEntry.Data)
		if err != nil {
			return nil, fmt.Errorf("unpackSEL record failed, err: %w", err)
		}
		out = append(out, sel)

		recordID = selEntry.NextRecordID
		if recordID == 0xffff {
			break
		}
	}

	return out, nil
}

func (c *Client) GetSELEntriesStream(ctx context.Context, startRecordID uint16) iter.Seq[*types.Result[types.SEL]] {
	return func(yield func(*types.Result[types.SEL]) bool) {
		var recordID uint16 = startRecordID

	loop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if _, err := c.GetSELInfo(ctx); err != nil {
					yield(&types.Result[types.SEL]{Err: err})
					return
				}

				selEntry, err := c.GetSELEntry(ctx, 0, recordID)
				if err != nil {
					yield(&types.Result[types.SEL]{Err: err})
					return
				}

				sel, err := types.ParseSEL(selEntry.Data)
				if err != nil {
					yield(&types.Result[types.SEL]{Err: err})
					return
				}

				if !yield(&types.Result[types.SEL]{Ok: sel}) {
					return
				}

				recordID = selEntry.NextRecordID
				if recordID == 0xffff {
					break loop
				}
			}
		}
	}
}

func (c *Client) GetSELTime(ctx context.Context) (response *storage.GetSELTimeResponse, err error) {
	request := &storage.GetSELTimeRequest{}
	response = &storage.GetSELTimeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetSELTimeUTCOffset is used to retrieve the SEL Time UTC Offset (timezone)
func (c *Client) GetSELTimeUTCOffset(ctx context.Context) (response *storage.GetSELTimeUTCOffsetResponse, err error) {
	request := &storage.GetSELTimeUTCOffsetRequest{}
	response = &storage.GetSELTimeUTCOffsetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// The command writes the specified byte or word to the FRU Inventory Info area. This is a low level direct interface to a non-volatile storage area. This means that the interface does not interpret or check any semantics or formatting for the data being written.
func (c *Client) WriteFRUData(ctx context.Context, fruDeviceID uint8, writeOffset uint16, writeData []byte) (response *storage.WriteFRUDataResponse, err error) {
	request := &storage.WriteFRUDataRequest{
		FRUDeviceID: fruDeviceID,
		WriteOffset: writeOffset,
		WriteData:   writeData,
	}
	response = &storage.WriteFRUDataResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command returns overall the size of the FRU Inventory Area in this device, in bytes.
func (c *Client) GetFRUInventoryAreaInfo(ctx context.Context, fruDeviceID uint8) (response *storage.GetFRUInventoryAreaInfoResponse, err error) {
	request := &storage.GetFRUInventoryAreaInfoRequest{
		FRUDeviceID: fruDeviceID,
	}
	response = &storage.GetFRUInventoryAreaInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ReserveSEL(ctx context.Context) (response *storage.ReserveSELResponse, err error) {
	request := &storage.ReserveSELRequest{}
	response = &storage.ReserveSELResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) AddSELEntry(ctx context.Context, sel *types.SEL) (response *storage.AddSELEntryResponse, err error) {
	request := &storage.AddSELEntryRequest{
		SEL: sel,
	}
	response = &storage.AddSELEntryResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSDRRepoAllocInfo(ctx context.Context) (response *storage.GetSDRRepoAllocInfoResponse, err error) {
	request := &storage.GetSDRRepoAllocInfoRequest{}
	response = &storage.GetSDRRepoAllocInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// SetSELTimeUTCOffset initializes and retrieve a UTC offset (timezone) that is associated with the SEL Time
func (c *Client) SetSELTimeUTCOffset(ctx context.Context, minutesOffset int16) (response *storage.SetSELTimeUTCOffsetResponse, err error) {
	request := &storage.SetSELTimeUTCOffsetRequest{
		MinutesOffset: minutesOffset,
	}
	response = &storage.SetSELTimeUTCOffsetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ClearSEL(ctx context.Context, reservationID uint16) (response *storage.ClearSELResponse, err error) {
	request := &storage.ClearSELRequest{
		ReservationID:        reservationID,
		GetErasureStatusFlag: false,
	}
	response = &storage.ClearSELResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command is used to obtain a Reservation ID.
func (c *Client) ReserveDeviceSDRRepo(ctx context.Context) (response *storage.ReserveDeviceSDRRepoResponse, err error) {
	request := &storage.ReserveDeviceSDRRepoRequest{}
	response = &storage.ReserveDeviceSDRRepoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSELInfo(ctx context.Context) (response *storage.GetSELInfoResponse, err error) {
	request := &storage.GetSELInfoRequest{}
	response = &storage.GetSELInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetSDR returns raw SDR record.
func (c *Client) GetSDR(ctx context.Context, recordID uint16) (response *storage.GetSDRResponse, err error) {
	request := &storage.GetSDRRequest{
		ReservationID: 0,
		RecordID:      recordID,
		ReadOffset:    0,
		ReadBytes:     0xff,
	}
	response = &storage.GetSDRResponse{}
	err = c.Exchange(ctx, request, response)

	if respErr, ok := types.IsResponseError(err); ok {
		if respErr.CompletionCode() == types.CompletionCodeCannotReturnRequestedDataBytes {
			return c.getSDR(ctx, recordID)
		}
	}

	return
}

func (c *Client) GetSDREnhanced(ctx context.Context, recordID uint16) (*types.SDR, error) {
	res, err := c.GetSDR(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %w", recordID, err)
	}

	sdr, err := types.ParseSDR(res.RecordData, res.NextRecordID)
	if err != nil {
		return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
	}

	if err := c.enhanceSDR(ctx, sdr); err != nil {
		return sdr, fmt.Errorf("enhanceSDR failed, err: %w", err)
	}

	return sdr, nil
}

// getSDR return SDR in a partial read way.
func (c *Client) getSDR(ctx context.Context, recordID uint16) (response *storage.GetSDRResponse, err error) {
	var data []byte

	dataLength := uint8(0)

	reservationID := uint16(0)
	readBytes := uint8(16)
	readTotal := uint8(0)
	readOffset := uint8(0)

	for {
		request := &storage.GetSDRRequest{
			ReservationID: reservationID,
			RecordID:      recordID,
			ReadOffset:    readOffset,
			ReadBytes:     readBytes,
		}
		response = &storage.GetSDRResponse{}
		if err = c.Exchange(ctx, request, response); err != nil {
			return
		}

		if readOffset == 0 {
			if len(response.RecordData) < types.SDRRecordHeaderSize {
				return nil, fmt.Errorf("too short record data for SDR header (%d/%d)", len(response.RecordData), types.SDRRecordHeaderSize)
			}
			dataLength = response.RecordData[4] + uint8(types.SDRRecordHeaderSize)
			data = make([]byte, dataLength)
		}

		copy(data[readOffset:readOffset+readBytes], response.RecordData[:])

		readOffset += uint8(len(response.RecordData))
		readTotal += uint8(len(response.RecordData))

		if readTotal >= dataLength {
			break
		}

		if readOffset+readBytes > dataLength {

			readBytes = dataLength - readOffset
		}

		rsp, err := c.ReserveSDRRepo(ctx)
		if err == nil {
			reservationID = rsp.ReservationID
		} else {
			reservationID = 0
		}
	}

	return &storage.GetSDRResponse{
		NextRecordID: response.NextRecordID,
		RecordData:   data,
	}, nil
}

func (c *Client) GetSDRBySensorID(ctx context.Context, sensorNumber uint8) (*types.SDR, error) {
	var recordID uint16 = 0
	for {
		res, err := c.GetSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %w", recordID, err)
		}
		sdr, err := types.ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
		}

		recordType := sdr.RecordHeader.RecordType

		if uint8(sdr.SensorNumber()) != sensorNumber || (recordType != types.SDRRecordTypeFullSensor && recordType != types.SDRRecordTypeCompactSensor && recordType != types.SDRRecordTypeEventOnly) {
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

func (c *Client) GetSDRBySensorName(ctx context.Context, sensorName string) (*types.SDR, error) {
	var recordID uint16 = 0
	for {
		res, err := c.GetSDR(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR failed for recordID (%#02x), err: %w", recordID, err)
		}
		sdr, err := types.ParseSDR(res.RecordData, res.NextRecordID)
		if err != nil {
			return nil, fmt.Errorf("ParseSDR failed, err: %w", err)
		}

		recordType := sdr.RecordHeader.RecordType

		if sdr.SensorName() != sensorName || (recordType != types.SDRRecordTypeFullSensor && recordType != types.SDRRecordTypeCompactSensor && recordType != types.SDRRecordTypeEventOnly) {
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
func (c *Client) GetSDRs(ctx context.Context, recordTypes ...types.SDRRecordType) ([]*types.SDR, error) {
	var recordID uint16 = 0
	var out = make([]*types.SDR, 0)
	for {
		sdr, err := c.GetSDREnhanced(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		if sdr.RecordHeader == nil {
			continue
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

func (c *Client) GetSDRsStream(ctx context.Context, recordTypes ...types.SDRRecordType) iter.Seq[*types.Result[types.SDR]] {
	return func(yield func(*types.Result[types.SDR]) bool) {
		var recordID uint16 = 0

	loop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				sdr, err := c.GetSDREnhanced(ctx, recordID)
				if err != nil {
					yield(&types.Result[types.SDR]{Err: err})
					return
				}

				if sdr.RecordHeader == nil {
					continue loop
				}

				if len(recordTypes) == 0 {
					if !yield(&types.Result[types.SDR]{Ok: sdr}) {
						return
					}
				}

				for _, v := range recordTypes {
					if sdr.RecordHeader.RecordType == v {
						if !yield(&types.Result[types.SDR]{Ok: sdr}) {
							return
						}
						break
					}
				}

				recordID = sdr.NextRecordID
				if recordID == 0xffff {
					break loop
				}
			}
		}
	}
}

// GetSDRsMap returns all Full/Compact SDRs grouped by GeneratorID and SensorNumber.
// The sensor name can only be got from SDR record.
// So use this method to construct a map from which you can get sensor name.
func (c *Client) GetSDRsMap(ctx context.Context) (types.SDRMapBySensorNumber, error) {
	var out = make(map[types.GeneratorID]map[types.SensorNumber]*types.SDR)

	var recordID uint16 = 0
	for {
		sdr, err := c.GetSDREnhanced(ctx, recordID)
		if err != nil {
			return nil, fmt.Errorf("GetSDR for recordID (%#0x) failed, err: %w", recordID, err)
		}

		var generatorID types.GeneratorID
		var sensorNumber types.SensorNumber

		recordType := sdr.RecordHeader.RecordType
		switch recordType {
		case types.SDRRecordTypeFullSensor:
			generatorID = sdr.Full.GeneratorID
			sensorNumber = sdr.Full.SensorNumber
		case types.SDRRecordTypeCompactSensor:
			generatorID = sdr.Compact.GeneratorID
			sensorNumber = sdr.Compact.SensorNumber
		}

		if recordType == types.SDRRecordTypeFullSensor || recordType == types.SDRRecordTypeCompactSensor {
			if _, ok := out[generatorID]; !ok {
				out[generatorID] = make(map[types.SensorNumber]*types.SDR)
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

// enhanceSDR fills live sensor reading data into a Full or Compact SDR record.
func (c *Client) enhanceSDR(ctx context.Context, sdr *types.SDR) error {
	if sdr == nil {
		return nil
	}
	if sdr.RecordHeader.RecordType != types.SDRRecordTypeFullSensor &&
		sdr.RecordHeader.RecordType != types.SDRRecordTypeCompactSensor {
		return nil
	}
	sensor, err := c.sdrToSensor(ctx, sdr)
	if err != nil {
		return fmt.Errorf("sdrToSensor failed, err: %w", err)
	}
	switch sdr.RecordHeader.RecordType {
	case types.SDRRecordTypeFullSensor:
		sdr.Full.SensorValue = sensor.Value
		sdr.Full.SensorStatus = sensor.Status()
	case types.SDRRecordTypeCompactSensor:
		sdr.Compact.SensorValue = sensor.Value
		sdr.Compact.SensorStatus = sensor.Status()
	}
	return nil
}

func (c *Client) GetSELAllocInfo(ctx context.Context) (response *storage.GetSELAllocInfoResponse, err error) {
	request := &storage.GetSELAllocInfoRequest{}
	response = &storage.GetSELAllocInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSDRRepoInfo(ctx context.Context) (response *storage.GetSDRRepoInfoResponse, err error) {
	request := &storage.GetSDRRepoInfoRequest{}
	response = &storage.GetSDRRepoInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSELTime(ctx context.Context, t time.Time) (response *storage.SetSELTimeResponse, err error) {
	request := &storage.SetSELTimeRequest{
		Time: t,
	}
	response = &storage.SetSELTimeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
