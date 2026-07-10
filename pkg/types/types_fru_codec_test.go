package types

import (
	"bytes"
	"testing"
)

func assertPackFRURoundTrip(t *testing.T, raw []byte) {
	t.Helper()
	fru, err := ParseFRU(raw)
	if err != nil {
		t.Fatalf("ParseFRU: %v", err)
	}
	got, err := PackFRUInventory(fru)
	if err != nil {
		t.Fatalf("PackFRUInventory: %v", err)
	}
	if !bytes.Equal(raw, got) {
		t.Fatalf("round-trip mismatch:\nwant % x\ngot  % x", raw, got)
	}
}

func TestPackFRU_ProductOnly(t *testing.T) {
	data, err := PackFRU(FRUPackConfig{
		Product: &FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "TestBMC",
			Version:      "1.0",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
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
	assertPackFRURoundTrip(t, data)
}

func TestPackFRU_MultiArea(t *testing.T) {
	data, err := PackFRU(FRUPackConfig{
		Chassis: &FRUPackChassis{Type: 0x17},
		Product: &FRUPackProduct{
			Manufacturer: "Acme",
			Name:         "TestBMC",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	hdr := &FRUCommonHeader{}
	if err := hdr.Unpack(data[:FRUCommonHeaderSize]); err != nil {
		t.Fatal(err)
	}
	if hdr.ChassisOffset8B == 0 || hdr.ProductOffset8B == 0 {
		t.Fatalf("expected chassis and product areas: %+v", hdr)
	}
	assertPackFRURoundTrip(t, data)
}

func TestFRUMultiRecord_HeaderChecksum(t *testing.T) {
	rec := &FRUMultiRecord{
		RecordType:    FRURecordType(0x00),
		FormatVersion: 0,
		RecordData:    bytes.Repeat([]byte{0x11}, 22),
	}
	wire := rec.Pack()
	if len(wire) != 5+22 {
		t.Fatalf("wire len: got %d want %d", len(wire), 5+22)
	}
	sum := 0
	for _, b := range wire[:5] {
		sum = (sum + int(b)) % 256
	}
	if sum != 0 {
		t.Fatalf("header checksum invalid: sum=%d wire=% x", sum, wire[:5])
	}
	// FRU spec Table 17-2 example: bytes [0,0,22,9,225] sum to zero mod 256.
	spec := []byte{0, 0, 22, 9, 225}
	sum = 0
	for _, b := range spec {
		sum = (sum + int(b)) % 256
	}
	if sum != 0 {
		t.Fatalf("spec example header sum: %d", sum)
	}
}

func TestPackFRU_InternalUseArea(t *testing.T) {
	// First data byte 0x02 must not be interpreted as a length field (FRU §9).
	data, err := PackFRUInventory(&FRU{
		CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion},
		InternalUseArea: &FRUInternalUseArea{
			Data: []byte{0x02, 0xAA, 0xBB, 0xCC},
		},
		ProductInfoArea: &FRUProductInfoArea{
			FormatVersion:          FRUFormatVersion,
			ManufacturerTypeLength: fruASCIITypeLength([]byte("Acme")),
			Manufacturer:           []byte("Acme"),
			NameTypeLength:         fruASCIITypeLength([]byte("BMC")),
			Name:                   []byte("BMC"),
			PartModelTypeLength:    TypeLength(0xC0),
			VersionTypeLength:      TypeLength(0xC0),
			SerialNumberTypeLength: TypeLength(0xC0),
			AssetTagTypeLength:     TypeLength(0xC0),
			FRUFileIDTypeLength:    TypeLength(0xC0),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseFRU(data)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.InternalUseArea == nil {
		t.Fatal("missing internal use area")
	}
	wantData := []byte{0x02, 0xAA, 0xBB, 0xCC, 0, 0, 0} // padded to 8 bytes after format version
	if !bytes.Equal(parsed.InternalUseArea.Data, wantData) {
		t.Fatalf("internal data: got % x want % x", parsed.InternalUseArea.Data, wantData)
	}
	if parsed.InternalUseArea.FormatVersion != FRUFormatVersion {
		t.Fatalf("format version: got %#02x", parsed.InternalUseArea.FormatVersion)
	}
	assertPackFRURoundTrip(t, data)
}

func TestPackFRUInventory_MultiRecord(t *testing.T) {
	dc := &FRURecordTypeDCOutput{
		OutputWhenOff:          true,
		OutputNumber:           1,
		NominalVoltage10mV:     1200,
		MaxNegativeVoltage10mV: -50,
		MaxPositiveVoltage10mV: 50,
		RippleNoise1mV:         10,
		MinCurrentDraw1mA:      100,
		MaxCurrentDraw1mA:      5000,
	}
	recordData := dc.Pack()

	fru := &FRU{
		CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion},
		ProductInfoArea: &FRUProductInfoArea{
			FormatVersion:          FRUFormatVersion,
			ManufacturerTypeLength: fruASCIITypeLength([]byte("Acme")),
			Manufacturer:           []byte("Acme"),
			NameTypeLength:         fruASCIITypeLength([]byte("PSU")),
			Name:                   []byte("PSU"),
			PartModelTypeLength:    TypeLength(0xC0),
			VersionTypeLength:      TypeLength(0xC0),
			SerialNumberTypeLength: TypeLength(0xC0),
			AssetTagTypeLength:     TypeLength(0xC0),
			FRUFileIDTypeLength:    TypeLength(0xC0),
		},
		MultiRecords: []*FRUMultiRecord{{
			RecordType:    FRURecordType(0x01),
			EndOfList:     true,
			FormatVersion: 0,
			RecordData:    recordData,
		}},
	}

	raw, err := PackFRUInventory(fru)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseFRU(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.MultiRecords) != 1 {
		t.Fatalf("expected 1 multi-record, got %d", len(parsed.MultiRecords))
	}
	got := &FRURecordTypeDCOutput{}
	if err := got.Unpack(parsed.MultiRecords[0].RecordData); err != nil {
		t.Fatal(err)
	}
	if got.NominalVoltage10mV != dc.NominalVoltage10mV || !got.OutputWhenOff {
		t.Fatalf("dc output mismatch: %+v", got)
	}
	assertPackFRURoundTrip(t, raw)
}

func TestPackFRURecordTypes_RoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		pack   func() []byte
		unpack func([]byte) error
	}{
		{
			name: "dc output",
			pack: func() []byte {
				return (&FRURecordTypeDCOutput{
					OutputNumber: 2, NominalVoltage10mV: 500,
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeDCOutput{}).Unpack(b)
			},
		},
		{
			name: "extended dc output",
			pack: func() []byte {
				return (&FRURecordTypeExtendedDCOutput{
					CurrentUnits100: true, OutputNumber: 1,
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeExtendedDCOutput{}).Unpack(b)
			},
		},
		{
			name: "dc load",
			pack: func() []byte {
				return (&FRURecordTypeDCLoad{OutputNumber: 3}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeDCLoad{}).Unpack(b)
			},
		},
		{
			name: "extended dc load",
			pack: func() []byte {
				return (&FRURecordTypeExtendedDCLoad{
					IsCurrentUnit100mA: true,
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeExtendedDCLoad{}).Unpack(b)
			},
		},
		{
			name: "management access",
			pack: func() []byte {
				return (&FRURecordTypeManagementAccess{
					SubRecordType: 0x02,
					Data:          []byte("bmc.example.com"),
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeManagementAccess{}).Unpack(b)
			},
		},
		{
			name: "base compatibility",
			pack: func() []byte {
				return (&FRURecordTypeBaseCompatibility{
					ManufacturerID: 0x123456,
					EntityID:       0x07,
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeBaseCompatibility{}).Unpack(b)
			},
		},
		{
			name: "extended compatibility",
			pack: func() []byte {
				return (&FRURecordTypeExtendedCompatibilityRecord{
					ManufacturerID: 0xabcdef,
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeExtendedCompatibilityRecord{}).Unpack(b)
			},
		},
		{
			name: "oem",
			pack: func() []byte {
				return (&FRURecordTypeOEM{
					ManufacturerID: 0x00aabb,
					Data:           []byte{0xde, 0xad},
				}).Pack()
			},
			unpack: func(b []byte) error {
				return (&FRURecordTypeOEM{}).Unpack(b)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := tc.pack()
			if err := tc.unpack(raw); err != nil {
				t.Fatalf("unpack: %v", err)
			}
			fru := &FRU{
				CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion},
				MultiRecords: []*FRUMultiRecord{{
					RecordType: FRURecordType(0xC0),
					EndOfList:  true,
					RecordData: raw,
				}},
			}
			wire, err := PackFRUInventory(fru)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := ParseFRU(wire)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, parsed.MultiRecords[0].RecordData) {
				t.Fatalf("record data mismatch: % x vs % x", raw, parsed.MultiRecords[0].RecordData)
			}
		})
	}
}
