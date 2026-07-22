package bmc

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// defaultSDRRepoSize is the synthetic repository capacity used for Info/AllocInfo.
// v2.0§33.9 encodes free space as uint16 where FFFEh means "64KB-2 or more";
// values larger than that must be clamped (see encodeSDRRepoFreeSpace).
const defaultSDRRepoSize = 64 * 1024

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

// SDRRepository reads SDR records from [hal.SDRStore] and implements
// repository semantics for Storage NetFn handlers (v2.0§33).
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
// Per v2.0§33.12: recordID 0000h maps to the first SDR; FFFFh maps to the last.
func (r *SDRRepository) GetRecord(ctx context.Context, recordID uint16) (record []byte, nextID uint16, err error) {
	ids, err := r.RecordIDs(ctx)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return nil, 0, hal.ErrNotFound
	}

	idx := -1
	switch recordID {
	case 0:
		idx = 0
	case 0xffff:
		idx = len(ids) - 1
	default:
		for i, id := range ids {
			if id == recordID {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		return nil, 0, hal.ErrNotFound
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

// Info returns the Get SDR Repository Info response per v2.0§33.9.
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
	now := r.clk.Now()
	return &storage.GetSDRRepoInfoResponse{
		SDRVersion:             types.SDRCommandSetVersion,
		RecordCount:            uint16(len(ids)),
		FreeSpaceBytes:         encodeSDRRepoFreeSpace(free),
		MostRecentAdditionTime: now,
		MostRecentEraseTime:    now,
		SDROperationSupport: storage.SDROperationSupport{
			SupportReserveSDRRepo:      true,
			SupportGetSDRRepoAllocInfo: true,
		},
	}, nil
}

// AllocInfo returns the Get SDR Repository Allocation Info response per v2.0§33.10.
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

// StorageMissing reports whether err indicates a missing FRU device or SDR record
// (mapped to completion code CBh by storage handlers).
func StorageMissing(err error) bool {
	return err == hal.ErrNotFound
}
