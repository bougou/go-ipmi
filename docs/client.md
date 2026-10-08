# Client

```go
import "github.com/bougou/go-ipmi/pkg/client"
```

## Interfaces

| Interface           | How                                        | Notes                                                    |
| ------------------- | ------------------------------------------ | -------------------------------------------------------- |
| `lanplus` (default) | `client.NewClient(host, port, user, pass)` | IPMI v2.0 / RMCP+ over UDP                               |
| `lan`               | `c.WithInterface(client.InterfaceLan)`     | IPMI v1.5 over UDP                                       |
| `open`              | `client.NewOpenClient()`                   | System interface: Linux OpenIPMI, Windows Microsoft_IPMI |
| `tool`              | `client.NewToolClient(path)`               | Runs an `ipmitool` binary or wrapper                     |

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

Backends live in `pkg/open`: Linux talks to `/dev/ipmiN` via
ioctl; Windows uses the Microsoft_IPMI WMI provider (COM by default, with a
PowerShell fallback).

### Linux architectures

The Linux backend uses native kernel layouts: C `int` and `unsigned int` are
32-bit fields, while the message ID is a native-width C `long`. MIPS and PowerPC
use different ioctl encoding constants from the other Linux architectures.
Production builds remain pure Go and do not require a C compiler or cgo.

CI compares every declared OpenIPMI struct's layout and scalar signedness, and
all ioctl numbers, against the target's `<linux/ipmi.h>`. It runs these checks
and the OpenIPMI tests on amd64, 386, arm (GOARM=6 and 7), arm64, ppc64le, s390x,
and mips, using QEMU for non-native targets. These checks validate the ABI;
access to an IPMI device still depends on the host's kernel, hardware and device
permissions. Emulation does not substitute for hardware testing.

For a native Linux check, run `CC=cc go test -v ./pkg/open`. The C oracle needs
Linux headers and static C libraries. Cross-testing also needs a target compiler
and emulator, for example:

```sh
GOARCH=arm GOARM=7 CGO_ENABLED=0 CC=arm-linux-gnueabihf-gcc \
  IPMI_UAPI_EXEC=qemu-arm go test -exec=qemu-arm -v ./pkg/open
```

Without an explicit `CC`, the oracle skips if `cc` is unavailable. CI sets `CC`
explicitly, so a missing compiler fails the check.

**Low-level API compatibility:** correcting the kernel declarations changes
exported field types in `IPMI_REQ`, `IPMI_RECV`, the address structs,
`IPMI_CMDSPEC_CHANS`, and `IPMI_TIMING_PARAMS`. Code constructing these structs
may need explicit conversions (for example, `MsgID: int(id)` and
`AddrLen: uint32(length)`). The transport-neutral `Request` and `Backend` APIs
are unchanged.

## Options

| Method                     | Effect                              |
| -------------------------- | ----------------------------------- |
| `WithDebug`                | Session / packet logging            |
| `WithInterface`            | `lan` / `lanplus` / `open` / `tool` |
| `WithTimeout`, `WithRetry` | Transport timing                    |
| `WithCipherSuiteID`        | Preferred RMCP+ cipher suites       |
| `WithMaxPrivilegeLevel`    | Cap session privilege               |
| `WithOpenBackend`          | Windows open backend selection      |
| `WithUDPProxy`             | Dial through a UDP proxy            |

## Concurrent LAN commands

LAN and LAN+ serialize each exchange from packet construction through response
parsing, including session keepalive traffic. This keeps request and session
sequence numbers aligned with wire order. A caller waiting for another exchange
can cancel that wait with its context. Configure and connect the client before
starting concurrent command calls.

LAN setup calls (`Connect`, `Connect15`, `Connect20`, and `ConnectAuto`) are
serialized and honor cancellation while waiting. Reconnecting stops and joins
the previous keepalive before changing session state. Do not overlap setup with
ordinary commands or configuration changes. Once setup starts keepalive and
succeeds, cancellation of its context does not undo the connection; call `Close`
to release the session.

`Close(ctx)` stops new commands and setup, cancels and joins any setup and
keepalive, attempts to close the BMC session within the supplied context, and
always closes the UDP socket. If its context expires while waiting for setup,
it skips session cleanup; canceled setup cannot start another keepalive.
Concurrent and repeated closes share the result; a waiting caller can cancel its
own wait. Create a new client to connect again after closing it.

Errors from canceled LAN command exchanges and setup calls match both the
context error and its cancellation cause with `errors.Is`. Work stopped by
`Close` matches `net.ErrClosed`, and a setup failure keeps its original error.

UDP hostname resolution, reads, writes, and exchange-slot waits honor caller
cancellation. Direct dialing retains IPv4 preference for dual-stack hostnames;
IPv6-only names and scoped IPv6 literals are also supported. The UDP
transport can be reused after cancellation or Close. Proxy dialers implementing
`proxy.ContextDialer` also receive cancellation. Legacy `proxy.Dialer` calls
cannot be interrupted internally; the caller can still return promptly, and any
connection returned later is closed. A legacy dialer that never returns can
retain its dial goroutine; use a context-aware dialer when bounded cleanup is
required. Custom request readers and datagram match callbacks must return
promptly because they run synchronously.

## Spec commands vs helpers

Specification commands are request/response pairs exposed as `Client` methods
that call `Exchange` underneath.

Some `ipmitool` subcommands map 1:1 onto a single IPMI command; others loop
(`ipmitool sdr list` repeatedly calls `GetSDR`). Helpers such as `GetSDRs` and
`GetSensors` are library conveniences, not single wire transactions. They are
marked with `*` in [Commands](./commands.md).

How to add a command: [Contributing](../CONTRIBUTING.md).
