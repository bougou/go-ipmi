package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/cmd/storage"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

func newTestBMCWithStorage(t *testing.T) (*bmc.BMC, *mock.HAL) {
	t.Helper()
	m := mock.New()
	info := bmc.DeviceInfo{
		DeviceID:       0x20,
		IPMIVersion:    0x20,
		ManufacturerID: 0x000157,
		ProductID:      0x0001,
	}
	var guid [16]byte
	return bmc.New(info, guid, m, bmc.WithClock(clock.Real)), m
}

func testFRUBytes() []byte {
	return types.PackFRU(types.FRUPackConfig{
		Product: &types.FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "TestBMC",
			Version:      "1.0",
			Serial:       "SN-001",
		},
	})
}

func TestHandleGetFRUInventoryAreaInfo(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	fru := testFRUBytes()
	_ = m.Storage().FRU().Write(context.Background(), 0, fru)

	hctx := &HandlerContext{BMC: b}
	resp, cc, err := handleGetFRUInventoryAreaInfo(context.Background(), hctx, []byte{0x00})
	if err != nil || cc != CodeOK {
		t.Fatalf("cc=%v err=%v", cc, err)
	}
	var decoded storage.GetFRUInventoryAreaInfoResponse
	if err := decoded.Unpack(resp); err != nil {
		t.Fatal(err)
	}
	if decoded.AreaSizeBytes != uint16(len(fru)) {
		t.Fatalf("size: want %d got %d", len(fru), decoded.AreaSizeBytes)
	}
}

func TestHandleReadFRUData_RoundTrip(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	fru := testFRUBytes()
	_ = m.Storage().FRU().Write(context.Background(), 0, fru)

	req := (&storage.ReadFRUDataRequest{FRUDeviceID: 0, ReadOffset: 0, ReadCount: 32}).Pack()
	hctx := &HandlerContext{BMC: b}
	resp, cc, err := handleReadFRUData(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("cc=%v err=%v", cc, err)
	}
	var decoded storage.ReadFRUDataResponse
	if err := decoded.Unpack(resp); err != nil {
		t.Fatal(err)
	}
	if decoded.CountReturned != 32 || len(decoded.Data) != 32 {
		t.Fatalf("unexpected read: %+v", decoded)
	}
}

func TestHandleReadFRUData_MissingDevice(t *testing.T) {
	b, _ := newTestBMCWithStorage(t)
	req := (&storage.ReadFRUDataRequest{FRUDeviceID: 1, ReadOffset: 0, ReadCount: 8}).Pack()
	hctx := &HandlerContext{BMC: b}
	_, cc, err := handleReadFRUData(context.Background(), hctx, req)
	if err != nil || cc != CodeRequestedRecordNotPresent {
		t.Fatalf("cc=%v err=%v", cc, err)
	}
}

func TestHandleGetSDR_Traverse(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	body := make([]byte, 32)
	for i := range body {
		body[i] = byte(i)
	}
	rec1 := append([]byte{0x01, 0x00, types.SDRCommandSetVersion, 0x02, byte(len(body))}, body...)
	rec2 := []byte{0x02, 0x00, types.SDRCommandSetVersion, 0x02, 0x02, 0x11, 0x22, 0x33}
	_ = m.Storage().SDR().Write(context.Background(), 1, rec1)
	_ = m.Storage().SDR().Write(context.Background(), 2, rec2)

	hctx := &HandlerContext{BMC: b}
	req := (&storage.GetSDRRequest{RecordID: 0, ReadOffset: 0, ReadBytes: 0xff}).Pack()
	_, cc, err := handleGetSDR(context.Background(), hctx, req)
	if err != nil || cc != CodeCannotReturnRequestedDataBytes {
		t.Fatalf("full read should trigger CAh: cc=%v err=%v", cc, err)
	}

	req = (&storage.GetSDRRequest{RecordID: 0, ReadOffset: 0, ReadBytes: 16}).Pack()
	resp, cc, err := handleGetSDR(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("partial read: cc=%v err=%v", cc, err)
	}
	var decoded storage.GetSDRResponse
	if err := decoded.Unpack(resp); err != nil {
		t.Fatal(err)
	}
	if decoded.NextRecordID != 2 {
		t.Fatalf("NextRecordID: want 2 got %d", decoded.NextRecordID)
	}
}

