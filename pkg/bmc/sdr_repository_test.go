package bmc

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestSDRRepositoryInfo_EmptyRepoFreeBytes(t *testing.T) {
	// BMC reports raw free capacity; §33.9 wire clamping is done by handlers.
	repo := NewSDRRepository(mock.New().Storage().SDR(), clock.Real)
	info, err := repo.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.RecordCount != 0 {
		t.Fatalf("RecordCount: want 0 got %d", info.RecordCount)
	}
	if info.FreeBytes != defaultSDRRepoSize {
		t.Fatalf("FreeBytes: want %d got %d", defaultSDRRepoSize, info.FreeBytes)
	}
	if info.SDRVersion != types.SDRCommandSetVersion {
		t.Fatalf("SDRVersion: want %#02x got %#02x", types.SDRCommandSetVersion, info.SDRVersion)
	}
	if !info.Capabilities.ReserveRepo || !info.Capabilities.GetAllocInfo {
		t.Fatalf("unexpected capabilities: %+v", info.Capabilities)
	}
}
