# Client

```go
import "github.com/bougou/go-ipmi/pkg/client"
```

Wire types and formatters are in [`pkg/types`](../pkg/types). Per-command
request/response structs are under [`pkg/command/<netfn>`](../pkg/command).
Every command method takes `context.Context` as its first argument (since
v0.6.0).

## Interfaces

| Interface | How | Notes |
| --------- | --- | ----- |
| `lanplus` (default) | `client.NewClient(host, port, user, pass)` | IPMI v2.0 / RMCP+ over UDP |
| `lan` | `c.WithInterface(client.InterfaceLan)` | IPMI v1.5 over UDP |
| `open` | `client.NewOpenClient()` | System interface: Linux OpenIPMI, Windows Microsoft_IPMI |
| `tool` | `client.NewToolClient(path)` | Runs an `ipmitool` binary or wrapper |

```go
c, err := client.NewClient(host, port, user, pass)
if err != nil {
	return err
}
c.WithInterface(client.InterfaceLan) // or InterfaceLanplus
```

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/client"
	"github.com/bougou/go-ipmi/pkg/types"
)

func main() {
	c, err := client.NewClient("10.0.0.1", 623, "root", "123456")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		panic(err)
	}
	defer c.Close(ctx)

	res, err := c.GetDeviceID(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Format())

	entries, err := c.GetSELEntries(ctx, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println(types.FormatSELs(entries, nil))
}
```

## Open (in-band)

```go
c, err := client.NewOpenClient()
if err != nil {
	panic(err)
}
// Windows: pick a WMI backend before Connect (no-op elsewhere).
// c.WithOpenBackend(client.OpenBackendCOM)
// c.WithOpenBackend(client.OpenBackendPowerShell)

ctx := context.Background()
if err := c.Connect(ctx); err != nil {
	panic(err)
}
defer c.Close(ctx)
```

Backends live in [`pkg/open`](../pkg/open): Linux talks to `/dev/ipmiN` via
ioctl; Windows uses the Microsoft_IPMI WMI provider (COM by default, with a
PowerShell fallback).

## Options

| Method | Effect |
| ------ | ------ |
| `WithDebug` | Session / packet logging |
| `WithInterface` | `lan` / `lanplus` / `open` / `tool` |
| `WithTimeout`, `WithRetry` | Transport timing |
| `WithCipherSuiteID` | Preferred RMCP+ cipher suites |
| `WithMaxPrivilegeLevel` | Cap session privilege |
| `WithOpenBackend` | Windows open backend selection |
| `WithUDPProxy` | Dial through a UDP proxy |

## Spec commands vs helpers

Specification commands are request/response pairs exposed as `Client` methods
that call `Exchange` underneath.

Some `ipmitool` subcommands map 1:1 onto a single IPMI command; others loop
(`ipmitool sdr list` repeatedly calls `GetSDR`). Helpers such as `GetSDRs` and
`GetSensors` are library conveniences, not single wire transactions. They are
marked with `*` in [Commands](./commands.md).

How to add a command: [Contributing](../CONTRIBUTING.md).
