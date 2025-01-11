package ipmi

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type LanConfigParameter interface {
	LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8)
	Parameter
}

// Table 23-4, LAN Configuration Parameters
// Parameter selector
type LanConfigParamSelector uint8

const (
	LanConfigParamSelector_SetInProgress                     LanConfigParamSelector = 0
	LanConfigParamSelector_AuthTypeSupport                   LanConfigParamSelector = 1 // read only
	LanConfigParamSelector_AuthTypeEnables                   LanConfigParamSelector = 2
	LanConfigParamSelector_IP                                LanConfigParamSelector = 3
	LanConfigParamSelector_IPSource                          LanConfigParamSelector = 4
	LanConfigParamSelector_MAC                               LanConfigParamSelector = 5 // can be read only
	LanConfigParamSelector_SubnetMask                        LanConfigParamSelector = 6
	LanConfigParamSelector_IPv4HeaderParams                  LanConfigParamSelector = 7
	LanConfigParamSelector_PrimaryRMCPPort                   LanConfigParamSelector = 8
	LanConfigParamSelector_SecondaryRMCPPort                 LanConfigParamSelector = 9
	LanConfigParamSelector_ARPControl                        LanConfigParamSelector = 10
	LanConfigParamSelector_GratuitousARPInterval             LanConfigParamSelector = 11
	LanConfigParamSelector_DefaultGatewayIP                  LanConfigParamSelector = 12
	LanConfigParamSelector_DefaultGatewayMAC                 LanConfigParamSelector = 13
	LanConfigParamSelector_BackupGatewayIP                   LanConfigParamSelector = 14
	LanConfigParamSelector_BackupGatewayMAC                  LanConfigParamSelector = 15
	LanConfigParamSelector_CommunityString                   LanConfigParamSelector = 16
	LanConfigParamSelector_AlertDestinationsCount            LanConfigParamSelector = 17 // read only
	LanConfigParamSelector_AlertDestinationType              LanConfigParamSelector = 18
	LanConfigParamSelector_AlertDestinationAddress           LanConfigParamSelector = 19
	LanConfigParamSelector_VLANID                            LanConfigParamSelector = 20
	LanConfigParamSelector_VLANPriority                      LanConfigParamSelector = 21
	LanConfigParamSelector_CipherSuitesSupport               LanConfigParamSelector = 22 // read only
	LanConfigParamSelector_CipherSuitesID                    LanConfigParamSelector = 23 // read only
	LanConfigParamSelector_CipherSuitesPrivLevel             LanConfigParamSelector = 24
	LanConfigParamSelector_AlertDestinationVLAN              LanConfigParamSelector = 25 // can be read only
	LanConfigParamSelector_BadPasswordThreshold              LanConfigParamSelector = 26
	LanConfigParamSelector_IPv6Support                       LanConfigParamSelector = 50 // read only
	LanConfigParamSelector_IPv6Enables                       LanConfigParamSelector = 51
	LanConfigParamSelector_IPv6StaticTrafficClass            LanConfigParamSelector = 52
	LanConfigParamSelector_IPv6StaticHopLimit                LanConfigParamSelector = 53
	LanConfigParamSelector_IPv6FlowLabel                     LanConfigParamSelector = 54
	LanConfigParamSelector_IPv6Status                        LanConfigParamSelector = 55 // read only
	LanConfigParamSelector_IPv6StaticAddress                 LanConfigParamSelector = 56
	LanConfigParamSelector_IPv6DHCPv6StaticDUIDCount         LanConfigParamSelector = 57 // read only
	LanConfigParamSelector_IPv6DHCPv6StaticDUID              LanConfigParamSelector = 58
	LanConfigParamSelector_IPv6DynamicAddress                LanConfigParamSelector = 59 // read only
	LanConfigParamSelector_IPv6DHCPv6DynamicDUIDCount        LanConfigParamSelector = 60 // read only
	LanConfigParamSelector_IPv6DHCPv6DynamicDUID             LanConfigParamSelector = 61
	LanConfigParamSelector_IPv6DHCPv6TimingConfigSupport     LanConfigParamSelector = 62 // read only
	LanConfigParamSelector_IPv6DHCPv6TimingConfig            LanConfigParamSelector = 63
	LanConfigParamSelector_IPv6RouterAddressConfigControl    LanConfigParamSelector = 64
	LanConfigParamSelector_IPv6StaticRouter1IP               LanConfigParamSelector = 65
	LanConfigParamSelector_IPv6StaticRouter1MAC              LanConfigParamSelector = 66
	LanConfigParamSelector_IPv6StaticRouter1PrefixLength     LanConfigParamSelector = 67
	LanConfigParamSelector_IPv6StaticRouter1PrefixValue      LanConfigParamSelector = 68
	LanConfigParamSelector_IPv6StaticRouter2IP               LanConfigParamSelector = 69
	LanConfigParamSelector_IPv6StaticRouter2MAC              LanConfigParamSelector = 70
	LanConfigParamSelector_IPv6StaticRouter2PrefixLength     LanConfigParamSelector = 71
	LanConfigParamSelector_IPv6StaticRouter2PrefixValue      LanConfigParamSelector = 72
	LanConfigParamSelector_IPv6DynamicRouterInfoCount        LanConfigParamSelector = 73 // read only
	LanConfigParamSelector_IPv6DynamicRouterInfoIP           LanConfigParamSelector = 74 // read only
	LanConfigParamSelector_IPv6DynamicRouterInfoMAC          LanConfigParamSelector = 75 // read only
	LanConfigParamSelector_IPv6DynamicRouterInfoPrefixLength LanConfigParamSelector = 76 // read only
	LanConfigParamSelector_IPv6DynamicRouterInfoPrefixValue  LanConfigParamSelector = 77 // read only
	LanConfigParamSelector_IPv6DynamicRouterReceivedHopLimit LanConfigParamSelector = 78 // read only
	LanConfigParamSelector_IPv6NDSLAACTimingConfigSupport    LanConfigParamSelector = 79 // read only, IPv6 Neighbor	Discovery / SLAAC
	LanConfigParamSelector_IPv6NDSLAACTimingConfig           LanConfigParamSelector = 80

	// OEM Parameters 192:255
	// This range is available for special OEM configuration parameters. The OEM is identified
	// according to the Manufacturer ID field returned by the Get Device ID command.

)

