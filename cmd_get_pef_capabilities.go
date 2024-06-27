package ipmi

import (
	"fmt"
)

// 30.1 Get PEF Capabilities Command
type GetPEFCapabilitiesRequest struct {
	// no request data
}

type GetPEFCapabilitiesResponse struct {
	PEFVersion uint8

	SupportOEMEventRecordFilter bool
	SupportDiagnosticInterrupt  bool
	SupportOEMAction            bool
	SupportPowerCycle           bool
	SupportReset                bool
	SupportPowerDown            bool
	SupportAlert                bool

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

	res.PEFVersion = bcdUint8(msg[0])
	b1 := msg[1]
	res.SupportOEMEventRecordFilter = isBit7Set(b1)
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
	return fmt.Sprintf(`PEF Version                  : %d
Event Filter Table Entries   : %d
Support OEM Event Filtering  : %s
Support Diagnostic Interrupt : %s
Support OEM Action           : %s
Support Power Cycle          : %s
Support Reset                : %s
Support Power Down           : %s
Support Alert                : %s`,
		res.PEFVersion,
		res.EventFilterTableEntries,
		formatBool(res.SupportOEMEventRecordFilter, "yes", "no"),
		formatBool(res.SupportDiagnosticInterrupt, "yes", "no"),
		formatBool(res.SupportOEMAction, "yes", "no"),
		formatBool(res.SupportPowerCycle, "yes", "no"),
		formatBool(res.SupportReset, "yes", "no"),
		formatBool(res.SupportPowerDown, "yes", "no"),
		formatBool(res.SupportAlert, "yes", "no"),
	)
}

func (c *Client) GetPEFCapabilities() (response *GetPEFCapabilitiesResponse, err error) {
	request := &GetPEFCapabilitiesRequest{}
	response = &GetPEFCapabilitiesResponse{}
	err = c.Exchange(request, response)
	return
}
