package bmc

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

const defaultSDRRepoSize = 64 * 1024

// SDRRepository reads SDR records from [hal.SDRStore] and implements
// repository semantics for Storage NetFn handlers (§33).
type SDRRepository struct {
	store hal.SDRStore
	clk   clock.Clock
}

// NewSDRRepository returns a repository backed by store.
func NewSDRRepository(store hal.SDRStore, clk clock.Clock) *SDRRepository {
	if clk == nil {
		clk = clock.Real
	}
	return &SDRRepository{store: store, clk: clk}
}

// RecordIDs returns the sorted list of stored record IDs.
func (r *SDRRepository) RecordIDs(ctx context.Context) ([]uint16, error) {
	return r.store.RecordIDs(ctx)
}

func (r *SDRRepository) usedBytes(ctx context.Context) (int, error) {
	total, _, err := r.scanRecords(ctx)
	return total, err
}

// scanRecords returns the total size across all records and the largest single record size.
func (r *SDRRepository) scanRecords(ctx context.Context) (total, maxRec int, err error) {
	ids, err := r.RecordIDs(ctx)
	if err != nil {
		return 0, 0, err
	}
	for _, id := range ids {
		rec, err := r.store.Read(ctx, id)
		if err != nil {
			return 0, 0, err
		}
		total += len(rec)
		if len(rec) > maxRec {
			maxRec = len(rec)
		}
	}
	return total, maxRec, nil
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

// Info returns the SDR Repository Info response per §33.9.
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

// AllocInfo returns the SDR Repository Alloc Info response per §33.10.
func (r *SDRRepository) AllocInfo(ctx context.Context) (*storage.GetSDRRepoAllocInfoResponse, error) {
	used, maxRec, err := r.scanRecords(ctx)
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
	if maxRec > 255 {
		maxRec = 255
	}
	maxRecSize := uint8(maxRec)
	if maxRecSize < 64 {
		maxRecSize = 64
	}
	return &storage.GetSDRRepoAllocInfoResponse{
		PossibleAllocUnits: uint16(totalUnits),
		AllocUnitsSize:     unitSize,
		FreeAllocUnits:     uint16(freeUnits),
		LargestFreeBlock:   uint16(freeUnits),
		MaximumRecordSize:  maxRecSize,
	}, nil
}

// StorageMissing reports whether err indicates the backing store is absent.
func StorageMissing(err error) bool {
	return err == hal.ErrNotSupported
}
