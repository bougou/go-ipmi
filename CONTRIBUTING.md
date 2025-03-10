# Contributing

Each command defined in the IPMI specification is a pair of request/response messages.
These IPMI commands are implemented as methods of the `ipmi.Client` struct in this library.

Using `ipmitool` as an example, some `ipmitool` command lines are implemented by calling just one underlying IPMI command,
while many others are not. For instance, `ipmitool sdr list` is a loop of `GetSDR` IPMI commands.

This library also implements some methods that are not IPMI commands defined
in the IPMI specification, but rather common helpers, like `GetSDRs` to get all SDRs.

## IPMI Command Guideline

For an IPMI Command `DoSomething`:

- You must define `DoSomethingRequest` which conforms to the `ipmi.Request` interface; it holds the request message data.
- You must define `DoSomethingResponse` which conforms to the `ipmi.Response` interface; it holds the response message data.
- You must define the `DoSomething` method on `ipmi.Client`

For the `DoSomething` method, you can pass `DoSomethingRequest` directly as the input parameter, like:

```go
func (c *Client) DoSomething(ctx context.Context, request *DoSomethingRequest) (response *DoSomethingResponse, err error) {
  response = &DoSomethingResponse{}
  err = c.Exchange(ctx, request, response)
  return
}
```

or, you can pass plain parameters and construct the `DoSomethingRequest` in the method body, like:

```go
func (c *Client) DoSomething(ctx context.Context, param1 string, param2 string) (response *DoSomethingResponse, err error) {
  request := &DoSomethingRequest{
    // construct by using input params
  }
  response = &DoSomethingResponse{}
  err = c.Exchange(ctx, request, response)
  return
}
```

Calling the `Exchange` method of `ipmi.Client` will handle all other complex underlying work.

## ipmi.Request interface

```go
type Request interface {
	// Pack encodes the object to data bytes
	Pack() []byte
	// Command return the IPMI command info (NetFn/Cmd).
	// All IPMI specification specified commands are already predefined in this repo.
	Command() Command
}

```
## ipmi.Response interface

```go
type Response interface {
	// Unpack decodes the object from data bytes
	Unpack(data []byte) error
	// CompletionCodes returns a map of command-specific completion codes
	CompletionCodes() map[uint8]string
	// Format return a formatted human friendly string
	Format() string
}
```

## IPMI Command Request

## IPMI Command Response

- Define necessary fields per IPMI specification, but DO NOT define the completion code field in the Response struct.
- If there are no command-specific completion codes, just return an empty map for the `CompletionCodes()` method.
