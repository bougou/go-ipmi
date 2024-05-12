package ipmi

import "fmt"

// 43.3 SDR Type 03h, Event-Only Record
type SDREventOnly struct {
	//
	// Record KEY
	//

	GeneratorID  GeneratorID
	SensorNumber SensorNumber // Unique number identifying the sensor behind a given slave address and LUN. Code FFh reserved.

	//
	// RECORD BODY
	//

	SensorEntityID       EntityID
	SensorEntityInstance EntityInstance
	// 0b = treat entity as a physical entity per Entity ID table
	// 1b = treat entity as a logical container entity. For example, if this bit is set,
	// and the Entity ID is "Processor", the container entity would be considered
	// to represent a logical "Processor Group" rather than a physical processor.
	// This bit is typically used in conjunction with an Entity Association record.
	SensorEntityIsLogical bool

	SensorType             SensorType
	SensorEventReadingType EventReadingType

	SensorDirection uint8

	IDStringInstanceModifierType uint8

	// Share count (number of sensors sharing this record). Sensor numbers sharing this
	// record are sequential starting with the sensor number specified by the Sensor
	// Number field for this record. E.g. if the starting sensor number was 10, and the share
	// count was 3, then sensors 10, 11, and 12 would share this record.
	ShareCount uint8

	EntityInstanceSharing bool

	// Multiple Discrete sensors can share the same sensor data record. The ID String Instance
	// Modifier and Modifier Offset are used to modify the Sensor ID String as follows:
	// Suppose sensor ID is "Temp " for "Temperature Sensor", share count = 3, ID string
	// instance modifier = numeric, instance modifier offset = 5 - then the sensors could be
	// identified as:
	// Temp 5, Temp 6, Temp 7
	// If the modifier = alpha, and offset = 26, then the sensors could be identified as:
	// Temp AA, Temp AB, Temp AC
	// (alpha characters are considered to be base 26 for ASCII)
	IDStringInstanceModifierOffset uint8

	IDStringTypeLength TypeLength
	IDStringBytes      []byte
}

func (eventOnly *SDREventOnly) String() string {
	return fmt.Sprintf(`Sensor ID              : %s (%#02x)
	Generator             : %d
	Entity ID             : %d.%d (%s)
	Sensor Type (%s) : %s (%#02x)`,
		string(eventOnly.IDStringBytes), eventOnly.SensorNumber,
		eventOnly.GeneratorID,
		uint8(eventOnly.SensorEntityID), uint8(eventOnly.SensorEntityInstance), eventOnly.SensorEntityID.String(),
		eventOnly.SensorEventReadingType.SensorClass(), eventOnly.SensorType.String(), uint8(eventOnly.SensorType),
	)
}

func parseSDREventOnly(data []byte, sdr *SDR) error {
	const SDREventOnlyMinSize int = 17
	minSize := SDREventOnlyMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize)
	}

	s := &SDREventOnly{}
	sdr.EventOnly = s

	generatorID, _, _ := unpackUint16L(data, 5)
	s.GeneratorID = GeneratorID(generatorID)

	sensorNumber, _, _ := unpackUint8(data, 7)
	s.SensorNumber = SensorNumber(sensorNumber)

	b8, _, _ := unpackUint8(data, 8)
	s.SensorEntityID = EntityID(b8)

	b9, _, _ := unpackUint8(data, 9)
	s.SensorEntityInstance = EntityInstance(b9 & 0x7f)
	s.SensorEntityIsLogical = isBit7Set(b9)

	sensorType, _, _ := unpackUint8(data, 10)
	s.SensorType = SensorType(sensorType)

	eventReadingType, _, _ := unpackUint8(data, 11)
	s.SensorEventReadingType = EventReadingType(eventReadingType)

	typeLength, _, _ := unpackUint8(data, 16)
	s.IDStringTypeLength = TypeLength(typeLength)

	idStrLen := int(s.IDStringTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (event-only) data must be longer than %d", minSize+idStrLen)
	}
	s.IDStringBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.4 SDR Type 08h - Entity Association Record
