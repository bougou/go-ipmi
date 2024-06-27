package ipmi

import (
	"fmt"
	"net"
	"strings"
)

// Table 23-4, LAN Configuration Parameters
// Parameter selector
type LanParamSelector uint8

const (
	LanParam_SetInProgress                    LanParamSelector = 0
	LanParam_AuthTypeSupported                LanParamSelector = 1
	LanParam_AuthTypeEnables                  LanParamSelector = 2
	LanParam_IP                               LanParamSelector = 3
	LanParam_IPSource                         LanParamSelector = 4
	LanParam_MAC                              LanParamSelector = 5 // read only
	LanParam_SubnetMask                       LanParamSelector = 6
	LanParam_IPv4HeaderParams                 LanParamSelector = 7
	LanParam_PrimaryRMCPPort                  LanParamSelector = 8
	LanParam_SecondaryRMCPPort                LanParamSelector = 9
	LanParam_ARPControl                       LanParamSelector = 10
	LanParam_GratuitousARPInterval            LanParamSelector = 11
	LanParam_DefaultGatewayIP                 LanParamSelector = 12
	LanParam_DefaultGatewayMAC                LanParamSelector = 13
	LanParam_BackupGatewayIP                  LanParamSelector = 14
	LanParam_BackupGatewayMAC                 LanParamSelector = 15
	LanParam_CommunityString                  LanParamSelector = 16
	LanParam_AlertDestinationsNumber          LanParamSelector = 17
	LanParam_AlertDestinationType             LanParamSelector = 18
	LanParam_AlertDestinationAddress          LanParamSelector = 19
	LanParam_VLANID                           LanParamSelector = 20
	LanParam_VLANPriority                     LanParamSelector = 21
	LanParam_CipherSuiteEntrySupport          LanParamSelector = 22
	LanParam_CipherSuiteEntries               LanParamSelector = 23
	LanParam_CipherSuitePrivilegeLevels       LanParamSelector = 24
	LanParam_AlertDestinationVLAN             LanParamSelector = 25 // read only
	LanParam_BadPasswordThreshold             LanParamSelector = 26
	LanParam_IP6Support                       LanParamSelector = 50
	LanParam_IP6Enables                       LanParamSelector = 51
	LanParam_IP6StaticTrafficClass            LanParamSelector = 52
	LanParam_IP6StaticHopLimit                LanParamSelector = 53
	LanParam_IP6FlowLabel                     LanParamSelector = 54
	LanParam_IP6Status                        LanParamSelector = 55
	LanParam_IP6StaticAddr                    LanParamSelector = 56
	LanParam_IP6DHCP6StaticDUIDLength         LanParamSelector = 57 // DHCPv6 uses DHCP Unique Identifier (DUID) to identify clients (and also clients identify the DHCPv6 server by its DUID)
	LanParam_IP6DHCP6StaticDUIDs              LanParamSelector = 58
	LanParam_IP6DynamicAddr                   LanParamSelector = 59
	LanParam_IP6DHCP6DynamicDUIDLength        LanParamSelector = 60
	LanParam_IP6DHCP6DynamicDUIDs             LanParamSelector = 61
	LanParam_IP6DHCP6TimingConfigSupport      LanParamSelector = 62
	LanParam_IP6DHCP6TimingConfig             LanParamSelector = 63
	LanParam_IP6RouterAddressConfigControl    LanParamSelector = 64
	LanParam_IP6StaticRouter1IP               LanParamSelector = 65
	LanParam_IP6StaticRouter1MAC              LanParamSelector = 66
	LanParam_IP6StaticRouter1PrefixLength     LanParamSelector = 67
	LanParam_IP6StaticRouter1PrefixValue      LanParamSelector = 68
	LanParam_IP6StaticRouter2IP               LanParamSelector = 69
	LanParam_IP6StaticRouter2MAC              LanParamSelector = 70
	LanParam_IP6StaticRouter2PrefixLength     LanParamSelector = 71
	LanParam_IP6StaticRouter2PrefixValue      LanParamSelector = 72
	LanParam_IP6DynamicRouterSetsNumber       LanParamSelector = 73 // read only
	LanParam_IP6DynamicRouterIP               LanParamSelector = 74
	LanParam_IP6DynamicRouterMAC              LanParamSelector = 75
	LanParam_IP6DynamicRouterPrefixLength     LanParamSelector = 76
	LanParam_IP6DynamicRouterPrefixValue      LanParamSelector = 77
	LanParam_IP6DynamicRouterReceivedHopLimit LanParamSelector = 78 // read only
	LanParam_IP6NDSLAACTimingConfigSupport    LanParamSelector = 79 // read only, IPv6 Neighbor	Discovery / SLAAC
	LanParam_IP6NDSLAACTiming                 LanParamSelector = 80

	// OEM Parameters 192:
	// 255
	// This range is available for special OEM configuration parameters. The OEM is identified
	// according to the Manufacturer ID field returned by the Get Device ID command.

)

