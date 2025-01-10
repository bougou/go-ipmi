package ipmi

import (
	"context"
	"fmt"
)

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParamsRequest struct {
	// [7] - 1b = get parameter revision only. 0b = get parameter
	// [6:0] - Parameter selector
	GetRevisionOnly bool
	ParamSelector   PEFConfigParamSelector

	SetSelector   uint8 // 00h if parameter does not require a Set Selector
	BlockSelector uint8 // 00h if parameter does not require a block number
}

type GetPEFConfigParamsResponse struct {
	// Parameter revision.
	//
	// Format:
	//  - MSN = present revision.
	//  - LSN = oldest revision parameter is backward compatible with.
	//  - 11h for parameters in this specification.
	Revision uint8

	// ParamData not returned when GetRevisionOnly is true
	ParamData []byte
}

func (req *GetPEFConfigParamsRequest) Command() Command {
	return CommandGetPEFConfigParams
}

func (req *GetPEFConfigParamsRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	b0 := uint8(req.ParamSelector) & 0x3f
	if req.GetRevisionOnly {
		b0 = setBit7(b0)
	}
	packUint8(b0, out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)

	return out
}

func (res *GetPEFConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.Revision = msg[0]

	if len(msg) > 1 {
		res.ParamData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParamsResponse) Format() string {
	return fmt.Sprintf(`
Parameter Revision           : %#02x (%d)
Configuration Parameter Data : %# 02x`,
		res.Revision, res.Revision,
		res.ParamData,
	)
}

func (c *Client) GetPEFConfigParams(ctx context.Context, getRevisionOnly bool, paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetPEFConfigParamsResponse, err error) {
	request := &GetPEFConfigParamsRequest{
		GetRevisionOnly: getRevisionOnly,
		ParamSelector:   paramSelector,
		SetSelector:     setSelector,
		BlockSelector:   blockSelector,
	}
	response = &GetPEFConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFConfigParamsFor(ctx context.Context, param PEFConfigParameter) error {
	paramSelector, setSelector, blockSelector := param.PEFConfigParameter()

	res, err := c.GetPEFConfigParams(ctx, false, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetPEFConfigParameters for param (%s) failed, err: %s", paramSelector, err)
	}

	if err := param.Unpack(res.ParamData); err != nil {
		return fmt.Errorf("unpack failed for param (%s), err: %s", paramSelector, err)
	}

	return nil
}

func (c *Client) GetPEFConfig(ctx context.Context) (pefConfig *PEFConfig, err error) {
	pefConfig = &PEFConfig{
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

	if err = c.GetPEFConfigFor(ctx, pefConfig); err != nil {
		return nil, fmt.Errorf("GetPEFConfig failed, err: %s", err)
	}

	return pefConfig, nil
}

func (c *Client) GetPEFConfigFor(ctx context.Context, pefConfig *PEFConfig) error {
	if pefConfig == nil {
		return nil
	}

	if pefConfig.SetInProgress != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.SetInProgress); err != nil {
			return err
		}
	}

	if pefConfig.Control != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.Control); err != nil {
			return err
		}
	}

	if pefConfig.ActionGlobalControl != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.ActionGlobalControl); err != nil {
			return err
		}
	}

	if pefConfig.StartupDelay != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.StartupDelay); err != nil {
			return err
		}
	}

	if pefConfig.AlertStartupDelay != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.AlertStartupDelay); err != nil {
			return err
		}
	}

	eventFiltersCount := uint8(0)
	if pefConfig.EventFiltersCount != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.AlertPoliciesCount); err != nil {
			return err
		}
		eventFiltersCount = pefConfig.EventFiltersCount.Value
	}

	if pefConfig.EventFilters != nil {
		if len(pefConfig.EventFilters) == 0 && eventFiltersCount > 0 {
			pefConfig.EventFilters = make([]*PEFConfigParam_EventFilter, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfig.EventFilters[i] = &PEFConfigParam_EventFilter{
					SetSelector: i + 1,
				}
			}
		}

		for _, eventFilter := range pefConfig.EventFilters {
			if err := c.GetPEFConfigParamsFor(ctx, eventFilter); err != nil {
				return err
			}
		}
	}

	if pefConfig.EventFiltersData1 != nil {
		if len(pefConfig.EventFiltersData1) == 0 && eventFiltersCount > 0 {
			pefConfig.EventFiltersData1 = make([]*PEFConfigParam_EventFilterData1, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfig.EventFiltersData1[i] = &PEFConfigParam_EventFilterData1{
					SetSelector: i + 1,
				}
			}
		}

		for _, eventFilterData1 := range pefConfig.EventFiltersData1 {
			if err := c.GetPEFConfigParamsFor(ctx, eventFilterData1); err != nil {
				return err
			}
		}
	}

	alertPoliciesCount := uint8(0)
	if pefConfig.AlertPoliciesCount != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.AlertPoliciesCount); err != nil {
			return err
		}
		alertPoliciesCount = pefConfig.AlertPoliciesCount.Value
	}

	if pefConfig.AlertPolicies != nil {
		if len(pefConfig.AlertPolicies) == 0 && alertPoliciesCount > 0 {
			pefConfig.AlertPolicies = make([]*PEFConfigParam_AlertPolicy, alertPoliciesCount)
			for i := uint8(0); i < alertPoliciesCount; i++ {
				pefConfig.AlertPolicies[i] = &PEFConfigParam_AlertPolicy{
					SetSelector: i + 1,
				}
			}
		}

		for _, alertPolicy := range pefConfig.AlertPolicies {
			if err := c.GetPEFConfigParamsFor(ctx, alertPolicy); err != nil {
				return err
			}
		}
	}

	if pefConfig.SystemGUID != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.SystemGUID); err != nil {
			return err
		}
	}

	alertStringsCount := uint8(0)
	if pefConfig.AlertStringsCount != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.AlertStringsCount); err != nil {
			return err
		}
	}

	if pefConfig.AlertStringKeys != nil {
		if len(pefConfig.AlertStringKeys) == 0 && alertStringsCount > 0 {
			pefConfig.AlertStringKeys = make([]*PEFConfigParam_AlertStringKey, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfig.AlertStringKeys[i] = &PEFConfigParam_AlertStringKey{
					SetSelector: i,
				}
			}
		}

		for _, alertStringKey := range pefConfig.AlertStringKeys {
			if err := c.GetPEFConfigParamsFor(ctx, alertStringKey); err != nil {
				return err
			}
		}
	}

	if pefConfig.AlertStrings != nil {
		if len(pefConfig.AlertStrings) == 0 && alertStringsCount > 0 {
			pefConfig.AlertStrings = make([]*PEFConfigParam_AlertString, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfig.AlertStrings[i] = &PEFConfigParam_AlertString{
					SetSelector: i,
				}
			}
		}

		for _, alertString := range pefConfig.AlertStrings {
			if err := c.GetPEFConfigParamsFor(ctx, alertString); err != nil {
				return err
			}
		}
	}

	groupControlsCount := uint8(0)
	if pefConfig.GroupControlsCount != nil {
		if err := c.GetPEFConfigParamsFor(ctx, pefConfig.GroupControlsCount); err != nil {
			return err
		}
	}

	if pefConfig.GroupControls != nil {
		if len(pefConfig.GroupControls) == 0 && groupControlsCount > 0 {
			pefConfig.GroupControls = make([]*PEFConfigParam_GroupControl, groupControlsCount)
			for i := uint8(0); i < groupControlsCount; i++ {
				pefConfig.GroupControls[i] = &PEFConfigParam_GroupControl{
					SetSelector: i,
				}
			}
		}

		for _, groupControl := range pefConfig.GroupControls {
			if err := c.GetPEFConfigParamsFor(ctx, groupControl); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) GetPEFConfig2(ctx context.Context) (pefConfig *PEFConfig, err error) {
	pefConfig = &PEFConfig{}

	{
		param := &PEFConfigParam_Control{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.Control = param
	}

	{
		param := &PEFConfigParam_ActionGlobalControl{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.ActionGlobalControl = param
	}

	{
		param := &PEFConfigParam_StartupDelay{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.StartupDelay = param
	}

	{
		param := &PEFConfigParam_AlertStartupDelay{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.AlertStartupDelay = param
	}

	{
		param := &PEFConfigParam_EventFiltersCount{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.EventFiltersCount = param
	}

	{
		for i := uint8(1); i <= pefConfig.EventFiltersCount.Value; i++ {
			param := &PEFConfigParam_EventFilter{}
			if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.EventFilters = append(pefConfig.EventFilters, param)
		}
	}

	{
		for i := uint8(1); i <= pefConfig.EventFiltersCount.Value; i++ {
			param := &PEFConfigParam_EventFilterData1{}
			if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.EventFiltersData1 = append(pefConfig.EventFiltersData1, param)
		}
	}

	{
		// ipmitool pef
		param := &PEFConfigParam_AlertPoliciesCount{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.AlertPoliciesCount = param
	}

	{
		for i := uint8(1); i < pefConfig.AlertPoliciesCount.Value; i++ {
			param := &PEFConfigParam_AlertPolicy{
				SetSelector: i,
			}

			if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertPolicies = append(pefConfig.AlertPolicies, param)
		}
	}

	{
		param := &PEFConfigParam_SystemGUID{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.SystemGUID = param
	}

	{
		param := &PEFConfigParam_AlertStringsCount{}
		if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
			return nil, err
		}
		pefConfig.AlertStringsCount = param
	}

	{
		for i := uint8(1); i <= pefConfig.AlertStringsCount.Value; i++ {
			param := &PEFConfigParam_AlertStringKey{
				SetSelector: i,
			}

			if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
				return nil, fmt.Errorf("get alert strings number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertStringKeys = append(pefConfig.AlertStringKeys, param)
		}
	}

	{
		for i := uint8(1); i < pefConfig.AlertStringsCount.Value; i++ {
			param := &PEFConfigParam_AlertString{
				SetSelector:   i,
				BlockSelector: uint8(1),
			}
			if err := c.GetPEFConfigParamsFor(ctx, param); err != nil {
				return nil, fmt.Errorf("get alert strings number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertStrings = append(pefConfig.AlertStrings, param)
		}
	}

	// {
	// 	param := &PEFConfigParam_NumberOfGroupControlTableEntries{}
	// 	if err := c.getPEFConfigFor(param); err != nil {
	// 		return nil, err
	// 	}
	// 	pefConfig.NumberOfGroupControlTableEntries = param
	// }

	// {
	// 	param := &PEFConfigParam_GroupControlTable{}
	// 	if err := c.getPEFConfigFor(param); err != nil {
	// 		return nil, err
	// 	}
	// 	pefConfig.GroupControlTable = param
	// }

	return pefConfig, nil
}