type SDREntityAssociation struct {
	//
	// Record KEY
	//

	ContainerEntityID       uint8
	ContainerEntityInstance uint8

	// [7] - 0b = contained entities specified as list
	//       1b = contained entities specified as range
	ContainedEntitiesAsRange bool
	// [6] - Record Link
	//       0b = no linked Entity Association records
	//       1b = linked Entity Association records exist
	LinkedEntityAssiactionExist bool
	// [5] - 0b = Container entity and contained entities can be assumed absent
	//            if presence sensor for container entity cannot be accessed.
	//            This value is also used if the entity does not have a presence sensor.
	//       1b = Presence sensor should always be accessible. Software should consider
	//            it an error if the presence sensor associated with the container entity
	//            is not accessible. If a presence sensor is accessible, then the
	//            presence sensor can still report that the container entity is absent.
	PresenceSensorAlwaysAccessible bool

	ContaineredEntity1ID       uint8
	ContaineredEntity1Instance uint8

	//
	// RECORD BODY
	//

	ContaineredEntity2ID       uint8
	ContaineredEntity2Instance uint8
	ContaineredEntity3ID       uint8
	ContaineredEntity3Instance uint8
	ContaineredEntity4ID       uint8
	ContaineredEntity4Instance uint8
}

func parseSDREntityAssociation(data []byte, sdr *SDR) error {
	const SDREntityAssociationSize int = 16
	if len(data) < SDREntityAssociationSize {
		return fmt.Errorf("sdr (entity association) data must be longer than %d", SDREntityAssociationSize)
	}

	s := &SDREntityAssociation{}
	sdr.EntityAssociation = s

	s.ContainerEntityID, _, _ = unpackUint8(data, 5)
	s.ContainerEntityInstance, _, _ = unpackUint8(data, 6)

	flag, _, _ := unpackUint8(data, 7)
	s.ContainedEntitiesAsRange = isBit7Set(flag)
	s.LinkedEntityAssiactionExist = isBit6Set(flag)
	s.PresenceSensorAlwaysAccessible = isBit5Set(flag)

	s.ContaineredEntity1ID, _, _ = unpackUint8(data, 8)
	s.ContaineredEntity1Instance, _, _ = unpackUint8(data, 9)
	s.ContaineredEntity2ID, _, _ = unpackUint8(data, 10)
	s.ContaineredEntity2Instance, _, _ = unpackUint8(data, 11)
	s.ContaineredEntity3ID, _, _ = unpackUint8(data, 12)
	s.ContaineredEntity3Instance, _, _ = unpackUint8(data, 13)
	s.ContaineredEntity4ID, _, _ = unpackUint8(data, 14)
	s.ContaineredEntity4Instance, _, _ = unpackUint8(data, 15)

	return nil
}

// 43.5 SDR Type 09h - Device-relative Entity Association Record
type SDRDeviceRelative struct {
	//
	// Record KEY
	//

	ContainerEntityID            uint8
	ContainerEntityInstance      uint8
	ContainerEntityDeviceAddress uint8
	ContainerEntityDeviceChannel uint8

	// [7] - 0b = contained entities specified as list
	//       1b = contained entities specified as range
	ContainedEntitiesAsRange bool
	// [6] - Record Link
	//       0b = no linked Entity Association records
	//       1b = linked Entity Association records exist
	LinkedEntityAssiactionExist bool
	// [5] - 0b = Container entity and contained entities can be assumed absent
	//            if presence sensor for container entity cannot be accessed.
	//            This value is also used if the entity does not have a presence sensor.
	//       1b = Presence sensor should always be accessible. Software should consider
	//            it an error if the presence sensor associated with the container entity
	//            is not accessible. If a presence sensor is accessible, then the
	//            presence sensor can still report that the container entity is absent.
	PresenceSensorAlwaysAccessible bool

	ContaineredEntity1DeviceAddress uint8
	ContaineredEntity1DeviceChannel uint8
	ContaineredEntity1ID            uint8
	ContaineredEntity1Instance      uint8

	//
	// RECORD BODY
	//

	ContaineredEntity2DeviceAddress uint8
	ContaineredEntity2DeviceChannel uint8
	ContaineredEntity2ID            uint8
	ContaineredEntity2Instance      uint8

	ContaineredEntity3DeviceAddress uint8
	ContaineredEntity3DeviceChannel uint8
	ContaineredEntity3ID            uint8
	ContaineredEntity3Instance      uint8

	ContaineredEntity4DeviceAddress uint8
	ContaineredEntity4DeviceChannel uint8
	ContaineredEntity4ID            uint8
	ContaineredEntity4Instance      uint8
}

