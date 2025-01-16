package ipmi

import "context"

type BridgeState uint8

const (
	BridgeState_Disabled  BridgeState = 0
	BridgeState_Enabled   BridgeState = 1
	BridgeState_Assigning BridgeState = 2
	BridgeState_Error     BridgeState = 3
)

func (s BridgeState) String() string {
	m := map[BridgeState]string{
		0: "disabled",
		1: "enabled",
		2: "assigning",
		3: "error",
	}
	if s, ok := m[s]; ok {
		return s
	}
	return "unknown"
}

type GetBridgeStateRequest struct {
}

type GetBridgeStateResponse struct {
	BridgeState BridgeState
}

func (req *GetBridgeStateRequest) Command() Command {
	return CommandGetBridgeState
}

func (req *GetBridgeStateRequest) Pack() []byte {
	return []byte{}
}

func (res *GetBridgeStateResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}
	res.BridgeState = BridgeState(msg[0])
	return nil
}

func (res *GetBridgeStateResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetBridgeStateResponse) Format() string {
	return res.BridgeState.String()
}

func (c *Client) GetBridgeState(ctx context.Context) (response *GetBridgeStateResponse, err error) {
	request := &GetBridgeStateRequest{}
	response = &GetBridgeStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
