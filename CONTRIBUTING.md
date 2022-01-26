# Contributing

Each command defined in the IPMI specification is a pair of request/response messages.
These IPMI commands are implemented as methods of the `ipmi.Client` struct in this library.

`ipmitool` as example, some `ipmitool` cmdline are realized by calling just one underlying IPMI command, but many others are not. Like `ipmitool sdr list`, it's a loop of `GetSDR` IPMI command.

So this library also implements some methods that are not IPMI commands defined
in IPMI sepcification, but just some common helpers, like `GetSDRs` to get all SDRs.

## IPMI Command Guideline

For a IPMI Command `DoSomething`:

- Must define `DoSomethingRequest` which conforms to the `ipmi.Request` interface, it holds the request message data.
- Must define `DoSomethingResponse` which conforms to the `ipmi.Response` interface, it holds the resonse message data.
- Must define `DoSomething` method on `ipmi.Client`

For `DoSomething` method, you can pass `DoSomethingRequest` directly as the input parameter, like:

```go
func (c *Client) DoSomething(request *DoSomethingRequest) (response *DoSomethingResponse, err error) {
  response := &DoSomethingResponse{}
  err := c.Exchange(request, response)
  return
}
```

or, you can pass some plain parametes, and construct the `DoSomethingRequest` in method body, like:

```go
func (c *Client) DoSomething(param1 string, param2 string) (response *DoSomethingResponse, err error) {
  request := &DoSomethingRequest{
    // construct by using input params
  }
  response := &DoSomethingResponse{}
  err := c.Exchange(request, response)
  return
}
```

Calling `Exchange` method of `ipmi.Client` will fullfil all other complex underlying works.

## ipmi.Request interface

## ipmi.Response interface
