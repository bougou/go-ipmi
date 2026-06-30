package handlers

// session_crypto.go contains the HMAC and key-derivation logic for IPMI 2.0
// RAKP authentication.  All inputs and outputs are raw bytes; callers own the
// returned slices.
//
// Algorithm references:
//   - IPMI 2.0 spec §13.31 – RAKP key exchange
//   - IPMI 2.0 spec §13.32 – Additional keying material (K1/K2)

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

// computeRAKP2AuthCode generates the Key Exchange Authentication Code that the
// BMC sends in RAKP Message 2.
//
// HMAC input (spec §13.31):
//
//	ConsoleID(4) || BMCID(4) || ConsoleRand(16) || BMCRand(16) || BMCGUID(16) || Role(1) || UserLen(1) || Username(N)
func computeRAKP2AuthCode(sess *bmc.Session, b *bmc.BMC) ([]byte, error) {
	if sess.AuthAlg == bmc.AuthAlgNone {
		return nil, nil
	}
	username := ""
	if sess.User != nil {
		username = sess.User.Name
	}

	buf := make([]byte, 4+4+16+16+16+1+1+len(username))
	binary.LittleEndian.PutUint32(buf[0:4], sess.ConsoleID)
	binary.LittleEndian.PutUint32(buf[4:8], sess.BMCID)
	copy(buf[8:24], sess.ConsoleRand[:])
	copy(buf[24:40], sess.BMCRand[:])
	copy(buf[40:56], b.GUID[:])
	buf[56] = sess.Role
	buf[57] = uint8(len(username))
	copy(buf[58:], username)

	key := hmacKey(sess, b)
	return computeHMAC(sess.AuthAlg, buf, key)
}

// computeRAKP3AuthCode generates the auth code the BMC expects in RAKP Message 3.
//
// HMAC input:
//
//	BMCRand(16) || ConsoleID(4) || Role(1) || UserLen(1) || Username(N)
func computeRAKP3AuthCode(sess *bmc.Session, b *bmc.BMC) ([]byte, error) {
	if sess.AuthAlg == bmc.AuthAlgNone {
		return nil, nil
	}
	username := ""
	if sess.User != nil {
		username = sess.User.Name
	}

	buf := make([]byte, 16+4+1+1+len(username))
	copy(buf[0:16], sess.BMCRand[:])
	binary.LittleEndian.PutUint32(buf[16:20], sess.ConsoleID)
	buf[20] = sess.Role
	buf[21] = uint8(len(username))
	copy(buf[22:], username)

	key := hmacKey(sess, b)
	return computeHMAC(sess.AuthAlg, buf, key)
}

// computeRAKP4AuthCode generates the confirmation code the BMC sends in RAKP4.
//
// HMAC input:
//
//	ConsoleRand(16) || BMCID(4) || BMCGUID(16)
//
// The HMAC key for RAKP4 is the Session Integrity Key (SIK), not Kuid/KG.
// The integrity algorithm (not auth algorithm) selects the HMAC variant.
func computeRAKP4AuthCode(sess *bmc.Session, b *bmc.BMC) ([]byte, error) {
	buf := make([]byte, 16+4+16)
	copy(buf[0:16], sess.ConsoleRand[:])
	binary.LittleEndian.PutUint32(buf[16:20], sess.BMCID)
	copy(buf[20:36], b.GUID[:])

	return computeHMACIntegrity(sess.IntegrityAlg, buf, sess.SIK)
}

