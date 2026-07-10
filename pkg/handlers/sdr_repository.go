package handlers

import (
	"context"
	"errors"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

const defaultSDRRepoSize = 64 * 1024

// SDRRepository reads SDR records from [hal.SDRStore] and implements
// repository semantics for Storage NetFn handlers.
type SDRRepository struct {
	store hal.SDRStore
	clk   clock.Clock
}

func newSDRRepository(store hal.SDRStore, clk clock.Clock) *SDRRepository {
	if clk == nil {
		clk = clock.Real
	}
	return &SDRRepository{store: store, clk: clk}
}

func (r *SDRRepository) RecordIDs(ctx context.Context) ([]uint16, error) {
	return r.store.RecordIDs(ctx)
}

func (r *SDRRepository) usedBytes(ctx context.Context) (int, error) {
	ids, err := r.RecordIDs(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, id := range ids {
		rec, err := r.store.Read(ctx, id)
		if err != nil {
			return 0, err
		}
		total += len(rec)
	}
	return total, nil
}

// GetRecord returns the wire record and next Record ID for repository traversal.
// recordID 0 maps to the lowest stored ID per §33.12.
func (r *SDRRepository) GetRecord(ctx context.Context, recordID uint16) (record []byte, nextID uint16, err error) {
	ids, err := r.RecordIDs(ctx)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return nil, 0, hal.ErrNotSupported
	}

	idx := -1
	if recordID == 0 {
		idx = 0
	} else {
		for i, id := range ids {
			if id == recordID {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		return nil, 0, hal.ErrNotSupported
	}

	id := ids[idx]
	record, err = r.store.Read(ctx, id)
	if err != nil {
		return nil, 0, err
	}
	if idx+1 < len(ids) {
		nextID = ids[idx+1]
	} else {
		nextID = 0xffff
	}
	return record, nextID, nil
}

func (r *SDRRepository) Info(ctx context.Context) (*storage.GetSDRRepoInfoResponse, error) {
	ids, err := r.RecordIDs(ctx)
	if err != nil {
		return nil, err
	}
	used, err := r.usedBytes(ctx)
	if err != nil {
		return nil, err
	}
	free := defaultSDRRepoSize - used
	if free < 0 {
		free = 0
	}
	now := r.clk.Now()
	return &storage.GetSDRRepoInfoResponse{
		SDRVersion:             types.SDRCommandSetVersion,
		RecordCount:            uint16(len(ids)),
		FreeSpaceBytes:         uint16(free),
		MostRecentAdditionTime: now,
		MostRecentEraseTime:    now,
		SDROperationSupport: storage.SDROperationSupport{
			SupportReserveSDRRepo:      true,
			SupportGetSDRRepoAllocInfo: true,
		},
	}, nil
}

func (r *SDRRepository) AllocInfo(ctx context.Context) (*storage.GetSDRRepoAllocInfoResponse, error) {
	used, err := r.usedBytes(ctx)
	if err != nil {
		return nil, err
	}
	unitSize := uint16(16)
	totalUnits := defaultSDRRepoSize / int(unitSize)
	usedUnits := (used + int(unitSize) - 1) / int(unitSize)
	freeUnits := totalUnits - usedUnits
	if freeUnits < 0 {
		freeUnits = 0
	}
	return &storage.GetSDRRepoAllocInfoResponse{
		PossibleAllocUnits: uint16(totalUnits),
		AllocUnitsSize:     unitSize,
		FreeAllocUnits:     uint16(freeUnits),
		LargestFreeBlock:   uint16(freeUnits),
		MaximumRecordSize:  64,
	}, nil
}

func isStorageMissing(err error) bool {
	return errors.Is(err, hal.ErrNotSupported)
}
