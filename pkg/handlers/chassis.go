package handlers

import (
	"context"
	"errors"

	"github.com/bougou/go-ipmi/pkg/cmd/chassis"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// IPMI Chassis command IDs (spec §28).
const (
	CmdGetChassisCapabilities uint8 = 0x00
	CmdGetChassisStatus       uint8 = 0x01
	CmdChassisControl         uint8 = 0x02
	CmdChassisIdentify        uint8 = 0x04
	CmdSetSystemBootOptions   uint8 = 0x08
	CmdGetSystemBootOptions   uint8 = 0x09
)

// RegisterChassisHandlers adds all Chassis command handlers to r.
func RegisterChassisHandlers(r *Registry) {
	r.Register(NetFnChassisRequest, CmdGetChassisCapabilities, HandlerFunc(handleGetChassisCapabilities))
	r.Register(NetFnChassisRequest, CmdGetChassisStatus, HandlerFunc(handleGetChassisStatus))
	r.Register(NetFnChassisRequest, CmdChassisControl, HandlerFunc(handleChassisControl))
	r.Register(NetFnChassisRequest, CmdChassisIdentify, HandlerFunc(handleChassisIdentify))
	r.Register(NetFnChassisRequest, CmdSetSystemBootOptions, HandlerFunc(handleSetSystemBootOptions))
	r.Register(NetFnChassisRequest, CmdGetSystemBootOptions, HandlerFunc(handleGetSystemBootOptions))
}

// handleGetChassisCapabilities returns a minimal static response.
// Real implementations may override this handler to report actual front-panel
// button and FRUS device addresses.
func handleGetChassisCapabilities(_ context.Context, _ *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	// Byte 0: capabilities flags (all disabled in reference impl)
	// Byte 1-4: FRU info device address, SDR device, SEL device, System management device
	return []byte{0x00, 0x20, 0x20, 0x20, 0x20}, CodeOK, nil
}

// handleGetChassisStatus implements Get Chassis Status (Chassis 0x01, spec §28.2).
// It builds a typed [chassis.GetChassisStatusResponse] from the HAL and
// serialises it with [chassis.GetChassisStatusResponse.Pack], eliminating
// hand-written byte assembly.
func handleGetChassisStatus(ctx context.Context, hctx *HandlerContext, _ []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()

	resp := &chassis.GetChassisStatusResponse{}
	if ch != nil {
		on, err := ch.PowerState(ctx)
		if err != nil {
			// Use codeFromErr, not codeFromHalErr: GetChassisStatus (§28.2
			// Table 28-3) does not define command-specific completion codes.
			// 80h (CodeBootParamNotSupported) is only valid for
			// Set/Get System Boot Options (§28.12/§28.13).
			return nil, codeFromErr(err), err
		}
		resp.PowerIsOn = on
		// IntrusionState is optional; absence (ErrNotSupported) leaves the
		// bit zero, matching the "no intrusion detected" wire value.
		if intruded, err := ch.IntrusionState(ctx); err == nil {
			resp.ChassisIntrusionActive = intruded
		}
		resp.ChassisIdentifySupported = true
	}
	return resp.Pack(), CodeOK, nil
}

// handleChassisControl implements Chassis Control (Chassis 0x02, spec §28.3).
// The request is decoded with [chassis.ChassisControlRequest.Unpack] and
// dispatched to the typed HAL method matching the spec Table 28-3 action.
// The action→HAL method mapping is a spec-fixed protocol dispatch; the
// upper-layer HAL implementation defines what each method means for the
// managed system.
func handleChassisControl(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}

	var typed chassis.ChassisControlRequest
	if err := typed.Unpack(req); err != nil {
		return nil, CodeRequestDataTruncated, nil
	}

	switch typed.ChassisControl {
	case chassis.ChassisControlPowerDown:
		return nil, codeFromErr(ch.SetPower(ctx, false)), nil
	case chassis.ChassisControlPowerUp:
		return nil, codeFromErr(ch.SetPower(ctx, true)), nil
	case chassis.ChassisControlPowerCycle:
		return nil, codeFromErr(ch.PowerCycle(ctx)), nil
	case chassis.ChassisControlHardReset:
		return nil, codeFromErr(ch.ColdReset(ctx)), nil
	case chassis.ChassisControlSoftShutdown:
		return nil, codeFromErr(ch.WarmReset(ctx)), nil
	case chassis.ChassisControlDiagnosticInterrupt:
		// No corresponding typed HAL method in the reference interface;
		// treat as an unsupported chassis control action per spec.
		return nil, CodeParamOutOfRange, nil
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

// handleSetSystemBootOptions implements Set System Boot Options (Chassis 0x08,
// spec §28.12 Table 28-12).  Supported parameter selectors:
//   - 0x04: Boot Info Acknowledge (§28.14 Table 28-14 param #4)
//   - 0x05: Boot Flags (§28.14 Table 28-14 param #5)
//
// The request format is:
//
//	byte 1: [7] parameter valid/invalid, [6:0] boot option parameter selector
//	byte 2:N: boot option parameter data (0 bytes allowed per spec)
//
// Per spec the BMC must return:
//   - 80h – parameter not supported
//   - 81h – attempt to set 'set in progress' when not in 'set complete' state
//   - 82h – attempt to write read-only parameter
func handleSetSystemBootOptions(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 1 {
		return nil, CodeRequestDataTruncated, nil
	}
	// bit 7 = parameter valid flag (§28.12 byte 1).
	_ = req[0] & 0x80 // accepted; HAL layer does not persist a lock state.
	paramSelector := types.BootOptionParamSelector(req[0] & 0x7f)
	paramData := req[1:]

	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}

	switch paramSelector {
	case types.BootOptionParamSelector_BootInfoAcknowledge:
		// 0 bytes of data → toggling the valid/lock bit only.
		if len(paramData) == 0 {
			return nil, CodeOK, nil
		}
		var ack types.BootOptionParam_BootInfoAcknowledge
		if err := ack.Unpack(paramData); err != nil {
			return nil, CodeRequestDataTruncated, nil
		}
		return nil, codeFromHalErr(ch.SetBootInfoAcknowledge(ctx, &ack)), nil

	case types.BootOptionParamSelector_BootFlags:
		if len(paramData) == 0 {
			return nil, CodeOK, nil
		}
		var flags types.BootOptionParam_BootFlags
		if err := flags.Unpack(paramData); err != nil {
			return nil, CodeRequestDataTruncated, nil
		}
		return nil, codeFromHalErr(ch.SetBootFlags(ctx, &flags)), nil

	default:
		// Per spec Table 28-12: unimplemented parameter → 80h.
		return nil, CodeBootParamNotSupported, nil
	}
}

// handleGetSystemBootOptions implements Get System Boot Options (Chassis 0x09,
// spec §28.13).  Supported parameter selectors:
//   - 0x04: Boot Info Acknowledge
//   - 0x05: Boot Flags
func handleGetSystemBootOptions(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if len(req) < 1 {
		return nil, CodeRequestDataTruncated, nil
	}
	paramSelector := types.BootOptionParamSelector(req[0] & 0x7f)

	ch := hctx.BMC.HAL().Chassis()
	if ch == nil {
		return nil, CodeNotSupportedInState, nil
	}

	switch paramSelector {
	case types.BootOptionParamSelector_BootInfoAcknowledge:
		ack, err := ch.GetBootInfoAcknowledge(ctx)
		if err != nil {
			return nil, codeFromHalErr(err), nil
		}
		resp := make([]byte, 0, 2+2)
		resp = append(resp, 0x01) // parameter version
		resp = append(resp, byte(paramSelector&0x7f))
		resp = append(resp, ack.Pack()...)
		return resp, CodeOK, nil

	case types.BootOptionParamSelector_BootFlags:
		flags, err := ch.GetBootFlags(ctx)
		if err != nil {
			return nil, codeFromHalErr(err), nil
		}
		resp := make([]byte, 0, 2+5)
		resp = append(resp, 0x01) // parameter version
		resp = append(resp, byte(paramSelector&0x7f))
		resp = append(resp, flags.Pack()...)
		return resp, CodeOK, nil

	default:
		return nil, CodeBootParamNotSupported, nil
	}
}

// codeFromErr maps a HAL error to a completion code.
// If err carries a [CompletionCode] (e.g. HAL returns [CodeNodeBusy] directly),
// that code is extracted; otherwise the error is mapped to [CodeUnspecifiedError].
func codeFromErr(err error) CompletionCode {
	if err == nil {
		return CodeOK
	}
	var cc CompletionCode
	if errors.As(err, &cc) {
		return cc
	}
	return CodeUnspecifiedError
}

// codeFromHalErr maps a HAL error to a completion code.
// It delegates to [codeFromErr] and additionally maps [hal.ErrNotSupported]
// to [CodeBootParamNotSupported].
func codeFromHalErr(err error) CompletionCode {
	if cc := codeFromErr(err); cc != CodeUnspecifiedError {
		return cc
	}
	if errors.Is(err, hal.ErrNotSupported) {
		return CodeBootParamNotSupported
	}
	return CodeUnspecifiedError
}
