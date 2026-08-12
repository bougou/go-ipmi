package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// activeRMCPSession builds an active RMCP+ session owned by an enabled user and
// returns it ready to hang off a HandlerContext.
func activeRMCPSession(t *testing.T, b *bmc.BMC, userID uint8, priv bmc.PrivilegeLevel) *bmc.Session {
	t.Helper()
	user, err := b.Users.Add(userID, "operator")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.Enabled = true

	sess, err := b.Sessions.Allocate(0x01020304, types.AuthAlg_HMAC_SHA1, types.IntegrityAlg_HMAC_SHA1_96, types.CryptAlg_AES_CBC_128, priv, lanChannelNumber)
	if err != nil {
		t.Fatalf("allocate session: %v", err)
	}
	sess.State = bmc.SessionStateActive
	sess.User = user
	sess.PrivilegeLevel = priv
	return sess
}

func TestHandleGetSessionInfoCurrentRMCP(t *testing.T) {
	b := newTestBMC()
	sess := activeRMCPSession(t, b, 5, bmc.PrivilegeLevelAdministrator)
	hctx := &HandlerContext{BMC: b, Session: sess}

	resp, cc, err := handleGetSessionInfo(context.Background(), hctx, []byte{sessionIndexCurrent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeOK {
		t.Fatalf("completion code = 0x%02x, want OK", uint8(cc))
	}
	if len(resp) != 6 {
		t.Fatalf("response length = %d, want 6", len(resp))
	}
	if resp[0] == 0 {
		t.Errorf("session handle must be non-zero")
	}
	if want := uint8(b.Sessions.Cap()); resp[1] != want {
		t.Errorf("possible sessions = %d, want %d", resp[1], want)
	}
	// The RMCP+ store holds exactly this one session.
	if resp[2] != 1 {
		t.Errorf("active sessions = %d, want 1", resp[2])
	}
	if resp[3] != 5 {
		t.Errorf("user id = %d, want 5", resp[3])
	}
	if resp[4] != uint8(bmc.PrivilegeLevelAdministrator) {
		t.Errorf("privilege level = 0x%02x, want 0x%02x", resp[4], uint8(bmc.PrivilegeLevelAdministrator))
	}
	if aux := resp[5] >> 4; aux != sessionAuxV20 {
		t.Errorf("auxiliary data = 0x%x, want IPMI v2.0 (0x%x)", aux, sessionAuxV20)
	}
	if ch := resp[5] & 0x0F; ch != lanChannelNumber {
		t.Errorf("channel = %d, want %d", ch, lanChannelNumber)
	}
}

func TestHandleGetSessionInfoCurrentV15(t *testing.T) {
	b := newTestBMC()
	user, err := b.Users.Add(3, "operator")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.Enabled = true

	sess, err := b.V15Sessions.CreatePending(bmc.V15AuthTypeMD5, user, [16]byte{}, lanChannelNumber)
	if err != nil {
		t.Fatalf("create v1.5 session: %v", err)
	}
	sess.SessionID = 0xAABBCCDD
	sess.State = bmc.V15SessionStateActive
	sess.PrivilegeLevel = bmc.PrivilegeLevelOperator

	hctx := &HandlerContext{BMC: b, V15Session: sess}

	resp, cc, err := handleGetSessionInfo(context.Background(), hctx, []byte{sessionIndexCurrent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeOK {
		t.Fatalf("completion code = 0x%02x, want OK", uint8(cc))
	}
	if len(resp) != 6 {
		t.Fatalf("response length = %d, want 6", len(resp))
	}
	if resp[0] != sess.Handle {
		t.Errorf("session handle = 0x%02x, want 0x%02x", resp[0], sess.Handle)
	}
	if resp[0] == 0 || resp[0] == 0xFF {
		t.Errorf("session handle = 0x%02x, want neither 0x00 (no session) nor reserved 0xFF", resp[0])
	}
	if resp[2] != 1 {
		t.Errorf("active sessions = %d, want 1", resp[2])
	}
	if resp[3] != 3 {
		t.Errorf("user id = %d, want 3", resp[3])
	}
	if resp[4] != uint8(bmc.PrivilegeLevelOperator) {
		t.Errorf("privilege level = 0x%02x, want 0x%02x", resp[4], uint8(bmc.PrivilegeLevelOperator))
	}
	if aux := resp[5] >> 4; aux != sessionAuxV15 {
		t.Errorf("auxiliary data = 0x%x, want IPMI v1.5 (0x%x)", aux, sessionAuxV15)
	}
	if ch := resp[5] & 0x0F; ch != lanChannelNumber {
		t.Errorf("channel = %d, want %d", ch, lanChannelNumber)
	}
}

func TestHandleGetSessionInfoRejectsUnsupportedForms(t *testing.T) {
	b := newTestBMC()
	sess := activeRMCPSession(t, b, 5, bmc.PrivilegeLevelAdministrator)
	hctx := &HandlerContext{BMC: b, Session: sess}

	for _, tc := range []struct {
		name string
		req  []byte
		want types.CompletionCode
	}{
		{"empty", nil, types.CodeRequestDataTruncated},
		{"by-index", []byte{0x02}, types.CodeRequestDataFieldInvalid},
		{"by-handle", []byte{sessionIndexByHandle, 0x01}, types.CodeRequestDataFieldInvalid},
		{"by-id", []byte{sessionIndexByID, 0x01, 0x02, 0x03, 0x04}, types.CodeRequestDataFieldInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, cc, err := handleGetSessionInfo(context.Background(), hctx, tc.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cc != tc.want {
				t.Fatalf("completion code = 0x%02x, want 0x%02x", uint8(cc), uint8(tc.want))
			}
		})
	}
}

func TestHandleGetSessionInfoNoCurrentSession(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	_, cc, err := handleGetSessionInfo(context.Background(), hctx, []byte{sessionIndexCurrent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeRequestDataFieldInvalid {
		t.Fatalf("completion code = 0x%02x, want field invalid", uint8(cc))
	}
}

func TestSessionHandlesUnique(t *testing.T) {
	// Handles are assigned at allocation and must be unique among live
	// sessions and never 0x00 ("no session") or the reserved 0xFF, so Get
	// Session Info is unambiguous about which session it describes.
	b := newTestBMC()
	seen := map[uint8]bool{}
	for i := 0; i < 4; i++ {
		sess, err := b.Sessions.Allocate(uint32(0x100+i), types.AuthAlg_None, types.IntegrityAlg_None, types.CryptAlg_None, bmc.PrivilegeLevelAdministrator, lanChannelNumber)
		if err != nil {
			t.Fatalf("allocate session %d: %v", i, err)
		}
		if sess.Handle == 0 || sess.Handle == 0xFF {
			t.Errorf("session %d: handle = 0x%02x, want neither 0x00 nor 0xFF", i, sess.Handle)
		}
		if seen[sess.Handle] {
			t.Errorf("session %d: handle 0x%02x already assigned to a live session", i, sess.Handle)
		}
		seen[sess.Handle] = true
	}
}
