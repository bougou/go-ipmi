package ipmi

type Channel uint8

// 6.4 Channel Protocol Type
type ChannelProtocol uint8

const (
	ChannelProtocolIPMB  ChannelProtocol = 0x01
	ChannelProtocolICMB  ChannelProtocol = 0x02 // 03 reserved
	ChannelProtocolSMBus ChannelProtocol = 0x04
	ChannelProtocolKCS   ChannelProtocol = 0x05
	ChannelProtocolSMIC  ChannelProtocol = 0x06
	ChannelProtocolBTv10 ChannelProtocol = 0x07
	ChannelProtocolBTv15 ChannelProtocol = 0x08
	ChannelProtocolTMode ChannelProtocol = 0x09
	ChannelProtocolOEM1  ChannelProtocol = 0x1c
	ChannelProtocolOEM2  ChannelProtocol = 0x1d
	ChannelProtocolOEM3  ChannelProtocol = 0x1e
	ChannelProtocolOEM4  ChannelProtocol = 0x1f
)

func (cp ChannelProtocol) String() string {
	m := map[ChannelProtocol]string{
		0x01: "IPMB-1.0",
		0x02: "ICMB-1.0",
		0x04: "IPMI-SMBus",
		0x05: "KCS",
		0x06: "SMIC",
		0x07: "BT-10",
		0x08: "BT-15",
		0x09: "TMode",
		0x1c: "OEM Protocol 1",
		0x1d: "OEM Protocol 2",
		0x1e: "OEM Protocol 3",
		0x1f: "OEM Protocol 4",
	}
	s, ok := m[cp]
	if ok {
		return s
	}
	return "reserved"
}

// 6.5 Channel Medium Type
type ChannelMedium uint8

const (
	ChannelMediumIPMB     ChannelMedium = 0x01
	ChannelMediumICMBv10  ChannelMedium = 0x02
	ChannelMediumICMBv09  ChannelMedium = 0x03
	ChannelMediumLAN      ChannelMedium = 0x04
	ChannelMediumSerial   ChannelMedium = 0x05
	ChannelMediumOtherLAN ChannelMedium = 0x06
	ChannelMediumSMBus    ChannelMedium = 0x07
	ChannelMediumSMBusv10 ChannelMedium = 0x08
	ChannelMediumSMBusv20 ChannelMedium = 0x09
	ChannelMediumUSBv1    ChannelMedium = 0x0a
	ChannelMediumUSBv2    ChannelMedium = 0x0b
	ChannelMediumSystem   ChannelMedium = 0x0c
)

func (cp ChannelMedium) String() string {
	m := map[ChannelMedium]string{
		0x01: "IPMB (I2C)",
		0x02: "ICMB v1.0",
		0x03: "ICMB v0.9",
		0x04: "802.3 LAN",
		0x05: "Asynch. Serial/Modem (RS-232)",
		0x06: "Other LAN",
		0x07: "PCI SMBus",
		0x08: "SMBus v1.0/1.1",
		0x09: "SMBus v2.0",
		0x0a: "USB 1.x",
		0x0b: "USB 2.x",
		0x0c: "System Interface (KCS, SMIC, or BT)",
	}
	s, ok := m[cp]
	if ok {
		return s
	}
	if cp >= 0x60 && cp <= 0x7f {
		return "OEM"
	}
	return "reserved"
}

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

func (l PrivilegeLevel) String() string {
	m := map[PrivilegeLevel]string{
		0x01: "callback",
		0x02: "user",
		0x03: "operator",
		0x04: "admin",
		0x05: "Oem",
	}
	s, ok := m[l]
	if ok {
		return s
	}
	return ""
}
