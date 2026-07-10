package types

// MCLocatorPackOpts configures [PackMCLocator].
// The defaults match a single local BMC with FRU inventory (§43.9 Type 12h),
// suitable for ipmitool fru list without seeding sensor SDR records.
type MCLocatorPackOpts struct {
	RecordID           uint16
	DeviceSlaveAddress uint8 // default 0x20 (BMC system software)
	DeviceSupport      uint8 // default 0x08 (FRU Inventory Device)
	EntityID           uint8 // default 0x07 (system board)
	EntityInstance     uint8
	Name               string // optional MC ID string (max 16 bytes)
}

// PackMCLocator builds a Type 12h Management Controller Device Locator wire
// record per §43.9. The layout matches [parseSDRManagementControllerDeviceLocator].
func PackMCLocator(opts MCLocatorPackOpts) []byte {
	name := []byte(opts.Name)
	if len(name) > 16 {
		name = name[:16]
	}
	n := len(name)

	// Fixed body bytes 5..15 (KEY + BODY through empty ID type/length); name follows at 16.
	const fixedBody = 11
	recLen := fixedBody + n
	rec := make([]byte, SDRRecordHeaderSize+recLen)

	if opts.RecordID == 0 {
		opts.RecordID = 1
	}
	slave := opts.DeviceSlaveAddress
	if slave == 0 {
		slave = 0x20
	}
	devSupport := opts.DeviceSupport
	if devSupport == 0 {
		devSupport = 0x08 // FRU Inventory Device
	}
	entityID := opts.EntityID
	if entityID == 0 {
		entityID = 0x07 // system board
	}

	PackUint16L(opts.RecordID, rec, 0)
	rec[2] = SDRCommandSetVersion
	rec[3] = byte(SDRRecordTypeManagementControllerDeviceLocator)
	rec[4] = byte(recLen)

	rec[5] = slave
	rec[6] = 0x00 // channel + reserved
	rec[7] = 0x00 // initialization / ACPI notification
	rec[8] = devSupport
	// bytes 9-11: reserved
	rec[12] = entityID
	rec[13] = opts.EntityInstance
	rec[14] = 0x00 // OEM
	rec[15] = 0xC0 | byte(n)
	if n > 0 {
		copy(rec[16:], name)
	}

	return rec
}

// CompactSensorPackOpts configures [PackCompactSensor].
type CompactSensorPackOpts struct {
	RecordID     uint16
	SensorNumber uint8
	SensorType   uint8 // e.g. SensorTypeTemperature (0x01)
	EntityID     uint8 // e.g. 0x07 system board
	Name         string
}

// PackCompactSensor builds a Type 02h Compact Sensor wire record per §43.2.
// The layout matches [parseSDRCompactSensor].
func PackCompactSensor(opts CompactSensorPackOpts) []byte {
	name := []byte(opts.Name)
	if len(name) > 16 {
		name = name[:16]
	}
	n := len(name)

	// Bytes 5..31 inclusive are fixed (type/length at 31); ID string follows at 32.
	const fixedBody = 27
	recLen := fixedBody + n
	rec := make([]byte, SDRRecordHeaderSize+recLen)

	PackUint16L(opts.RecordID, rec, 0)
	rec[2] = SDRCommandSetVersion
	rec[3] = byte(SDRRecordTypeCompactSensor)
	rec[4] = byte(recLen)

	PackUint16L(0x0020, rec, 5) // Owner ID: BMC system software
	rec[7] = opts.SensorNumber
	if opts.EntityID == 0 {
		rec[8] = 0x07 // system board
	} else {
		rec[8] = opts.EntityID
	}
	rec[9] = 0x00  // entity instance 0, physical
	rec[10] = 0x67 // sensor initialized; scanning + events enabled
	rec[11] = 0x00
	if opts.SensorType == 0 {
		rec[12] = byte(SensorTypeTemperature)
	} else {
		rec[12] = opts.SensorType
	}
	rec[13] = 0x01 // threshold event/reading type
	// bytes 14-19: assertion/deassertion/reading mask — zero
	rec[20] = 0x00 // sensor unit: unsigned, no rate/modifier
	rec[21] = byte(SensorUnitType_DegreesC)
	rec[22] = 0x00
	// bytes 23-26: direction / sharing / hysteresis — zero
	rec[31] = 0xC0 | byte(n)
	copy(rec[32:], name)

	return rec
}

// SplitSDRDump splits concatenated SDR repository bytes into individual records.
func SplitSDRDump(data []byte) map[uint16][]byte {
	out := map[uint16][]byte{}
	offset := 0
	for offset+5 <= len(data) {
		recLen := int(data[offset+4])
		total := 5 + recLen
		if offset+total > len(data) {
			break
		}
		rec := data[offset : offset+total]
		id := uint16(rec[0]) | uint16(rec[1])<<8
		out[id] = append([]byte{}, rec...)
		offset += total
	}
	return out
}
