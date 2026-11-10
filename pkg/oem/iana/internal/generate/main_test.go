package main

import (
	"io"
	"strings"
	"testing"
)

func TestWriteNamesRejectsInvalidRecords(t *testing.T) {
	for _, input := range []string{
		"1\n  First\n1\n  Second\n",
		"2\n  Later\n1\n  Earlier\n",
		"4294967296\n  Overflow\n",
		"1\n  \n",
		"1\n",
		"1\n  Invalid\tname\n",
	} {
		if err := writeNames(io.Discard, strings.NewReader(input)); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}
