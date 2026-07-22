package types

import (
	"bytes"
	"fmt"
	"testing"
)

func assertPackSDRRoundTrip(t *testing.T, raw []byte) {
	t.Helper()
	sdr, err := ParseSDR(raw, 0xffff)
	if err != nil {
		t.Fatalf("ParseSDR: %v", err)
	}
	got, err := PackSDR(sdr)
	if err != nil {
		t.Fatalf("PackSDR: %v", err)
	}
	if !bytes.Equal(raw, got) {
		t.Fatalf("round-trip mismatch:\nwant % x\ngot  % x", raw, got)
	}
}

func TestPackMCLocator_ParseRoundTrip(t *testing.T) {
	raw := PackMCLocator(MCLocatorPackOpts{RecordID: 1})
	sdr, err := ParseSDR(raw, 0xffff)
	if err != nil {
		t.Fatal(err)
	}
	if sdr.RecordHeader.RecordType != SDRRecordTypeManagementControllerDeviceLocator {
		t.Fatalf("type: got %v", sdr.RecordHeader.RecordType)
	}
	mc := sdr.MgmtControllerDeviceLocator
	if mc == nil {
		t.Fatal("nil MC locator")
	}
	if mc.DeviceSlaveAddress != 0x20 {
		t.Fatalf("slave: got %#02x", mc.DeviceSlaveAddress)
	}
	if !mc.DeviceCap_FRUInventoryDevice {
		t.Fatal("expected FRU inventory capability")
	}
	assertPackSDRRoundTrip(t, raw)
}

func TestPackCompactSensor_ParseRoundTrip(t *testing.T) {
	raw := PackCompactSensor(CompactSensorPackOpts{
		RecordID:     1,
		SensorNumber: 0x01,
		Name:         "CPU Temp",
	})
	sdr, err := ParseSDR(raw, 0xffff)
	if err != nil {
		t.Fatal(err)
	}
	if sdr.SensorName() != "CPU Temp" {
		t.Fatalf("name: want CPU Temp, got %q", sdr.SensorName())
	}
	if sdr.RecordHeader.RecordType != SDRRecordTypeCompactSensor {
		t.Fatalf("type: got %v", sdr.RecordHeader.RecordType)
	}
	assertPackSDRRoundTrip(t, raw)
}

// TestGenericDeviceLocatorChannelNumber verifies channel encoding per IPMI
// Table 43-6 (SDR Type 10h):
//
//	body[1] bit 0   = channel ms-bit (bit 3 of the 4-bit channel number)
//	body[2] bits 7:5 = channel ls-3 bits (bits 2:0 of the channel number)
func TestGenericDeviceLocatorChannelNumber(t *testing.T) {
	const slaveAddr = uint8(0x50) // 7-bit slave in bits 7:1

	for ch := uint8(0); ch < 16; ch++ {
		t.Run(fmt.Sprintf("channel_%d", ch), func(t *testing.T) {
			loc := &SDRGenericDeviceLocator{
				DeviceAccessAddress: 0x00,
				DeviceSlaveAddress:  slaveAddr,
				ChannelNumber:       ch,
				AccessLUN:           0x02,
				PrivateBusID:        0x03,
				DeviceType:          0x10,
				EntityID:            0x07,
				EntityInstance:      1,
				DeviceIDTypeLength:  TypeLength(0xC0 | 3),
				DeviceIDString:      []byte("DEV"),
			}
			raw := loc.Pack(0x0050)

			parsed := &SDR{}
			if err := parseSDRGenericLocator(raw, parsed); err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := parsed.GenericDeviceLocator.ChannelNumber; got != ch {
				t.Fatalf("round-trip: want %d got %d", ch, got)
			}

			body := raw[5:] // skip 5-byte SDR header
			msBit := body[1] & 0x01
			ls3 := (body[2] >> 5) & 0x07
			wantMS := (ch >> 3) & 0x01
			wantLS3 := ch & 0x07
			if msBit != wantMS || ls3 != wantLS3 {
				t.Fatalf("wire mismatch: ms-bit=%d ls-3=%d, want ms=%d ls-3=%d (body[1]=%#02x body[2]=%#02x)",
					msBit, ls3, wantMS, wantLS3, body[1], body[2])
			}
			if decoded := (msBit << 3) | ls3; decoded != ch {
				t.Fatalf("wire decode %d != channel %d", decoded, ch)
			}

			assertPackSDRRoundTrip(t, raw)
		})
	}

	// Golden wire bytes per Table 43-6 (LUN=0, PrivateBusID=0).
	golden := []struct {
		ch       uint8
		wantBody [2]byte // body[1], body[2]
	}{
		{5, [2]byte{0x50, 0xA0}},
		{8, [2]byte{0x51, 0x00}},
		{9, [2]byte{0x51, 0x20}},
	}
	for _, tc := range golden {
		t.Run(fmt.Sprintf("golden_%d", tc.ch), func(t *testing.T) {
			raw := (&SDRGenericDeviceLocator{
				DeviceSlaveAddress: slaveAddr,
				ChannelNumber:      tc.ch,
				DeviceType:         0x10,
				EntityID:           0x07,
				DeviceIDTypeLength: TypeLength(0xC0 | 3),
				DeviceIDString:     []byte("DEV"),
			}).Pack(0x0050)
			body := raw[5:]
			if body[1] != tc.wantBody[0] || body[2] != tc.wantBody[1] {
				t.Fatalf("channel %d: body[1:3]=%#02x %#02x, want %#02x %#02x",
					tc.ch, body[1], body[2], tc.wantBody[0], tc.wantBody[1])
			}
		})
	}
}

