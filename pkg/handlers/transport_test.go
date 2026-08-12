package handlers

import (
	"context"
	"testing"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/types"
)

// testIPConfig is a representative BMC NIC configuration used across the LAN
// parameter tests.
var testIPConfig = hal.IPConfig{
	IP:      [4]byte{192, 168, 1, 50},
	Mask:    [4]byte{255, 255, 255, 0},
	Gateway: [4]byte{192, 168, 1, 1},
	MAC:     [6]byte{0x52, 0x54, 0x00, 0x12, 0x34, 0x56},
	DHCP:    false,
}

// newTestBMCWithNetwork builds a BMC whose mock NetworkHAL reports cfg.
func newTestBMCWithNetwork(t *testing.T, cfg hal.IPConfig, opts ...bmc.Option) *bmc.BMC {
	t.Helper()
	m := mock.New()
	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, append([]bmc.Option{bmc.WithClock(clock.Real)}, opts...)...)
	if err := b.HAL().Network().SetConfig(context.Background(), &cfg); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	return b
}

// getLanParam dispatches Get LAN Configuration Parameters for one selector on
// channel 1 with zero set/block selectors, returning the raw response.
func getLanParam(t *testing.T, b *bmc.BMC, param types.LanConfigParamSelector) ([]byte, types.CompletionCode) {
	t.Helper()
	hctx := &HandlerContext{BMC: b}
	resp, cc, err := handleGetLanConfigParam(context.Background(), hctx, []byte{0x01, byte(param), 0x00, 0x00})
	if err != nil {
		t.Fatalf("handleGetLanConfigParam(%s): unexpected error: %v", param, err)
	}
	return resp, cc
}

// unpackParamData asserts a success response, checks the parameter revision, and
// feeds the parameter data through the given typed parameter's own decoder, the
// same path the real client takes. This proves the response bytes match what a
// console decodes.
func unpackParamData(t *testing.T, resp []byte, cc types.CompletionCode, param types.LanConfigParameter) {
	t.Helper()
	if cc != types.CodeOK {
		t.Fatalf("completion code = 0x%02x, want OK", uint8(cc))
	}
	if len(resp) < 1 {
		t.Fatalf("response too short: %d bytes", len(resp))
	}
	if resp[0] != lanParamRevision {
		t.Errorf("parameter revision = 0x%02x, want 0x%02x", resp[0], lanParamRevision)
	}
	if err := param.Unpack(resp[1:]); err != nil {
		t.Fatalf("decode parameter data % x: %v", resp[1:], err)
	}
}

func TestHandleGetLanConfigParamAddresses(t *testing.T) {
	b := newTestBMCWithNetwork(t, testIPConfig)

	t.Run("ip", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_IP)
		var p types.LanConfigParam_IP
		unpackParamData(t, resp, cc, &p)
		if got := p.IP.String(); got != "192.168.1.50" {
			t.Errorf("IP = %s, want 192.168.1.50", got)
		}
	})

	t.Run("subnet", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_SubnetMask)
		var p types.LanConfigParam_SubnetMask
		unpackParamData(t, resp, cc, &p)
		if got := p.SubnetMask.String(); got != "255.255.255.0" {
			t.Errorf("subnet = %s, want 255.255.255.0", got)
		}
	})

	t.Run("gateway", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_DefaultGatewayIP)
		var p types.LanConfigParam_DefaultGatewayIP
		unpackParamData(t, resp, cc, &p)
		if got := p.IP.String(); got != "192.168.1.1" {
			t.Errorf("gateway = %s, want 192.168.1.1", got)
		}
	})

	t.Run("mac", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_MAC)
		var p types.LanConfigParam_MAC
		unpackParamData(t, resp, cc, &p)
		if got := p.MAC.String(); got != "52:54:00:12:34:56" {
			t.Errorf("MAC = %s, want 52:54:00:12:34:56", got)
		}
	})
}

func TestHandleGetLanConfigParamIPSource(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		b := newTestBMCWithNetwork(t, testIPConfig)
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_IPSource)
		var p types.LanConfigParam_IPSource
		unpackParamData(t, resp, cc, &p)
		if p.Source != types.IPAddressSourceStatic {
			t.Errorf("source = %s, want static", p.Source)
		}
	})

	t.Run("dhcp", func(t *testing.T) {
		cfg := testIPConfig
		cfg.DHCP = true
		b := newTestBMCWithNetwork(t, cfg)
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_IPSource)
		var p types.LanConfigParam_IPSource
		unpackParamData(t, resp, cc, &p)
		if p.Source != types.IPAddressSourceDHCP {
			t.Errorf("source = %s, want DHCP", p.Source)
		}
	})
}

