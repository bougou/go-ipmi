package ipmi

import (
	"crypto/md5"
	"fmt"

	"github.com/bougou/go-ipmi/utils/md2"
)

// 22.17.1 AuthCode Algorithms
// Single Session AuthCode carried in IPMI message data for Activate Session Command
// to fill ActiveSessionRequest.Challenge
type AuthCodeSingleSessionInput struct {
	Password  string
	SessionID uint32
	Challenge []byte
}

func (a AuthCodeSingleSessionInput) AuthCode(authType AuthType) []byte {
	password := padBytes(a.Password, 16, 0x00)
	inputLength := 16 + 4 + len(a.Challenge) + 16

	var input = make([]byte, inputLength)
	packBytes(password, input, 0)
	packUint32L(a.SessionID, input, 16)
	packBytes(a.Challenge[:], input, 20)
	packBytes(password, input, 20+len(a.Challenge))

	var authCode []byte
	switch authType {
	case AuthTypePassword:
		authCode = password
	case AuthTypeMD2:
		authCode = md2.New().Sum(input)
		authCode = authCode[:16]
	case AuthTypeMD5:
		c := md5.Sum(input) // can not use md5.New().Sum(input)
		authCode = c[:]
	}

	return authCode[:16]
}

// 22.17.1 AuthCode Algorithms
// Multi-Session AuthCode carried in session header for all authenticated packets
type AuthCodeMultiSessionInput struct {
	Password   string
	SessionID  uint32
	SessionSeq uint32
	IPMIData   []byte
}

func (i *AuthCodeMultiSessionInput) AuthCode(authType AuthType) []byte {
	password := padBytes(i.Password, 16, 0x00)
	ipmiData := i.IPMIData

	// The Integrity Algorithm Number specifies the algorithm used to generate the contents
	// for the AuthCode signature field that accompanies authenticated IPMI v2.0/RMCP+ messages once the session has been
	// established.
	// Unless otherwise specified, the integrity algorithm is applied to the packet data starting with the
	// AuthType/Format field up to and including the field that immediately precedes the AuthCode field itself.
	authCodeInputLength := len(password) +
		4 + // session od uint32
		len(ipmiData) +
		4 + // session seq uint32
		len(password)

	var input = make([]byte, authCodeInputLength)
	packBytes(password, input, 0)
	packUint32L(i.SessionID, input, 16)
	packBytes(ipmiData, input, 20)
	packUint32L(i.SessionSeq, input, 20+len(ipmiData))
	packBytes(password, input, 20+len(ipmiData)+4)

	// c := md5.Sum(input)
	// authCode := c[:]

	var authCode []byte
	switch authType {
	case AuthTypePassword:
		authCode = password
	case AuthTypeMD2:
		authCode = md2.New().Sum(input)
		authCode = authCode[:16]
	case AuthTypeMD5:
		c := md5.Sum(input) // can not use md5.New().Sum(input)
		authCode = c[:]
	}

	return authCode[:16]
}

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
// see 22.17.1 AuthCode Algorithms
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
	switch c.session.v20.integrityAlg {
	case IntegrityAlg_None:
		//  If the Integrity Algorithm is none the AuthCode value is not calculated and
		// the AuthCode field in the message is not present (zero bytes).
		return []byte{}, nil

	case IntegrityAlg_MD5_128:
		data := []byte{}
		data = append(data, []byte(c.Password)[:]...)
		data = append(data, input...)
		data = append(data, []byte(c.Password)[:]...)
		h := md5.Sum(data)
		return h[:], nil

	case IntegrityAlg_HMAC_MD5_128:
		b, err := generate_hmac("md5", input, c.session.v20.k1)
		if err != nil {
			return nil, fmt.Errorf("generate hmac failed")
		}
		return b[0:16], nil

	case IntegrityAlg_HMAC_SHA1_96:

		b, err := generate_hmac("sha1", input, c.session.v20.k1)
		if err != nil {
			return nil, fmt.Errorf("generate hmac failed")
		}
		return b[0:12], nil

	case IntegrityAlg_HMAC_SHA256_128:
		b, err := generate_hmac("sha256", input, c.session.v20.k1)
		if err != nil {
			return nil, fmt.Errorf("generate hmac failed")
		}
		return b[0:16], nil

	default:
		return nil, fmt.Errorf("not support for integrity algorithm %x", c.session.v20.integrityAlg)
	}
}

