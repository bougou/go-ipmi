package handlers

import (
	"bytes"
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// setUsernameReq builds a Set User Name request body (userID + 16-byte name).
func setUsernameReq(userID uint8, name string) []byte {
	req := make([]byte, 1+bmc.MaxUserNameLen)
	req[0] = userID
	copy(req[1:], name)
	return req
}

// setPasswordReq builds a Set User Password request body.
func setPasswordReq(userID, operation uint8, stored20 bool, password string) []byte {
	b0 := userID & 0x3f
	plen := 16
	if stored20 {
		b0 |= 1 << 7
		plen = 20
	}
	req := []byte{b0, operation & 0x03}
	if operation == passwordOpSetPassword || operation == passwordOpTestPassword {
		pw := make([]byte, plen)
		copy(pw, password)
		req = append(req, pw...)
	}
	return req
}

func TestHandleGetUserAccess(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	// Populate slot 2 on channel 1 as an enabled operator with IPMI messaging.
	if _, err := b.Users.Add(2, "operator"); err != nil {
		t.Fatal(err)
	}
	if err := b.Users.Update(2, func(u *bmc.User) error {
		u.Enabled = true
		u.ChannelAccess[1] = bmc.UserChannelAccess{
			MaxPrivilege: bmc.PrivilegeLevelOperator,
			Enabled:      true,
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	resp, cc, err := handleGetUserAccess(ctx, hctx, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeOK {
		t.Fatalf("cc = 0x%02x, want OK", cc)
	}
	if len(resp) != 4 {
		t.Fatalf("resp len = %d, want 4", len(resp))
	}
	if resp[0] != bmc.MaxUsers {
		t.Errorf("max user count = %d, want %d", resp[0], bmc.MaxUsers)
	}
	if got := resp[1] >> 6; got != 0b01 {
		t.Errorf("enable status = 0b%02b, want 0b01 (enabled)", got)
	}
	// Slot 1 (anonymous) is enabled by default, plus the operator we added.
	if got := resp[1] & 0x3f; got != 2 {
		t.Errorf("enabled user count = %d, want 2", got)
	}
	// bit4 IPMI messaging + operator privilege (0x03).
	if resp[3]&(1<<4) == 0 {
		t.Errorf("IPMI messaging bit not set: 0x%02x", resp[3])
	}
	if got := resp[3] & 0x0f; got != uint8(bmc.PrivilegeLevelOperator) {
		t.Errorf("max priv = 0x%x, want 0x%x", got, uint8(bmc.PrivilegeLevelOperator))
	}
}

// TestHandleGetUserAccessEnumerable proves every slot 1..max answers OK, the
// property in-band enumerators depend on.
func TestHandleGetUserAccessEnumerable(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	for id := uint8(1); id <= b.Users.MaxUserCount(); id++ {
		_, cc, err := handleGetUserAccess(ctx, hctx, []byte{0x01, id})
		if err != nil || cc != types.CodeOK {
			t.Fatalf("slot %d: cc=0x%02x err=%v, want OK", id, cc, err)
		}
	}

	// Out-of-range slot must be rejected. At the default maximum of 63 the
	// user ID field's 6-bit mask makes a larger value unrepresentable, so the
	// upper bound is only testable with a lowered store maximum.
	b.Users = bmc.NewUserStore(bmc.WithMaxUsers(8))
	_, cc, _ := handleGetUserAccess(ctx, hctx, []byte{0x01, 9})
	if cc != types.CodeRequestDataFieldInvalid {
		t.Errorf("out-of-range slot cc = 0x%02x, want 0xCC", cc)
	}

	// Truncated request.
	_, cc, _ = handleGetUserAccess(ctx, hctx, []byte{0x01})
	if cc != types.CodeRequestDataTruncated {
		t.Errorf("truncated cc = 0x%02x, want 0xC6", cc)
	}
}

func TestHandleGetUsernameEmptySlotReturnsCC(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	// Slot 5 is unset -> exactly 0xCC.
	_, cc, _ := handleGetUsername(ctx, hctx, []byte{0x05})
	if cc != types.CodeRequestDataFieldInvalid {
		t.Fatalf("empty slot cc = 0x%02x, want 0xCC", cc)
	}

	// Set a name, then read it back null-padded.
	if _, cc, err := handleSetUsername(ctx, hctx, setUsernameReq(5, "alice")); err != nil || cc != types.CodeOK {
		t.Fatalf("set username: cc=0x%02x err=%v", cc, err)
	}
	resp, cc, err := handleGetUsername(ctx, hctx, []byte{0x05})
	if err != nil || cc != types.CodeOK {
		t.Fatalf("get username: cc=0x%02x err=%v", cc, err)
	}
	if len(resp) != bmc.MaxUserNameLen {
		t.Fatalf("name len = %d, want %d", len(resp), bmc.MaxUserNameLen)
	}
	if got := string(bytes.TrimRight(resp, "\x00")); got != "alice" {
		t.Errorf("name = %q, want %q", got, "alice")
	}
}

func TestHandleSetUserPassword(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	const secret = "s3cr3t"

	// Enable/disable/test on a missing slot must not create it and returns 0xCC.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpEnableUser, false, "")); cc != types.CodeRequestDataFieldInvalid {
		t.Errorf("enable missing slot cc = 0x%02x, want 0xCC", cc)
	}
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpTestPassword, false, secret)); cc != types.CodeRequestDataFieldInvalid {
		t.Errorf("test missing slot cc = 0x%02x, want 0xCC", cc)
	}
	if _, err := b.Users.Get(7); err == nil {
		t.Fatalf("slot 7 was created by a non-set operation")
	}

	// Set Password creates the slot (16-byte form).
	if _, cc, err := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpSetPassword, false, secret)); err != nil || cc != types.CodeOK {
		t.Fatalf("set password: cc=0x%02x err=%v", cc, err)
	}

	// Test Password matches with the size the password was stored with.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpTestPassword, false, secret)); cc != types.CodeOK {
		t.Errorf("test 16-byte cc = 0x%02x, want OK", cc)
	}
	// The size tag is part of the credential (spec §22.30): testing the same
	// secret with the other size fails with the dedicated wrong-size code.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpTestPassword, true, secret)); cc != types.CodeSetUserPasswordWrongSize {
		t.Errorf("test 20-byte cc = 0x%02x, want 0x81 (wrong size)", cc)
	}

	// Wrong password fails with the command-specific 0x80.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpTestPassword, false, "wrong")); cc != types.CodeSetUserPasswordDataMismatch {
		t.Errorf("test wrong cc = 0x%02x, want 0x80", cc)
	}

	// Enable then disable now works on the existing slot.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpEnableUser, false, "")); cc != types.CodeOK {
		t.Errorf("enable cc = 0x%02x, want OK", cc)
	}
	if u, _ := b.Users.Get(7); !u.Enabled {
		t.Errorf("user not enabled after enable op")
	}
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(7, passwordOpDisableUser, false, "")); cc != types.CodeOK {
		t.Errorf("disable cc = 0x%02x, want OK", cc)
	}
	if u, _ := b.Users.Get(7); u.Enabled {
		t.Errorf("user still enabled after disable op")
	}
}