func TestEventOnlySharingFieldsRoundTrip(t *testing.T) {
	// Golden bit packing per v2.0§43.3 Table 43-3:
	// Direction=10b (output), Modifier=01b (alpha), ShareCount=3 → byte1 = 0x93
	// EntitySharing=1, Offset=0x15 → byte2 = 0x95
	raw := (&SDREventOnly{
		GeneratorID:                    0x0020,
		SensorNumber:                   0x02,
		SensorEntityID:                 0x07,
		SensorType:                     SensorTypeFan,
		SensorEventReadingType:         EventReadingTypeThreshold,
		SensorDirection:                0x02,
		IDStringInstanceModifierType:   0x01,
		ShareCount:                     3,
		EntityInstanceSharing:          true,
		IDStringInstanceModifierOffset: 0x15,
		OEM:                            0xAB,
		IDStringTypeLength:             TypeLength(0xC0 | 3),
		IDStringBytes:                  []byte("Fan"),
	}).Pack(0x20)

	body := raw[5:] // skip 5-byte SDR header
	if body[7] != 0x93 || body[8] != 0x95 {
		t.Fatalf("sharing bytes: want 0x93 0x95, got %#02x %#02x", body[7], body[8])
	}
	if body[9] != 0x00 {
		t.Fatalf("reserved byte: want 0x00, got %#02x", body[9])
	}
	if body[10] != 0xAB {
		t.Fatalf("OEM byte: want 0xAB, got %#02x", body[10])
	}

	assertPackSDRRoundTrip(t, raw)

	sdr, err := ParseSDR(raw, 0xffff)
	if err != nil {
		t.Fatalf("ParseSDR: %v", err)
	}
	eo := sdr.EventOnly
	if eo.SensorDirection != 0x02 {
		t.Fatalf("SensorDirection: want 0x02 got %#02x", eo.SensorDirection)
	}
	if eo.IDStringInstanceModifierType != 0x01 {
		t.Fatalf("IDStringInstanceModifierType: want 0x01 got %#02x", eo.IDStringInstanceModifierType)
	}
	if eo.ShareCount != 3 {
		t.Fatalf("ShareCount: want 3 got %d", eo.ShareCount)
	}
	if !eo.EntityInstanceSharing {
		t.Fatal("EntityInstanceSharing: want true")
	}
	if eo.IDStringInstanceModifierOffset != 0x15 {
		t.Fatalf("IDStringInstanceModifierOffset: want 0x15 got %#02x", eo.IDStringInstanceModifierOffset)
	}
	if eo.OEM != 0xAB {
		t.Fatalf("OEM: want 0xAB got %#02x", eo.OEM)
	}
}

