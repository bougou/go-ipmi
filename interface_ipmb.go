package ipmi

type LUN uint8

// 7.2 BMC IPMB LUNs
const (
	IPMB_LUN_BMC   LUN = 0x00 // BMC commands and Event Request Messages
	IPMB_LUN_OEM_1 LUN = 0x01 // OEM LUN 1
	IPMB_LUN_SMS   LUN = 0x10 // SMS Message LUN (Intended for messages to System Management Software)
	IPMB_LUN_OEM_2 LUN = 0x11 // OEM LUN 2

	// the least significat bit
	// 0b (ID is a slave address)
	// 1b (ID is a Software ID)
	BMC_SA             uint8 = 0x20 // BMC's responder address
	RemoteConsole_SWID uint8 = 0x81 // Remote Console Software ID
)

type IPMBRequest struct {
	NetFn

	LUN

	Command

	Channel
}