func (lanConfigParam LanConfigParamSelector) String() string {
	m := map[LanConfigParamSelector]string{
		LanConfigParamSelector_SetInProgress:                     "Set In Progress",
		LanConfigParamSelector_AuthTypeSupport:                   "Authentication Type Support",
		LanConfigParamSelector_AuthTypeEnables:                   "Authentication Type Enables",
		LanConfigParamSelector_IP:                                "IP Address",
		LanConfigParamSelector_IPSource:                          "IP Address Source",
		LanConfigParamSelector_MAC:                               "MAC Address",
		LanConfigParamSelector_SubnetMask:                        "Subnet Mask",
		LanConfigParamSelector_IPv4HeaderParams:                  "IPv4 Header Params",
		LanConfigParamSelector_PrimaryRMCPPort:                   "Primary RMCP Port",
		LanConfigParamSelector_SecondaryRMCPPort:                 "Secondary RMCP Port",
		LanConfigParamSelector_ARPControl:                        "ARP Control",
		LanConfigParamSelector_GratuitousARPInterval:             "Gratuitous ARP Interval",
		LanConfigParamSelector_DefaultGatewayIP:                  "Default Gateway IP",
		LanConfigParamSelector_DefaultGatewayMAC:                 "Default Gateway MAC",
		LanConfigParamSelector_BackupGatewayIP:                   "Backup Gateway IP",
		LanConfigParamSelector_BackupGatewayMAC:                  "Backup Gateway MAC",
		LanConfigParamSelector_CommunityString:                   "Community String",
		LanConfigParamSelector_AlertDestinationsCount:            "Alert Destinations Count",
		LanConfigParamSelector_AlertDestinationType:              "Alert Destination Type",
		LanConfigParamSelector_AlertDestinationAddress:           "Alert Destination Address",
		LanConfigParamSelector_VLANID:                            "802.1q VLAN ID",
		LanConfigParamSelector_VLANPriority:                      "802.1q VLAN Priority",
		LanConfigParamSelector_CipherSuitesSupport:               "Cipher Suite Entries Support",
		LanConfigParamSelector_CipherSuitesID:                    "Cipher Suite Entries",
		LanConfigParamSelector_CipherSuitesPrivLevel:             "Cipher Suite Privilege Levels",
		LanConfigParamSelector_AlertDestinationVLAN:              "Alert Destination VLAN",
		LanConfigParamSelector_BadPasswordThreshold:              "Bad Password Threshold",
		LanConfigParamSelector_IPv6Support:                       "IPv6 Support",
		LanConfigParamSelector_IPv6Enables:                       "IPv6 Enables",
		LanConfigParamSelector_IPv6StaticTrafficClass:            "IPv6 Static Traffic Class",
		LanConfigParamSelector_IPv6StaticHopLimit:                "IPv6 Static Hop Limit",
		LanConfigParamSelector_IPv6FlowLabel:                     "IPv6 Flow Label",
		LanConfigParamSelector_IPv6Status:                        "IPv6 Status",
		LanConfigParamSelector_IPv6StaticAddress:                 "IPv6 Static Address",
		LanConfigParamSelector_IPv6DHCPv6StaticDUIDCount:         "IPv6 DHCPv6 Static DUID Count",
		LanConfigParamSelector_IPv6DHCPv6StaticDUID:              "IPv6 DHCPv6 Static DUID",
		LanConfigParamSelector_IPv6DynamicAddress:                "IPv6 Dynamic Address",
		LanConfigParamSelector_IPv6DHCPv6DynamicDUIDCount:        "IPv6 DHCPv6 Dynamic DUID Count",
		LanConfigParamSelector_IPv6DHCPv6DynamicDUID:             "IPv6 DHCPv6 Dynamic DUID",
		LanConfigParamSelector_IPv6DHCPv6TimingConfigSupport:     "IPv6 DHCPv6 Timing Config Support",
		LanConfigParamSelector_IPv6DHCPv6TimingConfig:            "IPv6 DHCPv6 Timing Config",
		LanConfigParamSelector_IPv6RouterAddressConfigControl:    "IPv6 Router Address Config Control",
		LanConfigParamSelector_IPv6StaticRouter1IP:               "IPv6 Static Router1 IP",
		LanConfigParamSelector_IPv6StaticRouter1MAC:              "IPv6 Static Router1 MAC",
		LanConfigParamSelector_IPv6StaticRouter1PrefixLength:     "IPv6 Static Router1 Prefix Length",
		LanConfigParamSelector_IPv6StaticRouter1PrefixValue:      "IPv6 Static Router1 Prefix Value",
		LanConfigParamSelector_IPv6StaticRouter2IP:               "IPv6 Static Router2 IP",
		LanConfigParamSelector_IPv6StaticRouter2MAC:              "IPv6 Static Router2 MAC",
		LanConfigParamSelector_IPv6StaticRouter2PrefixLength:     "IPv6 Static Router2 Prefix Length",
		LanConfigParamSelector_IPv6StaticRouter2PrefixValue:      "IPv6 Static Router2 Prefix Value",
		LanConfigParamSelector_IPv6DynamicRouterInfoCount:        "IPv6 Dynamic Router Sets Number",
		LanConfigParamSelector_IPv6DynamicRouterInfoIP:           "IPv6 Dynamic Router IP",
		LanConfigParamSelector_IPv6DynamicRouterInfoMAC:          "IPv6 Dynamic Router MAC",
		LanConfigParamSelector_IPv6DynamicRouterInfoPrefixLength: "IPv6 Dynamic Router Prefix Length",
		LanConfigParamSelector_IPv6DynamicRouterInfoPrefixValue:  "IPv6 Dynamic Router Prefix Value",
		LanConfigParamSelector_IPv6DynamicRouterReceivedHopLimit: "IPv6 Dynamic Router Received Hop Limit",
		LanConfigParamSelector_IPv6NDSLAACTimingConfigSupport:    "IPv6 ND/SLAAC Timing Config Support",
		LanConfigParamSelector_IPv6NDSLAACTimingConfig:           "IPv6 ND/SLAAC Timing Config",
	}

	if s, ok := m[lanConfigParam]; ok {
		return s
	}

	return "Unknown"
}

var (
	_ LanConfigParameter = (*LanConfigParam_SetInProgress)(nil)
	_ LanConfigParameter = (*LanConfigParam_AuthTypeSupport)(nil)
	_ LanConfigParameter = (*LanConfigParam_AuthTypeEnables)(nil)
	_ LanConfigParameter = (*LanConfigParam_IP)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPSource)(nil)
	_ LanConfigParameter = (*LanConfigParam_MAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_SubnetMask)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv4HeaderParams)(nil)
	_ LanConfigParameter = (*LanConfigParam_PrimaryRMCPPort)(nil)
	_ LanConfigParameter = (*LanConfigParam_SecondaryRMCPPort)(nil)
	_ LanConfigParameter = (*LanConfigParam_ARPControl)(nil)
	_ LanConfigParameter = (*LanConfigParam_GratuitousARPInterval)(nil)
	_ LanConfigParameter = (*LanConfigParam_DefaultGatewayIP)(nil)
	_ LanConfigParameter = (*LanConfigParam_DefaultGatewayMAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_BackupGatewayIP)(nil)
	_ LanConfigParameter = (*LanConfigParam_BackupGatewayMAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_CommunityString)(nil)
	_ LanConfigParameter = (*LanConfigParam_AlertDestinationsCount)(nil)
	_ LanConfigParameter = (*LanConfigParam_AlertDestinationType)(nil)
	_ LanConfigParameter = (*LanConfigParam_AlertDestinationAddress)(nil)
	_ LanConfigParameter = (*LanConfigParam_VLANID)(nil)
	_ LanConfigParameter = (*LanConfigParam_VLANPriority)(nil)
	_ LanConfigParameter = (*LanConfigParam_CipherSuitesSupport)(nil)
	_ LanConfigParameter = (*LanConfigParam_CipherSuitesID)(nil)
	_ LanConfigParameter = (*LanConfigParam_CipherSuitesPrivLevel)(nil)
	_ LanConfigParameter = (*LanConfigParam_AlertDestinationVLAN)(nil)
	_ LanConfigParameter = (*LanConfigParam_BadPasswordThreshold)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6Support)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6Enables)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticTrafficClass)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticHopLimit)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6FlowLabel)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6Status)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticAddress)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6StaticDUIDCount)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6StaticDUID)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicAddress)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6DynamicDUIDCount)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6DynamicDUID)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6TimingConfigSupport)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DHCPv6TimingConfig)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6RouterAddressConfigControl)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter1IP)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter1MAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter1PrefixLength)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter1PrefixValue)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter2IP)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter2MAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter2PrefixLength)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6StaticRouter2PrefixValue)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterInfoSets)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterInfoIP)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterInfoMAC)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterInfoPrefixLength)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterInfoPrefixValue)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6DynamicRouterReceivedHopLimit)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6NDSLAACTimingConfigSupport)(nil)
	_ LanConfigParameter = (*LanConfigParam_IPv6NDSLAACTimingConfig)(nil)
)

func isNilLanConfigParameter(param LanConfigParameter) bool {
	switch v := param.(type) {
	// MUST not put multiple types on the same case.
	case *LanConfigParam_SetInProgress:
		return v == nil
	case *LanConfigParam_AuthTypeSupport:
		return v == nil
	case *LanConfigParam_AuthTypeEnables:
		return v == nil
	case *LanConfigParam_IP:
		return v == nil
	case *LanConfigParam_IPSource:
		return v == nil
	case *LanConfigParam_MAC:
		return v == nil
	case *LanConfigParam_SubnetMask:
		return v == nil
	case *LanConfigParam_IPv4HeaderParams:
		return v == nil
	case *LanConfigParam_PrimaryRMCPPort:
		return v == nil
	case *LanConfigParam_SecondaryRMCPPort:
		return v == nil
	case *LanConfigParam_ARPControl:
		return v == nil
	case *LanConfigParam_GratuitousARPInterval:
		return v == nil
	case *LanConfigParam_DefaultGatewayIP:
		return v == nil
	case *LanConfigParam_DefaultGatewayMAC:
		return v == nil
	case *LanConfigParam_BackupGatewayIP:
		return v == nil
	case *LanConfigParam_BackupGatewayMAC:
		return v == nil
	case *LanConfigParam_CommunityString:
		return v == nil
	case *LanConfigParam_AlertDestinationsCount:
		return v == nil
	case *LanConfigParam_AlertDestinationType:
		return v == nil
	case *LanConfigParam_AlertDestinationAddress:
		return v == nil
	case *LanConfigParam_VLANID:
		return v == nil
	case *LanConfigParam_VLANPriority:
		return v == nil
	case *LanConfigParam_CipherSuitesSupport:
		return v == nil
	case *LanConfigParam_CipherSuitesID:
		return v == nil
	case *LanConfigParam_CipherSuitesPrivLevel:
		return v == nil
	case *LanConfigParam_AlertDestinationVLAN:
		return v == nil
	case *LanConfigParam_BadPasswordThreshold:
		return v == nil
	case *LanConfigParam_IPv6Support:
		return v == nil
	case *LanConfigParam_IPv6Enables:
		return v == nil
	case *LanConfigParam_IPv6StaticTrafficClass:
		return v == nil
	case *LanConfigParam_IPv6StaticHopLimit:
		return v == nil
	case *LanConfigParam_IPv6FlowLabel:
		return v == nil
	case *LanConfigParam_IPv6Status:
		return v == nil
	case *LanConfigParam_IPv6StaticAddress:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6StaticDUIDCount:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6StaticDUID:
		return v == nil
	case *LanConfigParam_IPv6DynamicAddress:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6DynamicDUIDCount:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6DynamicDUID:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6TimingConfigSupport:
		return v == nil
	case *LanConfigParam_IPv6DHCPv6TimingConfig:
		return v == nil
	case *LanConfigParam_IPv6RouterAddressConfigControl:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter1IP:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter1MAC:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter1PrefixLength:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter1PrefixValue:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter2IP:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter2MAC:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter2PrefixLength:
		return v == nil
	case *LanConfigParam_IPv6StaticRouter2PrefixValue:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterInfoSets:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterInfoIP:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterInfoMAC:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterInfoPrefixLength:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterInfoPrefixValue:
		return v == nil
	case *LanConfigParam_IPv6DynamicRouterReceivedHopLimit:
		return v == nil
	case *LanConfigParam_IPv6NDSLAACTimingConfigSupport:
		return v == nil
	case *LanConfigParam_IPv6NDSLAACTimingConfig:
		return v == nil
	default:
		return false
	}
}

