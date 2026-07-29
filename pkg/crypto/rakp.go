package crypto

import (
	"encoding/binary"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// PadPassword20 returns the user password padded to 20 bytes (Kuid).
// Spec: v2.0§13.31 (RAKP key exchange uses 160-bit Kuid).
func PadPassword20(password []byte) []byte {
	key := make([]byte, 20)
	copy(key, password)
	return key
}

// DeriveSIK computes the Session Integrity Key (v2.0§13.31):
//
//	SIK = HMAC(Kg or Kuid, ConsoleRand || BMCRand || Role || UserLen || Username)
//
// kgOrKuid should be Kg when configured, otherwise the 20-byte padded password.
func DeriveSIK(authAlg types.AuthAlg, consoleRand, bmcRand []byte, role uint8, username string, kgOrKuid []byte) ([]byte, error) {
	sikInput := make([]byte, 16+16+1+1+len(username))
	copy(sikInput[0:16], consoleRand)
	copy(sikInput[16:32], bmcRand)
	sikInput[32] = role
	sikInput[33] = uint8(len(username))
	copy(sikInput[34:], username)
	return AuthHMAC(authAlg, sikInput, kgOrKuid)
}

// DeriveK1 computes K1 = HMAC(SIK, 0x01 × 20) per v2.0§13.32.
func DeriveK1(authAlg types.AuthAlg, sik []byte) ([]byte, error) {
	const1 := [20]byte{
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	}
	return AuthHMAC(authAlg, const1[:], sik)
}

// DeriveK2 computes K2 = HMAC(SIK, 0x02 × 20) per v2.0§13.32.
func DeriveK2(authAlg types.AuthAlg, sik []byte) ([]byte, error) {
	const2 := [20]byte{
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
		0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
	}
	return AuthHMAC(authAlg, const2[:], sik)
}

// DeriveSessionKeys computes SIK, K1, and K2 from session parameters.
func DeriveSessionKeys(authAlg types.AuthAlg, consoleRand, bmcRand []byte, role uint8, username string, kgOrKuid []byte) (sik, k1, k2 []byte, err error) {
	sik, err = DeriveSIK(authAlg, consoleRand, bmcRand, role, username, kgOrKuid)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("derive SIK: %w", err)
	}
	k1, err = DeriveK1(authAlg, sik)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("derive K1: %w", err)
	}
	k2, err = DeriveK2(authAlg, sik)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("derive K2: %w", err)
	}
	return sik, k1, k2, nil
}

// RAKP2AuthCode generates the Key Exchange Authentication Code for RAKP Message 2
// (v2.0§13.31):
//
//	HMAC(Kuid, ConsoleID || BMCID || ConsoleRand || BMCRand || BMCGUID || Role || UserLen || Username)
func RAKP2AuthCode(authAlg types.AuthAlg, consoleID, bmcID uint32, consoleRand, bmcRand, bmcGUID []byte, role uint8, username string, kuid []byte) ([]byte, error) {
	if authAlg == types.AuthAlg_None {
		return nil, nil
	}
	buf := make([]byte, 4+4+16+16+16+1+1+len(username))
	binary.LittleEndian.PutUint32(buf[0:4], consoleID)
	binary.LittleEndian.PutUint32(buf[4:8], bmcID)
	copy(buf[8:24], consoleRand)
	copy(buf[24:40], bmcRand)
	copy(buf[40:56], bmcGUID)
	buf[56] = role
	buf[57] = uint8(len(username))
	copy(buf[58:], username)
	return AuthHMAC(authAlg, buf, kuid)
}

// RAKP3AuthCode generates the auth code for RAKP Message 3 (v2.0§13.31):
//
//	HMAC(Kuid, BMCRand || ConsoleID || Role || UserLen || Username)
func RAKP3AuthCode(authAlg types.AuthAlg, bmcRand []byte, consoleID uint32, role uint8, username string, kuid []byte) ([]byte, error) {
	if authAlg == types.AuthAlg_None {
		return nil, nil
	}
	buf := make([]byte, 16+4+1+1+len(username))
	copy(buf[0:16], bmcRand)
	binary.LittleEndian.PutUint32(buf[16:20], consoleID)
	buf[20] = role
	buf[21] = uint8(len(username))
	copy(buf[22:], username)
	return AuthHMAC(authAlg, buf, kuid)
}