// deriveSessKeys computes SIK, K1, and K2 from the session parameters per spec §13.31-13.32.
func deriveSessKeys(sess *bmc.Session, b *bmc.BMC) error {
	username := ""
	if sess.User != nil {
		username = sess.User.Name
	}

	// SIK = HMAC(Kg or Kuid, ConsoleRand || BMCRand || Role || UserLen || Username)
	sikInput := make([]byte, 16+16+1+1+len(username))
	copy(sikInput[0:16], sess.ConsoleRand[:])
	copy(sikInput[16:32], sess.BMCRand[:])
	sikInput[32] = sess.Role
	sikInput[33] = uint8(len(username))
	copy(sikInput[34:], username)

	var sikKey []byte
	if len(b.KG) > 0 {
		sikKey = b.KG
	} else {
		sikKey = paddedPassword(sess)
	}

	sik, err := computeHMAC(sess.AuthAlg, sikInput, sikKey)
	if err != nil {
		return fmt.Errorf("derive SIK: %w", err)
	}
	sess.SIK = sik

	// K1 = HMAC(SIK, 0x01 × 20)
	const20 := [20]byte{
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	}
	k1, err := computeHMAC(sess.AuthAlg, const20[:], sik)
	if err != nil {
		return fmt.Errorf("derive K1: %w", err)
	}
	sess.K1 = k1

	// K2 = HMAC(SIK, 0x02 × 20)
	const202 := [20]byte{
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	}
	k2, err := computeHMAC(sess.AuthAlg, const202[:], sik)
	if err != nil {
		return fmt.Errorf("derive K2: %w", err)
	}
	sess.K2 = k2
	return nil
}

// ---------------------------------------------------------------------------
// internal helpers
// ---------------------------------------------------------------------------

// hmacKey returns the HMAC key for RAKP2/3: Kuid (the user's password padded to 20 bytes).
func hmacKey(sess *bmc.Session, _ *bmc.BMC) []byte {
	return paddedPassword(sess)
}

// paddedPassword returns the user password padded to 20 bytes (Kuid).
func paddedPassword(sess *bmc.Session) []byte {
	if sess.User == nil {
		return make([]byte, 20)
	}
	key := make([]byte, 20)
	copy(key, sess.User.Password[:])
	return key
}

// computeHMAC selects the HMAC variant based on the auth algorithm.
// Returns the full-length auth code (RAKP2/RAKP3 use the full digest; SIK/K1/K2
// derivation also uses the full digest). Supported: RAKP-HMAC-SHA1 (20B) and
// RAKP-HMAC-SHA256 (32B).
func computeHMAC(alg bmc.AuthAlg, data, key []byte) ([]byte, error) {
	switch alg {
	case bmc.AuthAlgNone:
		return nil, nil
	case bmc.AuthAlgHMACSHA1:
		return doHMACSHA1(data, key), nil
	case bmc.AuthAlgHMACSHA256:
		return doHMACSHA256(data, key), nil
	default:
		return nil, fmt.Errorf("unsupported auth algorithm: %d", alg)
	}
}

// computeHMACIntegrity selects the HMAC variant based on the integrity algorithm
// for RAKP4 and session trailer AuthCode. The digest is truncated to the
// algorithm's integrity length: HMAC-SHA1-96 → 12 bytes, HMAC-SHA256-128 → 16
// bytes (spec §13.28).
func computeHMACIntegrity(alg bmc.IntegrityAlg, data, key []byte) ([]byte, error) {
	switch alg {
	case bmc.IntegrityAlgNone:
		return nil, nil
	case bmc.IntegrityAlgHMACSHA1_96:
		full := doHMACSHA1(data, key)
		return full[:12], nil // truncated to 96 bits
	case bmc.IntegrityAlgHMACSHA256_128:
		full := doHMACSHA256(data, key)
		return full[:16], nil // truncated to 128 bits
	default:
		return nil, fmt.Errorf("unsupported integrity algorithm: %d", alg)
	}
}

func doHMACSHA1(data, key []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func doHMACSHA256(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// rakp3AuthCodeLen returns the expected length of the auth code in RAKP3.
func rakp3AuthCodeLen(alg bmc.AuthAlg) int {
	switch alg {
	case bmc.AuthAlgHMACSHA1:
		return 20
	case bmc.AuthAlgHMACMD5:
		return 16
	case bmc.AuthAlgHMACSHA256:
		return 32
	default:
		return 0
	}
}

// hmacEqual compares two HMACs in constant time.
func hmacEqual(a, b []byte) bool {
	return hmac.Equal(a, b)
}
