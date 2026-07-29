package crypto

import (
	"crypto/md5"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// IntegrityAuthCodeLen returns the AuthCode (integrity data) field length for
// alg (v2.0§13.28.4). ok is false for unknown algorithms.
func IntegrityAuthCodeLen(alg types.IntegrityAlg) (int, bool) {
	switch alg {
	case types.IntegrityAlg_None:
		return 0, true
	case types.IntegrityAlg_HMAC_SHA1_96:
		return 12, true
	case types.IntegrityAlg_HMAC_MD5_128, types.IntegrityAlg_MD5_128, types.IntegrityAlg_HMAC_SHA256_128:
		return 16, true
	default:
		return 0, false
	}
}

// SessionIntegrityAuthCode computes the AuthCode for an authenticated RMCP+
// session packet (v2.0§13.28.4). Truncation matches the integrity algorithm:
//   - HMAC-SHA1-96 → 12 bytes
//   - HMAC-MD5-128 / HMAC-SHA256-128 → 16 bytes
//   - MD5-128 → MD5(password || input || password), 16 bytes
//
// password is only used for IntegrityAlg_MD5_128; HMAC variants use k1.
func SessionIntegrityAuthCode(alg types.IntegrityAlg, input, k1 []byte, password string) ([]byte, error) {
	switch alg {
	case types.IntegrityAlg_None:
		return []byte{}, nil
	case types.IntegrityAlg_MD5_128:
		data := make([]byte, 0, len(password)+len(input)+len(password))
		data = append(data, []byte(password)...)
		data = append(data, input...)
		data = append(data, []byte(password)...)
		h := md5.Sum(data)
		return h[:], nil
	case types.IntegrityAlg_HMAC_MD5_128:
		b, err := IntegrityHMAC(alg, input, k1)
		if err != nil {
			return nil, err
		}
		return b[0:16], nil
	case types.IntegrityAlg_HMAC_SHA1_96:
		b, err := IntegrityHMAC(alg, input, k1)
		if err != nil {
			return nil, err
		}
		return b[0:12], nil
	case types.IntegrityAlg_HMAC_SHA256_128:
		b, err := IntegrityHMAC(alg, input, k1)
		if err != nil {
			return nil, err
		}
		return b[0:16], nil
	default:
		return nil, fmt.Errorf("not support for integrity algorithm %x", alg)
	}
}