// RAKP4ICV generates the Integrity Check Value for RAKP Message 4.
//
// HMAC input (v2.0§13.31): ConsoleRand || BMCID || BMCGUID
//
// Per v2.0§13.28.1 / §13.28.1b the ICV truncation is selected by the
// *authentication* algorithm (not the session integrity algorithm), using SIK
// as the HMAC key:
//   - RAKP-HMAC-SHA1   → HMAC-SHA1-96   (12 bytes)
//   - RAKP-HMAC-SHA256 → HMAC-SHA256-128 (16 bytes)
//   - RAKP-HMAC-MD5    → HMAC-MD5-128    (16 bytes)
//   - RAKP-none        → absent
func RAKP4ICV(authAlg types.AuthAlg, consoleRand []byte, bmcID uint32, bmcGUID, sik []byte) ([]byte, error) {
	if authAlg == types.AuthAlg_None {
		return nil, nil
	}
	buf := make([]byte, 16+4+16)
	copy(buf[0:16], consoleRand)
	binary.LittleEndian.PutUint32(buf[16:20], bmcID)
	copy(buf[20:36], bmcGUID)

	full, err := AuthHMAC(authAlg, buf, sik)
	if err != nil {
		return nil, err
	}
	switch authAlg {
	case types.AuthAlg_HMAC_SHA1:
		if len(full) < 12 {
			return nil, fmt.Errorf("rakp4: hmac sha1 length %d too short for SHA1-96", len(full))
		}
		return full[:12], nil
	case types.AuthAlg_HMAC_SHA256:
		if len(full) < 16 {
			return nil, fmt.Errorf("rakp4: hmac sha256 length %d too short for SHA256-128", len(full))
		}
		return full[:16], nil
	case types.AuthAlg_HMAC_MD5:
		if len(full) < 16 {
			return nil, fmt.Errorf("rakp4: hmac md5 length %d too short for MD5-128", len(full))
		}
		return full[:16], nil
	default:
		return nil, fmt.Errorf("unsupported auth algorithm for RAKP4 ICV: %d", authAlg)
	}
}

// RAKP2AuthCodeLen returns the expected Key Exchange Authentication Code length
// in RAKP Message 2 for the given authentication algorithm.
func RAKP2AuthCodeLen(alg types.AuthAlg) int {
	switch alg {
	case types.AuthAlg_HMAC_SHA1:
		return 20
	case types.AuthAlg_HMAC_MD5:
		return 16
	case types.AuthAlg_HMAC_SHA256:
		return 32
	default:
		return 0
	}
}

// RAKP3AuthCodeLen returns the expected auth code length in RAKP Message 3.
func RAKP3AuthCodeLen(alg types.AuthAlg) int {
	return RAKP2AuthCodeLen(alg)
}

// RAKP4ICVLen returns the expected Integrity Check Value length in RAKP Message 4.
func RAKP4ICVLen(alg types.AuthAlg) int {
	switch alg {
	case types.AuthAlg_HMAC_SHA1:
		return 12
	case types.AuthAlg_HMAC_MD5, types.AuthAlg_HMAC_SHA256:
		return 16
	default:
		return 0
	}
}

// TruncateRAKP2AuthCode truncates a full AuthHMAC digest to the RAKP2 wire length.
func TruncateRAKP2AuthCode(alg types.AuthAlg, full []byte) ([]byte, error) {
	n := RAKP2AuthCodeLen(alg)
	if n == 0 {
		return full, nil
	}
	if len(full) < n {
		return nil, fmt.Errorf("hmac length %d too short for auth alg 0x%x (need %d)", len(full), alg, n)
	}
	return full[:n], nil
}
