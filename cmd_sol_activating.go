package ipmi

import "context"

// 26.1 SOL Activating Command
type SOLActivatingRequest struct {
	SessionState       uint8
	PayloadInstance    uint8
	FormatVersionMajor uint8
	FormatVersionMinor uint8
}

type SOLActivatingResponse struct {
}

func (req *SOLActivatingRequest) Command() Command {
	return CommandSOLActivating
}

func (req *SOLActivatingRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint8(req.SessionState, out, 0)
	packUint8(req.PayloadInstance, out, 1)
	packUint8(req.FormatVersionMajor, out, 2)
	packUint8(req.FormatVersionMinor, out, 3)
	return out
}

func (res *SOLActivatingResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SOLActivatingResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SOLActivatingResponse) Format() string {
	return ""
}

func (c *Client) SOLActivating(ctx context.Context, request *SOLActivatingRequest) (response *SOLActivatingResponse, err error) {
	response = &SOLActivatingResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
