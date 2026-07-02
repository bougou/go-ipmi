package bmc

// CipherSuiteID identifies an IPMI 2.0 RMCP+ cipher suite (spec §22.15.2).
//
// The bmc package mirrors this type (rather than importing pkg/types) to keep
// its dependency surface minimal, consistent with the existing AuthAlg /
// IntegrityAlg / CryptAlg mirrors in session.go.
type CipherSuiteID uint8

const (
	CipherSuiteID0  CipherSuiteID = 0  // RAKP-None + Integrity-None + Confidentiality-None (unauthenticated, unencrypted)
	CipherSuiteID1  CipherSuiteID = 1  // RAKP-HMAC-SHA1 + Integrity-None + Confidentiality-None
	CipherSuiteID2  CipherSuiteID = 2  // RAKP-HMAC-SHA1 + HMAC-SHA1-96 + Confidentiality-None
	CipherSuiteID3  CipherSuiteID = 3  // RAKP-HMAC-SHA1 + HMAC-SHA1-96 + AES-CBC-128 (spec-mandatory)
	CipherSuiteID15 CipherSuiteID = 15 // RAKP-HMAC-SHA256 + Integrity-None + Confidentiality-None
	CipherSuiteID16 CipherSuiteID = 16 // RAKP-HMAC-SHA256 + HMAC-SHA256-128 + Confidentiality-None
	CipherSuiteID17 CipherSuiteID = 17 // RAKP-HMAC-SHA256 + HMAC-SHA256-128 + AES-CBC-128
)

// DefaultCipherSuites is the cipher suite set advertised when no explicit
// configuration is provided. It contains the spec-mandatory suite 3 plus the
// recommended SHA256 suite 17.
var DefaultCipherSuites = []CipherSuiteID{CipherSuiteID3, CipherSuiteID17}

// CipherSuiteAlgorithms expands a cipher suite ID into its auth / integrity /
// confidentiality algorithm codes (spec §22.15.2). ok is false for IDs whose
// algorithm triple is not known to this build.
//
// Suite 0 (AuthAlgNone) is intentionally NOT in DefaultCipherSuites: selecting
// it disables RAKP authentication entirely, so the session is established
// without any password verification. Operators who want this must add it
// explicitly via [WithCipherSuites] / [BMC.SetCipherSuites].
func CipherSuiteAlgorithms(id CipherSuiteID) (auth AuthAlg, integ IntegrityAlg, crypt CryptAlg, ok bool) {
	switch id {
	case CipherSuiteID0:
		return AuthAlgNone, IntegrityAlgNone, CryptAlgNone, true
	case CipherSuiteID1:
		return AuthAlgHMACSHA1, IntegrityAlgNone, CryptAlgNone, true
	case CipherSuiteID2:
		return AuthAlgHMACSHA1, IntegrityAlgHMACSHA1_96, CryptAlgNone, true
	case CipherSuiteID3:
		return AuthAlgHMACSHA1, IntegrityAlgHMACSHA1_96, CryptAlgAESCBC128, true
	case CipherSuiteID15:
		return AuthAlgHMACSHA256, IntegrityAlgNone, CryptAlgNone, true
	case CipherSuiteID16:
		return AuthAlgHMACSHA256, IntegrityAlgHMACSHA256_128, CryptAlgNone, true
	case CipherSuiteID17:
		return AuthAlgHMACSHA256, IntegrityAlgHMACSHA256_128, CryptAlgAESCBC128, true
	default:
		return 0, 0, 0, false
	}
}

// serverImplementedAlgorithms reports whether the reference server can perform
// the given algorithm triple end-to-end (advertise, negotiate, compute).
func serverImplementedAlgorithms(auth AuthAlg, integ IntegrityAlg, crypt CryptAlg) bool {
	switch auth {
	case AuthAlgNone, AuthAlgHMACSHA1, AuthAlgHMACSHA256:
	default:
		return false
	}
	switch integ {
	case IntegrityAlgNone, IntegrityAlgHMACSHA1_96, IntegrityAlgHMACSHA256_128:
	default:
		return false
	}
	switch crypt {
	case CryptAlgNone, CryptAlgAESCBC128:
	default:
		return false
	}
	return true
}

// SupportedCipherSuite reports whether the reference server implements every
// algorithm in the named cipher suite. Configuring an unsupported suite would
// cause a runtime handshake failure, so callers validate with this before
// installing a cipher suite list.
func SupportedCipherSuite(id CipherSuiteID) bool {
	auth, integ, crypt, ok := CipherSuiteAlgorithms(id)
	if !ok {
		return false
	}
	return serverImplementedAlgorithms(auth, integ, crypt)
}
