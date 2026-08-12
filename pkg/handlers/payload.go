package handlers

// Payload management command handlers (spec v2.0 §24, "RMCP+ Support and
// Payload Commands") and SOL configuration handlers (§26).
//
// Only the standard SOL payload type (01h, Table 13-16) is implemented; the
// payload instance capacity is 1 (Table 24-6), matching the single shared
// serial port of the reference hardware model (§15.3).

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/hal"
	"github.com/bougou/go-ipmi/pkg/types"
)

// NetFnTransportRequest is the request NetFn for Transport commands (§23).
const NetFnTransportRequest uint8 = 0x0c

// payloadTypeSOL is the SOL payload type number (Table 13-16).
const payloadTypeSOL = 0x01

// solFormatVersion is the BCD major.minor of the implemented SOL payload
// format (Table 24-11): "The Format Version for the SOL payload implemented
// per this specification is 1.0 (10h)".
const solFormatVersion = 0x10

// RegisterPayloadHandlers adds the payload management commands (§24) to r.
func RegisterPayloadHandlers(r *Registry) {
	r.RegisterFunc(types.CommandActivatePayload, handleActivatePayload)
	r.RegisterFunc(types.CommandDeactivatePayload, handleDeactivatePayload)
	r.RegisterFunc(types.CommandGetPayloadActivationStatus, handleGetPayloadActivationStatus)
	r.RegisterFunc(types.CommandGetPayloadInstanceInfo, handleGetPayloadInstanceInfo)
	r.RegisterFunc(types.CommandSetUserPayloadAccess, handleSetUserPayloadAccess)
	r.RegisterFunc(types.CommandGetUserPayloadAccess, handleGetUserPayloadAccess)
	r.RegisterFunc(types.CommandGetChannelPayloadSupport, handleGetChannelPayloadSupport)
	r.RegisterFunc(types.CommandGetChannelPayloadVersion, handleGetChannelPayloadVersion)
	r.RegisterFunc(types.CommandSuspendResumePayloadEncryption, handleSuspendResumePayloadEncryption)
}

// RegisterSOLHandlers adds the SOL configuration commands (§26) to r.
//
// SOL Activating (§26.1) is intentionally absent: it is a BMC→console
// notification only meaningful with serial port sharing, which this BMC does
// not implement (§26 Table 26-1 footnote 2).
func RegisterSOLHandlers(r *Registry) {
	r.RegisterFunc(types.CommandSetSOLConfigParam, handleSetSOLConfigParam)
	r.RegisterFunc(types.CommandGetSOLConfigParam, handleGetSOLConfigParam)
}

// solCommandCC maps SOL store errors to the command-specific completion
// codes of the payload commands (named constants in pkg/types,
// Tables 24-2 / 24-3 / 24-5).
func solCommandCC(err error) types.CompletionCode {
	switch {
	case err == nil:
		return types.CodeOK
	case errors.Is(err, bmc.ErrSOLAlreadyActive):
		return types.CodeActivatePayloadAlreadyActive
	case errors.Is(err, bmc.ErrSOLNotActive):
		return types.CodeDeactivatePayloadAlreadyDeactivated
	case errors.Is(err, bmc.ErrSOLDisabled), errors.Is(err, hal.ErrNotSupported):
		// No redirectable console on the target reads as "payload type is
		// disabled" to the remote console.
		return types.CodeActivatePayloadTypeDisabled
	case errors.Is(err, bmc.ErrSOLEncryptionUnavailable):
		return types.CodeActivatePayloadCannotActivateWithEncryption
	case errors.Is(err, bmc.ErrSOLEncryptionRequired):
		return types.CodeActivatePayloadCannotActivateWithoutEncryption
	case errors.Is(err, bmc.ErrSOLAuthenticationUnavailable):
		return types.CodeRequestDataFieldInvalid
	case errors.Is(err, bmc.ErrSOLOperationUnsupported):
		return types.CodeSuspendResumePayloadEncryptionNotSupported
	case errors.Is(err, bmc.ErrSOLEncryptionForced):
		return types.CodeSuspendResumePayloadEncryptionNotAllowed
	case errors.Is(err, bmc.ErrSOLEncryptionUnavailableForSession):
		return types.CodeSuspendResumePayloadEncryptionNotAvailable
	case errors.Is(err, bmc.ErrSOLInstanceNotActive):
		return types.CodeSuspendResumePayloadEncryptionNotActive
	case errors.Is(err, bmc.ErrSOLPrivilege), errors.Is(err, bmc.ErrSOLNotOwner):
		return types.CodeInsufficientPrivilege
	default:
		return codeFromErr(err)
	}
}

