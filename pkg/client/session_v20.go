package client

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

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

func (c *Client) genSession20(payloadType types.PayloadType, rawPayload []byte) (*types.Session20, error) {
	c.lock()
	defer c.unlock()

	//
	// Session Header
	//
	sessionHeader := &types.SessionHeader20{
		AuthType:             types.AuthTypeRMCPPlus, // Auth Type / Format is always 0x06 for IPMI v2
		PayloadType:          payloadType,
		PayloadAuthenticated: false,
		PayloadEncrypted:     false,
		SessionID:            0,
		Sequence:             0,
		PayloadLength:        0, // PayloadLength would be updated later after encryption if necessary.
	}

	if c.session.v20.state == types.SessionStateActive {
		sessionHeader.PayloadAuthenticated = true
		sessionHeader.PayloadEncrypted = true
		sessionHeader.SessionID = c.session.v20.bmcSessionID // use bmc session id

		c.session.v20.sequence += 1
		sessionHeader.Sequence = c.session.v20.sequence
	}

	//
	// Session Payload
	//
	sessionPayload := rawPayload
	if c.session.v20.state == types.SessionStateActive && sessionHeader.PayloadEncrypted {
		e, err := c.encryptPayload(rawPayload, nil)
		if err != nil {
			return nil, fmt.Errorf("encrypt payload failed, err: %w", err)
		}
		sessionPayload = e
	}
	// now we can fill PayloadLength field of the SessionHeader
	sessionHeader.PayloadLength = uint16(len(sessionPayload))
	c.DebugBytes("sessionPayload(final)", sessionPayload, 16)

	sessionHeaderBytes := sessionHeader.Pack()

	c.DebugBytes("sessionHeader", sessionHeaderBytes, 16)
	//
	// Session Trailer
	//
	var sessionTrailer *types.SessionTrailer = nil
	var err error
	// For IPMI v2.0 RMCP+ packets, the IPMI Session Trailer is absent
	// whenever the Session ID is 0000_0000h, or the packet is unauthenticated
	if sessionHeader.PayloadAuthenticated && sessionHeader.SessionID != 0 {
		sessionTrailer, err = c.genSessionTrailer(sessionHeaderBytes, sessionPayload)
		if err != nil {
			return nil, fmt.Errorf("genSessionTrailer failed, err: %w", err)
		}
	}

	return &types.Session20{
		SessionHeader20: sessionHeader,
		SessionPayload:  sessionPayload,
		SessionTrailer:  sessionTrailer,
	}, nil
}

// genSessionTrailer will create the SessionTrailer.
//
// see 13.28.4 Integrity Algorithms
// Unless otherwise specified, the integrity algorithm is applied to the packet
// data starting with the AuthType/Format field up to and including the field
// that immediately precedes the AuthCode field itself.
func (c *Client) genSessionTrailer(sessionHeader []byte, sessionPayload []byte) (*types.SessionTrailer, error) {
	padSize := types.IntegrityPadLen(len(sessionHeader), len(sessionPayload))
	var pad = make([]byte, padSize)
	for i := 0; i < padSize; i++ {
		pad[i] = 0xff
	}

	sessionTrailer := &types.SessionTrailer{
		IntegrityPAD: pad,
		PadLength:    uint8(padSize),
		NextHeader:   0x07, /* Hardcoded per the spec, table 13-8 */
		AuthCode:     nil,
	}

	var input []byte = sessionHeader
	input = append(input, sessionPayload...)
	input = append(input, sessionTrailer.IntegrityPAD...)
	input = append(input, sessionTrailer.PadLength)
	input = append(input, sessionTrailer.NextHeader)

	c.DebugBytes("auth code input", input, 16)

	authCode, err := c.genIntegrityAuthCode(input)
	if err != nil {
		return nil, fmt.Errorf("generate integrity authcode failed, err: %w", err)
	}

	c.DebugBytes("generated auth code", authCode, 16)

	sessionTrailer.AuthCode = authCode

	return sessionTrailer, nil
}

