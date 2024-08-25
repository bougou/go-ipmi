package ipmi

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_ParseGUID(t *testing.T) {
	tests := []struct {
		input     [16]byte
		expecteds map[GUIDMode]string
	}{
		{
			input: [16]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			expecteds: map[GUIDMode]string{
				GUIDModeRFC4122: "00112233-4455-6677-8899-AABBCCDDEEFF",
				GUIDModeIPMI:    "FFEEDDCC-BBAA-9988-7766-554433221100",
				GUIDModeSMBIOS:  "33221100-5544-7766-8899-AABBCCDDEEFF",
			},
		},
	}

	for _, tt := range tests {
		for _, mode := range []GUIDMode{
			GUIDModeRFC4122,
			GUIDModeIPMI,
			GUIDModeSMBIOS,
		} {
			u, err := ParseGUID(tt.input[:], mode)
			if err != nil {
				t.Error(err)
				return
			}
			actual := strings.ToUpper(u.String())
			expected := tt.expecteds[mode]

			sec, nsec := u.Time().UnixTime()
			uTime := time.Unix(sec, nsec)

			fmt.Printf("mode: %s, string: %s, version: %s, variant: %s, time: %v\n", mode, actual, u.Version(), u.Variant(), uTime)
			if actual != expected {
				t.Errorf("not matched for GUIDMode (%s), actual (%s), expected (%s)", mode, actual, expected)
				return
			}
		}
	}
}

func Test_ParseGUID2(t *testing.T) {
	tests := []struct {
		inputs   map[GUIDMode][16]byte
		mode     GUIDMode
		expected string
	}{
		{
			inputs: map[GUIDMode][16]byte{
				GUIDModeRFC4122: {0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
				GUIDModeIPMI:    {0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00},
				GUIDModeSMBIOS:  {0x33, 0x22, 0x11, 0x00, 0x55, 0x44, 0x77, 0x66, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			},
			expected: "00112233-4455-6677-8899-AABBCCDDEEFF",
		},
	}

	for _, tt := range tests {
		for mode, input := range tt.inputs {
			u, err := ParseGUID(input[:], mode)
			if err != nil {
				t.Error(err)
				return
			}
			actual := strings.ToUpper(u.String())
			expected := tt.expected

			sec, nsec := u.Time().UnixTime()
			uTime := time.Unix(sec, nsec)

			fmt.Printf("mode: %s, string: %s, version: %s, variant: %s, time: %v\n", mode, actual, u.Version(), u.Variant(), uTime)
			if actual != expected {
				t.Errorf("not matched for GUIDMode (%s), actual (%s), expected (%s)", mode, actual, expected)
				return
			}
		}

	}
}
