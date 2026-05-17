package handlers

import (
	"context"
)

// IPMI Chassis command IDs.
const (
	CmdGetChassisCapabilities uint8 = 0x00
	CmdGetChassisStatus       uint8 = 0x01
	CmdChassisControl         uint8 = 0x02
	CmdChassisIdentify        uint8 = 0x04
)

// ChassisControl actions (Table 28-3).
const (
	ChassisControlPowerDown     uint8 = 0x00
	ChassisControlPowerUp       uint8 = 0x01
	ChassisControlPowerCycle    uint8 = 0x02
	ChassisControlHardReset     uint8 = 0x03
	ChassisControlDiagInterrupt uint8 = 0x04
	ChassisControlSoftShutdown  uint8 = 0x05
)

// RegisterChassisHandlers adds all Chassis command handlers to r.
func RegisterChassisHandlers(r *Registry) {
	r.Register(NetFnChassisRequest, CmdGetChassisCapabilities, HandlerFunc(handleGetChassisCapabilities))
	r.Register(NetFnChassisRequest, CmdGetChassisStatus, HandlerFunc(handleGetChassisStatus))
	r.Register(NetFnChassisRequest, CmdChassisControl, HandlerFunc(handleChassisControl))
	r.Register(NetFnChassisRequest, CmdChassisIdentify, HandlerFunc(handleChassisIdentify))
}

// handleGetChassisCapabilities returns a minimal static response.
// Real implementations may override this handler to report actual front-panel
// button and FRUS device addresses.
func handleGetChassisCapabilities(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	// Byte 0: capabilities flags (all disabled in reference impl)
	// Byte 1-4: FRU info device address, SDR device, SEL device, System management device
	return []byte{0x00, 0x20, 0x20, 0x20, 0x20}, CodeOK, nil
}

// handleGetChassisStatus implements Get Chassis Status (Chassis 0x01).
func handleGetChassisStatus(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()

	var powerOn bool
	if ch != nil {
		var err error
		powerOn, err = ch.PowerState(ctx)
		if err != nil {
			return nil, CodeUnspecifiedError, err
		}
	}

	byte0 := uint8(0x00)
	if powerOn {
		byte0 |= 0x01 // bit 0: power on
	}
	// Bytes 1-3: last power event, misc chassis state, FP button disables.
	// Return zeros (no specific event, no buttons disabled) for the reference impl.
	return []byte{byte0, 0x00, 0x00, 0x00}, CodeOK, nil
}

// handleChassisControl implements Chassis Control (Chassis 0x02).
func handleChassisControl(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 1 {
		return nil, CodeRequestDataTruncated, nil
	}
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}

	action := req[0] & 0x0F
	switch action {
	case ChassisControlPowerDown:
		return nil, codeFromErr(ch.SetPower(ctx, false)), nil
	case ChassisControlPowerUp:
		return nil, codeFromErr(ch.SetPower(ctx, true)), nil
	case ChassisControlHardReset:
		return nil, codeFromErr(ch.ColdReset(ctx)), nil
	case ChassisControlSoftShutdown:
		return nil, codeFromErr(ch.WarmReset(ctx)), nil
	default:
		return nil, CodeParamOutOfRange, nil
	}
}

// handleChassisIdentify implements Chassis Identify (Chassis 0x04).
func handleChassisIdentify(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}

	seconds := uint8(15) // default interval per spec
	if len(req) >= 1 {
		seconds = req[0]
	}
	// Byte 1 bit 0: Force Identify On – represented as math.MaxUint8 interval.
	if len(req) >= 2 && req[1]&0x01 != 0 {
		seconds = 0xFF
	}

	return nil, codeFromErr(ch.Identify(ctx, seconds)), nil
}

// codeFromErr maps a HAL error to a completion code.
func codeFromErr(err error) CompletionCode {
	if err == nil {
		return CodeOK
	}
	return CodeUnspecifiedError
}