func TestCompactSharingFieldsRoundTrip(t *testing.T) {
	// Same bit packing as Event-Only (v2.0§43.2 Table 43-2 bytes 24–25).
	raw := (&SDRCompact{
		GeneratorID:                    0x0020,
		SensorNumber:                   0x05,
		SensorEntityID:                 0x07,
		SensorType:                     SensorTypeTemperature,
		SensorEventReadingType:         EventReadingTypeThreshold,
		SensorDirection:                0x01, // input
		IDStringInstanceModifierType:   0x00, // numeric
		ShareCount:                     4,
		EntityInstanceSharing:          true,
		IDStringInstanceModifierOffset: 0x0A,
		SensorUnit: SensorUnit{
			BaseUnit: SensorUnitType_DegreesC,
		},
		IDStringTypeLength: TypeLength(0xC0 | 4),
		IDStringBytes:      []byte("Temp"),
	}).Pack(0x15)

	body := raw[5:]
	// Direction=01b, Modifier=00b, ShareCount=4 → 0x44
	// EntitySharing=1, Offset=0x0A → 0x8A
	if body[18] != 0x44 || body[19] != 0x8A {
		t.Fatalf("sharing bytes: want 0x44 0x8A, got %#02x %#02x", body[18], body[19])
	}

	assertPackSDRRoundTrip(t, raw)

	sdr, err := ParseSDR(raw, 0xffff)
	if err != nil {
		t.Fatalf("ParseSDR: %v", err)
	}
	c := sdr.Compact
	if c.SensorDirection != 0x01 || c.ShareCount != 4 || !c.EntityInstanceSharing {
		t.Fatalf("sharing fields: dir=%#02x count=%d sharing=%v", c.SensorDirection, c.ShareCount, c.EntityInstanceSharing)
	}
	if c.IDStringInstanceModifierOffset != 0x0A {
		t.Fatalf("offset: want 0x0A got %#02x", c.IDStringInstanceModifierOffset)
	}
}

func TestUnpackSensorRecordSharingGolden(t *testing.T) {
	dir, mod, count, sharing, offset := unpackSensorRecordSharing(0x93, 0x95)
	if dir != 0x02 || mod != 0x01 || count != 3 || !sharing || offset != 0x15 {
		t.Fatalf("got dir=%#02x mod=%#02x count=%d sharing=%v offset=%#02x", dir, mod, count, sharing, offset)
	}
	b1, b2 := packSensorRecordSharing(dir, mod, count, sharing, offset)
	if b1 != 0x93 || b2 != 0x95 {
		t.Fatalf("repack: want 0x93 0x95, got %#02x %#02x", b1, b2)
	}
}

func TestGenericDeviceLocatorSlaveAddressStripsChannelBit(t *testing.T) {
	const slaveAddr = uint8(0x50)
	raw := (&SDRGenericDeviceLocator{
		DeviceSlaveAddress: slaveAddr,
		ChannelNumber:      9, // sets bit 0 in body[1]
		DeviceType:         0x10,
		EntityID:           0x07,
		DeviceIDTypeLength: TypeLength(0xC0 | 3),
		DeviceIDString:     []byte("DEV"),
	}).Pack(0x0050)

	body := raw[5:]
	if body[1]&0x01 != 1 {
		t.Fatalf("wire body[1] bit 0 should be set for channel 9, got %#02x", body[1])
	}

	sdr := &SDR{}
	if err := parseSDRGenericLocator(raw, sdr); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := sdr.GenericDeviceLocator.DeviceSlaveAddress; got != slaveAddr {
		t.Fatalf("DeviceSlaveAddress: want %#02x (channel bit stripped), got %#02x", slaveAddr, got)
	}
	if got := sdr.GenericDeviceLocator.ChannelNumber; got != 9 {
		t.Fatalf("ChannelNumber: want 9 got %d", got)
	}
}

