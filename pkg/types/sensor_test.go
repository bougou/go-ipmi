package types

import (
	"fmt"
	"math"
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
