package ipmi

import "fmt"

type PEFConfigParamSelector uint8

const (
	PEFConfigParamSelector_SetInProgress       PEFConfigParamSelector = 0x00
	PEFConfigParamSelector_Control             PEFConfigParamSelector = 0x01
	PEFConfigParamSelector_ActionGlobalControl PEFConfigParamSelector = 0x02
	PEFConfigParamSelector_StartupDelay        PEFConfigParamSelector = 0x03
	PEFConfigParamSelector_AlertStartDelay     PEFConfigParamSelector = 0x04
	PEFConfigParamSelector_EventFiltersCount   PEFConfigParamSelector = 0x05
	PEFConfigParamSelector_EventFilter         PEFConfigParamSelector = 0x06
	PEFConfigParamSelector_EventFilterData1    PEFConfigParamSelector = 0x07
	PEFConfigParamSelector_AlertPoliciesCount  PEFConfigParamSelector = 0x08
	PEFConfigParamSelector_AlertPolicy         PEFConfigParamSelector = 0x09
	PEFConfigParamSelector_SystemGUID          PEFConfigParamSelector = 0x0a
	PEFConfigParamSelector_AlertStringsCount   PEFConfigParamSelector = 0x0b
	PEFConfigParamSelector_AlertStringKey      PEFConfigParamSelector = 0x0c
	PEFConfigParamSelector_AlertString         PEFConfigParamSelector = 0x0d
	PEFConfigParamSelector_GroupControlsCount  PEFConfigParamSelector = 0x0e
	PEFConfigParamSelector_GroupControl        PEFConfigParamSelector = 0x0f

	// 96:127
	// OEM Parameters (optional. Non-volatile or volatile as specified by OEM)
	// This range is available for special OEM configuration parameters.
	// The OEM is identified according to the Manufacturer ID field returned by the Get Device ID command.
)

func (p PEFConfigParamSelector) String() string {
	m := map[PEFConfigParamSelector]string{
		PEFConfigParamSelector_SetInProgress:       "Set In Progress",
		PEFConfigParamSelector_Control:             "Control",
		PEFConfigParamSelector_ActionGlobalControl: "Action Global Control",
		PEFConfigParamSelector_StartupDelay:        "Startup Delay",
		PEFConfigParamSelector_AlertStartDelay:     "Alert Start Delay",
		PEFConfigParamSelector_EventFiltersCount:   "Event Filters Count",
		PEFConfigParamSelector_EventFilter:         "Event Filter",
		PEFConfigParamSelector_EventFilterData1:    "Event Filter Data1",
		PEFConfigParamSelector_AlertPoliciesCount:  "Alert Policies Count",
		PEFConfigParamSelector_AlertPolicy:         "Alert Policy",
		PEFConfigParamSelector_SystemGUID:          "System GUID",
		PEFConfigParamSelector_AlertStringsCount:   "Alert Strings Count",
		PEFConfigParamSelector_AlertStringKey:      "Alert String Key",
		PEFConfigParamSelector_AlertString:         "Alert String",
		PEFConfigParamSelector_GroupControlsCount:  "Group Controls Count",
		PEFConfigParamSelector_GroupControl:        "Group Control",
	}

	if s, ok := m[p]; ok {
		return s
	}

	return fmt.Sprintf("Unknown (%#02x)", p)
}

type PEFConfigParameter interface {
	PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8)
	Parameter
}

var (
	_ PEFConfigParameter = (*PEFConfigParam_SetInProgress)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_Control)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_ActionGlobalControl)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_StartupDelay)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStartupDelay)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_EventFiltersCount)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_EventFilter)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertPoliciesCount)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertPolicy)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_SystemGUID)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStringsCount)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStringKey)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertString)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_GroupControlsCount)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_GroupControl)(nil)
)

type PEFConfig struct {
	SetInProgress       *PEFConfigParam_SetInProgress
	Control             *PEFConfigParam_Control
	ActionGlobalControl *PEFConfigParam_ActionGlobalControl
	StartupDelay        *PEFConfigParam_StartupDelay
	AlertStartupDelay   *PEFConfigParam_AlertStartupDelay
	EventFiltersCount   *PEFConfigParam_EventFiltersCount
	EventFilters        []*PEFConfigParam_EventFilter
	EventFiltersData1   []*PEFConfigParam_EventFilterData1
	AlertPoliciesCount  *PEFConfigParam_AlertPoliciesCount
	AlertPolicies       []*PEFConfigParam_AlertPolicy
	SystemGUID          *PEFConfigParam_SystemGUID
	AlertStringsCount   *PEFConfigParam_AlertStringsCount
	AlertStringKeys     []*PEFConfigParam_AlertStringKey
	AlertStrings        []*PEFConfigParam_AlertString
	GroupControlsCount  *PEFConfigParam_GroupControlsCount
	GroupControls       []*PEFConfigParam_GroupControl
}

