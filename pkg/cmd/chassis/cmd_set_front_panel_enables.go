package chassis

import (
	"github.com/bougou/go-ipmi/pkg/types"
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
		b = types.SetBit3(b)
	}
	if req.DisableDiagnosticButton {
		b = types.SetBit2(b)
	}
	if req.DisableResetButton {
		b = types.SetBit1(b)
	}
	if req.DisablePoweroffButton {
		b = types.SetBit0(b)
	}
	types.PackUint8(b, out, 0)
	return out
}

func (req *SetFrontPanelEnablesRequest) Command() types.Command {
	return types.CommandSetFrontPanelEnables
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
