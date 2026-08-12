package handlers

import (
	"context"
	"encoding/binary"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// CmdGetLanConfigParam is the Get LAN Configuration Parameters command byte,
// referenced by the privilege table. The Transport request NetFn is declared
// with the payload handlers.
const CmdGetLanConfigParam uint8 = 0x02

// LAN configuration parameter revision reported for every supported parameter.
// The high nibble is the "oldest revision supported" and the low nibble the
// "present revision"; 0x11 is the value real BMCs report (spec §23.2).
const lanParamRevision uint8 = 0x11

// standardPrimaryRMCPPort is the IANA-assigned RMCP port (623) used for IPMI
// over LAN. It is reported for param #8 unless the NetworkHAL advertises a
// non-zero port, which lets a BMC listening on a non-standard port make in-band
// software discover it.
const standardPrimaryRMCPPort uint16 = 623

// RegisterTransportHandlers adds all Transport (LAN) command handlers to r.
func RegisterTransportHandlers(r *Registry) {
	r.RegisterFunc(types.CommandGetLanConfigParam, handleGetLanConfigParam)
}

// handleGetLanConfigParam implements Get LAN Configuration Parameters
// (Transport 0x02, spec §23.2), the command in-band software and the metal
// agent use to discover the BMC's own network address.
//
// The request body is:
//
//	byte 1: [7] get parameter revision only, [3:0] channel number
//	byte 2: parameter selector
//	byte 3: set selector (block-based parameters only)
//	byte 4: block selector (block-based parameters only)
//
// The channel is validated to be a configured LAN channel (0x0E resolves to the
// arrival channel), but the reference BMC models a single NIC, so every LAN
// channel resolves to the same NetworkHAL configuration. The set/block
// selectors are not used by the parameters this handler serves. When bit 7 of
// the channel byte is set the caller wants only the parameter revision, so the
// data field is omitted (spec §23.2).
//
// The address-family parameters (IP, IP source, MAC, subnet, default gateway)
// are backed by [hal.NetworkHAL]; when Network() is nil the BMC has no NIC to
// describe and the command returns CannotExecuteCommandNotSupported. The set-in
// progress, authentication-type support and primary RMCP port parameters are
// static and answer without a NIC. Any other selector returns
// ParameterNotSupported (spec Table 23-4 permits a BMC to implement a subset).
//
// Only Get is implemented; Set LAN Configuration Parameters is out of scope.
func handleGetLanConfigParam(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	// The command's request is 4 bytes (channel, parameter selector, set
	// selector, block selector); the bundled client always packs all four.
	if len(req) < 4 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}

	revisionOnly := req[0]&0x80 != 0
	param := types.LanConfigParamSelector(req[1])

	// Validate the requested channel before anything else: it must be a
	// configured LAN channel, so channel 0x0F (system interface) or an unknown
	// channel does not return channel 1's NIC configuration.
	if !lanChannelValid(hctx, req[0]&0x0f) {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	// Resolve the parameter before honoring a revision-only query, so
	// revision-only for an unsupported selector still returns the
	// parameter-not-supported code rather than a spurious success. The data of
	// a revision-only query is discarded; the only cost is a NetworkHAL read.
	data, cc := lanParamData(ctx, hctx, param)
	if cc != types.CodeOK {
		return nil, cc, nil
	}
	if revisionOnly {
		return []byte{lanParamRevision}, types.CodeOK, nil
	}
	return lanParamResponse(data...), types.CodeOK, nil
}

// lanParamData returns the raw data bytes for one LAN configuration parameter.
// Its default arm is the single authority on which selectors are supported.
func lanParamData(ctx context.Context, hctx *HandlerContext, param types.LanConfigParamSelector) ([]byte, types.CompletionCode) {
	switch param {
	case types.LanConfigParamSelector_SetInProgress:
		// Report "set complete": the reference server has no in-progress LAN
		// configuration transaction.
		return []byte{byte(types.SetInProgress_SetComplete)}, types.CodeOK

	case types.LanConfigParamSelector_AuthTypeSupport:
		// Read-only bitmask (spec Table 23-4 param #1). Advertise the same v1.5
		// authentication types Get Channel Auth Capabilities reports, so the two
		// commands cannot contradict a v1.5-disabled or MD5-only deployment.
		return lanAuthTypeSupport(hctx.BMC).Pack(), types.CodeOK

	case types.LanConfigParamSelector_PrimaryRMCPPort:
		// 2-byte port, LS-first (spec Table 23-4 param #8). Reported from the
		// NetworkHAL configuration when it advertises a non-zero port, otherwise
		// the standard 623; this stays answerable without a NIC. A NetworkHAL read
		// error maps to the HAL completion code, like the address parameters,
		// rather than silently reporting 623.
		port := standardPrimaryRMCPPort

		if network := hctx.BMC.HAL().Network(); network != nil {
			cfg, err := network.GetConfig(ctx)
			if err != nil {
				return nil, codeFromHalErr(err)
			}

			if cfg.Port != 0 {
				port = cfg.Port
			}
		}

		out := make([]byte, 2)
		binary.LittleEndian.PutUint16(out, port)

		return out, types.CodeOK

	case types.LanConfigParamSelector_IP,
		types.LanConfigParamSelector_IPSource,
		types.LanConfigParamSelector_MAC,
		types.LanConfigParamSelector_SubnetMask,
		types.LanConfigParamSelector_DefaultGatewayIP:
		return lanAddressParamData(ctx, hctx, param)

	default:
		return nil, types.CodeParameterNotSupported
	}
}

// lanAddressParamData answers the address-family LAN parameters from the
// NetworkHAL configuration. The raw bytes it emits match what the client's
// GetLanConfigParamFor decoder expects (spec Table 23-4): 4 octets for the IPv4
// address, subnet mask and default gateway; 6 octets for the MAC; a single
// source byte for the IP address source.
func lanAddressParamData(ctx context.Context, hctx *HandlerContext, param types.LanConfigParamSelector) ([]byte, types.CompletionCode) {
	network := hctx.BMC.HAL().Network()
	if network == nil {
		return nil, types.CodeNotSupported
	}

	cfg, err := network.GetConfig(ctx)
	if err != nil {
		return nil, codeFromHalErr(err)
	}

	switch param {
	case types.LanConfigParamSelector_IP:
		return cfg.IP[:], types.CodeOK

	case types.LanConfigParamSelector_IPSource:
		source := types.IPAddressSourceStatic
		if cfg.DHCP {
			source = types.IPAddressSourceDHCP
		}
		return []byte{byte(source)}, types.CodeOK

	case types.LanConfigParamSelector_MAC:
		return cfg.MAC[:], types.CodeOK

	case types.LanConfigParamSelector_SubnetMask:
		return cfg.Mask[:], types.CodeOK

	case types.LanConfigParamSelector_DefaultGatewayIP:
		return cfg.Gateway[:], types.CodeOK

	default:
		// Unreachable: the caller only routes the parameters handled above.
		return nil, types.CodeParameterNotSupported
	}
}

// lanChannelValid resolves the request's channel nibble (0x0E means "this
// channel") and reports whether it names a configured LAN channel. Non-LAN and
// unknown channels are rejected so the command never describes the wrong NIC.
func lanChannelValid(hctx *HandlerContext, nibble uint8) bool {
	channel := nibble
	if channel == types.ChannelNumberSelf {
		if hctx.Channel != nil {
			channel = hctx.Channel.Number
		} else {
			channel = lanChannelNumber
		}
	}
	ch, err := hctx.BMC.Channels.Get(channel)
	if err != nil {
		return false
	}
	return ch.Medium == bmc.ChannelMediumLAN
}

// lanAuthTypeSupport derives the auth-type-support bitmask from the BMC's
// resolved v1.5 authentication types, the same source Get Channel Auth
// Capabilities uses. When v1.5 LAN is disabled no type is advertised.
func lanAuthTypeSupport(b *bmc.BMC) *types.LanConfigParam_AuthTypeSupport {
	support := &types.LanConfigParam_AuthTypeSupport{}
	if b == nil || !b.V15LANEnabled() {
		return support
	}
	for _, t := range b.ResolvedV15AuthTypes() {
		switch t {
		case bmc.V15AuthTypeNone:
			support.None = true
		case bmc.V15AuthTypeMD2:
			support.MD2 = true
		case bmc.V15AuthTypeMD5:
			support.MD5 = true
		case bmc.V15AuthTypePassword:
			support.Password = true
		case bmc.V15AuthTypeOEM:
			support.OEM = true
		}
	}
	return support
}

// lanParamResponse frames parameter data as a Get LAN Configuration Parameters
// response: the parameter revision byte followed by the parameter data.
func lanParamResponse(data ...byte) []byte {
	return append([]byte{lanParamRevision}, data...)
}