func parseSDRDeviceRelativeEntityAssociation(data []byte, sdr *SDR) error {
	const SDRDeviceRelativeEntityAssociationSize = 32
	if len(data) < SDRDeviceRelativeEntityAssociationSize {
		return fmt.Errorf("sdr (device-relative entity association) data must be longer than %d", SDRDeviceRelativeEntityAssociationSize)
	}

	s := &SDRDeviceRelative{}
	sdr.DeviceRelative = s

	s.ContainerEntityID, _, _ = unpackUint8(data, 5)
	s.ContainerEntityInstance, _, _ = unpackUint8(data, 6)
	s.ContainerEntityDeviceAddress, _, _ = unpackUint8(data, 7)
	s.ContainerEntityDeviceChannel, _, _ = unpackUint8(data, 8)

	flag, _, _ := unpackUint8(data, 9)
	s.ContainedEntitiesAsRange = isBit7Set(flag)
	s.LinkedEntityAssiactionExist = isBit6Set(flag)
	s.PresenceSensorAlwaysAccessible = isBit5Set(flag)

	s.ContaineredEntity1DeviceAddress, _, _ = unpackUint8(data, 10)
	s.ContaineredEntity1DeviceChannel, _, _ = unpackUint8(data, 11)
	s.ContaineredEntity1ID, _, _ = unpackUint8(data, 12)
	s.ContaineredEntity1Instance, _, _ = unpackUint8(data, 13)

	s.ContaineredEntity2DeviceAddress, _, _ = unpackUint8(data, 14)
	s.ContaineredEntity2DeviceChannel, _, _ = unpackUint8(data, 15)
	s.ContaineredEntity2ID, _, _ = unpackUint8(data, 16)
	s.ContaineredEntity2Instance, _, _ = unpackUint8(data, 17)

	s.ContaineredEntity3DeviceAddress, _, _ = unpackUint8(data, 18)
	s.ContaineredEntity3DeviceChannel, _, _ = unpackUint8(data, 19)
	s.ContaineredEntity3ID, _, _ = unpackUint8(data, 20)
	s.ContaineredEntity3Instance, _, _ = unpackUint8(data, 21)

	s.ContaineredEntity4DeviceAddress, _, _ = unpackUint8(data, 22)
	s.ContaineredEntity4DeviceChannel, _, _ = unpackUint8(data, 23)
	s.ContaineredEntity4ID, _, _ = unpackUint8(data, 24)
	s.ContaineredEntity4Instance, _, _ = unpackUint8(data, 25)

	unpackBytes(data, 26, 6) // last 6 bytes reserved
	return nil
}

// 43.7 SDR Type 10h - Generic Device Locator Record
// This record is used to store the location and type information for devices
// on the IPMB or management controller private busses that are neither
// IPMI FRU devices nor IPMI management controllers.
//
// These devices can either be common non-intelligent I2C devices, special management ASICs, or proprietary controllers.
//
// IPMI FRU Devices and Management Controllers are located via the FRU Device Locator
// and Management Controller Device Locator records described in following sections.
type SDRGenericDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceAccessAddress uint8 // Slave address of management controller used to access device. 0000000b if device is directly on IPMB
	DeviceSlaveAddress  uint8
	ChannelNumber       uint8 // Channel number for management controller used to access device
	AccessLUN           uint8 // LUN for Master Write-Read command. 00b if device is non-intelligent device directly on IPMB.
	PrivateBusID        uint8 // Private bus ID if bus = Private. 000b if device directly on IPMB

	//
	// RECORD BODY
	//

	AddressSpan        uint8
	DeviceType         uint8
	DeviceTypeModifier uint8
	EntityID           uint8
	EntityInstance     uint8

	DeviceIDTypeLength TypeLength
	DeviceIDString     []byte // Short ID string for the device
}