// sik (Session Integrite Key)
// Both the remote console and the managed system generate sik by using
// the same hmackey and hmac data, so they should be same.
// see 13.31
func (c *Client) generate_sik() ([]byte, error) {
	input := make([]byte, 34+len(c.Username))
	packBytes(c.session.v20.consoleRand[:], input, 0) // 16 bytes
	packBytes(c.session.v20.bmcRand[:], input, 16)    // 16 bytes
	packUint8(c.session.v20.role, input, 32)          // 1 bytes, Requested privilege level (entire byte)
	packUint8(uint8(len(c.Username)), input, 33)      // 1 bytes, Username length
	packBytes([]byte(c.Username), input, 34)          // N bytes, Usename (absent for null usernames)

	c.DebugBytes("sik mac input", input, 16)
	var hmacKey []byte
	// hmacKey shoud use 160-bit key Kg
	// and Kuid is used in place of Kg if "one-key" logins are being used.
	if len(c.session.v20.bmcKey) != 0 {
		hmacKey = c.session.v20.bmcKey
	} else {
		hmacKey = padBytes(c.Password, 20, 0x00) // 160 bit = 20 bytes
	}
	c.DebugBytes("sik mac key", hmacKey, 16)

	b, err := generate_auth_hmac(c.session.v20.authAlg, input, hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}

	c.DebugBytes("sik mac computed by the remote console:", b, 16)

	return b, nil
}

// see 13.32 Generating Additional Keying Material
//
// generate K1 key, the session integrity key (SIK) is used as hmac key.
func (c *Client) generate_k1() ([]byte, error) {
	var CONST_1 = [20]byte{
		0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01,
	}

	if c.session.v20.sik == nil {
		return nil, fmt.Errorf("sik not exists, generate sik first")
	}
	hmacKey := c.session.v20.sik
	b, err := generate_auth_hmac(c.session.v20.authAlg, CONST_1[:], hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}

	c.DebugBytes("generated k1:", b, 16)

	return b, nil
}

// see 13.32 Generating Additional Keying Material
//
// generate K2 key, the session integrity key (SIK) is used as hmac key.
func (c *Client) generate_k2() ([]byte, error) {
	var CONST_2 = [20]byte{
		0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02,
	}

	if c.session.v20.sik == nil {
		return nil, fmt.Errorf("sik not exists, generate sik first")
	}
	hmacKey := c.session.v20.sik
	b, err := generate_auth_hmac(c.session.v20.authAlg, CONST_2[:], hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}
	c.DebugBytes("generated k2:", b, 16)

	return b, nil
}

// used for verify rakp2
func (c *Client) generate_rakp2_authcode() ([]byte, error) {
	c.DebugBytes("bmc rand", c.session.v20.bmcRand[:], 16)

	bufferLen := 4 + 4 + 16 + 16 + 16 + 1 + 1 + len(c.Username)
	var buffer = make([]byte, bufferLen)
	packUint32L(c.session.v20.consoleSessionID, buffer, 0) // 4 bytes, Console session ID (SID)
	packUint32L(c.session.v20.bmcSessionID, buffer, 4)     // 4 bytes, bmc session ID (SID)
	packBytes(c.session.v20.consoleRand[:], buffer, 8)     // 16 bytes, Remote console random number
	packBytes(c.session.v20.bmcRand[:], buffer, 24)        // 16 bytes, BMC random number (RC)
	packBytes(c.session.v20.bmcGUID[:], buffer, 40)        // 16 bytes, BMC guid
	packUint8(c.session.v20.role, buffer, 56)              // 1 bytes, entire byte of privilegelevel of rakp1
	packUint8(uint8(len(c.Username)), buffer, 57)          // 1 bytes, Username length
	packBytes([]byte(c.Username), buffer, 58)              // N bytes, Usename (absent for null usernames)
	c.DebugBytes("rakp2 authcode input", buffer, 16)

	// The bmc also use user password to caculate authcode, so if the authcode does not match,
	// it may indicates the password is no right.
	hmacKey := padBytes(c.Password, 20, 0x00)
	c.DebugBytes("rakp2 authcode key", hmacKey, 16)

	b, err := generate_auth_hmac(c.session.v20.authAlg, buffer, hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}

	c.DebugBytes("rakp2 generated authcode", b, 16)

	var out = b

	switch c.session.v20.authAlg {
	case AuthAlgRAKP_None:
		// nothing need to do
	case AuthAlgRAKP_HMAC_MD5:
		// need to copy 16 bytes
		if len(b) < 16 {
			err = fmt.Errorf("hmac md5 length should be at least 16 bytes")
		}
		out = b[0:16]
	case AuthAlgRAKP_HMAC_SHA1:
		// need to copy 20 bytes
		if len(b) < 20 {
			err = fmt.Errorf("hmac sha1 length should be at least 20 bytes")
		}
		out = b[0:20]
	case AuthAlgRAKP_HMAC_SHA256:
		if len(b) < 32 {
			err = fmt.Errorf("hmac sha256 length should be at least 32 bytes")
		}
		out = b[0:32]
	default:
		err = fmt.Errorf("rakp2 message: no support for authentication algorithm 0x%x", c.session.v20.authAlg)
	}

	c.DebugBytes("rakp2 used authcode", out, 16)
	return out, err
}

