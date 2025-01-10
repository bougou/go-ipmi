package ipmi

import (
	"fmt"
)

type SetInProgress uint8

const (
	SetInProgress_SetComplete   SetInProgress = 0x00
	SetInProgress_SetInProgress SetInProgress = 0x01
	SetInProgress_CommitWrite   SetInProgress = 0x02
	SetInProgress_Reserved      SetInProgress = 0x03
)

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

type AuthTypesEnabled struct {
	OEM      bool
	Password bool
	MD5      bool
	MD2      bool
	None     bool
}

func (authTypeEnabled AuthTypesEnabled) String() string {
	if authTypeEnabled.OEM {
		return "OEM"
	}
	if authTypeEnabled.Password {
		return "Password"
	}
	if authTypeEnabled.MD5 {
		return "MD5"
	}
	if authTypeEnabled.MD2 {
		return "MD2"
	}
	if authTypeEnabled.None {
		return "None"
	}
	return "Unknown"
}

func packAuthTypesEnabled(a *AuthTypesEnabled) byte {
	b := uint8(0)
	b = setOrClearBit5(b, a.OEM)
	b = setOrClearBit4(b, a.Password)
	b = setOrClearBit2(b, a.MD5)
	b = setOrClearBit1(b, a.MD2)
	b = setOrClearBit0(b, a.None)
	return b
}

func unpackAuthTypesEnabled(b byte) *AuthTypesEnabled {
	return &AuthTypesEnabled{
		OEM:      isBit5Set(b),
		Password: isBit4Set(b),
		MD5:      isBit2Set(b),
		MD2:      isBit1Set(b),
		None:     isBit0Set(b),
	}
}

// see: LanConfigParameter_IPAddressSource (#4)
type LanIPAddressSource uint8

const (
	IPAddressSourceUnspecified LanIPAddressSource = 0x00
	IPAddressSourceStatic      LanIPAddressSource = 0x01
	IPAddressSourceDHCP        LanIPAddressSource = 0x02
	IPAddressSourceBIOS        LanIPAddressSource = 0x03
	IPAddressSourceOther       LanIPAddressSource = 0x04
)