func parseSDRGenericLocator(data []byte, sdr *SDR) error {
	const SDRGenericLocatorMinSize = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	minSize := SDRGenericLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize)
	}

	s := &SDRGenericDeviceLocator{}
	sdr.GenericDeviceLocator = s

	s.DeviceAccessAddress, _, _ = unpackUint8(data, 5)

	b, _, _ := unpackUint8(data, 6)
	s.DeviceSlaveAddress = b

	c, _, _ := unpackUint8(data, 7)
	s.ChannelNumber = ((b & 0x01) << 4) | (c >> 5)
	s.AccessLUN = (c & 0x1f) >> 3
	s.PrivateBusID = (c & 0x07)

	s.AddressSpan, _, _ = unpackUint8(data, 8)
	s.DeviceType, _, _ = unpackUint8(data, 10)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	idStrLen := int(s.DeviceIDTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (generic-locator) data must be longer than %d", minSize+idStrLen)
	}
	s.DeviceIDString, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.8 SDR Type 11h - FRU Device Locator Record
// 38. Accessing FRU Devices
type SDRFRUDeviceLocator struct {
	//
	// Record KEY
	//

	// [7:1] - Slave address of controller used to access device. 0000000b if device is directly on IPMB.
	// This field indicates whether the device is on a private bus or not.
	DeviceAccessAddress uint8

	// FRU Device ID / Device Slave Address
	//
	// For LOGICAL FRU DEVICE (accessed via FRU commands to mgmt. controller):
	// [7:0] - Number identifying FRU device within given IPM Controller. FFh = reserved.
	// The primary FRU device for a management controller is always device #0 at
	// LUN 00b. The primary FRU device is not reported via this FRU Device Locator
	// record - its presence is identified via the Device Capabilities field in the
	// Management Controller Device Locator record.
	//
	// For non-intelligent FRU device:
	// [7:1] - 7-bit I2C Slave Address[1] . This is relative to the bus the device is on. For
	// devices on the IPMB, this is the slave address of the device on the IPMB. For
	// devices on a private bus, this is the slave address of the device on the private
	// bus.
	// [0] - reserved
	FRUDeviceID_SlaveAddress uint8

	// [7] - logical/physical FRU device
	// 0b = device is not a logical FRU Device
	// 1b = device is logical FRU Device (accessed via FRU commands to mgmt. controller)
	IsLogicalFRUDevice bool

	// [4:3] - LUN for Master Write-Read command or FRU Command. 00b if device is nonintelligent device directly on IPMB.
	AccessLUN uint8

	// [2:0] - Private bus ID if bus = Private.
	// 000b if device directly on IPMB, or device is a logical FRU Device.
	PrivateBusID uint8

	// [7:4] - Channel number for management controller used to access device. 000b if
	// device directly on the primary IPMB, or if controller is on the primary IPMB. Msbit for channel number is kept in next byte. (For IPMI v1.5. This byte position
	// was reserved for IPMI v1.0.)
	// [3:0] - reserved
	ChannelNumber uint8

	//
	// RECORD BODY
	//

	DeviceType         DeviceType
	DeviceTypeModifier uint8
	FRUEntityID        uint8
	FRUEntityInstance  uint8

	DeviceIDTypeLength TypeLength
	DeviceIDBytes      []byte // Short ID string for the FRU Device
}

