package bmc

import (
	"github.com/bougou/go-ipmi/pkg/types"
)

// DefaultCipherSuites is the cipher suite set advertised when no explicit
// configuration is provided. It contains the spec-mandatory suite 3 plus the
// recommended SHA256 suite 17.
var DefaultCipherSuites = []types.CipherSuiteID{types.CipherSuiteID3, types.CipherSuiteID17}

// serverImplementedAlgorithms reports whether the reference server can perform
// the given algorithm triple end-to-end (advertise, negotiate, compute).
func serverImplementedAlgorithms(auth types.AuthAlg, integ types.IntegrityAlg, crypt types.CryptAlg) bool {
	switch auth {
	case types.AuthAlg_None, types.AuthAlg_HMAC_SHA1, types.AuthAlg_HMAC_SHA256:
	default:
		return false
	}
	switch integ {
	case types.IntegrityAlg_None, types.IntegrityAlg_HMAC_SHA1_96, types.IntegrityAlg_HMAC_SHA256_128:
	default:
		return false
	}
	switch crypt {
	case types.CryptAlg_None, types.CryptAlg_AES_CBC_128:
	default:
		return false
	}
	return true
}

// SupportedCipherSuite reports whether the reference server implements every
// algorithm in the named cipher suite. Configuring an unsupported suite would
// cause a runtime handshake failure, so callers validate with this before
// installing a cipher suite list.
func SupportedCipherSuite(id types.CipherSuiteID) bool {
	auth, integ, crypt, ok := types.GetCipherSuiteAlgorithms(id)
	if !ok {
		return false
	}
	return serverImplementedAlgorithms(auth, integ, crypt)
}