// 22.17.1 AuthCode Algorithms
func (c *Client) generate_rakp3_authcode() ([]byte, error) {

	// The auth code is an HMAC generated with the following content
	var input []byte = []byte{}
	input = append(input, c.session.v20.bmcRand[:]...) // 16 bytes, BMC random number (RC)

	buffer := make([]byte, 4)
	packUint32L(c.session.v20.consoleSessionID, buffer, 0)
	input = append(input, buffer...) // 4 bytes, Console session ID (SID)

	input = append(input, byte(c.session.v20.role)) // 1 bytes, Requested privilege level (entire byte)

	input = append(input, byte(len([]byte(c.Username)))) // 1 bytes, Username length

	input = append(input, []byte(c.Username)...) // N bytes, Usename (absent for null usernames)

	c.DebugBytes("rakp3 auth code input", input, 16)

	hmacKey := padBytes(c.Password, 20, 0x00)

	c.DebugBytes("rakp3 auth code key", hmacKey, 16)

	b, err := generate_auth_hmac(c.session.v20.authAlg, input, hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}

	c.DebugBytes("rakp3 generated authcode", b, 16)
	var out = b
	c.DebugBytes("rakp3 used authcode", out, 16)

	return out, err
}

// 13.31 RMCP+ Authenticated Key-Exchange Protocol (RAKP)
// 13.28
// the client use this method to verify the authcode returned in rakp4
func (c *Client) generate_rakp4_authcode() ([]byte, error) {
	var input []byte = []byte{}
	input = append(input, c.session.v20.consoleRand[:]...) // 16 bytes, Console random number

	buffer := make([]byte, 4)
	packUint32L(c.session.v20.bmcSessionID, buffer, 0)
	input = append(input, buffer...) // 4 bytes, BMC session ID (SID)

	input = append(input, c.session.v20.bmcGUID[:]...) // 16 bytes

	c.DebugBytes("rakp4 auth code input", input, 16)

	hmacKey := c.session.v20.sik
	c.DebugBytes("rakp4 auth code key", hmacKey, 16)

	b, err := generate_auth_hmac(AuthAlg(c.session.v20.integrityAlg), input, hmacKey)
	if err != nil {
		return nil, fmt.Errorf("generate hmac failed, err: %s", err)
	}

	c.DebugBytes("rakp4 generated authcode", b, 16)

	var errHmacLen = func(length int, integrityAlg IntegrityAlg) error {
		return fmt.Errorf("the length of generated mac is not long enough, should be at least (%d) for integrity algorithm (%0x)", len(b), integrityAlg)
	}

	var out = b

	integrityAlg := c.session.v20.integrityAlg
	switch integrityAlg {
	case IntegrityAlg_None:
		// nothing need to do
	case IntegrityAlg_HMAC_MD5_128:
		// need to copy 16 bytes
		if len(b) < 16 {
			err = errHmacLen(len(b), integrityAlg)
		}
		out = b[0:16]
	case IntegrityAlg_HMAC_SHA1_96:
		// need to copy 12 bytes
		if len(b) < 12 {
			err = errHmacLen(len(b), integrityAlg)
		}
		out = b[0:12]
	case IntegrityAlg_HMAC_SHA256_128:
		if len(b) < 16 {
			err = errHmacLen(len(b), integrityAlg)
		}
		out = b[0:16]
	default:
		err = fmt.Errorf("rakp4 message: no support for integrity algorithm %x", c.session.v20.integrityAlg)
	}
	c.DebugBytes("rakp4 used authcode", out, 16)

	return out, err
}
