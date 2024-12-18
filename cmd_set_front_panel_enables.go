package ipmi

import "context"

// 28.6 Set Front Panel Enables
// 定位
type SetFrontPanelEnablesRequest struct {
	DisableSleepButton      bool
	DisableDiagnosticButton bool
	DisableResetButton      bool
	DisablePoweroffButton   bool
}

type SetFrontPanelEnablesResponse struct {
	// empty
}

func (req *SetFrontPanelEnablesRequest) Pack() []byte {
	out := make([]byte, 1)

	var b uint8 = 0
	if req.DisableSleepButton {
		b = setBit3(b)
	}
	if req.DisableSleepButton {
		b = setBit2(b)
	}
	if req.DisableSleepButton {
		b = setBit1(b)
	}
	if req.DisableSleepButton {
		b = setBit0(b)
	}
	packUint8(b, out, 1)
	return out
}

func (req *SetFrontPanelEnablesRequest) Command() Command {
	return CommandSetFrontPanelEnables
}

func (res *SetFrontPanelEnablesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetFrontPanelEnablesResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetFrontPanelEnablesResponse) Format() string {
	return ""
}

// The following command is used to enable or disable the buttons on the front panel of the chassis.
func (c *Client) SetFrontPanelEnables(ctx context.Context, disableSleepButton bool, disableDiagnosticButton bool, disableResetButton bool, disablePoweroffButton bool) (response *SetFrontPanelEnablesResponse, err error) {
	request := &SetFrontPanelEnablesRequest{
		DisableSleepButton:      disableSleepButton,
		DisableDiagnosticButton: disableDiagnosticButton,
		DisableResetButton:      disableResetButton,
		DisablePoweroffButton:   disablePoweroffButton,
	}
	response = &SetFrontPanelEnablesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