type LanConfigParam_SetInProgress struct {
	Value SetInProgress
}

func (param *LanConfigParam_SetInProgress) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_SetInProgress, 0, 0
}

func (param *LanConfigParam_SetInProgress) Pack() []byte {
	return []byte{byte(param.Value)}
}

func (param *LanConfigParam_SetInProgress) Unpack(data []byte) error {
	if len(data) != 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = SetInProgress(data[0])
	return nil
}

func (param *LanConfigParam_SetInProgress) Format() string {
	return param.Value.String()
}

type LanConfigParam_AuthTypeSupport struct {
	OEM      bool
	Password bool
	MD5      bool
	MD2      bool
	None     bool
}

func (param *LanConfigParam_AuthTypeSupport) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AuthTypeSupport, 0, 0
}

func (param *LanConfigParam_AuthTypeSupport) Pack() []byte {
	b := uint8(0)
	b = setOrClearBit5(b, param.OEM)
	b = setOrClearBit4(b, param.Password)
	b = setOrClearBit2(b, param.MD5)
	b = setOrClearBit1(b, param.MD2)
	b = setOrClearBit0(b, param.None)
	return []byte{b}
}

func (param *LanConfigParam_AuthTypeSupport) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	b := data[0]
	param.OEM = isBit5Set(b)
	param.Password = isBit4Set(b)
	// Bit 3 Reserved
	param.MD5 = isBit2Set(b)
	param.MD2 = isBit1Set(b)
	param.None = isBit0Set(b)

	return nil
}

func (param *LanConfigParam_AuthTypeSupport) Format() string {
	var s string
	if param.OEM {
		s = "OEM"
	} else if param.Password {
		s = "Password"
	} else if param.MD5 {
		s = "MD5"
	} else if param.MD2 {
		s = "MD2"
	} else if param.None {
		s = "None"
	}

	return fmt.Sprintf(`%s, (OEM: %v, Password: %v, MD5: %v, MD2: %v, Non: %v)`,
		s, param.OEM, param.Password, param.MD5, param.MD2, param.None)
}

type LanConfigParam_AuthTypeEnables struct {
	Callback *AuthTypesEnabled
	User     *AuthTypesEnabled
	Operator *AuthTypesEnabled
	Admin    *AuthTypesEnabled
	OEM      *AuthTypesEnabled
}

func (param *LanConfigParam_AuthTypeEnables) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AuthTypeEnables, 0, 0
}

func (param *LanConfigParam_AuthTypeEnables) Unpack(data []byte) error {
	if len(data) < 5 {
		return ErrUnpackedDataTooShortWith(len(data), 5)
	}

	param.Callback = unpackAuthTypesEnabled(data[0])
	param.User = unpackAuthTypesEnabled(data[1])
	param.Operator = unpackAuthTypesEnabled(data[2])
	param.Admin = unpackAuthTypesEnabled(data[3])
	param.OEM = unpackAuthTypesEnabled(data[4])

	return nil
}

func (param *LanConfigParam_AuthTypeEnables) Pack() []byte {
	out := make([]byte, 5)

	out[0] = packAuthTypesEnabled(param.Callback)
	out[1] = packAuthTypesEnabled(param.User)
	out[2] = packAuthTypesEnabled(param.Operator)
	out[3] = packAuthTypesEnabled(param.Admin)
	out[4] = packAuthTypesEnabled(param.OEM)

	return out
}

func (param *LanConfigParam_AuthTypeEnables) Format() string {
	return fmt.Sprintf("Callback: %s, User: %s, Operator: %s, Admin: %s, OEM: %s",
		param.Callback.String(), param.User.String(), param.Operator.String(), param.Admin.String(), param.OEM.String())
}

type LanConfigParam_IP struct {
	IP net.IP
}

func (param *LanConfigParam_IP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IP, 0, 0
}

func (param *LanConfigParam_IP) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.IP = net.IPv4(data[0], data[1], data[2], data[3])

	return nil
}

func (param *LanConfigParam_IP) Pack() []byte {
	return []byte{param.IP[0], param.IP[1], param.IP[2], param.IP[3]}
}

func (param *LanConfigParam_IP) Format() string {
	return param.IP.String()
}

type LanConfigParam_IPSource struct {
	Source LanIPAddressSource
}

func (param *LanConfigParam_IPSource) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPSource, 0, 0
}

func (param *LanConfigParam_IPSource) Pack() []byte {
	return []byte{byte(param.Source)}
}

func (param *LanConfigParam_IPSource) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Source = LanIPAddressSource(data[0])
	return nil
}

func (param *LanConfigParam_IPSource) Format() string {
	return param.Source.String()
}

type LanConfigParam_MAC struct {
	MAC net.HardwareAddr
}

func (param *LanConfigParam_MAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_MAC, 0, 0
}

func (param *LanConfigParam_MAC) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}

	param.MAC = net.HardwareAddr(data[0:6])

	return nil
}

func (param *LanConfigParam_MAC) Pack() []byte {
	return param.MAC
}

func (param *LanConfigParam_MAC) Format() string {
	return param.MAC.String()
}

type LanConfigParam_SubnetMask struct {
	SubnetMask net.IP
}

func (param *LanConfigParam_SubnetMask) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_SubnetMask, 0, 0
}

func (param *LanConfigParam_SubnetMask) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.SubnetMask = net.IPv4(data[0], data[1], data[2], data[3])

	return nil
}

func (param *LanConfigParam_SubnetMask) Pack() []byte {
	return []byte{param.SubnetMask[0], param.SubnetMask[1], param.SubnetMask[2], param.SubnetMask[3]}
}

func (param *LanConfigParam_SubnetMask) Format() string {
	return param.SubnetMask.String()
}

type LanConfigParam_IPv4HeaderParams struct {
	// data 1 - Time-to-live. 1-based. (Default = 40h)
	// Value for time-to-live parameter in IP Header for RMCP packets and PET Traps
	// transmitted from this channel.
	TTL uint8

	// data 2
	//  - [7:5] - Flags. Sets value of bit 1 in the Flags field in the IP Header for packets transmitted
	//    by this channel. (Default = 010b  don't fragment)
	//  - [4:0] - reserved
	Flags uint8

	// data 3
	//  - [7:5] - Precedence (Default = 000b)
	Precedence uint8

	// data 3
	//  - [4:1] - Type of Service (Default = 1000b, 'minimize delay')
	//  - [0] - reserved
	TOS uint8
}

func (param *LanConfigParam_IPv4HeaderParams) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv4HeaderParams, 0, 0
}

func (param *LanConfigParam_IPv4HeaderParams) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.TTL = data[0]
	param.Flags = data[1]
	param.Precedence = data[2] >> 5
	param.TOS = data[2] & 0x1f

	return nil
}

func (param *LanConfigParam_IPv4HeaderParams) Pack() []byte {
	out := make([]byte, 3)

	out[0] = param.TTL
	out[1] = param.Flags
	out[2] = (param.Precedence << 5) | (param.TOS & 0x1f)

	return out
}

func (param *LanConfigParam_IPv4HeaderParams) Format() string {
	return fmt.Sprintf("TTL=%#02x Flags=%#02x Precedence=%#02x TOS=%#02x",
		param.TTL, param.Flags, param.Precedence, param.TOS)
}

type LanConfigParam_PrimaryRMCPPort struct {
	Port uint16
}

func (param *LanConfigParam_PrimaryRMCPPort) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_PrimaryRMCPPort, 0, 0
}

func (param *LanConfigParam_PrimaryRMCPPort) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}

	param.Port, _, _ = unpackUint16L(data, 0)

	return nil
}

func (param *LanConfigParam_PrimaryRMCPPort) Pack() []byte {
	out := make([]byte, 2)

	packUint16L(param.Port, out, 0)

	return out
}

func (param *LanConfigParam_PrimaryRMCPPort) Format() string {
	return fmt.Sprintf("%d", param.Port)
}

type LanConfigParam_SecondaryRMCPPort struct {
	Port uint16
}

func (param *LanConfigParam_SecondaryRMCPPort) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_SecondaryRMCPPort, 0, 0
}

func (param *LanConfigParam_SecondaryRMCPPort) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}

	param.Port, _, _ = unpackUint16L(data, 0)

	return nil
}

