package ipmi

import (
	"context"
	"fmt"
)

// 23.1 Set LAN Configuration Parameters Command
type SetLanConfigParamRequest struct {
	ChannelNumber uint8
	ParamSelector LanConfigParamSelector
	ParamData     []byte
}

type SetLanConfigParamResponse struct {
	// empty
}

func (req *SetLanConfigParamRequest) Pack() []byte {
	out := make([]byte, 2+len(req.ParamData))

	packUint8(req.ChannelNumber, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packBytes(req.ParamData, out, 2)

	return out
}

func (req *SetLanConfigParamRequest) Command() Command {
	return CommandSetLanConfigParam
}

func (res *SetLanConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported.",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state.",
		0x82: "attempt to write read-only parameter",
		0x83: "attempt to read write-only parameter",
	}
}

func (res *SetLanConfigParamResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetLanConfigParamResponse) Format() string {
	return ""
}

func (c *Client) SetLanConfigParam(ctx context.Context, channelNumber uint8, paramSelector LanConfigParamSelector, configData []byte) (response *SetLanConfigParamResponse, err error) {
	request := &SetLanConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		ParamData:     configData,
	}
	response = &SetLanConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetLanConfigParamFor(ctx context.Context, channelNumber uint8, param LanConfigParameter) error {
	paramSelector, _, _ := param.LanConfigParameter()
	c.DebugBytes(fmt.Sprintf(">> Set param data for (%s[%d]) ", paramSelector.String(), paramSelector), param.Pack(), 8)

	if _, err := c.SetLanConfigParam(ctx, channelNumber, paramSelector, param.Pack()); err != nil {
		c.Debugf("!!! Set LanConfigParam for paramSelector (%d) %s failed, err: %v\n", uint8(paramSelector), paramSelector, err)
		return err
	}

	return nil
}
