package ipmi

import (
	"fmt"
	"strconv"
	"strings"
)

// 17.7 Event Filter Table
type PEFEventFilter struct {
	// Filter Configuration
	//
	// FilterState (enabled or disabled)
	// [4:0] - reserved
	//
	// [7] - 1b = enable filter
	//       0b = disable filter
	FilterState bool
	// [6:5] - 11b = reserved
	//         10b = manufacturer pre-configured filter. The filter entry has been
	//               configured by the system integrator and should not be altered by software.
	//               Software is allowed to enable or disable the filter, however.
	//         01b = reserved
	//         00b = software configurable filter. The filter entry is available for
	//               configuration by system management software.
	FilterType PEFEventFilterType

	// 17.6 PEF Actions
	// All actions are optional for an implementation, with the exception of Alert
	// which is mandatory if alerting is supported for one or more channels.
	// The BMC will return 0b for unsupported actions.
	// Software can test for which actions are supported by writing 1's to the
	// specified fields and  reading back the result.
	// (Note that reserved bits must be written with 0's)
	ActionGroupControlOperation bool
	ActionDiagnosticInterrupt   bool
	ActionOEM                   bool
	ActionPowerCycle            bool
	ActionReset                 bool
	ActionPowerOff              bool

	// Either Event filter Action should be enabled or Power action should be present as channel alert is enabled.
	ActionAlert bool // Relates with AlertPolicyNumber

	// Used to select an alerting policy set from the Alert Policy Table.
	// The Alert Policy Table holds different policies that configure the order in which different alert destinations and alerting media are tried.
	//   [6:4] - group control selector (1-based). Selects entry from group control table. (see [ICMB)
	GroupControlSelector uint8
	//   [3:0] - policy number. Value is "don't care" if (ActionAlert=false) Alert is not selected in the Event Filter Action.
	AlertPolicyNumber uint8

	EventSeverity PEFEventSeverity

	GeneratorID GeneratorID

	SensorType SensorType

	SensorNumber SensorNumber

	EventReadingType EventReadingType

	EventData1EventOffsetMask uint16

	EventData1ANDMask uint8
	// Used to indicate whether each bit position's comparison is an exact comparison or not.
	EventData1Compare1 uint8
	EventData1Compare2 uint8

	EventData2ANDMask  uint8
	EventData2Compare1 uint8
	EventData2Compare2 uint8

	EventData3ANDMask  uint8
	EventData3Compare1 uint8
	EventData3Compare2 uint8
}

func (entry *PEFEventFilter) enabledActions() []string {
	out := make([]string, 0)
	if entry.ActionGroupControlOperation {
		out = append(out, "Group Control Operation")
	}
	if entry.ActionDiagnosticInterrupt {
		out = append(out, "DiagnosticInterrupt")
	}
	if entry.ActionOEM {
		out = append(out, "OEM-defined")
	}
	if entry.ActionPowerCycle {
		out = append(out, "PowerCycle")
	}
	if entry.ActionReset {
		out = append(out, "Reset")
	}
	if entry.ActionPowerOff {
		out = append(out, "PowerOff")
	}
	if entry.ActionAlert {
		out = append(out, "Alert")
	}
	return out
}

func (entry *PEFEventFilter) Unpack(data []byte) error {
	if len(data) < 20 {
		return ErrUnpackedDataTooShortWith(len(data), 20)
	}

	var b byte

	b = data[0]
	entry.FilterState = isBit7Set(b)
	entry.FilterType = PEFEventFilterType((b >> 5) & 0x03)

	b = data[1]
	entry.ActionGroupControlOperation = isBit6Set(b)
	entry.ActionDiagnosticInterrupt = isBit5Set(b)
	entry.ActionOEM = isBit4Set(b)
	entry.ActionPowerCycle = isBit3Set(b)
	entry.ActionReset = isBit2Set(b)
	entry.ActionPowerOff = isBit1Set(b)
	entry.ActionAlert = isBit0Set(b)

	b = data[2]
	entry.GroupControlSelector = (b >> 4) & 0x07
	entry.AlertPolicyNumber = b & 0x0f

	entry.EventSeverity = PEFEventSeverity(data[3])

	generatorID, _, _ := unpackUint16L(data, 4)
	entry.GeneratorID = GeneratorID(generatorID)

	entry.SensorType = SensorType(data[6])
	entry.SensorNumber = SensorNumber(data[7])
	entry.EventReadingType = EventReadingType(data[8])

	eventData1, _, _ := unpackUint16L(data, 9)
	entry.EventData1EventOffsetMask = eventData1

	entry.EventData1ANDMask = data[11]
	entry.EventData1Compare1 = data[12]
	entry.EventData1Compare2 = data[13]

	entry.EventData2ANDMask = data[14]
	entry.EventData2Compare1 = data[15]
	entry.EventData2Compare2 = data[16]

	entry.EventData3ANDMask = data[17]
	entry.EventData3Compare1 = data[18]
	entry.EventData3Compare2 = data[19]

	return nil
}

