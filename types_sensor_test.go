package ipmi

import (
	"fmt"
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
