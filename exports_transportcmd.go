package ipmi

import transportcmd "github.com/bougou/go-ipmi/pkg/cmd/transport"

const (
	ChannelSecurityKeysLockStatus_Locked      = transportcmd.ChannelSecurityKeysLockStatus_Locked
	ChannelSecurityKeysLockStatus_NotLockable = transportcmd.ChannelSecurityKeysLockStatus_NotLockable
	ChannelSecurityKeysLockStatus_Unlocked    = transportcmd.ChannelSecurityKeysLockStatus_Unlocked
	ChannelSecurityKeysOperationLock          = transportcmd.ChannelSecurityKeysOperationLock
	ChannelSecurityKeysOperationRead          = transportcmd.ChannelSecurityKeysOperationRead
	ChannelSecurityKeysOperationSet           = transportcmd.ChannelSecurityKeysOperationSet
	PayloadEncryptionOperationReinitialize    = transportcmd.PayloadEncryptionOperationReinitialize
	PayloadEncryptionOperationResume          = transportcmd.PayloadEncryptionOperationResume
	PayloadEncryptionOperationSuspend         = transportcmd.PayloadEncryptionOperationSuspend
	SharedSerialAlertBehavior_Defer           = transportcmd.SharedSerialAlertBehavior_Defer
	SharedSerialAlertBehavior_Fail            = transportcmd.SharedSerialAlertBehavior_Fail
	SharedSerialAlertBehavior_Success         = transportcmd.SharedSerialAlertBehavior_Success
)

type (
	ActivatePayloadRequest                 = transportcmd.ActivatePayloadRequest
	ActivatePayloadResponse                = transportcmd.ActivatePayloadResponse
	ChannelSecurityKeysLockStatus          = transportcmd.ChannelSecurityKeysLockStatus
	ChannelSecurityKeysOperation           = transportcmd.ChannelSecurityKeysOperation
	DeactivatePayloadRequest               = transportcmd.DeactivatePayloadRequest
	DeactivatePayloadResponse              = transportcmd.DeactivatePayloadResponse
	GetChannelOEMPayloadInfoRequest        = transportcmd.GetChannelOEMPayloadInfoRequest
	GetChannelOEMPayloadInfoResponse       = transportcmd.GetChannelOEMPayloadInfoResponse
	GetChannelPayloadSupportRequest        = transportcmd.GetChannelPayloadSupportRequest
	GetChannelPayloadSupportResponse       = transportcmd.GetChannelPayloadSupportResponse
	GetChannelPayloadVersionRequest        = transportcmd.GetChannelPayloadVersionRequest
	GetChannelPayloadVersionResponse       = transportcmd.GetChannelPayloadVersionResponse
	GetIPStatisticsRequest                 = transportcmd.GetIPStatisticsRequest
	GetIPStatisticsResponse                = transportcmd.GetIPStatisticsResponse
	GetLanConfigParamRequest               = transportcmd.GetLanConfigParamRequest
	GetLanConfigParamResponse              = transportcmd.GetLanConfigParamResponse
	GetPayloadActivationStatusRequest      = transportcmd.GetPayloadActivationStatusRequest
	GetPayloadActivationStatusResponse     = transportcmd.GetPayloadActivationStatusResponse
	GetPayloadInstanceInfoRequest          = transportcmd.GetPayloadInstanceInfoRequest
	GetPayloadInstanceInfoResponse         = transportcmd.GetPayloadInstanceInfoResponse
	GetSOLConfigParamRequest               = transportcmd.GetSOLConfigParamRequest
	GetSOLConfigParamResponse              = transportcmd.GetSOLConfigParamResponse
	PayloadEncryptionOperation             = transportcmd.PayloadEncryptionOperation
	SOLActivatingRequest                   = transportcmd.SOLActivatingRequest
	SOLActivatingResponse                  = transportcmd.SOLActivatingResponse
	SetChannelSecurityKeysRequest          = transportcmd.SetChannelSecurityKeysRequest
	SetChannelSecurityKeysResponse         = transportcmd.SetChannelSecurityKeysResponse
	SetLanConfigParamRequest               = transportcmd.SetLanConfigParamRequest
	SetLanConfigParamResponse              = transportcmd.SetLanConfigParamResponse
	SetSOLConfigParamRequest               = transportcmd.SetSOLConfigParamRequest
	SetSOLConfigParamResponse              = transportcmd.SetSOLConfigParamResponse
	SharedSerialAlertBehavior              = transportcmd.SharedSerialAlertBehavior
	SuspendARPsRequest                     = transportcmd.SuspendARPsRequest
	SuspendARPsResponse                    = transportcmd.SuspendARPsResponse
	SuspendResumePayloadEncryptionRequest  = transportcmd.SuspendResumePayloadEncryptionRequest
	SuspendResumePayloadEncryptionResponse = transportcmd.SuspendResumePayloadEncryptionResponse
)
