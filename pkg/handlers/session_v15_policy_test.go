package handlers

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

func TestV15AllowsAuthTypeNonePerMessageDisabled(t *testing.T) {
	ch := &bmc.Channel{Number: 1, PerMessageAuth: false, UserLevelAuth: true}
	sess := &bmc.V15Session{State: bmc.V15SessionStateActive, PrivilegeLevel: bmc.PrivilegeLevelAdministrator}

	if !V15AllowsAuthTypeNone(ch, NetFnChassisRequest, CmdChassisControl, sess) {
		t.Fatal("expected operator chassis control allowed without per-message auth")
	}
}

func TestV15AllowsAuthTypeNonePerMessageEnabled(t *testing.T) {
	ch := &bmc.Channel{Number: 1, PerMessageAuth: true, UserLevelAuth: true}
	sess := &bmc.V15Session{State: bmc.V15SessionStateActive, PrivilegeLevel: bmc.PrivilegeLevelAdministrator}

	if V15AllowsAuthTypeNone(ch, NetFnChassisRequest, CmdGetChassisStatus, sess) {
		t.Fatal("expected auth required when per-message auth enabled")
	}
}

func TestV15AllowsAuthTypeNoneUserLevelDisabled(t *testing.T) {
	ch := &bmc.Channel{Number: 1, PerMessageAuth: true, UserLevelAuth: false}
	sess := &bmc.V15Session{State: bmc.V15SessionStateActive, PrivilegeLevel: bmc.PrivilegeLevelUser}

	if !V15AllowsAuthTypeNone(ch, NetFnAppRequest, CmdGetDeviceID, sess) {
		t.Fatal("expected user-level get device id without auth when user-level auth disabled")
	}
	if V15AllowsAuthTypeNone(ch, NetFnChassisRequest, CmdChassisControl, sess) {
		t.Fatal("operator command should still require auth when only user-level auth disabled")
	}
}

func TestFillChannelAuthCapsByte4ReflectsChannel(t *testing.T) {
	b := newTestBMC()
	ch, _ := b.Channels.Get(lanChannelNumber)
	ch.PerMessageAuth = false
	ch.UserLevelAuth = false

	resp := make([]byte, 8)
	fillChannelAuthCapsByte4(resp, b, ch)

	if resp[2]&0x10 == 0 {
		t.Fatalf("expected per-message auth disabled bit, byte=0x%02x", resp[2])
	}
	if resp[2]&0x08 == 0 {
		t.Fatalf("expected user-level auth disabled bit, byte=0x%02x", resp[2])
	}
}
