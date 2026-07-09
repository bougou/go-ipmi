package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/cmd/chassis"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) SetBootParamSetInProgress(ctx context.Context, setInProgress types.SetInProgressState) error {
	param := &types.BootOptionParam_SetInProgress{
		Value: setInProgress,
	}

	if err := c.SetSystemBootOptionsParamFor(ctx, param); err != nil {
		return fmt.Errorf("SetSystemBootOptionsFor failed, err: %w", err)
	}

	return nil
}

func (c *Client) SetBootParamBootFlags(ctx context.Context, bootFlags *types.BootOptionParam_BootFlags) error {
	if err := c.SetBootParamSetInProgress(ctx, types.SetInProgress_SetInProgress); err != nil {
		goto OUT
	} else {
		if err := c.SetSystemBootOptionsParamFor(ctx, bootFlags); err != nil {
			return fmt.Errorf("SetSystemBootOptions failed, err: %w", err)
		}
	}

OUT:
	if err := c.SetBootParamSetInProgress(ctx, types.SetInProgress_SetComplete); err != nil {
		return fmt.Errorf("SetBootParamSetInProgress failed, err: %w", err)
	}

	return nil
}

func (c *Client) SetBootParamClearAck(ctx context.Context, by types.BootInfoAcknowledgeBy) error {
	param := &types.BootOptionParam_BootInfoAcknowledge{}

	switch by {
	case types.BootInfoAcknowledgeByBIOSPOST:
		param.ByBIOSPOST = true
	case types.BootInfoAcknowledgeByOSLoader:
		param.ByOSLoader = true
	case types.BootInfoAcknowledgeByOSServicePartition:
		param.ByOSServicePartition = true
	case types.BootInfoAcknowledgeBySMS:
		param.BySMS = true
	case types.BootInfoAcknowledgeByOEM:
		param.ByOEM = true
	}

	if err := c.SetSystemBootOptionsParamFor(ctx, param); err != nil {
		return fmt.Errorf("SetSystemBootOptionsFor failed, err: %w", err)
	}

	return nil
}

// SetBootDevice set the boot device for next boot.
// persist of false means it applies to next boot only.
// persist of true means this setting is persistent for all future boots.
func (c *Client) SetBootDevice(ctx context.Context, bootDeviceSelector types.BootDeviceSelector, bootType types.BIOSBootType, persist bool) error {
	param := &types.BootOptionParam_BootFlags{
		BootFlagsValid:     true,
		Persist:            persist,
		BIOSBootType:       bootType,
		BootDeviceSelector: bootDeviceSelector,
	}

	if err := c.SetSystemBootOptionsParamFor(ctx, param); err != nil {
		return fmt.Errorf("SetSystemBootOptions failed, err: %w", err)
	}

	return nil
}

