// Package ipmi provides an IPMI (Intelligent Platform Management Interface)
// client and server implementation in Go.
//
// The implementation is split across sub-packages:
//
//   - [github.com/bougou/go-ipmi/pkg/types] — protocol types, constants, and
//     wire-format helpers shared by client and server.
//   - [github.com/bougou/go-ipmi/pkg/crypto] — shared AES/HMAC/RAKP/v1.5 AuthCode
//     implementations (v2.0§13 / v1.5§18.15).
//   - [github.com/bougou/go-ipmi/pkg/rmcpplus] — RMCP+ Open Session and RAKP
//     message Pack/Unpack (v2.0§13.17–13.23).
//   - [github.com/bougou/go-ipmi/pkg/client] — IPMI client (LAN, LAN+, open,
//     ipmitool).
//   - github.com/bougou/go-ipmi/pkg/command/* — per-command request/response types
//     grouped by netFn (app, chassis, sensor, storage, transport, dcmi, oem).
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