func (entry *PEFEventFilter) Pack() []byte {
	out := make([]byte, 20)
	var b byte

	b = uint8(entry.FilterType) << 5
	b = setOrClearBit7(b, entry.FilterState)
	out[0] = b

	b = 0
	b = setOrClearBit6(b, entry.ActionGroupControlOperation)
	b = setOrClearBit5(b, entry.ActionDiagnosticInterrupt)
	b = setOrClearBit4(b, entry.ActionOEM)
	b = setOrClearBit3(b, entry.ActionPowerCycle)
	b = setOrClearBit2(b, entry.ActionReset)
	b = setOrClearBit1(b, entry.ActionPowerOff)
	b = setOrClearBit0(b, entry.ActionAlert)
	out[1] = b

	b = uint8(entry.GroupControlSelector) << 4
	b |= entry.AlertPolicyNumber & 0x0f
	out[2] = b

	out[3] = byte(entry.EventSeverity)

	packUint16L(uint16(entry.GeneratorID), out, 4)

	out[6] = byte(entry.SensorType)
	out[7] = byte(entry.SensorNumber)
	out[8] = byte(entry.EventReadingType)

	packUint16L(entry.EventData1EventOffsetMask, out, 9)

	out[11] = entry.EventData1ANDMask
	out[12] = entry.EventData1Compare1
	out[13] = entry.EventData1Compare2

	out[14] = entry.EventData2ANDMask
	out[15] = entry.EventData2Compare1
	out[16] = entry.EventData2Compare2

	out[17] = entry.EventData3ANDMask
	out[18] = entry.EventData3Compare1
	out[19] = entry.EventData3Compare2

	return out
}

func (entry *PEFEventFilter) Format() string {
	return "" +
		fmt.Sprintf("FilterType                  : %v\n", entry.FilterType) +
		fmt.Sprintf("FilterState                 : %v\n", formatBool(entry.FilterState, "enabled", "disabled")) +
		fmt.Sprintf("ActionGroupControlOperation : %v\n", formatBool(entry.ActionGroupControlOperation, "enabled", "disabled")) +
		fmt.Sprintf("ActionDiagnosticInterrupt   : %v\n", formatBool(entry.ActionDiagnosticInterrupt, "enabled", "disabled")) +
		fmt.Sprintf("ActionOEM                   : %v\n", formatBool(entry.ActionOEM, "enabled", "disabled")) +
		fmt.Sprintf("ActionPowerCycle            : %v\n", formatBool(entry.ActionPowerCycle, "enabled", "disabled")) +
		fmt.Sprintf("ActionReset                 : %v\n", formatBool(entry.ActionReset, "enabled", "disabled")) +
		fmt.Sprintf("ActionPowerOff              : %v\n", formatBool(entry.ActionPowerOff, "enabled", "disabled")) +
		fmt.Sprintf("ActionAlert                 : %v\n", formatBool(entry.ActionAlert, "enabled", "disabled")) +
		fmt.Sprintf("GroupControlSelector        : %v\n", entry.GroupControlSelector) +
		fmt.Sprintf("AlertPolicyNumber           : %v\n", entry.AlertPolicyNumber) +
		fmt.Sprintf("EventSeverity               : %v\n", entry.EventSeverity) +
		fmt.Sprintf("GeneratorID                 : %v\n", entry.GeneratorID) +
		fmt.Sprintf("SensorType                  : %v\n", entry.SensorType) +
		fmt.Sprintf("SensorNumber                : %v\n", entry.SensorNumber) +
		fmt.Sprintf("EventReadingType            : %v\n", entry.EventReadingType) +
		fmt.Sprintf("EventData1EventOffsetMask   : %v\n", entry.EventData1EventOffsetMask) +
		fmt.Sprintf("EventData1ANDMask           : %v\n", entry.EventData1ANDMask) +
		fmt.Sprintf("EventData1Compare1          : %v\n", entry.EventData1Compare1) +
		fmt.Sprintf("EventData1Compare2          : %v\n", entry.EventData1Compare2) +
		fmt.Sprintf("EventData2ANDMask           : %v\n", entry.EventData2ANDMask) +
		fmt.Sprintf("EventData2Compare1          : %v\n", entry.EventData2Compare1) +
		fmt.Sprintf("EventData2Compare2          : %v\n", entry.EventData2Compare2) +
		fmt.Sprintf("EventData3ANDMask           : %v\n", entry.EventData3ANDMask) +
		fmt.Sprintf("EventData3Compare1          : %v\n", entry.EventData3Compare1) +
		fmt.Sprintf("EventData3Compare2          : %v\n", entry.EventData3Compare2)
}

