package ipmi

import "context"

type SetBridgeProxyAddressRequest struct {
	ICMBAddr uint16
	IPMBAddr uint8
	LUN      uint8
}

type SetBridgeProxyAddressResponse struct {
}

func (req *SetBridgeProxyAddressRequest) Command() Command {
	return CommandSetBridgeProxyAddress
}

func (req *SetBridgeProxyAddressRequest) Pack() []byte {
	out := make([]byte, 4)
	packUint16L(req.ICMBAddr, out, 0)
	packUint8(req.IPMBAddr, out, 2)
	packUint8(req.LUN, out, 3)
	return out
}

func (res *SetBridgeProxyAddressResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetBridgeProxyAddressResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetBridgeProxyAddressResponse) Format() string {
	return ""
}

func (c *Client) SetBridgeProxyAddress(ctx context.Context, icmbAddr uint16, ipmbAddr uint8, lun uint8) (response *SetBridgeProxyAddressResponse, err error) {
	request := &SetBridgeProxyAddressRequest{
		ICMBAddr: icmbAddr,
		IPMBAddr: ipmbAddr,
		LUN:      lun,
	}
	response = &SetBridgeProxyAddressResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
