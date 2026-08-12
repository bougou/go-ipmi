package server

// End-to-end Get Session Info test driven through the real pkg/client, proving
// the RMCP+ keepalive (GetCurrentSessionInfo) the client fires periodically
// succeeds against the reference server and reports the caller's own session.

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestGetCurrentSessionInfoRMCP(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	admin := adminClient(t, ctx, port)
	defer admin.Close(ctx) //nolint:errcheck

	resp, err := admin.GetCurrentSessionInfo(ctx)
	if err != nil {
		t.Fatalf("GetCurrentSessionInfo: %v", err)
	}

	if resp.SessionHandle == 0 {
		t.Errorf("session handle must be non-zero")
	}
	// raceNewBMC seeds the admin in slot 2, and it connects at administrator
	// privilege over the LAN channel (1) as an RMCP+ (v2.0) session.
	if resp.UserID != 2 {
		t.Errorf("user id = %d, want 2", resp.UserID)
	}
	if resp.OperatingPrivilegeLevel != types.PrivilegeLevel(bmc.PrivilegeLevelAdministrator) {
		t.Errorf("privilege level = %d, want administrator", resp.OperatingPrivilegeLevel)
	}
	if resp.AuxiliaryData != 1 {
		t.Errorf("auxiliary data = %d, want 1 (IPMI v2.0/RMCP+)", resp.AuxiliaryData)
	}
	if resp.ChannelNumber != 1 {
		t.Errorf("channel number = %d, want 1", resp.ChannelNumber)
	}
	if want := uint8(b.Sessions.Cap()); resp.PossibleActiveSessions != want {
		t.Errorf("possible active sessions = %d, want %d", resp.PossibleActiveSessions, want)
	}
	if resp.CurrentActiveSessions < 1 {
		t.Errorf("current active sessions = %d, want >= 1", resp.CurrentActiveSessions)
	}
}
