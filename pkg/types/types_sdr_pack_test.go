package types

import "testing"

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
}
