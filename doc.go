// Package ipmi provides an IPMI (Intelligent Platform Management Interface)
// client and server implementation in Go.
//
// The implementation is split across sub-packages:
//
//   - [github.com/bougou/go-ipmi/pkg/types] — protocol types, constants, and
//     wire-format helpers shared by client and server.
//   - [github.com/bougou/go-ipmi/pkg/client] — IPMI client (LAN, LAN+, open,
//     ipmitool) and all command request/response types.
//   - [github.com/bougou/go-ipmi/pkg/server] — IPMI BMC server (RMCP+,
//     handler registry, session management).
//   - [github.com/bougou/go-ipmi/pkg/bmc] — BMC in-memory state (users,
//     channels, sessions).
//   - [github.com/bougou/go-ipmi/pkg/handlers] — IPMI command handlers.
//   - [github.com/bougou/go-ipmi/pkg/hal] — Hardware abstraction layer
//     interfaces.
//   - [github.com/bougou/go-ipmi/pkg/transport] — Network transport
//     abstraction (PacketConn).
package ipmi
