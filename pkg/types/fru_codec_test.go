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
		FormatVersion: FRUMultiRecordFormatVersion,
		RecordData:    bytes.Repeat([]byte{0x11}, 22),
	}
	wire := rec.Pack()
	if len(wire) != 5+22 {
		t.Fatalf("wire len: got %d want %d", len(wire), 5+22)
	}
	if wire[1]&0x0f != FRUMultiRecordFormatVersion {
		t.Fatalf("FormatVersion: want %#02x got %#02x", FRUMultiRecordFormatVersion, wire[1]&0x0f)
	}
	sum := 0
	for _, b := range wire[:5] {
		sum = (sum + int(b)) % 256
	}
	if sum != 0 {
		t.Fatalf("header checksum invalid: sum=%d wire=% x", sum, wire[:5])
	}
	// FRU spec Table 17-2 example header bytes sum to zero mod 256.
	spec := []byte{0, 0, 22, 9, 225}
	sum = 0
	for _, b := range spec {
		sum = (sum + int(b)) % 256
	}
	if sum != 0 {
		t.Fatalf("spec example header sum: %d", sum)
	}
}

func TestFRUMultiRecord_DefaultFormatVersion(t *testing.T) {
	// fru§16.2.3: FormatVersion 0 on Pack defaults to 02h.
	wire := (&FRUMultiRecord{
		RecordType: FRURecordType(0x01),
		EndOfList:  true,
		RecordData: []byte{0xaa},
	}).Pack()
	if wire[1]&0x0f != FRUMultiRecordFormatVersion {
		t.Fatalf("default FormatVersion: want %#02x got %#02x", FRUMultiRecordFormatVersion, wire[1]&0x0f)
	}
	if wire[1]&0x80 == 0 {
		t.Fatal("EndOfList bit not set")
	}
}

func TestPackMultiRecords_ForcesEndOfList(t *testing.T) {
	raw := PackMultiRecords([]*FRUMultiRecord{
		{RecordType: 0x01, RecordData: []byte{0x01}},
		{RecordType: 0x02, RecordData: []byte{0x02}},
	})
	if raw[1]&0x80 != 0 {
		t.Fatal("first record must not have EndOfList")
	}
	// second record header starts after 5+1 bytes
	if raw[7]&0x80 == 0 {
		t.Fatal("last record must have EndOfList")
	}
}

func TestPackFRU_InternalUseArea(t *testing.T) {
	// First data byte 0x02 must not be interpreted as a length field (fru§9).
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

func TestFRU_BCDFieldRoundTrip(t *testing.T) {
	// BCD-plus type/length 01b (fru§13): two digits per byte.
	// Raw 0x21 → chars "12" (low nibble first).
	bcdTL := TypeLength(0x40 | 1) // type=01b, length=1
	bcdRaw := []byte{0x21}

	fru := &FRU{
		CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion},
		ChassisInfoArea: &FRUChassisInfoArea{
			FormatVersion:          FRUFormatVersion,
			ChassisType:            0x17,
			PartNumberTypeLength:   bcdTL,
			PartNumber:             append([]byte(nil), bcdRaw...),
			SerialNumberTypeLength: TypeLength(0xC0),
		},
	}
	raw, err := PackFRUInventory(fru)
	if err != nil {
		t.Fatal(err)
	}
	assertPackFRURoundTrip(t, raw)

	parsed, err := ParseFRU(raw)
	if err != nil {
		t.Fatal(err)
	}
	got := parsed.ChassisInfoArea
	if got.PartNumberTypeLength != bcdTL {
		t.Fatalf("PartNumber TL: want %#02x got %#02x", bcdTL, got.PartNumberTypeLength)
	}
	if !bytes.Equal(got.PartNumber, bcdRaw) {
		t.Fatalf("PartNumber raw: want % x got % x", bcdRaw, got.PartNumber)
	}
	if s := FRUFieldString(got.PartNumberTypeLength, got.PartNumber); s != "12" {
		t.Fatalf("PartNumber decoded: want 12 got %q", s)
	}
}

func TestFRU_CustomFieldPreservesTypeLength(t *testing.T) {
	custom := FRUField{
		TypeLength: TypeLength(0x40 | 2), // BCD, 2 bytes
		Data:       []byte{0x21, 0x43},   // "1234"
	}
	fru := &FRU{
		CommonHeader: &FRUCommonHeader{FormatVersion: FRUFormatVersion},
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
			Custom:                 []FRUField{custom},
		},
	}
	raw, err := PackFRUInventory(fru)
	if err != nil {
		t.Fatal(err)
	}
	assertPackFRURoundTrip(t, raw)

	parsed, err := ParseFRU(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.ProductInfoArea.Custom) != 1 {
		t.Fatalf("custom count: want 1 got %d", len(parsed.ProductInfoArea.Custom))
	}
	got := parsed.ProductInfoArea.Custom[0]
	if got.TypeLength != custom.TypeLength {
		t.Fatalf("custom TL: want %#02x got %#02x", custom.TypeLength, got.TypeLength)
	}
	if !bytes.Equal(got.Data, custom.Data) {
		t.Fatalf("custom data: want % x got % x", custom.Data, got.Data)
	}
}

func TestFRUProduct_NullFieldEndsCustomWithoutC1(t *testing.T) {
	// Real BMCs often omit C1h and pad with null Type/Length (0xC0) + zeros.
	// Parsing must stop at the null field and not treat the checksum as Type/Length.
	body := []byte{
		FRUFormatVersion, 0, // version, length placeholder
		0x00, // language
	}
	body = append(body, packFRUASCIIField("Acme")...)
	body = append(body, packFRUASCIIField("BMC")...)
	body = append(body, packFRUASCIIField("")...)   // part
	body = append(body, packFRUASCIIField("")...)   // version
	body = append(body, packFRUASCIIField("")...)   // serial
	body = append(body, packFRUASCIIField("")...)   // asset
	body = append(body, packFRUASCIIField("")...)   // file id
	body = append(body, packFRUASCIIField("XY")...) // one custom (TL must not be C1h)
	body = append(body, 0xC0)                       // null field (no C1h)
	// Pad to 8-byte multiple including checksum byte.
	for (len(body)+1)%8 != 0 {
		body = append(body, 0x00)
	}
	body = append(body, 0) // checksum placeholder
	body[1] = uint8(len(body) / 8)
	body[len(body)-1] = fruPackChecksum(body[:len(body)-1])

	area := &FRUProductInfoArea{}
	if err := area.Unpack(body); err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if len(area.Custom) != 1 {
		t.Fatalf("custom count: want 1 got %d", len(area.Custom))
	}
	if string(area.Custom[0].Data) != "XY" {
		t.Fatalf("custom data: want XY got %q", area.Custom[0].Data)
	}
}
