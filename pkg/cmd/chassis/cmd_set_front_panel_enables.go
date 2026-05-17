package chassis

import (
	ipmi "github.com/bougou/go-ipmi/pkg/types"
	// 28.6 Set Front Panel Enables
	// 定位
)

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
		b = ipmi.SetBit3(b)
	}
	if req.DisableSleepButton {
		b = ipmi.SetBit2(b)
	}
	if req.DisableSleepButton {
		b = ipmi.SetBit1(b)
	}
	if req.DisableSleepButton {
		b = ipmi.SetBit0(b)
	}
	ipmi.PackUint8(b, out, 1)
	return out
}

func (req *SetFrontPanelEnablesRequest) Command() ipmi.Command {
	return ipmi.CommandSetFrontPanelEnables
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
