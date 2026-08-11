# [go-ipmi](https://github.com/bougou/go-ipmi)

Native Go implementation of IPMI. It does not shell out to `ipmitool`.

The library covers a remote/in-band client, a reference BMC useful for local
development and tests, and the shared protocol types used by both. Public
packages live under `pkg/`; start with `[pkg/client](./pkg/client)` for the
usual remote-console path.

## Compatibility

This module has not reached `v1.0.0`. The public API may change between
versions, including renamed packages, types, functions, and constants.
Pin to a specific version or commit for production use.

`v0` releases carry no backward-compatibility promise, per the
[Go module versioning convention](https://go.dev/ref/mod#versions).

## Install

```bash
go get github.com/bougou/go-ipmi
```

## Client

```go
package main

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/client"
)

func main() {
	c, err := client.NewClient("10.0.0.1", 623, "root", "123456")
	if err != nil {
		panic(err)
	}
	// c.WithInterface(client.InterfaceLan) // IPMI v1.5; default is lanplus (v2.0)
	// openClient, err := client.NewOpenClient() // in-band on Linux / Windows

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
}
```

Interfaces, options, and helpers: [docs/client.md](./docs/client.md).
Import path changes since `v0.8.x`: [docs/migration.md](./docs/migration.md).

## Server

`goipmi-server` is a dual-stack reference BMC (RMCP+ and IPMI v1.5 on one UDP
port). Build with `make build`, then run `_output/goipmi-server`. Defaults are
user/password `ADMIN`/`ADMIN` on port `623`.

Embedding the same stack in-process is described in [docs/server.md](./docs/server.md).

## Tools

```bash
make build
./_output/goipmi -I lanplus -H <bmc> -U <user> -P <pass> mc info
./_output/goipmi-server
```

`goipmi` mirrors common `ipmitool` subcommands so the library can be exercised
against real BMCs. It is a verification front-end, not a drop-in replacement.

## Docs

- [Architecture](./docs/architecture.md)
- [Client](./docs/client.md)
- [Server](./docs/server.md)
- [Commands](./docs/commands.md)
- [Migration](./docs/migration.md)
- [Contributing](./CONTRIBUTING.md)

## Specs

- [IPMI v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)
- [Platform Management FRU v1.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-platform-mgt-fru-info-storage-def-v1-0-rev-1-3-spec-update.pdf)
- [DCMI v1.5](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf)
- [PC SDRAM SPD](https://cdn.hackaday.io/files/10119432931296/Spdsd12b.pdf)

Checked-in copies (including IPMI v1.5) are under `[specs/](./specs/)`.
