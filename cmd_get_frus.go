package ipmi

import (
	"fmt"
)

// GetFRUData return all data bytes, the data size is firstly determined by
// GetFRUInventoryAreaInfoResponse.AreaSizeBytes
func (c *Client) GetFRUData(deviceID uint8) ([]byte, error) {
	fruAreaInfoRes, err := c.GetFRUInventoryAreaInfo(deviceID)
	if err != nil {
		return nil, fmt.Errorf("GetFRUInventoryAreaInfo failed, err: %s", err)
	}

	c.Debug("", fruAreaInfoRes.Format())
	if fruAreaInfoRes.AreaSizeBytes < 1 {
		return nil, fmt.Errorf("invalid FRU size %d", fruAreaInfoRes.AreaSizeBytes)
	}

	data, err := c.readFRUDataByLength(deviceID, 0, fruAreaInfoRes.AreaSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUDataAll failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	return data, nil
}

// GetFRU return FRU for the specified deviceID.
// The deviceName is not a must, pass empty string if not known.
func (c *Client) GetFRU(deviceID uint8, deviceName string) (*FRU, error) {
	c.Debugf("GetFRU device name (%s) id (%#02x)\n", deviceName, deviceID)

	fru := &FRU{
		deviceID:   deviceID,
		deviceName: deviceName,
	}

	fruAreaInfoRes, err := c.GetFRUInventoryAreaInfo(deviceID)
	if err != nil {
		if resErr, ok := err.(*ResponseError); ok {
			if resErr.CompletionCode() == CompletionCodeRequestedDataNotPresent {
				fru.deviceNotPresent = true
				fru.deviceNotPresentReason = "InventoryRecordNotExist"
				return fru, nil
			}
		}
		return nil, fmt.Errorf("GetFRUInventoryAreaInfo failed, err: %s", err)
	}

	c.Debug("", fruAreaInfoRes.Format())
	if fruAreaInfoRes.AreaSizeBytes < 1 {
		return nil, fmt.Errorf("invalid FRU size %d", fruAreaInfoRes.AreaSizeBytes)
	}

	// retrieve the FRU header, just fetch FRUCommonHeaderSize bytes to construct a FRU Header
	readFRURes, err := c.ReadFRUData(deviceID, 0, FRUCommonHeaderSize)
	if err != nil {
		if resErr, ok := err.(*ResponseError); ok {
			switch resErr.CompletionCode() {
			case CompletionCodeRequestedDataNotPresent:
				fru.deviceNotPresent = true
				fru.deviceNotPresentReason = "DataNotPresent"
				return fru, nil
			case CompletionCodeProcessTimeout:
				fru.deviceNotPresent = true
				fru.deviceNotPresentReason = "Timeout"
				return fru, nil
			}
		}
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}

	fruHeader := &FRUCommonHeader{}
	if err := fruHeader.Unpack(readFRURes.Data); err != nil {
		return nil, fmt.Errorf("unpack fru data failed, err: %s", err)
	}
	if fruHeader.FormatVersion != FRUFormatVersion {
		return nil, fmt.Errorf("unkown FRU header version %#02x", fruHeader.FormatVersion)
	}
	c.Debug("FRU Common Header", fruHeader)
	c.Debugf("%s\n\n", fruHeader.String())
	fru.CommonHeader = fruHeader

	if offset := uint16(fruHeader.ChassisOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Chassis, offset (%d)\n", offset)
		fruChassis, err := c.GetFRUAreaChassis(deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaChassis failed, err: %s", err)
		}

		c.Debug("FRU Area Chassis", fruChassis)
		fru.ChassisInfoArea = fruChassis
	}

	if offset := uint16(fruHeader.BoardOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Board, offset (%d)\n", offset)
		fruBoard, err := c.GetFRUAreaBoard(deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaBoard failed, err: %s", err)
		}
		c.Debug("FRU Area Board", fruBoard)
		fru.BoardInfoArea = fruBoard
	}

	if offset := uint16(fruHeader.ProductOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Product, offset (%d)\n", offset)
		fruProduct, err := c.GetFRUAreaProduct(deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaProduct failed, err: %s", err)
		}
		c.Debug("FRU Area Product", fruProduct)
		fru.ProductInfoArea = fruProduct
	}

	if offset := uint16(fruHeader.MultiRecordsOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Multi Records, offset (%d)\n", offset)
		fruMultiRecords, err := c.GetFRUAreaMultiRecords(deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaMultiRecord failed, err: %s", err)
		}
		c.Debug("FRU Area MultiRecords", fruMultiRecords)
		fru.MultiRecords = fruMultiRecords
	}

	c.Debug("FRU", fru)
	return fru, nil
}

func (c *Client) GetFRUs() ([]*FRU, error) {
	var frus = make([]*FRU, 0)

	// Do a Get Device ID command to determine device support
	deviceRes, err := c.GetDeviceID()
	if err != nil {
		return nil, fmt.Errorf("GetDeviceID failed, err: %s", err)
	}

	if deviceRes.AdditionalDeviceSupport.SupportFRUInventory {
		// FRU Device ID #00 at LUN 00b is predefined as being the FRU Device for the FRU that the management controller is located on.
		var deviceID uint8 = 0x00
		fru, err := c.GetFRU(deviceID, "Builtin FRU")
		if err != nil {
			return nil, fmt.Errorf("GetFRU device id (%#02x) failed, err: %s", deviceID, err)
		}
		frus = append(frus, fru)
	}

	// Walk the SDRs to look for FRU Devices and Management Controller Devices.
	// For FRU devices, print the FRU from the SDR locator record.
	// For MC devices, issue FRU commands to the satellite controller to print FRU data.
	sdrs, err := c.GetSDRs(SDRRecordTypeFRUDeviceLocator, SDRRecordTypeManagementControllerDeviceLocator)
	if err != nil {
		return nil, fmt.Errorf("GetSDRS failed, err: %s", err)
	}

	for _, sdr := range sdrs {
		switch sdr.RecordHeader.RecordType {
		case SDRRecordTypeFRUDeviceLocator:
			if !sdr.FRUDeviceLocator.IsLogicalFRUDevice {
				// only logical FRU Device can be accessed via FRU commands to mgmt controller
				continue
			}

			deviceType := sdr.FRUDeviceLocator.DeviceType
			deviceTypeModifier := sdr.FRUDeviceLocator.DeviceTypeModifier
			deviceID := sdr.FRUDeviceLocator.FRUDeviceID_SlaveAddress
			deviceName := string(sdr.FRUDeviceLocator.DeviceIDBytes)

			if deviceType != 0x10 && (deviceType < 0x08 || deviceType > 0x0f || deviceTypeModifier != 0x02) {
				// ignore
				continue
			}

			if sdr.FRUDeviceLocator.DeviceAccessAddress == BMC_SA && deviceID == 0x00 {
				continue
			}

			switch deviceTypeModifier {
			case 0x00, 0x02:
				fru, err := c.GetFRU(deviceID, deviceName)
				if err != nil {
					return nil, fmt.Errorf("GetFRU sdr device id (%#02x) failed, err: %s", deviceID, err)
				}
				frus = append(frus, fru)

			case 0x01:
				// *   0x01 = DIMM Memory ID
				fruData, err := c.GetFRUData(deviceID)
				if err != nil {
					return nil, fmt.Errorf("GetFRUData failed, err: %s", err)
				}
				c.DebugBytes("FRU Data", fruData, 16)
				// Todo, parse SPD

			default:
			}

		case SDRRecordTypeManagementControllerDeviceLocator:

		}
	}

	return frus, nil
}

func (c *Client) GetFRUAreaChassis(deviceID uint8, offset uint16) (*FRUChassisInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUDataAll failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruChassis := &FRUChassisInfoArea{}
	if err := fruChassis.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru chassis failed, err: %s", err)
	}

	return fruChassis, nil
}