func (param *LanConfigParam_SecondaryRMCPPort) Pack() []byte {
	out := make([]byte, 2)

	packUint16L(param.Port, out, 0)

	return out
}

func (param *LanConfigParam_SecondaryRMCPPort) Format() string {
	return fmt.Sprintf("%d", param.Port)
}

type LanConfigParam_ARPControl struct {
	ARPResponseEnabled   bool
	GratuitousARPEnabled bool
}

func (param *LanConfigParam_ARPControl) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_ARPControl, 0, 0
}

func (param *LanConfigParam_ARPControl) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.ARPResponseEnabled = isBit1Set(data[0])
	param.GratuitousARPEnabled = isBit0Set(data[0])
	return nil
}

func (param *LanConfigParam_ARPControl) Pack() []byte {
	out := make([]byte, 1)
	b := uint8(0)
	b = setOrClearBit1(b, param.ARPResponseEnabled)
	b = setOrClearBit0(b, param.GratuitousARPEnabled)
	out[0] = b

	return out
}

func (param *LanConfigParam_ARPControl) Format() string {
	return fmt.Sprintf("ARP Responses %s, Gratuitous ARP %s",
		formatBool(param.ARPResponseEnabled, "enabled", "disabled"),
		formatBool(param.GratuitousARPEnabled, "enabled", "disabled"))
}

type LanConfigParam_GratuitousARPInterval struct {
	// Gratuitous ARP interval in 500 millisecond increments. 0-based.
	MilliSec uint32
}

func (param *LanConfigParam_GratuitousARPInterval) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_GratuitousARPInterval, 0, 0
}

func (param *LanConfigParam_GratuitousARPInterval) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.MilliSec = uint32(data[0]) * 500
	return nil
}

func (param *LanConfigParam_GratuitousARPInterval) Pack() []byte {
	out := make([]byte, 1)
	out[0] = uint8(param.MilliSec / 500)
	return out
}

func (param *LanConfigParam_GratuitousARPInterval) Format() string {
	return fmt.Sprintf("%.1f seconds", float64(param.MilliSec/1000.0))
}

type LanConfigParam_DefaultGatewayIP struct {
	IP net.IP
}

func (param *LanConfigParam_DefaultGatewayIP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_DefaultGatewayIP, 0, 0
}

func (param *LanConfigParam_DefaultGatewayIP) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.IP = net.IPv4(data[0], data[1], data[2], data[3])

	return nil
}

func (param *LanConfigParam_DefaultGatewayIP) Pack() []byte {
	out := make([]byte, 4)

	out[0] = param.IP[0]
	out[1] = param.IP[1]
	out[2] = param.IP[2]
	out[3] = param.IP[3]

	return out
}

func (param *LanConfigParam_DefaultGatewayIP) Format() string {
	return param.IP.String()
}

type LanConfigParam_DefaultGatewayMAC struct {
	MAC net.HardwareAddr
}

func (param *LanConfigParam_DefaultGatewayMAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_DefaultGatewayMAC, 0, 0
}

func (param *LanConfigParam_DefaultGatewayMAC) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}

	param.MAC = net.HardwareAddr(data[0:6])

	return nil
}

func (param *LanConfigParam_DefaultGatewayMAC) Pack() []byte {
	out := make([]byte, 6)

	out[0] = param.MAC[0]
	out[1] = param.MAC[1]
	out[2] = param.MAC[2]
	out[3] = param.MAC[3]
	out[4] = param.MAC[4]
	out[5] = param.MAC[5]

	return out
}

func (param *LanConfigParam_DefaultGatewayMAC) Format() string {
	return param.MAC.String()
}

type LanConfigParam_BackupGatewayIP struct {
	IP net.IP
}

func (param *LanConfigParam_BackupGatewayIP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_BackupGatewayIP, 0, 0
}

func (param *LanConfigParam_BackupGatewayIP) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.IP = net.IPv4(data[0], data[1], data[2], data[3])

	return nil
}

func (param *LanConfigParam_BackupGatewayIP) Pack() []byte {
	out := make([]byte, 4)

	out[0] = param.IP[0]
	out[1] = param.IP[1]
	out[2] = param.IP[2]
	out[3] = param.IP[3]

	return out
}

func (param *LanConfigParam_BackupGatewayIP) Format() string {
	return param.IP.String()
}

type LanConfigParam_BackupGatewayMAC struct {
	MAC net.HardwareAddr
}

func (param *LanConfigParam_BackupGatewayMAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_BackupGatewayMAC, 0, 0
}

func (param *LanConfigParam_BackupGatewayMAC) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}

	param.MAC = net.HardwareAddr(data[0:6])

	return nil
}

func (param *LanConfigParam_BackupGatewayMAC) Pack() []byte {
	out := make([]byte, 6)

	out[0] = param.MAC[0]
	out[1] = param.MAC[1]
	out[2] = param.MAC[2]
	out[3] = param.MAC[3]
	out[4] = param.MAC[4]
	out[5] = param.MAC[5]

	return out
}

func (param *LanConfigParam_BackupGatewayMAC) Format() string {
	return param.MAC.String()
}

type LanConfigParam_CommunityString struct {
	CommunityString CommunityString
}

func (param *LanConfigParam_CommunityString) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_CommunityString, 0, 0
}

func (param *LanConfigParam_CommunityString) Unpack(data []byte) error {
	if len(data) < 18 {
		return ErrUnpackedDataTooShortWith(len(data), 18)
	}

	var cs CommunityString
	for i := 0; i < 18; i++ {
		cs[i] = data[i]
	}
	param.CommunityString = cs

	return nil
}

func (param *LanConfigParam_CommunityString) Pack() []byte {
	return param.CommunityString[:]
}

func (param *LanConfigParam_CommunityString) Format() string {
	return string(param.CommunityString[:])
}

// Number of LAN Alert Destinations supported on this channel. (Read Only).
//
// At least one set of non-volatile destination information is required if LAN alerting is supported.
//
// Additional non-volatile destination parameters can optionally be provided for supporting an
// alert 'call down' list policy.
//
// A maximum of fifteen (1h to Fh) non-volatile destinations are supported in this specification.
// Destination 0 is always present as a volatile destination that is used with the Alert Immediate command.
type LanConfigParam_AlertDestinationsCount struct {
	// [7:4] - reserved.
	// [3:0] - Number LAN Destinations. A count of 0h indicates LAN Alerting is not supported.
	// This value is number of non-volatile destinations.
	Count uint8
}

func (param *LanConfigParam_AlertDestinationsCount) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AlertDestinationsCount, 0, 0
}

func (param *LanConfigParam_AlertDestinationsCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Count = data[0] & 0x0f

	return nil
}

func (param *LanConfigParam_AlertDestinationsCount) Pack() []byte {
	out := make([]byte, 1)

	out[0] = param.Count & 0x0f

	return out
}

func (param *LanConfigParam_AlertDestinationsCount) Format() string {
	return fmt.Sprintf("%d", param.Count)
}

type LanConfigParam_AlertDestinationType struct {
	// Destination selector, 0 based.
	// Destination 0 is always present as a volatile destination that is used with the Alert Immediate command.
	SetSelector uint8

	// Alert Acknowledge
	//  - 0b = Unacknowledged.
	//         Alert is assumed successful if transmission occurs without error.
	//         This value is also used with Callback numbers.
	//  - 1b = Acknowledged.
	//         Alert is assumed successful only if acknowledged is returned.
	//         Note, some alert types, such as Dial Page, do not support an acknowledge
	AlertAcknowledged bool

	// Destination Type
	//  - 000b = PET Trap destination
	//  - 001b - 101b = reserved
	//  - 110b = OEM 1
	//  - 111b = OEM 2
	DestinationType uint8

	// Alert Acknowledge Timeout / Retry Interval, in seconds, 0-based (i.e. minimum
	// timeout = 1 second)
	AlertAcknowledgeTimeout uint8

	// Number of times to retry alert to given destination.
	Retries uint8
}

func (param *LanConfigParam_AlertDestinationType) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AlertDestinationType, param.SetSelector, 0
}

func (param *LanConfigParam_AlertDestinationType) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.SetSelector = data[0]
	param.AlertAcknowledged = isBit7Set(data[1])
	param.DestinationType = data[1] & 0x03
	param.AlertAcknowledgeTimeout = data[2]
	param.Retries = data[3] & 0x07

	return nil
}

func (param *LanConfigParam_AlertDestinationType) Pack() []byte {
	out := make([]byte, 4)

	out[0] = param.SetSelector

	b := param.DestinationType & 0x03
	b = setOrClearBit7(b, param.AlertAcknowledged)
	out[1] = b

	out[2] = param.AlertAcknowledgeTimeout
	out[3] = param.Retries

	return out
}

func (param *LanConfigParam_AlertDestinationType) Format() string {
	return fmt.Sprintf("%12s %2d, %v, %d, %d, %d",
		formatBool(param.SetSelector == 0, "volatile", "non-volatile"),
		param.SetSelector,
		formatBool(param.AlertAcknowledged, "acknowledged", "unacknowledged"),
		param.DestinationType,
		param.AlertAcknowledgeTimeout,
		param.Retries,
	)
}

