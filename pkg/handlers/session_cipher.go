package handlers

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
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
		auth, integ, crypt, ok := types.GetCipherSuiteAlgorithms(id)
		if !ok {
			continue
		}
		out = append(out, 0xC0, byte(id))
		out = append(out, 0x00|byte(auth))
		if integ != types.IntegrityAlg_None {
			out = append(out, 0x40|byte(integ))
		}
		if crypt != types.CryptAlg_None {
			out = append(out, 0x80|byte(crypt))
		}
	}
	return out
}

// isCipherSuiteAllowed checks whether the (auth, integ, crypt) triple from an
// Open Session Request matches at least one configured cipher suite (spec
// §22.15.2, §13.17). It validates the triple as a whole — each algorithm must
// come from the same suite. Cross-suite recombinations (where each algorithm
// appears in some configured suite but the triple as a unit was never
// advertised via Get Channel Cipher Suites) are rejected.
//
// When the triple is rejected, the error code attributes the failure to the
// first algorithm that does not appear in any configured suite. If all three
// algorithms exist individually but the triple is not a recognised suite
// combination, 0x04 (invalid authentication algorithm) is returned.
func isCipherSuiteAllowed(b *bmc.BMC, auth types.AuthAlg, integ types.IntegrityAlg, crypt types.CryptAlg) (ok bool, errCode uint8) {
	authKnown := false
	integKnown := false
	cryptKnown := false
	for _, id := range b.ResolvedCipherSuites() {
		a, i, c, ok := types.GetCipherSuiteAlgorithms(id)
		if !ok {
			continue
		}
		if !authKnown {
			authKnown = a == auth
		}
		if !integKnown {
			integKnown = i == integ
		}
		if !cryptKnown {
			cryptKnown = c == crypt
		}
		if a == auth && i == integ && c == crypt {
			return true, 0
		}
	}
	if !authKnown {
		return false, 0x04 // Invalid authentication algorithm
	}
	if !integKnown {
		return false, 0x05 // Invalid integrity algorithm
	}
	if !cryptKnown {
		return false, 0x10 // Invalid confidentiality algorithm
	}
	// All three algorithms appear individually in some configured suite, but
	// no single suite contains this triple — a cross-suite recombination.
	return false, 0x04
}