func (c *Client) GetFRUAreaBoard(deviceID uint8, offset uint16) (*FRUBoardInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUDataAll failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruBoard := &FRUBoardInfoArea{}
	if err := fruBoard.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru board failed, err: %s", err)
	}

	return fruBoard, nil
}

func (c *Client) GetFRUAreaProduct(deviceID uint8, offset uint16) (*FRUProductInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUDataAll failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruProduct := &FRUProductInfoArea{}
	if err := fruProduct.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru board failed, err: %s", err)
	}

	return fruProduct, nil
}

func (c *Client) GetFRUAreaMultiRecords(deviceID uint8, offset uint16) ([]*FRUMultiRecord, error) {
	records := make([]*FRUMultiRecord, 0)

	for {
		// read enough (5 bytes) to check the length of each record
		// For a MultiRecord, the first 5 bytes contains the Record Header,
		// and the third byte holds the data length.
		//
		// see: FRU/16.1 Record Header
		res, err := c.ReadFRUData(deviceID, offset, 5)
		if err != nil {
			return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
		}
		length := uint16(res.Data[2])

		// now read full data for this record
		recordSize := 5 + length // Record Header + Data Length
		data, err := c.readFRUDataByLength(deviceID, offset, recordSize)
		if err != nil {
			return nil, fmt.Errorf("ReadFRUDataAll failed, err: %s", err)
		}
		c.Debugf("Got %d fru data\n", len(data))

		record := &FRUMultiRecord{}
		if err := record.Unpack(data); err != nil {
			return nil, fmt.Errorf("unpack fru multi record failed, err: %s", err)
		}
		c.Debug("Multi record", record)
		records = append(records, record)

		// update offset for the next record
		offset += uint16(5 + record.RecordLength)

		if record.EndOfList {
			break
		}
	}

	return records, nil
}
