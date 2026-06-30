package handlers

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
)

// cipherSuiteRecords builds the wire bytes for the Get Channel Cipher Suites
// response (spec §22.15.1) from the BMC's configured cipher suite list.
//
// Each standard record is:
//
//	0xC0 <cipherSuiteID> 0x00|authAlg [0x40|integAlg] [0x80|cryptAlg]
//
// where the integrity and confidentiality entries are omitted when their
// algorithm is None. Auth is always present per spec.
func cipherSuiteRecords(b *bmc.BMC) []byte {
	ids := b.ResolvedCipherSuites()
	out := make([]byte, 0, len(ids)*5)
	for _, id := range ids {
		auth, integ, crypt, ok := bmc.CipherSuiteAlgorithms(id)
		if !ok {
			continue
		}
		out = append(out, 0xC0, byte(id))
		out = append(out, 0x00|byte(auth))
		if integ != bmc.IntegrityAlgNone {
			out = append(out, 0x40|byte(integ))
		}
		if crypt != bmc.CryptAlgNone {
			out = append(out, 0x80|byte(crypt))
		}
	}
	return out
}

// allowedAuthAlgorithms / allowedIntegrityAlgorithms / allowedCryptAlgorithms
// return the sets of algorithm codes the server will accept in an Open Session
// Request, derived from the configured cipher suites. An algorithm is accepted
// if at least one configured suite uses it (or None, which is always accepted
// for auth/integrity/crypt since the spec allows negotiating "no integrity" /
// "no confidentiality" within a suite that supports them).
func allowedAlgorithms(b *bmc.BMC) (auth map[bmc.AuthAlg]bool, integ map[bmc.IntegrityAlg]bool, crypt map[bmc.CryptAlg]bool) {
	auth = map[bmc.AuthAlg]bool{bmc.AuthAlgNone: true}
	integ = map[bmc.IntegrityAlg]bool{bmc.IntegrityAlgNone: true}
	crypt = map[bmc.CryptAlg]bool{bmc.CryptAlgNone: true}
	for _, id := range b.ResolvedCipherSuites() {
		a, i, c, ok := bmc.CipherSuiteAlgorithms(id)
		if !ok {
			continue
		}
		auth[a] = true
		integ[i] = true
		crypt[c] = true
	}
	return
}
