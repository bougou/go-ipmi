package client

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

// Re-export v1.5 AuthCode input types from the shared crypto package.
type (
	AuthCodeSingleSessionInput = crypto.SingleSessionInput
	AuthCodeMultiSessionInput  = crypto.MultiSessionInput
)

func (c *Client) genAuthCodeForSingleSession() []byte {
	input := &AuthCodeSingleSessionInput{
		Password:  c.Password,
		SessionID: c.session.v15.sessionID,
		Challenge: c.session.v15.challenge[:],
	}

	authCode := input.AuthCode(c.session.authType)
	c.DebugBytes(fmt.Sprintf("authtype (%d) gen authcode", c.session.authType), authCode, 16)
	return authCode
}

// only be used for ActivateSession (IPMI v1.5)
// see v1.5§18.15.1 / v2.0§22.17.1 AuthCode Algorithms
func (c *Client) genAuthCodeForMultiSession(ipmiMsg []byte) []byte {
	input := &AuthCodeMultiSessionInput{
		Password:   c.Password,
		SessionID:  c.session.v15.sessionID,
		SessionSeq: c.session.v15.inSeq,
		IPMIData:   ipmiMsg,
	}

	authCode := input.AuthCode(c.session.authType)
	c.DebugBytes(fmt.Sprintf("authtype (%d) gen authcode", c.session.authType), authCode, 16)
	return authCode
}

func (c *Client) genSession15(rawPayload []byte) (*types.Session15, error) {
	c.lock()
	defer c.unlock()

	sessionHeader := &types.SessionHeader15{
		AuthType:      types.AuthTypeNone,
		Sequence:      0,
		SessionID:     0,
		AuthCode:      nil, // AuthCode would be filled afterward
		PayloadLength: uint8(len(rawPayload)),
	}

	if c.session.v15.preSession || c.session.v15.active {
		sessionHeader.AuthType = c.session.authType
		sessionHeader.SessionID = c.session.v15.sessionID
	}

	if c.session.v15.active {
		c.session.v15.inSeq++
		// Spec reserves sequence 0 for pre-session packets; skip on wrap
		// (same as ipmitool lan.c).
		if c.session.v15.inSeq == 0 {
			c.session.v15.inSeq = 1
		}
		sessionHeader.Sequence = c.session.v15.inSeq
	}

	if sessionHeader.AuthType != types.AuthTypeNone {
		authCode := c.genAuthCodeForMultiSession(rawPayload)
		sessionHeader.AuthCode = authCode
	}

	return &types.Session15{
		SessionHeader15: sessionHeader,
		Payload:         rawPayload,
	}, nil
}
