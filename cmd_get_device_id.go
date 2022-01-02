package ipmi

import "fmt"

// 20.1
type GetDeviceIDRequest struct {
	// empty
}

type GetDeviceIDResponse struct {
	DeviceID uint8

	// [7] 1 = device provides Device SDRs
	// 0 = device does not provide Device SDRs
	// [6:4] reserved. Return as 0.
	// [3:0] Device Revision, binary encoded
	DeviceProvideSDRs bool
	DeviceRevision    uint8

	// [7] Device available: 0=normal operation, 1= device firmware, SDR
	// Repository update or self-initialization in progress. [Firmware / SDR
	// Repository updates can be differentiated by issuing a Get SDR
	// command and checking the completion code.]
	// [6:0] Major Firmware Revision, binary encoded
	DeviceAvailable       bool
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

func (req *GetDeviceIDRequest) Command() Command {
	return CommandGetDeviceID
}

func (req *GetDeviceIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetDeviceIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDeviceIDResponse) Unpack(msg []byte) error {
	if len(msg) < 11 {
		return ErrUnpackedDataTooShort
	}

	res.DeviceID, _, _ = unpackUint8(msg, 0)
	b2, _, _ := unpackUint8(msg, 1)
	res.DeviceProvideSDRs = isBit7Set(b2)
	res.DeviceRevision = b2 & 0x0f

	b3, _, _ := unpackUint8(msg, 2)
	res.DeviceAvailable = isBit7Set(b3)
	res.MajorFirmwareRevision = b3 & 0x3f // binary encoded

	res.MinorFirmwareRevision, _, _ = unpackUint8(msg, 3) // BCD encoded
	res.IPMIVersion, _, _ = unpackUint8(msg, 4)           // BCD encoded

	b6, _, _ := unpackUint8(msg, 5) // BCD encoded

	res.SupportChassis = isBit7Set(b6)
	res.SupportBridge = isBit6Set(b6)
	res.SupportIPMBEventGenerator = isBit5Set(b6)
	res.SupportIPMBEventReceiver = isBit4Set(b6)
	res.SupportFRUInventory = isBit3Set(b6)
	res.SupportSEL = isBit2Set(b6)
	res.SupportSDRRepo = isBit1Set(b6)
	res.SupportSensor = isBit0Set(b6)

	res.ManufacturerID, _, _ = unpackUint24L(msg, 6)
	res.ProductID, _, _ = unpackUint16L(msg, 9)

	if len(msg) > 11 && len(msg) < 15 {
		return ErrUnpackedDataTooShort
	} else {
		res.AuxiliaryFirmwareRevision, _, _ = unpackUint32L(msg, 11)
	}
	return nil
}

func (res *GetDeviceIDResponse) Format() string {
	return fmt.Sprintf(`Device ID                 : %d
Device Revision           : %d
Firmware Revision         : %d.%d
IPMI Version              : %d
Manufacturer ID           : %d
Manufacturer Name         :
Product ID                : %d (%#02x)
Product Name              :
Device Available          : %s
Provides Device SDRs      : %s
Additional Device Support :
Aux Firmware Rev Info     :`,
		res.DeviceID,
		res.DeviceRevision,
		res.MajorFirmwareRevision, res.MinorFirmwareRevision,
		res.IPMIVersion,
		res.ManufacturerID,
		res.ProductID, res.ProductID,
		formatBool(res.DeviceAvailable, "yes", "no"),
		formatBool(res.DeviceProvideSDRs, "yes", "no"),
	)
}

func (c *Client) GetDeviceID() (response *GetDeviceIDResponse, err error) {
	request := &GetDeviceIDRequest{}
	response = &GetDeviceIDResponse{}
	err = c.Exchange(request, response)
	return
}
