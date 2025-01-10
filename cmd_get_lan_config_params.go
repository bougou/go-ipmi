package ipmi

import (
	"context"
	"fmt"
)

// 23.2 Get LAN Configuration Parameters Command
type GetLanConfigParamsRequest struct {
	ChannelNumber uint8
	ParamSelector LanConfigParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

type GetLanConfigParamsResponse struct {
	ParameterRevision uint8
	ParamData         []byte
}

func (req *GetLanConfigParamsRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packUint8(req.BlockSelector, out, 3)
	return out
}

func (req *GetLanConfigParamsRequest) Command() Command {
	return CommandGetLanConfigParams
}

func (res *GetLanConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
	}
}

func (res *GetLanConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.ParameterRevision, _, _ = unpackUint8(msg, 0)
	res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	return nil
}

func (res *GetLanConfigParamsResponse) Format() string {
	out := `
Parameter Revision    : %d
Param Data            : %v
Length of Config Data : %d
`

	return fmt.Sprintf(out, res.ParameterRevision, res.ParamData, len(res.ParamData))
}

func (c *Client) GetLanConfigParams(ctx context.Context, channelNumber uint8, paramSelector LanConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetLanConfigParamsResponse, err error) {
	request := &GetLanConfigParamsRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &GetLanConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetLanConfigParamsFor get the lan config for a specific parameter.
//
// The param is a pointer to a struct that implements the LanConfigParameter interface.
func (c *Client) GetLanConfigParamsFor(ctx context.Context, channelNumber uint8, param LanConfigParameter) error {
	paramSelector, setSelector, blockSelector := param.LanConfigParamSelector()
	c.Debugf(">> Get LanConfigParam for paramSelector (%d) %s, setSelector %d, blockSelector %d\n", uint8(paramSelector), paramSelector, setSelector, blockSelector)

	response, err := c.GetLanConfigParams(ctx, channelNumber, paramSelector, setSelector, blockSelector)
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
	lanConfig := &LanConfig{
		SetInProgress:         &LanConfigParam_SetInProgress{},
		AuthTypeSupport:       &LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:       &LanConfigParam_AuthTypeEnables{},
		IP:                    &LanConfigParam_IP{},
		IPSource:              &LanConfigParam_IPSource{},
		MAC:                   &LanConfigParam_MAC{},
		SubnetMask:            &LanConfigParam_SubnetMask{},
		IPHeaderParams:        &LanConfigParam_IPv4HeaderParams{},
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

	if err := c.GetLanConfigFor(ctx, channelNumber, lanConfig); err != nil {
		return nil, err
	}

	return lanConfig, nil
}

func (c *Client) GetLanConfigFull(ctx context.Context, channelNumber uint8) (*LanConfig, error) {
	lanConfig := &LanConfig{
		SetInProgress:                     &LanConfigParam_SetInProgress{},
		AuthTypeSupport:                   &LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:                   &LanConfigParam_AuthTypeEnables{},
		IP:                                &LanConfigParam_IP{},
		IPSource:                          &LanConfigParam_IPSource{},
		MAC:                               &LanConfigParam_MAC{},
		SubnetMask:                        &LanConfigParam_SubnetMask{},
		IPHeaderParams:                    &LanConfigParam_IPv4HeaderParams{},
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

	if err := c.GetLanConfigFor(ctx, channelNumber, lanConfig); err != nil {
		return nil, err
	}

	return lanConfig, nil
}

// GetLanConfigFor get the lan config partially.
// You initialize the LanConfig struct and ONLY initialize the fields you want to get.
func (c *Client) GetLanConfigFor(ctx context.Context, channelNumber uint8, lanConfig *LanConfig) error {
	if lanConfig == nil {
		return nil
	}

	// If the err is a ResponseError and the completion code wrapped
	// in ResponseError can be safely ignored
	var canIgnoreError = func(err error) error {
		if respErr, ok := err.(*ResponseError); ok {
			cc := respErr.CompletionCode()

			switch cc {
			case
				0x80: // parameter not supported

				return nil
			}

		}
		return err
	}

	if lanConfig.SetInProgress != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.SetInProgress); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.AuthTypeSupport != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.AuthTypeSupport); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.AuthTypeEnables != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.AuthTypeEnables); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IP != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IP); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPSource != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPSource); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.MAC != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.MAC); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.SubnetMask != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.SubnetMask); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPHeaderParams != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPHeaderParams); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.PrimaryRMCPPort != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.PrimaryRMCPPort); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.SecondaryRMCPPort != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.SecondaryRMCPPort); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.ARPControl != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.ARPControl); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.GratuitousARPInterval != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.GratuitousARPInterval); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.DefaultGatewayIP != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.DefaultGatewayIP); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.DefaultGatewayMAC != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.DefaultGatewayMAC); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.BackupGatewayIP != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.BackupGatewayIP); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.BackupGatewayMAC != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.BackupGatewayMAC); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.CommunityString != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.CommunityString); canIgnoreError(err) != nil {
			return err
		}
	}

	alertDestinationsCount := uint8(0)
	if lanConfig.AlertDestinationsCount != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.AlertDestinationsCount); canIgnoreError(err) != nil {
			return err
		}
		alertDestinationsCount = lanConfig.AlertDestinationsCount.Count
	}

	if lanConfig.AlertDestinationTypes != nil {

		if len(lanConfig.AlertDestinationTypes) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1

			lanConfig.AlertDestinationTypes = make([]*LanConfigParam_AlertDestinationType, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.AlertDestinationTypes[i] = &LanConfigParam_AlertDestinationType{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationType := range lanConfig.AlertDestinationTypes {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, alertDestinationType); err != nil {
				return err
			}
		}
	}

	if lanConfig.AlertDestinationAddresses != nil {

		if len(lanConfig.AlertDestinationAddresses) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1

			lanConfig.AlertDestinationAddresses = make([]*LanConfigParam_AlertDestinationAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.AlertDestinationAddresses[i] = &LanConfigParam_AlertDestinationAddress{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationAddress := range lanConfig.AlertDestinationAddresses {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, alertDestinationAddress); err != nil {
				return err
			}
		}
	}

	if lanConfig.VLANID != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.VLANID); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.VLANPriority != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.VLANPriority); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.CipherSuitesSupport != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.CipherSuitesSupport); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.CipherSuitesID != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.CipherSuitesID); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.CipherSuitesPrivLevel != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.CipherSuitesPrivLevel); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.AlertDestinationVLANs != nil {
		if len(lanConfig.AlertDestinationVLANs) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1

			lanConfig.AlertDestinationVLANs = make([]*LanConfigParam_AlertDestinationVLAN, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.AlertDestinationVLANs[i] = &LanConfigParam_AlertDestinationVLAN{
					SetSelector: i,
				}
			}
		}

		for _, alertDestinationVLAN := range lanConfig.AlertDestinationVLANs {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, alertDestinationVLAN); err != nil {
				return err
			}
		}
	}

	if lanConfig.BadPasswordThreshold != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.BadPasswordThreshold); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6Support != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6Support); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6Enables != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6Enables); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticTrafficClass != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticTrafficClass); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticHopLimit != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticHopLimit); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6FlowLabel != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6FlowLabel); canIgnoreError(err) != nil {
			return err
		}
	}

	var ipv6StaticAddressMax uint8
	var ipv6DynamicAddressMax uint8
	if lanConfig.IPv6Status != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6Status); canIgnoreError(err) != nil {
			return err
		}
		ipv6StaticAddressMax = lanConfig.IPv6Status.StaticAddressMax
		ipv6DynamicAddressMax = lanConfig.IPv6Status.DynamicAddressMax
	}

	if lanConfig.IPv6StaticAddresses != nil {
		if len(lanConfig.IPv6StaticAddresses) == 0 && ipv6StaticAddressMax > 0 {
			count := ipv6StaticAddressMax

			lanConfig.IPv6StaticAddresses = make([]*LanConfigParam_IPv6StaticAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6StaticAddresses[i] = &LanConfigParam_IPv6StaticAddress{
					SetSelector: i,
				}
			}
		}

		for _, ipv6StaticAddress := range lanConfig.IPv6StaticAddresses {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6StaticAddress); err != nil {
				return err
			}
		}
	}

	var ipv6DHCPv6StaticDUIDCount uint8
	if lanConfig.IPv6DHCPv6StaticDUIDCount != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6DHCPv6StaticDUIDCount); canIgnoreError(err) != nil {
			return err
		}

		ipv6DHCPv6StaticDUIDCount = lanConfig.IPv6DHCPv6StaticDUIDCount.Max
	}

	if lanConfig.IPv6DHCPv6StaticDUIDs != nil {
		if len(lanConfig.IPv6DHCPv6StaticDUIDs) == 0 && ipv6DHCPv6StaticDUIDCount > 0 {
			count := ipv6DHCPv6StaticDUIDCount

			lanConfig.IPv6DHCPv6StaticDUIDs = make([]*LanConfigParam_IPv6DHCPv6StaticDUID, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DHCPv6StaticDUIDs[i] = &LanConfigParam_IPv6DHCPv6StaticDUID{
					SetSelector: i,
				}
			}
		}

		for _, ipv6DHCPv6StaticDUID := range lanConfig.IPv6DHCPv6StaticDUIDs {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DHCPv6StaticDUID); err != nil {
				return err
			}
		}
	}

	if lanConfig.IPv6DynamicAddresses != nil {
		if len(lanConfig.IPv6DynamicAddresses) == 0 && ipv6DynamicAddressMax > 0 {
			count := ipv6DynamicAddressMax

			lanConfig.IPv6DynamicAddresses = make([]*LanConfigParam_IPv6DynamicAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DynamicAddresses[i] = &LanConfigParam_IPv6DynamicAddress{
					SetSelector: i,
				}
			}
		}

		for _, ipv6DynamicAddress := range lanConfig.IPv6DynamicAddresses {
			if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DynamicAddress); err != nil {
				return err
			}
		}
	}

	var ipv6DHCPv6DynamicDUIDCount uint8
	if lanConfig.IPv6DHCPv6DynamicDUIDCount != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6DHCPv6DynamicDUIDCount); canIgnoreError(err) != nil {
			return err
		}

		ipv6DHCPv6DynamicDUIDCount = lanConfig.IPv6DHCPv6DynamicDUIDCount.Max
	}

	if lanConfig.IPv6DHCPv6DynamicDUIDs != nil {
		if len(lanConfig.IPv6DHCPv6DynamicDUIDs) == 0 && ipv6DHCPv6DynamicDUIDCount > 0 {
			// Todo

			// 	count := ipv6DHCPv6DynamicDUIDCount

			// 	lanConfig.IPv6DHCPv6DynamicDUIDs = make([]*LanConfigParam_IPv6DHCPv6DynamicDUID, count)
			// 	for i := uint8(0); i < count; i++ {
			// 		lanConfig.IPv6DHCPv6DynamicDUIDs[i] = &LanConfigParam_IPv6DHCPv6DynamicDUID{
			// 			SetSelector: i,
			// 		}
			// 	}

			// 	for _, ipv6DHCPv6DynamicDUID := range lanConfig.IPv6DHCPv6DynamicDUIDs {
			// 		if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DHCPv6DynamicDUID); err != nil {
			// 			return err
			// 		}
			// 	}
		}
	}

	if lanConfig.IPv6DHCPv6TimingConfigSupport != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6DHCPv6TimingConfigSupport); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6DHCPv6TimingConfig != nil {
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
		// 		if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DHCPv6TimingConfig); err != nil {
		// 			return err
		// 		}
		// 	}
		// }
	}

	if lanConfig.IPv6RouterAddressConfigControl != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6RouterAddressConfigControl); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter1IP != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter1IP); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter1MAC != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter1MAC); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter1PrefixLength != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter1PrefixLength); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter1PrefixValue != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter1PrefixValue); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter2IP != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter2IP); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter2MAC != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter2MAC); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter2PrefixLength != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter2PrefixLength); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6StaticRouter2PrefixValue != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6StaticRouter2PrefixValue); canIgnoreError(err) != nil {
			return err
		}
	}

	var ipv6DynamicRouterInfoCount uint8
	if lanConfig.IPv6DynamicRouterInfoSets != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6DynamicRouterInfoSets); canIgnoreError(err) != nil {
			return err
		}

		ipv6DynamicRouterInfoCount = lanConfig.IPv6DynamicRouterInfoSets.Count
	}

	if lanConfig.IPv6DynamicRouterInfoIP != nil {
		if len(lanConfig.IPv6DynamicRouterInfoIP) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount

			lanConfig.IPv6DynamicRouterInfoIP = make([]*LanConfigParam_IPv6DynamicRouterInfoIP, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DynamicRouterInfoIP[i] = &LanConfigParam_IPv6DynamicRouterInfoIP{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoIP := range lanConfig.IPv6DynamicRouterInfoIP {
				if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DynamicRouterInfoIP); err != nil {
					return err
				}
			}
		}
	}

	if lanConfig.IPv6DynamicRouterInfoMAC != nil {
		if len(lanConfig.IPv6DynamicRouterInfoMAC) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount

			lanConfig.IPv6DynamicRouterInfoMAC = make([]*LanConfigParam_IPv6DynamicRouterInfoMAC, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DynamicRouterInfoMAC[i] = &LanConfigParam_IPv6DynamicRouterInfoMAC{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoMAC := range lanConfig.IPv6DynamicRouterInfoMAC {
				if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DynamicRouterInfoMAC); err != nil {
					return err
				}
			}
		}
	}

	if lanConfig.IPv6DynamicRouterInfoPrefixLength != nil {
		if len(lanConfig.IPv6DynamicRouterInfoPrefixLength) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount

			lanConfig.IPv6DynamicRouterInfoPrefixLength = make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixLength, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DynamicRouterInfoPrefixLength[i] = &LanConfigParam_IPv6DynamicRouterInfoPrefixLength{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoPrefixLength := range lanConfig.IPv6DynamicRouterInfoPrefixLength {
				if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DynamicRouterInfoPrefixLength); err != nil {
					return err
				}
			}
		}
	}

	if lanConfig.IPv6DynamicRouterInfoPrefixValue != nil {
		if len(lanConfig.IPv6DynamicRouterInfoPrefixValue) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount

			lanConfig.IPv6DynamicRouterInfoPrefixValue = make([]*LanConfigParam_IPv6DynamicRouterInfoPrefixValue, count)
			for i := uint8(0); i < count; i++ {
				lanConfig.IPv6DynamicRouterInfoPrefixValue[i] = &LanConfigParam_IPv6DynamicRouterInfoPrefixValue{
					SetSelector: i,
				}
			}

			for _, ipv6DynamicRouterInfoPrefixValue := range lanConfig.IPv6DynamicRouterInfoPrefixValue {
				if err := c.GetLanConfigParamsFor(ctx, channelNumber, ipv6DynamicRouterInfoPrefixValue); err != nil {
					return err
				}
			}
		}
	}

	if lanConfig.IPv6DynamicRouterReceivedHopLimit != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6DynamicRouterReceivedHopLimit); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6NDSLAACTimingConfigSupport != nil {
		if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfig.IPv6NDSLAACTimingConfigSupport); canIgnoreError(err) != nil {
			return err
		}
	}

	if lanConfig.IPv6NDSLAACTimingConfig != nil {
		// Todo

	}
	return nil
}
