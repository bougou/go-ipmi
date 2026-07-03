package handlers

import (
	"crypto/md5"
	"crypto/subtle"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/utils/md2"
)

// genV15AuthCode computes the IPMI v1.5 multi-session AuthCode per spec
// §18.15.1 Figure 18-1. Must match client [client.AuthCodeMultiSessionInput].
func GenV15AuthCode(password []byte, authType bmc.V15AuthType, sessionID uint32, ipmiData []byte, sessionSeq uint32) []byte {
	padded := padV15Password(password)
	inputLen := 16 + 4 + len(ipmiData) + 4 + 16
	input := make([]byte, inputLen)
	copy(input[0:16], padded)
	packUint32LE(sessionID, input, 16)
	copy(input[20:], ipmiData)
	packUint32LE(sessionSeq, input, 20+len(ipmiData))
	copy(input[20+len(ipmiData)+4:], padded)

	var authCode []byte
	switch authType {
	case bmc.V15AuthTypePassword:
		authCode = padded
	case bmc.V15AuthTypeMD2:
		h := md2.New()
		h.Write(input)
		authCode = h.Sum(nil)
		authCode = authCode[:16]
	case bmc.V15AuthTypeMD5:
		h := md5.Sum(input)
		authCode = h[:]
	default:
		return nil
	}
	return authCode[:16]
}

// VerifyV15AuthCode returns true when got matches the expected AuthCode.
func VerifyV15AuthCode(password []byte, authType bmc.V15AuthType, sessionID uint32, ipmiData []byte, sessionSeq uint32, got []byte) bool {
	if len(got) != 16 {
		return false
	}
	expected := GenV15AuthCode(password, authType, sessionID, ipmiData, sessionSeq)
	if expected == nil {
		return false
	}
	return subtle.ConstantTimeCompare(expected, got) == 1
}

func padV15Password(password []byte) []byte {
	var p [16]byte
	copy(p[:], password)
	return p[:]
}

func packUint32LE(v uint32, dst []byte, off int) {
	dst[off] = byte(v)
	dst[off+1] = byte(v >> 8)
	dst[off+2] = byte(v >> 16)
	dst[off+3] = byte(v >> 24)
}
