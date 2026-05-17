package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
)

func newTestBMC() *bmc.BMC {
	info := bmc.DeviceInfo{
		DeviceID:                1,
		DeviceRevision:          1,
		FirmwareMajor:           2,
		FirmwareMinor:           0x00,
		IPMIVersion:             0x20,
		ManufacturerID:          0x000157, // PICMG
		ProductID:               0x0001,
		AdditionalDeviceSupport: 0x3D,
	}
	var guid [16]byte
	return bmc.New(info, guid, mock.New(), bmc.WithClock(clock.Real))
}

func TestHandleGetDeviceID(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	resp, cc, err := handleGetDeviceID(context.Background(), hctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if len(resp) < 11 {
		t.Fatalf("response too short: %d bytes", len(resp))
	}

	tests := []struct {
		name string
		idx  int
		want byte
	}{
		{"DeviceID", 0, 1},
		{"IPMIVersion", 4, 0x20},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if resp[tc.idx] != tc.want {
				t.Errorf("byte[%d]: want 0x%02x, got 0x%02x", tc.idx, tc.want, resp[tc.idx])
			}
		})
	}
}

func TestHandleGetDeviceGUID(t *testing.T) {
	b := newTestBMC()
	b.GUID[0] = 0xDE
	b.GUID[15] = 0xAD

	hctx := &HandlerContext{BMC: b}
	resp, cc, _ := handleGetDeviceGUID(context.Background(), hctx, nil)

	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if len(resp) != 16 {
		t.Fatalf("want 16 bytes, got %d", len(resp))
	}
	if resp[0] != 0xDE || resp[15] != 0xAD {
		t.Errorf("GUID bytes not copied correctly")
	}
}

func TestHandleGetSelfTestResults(t *testing.T) {
	resp, cc, _ := handleGetSelfTestResults(context.Background(), nil, nil)
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if len(resp) < 2 || resp[0] != 0x55 {
		t.Errorf("unexpected self-test response: %v", resp)
	}
}

func TestHandleColdReset(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	_, cc, _ := handleColdReset(context.Background(), hctx, nil)
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}

	chassis := b.HAL().Chassis().(*mock.Chassis)
	if chassis.ColdResets != 1 {
		t.Errorf("ColdResets: want 1, got %d", chassis.ColdResets)
	}
}
