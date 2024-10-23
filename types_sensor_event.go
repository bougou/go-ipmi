package ipmi

import "fmt"

type SensorEvent struct {
	SensorClass SensorClass

	ThresholdType SensorThresholdType
	Assert        bool // true -> assertion events; false -> deassertion events
	High          bool // true -> going high; false -> going low

	State uint8 // state 0-14 (total 15 possible states)
}

func (e SensorEvent) String() string {
	switch e.SensorClass {
	case SensorClassThreshold:
		out := e.ThresholdType.Abbr()
		if e.High {
			out += "+"
		} else {
			out += "-"
		}
		return out
	case SensorClassDiscrete:
		return fmt.Sprintf("state%d", e.State)
	}
	return ""
}

type SensorEvents []SensorEvent

func (events SensorEvents) Strings() []string {
	out := make([]string, 0)
	return out
}
func (events SensorEvents) FilterAssert() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.Assert {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterDeassert() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if !event.Assert {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterThreshold() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.SensorClass == SensorClassThreshold {
			out = append(out, event)
		}
	}
	return out
}

func (events SensorEvents) FilterDiscrete() SensorEvents {
	out := make([]SensorEvent, 0)
	for _, event := range events {
		if event.SensorClass == SensorClassDiscrete {
			out = append(out, event)
		}
	}
	return out
}

// SensorEventFlag holds a struct with fields indicating the specified sensor event is set or not.
// SensorEventFlag was embedded in Sensor related commands.
type SensorEventFlag struct {
	SensorEvent_UNC_High_Assert bool
	SensorEvent_UNC_Low_Assert  bool
	SensorEvent_LNR_High_Assert bool
	SensorEvent_LNR_Low_Assert  bool
	SensorEvent_LCR_High_Assert bool
	SensorEvent_LCR_Low_Assert  bool
	SensorEvent_LNC_High_Assert bool
	SensorEvent_LNC_Low_Assert  bool
	SensorEvent_State_7_Assert  bool
	SensorEvent_State_6_Assert  bool
	SensorEvent_State_5_Assert  bool
	SensorEvent_State_4_Assert  bool
	SensorEvent_State_3_Assert  bool
	SensorEvent_State_2_Assert  bool
	SensorEvent_State_1_Assert  bool
	SensorEvent_State_0_Assert  bool

	SensorEvent_UNR_High_Assert bool
	SensorEvent_UNR_Low_Assert  bool
	SensorEvent_UCR_High_Assert bool
	SensorEvent_UCR_Low_Assert  bool
	SensorEvent_State_14_Assert bool
	SensorEvent_State_13_Assert bool
	SensorEvent_State_12_Assert bool
	SensorEvent_State_11_Assert bool
	SensorEvent_State_10_Assert bool
	SensorEvent_State_9_Assert  bool
	SensorEvent_State_8_Assert  bool

	SensorEvent_UNC_High_Deassert bool
	SensorEvent_UNC_Low_Deassert  bool
	SensorEvent_LNR_High_Deassert bool
	SensorEvent_LNR_Low_Deassert  bool
	SensorEvent_LCR_High_Deassert bool
	SensorEvent_LCR_Low_Deassert  bool
	SensorEvent_LNC_High_Deassert bool
	SensorEvent_LNC_Low_Deassert  bool
	SensorEvent_State_7_Deassert  bool
	SensorEvent_State_6_Deassert  bool
	SensorEvent_State_5_Deassert  bool
	SensorEvent_State_4_Deassert  bool
	SensorEvent_State_3_Deassert  bool
	SensorEvent_State_2_Deassert  bool
	SensorEvent_State_1_Deassert  bool
	SensorEvent_State_0_Deassert  bool

	SensorEvent_UNR_High_Deassert bool
	SensorEvent_UNR_Low_Deassert  bool
	SensorEvent_UCR_High_Deassert bool
	SensorEvent_UCR_Low_Deassert  bool
	SensorEvent_State_14_Deassert bool
	SensorEvent_State_13_Deassert bool
	SensorEvent_State_12_Deassert bool
	SensorEvent_State_11_Deassert bool
	SensorEvent_State_10_Deassert bool
	SensorEvent_State_9_Deassert  bool
	SensorEvent_State_8_Deassert  bool
}