type LanConfigParam_AlertDestinationAddress struct {
	SetSelector uint8

	IsIPv6 bool

	//  - 0b = use default gateway first, then backup gateway
	//         (Note: older implementations (errata 4 or earlier) may only send to the default gateway.)
	//  - 1b = use backup gateway
	UseBackupGateway bool

	IPv4 net.IP
	MAC  net.HardwareAddr

	IPv6 net.IP
}

func (param *LanConfigParam_AlertDestinationAddress) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AlertDestinationAddress, param.SetSelector, 0
}

func (param *LanConfigParam_AlertDestinationAddress) Unpack(data []byte) error {
	if len(data) < 13 {
		return ErrUnpackedDataTooShortWith(len(data), 13)
	}

	param.SetSelector = data[0]
	param.IsIPv6 = isBit4Set(data[1])

	if param.IsIPv6 {
		if len(data) < 18 {
			return ErrUnpackedDataTooShortWith(len(data), 18)
		}
		param.IPv6 = net.IP(data[3:18])
	} else {
		param.UseBackupGateway = isBit7Set(data[2])
		param.IPv4 = net.IP(data[3:7])
		param.MAC = net.HardwareAddr(data[7:13])

	}

	return nil
}

func (param *LanConfigParam_AlertDestinationAddress) Pack() []byte {
	out := make([]byte, 2)

	return out
}

func (param *LanConfigParam_AlertDestinationAddress) Format() string {
	return fmt.Sprintf("%12s %2d, %s, %s, %s",
		formatBool(param.SetSelector == 0, "volatile", "non-volatile"),
		param.SetSelector,
		formatBool(param.UseBackupGateway, "backup gateway", "default gateway"),
		formatBool(param.IsIPv6, "IPv6", "IPv4"),
		formatBool(param.IsIPv6, param.IPv6.String(), fmt.Sprintf("%s/%s", param.IPv4, param.MAC)),
	)
}

type LanConfigParam_VLANID struct {
	Enabled bool
	ID      uint16
}

func (param *LanConfigParam_VLANID) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_VLANID, 0, 0
}

func (param *LanConfigParam_VLANID) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}

	param.Enabled = isBit7Set(data[1])

	id := uint16(data[1]) & 0x0f
	id <<= 12
	id |= uint16(data[0])
	param.ID = id

	return nil
}

func (param *LanConfigParam_VLANID) Pack() []byte {
	out := make([]byte, 2)

	out[0] = byte(param.ID & 0xff)

	b := byte(param.ID >> 8)
	b = setOrClearBit7(b, param.Enabled)
	out[1] = b

	return out
}

func (param *LanConfigParam_VLANID) Format() string {
	return formatBool(param.Enabled, fmt.Sprintf("%d", param.ID), "disabled")
}

type LanConfigParam_VLANPriority struct {
	Priority uint8
}

func (param *LanConfigParam_VLANPriority) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_VLANPriority, 0, 0
}

func (param *LanConfigParam_VLANPriority) Pack() []byte {
	out := make([]byte, 1)

	out[0] = param.Priority & 0x07

	return out
}

func (param *LanConfigParam_VLANPriority) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Priority = data[0] & 0x07

	return nil
}

func (param *LanConfigParam_VLANPriority) Format() string {
	return fmt.Sprintf("%#2x", param.Priority)
}

type LanConfigParam_CipherSuitesSupport struct {
	// Cipher Suite Entry count. Number of Cipher Suite entries, 1-based, 10h max.
	Count uint8
}

func (param *LanConfigParam_CipherSuitesSupport) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_CipherSuitesSupport, 0, 0
}

func (param *LanConfigParam_CipherSuitesSupport) Pack() []byte {
	out := make([]byte, 1)

	out[0] = param.Count & 0x1f

	return out
}

func (param *LanConfigParam_CipherSuitesSupport) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Count = data[0] & 0x1f

	return nil
}

func (param *LanConfigParam_CipherSuitesSupport) Format() string {
	return fmt.Sprintf("%d", param.Count)
}

type LanConfigParam_CipherSuitesID struct {
	IDs [16]CipherSuiteID
}

func (param *LanConfigParam_CipherSuitesID) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_CipherSuitesID, 0, 0
}

func (param *LanConfigParam_CipherSuitesID) Pack() []byte {
	out := make([]uint8, 17)
	out[0] = 0

	for i, id := range param.IDs {
		out[i+1] = uint8(id)
	}

	return out
}

func (param *LanConfigParam_CipherSuitesID) Unpack(data []byte) error {
	if len(data) > 17 {
		data = data[:17]
	}

	for i, v := range data {
		if i == 0 {
			// first byte is reserved
			continue
		}

		param.IDs[i-1] = CipherSuiteID(v)
	}

	return nil
}

func (param *LanConfigParam_CipherSuitesID) Format() string {
	ss := make([]string, 0)
	for i, v := range param.IDs[:] {
		if i != 0 && v == 0 {
			// Only the first ID can be CipherSuiteID0, all other 0s means empty slot.
			continue
		}
		ss = append(ss, strconv.Itoa(int(v)))
	}

	return strings.Join(ss, ",")
}

type LanConfigParam_CipherSuitesPrivLevel struct {
	PrivLevels [16]PrivilegeLevel
}

func (param *LanConfigParam_CipherSuitesPrivLevel) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_CipherSuitesPrivLevel, 0, 0
}

func (param *LanConfigParam_CipherSuitesPrivLevel) Pack() []byte {
	out := make([]byte, 9)
	out[0] = 0

	for i := 0; i < 8; i++ {
		o := byte(param.PrivLevels[2*i] & 0x0f)
		o |= byte(param.PrivLevels[2*i+1] & 0x0f)

		out[i+1] = o
	}

	return out
}

func (param *LanConfigParam_CipherSuitesPrivLevel) Unpack(data []byte) error {
	if len(data) < 9 {
		return ErrUnpackedDataTooShortWith(len(data), 9)
	}

	for i, v := range data[0:9] {
		if i == 0 {
			// first byte is reserved
			continue
		}

		param.PrivLevels[2*i-2] = PrivilegeLevel(v & 0x0f)
		param.PrivLevels[2*i-1] = PrivilegeLevel(v & 0xf0 >> 4)
	}

	return nil
}

func (param *LanConfigParam_CipherSuitesPrivLevel) Format() string {
	ss := []string{}
	for _, v := range param.PrivLevels[:] {
		ss = append(ss, v.Symbol())
	}

	return fmt.Sprintf(`%s
            :     X=Cipher Suite Unused
            :     c=CALLBACK
            :     u=USER
            :     o=OPERATOR
            :     a=ADMIN
            :     O=OEM`,
		strings.Join(ss, ""))
}

type LanConfigParam_AlertDestinationVLAN struct {
	// data 1 - Set Selector = Destination Selector.
	//  - [7:4] - reserved
	//  - [3:0] - Destination selector.
	// Destination 0 is always present as a volatile destination that is used with the Alert Immediate command.
	SetSelector uint8

	// Address Format.
	// VLAN ID is used with this destination
	//  - 0h = VLAN ID not used with this destination
	//  - 1h = 802.1q VLAN TAG
	Enabled bool

	// data 3-4 - VLAN TAG
	//  - [7:0] - VLAN ID, least-significant byte
	//  - [11:8] - VLAN ID, most-significant nibble
	//  - [12] - CFI (Canonical Format Indicator. Set to 0b)
	//  - [15:13] - User priority (000b, typical)
	VLANID   uint16
	CFI      bool
	Priority uint8
}

func (param *LanConfigParam_AlertDestinationVLAN) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_AlertDestinationVLAN, param.SetSelector, 0
}

func (param *LanConfigParam_AlertDestinationVLAN) Pack() []byte {
	out := make([]byte, 4)

	out[0] = param.SetSelector

	b1 := uint8(0)
	b1 = setOrClearBit5(b1, param.Enabled)
	out[1] = b1

	out[2] = uint8(param.VLANID)

	b3 := uint8(param.VLANID>>8) & 0x0f
	b3 = setOrClearBit4(b3, param.CFI)
	b3 |= param.Priority << 5
	out[3] = b3

	return out
}

func (param *LanConfigParam_AlertDestinationVLAN) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.SetSelector = data[0]
	param.Enabled = isBit4Set(data[1])

	param.CFI = isBit4Set(data[3])
	param.Priority = data[3] >> 5

	param.VLANID = uint16(data[3]&0x0f) << 12
	param.VLANID |= uint16(data[2])

	return nil
}

func (param *LanConfigParam_AlertDestinationVLAN) Format() string {
	return fmt.Sprintf("%12s %d, %s, %d, %s, %d",
		formatBool(param.SetSelector == 0, "volatile", "non-volatile"),
		param.SetSelector,
		formatBool(param.Enabled, "enabled", "disabled"),
		param.VLANID,
		formatBool(param.CFI, "1", "0"),
		param.Priority,
	)
}