func TestHandleSetUserAccess(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	// Change bit clear still applies the byte-3 privilege limit (spec §22.26 bit
	// 7 gates only the byte-1 access flags), so it creates/updates the slot with
	// the operator limit while leaving the access flags at their defaults.
	if _, cc, _ := handleSetUserAccess(ctx, hctx, []byte{0x01, 0x03, uint8(bmc.PrivilegeLevelOperator)}); cc != types.CodeOK {
		t.Errorf("change-bit-clear cc = 0x%02x, want OK", cc)
	}
	u, err := b.Users.Get(3)
	if err != nil {
		t.Fatalf("slot 3 not created by change-bit-clear privilege-limit set: %v", err)
	}
	if ca := u.ChannelAccess[1]; ca.MaxPrivilege != bmc.PrivilegeLevelOperator {
		t.Errorf("priv limit = 0x%x, want operator", uint8(ca.MaxPrivilege))
	} else if ca.CallbackOnly || ca.Enabled {
		t.Errorf("change-bit-clear altered access flags: %+v", ca)
	}

	// Change bit set on channel 1: callback-only + IPMI messaging + admin priv.
	flags := uint8(1<<7 | 1<<6 | 1<<4 | 0x01)
	if _, cc, _ := handleSetUserAccess(ctx, hctx, []byte{flags, 0x03, uint8(bmc.PrivilegeLevelAdministrator)}); cc != types.CodeOK {
		t.Fatalf("set access cc = 0x%02x, want OK", cc)
	}
	u, err = b.Users.Get(3)
	if err != nil {
		t.Fatalf("slot 3 not found: %v", err)
	}
	if ca := u.ChannelAccess[1]; !ca.CallbackOnly || !ca.Enabled || ca.MaxPrivilege != bmc.PrivilegeLevelAdministrator {
		t.Errorf("channel access = %+v, want callback+enabled+admin", ca)
	}

	// A later change-bit-clear updates only the privilege limit (read-modify-
	// write), preserving the byte-1 flags set above.
	if _, cc, _ := handleSetUserAccess(ctx, hctx, []byte{0x01, 0x03, uint8(bmc.PrivilegeLevelUser)}); cc != types.CodeOK {
		t.Fatalf("change-bit-clear update cc = 0x%02x, want OK", cc)
	}
	u, _ = b.Users.Get(3)
	if ca := u.ChannelAccess[1]; !ca.CallbackOnly || !ca.Enabled {
		t.Errorf("change-bit-clear cleared existing access flags: %+v", ca)
	} else if ca.MaxPrivilege != bmc.PrivilegeLevelUser {
		t.Errorf("priv limit = 0x%x, want user", uint8(ca.MaxPrivilege))
	}
}

