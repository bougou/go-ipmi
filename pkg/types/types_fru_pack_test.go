package types

import "testing"

func TestPackFRU_ProductOnly(t *testing.T) {
	data := PackFRU(FRUPackConfig{
		Product: &FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "TestBMC",
			Version:      "1.0",
		},
	})
	if len(data) < int(FRUCommonHeaderSize) {
		t.Fatalf("FRU too short: %d", len(data))
	}
	hdr := &FRUCommonHeader{}
	if err := hdr.Unpack(data[:FRUCommonHeaderSize]); err != nil {
		t.Fatal(err)
	}
	if !hdr.Valid() {
		t.Fatal("invalid common header checksum")
	}
	if hdr.ProductOffset8B == 0 {
		t.Fatal("expected product area")
	}
}

func TestPackFRU_MultiArea(t *testing.T) {
	data := PackFRU(FRUPackConfig{
		Chassis: &FRUPackChassis{Type: 0x17},
		Product: &FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "TestBMC",
		},
	})
	hdr := &FRUCommonHeader{}
	if err := hdr.Unpack(data[:FRUCommonHeaderSize]); err != nil {
		t.Fatal(err)
	}
	if hdr.ChassisOffset8B == 0 || hdr.ProductOffset8B == 0 {
		t.Fatalf("expected chassis and product areas: %+v", hdr)
	}
}

func TestSplitSDRDump(t *testing.T) {
	dump := []byte{
		0x01, 0x00, 0x51, 0x02, 0x02, 0xaa, 0xbb,
		0x02, 0x00, 0x51, 0x02, 0x01, 0xcc,
	}
	got := SplitSDRDump(dump)
	if len(got) != 2 {
		t.Fatalf("want 2 records, got %d", len(got))
	}
	if len(got[1]) != 7 || len(got[2]) != 6 {
		t.Fatalf("unexpected record sizes: %d %d", len(got[1]), len(got[2]))
	}
}
