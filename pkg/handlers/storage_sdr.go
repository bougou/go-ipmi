package handlers

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/command/storage"
	"github.com/bougou/go-ipmi/pkg/types"
)

// encodeSDRRepoFreeSpace maps a free-byte count onto the Get SDR Repository
// Info Free Space field (v2.0§33.9): 0000h = full, FFFEh = 64KB-2 or more,
// FFFFh = unspecified.
func encodeSDRRepoFreeSpace(free int) uint16 {
	if free <= 0 {
		return 0
	}
	if free >= 0xFFFE {
		return 0xFFFE
	}
	return uint16(free)
}

func handleGetSDRRepoInfo(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	store := storageHAL(hctx)
	if store == nil || store.SDR() == nil {
		return nil, types.CodeNotSupported, nil
	}
	_ = req

	repo := hctx.BMC.SDRRepository()
	if repo == nil {
		return nil, types.CodeUnspecifiedError, nil
	}
	info, err := repo.Info(ctx)
	if err != nil {
		return nil, codeFromErr(err), err
	}
	resp := &storage.GetSDRRepoInfoResponse{
		SDRVersion:             info.SDRVersion,
		RecordCount:            info.RecordCount,
		FreeSpaceBytes:         encodeSDRRepoFreeSpace(info.FreeBytes),
		MostRecentAdditionTime: info.MostRecentAdd,
		MostRecentEraseTime:    info.MostRecentErase,
		SDROperationSupport: storage.SDROperationSupport{
			Overflow:                     info.Overflow,
			SupportModalSDRRepoUpdate:    info.Capabilities.ModalUpdate,
			SupportNonModalSDRRepoUpdate: info.Capabilities.NonModalUpdate,
			SupportDeleteSDR:             info.Capabilities.DeleteSDR,
			SupportPartialAddSDR:         info.Capabilities.PartialAddSDR,
			SupportReserveSDRRepo:        info.Capabilities.ReserveRepo,
			SupportGetSDRRepoAllocInfo:   info.Capabilities.GetAllocInfo,
		},
	}
	return resp.Pack(), types.CodeOK, nil
}

func handleGetSDRRepoAllocInfo(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	store := storageHAL(hctx)
	if store == nil || store.SDR() == nil {
		return nil, types.CodeNotSupported, nil
	}
	_ = req

	repo := hctx.BMC.SDRRepository()
	if repo == nil {
		return nil, types.CodeUnspecifiedError, nil
	}
	info, err := repo.AllocInfo(ctx)
	if err != nil {
		return nil, codeFromErr(err), err
	}
	resp := &storage.GetSDRRepoAllocInfoResponse{
		PossibleAllocUnits: info.PossibleAllocUnits,
		AllocUnitsSize:     info.AllocUnitSize,
		FreeAllocUnits:     info.FreeAllocUnits,
		LargestFreeBlock:   info.LargestFreeBlock,
		MaximumRecordSize:  info.MaximumRecordSize,
	}
	return resp.Pack(), types.CodeOK, nil
}

func handleReserveSDRRepo(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	store := storageHAL(hctx)
	if store == nil || store.SDR() == nil {
		return nil, types.CodeNotSupported, nil
	}
	_ = req

	if hctx.BMC.SDRRepo == nil {
		return nil, types.CodeUnspecifiedError, nil
	}
	id := hctx.BMC.SDRRepo.Reserve()
	resp := &storage.ReserveSDRRepoResponse{ReservationID: id}
	return resp.Pack(), types.CodeOK, nil
}

func handleGetSDR(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	store := storageHAL(hctx)
	if store == nil || store.SDR() == nil {
		return nil, types.CodeNotSupported, nil
	}

	var typed storage.GetSDRRequest
	if err := typed.Unpack(req); err != nil {
		return nil, types.CodeRequestDataTruncated, nil
	}

	if typed.ReadOffset > 0 {
		if hctx.BMC.SDRRepo == nil || !hctx.BMC.SDRRepo.Validate(typed.ReservationID) {
			return nil, types.CodeReservationCanceled, nil
		}
	}

	repo := hctx.BMC.SDRRepository()
	if repo == nil {
		return nil, types.CodeUnspecifiedError, nil
	}
	record, nextID, err := repo.GetRecord(ctx, typed.RecordID)
	if err != nil {
		if bmc.StorageMissing(err) {
			return nil, types.CodeRequestedDataNotPresent, nil
		}
		return nil, codeFromErr(err), err
	}

	if int(typed.ReadOffset) >= len(record) {
		return nil, types.CodeRequestedDataNotPresent, nil
	}

	want := int(typed.ReadBytes)
	if typed.ReadBytes == 0xff {
		want = len(record) - int(typed.ReadOffset)
	}
	avail := len(record) - int(typed.ReadOffset)
	if want > avail {
		want = avail
	}
	if want > maxSDRReadBytes {
		return nil, types.CodeCannotReturnRequestedDataBytes, nil
	}

	start := int(typed.ReadOffset)
	chunk := record[start : start+want]
	resp := &storage.GetSDRResponse{
		NextRecordID: nextID,
		RecordData:   chunk,
	}
	return resp.Pack(), types.CodeOK, nil
}
