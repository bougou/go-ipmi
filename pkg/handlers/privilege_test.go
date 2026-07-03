package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
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
	if cc != CodeInsufficientPrivilege {
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
	if cc != CodeParamOutOfRange {
		t.Fatalf("want param out of range for privilege 0, got %02x", cc)
	}
}