type LanConfigParam_BadPasswordThreshold struct {
	// Generate Session Audit Event
	//  - 0b = do not generate an event message when the user is disabled.
	//  - 1b = generate a Session Audit sensor "Invalid password disable" event message.
	GenerateSessionAuditEvent bool

	// Bad Password Threshold number
	Threshold uint8

	// Attempt Count Reset Interval.
	// The raw data occupies 2 bytes, and the unit is in tens of seconds.
	//
	// 0 means the Attempt Count Reset Interval is disabled.
	// The count of bad password attempts is retained as long as
	// the BMC remains powered and is not reinitialized.
	AttemptCountResetIntervalSec uint32

	// User Lockout Interval
	// The raw data occupies 2 bytes, and the unit is in tens of seconds.
	//
	// 0 means the User Lockout Interval is disabled.
	// If a user was automatically disabled due to the Bad Password threshold,
	// the user will remain disabled until re-enabled via the Set User Access command.
	UserLockoutIntervalSec uint32
}

func (param *LanConfigParam_BadPasswordThreshold) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_BadPasswordThreshold, 0, 0
}

func (param *LanConfigParam_BadPasswordThreshold) Pack() []byte {
	out := make([]byte, 0)

	b := uint8(0)
	b = setOrClearBit0(b, param.GenerateSessionAuditEvent)
	out[0] = b

	out[1] = param.Threshold

	resetInterval := uint16(param.AttemptCountResetIntervalSec / 10)
	lockInterval := uint16(param.UserLockoutIntervalSec / 10)
	packUint16L(resetInterval, out, 2)
	packUint16L(lockInterval, out, 4)

	return out
}

func (param *LanConfigParam_BadPasswordThreshold) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}

	param.GenerateSessionAuditEvent = isBit0Set(data[0])
	param.Threshold = data[1]

	resetInterval, _, _ := unpackUint16L(data, 2)
	lockInterval, _, _ := unpackUint16L(data, 4)
	param.AttemptCountResetIntervalSec = uint32(resetInterval) * 10
	param.UserLockoutIntervalSec = uint32(lockInterval) * 10

	return nil
}

func (param *LanConfigParam_BadPasswordThreshold) Format() string {
	return fmt.Sprintf(`
        Threshold                    : %d
        Generate Session Audit Event : %v
        Attempt Count Reset Interval : %d
        User Lockout Interval        : %d
`,
		param.Threshold,
		param.GenerateSessionAuditEvent,
		param.AttemptCountResetIntervalSec,
		param.UserLockoutIntervalSec,
	)
}

type LanConfigParam_IPv6Support struct {
	// Implementation supports IPv6 Destination Addresses for LAN Alerting.
	SupportIPv6AlertDestination bool
	// Implementation can be configured to use both IPv4 and IPv6 addresses simultaneously
	CanUseBothIPv4AndIPv6 bool
	// Implementation can be configured to use IPv6 addresses only.
	CanUseIPv6Only bool
}

func (param *LanConfigParam_IPv6Support) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6Support, 0, 0
}

func (param *LanConfigParam_IPv6Support) Pack() []byte {
	out := make([]byte, 1)

	var b byte
	b = setOrClearBit2(b, param.SupportIPv6AlertDestination)
	b = setOrClearBit1(b, param.CanUseBothIPv4AndIPv6)
	b = setOrClearBit0(b, param.CanUseIPv6Only)

	out[1] = b

	return out
}

func (param *LanConfigParam_IPv6Support) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.SupportIPv6AlertDestination = isBit2Set(data[0])
	param.CanUseBothIPv4AndIPv6 = isBit1Set(data[0])
	param.CanUseIPv6Only = isBit0Set(data[0])

	return nil
}

func (param *LanConfigParam_IPv6Support) Format() string {
	return fmt.Sprintf("%s %s %s",
		formatBool(param.SupportIPv6AlertDestination, "ipv6(supported)", "ipv6(not-supported)"),
		formatBool(param.CanUseBothIPv4AndIPv6, "ipv4-and-ipv6(supported)", "ipv4-and-ipv6(not-supported)"),
		formatBool(param.CanUseIPv6Only, "ipv6-only(supported)", "ipv6-only(not-supported)"),
	)
}

type LanConfigParam_IPv6Enables struct {
	EnableMode LanIPv6EnableMode
}

func (enableMode LanIPv6EnableMode) String() string {
	m := map[LanIPv6EnableMode]string{
		LanIPv6EnableMode_IPv6Disabled: "IPv6 disabled",
		LanIPv6EnableMode_IPv6Only:     "IPv6 only",
		LanIPv6EnableMode_IPv4AndIPv6:  "IPv4 and IPv6",
	}
	s, ok := m[enableMode]
	if ok {
		return s
	}
	return "Unknown"
}

func (param *LanConfigParam_IPv6Enables) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6Enables, 0, 0
}

func (param *LanConfigParam_IPv6Enables) Pack() []byte {
	out := make([]byte, 1)
	out[0] = uint8(param.EnableMode)
	return out
}

func (param *LanConfigParam_IPv6Enables) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.EnableMode = LanIPv6EnableMode(data[0])

	return nil
}

func (param *LanConfigParam_IPv6Enables) Format() string {
	return param.EnableMode.String()
}

type LanConfigParam_IPv6StaticTrafficClass struct {
	TrafficClass uint8
}

func (param *LanConfigParam_IPv6StaticTrafficClass) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticTrafficClass, 0, 0
}

func (param *LanConfigParam_IPv6StaticTrafficClass) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.TrafficClass
	return out
}

func (param *LanConfigParam_IPv6StaticTrafficClass) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.TrafficClass = data[0]

	return nil
}

func (param *LanConfigParam_IPv6StaticTrafficClass) Format() string {
	return fmt.Sprintf("%d", param.TrafficClass)
}

type LanConfigParam_IPv6StaticHopLimit struct {
	HopLimit uint8
}

func (param *LanConfigParam_IPv6StaticHopLimit) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticHopLimit, 0, 0
}

func (param *LanConfigParam_IPv6StaticHopLimit) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.HopLimit
	return out
}

func (param *LanConfigParam_IPv6StaticHopLimit) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.HopLimit = data[0]

	return nil
}

func (param *LanConfigParam_IPv6StaticHopLimit) Format() string {
	return fmt.Sprintf("%d", param.HopLimit)
}

type LanConfigParam_IPv6FlowLabel struct {
	// Flow Label, 20-bits, right justified, MS Byte first. Default = 0.
	//
	// Three bytes.
	//
	// If this configuration parameter is not supported, the Flow Label shall be set to 0 per [RFC2460].
	// Bits [23:20] = reserved - set to 0b.
	// see: https://datatracker.ietf.org/doc/html/rfc2460#page-25
	FlowLabel uint32
}

func (param *LanConfigParam_IPv6FlowLabel) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6FlowLabel, 0, 0
}

func (param *LanConfigParam_IPv6FlowLabel) Pack() []byte {
	out := make([]byte, 3)
	packUint24(param.FlowLabel, out, 0)
	return out
}

func (param *LanConfigParam_IPv6FlowLabel) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.FlowLabel, _, _ = unpackUint24(data, 0)

	return nil
}

func (param *LanConfigParam_IPv6FlowLabel) Format() string {
	return fmt.Sprintf("%d", param.FlowLabel)
}

type LanConfigParam_IPv6Status struct {
	// Maximum number of static IPv6 addresses for establishing connections to the BMC.
	// Note: in some implementations this may exceed the number of simultaneous sessions supported on
	// the channel. 0 indicates that static address configuration is not available.
	StaticAddressMax uint8

	// Maximum number of Dynamic (SLAAC/ DHCPv6) IPv6 addresses that can be obtained for
	// establishing connections to the BMC.
	//Note: in some implementations this may exceed the number of simultaneous sessions supported on the channel.
	// 0 = Dynamic addressing is not supported by the BMC.
	DynamicAddressMax uint8

	// data 3: -
	//  - [7:2] - reserved
	//  - [1] - 1b = SLAAC addressing is supported by the BMC
	//  - [0] - 1b = DHCPv6 addressing is supported by the BMC (optional)
	SupportSLAACAddressing  bool
	SupportDHCPv6Addressing bool
}

func (param *LanConfigParam_IPv6Status) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6Status, 0, 0
}

func (param *LanConfigParam_IPv6Status) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.StaticAddressMax = data[0]
	param.DynamicAddressMax = data[1]

	param.SupportSLAACAddressing = isBit1Set(data[2])
	param.SupportDHCPv6Addressing = isBit0Set(data[2])

	return nil
}

func (param *LanConfigParam_IPv6Status) Pack() []byte {
	out := make([]byte, 3)

	out[0] = param.StaticAddressMax
	out[1] = param.DynamicAddressMax

	var b uint8
	b = setOrClearBit1(b, param.SupportSLAACAddressing)
	b = setOrClearBit0(b, param.SupportDHCPv6Addressing)
	out[2] = b

	return out
}

func (param *LanConfigParam_IPv6Status) Format() string {
	return fmt.Sprintf("Static Addr Max: %d, Dynamic Addr Max: %d, SupportSLAAC: %v, SupportDHCPv6: %v",
		param.StaticAddressMax, param.DynamicAddressMax, param.SupportSLAACAddressing, param.SupportDHCPv6Addressing)
}