var LanParams = []LanParam{
	{Selector: LanParam_SetInProgress, DataSize: 1, Name: "Set in Progress"},
	{Selector: LanParam_AuthTypeSupported, DataSize: 1, Name: "Auth Type Support"},
	{Selector: LanParam_AuthTypeEnables, DataSize: 5, Name: "Auth Type Enable"},
	{Selector: LanParam_IP, DataSize: 4, Name: "IP Address"},
	{Selector: LanParam_IPSource, DataSize: 1, Name: "IP Address Source"},
	{Selector: LanParam_MAC, DataSize: 6, Name: "MAC Address"},
	{Selector: LanParam_SubnetMask, DataSize: 4, Name: "Subnet Mask"},
	{Selector: LanParam_IPv4HeaderParams, DataSize: 3, Name: "IP Header"},
	{Selector: LanParam_PrimaryRMCPPort, DataSize: 2, Name: "Primary RMCP Port"},
	{Selector: LanParam_SecondaryRMCPPort, DataSize: 2, Name: "Secondary RMCP Port"},
	{Selector: LanParam_ARPControl, DataSize: 1, Name: "BMC ARP Control"},
	{Selector: LanParam_GratuitousARPInterval, DataSize: 1, Name: "Gratuitous ARP Interval"},
	{Selector: LanParam_DefaultGatewayIP, DataSize: 4, Name: "Default Gateway IP"},
	{Selector: LanParam_DefaultGatewayMAC, DataSize: 6, Name: "Default Gateway MAC"},
	{Selector: LanParam_BackupGatewayIP, DataSize: 4, Name: "Backup Gateway IP"},
	{Selector: LanParam_BackupGatewayMAC, DataSize: 6, Name: "Backup Gateway MAC"},
	{Selector: LanParam_CommunityString, DataSize: 18, Name: "SNMP Community String"},
	{Selector: LanParam_AlertDestinationsNumber, DataSize: 1, Name: "Number of Destinations"},
	{Selector: LanParam_AlertDestinationType, DataSize: 4, Name: "Destination Type"},
	{Selector: LanParam_AlertDestinationAddress, DataSize: 18, Name: "Destination Addresses"},
	{Selector: LanParam_VLANID, DataSize: 2, Name: "802.1q VLAN ID"},
	{Selector: LanParam_VLANPriority, DataSize: 1, Name: "802.1q VLAN Priority"},
	{Selector: LanParam_CipherSuiteEntrySupport, DataSize: 1, Name: "RMCP+ Cipher Suite Count"},
	{Selector: LanParam_CipherSuiteEntries, DataSize: 17, Name: "RMCP+ Cipher Suites"},
	{Selector: LanParam_CipherSuitePrivilegeLevels, DataSize: 9, Name: "Cipher Suite Priv Max"},
	{Selector: LanParam_BadPasswordThreshold, DataSize: 4, Name: "Bad Password Threshold"},
}

func (lanParam LanParamSelector) String() string {
	for _, v := range LanParams {
		if v.Selector == lanParam {
			return v.Name
		}
	}
	return ""
}

