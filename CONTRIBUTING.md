# Contributing

An IPMI command is a request/response pair. In this tree, command methods live
on `client.Client` (`pkg/client`). Request/response wire types for a NetFn
usually sit under `pkg/command/<netfn>/`; shared interfaces and helpers are in
`pkg/types`.

`ipmitool` is a useful comparison point: some of its subcommands are one IPMI
transaction (`mc info` → `GetDeviceID`), others are loops (`sdr list` →
repeated `GetSDR`). This library also ships conveniences such as `GetSDRs`
that are not themselves single specification commands. Coverage:
[docs/commands.md](./docs/commands.md).

## Adding a command

For a command `DoSomething`:

1. Define `DoSomethingRequest` implementing `types.Request`.
2. Define `DoSomethingResponse` implementing `types.Response`.
3. Add `DoSomething` on `client.Client`.

Either accept the request object:

```go
func (c *Client) DoSomething(ctx context.Context, request *DoSomethingRequest) (*DoSomethingResponse, error) {
	response := &DoSomethingResponse{}
	err := c.Exchange(ctx, request, response)
	return response, err
}
```

or accept plain parameters and build the request inside the method:

```go
func (c *Client) DoSomething(ctx context.Context, param1, param2 string) (*DoSomethingResponse, error) {
	request := &DoSomethingRequest{
		// fill from param1, param2
	}
	response := &DoSomethingResponse{}
	err := c.Exchange(ctx, request, response)
	return response, err
}
```

`Exchange` owns the session/transport work.

## `types.Request` / `types.Response`

```go
type Request interface {
	Pack() []byte
	Command() Command // NetFn/Cmd; predefined commands live in this repo
}

type Response interface {
	Unpack(data []byte) error
	Format() string
}
```

Both interfaces (and `Command`) are in `pkg/types`.

## Responses and completion codes

- Put the fields the specification defines on the response struct. Do not add a
  completion-code field there.
- Command-specific codes (80h–BEh) are tabulated in
  `pkg/types/types_completion_code.go`, keyed by `Command.Key()` (NetFn + Cmd,
  without Name). Reuse shared maps such as `paramConfigSetCC` / `selEraseCC`
  when the wording matches. Commands with only generic codes need no entry.
- At runtime use `types.StrCC(cmd, ccode)` (also `types.AllCC` /
  `types.CommandSpecificCC`). Prefer `request.Command()` as the identity — that
  is what went on the wire.
