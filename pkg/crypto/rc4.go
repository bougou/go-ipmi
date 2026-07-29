package crypto

import (
	"crypto/rc4"
	"fmt"
)

// EncryptRC4 / DecryptRC4 apply the RC4 stream cipher (xRC4 confidentiality,
// v2.0§13.30). IV is unused by the cipher itself but kept in the signature for
// call-site parity with the AES helpers.
func EncryptRC4(plainText, cipherKey, _ []byte) ([]byte, error) {
	c, err := rc4.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}
	out := make([]byte, len(plainText))
	c.XORKeyStream(out, plainText)
	return out, nil
}

// DecryptRC4 is identical to EncryptRC4 for RC4 (XOR stream).
func DecryptRC4(cipherText, cipherKey, iv []byte) ([]byte, error) {
	return EncryptRC4(cipherText, cipherKey, iv)
}