func (pefConfig *PEFConfig) Format() string {
	var out string

	format := func(param PEFConfigParameter) string {
		paramSelector, _, _ := param.PEFConfigParameter()
		content := param.Format()
		if content[len(content)-1] != '\n' {
			content += "\n"
		}
		return fmt.Sprintf("[%2d] %s : %s", paramSelector, paramSelector.String(), content)
	}

	if pefConfig.SetInProgress != nil {
		out += format(pefConfig.SetInProgress)
	}
	if pefConfig.Control != nil {
		out += format(pefConfig.Control)
	}
	if pefConfig.ActionGlobalControl != nil {
		out += format(pefConfig.ActionGlobalControl)
	}
	if pefConfig.StartupDelay != nil {
		out += format(pefConfig.StartupDelay)
	}
	if pefConfig.AlertStartupDelay != nil {
		out += format(pefConfig.AlertStartupDelay)
	}
	if pefConfig.EventFiltersCount != nil {
		out += format(pefConfig.EventFiltersCount)
	}
	if pefConfig.EventFilters != nil {
		for i := range pefConfig.EventFilters {
			out += format(pefConfig.EventFilters[i])
		}
	}
	if pefConfig.EventFiltersData1 != nil {
		for i := range pefConfig.EventFiltersData1 {
			out += format(pefConfig.EventFiltersData1[i])
		}
	}
	if pefConfig.AlertPoliciesCount != nil {
		out += format(pefConfig.AlertPoliciesCount)
	}
	if pefConfig.AlertPolicies != nil {
		for i := range pefConfig.AlertPolicies {
			out += format(pefConfig.AlertPolicies[i])
		}
	}
	if pefConfig.SystemGUID != nil {
		out += format(pefConfig.SystemGUID)
	}
	if pefConfig.AlertStringsCount != nil {
		out += format(pefConfig.AlertStringsCount)
	}
	if pefConfig.AlertStringKeys != nil {
		for i := range pefConfig.AlertStringKeys {
			out += format(pefConfig.AlertStringKeys[i])
		}
	}
	if pefConfig.AlertStrings != nil {
		for i := range pefConfig.AlertStrings {
			out += format(pefConfig.AlertStrings[i])
		}
	}
	if pefConfig.GroupControlsCount != nil {
		out += format(pefConfig.GroupControlsCount)
	}
	if pefConfig.GroupControls != nil {
		for i := range pefConfig.GroupControls {
			out += format(pefConfig.GroupControls[i])
		}
	}

	return out
}

type PEFConfigParam_SetInProgress struct {
	Value SetInProgress
}

func (param *PEFConfigParam_SetInProgress) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_SetInProgress, 0, 0
}

func (param *PEFConfigParam_SetInProgress) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = SetInProgress(data[0])
	return nil
}

func (param *PEFConfigParam_SetInProgress) Pack() []byte {
	return []byte{byte(param.Value)}
}

func (param *PEFConfigParam_SetInProgress) Format() string {
	return fmt.Sprintf("%v", param.Value)
}

type PEFConfigParam_Control struct {
	EnablePEFAlertStartupDelay bool
	EnablePEFStartupDelay      bool
	EnableEventMessage         bool
	EnablePEF                  bool
}

func (param *PEFConfigParam_Control) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_Control, 0, 0
}

func (param *PEFConfigParam_Control) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.EnablePEFAlertStartupDelay = isBit3Set(data[0])
	param.EnablePEFStartupDelay = isBit2Set(data[0])
	param.EnableEventMessage = isBit1Set(data[0])
	param.EnablePEF = isBit0Set(data[0])

	return nil
}

func (param *PEFConfigParam_Control) Pack() []byte {
	b := uint8(0x00)

	b = setOrClearBit3(b, param.EnablePEFAlertStartupDelay)
	b = setOrClearBit2(b, param.EnablePEFStartupDelay)
	b = setOrClearBit1(b, param.EnableEventMessage)
	b = setOrClearBit0(b, param.EnablePEF)

	return []byte{b}
}

