package handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

func TestHandleGetChannelAuthCapsAdvertisesRMCPPlusOnly(t *testing.T) {
	resp, cc, err := handleGetChannelAuthCaps(context.Background(), &HandlerContext{}, []byte{0x8e, 0x04})
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
	if resp[3]&0x02 == 0 {
		t.Fatalf("expected IPMI v2.0 connection support, byte=0x%02x", resp[3])
	}
	if resp[3]&0x01 != 0 {
		t.Fatalf("must not advertise IPMI v1.5 session support, byte=0x%02x", resp[3])
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
	sess, err := b.Sessions.Allocate(0x01020304, bmc.AuthAlgHMACSHA1, bmc.IntegrityAlgHMACSHA1_96, bmc.CryptAlgAESCBC128)
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
	sess, err := b.Sessions.Allocate(0x01020304, bmc.AuthAlgHMACSHA1, bmc.IntegrityAlgHMACSHA1_96, bmc.CryptAlgAESCBC128)
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
	b.SetCipherSuites([]bmc.CipherSuiteID{bmc.CipherSuiteID17})
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
		uint8(bmc.AuthAlgHMACSHA256), uint8(bmc.IntegrityAlgHMACSHA256_128), uint8(bmc.CryptAlgAESCBC128))

	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 36 || resp[1] != 0x00 {
		t.Fatalf("want success 36-byte response, got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
	if resp[16] != uint8(bmc.AuthAlgHMACSHA256) {
		t.Fatalf("auth alg not echoed: 0x%02x", resp[16])
	}
	if resp[24] != uint8(bmc.IntegrityAlgHMACSHA256_128) {
		t.Fatalf("integrity alg not echoed: 0x%02x", resp[24])
	}
	if resp[32] != uint8(bmc.CryptAlgAESCBC128) {
		t.Fatalf("crypt alg not echoed: 0x%02x", resp[32])
	}
}

func TestHandleOpenSession_RejectsUnsupported(t *testing.T) {
	b := newTestBMC()
	// MD5 auth (0x02) is not part of any configured suite (default {3,17}).
	payload := openSessionPayload(0x01, 0x04, 0x01020304,
		uint8(bmc.AuthAlgHMACMD5), uint8(bmc.IntegrityAlgHMACSHA1_96), uint8(bmc.CryptAlgAESCBC128))

	resp, err := HandleOpenSession(context.Background(), b, payload)
	if err != nil {
		t.Fatalf("HandleOpenSession: %v", err)
	}
	if len(resp) != 8 || resp[1] != 0x04 {
		t.Fatalf("want status 0x04 (invalid auth alg), got len=%d status=0x%02x", len(resp), safeStatus(resp))
	}
}

func safeStatus(resp []byte) uint8 {
	if len(resp) < 2 {
		return 0xff
	}
	return resp[1]
}
