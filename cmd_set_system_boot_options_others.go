package ipmi

import "fmt"

func (c *Client) SetBootParamSetInProgressState(progressState BOP_SetInProgressState) error {
	r := &SetSystemBootOptionsRequest{
		MarkParameterInvalid: false,
		ParameterSelector:    BOPS_SetInProgressState,

		BootOptionParameter: BootOptionParameter{
			SetInProgressState: &progressState,
		},
	}

	_, err := c.SetSystemBootOptions(r)
	if err != nil {
		return fmt.Errorf("SetSystemBootOptions failed, err: %s", err)
	}

	return nil
}

func (c *Client) SetBootParamBootFlags(bootFlags *BOP_BootFlags) error {
	if err := c.SetBootParamSetInProgressState(SetInProgressState_SetInProgress); err != nil {
		goto OUT
	} else {
		r := &SetSystemBootOptionsRequest{
			MarkParameterInvalid: false,
			ParameterSelector:    BOPS_BootFlags,
			BootOptionParameter: BootOptionParameter{
				BootFlags: bootFlags,
			},
		}

		_, err := c.SetSystemBootOptions(r)
		if err != nil {
			return fmt.Errorf("SetSystemBootOptions failed, err: %s", err)
		}
	}

OUT:
	if err := c.SetBootParamSetInProgressState(SetInProgressState_SetComplete); err != nil {
		return fmt.Errorf("SetBootParamSetInProgressState failed, err: %s", err)
	}

	return nil
}

func (c *Client) SetBootParamClearAck(by BootInfoAcknowledgeBy) error {
	ack := &BOP_BootInfoAcknowledge{}

	switch by {
	case BootInfoAcknowledgeByBIOSPOST:
		ack.ByBIOSPOST = true
	case BootInfoAcknowledgeByOSLoader:
		ack.ByOSLoader = true
	case BootInfoAcknowledgeByOSServicePartition:
		ack.ByOSServicePartition = true
	case BootInfoAcknowledgeBySMS:
		ack.BySMS = true
	case BootInfoAcknowledgeByOEM:
		ack.ByOEM = true
	}

	r := &SetSystemBootOptionsRequest{
		MarkParameterInvalid: false,
		ParameterSelector:    BOPS_BootInfoAcknowledge,
		BootOptionParameter: BootOptionParameter{
			BootInfoAcknowledge: ack,
		},
	}

	_, err := c.SetSystemBootOptions(r)
	if err != nil {
		return fmt.Errorf("SetSystemBootOptions failed, err: %s", err)
	}

	return nil
}
