package iana

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLookupMatchesRegistry(t *testing.T) {
	want := readIANARegistryForTest(t, "data/enterprise-numbers.txt")
	if got := len(registryNames()); got != len(want) {
		t.Fatalf("embedded %d entries, want %d", got, len(want))
	}
	for id, name := range want {
		if got, ok := Lookup(id); !ok || got != name {
			t.Fatalf("PEN %d: got %q, %v; want %q", id, got, ok, name)
		}
	}
	if _, ok := Lookup(^uint32(0)); ok {
		t.Fatal("unexpected entry for unassigned PEN")
	}
}

func readIANARegistryForTest(t *testing.T, path string) map[uint32]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	entries := make(map[uint32]string)
	scanner := bufio.NewScanner(f)
	var pending uint32
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
		pending = uint32(id)
		havePending = true
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return entries
}

func BenchmarkLookup(b *testing.B) {
	Lookup(674) // Measure cached lookups.
	b.ReportAllocs()
	for b.Loop() {
		Lookup(674)
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
			if _, err := parseNames(strings.NewReader(tt.input)); err == nil {
				t.Fatal("parseNames() error = nil, want invalid record error")
			}
		})
	}
}
