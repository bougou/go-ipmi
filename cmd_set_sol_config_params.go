package ipmi

import (
	"context"
	"fmt"
)

// 26.2 Set SOL Configuration Parameters Command
type SetSOLConfigParamsRequest struct {
	ChannelNumber uint8
	ParamSelector SOLConfigParamSelector
	ParamData     []byte
}

type SetSOLConfigParamsResponse struct {
}

func (req *SetSOLConfigParamsRequest) Command() Command {
	return CommandSetSOLConfigParams
}

func (req *SetSOLConfigParamsRequest) Pack() []byte {
	out := make([]byte, 2+len(req.ParamData))
	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packBytes(req.ParamData, out, 2)
	return out
}

func (res *SetSOLConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetSOLConfigParamsResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSOLConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) SetSOLConfigParams(ctx context.Context, channelNumber uint8, paramSelector SOLConfigParamSelector, paramData []byte) (response *SetSOLConfigParamsResponse, err error) {
	request := &SetSOLConfigParamsRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &SetSOLConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSOLConfigParamsFor(ctx context.Context, channelNumber uint8, param SOLConfigParameter) error {
	paramSelector, _, _ := param.SOLConfigParameter()
	paramData := param.Pack()

	_, err := c.SetSOLConfigParams(ctx, channelNumber, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSOLConfigParams failed, err: %s", err)
	}

	return nil
}
