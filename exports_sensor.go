package ipmi

import sensor "github.com/bougou/go-ipmi/pkg/cmd/sensor"

const (
	AlertImmediateOperationClearStatus   = sensor.AlertImmediateOperationClearStatus
	AlertImmediateOperationGetStatus     = sensor.AlertImmediateOperationGetStatus
	AlertImmediateOperationInitiateAlert = sensor.AlertImmediateOperationInitiateAlert
	AlertImmediateOperationReserved      = sensor.AlertImmediateOperationReserved
	AlertImmediateStatusFailedRetry      = sensor.AlertImmediateStatusFailedRetry
	AlertImmediateStatusFailedWaitACK    = sensor.AlertImmediateStatusFailedWaitACK
	AlertImmediateStatusInProgress       = sensor.AlertImmediateStatusInProgress
	AlertImmediateStatusNoStatus         = sensor.AlertImmediateStatusNoStatus
	AlertImmediateStatusNormalEnd        = sensor.AlertImmediateStatusNormalEnd
	SetSensorEventEnableModeDisable      = sensor.SetSensorEventEnableModeDisable
	SetSensorEventEnableModeEnable       = sensor.SetSensorEventEnableModeEnable
	SetSensorEventEnableModeNoChange     = sensor.SetSensorEventEnableModeNoChange
)

type (
	AlertImmediateOperation                = sensor.AlertImmediateOperation
	AlertImmediateRequest                  = sensor.AlertImmediateRequest
	AlertImmediateResponse                 = sensor.AlertImmediateResponse
	AlertImmediateStatus                   = sensor.AlertImmediateStatus
	ArmPEFPostponeTimerRequest             = sensor.ArmPEFPostponeTimerRequest
	ArmPEFPostponeTimerResponse            = sensor.ArmPEFPostponeTimerResponse
	ClearMessageFlagsRequest               = sensor.ClearMessageFlagsRequest
	ClearMessageFlagsResponse              = sensor.ClearMessageFlagsResponse
	GetBMCGlobalEnablesRequest             = sensor.GetBMCGlobalEnablesRequest
	GetBMCGlobalEnablesResponse            = sensor.GetBMCGlobalEnablesResponse
	GetEventReceiverRequest                = sensor.GetEventReceiverRequest
	GetEventReceiverResponse               = sensor.GetEventReceiverResponse
	GetLastProcessedEventIdRequest         = sensor.GetLastProcessedEventIdRequest
	GetLastProcessedEventIdResponse        = sensor.GetLastProcessedEventIdResponse
	GetMessageFlagsRequest                 = sensor.GetMessageFlagsRequest
	GetMessageFlagsResponse                = sensor.GetMessageFlagsResponse
	GetMessageRequest                      = sensor.GetMessageRequest
	GetMessageResponse                     = sensor.GetMessageResponse
	GetPEFCapabilitiesRequest              = sensor.GetPEFCapabilitiesRequest
	GetPEFCapabilitiesResponse             = sensor.GetPEFCapabilitiesResponse
	GetPEFConfigParamRequest               = sensor.GetPEFConfigParamRequest
	GetPEFConfigParamResponse              = sensor.GetPEFConfigParamResponse
	GetSensorEventEnableRequest            = sensor.GetSensorEventEnableRequest
	GetSensorEventEnableResponse           = sensor.GetSensorEventEnableResponse
	GetSensorEventStatusRequest            = sensor.GetSensorEventStatusRequest
	GetSensorEventStatusResponse           = sensor.GetSensorEventStatusResponse
	GetSensorHysteresisRequest             = sensor.GetSensorHysteresisRequest
	GetSensorHysteresisResponse            = sensor.GetSensorHysteresisResponse
	GetSensorReadingFactorsRequest         = sensor.GetSensorReadingFactorsRequest
	GetSensorReadingFactorsResponse        = sensor.GetSensorReadingFactorsResponse
	GetSensorReadingRequest                = sensor.GetSensorReadingRequest
	GetSensorReadingResponse               = sensor.GetSensorReadingResponse
	GetSensorThresholdsRequest             = sensor.GetSensorThresholdsRequest
	GetSensorThresholdsResponse            = sensor.GetSensorThresholdsResponse
	GetSensorTypeRequest                   = sensor.GetSensorTypeRequest
	GetSensorTypeResponse                  = sensor.GetSensorTypeResponse
	PETAcknowledgeRequest                  = sensor.PETAcknowledgeRequest
	PETAcknowledgeResponse                 = sensor.PETAcknowledgeResponse
	PlatformEventMessageRequest            = sensor.PlatformEventMessageRequest
	PlatformEventMessageResponse           = sensor.PlatformEventMessageResponse
	ReadEventMessageBufferRequest          = sensor.ReadEventMessageBufferRequest
	ReadEventMessageBufferResponse         = sensor.ReadEventMessageBufferResponse
	RearmSensorEventsRequest               = sensor.RearmSensorEventsRequest
	RearmSensorEventsResponse              = sensor.RearmSensorEventsResponse
	SetBMCGlobalEnablesRequest             = sensor.SetBMCGlobalEnablesRequest
	SetBMCGlobalEnablesResponse            = sensor.SetBMCGlobalEnablesResponse
	SetEventReceiverRequest                = sensor.SetEventReceiverRequest
	SetEventReceiverResponse               = sensor.SetEventReceiverResponse
	SetLastProcessedEventIdRequest         = sensor.SetLastProcessedEventIdRequest
	SetLastProcessedEventIdResponse        = sensor.SetLastProcessedEventIdResponse
	SetPEFConfigParamRequest               = sensor.SetPEFConfigParamRequest
	SetPEFConfigParamResponse              = sensor.SetPEFConfigParamResponse
	SetSensorEventEnableMode               = sensor.SetSensorEventEnableMode
	SetSensorEventEnableRequest            = sensor.SetSensorEventEnableRequest
	SetSensorEventEnableResponse           = sensor.SetSensorEventEnableResponse
	SetSensorHysteresisRequest             = sensor.SetSensorHysteresisRequest
	SetSensorHysteresisResponse            = sensor.SetSensorHysteresisResponse
	SetSensorReadingAndEventStatusRequest  = sensor.SetSensorReadingAndEventStatusRequest
	SetSensorReadingAndEventStatusResponse = sensor.SetSensorReadingAndEventStatusResponse
	SetSensorThresholdsRequest             = sensor.SetSensorThresholdsRequest
	SetSensorThresholdsResponse            = sensor.SetSensorThresholdsResponse
	SetSensorTypeRequest                   = sensor.SetSensorTypeRequest
	SetSensorTypeResponse                  = sensor.SetSensorTypeResponse
)
