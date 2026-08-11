package handlers

import (
	"context"
	"encoding/binary"

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
