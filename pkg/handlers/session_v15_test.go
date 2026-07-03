package handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/client"
	ipmi "github.com/bougou/go-ipmi/pkg/types"
)

func TestHandleGetChannelAuthCapsAdvertisesV15AndV20(t *testing.T) {
	b := newTestBMC()
	resp, cc, err := handleGetChannelAuthCaps(context.Background(), &HandlerContext{BMC: b}, []byte{0x8e, 0x04})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if resp[1]&0x04 == 0 {
		t.Fatalf("expected MD5 auth type advertised, byte=0x%02x", resp[1])
	}
	if resp[3]&0x01 == 0 {
		t.Fatalf("expected IPMI v1.5 connection support, byte=0x%02x", resp[3])
	}
	if resp[3]&0x02 == 0 {
		t.Fatalf("expected IPMI v2.0 connection support, byte=0x%02x", resp[3])
	}
}

func TestHandleGetSessionChallenge(t *testing.T) {
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

	req := make([]byte, 17)
	req[0] = uint8(bmc.V15AuthTypeMD5)
	copy(req[1:], []byte("ADMIN"))

	resp, cc, err := handleGetSessionChallenge(context.Background(), &HandlerContext{BMC: b}, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CodeOK {
		t.Fatalf("want CodeOK, got %d", cc)
	}
	if len(resp) != 20 {
		t.Fatalf("want 20 response bytes, got %d", len(resp))
	}
	tempID := binary.LittleEndian.Uint32(resp[0:4])
	if tempID == 0 {
		t.Fatal("expected non-zero temporary session ID")
	}
	sess, err := b.V15Sessions.Get(tempID)
	if err != nil {
		t.Fatalf("pending session not found: %v", err)
	}
	if sess.State != bmc.V15SessionStatePending {
		t.Fatalf("want pending state, got %v", sess.State)
	}
	if !bytes.Equal(sess.Challenge[:], resp[4:20]) {
		t.Fatal("challenge mismatch")
	}
}

func TestGenV15AuthCodeMatchesClient(t *testing.T) {
	password := []byte("ADMIN")
	sessionID := uint32(0xAABBCCDD)
	sessionSeq := uint32(0)
	ipmiData := []byte{0x20, 0x18, 0xc8, 0x81, 0x04, 0x3a}

	serverCode := GenV15AuthCode(password, bmc.V15AuthTypeMD5, sessionID, ipmiData, sessionSeq)

	clientInput := &client.AuthCodeMultiSessionInput{
		Password:   string(password),
		SessionID:  sessionID,
		SessionSeq: sessionSeq,
		IPMIData:   ipmiData,
	}
	clientCode := clientInput.AuthCode(ipmi.AuthTypeMD5)

	if !bytes.Equal(serverCode, clientCode) {
		t.Fatalf("auth code mismatch:\n server=%x\n client=%x", serverCode, clientCode)
	}
}

func TestV15SessionActivationFlow(t *testing.T) {
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

	challengeReq := make([]byte, 17)
	challengeReq[0] = uint8(bmc.V15AuthTypeMD5)
	copy(challengeReq[1:], []byte("ADMIN"))
	challengeResp, cc, err := handleGetSessionChallenge(context.Background(), &HandlerContext{BMC: b}, challengeReq)
	if err != nil || cc != CodeOK {
		t.Fatalf("GetSessionChallenge: cc=%d err=%v", cc, err)
	}
	tempID := binary.LittleEndian.Uint32(challengeResp[0:4])
	sess, _ := b.V15Sessions.Get(tempID)

	activateReq := make([]byte, 22)
	activateReq[0] = uint8(bmc.V15AuthTypeMD5)
	activateReq[1] = uint8(bmc.PrivilegeLevelAdministrator)
	copy(activateReq[2:18], challengeResp[4:20])
	binary.LittleEndian.PutUint32(activateReq[18:22], 0x01020304)

	hctx := &HandlerContext{BMC: b, V15Session: sess, User: user}
	activateResp, cc, err := handleActivateSession(context.Background(), hctx, activateReq)
	if err != nil || cc != CodeOK {
		t.Fatalf("ActivateSession: cc=%d err=%v", cc, err)
	}
	if len(activateResp) != 10 {
		t.Fatalf("want 10-byte response, got %d", len(activateResp))
	}
	permID := binary.LittleEndian.Uint32(activateResp[1:5])
	if permID == 0 || permID == tempID {
		t.Fatalf("unexpected permanent session ID: 0x%08x", permID)
	}
	got, err := b.V15Sessions.Get(permID)
	if err != nil {
		t.Fatalf("active session not found: %v", err)
	}
	if got.State != bmc.V15SessionStateActive {
		t.Fatalf("want active state, got %v", got.State)
	}
	if got.PrivilegeLevel != bmc.PrivilegeLevelUser {
		t.Fatalf("initial privilege want USER, got %v", got.PrivilegeLevel)
	}
	if got.MaxPrivilege != bmc.PrivilegeLevelAdministrator {
		t.Fatalf("max privilege want ADMIN, got %v", got.MaxPrivilege)
	}
	if got.OutboundSeq != 0x01020304 {
		t.Fatalf("outbound seq: want 0x01020304, got 0x%08x", got.OutboundSeq)
	}
}

func TestActivateSessionRejectsChallengeMismatch(t *testing.T) {
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

	challengeReq := make([]byte, 17)
	challengeReq[0] = uint8(bmc.V15AuthTypeMD5)
	copy(challengeReq[1:], []byte("ADMIN"))
	challengeResp, cc, err := handleGetSessionChallenge(context.Background(), &HandlerContext{BMC: b}, challengeReq)
	if err != nil || cc != CodeOK {
		t.Fatalf("GetSessionChallenge: cc=%d err=%v", cc, err)
	}
	tempID := binary.LittleEndian.Uint32(challengeResp[0:4])
	sess, _ := b.V15Sessions.Get(tempID)

	activateReq := make([]byte, 22)
	activateReq[0] = uint8(bmc.V15AuthTypeMD5)
	activateReq[1] = uint8(bmc.PrivilegeLevelAdministrator)
	copy(activateReq[2:18], challengeResp[4:20])
	activateReq[2] ^= 0xff // corrupt challenge
	binary.LittleEndian.PutUint32(activateReq[18:22], 0x01020304)

	hctx := &HandlerContext{BMC: b, V15Session: sess, User: user}
	_, cc, err = handleActivateSession(context.Background(), hctx, activateReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != CCV15InvalidSessionID {
		t.Fatalf("want invalid session id, got cc=%d", cc)
	}
}

func TestV15InboundSeqValid(t *testing.T) {
	tests := []struct {
		high, rcvd, seq uint32
		want            bool
	}{
		{100, 1, 101, true},
		{100, 1, 108, true},
		{100, 1, 109, false},
		{100, 1, 100, false},
		{100, 1, 0, false},
		{100, 1, 95, true},
		{100, 1, 92, true},
		{100, 1, 91, false},
	}
	for _, tc := range tests {
		sess := &bmc.V15Session{InboundSeq: tc.high, InboundRcvd: uint8(tc.rcvd)}
		got := bmc.V15InboundSeqValid(sess, tc.seq)
		if got != tc.want {
			t.Errorf("V15InboundSeqValid(high=%d, seq=%d) = %v, want %v", tc.high, tc.seq, got, tc.want)
		}
	}
}

func TestV15InboundSeqDuplicateRejected(t *testing.T) {
	sess := &bmc.V15Session{InboundSeq: 100, InboundRcvd: 1}
	if !sess.TryAcceptInboundSeq(101) {
		t.Fatal("expected first accept")
	}
	if sess.TryAcceptInboundSeq(101) {
		t.Fatal("duplicate seq should be rejected")
	}
}