func (param *PEFConfigParam_Control) Format() string {
	return fmt.Sprintf(`
    PEF startup delay   : %s
    Alert startup delay : %s
    PEF event messages  : %s
    PEF                 : %s
`,
		formatBool(param.EnablePEFAlertStartupDelay, "enabled", "disabled"),
		formatBool(param.EnablePEFStartupDelay, "enabled", "disabled"),
		formatBool(param.EnableEventMessage, "enabled", "disabled"),
		formatBool(param.EnablePEF, "enabled", "disabled"),
	)
}

type PEFConfigParam_ActionGlobalControl struct {
	DiagnosticInterruptEnabled bool
	OEMActionEnabled           bool
	PowerCycleActionEnabled    bool
	ResetActionEnabled         bool
	PowerDownActionEnabled     bool
	AlertActionEnabled         bool
}

func (param *PEFConfigParam_ActionGlobalControl) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_ActionGlobalControl, 0, 0
}

func (param *PEFConfigParam_ActionGlobalControl) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DiagnosticInterruptEnabled = isBit5Set(data[0])
	param.OEMActionEnabled = isBit4Set(data[0])
	param.PowerCycleActionEnabled = isBit3Set(data[0])
	param.ResetActionEnabled = isBit2Set(data[0])
	param.PowerDownActionEnabled = isBit1Set(data[0])
	param.AlertActionEnabled = isBit0Set(data[0])

	return nil
}

func (param *PEFConfigParam_ActionGlobalControl) Pack() []byte {
	b := uint8(0x00)

	b = setOrClearBit5(b, param.DiagnosticInterruptEnabled)
	b = setOrClearBit4(b, param.OEMActionEnabled)
	b = setOrClearBit3(b, param.PowerCycleActionEnabled)
	b = setOrClearBit2(b, param.ResetActionEnabled)
	b = setOrClearBit1(b, param.PowerDownActionEnabled)
	b = setOrClearBit0(b, param.AlertActionEnabled)

	return []byte{b}
}

func (param *PEFConfigParam_ActionGlobalControl) Format() string {
	return fmt.Sprintf(`
    Diagnostic-interrupt : %s
    OEM-defined          : %s
    Power-cycle          : %s
    Reset                : %s
    Power-off            : %s
    Alert                : %s
`,
		formatBool(param.DiagnosticInterruptEnabled, "active", "inactive"),
		formatBool(param.OEMActionEnabled, "active", "inactive"),
		formatBool(param.PowerCycleActionEnabled, "active", "inactive"),
		formatBool(param.ResetActionEnabled, "active", "inactive"),
		formatBool(param.PowerDownActionEnabled, "active", "inactive"),
		formatBool(param.AlertActionEnabled, "active", "inactive"),
	)
}

type PEFConfigParam_StartupDelay struct {
	DelaySec uint8
}

func (param *PEFConfigParam_StartupDelay) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_StartupDelay, 0, 0
}

func (param *PEFConfigParam_StartupDelay) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DelaySec = data[0]

	return nil
}

func (param *PEFConfigParam_StartupDelay) Pack() []byte {
	return []byte{param.DelaySec}
}

func (param *PEFConfigParam_StartupDelay) Format() string {
	return fmt.Sprintf("%v", param.DelaySec)
}

type PEFConfigParam_AlertStartupDelay struct {
	DelaySec uint8
}

func (param *PEFConfigParam_AlertStartupDelay) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertStartDelay, 0, 0
}

func (param *PEFConfigParam_AlertStartupDelay) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DelaySec = data[0]

	return nil
}

func (param *PEFConfigParam_AlertStartupDelay) Pack() []byte {
	return []byte{param.DelaySec}
}

func (param *PEFConfigParam_AlertStartupDelay) Format() string {
	return fmt.Sprintf("%v", param.DelaySec)
}

// Number of event filters supported. 1-based.
// This parameter does not need to be supported if Alerting is not supported.
// READ ONLY
type PEFConfigParam_EventFiltersCount struct {
	Value uint8
}

func (param *PEFConfigParam_EventFiltersCount) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_EventFiltersCount, 0, 0
}

func (param *PEFConfigParam_EventFiltersCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]
	return nil
}

func (param *PEFConfigParam_EventFiltersCount) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_EventFiltersCount) Format() string {
	return fmt.Sprintf("%d", param.Value)
}

type PEFConfigParam_EventFilter struct {
	// Set Selector = filter number. 1-based. 00h = reserved.
	SetSelector uint8

	Filter *PEFEventFilter
}

