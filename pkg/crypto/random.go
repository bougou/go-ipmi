package crypto

import "crypto/rand"

// RandomBytes returns n cryptographically random bytes.
// Panics if crypto/rand is unavailable (same contract as prior client/server helpers).
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return b
}
