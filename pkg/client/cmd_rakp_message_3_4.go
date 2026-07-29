package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

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
		return false, fmt.Errorf("rakp4 status code not ok, %x", response.RmcpStatusCode)
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
		return false, fmt.Errorf("rakp4 returned integrity check not passed, console mac %0x, bmc mac: %0x", authCode, response.IntegrityCheckValue)
	}
	return true, nil
}
