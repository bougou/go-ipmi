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
	assertPackFRURoundTrip(t, data)
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