func FormatEventFilters(eventFilters []*PEFEventFilter) string {
	rows := make([]map[string]string, len(eventFilters))

	for i, f := range eventFilters {
		rows[i] = map[string]string{
			"Filter State":           formatBool(f.FilterState, "enabled", "disabled"),
			"Filter Type":            f.FilterType.String(),
			"Actions":                strings.Join(f.enabledActions(), ","),
			"Group Control Selector": fmt.Sprintf("%v", f.GroupControlSelector),
			"Alert Policy Number":    fmt.Sprintf("%v", f.AlertPolicyNumber),
			"Event Severity":         fmt.Sprintf("%v", f.EventSeverity),
			"Generator ID":           fmt.Sprintf("%v", f.GeneratorID),
			"Sensor Type":            fmt.Sprintf("%v", f.SensorType),
			"Sensor Number":          fmt.Sprintf("%#02x", f.SensorNumber),
			"Event Reading Type":     fmt.Sprintf("%v", f.EventReadingType),
			"ED1 Event Offset Mask":  fmt.Sprintf("%v", f.EventData1EventOffsetMask),
			// "ED1 AND Mask": fmt.Sprintf("%v", f.EventData1ANDMask),
			// "ED1 Compare 1": fmt.Sprintf("%v", f.EventData1Compare1),
			// "ED1 Compare 2": fmt.Sprintf("%v", f.EventData1Compare2),
			// "ED2 AND Mask": fmt.Sprintf("%v", f.EventData2ANDMask),
			// "ED2 Compare 1": fmt.Sprintf("%v", f.EventData2Compare1),
			// "ED2 Compare 2": fmt.Sprintf("%v", f.EventData2Compare2),
			// "ED3 AND Mask": fmt.Sprintf("%v", f.EventData3ANDMask),
			// "ED3 Compare 1": fmt.Sprintf("%v", f.EventData3Compare1),
			// "ED3 Compare 2": fmt.Sprintf("%v", f.EventData3Compare2),
		}
	}

	headers := []string{
		"Filter State",
		"Filter Type",
		"Actions",
		"Group Control Selector",
		"Alert Policy Number",
		"Event Severity",
		"Generator ID",
		"Sensor Type",
		"Sensor Number",
		"Event Reading Type",
		"ED1 Event Offset Mask",
		// "ED1 AND Mask",
		// "ED1 Compare 1",
		// "ED1 Compare 2",
		// "ED2 AND Mask",
		// "ED2 Compare 1",
		// "ED2 Compare 2",
		// "ED3 AND Mask",
	}

	return formatTable(headers, rows)
}

// PEFEventFilterType:
//   - manufacturer pre-configured filter.
//     The filter entry has been configured by the system integrator and
//     should not be altered by software. Software is allowed to enable or
//     disable the filter, however.
//   - software configurable filter.
//     The filter entry is available for configuration by system management software.
type PEFEventFilterType uint8

const (
	PEFEventFilterType_Configurable  PEFEventFilterType = 0x00
	PEFEventFilterType_PreConfigured PEFEventFilterType = 0x10
)

