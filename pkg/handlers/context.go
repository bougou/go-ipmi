// Package handlers provides the IPMI command dispatch infrastructure and a set
// of default BMC command handlers.
//
// # Architecture
//
// A [Handler] receives a raw IPMI request body, uses [HandlerContext] to access
// BMC state and hardware, and returns a raw response body plus a completion
// code.  The server layer handles all RMCP/IPMI framing, encryption, and
// sequence-number tracking – handlers never touch wire bytes.
//
// # Composability
//
// Callers can:
//   - Replace individual handlers via [Registry.Register].
//   - Wrap all handlers with [Registry.Use] middleware (e.g., for audit logging).
//   - Merge registries to add OEM command namespaces.
package handlers

import (
	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/types"
)

// HandlerContext carries per-request BMC state to a [Handler].
//
// Concurrency contract:
//   - Session / V15Session: the server holds the session's ProcMu for the
//     entire dispatch, so a handler may read and write these session fields
//     directly (only Set Session Privilege Level does, on PrivilegeLevel).
//     Registry-dispatched handlers must NOT take ProcMu themselves; it is
//     already held, and the mutex is not reentrant. The RAKP handlers
//     ([HandleOpenSession], [HandleRAKP1], [HandleRAKP3]) are the exception:
//     they are dispatched outside the in-session path and take ProcMu
//     themselves, so calling them from a registry-dispatched handler
//     self-deadlocks.
//   - User / Channel: independent snapshot copies, safe to read without locking
//     but not connected to the store. To change a user or channel at runtime,
//     use [bmc.UserStore.Update] or [bmc.ChannelStore.Set].
//   - BMC: goroutine-safe; mutate shared state only through its store methods.
type HandlerContext struct {
	// Command identifies the request being dispatched.  [Registry.Dispatch]
	// fills it in from the command table, so middleware and handlers can name
	// the request without re-deriving it from NetFn/Cmd.  For a command with no
	// entry in the table, only ID and NetFn are populated and Name is empty.
	Command types.Command

	// BMC is the top-level BMC state.
	BMC *bmc.BMC

	// Session is the authenticated RMCP+ session, or nil for pre-session requests.
	Session *bmc.Session

	// V15Session is the authenticated IPMI v1.5 session, or nil.
	V15Session *bmc.V15Session

	// Channel is the channel the request arrived on.
	Channel *bmc.Channel

	// User is the authenticated user for this session, or nil for anonymous.
	User *bmc.User
}