// The following command is used to enable or disable the buttons on the front panel of the chassis.
func (c *Client) SetFrontPanelEnables(ctx context.Context, disableSleepButton bool, disableDiagnosticButton bool, disableResetButton bool, disablePoweroffButton bool) (response *chassis.SetFrontPanelEnablesResponse, err error) {
	request := &chassis.SetFrontPanelEnablesRequest{
		DisableSleepButton:      disableSleepButton,
		DisableDiagnosticButton: disableDiagnosticButton,
		DisableResetButton:      disableResetButton,
		DisablePoweroffButton:   disablePoweroffButton,
	}
	response = &chassis.SetFrontPanelEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) ChassisControl(ctx context.Context, control chassis.ChassisControl) (response *chassis.ChassisControlResponse, err error) {
	request := &chassis.ChassisControlRequest{
		ChassisControl: control,
	}
	response = &chassis.ChassisControlResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command causes the chassis to physically identify itself by a mechanism
// chosen by the system implementation; such as turning on blinking user-visible lights
// or emitting beeps via a speaker, LCD panel, etc.
func (c *Client) ChassisIdentify(ctx context.Context, interval uint8, force bool) (response *chassis.ChassisIdentifyResponse, err error) {
	request := &chassis.ChassisIdentifyRequest{
		IdentifyInterval: interval,
		ForceIdentifyOn:  force,
	}
	response = &chassis.ChassisIdentifyResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetChassisCapabilities(ctx context.Context, request *chassis.SetChassisCapabilitiesRequest) (response *chassis.SetChassisCapabilitiesResponse, err error) {
	response = &chassis.SetChassisCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemRestartCause(ctx context.Context) (response *chassis.GetSystemRestartCauseResponse, err error) {
	request := &chassis.GetSystemRestartCauseRequest{}
	response = &chassis.GetSystemRestartCauseResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command is provided to allow system software to tell a controller the present ACPI power state of the system.
func (c *Client) SetACPIPowerState(ctx context.Context, request *chassis.SetACPIPowerStateRequest) (err error) {
	response := &chassis.SetACPIPowerStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command returns the present reading of the POH (Power-On Hours) counter, plus the number of counts per hour.
func (c *Client) GetPOHCounter(ctx context.Context) (response *chassis.GetPOHCounterResponse, err error) {
	request := &chassis.GetPOHCounterRequest{}
	response = &chassis.GetPOHCounterResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetChassisCapabilities(ctx context.Context) (response *chassis.GetChassisCapabilitiesResponse, err error) {
	request := &chassis.GetChassisCapabilitiesRequest{}
	response = &chassis.GetChassisCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetChassisStatus(ctx context.Context) (response *chassis.GetChassisStatusResponse, err error) {
	request := &chassis.GetChassisStatusRequest{}
	response = &chassis.GetChassisStatusResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// This command is used to set parameters that direct the system boot following a system power up or reset.
// The boot flags only apply for one system restart. It is the responsibility of the system BIOS
// to read these settings from the BMC and then clear the boot flags
func (c *Client) SetSystemBootOptionsParam(ctx context.Context, request *chassis.SetSystemBootOptionsParamRequest) (response *chassis.SetSystemBootOptionsParamResponse, err error) {
	response = &chassis.SetSystemBootOptionsParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetSystemBootOptionsParamFor(ctx context.Context, param types.BootOptionParameter) error {
	if types.IsNilBootOptionParameter(param) {
		return fmt.Errorf("param is nil")
	}
	paramSelector, _, _ := param.BootOptionParameter()
	paramData := param.Pack()

	request := &chassis.SetSystemBootOptionsParamRequest{
		MarkParameterInvalid: false,
		ParamSelector:        paramSelector,
		ParamData:            paramData,
	}
	if _, err := c.SetSystemBootOptionsParam(ctx, request); err != nil {
		return fmt.Errorf("SetSystemBootOptionsParam failed, err: %w", err)
	}

	return nil
}

// This command was used with early versions of the ICMB.
// It has been superseded by the Chassis Control command
// For host systems, this corresponds to a system hard reset.
func (c *Client) ChassisReset(ctx context.Context) (response *chassis.ChassisResetResponse, err error) {
	request := &chassis.ChassisResetRequest{}
	response = &chassis.ChassisResetResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetPowerRestorePolicy(ctx context.Context, policy chassis.PowerRestorePolicy) (response *chassis.SetPowerRestorePolicyResponse, err error) {
	request := &chassis.SetPowerRestorePolicyRequest{
		PowerRestorePolicy: policy,
	}
	response = &chassis.SetPowerRestorePolicyResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetPowerCycleInterval(ctx context.Context, intervalInSec uint8) (response *chassis.SetPowerCycleIntervalResponse, err error) {
	request := &chassis.SetPowerCycleIntervalRequest{
		IntervalInSec: intervalInSec,
	}
	response = &chassis.SetPowerCycleIntervalResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemBootOptionsParam(ctx context.Context, paramSelector types.BootOptionParamSelector, setSelector uint8, blockSelector uint8) (response *chassis.GetSystemBootOptionsParamResponse, err error) {
	request := &chassis.GetSystemBootOptionsParamRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		BlockSelector: blockSelector,
	}
	response = &chassis.GetSystemBootOptionsParamResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetSystemBootOptionsParamFor(ctx context.Context, param types.BootOptionParameter) error {
	if types.IsNilBootOptionParameter(param) {
		return nil
	}
	paramSelector, setSelector, blockSelector := param.BootOptionParameter()

	response, err := c.GetSystemBootOptionsParam(ctx, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetSystemBootOptions for param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

// GetSystemBootOptionsParams get all parameters of boot options.
func (c *Client) GetSystemBootOptionsParams(ctx context.Context) (*types.BootOptionsParams, error) {
	bootOptionsParams := &types.BootOptionsParams{
		SetInProgress:            &types.BootOptionParam_SetInProgress{},
		ServicePartitionSelector: &types.BootOptionParam_ServicePartitionSelector{},
		ServicePartitionScan:     &types.BootOptionParam_ServicePartitionScan{},
		BMCBootFlagValidBitClear: &types.BootOptionParam_BMCBootFlagValidBitClear{},
		BootInfoAcknowledge:      &types.BootOptionParam_BootInfoAcknowledge{},
		BootFlags:                &types.BootOptionParam_BootFlags{},
		BootInitiatorInfo:        &types.BootOptionParam_BootInitiatorInfo{},
		BootInitiatorMailbox:     &types.BootOptionParam_BootInitiatorMailbox{},
	}

	if err := c.GetSystemBootOptionsParamsFor(ctx, bootOptionsParams); err != nil {
		return nil, fmt.Errorf("GetBootOptionsFor failed, err: %w", err)
	}

	return bootOptionsParams, nil
}

func (c *Client) GetSystemBootOptionsParamsFor(ctx context.Context, bootOptionsParams *types.BootOptionsParams) error {
	if bootOptionsParams == nil {
		return nil
	}

	if bootOptionsParams.SetInProgress != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.SetInProgress); err != nil {
			return err
		}
	}

	if bootOptionsParams.ServicePartitionSelector != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.ServicePartitionSelector); err != nil {
			return err
		}
	}

	if bootOptionsParams.ServicePartitionScan != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.ServicePartitionScan); err != nil {
			return err
		}
	}

	if bootOptionsParams.BMCBootFlagValidBitClear != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BMCBootFlagValidBitClear); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootInfoAcknowledge != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInfoAcknowledge); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootFlags != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootFlags); err != nil {
			return err
		}
	}

	if bootOptionsParams.BootInitiatorInfo != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInitiatorInfo); err != nil {
			return err
		}

	}

	if bootOptionsParams.BootInitiatorMailbox != nil {
		if err := c.GetSystemBootOptionsParamFor(ctx, bootOptionsParams.BootInitiatorMailbox); err != nil {
			return err
		}
	}
	return nil
}

// This command is provided to allow system software to tell a controller the present ACPI power state of the system.
func (c *Client) GetACPIPowerState(ctx context.Context) (response *chassis.GetACPIPowerStateResponse, err error) {
	request := &chassis.GetACPIPowerStateRequest{}
	response = &chassis.GetACPIPowerStateResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
