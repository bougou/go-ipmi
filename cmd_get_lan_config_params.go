package ipmi

import (
	"context"
	"fmt"
)

// 23.2 Get LAN Configuration Parameters Command
type GetLanConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector LanConfigParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

type GetLanConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetLanConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packUint8(req.BlockSelector, out, 3)
	return out
}

func (req *GetLanConfigParamRequest) Command() Command {
	return CommandGetLanConfigParam
}

func (res *GetLanConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *GetLanConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ParamRevision, _, _ = unpackUint8(msg, 0)
	res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetLanConfigParamResponse) Format() string {

	return "" +
		fmt.Sprintf("Parameter Revision    : %d\n", res.ParamRevision) +
		fmt.Sprintf("Param Data            : %v\n", res.ParamData) +
		fmt.Sprintf("Length of Config Data : %d\n", len(res.ParamData))
}

func (c *Client) GetLanConfigParam(ctx context.Context, channelNumber uint8, paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetLanConfigParamResponse, err error) {
	request := &GetLanConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &GetLanConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetLanConfigParamFor get the lan config for a specific parameter.
//
// The param is a pointer to a struct that implements the LanConfigParameter interface.
func (c *Client) GetLanConfigParamFor(ctx context.Context, channelNumber uint8, param LanConfigParameter) error {
	if isNilLanConfigParameter(param) {
		return nil
	}

	paramSelector, setSelector, blockSelector := param.LanConfigParameter()
	c.Debugf(">> Get LanConfigParam for paramSelector (%d) %s, setSelector %d, blockSelector %d\n", uint8(paramSelector), paramSelector, setSelector, blockSelector)

	response, err := c.GetLanConfigParam(ctx, channelNumber, paramSelector, setSelector, blockSelector)
	if err != nil {
		c.Debugf("!!! Get LanConfigParam for paramSelector (%d) %s, setSelector %d failed, err: %v\n", uint8(paramSelector), paramSelector, setSelector, err)
		return err
	}

	c.DebugBytes(fmt.Sprintf("<< Got param data for (%s[%d]) ", paramSelector.String(), paramSelector), response.ParamData, 8)
	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack lan config param (%s [%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetLanConfig(ctx context.Context, channelNumber uint8) (*LanConfig, error) {
	lanConfigParams, err := c.GetLanConfigParams(ctx, channelNumber)
	if err != nil {
		return nil, fmt.Errorf("GetLanConfigParams failed, err: %w", err)
	}

	return lanConfigParams.ToLanConfig(), nil
}

func (c *Client) GetLanConfigParams(ctx context.Context, channelNumber uint8) (*LanConfigParams, error) {
	lanConfigParams := &LanConfigParams{
		SetInProgress:         &LanConfigParam_SetInProgress{},
		AuthTypeSupport:       &LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:       &LanConfigParam_AuthTypeEnables{},
		IP:                    &LanConfigParam_IP{},
		IPSource:              &LanConfigParam_IPSource{},
		MAC:                   &LanConfigParam_MAC{},
		SubnetMask:            &LanConfigParam_SubnetMask{},
		IPv4HeaderParams:      &LanConfigParam_IPv4HeaderParams{},
		PrimaryRMCPPort:       &LanConfigParam_PrimaryRMCPPort{},
		SecondaryRMCPPort:     &LanConfigParam_SecondaryRMCPPort{},
		ARPControl:            &LanConfigParam_ARPControl{},
		GratuitousARPInterval: &LanConfigParam_GratuitousARPInterval{},
		DefaultGatewayIP:      &LanConfigParam_DefaultGatewayIP{},
		DefaultGatewayMAC:     &LanConfigParam_DefaultGatewayMAC{},
		BackupGatewayIP:       &LanConfigParam_BackupGatewayIP{},
		BackupGatewayMAC:      &LanConfigParam_BackupGatewayMAC{},
		CommunityString:       &LanConfigParam_CommunityString{},
		// AlertDestinationsCount:            &LanConfigParam_AlertDestinationsCount{},
		// AlertDestinationTypes:             make([]*LanConfigParam_AlertDestinationType, 0),
		// AlertDestinationAddresses:         make([]*LanConfigParam_AlertDestinationAddress, 0),
		VLANID:                &LanConfigParam_VLANID{},
		VLANPriority:          &LanConfigParam_VLANPriority{},
		CipherSuitesSupport:   &LanConfigParam_CipherSuitesSupport{},
		CipherSuitesID:        &LanConfigParam_CipherSuitesID{},
		CipherSuitesPrivLevel: &LanConfigParam_CipherSuitesPrivLevel{},
		AlertDestinationVLANs: make([]*LanConfigParam_AlertDestinationVLAN, 0),
		BadPasswordThreshold:  &LanConfigParam_BadPasswordThreshold{},
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
	}

	if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfigParams); err != nil {
		return nil, err
	}

	return lanConfigParams, nil
}

func (c *Client) GetLanConfigParamsFull(ctx context.Context, channelNumber uint8) (*LanConfigParams, error) {
	lanConfigParams := &LanConfigParams{
		SetInProgress:                     &LanConfigParam_SetInProgress{},
		AuthTypeSupport:                   &LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:                   &LanConfigParam_AuthTypeEnables{},
		IP:                                &LanConfigParam_IP{},
		IPSource:                          &LanConfigParam_IPSource{},
		MAC:                               &LanConfigParam_MAC{},
		SubnetMask:                        &LanConfigParam_SubnetMask{},
		IPv4HeaderParams:                  &LanConfigParam_IPv4HeaderParams{},
		PrimaryRMCPPort:                   &LanConfigParam_PrimaryRMCPPort{},
		SecondaryRMCPPort:                 &LanConfigParam_SecondaryRMCPPort{},
		ARPControl:                        &LanConfigParam_ARPControl{},
		GratuitousARPInterval:             &LanConfigParam_GratuitousARPInterval{},
		DefaultGatewayIP:                  &LanConfigParam_DefaultGatewayIP{},
		DefaultGatewayMAC:                 &LanConfigParam_DefaultGatewayMAC{},
		BackupGatewayIP:                   &LanConfigParam_BackupGatewayIP{},
		BackupGatewayMAC:                  &LanConfigParam_BackupGatewayMAC{},
		CommunityString:                   &LanConfigParam_CommunityString{},
		AlertDestinationsCount:            &LanConfigParam_AlertDestinationsCount{},
		AlertDestinationTypes:             make([]*LanConfigParam_AlertDestinationType, 0),
		AlertDestinationAddresses:         make([]*LanConfigParam_AlertDestinationAddress, 0),
		VLANID:                            &LanConfigParam_VLANID{},
		VLANPriority:                      &LanConfigParam_VLANPriority{},
		CipherSuitesSupport:               &LanConfigParam_CipherSuitesSupport{},
		CipherSuitesID:                    &LanConfigParam_CipherSuitesID{},
		CipherSuitesPrivLevel:             &LanConfigParam_CipherSuitesPrivLevel{},
		AlertDestinationVLANs:             make([]*LanConfigParam_AlertDestinationVLAN, 0),
		BadPasswordThreshold:              &LanConfigParam_BadPasswordThreshold{},
		IPv6Support:                       &LanConfigParam_IPv6Support{},
		IPv6Enables:                       &LanConfigParam_IPv6Enables{},
		IPv6StaticTrafficClass:            &LanConfigParam_IPv6StaticTrafficClass{},
		IPv6StaticHopLimit:                &LanConfigParam_IPv6StaticHopLimit{},
		IPv6FlowLabel:                     &LanConfigParam_IPv6FlowLabel{},
		IPv6Status:                        &LanConfigParam_IPv6Status{},
		IPv6StaticAddresses:               make([]*LanConfigParam_IPv6StaticAddress, 0),
		IPv6DHCPv6StaticDUIDCount:         &LanConfigParam_IPv6DHCPv6StaticDUIDCount{},
		IPv6DHCPv6StaticDUIDs:             make([]*LanConfigParam_IPv6DHCPv6StaticDUID, 0),
		IPv6DynamicAddresses:              make([]*LanConfigParam_IPv6DynamicAddress, 0),
		IPv6DHCPv6DynamicDUIDCount:        &LanConfigParam_IPv6DHCPv6DynamicDUIDCount{},
		IPv6DHCPv6DynamicDUIDs:            make([]*LanConfigParam_IPv6DHCPv6DynamicDUID, 0),
		IPv6DHCPv6TimingConfigSupport:     &LanConfigParam_IPv6DHCPv6TimingConfigSupport{},
		IPv6DHCPv6TimingConfig:            make([]*LanConfigParam_IPv6DHCPv6TimingConfig, 0),
		IPv6RouterAddressConfigControl:    &LanConfigParam_IPv6RouterAddressConfigControl{},
		IPv6StaticRouter1IP:               &LanConfigParam_IPv6StaticRouter1IP{},
		IPv6StaticRouter1MAC:              &LanConfigParam_IPv6StaticRouter1MAC{},
		IPv6StaticRouter1PrefixLength:     &LanConfigParam_IPv6StaticRouter1PrefixLength{},
		IPv6StaticRouter1PrefixValue:      &LanConfigParam_IPv6StaticRouter1PrefixValue{},
		IPv6StaticRouter2IP:               &LanConfigParam_IPv6StaticRouter2IP{},
		IPv6StaticRouter2MAC:              &LanConfigParam_IPv6StaticRouter2MAC{},
		IPv6StaticRouter2PrefixLength:     &LanConfigParam_IPv6StaticRouter2PrefixLength{},
		IPv6StaticRouter2PrefixValue:      &LanConfigParam_IPv6StaticRouter2PrefixValue{},
		IPv6DynamicRouterInfoSets:         &LanConfigParam_IPv6DynamicRouterInfoSets{},
		IPv6DynamicRouterInfoIP:           make([]*LanConfigParam_IPv6DynamicRouterInfoIP, 0),
		IPv6DynamicRouterInfoMAC:          make([]*LanConfigParam_IPv6DynamicRouterInfoMAC, 0),
		IPv6DynamicRouterInfoPrefixLength: make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixLength, 0),
		IPv6DynamicRouterInfoPrefixValue:  make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixValue, 0),
		IPv6DynamicRouterReceivedHopLimit: &LanConfigParam_IPv6DynamicRouterReceivedHopLimit{},
		IPv6NDSLAACTimingConfigSupport:    &LanConfigParam_IPv6NDSLAACTimingConfigSupport{},
		IPv6NDSLAACTimingConfig:           make([]*LanConfigParam_IPv6NDSLAACTimingConfig, 0),
	}

	if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfigParams); err != nil {
		return nil, err
	}

	return lanConfigParams, nil
}

