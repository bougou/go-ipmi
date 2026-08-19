package sensor

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

// TestGetSensorReadingResponseThresholdStatus checks the threshold comparison
// status decoded from byte 3 of the Get Sensor Reading response (IPMI v2.0
// section 35.14), where bits 5-0 mean, in order, at or above UNR, UCR and UNC,
// and at or below LNR, LCR and LNC. Every asserted threshold must map to its own
// status, and when several are asserted at once the most severe one wins.
func TestGetSensorReadingResponseThresholdStatus(t *testing.T) {
	cases := []struct {
		name string
		// data is the response starting at the reading byte. Byte 1 has the
		// reading valid and sensor scanning enabled bits set (0xc0).
		data []byte
		want types.SensorThresholdStatus
	}{
		{
			name: "no threshold crossed",
			data: []byte{0x40, 0xc0, 0x00},
			want: types.SensorThresholdStatus_OK,
		},
		{
			name: "at or above upper non-critical",
			data: []byte{0x50, 0xc0, 0x08},
			want: types.SensorThresholdStatus_UNC,
		},
		{
			name: "at or above upper critical",
			data: []byte{0x60, 0xc0, 0x18},
			want: types.SensorThresholdStatus_UCR,
		},
		{
			name: "at or above upper non-recoverable",
			data: []byte{0x70, 0xc0, 0x38},
			want: types.SensorThresholdStatus_UNR,
		},
		{
			name: "at or below lower non-critical",
			data: []byte{0x20, 0xc0, 0x01},
			want: types.SensorThresholdStatus_LNC,
		},
		{
			name: "at or below lower critical",
			data: []byte{0x10, 0xc0, 0x03},
			want: types.SensorThresholdStatus_LCR,
		},
		{
			name: "at or below lower non-recoverable",
			data: []byte{0x00, 0xc0, 0x07},
			want: types.SensorThresholdStatus_LNR,
		},
		{
			// Only the critical bit asserted, without the non-critical one
			// below it: some BMCs report just the threshold that was crossed.
			name: "upper critical without upper non-critical",
			data: []byte{0x60, 0xc0, 0x10},
			want: types.SensorThresholdStatus_UCR,
		},
		{
			// A response without the optional threshold comparison byte.
			name: "no comparison status byte",
			data: []byte{0x40, 0xc0},
			want: types.SensorThresholdStatus_OK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := &GetSensorReadingResponse{}
			if err := res.Unpack(tc.data); err != nil {
				t.Fatalf("Unpack failed: %v", err)
			}
			if got := res.ThresholdStatus(); got != tc.want {
				t.Errorf("ThresholdStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}