func (filterType PEFEventFilterType) String() string {
	m := map[PEFEventFilterType]string{
		PEFEventFilterType_Configurable:  "Configurable",
		PEFEventFilterType_PreConfigured: "Pre-Configured",
	}
	s, ok := m[filterType]
	if ok {
		return s
	}
	return fmt.Sprintf("%#02x", filterType)
}

type PEFEventSeverity uint8

const (
	PEFEventSeverityUnspecified    PEFEventSeverity = 0x00
	PEFEventSeverityMonitor        PEFEventSeverity = 0x01
	PEFEventSeverityInformation    PEFEventSeverity = 0x02
	PEFEventSeverityOK             PEFEventSeverity = 0x04
	PEFEventSeverityNonCritical    PEFEventSeverity = 0x08 // aka Warning
	PEFEventSeverityCritical       PEFEventSeverity = 0x10
	PEFEventSeverityNonRecoverable PEFEventSeverity = 0x20
)

func (severity PEFEventSeverity) String() string {
	m := map[PEFEventSeverity]string{
		PEFEventSeverityUnspecified:    "Unspecified",
		PEFEventSeverityMonitor:        "Monitor",
		PEFEventSeverityInformation:    "Information",
		PEFEventSeverityOK:             "OK",
		PEFEventSeverityNonCritical:    "Non-Critical",
		PEFEventSeverityCritical:       "Critical",
		PEFEventSeverityNonRecoverable: "Non-Recoverable",
	}
	if s, ok := m[severity]; ok {
		return s
	}

	return ""
}

// 17.11 Alert Policy Table
type PEFAlertPolicy struct {
	// [7:4] - policy number. 1 based. 0000b = reserved.
	PolicyNumber uint8
	// PolicyState (enabled or disabled)
	//
	// [3] - 0b = this entry is disabled. Skip to next entry in policy, if any.
	//       1b = this entry is enabled.
	PolicyState bool
	// [2:0] - policy
	PolicyAction PEFAlertPolicyAction

	// [7:4] = ChannelNumber Number.
	ChannelNumber uint8
	// [3:0] = Destination selector.
	Destination uint8

	// [7] - Event-specific Alert String
	//   1b = Alert String look-up is event specific. The following Alert String Set / Selector sub-
	//        field is interpreted as an Alert String Set Number that is used in conjunction with
	//        the Event Filter Number to lookup the Alert String from the PEF Configuration Parameters.
	//   0b = Alert String is not event specific. The following Alert String Set / Selector sub-field
	//        is interpreted as an Alert String Selector that provides a direct pointer to the
	//        desired Alert String from the PEF Configuration Parameters.
	IsEventSpecific bool
	// [6:0] - Alert String Set / Selector.
	// This value identifies one or more Alert Strings in the Alert String table.
	//   - When used as an Alert String Set Number (IsEventSpecific = true), it is used in conjunction with the Event Filter Number to uniquely identify an Alert String.
	//   - When used as an Alert String Selector (IsEventSpecific = false), it directly selects an Alert String from the PEF Configuration Parameters.
	AlertStringKey uint8
}

func (entry *PEFAlertPolicy) Pack() []byte {
	out := make([]byte, 3)

	var b uint8

	b = uint8(entry.PolicyAction) & 0x07
	b = setOrClearBit3(b, entry.PolicyState)
	b |= entry.PolicyNumber << 4
	out[0] = b

	b = entry.Destination & 0x0F
	b |= entry.ChannelNumber << 4
	out[1] = b

	b = entry.AlertStringKey & 0x7F
	b = setOrClearBit7(b, entry.IsEventSpecific)
	out[2] = b

	return out
}

func (entry *PEFAlertPolicy) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	entry.PolicyNumber = data[0] >> 4
	entry.PolicyState = isBit3Set(data[0])
	entry.PolicyAction = PEFAlertPolicyAction(data[0] & 0x07)

	entry.ChannelNumber = data[1] >> 4
	entry.Destination = data[1] & 0x0F

	entry.IsEventSpecific = isBit7Set(data[2])
	entry.AlertStringKey = data[2] & 0x7F

	return nil
}

