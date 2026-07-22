package hal

import "context"

// StorageHAL groups persistent blob stores for Storage NetFn data (§33–34).
// Each sub-store may be nil when the backing hardware is absent.
type StorageHAL interface {
	FRU() FRUStore
	SDR() SDRStore
}

// FRUStore holds wire-format FRU inventory blobs (§34).
// DeviceID 0 is the builtin MC FRU at LUN 00b.
type FRUStore interface {
	Read(ctx context.Context, deviceID uint8) ([]byte, error)
	Write(ctx context.Context, deviceID uint8, data []byte) error
	Delete(ctx context.Context, deviceID uint8) error
	DeviceIDs(ctx context.Context) ([]uint8, error)
}

// SDRStore holds wire-format SDR repository records (v2.0§33).
// RecordID 0 is not a valid stored ID; [bmc.SDRRepository.GetRecord] maps
// Get SDR(0000h) to the first record and Get SDR(FFFFh) to the last (v2.0§33.12).
type SDRStore interface {
	Read(ctx context.Context, recordID uint16) ([]byte, error)
	Write(ctx context.Context, recordID uint16, data []byte) error
	Delete(ctx context.Context, recordID uint16) error
	RecordIDs(ctx context.Context) ([]uint16, error)
}
