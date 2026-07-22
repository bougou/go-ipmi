package bmc

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/hal"
)

// FRUInventory reads FRU inventory blobs from [hal.FRUStore] and implements
// FRU Device semantics for Storage NetFn handlers (v2.0§34).
type FRUInventory struct {
	store hal.FRUStore
}

// NewFRUInventory returns an inventory backed by store.
func NewFRUInventory(store hal.FRUStore) *FRUInventory {
	return &FRUInventory{store: store}
}

// Read returns the full FRU inventory blob for deviceID (v2.0§34.2).
func (f *FRUInventory) Read(ctx context.Context, deviceID uint8) ([]byte, error) {
	return f.store.Read(ctx, deviceID)
}

// AreaSize returns the FRU inventory area size in bytes (v2.0§34.1).
func (f *FRUInventory) AreaSize(ctx context.Context, deviceID uint8) (uint16, error) {
	data, err := f.store.Read(ctx, deviceID)
	if err != nil {
		return 0, err
	}
	return uint16(len(data)), nil
}
