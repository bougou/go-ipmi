package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/cmd/dcmi"
	"github.com/bougou/go-ipmi/pkg/types"
)

// SetDCMIPowerLimit sends a DCMI "Set Power Limit" command.
// See [SetDCMIPowerLimitRequest] for details.
func (c *Client) SetDCMIPowerLimit(ctx context.Context, request *dcmi.SetDCMIPowerLimitRequest) (response *dcmi.SetDCMIPowerLimitResponse, err error) {
	response = &dcmi.SetDCMIPowerLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetDCMISensorInfo sends a DCMI "Get Sensor Info" command.
// See [GetDCMISensorInfoRequest] for details.
func (c *Client) GetDCMISensorInfo(ctx context.Context, request *dcmi.GetDCMISensorInfoRequest) (response *dcmi.GetDCMISensorInfoResponse, err error) {
	response = &dcmi.GetDCMISensorInfoResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMISensors(ctx context.Context, entityIDs ...types.EntityID) ([]*types.SDR, error) {
	out := make([]*types.SDR, 0)

	for _, entityID := range entityIDs {
		request := &dcmi.GetDCMISensorInfoRequest{
			SensorType:          types.SensorTypeTemperature,
			EntityID:            entityID,
			EntityInstance:      0x00,
			EntityInstanceStart: 0,
		}

		response := &dcmi.GetDCMISensorInfoResponse{}
		if err := c.Exchange(ctx, request, response); err != nil {
			return nil, err
		}

		for _, recordID := range response.SDRRecordID {
			sdr, err := c.GetSDREnhanced(ctx, recordID)
			if err != nil {
				return nil, fmt.Errorf("GetSDRDetail failed for recordID (%#02x), err: %w", recordID, err)
			}
			out = append(out, sdr)
		}
	}

	return out, nil
}

func (c *Client) GetDCMIThermalLimit(ctx context.Context, entityID types.EntityID, entityInstance types.EntityInstance) (response *dcmi.GetDCMIThermalLimitResponse, err error) {
	if uint8(entityID) != 0x37 && uint8(entityID) != 0x40 {
		return nil, errors.New("only Inlet Temperature entityID (0x37 or 0x40) is supported")
	}
	request := &dcmi.GetDCMIThermalLimitRequest{
		EntityID:       entityID,
		EntityInstance: entityInstance,
	}
	response = &dcmi.GetDCMIThermalLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetDCMIPowerLimit sends a DCMI "Get Power Limit" command.
// See [GetDCMIPowerLimitRequest] for details.
func (c *Client) GetDCMIPowerLimit(ctx context.Context) (response *dcmi.GetDCMIPowerLimitResponse, err error) {
	request := &dcmi.GetDCMIPowerLimitRequest{}
	response = &dcmi.GetDCMIPowerLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetDCMIAssetTag(ctx context.Context, offset uint8, writeBytes uint8, assetTag []byte) (response *dcmi.SetDCMIAssetTagResponse, err error) {
	request := &dcmi.SetDCMIAssetTagRequest{
		Offset:     offset,
		WriteBytes: writeBytes,
		AssetTag:   assetTag,
	}
	response = &dcmi.SetDCMIAssetTagResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetDCMIAssetTagFull(ctx context.Context, assetTag []byte) (err error) {
	if len(assetTag) > 63 {
		return fmt.Errorf("the asset tag must be at most 63 bytes")
	}

	var offset uint8 = 0
	var writeBytes uint8 = 16
	if len(assetTag) < 16 {
		writeBytes = uint8(len(assetTag))
	}

	for {
		offsetEnd := offset + writeBytes
		_, err := c.SetDCMIAssetTag(ctx, offset, writeBytes, assetTag[offset:offsetEnd])
		if err != nil {
			return fmt.Errorf("SetDCMIAssetTag failed, err: %w", err)
		}

		offset = offset + writeBytes
		if offset >= uint8(len(assetTag)) {
			break
		}
		if offset+writeBytes > uint8(len(assetTag)) {
			writeBytes = uint8(len(assetTag)) - offset
		}
	}

	return nil
}

func (c *Client) GetDCMITemperatureReadings(ctx context.Context, request *dcmi.GetDCMITemperatureReadingsRequest) (response *dcmi.GetDCMITemperatureReadingsResponse, err error) {
	response = &dcmi.GetDCMITemperatureReadingsResponse{
		EntityID: request.EntityID,
	}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMITemperatureReadingsForEntities(ctx context.Context, entityIDs ...types.EntityID) ([]dcmi.DCMITemperatureReading, error) {
	out := make([]dcmi.DCMITemperatureReading, 0)

	for _, entityID := range entityIDs {
		request := &dcmi.GetDCMITemperatureReadingsRequest{
			SensorType:          types.SensorTypeTemperature,
			EntityID:            entityID,
			EntityInstance:      0x00,
			EntityInstanceStart: 0,
		}

		response, err := c.GetDCMITemperatureReadings(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("GetDCMITemperatureReadings failed for entityID (%#02x), err: %w", entityID, err)
		}

		out = append(out, response.TemperatureReadings...)
	}

	return out, nil
}

func (c *Client) SetDCMIThermalLimit(ctx context.Context, request *dcmi.SetDCMIThermalLimitRequest) (response *dcmi.SetDCMIThermalLimitResponse, err error) {
	response = &dcmi.SetDCMIThermalLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIConfigParam(ctx context.Context, paramSelector types.DCMIConfigParamSelector, setSelector uint8) (response *dcmi.GetDCMIConfigParamResponse, err error) {
	request := &dcmi.GetDCMIConfigParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
	}
	response = &dcmi.GetDCMIConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIConfigParamFor(ctx context.Context, param types.DCMIConfigParameter) error {
	if types.IsNilDCMIConfigParameter(param) {
		return nil
	}

	paramSelector, setSelector := param.DCMIConfigParameter()

	request := &dcmi.GetDCMIConfigParamRequest{ParamSelector: paramSelector, SetSelector: setSelector}
	response := &dcmi.GetDCMIConfigParamResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetDCMIConfigParams(ctx context.Context) (*types.DCMIConfigParams, error) {
	dcmiConfigParams := &types.DCMIConfigParams{
		ActivateDHCP:           &types.DCMIConfigParam_ActivateDHCP{},
		DiscoveryConfiguration: &types.DCMIConfigParam_DiscoveryConfiguration{},
		DHCPTiming1:            &types.DCMIConfigParam_DHCPTiming1{},
		DHCPTiming2:            &types.DCMIConfigParam_DHCPTiming2{},
		DHCPTiming3:            &types.DCMIConfigParam_DHCPTiming3{},
	}

	if err := c.GetDCMIConfigParamsFor(ctx, dcmiConfigParams); err != nil {
		return nil, err
	}

	return dcmiConfigParams, nil
}

func (c *Client) GetDCMIConfigParamsFor(ctx context.Context, dcmiConfigParams *types.DCMIConfigParams) error {
	if dcmiConfigParams == nil {
		return nil
	}

	if dcmiConfigParams.ActivateDHCP != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.ActivateDHCP); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DiscoveryConfiguration != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DiscoveryConfiguration); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming1 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming1); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming2 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming2); err != nil {
			return err
		}
	}

	if dcmiConfigParams.DHCPTiming3 != nil {
		if err := c.GetDCMIConfigParamFor(ctx, dcmiConfigParams.DHCPTiming3); err != nil {
			return err
		}
	}

	return nil
}

// GetDCMIMgmtControllerIdentifier sends a DCMI "Get Management Controller Identifier" command.
// See [GetDCMIMgmtControllerIdentifierRequest] for details.
func (c *Client) GetDCMIMgmtControllerIdentifier(ctx context.Context, offset uint8) (response *dcmi.GetDCMIMgmtControllerIdentifierResponse, err error) {
	request := &dcmi.GetDCMIMgmtControllerIdentifierRequest{Offset: offset}
	response = &dcmi.GetDCMIMgmtControllerIdentifierResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIMgmtControllerIdentifierFull(ctx context.Context) ([]byte, error) {
	id := make([]byte, 0)
	offset := uint8(0)
	for {
		resp, err := c.GetDCMIMgmtControllerIdentifier(ctx, offset)
		if err != nil {
			return nil, fmt.Errorf("GetDCMIMgmtControllerIdentifier failed, err: %w", err)
		}
		id = append(id, resp.IDStr...)
		if resp.IDStrLength <= offset+uint8(len(resp.IDStr)) {
			break
		}
		offset += uint8(len(resp.IDStr))
	}

	return id, nil
}

// ActivateDCMIPowerLimit activate or deactivate the power limit set.
// Setting the param 'activate' to true means to activate the power limit, false means to deactivate the power limit
func (c *Client) ActivateDCMIPowerLimit(ctx context.Context, activate bool) (response *dcmi.ActivateDCMIPowerLimitResponse, err error) {
	request := &dcmi.ActivateDCMIPowerLimitRequest{
		Activate: activate,
	}
	response = &dcmi.ActivateDCMIPowerLimitResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// GetDCMIPowerReading sends a DCMI "Get Power Reading" command.
// See [GetDCMIPowerReadingRequest] for details.
func (c *Client) GetDCMIPowerReading(ctx context.Context) (response *dcmi.GetDCMIPowerReadingResponse, err error) {
	request := &dcmi.GetDCMIPowerReadingRequest{}
	response = &dcmi.GetDCMIPowerReadingResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMICapParam(ctx context.Context, paramSelector types.DCMICapParamSelector) (response *dcmi.GetDCMICapParamResponse, err error) {
	request := &dcmi.GetDCMICapParamRequest{ParamSelector: paramSelector}
	response = &dcmi.GetDCMICapParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMICapParamFor(ctx context.Context, param types.DCMICapParameter) error {
	if types.IsNilDCMICapParameter(param) {
		return nil
	}

	paramSelector := param.DCMICapParameter()

	request := &dcmi.GetDCMICapParamRequest{ParamSelector: paramSelector}
	response := &dcmi.GetDCMICapParamResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param data failed, err: %w", err)
	}

	return nil
}

func (c *Client) GetDCMICapParams(ctx context.Context) (*types.DCMICapParams, error) {
	dcmiCapParams := &types.DCMICapParams{
		SupportedDCMICapabilities:               &types.DCMICapParam_SupportedDCMICapabilities{},
		MandatoryPlatformAttributes:             &types.DCMICapParam_MandatoryPlatformAttributes{},
		OptionalPlatformAttributes:              &types.DCMICapParam_OptionalPlatformAttributes{},
		ManageabilityAccessAttributes:           &types.DCMICapParam_ManageabilityAccessAttributes{},
		EnhancedSystemPowerStatisticsAttributes: &types.DCMICapParam_EnhancedSystemPowerStatisticsAttributes{},
	}

	if err := c.GetDCMICapParamsFor(ctx, dcmiCapParams); err != nil {
		return nil, err
	}

	return dcmiCapParams, nil
}

func (c *Client) GetDCMICapParamsFor(ctx context.Context, dcmiCapParams *types.DCMICapParams) error {
	if dcmiCapParams == nil {
		return nil
	}

	if dcmiCapParams.SupportedDCMICapabilities != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.SupportedDCMICapabilities); err != nil {
			return err
		}
	}

	if dcmiCapParams.MandatoryPlatformAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.MandatoryPlatformAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.OptionalPlatformAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.OptionalPlatformAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.ManageabilityAccessAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.ManageabilityAccessAttributes); err != nil {
			return err
		}
	}

	if dcmiCapParams.EnhancedSystemPowerStatisticsAttributes != nil {
		if err := c.GetDCMICapParamFor(ctx, dcmiCapParams.EnhancedSystemPowerStatisticsAttributes); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) SetDCMIMgmtControllerIdentifier(ctx context.Context, offset uint8, writeBytes uint8, idStr []byte) (response *dcmi.SetDCMIMgmtControllerIdentifierResponse, err error) {
	request := &dcmi.SetDCMIMgmtControllerIdentifierRequest{
		Offset:     offset,
		WriteBytes: writeBytes,
		IDStr:      idStr,
	}
	response = &dcmi.SetDCMIMgmtControllerIdentifierResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetDCMIMgmtControllerIdentifierFull(ctx context.Context, idStr []byte) (err error) {
	if len(idStr) > 63 {
		return fmt.Errorf("the id str must be at most 63 bytes")
	}

	if len(idStr) == 0 || idStr[len(idStr)-1] != 0x00 {
		idStr = append(idStr, 0x00)
	}

	var offset uint8 = 0
	var writeBytes uint8 = 16
	if len(idStr) < 16 {
		writeBytes = uint8(len(idStr))
	}

	for {
		offsetEnd := offset + writeBytes
		_, err := c.SetDCMIMgmtControllerIdentifier(ctx, offset, writeBytes, idStr[offset:offsetEnd])
		if err != nil {
			return fmt.Errorf("SetDCMIMgmtControllerIdentifier failed, err: %w", err)
		}

		offset = offset + writeBytes
		if offset >= uint8(len(idStr)) {
			break
		}
		if offset+writeBytes > uint8(len(idStr)) {
			writeBytes = uint8(len(idStr)) - offset
		}
	}

	return nil
}

func (c *Client) SetDCMIConfigParam(ctx context.Context, paramSelector types.DCMIConfigParamSelector, setSelector uint8, paramData []byte) (response *dcmi.SetDCMIConfigParamResponse, err error) {
	request := &dcmi.SetDCMIConfigParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		ParamData:     paramData,
	}
	response = &dcmi.SetDCMIConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetDCMIConfigParamFor(ctx context.Context, param types.DCMIConfigParameter) (response *dcmi.SetDCMIConfigParamResponse, err error) {
	if types.IsNilDCMIConfigParameter(param) {
		return nil, fmt.Errorf("param is nil")
	}

	paramSelector, setSelector := param.DCMIConfigParameter()
	paramData := param.Pack()

	response, err = c.SetDCMIConfigParam(ctx, paramSelector, setSelector, paramData)
	if err != nil {
		return nil, fmt.Errorf("SetDCMIConfigParam failed, err: %w", err)
	}

	return
}

// GetDCMIAssetTag sends a DCMI "Get Asset Tag" command.
// See [GetDCMIAssetTagRequest] for details.
func (c *Client) GetDCMIAssetTag(ctx context.Context, offset uint8) (response *dcmi.GetDCMIAssetTagResponse, err error) {
	request := &dcmi.GetDCMIAssetTagRequest{Offset: offset}
	response = &dcmi.GetDCMIAssetTagResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIAssetTagFull(ctx context.Context) (assetTagRaw []byte, typeLength types.TypeLength, err error) {
	assetTagRaw = make([]byte, 0)

	typeCode := uint8(0)
	offset := uint8(0)

	for {
		resp, err := c.GetDCMIAssetTag(ctx, offset)
		if err != nil {
			if respErr, ok := types.IsResponseError(err); ok {
				cc := uint8(respErr.CompletionCode())
				switch cc {

				case 0x80:
					typeCode = 0b00
				case 0x81:
					typeCode = 0b01
				case 0x82:
					typeCode = 0x10
				case 0x83:
					typeCode = 0x11
				default:
					return nil, 0, fmt.Errorf("GetDCMIAssetTag failed, err: %w", respErr)
				}
			} else {
				return nil, 0, fmt.Errorf("GetDCMIAssetTag failed, err: %w", err)
			}
		}

		assetTagRaw = append(assetTagRaw, resp.AssetTag...)
		if resp.TotalLength <= offset+uint8(len(resp.AssetTag)) {
			break
		}
		offset += uint8(len(resp.AssetTag))
	}

	typeLength = types.TypeLength(typeCode<<6 | uint8(len(assetTagRaw)))

	return
}
