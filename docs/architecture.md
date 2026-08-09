# Architecture

Earlier releases exposed a single flat package at the module root. The tree is
now split under `pkg/` so the client and the reference BMC can share protocol
code without dragging each other into the same API surface.

Import the package you need (`pkg/client`, `pkg/types`, `pkg/server`, …). If you
are updating code written against `v0.8.x`, see [Migration](./migration.md).

## Layout

```
github.com/bougou/go-ipmi/
├── cmd/
│   ├── goipmi/           # ipmitool-style CLI (library check-out tool)
│   └── goipmi-server/    # reference BMC
├── pkg/
│   ├── types/            # wire types, constants, pack/unpack
│   ├── crypto/           # AES / HMAC / RAKP / v1.5 AuthCode
│   ├── rmcpplus/         # Open Session + RAKP codecs
│   ├── protocol/         # payload IDs and framing helpers
│   ├── command/          # request/response types by NetFn
│   │   ├── app/
│   │   ├── chassis/
│   │   ├── sensor/
│   │   ├── storage/
│   │   ├── transport/
│   │   ├── dcmi/
│   │   └── oem/
│   ├── client/           # LAN, LAN+, Open, Tool
│   ├── open/             # in-band backends (Linux / Windows)
│   ├── server/           # serve loop, sessions, dispatch
│   ├── bmc/              # users, channels, sessions, device state
│   ├── handlers/         # command handlers
│   ├── hal/              # hardware abstraction (+ mock)
│   ├── transport/        # PacketConn (+ udp)
│   ├── clock/
│   └── utils/
├── specs/                # IPMI / DCMI / FRU PDFs
└── test/e2e/
```

## How the pieces connect

```
cmd/goipmi
    └─ pkg/client ──┬─ pkg/command/* (request/response)
                    ├─ pkg/open      (in-band)
                    └─ shared: types, crypto, rmcpplus, protocol
                                    │
cmd/goipmi-server ─ pkg/server ─ handlers ─ bmc ─ hal
                         └─ transport/udp
```

A client builds a session (`Connect`), then issues typed methods or raw
`Exchange` calls. The server binds a `transport.PacketConn`, owns BMC state
through `bmc.BMC`, and dispatches into `handlers`.

## Spec references in code

Comments cite the sections they implement, for example `v2.0§13` (RMCP+),
`v2.0§22` / `v1.5§18` (messaging), `v1.5§18.15` (AuthCode), `dcmi§…`, `fru§…`.
The PDFs themselves are under `specs/`.