func (param *PEFConfigParam_EventFilter) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_EventFilter, param.SetSelector, 0
}

func (param *PEFConfigParam_EventFilter) Unpack(data []byte) error {
	if len(data) < 21 {
		return ErrUnpackedDataTooShortWith(len(data), 21)
	}

	param.SetSelector = data[0]

	eventFilter := &PEFEventFilter{}
	if err := eventFilter.Unpack(data[1:21]); err != nil {
		return fmt.Errorf("unpack entry failed, err: %s", err)
	}
	param.Filter = eventFilter

	return nil
}

func (param *PEFConfigParam_EventFilter) Pack() []byte {
	entryData := param.Filter.Pack()
	out := make([]byte, len(entryData))

	out[0] = param.SetSelector
	packBytes(entryData, out, 1)

	return out
}

func (param *PEFConfigParam_EventFilter) Format() string {
	return fmt.Sprintf(`
    Event Filter Number:   %d
    Event Filter:
%v
`, param.SetSelector, param.Filter.Format())
}

// This parameter provides an aliased access to the first byte of the event filter data.
// This is provided to simplify the act of enabling and disabling individual filters
// by avoiding the need to do a read-modify-write of the entire filter data.
type PEFConfigParam_EventFilterData1 struct {
	// Set Selector = filter number
	SetSelector uint8

	// data byte 1 of event filter data

	// [7] - 1b = enable filter
	//       0b = disable filter
	FilterEnabled bool
	// [6:5] - 11b = reserved
	//         10b = manufacturer pre-configured filter. The filter entry has been
	//               configured by the system integrator and should not be altered by software.
	//               Software is allowed to enable or disable the filter, however.
	//         01b = reserved
	//         00b = software configurable filter. The filter entry is available for
	//               configuration by system management software.
	FilterType PEFEventFilterType
}

func (param *PEFConfigParam_EventFilterData1) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_EventFilterData1, param.SetSelector, 0
}

func (param *PEFConfigParam_EventFilterData1) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 21)
	}

	param.SetSelector = data[0]

	b := data[1]
	param.FilterEnabled = isBit7Set(b)
	param.FilterType = PEFEventFilterType((b >> 5) & 0x03)

	return nil
}

func (param *PEFConfigParam_EventFilterData1) Pack() []byte {
	out := make([]byte, 21)

	out[0] = param.SetSelector

	var b byte
	b = uint8(param.FilterType) << 5
	b = setOrClearBit7(b, param.FilterEnabled)
	out[1] = b

	return out
}

func (param *PEFConfigParam_EventFilterData1) Format() string {
	return fmt.Sprintf(`FilterNumber: %d, FilterEnabled: %v, FilterType: %v`, param.SetSelector, param.FilterEnabled, param.FilterType)
}

// Number of alert policy entries supported. 1-based.
// This parameter does not need to be supported if Alerting is not supported.
// READ ONLY
type PEFConfigParam_AlertPoliciesCount struct {
	Value uint8
}

func (param *PEFConfigParam_AlertPoliciesCount) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertPoliciesCount, 0, 0
}

func (param *PEFConfigParam_AlertPoliciesCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_AlertPoliciesCount) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_AlertPoliciesCount) Format() string {
	return fmt.Sprintf("%d", param.Value)
}

type PEFConfigParam_AlertPolicy struct {
	// Set Selector = entry number
	//  - [7] - reserved
	//  - [6:0] - alert policy entry number. 1-based.
	SetSelector uint8

	Policy *PEFAlertPolicy
}

func (param *PEFConfigParam_AlertPolicy) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertPolicy, param.SetSelector, 0
}

func (param *PEFConfigParam_AlertPolicy) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.SetSelector = data[0]

	b := &PEFAlertPolicy{}
	if err := b.Unpack(data[1:]); err != nil {
		return err
	}
	param.Policy = b

	return nil
}

func (param *PEFConfigParam_AlertPolicy) Pack() []byte {
	entryData := param.Policy.Pack()

	out := make([]byte, 1+len(entryData))

	out[0] = param.SetSelector
	packBytes(entryData, out, 1)

	return out
}

func (param *PEFConfigParam_AlertPolicy) Format() string {
	return fmt.Sprintf(`
    Entry Number %d : %v
`, param.SetSelector, param.Policy.Format())
}

