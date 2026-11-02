package types

import "testing"

func TestOEMString(t *testing.T) {
	tests := []struct {
		name string
		oem  OEM
		want string
	}{
		{
			name: "gigabyte",
			oem:  OEM(15370),
			want: "GIGA-BYTE TECHNOLOGY CO., LTD",
		},
		{
			name: "id not in table",
			oem:  OEM(0x0FFFFE),
			want: "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.oem.String(); got != tt.want {
				t.Errorf("OEM(%d).String() = %q, want %q", uint32(tt.oem), got, tt.want)
			}
		})
	}
}
