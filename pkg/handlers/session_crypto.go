package handlers

// HMAC and key-derivation logic for IPMI 2.0 RAKP authentication.
// Implementations live in pkg/crypto; this file adapts BMC session state.

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/crypto"
)

// computeRAKP2AuthCode generates the Key Exchange Authentication Code that the
// BMC sends in RAKP Message 2 (v2.0§13.31).
func computeRAKP2AuthCode(sess *bmc.Session, b *bmc.BMC) ([]byte, error) {
	return crypto.RAKP2AuthCode(
		sess.AuthAlg,
		sess.ConsoleID,
		sess.BMCID,
		sess.ConsoleRand[:],
		sess.BMCRand[:],
		b.GUID[:],
		sess.Role,
		sessionUsername(sess),
		paddedPassword(sess),
	)
}

// computeRAKP3AuthCode generates the auth code the BMC expects in RAKP Message 3.
func computeRAKP3AuthCode(sess *bmc.Session, _ *bmc.BMC) ([]byte, error) {
	return crypto.RAKP3AuthCode(
		sess.AuthAlg,
		sess.BMCRand[:],
		sess.ConsoleID,
		sess.Role,
		sessionUsername(sess),
		paddedPassword(sess),
	)
}

// computeRAKP4AuthCode generates the Integrity Check Value the BMC sends in
// RAKP Message 4 (v2.0§13.28.1 / §13.31).
func computeRAKP4AuthCode(sess *bmc.Session, b *bmc.BMC) ([]byte, error) {
	return crypto.RAKP4ICV(sess.AuthAlg, sess.ConsoleRand[:], sess.BMCID, b.GUID[:], sess.SIK)
}

// deriveSessKeys computes SIK, K1, and K2 from the session parameters per spec §13.31-13.32.
func deriveSessKeys(sess *bmc.Session, b *bmc.BMC) error {
	var sikKey []byte
	if len(b.KG) > 0 {
		sikKey = b.KG
	} else {
		sikKey = paddedPassword(sess)
	}

	sik, k1, k2, err := crypto.DeriveSessionKeys(
		sess.AuthAlg,
		sess.ConsoleRand[:],
		sess.BMCRand[:],
		sess.Role,
		sessionUsername(sess),
		sikKey,
	)
	if err != nil {
		return err
	}
	sess.SIK = sik
	sess.K1 = k1
	sess.K2 = k2
	return nil
}

func sessionUsername(sess *bmc.Session) string {
	if sess.User == nil {
		return ""
	}
	return sess.User.Name
}

// paddedPassword returns the user password padded to 20 bytes (Kuid).
func paddedPassword(sess *bmc.Session) []byte {
	if sess.User == nil {
		return crypto.PadPassword20(nil)
	}
	return crypto.PadPassword20(sess.User.Password[:])
}

func hmacEqual(a, b []byte) bool {
	return crypto.Equal(a, b)
}
