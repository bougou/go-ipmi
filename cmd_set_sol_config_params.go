package ipmi

import (
	"context"
	"fmt"
)

// 26.2 Set SOL Configuration Parameters Command
type SetSOLConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector SOLConfigParamSelector
	ParamData     []byte
}

type SetSOLConfigParamResponse struct {
}

func (req *SetSOLConfigParamRequest) Command() Command {
	return CommandSetSOLConfigParam
}

func (req *SetSOLConfigParamRequest) Pack() []byte {
	out := make([]byte, 2+len(req.ParamData))
	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packBytes(req.ParamData, out, 2)
	return out
}

func (res *SetSOLConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetSOLConfigParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSOLConfigParamResponse) Format() string {
	return ""
}

func (c *Client) SetSOLConfigParam(ctx context.Context, channelNumber uint8, paramSelector SOLConfigParamSelector, paramData []byte) (response *SetSOLConfigParamResponse, err error) {
	request := &SetSOLConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &SetSOLConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSOLConfigParamFor(ctx context.Context, channelNumber uint8, param SOLConfigParameter) error {
	paramSelector, _, _ := param.SOLConfigParameter()
	paramData := param.Pack()

	_, err := c.SetSOLConfigParam(ctx, channelNumber, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSOLConfigParam failed, err: %w", err)
	}

	return nil
}
