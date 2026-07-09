package handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestHandleGetChannelAuthCapsAdvertisesRMCPPlusOnly(t *testing.T) {
	b := newTestBMC()
	resp, cc, err := handleGetChannelAuthCaps(context.Background(), &HandlerContext{BMC: b}, []byte{0x8e, 0x04})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if len(resp) != 8 {
		t.Fatalf("want 8 response bytes, got %d", len(resp))
	}
	if resp[1]&0x80 == 0 {
		t.Fatalf("expected IPMI v2.0 extended capabilities to be available, byte=0x%02x", resp[1])
	}
	if resp[1]&0x04 == 0 {
		t.Fatalf("expected MD5 v1.5 auth type, byte=0x%02x", resp[1])
	}
	if resp[3]&0x02 == 0 {
		t.Fatalf("expected IPMI v2.0 connection support, byte=0x%02x", resp[3])
	}
	if resp[3]&0x01 == 0 {
		t.Fatalf("expected IPMI v1.5 connection support, byte=0x%02x", resp[3])
	}
}

func TestHandleRAKP1RejectsUnauthorizedPrivilege(t *testing.T) {
	b := newTestBMC()
	user, err := b.Users.Add(2, "operator")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte("secret"))
	user.Enabled = true
	user.ChannelAccess[lanChannelNumber] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelUser,
		Enabled:      true,
	}
	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	sess.Channel = lanChannelNumber
	sess.MaxPrivilege = bmc.PrivilegeLevelAdministrator

	resp, err := HandleRAKP1(context.Background(), b, rakp1Payload(sess.BMCID, bmc.PrivilegeLevelAdministrator, "operator"))
	if err != nil {
		t.Fatalf("HandleRAKP1: %v", err)
	}
	if len(resp) < 2 || resp[1] != 0x0A {
		t.Fatalf("want unauthorized privilege status 0x0a, got %x", resp)
	}
	if _, err := b.Sessions.Get(sess.BMCID); err == nil {
		t.Fatalf("unauthorized session should be closed")
	}
}

func TestHandleRAKP1AcceptsAuthorizedPrivilege(t *testing.T) {
	b := newTestBMC()
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[lanChannelNumber] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	sess.Channel = lanChannelNumber
	sess.MaxPrivilege = bmc.PrivilegeLevelAdministrator

	resp, err := HandleRAKP1(context.Background(), b, rakp1Payload(sess.BMCID, bmc.PrivilegeLevelAdministrator, "ADMIN"))
	if err != nil {
		t.Fatalf("HandleRAKP1: %v", err)
	}
	if len(resp) < 40 || resp[1] != 0x00 {
		t.Fatalf("want successful RAKP2 response, got %x", resp)
	}
	if sess.User != user {
		t.Fatalf("session user was not recorded")
	}
}

func rakp1Payload(bmcSessionID uint32, role bmc.PrivilegeLevel, username string) []byte {
	payload := make([]byte, 28+len(username))
	payload[0] = 0x01
	binary.LittleEndian.PutUint32(payload[4:8], bmcSessionID)
	for i := range payload[8:24] {
		payload[8+i] = byte(i + 1)
	}
	payload[24] = uint8(role)
	payload[27] = uint8(len(username))
	copy(payload[28:], username)
	return payload
}

// openSessionPayload builds a 32-byte RMCP+ Open Session Request payload with
// the given algorithm codes. Record layout per spec §13.17:
//
//	tag(1) maxPriv(1) reserved(2) consoleID(4)
//	[auth: type=0 reserved(2) 0x08 alg(1) reserved(3)]
//	[integ: type=1 reserved(2) 0x08 alg(1) reserved(3)]
//	[crypt: type=2 reserved(2) 0x08 alg(1) reserved(3)]
func openSessionPayload(tag, maxPriv uint8, consoleID uint32, authAlg, intAlg, cryptAlg uint8) []byte {
	p := make([]byte, 32)
	p[0] = tag
	p[1] = maxPriv
	binary.LittleEndian.PutUint32(p[4:8], consoleID)
	p[11], p[12], p[15] = 0x08, authAlg, 0x08
	p[19], p[20], p[23] = 0x08, intAlg, 0x08
	p[27], p[28], p[31] = 0x08, cryptAlg, 0x08
	p[8] = 0x00  // auth payload type
	p[16] = 0x01 // integrity payload type
	p[24] = 0x02 // confidentiality payload type
	return p
}

