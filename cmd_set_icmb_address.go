package ipmi

import "context"

type SetICMBAddressRequest struct {
	ICMBAddr uint16
}

type SetICMBAddressResponse struct {
}

func (req *SetICMBAddressRequest) Command() Command {
	return CommandSetICMBAddress
}

func (req *SetICMBAddressRequest) Pack() []byte {
	out := make([]byte, 2)
	packUint16L(req.ICMBAddr, out, 0)
	return out
}

func (res *SetICMBAddressResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetICMBAddressResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetICMBAddressResponse) Format() string {
	return ""
}

func (c *Client) SetICMBAddress(ctx context.Context, icmbAddr uint16) (response *SetICMBAddressResponse, err error) {
	request := &SetICMBAddressRequest{
		ICMBAddr: icmbAddr,
	}
	response = &SetICMBAddressResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
