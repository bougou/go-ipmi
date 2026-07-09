package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/cmd/transport"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) GetIPStatistics(ctx context.Context, channelNumber uint8, clearAllStatistics bool) (response *transport.GetIPStatisticsResponse, err error) {
	request := &transport.GetIPStatisticsRequest{
		ChannelNumber:      channelNumber,
		ClearAllStatistics: clearAllStatistics,
	}
	response = &transport.GetIPStatisticsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ActivatePayload(ctx context.Context, request *transport.ActivatePayloadRequest) (response *transport.ActivatePayloadResponse, err error) {
	response = &transport.ActivatePayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetChannelPayloadSupport(ctx context.Context, channelNumber uint8) (response *transport.GetChannelPayloadSupportResponse, err error) {
	request := &transport.GetChannelPayloadSupportRequest{
		ChannelNumber: channelNumber,
	}
	response = &transport.GetChannelPayloadSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPayloadActivationStatus(ctx context.Context, payloadType types.PayloadType) (response *transport.GetPayloadActivationStatusResponse, err error) {
	request := &transport.GetPayloadActivationStatusRequest{
		PayloadType: payloadType,
	}
	response = &transport.GetPayloadActivationStatusResponse{}
	response.PayloadType = request.PayloadType
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SuspendARPs(ctx context.Context, channelNumber uint8, suspendARP bool, suspendGratuitousARP bool) (response *transport.SuspendARPsResponse, err error) {
	request := &transport.SuspendARPsRequest{
		ChannelNumber:        channelNumber,
		SuspendARP:           suspendARP,
		SuspendGratuitousARP: suspendGratuitousARP,
	}
	response = &transport.SuspendARPsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSOLConfigParam(ctx context.Context, channelNumber uint8, paramSelector types.SOLConfigParamSelector, setSelector, blockSelector uint8) (response *transport.GetSOLConfigParamResponse, err error) {
	request := &transport.GetSOLConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &transport.GetSOLConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSOLConfigParamFor(ctx context.Context, channelNumber uint8, param types.SOLConfigParameter) error {
	if types.IsNilSOLConfigParameter(param) {
		return nil
	}
	paramSelector, setSelector, blockSelector := param.SOLConfigParameter()
	res, err := c.GetSOLConfigParam(ctx, channelNumber, paramSelector, setSelector, blockSelector)

	if err != nil {
		return fmt.Errorf("GetSOLConfigParam for param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(res.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSOLConfigParams(ctx context.Context, channelNumber uint8) (*types.SOLConfigParams, error) {
	solConfigParams := &types.SOLConfigParams{
		SetInProgress:      &types.SOLConfigParam_SetInProgress{},
		SOLEnable:          &types.SOLConfigParam_SOLEnable{},
		SOLAuthentication:  &types.SOLConfigParam_SOLAuthentication{},
		Character:          &types.SOLConfigParam_Character{},
		SOLRetry:           &types.SOLConfigParam_SOLRetry{},
		NonVolatileBitRate: &types.SOLConfigParam_NonVolatileBitRate{},
		VolatileBitRate:    &types.SOLConfigParam_VolatileBitRate{},
		PayloadChannel:     &types.SOLConfigParam_PayloadChannel{},
		PayloadPort:        &types.SOLConfigParam_PayloadPort{},
	}

	if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfigParams); err != nil {
		return nil, fmt.Errorf("GetSOLConfigParamFor failed, err: %w", err)
	}

	return solConfigParams, nil
}

func (c *Client) GetSOLConfigParamsFor(ctx context.Context, channelNumber uint8, solConfigParams *types.SOLConfigParams) error {
	if solConfigParams == nil {
		return nil
	}

	if solConfigParams.SetInProgress != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SetInProgress); err != nil {
			return err
		}
	}

	if solConfigParams.SOLEnable != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLEnable); err != nil {
			return err
		}
	}

	if solConfigParams.SOLAuthentication != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLAuthentication); err != nil {
			return err
		}
	}

	if solConfigParams.Character != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.Character); err != nil {
			return err
		}
	}

	if solConfigParams.SOLRetry != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLRetry); err != nil {
			return err
		}
	}

	if solConfigParams.NonVolatileBitRate != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.NonVolatileBitRate); err != nil {
			return err
		}
	}

	if solConfigParams.VolatileBitRate != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.VolatileBitRate); err != nil {
			return err
		}
	}

	if solConfigParams.PayloadChannel != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.PayloadChannel); err != nil {
			return err
		}
	}

	if solConfigParams.PayloadPort != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.PayloadPort); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetChannelPayloadVersion(ctx context.Context, channelNumber uint8, payloadType types.PayloadType) (response *transport.GetChannelPayloadVersionResponse, err error) {
	request := &transport.GetChannelPayloadVersionRequest{
		ChannelNumber: channelNumber,
		PayloadType:   payloadType,
	}
	response = &transport.GetChannelPayloadVersionResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetChannelOEMPayloadInfo(ctx context.Context, request *transport.GetChannelOEMPayloadInfoRequest) (response *transport.GetChannelOEMPayloadInfoResponse, err error) {
	response = &transport.GetChannelOEMPayloadInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetChannelSecurityKeys(ctx context.Context, request *transport.SetChannelSecurityKeysRequest) (response *transport.SetChannelSecurityKeysResponse, err error) {
	response = &transport.SetChannelSecurityKeysResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SuspendResumePayloadEncryption(ctx context.Context, payloadType types.PayloadType, payloadInstance uint8, operation transport.PayloadEncryptionOperation) (response *transport.SuspendResumePayloadEncryptionResponse, err error) {
	request := &transport.SuspendResumePayloadEncryptionRequest{
		PayloadType:     payloadType,
		PayloadInstance: payloadInstance,
		Operation:       operation,
	}
	response = &transport.SuspendResumePayloadEncryptionResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPayloadInstanceInfo(ctx context.Context, payloadType types.PayloadType, payloadInstance uint8) (response *transport.GetPayloadInstanceInfoResponse, err error) {
	request := &transport.GetPayloadInstanceInfoRequest{
		PayloadType:     payloadType,
		PayloadInstance: payloadInstance,
	}
	response = &transport.GetPayloadInstanceInfoResponse{}
	response.PayloadType = request.PayloadType
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SOLActivating(ctx context.Context, request *transport.SOLActivatingRequest) (response *transport.SOLActivatingResponse, err error) {
	response = &transport.SOLActivatingResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSOLConfigParam(ctx context.Context, channelNumber uint8, paramSelector types.SOLConfigParamSelector, paramData []byte) (response *transport.SetSOLConfigParamResponse, err error) {
	request := &transport.SetSOLConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &transport.SetSOLConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSOLConfigParamFor(ctx context.Context, channelNumber uint8, param types.SOLConfigParameter) error {
	if types.IsNilSOLConfigParameter(param) {
		return fmt.Errorf("param is nil")
	}

	paramSelector, _, _ := param.SOLConfigParameter()
	paramData := param.Pack()

	_, err := c.SetSOLConfigParam(ctx, channelNumber, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSOLConfigParam failed, err: %w", err)
	}

	return nil
}

func (c *Client) DeactivatePayload(ctx context.Context, request *transport.DeactivatePayloadRequest) (response *transport.DeactivatePayloadResponse, err error) {
	response = &transport.DeactivatePayloadResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetLanConfigParam(ctx context.Context, channelNumber uint8, paramSelector types.LanConfigParamSelector, setSelector uint8, blockSelector uint8) (response *transport.GetLanConfigParamResponse, err error) {
	request := &transport.GetLanConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &transport.GetLanConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetLanConfigParamFor get the lan config for a specific parameter.
//
// The param is a pointer to a struct that implements the LanConfigParameter interface.
func (c *Client) GetLanConfigParamFor(ctx context.Context, channelNumber uint8, param types.LanConfigParameter) error {
	if types.IsNilLanConfigParameter(param) {
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

func (c *Client) GetLanConfig(ctx context.Context, channelNumber uint8) (*types.LanConfig, error) {
	lanConfigParams, err := c.GetLanConfigParams(ctx, channelNumber)
	if err != nil {
		return nil, fmt.Errorf("GetLanConfigParams failed, err: %w", err)
	}

	return lanConfigParams.ToLanConfig(), nil
}

func (c *Client) GetLanConfigParams(ctx context.Context, channelNumber uint8) (*types.LanConfigParams, error) {
	lanConfigParams := &types.LanConfigParams{
		SetInProgress:         &types.LanConfigParam_SetInProgress{},
		AuthTypeSupport:       &types.LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:       &types.LanConfigParam_AuthTypeEnables{},
		IP:                    &types.LanConfigParam_IP{},
		IPSource:              &types.LanConfigParam_IPSource{},
		MAC:                   &types.LanConfigParam_MAC{},
		SubnetMask:            &types.LanConfigParam_SubnetMask{},
		IPv4HeaderParams:      &types.LanConfigParam_IPv4HeaderParams{},
		PrimaryRMCPPort:       &types.LanConfigParam_PrimaryRMCPPort{},
		SecondaryRMCPPort:     &types.LanConfigParam_SecondaryRMCPPort{},
		ARPControl:            &types.LanConfigParam_ARPControl{},
		GratuitousARPInterval: &types.LanConfigParam_GratuitousARPInterval{},
		DefaultGatewayIP:      &types.LanConfigParam_DefaultGatewayIP{},
		DefaultGatewayMAC:     &types.LanConfigParam_DefaultGatewayMAC{},
		BackupGatewayIP:       &types.LanConfigParam_BackupGatewayIP{},
		BackupGatewayMAC:      &types.LanConfigParam_BackupGatewayMAC{},
		CommunityString:       &types.LanConfigParam_CommunityString{},

		VLANID:                &types.LanConfigParam_VLANID{},
		VLANPriority:          &types.LanConfigParam_VLANPriority{},
		CipherSuitesSupport:   &types.LanConfigParam_CipherSuitesSupport{},
		CipherSuitesID:        &types.LanConfigParam_CipherSuitesID{},
		CipherSuitesPrivLevel: &types.LanConfigParam_CipherSuitesPrivLevel{},
		AlertDestinationVLANs: make([]*types.LanConfigParam_AlertDestinationVLAN, 0),
		BadPasswordThreshold:  &types.LanConfigParam_BadPasswordThreshold{},
	}

	if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfigParams); err != nil {
		return nil, err
	}

	return lanConfigParams, nil
}

func (c *Client) GetLanConfigParamsFull(ctx context.Context, channelNumber uint8) (*types.LanConfigParams, error) {
	lanConfigParams := &types.LanConfigParams{
		SetInProgress:                     &types.LanConfigParam_SetInProgress{},
		AuthTypeSupport:                   &types.LanConfigParam_AuthTypeSupport{},
		AuthTypeEnables:                   &types.LanConfigParam_AuthTypeEnables{},
		IP:                                &types.LanConfigParam_IP{},
		IPSource:                          &types.LanConfigParam_IPSource{},
		MAC:                               &types.LanConfigParam_MAC{},
		SubnetMask:                        &types.LanConfigParam_SubnetMask{},
		IPv4HeaderParams:                  &types.LanConfigParam_IPv4HeaderParams{},
		PrimaryRMCPPort:                   &types.LanConfigParam_PrimaryRMCPPort{},
		SecondaryRMCPPort:                 &types.LanConfigParam_SecondaryRMCPPort{},
		ARPControl:                        &types.LanConfigParam_ARPControl{},
		GratuitousARPInterval:             &types.LanConfigParam_GratuitousARPInterval{},
		DefaultGatewayIP:                  &types.LanConfigParam_DefaultGatewayIP{},
		DefaultGatewayMAC:                 &types.LanConfigParam_DefaultGatewayMAC{},
		BackupGatewayIP:                   &types.LanConfigParam_BackupGatewayIP{},
		BackupGatewayMAC:                  &types.LanConfigParam_BackupGatewayMAC{},
		CommunityString:                   &types.LanConfigParam_CommunityString{},
		AlertDestinationsCount:            &types.LanConfigParam_AlertDestinationsCount{},
		AlertDestinationTypes:             make([]*types.LanConfigParam_AlertDestinationType, 0),
		AlertDestinationAddresses:         make([]*types.LanConfigParam_AlertDestinationAddress, 0),
		VLANID:                            &types.LanConfigParam_VLANID{},
		VLANPriority:                      &types.LanConfigParam_VLANPriority{},
		CipherSuitesSupport:               &types.LanConfigParam_CipherSuitesSupport{},
		CipherSuitesID:                    &types.LanConfigParam_CipherSuitesID{},
		CipherSuitesPrivLevel:             &types.LanConfigParam_CipherSuitesPrivLevel{},
		AlertDestinationVLANs:             make([]*types.LanConfigParam_AlertDestinationVLAN, 0),
		BadPasswordThreshold:              &types.LanConfigParam_BadPasswordThreshold{},
		IPv6Support:                       &types.LanConfigParam_IPv6Support{},
		IPv6Enables:                       &types.LanConfigParam_IPv6Enables{},
		IPv6StaticTrafficClass:            &types.LanConfigParam_IPv6StaticTrafficClass{},
		IPv6StaticHopLimit:                &types.LanConfigParam_IPv6StaticHopLimit{},
		IPv6FlowLabel:                     &types.LanConfigParam_IPv6FlowLabel{},
		IPv6Status:                        &types.LanConfigParam_IPv6Status{},
		IPv6StaticAddresses:               make([]*types.LanConfigParam_IPv6StaticAddress, 0),
		IPv6DHCPv6StaticDUIDCount:         &types.LanConfigParam_IPv6DHCPv6StaticDUIDCount{},
		IPv6DHCPv6StaticDUIDs:             make([]*types.LanConfigParam_IPv6DHCPv6StaticDUID, 0),
		IPv6DynamicAddresses:              make([]*types.LanConfigParam_IPv6DynamicAddress, 0),
		IPv6DHCPv6DynamicDUIDCount:        &types.LanConfigParam_IPv6DHCPv6DynamicDUIDCount{},
		IPv6DHCPv6DynamicDUIDs:            make([]*types.LanConfigParam_IPv6DHCPv6DynamicDUID, 0),
		IPv6DHCPv6TimingConfigSupport:     &types.LanConfigParam_IPv6DHCPv6TimingConfigSupport{},
		IPv6DHCPv6TimingConfig:            make([]*types.LanConfigParam_IPv6DHCPv6TimingConfig, 0),
		IPv6RouterAddressConfigControl:    &types.LanConfigParam_IPv6RouterAddressConfigControl{},
		IPv6StaticRouter1IP:               &types.LanConfigParam_IPv6StaticRouter1IP{},
		IPv6StaticRouter1MAC:              &types.LanConfigParam_IPv6StaticRouter1MAC{},
		IPv6StaticRouter1PrefixLength:     &types.LanConfigParam_IPv6StaticRouter1PrefixLength{},
		IPv6StaticRouter1PrefixValue:      &types.LanConfigParam_IPv6StaticRouter1PrefixValue{},
		IPv6StaticRouter2IP:               &types.LanConfigParam_IPv6StaticRouter2IP{},
		IPv6StaticRouter2MAC:              &types.LanConfigParam_IPv6StaticRouter2MAC{},
		IPv6StaticRouter2PrefixLength:     &types.LanConfigParam_IPv6StaticRouter2PrefixLength{},
		IPv6StaticRouter2PrefixValue:      &types.LanConfigParam_IPv6StaticRouter2PrefixValue{},
		IPv6DynamicRouterInfoSets:         &types.LanConfigParam_IPv6DynamicRouterInfoSets{},
		IPv6DynamicRouterInfoIP:           make([]*types.LanConfigParam_IPv6DynamicRouterInfoIP, 0),
		IPv6DynamicRouterInfoMAC:          make([]*types.LanConfigParam_IPv6DynamicRouterInfoMAC, 0),
		IPv6DynamicRouterInfoPrefixLength: make([]*types.LanConfigParam_IPv6DynamicRouterInfoPrefixLength, 0),
		IPv6DynamicRouterInfoPrefixValue:  make([]*types.LanConfigParam_IPv6DynamicRouterInfoPrefixValue, 0),
		IPv6DynamicRouterReceivedHopLimit: &types.LanConfigParam_IPv6DynamicRouterReceivedHopLimit{},
		IPv6NDSLAACTimingConfigSupport:    &types.LanConfigParam_IPv6NDSLAACTimingConfigSupport{},
		IPv6NDSLAACTimingConfig:           make([]*types.LanConfigParam_IPv6NDSLAACTimingConfig, 0),
	}

	if err := c.GetLanConfigParamsFor(ctx, channelNumber, lanConfigParams); err != nil {
		return nil, err
	}

	return lanConfigParams, nil
}

// GetLanConfigParamsFor get the lan config params.
// You can initialize specific fields of LanConfigParams struct, which indicates to only get params for those fields.
func (c *Client) GetLanConfigParamsFor(ctx context.Context, channelNumber uint8, lanConfigParams *types.LanConfigParams) error {
	if lanConfigParams == nil {
		return nil
	}

	var canIgnore = buildCanIgnoreFn(
		0x80,
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
			param = &types.LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationTypes) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1
			lanConfigParams.AlertDestinationTypes = make([]*types.LanConfigParam_AlertDestinationType, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationTypes[i] = &types.LanConfigParam_AlertDestinationType{
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
			param = &types.LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationAddresses) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1
			lanConfigParams.AlertDestinationAddresses = make([]*types.LanConfigParam_AlertDestinationAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationAddresses[i] = &types.LanConfigParam_AlertDestinationAddress{
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
			param = &types.LanConfigParam_AlertDestinationsCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		alertDestinationsCount := param.Count

		if len(lanConfigParams.AlertDestinationVLANs) == 0 && alertDestinationsCount > 0 {
			count := alertDestinationsCount + 1

			lanConfigParams.AlertDestinationVLANs = make([]*types.LanConfigParam_AlertDestinationVLAN, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.AlertDestinationVLANs[i] = &types.LanConfigParam_AlertDestinationVLAN{
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
			param = &types.LanConfigParam_IPv6Status{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6StaticAddressMax := param.StaticAddressMax

		if len(lanConfigParams.IPv6StaticAddresses) == 0 && ipv6StaticAddressMax > 0 {
			count := ipv6StaticAddressMax
			lanConfigParams.IPv6StaticAddresses = make([]*types.LanConfigParam_IPv6StaticAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6StaticAddresses[i] = &types.LanConfigParam_IPv6StaticAddress{
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
			ipv6Status = &types.LanConfigParam_IPv6DHCPv6StaticDUIDCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6Status); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DHCPv6StaticDUIDCount := ipv6Status.Max

		if len(lanConfigParams.IPv6DHCPv6StaticDUIDs) == 0 && ipv6DHCPv6StaticDUIDCount > 0 {
			count := ipv6DHCPv6StaticDUIDCount
			lanConfigParams.IPv6DHCPv6StaticDUIDs = make([]*types.LanConfigParam_IPv6DHCPv6StaticDUID, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DHCPv6StaticDUIDs[i] = &types.LanConfigParam_IPv6DHCPv6StaticDUID{
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
			ipv6Status = &types.LanConfigParam_IPv6Status{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, ipv6Status); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicAddressMax := ipv6Status.DynamicAddressMax

		if len(lanConfigParams.IPv6DynamicAddresses) == 0 && ipv6DynamicAddressMax > 0 {
			count := ipv6DynamicAddressMax
			lanConfigParams.IPv6DynamicAddresses = make([]*types.LanConfigParam_IPv6DynamicAddress, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicAddresses[i] = &types.LanConfigParam_IPv6DynamicAddress{
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
			param = &types.LanConfigParam_IPv6DHCPv6DynamicDUIDCount{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DHCPv6DynamicDUIDCount := param.Max

		if len(lanConfigParams.IPv6DHCPv6DynamicDUIDs) == 0 && ipv6DHCPv6DynamicDUIDCount > 0 {

		}
	}

	if err := c.GetLanConfigParamFor(ctx, channelNumber, lanConfigParams.IPv6DHCPv6TimingConfigSupport); canIgnore(err) != nil {
		return err
	}

	if lanConfigParams.IPv6DHCPv6TimingConfig != nil {

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
			param = &types.LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoIP) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoIP = make([]*types.LanConfigParam_IPv6DynamicRouterInfoIP, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoIP[i] = &types.LanConfigParam_IPv6DynamicRouterInfoIP{
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
			param = &types.LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoMAC) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoMAC = make([]*types.LanConfigParam_IPv6DynamicRouterInfoMAC, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoMAC[i] = &types.LanConfigParam_IPv6DynamicRouterInfoMAC{
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
			param = &types.LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoPrefixLength) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoPrefixLength = make([]*types.LanConfigParam_IPv6DynamicRouterInfoPrefixLength, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoPrefixLength[i] = &types.LanConfigParam_IPv6DynamicRouterInfoPrefixLength{
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
			param = &types.LanConfigParam_IPv6DynamicRouterInfoSets{}
			if err := c.GetLanConfigParamFor(ctx, channelNumber, param); canIgnore(err) != nil {
				return err
			}
		}
		ipv6DynamicRouterInfoCount := param.Count

		if len(lanConfigParams.IPv6DynamicRouterInfoPrefixValue) == 0 && ipv6DynamicRouterInfoCount > 0 {
			count := ipv6DynamicRouterInfoCount
			lanConfigParams.IPv6DynamicRouterInfoPrefixValue = make([]*types.LanConfigParam_IPv6DynamicRouterInfoPrefixValue, count)
			for i := uint8(0); i < count; i++ {
				lanConfigParams.IPv6DynamicRouterInfoPrefixValue[i] = &types.LanConfigParam_IPv6DynamicRouterInfoPrefixValue{
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

	}
	return nil
}

func (c *Client) SetLanConfigParam(ctx context.Context, channelNumber uint8, paramSelector types.LanConfigParamSelector, configData []byte) (response *transport.SetLanConfigParamResponse, err error) {
	request := &transport.SetLanConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ParamData:     configData,
	}
	response = &transport.SetLanConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetLanConfigParamFor(ctx context.Context, channelNumber uint8, param types.LanConfigParameter) error {
	paramSelector, _, _ := param.LanConfigParameter()
	c.DebugBytes(fmt.Sprintf(">> Set param data for (%s[%d]) ", paramSelector.String(), paramSelector), param.Pack(), 8)

	if _, err := c.SetLanConfigParam(ctx, channelNumber, paramSelector, param.Pack()); err != nil {
		c.Debugf("!!! Set LanConfigParam for paramSelector (%d) %s failed, err: %v\n", uint8(paramSelector), paramSelector, err)
		return err
	}

	return nil
}