// resolveChannel maps the request channel nibble to an effective channel
// number: 0Eh selects "the channel this request was issued over"
// (spec v2.0 §6.6 note on channel numbering).
func resolveChannel(hctx *HandlerContext, reqChannel uint8) uint8 {
	if reqChannel == 0x0e && hctx.Session != nil {
		return hctx.Session.Channel
	}
	return reqChannel
}

// parseSOLSelector validates the payload type / instance pair carried by the
// payload activation commands. ok=false means the completion code is set.
func parseSOLSelector(data []byte, needInstance bool) (instance uint8, cc types.CompletionCode, ok bool) {
	if len(data) < 1 || (needInstance && len(data) < 2) {
		return 0, types.CodeRequestDataLengthInvalid, false
	}
	if data[0]&0x3f != payloadTypeSOL {
		// 0h (IPMI Message) is not activatable; OEM/other types are not
		// implemented by this BMC (Table 24-2 note).
		return 0, types.CodeRequestDataFieldInvalid, false
	}
	if !needInstance {
		return 0, types.CodeOK, true
	}
	instance = data[1] & 0x0f
	if instance == 0 || instance > bmc.SOLMaxInstances {
		return 0, types.CodeParameterOutOfRange, false
	}
	return instance, types.CodeOK, true
}

func handleActivatePayload(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	_, cc, ok := parseSOLSelector(data, true)
	if !ok {
		return nil, cc, nil
	}
	// Payload activation is an RMCP+ concept; a v1.5 session has no payload
	// carrier to activate onto.
	if hctx.Session == nil {
		return nil, types.CodeNotSupported, nil
	}

	var aux uint8
	if len(data) >= 3 {
		aux = data[2]
	}
	// Bit [5] (Test Mode) is accepted but never enabled: the auxiliary
	// response data reports test mode inactive (Table 24-2).
	if _, err := hctx.BMC.SOL.Activate(ctx, hctx.Session, aux&0x80 != 0, aux&0x40 != 0); err != nil {
		return nil, solCommandCC(err), nil
	}

	port := hctx.BMC.SOL.Config().PayloadPort

	// Table 24-2 response: aux data (4 bytes LE, bit0 = test mode enabled),
	// payload sizes, UDP port, VLAN (FFFFh = not used).
	resp := make([]byte, 12)
	binary.LittleEndian.PutUint16(resp[4:6], 255) // inbound payload size incl. 4-byte SOL header
	binary.LittleEndian.PutUint16(resp[6:8], 255) // outbound payload size
	binary.LittleEndian.PutUint16(resp[8:10], port)
	binary.LittleEndian.PutUint16(resp[10:12], 0xffff)
	return resp, types.CodeOK, nil
}

// handleSuspendResumePayloadEncryption implements Table 24-5 for the SOL
// payload (Table 24-4: the command controls whether SOL payload data from
// the BMC is encrypted). Run-time control only exists when the channel can
// do encryption at all, which is exactly when the command is mandatory
// (Table 24-1 footnote 3).
func handleSuspendResumePayloadEncryption(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	_, cc, ok := parseSOLSelector(data, true)
	if !ok {
		return nil, cc, nil
	}
	if len(data) < 3 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	// Run-time encryption control is an RMCP+ concept; a v1.5 session has no
	// payload protection to toggle.
	if hctx.Session == nil {
		return nil, types.CodeNotSupported, nil
	}
	if err := hctx.BMC.SOL.SuspendResumeEncryption(hctx.Session, data[2]&0x1f); err != nil {
		return nil, solCommandCC(err), nil
	}
	return nil, types.CodeOK, nil
}

func handleDeactivatePayload(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	_, cc, ok := parseSOLSelector(data, true)
	if !ok {
		return nil, cc, nil
	}
	if hctx.Session == nil {
		return nil, types.CodeNotSupported, nil
	}
	if err := hctx.BMC.SOL.Deactivate(hctx.Session); err != nil {
		return nil, solCommandCC(err), nil
	}
	return nil, types.CodeOK, nil
}

