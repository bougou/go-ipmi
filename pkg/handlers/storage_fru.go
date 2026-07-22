package handlers

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/hal"
)

func handleGetFRUInventoryAreaInfo(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	fru := hctx.BMC.FRUInventory()
	if fru == nil {
		return nil, CodeNotSupportedInState, nil
	}

	var typed storage.GetFRUInventoryAreaInfoRequest
	if err := typed.Unpack(req); err != nil {
		return nil, CodeRequestDataTruncated, nil
	}

	size, err := fru.AreaSize(ctx, typed.FRUDeviceID)
	if err != nil {
		if bmc.StorageMissing(err) {
			return nil, CodeRequestedRecordNotPresent, nil
		}
		return nil, codeFromErr(err), err
	}

	resp := &storage.GetFRUInventoryAreaInfoResponse{
		AreaSizeBytes:         size,
		DeviceAccessedByWords: false,
	}
	return resp.Pack(), CodeOK, nil
}

func handleReadFRUData(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	fru := hctx.BMC.FRUInventory()
	if fru == nil {
		return nil, CodeNotSupportedInState, nil
	}

	var typed storage.ReadFRUDataRequest
	if err := typed.Unpack(req); err != nil {
		return nil, CodeRequestDataTruncated, nil
	}

	data, err := fru.Read(ctx, typed.FRUDeviceID)
	if err != nil {
		if bmc.StorageMissing(err) {
			return nil, CodeRequestedRecordNotPresent, nil
		}
		return nil, codeFromErr(err), err
	}

	size := len(data)
	if int(typed.ReadOffset) >= size {
		return nil, CodeRequestedRecordNotPresent, nil
	}

	avail := size - int(typed.ReadOffset)
	count := int(typed.ReadCount)
	if typed.ReadCount == 0 {
		count = 0
	} else if count > avail {
		count = avail
	}

	chunk := data[typed.ReadOffset : typed.ReadOffset+uint16(count)]
	resp := &storage.ReadFRUDataResponse{
		CountReturned: uint8(len(chunk)),
		Data:          chunk,
	}
	return resp.Pack(), CodeOK, nil
}

func storageHAL(hctx *HandlerContext) hal.StorageHAL {
	if hctx == nil || hctx.BMC == nil || hctx.BMC.HAL() == nil {
		return nil
	}
	return hctx.BMC.HAL().Storage()
}

func hasFRUDevice(ctx context.Context, store hal.StorageHAL, deviceID uint8) bool {
	fru := store.FRU()
	if fru == nil {
		return false
	}
	_, err := fru.Read(ctx, deviceID)
	return err == nil
}

func hasSDRRecords(ctx context.Context, store hal.StorageHAL) bool {
	sdr := store.SDR()
	if sdr == nil {
		return false
	}
	ids, err := sdr.RecordIDs(ctx)
	return err == nil && len(ids) > 0
}
