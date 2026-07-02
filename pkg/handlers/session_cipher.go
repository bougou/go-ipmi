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

// allowedAlgorithms returns the sets of auth / integrity / confidentiality
// algorithm codes the server will accept in an RMCP+ Open Session Request,
// derived strictly from the configured cipher suites (spec §22.15.2,
// §13.17). An algorithm is accepted if and only if at least one configured
// cipher suite uses it.
//
// None is NOT implicitly accepted. The default cipher suite set
// ([bmc.DefaultCipherSuites], suites 3 and 17) does not contain any
// unauthenticated or unencrypted suite, so AuthAlgNone / IntegrityAlgNone /
// CryptAlgNone are rejected unless an operator explicitly configures a suite
// that uses them (e.g. suite 0 for fully unauthenticated, suite 1/15 for
// authenticated-but-unencrypted). This keeps Open Session negotiation
// consistent with the suites advertised via Get Channel Cipher Suites.
func allowedAlgorithms(b *bmc.BMC) (auth map[bmc.AuthAlg]bool, integ map[bmc.IntegrityAlg]bool, crypt map[bmc.CryptAlg]bool) {
	auth = make(map[bmc.AuthAlg]bool)
	integ = make(map[bmc.IntegrityAlg]bool)
	crypt = make(map[bmc.CryptAlg]bool)
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
