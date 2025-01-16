package ipmi

import (
	"context"
	"fmt"
)

type SendICMBConnectionIDRequest struct {
	ICMBAddr uint16
}

type SendICMBConnectionIDResponse struct {
	Completion uint8
	NumericID  uint8
	TypeLength TypeLength
	IDString   []byte
}

func (req *SendICMBConnectionIDRequest) Command() Command {
	return CommandSendICMBConnectionID
}

func (req *SendICMBConnectionIDRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint16L(req.ICMBAddr, out, 0)
	return out
}

func (res *SendICMBConnectionIDResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.Completion = msg[0]
	res.NumericID = msg[1]
	res.TypeLength = TypeLength(msg[2])

	if len(msg) > 3 {
		res.IDString, _, _ = unpackBytes(msg, 3, len(msg)-3)
	}

	return nil
}

func (res *SendICMBConnectionIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SendICMBConnectionIDResponse) Format() string {
	return fmt.Sprintf(`
	      Completion  : %#02x
				Numeric ID  : %#02x
				Type Length : %#02x
				ID String   : %s
`,
		res.Completion,
		res.NumericID,
		res.TypeLength,
		res.IDString)
}

func (c *Client) SendICMBConnectionID(ctx context.Context, icmbAddr uint16) (response *SendICMBConnectionIDResponse, err error) {
	request := &SendICMBConnectionIDRequest{
		ICMBAddr: icmbAddr,
	}
	response = &SendICMBConnectionIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
