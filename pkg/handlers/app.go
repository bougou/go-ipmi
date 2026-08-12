package handlers

import (
	"context"
	"encoding/binary"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// NetFn and command codes for the requests this package inspects as raw wire
// bytes, before dispatch has resolved them to a [types.Command] (privilege
// rules, v1.5 authentication policy, session state machine). Registration uses
// the [types] command table directly and needs none of these.
//
// TestWireConstantsMatchCommandTable keeps the values in step with that table.
const (
	NetFnAppRequest     uint8 = 0x06
	NetFnChassisRequest uint8 = 0x00
)

// IPMI App command IDs.
const (
	CmdColdReset              uint8 = 0x02
	CmdWarmReset              uint8 = 0x03
	CmdGetChannelCipherSuites uint8 = 0x54
)

// RegisterAppHandlers adds all App/Global command handlers to r.
func RegisterAppHandlers(r *Registry) {
	r.RegisterFunc(types.CommandGetDeviceID, handleGetDeviceID)
	r.RegisterFunc(types.CommandColdReset, handleColdReset)
	r.RegisterFunc(types.CommandWarmReset, handleWarmReset)
	r.RegisterFunc(types.CommandGetSelfTestResults, handleGetSelfTestResults)
	r.RegisterFunc(types.CommandGetDeviceGUID, handleGetDeviceGUID)
	r.RegisterFunc(types.CommandGetChannelInfo, handleGetChannelInfo)
}

// handleGetDeviceID implements Get Device ID (App 0x01).
// Response format follows Table 20-2 of the IPMI 2.0 spec.
func handleGetDeviceID(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
	info := hctx.BMC.Info
	deviceRev := info.DeviceRevision & 0x0F
	additional := info.AdditionalDeviceSupport
	if store := storageHAL(hctx); store != nil {
		if hasSDRRecords(ctx, store) {
			additional |= 0x02 // bit 1: SDR Repository Device (Table 20-2)
		}
		if hasFRUDevice(ctx, store, 0) {
			additional |= 0x08 // bit 3: FRU Inventory Device (Table 20-2)
		}
	}
	resp := make([]byte, 11)
	resp[0] = info.DeviceID
	resp[1] = deviceRev
	resp[2] = info.FirmwareMajor & 0x7F // bits 6:0; bit 7 = update in progress
	resp[3] = info.FirmwareMinor        // BCD
	resp[4] = info.IPMIVersion          // 0x20 for IPMI 2.0
	resp[5] = additional
	// Manufacturer ID: 3 bytes LS-first (bits 23:0)
	mid := info.ManufacturerID
	resp[6] = uint8(mid)
	resp[7] = uint8(mid >> 8)
	resp[8] = uint8(mid >> 16)
	// Product ID: 2 bytes LS-first
	binary.LittleEndian.PutUint16(resp[9:11], info.ProductID)
	return resp, types.CodeOK, nil
}

// handleColdReset implements Cold Reset (App 0x02).
func handleColdReset(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, types.CodeNotSupported, nil
	}
	if err := ch.ColdReset(ctx); err != nil {
		return nil, types.CodeUnspecifiedError, err
	}
	// A BMC cold reset aborts any SOL parameter set in progress: the
	// volatile SOL configuration (#0 set in progress, #6 bit rate) returns
	// to its power-up state (Table 26-3, Table 26-5).
	hctx.BMC.SOL.Config().ResetVolatile()
	return nil, types.CodeOK, nil
}

// handleWarmReset implements Warm Reset (App 0x03).
func handleWarmReset(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, types.CodeNotSupported, nil
	}
	if err := ch.WarmReset(ctx); err != nil {
		return nil, types.CodeUnspecifiedError, err
	}
	return nil, types.CodeOK, nil
}

// handleGetSelfTestResults implements Get Self Test Results (App 0x04).
// Returns "No error" (0x55 0x00) as a static response; real implementations
// should perform an actual self-test and return the result.
func handleGetSelfTestResults(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
	// 0x55 = "No error", 0x00 = test result byte (all tests passed)
	return []byte{0x55, 0x00}, types.CodeOK, nil
}

// handleGetDeviceGUID implements Get Device GUID (App 0x08).
// Returns the 16-byte GUID from the BMC config (stored LS-byte first per spec).
func handleGetDeviceGUID(_ context.Context, hctx *HandlerContext, _ []byte) ([]byte, types.CompletionCode, error) {
	g := hctx.BMC.GUID
	resp := make([]byte, 16)
	copy(resp, g[:])
	return resp, types.CodeOK, nil
}

// ---------------------------------------------------------------------------
// Get Channel Info
// ---------------------------------------------------------------------------

