package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bougou/go-ipmi/pkg/handlers"
	"github.com/bougou/go-ipmi/pkg/types"
)

// tracingRegistry builds the standard command set with [traceCommands] wrapped
// around it. Middleware is not applied retroactively, so Use must come before
// the handlers are registered — which is also why this cannot just decorate the
// registry [server.NewServer] would have built by default.
func tracingRegistry() *handlers.Registry {
	reg := handlers.NewRegistry()
	reg.Use(traceCommands)
	handlers.RegisterAllHandlers(reg)
	return reg
}

// traceCommands logs one line per dispatched command. Driving this server with
// `ipmitool -I lanplus ... chassis power on` produces, abridged:
//
//	trace "Get Channel Authentication Capabilities" netfn=0x06 cmd=0x38 user=-         priv=-             cc=0x00 (Command completed normally) req=2B resp=8B 60.828µs
//	trace "Get Channel Cipher Suites"               netfn=0x06 cmd=0x54 user=-         priv=-             cc=0x00 (Command completed normally) req=3B resp=11B 8.759µs
//	trace "Set Session Privilege Level"             netfn=0x06 cmd=0x3b user=ADMIN     priv=ADMINISTRATOR cc=0x00 (Command completed normally) req=1B resp=1B 422ns
//	trace <unregistered>                            netfn=0x2c cmd=0x3e user=ADMIN     priv=ADMINISTRATOR cc=0xc1 (Invalid command) req=2B resp=0B 1.953µs err=no handler for netFn=0x2c cmd=0x3e
//	trace "Get Device ID"                           netfn=0x06 cmd=0x01 user=ADMIN     priv=ADMINISTRATOR cc=0x00 (Command completed normally) req=0B resp=11B 3.908µs
//	trace "Chassis Control"                         netfn=0x00 cmd=0x02 user=ADMIN     priv=ADMINISTRATOR cc=0x00 (Command completed normally) req=1B resp=0B 963ns
//	trace "Close Session"                           netfn=0x06 cmd=0x3c user=ADMIN     priv=ADMINISTRATOR cc=0x00 (Command completed normally) req=4B resp=0B 1.001µs
//
// Every field comes from what the dispatcher already knew, which is the point:
// middleware is the layer that wants to describe a request, so
// [handlers.HandlerContext] hands it the description. Naming the command needs
// no table of its own here — without [types.Command] on the context this would
// print bare numbers, or each handler would have to report its own name upward,
// putting the knowledge in the layer least able to act on it.
//
// The two lines that are not a chassis power-on show why that matters. The
// pre-session commands carry no user or privilege because they are the exchange
// that establishes them, and they are exactly the commands the privilege check
// exempts. The 0x2c/0x3e line is ipmitool probing for DCMI support, which this
// server does not implement: unregistered commands run middleware before
// answering [types.CodeInvalidCommand], because an initiator probing for
// commands the BMC does not have is what an audit log exists to catch, and
// short-circuiting would make it invisible.
func traceCommands(next handlers.Handler) handlers.Handler {
	return handlers.HandlerFunc(func(ctx context.Context, hctx *handlers.HandlerContext, req []byte) ([]byte, types.CompletionCode, error) {
		start := time.Now()
		resp, cc, err := next.Handle(ctx, hctx, req)

		fmt.Fprintln(os.Stderr, "goipmi-server: "+traceLine(hctx, cc, err, len(req), len(resp), time.Since(start)))

		return resp, cc, err
	})
}

// traceLine formats one trace record. Splitting it from the middleware keeps
// the interesting half testable without capturing stderr.
func traceLine(hctx *handlers.HandlerContext, cc types.CompletionCode, err error, reqLen, respLen int, elapsed time.Duration) string {
	// 41 columns fits the widest name this server registers, "Get Channel
	// Authentication Capabilities", in quotes. cc name comes from
	// [types.StrCC] so command-specific 80h-BEh codes resolve without
	// a decoded Response.
	line := fmt.Sprintf("trace %-41s netfn=0x%02x cmd=0x%02x user=%-9s priv=%-13s cc=0x%02x (%s) req=%dB resp=%dB %v",
		commandName(hctx), uint8(hctx.Command.NetFn), hctx.Command.ID,
		userName(hctx), privilege(hctx), uint8(cc), types.StrCC(hctx.Command, uint8(cc)),
		reqLen, respLen, elapsed)
	if err != nil {
		line += " err=" + err.Error()
	}
	return line
}

// commandName prefers the spec name the registry recorded. Commands with no
// handler never got one, and are worth telling apart from the rest.
func commandName(hctx *handlers.HandlerContext) string {
	if hctx.Command.Name == "" {
		return "<unregistered>"
	}
	return `"` + hctx.Command.Name + `"`
}

func userName(hctx *handlers.HandlerContext) string {
	if hctx.User == nil || hctx.User.Name == "" {
		return "-"
	}
	return hctx.User.Name
}

// privilege reports the level the session negotiated, which is what the
// privilege check inside the dispatch chain compared against. A request with no
// session yet — Get Channel Authentication Capabilities and the rest of the
// setup exchange — has none, and those are the commands exempt from the check.
// The conversion is what [bmc.PrivilegeLevel] documents itself as: a mirror of
// the wire type, kept distinct so bmc holds no wire conversions. Only the wire
// type spells the levels out.
func privilege(hctx *handlers.HandlerContext) string {
	switch {
	case hctx.Session != nil:
		return types.PrivilegeLevel(hctx.Session.PrivilegeLevel).String()
	case hctx.V15Session != nil:
		return types.PrivilegeLevel(hctx.V15Session.PrivilegeLevel).String()
	default:
		return "-"
	}
}
