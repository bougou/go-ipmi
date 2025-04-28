package ipmi

import (
	"context"
	"fmt"
)

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParamRequest struct {
	// [7] - 1b = get parameter revision only. 0b = get parameter
	// [6:0] - Parameter selector
	GetParamRevisionOnly bool
	ParamSelector        PEFConfigParamSelector

	SetSelector   uint8 // 00h if parameter does not require a Set Selector
	BlockSelector uint8 // 00h if parameter does not require a block number
}

type GetPEFConfigParamResponse struct {
	// Parameter revision.
	//
	// Format:
	//  - MSN = present revision.
	//  - LSN = oldest revision parameter is backward compatible with.
	//  - 11h for parameters in this specification.
	ParamRevision uint8

	// ParamData not returned when GetParamRevisionOnly is true
	ParamData []byte
}

func (req *GetPEFConfigParamRequest) Command() Command {
	return CommandGetPEFConfigParam
}

func (req *GetPEFConfigParamRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	b0 := uint8(req.ParamSelector) & 0x3f
	if req.GetParamRevisionOnly {
		b0 = setBit7(b0)
	}
	packUint8(b0, out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)

	return out
}

func (res *GetPEFConfigParamResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.ParamRevision = msg[0]

	if len(msg) > 1 {
		res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParamResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParamResponse) Format() string {
	return "" +
		fmt.Sprintf("Parameter Revision           : %#02x (%d)\n", res.ParamRevision, res.ParamRevision) +
		fmt.Sprintf("Configuration Parameter Data : %# 02x\n", res.ParamData)
}

func (c *Client) GetPEFConfigParam(ctx context.Context, getRevisionOnly bool, paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetPEFConfigParamResponse, err error) {
	request := &GetPEFConfigParamRequest{
		GetParamRevisionOnly: getRevisionOnly,
		ParamSelector:        paramSelector,
		SetSelector:          setSelector,
		BlockSelector:        blockSelector,
	}
	response = &GetPEFConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFConfigParamFor(ctx context.Context, param PEFConfigParameter) error {
	if isNilPEFConfigParameter(param) {
		return nil
	}

	paramSelector, setSelector, blockSelector := param.PEFConfigParameter()

	res, err := c.GetPEFConfigParam(ctx, false, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetPEFConfigParameters for param (%s) failed, err: %w", paramSelector, err)
	}

	if err := param.Unpack(res.ParamData); err != nil {
		return fmt.Errorf("unpack failed for param (%s), err: %w", paramSelector, err)
	}

	return nil
}

func (c *Client) GetPEFConfigParams(ctx context.Context) (pefConfigParams *PEFConfigParams, err error) {
	pefConfigParams = &PEFConfigParams{
		SetInProgress:       &PEFConfigParam_SetInProgress{},
		Control:             &PEFConfigParam_Control{},
		ActionGlobalControl: &PEFConfigParam_ActionGlobalControl{},
		StartupDelay:        &PEFConfigParam_StartupDelay{},
		AlertStartupDelay:   &PEFConfigParam_AlertStartupDelay{},
		EventFiltersCount:   &PEFConfigParam_EventFiltersCount{},
		EventFilters:        []*PEFConfigParam_EventFilter{},
		EventFiltersData1:   []*PEFConfigParam_EventFilterData1{},
		AlertPoliciesCount:  &PEFConfigParam_AlertPoliciesCount{},
		AlertPolicies:       []*PEFConfigParam_AlertPolicy{},
		SystemGUID:          &PEFConfigParam_SystemGUID{},
		AlertStringsCount:   &PEFConfigParam_AlertStringsCount{},
		AlertStringKeys:     []*PEFConfigParam_AlertStringKey{},
		AlertStrings:        []*PEFConfigParam_AlertString{},
		// GroupControlsCount:  &PEFConfigParam_GroupControlsCount{},
		// GroupControls:       []*PEFConfigParam_GroupControl{},
	}

	if err = c.GetPEFConfigParamsFor(ctx, pefConfigParams); err != nil {
		return nil, fmt.Errorf("GetPEFConfigParamsFor failed, err: %w", err)
	}

	return pefConfigParams, nil
}

func (c *Client) GetPEFConfigParamsFor(ctx context.Context, pefConfigParams *PEFConfigParams) error {
	if pefConfigParams == nil {
		return nil
	}

	if pefConfigParams.SetInProgress != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.SetInProgress); err != nil {
			return err
		}
	}

	if pefConfigParams.Control != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.Control); err != nil {
			return err
		}
	}

	if pefConfigParams.ActionGlobalControl != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.ActionGlobalControl); err != nil {
			return err
		}
	}

	if pefConfigParams.StartupDelay != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.StartupDelay); err != nil {
			return err
		}
	}

	if pefConfigParams.AlertStartupDelay != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.AlertStartupDelay); err != nil {
			return err
		}
	}

	eventFiltersCount := uint8(0)
	if pefConfigParams.EventFiltersCount != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.AlertPoliciesCount); err != nil {
			return err
		}
		eventFiltersCount = pefConfigParams.EventFiltersCount.Value
	}

	if pefConfigParams.EventFilters != nil {
		if len(pefConfigParams.EventFilters) == 0 && eventFiltersCount > 0 {
			pefConfigParams.EventFilters = make([]*PEFConfigParam_EventFilter, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfigParams.EventFilters[i] = &PEFConfigParam_EventFilter{
					SetSelector: i + 1,
				}
			}
		}

		for _, eventFilter := range pefConfigParams.EventFilters {
			if err := c.GetPEFConfigParamFor(ctx, eventFilter); err != nil {
				return err
			}
		}
	}

	if pefConfigParams.EventFiltersData1 != nil {
		if len(pefConfigParams.EventFiltersData1) == 0 && eventFiltersCount > 0 {
			pefConfigParams.EventFiltersData1 = make([]*PEFConfigParam_EventFilterData1, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfigParams.EventFiltersData1[i] = &PEFConfigParam_EventFilterData1{
					SetSelector: i + 1,
				}
			}
		}

		for _, eventFilterData1 := range pefConfigParams.EventFiltersData1 {
			if err := c.GetPEFConfigParamFor(ctx, eventFilterData1); err != nil {
				return err
			}
		}
	}

	alertPoliciesCount := uint8(0)
	if pefConfigParams.AlertPoliciesCount != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.AlertPoliciesCount); err != nil {
			return err
		}
		alertPoliciesCount = pefConfigParams.AlertPoliciesCount.Value
	}

	if pefConfigParams.AlertPolicies != nil {
		if len(pefConfigParams.AlertPolicies) == 0 && alertPoliciesCount > 0 {
			pefConfigParams.AlertPolicies = make([]*PEFConfigParam_AlertPolicy, alertPoliciesCount)
			for i := uint8(0); i < alertPoliciesCount; i++ {
				pefConfigParams.AlertPolicies[i] = &PEFConfigParam_AlertPolicy{
					SetSelector: i + 1,
				}
			}
		}

		for _, alertPolicy := range pefConfigParams.AlertPolicies {
			if err := c.GetPEFConfigParamFor(ctx, alertPolicy); err != nil {
				return err
			}
		}
	}

	if pefConfigParams.SystemGUID != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.SystemGUID); err != nil {
			return err
		}
	}

	alertStringsCount := uint8(0)
	if pefConfigParams.AlertStringsCount != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.AlertStringsCount); err != nil {
			return err
		}
	}

	if pefConfigParams.AlertStringKeys != nil {
		if len(pefConfigParams.AlertStringKeys) == 0 && alertStringsCount > 0 {
			pefConfigParams.AlertStringKeys = make([]*PEFConfigParam_AlertStringKey, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfigParams.AlertStringKeys[i] = &PEFConfigParam_AlertStringKey{
					SetSelector: i,
				}
			}
		}

		for _, alertStringKey := range pefConfigParams.AlertStringKeys {
			if err := c.GetPEFConfigParamFor(ctx, alertStringKey); err != nil {
				return err
			}
		}
	}

	if pefConfigParams.AlertStrings != nil {
		if len(pefConfigParams.AlertStrings) == 0 && alertStringsCount > 0 {
			pefConfigParams.AlertStrings = make([]*PEFConfigParam_AlertString, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfigParams.AlertStrings[i] = &PEFConfigParam_AlertString{
					SetSelector: i,
				}
			}
		}

		for _, alertString := range pefConfigParams.AlertStrings {
			if err := c.GetPEFConfigParamFor(ctx, alertString); err != nil {
				return err
			}
		}
	}

	groupControlsCount := uint8(0)
	if pefConfigParams.GroupControlsCount != nil {
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.GroupControlsCount); err != nil {
			return err
		}
	}

	if pefConfigParams.GroupControls != nil {
		if len(pefConfigParams.GroupControls) == 0 && groupControlsCount > 0 {
			pefConfigParams.GroupControls = make([]*PEFConfigParam_GroupControl, groupControlsCount)
			for i := uint8(0); i < groupControlsCount; i++ {
				pefConfigParams.GroupControls[i] = &PEFConfigParam_GroupControl{
					SetSelector: i,
				}
			}
		}

		for _, groupControl := range pefConfigParams.GroupControls {
			if err := c.GetPEFConfigParamFor(ctx, groupControl); err != nil {
				return err
			}
		}
	}

	return nil
}
