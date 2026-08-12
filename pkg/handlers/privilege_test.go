package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestChassisControlRequiresOperatorPrivilege(t *testing.T) {
	b := newTestBMC()
	reg := NewRegistry()
	RegisterChassisHandlers(reg)

	sess := &bmc.V15Session{
		State:          bmc.V15SessionStateActive,
		PrivilegeLevel: bmc.PrivilegeLevelUser,
		MaxPrivilege:   bmc.PrivilegeLevelAdministrator,
	}
	hctx := &HandlerContext{BMC: b, V15Session: sess}

	_, cc, err := reg.Dispatch(context.Background(), hctx, NetFnChassisRequest, CmdChassisControl, []byte{0x00})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if cc != types.CodeInsufficientPrivilege {
		t.Fatalf("want insufficient privilege, got %02x", cc)
	}
}

func TestActivateSessionRejectsReservedPrivilegeZero(t *testing.T) {
	b := newTestBMC()
	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		t.Fatalf("add user: %v", err)
	}
	user.Enabled = true
	user.ChannelAccess[lanChannelNumber] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}

	var challenge [16]byte
	sess, err := b.V15Sessions.CreatePending(bmc.V15AuthTypeMD5, user, challenge, lanChannelNumber)
	if err != nil {
		t.Fatalf("CreatePending: %v", err)
	}

	req := make([]byte, 22)
	req[0] = uint8(bmc.V15AuthTypeMD5)
	req[1] = 0 // reserved
	copy(req[2:18], challenge[:])

	hctx := &HandlerContext{BMC: b, V15Session: sess, User: user}
	_, cc, err := handleActivateSession(context.Background(), hctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeParameterOutOfRange {
		t.Fatalf("want param out of range for privilege 0, got %02x", cc)
	}
}

// TestPreSessionRejectsPrivilegedCommands proves an unauthenticated LAN packet
// (no session) cannot invoke a privileged command such as Set User Password.
// Without the gate a remote attacker could create an administrator account and
// then log in normally. It drives the real registry so the check is exercised
// through the same wrapper every dispatch path uses, and asserts no user slot
// was created as a side effect.
func TestPreSessionRejectsPrivilegedCommands(t *testing.T) {
	b := newTestBMC()
	reg := NewRegistry()
	RegisterAllHandlers(reg)
	ctx := context.Background()

	setPassword := make([]byte, 18)
	setPassword[0] = 4    // user slot
	setPassword[1] = 0x02 // set password
	copy(setPassword[2:], "attacker")

	// Session-less (pre-session LAN): rejected, and the user store is untouched.
	_, cc, _ := reg.Dispatch(ctx, &HandlerContext{BMC: b}, NetFnAppRequest, CmdSetUserPassword, setPassword)
	if cc != types.CodeInsufficientPrivilege {
		t.Errorf("pre-session Set User Password cc = 0x%02x, want insufficient privilege", uint8(cc))
	}
	if _, err := b.Users.Get(4); err == nil {
		t.Error("pre-session Set User Password created a user slot")
	}

	// The pre-session exempt commands stay reachable without a session.
	_, cc, _ = reg.Dispatch(ctx, &HandlerContext{BMC: b}, NetFnAppRequest, CmdGetChannelAuthCapabilities, []byte{0x0e, 0x04})
	if cc != types.CodeOK {
		t.Errorf("pre-session Get Channel Auth Caps cc = 0x%02x, want OK", uint8(cc))
	}

	// An authenticated Administrator session runs the command normally.
	admin := &HandlerContext{BMC: b, Session: &bmc.Session{PrivilegeLevel: bmc.PrivilegeLevelAdministrator}}
	if _, cc, _ := reg.Dispatch(ctx, admin, NetFnAppRequest, CmdSetUserPassword, setPassword); cc != types.CodeOK {
		t.Errorf("admin-session Set User Password cc = 0x%02x, want OK", uint8(cc))
	}
	if _, err := b.Users.Get(4); err != nil {
		t.Errorf("admin-session Set User Password did not create the slot: %v", err)
	}
}
