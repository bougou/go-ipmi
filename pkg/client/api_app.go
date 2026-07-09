package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/cmd/app"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) SetWatchdogTimer(ctx context.Context) (response *app.SetWatchdogTimerResponse, err error) {
	request := &app.SetWatchdogTimerRequest{}
	response = &app.SetWatchdogTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// The command selects which of the BMC-supported authentication types the Remote Console would like to use,
// and a username that selects which set of user information should be used for the session
func (c *Client) GetSessionChallenge(ctx context.Context) (response *app.GetSessionChallengeResponse, err error) {
	username := padBytes(c.Username, 16, 0x00)
	request := &app.GetSessionChallengeRequest{
		AuthType: c.session.authType,
		Username: array16(username),
	}

	response = &app.GetSessionChallengeResponse{}
	err = c.Exchange(ctx, request, response)
	if err != nil {
		return
	}

	c.session.v15.sessionID = response.TemporarySessionID
	c.session.v15.challenge = response.Challenge

	return
}

func (c *Client) GetSystemGUID(ctx context.Context) (response *app.GetSystemGUIDResponse, err error) {
	request := &app.GetSystemGUIDRequest{}
	response = &app.GetSystemGUIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) EnableMessageChannelReceive(ctx context.Context, channelNumber uint8, channelState uint8) (response *app.EnableMessageChannelReceiveResponse, err error) {
	request := &app.EnableMessageChannelReceiveRequest{
		ChannelNumber: channelNumber,
		ChannelState:  channelState,
	}
	response = &app.EnableMessageChannelReceiveResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemInfoParam(ctx context.Context, paramSelector types.SystemInfoParamSelector, setSelector uint8, blockSelector uint8) (response *app.GetSystemInfoParamResponse, err error) {
	request := &app.GetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &app.GetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemInfoParamFor(ctx context.Context, param types.SystemInfoParameter) error {
	if types.IsNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, setSelector, blockSelector := param.SystemInfoParameter()
	response, err := c.GetSystemInfoParam(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSystemInfoParams(ctx context.Context) (*types.SystemInfoParams, error) {
	systemInfo := &types.SystemInfoParams{
		SetInProgress:          &types.SystemInfoParam_SetInProgress{},
		SystemFirmwareVersions: make([]*types.SystemInfoParam_SystemFirmwareVersion, 0),
		SystemNames:            make([]*types.SystemInfoParam_SystemName, 0),
		PrimaryOSNames:         make([]*types.SystemInfoParam_PrimaryOSName, 0),
		OSNames:                make([]*types.SystemInfoParam_OSName, 0),
		OSVersions:             make([]*types.SystemInfoParam_OSVersion, 0),
		BMCURLs:                make([]*types.SystemInfoParam_BMCURL, 0),
		ManagementURLs:         make([]*types.SystemInfoParam_ManagementURL, 0),
	}

	if err := c.GetSystemInfoParamsFor(ctx, systemInfo); err != nil {
		return nil, err
	}

	return systemInfo, nil
}

func (c *Client) GetSystemInfoParamsFor(ctx context.Context, params *types.SystemInfoParams) error {
	if params == nil {
		return nil
	}

	canIgnore := buildCanIgnoreFn(
		0x80,
	)

	calculateSetsCount := func(blockData []byte) uint8 {

		if len(blockData) < 2 {
			return 0
		}

		stringLength := uint8(blockData[1])
		totalLength := 2 + stringLength

		return (totalLength-1)/16 + 1
	}

	if err := c.GetSystemInfoParamFor(ctx, params.SetInProgress); canIgnore(err) != nil {
		return err
	}

	if params.SystemFirmwareVersions != nil {
		if len(params.SystemFirmwareVersions) == 0 {
			p := &types.SystemInfoParam_SystemFirmwareVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.SystemFirmwareVersions = make([]*types.SystemInfoParam_SystemFirmwareVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_SystemFirmwareVersion{
					SetSelector: i,
				}
				params.SystemFirmwareVersions[i] = p
			}
		}

		for _, param := range params.SystemFirmwareVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.SystemNames != nil {
		if len(params.SystemNames) == 0 {
			p := &types.SystemInfoParam_SystemName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.SystemNames = make([]*types.SystemInfoParam_SystemName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_SystemName{
					SetSelector: i,
				}
				params.SystemNames[i] = p
			}
		}

		for _, param := range params.SystemNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.PrimaryOSNames != nil {
		if len(params.PrimaryOSNames) == 0 {
			p := &types.SystemInfoParam_PrimaryOSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.PrimaryOSNames = make([]*types.SystemInfoParam_PrimaryOSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_PrimaryOSName{
					SetSelector: i,
				}
				params.PrimaryOSNames[i] = p
			}
		}

		for _, param := range params.PrimaryOSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSNames != nil {
		if len(params.OSNames) == 0 {
			p := &types.SystemInfoParam_OSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.OSNames = make([]*types.SystemInfoParam_OSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_OSName{
					SetSelector: i,
				}
				params.OSNames[i] = p
			}
		}

		for _, param := range params.OSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSVersions != nil {
		if len(params.OSVersions) == 0 {
			p := &types.SystemInfoParam_OSVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.OSVersions = make([]*types.SystemInfoParam_OSVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_OSVersion{
					SetSelector: i,
				}
				params.OSVersions[i] = p
			}
		}

		for _, param := range params.OSVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.BMCURLs != nil {
		if len(params.BMCURLs) == 0 {
			p := &types.SystemInfoParam_BMCURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.BMCURLs = make([]*types.SystemInfoParam_BMCURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_BMCURL{
					SetSelector: i,
				}
				params.BMCURLs[i] = p
			}
		}

		for _, param := range params.BMCURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.ManagementURLs != nil {
		if len(params.ManagementURLs) == 0 {
			p := &types.SystemInfoParam_ManagementURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.ManagementURLs = make([]*types.SystemInfoParam_ManagementURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_ManagementURL{
					SetSelector: i,
				}
				params.ManagementURLs[i] = p
			}
		}

		for _, param := range params.ManagementURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	systemInfoParams, err := c.GetSystemInfoParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetSystemInfo failed, err: %w", err)
	}
	return systemInfoParams.ToSystemInfo(), nil
}

func (c *Client) GetUserAccess(ctx context.Context, channelNumber uint8, userID uint8) (response *app.GetUserAccessResponse, err error) {
	request := &app.GetUserAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &app.GetUserAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUsers(ctx context.Context, channelNumber uint8) ([]*app.User, error) {
	var users = make([]*app.User, 0)

	var userID uint8 = 1
	var username string
	for {
		res, err := c.GetUserAccess(ctx, channelNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("GetUserAccess for userID %d failed, err: %w", userID, err)
		}

		res2, err := c.GetUsername(ctx, userID)
		if err != nil {
			if respErr, ok := types.IsResponseError(err); ok {
				if respErr.CompletionCode() == types.CompletionCodeRequestDataFieldInvalid {

					username = ""
				}
			} else {
				return nil, fmt.Errorf("GetUsername for userID %d failed, err: %w", userID, err)
			}
		} else {
			username = res2.Username
		}

		user := &app.User{
			ID:                   userID,
			Name:                 username,
			Callin:               !res.CallbackOnly,
			LinkAuthEnabled:      res.LinkAuthEnabled,
			IPMIMessagingEnabled: res.IPMIMessagingEnabled,
			MaxPrivLevel:         res.MaxPrivLevel,
		}
		users = append(users, user)

		if userID >= res.MaxUsersIDCount {
			break
		}
		userID += 1
	}

	return users, nil
}

func (c *Client) SendMessage(ctx context.Context, channelNumber uint8, authenticated bool, encrypted bool, trackMask uint8, data []byte) (response *app.SendMessageResponse, err error) {
	request := &app.SendMessageRequest{
		ChannelNumber: channelNumber,
		Authenticated: authenticated,
		Encrypted:     encrypted,
		TrackMask:     trackMask,
		MessageData:   data,
	}
	response = &app.SendMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetUserAccess(ctx context.Context, request *app.SetUserAccessRequest) (response *app.SetUserAccessResponse, err error) {
	response = &app.SetUserAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSessionInfo(ctx context.Context, request *app.GetSessionInfoRequest) (response *app.GetSessionInfoResponse, err error) {
	response = &app.GetSessionInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetCurrentSessionInfo(ctx context.Context) (response *app.GetSessionInfoResponse, err error) {
	request := &app.GetSessionInfoRequest{
		SessionIndex: 0x00,
	}
	response = &app.GetSessionInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUserPayloadAccess(ctx context.Context, channelNumber uint8, userID uint8) (response *app.GetUserPayloadAccessResponse, err error) {
	request := &app.GetUserPayloadAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &app.GetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUsername(ctx context.Context, userID uint8) (response *app.GetUsernameResponse, err error) {
	request := &app.GetUsernameRequest{
		UserID: userID,
	}
	response = &app.GetUsernameResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetConfigurableCommands(ctx context.Context, channelNumber uint8, commandRangeMask app.CommandRangeMask, netFn types.NetFn, lun uint8, code uint8, oemIANA uint32) (response *app.GetConfigurableCommandsResponse, err error) {
	request := &app.GetConfigurableCommandsRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEMIANA:          oemIANA,
	}
	response = &app.GetConfigurableCommandsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDeviceID(ctx context.Context) (response *app.GetDeviceIDResponse, err error) {
	request := &app.GetDeviceIDRequest{}
	response = &app.GetDeviceIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSelfTestResults(ctx context.Context) (response *app.GetSelfTestResultsResponse, err error) {
	request := &app.GetSelfTestResultsRequest{}
	response = &app.GetSelfTestResultsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetCommandSubfunctionEnables(ctx context.Context, request *app.SetCommandSubfunctionEnablesRequest) (response *app.SetCommandSubfunctionEnablesResponse, err error) {
	response = &app.SetCommandSubfunctionEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSessionPrivilegeLevel(ctx context.Context, privilegeLevel types.PrivilegeLevel) (response *app.SetSessionPrivilegeLevelResponse, err error) {
	request := &app.SetSessionPrivilegeLevelRequest{
		PrivilegeLevel: privilegeLevel,
	}
	response = &app.SetSessionPrivilegeLevelResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetUserPassword(ctx context.Context, userID uint8, password string, stored20 bool) (response *app.SetUserPasswordResponse, err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: app.PasswordOperationSetPassword,
		Password:  password,
	}
	response = &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) TestUserPassword(ctx context.Context, userID uint8, password string, stored20 bool) (response *app.SetUserPasswordResponse, err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: app.PasswordOperationTestPassword,
		Password:  password,
	}
	response = &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) DisableUser(ctx context.Context, userID uint8) (err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Operation: app.PasswordOperationDisableUser,
	}
	response := &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return err
}

func (c *Client) EnableUser(ctx context.Context, userID uint8) (err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Operation: app.PasswordOperationEnableUser,
	}
	response := &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return err
}

func (c *Client) SetSystemInfoParam(ctx context.Context, paramSelector types.SystemInfoParamSelector, paramData []byte) (response *app.SetSystemInfoParamResponse, err error) {
	request := &app.SetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &app.SetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSystemInfoParamFor(ctx context.Context, param types.SystemInfoParameter) error {
	if types.IsNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, _, _ := param.SystemInfoParameter()
	paramData := param.Pack()
	_, err := c.SetSystemInfoParam(ctx, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetCommandSupport(ctx context.Context, channelNumber uint8, commandRangeMask app.CommandRangeMask, netFn types.NetFn, lun uint8, code uint8, oemIANA uint32) (response *app.GetCommandSupportResponse, err error) {
	request := &app.GetCommandSupportRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEMIANA:          oemIANA,
	}
	response = &app.GetCommandSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetChannelAccess(ctx context.Context, channelNumber uint8, accessOption types.ChannelAccessOption) (response *app.GetChannelAccessResponse, err error) {
	request := &app.GetChannelAccessRequest{
		ChannelNumber: channelNumber,
		AccessOption:  accessOption,
	}
	response = &app.GetChannelAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetChannelAccess(ctx context.Context, request *app.SetChannelAccessRequest) (response *app.SetChannelAccessResponse, err error) {
	response = &app.SetChannelAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetCommandEnables(ctx context.Context, request *app.SetCommandEnablesRequest) (response *app.SetCommandEnablesResponse, err error) {
	response = &app.SetCommandEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetUserPayloadAccess(ctx context.Context, payloadType types.PayloadType, payloadInstance uint8) (response *app.SetUserPayloadAccessResponse, err error) {
	request := &app.SetUserPayloadAccessRequest{}
	response = &app.SetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetCommandSubfunctionSupport(ctx context.Context, channelNumber uint8, netFn types.NetFn, lun uint8, code uint8, oemIANA uint32) (response *app.GetCommandSubfunctionSupportResponse, err error) {
	request := &app.GetCommandSubfunctionSupportRequest{
		ChannelNumber:  channelNumber,
		NetFn:          netFn,
		LUN:            lun,
		CodeForNetFn2C: code,
		OEMIANA:        oemIANA,
	}
	response = &app.GetCommandSubfunctionSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetBTInterfaceCapabilities(ctx context.Context) (response *app.GetBTInterfaceCapabilitiesResponse, err error) {
	request := &app.GetBTInterfaceCapabilitiesRequest{}
	response = &app.GetBTInterfaceCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetWatchdogTimer(ctx context.Context) (response *app.GetWatchdogTimerResponse, err error) {
	request := &app.GetWatchdogTimerRequest{}
	response = &app.GetWatchdogTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) MasterWriteRead(ctx context.Context, request *app.MasterWriteReadRequest) (*app.MasterWriteReadResponse, error) {
	response := &app.MasterWriteReadResponse{}
	err := c.Exchange(ctx, request, response)
	return response, err
}

func (c *Client) SetUsername(ctx context.Context, userID uint8, username string) (response *app.SetUsernameResponse, err error) {
	request := &app.SetUsernameRequest{
		UserID:   userID,
		Username: username,
	}
	response = &app.SetUsernameResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ColdReset(ctx context.Context) (err error) {
	request := &app.ColdResetRequest{}
	response := &app.ColdResetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetNetFnSupport(ctx context.Context, channelNumber uint8) (response *app.GetNetFnSupportResponse, err error) {
	request := &app.GetNetFnSupportRequest{
		ChannelNumber: channelNumber,
	}
	response = &app.GetNetFnSupportResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// ActivateSession is only used for IPMI v1.5
func (c *Client) ActivateSession(ctx context.Context) (response *app.ActivateSessionResponse, err error) {
	request := &app.ActivateSessionRequest{
		AuthTypeForSession: c.session.authType,
		MaxPrivilegeLevel:  c.maxPrivilegeLevel,
		Challenge:          c.session.v15.challenge,

		InitialOutboundSequenceNumber: randomUint32(),
	}
	c.session.v15.outSeq = request.InitialOutboundSequenceNumber

	response = &app.ActivateSessionResponse{}

	err = c.Exchange(ctx, request, response)
	if err != nil {
		return
	}
	c.session.v15.active = true
	c.session.v15.preSession = false

	c.session.v15.sessionID = response.SessionID

	c.session.v15.inSeq = response.InitialInboundSequenceNumber

	return
}

func (c *Client) CloseSession(ctx context.Context, request *app.CloseSessionRequest) (response *app.CloseSessionResponse, err error) {
	response = &app.CloseSessionResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetConfigurableCommandSubfunctions(ctx context.Context, request *app.GetConfigurableCommandSubfunctionsRequest) (response *app.GetConfigurableCommandSubfunctionsResponse, err error) {
	response = &app.GetConfigurableCommandSubfunctionsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetCommandSubfunctionEnables(ctx context.Context, request *app.GetCommandSubfunctionEnablesRequest) (response *app.GetCommandSubfunctionEnablesResponse, err error) {
	response = &app.GetCommandSubfunctionEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDeviceGUID(ctx context.Context) (response *app.GetDeviceGUIDResponse, err error) {
	request := &app.GetDeviceGUIDRequest{}
	response = &app.GetDeviceGUIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// If the device supports a "manufacturing test mode", this command is reserved to turn that mode on.
func (c *Client) ManufacturingTestOn(ctx context.Context) (response *app.ManufacturingTestOnResponse, err error) {
	request := &app.ManufacturingTestOnRequest{}
	response = &app.ManufacturingTestOnResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetCommandEnables(ctx context.Context, channelNumber uint8, commandRangeMask app.CommandRangeMask, netFn types.NetFn, lun uint8, code uint8, oemIANA uint32) (response *app.GetCommandEnablesResponse, err error) {
	request := &app.GetCommandEnablesRequest{
		ChannelNumber:    channelNumber,
		CommandRangeMask: commandRangeMask,
		NetFn:            netFn,
		LUN:              lun,
		CodeForNetFn2C:   code,
		OEMIANA:          oemIANA,
	}
	response = &app.GetCommandEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetChannelAuthenticationCapabilities is used to retrieve capability information
// about the channel that the message is delivered over, or for a particular channel.
// The command returns the authentication algorithm support for the given privilege level.
//
// This command is sent in unauthenticated (clear) format.
//
// When activating a session, the privilege level passed in this command will
// normally be the same Requested Maximum Privilege level that will be used
// for a subsequent Activate Session command.
func (c *Client) GetChannelAuthenticationCapabilities(ctx context.Context, channelNumber uint8, privilegeLevel types.PrivilegeLevel) (response *app.GetChannelAuthenticationCapabilitiesResponse, err error) {
	request := &app.GetChannelAuthenticationCapabilitiesRequest{
		IPMIv20Extended:       true,
		ChannelNumber:         channelNumber,
		MaximumPrivilegeLevel: privilegeLevel,
	}

	response = &app.GetChannelAuthenticationCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	if err != nil {
		return
	}

	if !response.AnonymousLoginEnabled {
		if c.Username == "" {
			return nil, fmt.Errorf("anonymous login is not enabled, username (%s) is empty", c.Username)
		}
	}

	c.session.authType = response.ChooseAuthType()

	return
}

func (c *Client) GetSystemInterfaceCapabilities(ctx context.Context, interfaceType app.SystemInterfaceType) (response *app.GetSystemInterfaceCapabilitiesResponse, err error) {
	request := &app.GetSystemInterfaceCapabilitiesRequest{
		SystemInterfaceType: interfaceType,
	}
	response = &app.GetSystemInterfaceCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ResetWatchdogTimer(ctx context.Context) (response *app.ResetWatchdogTimerResponse, err error) {
	request := &app.ResetWatchdogTimerRequest{}
	response = &app.ResetWatchdogTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) WarmReset(ctx context.Context) (err error) {
	request := &app.WarmResetRequest{}
	response := &app.WarmResetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command can be executed prior to establishing a session with the BMC.
// The command is used to look up what authentication, integrity, and confidentiality algorithms are supported.
// The algorithms are used in combination as 'Cipher Suites'.
// This command only applies to implementations that support IPMI v2.0/RMCP+ sessions.
func (c *Client) GetChannelCipherSuites(ctx context.Context, channelNumber uint8, index uint8) (response *app.GetChannelCipherSuitesResponse, err error) {
	request := &app.GetChannelCipherSuitesRequest{
		ChannelNumber: channelNumber,
		PayloadType:   types.PayloadTypeIPMI,
		ListIndex:     index,
	}
	response = &app.GetChannelCipherSuitesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetAllChannelCipherSuites(ctx context.Context, channelNumber uint8) ([]types.CipherSuiteRecord, error) {
	var index uint8 = 0
	var cipherSuitesData = make([]byte, 0)
	for ; index < app.MaxCipherSuiteListIndex; index++ {
		res, err := c.GetChannelCipherSuites(ctx, channelNumber, index)
		if err != nil {
			return nil, fmt.Errorf("cmd GetChannelCipherSuites failed, err: %w", err)
		}
		cipherSuitesData = append(cipherSuitesData, res.CipherSuiteRecords...)
		if len(res.CipherSuiteRecords) < 16 {
			break
		}
	}

	c.DebugBytes("cipherSuitesData", cipherSuitesData, 16)
	return app.ParseCipherSuitesData(cipherSuitesData)
}

func (c *Client) GetChannelInfo(ctx context.Context, channelNumber uint8) (response *app.GetChannelInfoResponse, err error) {
	request := &app.GetChannelInfoRequest{
		ChannelNumber: channelNumber,
	}
	response = &app.GetChannelInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
