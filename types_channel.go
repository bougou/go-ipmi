package ipmi

type Channel uint8

// 6.4
type ChannelProtocol uint8

const (
	ChannelProtocolIPMB  = 0x01
	ChannelProtocolICMB  = 0x02 // 03 reserved
	ChannelProtocolSMBus = 0x04
	ChannelProtocolKCS   = 0x05
	ChannelProtocolSMIC  = 0x06
	ChannelProtocolBTv10 = 0x07
	ChannelProtocolBTv15 = 0x08
	ChannelProtocolTMode = 0x09
)

// 6.5
type ChannelMedium uint8

const (
	ChannelMediumIPMB     = 0x01
	ChannelMediumICMBv10  = 0x02
	ChannelMediumICMBv09  = 0x03
	ChannelMediumLAN      = 0x04
	ChannelMediumSerial   = 0x05
	ChannelMediumOtherLAN = 0x06
	ChannelMediumSMBus    = 0x07
	ChannelMediumSMBusv10 = 0x08
	ChannelMediumSMBusv20 = 0x09
	ChannelMediumUSBv1    = 0x0a
	ChannelMediumUSBv2    = 0x0b
	ChannelMediumSystem   = 0x0c
)

// 6.6 Channel Access Modes

const (
	ChannelAccessModePrebootOnly = "pre-boot only"
	ChannelAccessModeAlways      = "always available"
	ChannelAccessModeShared      = "shared"
	ChannelAccessModeDisabled    = "disabled"
)

// 6.8 Channel Privilege Levels
type PrivilegeLevel uint8

const (
	PrivilegeLevelAutoDetect    PrivilegeLevel = 0x00
	PrivilegeLevelCallback      PrivilegeLevel = 0x01
	PrivilegeLevelUser          PrivilegeLevel = 0x02
	PrivilegeLevelOperator      PrivilegeLevel = 0x03
	PrivilegeLevelAdministrator PrivilegeLevel = 0x04
	PrivilegeLevelOEM           PrivilegeLevel = 0x05
)
