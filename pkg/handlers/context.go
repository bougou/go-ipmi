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
)

// HandlerContext carries per-request BMC state to a [Handler].
// All fields are read-only from the handler's perspective; mutations must go
// through the store methods (which are goroutine-safe).
type HandlerContext struct {
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
