package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) OpenSession(ctx context.Context) (response *rmcpplus.OpenSessionResponse, err error) {
	cipherSuiteID := c.session.v20.cipherSuiteID

	authAlg, integrityAlg, cryptAlg, ok := types.GetCipherSuiteAlgorithms(cipherSuiteID)
	if !ok {
		return nil, fmt.Errorf("not valid cipher suite id: %#02x", cipherSuiteID)
	}
	c.session.v20.requestedAuthAlg = authAlg
	c.session.v20.requestedIntegrityAlg = integrityAlg
	c.session.v20.requestedEncryptAlg = cryptAlg

	// Choose our session ID for easy recognition in the packet dump
	var remoteConsoleSessionID uint32 = 0xa0a1a2a3

	authPayload, integPayload, cryptPayload := rmcpplus.NewAlgorithmPayloads(
		c.session.v20.requestedAuthAlg,
		c.session.v20.requestedIntegrityAlg,
		c.session.v20.requestedEncryptAlg,
	)

	request := &rmcpplus.OpenSessionRequest{
		MessageTag:                     0x00,
		RequestedMaximumPrivilegeLevel: 0, // Request the highest level matching proposed algorithms
		RemoteConsoleSessionID:         remoteConsoleSessionID,
		AuthenticationPayload:          authPayload,
		IntegrityPayload:               integPayload,
		ConfidentialityPayload:         cryptPayload,
	}

	response = &rmcpplus.OpenSessionResponse{}

	c.session.v20.state = types.SessionStateOpenSessionSent

	err = c.Exchange(ctx, request, response)
	if err != nil {
		return nil, fmt.Errorf("client exchange failed, err: %w", err)
	}

	c.Debug("OPEN SESSION RESPONSE", response.Format())

	if response.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		return response, types.NewRmcpStatusError(response.RmcpStatusCode)
	}

	c.session.v20.state = types.SessionStateOpenSessionReceived

	c.session.v20.authAlg = types.AuthAlg(response.AuthAlg)
	c.session.v20.integrityAlg = types.IntegrityAlg(response.IntegrityAlg)
	c.session.v20.cryptAlg = types.CryptAlg(response.CryptAlg)
	c.session.v20.consoleSessionID = response.RemoteConsoleSessionID
	c.session.v20.bmcSessionID = response.ManagedSystemSessionID

	return
}

// ValidateRAKP2 validates RAKP Message 2 returned by BMC.
func (c *Client) ValidateRAKP2(ctx context.Context, rakp2 *rmcpplus.RAKPMessage2) (bool, error) {
	if rakp2.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		return false, types.NewRmcpStatusError(rakp2.RmcpStatusCode)
	}

	if c.session.v20.consoleSessionID != rakp2.RemoteConsoleSessionID {
		return false, fmt.Errorf("session id not matched, cached console session id: %x, rakp2 returned session id: %x", c.session.v20.consoleSessionID, rakp2.RemoteConsoleSessionID)
	}

	authcode, err := c.generate_rakp2_authcode()
	if err != nil {
		return false, fmt.Errorf("generate rakp2 authcode failed, err: %w", err)
	}

	c.DebugBytes("rakp2 returned auth code", rakp2.KeyExchangeAuthenticationCode, 16)

	if !isByteSliceEqual(authcode, rakp2.KeyExchangeAuthenticationCode) {
		return false, fmt.Errorf("RAKP2 authentication code mismatch: %w", ErrRAKPAuthentication)
	}
	return true, nil
}