type LanParam struct {
	Selector LanParamSelector
	DataSize uint8
	Name     string
}

type LanConfig struct {
	SetInProgress                 SetInProgress
	AuthTypeSupport               AuthTypeSupport
	AuthTypeEnables               AuthTypeEnables
	IP                            net.IP
	IPSource                      IPAddressSource
	MAC                           net.HardwareAddr
	SubnetMask                    net.IP
	IPHeaderParams                IPHeaderParams
	PrimaryRMCPPort               uint16
	SecondaryRMCPPort             uint16
	ARPControl                    ARPControl
	GratuitousARPIntervalMilliSec int32
	DefaultGatewayIP              net.IP
	DefaultGatewayMAC             net.HardwareAddr
	BackupGatewayIP               net.IP
	BackupGatewayMAC              net.HardwareAddr
	CommunityString               CommunityString
	AlertDestinationsNumber       uint8
	AlertDestinationType          AlertDestinationType
	AlertDestinationAddress       AlertDestinationAddress
	VLANEnabled                   bool
	VLANID                        uint16
	VLANPriority                  uint8
	RMCPCipherSuitesCount         uint8
	RMCPCipherSuiteEntries        []CipherSuiteID
	RMCPCipherSuitesMaxPrivLevel  []PrivilegeLevel
	AlertDestinationVLAN          AlertDestinationVLAN
	BadPasswordThreshold          BadPasswordThreshold

	IP6Support IP6Support
}

func (lanConfig *LanConfig) Format() string {
	cipherSuites := []string{}
	for _, v := range lanConfig.RMCPCipherSuiteEntries {
		cipherSuites = append(cipherSuites, fmt.Sprintf("%d", v))
	}
	cipherSuitesStr := strings.Join(cipherSuites, ",")

	levels := []string{}
	for _, v := range lanConfig.RMCPCipherSuitesMaxPrivLevel {
		levels = append(levels, v.Short())
	}
	levelsStr := strings.Join(levels, "")

	return fmt.Sprintf(`
Set in Progress         : %s
IP Address Source       : %s
IP Address              : %s
Subnet Mask             : %s
MAC Address             : %s
SNMP Community String   : %s
IP Header               : TTL=%#02x Flags=%#02x Precedence=%#02x TOS=%#02x
Default Gateway IP      : %s
802.1q VLAN ID          : %d
RMCP+ Cipher Suites     : %s
Cipher Suite Priv Max   : %s
                        :     X=Cipher Suite Unused
                        :     c=CALLBACK
                        :     u=USER
                        :     o=OPERATOR
                        :     a=ADMIN
                        :     O=OEM
Bad Password Threshold  : %d`,
		lanConfig.SetInProgress,
		lanConfig.IPSource,
		lanConfig.IP,
		lanConfig.SubnetMask,
		lanConfig.MAC,
		lanConfig.CommunityString,
		lanConfig.IPHeaderParams.TTL, lanConfig.IPHeaderParams.Flags, lanConfig.IPHeaderParams.Precedence, lanConfig.IPHeaderParams.TOS,
		lanConfig.DefaultGatewayIP,
		lanConfig.VLANID,
		cipherSuitesStr,
		levelsStr,
		lanConfig.BadPasswordThreshold.Threshold,
	)
}

type SetInProgress uint8

func (p SetInProgress) String() string {
	m := map[SetInProgress]string{
		0x00: "set complete",
		0x01: "set in progress",
		0x02: "commit write",
		0x03: "reserved",
	}
	s, ok := m[p]
	if ok {
		return s
	}
	return ""
}

type AuthTypeSupport struct {
	OEM      bool
	Password bool
	MD5      bool
	MD2      bool
	None     bool
}

type AuthTypeEnabled struct {
	OEM      bool
	Password bool
	MD5      bool
	MD2      bool
	None     bool
}

type AuthTypeEnables struct {
	Callback AuthTypeEnabled
	User     AuthTypeEnabled
	Operator AuthTypeEnabled
	Admin    AuthTypeEnabled
	OEM      AuthTypeEnabled
}

