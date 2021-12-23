package ipmi

// 20.1
type GetDeviceIDRequest struct {
}

func (r *GetDeviceIDRequest) Pack() []byte {
	return nil
}

type GetDeviceResponse struct {
	CompletionCode

	DeviceID uint8

	// [7] 1 = device provides Device SDRs
	// 0 = device does not provide Device SDRs
	// [6:4] reserved. Return as 0.
	// [3:0] Device Revision, binary encoded
	ProvideDeviceSDRs bool
	DeviceRevision    uint8

	// [7] Device available: 0=normal operation, 1= device firmware, SDR
	// Repository update or self-initialization in progress. [Firmware / SDR
	// Repository updates can be differentiated by issuing a Get SDR
	// command and checking the completion code.]
	// [6:0] Major Firmware Revision, binary encoded
	MajorFirmwareRevision uint8

	// BCD encoded
	MinorFirmwareRevision uint8

	// Holds IPMI Command Specification Version. BCD encoded.
	// 00h = reserved.
	// Bits 7:4 hold the Least Significant digit of the revision, while
	// bits 3:0 hold the Most Significant bits.
	// E.g. a value of 51h indicates revision 1.5 functionality.
	// 02h for implementations that provide IPMI v2.0 capabilities
	// per this specification.
	IPMIVersion uint8

	// Additional Device Support (formerly called IPM Device Support). Lists the
	// IPMI 'logical device' commands and functions that the controller supports that
	// are in addition to the mandatory IPM and Application commands.
	// [7] Chassis Device (device functions as chassis device per ICMB spec.)
	// [6] Bridge (device responds to Bridge NetFn commands)
	// [5] IPMB Event Generator (device generates event messages [platform
	// event request messages] onto the IPMB)
	// [4] IPMB Event Receiver (device accepts event messages [platform event
	// request messages] from the IPMB)
	// [3] FRU Inventory Device
	// [2] SEL Device
	// [1] SDR Repository Device
	// [0] Sensor Device
	SupportChassis            bool
	SupportBridge             bool
	SupportIPMBEventGenerator bool
	SupportIPMBEventReceiver  bool
	SupportFRUInventory       bool
	SupportSEL                bool
	SupportSDRRepo            bool
	SupportSensor             bool

	// Manufacturer ID, LS Byte first. The manufacturer ID is a 20-bit value that is
	// derived from the IANA Private Enterprise ID (see below).
	// Most significant four bits = reserved (0000b).
	// 000000h = unspecified. 0FFFFFh = reserved. This value is binary encoded.
	// E.g. the ID for the IPMI forum is 7154 decimal, which is 1BF2h, which would
	// be stored in this record as F2h, 1Bh, 00h for bytes 8 through 10, respectively
	ManufacturerID uint32 // only 3 bytes used

	// Product ID, LS Byte first. This field can be used to provide a number that
	// identifies a particular system, module, add-in card, or board set. The number
	// is specified according to the manufacturer given by Manufacturer ID (see
	// below).
	// 0000h = unspecified. FFFFh = reserved.
	ProductID uint16

	// Auxiliary Firmware Revision Information. This field is optional. If present, it
	// holds additional information about the firmware revision, such as boot block or
	// internal data structure version numbers. The meanings of the numbers are
	// specific to the vendor identified by Manufacturer ID (see below). When the
	// vendor-specific definition is not known, generic utilities should display each
	// byte as 2-digit hexadecimal numbers, with byte 13 displayed first as the mostsignificant byte.
	AuxiliaryFirmwareRevision uint32
}
