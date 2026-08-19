package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) GetSystemInfoParam(ctx context.Context, paramSelector types.SystemInfoParamSelector, setSelector uint8, blockSelector uint8) (response *app.GetSystemInfoParamResponse, err error) {
	request := &app.GetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &app.GetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemInfoParamFor(ctx context.Context, param types.SystemInfoParameter) error {
	if types.IsNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, setSelector, blockSelector := param.SystemInfoParameter()
	response, err := c.GetSystemInfoParam(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetSystemInfoParams(ctx context.Context) (*types.SystemInfoParams, error) {
	systemInfo := &types.SystemInfoParams{
		SetInProgress:          &types.SystemInfoParam_SetInProgress{},
		SystemFirmwareVersions: make([]*types.SystemInfoParam_SystemFirmwareVersion, 0),
		SystemNames:            make([]*types.SystemInfoParam_SystemName, 0),
		PrimaryOSNames:         make([]*types.SystemInfoParam_PrimaryOSName, 0),
		OSNames:                make([]*types.SystemInfoParam_OSName, 0),
		OSVersions:             make([]*types.SystemInfoParam_OSVersion, 0),
		BMCURLs:                make([]*types.SystemInfoParam_BMCURL, 0),
		ManagementURLs:         make([]*types.SystemInfoParam_ManagementURL, 0),
	}

	if err := c.GetSystemInfoParamsFor(ctx, systemInfo); err != nil {
		return nil, err
	}

	return systemInfo, nil
}

func (c *Client) GetSystemInfoParamsFor(ctx context.Context, params *types.SystemInfoParams) error {
	if params == nil {
		return nil
	}

	canIgnore := buildCanIgnoreFn(
		0x80,
	)

	calculateSetsCount := func(blockData []byte) uint8 {

		if len(blockData) < 2 {
			return 0
		}

		stringLength := uint8(blockData[1])
		totalLength := 2 + stringLength

		return (totalLength-1)/16 + 1
	}

	if err := c.GetSystemInfoParamFor(ctx, params.SetInProgress); canIgnore(err) != nil {
		return err
	}

	if params.SystemFirmwareVersions != nil {
		if len(params.SystemFirmwareVersions) == 0 {
			p := &types.SystemInfoParam_SystemFirmwareVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.SystemFirmwareVersions = make([]*types.SystemInfoParam_SystemFirmwareVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_SystemFirmwareVersion{
					SetSelector: i,
				}
				params.SystemFirmwareVersions[i] = p
			}
		}

		for _, param := range params.SystemFirmwareVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.SystemNames != nil {
		if len(params.SystemNames) == 0 {
			p := &types.SystemInfoParam_SystemName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.SystemNames = make([]*types.SystemInfoParam_SystemName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_SystemName{
					SetSelector: i,
				}
				params.SystemNames[i] = p
			}
		}

		for _, param := range params.SystemNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.PrimaryOSNames != nil {
		if len(params.PrimaryOSNames) == 0 {
			p := &types.SystemInfoParam_PrimaryOSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.PrimaryOSNames = make([]*types.SystemInfoParam_PrimaryOSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_PrimaryOSName{
					SetSelector: i,
				}
				params.PrimaryOSNames[i] = p
			}
		}

		for _, param := range params.PrimaryOSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSNames != nil {
		if len(params.OSNames) == 0 {
			p := &types.SystemInfoParam_OSName{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.OSNames = make([]*types.SystemInfoParam_OSName, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_OSName{
					SetSelector: i,
				}
				params.OSNames[i] = p
			}
		}

		for _, param := range params.OSNames {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.OSVersions != nil {
		if len(params.OSVersions) == 0 {
			p := &types.SystemInfoParam_OSVersion{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.OSVersions = make([]*types.SystemInfoParam_OSVersion, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_OSVersion{
					SetSelector: i,
				}
				params.OSVersions[i] = p
			}
		}

		for _, param := range params.OSVersions {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.BMCURLs != nil {
		if len(params.BMCURLs) == 0 {
			p := &types.SystemInfoParam_BMCURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.BMCURLs = make([]*types.SystemInfoParam_BMCURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_BMCURL{
					SetSelector: i,
				}
				params.BMCURLs[i] = p
			}
		}

		for _, param := range params.BMCURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	if params.ManagementURLs != nil {
		if len(params.ManagementURLs) == 0 {
			p := &types.SystemInfoParam_ManagementURL{
				SetSelector: 0,
			}
			if err := c.GetSystemInfoParamFor(ctx, p); canIgnore(err) != nil {
				return err
			}
			setsCount := calculateSetsCount(p.BlockData)
			if setsCount == 0 {
				return nil
			}

			params.ManagementURLs = make([]*types.SystemInfoParam_ManagementURL, setsCount)
			for i := uint8(0); i < setsCount; i++ {
				p := &types.SystemInfoParam_ManagementURL{
					SetSelector: i,
				}
				params.ManagementURLs[i] = p
			}
		}

		for _, param := range params.ManagementURLs {
			if err := c.GetSystemInfoParamFor(ctx, param); canIgnore(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	systemInfoParams, err := c.GetSystemInfoParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetSystemInfo failed, err: %w", err)
	}
	return systemInfoParams.ToSystemInfo(), nil
}

func (c *Client) SetSystemInfoParam(ctx context.Context, paramSelector types.SystemInfoParamSelector, paramData []byte) (response *app.SetSystemInfoParamResponse, err error) {
	request := &app.SetSystemInfoParamRequest{
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &app.SetSystemInfoParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSystemInfoParamFor(ctx context.Context, param types.SystemInfoParameter) error {
	if types.IsNilSystemInfoParamete(param) {
		return nil
	}

	paramSelector, _, _ := param.SystemInfoParameter()
	paramData := param.Pack()
	_, err := c.SetSystemInfoParam(ctx, paramSelector, paramData)
	if err != nil {
		return fmt.Errorf("SetSystemInfoParam for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}
