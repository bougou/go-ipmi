package ipmi

import (
	"context"
	"fmt"
)

// GetFRUData return all data bytes, the data size is firstly determined by
// GetFRUInventoryAreaInfoResponse.AreaSizeBytes
func (c *Client) GetFRUData(ctx context.Context, deviceID uint8) ([]byte, error) {
	fruAreaInfoRes, err := c.GetFRUInventoryAreaInfo(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("GetFRUInventoryAreaInfo failed, err: %s", err)
	}

	c.Debug("", fruAreaInfoRes.Format())
	if fruAreaInfoRes.AreaSizeBytes < 1 {
		return nil, fmt.Errorf("invalid FRU size %d", fruAreaInfoRes.AreaSizeBytes)
	}

	data, err := c.readFRUDataByLength(ctx, deviceID, 0, fruAreaInfoRes.AreaSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("read full fru area data failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	return data, nil
}

// GetFRU return FRU for the specified deviceID.
// The deviceName is not a must, pass empty string if not known.
func (c *Client) GetFRU(ctx context.Context, deviceID uint8, deviceName string) (*FRU, error) {
	c.Debugf("GetFRU device name (%s) id (%#02x)\n", deviceName, deviceID)

	fru := &FRU{
		deviceID:   deviceID,
		deviceName: deviceName,
	}

	fruAreaInfoRes, err := c.GetFRUInventoryAreaInfo(ctx, deviceID)
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
	readFRURes, err := c.ReadFRUData(ctx, deviceID, 0, FRUCommonHeaderSize)
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
		return nil, fmt.Errorf("unknown FRU header version %#02x", fruHeader.FormatVersion)
	}
	c.Debug("FRU Common Header", fruHeader)
	c.Debugf("%s\n\n", fruHeader.String())
	fru.CommonHeader = fruHeader

	if offset := uint16(fruHeader.ChassisOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Chassis, offset (%d)\n", offset)
		fruChassis, err := c.GetFRUAreaChassis(ctx, deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaChassis failed, err: %s", err)
		}

		c.Debug("FRU Area Chassis", fruChassis)
		fru.ChassisInfoArea = fruChassis
	}

	if offset := uint16(fruHeader.BoardOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Board, offset (%d)\n", offset)
		fruBoard, err := c.GetFRUAreaBoard(ctx, deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaBoard failed, err: %s", err)
		}
		c.Debug("FRU Area Board", fruBoard)
		fru.BoardInfoArea = fruBoard
	}

	if offset := uint16(fruHeader.ProductOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Product, offset (%d)\n", offset)
		fruProduct, err := c.GetFRUAreaProduct(ctx, deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaProduct failed, err: %s", err)
		}
		c.Debug("FRU Area Product", fruProduct)
		fru.ProductInfoArea = fruProduct
	}

	if offset := uint16(fruHeader.MultiRecordsOffset8B) * 8; offset > 0 && offset < fruAreaInfoRes.AreaSizeBytes {
		c.Debugf("Get FRU Area Multi Records, offset (%d)\n", offset)
		fruMultiRecords, err := c.GetFRUAreaMultiRecords(ctx, deviceID, offset)
		if err != nil {
			return nil, fmt.Errorf("GetFRUAreaMultiRecord failed, err: %s", err)
		}
		c.Debug("FRU Area MultiRecords", fruMultiRecords)
		fru.MultiRecords = fruMultiRecords
	}

	c.Debug("FRU", fru)
	return fru, nil
}

func (c *Client) GetFRUs(ctx context.Context) ([]*FRU, error) {
	var frus = make([]*FRU, 0)

	// Do a Get Device ID command to determine device support
	deviceRes, err := c.GetDeviceID(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetDeviceID failed, err: %s", err)
	}

	c.Debug("deviceRes", deviceRes)

	if deviceRes.AdditionalDeviceSupport.SupportFRUInventory {
		// FRU Device ID #00 at LUN 00b is predefined as being the FRU Device
		// for the FRU that the management controller is located on.
		var deviceID uint8 = 0x00
		fru, err := c.GetFRU(ctx, deviceID, "Builtin FRU")
		if err != nil {
			return nil, fmt.Errorf("GetFRU device id (%#02x) failed, err: %s", deviceID, err)
		}
		frus = append(frus, fru)
	}

	// Walk the SDRs to look for FRU Devices and Management Controller Devices.
	// For FRU devices, print the FRU from the SDR locator record.
	// For MC devices, issue FRU commands to the satellite controller to print FRU data.
	sdrs, err := c.GetSDRs(ctx, SDRRecordTypeFRUDeviceLocator, SDRRecordTypeManagementControllerDeviceLocator)
	if err != nil {
		return nil, fmt.Errorf("GetSDRS failed, err: %s", err)
	}

	for _, sdr := range sdrs {
		switch sdr.RecordHeader.RecordType {

		case SDRRecordTypeFRUDeviceLocator:

			deviceType := sdr.FRUDeviceLocator.DeviceType
			deviceTypeModifier := sdr.FRUDeviceLocator.DeviceTypeModifier

			deviceName := string(sdr.FRUDeviceLocator.DeviceIDBytes)
			deviceAccessAddress := sdr.FRUDeviceLocator.DeviceAccessAddress         // controller
			accessLUN := sdr.FRUDeviceLocator.AccessLUN                             // LUN
			privateBusID := sdr.FRUDeviceLocator.PrivateBusID                       // Private bus
			deviceIDOrSlaveAddress := sdr.FRUDeviceLocator.FRUDeviceID_SlaveAddress // device

			fruLocation := sdr.FRUDeviceLocator.Location()

			c.Debugf("fruLocation: (%s), deviceType: (%s [%#02x]), deviceTypeModifier: (%#02x), deviceIDOrSlaveAddress: (%#02x), deviceName: (%s), isLogical: (%v), "+
				"DeviceAccessAddress (%#02x), AccessLUN: (%#02x), PrivateBusID(%#02x)\n",
				fruLocation, deviceType.String(), uint8(deviceType), deviceTypeModifier, deviceIDOrSlaveAddress, deviceName, sdr.FRUDeviceLocator.IsLogicalFRUDevice,
				deviceAccessAddress, accessLUN, privateBusID,
			)

			// see 38. Accessing FRU Devices
			switch fruLocation {
			case FRULocation_MgmtController:
				if accessLUN == 0x00 && deviceIDOrSlaveAddress == 0x00 {
					// this is the Builtin FRU device, already got
					continue
				}

				// Todo, accessed using Read/Write FRU commands at LUN other than 00b
				fru, err := c.GetFRU(ctx, deviceIDOrSlaveAddress, deviceName)
				if err != nil {
					return nil, fmt.Errorf("GetFRU sdr device id (%#02x) failed, err: %s", deviceIDOrSlaveAddress, err)
				}
				frus = append(frus, fru)

			case FRULocation_PrivateBus:
				// Todo,
				switch deviceType {
				case 0x10:
					// Todo, refactor BuildIPMIRequest to use LUN
					// if sdr.FRUDeviceLocator.DeviceAccessAddress == BMC_SA && deviceID == 0x00 {
					// 	continue
					// }

					switch deviceTypeModifier {
					// 0x00, 0x02 = IPMI FRU Inventory
					case 0x00, 0x02:

					// 0x01 = DIMM Memory ID
					case 0x01:

					// 03h = System Processor Cartridge FRU / PIROM (processor information ROM)
					case 0x03:

					}

				case 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f:
					// Todo
				}

			case FRULocation_IPMB:

			}

		case SDRRecordTypeManagementControllerDeviceLocator:

		}
	}

	return frus, nil
}

func (c *Client) GetFRUAreaChassis(ctx context.Context, deviceID uint8, offset uint16) (*FRUChassisInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(ctx, deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(ctx, deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("read full fru area data failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruChassis := &FRUChassisInfoArea{}
	if err := fruChassis.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru chassis failed, err: %s", err)
	}

	return fruChassis, nil
}

func (c *Client) GetFRUAreaBoard(ctx context.Context, deviceID uint8, offset uint16) (*FRUBoardInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(ctx, deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(ctx, deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("read full fru area data failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruBoard := &FRUBoardInfoArea{}
	if err := fruBoard.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru board failed, err: %s", err)
	}

	return fruBoard, nil
}

func (c *Client) GetFRUAreaProduct(ctx context.Context, deviceID uint8, offset uint16) (*FRUProductInfoArea, error) {
	// read enough (2 bytes) to check the length field
	res, err := c.ReadFRUData(ctx, deviceID, offset, 2)
	if err != nil {
		return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
	}
	length := uint16(res.Data[1]) * 8 // in multiples of 8 bytes

	// now read full area data
	data, err := c.readFRUDataByLength(ctx, deviceID, offset, length)
	if err != nil {
		return nil, fmt.Errorf("read full fru area data failed, err: %s", err)
	}
	c.Debugf("Got %d fru data\n", len(data))

	fruProduct := &FRUProductInfoArea{}
	if err := fruProduct.Unpack(data); err != nil {
		return nil, fmt.Errorf("unpack fru board failed, err: %s", err)
	}

	return fruProduct, nil
}

func (c *Client) GetFRUAreaMultiRecords(ctx context.Context, deviceID uint8, offset uint16) ([]*FRUMultiRecord, error) {
	records := make([]*FRUMultiRecord, 0)

	for {
		// read enough (5 bytes) to check the length of each record
		// For a MultiRecord, the first 5 bytes contains the Record Header,
		// and the third byte holds the data length.
		//
		// see: FRU/16.1 Record Header
		res, err := c.ReadFRUData(ctx, deviceID, offset, 5)
		if err != nil {
			return nil, fmt.Errorf("ReadFRUData failed, err: %s", err)
		}
		length := uint16(res.Data[2])

		// now read full data for this record
		recordSize := 5 + length // Record Header + Data Length
		data, err := c.readFRUDataByLength(ctx, deviceID, offset, recordSize)
		if err != nil {
			return nil, fmt.Errorf("read full fru area data failed, err: %s", err)
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
