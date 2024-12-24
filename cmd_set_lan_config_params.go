package ipmi

import (
	"context"
	"fmt"
)

// 23.1 Set LAN Configuration Parameters Command
type SetLanConfigParamsRequest struct {
	ChannelNumber uint8
	ParamSelector LanConfigParamSelector
	ConfigData    []byte
}

type SetLanConfigParamsResponse struct {
	// empty
}

func (req *SetLanConfigParamsRequest) Pack() []byte {
	return []byte{}
}

func (req *SetLanConfigParamsRequest) Command() Command {
	return CommandSetLanConfigParams
}

func (res *SetLanConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetLanConfigParamsResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetLanConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) SetLanConfigParams(ctx context.Context, channelNumber uint8, paramSelector LanConfigParamSelector, configData []byte) (response *SetLanConfigParamsResponse, err error) {
	request := &SetLanConfigParamsRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ConfigData:    configData,
	}
	response = &SetLanConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetLanConfigParamsFor(ctx context.Context, channelNumber uint8, param LanConfigParameter) error {
	paramSelector, _, _ := param.LanConfigParamSelector()
	c.DebugBytes(fmt.Sprintf(">> Set param data for (%s[%d]) ", paramSelector.String(), paramSelector), param.Pack(), 8)

	if _, err := c.SetLanConfigParams(ctx, channelNumber, paramSelector, param.Pack()); err != nil {
		c.Debugf("!!! Set LanConfigParam for paramSelector (%d) %s failed, err: %v\n", uint8(paramSelector), paramSelector, err)
		return err
	}

	return nil
}
