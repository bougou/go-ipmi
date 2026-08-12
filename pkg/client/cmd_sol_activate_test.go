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
		{name: "terminate", input: "~X"},
		{name: "terminate after newline", input: "data\n~X", want: "data\n"},
		{name: "ordinary tilde", input: "~x", want: "~x"},
		{name: "single tilde dot", input: "~.", want: "~."},
		{name: "transparent outside line start", input: "data~X", want: "data~X"},
		{name: "insufficient characters", input: "~", want: "~"},
		{name: "tilde and lowercase x", input: "~x", want: "~x"},
		{name: "tilde and other char", input: "~?", want: "~?"},
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
