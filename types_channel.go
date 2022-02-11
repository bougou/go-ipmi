package ipmi

// 6.3 Channel Numbers
// Only the channel number assignments for the primary IPMB and the System Interface are fixed,
// the assignment of other channel numbers can vary on a per-platform basis
type Channel uint8

const (
	ChannelPrimaryIPMB Channel = 0x0
	ChannelSystem      Channel = 0xf
)

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
	ChannelMediumIPMB            ChannelMedium = 0x01
	ChannelMediumICMBv10         ChannelMedium = 0x02
	ChannelMediumICMBv09         ChannelMedium = 0x03
	ChannelMediumLAN             ChannelMedium = 0x04
	ChannelMediumSerial          ChannelMedium = 0x05
	ChannelMediumOtherLAN        ChannelMedium = 0x06
	ChannelMediumSMBus           ChannelMedium = 0x07
	ChannelMediumSMBusv10        ChannelMedium = 0x08
	ChannelMediumSMBusv20        ChannelMedium = 0x09
	ChannelMediumUSBv1           ChannelMedium = 0x0a
	ChannelMediumUSBv2           ChannelMedium = 0x0b
	ChannelMediumSystemInterface ChannelMedium = 0x0c
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

// 6.8 Channel Privilege Levels
type PrivilegeLevel uint8

const (
	PrivilegeLevelUnspecified   PrivilegeLevel = 0x00
	PrivilegeLevelCallback      PrivilegeLevel = 0x01
	PrivilegeLevelUser          PrivilegeLevel = 0x02
	PrivilegeLevelOperator      PrivilegeLevel = 0x03
	PrivilegeLevelAdministrator PrivilegeLevel = 0x04
	PrivilegeLevelOEM           PrivilegeLevel = 0x05
)

func (l PrivilegeLevel) Short() string {
	// :     X=Cipher Suite Unused
	// :     c=CALLBACK
	// :     u=USER
	// :     o=OPERATOR
	// :     a=ADMIN
	// :     O=OEM
	m := map[PrivilegeLevel]string{
		0x00: "X",
		0x01: "c",
		0x02: "u",
		0x03: "o",
		0x04: "a",
		0x05: "O",
	}
	s, ok := m[l]
	if ok {
		return s
	}
	return "-"
}

func (l PrivilegeLevel) String() string {
	m := map[PrivilegeLevel]string{
		0x00: "Unspecified",
		0x01: "CALLBACK",
		0x02: "USER",
		0x03: "OPERATOR",
		0x04: "ADMINISTRATOR",
		0x05: "OEM",
	}
	s, ok := m[l]
	if ok {
		return s
	}
	return "NO ACCESS"
}

// see: Table 22-28, Get Channel Access Command

type ChannelAccessOption uint8

const (
	ChannelAccessOption_NoChange    ChannelAccessOption = 0
	ChannelAccessOption_NonVolatile ChannelAccessOption = 1
	ChannelAccessOption_Volatile    ChannelAccessOption = 2
)

// 6.6 Channel Access Modes
type ChannelAccessMode uint8

const (
	ChannelAccessMode_Disabled        ChannelAccessMode = 0
	ChannelAccessMode_PrebootOnly     ChannelAccessMode = 1
	ChannelAccessMode_AlwaysAvailable ChannelAccessMode = 2
	ChannelAccessMode_Shared          ChannelAccessMode = 3
)

func (mode ChannelAccessMode) String() string {
	m := map[ChannelAccessMode]string{
		0: "disabled",
		1: "pre-boot only",
		2: "always available",
		3: "shared",
	}
	s, ok := m[mode]
	if ok {
		return s
	}
	return ""
}

type ChannelPrivilegeOption uint8

const (
	ChanenlPrivilegeOption_NoChange    ChannelPrivilegeOption = 0
	ChannelPrivilegeOption_NonVolatile ChannelPrivilegeOption = 1
	ChannelPrivilegeOption_Volatile    ChannelPrivilegeOption = 2
)
