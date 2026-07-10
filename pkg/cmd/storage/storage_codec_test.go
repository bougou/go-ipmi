package storage

import (
	"testing"
	"time"
)

func TestGetFRUInventoryAreaInfoCodecRoundTrip(t *testing.T) {
	reqOrig := &GetFRUInventoryAreaInfoRequest{FRUDeviceID: 0x05}
	var req GetFRUInventoryAreaInfoRequest
	if err := req.Unpack(reqOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if req.FRUDeviceID != reqOrig.FRUDeviceID {
		t.Fatalf("FRUDeviceID: want %d, got %d", reqOrig.FRUDeviceID, req.FRUDeviceID)
	}

	resOrig := &GetFRUInventoryAreaInfoResponse{
		AreaSizeBytes:         512,
		DeviceAccessedByWords: false,
	}
	var res GetFRUInventoryAreaInfoResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if res.AreaSizeBytes != resOrig.AreaSizeBytes || res.DeviceAccessedByWords != resOrig.DeviceAccessedByWords {
		t.Fatalf("response mismatch: %+v vs %+v", resOrig, res)
	}
}

func TestReadFRUDataCodecRoundTrip(t *testing.T) {
	reqOrig := &ReadFRUDataRequest{FRUDeviceID: 0, ReadOffset: 16, ReadCount: 32}
	var req ReadFRUDataRequest
	if err := req.Unpack(reqOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if req != *reqOrig {
		t.Fatalf("request mismatch: %+v vs %+v", reqOrig, req)
	}

	resOrig := &ReadFRUDataResponse{CountReturned: 4, Data: []byte{0x01, 0x02, 0x03, 0x04}}
	var res ReadFRUDataResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if res.CountReturned != resOrig.CountReturned || len(res.Data) != len(resOrig.Data) {
		t.Fatalf("response mismatch: %+v vs %+v", resOrig, res)
	}
}

func TestGetSDRRepoInfoCodecRoundTrip(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	resOrig := &GetSDRRepoInfoResponse{
		SDRVersion:             0x51,
		RecordCount:            42,
		FreeSpaceBytes:         1024,
		MostRecentAdditionTime: ts,
		MostRecentEraseTime:    ts,
		SDROperationSupport: SDROperationSupport{
			SupportReserveSDRRepo:      true,
			SupportGetSDRRepoAllocInfo: true,
		},
	}
	var res GetSDRRepoInfoResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if res.SDRVersion != resOrig.SDRVersion || res.RecordCount != resOrig.RecordCount {
		t.Fatalf("basic fields mismatch")
	}
	if !res.SDROperationSupport.SupportReserveSDRRepo || !res.SDROperationSupport.SupportGetSDRRepoAllocInfo {
		t.Fatalf("SDROperationSupport flags not preserved")
	}
}

func TestGetSDRRepoAllocInfoCodecRoundTrip(t *testing.T) {
	resOrig := &GetSDRRepoAllocInfoResponse{
		PossibleAllocUnits: 64,
		AllocUnitsSize:     16,
		FreeAllocUnits:     32,
		LargestFreeBlock:   8,
		MaximumRecordSize:  64,
	}
	var res GetSDRRepoAllocInfoResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if *resOrig != res {
		t.Fatalf("response mismatch: %+v vs %+v", resOrig, res)
	}
}

func TestReserveSDRRepoCodecRoundTrip(t *testing.T) {
	resOrig := &ReserveSDRRepoResponse{ReservationID: 0x1234}
	var res ReserveSDRRepoResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if res.ReservationID != resOrig.ReservationID {
		t.Fatalf("ReservationID: want %d, got %d", resOrig.ReservationID, res.ReservationID)
	}
}

func TestGetSDRCodecRoundTrip(t *testing.T) {
	reqOrig := &GetSDRRequest{
		ReservationID: 0x0001,
		RecordID:      0x0002,
		ReadOffset:    8,
		ReadBytes:     0xff,
	}
	var req GetSDRRequest
	if err := req.Unpack(reqOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if *reqOrig != req {
		t.Fatalf("request mismatch: %+v vs %+v", reqOrig, req)
	}

	resOrig := &GetSDRResponse{
		NextRecordID: 0xffff,
		RecordData:   []byte{0x01, 0x00, 0x51, 0x01, 0x10},
	}
	var res GetSDRResponse
	if err := res.Unpack(resOrig.Pack()); err != nil {
		t.Fatal(err)
	}
	if res.NextRecordID != resOrig.NextRecordID {
		t.Fatalf("NextRecordID mismatch")
	}
	if len(res.RecordData) != len(resOrig.RecordData) {
		t.Fatalf("RecordData length mismatch")
	}
}

func TestGetSDRRequestUnpackTruncated(t *testing.T) {
	var req GetSDRRequest
	if err := req.Unpack([]byte{0x00}); err == nil {
		t.Fatal("expected error for truncated request")
	}
}