type PEFConfigParam_SystemGUID struct {
	// Used to fill in the GUID field in a PET Trap.
	//   [7:1] - reserved
	//   [0]
	//    1b = BMC uses following value in PET Trap.
	//    0b = BMC ignores following value and uses value returned from Get System GUID command instead.
	UseGUID bool
	GUID    [16]byte
}

func (param *PEFConfigParam_SystemGUID) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_SystemGUID, 0, 0
}

func (param *PEFConfigParam_SystemGUID) Unpack(configData []byte) error {
	if len(configData) < 17 {
		return ErrUnpackedDataTooShortWith(len(configData), 17)
	}

	param.UseGUID = isBit0Set(configData[0])
	param.GUID = array16(configData[1:17])
	return nil
}

func (param *PEFConfigParam_SystemGUID) Pack() []byte {
	out := make([]byte, 17)

	out[0] = setOrClearBit0(0x00, param.UseGUID)
	copy(out[1:], param.GUID[:])

	return out
}

func (param *PEFConfigParam_SystemGUID) Format() string {
	var guidStr string

	guid, err := ParseGUID(param.GUID[:], GUIDModeSMBIOS)
	if err != nil {
		guidStr = fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	} else {
		guidStr = guid.String()
	}

	return fmt.Sprintf(`
    UseGUID : %v
    GUID    : %s
`, param.UseGUID, guidStr)
}

// Number of alert strings supported in addition to Alert String 0. 1-based.
// This parameter does not need to be supported if Alerting is not supported.
// READ ONLY
type PEFConfigParam_AlertStringsCount struct {
	Value uint8
}

func (param *PEFConfigParam_AlertStringsCount) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertStringsCount, 0, 0
}

func (param *PEFConfigParam_AlertStringsCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_AlertStringsCount) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_AlertStringsCount) Format() string {
	return fmt.Sprintf("%d", param.Value)
}

// Sets the keys used to look up Alert String data in PEF.
// This parameter does not need to be supported if Alerting is not supported.
//
// It's purpose is to get the AlertStringSelector from combination of the (Event) FilterNumber and AlertStringSet.
type PEFConfigParam_AlertStringKey struct {
	// Set Selector = Alert string selector.
	//   - 0 = selects volatile string parameters
	//   - 01h-7Fh = non-volatile string selectors
	SetSelector uint8

	// [6:0] - Filter number. 1-based. 00h = unspecified.
	FilterNumber uint8

	// [6:0] - Set number for string. 1-based. 00h = unspecified.
	AlertStringSet uint8
}

func (param *PEFConfigParam_AlertStringKey) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertStringKey, param.SetSelector, 0
}

func (param *PEFConfigParam_AlertStringKey) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.SetSelector = data[0]
	param.FilterNumber = data[1]
	param.AlertStringSet = data[2]

	return nil
}

func (param *PEFConfigParam_AlertStringKey) Pack() []byte {
	return []byte{param.SetSelector, param.FilterNumber, param.AlertStringSet}
}

func (param *PEFConfigParam_AlertStringKey) Format() string {
	return fmt.Sprintf(`Set Selector: %d, Event Filter Number: %d, Alert String Set: %d`,
		param.SetSelector, param.FilterNumber, param.AlertStringSet)
}

type PEFConfigParam_AlertString struct {
	// Set Selector = string selector.
	//   - 0 = selects volatile string
	//   - 01h-7Fh = non-volatile string selectors
	SetSelector uint8

	// Block Selector = string block number to set, 1 based. Blocks are 16 bytes.
	BlockSelector uint8

	// String data. Null terminated 8-bit ASCII string. 16-bytes max. per block.
	StringData []byte
}

func (param *PEFConfigParam_AlertString) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_AlertString, param.SetSelector, param.BlockSelector
}

func (param *PEFConfigParam_AlertString) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.SetSelector = data[0]
	param.BlockSelector = data[1]
	param.StringData, _, _ = unpackBytes(data, 2, len(data)-2)

	return nil
}

func (param *PEFConfigParam_AlertString) Pack() []byte {
	out := make([]byte, 2+len(param.StringData))

	out[0] = param.SetSelector
	out[1] = param.BlockSelector
	packBytes(param.StringData, out, 2)

	return out
}

func (param *PEFConfigParam_AlertString) Format() string {
	return fmt.Sprintf(`AlertStringSelector: %d, BlockSelector: %d, StringData: %s`, param.SetSelector, param.BlockSelector, string(param.StringData))
}

// READ ONLY
type PEFConfigParam_GroupControlsCount struct {
	Value uint8
}