// TrueEvents returns a slice of SensorEvent those are set to true in the SensorEventFlag.
func (flag *SensorEventFlag) TrueEvents() []SensorEvent {
	out := make([]SensorEvent, 0)
	if flag.SensorEvent_UNC_High_Assert {
		out = append(out, SensorEvent_UNC_High_Assert)
	}
	if flag.SensorEvent_UNC_Low_Assert {
		out = append(out, SensorEvent_UNC_Low_Deassert)
	}
	if flag.SensorEvent_LNR_High_Assert {
		out = append(out, SensorEvent_LNR_High_Assert)
	}
	if flag.SensorEvent_LNR_Low_Assert {
		out = append(out, SensorEvent_LNR_Low_Assert)
	}
	if flag.SensorEvent_LCR_High_Assert {
		out = append(out, SensorEvent_LCR_High_Assert)
	}
	if flag.SensorEvent_LCR_Low_Assert {
		out = append(out, SensorEvent_LCR_Low_Assert)
	}
	if flag.SensorEvent_LNC_High_Assert {
		out = append(out, SensorEvent_LNC_High_Assert)
	}
	if flag.SensorEvent_LNC_Low_Assert {
		out = append(out, SensorEvent_LNC_Low_Assert)
	}
	if flag.SensorEvent_State_7_Assert {
		out = append(out, SensorEvent_State_7_Assert)
	}
	if flag.SensorEvent_State_6_Assert {
		out = append(out, SensorEvent_State_6_Assert)
	}
	if flag.SensorEvent_State_5_Assert {
		out = append(out, SensorEvent_State_5_Assert)
	}
	if flag.SensorEvent_State_4_Assert {
		out = append(out, SensorEvent_State_4_Assert)
	}
	if flag.SensorEvent_State_3_Assert {
		out = append(out, SensorEvent_State_3_Assert)
	}
	if flag.SensorEvent_State_2_Assert {
		out = append(out, SensorEvent_State_2_Assert)
	}
	if flag.SensorEvent_State_1_Assert {
		out = append(out, SensorEvent_State_1_Assert)
	}
	if flag.SensorEvent_State_0_Assert {
		out = append(out, SensorEvent_State_0_Assert)
	}
	if flag.SensorEvent_UNR_High_Assert {
		out = append(out, SensorEvent_UNR_High_Assert)
	}
	if flag.SensorEvent_UNR_Low_Assert {
		out = append(out, SensorEvent_UNR_Low_Assert)
	}
	if flag.SensorEvent_UCR_High_Assert {
		out = append(out, SensorEvent_UCR_High_Assert)
	}
	if flag.SensorEvent_UCR_Low_Assert {
		out = append(out, SensorEvent_UCR_Low_Assert)
	}
	if flag.SensorEvent_State_14_Assert {
		out = append(out, SensorEvent_State_14_Assert)
	}
	if flag.SensorEvent_State_13_Assert {
		out = append(out, SensorEvent_State_13_Assert)
	}
	if flag.SensorEvent_State_12_Assert {
		out = append(out, SensorEvent_State_12_Assert)
	}
	if flag.SensorEvent_State_11_Assert {
		out = append(out, SensorEvent_State_11_Assert)
	}
	if flag.SensorEvent_State_10_Assert {
		out = append(out, SensorEvent_State_10_Assert)
	}
	if flag.SensorEvent_State_9_Assert {
		out = append(out, SensorEvent_State_9_Assert)
	}
	if flag.SensorEvent_State_8_Assert {
		out = append(out, SensorEvent_State_8_Assert)
	}
	if flag.SensorEvent_UNC_High_Deassert {
		out = append(out, SensorEvent_UNC_High_Deassert)
	}
	if flag.SensorEvent_UNC_Low_Deassert {
		out = append(out, SensorEvent_UNC_Low_Deassert)
	}
	if flag.SensorEvent_LNR_High_Deassert {
		out = append(out, SensorEvent_LNR_High_Deassert)
	}
	if flag.SensorEvent_LNR_Low_Deassert {
		out = append(out, SensorEvent_LNR_Low_Deassert)
	}
	if flag.SensorEvent_LCR_High_Deassert {
		out = append(out, SensorEvent_LCR_High_Deassert)
	}
	if flag.SensorEvent_LCR_Low_Deassert {
		out = append(out, SensorEvent_LCR_Low_Deassert)
	}
	if flag.SensorEvent_LNC_High_Deassert {
		out = append(out, SensorEvent_LNC_High_Deassert)
	}
	if flag.SensorEvent_LNC_Low_Deassert {
		out = append(out, SensorEvent_LNC_Low_Deassert)
	}
	if flag.SensorEvent_State_7_Deassert {
		out = append(out, SensorEvent_State_7_Deassert)
	}
	if flag.SensorEvent_State_6_Deassert {
		out = append(out, SensorEvent_State_6_Deassert)
	}
	if flag.SensorEvent_State_5_Deassert {
		out = append(out, SensorEvent_State_5_Deassert)
	}
	if flag.SensorEvent_State_4_Deassert {
		out = append(out, SensorEvent_State_4_Deassert)
	}
	if flag.SensorEvent_State_3_Deassert {
		out = append(out, SensorEvent_State_3_Deassert)
	}
	if flag.SensorEvent_State_2_Deassert {
		out = append(out, SensorEvent_State_2_Deassert)
	}
	if flag.SensorEvent_State_1_Deassert {
		out = append(out, SensorEvent_State_1_Deassert)
	}
	if flag.SensorEvent_State_0_Deassert {
		out = append(out, SensorEvent_State_0_Deassert)
	}
	if flag.SensorEvent_UNR_High_Deassert {
		out = append(out, SensorEvent_UNR_High_Deassert)
	}
	if flag.SensorEvent_UNR_Low_Deassert {
		out = append(out, SensorEvent_UNR_Low_Deassert)
	}
	if flag.SensorEvent_UCR_High_Deassert {
		out = append(out, SensorEvent_UCR_High_Deassert)
	}
	if flag.SensorEvent_UCR_Low_Deassert {
		out = append(out, SensorEvent_UCR_Low_Deassert)
	}
	if flag.SensorEvent_State_14_Deassert {
		out = append(out, SensorEvent_State_14_Deassert)
	}
	if flag.SensorEvent_State_13_Deassert {
		out = append(out, SensorEvent_State_13_Deassert)
	}
	if flag.SensorEvent_State_12_Deassert {
		out = append(out, SensorEvent_State_12_Deassert)
	}
	if flag.SensorEvent_State_11_Deassert {
		out = append(out, SensorEvent_State_11_Deassert)
	}
	if flag.SensorEvent_State_10_Deassert {
		out = append(out, SensorEvent_State_10_Deassert)
	}
	if flag.SensorEvent_State_9_Deassert {
		out = append(out, SensorEvent_State_9_Deassert)
	}
	if flag.SensorEvent_State_8_Deassert {
		out = append(out, SensorEvent_State_8_Deassert)
	}
	return out
}