// Table 38-1, FRU Device Locator Field Usage
func (sdrFRU *SDRFRUDeviceLocator) Location() FRULocation {
	if sdrFRU.DeviceAccessAddress == 0x00 {
		// Use MasterWriteRead command to access FRU.
		// FRUDeviceID_SlaveAddress (slave address Of SEEPROM on the IPMB)
		return FRULocation_IPMB
	}

	if sdrFRU.IsLogicalFRUDevice {
		// Use Read/WriteFRUData command to access FRU.
		// DeviceAccessAddress (Slave Address of IPMB)
		// FRUDeviceID_SlaveAddress
		return FRULocation_MgmtController
	}

	// Use MasterWriteRead command to access FRU.
	// DeviceAccessAddress (Slave Address of IPMB)
	// PrivateBusID
	// FRUDeviceID_SlaveAddress (Slave Address of SEEPROM on the Private Bus)
	return FRULocation_PrivateBus
}

func parseSDRFRUDeviceLocator(data []byte, sdr *SDR) error {
	const SDRFRUDeviceLocatorMinSize = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	minSize := SDRFRUDeviceLocatorMinSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize)
	}

	s := &SDRFRUDeviceLocator{}
	sdr.FRUDeviceLocator = s

	b5, _, _ := unpackUint8(data, 5)
	s.DeviceAccessAddress = b5

	b7, _, _ := unpackUint8(data, 6)
	// Todo
	s.FRUDeviceID_SlaveAddress = b7

	b8, _, _ := unpackUint8(data, 7)
	s.IsLogicalFRUDevice = isBit7Set(b8)
	s.AccessLUN = (b8 & 0x1f) >> 3
	s.PrivateBusID = b8 & 0x07

	b9, _, _ := unpackUint8(data, 8)
	s.ChannelNumber = b9 >> 4

	deviceType, _, _ := unpackUint8(data, 10)
	s.DeviceType = DeviceType(deviceType)
	s.DeviceTypeModifier, _, _ = unpackUint8(data, 11)

	s.FRUEntityID, _, _ = unpackUint8(data, 12)
	s.FRUEntityInstance, _, _ = unpackUint8(data, 13)

	// index 14 Reserved for OEM use.

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	var idStrLen int
	if s.DeviceIDTypeLength.TypeCode() == 0x00 {
		// unspecified type
		idStrLen = len(data) - minSize
	} else {
		idStrLen = int(s.DeviceIDTypeLength.Length())
	}
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (fru device) data must be longer than %d", minSize+idStrLen)
	}

	s.DeviceIDBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.9 SDR Type 12h - Management Controller Device Locator Record
type SDRMgmtControllerDeviceLocator struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8 // 7-bit I2C Slave Address[1] of device on channel
	ChannelNumber      uint8

	//
	// RECORD BODY
	//

	ACPISystemPowerStateNotificationRequired bool
	ACPIDevicePowerStateNotificationRequired bool
	ControllerLogsInitializationAgentErrors  bool
	LogInitializationAgentErrors             bool

	DeviceCap_ChassisDevice      bool // device functions as chassis device
	DeviceCap_Bridge             bool // Controller responds to Bridge NetFn command
	DeviceCap_IPMBEventGenerator bool // device generates event messages on IPMB
	DeviceCap_IPMBEventReceiver  bool // device accepts event messages from IPMB
	DeviceCap_FRUInventoryDevice bool // accepts FRU commands to FRU Device #0 at LUN 00b
	DeviceCap_SELDevice          bool // provides interface to SEL
	DeviceCap_SDRRepoDevice      bool // For BMC, indicates BMC provides interface to	1b = SDR Repository. For other controller, indicates controller accepts Device SDR commands
	DeviceCap_SensorDevice       bool // device accepts sensor commands

	EntityID       uint8
	EntityInstance uint8

	DeviceIDTypeLength TypeLength
	DeviceIDBytes      []byte
}