func (param *PEFConfigParam_GroupControlsCount) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_GroupControlsCount, 0, 0
}

func (param *PEFConfigParam_GroupControlsCount) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_GroupControlsCount) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_GroupControlsCount) Format() string {
	return fmt.Sprintf("%d", param.Value)
}

type PEFConfigParam_GroupControl struct {
	// Set Selector (Entry Selector) = group control table entry selector.
	SetSelector uint8

	ForceControlOperation bool
	DelayedControl        bool
	ChannelNumber         uint8

	GroupID0              uint8
	MemberID0             uint8
	DisableMemberID0Check bool

	GroupID1              uint8
	MemberID1             uint8
	DisableMemberID1Check bool

	GroupID2              uint8
	MemberID2             uint8
	DisableMemberID2Check bool

	GroupID3              uint8
	MemberID3             uint8
	DisableMemberID3Check bool

	RetryCount uint8

	Operation uint8
}

func (param *PEFConfigParam_GroupControl) PEFConfigParameter() (paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) {
	return PEFConfigParamSelector_GroupControl, param.SetSelector, 0
}

func (param *PEFConfigParam_GroupControl) Unpack(data []byte) error {
	if len(data) < 11 {
		return ErrUnpackedDataTooShortWith(len(data), 11)
	}

	param.SetSelector = data[0]

	param.ForceControlOperation = isBit5Set(data[1])
	param.DelayedControl = isBit4Set(data[1])
	param.ChannelNumber = data[1] & 0x0F

	param.GroupID0 = data[2]
	param.MemberID0 = data[3] & 0x0F
	param.DisableMemberID0Check = isBit4Set(data[3])

	param.GroupID1 = data[4]
	param.MemberID1 = data[5] & 0x0F
	param.DisableMemberID1Check = isBit4Set(data[5])

	param.GroupID2 = data[6]
	param.MemberID2 = data[7] & 0x0F
	param.DisableMemberID2Check = isBit4Set(data[7])

	param.GroupID3 = data[8]
	param.MemberID3 = data[9] & 0x0F
	param.DisableMemberID3Check = isBit4Set(data[9])

	// data 11: - Retries and Operation
	// [7] - reserved
	// [6:4] - number of times to retry sending the command to perform
	// the group operation [For ICMB, the BMC broadcasts a
	// Group Chassis Control command] (1-based)
	param.RetryCount = (data[10] & 0x7F) >> 4
	param.Operation = data[10] & 0x0F
	return nil
}

func (param *PEFConfigParam_GroupControl) Pack() []byte {
	var b uint8

	out := make([]byte, 11)
	out[0] = param.SetSelector

	b = param.ChannelNumber & 0x0F
	b = setOrClearBit5(b, param.ForceControlOperation)
	b = setOrClearBit4(b, param.DelayedControl)
	out[1] = b

	out[2] = param.GroupID0
	b = param.MemberID0 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID0Check)
	out[3] = b

	out[4] = param.GroupID1
	b = param.MemberID1 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID1Check)
	out[5] = b

	out[6] = param.GroupID2
	b = param.MemberID2 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID2Check)
	out[7] = b

	out[8] = param.GroupID3
	b = param.MemberID3 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID3Check)
	out[9] = b

	b = param.RetryCount << 4
	b |= param.Operation
	out[10] = b

	return out
}

func (param *PEFConfigParam_GroupControl) Format() string {
	return fmt.Sprintf(`
    EntrySelector:          %d
    ForceControlOperation:  %v
    DelayedControl:         %v
    ChannelNumber:          %d
    GroupID0:               %d
    MemberID0:              %d
    DisableMemberID0Check:  %v
    GroupID1:               %d
    MemberID1:              %d
    DisableMemberID1Check:  %v
    GroupID2:               %d
    MemberID2:              %d
    DisableMemberID2Check:  %v
    GroupID3:               %d
    MemberID3:              %d
    DisableMemberID3Check:  %v
    RetryCount:             %d
    Operation:              %d
`,
		param.SetSelector,
		param.ForceControlOperation,
		param.DelayedControl,
		param.ChannelNumber,
		param.GroupID0,
		param.MemberID0,
		param.DisableMemberID0Check,
		param.GroupID1,
		param.MemberID1,
		param.DisableMemberID1Check,
		param.GroupID2,
		param.MemberID2,
		param.DisableMemberID2Check,
		param.GroupID3,
		param.MemberID3,
		param.DisableMemberID3Check,
		param.RetryCount,
		param.Operation)
}
