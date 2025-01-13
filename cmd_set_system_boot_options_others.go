package ipmi

import (
	"context"
	"fmt"
)

func (c *Client) SetBootParamSetInProgress(ctx context.Context, setInProgress SetInProgressState) error {
	param := &BootOptionParam_SetInProgress{
		Value: setInProgress,
	}

	if err := c.SetSystemBootOptionsParamFor(ctx, param); err != nil {
		return fmt.Errorf("SetSystemBootOptionsFor failed, err: %w", err)
	}

	return nil
}

func (c *Client) SetBootParamBootFlags(ctx context.Context, bootFlags *BootOptionParam_BootFlags) error {
	if err := c.SetBootParamSetInProgress(ctx, SetInProgress_SetInProgress); err != nil {
		goto OUT
	} else {
		if err := c.SetSystemBootOptionsParamFor(ctx, bootFlags); err != nil {
			return fmt.Errorf("SetSystemBootOptions failed, err: %w", err)
		}
	}

OUT:
	if err := c.SetBootParamSetInProgress(ctx, SetInProgress_SetComplete); err != nil {
		return fmt.Errorf("SetBootParamSetInProgress failed, err: %w", err)
	}

	return nil
}

func (c *Client) SetBootParamClearAck(ctx context.Context, by BootInfoAcknowledgeBy) error {
	param := &BootOptionParam_BootInfoAcknowledge{}

	switch by {
	case BootInfoAcknowledgeByBIOSPOST:
		param.ByBIOSPOST = true
	case BootInfoAcknowledgeByOSLoader:
		param.ByOSLoader = true
	case BootInfoAcknowledgeByOSServicePartition:
		param.ByOSServicePartition = true
	case BootInfoAcknowledgeBySMS:
		param.BySMS = true
	case BootInfoAcknowledgeByOEM:
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
func (c *Client) SetBootDevice(ctx context.Context, bootDeviceSelector BootDeviceSelector, bootType BIOSBootType, persist bool) error {
	param := &BootOptionParam_BootFlags{
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