func TestPackSDR_AllRecordTypes(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
	}{
		{
			name: "full sensor",
			raw: (&SDRFull{
				GeneratorID:    0x0020,
				SensorNumber:   0x01,
				SensorEntityID: 0x07,
				SensorInitialization: SensorInitialization{
					InitScanning:           true,
					InitEvents:             true,
					EventGenerationEnabled: true,
					SensorScanningEnabled:  true,
				},
				SensorType:             SensorTypeTemperature,
				SensorEventReadingType: EventReadingTypeThreshold,
				SensorUnit: SensorUnit{
					BaseUnit: SensorUnitType_DegreesC,
				},
				ReadingFactors:     ReadingFactors{M: 1, B: 0},
				IDStringTypeLength: TypeLength(0xC0 | 4),
				IDStringBytes:      []byte("Temp"),
			}).Pack(0x10),
		},
		{
			name: "event only",
			raw: (&SDREventOnly{
				GeneratorID:            0x0020,
				SensorNumber:           0x02,
				SensorEntityID:         0x07,
				SensorType:             SensorTypeFan,
				SensorEventReadingType: EventReadingTypeThreshold,
				IDStringTypeLength:     TypeLength(0xC0 | 3),
				IDStringBytes:          []byte("Fan"),
			}).Pack(0x20),
		},
		{
			name: "entity association",
			raw: (&SDREntityAssociation{
				ContainerEntityID:        0x07,
				ContainerEntityInstance:  1,
				ContainedEntity1ID:       0x03,
				ContainedEntity1Instance: 1,
			}).Pack(0x30),
		},
		{
			name: "device relative",
			raw: (&SDRDeviceRelative{
				ContainerEntityID:             0x07,
				ContainerEntityInstance:       1,
				ContainerEntityDeviceAddress:  0x20,
				ContainedEntity1DeviceAddress: 0x22,
				ContainedEntity1ID:            0x03,
				ContainedEntity1Instance:      1,
			}).Pack(0x40),
		},
		{
			name: "generic locator",
			raw: (&SDRGenericDeviceLocator{
				DeviceSlaveAddress: 0x50,
				DeviceType:         0x10,
				EntityID:           0x07,
				DeviceIDTypeLength: TypeLength(0xC0 | 3),
				DeviceIDString:     []byte("DEV"),
			}).Pack(0x50),
		},
		{
			name: "fru device locator",
			raw: (&SDRFRUDeviceLocator{
				FRUDeviceID_SlaveAddress: 0x01,
				IsLogicalFRUDevice:       true,
				FRUEntityID:              0x07,
				DeviceIDTypeLength:       TypeLength(0xC0 | 3),
				DeviceIDBytes:            []byte("FRU"),
			}).Pack(0x60),
		},
		{
			name: "mc confirmation",
			raw: (&SDRMgmtControllerConfirmation{
				DeviceSlaveAddress:    0x20,
				DeviceID:              0x01,
				FirmwareMajorRevision: 1,
				FirmwareMinorRevision: 0x10,
				MajorIPMIVersion:      2,
				MinorIPMIVersion:      0,
				ManufacturerID:        0x123456,
				ProductID:             0x789a,
				DeviceGUID:            bytes.Repeat([]byte{0xab}, 16),
			}).Pack(0x70),
		},
		{
			name: "bmc channel info",
			raw: (&SDRBMCChannelInfo{
				Channel0: ChannelInfo{
					TransmitSupported: true,
					MessageReceiveLUN: 0,
					ChannelProtocol:   0x01,
				},
				MessagingInterruptType: 0x01,
			}).Pack(0x80),
		},
		{
			name: "oem",
			raw: (&SDROEM{
				ManufacturerID: 0x00aabb,
				OEMData:        []byte{0x01, 0x02, 0x03},
			}).Pack(0x90),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertPackSDRRoundTrip(t, tc.raw)
		})
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
