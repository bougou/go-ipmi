package ipmi

import (
	"context"
	"fmt"
)

type GetICMBAddressRequest struct {
}

type GetICMBAddressResponse struct {
	ICMBAddr uint16
}

func (req *GetICMBAddressRequest) Command() Command {
	return CommandGetICMBAddress
}

func (req *GetICMBAddressRequest) Pack() []byte {
	return []byte{}
}

func (res *GetICMBAddressResponse) Unpack(msg []byte) error {
	if len(msg) < 2 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}
	res.ICMBAddr, _, _ = unpackUint16L(msg, 0)
	return nil
}

func (res *GetICMBAddressResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetICMBAddressResponse) Format() string {
	return fmt.Sprintf("#%02x", res.ICMBAddr)
}

func (c *Client) GetICMBAddress(ctx context.Context) (response *GetICMBAddressResponse, err error) {
	request := &GetICMBAddressRequest{}
	response = &GetICMBAddressResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