func handleGetPayloadActivationStatus(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	_, cc, ok := parseSOLSelector(data, false)
	if !ok {
		return nil, cc, nil
	}
	capacity, active1to8, active9to16 := hctx.BMC.SOL.ActivationStatus()
	return []byte{capacity, active1to8, active9to16}, types.CodeOK, nil
}

func handleGetPayloadInstanceInfo(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	instance, cc, ok := parseSOLSelector(data, true)
	if !ok {
		return nil, cc, nil
	}
	resp := make([]byte, 12)
	binary.LittleEndian.PutUint32(resp[0:4], hctx.BMC.SOL.ActiveSessionID(instance))
	// Table 24-7 SOL payload-specific data byte 1: system serial port number
	// being redirected (1-based; this BMC redirects exactly one port).
	resp[4] = 1
	return resp, types.CodeOK, nil
}

func handleSetUserPayloadAccess(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 6 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	op := data[1] >> 6
	if op > 1 {
		return nil, types.CodeRequestDataFieldInvalid, nil // 10b/11b reserved (Table 24-8)
	}
	userID := data[1] & 0x3f
	// The read-modify-write runs on the live user under the store's write
	// lock, so it is atomic against a concurrent Set/Get on another session
	// authenticated as the same user.
	if err := hctx.BMC.Users.Update(userID, func(u *bmc.User) error {
		u.SetPayloadAccess(resolveChannel(hctx, data[0]&0x0f), op == 0, data[2], data[4])
		return nil
	}); err != nil {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}
	return nil, types.CodeOK, nil
}

func handleGetUserPayloadAccess(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 2 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	user, err := hctx.BMC.Users.Get(data[1] & 0x3f)
	if err != nil {
		return nil, types.CodeRequestDataFieldInvalid, nil
	}
	access := user.PayloadAccessFor(resolveChannel(hctx, data[0]&0x0f))
	return []byte{access.Standard1, 0, access.OEM1, 0}, types.CodeOK, nil
}

func handleGetChannelPayloadSupport(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 1 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	// §24.8 response: standard payload enables 1-2, session-setup payload
	// enables 1-2, OEM payload enables 1-2, then two reserved bytes — 8 in
	// total; spec-conformant clients reject shorter replies.
	resp := make([]byte, 8)
	// Standard payload types 0-7: IPMI Message is always carried; SOL only
	// when a console exists and the type is enabled (Table 26-5 #1).
	resp[0] = 0x01
	if hctx.BMC.SOL.Supported() {
		resp[0] |= 0x02
	}
	// Session setup payload types 0-7 (Table 13-16): Open Session request/
	// response + RAKP messages 1-4.
	resp[2] = 0x3f
	return resp, types.CodeOK, nil
}

func handleGetChannelPayloadVersion(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 2 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	if data[1]&0x3f != payloadTypeSOL || !hctx.BMC.SOL.Supported() {
		return nil, types.CodeGetChannelPayloadVersionNotAvailable, nil
	}
	return []byte{solFormatVersion}, types.CodeOK, nil
}

func handleSetSOLConfigParam(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 2 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	// The channel nibble selects which channel's parameters to write; this
	// BMC keeps one SOL parameter set (single LAN channel model), so the
	// selector alone decides.
	cc := hctx.BMC.SOL.Config().SetParam(data[1], data[2:])
	return nil, cc, nil
}

func handleGetSOLConfigParam(ctx context.Context, hctx *HandlerContext, data []byte) ([]byte, types.CompletionCode, error) {
	if len(data) < 4 {
		return nil, types.CodeRequestDataLengthInvalid, nil
	}
	revisionOnly := data[0]&0x80 != 0
	paramData, ok := hctx.BMC.SOL.Config().GetParam(data[1])
	if !ok {
		return nil, types.CodeParameterNotSupported, nil
	}
	// Parameter revision 11h: present revision 1, backward compatible to 1
	// (Table 26-4).
	resp := []byte{0x11}
	if !revisionOnly {
		resp = append(resp, paramData...)
	}
	return resp, types.CodeOK, nil
}