var (
	SensorEvent_UNC_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UNC_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LNR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LNR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LCR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LCR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_LNC_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        true,
		High:          true,
	}

	SensorEvent_LNC_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        true,
		High:          false,
	}

	SensorEvent_UNR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UNR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_UCR_High_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        true,
		High:          true,
	}

	SensorEvent_UCR_Low_Assert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        true,
		High:          false,
	}

	SensorEvent_State_14_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       14,
	}

	SensorEvent_State_13_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       13,
	}

	SensorEvent_State_12_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       12,
	}

	SensorEvent_State_11_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       11,
	}

	SensorEvent_State_10_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       10,
	}

	SensorEvent_State_9_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       9,
	}

	SensorEvent_State_8_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       8,
	}

	SensorEvent_State_7_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       7,
	}

	SensorEvent_State_6_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       6,
	}

	SensorEvent_State_5_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       5,
	}

	SensorEvent_State_4_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       4,
	}

	SensorEvent_State_3_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       3,
	}

	SensorEvent_State_2_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       2,
	}

	SensorEvent_State_1_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       1,
	}

	SensorEvent_State_0_Assert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      true,
		State:       0,
	}

	SensorEvent_UNC_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        false,
		High:          true,
	}

	// Deassert Events

	SensorEvent_UNC_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNC,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_LCR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LCR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LCR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_LNC_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        false,
		High:          true,
	}

	SensorEvent_LNC_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_LNC,
		Assert:        false,
		High:          false,
	}

	SensorEvent_UNR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_UNR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UNR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_UCR_High_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        false,
		High:          true,
	}

	SensorEvent_UCR_Low_Deassert = SensorEvent{
		SensorClass:   SensorClassThreshold,
		ThresholdType: SensorThresholdType_UCR,
		Assert:        false,
		High:          false,
	}

	SensorEvent_State_14_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       14,
	}

	SensorEvent_State_13_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       13,
	}

	SensorEvent_State_12_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       12,
	}

	SensorEvent_State_11_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       11,
	}

	SensorEvent_State_10_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       10,
	}

	SensorEvent_State_9_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       9,
	}

	SensorEvent_State_8_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       8,
	}

	SensorEvent_State_7_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       7,
	}

	SensorEvent_State_6_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       6,
	}

	SensorEvent_State_5_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       5,
	}

	SensorEvent_State_4_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       4,
	}

	SensorEvent_State_3_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       3,
	}

	SensorEvent_State_2_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       2,
	}

	SensorEvent_State_1_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       1,
	}

	SensorEvent_State_0_Deassert = SensorEvent{
		SensorClass: SensorClassDiscrete,
		Assert:      false,
		State:       0,
	}
)
