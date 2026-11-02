package chassis

import (
	"strings"
	"testing"
)

// TestGetChassisStatusResponseFormatLastPowerEvent checks that Format reports
// the Last Power Event byte (IPMI v2.0 28.2, Table 28-3, byte 2) with the
// keywords ipmitool prints for "chassis status".
func TestGetChassisStatusResponseFormatLastPowerEvent(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{
			// Powered on with restore policy "previous" after AC/mains returned.
			name: "ac failed",
			data: []byte{0x21, 0x01, 0x00, 0xf0},
			want: "Last Power Event     : ac-failed",
		},
		{
			// Chassis intrusion is byte 3 and is not a power event.
			name: "chassis intrusion only",
			data: []byte{0x21, 0x00, 0x01},
			want: "Last Power Event     : ",
		},
		{
			name: "power on by command",
			data: []byte{0x01, 0x10, 0x00},
			want: "Last Power Event     : command",
		},
		{
			name: "all events",
			data: []byte{0x21, 0x1f, 0x00},
			want: "Last Power Event     : ac-failed overload interlock fault command",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var res GetChassisStatusResponse
			if err := res.Unpack(tc.data); err != nil {
				t.Fatalf("Unpack: %v", err)
			}
			out := res.Format()
			for _, line := range strings.Split(out, "\n") {
				if strings.HasPrefix(line, "Last Power Event") {
					if line != tc.want {
						t.Fatalf("want %q, got %q", tc.want, line)
					}
					return
				}
			}
			t.Fatalf("no Last Power Event line in:\n%s", out)
		})
	}
}
