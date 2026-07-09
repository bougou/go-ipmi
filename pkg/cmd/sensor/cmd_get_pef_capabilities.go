package sensor

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 30.1 Get PEF Capabilities Command
type GetPEFCapabilitiesRequest struct {
	// empty
}

type GetPEFCapabilitiesResponse struct {
	// PEF Version (BCD encoded, LSN first. 51h version 1.5)
	PEFVersion uint8

	SupportOEMEventFilter      bool
	SupportDiagnosticInterrupt bool
	SupportOEMAction           bool
	SupportPowerCycle          bool
	SupportReset               bool
	SupportPowerDown           bool
	SupportAlert               bool

	EventFilterTableEntries uint8
}

func (req *GetPEFCapabilitiesRequest) Command() types.Command {
	return types.CommandGetPEFCapabilities
}

func (req *GetPEFCapabilitiesRequest) Pack() []byte {
	// empty request data
	return []byte{}
}

func (res *GetPEFCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.PEFVersion = msg[0]

	b1 := msg[1]
	res.SupportOEMEventFilter = types.IsBit7Set(b1)
	res.SupportDiagnosticInterrupt = types.IsBit5Set(b1)
	res.SupportOEMAction = types.IsBit4Set(b1)
	res.SupportPowerCycle = types.IsBit3Set(b1)
	res.SupportReset = types.IsBit2Set(b1)
	res.SupportPowerDown = types.IsBit1Set(b1)
	res.SupportAlert = types.IsBit0Set(b1)

	res.EventFilterTableEntries = msg[2]

	return nil
}

func (r *GetPEFCapabilitiesResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetPEFCapabilitiesResponse) Format() string {
	return "" +
		fmt.Sprintf("PEF Version                  : %#2x\n", res.PEFVersion) +
		fmt.Sprintf("Event Filter Table Entries   : %d\n", res.EventFilterTableEntries) +
		fmt.Sprintf("Support OEM Event Filtering  : %s\n", types.FormatBool(res.SupportOEMEventFilter, "supported", "not-supported")) +
		fmt.Sprintf("Support Diagnostic Interrupt : %s\n", types.FormatBool(res.SupportDiagnosticInterrupt, "supported", "not-supported")) +
		fmt.Sprintf("Support OEM Action           : %s\n", types.FormatBool(res.SupportOEMAction, "supported", "not-supported")) +
		fmt.Sprintf("Support Power Cycle          : %s\n", types.FormatBool(res.SupportPowerCycle, "supported", "not-supported")) +
		fmt.Sprintf("Support Reset                : %s\n", types.FormatBool(res.SupportReset, "supported", "not-supported")) +
		fmt.Sprintf("Support Power Down           : %s\n", types.FormatBool(res.SupportPowerDown, "supported", "not-supported")) +
		fmt.Sprintf("Support Alert                : %s\n", types.FormatBool(res.SupportAlert, "supported", "not-supported"))
}
