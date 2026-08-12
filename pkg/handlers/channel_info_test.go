package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// wantChannelInfo asserts the common response invariants: the standard 9-byte
// length and the IPMI-forum IANA (7154 = 0xF2 0x1B 0x00, LS-first).
func wantChannelInfo(t *testing.T, resp []byte, cc types.CompletionCode) {
	t.Helper()
	if cc != types.CodeOK {
		t.Fatalf("completion code = 0x%02x, want OK", uint8(cc))
	}
	if len(resp) != 9 {
		t.Fatalf("response length = %d, want 9", len(resp))
	}
	if resp[4] != 0xF2 || resp[5] != 0x1B || resp[6] != 0x00 {
		t.Errorf("vendor IANA = % x, want f2 1b 00 (7154 LS-first)", resp[4:7])
	}
}

func TestHandleGetChannelInfoLAN(t *testing.T) {
	b := newTestBMC()
	// A concrete LAN channel arrival, so 0x0E must resolve to it.
	ch, _ := b.Channels.Get(lanChannelNumber)
	hctx := &HandlerContext{BMC: b, Channel: ch}

	resp, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{lanChannelNumber})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantChannelInfo(t, resp, cc)

	if resp[0] != lanChannelNumber {
		t.Errorf("channel number = %d, want %d", resp[0], lanChannelNumber)
	}
	if resp[1] != uint8(bmc.ChannelMediumLAN) {
		t.Errorf("medium = 0x%02x, want 0x%02x (LAN)", resp[1], uint8(bmc.ChannelMediumLAN))
	}
	if resp[2] != uint8(types.ChannelProtocolIPMB) {
		t.Errorf("protocol = 0x%02x, want 0x%02x (IPMB)", resp[2], uint8(types.ChannelProtocolIPMB))
	}
	if support := resp[3] >> 6; support != channelMultiSession {
		t.Errorf("session support = %d, want %d (multi-session)", support, channelMultiSession)
	}
	if count := resp[3] & 0x3F; count != 0 {
		t.Errorf("active session count = %d, want 0", count)
	}
}

func TestHandleGetChannelInfoCurrentChannel(t *testing.T) {
	b := newTestBMC()
	ch, _ := b.Channels.Get(lanChannelNumber)
	hctx := &HandlerContext{BMC: b, Channel: ch}

	// 0x0E ("this channel") must resolve to the channel the request arrived on.
	resp, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{types.ChannelNumberSelf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantChannelInfo(t, resp, cc)
	if resp[0] != lanChannelNumber {
		t.Errorf("resolved channel = %d, want %d", resp[0], lanChannelNumber)
	}
}

func TestHandleGetChannelInfoCurrentChannelNoContext(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b} // no arrival channel recorded

	// Without an arrival channel, 0x0E falls back to the LAN channel.
	resp, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{types.ChannelNumberSelf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantChannelInfo(t, resp, cc)
	if resp[0] != lanChannelNumber {
		t.Errorf("resolved channel = %d, want %d", resp[0], lanChannelNumber)
	}
}

func TestHandleGetChannelInfoSystemInterface(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	resp, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{0x0F})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantChannelInfo(t, resp, cc)
	if resp[0] != 0x0F {
		t.Errorf("channel number = %d, want 15", resp[0])
	}
	if resp[1] != uint8(bmc.ChannelMediumSystemIF) {
		t.Errorf("medium = 0x%02x, want 0x%02x (system interface)", resp[1], uint8(bmc.ChannelMediumSystemIF))
	}
	if resp[2] != uint8(types.ChannelProtocolKCS) {
		t.Errorf("protocol = 0x%02x, want 0x%02x (KCS)", resp[2], uint8(types.ChannelProtocolKCS))
	}
	if support := resp[3] >> 6; support != channelSessionLess {
		t.Errorf("session support = %d, want %d (session-less)", support, channelSessionLess)
	}
}

func TestHandleGetChannelInfoMediumFromStore(t *testing.T) {
	b := newTestBMC()
	// Reconfigure a channel's medium and confirm the handler reports it rather
	// than a hardcoded value.
	b.Channels.Set(&bmc.Channel{Number: 3, Medium: bmc.ChannelMediumSerial, AccessMode: bmc.ChannelAccessAlways})
	hctx := &HandlerContext{BMC: b}

	resp, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantChannelInfo(t, resp, cc)
	if resp[1] != uint8(bmc.ChannelMediumSerial) {
		t.Errorf("medium = 0x%02x, want 0x%02x (serial)", resp[1], uint8(bmc.ChannelMediumSerial))
	}
}

func TestHandleGetChannelInfoErrors(t *testing.T) {
	b := newTestBMC()
	hctx := &HandlerContext{BMC: b}

	t.Run("truncated", func(t *testing.T) {
		_, cc, err := handleGetChannelInfo(context.Background(), hctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cc != types.CodeRequestDataTruncated {
			t.Fatalf("completion code = 0x%02x, want truncated", uint8(cc))
		}
	})

	t.Run("unknown-channel", func(t *testing.T) {
		_, cc, err := handleGetChannelInfo(context.Background(), hctx, []byte{0x07})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cc != types.CodeRequestDataFieldInvalid {
			t.Fatalf("completion code = 0x%02x, want field invalid", uint8(cc))
		}
	})
}
