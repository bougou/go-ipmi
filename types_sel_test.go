package ipmi

import (
	"encoding/hex"
	"testing"
)

func TestParseSEL(t *testing.T) {
	type args struct {
		msg []byte
	}
	s, _ := hex.DecodeString("4d150290b3c66741000409010b03ffff")
	tests := []struct {
		name         string
		args         args
		wantSeverity EventSeverity
		wantErr      bool
	}{
		{
			name: "SEL",
			args: args{
				msg: s,
			},
			wantSeverity: EventSeverityCritical,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSEL(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSEL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sel.Standard.EventSeverity() != tt.wantSeverity {
				t.Errorf("ParseSEL() = %v, want %v", sel.Standard.EventSeverity(), tt.wantSeverity)
			}
		})
	}
}
