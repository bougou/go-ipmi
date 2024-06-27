package ipmi

import (
	"fmt"
)

// 13.22 RAKP Message 3
type RAKPMessage3 struct {
	// Selected by remote console. Used by remote console to help match
	// responses up with requests.
	MessageTag uint8

	// Identifies the status of the previous message.
	RmcpStatusCode RmcpStatusCode

	// The Managed System's Session ID for this session, returned by the managed system on the previous RMCP+ Open Session Response message.
	ManagedSystemSessionID uint32

	// An integrity check value over the relevant items specified by the RAKP
	// authentication algorithm identified in RAKP Message 1.
	// The size of this field depends on the specific authentication algorithm.
	//
	// This field may be 0 bytes (absent) for some algorithms (e.g. RAKP-none).
	KeyExchangeAuthenticationCode []byte
}

type RAKPMessage4 struct {
	authAlg AuthAlg

	MessageTag uint8

	RmcpStatusCode RmcpStatusCode

	MgmtConsoleSessionID uint32

	// An integrity check value over the relevant items specified by
	// the RAKP authentication algorithm that was identified in RAKP Message 1.
	//
	// The size of this field depends on the specific authentication algorithm.
	//
	// For example, the RAKP-HMAC-SHA1 specifies that an HMACSHA1-96 algorithm be used for calculating this field.
	// See Section 13.28
	// Authentication, Integrity, and Confidentiality Algorithm Numbers for info on
	// the algorithm to be used for this field.
	//
	// This field may be 0 bytes (absent) for some authentication algorithms (e.g. RAKP-none)
	IntegrityCheckValue []byte
}

func (req *RAKPMessage3) Command() Command {
	return CommandNone
}

func (req *RAKPMessage3) Pack() []byte {
	var msg = make([]byte, 8+len(req.KeyExchangeAuthenticationCode))
	packUint8(req.MessageTag, msg, 0)
	packUint8(uint8(req.RmcpStatusCode), msg, 1)
	packUint16(0, msg, 2) // reserved
	packUint32L(req.ManagedSystemSessionID, msg, 4)
	packBytes(req.KeyExchangeAuthenticationCode, msg, 8)
	return msg
}

func (res *RAKPMessage4) Unpack(msg []byte) error {
	authCodeLen := 0
	switch res.authAlg {
	case AuthAlgRAKP_None:
		// nothing need to do
	case AuthAlgRAKP_HMAC_MD5:
		// need to copy 16 bytes
		authCodeLen = 16
	case AuthAlgRAKP_HMAC_SHA1:
		// need to copy 12 bytes
		authCodeLen = 12
	case AuthAlgRAKP_HMAC_SHA256:
		authCodeLen = 16
	default:
	}

	if len(msg) < 8+authCodeLen {
		return ErrUnpackedDataTooShortWith(len(msg), 8+authCodeLen)
	}

	res.MessageTag, _, _ = unpackUint8(msg, 0)
	b1, _, _ := unpackUint8(msg, 1)
	res.RmcpStatusCode = RmcpStatusCode(b1)
	res.MgmtConsoleSessionID, _, _ = unpackUint32L(msg, 4)
	res.IntegrityCheckValue, _, _ = unpackBytes(msg, 8, authCodeLen)
	return nil
}

func (*RAKPMessage4) CompletionCodes() map[uint8]string {
	// no command-specific cc
	return map[uint8]string{}
}

func (res *RAKPMessage4) Format() string {
	return fmt.Sprintf("%v", res)
}

// authAlg is used to parse the returned RAKPMessage4 message
func (c *Client) RAKPMessage3() (response *RAKPMessage4, err error) {
	// create session integrity key
	sik, err := c.generate_sik()
	if err != nil {
		err = fmt.Errorf("generate sik failed, err: %s", err)
		return
	}
	c.session.v20.sik = sik

	k1, err := c.generate_k1()
	if err != nil {
		err = fmt.Errorf("generate k1 failed, err: %s", err)
		return
	}
	c.session.v20.k1 = k1

	k2, err := c.generate_k2()
	if err != nil {
		err = fmt.Errorf("generate k2 failed, err: %s", err)
		return
	}
	c.session.v20.k2 = k2

	authCode, err := c.generate_rakp3_authcode()
	if err != nil {
		return nil, fmt.Errorf("generate rakp3 auth code failed, err: %s", err)
	}

	request := &RAKPMessage3{
		MessageTag:                    0,
		RmcpStatusCode:                RmcpStatusCode(c.session.v20.rakp2ReturnCode),
		ManagedSystemSessionID:        c.session.v20.bmcSessionID,
		KeyExchangeAuthenticationCode: authCode,
	}

	response = &RAKPMessage4{
		authAlg: c.session.v20.authAlg,
	}
	c.session.v20.state = SessionStateRakp3Sent

	err = c.Exchange(request, response)
	if err != nil {
		return nil, err
	}

	if _, err = c.ValidateRAKP4(response); err != nil {
		return nil, fmt.Errorf("validate rakp4 failed, err: %s", err)
	}

	c.session.v20.state = SessionStateActive

	return response, nil
}

func (c *Client) ValidateRAKP4(response *RAKPMessage4) (bool, error) {
	if response.RmcpStatusCode != RmcpStatusCodeNoErrors {
		return false, fmt.Errorf("rakp4 status code not ok, %x", response.RmcpStatusCode)
	}
	// verify
	if c.session.v20.consoleSessionID != response.MgmtConsoleSessionID {
		return false, fmt.Errorf("session not activated")
	}

	authCode, err := c.generate_rakp4_authcode()
	if err != nil {
		return false, fmt.Errorf("generate rakp4 auth code failed, err: %s", err)
	}

	c.DebugBytes("rakp4 console computed authcode", authCode, 16)
	c.DebugBytes("rakp4 bmc returned authcode", response.IntegrityCheckValue, 16)

	if !isByteSliceEqual(response.IntegrityCheckValue, authCode) {
		return false, fmt.Errorf("rakp4 returned integrity check not passed, console mac %0x, bmc mac: %0x", authCode, response.IntegrityCheckValue)
	}
	return true, nil
}