type IPAddressSource uint8

const (
	IPAddressSourceUnspecified IPAddressSource = 0x00
	IPAddressSourceStatic      IPAddressSource = 0x01
	IPAddressSourceDHCP        IPAddressSource = 0x02
	IPAddressSourceBIOS        IPAddressSource = 0x03
	IPAddressSourceOther       IPAddressSource = 0x04
)

func (i IPAddressSource) String() string {
	m := map[IPAddressSource]string{
		0x00: "unspecified",
		0x01: "static",
		0x02: "dhcp",
		0x03: "bios",
		0x04: "other",
	}
	s, ok := m[i]
	if ok {
		return s
	}
	return ""
}

type IPHeaderParams struct {
	TTL        uint8
	Flags      uint8
	Precedence uint8
	TOS        uint8
}

type ARPControl struct {
	ARPResponseEnabled   bool
	GratuitousARPEnabled bool
}

type CommunityString [18]byte

func (c CommunityString) String() string {
	s := []byte{}
	for _, v := range c {
		if v == 0x00 { // null
			break
		}
		s = append(s, v)
	}
	return string(s)
}

func NewCommunityString(s string) CommunityString {
	o := [18]byte{}

	b := []byte(s)
	for i := 0; i < 18; i++ {
		if i < len(b) {
			o[i] = b[i]
		}
		o[i] = 0x00
	}

	return CommunityString(o)
}

type AlertDestinationType struct {
	SetSelector uint8 // Destination selector.

	// 0b = Unacknowledged. Alert is assumed successful if transmission occurs
	// without error. This value is also used with Callback numbers.
	// 1b = Acknowledged. Alert is assumed successful only if acknowledged is
	// returned. Note, some alert types, such as Dial Page, do not support an
	// acknowledge
	AlertSupportAcknowledge bool

	// 000b = PET Trap destination
	// 001b - 101b = reserved
	// 110b = OEM 1
	// 111b = OEM 2
	DestinationType uint8

	// Alert Acknowledge Timeout / Retry Interval, in seconds, 0-based (i.e. minimum
	// timeout = 1 second)
	AlertAcknowledgeTimeout uint8

	// Number of times to retry alert to given destination.
	Retries uint8
}

type AlertDestinationAddress struct {
	SetSelector uint8

	// 0h = IPv4 IP Address followed by DIX Ethernet/802.3 MAC Address
	// 1h = IPv6 IP Address
	AddressFormat uint8

	IP4UseBackupGateway bool
	IP4IP               net.IP
	IP4MAC              net.HardwareAddr

	IP6IP net.IP
}

type VLAN struct {
	Enabled  bool
	ID       uint16
	Priority uint8
}

type AlertDestinationVLAN struct {
	SetSelector uint8

	// 0h = VLAN ID not used with this destination
	// 1h = 802.1q VLAN TAG
	AddressFormat uint8

	VLANID uint16
	// CFI (Canonical Format Indicator. Set to 0b)
	CFI      bool
	Priority uint8
}

type BadPasswordThreshold struct {
	// generate a Session Audit sensor "Invalid password disable" event message.
	GenerateSessionAuditEvent bool

	// Bad Password Threshold number
	Threshold uint8

	// Attempt Count Reset Interval.
	// The raw data is 2 byte, and the unit is in tens of seconds, the program should convert to seconds.
	AttemptCountResetIntervalSec uint32

	// User Lockout Interval
	// The raw data is 2 byte, and the unit is in tens of seconds, the program should convert to seconds.
	UserLockoutIntervalSec uint32
}

type IP6Support struct {
	// Implementation supports IPv6 Destination Addresses for LAN Alerting.
	SupportIP6AlertDestination bool
	//  Implementation can be configured to use both IPv4 and IPv6 addresses simultaneously
	CanUseBothIP4AndIP6 bool
	// Implementation can be configured to use IPv6 addresses only.
	CanUseIP6Only bool
}
