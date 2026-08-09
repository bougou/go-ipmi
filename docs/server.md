# Server

The reference BMC is meant for development, simulation, and end-to-end tests.
It is not a substitute for production BMC firmware.

| Package | Responsibility |
| ------- | -------------- |
| [`pkg/server`](../pkg/server) | Serve loop, session framing, dispatch |
| [`pkg/bmc`](../pkg/bmc) | Users, channels, sessions, device info |
| [`pkg/handlers`](../pkg/handlers) | Per-command handlers |
| [`pkg/hal`](../pkg/hal) | Hardware abstraction; `hal/mock` for tests |
| [`pkg/transport`](../pkg/transport) | `PacketConn`; `transport/udp` for UDP |

One UDP port serves both IPMI v2.0 / RMCP+ (`-I lanplus`) and IPMI v1.5
(`-I lan`, e.g. `-A MD5`).

## `goipmi-server`

```bash
make build
./_output/goipmi-server
```

| Variable | Default | Meaning |
| -------- | ------- | ------- |
| `GOIPMI_SERVER_PORT` | `623` | UDP listen port |
| `GOIPMI_SERVER_USER` | `ADMIN` | Username |
| `GOIPMI_SERVER_PASS` | `ADMIN` | Password |
| `GOIPMI_SERVER_CIPHER_SUITES` | `3,17` | Advertised RMCP+ cipher suite IDs |
| `GOIPMI_SERVER_V15_AUTH_TYPES` | `md5` | v1.5 auth types: `none`, `md2`, `md5`, `password`, `oem` |
| `GOIPMI_SERVER_V15` | `1` | `0` / `false` disables v1.5; lanplus stays up |
| `GOIPMI_SERVER_TRACE` | `0` | Log dispatched commands to stderr |

```bash
./_output/goipmi -I lanplus -H 127.0.0.1 -p 623 -U ADMIN -P ADMIN mc info
./_output/goipmi -I lan -H 127.0.0.1 -p 623 -U ADMIN -P ADMIN mc info
```

`test/e2e/` covers client→simulator, ipmitool→server, and goipmi→goipmi-server.
Run the full set with `make test-e2e`.

## Embedding

Roughly what `cmd/goipmi-server` does:

```go
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
)

func main() {
	info := bmc.DeviceInfo{
		DeviceID:       32,
		DeviceRevision: 1,
		FirmwareMajor:  1,
		FirmwareMinor:  0,
		IPMIVersion:    0x20,
		ManufacturerID: 0x000157,
		ProductID:      0x0001,
	}
	var guid [16]byte
	copy(guid[:], "example-bmc\x00\x00\x00\x00\x00")

	b := bmc.New(info, guid, mock.New(), bmc.WithClock(clock.Real))

	user, err := b.Users.Add(2, "ADMIN")
	if err != nil {
		panic(err)
	}
	user.SetPassword([]byte("ADMIN"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{
		MaxPrivilege: bmc.PrivilegeLevelAdministrator,
		Enabled:      true,
	}

	conn, err := udp.Listen(":623")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	srv := server.NewServer(b, conn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	if err := srv.Serve(ctx); err != nil {
		panic(err)
	}
}
```

Useful knobs:

- `server.WithHandlerRegistry` — replace or wrap handlers (OEM commands, tracing)
- `server.WithCipherSuites` / `bmc.WithCipherSuites` — advertised RMCP+ suites
- `server.WithV15AuthTypes` / `server.WithV15Disabled` — v1.5 auth policy
- a custom `hal.HAL` instead of `hal/mock`
- a custom `transport.PacketConn` if you already own the socket
