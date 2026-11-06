package types

import (
	"fmt"
	"math"
	"slices"
	"testing"
)

func Test_ConvertReading(t *testing.T) {

	tests := []struct {
		raw               uint8
		analogDataFormat  SensorAnalogUnitFormat
		factors           ReadingFactors
		linearizationFunc LinearizationFunc
	}{
		{
			raw:               0,
			analogDataFormat:  SensorAnalogUnitFormat_Unsigned,
			factors:           ReadingFactors{},
			linearizationFunc: LinearizationFunc_Linear,
		},
		{
			raw:               0,
			analogDataFormat:  SensorAnalogUnitFormat_NotAnalog,
			factors:           ReadingFactors{},
			linearizationFunc: LinearizationFunc_Linear,
		},
		{
			raw:              0,
			analogDataFormat: SensorAnalogUnitFormat_1sComplement,
			factors: ReadingFactors{
				M:            1,
				Tolerance:    0,
				B:            0,
				Accuracy:     0,
				Accuracy_Exp: 0,
				B_Exp:        0,
			},
			linearizationFunc: LinearizationFunc_Linear,
		},
		{
			raw:               0,
			analogDataFormat:  SensorAnalogUnitFormat_2sComplement,
			factors:           ReadingFactors{},
			linearizationFunc: LinearizationFunc_Linear,
		},
	}

	for _, tt := range tests {
		v := ConvertReading(tt.raw, tt.analogDataFormat, tt.factors, tt.linearizationFunc)
		fmt.Println(v)
		// Todo
	}
}

func TestSensorEventsStrings(t *testing.T) {
	events := SensorEvents{
		SensorEvent_UNC_High_Assert,
		SensorEvent_UNC_Low_Deassert,
		SensorEvent_UCR_High_Assert,
		SensorEvent_LNC_Low_Deassert,
	}

	want := []string{"unc+", "unc-", "ucr+", "lnc-"}
	if got := events.Strings(); !slices.Equal(got, want) {
		t.Fatalf("Strings() = %v, want %v", got, want)
	}
}

func TestSensorThresholdString(t *testing.T) {
	s := &Sensor{EventReadingType: EventReadingTypeThreshold}

	s.Threshold.Mask = Mask_Thresholds{
		LNR: Mask_Threshold{Readable: true},
		LCR: Mask_Threshold{Readable: true},
		LNC: Mask_Threshold{Readable: true},
		UNC: Mask_Threshold{Readable: true, High_Assert: true, High_Deassert: true},
		UCR: Mask_Threshold{Readable: true, High_Assert: true, High_Deassert: true},
		UNR: Mask_Threshold{Readable: true},
	}

	s.Threshold.LNC = 0
	s.Threshold.LCR = 0
	s.Threshold.LNR = 0
	s.Threshold.UNC = 81
	s.Threshold.UCR = 83
	s.Threshold.UNR = 85

	s.Threshold.PositiveHysteresisRaw = 1
	s.Threshold.NegativeHysteresisRaw = 1
	s.Threshold.PositiveHysteresis = 1.0
	s.Threshold.NegativeHysteresis = 1.0

	want := "" +
		" Lower Non-Recoverable : 0.000\n" +
		" Lower Critical        : 0.000\n" +
		" Lower Non-Critical    : 0.000\n" +
		" Upper Non-Critical    : 81.000\n" +
		" Upper Critical        : 83.000\n" +
		" Upper Non-Recoverable : 85.000\n" +
		" Positive Hysteresis   : 1.000\n" +
		" Negative Hysteresis   : 1.000\n" +
		" Assertions Enabled    : unc+ ucr+\n" +
		" Deassertions Enabled  : unc+ ucr+\n"

	if got := s.thresholdString(); got != want {
		t.Fatalf("thresholdString() mismatch:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestSensorThresholdStringNonThreshold(t *testing.T) {
	s := &Sensor{EventReadingType: EventReadingTypeSensorSpecific}
	if got := s.thresholdString(); got != "" {
		t.Fatalf("thresholdString() for non-threshold sensor = %q, want empty", got)
	}
}

func TestConvertReadingEXP10FractionalExponent(t *testing.T) {
	tests := []struct {
		name string
		raw  uint8
		want float64
	}{
		{name: "positive", raw: 15, want: 31.622776601683793},
		{name: "negative", raw: 0xf1, want: 0.03162277660168379},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Scaling signed readings of +/-15 gives exponents of +/-1.5.
			got := ConvertReading(tt.raw, SensorAnalogUnitFormat_2sComplement,
				ReadingFactors{M: 1, R_Exp: -1}, LinearizationFunc_EXP10)
			if math.IsNaN(got) || math.Abs(got-tt.want) > 1e-12*tt.want {
				t.Fatalf("ConvertReading() = %.17g, want %.17g", got, tt.want)
			}
		})
	}
}
