package ipmi

import (
	"context"
	"fmt"
)

// 22.14a Set System Info Parameters Command
type SetSystemInfoParamRequest struct {
	ParamSelector SystemInfoParamSelector
	ParamData     []byte
}

type SetSystemInfoParamResponse struct {
}

func (req *SetSystemInfoParamRequest) Pack() []byte {
	out := make([]byte, 1+len(req.ParamData))
	out[0] = byte(req.ParamSelector)
	packBytes(req.ParamData, out, 1)
	return out
}

func (req *SetSystemInfoParamRequest) Command() Command {
	return CommandSetSystemInfoParam
}

func (res *SetSystemInfoParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
	}
}

func (res *SetSystemInfoParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemInfoParamResponse) Format() string {
	return ""
}

func (c *Client) SetSystemInfoParam(ctx context.Context, paramSelector SystemInfoParamSelector, paramData []byte) (response *SetSystemInfoParamResponse, err error) {
	request := &SetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &SetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSystemInfoParamFor(ctx context.Context, param SystemInfoParameter) error {
	if isNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, _, _ := param.SystemInfoParameter()
	paramData := param.Pack()
	_, err := c.SetSystemInfoParam(ctx, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}
