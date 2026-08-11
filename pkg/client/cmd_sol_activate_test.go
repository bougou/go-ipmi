package client

import (
	"io"
	"strings"
	"testing"
)

func TestSOLEscapeReader(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "terminate", input: "~."},
		{name: "terminate after newline", input: "data\n~.", want: "data\n"},
		{name: "ordinary tilde", input: "~x", want: "~x"},
		{name: "repeated escape prefix", input: "~~.", want: "~"},
		{name: "transparent outside line start", input: "data~.", want: "data~."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := io.ReadAll(newSOLEscapeReader(strings.NewReader(tt.input)))
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