func TestHandleGetLanConfigParamStatic(t *testing.T) {
	b := newTestBMCWithNetwork(t, testIPConfig)

	t.Run("set-in-progress", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_SetInProgress)
		var p types.LanConfigParam_SetInProgress
		unpackParamData(t, resp, cc, &p)
		if p.Value != types.SetInProgress_SetComplete {
			t.Errorf("set in progress = %s, want set complete", p.Value)
		}
	})

	t.Run("auth-type-support", func(t *testing.T) {
		// Configure a known set of v1.5 auth types and assert the parameter
		// reflects exactly them, the same source Get Channel Auth Caps uses.
		bb := newTestBMCWithNetwork(t, testIPConfig, bmc.WithV15AuthTypes([]bmc.V15AuthType{bmc.V15AuthTypeMD5, bmc.V15AuthTypePassword}))
		resp, cc := getLanParam(t, bb, types.LanConfigParamSelector_AuthTypeSupport)
		var p types.LanConfigParam_AuthTypeSupport
		unpackParamData(t, resp, cc, &p)
		if !p.MD5 || !p.Password {
			t.Errorf("auth type support = %+v, want MD5 and Password advertised", p)
		}
		if p.MD2 || p.OEM || p.None {
			t.Errorf("auth type support = %+v, want only MD5 and Password", p)
		}
	})

	t.Run("auth-type-support-v15-disabled", func(t *testing.T) {
		// With v1.5 LAN disabled the parameter must not advertise any v1.5 auth
		// type, matching Get Channel Auth Caps rather than contradicting it.
		bb := newTestBMCWithNetwork(t, testIPConfig, bmc.WithV15Disabled())
		resp, cc := getLanParam(t, bb, types.LanConfigParamSelector_AuthTypeSupport)
		var p types.LanConfigParam_AuthTypeSupport
		unpackParamData(t, resp, cc, &p)
		if p.MD5 || p.Password || p.MD2 || p.OEM || p.None {
			t.Errorf("auth type support = %+v, want none advertised", p)
		}
	})

	t.Run("primary-rmcp-port", func(t *testing.T) {
		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_PrimaryRMCPPort)
		var p types.LanConfigParam_PrimaryRMCPPort
		unpackParamData(t, resp, cc, &p)
		if p.Port != standardPrimaryRMCPPort {
			t.Errorf("primary RMCP port = %d, want %d", p.Port, standardPrimaryRMCPPort)
		}
	})
}

func TestHandleGetLanConfigParamRevisionOnly(t *testing.T) {
	b := newTestBMCWithNetwork(t, testIPConfig)
	hctx := &HandlerContext{BMC: b}

	// Bit 7 of the channel byte requests the parameter revision only, so the
	// response carries just the revision byte and no parameter data.
	resp, cc, err := handleGetLanConfigParam(context.Background(), hctx, []byte{0x81, byte(types.LanConfigParamSelector_IP), 0x00, 0x00})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc != types.CodeOK {
		t.Fatalf("completion code = 0x%02x, want OK", uint8(cc))
	}
	if len(resp) != 1 || resp[0] != lanParamRevision {
		t.Errorf("response = % x, want just the revision byte 0x%02x", resp, lanParamRevision)
	}
}

