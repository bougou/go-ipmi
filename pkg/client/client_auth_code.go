package client

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
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

// When the HMAC-SHA1-96 Integrity Algorithm is used the resulting AuthCode field is 12 bytes (96 bits).
// When the HMAC-SHA256-128 and HMAC-MD5-128 Integrity Algorithms are used the resulting AuthCode field is 16-bytes (128 bits).
func (c *Client) genIntegrityAuthCode(input []byte) ([]byte, error) {
	return crypto.SessionIntegrityAuthCode(c.session.v20.integrityAlg, input, c.session.v20.k1, c.Password)
}

// sik (Session Integrity Key)
// Both the remote console and the managed system generate sik by using
// the same hmackey and hmac data, so they should be same.
// see 13.31
func (c *Client) generate_sik() ([]byte, error) {
	var hmacKey []byte
	// hmacKey should use 160-bit key Kg
	// and Kuid is used in place of Kg if "one-key" logins are being used.
	if len(c.session.v20.bmcKey) != 0 {
		hmacKey = c.session.v20.bmcKey
	} else {
		hmacKey = crypto.PadPassword20([]byte(c.Password))
	}
	c.DebugBytes("sik mac key", hmacKey, 16)

	b, err := crypto.DeriveSIK(
		c.session.v20.authAlg,
		c.session.v20.consoleRand[:],
		c.session.v20.bmcRand[:],
		c.session.v20.role,
		c.Username,
		hmacKey,
	)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %w", err)
	}

	c.DebugBytes("sik mac computed by the remote console:", b, 16)
	return b, nil
}

// see 13.32 Generating Additional Keying Material
func (c *Client) generate_k1() ([]byte, error) {
	if c.session.v20.sik == nil {
		return nil, fmt.Errorf("sik not exists, generate sik first")
	}
	b, err := crypto.DeriveK1(c.session.v20.authAlg, c.session.v20.sik)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %w", err)
	}
	c.DebugBytes("generated k1:", b, 16)
	return b, nil
}

// see 13.32 Generating Additional Keying Material
func (c *Client) generate_k2() ([]byte, error) {
	if c.session.v20.sik == nil {
		return nil, fmt.Errorf("sik not exists, generate sik first")
	}
	b, err := crypto.DeriveK2(c.session.v20.authAlg, c.session.v20.sik)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %w", err)
	}
	c.DebugBytes("generated k2:", b, 16)
	return b, nil
}

// used for verify rakp2
func (c *Client) generate_rakp2_authcode() ([]byte, error) {
	c.DebugBytes("bmc rand", c.session.v20.bmcRand[:], 16)

	hmacKey := crypto.PadPassword20([]byte(c.Password))
	c.DebugBytes("rakp2 authcode key", hmacKey, 16)

	b, err := crypto.RAKP2AuthCode(
		c.session.v20.authAlg,
		c.session.v20.consoleSessionID,
		c.session.v20.bmcSessionID,
		c.session.v20.consoleRand[:],
		c.session.v20.bmcRand[:],
		c.session.v20.bmcGUID[:],
		c.session.v20.role,
		c.Username,
		hmacKey,
	)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %w", err)
	}

	c.DebugBytes("rakp2 generated authcode", b, 16)

	out, err := crypto.TruncateRAKP2AuthCode(c.session.v20.authAlg, b)
	if err != nil {
		return nil, err
	}
	c.DebugBytes("rakp2 used authcode", out, 16)
	return out, nil
}

// v1.5§18.15.1 / v2.0§22.17.1 AuthCode Algorithms — RAKP3 uses the RAKP auth HMAC.
func (c *Client) generate_rakp3_authcode() ([]byte, error) {
	hmacKey := crypto.PadPassword20([]byte(c.Password))
	c.DebugBytes("rakp3 auth code key", hmacKey, 16)

	b, err := crypto.RAKP3AuthCode(
		c.session.v20.authAlg,
		c.session.v20.bmcRand[:],
		c.session.v20.consoleSessionID,
		c.session.v20.role,
		c.Username,
		hmacKey,
	)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %w", err)
	}
	c.DebugBytes("rakp3 generated authcode", b, 16)
	c.DebugBytes("rakp3 used authcode", b, 16)
	return b, nil
}

// generate_rakp4_authcode computes the Integrity Check Value the remote
// console uses to verify RAKP Message 4 from the BMC (v2.0§13.28 / §13.31).
func (c *Client) generate_rakp4_authcode() ([]byte, error) {
	out, err := crypto.RAKP4ICV(
		c.session.v20.authAlg,
		c.session.v20.consoleRand[:],
		c.session.v20.bmcSessionID,
		c.session.v20.bmcGUID[:],
		c.session.v20.sik,
	)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return []byte{}, nil
	}
	c.DebugBytes("rakp4 used authcode", out, 16)
	return out, nil
}
