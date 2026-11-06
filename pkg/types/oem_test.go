package types

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestOEMString(t *testing.T) {
	tests := []struct {
		name string
		oem  OEM
		want string
	}{
		{name: "unknown", oem: OEM_UNKNOWN, want: "Unknown"},
		{name: "IPMI reserved", oem: OEM_RESERVED, want: "Unspecified"},
		{name: "debug", oem: OEM_DEBUG, want: "A Debug Assisting Company, Ltd."},
		{name: "entry absent from old table", oem: OEM(1), want: "NxNetworks"},
		{name: "Dell", oem: OEM_DELL, want: "Dell Inc."},
		{name: "gigabyte", oem: OEM_GIGABYTE, want: "GIGA-BYTE TECHNOLOGY CO., LTD"},
		{name: "registry overrides legacy constant name", oem: OEM_SUPERMICRO_47488, want: "Nedam ENG. Co., ltd."},
		{name: "id not in registry", oem: OEM(0x0FFFFE), want: "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.oem.String(); got != tt.want {
				t.Errorf("OEM(%d).String() = %q, want %q", uint32(tt.oem), got, tt.want)
			}
		})
	}
}

func TestOEMNamesMatchIANARegistry(t *testing.T) {
	want := readIANARegistryForTest(t, "data/enterprise-numbers.txt")
	got, err := loadOEMNames()
	if err != nil {
		t.Fatalf("loadOEMNames() error = %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("embedded registry has %d entries, source registry has %d", len(got), len(want))
	}
	for oem, wantName := range want {
		if gotName, ok := got[oem]; !ok || gotName != wantName {
			t.Errorf("OEM(%d) = %q, %v; source registry = %q", uint32(oem), gotName, ok, wantName)
		}
	}
}

func TestParseOEMNamesRejectsInvalidRecords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing separator", input: "1 NxNetworks\n"},
		{name: "empty organization", input: "1\t\n"},
		{name: "invalid PEN", input: "x\tNxNetworks\n"},
		{name: "duplicate PEN", input: "1\tNxNetworks\n1\tOther\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseOEMNames(strings.NewReader(tt.input)); err == nil {
				t.Fatal("parseOEMNames() error = nil, want invalid record error")
			}
		})
	}
}

func readIANARegistryForTest(t *testing.T, path string) map[OEM]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	entries := make(map[OEM]string)
	scanner := bufio.NewScanner(f)
	var pending OEM
	havePending := false
	for scanner.Scan() {
		line := scanner.Text()
		if havePending {
			havePending = false
			if len(line) >= 3 && line[0] == ' ' && line[1] == ' ' && line[2] != ' ' {
				entries[pending] = line[2:]
			}
			continue
		}
		if line == "" || line[0] < '0' || line[0] > '9' {
			continue
		}
		end := 0
		for end < len(line) && line[end] >= '0' && line[end] <= '9' {
			end++
		}
		id, err := strconv.ParseUint(line[:end], 10, 32)
		if err != nil {
			t.Fatalf("parse PEN %q: %v", line[:end], err)
		}
		pending = OEM(id)
		havePending = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return entries
}
