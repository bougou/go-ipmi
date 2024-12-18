package ipmi

import "context"

// 20.3 Warm Reset Command
type WarmResetRequest struct {
	// empty
}

type WarmResetResponse struct {
}

func (req *WarmResetRequest) Command() Command {
	return CommandWarmReset
}

func (req *WarmResetRequest) Pack() []byte {
	return []byte{}
}

func (res *WarmResetResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *WarmResetResponse) Unpack(msg []byte) error {
	return nil
}

func (res *WarmResetResponse) Format() string {
	return ""
}

func (c *Client) WarmReset(ctx context.Context) (err error) {
	request := &WarmResetRequest{}
	response := &WarmResetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
