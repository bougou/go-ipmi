package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/rmcpplus"
	"github.com/bougou/go-ipmi/pkg/types"
)

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