func TestHandleGetSDR_ReservationRequired(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	rec := []byte{0x01, 0x00, types.SDRCommandSetVersion, 0x02, 0x10, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	_ = m.Storage().SDR().Write(context.Background(), 1, rec)

	hctx := &HandlerContext{BMC: b}
	req := (&storage.GetSDRRequest{RecordID: 1, ReadOffset: 16, ReadBytes: 4}).Pack()
	_, cc, err := handleGetSDR(context.Background(), hctx, req)
	if err != nil || cc != CodeReservationCanceled {
		t.Fatalf("want C5h without reservation: cc=%v err=%v", cc, err)
	}

	reserveResp, cc, err := handleReserveSDRRepo(context.Background(), hctx, nil)
	if err != nil || cc != CodeOK {
		t.Fatalf("reserve: cc=%v err=%v", cc, err)
	}
	var reserve storage.ReserveSDRRepoResponse
	if err := reserve.Unpack(reserveResp); err != nil {
		t.Fatal(err)
	}

	req = (&storage.GetSDRRequest{
		ReservationID: reserve.ReservationID,
		RecordID:      1,
		ReadOffset:    16,
		ReadBytes:     4,
	}).Pack()
	resp, cc, err := handleGetSDR(context.Background(), hctx, req)
	if err != nil || cc != CodeOK {
		t.Fatalf("reserved partial: cc=%v err=%v", cc, err)
	}
	var decoded storage.GetSDRResponse
	if err := decoded.Unpack(resp); err != nil {
		t.Fatal(err)
	}
	if len(decoded.RecordData) != 4 {
		t.Fatalf("want 4 bytes, got %d", len(decoded.RecordData))
	}
}

func TestHandleGetSDRRepoInfo(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	_ = m.Storage().SDR().Write(context.Background(), 1, []byte{0x01, 0x00, types.SDRCommandSetVersion, 0x02, 0x01, 0x00})

	hctx := &HandlerContext{BMC: b}
	resp, cc, err := handleGetSDRRepoInfo(context.Background(), hctx, nil)
	if err != nil || cc != CodeOK {
		t.Fatalf("cc=%v err=%v", cc, err)
	}
	var info storage.GetSDRRepoInfoResponse
	if err := info.Unpack(resp); err != nil {
		t.Fatal(err)
	}
	if info.RecordCount != 1 || info.SDRVersion != types.SDRCommandSetVersion {
		t.Fatalf("unexpected info: %+v", info)
	}
	if !info.SDROperationSupport.SupportReserveSDRRepo {
		t.Fatal("expected reserve support")
	}
}

func TestHandleGetDeviceID_StorageBits(t *testing.T) {
	b, m := newTestBMCWithStorage(t)
	_ = m.Storage().FRU().Write(context.Background(), 0, testFRUBytes())
	_ = m.Storage().SDR().Write(context.Background(), 1, []byte{0x01, 0x00, types.SDRCommandSetVersion, 0x02, 0x01, 0x00})

	hctx := &HandlerContext{BMC: b}
	resp, cc, err := handleGetDeviceID(context.Background(), hctx, nil)
	if err != nil || cc != CodeOK {
		t.Fatalf("cc=%v err=%v", cc, err)
	}
	if resp[1]&0x80 != 0 {
		t.Fatal("byte 1 bit 7 is ProvideDeviceSDRs, must not be set unless Device SDR is supported")
	}
	if resp[5]&0x02 == 0 {
		t.Fatal("SDR repository bit not set in additional device support")
	}
	if resp[5]&0x08 == 0 {
		t.Fatal("FRU inventory bit not set in additional support")
	}
}

func TestStorageHandlers_NilHAL(t *testing.T) {
	info := bmc.DeviceInfo{DeviceID: 1, IPMIVersion: 0x20}
	var guid [16]byte
	b := bmc.New(info, guid, nil)
	hctx := &HandlerContext{BMC: b}

	_, cc, err := handleGetFRUInventoryAreaInfo(context.Background(), hctx, []byte{0})
	if err != nil || cc != CodeNotSupportedInState {
		t.Fatalf("fru info: cc=%v err=%v", cc, err)
	}
	_, cc, err = handleGetSDR(context.Background(), hctx, (&storage.GetSDRRequest{}).Pack())
	if err != nil || cc != CodeNotSupportedInState {
		t.Fatalf("get sdr: cc=%v err=%v", cc, err)
	}
}

func TestSeedFRU_ClientParse(t *testing.T) {
	data := types.PackFRU(types.FRUPackConfig{
		Chassis: &types.FRUPackChassis{Type: 0x17},
		Board: &types.FRUPackBoard{
			Mfg:     "Acme",
			Product: "Board-A",
		},
		Product: &types.FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "Product-A",
		},
	})
	hdr := &types.FRUCommonHeader{}
	if err := hdr.Unpack(data[:types.FRUCommonHeaderSize]); err != nil {
		t.Fatal(err)
	}
	if !hdr.Valid() {
		t.Fatal("invalid common header checksum")
	}
}
