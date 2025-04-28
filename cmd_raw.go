package ipmi

import (
	"context"
	"fmt"
	"strings"
)

type CommandRawRequest struct {
	NetFn NetFn
	Cmd   uint8
	Data  []byte
	Name  string
}

type CommandRawResponse struct {
	Response []byte
}

func (req *CommandRawRequest) Command() Command {
	return Command{ID: req.Cmd, NetFn: req.NetFn, Name: req.Name}
}

func (req *CommandRawRequest) Pack() []byte {
	out := make([]byte, len(req.Data))

	packBytes(req.Data, out, 0)

	return out
}

func (res *CommandRawResponse) Unpack(msg []byte) error {
	res.Response = msg
	return nil
}

func (res *CommandRawResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *CommandRawResponse) Format() string {
	// convert the byte array to a slice of hex strings
	hexStrings := make([]string, len(res.Response))
	for i, b := range res.Response {
		hexStrings[i] = fmt.Sprintf("0x%02X", b)
	}

	// join the hex strings with commas
	hexString := strings.Join(hexStrings, ", ")

	return fmt.Sprintf("raw.Response = %s", hexString)
}

func (c *Client) RawCommand(ctx context.Context, netFn NetFn, cmd uint8, data []byte, name string) (response *CommandRawResponse, err error) {
	request := &CommandRawRequest{
		NetFn: netFn,
		Cmd:   cmd,
		Data:  data,
		Name:  name,
	}
	response = &CommandRawResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
