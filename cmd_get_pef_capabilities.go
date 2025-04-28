package ipmi

import (
	"context"
	"fmt"
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

func (req *GetPEFCapabilitiesRequest) Command() Command {
	return CommandGetPEFCapabilities
}

func (req *GetPEFCapabilitiesRequest) Pack() []byte {
	// empty request data
	return []byte{}
}

func (res *GetPEFCapabilitiesResponse) Unpack(msg []byte) error {
	if len(msg) < 3 {
		return ErrUnpackedDataTooShortWith(len(msg), 3)
	}

	res.PEFVersion = msg[0]

	b1 := msg[1]
	res.SupportOEMEventFilter = isBit7Set(b1)
	res.SupportDiagnosticInterrupt = isBit5Set(b1)
	res.SupportOEMAction = isBit4Set(b1)
	res.SupportPowerCycle = isBit3Set(b1)
	res.SupportReset = isBit2Set(b1)
	res.SupportPowerDown = isBit1Set(b1)
	res.SupportAlert = isBit0Set(b1)

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
		fmt.Sprintf("Support OEM Event Filtering  : %s\n", formatBool(res.SupportOEMEventFilter, "supported", "not-supported")) +
		fmt.Sprintf("Support Diagnostic Interrupt : %s\n", formatBool(res.SupportDiagnosticInterrupt, "supported", "not-supported")) +
		fmt.Sprintf("Support OEM Action           : %s\n", formatBool(res.SupportOEMAction, "supported", "not-supported")) +
		fmt.Sprintf("Support Power Cycle          : %s\n", formatBool(res.SupportPowerCycle, "supported", "not-supported")) +
		fmt.Sprintf("Support Reset                : %s\n", formatBool(res.SupportReset, "supported", "not-supported")) +
		fmt.Sprintf("Support Power Down           : %s\n", formatBool(res.SupportPowerDown, "supported", "not-supported")) +
		fmt.Sprintf("Support Alert                : %s\n", formatBool(res.SupportAlert, "supported", "not-supported"))
}

func (c *Client) GetPEFCapabilities(ctx context.Context) (response *GetPEFCapabilitiesResponse, err error) {
	request := &GetPEFCapabilitiesRequest{}
	response = &GetPEFCapabilitiesResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