func (entry *PEFAlertPolicy) Format() string {
	return "" +
		fmt.Sprintf("PolicyNumber    : %d\n", entry.PolicyNumber) +
		fmt.Sprintf("PolicyState     : %v\n", entry.PolicyState) +
		fmt.Sprintf("PolicyAction    : %v\n", entry.PolicyAction) +
		fmt.Sprintf("Channel         : %d\n", entry.ChannelNumber) +
		fmt.Sprintf("Destination     : %d\n", entry.Destination) +
		fmt.Sprintf("IsEventSpecific : %v\n", entry.IsEventSpecific) +
		fmt.Sprintf("AlertStringKey  : %d\n", entry.AlertStringKey)
}

func FormatPEFAlertPolicyTable(alertPolicies []*PEFAlertPolicy) string {

	rows := make([]map[string]string, len(alertPolicies))
	for i, alertPolicy := range alertPolicies {
		rows[i] = map[string]string{
			"Entry":           strconv.Itoa(i + 1),
			"PolicyNumber":    fmt.Sprintf("%d", alertPolicy.PolicyNumber),
			"PolicyState":     fmt.Sprintf("%v", alertPolicy.PolicyState),
			"PolicyAction":    alertPolicy.PolicyAction.ShortString(),
			"Channel":         fmt.Sprintf("%d", alertPolicy.ChannelNumber),
			"Destination":     fmt.Sprintf("%d", alertPolicy.Destination),
			"IsEventSpecific": fmt.Sprintf("%v", alertPolicy.IsEventSpecific),
			"AlertStringKey":  fmt.Sprintf("%d", alertPolicy.AlertStringKey),
		}
	}

	headers := []string{
		"Entry",
		"PolicyNumber",
		"PolicyState",
		"PolicyAction",
		"Channel",
		"Destination",
		"IsEventSpecific",
		"AlertStringKey",
	}

	return formatTable(headers, rows)
}

type PEFAlertPolicyAction uint8

const (
	// always send alert to this destination.
	PEFAlertPolicyAction_Always PEFAlertPolicyAction = 0

	// if alert to previous destination was successful, do not send alert to this destination.
	// Proceed to next entry in this policy set.
	PEFAlertPolicyAction_ProceedNext PEFAlertPolicyAction = 1

	// if alert to previous destination was successful, do not send alert to this destination.
	// Do not process any more entries in this policy set.
	PEFAlertPolicyAction_NoProceed PEFAlertPolicyAction = 2

	// if alert to previous destination was successful, do not send alert to this destination.
	// Proceed to next entry in this policy set that is to a different channel.
	PEFAlertPolicyAction_ProceedNextDifferentChannel PEFAlertPolicyAction = 3

	// if alert to previous destination was successful, do not send alert to this destination.
	// Proceed to next entry in this policy set that is to a different destination type.
	PEFAlertPolicyAction_ProceedNextDifferentDestination PEFAlertPolicyAction = 4
)

func (action PEFAlertPolicyAction) String() string {
	m := map[PEFAlertPolicyAction]string{
		PEFAlertPolicyAction_Always:                          "Always send alert to this destination",
		PEFAlertPolicyAction_ProceedNext:                     "If previous successful, skip this and continue (if configured)",
		PEFAlertPolicyAction_NoProceed:                       "If previous successful, stop alerting further",
		PEFAlertPolicyAction_ProceedNextDifferentChannel:     "If previous successful, switch to another channel (if configured)",
		PEFAlertPolicyAction_ProceedNextDifferentDestination: "If previous successful, switch to another destination (if configured)",
	}

	if s, ok := m[action]; ok {
		return s
	}

	return "Unknown"
}

func (action PEFAlertPolicyAction) ShortString() string {
	m := map[PEFAlertPolicyAction]string{
		PEFAlertPolicyAction_Always:                          "Match-always",
		PEFAlertPolicyAction_ProceedNext:                     "Proceed-next",
		PEFAlertPolicyAction_NoProceed:                       "No-proceed",
		PEFAlertPolicyAction_ProceedNextDifferentChannel:     "Different-channel",
		PEFAlertPolicyAction_ProceedNextDifferentDestination: "Different-destination",
	}

	if s, ok := m[action]; ok {
		return s
	}

	return "Unknown"
}
