package crypto

import (
	"crypto/md5"
	"crypto/subtle"

	"github.com/bougou/go-ipmi/pkg/types"
	"github.com/bougou/go-ipmi/utils/md2"
)

// SingleSessionInput is the AuthCode carried in Activate Session request data
// (v1.5§18.15.1 / v2.0§22.17.1).
type SingleSessionInput struct {
	Password  string
	SessionID uint32
	Challenge []byte
}

// AuthCode computes the single-session AuthCode for authType.
func (a SingleSessionInput) AuthCode(authType types.AuthType) []byte {
	return AuthCodeSingleSession(authType, []byte(a.Password), a.SessionID, a.Challenge)
}

// MultiSessionInput is the AuthCode carried in the v1.5 session header for
// authenticated packets (v1.5§18.15.1 / v2.0§22.17.1).
type MultiSessionInput struct {
	Password   string
	SessionID  uint32
	SessionSeq uint32
	IPMIData   []byte
}

// AuthCode computes the multi-session AuthCode for authType.
func (i MultiSessionInput) AuthCode(authType types.AuthType) []byte {
	return AuthCodeMultiSession(authType, []byte(i.Password), i.SessionID, i.SessionSeq, i.IPMIData)
}

// AuthCodeSingleSession computes the Activate Session AuthCode.
func AuthCodeSingleSession(authType types.AuthType, password []byte, sessionID uint32, challenge []byte) []byte {
	padded := padPassword16(password)
	input := make([]byte, 16+4+len(challenge)+16)
	copy(input[0:16], padded)
	packUint32LE(sessionID, input, 16)
	copy(input[20:], challenge)
	copy(input[20+len(challenge):], padded)
	return hashAuthCode(authType, padded, input)
}

// AuthCodeMultiSession computes the multi-session session-header AuthCode.
func AuthCodeMultiSession(authType types.AuthType, password []byte, sessionID, sessionSeq uint32, ipmiData []byte) []byte {
	padded := padPassword16(password)
	input := make([]byte, 16+4+len(ipmiData)+4+16)
	copy(input[0:16], padded)
	packUint32LE(sessionID, input, 16)
	copy(input[20:], ipmiData)
	packUint32LE(sessionSeq, input, 20+len(ipmiData))
	copy(input[20+len(ipmiData)+4:], padded)
	return hashAuthCode(authType, padded, input)
}

// VerifyMultiSessionAuthCode returns true when got matches the expected AuthCode.
func VerifyMultiSessionAuthCode(authType types.AuthType, password []byte, sessionID, sessionSeq uint32, ipmiData, got []byte) bool {
	if len(got) != 16 {
		return false
	}
	expected := AuthCodeMultiSession(authType, password, sessionID, sessionSeq, ipmiData)
	if expected == nil {
		return false
	}
	return subtle.ConstantTimeCompare(expected, got) == 1
}

func hashAuthCode(authType types.AuthType, paddedPassword, input []byte) []byte {
	var authCode []byte
	switch authType {
	case types.AuthTypePassword:
		authCode = paddedPassword
	case types.AuthTypeMD2:
		h := md2.New()
		h.Write(input)
		authCode = h.Sum(nil)
	case types.AuthTypeMD5:
		sum := md5.Sum(input)
		authCode = sum[:]
	default:
		return nil
	}
	return authCode[:16]
}

func padPassword16(password []byte) []byte {
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