func parseSDRManagementControllerDeviceLocator(data []byte, sdr *SDR) error {
	const SDRManagementControllerDeviceLocatorMinSize = 16 // plus the ID String Bytes (optional 16 bytes maximum)
	minSize := SDRManagementControllerDeviceLocatorMinSize

	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerDeviceLocator{}
	sdr.MgmtControllerDeviceLocator = s

	b6, _, _ := unpackUint8(data, 5)
	s.DeviceSlaveAddress = b6

	b7, _, _ := unpackUint8(data, 6)
	s.ChannelNumber = b7

	b8, _, _ := unpackUint8(data, 7)
	s.ACPISystemPowerStateNotificationRequired = isBit7Set(b8)
	s.ACPIDevicePowerStateNotificationRequired = isBit6Set(b8)
	s.ControllerLogsInitializationAgentErrors = isBit3Set(b8)
	s.LogInitializationAgentErrors = isBit2Set(b8)

	b9, _, _ := unpackUint8(data, 8)
	s.DeviceCap_ChassisDevice = isBit7Set(b9)
	s.DeviceCap_Bridge = isBit6Set(b9)
	s.DeviceCap_IPMBEventGenerator = isBit5Set(b9)
	s.DeviceCap_IPMBEventReceiver = isBit4Set(b9)
	s.DeviceCap_FRUInventoryDevice = isBit3Set(b9)
	s.DeviceCap_SELDevice = isBit2Set(b9)
	s.DeviceCap_SDRRepoDevice = isBit1Set(b9)
	s.DeviceCap_SensorDevice = isBit0Set(b9)

	s.EntityID, _, _ = unpackUint8(data, 12)
	s.EntityInstance, _, _ = unpackUint8(data, 13)

	typeLength, _, _ := unpackUint8(data, 15)
	s.DeviceIDTypeLength = TypeLength(typeLength)

	idStrLen := int(s.DeviceIDTypeLength.Length())
	if len(data) < minSize+idStrLen {
		return fmt.Errorf("sdr (mgmt controller device locator) data must be longer than %d", minSize+idStrLen)
	}
	s.DeviceIDBytes, _, _ = unpackBytes(data, minSize, idStrLen)
	return nil
}

// 43.10 SDR Type 13h - Management Controller Confirmation Record
type SDRMgmtControllerConfirmation struct {
	//
	// Record KEY
	//

	DeviceSlaveAddress uint8 // 7-bit I2C Slave Address[1] of device on IPMB.
	DeviceID           uint8
	ChannelNumber      uint8
	DeviceRevision     uint8

	//
	// RECORD BODY
	//

	FirmwareMajorRevision uint8 // [6:0] - Major Firmware Revision, binary encoded.
	FirmwareMinorRevision uint8 // Minor Firmware Revision. BCD encoded.

	// IPMI Version from Get Device ID command. Holds IPMI Command Specification
	// Version. BCD encoded. 00h = reserved. Bits 7:4 hold the Least Significant digit of the
	// revision, while bits 3:0 hold the Most Significant bits. E.g. a value of 01h indicates
	// revision 1.0
	MajorIPMIVersion uint8
	MinorIPMIVersion uint8

	ManufacturerID uint32 // 3 bytes only
	ProductID      uint16
	DeviceGUID     []byte // 16 bytes
}

func parseSDRManagementControllerConfirmation(data []byte, sdr *SDR) error {
	const SDRManagementControllerConfirmationSize = 32
	minSize := SDRManagementControllerConfirmationSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (mgmt controller confirmation) data must be longer than %d", minSize)
	}

	s := &SDRMgmtControllerConfirmation{}
	sdr.MgmtControllerConfirmation = s

	b6, _, _ := unpackUint8(data, 5)
	s.DeviceSlaveAddress = b6

	s.DeviceID, _, _ = unpackUint8(data, 6)

	b8, _, _ := unpackUint8(data, 7)
	s.ChannelNumber = b8 >> 4
	s.DeviceRevision = b8 & 0x0f

	b9, _, _ := unpackUint8(data, 8)
	s.FirmwareMajorRevision = b9 & 0x7f

	s.FirmwareMinorRevision, _, _ = unpackUint8(data, 9)

	ipmiVersionBCD, _, _ := unpackUint8(data, 10)
	s.MajorIPMIVersion = ipmiVersionBCD & 0x0f
	s.MinorIPMIVersion = ipmiVersionBCD >> 4

	s.ManufacturerID, _, _ = unpackUint24L(data, 11)
	s.ProductID, _, _ = unpackUint16L(data, 14)
	s.DeviceGUID, _, _ = unpackBytes(data, 16, 16)
	return nil
}

