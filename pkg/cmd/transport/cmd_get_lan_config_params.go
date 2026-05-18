package transport

import (
	"fmt"

	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// 23.2 Get LAN Configuration Parameters Command
type GetLanConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector ipmi.LanConfigParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

type GetLanConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetLanConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	ipmi.PackUint8(req.ChannelNumber, out, 0)
	ipmi.PackUint8(uint8(req.ParamSelector), out, 1)
	ipmi.PackUint8(req.SetSelector, out, 2)
	ipmi.PackUint8(req.BlockSelector, out, 3)
	return out
}

func (req *GetLanConfigParamRequest) Command() ipmi.Command {
	return ipmi.CommandGetLanConfigParam
}

func (res *GetLanConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *GetLanConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ipmi.ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ParamRevision, _, _ = ipmi.UnpackUint8(msg, 0)
	res.ParamData, _, _ = ipmi.UnpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetLanConfigParamResponse) Format() string {

	return "" +
		fmt.Sprintf("Parameter Revision    : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param Data            : %v\n", res.ParamData) +
		fmt.Sprintf("Length of Config Data : %d\n", len(res.ParamData))
}

// GetLanConfigParamFor get the lan config for a specific parameter.
//
// The param is a pointer to a struct that implements the LanConfigParameter interface.

// AlertDestinationsCount:            &LanConfigParam_AlertDestinationsCount{},
// AlertDestinationTypes:             make([]*LanConfigParam_AlertDestinationType, 0),
// AlertDestinationAddresses:         make([]*LanConfigParam_AlertDestinationAddress, 0),

// IPv6Support:                       &LanConfigParam_IPv6Support{},
// IPv6Enables:                       &LanConfigParam_IPv6Enables{},
// IPv6StaticTrafficClass:            &LanConfigParam_IPv6StaticTrafficClass{},
// IPv6StaticHopLimit:                &LanConfigParam_IPv6StaticHopLimit{},
// IPv6FlowLabel:                     &LanConfigParam_IPv6FlowLabel{},
// IPv6Status:                        &LanConfigParam_IPv6Status{},
// IPv6StaticAddresses:               make([]*LanConfigParam_IPv6StaticAddress, 0),
// IPv6DHCPv6StaticDUIDCount:         &LanConfigParam_IPv6DHCPv6StaticDUIDCount{},
// IPv6DHCPv6StaticDUIDs:             make([]*LanConfigParam_IPv6DHCPv6StaticDUID, 0),
// IPv6DynamicAddresses:              make([]*LanConfigParam_IPv6DynamicAddress, 0),
// IPv6DHCPv6DynamicDUIDCount:        &LanConfigParam_IPv6DHCPv6DynamicDUIDCount{},
// IPv6DHCPv6DynamicDUIDs:            make([]*LanConfigParam_IPv6DHCPv6DynamicDUID, 0),
// IPv6DHCPv6TimingConfigSupport:     &LanConfigParam_IPv6DHCPv6TimingConfigSupport{},
// IPv6DHCPv6TimingConfig:            make([]*LanConfigParam_IPv6DHCPv6TimingConfig, 0),
// IPv6RouterAddressConfigControl:    &LanConfigParam_IPv6RouterAddressConfigControl{},
// IPv6StaticRouter1IP:               &LanConfigParam_IPv6StaticRouter1IP{},
// IPv6StaticRouter1MAC:              &LanConfigParam_IPv6StaticRouter1MAC{},
// IPv6StaticRouter1PrefixLength:     &LanConfigParam_IPv6StaticRouter1PrefixLength{},
// IPv6StaticRouter1PrefixValue:      &LanConfigParam_IPv6StaticRouter1PrefixValue{},
// IPv6StaticRouter2IP:               &LanConfigParam_IPv6StaticRouter2IP{},
// IPv6StaticRouter2MAC:              &LanConfigParam_IPv6StaticRouter2MAC{},
// IPv6StaticRouter2PrefixLength:     &LanConfigParam_IPv6StaticRouter2PrefixLength{},
// IPv6StaticRouter2PrefixValue:      &LanConfigParam_IPv6StaticRouter2PrefixValue{},
// IPv6DynamicRouterInfoSets:         &LanConfigParam_IPv6DynamicRouterInfoSets{},
// IPv6DynamicRouterInfoIP:           make([]*LanConfigParam_IPv6DynamicRouterInfoIP, 0),
// IPv6DynamicRouterInfoMAC:          make([]*LanConfigParam_IPv6DynamicRouterInfoMAC, 0),
// IPv6DynamicRouterInfoPrefixLength: make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixLength, 0),
// IPv6DynamicRouterInfoPrefixValue:  make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixValue, 0),
// IPv6DynamicRouterReceivedHopLimit: &LanConfigParam_IPv6DynamicRouterReceivedHopLimit{},
// IPv6NDSLAACTimingConfigSupport:    &LanConfigParam_IPv6NDSLAACTimingConfigSupport{},
// IPv6NDSLAACTimingConfig:           make([]*LanConfigParam_IPv6NDSLAACTimingConfig, 0),

// GetLanConfigParamsFor get the lan config params.
// You can initialize specific fields of LanConfigParams struct, which indicates to only get params for those fields.

// parameter not supported

// Todo

// 	count := ipv6DHCPv6DynamicDUIDCount

// 	lanConfig.IPv6DHCPv6DynamicDUIDs = make([]*LanConfigParam_IPv6DHCPv6DynamicDUID, count)
// 	for i := uint8(0); i < count; i++ {
// 		lanConfig.IPv6DHCPv6DynamicDUIDs[i] = &LanConfigParam_IPv6DHCPv6DynamicDUID{
// 			SetSelector: i,
// 		}
// 	}

// 	for _, ipv6DHCPv6DynamicDUID := range lanConfig.IPv6DHCPv6DynamicDUIDs {
// 		if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DHCPv6DynamicDUID); err != nil {
// 			return err
// 		}
// 	}

// Todo

// if len(lanConfig.IPv6DHCPv6TimingConfig) == 0 && ipv6DynamicAddressMax > 0 {
// 	count := ipv6DynamicAddressMax

// 	lanConfig.IPv6DHCPv6TimingConfig = make([]*LanConfigParam_IPv6DHCPv6TimingConfig, count)
// 	for i := uint8(0); i < count; i++ {
// 		lanConfig.IPv6DHCPv6TimingConfig[i] = &LanConfigParam_IPv6DHCPv6TimingConfig{
// 			SetSelector: i,
// 		}
// 	}

// 	for _, ipv6DHCPv6TimingConfig := range lanConfig.IPv6DHCPv6TimingConfig {
// 		if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DHCPv6TimingConfig); err != nil {
// 			return err
// 		}
// 	}
// }

// Todo
