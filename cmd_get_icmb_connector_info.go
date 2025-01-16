package ipmi

import (
	"context"
	"fmt"
)

type GetICMBConnectorInfoRequest struct {
	NumericID uint8
}

type GetICMBConnectorInfoResponse struct {
	Flags           uint8
	ConnectorsCount uint8
	TypeLength      TypeLength
	IDString        []byte
}

func (req *GetICMBConnectorInfoRequest) Command() Command {
	return CommandGetICMBConnectorInfo
}

func (req *GetICMBConnectorInfoRequest) Pack() []byte {
	return []byte{}
}

func (res *GetICMBConnectorInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.Flags = msg[0]
	res.ConnectorsCount = msg[1]
	res.TypeLength = TypeLength(msg[2])

	if len(msg) > 3 {
		res.IDString, _, _ = unpackBytes(msg, 3, len(msg)-3)
	}

	return nil
}

func (res *GetICMBConnectorInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetICMBConnectorInfoResponse) Format() string {
	return fmt.Sprintf(`
        Flags            : %#02x
				Connectors Count : %#02x
				Type Length      : %#02x
				ID String        : %s
`,
		res.Flags,
		res.ConnectorsCount,
		res.TypeLength,
		res.IDString)
}

func (c *Client) GetICMBConnectorInfo(ctx context.Context, numericID uint8) (response *GetICMBConnectorInfoResponse, err error) {
	request := &GetICMBConnectorInfoRequest{
		NumericID: numericID,
	}
	response = &GetICMBConnectorInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
