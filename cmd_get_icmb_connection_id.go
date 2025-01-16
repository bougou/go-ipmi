package ipmi

import (
	"context"
	"fmt"
)

type GetICMBConnectionIDRequest struct {
	NumericID uint8
}

type GetICMBConnectionIDResponse struct {
	NumericID  uint8
	TypeLength TypeLength
	IDString   []byte
}

func (req *GetICMBConnectionIDRequest) Command() Command {
	return CommandGetICMBConnectionID
}

func (req *GetICMBConnectionIDRequest) Pack() []byte {
	return []byte{}
}

func (res *GetICMBConnectionIDResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	res.NumericID = msg[0]
	res.TypeLength = TypeLength(msg[1])

	if len(msg) > 2 {
		res.IDString, _, _ = unpackBytes(msg, 2, len(msg)-2)
	}

	return nil
}

func (res *GetICMBConnectionIDResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetICMBConnectionIDResponse) Format() string {
	return fmt.Sprintf(`
				Numeric ID  : %#02x
				Type Length : %#02x
				ID String   : %s
`,
		res.NumericID,
		res.TypeLength,
		res.IDString)
}

func (c *Client) GetICMBConnectionID(ctx context.Context) (response *GetICMBConnectionIDResponse, err error) {
	request := &GetICMBConnectionIDRequest{}
	response = &GetICMBConnectionIDResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
