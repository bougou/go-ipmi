package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/handlers"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

// TestRAKP_SHA256_CrossValidation drives the reference server's RAKP1/RAKP3
// handlers with cipher suite 17 (RAKP-HMAC-SHA256 + HMAC-SHA256-128) and
// verifies every auth code / derived key against the client-side generators.
// This is the cross-package equivalent of an ipmitool -C 17 handshake.
func TestRAKP_SHA256_CrossValidation(t *testing.T) {
	const (
		username = "ADMIN"
		password = "ADMIN"
	)
	consoleID := uint32(0x11223344)
	role := uint8(bmc.PrivilegeLevelAdministrator)

	b := newSHA256TestBMC(t, username, password)

	sess, err := b.Sessions.Allocate(consoleID,
		bmc.AuthAlgHMACSHA256, bmc.IntegrityAlgHMACSHA256_128, bmc.CryptAlgAESCBC128)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	sess.Channel = 1
	sess.MaxPrivilege = bmc.PrivilegeLevelAdministrator

	// Build RAKP1 with a fixed console random for reproducibility.
	var consoleRand [16]byte
	for i := range consoleRand {
		consoleRand[i] = byte(0xA0 + i)
	}
	rakp1 := buildRAKP1(sess.BMCID, role, consoleRand, username)

	rakp2, err := handlers.HandleRAKP1(context.Background(), b, rakp1)
	if err != nil {
		t.Fatalf("HandleRAKP1: %v", err)
	}
	if len(rakp2) < 40 || rakp2[1] != 0x00 {
		t.Fatalf("RAKP2 not successful: len=%d status=0x%02x", len(rakp2), rakp2Status(rakp2))
	}

	// Parse RAKP2: tag(1) status(1) reserved(2) consoleID(4) bmcRand(16) bmcGUID(16) authCode(N).
	var bmcRand [16]byte
	copy(bmcRand[:], rakp2[8:24])
	var bmcGUID [16]byte
	copy(bmcGUID[:], rakp2[24:40])
	serverRAKP2Code := rakp2[40:] // SHA256 → 32 bytes

	// Configure the client session to mirror the exchange and let the client
	// recompute the RAKP2 auth code.
	c := newTestClient(t, username, password)
	c.session.v20.authAlg = ipmi.AuthAlgRAKP_HMAC_SHA256
	c.session.v20.integrityAlg = ipmi.IntegrityAlg_HMAC_SHA256_128
	c.session.v20.consoleSessionID = consoleID
	c.session.v20.bmcSessionID = sess.BMCID
	c.session.v20.consoleRand = consoleRand
	c.session.v20.bmcRand = bmcRand
	c.session.v20.bmcGUID = bmcGUID
	c.session.v20.role = role

	clientRAKP2Code, err := c.generate_rakp2_authcode()
	if err != nil {
		t.Fatalf("client generate_rakp2_authcode: %v", err)
	}
	if !bytes.Equal(clientRAKP2Code, serverRAKP2Code) {
		t.Fatalf("RAKP2 auth code mismatch:\n client=%x\n server=%x", clientRAKP2Code, serverRAKP2Code)
	}

	// Build RAKP3 with the client's RAKP3 auth code and feed it to the server.
	clientRAKP3Code, err := c.generate_rakp3_authcode()
	if err != nil {
		t.Fatalf("client generate_rakp3_authcode: %v", err)
	}
	rakp3 := buildRAKP3(sess.BMCID, clientRAKP3Code)

	rakp4, err := handlers.HandleRAKP3(context.Background(), b, rakp3)
	if err != nil {
		t.Fatalf("HandleRAKP3: %v", err)
	}
	if len(rakp4) < 8 || rakp4[1] != 0x00 {
		t.Fatalf("RAKP4 not successful: len=%d status=0x%02x", len(rakp4), rakp2Status(rakp4))
	}

	// SIK / K1 / K2 must match between client and server. The client stores SIK
	// on the session before deriving K1/K2.
	clientSIK, err := c.generate_sik()
	if err != nil {
		t.Fatalf("client generate_sik: %v", err)
	}
	c.session.v20.sik = clientSIK
	if !bytes.Equal(clientSIK, sess.SIK) {
		t.Fatalf("SIK mismatch:\n client=%x\n server=%x", clientSIK, sess.SIK)
	}
	clientK1, err := c.generate_k1()
	if err != nil {
		t.Fatalf("client generate_k1: %v", err)
	}
	if !bytes.Equal(clientK1, sess.K1) {
		t.Fatalf("K1 mismatch:\n client=%x\n server=%x", clientK1, sess.K1)
	}
	clientK2, err := c.generate_k2()
	if err != nil {
		t.Fatalf("client generate_k2: %v", err)
	}
	if !bytes.Equal(clientK2, sess.K2) {
		t.Fatalf("K2 mismatch:\n client=%x\n server=%x", clientK2, sess.K2)
	}

	// RAKP4 auth code (HMAC-SHA256-128, 16 bytes) must match.
	clientRAKP4Code, err := c.generate_rakp4_authcode()
	if err != nil {
		t.Fatalf("client generate_rakp4_authcode: %v", err)
	}
	serverRAKP4Code := rakp4[8:]
	if !bytes.Equal(clientRAKP4Code, serverRAKP4Code) {
		t.Fatalf("RAKP4 auth code mismatch:\n client=%x\n server=%x", clientRAKP4Code, serverRAKP4Code)
	}
	if len(serverRAKP4Code) != 16 {
		t.Fatalf("RAKP4 auth code length: want 16 (SHA256-128), got %d", len(serverRAKP4Code))
	}
}

func newSHA256TestBMC(t *testing.T, username, password string) *bmc.BMC {
	t.Helper()
	info := bmc.DeviceInfo{DeviceID: 1, IPMIVersion: 0x20, ManufacturerID: 0x000157, ProductID: 0x0001}
	var guid [16]byte
	for i := range guid {
		guid[i] = byte(0x10 + i)
	}
	b := bmc.New(info, guid, mock.New(), bmc.WithClock(clock.Real))
	user, err := b.Users.Add(2, username)
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.SetPassword([]byte(password))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}
	return b
}

func newTestClient(t *testing.T, username, password string) *Client {
	t.Helper()
	c, err := NewClient("127.0.0.1", 623, username, password)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// buildRAKP1 constructs a RAKP Message 1 payload (spec §13.21):
//
//	tag(1) status(1) reserved(2) bmcSessionID(4) consoleRand(16) role(1) reserved(2) userLen(1) username(N)
func buildRAKP1(bmcSessionID uint32, role uint8, consoleRand [16]byte, username string) []byte {
	p := make([]byte, 28+len(username))
	p[0] = 0x01 // tag
	binary.LittleEndian.PutUint32(p[4:8], bmcSessionID)
	copy(p[8:24], consoleRand[:])
	p[24] = role
	p[27] = uint8(len(username))
	copy(p[28:], username)
	return p
}

// buildRAKP3 constructs a RAKP Message 3 payload (spec §13.23):
//
//	tag(1) status(1) reserved(2) bmcSessionID(4) authCode(N)
func buildRAKP3(bmcSessionID uint32, authCode []byte) []byte {
	p := make([]byte, 8+len(authCode))
	p[0] = 0x03 // tag
	binary.LittleEndian.PutUint32(p[4:8], bmcSessionID)
	copy(p[8:], authCode)
	return p
}

func rakp2Status(resp []byte) uint8 {
	if len(resp) < 2 {
		return 0xff
	}
	return resp[1]
}
