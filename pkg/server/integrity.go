package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	rmcpHeaderSize        = 4
	rmcpPlusHeaderSize    = 12
	rmcpPlusPayloadOffset = rmcpHeaderSize + rmcpPlusHeaderSize
	rmcpPlusNextHeader    = 0x07
)

func appendRMCPPlusIntegrity(pkt []byte, sess *bmc.Session) ([]byte, bool) {
	authCodeLen, ok := rmcpPlusIntegrityAuthCodeLen(sess.IntegrityAlg)
	if !ok {
		return nil, false
	}
	if authCodeLen == 0 {
		return pkt, true
	}
	if len(sess.K1) == 0 || len(pkt) < rmcpPlusPayloadOffset {
		return nil, false
	}

	padLen := rmcpPlusIntegrityPadLen(len(pkt[rmcpHeaderSize:rmcpPlusPayloadOffset]), len(pkt[rmcpPlusPayloadOffset:]))
	out := make([]byte, 0, len(pkt)+padLen+2+authCodeLen)
	out = append(out, pkt...)
	for i := 0; i < padLen; i++ {
		out = append(out, 0xff)
	}
	out = append(out, byte(padLen), rmcpPlusNextHeader)

	authCode := rmcpPlusIntegrityAuthCode(sess.IntegrityAlg, out[rmcpHeaderSize:], sess.K1)
	out = append(out, authCode...)
	return out, true
}

func verifyRMCPPlusIntegrity(pkt []byte, sess *bmc.Session, authenticated bool) bool {
	authCodeLen, ok := rmcpPlusIntegrityAuthCodeLen(sess.IntegrityAlg)
	if !ok {
		return false
	}
	if authCodeLen == 0 {
		return true
	}
	if !authenticated || len(sess.K1) == 0 || len(pkt) < rmcpPlusPayloadOffset {
		return false
	}

	payloadLen := int(binary.LittleEndian.Uint16(pkt[14:16]))
	payloadEnd := rmcpPlusPayloadOffset + payloadLen
	if len(pkt) < payloadEnd {
		return false
	}

	padLen := rmcpPlusIntegrityPadLen(rmcpPlusHeaderSize, payloadLen)
	authCodeStart := payloadEnd + padLen + 2
	if len(pkt) != authCodeStart+authCodeLen {
		return false
	}

	for _, b := range pkt[payloadEnd : payloadEnd+padLen] {
		if b != 0xff {
			return false
		}
	}
	if pkt[payloadEnd+padLen] != byte(padLen) || pkt[payloadEnd+padLen+1] != rmcpPlusNextHeader {
		return false
	}

	expected := rmcpPlusIntegrityAuthCode(sess.IntegrityAlg, pkt[rmcpHeaderSize:authCodeStart], sess.K1)
	return hmac.Equal(expected, pkt[authCodeStart:])
}

func rmcpPlusIntegrityPadLen(sessionHeaderLen, payloadLen int) int {
	n := sessionHeaderLen + payloadLen + 1 + 1
	if n%4 == 0 {
		return 0
	}
	return 4 - n%4
}

func rmcpPlusIntegrityAuthCodeLen(alg types.IntegrityAlg) (int, bool) {
	switch alg {
	case types.IntegrityAlg_None:
		return 0, true
	case types.IntegrityAlg_HMAC_SHA1_96:
		return 12, true
	case types.IntegrityAlg_HMAC_SHA256_128:
		return 16, true
	default:
		return 0, false
	}
}

func rmcpPlusIntegrityAuthCode(alg types.IntegrityAlg, data, key []byte) []byte {
	switch alg {
	case types.IntegrityAlg_HMAC_SHA1_96:
		h := hmac.New(sha1.New, key)
		h.Write(data)
		return h.Sum(nil)[:12]
	case types.IntegrityAlg_HMAC_SHA256_128:
		h := hmac.New(sha256.New, key)
		h.Write(data)
		return h.Sum(nil)[:16]
	default:
		return nil
	}
}
