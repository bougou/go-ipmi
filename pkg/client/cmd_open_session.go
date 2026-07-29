package client

import (
	"context"
	"fmt"

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
		err = fmt.Errorf("rakp status code error: (%#02x) %s", uint8(response.RmcpStatusCode), response.RmcpStatusCode)
		return
	}

	c.session.v20.state = types.SessionStateOpenSessionReceived

	c.session.v20.authAlg = types.AuthAlg(response.AuthAlg)
	c.session.v20.integrityAlg = types.IntegrityAlg(response.IntegrityAlg)
	c.session.v20.cryptAlg = types.CryptAlg(response.CryptAlg)
	c.session.v20.consoleSessionID = response.RemoteConsoleSessionID
	c.session.v20.bmcSessionID = response.ManagedSystemSessionID

	return
}