func TestHandleGetChannelCipherSuites_Default(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	// listIndex 0: channel byte + 10 record bytes (suite 3 + suite 17).
	resp, cc, err := handleGetChannelCipherSuites(context.Background(), hctx, []byte{0x8e, 0x00, 0x00})
	if err != nil || cc != CodeOK {
		t.Fatalf("unexpected cc=%d err=%v", cc, err)
	}
	if len(resp) != 11 {
		t.Fatalf("want 11 bytes (1 channel + 10 record), got %d", len(resp))
	}
	wantRecords := []byte{0xC0, 0x03, 0x01, 0x41, 0x81, 0xC0, 0x11, 0x03, 0x44, 0x81}
	if !bytes.Equal(resp[1:], wantRecords) {
		t.Fatalf("records: want %x, got %x", wantRecords, resp[1:])
	}

	// listIndex 1: beyond the 10-byte record window, only channel byte remains.
	resp2, _, _ := handleGetChannelCipherSuites(context.Background(), hctx, []byte{0x8e, 0x00, 0x01})
	if len(resp2) != 1 {
		t.Fatalf("listIndex 1: want 1 byte (channel only), got %d", len(resp2))
	}
}

func TestHandleGetChannelCipherSuites_Custom(t *testing.T) {
	b := newTestBMC()
	b.SetCipherSuites([]types.CipherSuiteID{types.CipherSuiteID17})
	hctx := &HandlerContext{BMC: b}

	resp, cc, err := handleGetChannelCipherSuites(context.Background(), hctx, []byte{0x8e, 0x00, 0x00})
	if err != nil || cc != CodeOK {
		t.Fatalf("unexpected cc=%d err=%v", cc, err)
	}
	if len(resp) != 6 {
		t.Fatalf("want 6 bytes (1 channel + 5 record), got %d", len(resp))
	}
	wantRecords := []byte{0xC0, 0x11, 0x03, 0x44, 0x81}
	if !bytes.Equal(resp[1:], wantRecords) {
		t.Fatalf("records: want %x, got %x", wantRecords, resp[1:])
	}
}

