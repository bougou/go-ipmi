package ipmi

import (
	"context"
	"fmt"
)

// 26.3 Get SOL Configuration Parameters Command
type GetSOLConfigParamRequest struct {
	GetParamRevisionOnly bool
	ChannelNumber        uint8
	ParamSelector        SOLConfigParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSOLConfigParamResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSOLConfigParamRequest) Command() Command {
	return CommandGetSOLConfigParam
}

func (req *GetSOLConfigParamRequest) Pack() []byte {
	out := make([]byte, 4)
	b := req.ChannelNumber
	if req.GetParamRevisionOnly {
		b = setBit7(b)
	}

	packUint8(b, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packUint8(req.BlockSelector, out, 3)
	return out
}

func (res *GetSOLConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSOLConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision = msg[0]
	if len(msg) > 1 {
		res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSOLConfigParamResponse) Format() string {
	return ""
}

func (c *Client) GetSOLConfigParam(ctx context.Context, channelNumber uint8, paramSelector SOLConfigParamSelector, setSelector, blockSelector uint8) (response *GetSOLConfigParamResponse, err error) {
	request := &GetSOLConfigParamRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   0x00,
		BlockSelector: 0x00,
	}
	response = &GetSOLConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSOLConfigParamFor(ctx context.Context, channelNumber uint8, param SOLConfigParameter) error {
	if isNilSOLConfigParameter(param) {
		return nil
	}
	paramSelector, setSelector, blockSelector := param.SOLConfigParameter()
	res, err := c.GetSOLConfigParam(ctx, channelNumber, paramSelector, setSelector, blockSelector)

	if err != nil {
		return fmt.Errorf("GetSOLConfigParam for param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(res.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSOLConfigParams(ctx context.Context, channelNumber uint8) (*SOLConfigParams, error) {
	solConfigParams := &SOLConfigParams{
		SetInProgress:      &SOLConfigParam_SetInProgress{},
		SOLEnable:          &SOLConfigParam_SOLEnable{},
		SOLAuthentication:  &SOLConfigParam_SOLAuthentication{},
		Character:          &SOLConfigParam_Character{},
		SOLRetry:           &SOLConfigParam_SOLRetry{},
		NonVolatileBitRate: &SOLConfigParam_NonVolatileBitRate{},
		VolatileBitRate:    &SOLConfigParam_VolatileBitRate{},
		PayloadChannel:     &SOLConfigParam_PayloadChannel{},
		PayloadPort:        &SOLConfigParam_PayloadPort{},
	}

	if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfigParams); err != nil {
		return nil, fmt.Errorf("GetSOLConfigParamFor failed, err: %w", err)
	}

	return solConfigParams, nil
}

func (c *Client) GetSOLConfigParamsFor(ctx context.Context, channelNumber uint8, solConfigParams *SOLConfigParams) error {
	if solConfigParams == nil {
		return nil
	}

	if solConfigParams.SetInProgress != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SetInProgress); err != nil {
			return err
		}
	}

	if solConfigParams.SOLEnable != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLEnable); err != nil {
			return err
		}
	}

	if solConfigParams.SOLAuthentication != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLAuthentication); err != nil {
			return err
		}
	}

	if solConfigParams.Character != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.Character); err != nil {
			return err
		}
	}

	if solConfigParams.SOLRetry != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.SOLRetry); err != nil {
			return err
		}
	}

	if solConfigParams.NonVolatileBitRate != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.NonVolatileBitRate); err != nil {
			return err
		}
	}

	if solConfigParams.VolatileBitRate != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.VolatileBitRate); err != nil {
			return err
		}
	}

	if solConfigParams.PayloadChannel != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.PayloadChannel); err != nil {
			return err
		}
	}

	if solConfigParams.PayloadPort != nil {
		if err := c.GetSOLConfigParamFor(ctx, channelNumber, solConfigParams.PayloadPort); err != nil {
			return err
		}
	}

	return nil
}