func TestHandleGetLanConfigParamErrors(t *testing.T) {
	b := newTestBMCWithNetwork(t, testIPConfig)
	hctx := &HandlerContext{BMC: b}

	t.Run("truncated", func(t *testing.T) {
		// The request is 4 bytes; anything shorter (including the 2- and 3-byte
		// forms) is rejected rather than treating the missing selectors as zero.
		for _, req := range [][]byte{
			{0x01},
			{0x01, byte(types.LanConfigParamSelector_IP)},
			{0x01, byte(types.LanConfigParamSelector_IP), 0x00},
		} {
			_, cc, err := handleGetLanConfigParam(context.Background(), hctx, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cc != types.CodeRequestDataTruncated {
				t.Fatalf("%d-byte request: completion code = 0x%02x, want truncated", len(req), uint8(cc))
			}
		}
	})

	t.Run("unsupported-selector", func(t *testing.T) {
		// Community string (param 16) is not implemented by the reference server.
		_, cc := getLanParam(t, b, types.LanConfigParamSelector_CommunityString)
		if cc != types.CodeParameterNotSupported {
			t.Fatalf("completion code = 0x%02x, want parameter not supported", uint8(cc))
		}
	})

	t.Run("revision-only-unsupported-selector", func(t *testing.T) {
		// Revision-only must validate the selector first, so an unsupported one
		// returns parameter-not-supported instead of a spurious success.
		_, cc, err := handleGetLanConfigParam(context.Background(), hctx,
			[]byte{0x81, byte(types.LanConfigParamSelector_CommunityString), 0x00, 0x00})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cc != types.CodeParameterNotSupported {
			t.Fatalf("completion code = 0x%02x, want parameter not supported", uint8(cc))
		}
	})

	t.Run("non-lan-channel", func(t *testing.T) {
		// Channel 0x0F is the system interface, not a LAN channel, so it must not
		// return channel 1's NIC configuration.
		_, cc, err := handleGetLanConfigParam(context.Background(), hctx,
			[]byte{0x0F, byte(types.LanConfigParamSelector_IP), 0x00, 0x00})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cc != types.CodeRequestDataFieldInvalid {
			t.Fatalf("completion code = 0x%02x, want field invalid", uint8(cc))
		}
	})

	t.Run("unknown-channel", func(t *testing.T) {
		_, cc, err := handleGetLanConfigParam(context.Background(), hctx,
			[]byte{0x07, byte(types.LanConfigParamSelector_IP), 0x00, 0x00})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cc != types.CodeRequestDataFieldInvalid {
			t.Fatalf("completion code = 0x%02x, want field invalid", uint8(cc))
		}
	})
}

// noNetworkHAL wraps a HAL but reports no NIC, so address-family LAN parameters
// have nothing to describe.
type noNetworkHAL struct {
	hal.HAL
}

func (noNetworkHAL) Network() hal.NetworkHAL { return nil }

func TestHandleGetLanConfigParamNoNetwork(t *testing.T) {
	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, noNetworkHAL{mock.New()}, bmc.WithClock(clock.Real))
	hctx := &HandlerContext{BMC: b}

	// Address-family parameters require a NIC and must report the command as
	// unsupported when Network() is nil.
	for _, param := range []types.LanConfigParamSelector{
		types.LanConfigParamSelector_IP,
		types.LanConfigParamSelector_IPSource,
		types.LanConfigParamSelector_MAC,
		types.LanConfigParamSelector_SubnetMask,
		types.LanConfigParamSelector_DefaultGatewayIP,
	} {
		_, cc, err := handleGetLanConfigParam(context.Background(), hctx, []byte{0x01, byte(param), 0x00, 0x00})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", param, err)
		}
		if cc != types.CodeNotSupported {
			t.Errorf("%s: completion code = 0x%02x, want command not supported", param, uint8(cc))
		}
	}

	// The static parameters do not depend on a NIC and still answer.
	_, cc, err := handleGetLanConfigParam(context.Background(), hctx, []byte{0x01, byte(types.LanConfigParamSelector_PrimaryRMCPPort), 0x00, 0x00})
	if err != nil {
		t.Fatalf("primary RMCP port: unexpected error: %v", err)
	}
	if cc != types.CodeOK {
		t.Errorf("primary RMCP port without NIC: completion code = 0x%02x, want OK", uint8(cc))
	}
}

// TestHandleGetLanConfigParamPrimaryRMCPPort proves param #8 reports a
// non-standard RMCP port from the NetworkHAL configuration and falls back to
// 623 when none is configured. This is how an emulated BMC listening on an
// OS-assigned port makes in-band software discover it.
func TestHandleGetLanConfigParamPrimaryRMCPPort(t *testing.T) {
	t.Run("custom port from config", func(t *testing.T) {
		cfg := testIPConfig
		cfg.Port = 62345

		b := newTestBMCWithNetwork(t, cfg)

		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_PrimaryRMCPPort)

		var param types.LanConfigParam_PrimaryRMCPPort

		unpackParamData(t, resp, cc, &param)

		if param.Port != 62345 {
			t.Errorf("port = %d, want 62345", param.Port)
		}
	})

	t.Run("zero config port falls back to 623", func(t *testing.T) {
		b := newTestBMCWithNetwork(t, testIPConfig) // Port left zero

		resp, cc := getLanParam(t, b, types.LanConfigParamSelector_PrimaryRMCPPort)

		var param types.LanConfigParam_PrimaryRMCPPort

		unpackParamData(t, resp, cc, &param)

		if param.Port != 623 {
			t.Errorf("port = %d, want 623", param.Port)
		}
	})
}