// the input data only represents the serialized ipmi msg request bytes.
// the output bytes contains the
//   - Confidentiality Header (clear text)
//   - Encrypted Payload.
//   - the cipher text of both rawPayload
//   - padded Confidentiality Trailer.
func (c *Client) encryptPayload(rawPayload []byte, iv []byte) ([]byte, error) {

	switch c.session.v20.cryptAlg {
	case types.CryptAlg_None:
		return rawPayload, nil

	case types.CryptAlg_AES_CBC_128:
		// The input to the AES encryption algorithm has to be a multiple of the block size (16 bytes).
		// The extra byte we are adding is the pad length byte.
		var paddedData = rawPayload
		var padLength uint8
		if mod := (len(rawPayload) + 1) % int(types.Encryption_AES_CBS_128_BlockSize); mod > 0 {
			padLength = types.Encryption_AES_CBS_128_BlockSize - uint8(mod)
		} else {
			padLength = 0
		}
		for i := uint8(0); i < padLength; i++ {
			paddedData = append(paddedData, i+1)
		}
		paddedData = append(paddedData, padLength) // now, the length of data SHOULD be multiple of 16
		c.DebugBytes("padded data (before encrypt)", paddedData, 16)

		// see 13.29 AES-CBC Encrypted Payload Fields
		if len(iv) == 0 {
			iv = randomBytes(16) // Initialization Vector
		}
		c.DebugBytes("random iv", iv, 16)

		// see 13.29.2 Encryption with AES
		// AES-128 uses a 128-bit Cipher Key. The Cipher Key is the first 128-bits of key K2
		cipherKey := c.session.v20.k2[0:16]
		c.DebugBytes("cipher key (k2)", cipherKey, 16)

		encryptedPayload, err := crypto.EncryptAES(paddedData, cipherKey, iv)
		if err != nil {
			return nil, fmt.Errorf("encrypt payload with AES_CBC_128 failed, err: %w", err)
		}
		c.DebugBytes("encrypted data", encryptedPayload, 16)

		var out []byte

		// write Confidentiality Header
		out = append(out, iv...)
		// write Encrypted Payload
		out = append(out, encryptedPayload...)

		c.DebugBytes("encrypted session payload", out, 16)

		return out, nil

	case types.CryptAlg_xRC4_40, types.CryptAlg_xRC4_128:
		var out []byte

		// see 13.30 xRC4-Encrypted Payload Fields
		var confidentialityHeader []byte
		var offset = make([]byte, 4)
		if c.session.v20.accumulatedPayloadSize == 0 {
			// means this is the first sent packet
			for i := 0; i < 4; i++ {
				offset[i] = 0
			}
			c.session.v20.rc4EncryptIV = array16(randomBytes(16))
			confidentialityHeader = append(offset, c.session.v20.rc4EncryptIV[:]...)
		} else {
			binary.BigEndian.PutUint32(offset, c.session.v20.accumulatedPayloadSize)
			confidentialityHeader = offset
		}

		c.session.v20.accumulatedPayloadSize += uint32(len(rawPayload))

		iv := c.session.v20.rc4EncryptIV[:]
		out = append(out, confidentialityHeader...)

		input := append(c.session.v20.k2, iv...)
		keyRC := md5.Sum(input)

		var cipherKey []byte
		switch c.session.v20.cryptAlg {
		case types.CryptAlg_xRC4_40:
			// For xRC4 using a 40-bit key, only the most significant forty bits of Krc are used
			cipherKey = keyRC[:5]

		case types.CryptAlg_xRC4_128:
			// For xRC4 using a 128-bit key, all bits of Krc are used for initialization
			cipherKey = keyRC[:16]
		}

		encryptedPayload, err := crypto.EncryptRC4(rawPayload, cipherKey, iv)
		if err != nil {
			return nil, fmt.Errorf("encrypt payload with xRC4_40 or xRC4_128 failed, err: %w", err)
		}
		// write Encrypted Payload
		out = append(out, encryptedPayload...)
		// xRC4 does not use a confidentiality trailer.
		return out, nil

	default:

		return nil, fmt.Errorf("not supported encryption algorithm %x", c.session.v20.cryptAlg)
	}
}

// the input data is the encrypted session payload.
// the output bytes is the decrypted IPMI Message bytes with padding removed.
func (c *Client) decryptPayload(data []byte) ([]byte, error) {
	switch c.session.v20.cryptAlg {

	case types.CryptAlg_None:
		return data, nil

	case types.CryptAlg_AES_CBC_128:
		iv := data[0:16] // the first 16 byte is the initialization vector
		cipherText := data[16:]
		cipherKey := c.session.v20.k2[0:16]
		d, err := crypto.DecryptAES(cipherText, cipherKey, iv)
		if err != nil {
			return nil, fmt.Errorf("decrypt payload with AES_CBC_128 failed, err: %w", err)
		}
		padLength := d[len(d)-1]
		dEnd := len(d) - int(padLength) - 1
		return d[0:dEnd], nil

	case types.CryptAlg_xRC4_40, types.CryptAlg_xRC4_128:
		// the first received packet
		if data[0] == 0x0 && data[1] == 0x0 && data[2] == 0x0 && data[3] == 0x0 {
			c.session.v20.rc4DecryptIV = array16(data[4:20])
		}

		iv := c.session.v20.rc4DecryptIV[:]
		input := append(c.session.v20.k2, iv...)
		keyRC := md5.Sum(input)
		var cipherKey []byte
		switch c.session.v20.cryptAlg {
		case types.CryptAlg_xRC4_40:
			// For xRC4 using a 40-bit key, only the most significant forty bits of Krc are used
			cipherKey = keyRC[:5]

		case types.CryptAlg_xRC4_128:
			// For xRC4 using a 128-bit key, all bits of Krc are used for initialization
			cipherKey = keyRC[:16]
		}

		payloadData := data[20:]
		b, err := crypto.DecryptRC4(payloadData, cipherKey, iv)
		if err != nil {
			return nil, fmt.Errorf("decrypt payload with xRC4_128 failed, err: %w", err)
		}
		return b, nil

	default:
		return nil, fmt.Errorf("not supported encryption algorithm %0x", c.session.v20.cryptAlg)
	}
}
