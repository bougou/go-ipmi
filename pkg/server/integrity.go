package server

import (
	"encoding/binary"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/crypto"
	"github.com/bougou/go-ipmi/pkg/types"
)

const (
	rmcpHeaderSize        = 4
	rmcpPlusHeaderSize    = 12
	rmcpPlusPayloadOffset = rmcpHeaderSize + rmcpPlusHeaderSize
	rmcpPlusNextHeader    = 0x07 // v2.0 Table 13-8
)

func appendRMCPPlusIntegrity(pkt []byte, sess *bmc.Session) ([]byte, bool) {
	authCodeLen, ok := crypto.IntegrityAuthCodeLen(sess.IntegrityAlg)
	if !ok {
		return nil, false
	}
	if authCodeLen == 0 {
		return pkt, true
	}
	if len(sess.K1) == 0 || len(pkt) < rmcpPlusPayloadOffset {
		return nil, false
	}

	padLen := types.IntegrityPadLen(len(pkt[rmcpHeaderSize:rmcpPlusPayloadOffset]), len(pkt[rmcpPlusPayloadOffset:]))
	out := make([]byte, 0, len(pkt)+padLen+2+authCodeLen)
	out = append(out, pkt...)
	for i := 0; i < padLen; i++ {
		out = append(out, 0xff)
	}
	out = append(out, byte(padLen), rmcpPlusNextHeader)

	authCode, err := crypto.SessionIntegrityAuthCode(sess.IntegrityAlg, out[rmcpHeaderSize:], sess.K1, "")
	if err != nil || len(authCode) != authCodeLen {
		return nil, false
	}
	out = append(out, authCode...)
	return out, true
}

func verifyRMCPPlusIntegrity(pkt []byte, sess *bmc.Session, authenticated bool) bool {
	authCodeLen, ok := crypto.IntegrityAuthCodeLen(sess.IntegrityAlg)
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

	padLen := types.IntegrityPadLen(rmcpPlusHeaderSize, payloadLen)
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

	expected, err := crypto.SessionIntegrityAuthCode(sess.IntegrityAlg, pkt[rmcpHeaderSize:authCodeStart], sess.K1, "")
	if err != nil {
		return false
	}
	return crypto.Equal(expected, pkt[authCodeStart:])
}
