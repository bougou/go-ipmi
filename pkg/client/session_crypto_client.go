package client

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/bougou/go-ipmi/pkg/types"
)

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
		c.session.v15.inSeq += 1
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

func genSessionTrailerPadLength(sessionHeader []byte, sessionPayload []byte) int {

	// (12) sessionHeader length
	// sessionPayload length
	// (1) pad length field
	// (1) next header field
	length := len(sessionHeader) + len(sessionPayload) + 1 + 1

	var padSize int = 0
	if length%4 != 0 {
		padSize = 4 - int(length%4)
	}
	return padSize
}

// genSessionTrailer will create the SessionTrailer.
//
// see 13.28.4 Integrity Algorithms
// Unless otherwise specified, the integrity algorithm is applied to the packet
// data starting with the AuthType/Format field up to and including the field
// that immediately precedes the AuthCode field itself.
func (c *Client) genSessionTrailer(sessionHeader []byte, sessionPayload []byte) (*types.SessionTrailer, error) {
	padSize := genSessionTrailerPadLength(sessionHeader, sessionPayload)
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

		encryptedPayload, err := encryptAES(paddedData, cipherKey, iv)
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

		encryptedPayload, err := encryptRC4(rawPayload, cipherKey, iv)
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
		d, err := decryptAES(cipherText, cipherKey, iv)
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
		b, err := decryptRC4(payloadData, cipherKey, iv)
		if err != nil {
			return nil, fmt.Errorf("decrypt payload with xRC4_128 failed, err: %w", err)
		}
		return b, nil

	default:
		return nil, fmt.Errorf("not supported encryption algorithm %0x", c.session.v20.cryptAlg)
	}
}