func TestHandleOpenSession_AcceptsSHA256(t *testing.T) {
	b := newTestBMC()
	payload := openSessionPayload(0x01, 0x04, 0x01020304,
		uint8(types.AuthAlg_HMAC_SHA256), uint8(types.IntegrityAlg_HMAC_SHA256_128), uint8(types.CryptAlg_AES_CBC_128))

	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 36 || resp[1] != 0x00 {
		t.Fatalf("want success 36-byte response, got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
	if resp[16] != uint8(types.AuthAlg_HMAC_SHA256) {
		t.Fatalf("auth alg not echoed: 0x%02x", resp[16])
	}
	if resp[24] != uint8(types.IntegrityAlg_HMAC_SHA256_128) {
		t.Fatalf("integrity alg not echoed: 0x%02x", resp[24])
	}
	if resp[32] != uint8(types.CryptAlg_AES_CBC_128) {
		t.Fatalf("crypt alg not echoed: 0x%02x", resp[32])
	}
}

func TestHandleOpenSession_RejectsUnsupported(t *testing.T) {
	b := newTestBMC()
	// MD5 auth (0x02) is not part of any configured suite (default {3,17}).
	payload := openSessionPayload(0x01, 0x04, 0x01020304,
		uint8(types.AuthAlg_HMAC_MD5), uint8(types.IntegrityAlg_HMAC_SHA1_96), uint8(types.CryptAlg_AES_CBC_128))

	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 8 || resp[1] != 0x04 {
		t.Fatalf("want status 0x04 (invalid auth alg), got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
}

// TestHandleOpenSession_RejectsNoneByDefault verifies that the default cipher
// suite set ({3, 17}) does NOT allow negotiating AuthAlgNone / IntegrityAlgNone
// / CryptAlgNone. None is only reachable by explicitly configuring a suite that
// uses it (e.g. suite 0). This guards against an authentication bypass where
// AuthAlgNone skips RAKP password verification.
func TestHandleOpenSession_RejectsNoneByDefault(t *testing.T) {
	cases := []struct {
		name    string
		auth    types.AuthAlg
		integ   types.IntegrityAlg
		crypt   types.CryptAlg
		wantErr uint8
	}{
		{"auth none", types.AuthAlg_None, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, 0x04},
		{"integrity none", types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_None, types.CryptAlg_AES_CBC_128, 0x05},
		{"crypt none", types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_None, 0x10},
		{"all none", types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None, 0x04},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newTestBMC()
			payload := openSessionPayload(0x01, 0x04, 0x01020304,
				uint8(tc.auth), uint8(tc.integ), uint8(tc.crypt))
			resp, err := HandleOpenSession(context.Background(), b, payload)
			if err != nil {
				t.Fatalf("HandleOpenSession: %v", err)
			}
			if len(resp) != 8 || resp[1] != tc.wantErr {
				t.Fatalf("want status 0x%02x, got len=%d status=0x%02x", tc.wantErr, len(resp), safeStatus(resp))
			}
		})
	}
}

// TestHandleOpenSession_AcceptsNoneWhenSuite0Configured verifies that an
// operator who explicitly configures suite 0 (unauthenticated) can still
// negotiate AuthAlgNone. This is the spec-defined way to opt in to
// unauthenticated sessions; the security choice is the operator's.
func TestHandleOpenSession_AcceptsNoneWhenSuite0Configured(t *testing.T) {
	b := newTestBMC()
	b.SetCipherSuites([]types.CipherSuiteID{types.CipherSuiteID0})

	payload := openSessionPayload(0x01, 0x04, 0x01020304,
		uint8(types.AuthAlg_None), uint8(types.IntegrityAlg_None), uint8(types.CryptAlg_None))
	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 36 || resp[1] != 0x00 {
		t.Fatalf("want success 36-byte response, got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
}

// TestHandleOpenSession_AcceptsMixedSuites verifies that configuring a mix of
// authenticated and unencrypted suites (e.g. 3 + 15) allows each suite's
// algorithm triple to be negotiated individually, including IntegrityAlgNone
// and CryptAlgNone from suite 15.
func TestHandleOpenSession_AcceptsMixedSuites(t *testing.T) {
	b := newTestBMC()
	b.SetCipherSuites([]types.CipherSuiteID{types.CipherSuiteID3, types.CipherSuiteID15})

	// Suite 15 triple: HMAC-SHA256 auth + None integ + None crypt.
	payload := openSessionPayload(0x01, 0x04, 0x01020304,
		uint8(types.AuthAlg_HMAC_SHA256), uint8(types.IntegrityAlg_None), uint8(types.CryptAlg_None))
	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 36 || resp[1] != 0x00 {
		t.Fatalf("want success 36-byte response, got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
}

// TestHandleOpenSession_RejectsCrossSuiteRecombination verifies that the
// server rejects algorithm triples formed by recombining algorithms from
// different configured suites. Configuring suites {2, 17} must NOT accept
// suite 3 (SHA1 auth + SHA1-96 integ + AES crypt), even though each
// component exists individually (SHA1 and SHA1-96 from suite 2, AES from
// suite 17). The triple must come from a single configured suite.
func TestHandleOpenSession_RejectsCrossSuiteRecombination(t *testing.T) {
	type testCase struct {
		name       string
		suites     []types.CipherSuiteID
		auth       types.AuthAlg
		integ      types.IntegrityAlg
		crypt      types.CryptAlg
		wantStatus uint8
	}

	cases := []testCase{
		{
			// Suites {2, 17}: each component of suite 3 exists individually
			// (SHA1/SHA1-96 from suite 2, AES from suite 17), but no single
			// configured suite contains the triple.
			name:       "{2,17} should reject suite 3 (cross-suite SHA1+SHA1-96+AES)",
			suites:     []types.CipherSuiteID{types.CipherSuiteID2, types.CipherSuiteID17},
			auth:       types.AuthAlg_HMAC_SHA1,
			integ:      types.IntegrityAlg_HMAC_SHA1_96,
			crypt:      types.CryptAlg_AES_CBC_128,
			wantStatus: 0x04,
		},
		{
			// Suites {2, 17}: SHA256+SHA256-128+None = suite 16. Suite 16
			// is not configured; SHA256/SHA256-128 are from suite 17, None
			// is from suite 2's crypt.
			name:       "{2,17} should reject suite 16 (cross-suite SHA256+SHA256-128+None)",
			suites:     []types.CipherSuiteID{types.CipherSuiteID2, types.CipherSuiteID17},
			auth:       types.AuthAlg_HMAC_SHA256,
			integ:      types.IntegrityAlg_HMAC_SHA256_128,
			crypt:      types.CryptAlg_None,
			wantStatus: 0x04,
		},
		{
			// Auth bypass: suites {3, 15} create auth={SHA1,SHA256},
			// integ={SHA1-96,None}, crypt={AES,None}. Without triple
			// validation, suite 1 (SHA1+None+None — authenticated but no
			// integrity check) would be accepted even though it was never
			// configured. This is the most dangerous recombination.
			name:       "{3,15} should reject suite 1 (auth bypass SHA1+None+None)",
			suites:     []types.CipherSuiteID{types.CipherSuiteID3, types.CipherSuiteID15},
			auth:       types.AuthAlg_HMAC_SHA1,
			integ:      types.IntegrityAlg_None,
			crypt:      types.CryptAlg_None,
			wantStatus: 0x04,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newTestBMC()
			b.SetCipherSuites(tc.suites)

			payload := openSessionPayload(0x01, 0x04, 0x01020304,
				uint8(tc.auth), uint8(tc.integ), uint8(tc.crypt))
			resp, err := HandleOpenSession(context.Background(), b, payload)
			if err != nil {
				t.Fatalf("HandleOpenSession: %v", err)
			}
			if len(resp) != 8 || resp[1] != tc.wantStatus {
				t.Fatalf("want status 0x%02x, got len=%d status=0x%02x", tc.wantStatus, len(resp), safeStatus(resp))
			}
		})
	}
}

// TestComputeRAKP4AuthCode_UsesAuthAlgorithm verifies that the RAKP4 Integrity
// Check Value is selected by the *authentication* algorithm (spec §13.28.1 /
// §13.28.1b / §13.31), not the session integrity algorithm. This matters for
// suites that pair a non-None auth algorithm with Integrity=None (suites 1 and
// 15): the ICV must still be 12 / 16 bytes, not absent.
func TestComputeRAKP4AuthCode_UsesAuthAlgorithm(t *testing.T) {
	b := newTestBMC()
	cases := []struct {
		name    string
		auth    types.AuthAlg
		integ   types.IntegrityAlg
		wantLen int
	}{
		{"SHA1 auth + None integ (suite 1)", types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_None, 12},
		{"SHA256 auth + None integ (suite 15)", types.AuthAlg_HMAC_SHA256, types.IntegrityAlg_None, 16},
		{"SHA1 auth + SHA1-96 integ (suite 3)", types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, 12},
		{"SHA256 auth + SHA256-128 integ (suite 17)", types.AuthAlg_HMAC_SHA256, types.IntegrityAlg_HMAC_SHA256_128, 16},
		{"None auth + None integ (suite 0)", types.AuthAlg_None, types.IntegrityAlg_None, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess, err := b.Sessions.Allocate(0x11223344, tc.auth, tc.integ, types.CryptAlg_None)
			if err != nil {
				t.Fatalf("allocate session: %v", err)
			}
			for i := range sess.ConsoleRand {
				sess.ConsoleRand[i] = byte(0xA0 + i)
			}
			sess.SIK = bytes.Repeat([]byte{0x5a}, 32) // any key longer than block size is fine

			code, err := computeRAKP4AuthCode(sess, b)
			if err != nil {
				t.Fatalf("computeRAKP4AuthCode: %v", err)
			}
			if len(code) != tc.wantLen {
				t.Fatalf("ICV length: want %d, got %d (code=%x)", tc.wantLen, len(code), code)
			}
		})
	}
}

func safeStatus(resp []byte) uint8 {
	if len(resp) < 2 {
		return 0xff
	}
	return resp[1]
}
