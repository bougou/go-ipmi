package handlers

import (
	"context"

	"github.com/bougou/go-ipmi/pkg/bmc"
)

// privilegeExempt reports commands that do not require session privilege checks.
func privilegeExempt(netFn, cmd uint8) bool {
	if netFn != NetFnAppRequest {
		return false
	}
	switch cmd {
	case CmdGetChannelAuthCapabilities,
		CmdGetSessionChallenge,
		CmdActivateSession,
		CmdGetChannelCipherSuites:
		return true
	}
	return false
}

func sessionPrivilege(hctx *HandlerContext) (bmc.PrivilegeLevel, bool) {
	if hctx.V15Session != nil && hctx.V15Session.State == bmc.V15SessionStateActive {
		return hctx.V15Session.PrivilegeLevel, true
	}
	if hctx.Session != nil {
		return hctx.Session.PrivilegeLevel, true
	}
	return 0, false
}

// checkCommandPrivilege enforces per-command minimum privilege (spec §18.16).
func checkCommandPrivilege(hctx *HandlerContext, netFn, cmd uint8) CompletionCode {
	if privilegeExempt(netFn, cmd) {
		return CodeOK
	}
	priv, ok := sessionPrivilege(hctx)
	if !ok {
		return CodeOK
	}
	if priv < MinimumPrivilege(netFn, cmd) {
		return CodeInsufficientPrivilege
	}
	return CodeOK
}

type dispatchingHandler struct {
	inner Handler
	netFn uint8
	cmd   uint8
}

func (d *dispatchingHandler) Handle(ctx context.Context, hctx *HandlerContext, req []byte) ([]byte, CompletionCode, error) {
	if cc := checkCommandPrivilege(hctx, d.netFn, d.cmd); cc != CodeOK {
		return nil, cc, nil
	}
	return d.inner.Handle(ctx, hctx, req)
}
