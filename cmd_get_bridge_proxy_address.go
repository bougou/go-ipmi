package ipmi

import (
	"context"
	"fmt"
)

type GetBridgeProxyAddressRequest struct {
}

type GetBridgeProxyAddressResponse struct {
	ICMBAddr uint16
	IPMBAddr uint8
	LUN      uint8
}

func (req *GetBridgeProxyAddressRequest) Command() Command {
	return CommandGetBridgeProxyAddress
}

func (req *GetBridgeProxyAddressRequest) Pack() []byte {
	return []byte{}
}

func (res *GetBridgeProxyAddressResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	res.ICMBAddr, _, _ = unpackUint16L(msg, 0)
	res.IPMBAddr, _, _ = unpackUint8(msg, 2)
	res.LUN, _, _ = unpackUint8(msg, 3)
	return nil
}

func (res *GetBridgeProxyAddressResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetBridgeProxyAddressResponse) Format() string {
	return fmt.Sprintf(`
			ICMB Address : %#04x
			IPMB Address : %#02x
			LUN          : %#02x
		`,
		res.ICMBAddr, res.IPMBAddr, res.LUN,
	)
}

func (c *Client) GetBridgeProxyAddress(ctx context.Context) (response *GetBridgeProxyAddressResponse, err error) {
	request := &GetBridgeProxyAddressRequest{}
	response = &GetBridgeProxyAddressResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
