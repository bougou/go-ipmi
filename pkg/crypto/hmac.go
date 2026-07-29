package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// HMAC computes a hash-based MAC. Supported algorithms: "md5", "sha1", "sha256".
func HMAC(alg string, data, key []byte) ([]byte, error) {
	switch alg {
	case "md5":
		h := hmac.New(md5.New, key)
		if _, err := h.Write(data); err != nil {
			return nil, fmt.Errorf("hmac md5 failed, err: %w", err)
		}
		return h.Sum(nil), nil
	case "sha1":
		h := hmac.New(sha1.New, key)
		if _, err := h.Write(data); err != nil {
			return nil, fmt.Errorf("hmac sha1 failed, err: %w", err)
		}
		return h.Sum(nil), nil
	case "sha256":
		h := hmac.New(sha256.New, key)
		if _, err := h.Write(data); err != nil {
			return nil, fmt.Errorf("hmac sha256 failed, err: %w", err)
		}
		return h.Sum(nil), nil
	default:
		return nil, fmt.Errorf("not support for hmac algorithm %s", alg)
	}
}

// AuthHMAC selects the HMAC variant based on the RAKP authentication algorithm
// (v2.0§13.28). Returns the full digest (RAKP2/RAKP3 and SIK/K1/K2 use the
// full digest; RAKP4 truncates separately).
func AuthHMAC(alg types.AuthAlg, data, key []byte) ([]byte, error) {
	switch alg {
	case types.AuthAlg_None:
		return nil, nil
	case types.AuthAlg_HMAC_MD5:
		return HMAC("md5", data, key)
	case types.AuthAlg_HMAC_SHA1:
		return HMAC("sha1", data, key)
	case types.AuthAlg_HMAC_SHA256:
		return HMAC("sha256", data, key)
	default:
		return nil, fmt.Errorf("unsupported auth algorithm: %d", alg)
	}
}

// IntegrityHMAC selects the HMAC variant based on the session integrity
// algorithm (v2.0§13.28.4). Returns the full digest; callers truncate.
func IntegrityHMAC(alg types.IntegrityAlg, data, key []byte) ([]byte, error) {
	switch alg {
	case types.IntegrityAlg_None:
		return nil, nil
	case types.IntegrityAlg_HMAC_MD5_128:
		return HMAC("md5", data, key)
	case types.IntegrityAlg_HMAC_SHA1_96:
		return HMAC("sha1", data, key)
	case types.IntegrityAlg_HMAC_SHA256_128:
		return HMAC("sha256", data, key)
	default:
		return nil, fmt.Errorf("unsupported integrity algorithm: %d", alg)
	}
}

// Equal compares two HMACs in constant time.
func Equal(a, b []byte) bool {
	return hmac.Equal(a, b)
}