// Channel session-support encodings (response byte 4, bits [7:6]) per IPMI spec
// Table 22-30. The full set is session-less (0), single-session (1),
// multi-session (2) and session-based (3); the reference server only reports the
// first and third.
const (
	channelSessionLess  uint8 = 0x00 // no sessions (e.g. the system interface)
	channelMultiSession uint8 = 0x02 // multiple concurrent sessions (LAN)
)

// ipmiForumIANA is the IPMI-forum IANA enterprise number reported as the channel
// protocol vendor for the standard IPMB protocol. It is sent LS-byte first as a
// 3-byte field (0xF2 0x1B 0x00).
const ipmiForumIANA uint32 = 7154

// handleGetChannelInfo implements Get Channel Info (App 0x42).
//
// The response is built from the BMC's [bmc.ChannelStore] rather than from
// hardcoded media or numbers, so it stays truthful as channels are reconfigured.
// The channel-number request field may be 0x0E ("this channel"), which resolves
// to the channel the request arrived on. An unknown channel returns
// CodeRequestDataFieldInvalid. Response format follows Table 22-30 of the IPMI
// 2.0 spec.
//
// Two fields are deliberate simplifications: the per-channel active-session
// count is reported as 0 (session occupancy is deferred to Get Session Info),
// and the protocol and session-support fields are derived from the channel
// medium. That derivation is faithful for the modeled LAN and system-interface
// channels but only approximate for other media a caller might configure.
func handleGetChannelInfo(_ context.Context, hctx *HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
	if len(req) < 1 {
		return nil, types.CodeRequestDataTruncated, nil
	}
	if hctx == nil || hctx.BMC == nil {
		return nil, types.CodeNotSupported, nil
	}

	chNum := req[0] & 0x0F
	if chNum == types.ChannelNumberSelf {
		// 0x0E means "the channel this request was received on". The server
		// records that channel on the context; fall back to the LAN channel when
		// it is absent (e.g. requests over the system interface in a bare setup).
		if hctx.Channel != nil {
			chNum = hctx.Channel.Number
		} else {
			chNum = lanChannelNumber
		}
	}

	ch, err := hctx.BMC.Channels.Get(chNum)
	if err != nil {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}

	resp := make([]byte, 9)
	resp[0] = ch.Number
	resp[1] = uint8(ch.Medium)
	resp[2] = uint8(channelProtocolForMedium(ch.Medium))
	// Byte 4: bits [7:6] session support, bits [5:0] active session count. All
	// sessions live on the LAN channel (the only session-based channel this
	// server models), so its count is both tables' occupancy; other channels
	// are session-less and report zero. The RMCP+ table's occupancy stands in
	// for its active count for the reason given in Get Session Info.
	sessions := 0
	if ch.Number == lanChannelNumber {
		sessions = hctx.BMC.Sessions.Count() + hctx.BMC.V15Sessions.CountActiveSessions()
	}
	resp[3] = channelSessionSupportForMedium(ch.Medium)<<6 | clampSessionCount(sessions)
	// Bytes 5:7: channel protocol vendor ID (IANA), LS-first.
	resp[4] = uint8(ipmiForumIANA & 0xFF)
	resp[5] = uint8((ipmiForumIANA >> 8) & 0xFF)
	resp[6] = uint8((ipmiForumIANA >> 16) & 0xFF)
	// Bytes 8:9: auxiliary channel info. For the system interface these carry
	// the SMS and event-message-buffer interrupt types, where 0x00 means IRQ 0;
	// 0xFF is the defined "no interrupt / unspecified" value (Table 22-24). For
	// a LAN channel the bytes are unused and zero.
	if ch.Medium == bmc.ChannelMediumSystemIF {
		resp[7] = 0xFF
		resp[8] = 0xFF
	}
	return resp, types.CodeOK, nil
}

// channelProtocolForMedium maps a channel medium to the message protocol it
// speaks. LAN and IPMB-based channels carry IPMB-formatted messages; the system
// interface uses KCS.
func channelProtocolForMedium(m bmc.ChannelMedium) types.ChannelProtocol {
	if m == bmc.ChannelMediumSystemIF {
		return types.ChannelProtocolKCS
	}
	return types.ChannelProtocolIPMB
}

// channelSessionSupportForMedium reports whether a channel is session-based. LAN
// channels support multiple concurrent sessions; other media (notably the
// system interface) are session-less.
func channelSessionSupportForMedium(m bmc.ChannelMedium) uint8 {
	if m == bmc.ChannelMediumLAN {
		return channelMultiSession
	}
	return channelSessionLess
}
