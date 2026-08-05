package types

import "testing"

func TestCommandKeyIgnoresName(t *testing.T) {
	named := CommandCloseSession
	raw := Command{ID: named.ID, NetFn: named.NetFn}
	if named.Key() != raw.Key() {
		t.Fatalf("Key() must ignore Name: named=%v raw=%v", named.Key(), raw.Key())
	}
	if StrCC(raw, 0x87) != StrCC(named, 0x87) {
		t.Fatalf("StrCC must resolve without Name")
	}
}

func TestStrCC(t *testing.T) {
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
			name:  "shared parameter-config set applies to sibling commands",
			cmd:   CommandSetPEFConfigParam,
			ccode: 0x83,
			want:  "Attempt to read write-only parameter",
		},
		{
			name:  "set-in-progress conflict on boot options",
			cmd:   CommandSetSystemBootOptions,
			ccode: 0x81,
			want:  "Attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state",
		},
		{
			name:  "SEL erase in progress",
			cmd:   CommandGetSELInfo,
			ccode: 0x81,
			want:  "Cannot execute command, SEL erase in progress",
		},
		{
			name:  "codes above 84h are reachable",
			cmd:   CommandCloseSession,
			ccode: 0x88,
			want:  "Invalid Session Handle in request",
		},
		{
			name:  "DCMI set power limit out of range",
			cmd:   CommandSetDCMIPowerLimit,
			ccode: 0x84,
			want:  "Power Limit out of range",
		},
		{
			name:  "Write FRU Data write-protected offset",
			cmd:   CommandWriteFRUData,
			ccode: 0x80,
			want:  "Write-protected offset",
		},
		{
			name:  "Partial Add SEL Entry record length mismatch",
			cmd:   CommandPartialAddSELEntry,
			ccode: 0x80,
			want:  "Record rejected due to mismatch between record length in header data and number of bytes written",
		},
		{
			name:  "Get Channel Access uses Get wording",
			cmd:   CommandGetChannelAccess,
			ccode: 0x82,
			want:  "Command not supported for selected channel (e.g. channel is session-less)",
		},
		{
			name:  "Get Command Sub-function Enables",
			cmd:   CommandGetCommandSubfunctionEnables,
			ccode: 0x80,
			want:  "Attempt to get an unsupported or un-configurable sub-function",
		},
		{
			name:  "Get Configurable Command Sub-functions has no command-specific codes",
			cmd:   CommandGetConfigurableCommandSubfunctions,
			ccode: 0x80,
			want:  "0x80",
		},
		{
			name:  "DCMI Get Asset Tag encoding",
			cmd:   CommandGetDCMIAssetTag,
			ccode: 0x81,
			want:  "Encoding type in FRU is BCD Plus",
		},
		{
			name:  "PET Acknowledge has no command-specific codes",
			cmd:   CommandPETAcknowledge,
			ccode: 0x81,
			want:  "0x81",
		},
		{
			name:  "Suspend BMC ARPs has no command-specific codes",
			cmd:   CommandSuspendARPs,
			ccode: 0x80,
			want:  "0x80",
		},
		{
			name:  "zero command falls back to hex for command-specific range",
			cmd:   Command{},
			ccode: 0x81,
			want:  "0x81",
		},
		{
			name:  "Set User Password command-specific code",
			cmd:   CommandSetUserPassword,
			ccode: 0x80,
			want:  "Password test failed. Password size correct, but password data does not match stored value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrCC(tt.cmd, tt.ccode); got != tt.want {
				t.Errorf("StrCC(%q, 0x%02x) = %q, want %q", tt.cmd.Name, tt.ccode, got, tt.want)
			}
			if _, ok := commandSpecificCC[tt.cmd.Key()]; ok {
				if CommandSpecificCC(tt.cmd) == nil {
					t.Errorf("CommandSpecificCC(%q) = nil, want command-specific map", tt.cmd.Name)
				}
			} else if CommandSpecificCC(tt.cmd) != nil {
				t.Errorf("CommandSpecificCC(%q) = non-nil, want nil when no command-specific codes", tt.cmd.Name)
			}
			if got := AllCC(tt.cmd)[tt.ccode]; tt.want != "0x80" && tt.want != "0x81" && got != tt.want {
				// AllCC only has named entries; hex fallbacks are String()-only.
				m := CommandSpecificCC(tt.cmd)
				if _, ok := genericCC[tt.ccode]; ok || (m != nil && m[tt.ccode] != "") {
					if got != tt.want {
						t.Errorf("AllCC(%q)[0x%02x] = %q, want %q", tt.cmd.Name, tt.ccode, got, tt.want)
					}
				}
			}
		})
	}
}