// 43.11 SDR Type 14h - BMC Message Channel Info Record
type SDRBMCChannelInfo struct {
	//
	// NO Record KEY
	//

	//
	// RECORD BODY
	//

	Channel0 ChannelInfo
	Channel1 ChannelInfo
	Channel2 ChannelInfo
	Channel3 ChannelInfo
	Channel4 ChannelInfo
	Channel5 ChannelInfo
	Channel6 ChannelInfo
	Channel7 ChannelInfo

	MessagingInterruptType uint8

	EventMessageBufferInterruptType uint8
}

type ChannelInfo struct {
	TransmitSupported bool // false means  receive message queue access only
	MessageReceiveLUN uint8
	ChannelProtocol   uint8
}

func parseChannelInfo(b uint8) ChannelInfo {
	return ChannelInfo{
		TransmitSupported: isBit7Set(b),
		MessageReceiveLUN: (b & 0x7f) >> 4,
		ChannelProtocol:   b & 0x0f,
	}
}

func parseSDRBMCMessageChannelInfo(data []byte, sdr *SDR) error {
	const SDRBMCMessageChannelInfoSize = 16
	minSize := SDRBMCMessageChannelInfoSize
	if len(data) < minSize {
		return fmt.Errorf("sdr (bmc message channel info) data must be longer than %d", minSize)
	}

	s := &SDRBMCChannelInfo{}
	sdr.BMCChannelInfo = s

	s.Channel0 = parseChannelInfo(data[5])
	s.Channel1 = parseChannelInfo(data[6])
	s.Channel2 = parseChannelInfo(data[7])
	s.Channel3 = parseChannelInfo(data[8])
	s.Channel4 = parseChannelInfo(data[9])
	s.Channel5 = parseChannelInfo(data[10])
	s.Channel6 = parseChannelInfo(data[11])
	s.Channel7 = parseChannelInfo(data[12])

	s.MessagingInterruptType, _, _ = unpackUint8(data, 13)
	s.EventMessageBufferInterruptType, _, _ = unpackUint8(data, 14)
	return nil
}

// 43.12 SDR Type C0h - OEM Record
type SDROEM struct {
	//
	// NO Record KEY
	//

	//
	// RECORD BODY
	//

	ManufacturerID uint32 // 3 bytes only
	OEMData        []byte
}

func parseSDROEM(data []byte, sdr *SDR) error {
	const SDROEMMinSize = 8
	const SDROEMMaxSize = 64 // OEM defined records are limited to a maximum of 64 bytes, including the header

	if len(data) < SDROEMMinSize {
		return fmt.Errorf("sdr (bmc message channel info) data must be longer than %d", SDROEMMinSize)
	}

	s := &SDROEM{}
	sdr.OEM = s

	s.ManufacturerID, _, _ = unpackUint24L(data, 5)
	s.OEMData, _, _ = unpackBytesMost(data, 8, SDROEMMaxSize-8)
	return nil
}

// 43.6 SDR Type 0Ah:0Fh - Reserved Records
type SDRReserved struct {
}

// 43.15 Type/Length Byte Format
//
//  7:6 00 = Unicode
//           00b define a Unicode string in the IPMI specification,
//           whereas they specify a binary field in the Platform Management FRU specification.
//      01 = BCD plus (see below)
//      10 = 6-bit ASCII, packed
//      11 = 8-bit ASCII + Latin 1.
//          At least two bytes of data must be present when this type is used.
//          Therefore, the length (number of data bytes) will be >1 if data is present,
//          0 if data is not present. A length of 1 is reserved.
//  5 reserved.
//      the bit 5 is reserved in the IPMI specification type/length byte,
//      where it is part of the length field in the Platform Management FRU specification
//  4:0 length of following data, in characters.
//      00000b indicates 'none following'.
//      11111b = reserved.
type TypeLength uint8