type LanConfigParam_IPv6StaticAddress struct {
	SetSelector uint8

	// Address Enabled
	//  - [7]- enable=1/disable=0
	Enabled bool

	// Address Source
	// [3:0]- source/type
	//  - 0h = Static
	//  - All other = reserved
	Source LanIPv6StaticAddressSource

	IPv6 net.IP

	PrefixLength uint8

	Status LanIPv6AddressStatus
}

func (param *LanConfigParam_IPv6StaticAddress) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticAddress, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6StaticAddress) Pack() []byte {
	out := make([]byte, 20)

	out[0] = param.SetSelector

	var b1 uint8
	b1 = setOrClearBit7(b1, param.Enabled)
	b1 |= uint8(param.Source) & 0x0f
	out[1] = b1

	// 16-byte (IPv6)
	packBytes(param.IPv6, out, 2)

	out[18] = param.PrefixLength
	out[19] = byte(param.Status)

	return out
}

func (param *LanConfigParam_IPv6StaticAddress) Unpack(data []byte) error {
	if len(data) < 20 {
		return ErrUnpackedDataTooShortWith(len(data), 20)
	}

	param.SetSelector = data[0]

	param.Enabled = isBit7Set(data[1])
	param.Source = LanIPv6StaticAddressSource(data[1] & 0x0f)

	param.IPv6 = net.IP(data[2:18])
	param.PrefixLength = data[18]
	param.Status = LanIPv6AddressStatus(data[19])

	return nil
}

func (param *LanConfigParam_IPv6StaticAddress) Format() string {
	return fmt.Sprintf("%d, Enabled: %v, Source: %d, IPv6: %s, PrefixLength: %d, Status: %s",
		param.SetSelector, param.Enabled, param.Source, param.IPv6, param.PrefixLength, param.Status)
}

type LanConfigParam_IPv6DHCPv6StaticDUIDCount struct {
	// The maximum number of 16-byte blocks that can be used for storing each DUID via
	// the IPv6 DHCPv6 Static DUIDs parameter. 1-based. Returns 0 if IPv6 Static Address
	// configuration is not supported.
	Max uint8
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUIDCount) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6StaticDUIDCount, 0, 0
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUIDCount) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.Max
	return out
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUIDCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.Max = data[0]
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUIDCount) Format() string {
	return fmt.Sprintf("%d", param.Max)
}

type LanConfigParam_IPv6DHCPv6StaticDUID struct {
	SetSelector   uint8
	BlockSelector uint8

	DUID [16]byte
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUID) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6StaticDUID, param.SetSelector, param.BlockSelector
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUID) Pack() []byte {
	out := make([]byte, 18)
	out[0] = param.SetSelector
	out[1] = param.BlockSelector
	copy(out[2:], param.DUID[:])
	return out
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUID) Unpack(data []byte) error {
	if len(data) < 18 {
		return ErrUnpackedDataTooShortWith(len(data), 18)
	}
	param.SetSelector = data[0]
	param.BlockSelector = data[1]
	copy(param.DUID[:], data[2:18])
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6StaticDUID) Format() string {
	return fmt.Sprintf("%d, %d, %x", param.SetSelector, param.BlockSelector, param.DUID)
}

type LanConfigParam_IPv6DynamicAddress struct {
	SetSelector uint8

	Enabled bool
	Source  LanIPv6DynamicAddressSource

	IPv6 net.IP

	PrefixLength uint8

	Status LanIPv6AddressStatus
}

func (param *LanConfigParam_IPv6DynamicAddress) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicAddress, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6DynamicAddress) Pack() []byte {
	out := make([]byte, 20)

	out[0] = param.SetSelector

	var b1 uint8
	b1 = setOrClearBit7(b1, param.Enabled)
	b1 |= uint8(param.Source) & 0x0f
	out[1] = b1

	// 16-byte (IPv6)
	packBytes(param.IPv6, out, 2)

	out[18] = param.PrefixLength
	out[19] = byte(param.Status)

	return out
}

func (param *LanConfigParam_IPv6DynamicAddress) Unpack(data []byte) error {
	if len(data) < 20 {
		return ErrUnpackedDataTooShortWith(len(data), 20)
	}

	param.SetSelector = data[0]

	param.Enabled = isBit7Set(data[1])
	param.Source = LanIPv6DynamicAddressSource(data[1] & 0x0f)

	param.IPv6 = net.IP(data[2:18])
	param.PrefixLength = data[18]
	param.Status = LanIPv6AddressStatus(data[19])

	return nil
}

func (param *LanConfigParam_IPv6DynamicAddress) Format() string {
	return fmt.Sprintf("%d, Enabled: %v, Source: %d, IPv6: %s, PrefixLength: %d, Status: %s",
		param.SetSelector, param.Enabled, param.Source, param.IPv6, param.PrefixLength, param.Status)
}

type LanConfigParam_IPv6DHCPv6DynamicDUIDCount struct {
	// The maximum number of 16-byte blocks that can be used for storing each DUID via
	// the IPv6 DHCPv6 Static DUIDs parameter. 1-based. Returns 0 if IPv6 Static Address
	// configuration is not supported.
	Max uint8
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUIDCount) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6DynamicDUIDCount, 0, 0
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUIDCount) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.Max
	return out
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUIDCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.Max = data[0]
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUIDCount) Format() string {
	return fmt.Sprintf("%d", param.Max)
}

type LanConfigParam_IPv6DHCPv6DynamicDUID struct {
	SetSelector   uint8
	BlockSelector uint8

	DUID [16]byte
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUID) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6DynamicDUID, param.SetSelector, param.BlockSelector
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUID) Pack() []byte {
	out := make([]byte, 18)
	out[0] = param.SetSelector
	out[1] = param.BlockSelector
	copy(out[2:], param.DUID[:])
	return out
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUID) Unpack(data []byte) error {
	if len(data) < 18 {
		return ErrUnpackedDataTooShortWith(len(data), 18)
	}
	param.SetSelector = data[0]
	param.BlockSelector = data[1]
	copy(param.DUID[:], data[2:18])
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6DynamicDUID) Format() string {
	return fmt.Sprintf("%d, %d, %x", param.SetSelector, param.BlockSelector, param.DUID)
}

type LanConfigParam_IPv6DHCPv6TimingConfigSupport struct {
	Mode LanIPv6DHCPv6TimingConfigMode
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfigSupport) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6TimingConfigSupport, 0, 0
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfigSupport) Pack() []byte {
	out := make([]byte, 1)
	out[0] = byte(param.Mode)
	return out
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfigSupport) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.Mode = LanIPv6DHCPv6TimingConfigMode(data[0])
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfigSupport) Format() string {
	return fmt.Sprintf("%s (%d)", param.Mode.String(), param.Mode)
}

type LanConfigParam_IPv6DHCPv6TimingConfig struct {
	SetSelector   uint8
	BlockSelector uint8
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfig) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DHCPv6TimingConfig, param.SetSelector, param.BlockSelector
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfig) Pack() []byte {
	out := make([]byte, 2)
	out[0] = param.SetSelector
	out[1] = param.BlockSelector
	return out
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfig) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}
	param.SetSelector = data[0]
	param.BlockSelector = data[1]
	return nil
}

func (param *LanConfigParam_IPv6DHCPv6TimingConfig) Format() string {
	return fmt.Sprintf("%d, %d", param.SetSelector, param.BlockSelector)
}

type LanConfigParam_IPv6RouterAddressConfigControl struct {
	// enable dynamic router address configuration via router advertisement messages.
	EnableDynamic bool

	// enable static router address
	EnableStatic bool
}

func (param *LanConfigParam_IPv6RouterAddressConfigControl) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6RouterAddressConfigControl, 0, 0
}

func (param *LanConfigParam_IPv6RouterAddressConfigControl) Pack() []byte {
	out := make([]byte, 1)

	var b uint8
	b = setOrClearBit1(b, param.EnableDynamic)
	b = setOrClearBit0(b, param.EnableStatic)
	out[0] = b

	return out
}

func (param *LanConfigParam_IPv6RouterAddressConfigControl) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.EnableDynamic = isBit1Set(data[0])
	param.EnableStatic = isBit0Set(data[0])
	return nil
}

func (param *LanConfigParam_IPv6RouterAddressConfigControl) Format() string {
	return fmt.Sprintf("dynamic: %s, static: %s",
		formatBool(param.EnableDynamic, "enabled", "disabled"),
		formatBool(param.EnableStatic, "enabled", "disabled"),
	)
}

type LanConfigParam_IPv6StaticRouter1IP struct {
	IPv6 net.IP
}

func (param *LanConfigParam_IPv6StaticRouter1IP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter1IP, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter1IP) Pack() []byte {
	out := make([]byte, 16)
	copy(out, param.IPv6)
	return out
}

func (param *LanConfigParam_IPv6StaticRouter1IP) Unpack(data []byte) error {
	if len(data) < 16 {
		return ErrUnpackedDataTooShortWith(len(data), 16)
	}
	param.IPv6 = data
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter1IP) Format() string {
	return param.IPv6.String()
}

