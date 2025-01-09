package ipmi

import (
	"context"
	"fmt"
)

// 28.13 Get System Boot Options Command
type GetSystemBootOptionsRequest struct {
	ParamSelector BootOptionParamSelector
	SetSelector   uint8
	BlockSelector uint8
}

// Table 28-14, Boot Option Parameters

type GetSystemBootOptionsResponse struct {
	ParameterVersion uint8

	// [7] - 1b = mark parameter invalid / locked
	// 0b = mark parameter valid / unlocked
	ParameterInValid bool
	// [6:0] - boot option parameter selector
	ParamSelector BootOptionParamSelector

	ParamData []byte // origin parameter data

	Parameter BootOptionParameter
}

func (req *GetSystemBootOptionsRequest) Pack() []byte {
	out := make([]byte, 3)
	packUint8(uint8(req.ParamSelector), out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)
	return out
}

func (req *GetSystemBootOptionsRequest) Command() Command {
	return CommandGetSystemBootOptions
}

func (res *GetSystemBootOptionsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetSystemBootOptionsResponse) Unpack(msg []byte) error {
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

func (res *GetSystemBootOptionsResponse) Format() string {

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

func (c *Client) GetSystemBootOptions(ctx context.Context, paramSelector BootOptionParamSelector, setSelector uint8, blockSelector uint8) (response *GetSystemBootOptionsResponse, err error) {
	request := &GetSystemBootOptionsRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &GetSystemBootOptionsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemBootOptionsFor(ctx context.Context, param BootOptionParameter) error {
	paramSelector, setSelector, blockSelector := param.BootOptionParameter()

	response, err := c.GetSystemBootOptions(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemBootOptions for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

// GetBootOptions get all parameters for boot options.
func (c *Client) GetBootOptions(ctx context.Context) (*BootOptions, error) {
	bootOptions := &BootOptions{
		SetInProgress:            &BootOptionParam_SetInProgress{},
		ServicePartitionSelector: &BootOptionParam_ServicePartitionSelector{},
		ServicePartitionScan:     &BootOptionParam_ServicePartitionScan{},
		BMCBootFlagValidBitClear: &BootOptionParam_BMCBootFlagValidBitClear{},
		BootInfoAcknowledge:      &BootOptionParam_BootInfoAcknowledge{},
		BootFlags:                &BootOptionParam_BootFlags{},
		BootInitiatorInfo:        &BootOptionParam_BootInitiatorInfo{},
		BootInitiatorMailbox:     &BootOptionParam_BootInitiatorMailbox{},
	}

	if err := c.GetBootOptionsFor(ctx, bootOptions); err != nil {
		return nil, fmt.Errorf("GetBootOptionsFor failed, err: %s", err)
	}

	return bootOptions, nil
}

func (c *Client) GetBootOptionsFor(ctx context.Context, bootOptions *BootOptions) error {
	if bootOptions == nil {
		return nil
	}

	if bootOptions.SetInProgress != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.SetInProgress); err != nil {
			return err
		}
	}

	if bootOptions.ServicePartitionSelector != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.ServicePartitionSelector); err != nil {
			return err
		}
	}

	if bootOptions.ServicePartitionScan != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.ServicePartitionScan); err != nil {
			return err
		}
	}

	if bootOptions.BMCBootFlagValidBitClear != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.BMCBootFlagValidBitClear); err != nil {
			return err
		}
	}

	if bootOptions.BootInfoAcknowledge != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.BootInfoAcknowledge); err != nil {
			return err
		}
	}

	if bootOptions.BootFlags != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.BootFlags); err != nil {
			return err
		}
	}

	if bootOptions.BootInitiatorInfo != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.BootInitiatorInfo); err != nil {
			return err
		}

	}

	if bootOptions.BootInitiatorMailbox != nil {
		if err := c.GetSystemBootOptionsFor(ctx, bootOptions.BootInitiatorMailbox); err != nil {
			return err
		}
	}
	return nil
}
