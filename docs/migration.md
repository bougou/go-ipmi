# Migration from `v0.8.x`

Through `v0.8.x` the client API lived in one package:

```go
import "github.com/bougou/go-ipmi"

c, err := ipmi.NewClient(host, port, user, pass)
```

That package is gone. Import the subpackages under `pkg/` instead. For the
common `NewClient` / method-call path the edit is mostly mechanical.

```go
// before
import "github.com/bougou/go-ipmi"
c, err := ipmi.NewClient(host, port, user, pass)
c.WithInterface(ipmi.InterfaceLanplus)

// after
import "github.com/bougou/go-ipmi/pkg/client"
c, err := client.NewClient(host, port, user, pass)
c.WithInterface(client.InterfaceLanplus)
```

`*Client` methods (`Connect`, `GetDeviceID`, …) keep the same names.

| Was under `ipmi`                                   | Now                                                |
| -------------------------------------------------- | -------------------------------------------------- |
| `NewClient`, `Interface*`, `RenderTable`, …        | `pkg/client`                                       |
| `PrivilegeLevel`, `FormatSELs`, `CipherSuiteID`, … | `pkg/types`                                        |
| `GetDeviceIDRequest`, …                            | `pkg/command/<netfn>` (e.g. `pkg/command/app`)     |
| `RAKPMessage1`, Open Session types, …              | `pkg/rmcpplus`                                     |
| _(server APIs did not exist in v0.8.x)_            | `pkg/server`, `pkg/bmc`, `pkg/handlers`, `pkg/hal` |

Most programs only need `pkg/client`. Pull in `pkg/types` (or a
`pkg/command/...` package) when you reference types or formatters that used to
share the root package with `NewClient`.

The module path is still `github.com/bougou/go-ipmi`; there is just nothing
importable at that exact path. Use a `pkg/...` import.
