//go:build windows
// +build windows

package open

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseByteCSV(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{name: "empty", input: "", want: []byte{}},
		{name: "whitespace only", input: "   ", want: []byte{}},
		{name: "single byte", input: "0", want: []byte{0}},
		{name: "simple list", input: "1,2,3", want: []byte{1, 2, 3}},
		{name: "with spaces", input: " 1 , 2 , 3 ", want: []byte{1, 2, 3}},
		{name: "max byte", input: "255", want: []byte{0xff}},
		{name: "overflow rejected", input: "256", wantErr: true},
		{name: "non-numeric rejected", input: "abc", wantErr: true},
		{name: "negative rejected", input: "-1", wantErr: true},
		{name: "mixed valid invalid", input: "1,bad,3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseByteCSV(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len mismatch: want %v, got %v", tt.want, got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("byte %d: want %v, got %v", i, tt.want[i], got[i])
				}
			}
		})
	}
}

// TestParseWinIPMIResponse pins the transport contract of
// PowerShellConn.SendCommand's response handling:
//
//   - The Microsoft_IPMI provider returns the FULL IPMI response in
//     ResponseData with the completion code as byte 0. The bytes are
//     returned as-is — nothing is prepended.
//   - Non-zero completion codes are NOT wrapped into errors at this layer;
//     the client layer (exchangeOpen) reads recv[0] and wraps them as
//     ResponseError, matching the Linux and COM contracts.
//   - The CompletionCode JSON property mirrors ResponseData[0] and is
//     deliberately not consulted.
func TestParseWinIPMIResponse(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    []byte
		wantErr string
	}{
		{
			name: "success, cc 0x00 first then payload",
			json: `{"CompletionCode":0,"ResponseData":"0,17,34,51","ResponseDataSize":4}`,
			want: []byte{0x00, 17, 34, 51},
		},
		{
			name: "non-zero cc returned as-is, not wrapped",
			json: `{"CompletionCode":192,"ResponseData":"192","ResponseDataSize":1}`,
			want: []byte{0xC0},
		},
		{
			name: "CompletionCode property is ignored",
			// Contrived mismatch: the returned bytes still come verbatim
			// from ResponseData; only ResponseData defines the output.
			json: `{"CompletionCode":0,"ResponseData":"204","ResponseDataSize":1}`,
			want: []byte{0xCC},
		},
		{
			name: "ResponseDataSize truncates trailing bytes",
			json: `{"CompletionCode":0,"ResponseData":"0,1,2,3,4","ResponseDataSize":3}`,
			want: []byte{0, 1, 2},
		},
		{
			name:    "empty ResponseData rejected (no completion code)",
			json:    `{"CompletionCode":0,"ResponseData":"","ResponseDataSize":0}`,
			wantErr: "ResponseData empty",
		},
		{
			name:    "invalid JSON rejected",
			json:    `not json`,
			wantErr: "parse Microsoft_IPMI response failed",
		},
		{
			name:    "invalid byte in CSV rejected",
			json:    `{"CompletionCode":0,"ResponseData":"0,bad","ResponseDataSize":2}`,
			wantErr: "parse Microsoft_IPMI response data failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWinIPMIResponse([]byte(tt.json))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("want error %q, got result %v", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}
