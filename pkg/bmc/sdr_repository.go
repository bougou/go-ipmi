package bmc

import (
	"context"
	"time"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// defaultSDRRepoSize is the synthetic repository capacity used for Info/AllocInfo.
// v2.0§33.9 encodes free space as uint16 where FFFEh means "64KB-2 or more";
// handlers clamp values larger than that when packing the wire response.
const defaultSDRRepoSize = 64 * 1024

// SDRCapabilities describes which SDR repository operations this BMC supports.
// Handlers map these flags onto storage.SDROperationSupport (v2.0§33.9).
type SDRCapabilities struct {
	ModalUpdate    bool
	NonModalUpdate bool
	DeleteSDR      bool
	PartialAddSDR  bool
	ReserveRepo    bool
	GetAllocInfo   bool
}

// SDRRepoInfo is BMC-side repository status (not a wire response).
// Handlers map this to storage.GetSDRRepoInfoResponse (v2.0§33.9).
type SDRRepoInfo struct {
	SDRVersion      uint8
	RecordCount     uint16
	FreeBytes       int // raw free capacity; §33.9 wire encoding is the handler's job
	MostRecentAdd   time.Time
	MostRecentErase time.Time
	Overflow        bool
	Capabilities    SDRCapabilities
}

// SDRRepoAllocInfo is BMC-side allocation accounting (not a wire response).
// Handlers map this to storage.GetSDRRepoAllocInfoResponse (v2.0§33.10).
type SDRRepoAllocInfo struct {
	PossibleAllocUnits uint16
	AllocUnitSize      uint16
	FreeAllocUnits     uint16
	LargestFreeBlock   uint16
	MaximumRecordSize  uint8
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

// Info returns BMC-side SDR repository status per v2.0§33.9 semantics.
func (r *SDRRepository) Info(ctx context.Context) (*SDRRepoInfo, error) {
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
	return &SDRRepoInfo{
		SDRVersion:      types.SDRCommandSetVersion,
		RecordCount:     uint16(len(ids)),
		FreeBytes:       free,
		MostRecentAdd:   now,
		MostRecentErase: now,
		Capabilities: SDRCapabilities{
			ReserveRepo:  true,
			GetAllocInfo: true,
		},
	}, nil
}

// AllocInfo returns BMC-side allocation accounting per v2.0§33.10 semantics.
func (r *SDRRepository) AllocInfo(ctx context.Context) (*SDRRepoAllocInfo, error) {
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
	return &SDRRepoAllocInfo{
		PossibleAllocUnits: uint16(totalUnits),
		AllocUnitSize:      unitSize,
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
