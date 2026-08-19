package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/command/app"
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

	// Seed to N-1 so the pre-increment in genSession15 emits N on the first
	// post-Activate request (spec v1.5§18.15 / v2.0§6.12.9).
	c.session.v15.inSeq = v15SeedInSeq(response.InitialInboundSequenceNumber)

	// Spec v1.5§18.15 / v2.0§22.17: the AuthType in the Activate Session response is the auth
	// type the BMC wants used for the remainder of the session — it may
	// differ from what we requested (some BMCs, e.g. ASUS, return None even
	// when the session was activated with MD5). Adopt it so subsequent
	// packets are framed with the AuthType the BMC expects; otherwise it
	// silently drops them. Matches ipmitool behaviour.
	c.session.authType = response.AuthType

	return
}

// v15SeedInSeq returns the inSeq high-water to store after Activate so the
// next pre-increment emits starting. Seq 0 is reserved; a non-compliant BMC
// returning 0 is remapped to 1 (same as GenerateInboundSeq) so the first
// packet is 1 rather than wrapping through 0xffffffff to 0.
func v15SeedInSeq(starting uint32) uint32 {
	if starting == 0 {
		starting = 1
	}
	return starting - 1
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