// GetLanConfigParamsFor get the lan config params.
// You can initialize specific fields of LanConfigParams struct, which indicates to only get params for those fields.
func (c *Client) GetLanConfigParamsFor(ctx context.Context, channelNumber uint8, lanConfigParams *LanConfigParams) error {
	if lanConfigParams == nil {
		return nil
	}

	var canIgnore = buildCanIgnoreFn(
		0x80, // parameter not supported
	)

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.SetInProgress); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.AuthTypeSupport); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.AuthTypeEnables); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IP); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPSource); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.MAC); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.SubnetMask); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv4HeaderParams); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.PrimaryRMCPPort); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.SecondaryRMCPPort); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.ARPControl); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.GratuitousARPInterval); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.DefaultGatewayIP); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.DefaultGatewayMAC); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.BackupGatewayIP); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.BackupGatewayMAC); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.CommunityString); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.AlertDestinationsCount); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.AlertDestinationTypes != nil {
		param := lanConfigParams.AlertDestinationsCount
		if param == nil {
			param = &LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationTypes) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1
			lanConfigParams.AlertDestinationTypes = make([]*LanConfigParam_AlertDestinationType, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationTypes[i] = &LanConfigParam_AlertDestinationType{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationType := range lanConfigParams.AlertDestinationTypes {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, alertDestinationType); canIgnore(err) != nil {
				return err
			}
		}
	}

	if lanConfigParams.AlertDestinationAddresses != nil {
		param := lanConfigParams.AlertDestinationsCount
		if param == nil {
			param = &LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationAddresses) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1
			lanConfigParams.AlertDestinationAddresses = make([]*LanConfigParam_AlertDestinationAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationAddresses[i] = &LanConfigParam_AlertDestinationAddress{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationAddress := range lanConfigParams.AlertDestinationAddresses {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, alertDestinationAddress); canIgnore(err) != nil {
				return err
			}
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.VLANID); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.VLANPriority); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.CipherSuitesSupport); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.CipherSuitesID); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.CipherSuitesPrivLevel); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.AlertDestinationVLANs != nil {
		param := lanConfigParams.AlertDestinationsCount
		if param == nil {
			param = &LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationVLANs) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1

			lanConfigParams.AlertDestinationVLANs = make([]*LanConfigParam_AlertDestinationVLAN, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationVLANs[i] = &LanConfigParam_AlertDestinationVLAN{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationVLAN := range lanConfigParams.AlertDestinationVLANs {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, alertDestinationVLAN); canIgnore(err) != nil {
				return err
			}
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.BadPasswordThreshold); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6Support); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6Enables); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticTrafficClass); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticHopLimit); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6FlowLabel); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6Status); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6StaticAddresses != nil {
		param := lanConfigParams.IPv6Status
		if param == nil {
			param = &LanConfigParam_IPv6Status{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6StaticAddressMax := param.StaticAddressMax

		if len(lanConfigParams.IPv6StaticAddresses) == 0 && ipv6StaticAddressMax > 0 {
			count := ipv6StaticAddressMax
			lanConfigParams.IPv6StaticAddresses = make([]*LanConfigParam_IPv6StaticAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6StaticAddresses[i] = &LanConfigParam_IPv6StaticAddress{
					SetSelector: i,
				}
			}
		}

		for _, ipv6StaticAddress := range lanConfigParams.IPv6StaticAddresses {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6StaticAddress); canIgnore(err) != nil {
				return err
			}
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DHCPv6StaticDUIDCount); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6DHCPv6StaticDUIDs != nil {
		ipv6Status := lanConfigParams.IPv6DHCPv6StaticDUIDCount
		if ipv6Status == nil {
			ipv6Status = &LanConfigParam_IPv6DHCPv6StaticDUIDCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6Status); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DHCPv6StaticDUIDCount := ipv6Status.Max

		if len(lanConfigParams.IPv6DHCPv6StaticDUIDs) == 0 && ipv6DHCPv6StaticDUIDCount > 0 {
			count := ipv6DHCPv6StaticDUIDCount
			lanConfigParams.IPv6DHCPv6StaticDUIDs = make([]*LanConfigParam_IPv6DHCPv6StaticDUID, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DHCPv6StaticDUIDs[i] = &LanConfigParam_IPv6DHCPv6StaticDUID{
					SetSelector: i,
				}
			}
		}

		for _, ipv6DHCPv6StaticDUID := range lanConfigParams.IPv6DHCPv6StaticDUIDs {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DHCPv6StaticDUID); canIgnore(err) != nil {
				return err
			}
		}
	}

	if lanConfigParams.IPv6DynamicAddresses != nil {
		ipv6Status := lanConfigParams.IPv6Status
		if ipv6Status == nil {
			ipv6Status = &LanConfigParam_IPv6Status{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6Status); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicAddressMax := ipv6Status.DynamicAddressMax

		if len(lanConfigParams.IPv6DynamicAddresses) == 0 && ipv6DynamicAddressMax > 0 {
			count := ipv6DynamicAddressMax
			lanConfigParams.IPv6DynamicAddresses = make([]*LanConfigParam_IPv6DynamicAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicAddresses[i] = &LanConfigParam_IPv6DynamicAddress{
					SetSelector: i,
				}
			}
		}

		for _, ipv6DynamicAddress := range lanConfigParams.IPv6DynamicAddresses {
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DynamicAddress); canIgnore(err) != nil {
				return err
			}
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DHCPv6DynamicDUIDCount); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6DHCPv6DynamicDUIDs != nil {
		param := lanConfigParams.IPv6DHCPv6DynamicDUIDCount
		if param == nil {
			param = &LanConfigParam_IPv6DHCPv6DynamicDUIDCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DHCPv6DynamicDUIDCount := param.Max

		if len(lanConfigParams.IPv6DHCPv6DynamicDUIDs) == 0 && ipv6DHCPv6DynamicDUIDCount > 0 {
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
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DHCPv6TimingConfigSupport); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6DHCPv6TimingConfig != nil {
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
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6RouterAddressConfigControl); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter1IP); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter1MAC); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter1PrefixLength); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter1PrefixValue); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter2IP); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter2MAC); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter2PrefixLength); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6StaticRouter2PrefixValue); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DynamicRouterInfoSets); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6DynamicRouterInfoIP != nil {
		param := lanConfigParams.IPv6DynamicRouterInfoSets
		if param == nil {
			param = &LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoIP) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoIP = make([]*LanConfigParam_IPv6DynamicRouterInfoIP, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoIP[i] = &LanConfigParam_IPv6DynamicRouterInfoIP{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoIP := range lanConfigParams.IPv6DynamicRouterInfoIP {
				if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DynamicRouterInfoIP); err != nil {
					return err
				}
			}
		}
	}

	if lanConfigParams.IPv6DynamicRouterInfoMAC != nil {
		param := lanConfigParams.IPv6DynamicRouterInfoSets
		if param == nil {
			param = &LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoMAC) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoMAC = make([]*LanConfigParam_IPv6DynamicRouterInfoMAC, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoMAC[i] = &LanConfigParam_IPv6DynamicRouterInfoMAC{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoMAC := range lanConfigParams.IPv6DynamicRouterInfoMAC {
				if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DynamicRouterInfoMAC); err != nil {
					return err
				}
			}
		}
	}

	if lanConfigParams.IPv6DynamicRouterInfoPrefixLength != nil {
		param := lanConfigParams.IPv6DynamicRouterInfoSets
		if param == nil {
			param = &LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoPrefixLength) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoPrefixLength = make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixLength, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoPrefixLength[i] = &LanConfigParam_IPv6DynamicRouterInfoPrefixLength{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoPrefixLength := range lanConfigParams.IPv6DynamicRouterInfoPrefixLength {
				if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DynamicRouterInfoPrefixLength); err != nil {
					return err
				}
			}
		}
	}

	if lanConfigParams.IPv6DynamicRouterInfoPrefixValue != nil {
		param := lanConfigParams.IPv6DynamicRouterInfoSets
		if param == nil {
			param = &LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoPrefixValue) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoPrefixValue = make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixValue, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoPrefixValue[i] = &LanConfigParam_IPv6DynamicRouterInfoPrefixValue{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoPrefixValue := range lanConfigParams.IPv6DynamicRouterInfoPrefixValue {
				if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6DynamicRouterInfoPrefixValue); err != nil {
					return err
				}
			}
		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DynamicRouterReceivedHopLimit); canIgnore(err) != nil {
		return err
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6NDSLAACTimingConfigSupport); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6NDSLAACTimingConfig != nil {
		// Todo

	}
	return nil
}
