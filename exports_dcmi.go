package ipmi

import dcmi "github.com/bougou/go-ipmi/pkg/cmd/dcmi"

type (
	ActivateDCMIPowerLimitRequest           = dcmi.ActivateDCMIPowerLimitRequest
	ActivateDCMIPowerLimitResponse          = dcmi.ActivateDCMIPowerLimitResponse
	DCMITemperatureReading                  = dcmi.DCMITemperatureReading
	GetDCMIAssetTagRequest                  = dcmi.GetDCMIAssetTagRequest
	GetDCMIAssetTagResponse                 = dcmi.GetDCMIAssetTagResponse
	GetDCMICapParamRequest                  = dcmi.GetDCMICapParamRequest
	GetDCMICapParamResponse                 = dcmi.GetDCMICapParamResponse
	GetDCMIConfigParamRequest               = dcmi.GetDCMIConfigParamRequest
	GetDCMIConfigParamResponse              = dcmi.GetDCMIConfigParamResponse
	GetDCMIMgmtControllerIdentifierRequest  = dcmi.GetDCMIMgmtControllerIdentifierRequest
	GetDCMIMgmtControllerIdentifierResponse = dcmi.GetDCMIMgmtControllerIdentifierResponse
	GetDCMIPowerLimitRequest                = dcmi.GetDCMIPowerLimitRequest
	GetDCMIPowerLimitResponse               = dcmi.GetDCMIPowerLimitResponse
	GetDCMIPowerReadingRequest              = dcmi.GetDCMIPowerReadingRequest
	GetDCMIPowerReadingResponse             = dcmi.GetDCMIPowerReadingResponse
	GetDCMISensorInfoRequest                = dcmi.GetDCMISensorInfoRequest
	GetDCMISensorInfoResponse               = dcmi.GetDCMISensorInfoResponse
	GetDCMITemperatureReadingsRequest       = dcmi.GetDCMITemperatureReadingsRequest
	GetDCMITemperatureReadingsResponse      = dcmi.GetDCMITemperatureReadingsResponse
	GetDCMIThermalLimitRequest              = dcmi.GetDCMIThermalLimitRequest
	GetDCMIThermalLimitResponse             = dcmi.GetDCMIThermalLimitResponse
	SetDCMIAssetTagRequest                  = dcmi.SetDCMIAssetTagRequest
	SetDCMIAssetTagResponse                 = dcmi.SetDCMIAssetTagResponse
	SetDCMIConfigParamRequest               = dcmi.SetDCMIConfigParamRequest
	SetDCMIConfigParamResponse              = dcmi.SetDCMIConfigParamResponse
	SetDCMIMgmtControllerIdentifierRequest  = dcmi.SetDCMIMgmtControllerIdentifierRequest
	SetDCMIMgmtControllerIdentifierResponse = dcmi.SetDCMIMgmtControllerIdentifierResponse
	SetDCMIPowerLimitRequest                = dcmi.SetDCMIPowerLimitRequest
	SetDCMIPowerLimitResponse               = dcmi.SetDCMIPowerLimitResponse
	SetDCMIThermalLimitRequest              = dcmi.SetDCMIThermalLimitRequest
	SetDCMIThermalLimitResponse             = dcmi.SetDCMIThermalLimitResponse
)

var (
	FormatDCMITemperatureReadings = dcmi.FormatDCMITemperatureReadings
)
