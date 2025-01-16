package ipmi

import "context"

type SetBridgeStateRequest struct {
	BridgeState BridgeState
}

type SetBridgeStateResponse struct {
}

func (req *SetBridgeStateRequest) Command() Command {
	return CommandSetBridgeState
}

func (req *SetBridgeStateRequest) Pack() []byte {
	return []byte{uint8(req.BridgeState)}
}

func (res *SetBridgeStateResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetBridgeStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetBridgeStateResponse) Format() string {
	return ""
}

func (c *Client) SetBridgeState(ctx context.Context, bridgeState BridgeState) (response *SetBridgeStateResponse, err error) {
	request := &SetBridgeStateRequest{
		BridgeState: bridgeState,
	}
	response = &SetBridgeStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
