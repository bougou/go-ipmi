package server

// End-to-end Get Channel Info test driven through the real pkg/client, proving a
// remote console can read the LAN channel description from the reference server.

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestGetChannelInfoLAN(t *testing.T) {
	b := raceNewBMC(t)
	port, ctx, stop := raceStartServer(t, b)
	defer stop()

	admin := adminClient(t, ctx, port)
	defer admin.Close(ctx) //nolint:errcheck

	// 0x0E asks for "this channel", which the session arrived on: the LAN channel.
	resp, err := admin.GetChannelInfo(ctx, types.ChannelNumberSelf)
	if err != nil {
		t.Fatalf("GetChannelInfo: %v", err)
	}

	if resp.ActualChannelNumber != 1 {
		t.Errorf("channel number = %d, want 1", resp.ActualChannelNumber)
	}
	if resp.ChannelMedium != types.ChannelMediumLAN {
		t.Errorf("medium = %s, want 802.3 LAN", resp.ChannelMedium)
	}
	if resp.ChannelProtocol != types.ChannelProtocolIPMB {
		t.Errorf("protocol = %s, want IPMB", resp.ChannelProtocol)
	}
	if resp.SessionSupport != 0x02 {
		t.Errorf("session support = %d, want 2 (multi-session)", resp.SessionSupport)
	}
	if resp.VendorID != 7154 {
		t.Errorf("vendor IANA = %d, want 7154", resp.VendorID)
	}
}
