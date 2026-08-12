package server

// End-to-end Get LAN Configuration Parameters test driven through the real
// pkg/client, proving a remote console (and the metal agent) can read the BMC's
// own IPv4 address and RMCP port from the reference server.

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

func TestGetLanConfigParamOverClient(t *testing.T) {
	b := raceNewBMC(t)
	if err := b.HAL().Network().SetConfig(context.Background(), &hal.IPConfig{
		IP:      [4]byte{192, 168, 1, 50},
		Mask:    [4]byte{255, 255, 255, 0},
		Gateway: [4]byte{192, 168, 1, 1},
		MAC:     [6]byte{0x52, 0x54, 0x00, 0x12, 0x34, 0x56},
	}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	admin := adminClient(t, ctx, port)
	defer admin.Close(ctx) //nolint:errcheck

	// Channel 1 is the LAN channel the admin session arrived on.
	const lanChannel = 1

	// Param 3: the agent's hard requirement, the BMC's own IPv4 address.
	ip := &types.LanConfigParam_IP{}
	if err := admin.GetLanConfigParamFor(ctx, lanChannel, ip); err != nil {
		t.Fatalf("GetLanConfigParamFor(IP): %v", err)
	}
	if got := ip.IP.String(); got != "192.168.1.50" {
		t.Errorf("IP = %s, want 192.168.1.50", got)
	}

	// Param 8: the primary RMCP port, reported as the standard 623.
	rmcp := &types.LanConfigParam_PrimaryRMCPPort{}
	if err := admin.GetLanConfigParamFor(ctx, lanChannel, rmcp); err != nil {
		t.Fatalf("GetLanConfigParamFor(PrimaryRMCPPort): %v", err)
	}
	if rmcp.Port != 623 {
		t.Errorf("primary RMCP port = %d, want 623", rmcp.Port)
	}
}
