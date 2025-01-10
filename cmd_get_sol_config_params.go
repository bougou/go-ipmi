package ipmi

import (
	"context"
	"fmt"
)

// 26.3 Get SOL Configuration Parameters Command
type GetSOLConfigParamsRequest struct {
	GetParamRevisionOnly bool
	ChannelNumber        uint8
	ParamSelector        SOLConfigParamSelector
	SetSelector          uint8
	BlockSelector        uint8
}

type GetSOLConfigParamsResponse struct {
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetSOLConfigParamsRequest) Command() Command {
	return CommandGetSOLConfigParams
}

func (req *GetSOLConfigParamsRequest) Pack() []byte {
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

func (res *GetSOLConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetSOLConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 1)
	}

	res.ParamRevision = msg[0]
	if len(msg) > 1 {
		res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (res *GetSOLConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) GetSOLConfigParams(ctx context.Context, channelNumber uint8, paramSelector SOLConfigParamSelector, setSelector, blockSelector uint8) (response *GetSOLConfigParamsResponse, err error) {
	request := &GetSOLConfigParamsRequest{
		ChannelNumber: channelNumber,
		ParamSelector: paramSelector,
		SetSelector:   0x00,
		BlockSelector: 0x00,
	}
	response = &GetSOLConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSOLConfigParamsFor(ctx context.Context, channelNumber uint8, param SOLConfigParameter) error {
	paramSelector, setSelector, blockSelector := param.SOLConfigParameter()
	res, err := c.GetSOLConfigParams(ctx, channelNumber, paramSelector, setSelector, blockSelector)

	if err != nil {
		return fmt.Errorf("GetSOLConfigParams for param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(res.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%2d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSOLConfig(ctx context.Context, channelNumber uint8) (*SOLConfig, error) {
	solConfig := &SOLConfig{
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

	if err := c.GetSOLConfigFor(ctx, channelNumber, solConfig); err != nil {
		return nil, fmt.Errorf("GetSOLConfigParamsFor failed, err: %w", err)
	}

	return solConfig, nil
}

func (c *Client) GetSOLConfigFor(ctx context.Context, channelNumber uint8, solConfig *SOLConfig) error {
	if solConfig == nil {
		return nil
	}

	if solConfig.SetInProgress != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.SetInProgress); err != nil {
			return err
		}
	}

	if solConfig.SOLEnable != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.SOLEnable); err != nil {
			return err
		}
	}

	if solConfig.SOLAuthentication != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.SOLAuthentication); err != nil {
			return err
		}
	}

	if solConfig.Character != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.Character); err != nil {
			return err
		}
	}

	if solConfig.SOLRetry != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.SOLRetry); err != nil {
			return err
		}
	}

	if solConfig.NonVolatileBitRate != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.NonVolatileBitRate); err != nil {
			return err
		}
	}

	if solConfig.VolatileBitRate != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.VolatileBitRate); err != nil {
			return err
		}
	}

	if solConfig.PayloadChannel != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.PayloadChannel); err != nil {
			return err
		}
	}

	if solConfig.PayloadPort != nil {
		if err := c.GetSOLConfigParamsFor(ctx, channelNumber, solConfig.PayloadPort); err != nil {
			return err
		}
	}

	return nil
}