func (tl TypeLength) String() string {
	return fmt.Sprintf("Byte: (%#02x) / Type: (%s) / Length: %d / Size: %d / ", uint8(tl), tl.Type(), tl.Length(), tl.Size())
}

func (tl TypeLength) Type() string {
	typecode := tl.TypeCode()

	var s string
	switch typecode {
	case 0:
		s = "Binary"
	case 1:
		s = "BCD plus"
	case 2:
		s = "6-bit ASCII"
	case 3:
		s = "8-bit ASCII"
	}

	return s
}

func (tl TypeLength) TypeCode() uint8 {
	return (uint8(tl) & 0xc0) >> 6 // the highest 2 bits
}

// Length returns the length of bytes occupied that packed the chars.
// But it is not the length of chars.
// For BCD plus type, one byte packs two chars.
func (tl TypeLength) Length() uint8 {
	return uint8(tl) & 0x3f // the lowest 6 bits
}

// Size returns the length of chars.
func (tl TypeLength) Size() uint8 {
	typecode := tl.TypeCode()
	l := tl.Length()

	var size uint8
	switch typecode {
	case 0: /* 00b: binary data */
		size = l
	case 1: /* 01b: BCD plus (binary-coded decimal) */
		// one byte packs two chars
		/* hex dump or BCD -> 2x length */
		size = l * 2
	case 2: /* 10b: 6-bit ASCII packed */
		// three bytes packs four chars
		/* 4 chars per group of 1-3 bytes, round up to 4 bytes boundary */
		size = (l/3 + 1) * 4
	case 3: /* 11b: 8-bit ASCII + Latin 1 */
		/* no length adjustment */
		size = l
	}

	return size
}

// Chars decodes the raw bytes to ASCII chars according to the encoding type code of TypeLength
func (tl TypeLength) Chars(raw []byte) (chars []byte, err error) {
	if len(raw) != int(tl.Length()) {
		err = fmt.Errorf("passed raw not equal to length")
		return
	}

	size := int(tl.Size())
	chars = make([]byte, size)

	switch tl.TypeCode() {
	case 0: // 00b - Binary
		for i := 0; i < size; i++ {
			chars[i] = raw[i]
		}

	case 1: // 01b - BCD Plus
		var bcdPlusChars = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ' ', '-', '.', ':', ',', '_'}

		for i := 0; i < size; i++ {
			var charIndex uint8
			if i%2 == 0 {
				charIndex = raw[i/2] >> 0 & 0x0f
			} else {
				charIndex = raw[i/2] >> 4
			}
			chars[i] = bcdPlusChars[charIndex]
		}

	case 2: // 10b - 6-bit ASCII
		// 6-bit ASCII definition
		var ascci6bit = [64]byte{
			' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
			'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
			'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
		}

		var leftover byte
		var s []byte

		for i := 0; i < len(raw); i++ {
			// every 3 bytes pack 4 chars, so we can calculate
			// character positions in a byte based on the remainder of division by 3.
			switch i % 3 {
			case 0:
				idx := raw[i] & 0x3f            // 6 right bits are an index of one char
				leftover = (raw[i] & 0xc0) >> 6 // 2 left bits are leftovers

				s = append(s, ascci6bit[idx])
			case 1:
				idx := leftover | (raw[i]&0x0f)<<2 // index of one char is 2-bit leftover as prefix plus 4 right bits
				leftover = (raw[i] & 0xf0) >> 4    // 4 left bits are leftovers

				s = append(s, ascci6bit[idx])
			case 2:
				idx := (raw[i]&0x03)<<4 | leftover // index of one char is 2 right bits plus 4-bit leftover as suffix
				leftover = 0                       // cleanup leftover calculation

				s = append(s, ascci6bit[idx])

				idx = (raw[i] & 0xfc) >> 2 // 6 left bits are an index of one char
				s = append(s, ascci6bit[idx])
			}
		}

		chars = s

	case 3: // 11b - 8-bit ASCII
		for i := 0; i < size; i++ {
			chars[i] = raw[i]
		}
	}

	return
}
