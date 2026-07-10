package types

import (
	"bytes"
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