// TestHandleSetUserAccessResolvesCurrentChannel proves channel nibble 0x0E
// ("this channel") resolves to the request's arrival channel for both Set and
// Get User Access, rather than being used as literal pseudo-channel 14.
func TestHandleSetUserAccessResolvesCurrentChannel(t *testing.T) {
	b := newTestBMC()
	ch, _ := b.Channels.Get(lanChannelNumber)
	hctx := &HandlerContext{BMC: b, Channel: ch}
	ctx := context.Background()

	flags := uint8(1<<7 | 1<<4 | 0x0e) // change bit + IPMI messaging + channel 0x0E
	if _, cc, _ := handleSetUserAccess(ctx, hctx, []byte{flags, 0x04, uint8(bmc.PrivilegeLevelOperator)}); cc != types.CodeOK {
		t.Fatalf("set access cc = 0x%02x, want OK", cc)
	}
	u, _ := b.Users.Get(4)
	if _, ok := u.ChannelAccess[0x0e]; ok {
		t.Errorf("access wrongly stored under pseudo-channel 14")
	}
	if ca, ok := u.ChannelAccess[lanChannelNumber]; !ok || !ca.Enabled || ca.MaxPrivilege != bmc.PrivilegeLevelOperator {
		t.Errorf("channel %d access = %+v ok=%v, want enabled operator", lanChannelNumber, ca, ok)
	}

	// Get User Access for 0x0E must read the same resolved channel back.
	resp, cc, _ := handleGetUserAccess(ctx, hctx, []byte{0x0e, 0x04})
	if cc != types.CodeOK {
		t.Fatalf("get access cc = 0x%02x, want OK", cc)
	}
	if resp[3]&(1<<4) == 0 || resp[3]&0x0f != uint8(bmc.PrivilegeLevelOperator) {
		t.Errorf("get access byte 3 = 0x%02x, want IPMI messaging + operator for resolved channel", resp[3])
	}
}

// TestHandleSetUsernameRejectsDuplicate proves Set User Name refuses to give a
// second slot a name already held by another, keeping name-based session lookup
// deterministic.
func TestHandleSetUsernameRejectsDuplicate(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	if _, cc, err := handleSetUsername(ctx, hctx, setUsernameReq(3, "alice")); err != nil || cc != types.CodeOK {
		t.Fatalf("set name slot 3: cc=0x%02x err=%v", cc, err)
	}
	if _, cc, _ := handleSetUsername(ctx, hctx, setUsernameReq(4, "alice")); cc != types.CodeRequestDataFieldInvalid {
		t.Errorf("duplicate name cc = 0x%02x, want 0xCC", cc)
	}
	if _, err := b.Users.Get(4); err == nil {
		t.Errorf("slot 4 created despite duplicate-name rejection")
	}
	// Re-setting the same name on the same slot stays allowed (idempotent).
	if _, cc, _ := handleSetUsername(ctx, hctx, setUsernameReq(3, "alice")); cc != types.CodeOK {
		t.Errorf("idempotent rename cc = 0x%02x, want OK", cc)
	}
}

// TestHandleSetUserPassword20ByteNoPrefixMatch proves a 16-byte Test cannot
// match a genuine 20-byte password on its first 16 bytes.
func TestHandleSetUserPassword20ByteNoPrefixMatch(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}
	ctx := context.Background()

	const full20 = "abcdefghijklmnopqrst" // exactly 20 bytes
	if _, cc, err := handleSetUserPassword(ctx, hctx, setPasswordReq(8, passwordOpSetPassword, true, full20)); err != nil || cc != types.CodeOK {
		t.Fatalf("set 20-byte password: cc=0x%02x err=%v", cc, err)
	}
	// A 16-byte Test carrying only the first 16 bytes must fail: the size tag
	// differs before any byte comparison happens.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(8, passwordOpTestPassword, false, full20[:16])); cc != types.CodeSetUserPasswordWrongSize {
		t.Errorf("16-byte prefix test cc = 0x%02x, want 0x81 (wrong size)", cc)
	}
	// The genuine full 20-byte secret still matches.
	if _, cc, _ := handleSetUserPassword(ctx, hctx, setPasswordReq(8, passwordOpTestPassword, true, full20)); cc != types.CodeOK {
		t.Errorf("full 20-byte test cc = 0x%02x, want OK", cc)
	}
}
