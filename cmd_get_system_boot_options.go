package ipmi

import (
	"context"
	"fmt"
)

// 28.13 Get System Boot Options Command
type GetSystemBootOptionsParamRequest struct {
	ParamSelector BootOptionParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

// Table 28-14, Boot Option Parameters

type GetSystemBootOptionsParamResponse struct {
	ParameterVersion uint8

	// [7] - 1b = mark parameter invalid / locked
	// 0b = mark parameter valid / unlocked
	ParameterInValid bool
	// [6:0] - boot option parameter selector
	ParamSelector BootOptionParamSelector

	ParamData []byte // origin parameter data
}

func (req *GetSystemBootOptionsParamRequest) Pack() []byte {
	out := make([]byte, 3)
	packUint8(uint8(req.ParamSelector), out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)
	return out
}

func (req *GetSystemBootOptionsParamRequest) Command() Command {
	return CommandGetSystemBootOptions
}

func (res *GetSystemBootOptionsParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemBootOptionsParamResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}
	res.ParameterVersion, _, _ = unpackUint8(msg, 0)
	b, _, _ := unpackUint8(msg, 1)
	res.ParameterInValid = isBit7Set(b)
	res.ParamSelector = BootOptionParamSelector(b & 0x7f) // clear bit 7

	res.ParamData, _, _ = unpackBytes(msg, 2, len(msg)-2)
	return nil
}

func (res *GetSystemBootOptionsParamResponse) Format() string {

	var paramDataFormatted string

	var param BootOptionParameter

	switch res.ParamSelector {
	case BootOptionParamSelector_SetInProgress:
		param = &BootOptionParam_SetInProgress{}
	case BootOptionParamSelector_ServicePartitionSelector:
		param = &BootOptionParam_ServicePartitionSelector{}
	case BootOptionParamSelector_ServicePartitionScan:
		param = &BootOptionParam_ServicePartitionScan{}
	case BootOptionParamSelector_BMCBootFlagValidBitClear:
		param = &BootOptionParam_BMCBootFlagValidBitClear{}
	case BootOptionParamSelector_BootInfoAcknowledge:
		param = &BootOptionParam_BootInfoAcknowledge{}
	case BootOptionParamSelector_BootFlags:
		param = &BootOptionParam_BootFlags{}
	case BootOptionParamSelector_BootInitiatorInfo:
		param = &BootOptionParam_BootInitiatorInfo{}
	case BootOptionParamSelector_BootInitiatorMailbox:
		param = &BootOptionParam_BootInitiatorMailbox{}
	}

	if param != nil {
		if err := param.Unpack(res.ParamData); err == nil {
			paramDataFormatted = param.Format()
		}
	}

	return fmt.Sprintf(`Boot parameter version: %d
Boot parameter %d is %s
Boot parameter data: %02x
  %s : %s`,
		res.ParameterVersion,
		res.ParamSelector, formatBool(res.ParameterInValid, "invalid/locked", "valid/unlocked"),
		res.ParamData,
		res.ParamSelector.String(),
		paramDataFormatted,
	)
}

func (c *Client) GetSystemBootOptionsParam(ctx context.Context, paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) (response *GetSystemBootOptionsParamResponse, err error) {
	request := &GetSystemBootOptionsParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &GetSystemBootOptionsParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemBootOptionsParamFor(ctx context.Context, param BootOptionParameter) error {
	if isNilBootOptionParameter(param) {
		return nil
	}
	paramSelector, setSelector, blockSelector := param.BootOptionParameter()

	response, err := c.GetSystemBootOptionsParam(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemBootOptions for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

// GetSystemBootOptionsParams get all parameters of boot options.
func (c *Client) GetSystemBootOptionsParams(ctx context.Context) (*BootOptionsParams, error) {
	bootOptionsParams := &BootOptionsParams{
		SetInProgress:            &BootOptionParam_SetInProgress{},
		ServicePartitionSelector: &BootOptionParam_ServicePartitionSelector{},
		ServicePartitionScan:     &BootOptionParam_ServicePartitionScan{},
		BMCBootFlagValidBitClear: &BootOptionParam_BMCBootFlagValidBitClear{},
		BootInfoAcknowledge:      &BootOptionParam_BootInfoAcknowledge{},
		BootFlags:                &BootOptionParam_BootFlags{},
		BootInitiatorInfo:        &BootOptionParam_BootInitiatorInfo{},
		BootInitiatorMailbox:     &BootOptionParam_BootInitiatorMailbox{},
	}

	if err := c.GetSystemBootOptionsParamsFor(ctx, bootOptionsParams); err != nil {
		return nil, fmt.Errorf("GetBootOptionsFor failed, err: %w", err)
	}

	return bootOptionsParams, nil
}

func (c *Client) GetSystemBootOptionsParamsFor(ctx context.Context, bootOptionsParams *BootOptionsParams) error {
	if bootOptionsParams == nil {
		return nil
	}

	if bootOptionsParams.SetInProgress != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.SetInProgress); err != nil {
			return err
		}
	}

	if bootOptionsParams.ServicePartitionSelector != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.ServicePartitionSelector); err != nil {
			return err
		}
	}

	if bootOptionsParams.ServicePartitionScan != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.ServicePartitionScan); err != nil {
			return err
		}
	}

	if bootOptionsParams.BMCBootFlagValidBitClear != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BMCBootFlagValidBitClear); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootInfoAcknowledge != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInfoAcknowledge); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootFlags != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootFlags); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootInitiatorInfo != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInitiatorInfo); err != nil {
			return err
		}

	}

	if bootOptionsParams.BootInitiatorMailbox != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInitiatorMailbox); err != nil {
			return err
		}
	}
	return nil
}