func (c *Client) RAKPMessage1(ctx context.Context) (response *rmcpplus.RAKPMessage2, err error) {

	c.session.v20.consoleRand = array16(crypto.RandomBytes(16))
	c.DebugBytes("console generate console random number", c.session.v20.consoleRand[:], 16)

	request := &rmcpplus.RAKPMessage1{
		MessageTag:                     0,
		ManagedSystemSessionID:         c.session.v20.bmcSessionID, // set by previous RMCP+ Open Session Request
		RemoteConsoleRandomNumber:      c.session.v20.consoleRand,
		RequestedMaximumPrivilegeLevel: c.maxPrivilegeLevel,
		NameOnlyLookup:                 true,
		UsernameLength:                 uint8(len(c.Username)),
		Username:                       []byte(c.Username),
	}

	c.session.v20.role = request.Role()

	response = &rmcpplus.RAKPMessage2{
		AuthAlg: c.session.v20.authAlg,
	}
	c.session.v20.state = types.SessionStateRakp1Sent

	err = c.Exchange(ctx, request, response)
	if err != nil {
		return nil, err
	}

	// the following fields must be set before generate_sik/generate_k1/generate_k2
	c.session.v20.rakp2ReturnCode = uint8(response.RmcpStatusCode)
	c.session.v20.bmcGUID = response.ManagedSystemGUID
	c.session.v20.bmcRand = response.ManagedSystemRandomNumber // will be used in rakp3 to generate authCode

	if _, err = c.ValidateRAKP2(ctx, response); err != nil {
		err = fmt.Errorf("validate rakp2 message failed, err: %w", err)
		return
	}

	c.session.v20.state = types.SessionStateRakp2Received

	return
}

// RAKPMessage3 sends RAKP Message 3 and validates RAKP Message 4.
func (c *Client) RAKPMessage3(ctx context.Context) (response *rmcpplus.RAKPMessage4, err error) {
	sik, err := c.generate_sik()
	if err != nil {
		err = fmt.Errorf("generate sik failed, err: %w", err)
		return
	}
	c.session.v20.sik = sik

	k1, err := c.generate_k1()
	if err != nil {
		err = fmt.Errorf("generate k1 failed, err: %w", err)
		return
	}
	c.session.v20.k1 = k1

	k2, err := c.generate_k2()
	if err != nil {
		err = fmt.Errorf("generate k2 failed, err: %w", err)
		return
	}
	c.session.v20.k2 = k2

	authCode, err := c.generate_rakp3_authcode()
	if err != nil {
		return nil, fmt.Errorf("generate rakp3 auth code failed, err: %w", err)
	}

	request := &rmcpplus.RAKPMessage3{
		MessageTag:                    0,
		RmcpStatusCode:                types.RmcpStatusCode(c.session.v20.rakp2ReturnCode),
		ManagedSystemSessionID:        c.session.v20.bmcSessionID,
		KeyExchangeAuthenticationCode: authCode,
	}

	response = &rmcpplus.RAKPMessage4{
		AuthAlg: c.session.v20.authAlg,
	}
	c.session.v20.state = types.SessionStateRakp3Sent

	err = c.Exchange(ctx, request, response)
	if err != nil {
		return nil, err
	}

	if _, err = c.ValidateRAKP4(ctx, response); err != nil {
		return nil, fmt.Errorf("validate rakp4 failed, err: %w", err)
	}

	c.session.v20.state = types.SessionStateActive

	return response, nil
}

func (c *Client) ValidateRAKP4(ctx context.Context, response *rmcpplus.RAKPMessage4) (bool, error) {
	if response.RmcpStatusCode != types.RmcpStatusCodeNoErrors {
		return false, types.NewRmcpStatusError(response.RmcpStatusCode)
	}
	if c.session.v20.consoleSessionID != response.MgmtConsoleSessionID {
		return false, fmt.Errorf("session not activated")
	}

	authCode, err := c.generate_rakp4_authcode()
	if err != nil {
		return false, fmt.Errorf("generate rakp4 auth code failed, err: %w", err)
	}

	c.DebugBytes("rakp4 console computed authcode", authCode, 16)
	c.DebugBytes("rakp4 bmc returned authcode", response.IntegrityCheckValue, 16)

	if !isByteSliceEqual(response.IntegrityCheckValue, authCode) {
		return false, fmt.Errorf("RAKP4 integrity check value does not match: %w", ErrRAKPAuthentication)
	}
	return true, nil
}
