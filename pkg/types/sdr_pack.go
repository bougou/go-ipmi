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

	mc := &SDRMgmtControllerDeviceLocator{
		DeviceSlaveAddress:           slave,
		DeviceCap_FRUInventoryDevice: devSupport&0x08 != 0,
		DeviceCap_SELDevice:          devSupport&0x04 != 0,
		DeviceCap_SDRRepoDevice:      devSupport&0x02 != 0,
		DeviceCap_SensorDevice:       devSupport&0x01 != 0,
		DeviceCap_ChassisDevice:      devSupport&0x80 != 0,
		DeviceCap_Bridge:             devSupport&0x40 != 0,
		DeviceCap_IPMBEventGenerator: devSupport&0x20 != 0,
		DeviceCap_IPMBEventReceiver:  devSupport&0x10 != 0,
		EntityID:                     entityID,
		EntityInstance:               opts.EntityInstance,
		DeviceIDBytes:                name,
	}
	if len(name) > 0 {
		mc.DeviceIDTypeLength = TypeLength(0xC0 | byte(len(name)))
	}
	return mc.Pack(opts.RecordID)
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

	entityID := opts.EntityID
	if entityID == 0 {
		entityID = 0x07
	}
	sensorType := opts.SensorType
	if sensorType == 0 {
		sensorType = byte(SensorTypeTemperature)
	}

	c := &SDRCompact{
		GeneratorID:    0x0020,
		SensorNumber:   SensorNumber(opts.SensorNumber),
		SensorEntityID: EntityID(entityID),
		SensorInitialization: SensorInitialization{
			InitScanning:           true,
			InitEvents:             true,
			EventGenerationEnabled: true,
			SensorScanningEnabled:  true,
		},
		SensorType:             SensorType(sensorType),
		SensorEventReadingType: EventReadingTypeThreshold,
		SensorUnit: SensorUnit{
			BaseUnit: SensorUnitType_DegreesC,
		},
		IDStringTypeLength: TypeLength(0xC0 | byte(len(name))),
		IDStringBytes:      name,
	}
	return c.Pack(opts.RecordID)
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