func (i LanIPAddressSource) String() string {
	m := map[LanIPAddressSource]string{
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

type LanIPv6EnableMode uint8

const (
	// 00h = IPv6 addressing disabled.
	LanIPv6EnableMode_IPv6Disabled LanIPv6EnableMode = 0
	// 01h = Enable IPv6 addressing only. IPv4 addressing is disabled.
	LanIPv6EnableMode_IPv6Only LanIPv6EnableMode = 1
	// 02h = Enable IPv6 and IPv4 addressing simultaneously.
	LanIPv6EnableMode_IPv4AndIPv6 LanIPv6EnableMode = 2
)

// Address Status (Read-only parameter)
//   - 00h = Active (in-use)
//   - 01h = Disabled
//   - 02h = Pending (currently undergoing DAD [duplicate address detection], optional)
//   - 03h = Failed (duplicate address found, optional)
//   - 04h = Deprecated (preferred timer has expired, optional)
//   - 05h = Invalid (validity timer has expired, optional)
//   - All other = reserved
type LanIPv6AddressStatus uint8

const (
	LanIPv6AddressStatus_Active     LanIPv6AddressStatus = 0
	LanIPv6AddressStatus_Disabled   LanIPv6AddressStatus = 1
	LanIPv6AddressStatus_Pending    LanIPv6AddressStatus = 2
	LanIPv6AddressStatus_Failed     LanIPv6AddressStatus = 3
	LanIPv6AddressStatus_Deprecated LanIPv6AddressStatus = 4
	LanIPv6AddressStatus_Invalid    LanIPv6AddressStatus = 5
)

func (addressStatus LanIPv6AddressStatus) String() string {
	m := map[LanIPv6AddressStatus]string{
		0x00: "active",
		0x01: "disabled",
		0x02: "pending",
		0x03: "failed",
		0x04: "deprecated",
		0x05: "invalid",
	}
	s, ok := m[addressStatus]
	if ok {
		return s
	}
	return "reserved"
}

// IPv6 Static Address Source
//   - 0h = Static
//   - All other = reserved
type LanIPv6StaticAddressSource uint8

const (
	LanIPv6StaticAddressSource_Static LanIPv6StaticAddressSource = 0
)

func (addressSource LanIPv6StaticAddressSource) String() string {
	m := map[LanIPv6StaticAddressSource]string{
		0: "static",
	}
	s, ok := m[addressSource]
	if ok {
		return s
	}
	return "reserved"
}

// Address source/type
//   - 0 - Reserved
//   - 1 - SLAAC (StateLess Address Auto Configuration)
//   - 2 - DHCPv6 (optional)
//   - Other - reserved
type LanIPv6DynamicAddressSource uint8

const (
	LanIPv6AddressSource_SLAAC  LanIPv6DynamicAddressSource = 1
	LanIPv6AddressSource_DHCPv6 LanIPv6DynamicAddressSource = 2
)

func (addressSource LanIPv6DynamicAddressSource) String() string {
	m := map[LanIPv6DynamicAddressSource]string{
		1: "SLAAC",
		2: "DHCPv6",
	}
	s, ok := m[addressSource]
	if ok {
		return s
	}
	return "reserved"
}

// DHCPv6 Timing Configuration Mode
//   - 00b = `Not Supported`
//     DHCPv6 timing configuration per IPMI is not supported.
//   - 01b = `Global`
//     Timing configuration applies across all interfaces (IAs) that use
//     dynamic addressing and have DHCPv6 is enabled.
//   - 10b = `Per Interface`
//     Timing is configurable for each interface and used when DHCPv6 is enabled for the given interface (IA).
//   - 11b = reserved
type LanIPv6DHCPv6TimingConfigMode uint8

const (
	LanIPv6DHCPv6TimingConfigMode_NotSupported LanIPv6DHCPv6TimingConfigMode = 0
	LanIPv6DHCPv6TimingConfigMode_Global       LanIPv6DHCPv6TimingConfigMode = 1
	LanIPv6DHCPv6TimingConfigMode_PerInterface LanIPv6DHCPv6TimingConfigMode = 2
)

func (mode LanIPv6DHCPv6TimingConfigMode) String() string {
	m := map[LanIPv6DHCPv6TimingConfigMode]string{
		0: "not supported",
		1: "global",
		2: "per interface",
	}
	s, ok := m[mode]
	if ok {
		return s
	}
	return "reserved"
}

type LanIPv6NDSLAACTimingConfigMode uint8

const (
	LanIPv6NDSLAACTimingConfigMode_NotSupported LanIPv6DHCPv6TimingConfigMode  = 0
	LanIPv6NDSLAACTimingConfigMode_Global       LanIPv6NDSLAACTimingConfigMode = 1
	LanIPv6NDSLAACTimingConfigMode_PerInterface LanIPv6NDSLAACTimingConfigMode = 2
)

func (mode LanIPv6NDSLAACTimingConfigMode) String() string {
	m := map[LanIPv6NDSLAACTimingConfigMode]string{
		0: "not supported",
		1: "global",
		2: "per interface",
	}
	s, ok := m[mode]
	if ok {
		return s
	}
	return "reserved"
}

type LanConfig struct {
	SetInProgress                     *LanConfigParam_SetInProgress                       // #0, Read Only
	AuthTypeSupport                   *LanConfigParam_AuthTypeSupport                     // #1
	AuthTypeEnables                   *LanConfigParam_AuthTypeEnables                     // #2
	IP                                *LanConfigParam_IP                                  // #3
	IPSource                          *LanConfigParam_IPSource                            // #4
	MAC                               *LanConfigParam_MAC                                 // #5, can be Read Only. An implementation can either allow this parameter to be settable, or it can be implemented as Read Only.
	SubnetMask                        *LanConfigParam_SubnetMask                          // #6
	IPHeaderParams                    *LanConfigParam_IPv4HeaderParams                    // #7
	PrimaryRMCPPort                   *LanConfigParam_PrimaryRMCPPort                     // #8
	SecondaryRMCPPort                 *LanConfigParam_SecondaryRMCPPort                   // #9
	ARPControl                        *LanConfigParam_ARPControl                          // #10
	GratuitousARPInterval             *LanConfigParam_GratuitousARPInterval               // #11
	DefaultGatewayIP                  *LanConfigParam_DefaultGatewayIP                    // #12
	DefaultGatewayMAC                 *LanConfigParam_DefaultGatewayMAC                   // #13
	BackupGatewayIP                   *LanConfigParam_BackupGatewayIP                     // #14
	BackupGatewayMAC                  *LanConfigParam_BackupGatewayMAC                    // #15
	CommunityString                   *LanConfigParam_CommunityString                     // #16
	AlertDestinationsCount            *LanConfigParam_AlertDestinationsCount              // #17, Read Only
	AlertDestinationTypes             []*LanConfigParam_AlertDestinationType              // #18
	AlertDestinationAddresses         []*LanConfigParam_AlertDestinationAddress           // #19
	VLANID                            *LanConfigParam_VLANID                              // #20
	VLANPriority                      *LanConfigParam_VLANPriority                        // #21
	CipherSuitesSupport               *LanConfigParam_CipherSuitesSupport                 // #22, Read Only
	CipherSuitesID                    *LanConfigParam_CipherSuitesID                      // #23, Read Only
	CipherSuitesPrivLevel             *LanConfigParam_CipherSuitesPrivLevel               // #24
	AlertDestinationVLANs             []*LanConfigParam_AlertDestinationVLAN              // #25, can be READ ONLY
	BadPasswordThreshold              *LanConfigParam_BadPasswordThreshold                // #26
	IPv6Support                       *LanConfigParam_IPv6Support                         // #50, Read Only
	IPv6Enables                       *LanConfigParam_IPv6Enables                         // #51
	IPv6StaticTrafficClass            *LanConfigParam_IPv6StaticTrafficClass              // #52
	IPv6StaticHopLimit                *LanConfigParam_IPv6StaticHopLimit                  // #53
	IPv6FlowLabel                     *LanConfigParam_IPv6FlowLabel                       // #54
	IPv6Status                        *LanConfigParam_IPv6Status                          // #55, Read Only
	IPv6StaticAddresses               []*LanConfigParam_IPv6StaticAddress                 // #56
	IPv6DHCPv6StaticDUIDCount         *LanConfigParam_IPv6DHCPv6StaticDUIDCount           // #57, Read Only
	IPv6DHCPv6StaticDUIDs             []*LanConfigParam_IPv6DHCPv6StaticDUID              // #58
	IPv6DynamicAddresses              []*LanConfigParam_IPv6DynamicAddress                // #59, Read Only
	IPv6DHCPv6DynamicDUIDCount        *LanConfigParam_IPv6DHCPv6DynamicDUIDCount          // #60, Read Only
	IPv6DHCPv6DynamicDUIDs            []*LanConfigParam_IPv6DHCPv6DynamicDUID             // #61
	IPv6DHCPv6TimingConfigSupport     *LanConfigParam_IPv6DHCPv6TimingConfigSupport       // #62, Read Only
	IPv6DHCPv6TimingConfig            []*LanConfigParam_IPv6DHCPv6TimingConfig            // #63
	IPv6RouterAddressConfigControl    *LanConfigParam_IPv6RouterAddressConfigControl      // #64
	IPv6StaticRouter1IP               *LanConfigParam_IPv6StaticRouter1IP                 // #65
	IPv6StaticRouter1MAC              *LanConfigParam_IPv6StaticRouter1MAC                // #66
	IPv6StaticRouter1PrefixLength     *LanConfigParam_IPv6StaticRouter1PrefixLength       // #67
	IPv6StaticRouter1PrefixValue      *LanConfigParam_IPv6StaticRouter1PrefixValue        // #68
	IPv6StaticRouter2IP               *LanConfigParam_IPv6StaticRouter2IP                 // #69
	IPv6StaticRouter2MAC              *LanConfigParam_IPv6StaticRouter2MAC                // #70
	IPv6StaticRouter2PrefixLength     *LanConfigParam_IPv6StaticRouter2PrefixLength       // #71
	IPv6StaticRouter2PrefixValue      *LanConfigParam_IPv6StaticRouter2PrefixValue        // #72
	IPv6DynamicRouterInfoSets         *LanConfigParam_IPv6DynamicRouterInfoSets           // #73, Read Only
	IPv6DynamicRouterInfoIP           []*LanConfigParam_IPv6DynamicRouterInfoIP           // #74
	IPv6DynamicRouterInfoMAC          []*LanConfigParam_IPv6DynamicRouterInfoMAC          // #75
	IPv6DynamicRouterInfoPrefixLength []*LanConfigParam_IPv6DynamicRouterInfoPrefixLength // #76
	IPv6DynamicRouterInfoPrefixValue  []*LanConfigParam_IPv6DynamicRouterInfoPrefixValue  // #77
	IPv6DynamicRouterReceivedHopLimit *LanConfigParam_IPv6DynamicRouterReceivedHopLimit   // #78
	IPv6NDSLAACTimingConfigSupport    *LanConfigParam_IPv6NDSLAACTimingConfigSupport      // #79, Read Only
	IPv6NDSLAACTimingConfig           []*LanConfigParam_IPv6NDSLAACTimingConfig           // #80
}

func (lanConfigParams *LanConfig) Format() string {
	format := func(param LanConfigParameter) string {
		paramSelector, _, _ := param.LanConfigParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%2d] %-40s: %s", paramSelector, paramSelector.String(), content)
	}

	out := ""

	if lanConfigParams.SetInProgress != nil {
		out += format(lanConfigParams.SetInProgress)
	}
	if lanConfigParams.AuthTypeSupport != nil {
		out += format(lanConfigParams.AuthTypeSupport)
	}
	if lanConfigParams.AuthTypeEnables != nil {
		out += format(lanConfigParams.AuthTypeEnables)
	}
	if lanConfigParams.IP != nil {
		out += format(lanConfigParams.IP)
	}
	if lanConfigParams.IPSource != nil {
		out += format(lanConfigParams.IPSource)
	}
	if lanConfigParams.MAC != nil {
		out += format(lanConfigParams.MAC)
	}
	if lanConfigParams.SubnetMask != nil {
		out += format(lanConfigParams.SubnetMask)
	}
	if lanConfigParams.IPHeaderParams != nil {
		out += format(lanConfigParams.IPHeaderParams)
	}
	if lanConfigParams.PrimaryRMCPPort != nil {
		out += format(lanConfigParams.PrimaryRMCPPort)
	}
	if lanConfigParams.SecondaryRMCPPort != nil {
		out += format(lanConfigParams.SecondaryRMCPPort)
	}
	if lanConfigParams.ARPControl != nil {
		out += format(lanConfigParams.ARPControl)
	}
	if lanConfigParams.GratuitousARPInterval != nil {
		out += format(lanConfigParams.GratuitousARPInterval)
	}
	if lanConfigParams.DefaultGatewayIP != nil {
		out += format(lanConfigParams.DefaultGatewayIP)
	}
	if lanConfigParams.DefaultGatewayMAC != nil {
		out += format(lanConfigParams.DefaultGatewayMAC)
	}
	if lanConfigParams.BackupGatewayIP != nil {
		out += format(lanConfigParams.BackupGatewayIP)
	}
	if lanConfigParams.BackupGatewayMAC != nil {
		out += format(lanConfigParams.BackupGatewayMAC)
	}
	if lanConfigParams.CommunityString != nil {
		out += format(lanConfigParams.CommunityString)
	}
	if lanConfigParams.AlertDestinationsCount != nil {
		out += format(lanConfigParams.AlertDestinationsCount)
	}
	if lanConfigParams.AlertDestinationTypes != nil {
		for _, alertDestinationType := range lanConfigParams.AlertDestinationTypes {
			out += format(alertDestinationType)
		}
	}
	if lanConfigParams.AlertDestinationAddresses != nil {
		for _, alertDestinationAddress := range lanConfigParams.AlertDestinationAddresses {
			out += format(alertDestinationAddress)
		}
	}
	if lanConfigParams.VLANID != nil {
		out += format(lanConfigParams.VLANID)
	}
	if lanConfigParams.VLANPriority != nil {
		out += format(lanConfigParams.VLANPriority)
	}
	if lanConfigParams.CipherSuitesSupport != nil {
		out += format(lanConfigParams.CipherSuitesSupport)
	}
	if lanConfigParams.CipherSuitesID != nil {
		out += format(lanConfigParams.CipherSuitesID)
	}
	if lanConfigParams.CipherSuitesPrivLevel != nil {
		out += format(lanConfigParams.CipherSuitesPrivLevel)
	}
	if lanConfigParams.AlertDestinationVLANs != nil {
		for _, alertDestinationVLAN := range lanConfigParams.AlertDestinationVLANs {
			out += format(alertDestinationVLAN)
		}
	}
	if lanConfigParams.BadPasswordThreshold != nil {
		out += format(lanConfigParams.BadPasswordThreshold)
	}
	if lanConfigParams.IPv6Support != nil {
		out += format(lanConfigParams.IPv6Support)
	}
	if lanConfigParams.IPv6Enables != nil {
		out += format(lanConfigParams.IPv6Enables)
	}
	if lanConfigParams.IPv6StaticTrafficClass != nil {
		out += format(lanConfigParams.IPv6StaticTrafficClass)
	}
	if lanConfigParams.IPv6StaticHopLimit != nil {
		out += format(lanConfigParams.IPv6StaticHopLimit)
	}
	if lanConfigParams.IPv6FlowLabel != nil {
		out += format(lanConfigParams.IPv6FlowLabel)
	}
	if lanConfigParams.IPv6Status != nil {
		out += format(lanConfigParams.IPv6Status)
	}
	if lanConfigParams.IPv6StaticAddresses != nil {
		for _, ipv6StaticAddress := range lanConfigParams.IPv6StaticAddresses {
			out += format(ipv6StaticAddress)
		}
	}
	if lanConfigParams.IPv6DHCPv6StaticDUIDCount != nil {
		out += format(lanConfigParams.IPv6DHCPv6StaticDUIDCount)
	}
	if lanConfigParams.IPv6DHCPv6StaticDUIDs != nil {
		for _, ipv6DHCPv6StaticDUID := range lanConfigParams.IPv6DHCPv6StaticDUIDs {
			out += format(ipv6DHCPv6StaticDUID)
		}
	}
	if lanConfigParams.IPv6DynamicAddresses != nil {
		for _, ipv6DynamicAddress := range lanConfigParams.IPv6DynamicAddresses {
			out += format(ipv6DynamicAddress)
		}
	}
	if lanConfigParams.IPv6DHCPv6DynamicDUIDCount != nil {
		out += format(lanConfigParams.IPv6DHCPv6DynamicDUIDCount)
	}
	if lanConfigParams.IPv6DHCPv6DynamicDUIDs != nil {
		for _, ipv6DHCPv6DynamicDUID := range lanConfigParams.IPv6DHCPv6DynamicDUIDs {
			out += format(ipv6DHCPv6DynamicDUID)
		}
	}
	if lanConfigParams.IPv6DHCPv6TimingConfigSupport != nil {
		out += format(lanConfigParams.IPv6DHCPv6TimingConfigSupport)
	}
	if lanConfigParams.IPv6DHCPv6TimingConfig != nil {
		for _, ipv6DHCPv6TimingConfig := range lanConfigParams.IPv6DHCPv6TimingConfig {
			out += format(ipv6DHCPv6TimingConfig)
		}
	}
	if lanConfigParams.IPv6RouterAddressConfigControl != nil {
		out += format(lanConfigParams.IPv6RouterAddressConfigControl)
	}
	if lanConfigParams.IPv6StaticRouter1IP != nil {
		out += format(lanConfigParams.IPv6StaticRouter1IP)
	}
	if lanConfigParams.IPv6StaticRouter1MAC != nil {
		out += format(lanConfigParams.IPv6StaticRouter1MAC)
	}
	if lanConfigParams.IPv6StaticRouter1PrefixLength != nil {
		out += format(lanConfigParams.IPv6StaticRouter1PrefixLength)
	}
	if lanConfigParams.IPv6StaticRouter1PrefixValue != nil {
		out += format(lanConfigParams.IPv6StaticRouter1PrefixValue)
	}
	if lanConfigParams.IPv6StaticRouter2IP != nil {
		out += format(lanConfigParams.IPv6StaticRouter2IP)
	}
	if lanConfigParams.IPv6StaticRouter2MAC != nil {
		out += format(lanConfigParams.IPv6StaticRouter2MAC)
	}
	if lanConfigParams.IPv6StaticRouter2PrefixLength != nil {
		out += format(lanConfigParams.IPv6StaticRouter2PrefixLength)
	}
	if lanConfigParams.IPv6StaticRouter2PrefixValue != nil {
		out += format(lanConfigParams.IPv6StaticRouter2PrefixValue)
	}
	if lanConfigParams.IPv6DynamicRouterInfoSets != nil {
		out += format(lanConfigParams.IPv6DynamicRouterInfoSets)
	}
	if lanConfigParams.IPv6DynamicRouterInfoIP != nil {
		for _, ipv6DynamicRouterInfoIP := range lanConfigParams.IPv6DynamicRouterInfoIP {
			out += format(ipv6DynamicRouterInfoIP)
		}
	}
	if lanConfigParams.IPv6DynamicRouterInfoMAC != nil {
		for _, ipv6DynamicRouterInfoMAC := range lanConfigParams.IPv6DynamicRouterInfoMAC {
			out += format(ipv6DynamicRouterInfoMAC)
		}
	}
	if lanConfigParams.IPv6DynamicRouterInfoPrefixLength != nil {
		for _, ipv6DynamicRouterInfoPrefixLength := range lanConfigParams.IPv6DynamicRouterInfoPrefixLength {
			out += format(ipv6DynamicRouterInfoPrefixLength)
		}
	}
	if lanConfigParams.IPv6DynamicRouterInfoPrefixValue != nil {
		for _, ipv6DynamicRouterInfoPrefixValue := range lanConfigParams.IPv6DynamicRouterInfoPrefixValue {
			out += format(ipv6DynamicRouterInfoPrefixValue)
		}
	}
	if lanConfigParams.IPv6DynamicRouterReceivedHopLimit != nil {
		out += format(lanConfigParams.IPv6DynamicRouterReceivedHopLimit)
	}
	if lanConfigParams.IPv6NDSLAACTimingConfigSupport != nil {
		out += format(lanConfigParams.IPv6NDSLAACTimingConfigSupport)
	}
	if lanConfigParams.IPv6NDSLAACTimingConfig != nil {
		for _, ipv6NDSLAACTimingConfig := range lanConfigParams.IPv6NDSLAACTimingConfig {
			out += format(ipv6NDSLAACTimingConfig)
		}
	}

	return out
}
