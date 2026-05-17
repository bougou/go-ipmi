package handlers

import (
	"context"
	"encoding/binary"
)

// IPMI NetFn codes used in this file.
const (
	NetFnAppRequest     uint8 = 0x06
	NetFnChassisRequest uint8 = 0x00
	NetFnStorageRequest uint8 = 0x0A
	NetFnSensorRequest  uint8 = 0x04
)

// IPMI App command IDs.
const (
	CmdGetDeviceID            uint8 = 0x01
	CmdColdReset              uint8 = 0x02
	CmdWarmReset              uint8 = 0x03
	CmdGetSelfTestResults     uint8 = 0x04
	CmdGetDeviceGUID          uint8 = 0x08
	CmdSetBMCGlobalEnables    uint8 = 0x2E
	CmdGetBMCGlobalEnables    uint8 = 0x2F
	CmdGetChannelAuthCaps     uint8 = 0x38
	CmdGetChannelCipherSuites uint8 = 0x54
)

// RegisterAppHandlers adds all App/Global command handlers to r.
func RegisterAppHandlers(r *Registry) {
	r.Register(NetFnAppRequest, CmdGetDeviceID, HandlerFunc(handleGetDeviceID))
	r.Register(NetFnAppRequest, CmdColdReset, HandlerFunc(handleColdReset))
	r.Register(NetFnAppRequest, CmdWarmReset, HandlerFunc(handleWarmReset))
	r.Register(NetFnAppRequest, CmdGetSelfTestResults, HandlerFunc(handleGetSelfTestResults))
	r.Register(NetFnAppRequest, CmdGetDeviceGUID, HandlerFunc(handleGetDeviceGUID))
}

// handleGetDeviceID implements Get Device ID (App 0x01).
// Response format follows Table 20-2 of the IPMI 2.0 spec.
func handleGetDeviceID(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	info := hctx.BMC.Info
	resp := make([]byte, 11)
	resp[0] = info.DeviceID
	resp[1] = info.DeviceRevision & 0x0F // bits 3:0 only; bit 7 = SDR present (set separately)
	resp[2] = info.FirmwareMajor & 0x7F  // bits 6:0; bit 7 = update in progress
	resp[3] = info.FirmwareMinor         // BCD
	resp[4] = info.IPMIVersion           // 0x20 for IPMI 2.0
	resp[5] = info.AdditionalDeviceSupport
	// Manufacturer ID: 3 bytes LS-first (bits 23:0)
	mid := info.ManufacturerID
	resp[6] = uint8(mid)
	resp[7] = uint8(mid >> 8)
	resp[8] = uint8(mid >> 16)
	// Product ID: 2 bytes LS-first
	binary.LittleEndian.PutUint16(resp[9:11], info.ProductID)
	return resp, CodeOK, nil
}

// handleColdReset implements Cold Reset (App 0x02).
func handleColdReset(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}
	if err := ch.ColdReset(ctx); err != nil {
		return nil, CodeUnspecifiedError, err
	}
	return nil, CodeOK, nil
}

// handleWarmReset implements Warm Reset (App 0x03).
func handleWarmReset(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}
	if err := ch.WarmReset(ctx); err != nil {
		return nil, CodeUnspecifiedError, err
	}
	return nil, CodeOK, nil
}

// handleGetSelfTestResults implements Get Self Test Results (App 0x04).
// Returns "No error" (0x55 0x00) as a static response; real implementations
// should perform an actual self-test and return the result.
func handleGetSelfTestResults(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	// 0x55 = "No error", 0x00 = test result byte (all tests passed)
	return []byte{0x55, 0x00}, CodeOK, nil
}

// handleGetDeviceGUID implements Get Device GUID (App 0x08).
// Returns the 16-byte GUID from the BMC config (stored LS-byte first per spec).
func handleGetDeviceGUID(_ context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	g := hctx.BMC.GUID
	resp := make([]byte, 16)
	copy(resp, g[:])
	return resp, CodeOK, nil
}
