package ipmi

import chassis "github.com/bougou/go-ipmi/pkg/cmd/chassis"

const (
	ChassisControlDiagnosticInterrupt = chassis.ChassisControlDiagnosticInterrupt
	ChassisControlHardReset           = chassis.ChassisControlHardReset
	ChassisControlPowerCycle          = chassis.ChassisControlPowerCycle
	ChassisControlPowerDown           = chassis.ChassisControlPowerDown
	ChassisControlPowerUp             = chassis.ChassisControlPowerUp
	ChassisControlSoftShutdown        = chassis.ChassisControlSoftShutdown
	ChassisIdentifyStateIndefiniteOn  = chassis.ChassisIdentifyStateIndefiniteOn
	ChassisIdentifyStateOff           = chassis.ChassisIdentifyStateOff
	ChassisIdentifyStateTemporaryOn   = chassis.ChassisIdentifyStateTemporaryOn
	DevicePowerStateD0                = chassis.DevicePowerStateD0
	DevicePowerStateD1                = chassis.DevicePowerStateD1
	DevicePowerStateD2                = chassis.DevicePowerStateD2
	DevicePowerStateD3                = chassis.DevicePowerStateD3
	DevicePowerStateNoChange          = chassis.DevicePowerStateNoChange
	DevicePowerStateUnknown           = chassis.DevicePowerStateUnknown
	PowerRestorePolicyAlwaysOff       = chassis.PowerRestorePolicyAlwaysOff
	PowerRestorePolicyAlwaysOn        = chassis.PowerRestorePolicyAlwaysOn
	PowerRestorePolicyPrevious        = chassis.PowerRestorePolicyPrevious
	SystemPowerStateG1Sleeping        = chassis.SystemPowerStateG1Sleeping
	SystemPowerStateG3                = chassis.SystemPowerStateG3
	SystemPowerStateLegacyOff         = chassis.SystemPowerStateLegacyOff
	SystemPowerStateLegacyOn          = chassis.SystemPowerStateLegacyOn
	SystemPowerStateNoChange          = chassis.SystemPowerStateNoChange
	SystemPowerStateOverride          = chassis.SystemPowerStateOverride
	SystemPowerStateS0G0              = chassis.SystemPowerStateS0G0
	SystemPowerStateS1                = chassis.SystemPowerStateS1
	SystemPowerStateS2                = chassis.SystemPowerStateS2
	SystemPowerStateS3                = chassis.SystemPowerStateS3
	SystemPowerStateS4                = chassis.SystemPowerStateS4
	SystemPowerStateS4S5              = chassis.SystemPowerStateS4S5
	SystemPowerStateS5G2              = chassis.SystemPowerStateS5G2
	SystemPowerStateSleeping          = chassis.SystemPowerStateSleeping
	SystemPowerStateUnknown           = chassis.SystemPowerStateUnknown
)

var (
	SupportedPowerRestorePolicies = chassis.SupportedPowerRestorePolicies
)

type (
	ChassisControl                    = chassis.ChassisControl
	ChassisControlRequest             = chassis.ChassisControlRequest
	ChassisControlResponse            = chassis.ChassisControlResponse
	ChassisIdentifyRequest            = chassis.ChassisIdentifyRequest
	ChassisIdentifyResponse           = chassis.ChassisIdentifyResponse
	ChassisIdentifyState              = chassis.ChassisIdentifyState
	ChassisResetRequest               = chassis.ChassisResetRequest
	ChassisResetResponse              = chassis.ChassisResetResponse
	DevicePowerState                  = chassis.DevicePowerState
	GetACPIPowerStateRequest          = chassis.GetACPIPowerStateRequest
	GetACPIPowerStateResponse         = chassis.GetACPIPowerStateResponse
	GetChassisCapabilitiesRequest     = chassis.GetChassisCapabilitiesRequest
	GetChassisCapabilitiesResponse    = chassis.GetChassisCapabilitiesResponse
	GetChassisStatusRequest           = chassis.GetChassisStatusRequest
	GetChassisStatusResponse          = chassis.GetChassisStatusResponse
	GetPOHCounterRequest              = chassis.GetPOHCounterRequest
	GetPOHCounterResponse             = chassis.GetPOHCounterResponse
	GetSystemBootOptionsParamRequest  = chassis.GetSystemBootOptionsParamRequest
	GetSystemBootOptionsParamResponse = chassis.GetSystemBootOptionsParamResponse
	GetSystemRestartCauseRequest      = chassis.GetSystemRestartCauseRequest
	GetSystemRestartCauseResponse     = chassis.GetSystemRestartCauseResponse
	PowerRestorePolicy                = chassis.PowerRestorePolicy
	SetACPIPowerStateRequest          = chassis.SetACPIPowerStateRequest
	SetACPIPowerStateResponse         = chassis.SetACPIPowerStateResponse
	SetChassisCapabilitiesRequest     = chassis.SetChassisCapabilitiesRequest
	SetChassisCapabilitiesResponse    = chassis.SetChassisCapabilitiesResponse
	SetFrontPanelEnablesRequest       = chassis.SetFrontPanelEnablesRequest
	SetFrontPanelEnablesResponse      = chassis.SetFrontPanelEnablesResponse
	SetPowerCycleIntervalRequest      = chassis.SetPowerCycleIntervalRequest
	SetPowerCycleIntervalResponse     = chassis.SetPowerCycleIntervalResponse
	SetPowerRestorePolicyRequest      = chassis.SetPowerRestorePolicyRequest
	SetPowerRestorePolicyResponse     = chassis.SetPowerRestorePolicyResponse
	SetSystemBootOptionsParamRequest  = chassis.SetSystemBootOptionsParamRequest
	SetSystemBootOptionsParamResponse = chassis.SetSystemBootOptionsParamResponse
	SystemPowerState                  = chassis.SystemPowerState
	SystemRestartCause                = chassis.SystemRestartCause
)
