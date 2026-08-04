package types

import "testing"

func TestStrCCForCommand(t *testing.T) {
	tests := []struct {
		name  string
		cmd   Command
		ccode uint8
		want  string
	}{
		{
			name:  "command-specific code named per command",
			cmd:   CommandSetSystemBootOptions,
			ccode: 0x80,
			want:  "Parameter not supported",
		},
		{
			name:  "command-specific code named for get variant",
			cmd:   CommandGetSystemBootOptions,
			ccode: 0x80,
			want:  "Parameter not supported",
		},
		{
			name:  "generic code named regardless of command",
			cmd:   CommandSetSystemBootOptions,
			ccode: 0xc0,
			want:  "Node busy",
		},
		{
			name:  "command-specific code on another command falls back to hex",
			cmd:   CommandChassisControl,
			ccode: 0x80,
			want:  "0x80",
		},
		{
			name:  "zero command falls back to hex for command-specific range",
			cmd:   Command{},
			ccode: 0x81,
			want:  "0x81",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrCCForCommand(tt.cmd, tt.ccode); got != tt.want {
				t.Errorf("StrCCForCommand(%q, 0x%02x) = %q, want %q", tt.cmd.Name, tt.ccode, got, tt.want)
			}
		})
	}
}