type LanConfigParam_IPv6StaticRouter1MAC struct {
	MAC net.HardwareAddr
}

func (param *LanConfigParam_IPv6StaticRouter1MAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter1MAC, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter1MAC) Pack() []byte {
	out := make([]byte, 6)
	copy(out, param.MAC)
	return out
}

func (param *LanConfigParam_IPv6StaticRouter1MAC) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}
	param.MAC = data
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter1MAC) Format() string {
	return param.MAC.String()
}

type LanConfigParam_IPv6StaticRouter1PrefixLength struct {
	PrefixLength uint8
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixLength) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter1PrefixLength, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixLength) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.PrefixLength
	return out
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixLength) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.PrefixLength = data[0]
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixLength) Format() string {
	return fmt.Sprintf("%d", param.PrefixLength)
}

type LanConfigParam_IPv6StaticRouter1PrefixValue struct {
	PrefixValue [16]byte
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixValue) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter1PrefixValue, 0, 0
}
func (param *LanConfigParam_IPv6StaticRouter1PrefixValue) Pack() []byte {
	out := make([]byte, 16)
	copy(out[0:], param.PrefixValue[:])
	return out
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixValue) Unpack(data []byte) error {
	if len(data) < 16 {
		return ErrUnpackedDataTooShortWith(len(data), 16)
	}
	copy(param.PrefixValue[:], data[0:])
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter1PrefixValue) Format() string {
	return fmt.Sprintf("%s", param.PrefixValue)
}

type LanConfigParam_IPv6StaticRouter2IP struct {
	IPv6 net.IP
}

func (param *LanConfigParam_IPv6StaticRouter2IP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter2IP, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter2IP) Pack() []byte {
	out := make([]byte, 16)
	copy(out, param.IPv6)
	return out
}

func (param *LanConfigParam_IPv6StaticRouter2IP) Unpack(data []byte) error {
	if len(data) < 16 {
		return ErrUnpackedDataTooShortWith(len(data), 16)
	}
	param.IPv6 = data
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter2IP) Format() string {
	return param.IPv6.String()
}

type LanConfigParam_IPv6StaticRouter2MAC struct {
	MAC net.HardwareAddr
}

func (param *LanConfigParam_IPv6StaticRouter2MAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter2MAC, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter2MAC) Pack() []byte {
	out := make([]byte, 6)
	copy(out, param.MAC)
	return out
}

func (param *LanConfigParam_IPv6StaticRouter2MAC) Unpack(data []byte) error {
	if len(data) < 6 {
		return ErrUnpackedDataTooShortWith(len(data), 6)
	}
	param.MAC = data
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter2MAC) Format() string {
	return param.MAC.String()
}

type LanConfigParam_IPv6StaticRouter2PrefixLength struct {
	PrefixLength uint8
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixLength) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter2PrefixLength, 0, 0
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixLength) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.PrefixLength
	return out
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixLength) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.PrefixLength = data[0]
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixLength) Format() string {
	return fmt.Sprintf("%d", param.PrefixLength)
}

type LanConfigParam_IPv6StaticRouter2PrefixValue struct {
	PrefixValue [16]byte
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixValue) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6StaticRouter2PrefixValue, 0, 0
}
func (param *LanConfigParam_IPv6StaticRouter2PrefixValue) Pack() []byte {
	out := make([]byte, 16)
	copy(out[0:], param.PrefixValue[:])
	return out
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixValue) Unpack(data []byte) error {
	if len(data) < 16 {
		return ErrUnpackedDataTooShortWith(len(data), 16)
	}
	copy(param.PrefixValue[:], data[0:])
	return nil
}

func (param *LanConfigParam_IPv6StaticRouter2PrefixValue) Format() string {
	return fmt.Sprintf("%s", param.PrefixValue)
}

type LanConfigParam_IPv6DynamicRouterInfoSets struct {
	// Number of dynamic Router Address information entries
	Count uint8
}

func (param *LanConfigParam_IPv6DynamicRouterInfoSets) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterInfoCount, 0, 0
}

func (param *LanConfigParam_IPv6DynamicRouterInfoSets) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.Count
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterInfoSets) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.Count = data[0]
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterInfoSets) Format() string {
	return fmt.Sprintf("%d", param.Count)
}

type LanConfigParam_IPv6DynamicRouterInfoIP struct {
	SetSelector uint8
	IPv6        net.IP
}

func (param *LanConfigParam_IPv6DynamicRouterInfoIP) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterInfoIP, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6DynamicRouterInfoIP) Pack() []byte {
	out := make([]byte, 17)
	out[0] = param.SetSelector
	copy(out[1:], param.IPv6)
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterInfoIP) Unpack(data []byte) error {
	if len(data) < 17 {
		return ErrUnpackedDataTooShortWith(len(data), 17)
	}
	param.SetSelector = data[0]
	param.IPv6 = data[1:]
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterInfoIP) Format() string {
	return fmt.Sprintf("%d, %s", param.SetSelector, param.IPv6)
}

type LanConfigParam_IPv6DynamicRouterInfoMAC struct {
	SetSelector uint8
	MAC         net.HardwareAddr
}

func (param *LanConfigParam_IPv6DynamicRouterInfoMAC) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterInfoMAC, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6DynamicRouterInfoMAC) Pack() []byte {
	out := make([]byte, 7)
	out[0] = param.SetSelector
	copy(out[1:], param.MAC)
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterInfoMAC) Unpack(data []byte) error {
	if len(data) < 7 {
		return ErrUnpackedDataTooShortWith(len(data), 7)
	}
	param.SetSelector = data[0]
	param.MAC = data[1:]
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterInfoMAC) Format() string {
	return fmt.Sprintf("%d, %s", param.SetSelector, param.MAC)
}

type LanConfigParam_IPv6DynamicRouterInfoPrefixLength struct {
	SetSelector  uint8
	PrefixLength uint8
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixLength) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterInfoPrefixLength, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixLength) Pack() []byte {
	out := make([]byte, 2)
	out[0] = param.SetSelector
	out[1] = param.PrefixLength
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixLength) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}
	param.SetSelector = data[0]
	param.PrefixLength = data[1]
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixLength) Format() string {
	return fmt.Sprintf("%d, %d", param.SetSelector, param.PrefixLength)
}

type LanConfigParam_IPv6DynamicRouterInfoPrefixValue struct {
	SetSelector uint8
	PrefixValue [16]byte
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixValue) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterInfoPrefixValue, param.SetSelector, 0
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixValue) Pack() []byte {
	out := make([]byte, 17)
	out[0] = param.SetSelector
	copy(out[1:], param.PrefixValue[:])
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixValue) Unpack(data []byte) error {
	if len(data) < 17 {
		return ErrUnpackedDataTooShortWith(len(data), 17)
	}
	param.SetSelector = data[0]
	copy(param.PrefixValue[:], data[1:])
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterInfoPrefixValue) Format() string {
	return fmt.Sprintf("%d, %s", param.SetSelector, param.PrefixValue)
}

type LanConfigParam_IPv6DynamicRouterReceivedHopLimit struct {
	HopLimit uint8
}

func (param *LanConfigParam_IPv6DynamicRouterReceivedHopLimit) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6DynamicRouterReceivedHopLimit, 0, 0
}

func (param *LanConfigParam_IPv6DynamicRouterReceivedHopLimit) Pack() []byte {
	out := make([]byte, 1)
	out[0] = param.HopLimit
	return out
}

func (param *LanConfigParam_IPv6DynamicRouterReceivedHopLimit) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.HopLimit = data[0]
	return nil
}

func (param *LanConfigParam_IPv6DynamicRouterReceivedHopLimit) Format() string {
	return fmt.Sprintf("%d", param.HopLimit)
}

type LanConfigParam_IPv6NDSLAACTimingConfigSupport struct {
	Mode LanIPv6NDSLAACTimingConfigMode
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfigSupport) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6NDSLAACTimingConfigSupport, 0, 0
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfigSupport) Pack() []byte {
	out := make([]byte, 1)
	out[0] = byte(param.Mode)
	return out
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfigSupport) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}
	param.Mode = LanIPv6NDSLAACTimingConfigMode(data[0])
	return nil
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfigSupport) Format() string {
	return fmt.Sprintf("%s (%d)", param.Mode.String(), param.Mode)
}

type LanConfigParam_IPv6NDSLAACTimingConfig struct {
	SetSelector   uint8
	BlockSelector uint8
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfig) LanConfigParameter() (paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return LanConfigParamSelector_IPv6NDSLAACTimingConfig, param.SetSelector, param.BlockSelector
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfig) Pack() []byte {
	out := make([]byte, 2)
	out[0] = param.SetSelector
	out[1] = param.BlockSelector
	return out
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfig) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 2)
	}
	param.SetSelector = data[0]
	param.BlockSelector = data[1]
	return nil
}

func (param *LanConfigParam_IPv6NDSLAACTimingConfig) Format() string {
	return fmt.Sprintf("%d, %d", param.SetSelector, param.BlockSelector)
}
