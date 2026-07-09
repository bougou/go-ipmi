package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/cmd/sensor"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) SetLastProcessedEventId(ctx context.Context, recordID uint16, byBMC bool) (response *sensor.SetLastProcessedEventIdResponse, err error) {
	request := &sensor.SetLastProcessedEventIdRequest{
		ByBMC:    byBMC,
		RecordID: recordID,
	}
	response = &sensor.SetLastProcessedEventIdResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) PlatformEventMessage(ctx context.Context, request *sensor.PlatformEventMessageRequest) (response *sensor.PlatformEventMessageResponse, err error) {

	response = &sensor.PlatformEventMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command provides a mechanism for setting the hysteresis values associated
// with the thresholds of a sensor that has threshold based event generation.
func (c *Client) SetSensorHysteresis(ctx context.Context, sensorNumber uint8, positiveHysteresis uint8, negativeHysteresis uint8) (response *sensor.SetSensorHysteresisResponse, err error) {
	request := &sensor.SetSensorHysteresisRequest{
		SensorNumber:       sensorNumber,
		PositiveHysteresis: positiveHysteresis,
		NegativeHysteresis: negativeHysteresis,
	}
	response = &sensor.SetSensorHysteresisResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFCapabilities(ctx context.Context) (response *sensor.GetPEFCapabilitiesResponse, err error) {
	request := &sensor.GetPEFCapabilitiesRequest{}
	response = &sensor.GetPEFCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetEventReceiver(ctx context.Context, slaveAddress uint8, lun uint8) (response *sensor.SetEventReceiverResponse, err error) {
	request := &sensor.SetEventReceiverRequest{
		SlaveAddress: slaveAddress,
		LUN:          lun,
	}
	response = &sensor.SetEventReceiverResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetEventReceiverDisable(ctx context.Context, lun uint8) (response *sensor.SetEventReceiverResponse, err error) {
	return c.SetEventReceiver(ctx, 0xff, lun)
}

func (c *Client) SetSensorType(ctx context.Context, sensorNumber uint8, sensorType types.SensorType, eventReadingType types.EventReadingType) (response *sensor.SetSensorTypeResponse, err error) {
	request := &sensor.SetSensorTypeRequest{
		SensorNumber:     sensorNumber,
		SensorType:       sensorType,
		EventReadingType: eventReadingType,
	}
	response = &sensor.SetSensorTypeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetPEFConfigParam(ctx context.Context, paramSelector types.PEFConfigParamSelector, paramData []byte) (response *sensor.SetPEFConfigParamResponse, err error) {
	request := &sensor.SetPEFConfigParamRequest{
		ParamSelector: paramSelector,
		ParamData:     paramData,
	}
	response = &sensor.SetPEFConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSensorReadingAndEventStatus(ctx context.Context, request *sensor.SetSensorReadingAndEventStatusRequest) (response *sensor.SetSensorReadingAndEventStatusResponse, err error) {
	response = &sensor.SetSensorReadingAndEventStatusResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) PETAcknowledge(ctx context.Context, request *sensor.PETAcknowledgeRequest) (response *sensor.PETAcknowledgeResponse, err error) {
	response = &sensor.PETAcknowledgeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) AlertImmediate(ctx context.Context, request *sensor.AlertImmediateRequest) (response *sensor.AlertImmediateResponse, err error) {
	response = &sensor.AlertImmediateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) RearmSensorEvents(ctx context.Context, request *sensor.RearmSensorEventsRequest) (response *sensor.RearmSensorEventsResponse, err error) {
	response = &sensor.RearmSensorEventsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSensorType(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorTypeResponse, err error) {
	request := &sensor.GetSensorTypeRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorTypeResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetBMCGlobalEnables(ctx context.Context, enableSystemEventLogging bool, enableEventMessageBuffer bool, enableEventMessageBufferFullInterrupt bool, enableReceiveMessageQueueInterrupt bool) (response *sensor.SetBMCGlobalEnablesResponse, err error) {
	getRes, err := c.GetBMCGlobalEnables(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetBMCGlobalEnables failed, err: %w", err)
	}

	request := &sensor.SetBMCGlobalEnablesRequest{
		EnableOEM2: getRes.OEM2Enabled,
		EnableOEM1: getRes.OEM1Enabled,
		EnableOEM0: getRes.OEM0Enabled,

		EnableSystemEventLogging:              enableSystemEventLogging,
		EnableEventMessageBuffer:              enableEventMessageBuffer,
		EnableEventMessageBufferFullInterrupt: enableEventMessageBufferFullInterrupt,
		EnableReceiveMessageQueueInterrupt:    enableReceiveMessageQueueInterrupt,
	}
	response = &sensor.SetBMCGlobalEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ArmPEFPostponeTimer(ctx context.Context, timeout uint8) (response *sensor.ArmPEFPostponeTimerResponse, err error) {
	request := &sensor.ArmPEFPostponeTimerRequest{
		Timeout: timeout,
	}
	response = &sensor.ArmPEFPostponeTimerResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetLastProcessedEventId(ctx context.Context) (response *sensor.GetLastProcessedEventIdResponse, err error) {
	request := &sensor.GetLastProcessedEventIdRequest{}
	response = &sensor.GetLastProcessedEventIdResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command retrieves the present hysteresis values for the specified sensor.
// If the sensor hysteresis values are "fixed", then the hysteresis values can be obtained from the SDR for the sensor.
func (c *Client) GetSensorHysteresis(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorHysteresisResponse, err error) {
	request := &sensor.GetSensorHysteresisRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorHysteresisResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetEventReceiver(ctx context.Context) (response *sensor.GetEventReceiverResponse, err error) {
	request := &sensor.GetEventReceiverRequest{}
	response = &sensor.GetEventReceiverResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetBMCGlobalEnables(ctx context.Context) (response *sensor.GetBMCGlobalEnablesResponse, err error) {
	request := &sensor.GetBMCGlobalEnablesRequest{}
	response = &sensor.GetBMCGlobalEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSensorEventEnable(ctx context.Context, request *sensor.SetSensorEventEnableRequest) (response *sensor.SetSensorEventEnableResponse, err error) {
	response = &sensor.SetSensorEventEnableResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetMessageFlags(ctx context.Context) (response *sensor.GetMessageFlagsResponse, err error) {
	request := &sensor.GetMessageFlagsRequest{}
	response = &sensor.GetMessageFlagsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command returns the Sensor Reading Factors fields for the specified reading value on the specified sensor.
func (c *Client) GetSensorReadingFactors(ctx context.Context, sensorNumber uint8, reading uint8) (response *sensor.GetSensorReadingFactorsResponse, err error) {
	request := &sensor.GetSensorReadingFactorsRequest{
		SensorNumber: sensorNumber,
		Reading:      reading,
	}
	response = &sensor.GetSensorReadingFactorsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetMessage(ctx context.Context) (response *sensor.GetMessageResponse, err error) {
	request := &sensor.GetMessageRequest{}
	response = &sensor.GetMessageResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ReadEventMessageBuffer(ctx context.Context) (response *sensor.ReadEventMessageBufferResponse, err error) {
	request := &sensor.ReadEventMessageBufferRequest{}
	response = &sensor.ReadEventMessageBufferResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSensorEventStatus(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorEventStatusResponse, err error) {
	request := &sensor.GetSensorEventStatusRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorEventStatusResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFConfigParam(ctx context.Context, getRevisionOnly bool, paramSelector types.PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *sensor.GetPEFConfigParamResponse, err error) {
	request := &sensor.GetPEFConfigParamRequest{
		GetParamRevisionOnly: getRevisionOnly,
		ParamSelector:        paramSelector,
		SetSelector:          setSelector,
		BlockSelector:        blockSelector,
	}
	response = &sensor.GetPEFConfigParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetPEFConfigParamFor(ctx context.Context, param types.PEFConfigParameter) error {
	if types.IsNilPEFConfigParameter(param) {
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

func (c *Client) GetPEFConfigParams(ctx context.Context) (pefConfigParams *types.PEFConfigParams, err error) {
	pefConfigParams = &types.PEFConfigParams{
		SetInProgress:       &types.PEFConfigParam_SetInProgress{},
		Control:             &types.PEFConfigParam_Control{},
		ActionGlobalControl: &types.PEFConfigParam_ActionGlobalControl{},
		StartupDelay:        &types.PEFConfigParam_StartupDelay{},
		AlertStartupDelay:   &types.PEFConfigParam_AlertStartupDelay{},
		EventFiltersCount:   &types.PEFConfigParam_EventFiltersCount{},
		EventFilters:        []*types.PEFConfigParam_EventFilter{},
		EventFiltersData1:   []*types.PEFConfigParam_EventFilterData1{},
		AlertPoliciesCount:  &types.PEFConfigParam_AlertPoliciesCount{},
		AlertPolicies:       []*types.PEFConfigParam_AlertPolicy{},
		SystemGUID:          &types.PEFConfigParam_SystemGUID{},
		AlertStringsCount:   &types.PEFConfigParam_AlertStringsCount{},
		AlertStringKeys:     []*types.PEFConfigParam_AlertStringKey{},
		AlertStrings:        []*types.PEFConfigParam_AlertString{},
		GroupControlsCount:  &types.PEFConfigParam_GroupControlsCount{},
		GroupControls:       []*types.PEFConfigParam_GroupControl{},
	}

	if err = c.GetPEFConfigParamsFor(ctx, pefConfigParams); err != nil {
		return nil, fmt.Errorf("GetPEFConfigParamsFor failed, err: %w", err)
	}

	return pefConfigParams, nil
}

func (c *Client) GetPEFConfigParamsFor(ctx context.Context, pefConfigParams *types.PEFConfigParams) error {
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
		if err := c.GetPEFConfigParamFor(ctx, pefConfigParams.EventFiltersCount); err != nil {
			return err
		}
		eventFiltersCount = pefConfigParams.EventFiltersCount.Value
	}

	if pefConfigParams.EventFilters != nil {
		if len(pefConfigParams.EventFilters) == 0 && eventFiltersCount > 0 {
			pefConfigParams.EventFilters = make([]*types.PEFConfigParam_EventFilter, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfigParams.EventFilters[i] = &types.PEFConfigParam_EventFilter{
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
			pefConfigParams.EventFiltersData1 = make([]*types.PEFConfigParam_EventFilterData1, eventFiltersCount)
			for i := uint8(0); i < eventFiltersCount; i++ {
				pefConfigParams.EventFiltersData1[i] = &types.PEFConfigParam_EventFilterData1{
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
			pefConfigParams.AlertPolicies = make([]*types.PEFConfigParam_AlertPolicy, alertPoliciesCount)
			for i := uint8(0); i < alertPoliciesCount; i++ {
				pefConfigParams.AlertPolicies[i] = &types.PEFConfigParam_AlertPolicy{
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
		alertStringsCount = pefConfigParams.AlertStringsCount.Value
	}

	if pefConfigParams.AlertStringKeys != nil {
		if len(pefConfigParams.AlertStringKeys) == 0 && alertStringsCount > 0 {
			pefConfigParams.AlertStringKeys = make([]*types.PEFConfigParam_AlertStringKey, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfigParams.AlertStringKeys[i] = &types.PEFConfigParam_AlertStringKey{
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
			pefConfigParams.AlertStrings = make([]*types.PEFConfigParam_AlertString, alertStringsCount)
			for i := uint8(0); i < alertStringsCount; i++ {
				pefConfigParams.AlertStrings[i] = &types.PEFConfigParam_AlertString{
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
		groupControlsCount = pefConfigParams.GroupControlsCount.Value
	}

	if pefConfigParams.GroupControls != nil {
		if len(pefConfigParams.GroupControls) == 0 && groupControlsCount > 0 {
			pefConfigParams.GroupControls = make([]*types.PEFConfigParam_GroupControl, groupControlsCount)
			for i := uint8(0); i < groupControlsCount; i++ {
				pefConfigParams.GroupControls[i] = &types.PEFConfigParam_GroupControl{
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

// SetSensorThresholds is to set the specified threshold for the given sensor.
// Note that the application issuing this command is responsible for ensuring that
// thresholds for a sensor are set in the proper order (e.g. that
// the upper critical threshold is set higher than the upper non-critical threshold)
//
//	Upper Non Recoverable area
//	-----------------UNR threshold
//	Upper Critical area
//	-----------------UCR threshold
//	Upper Non Critical area
//	-----------------UNC threshold
//	OK area
//	-----------------LNC threshold
//	Lower Non Critical area
//	-----------------LCR threshold
//	Lower Critical area
//	-----------------LNR threshold
//	Lower NonRecoverable area
func (c *Client) SetSensorThresholds(ctx context.Context, request *sensor.SetSensorThresholdsRequest) (response *sensor.SetSensorThresholdsResponse, err error) {
	response = &sensor.SetSensorThresholdsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSensorReading(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorReadingResponse, err error) {
	request := &sensor.GetSensorReadingRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorReadingResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ClearMessageFlags(ctx context.Context, request *sensor.ClearMessageFlagsRequest) (response *sensor.ClearMessageFlagsResponse, err error) {
	response = &sensor.ClearMessageFlagsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command retrieves the threshold for the given sensor.
func (c *Client) GetSensorThresholds(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorThresholdsResponse, err error) {
	request := &sensor.GetSensorThresholdsRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorThresholdsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSensorEventEnable(ctx context.Context, sensorNumber uint8) (response *sensor.GetSensorEventEnableResponse, err error) {
	request := &sensor.GetSensorEventEnableRequest{
		SensorNumber: sensorNumber,
	}
	response = &sensor.GetSensorEventEnableResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
