package ipmi

import (
	"reflect"
	"testing"
)

func TestTypeLength_Chars(t *testing.T) {
	t.Parallel()

	type args struct {
		raw []byte
	}

	tests := []struct {
		name      string
		tl        TypeLength
		args      args
		wantChars []byte
		wantErr   bool
	}{
		{
			name:      "empty tl & raw",
			tl:        0x00,
			args:      args{raw: []byte{}},
			wantChars: []byte{},
			wantErr:   false,
		},
		{
			name:      "differen lenght for tl & raw",
			tl:        0xc0,
			args:      args{raw: []byte{57, 84}},
			wantChars: nil,
			wantErr:   true,
		},
		{
			name:      "valid 8-bit ASCII",
			tl:        0xcd,
			args:      args{raw: []byte{89, 49, 85, 85, 66, 51, 79, 71, 88, 89, 48, 75, 70}},
			wantChars: []byte("Y1UUB3OGXY0KF"),
			wantErr:   false,
		},
		{
			name:      "untrimmed 8-bit ASCII",
			tl:        0xc8,
			args:      args{raw: []byte{83, 97, 109, 115, 117, 110, 103, 0}},
			wantChars: []byte{83, 97, 109, 115, 117, 110, 103, 0}, // "Samsung\0"
			wantErr:   false,
		},
		{
			// Every 3 bytes contains 4 chars.
			// 'Y' in 6-bit ASCII is 111001
			// Three 'Y's are 01111001 10011110 00000011
			name:      "3-symbol word (empty 4th char) 6-bit ASCII",
			tl:        0x83,
			args:      args{raw: []byte{121, 158, 3}},
			wantChars: []byte("YYY "),
			wantErr:   false,
		},
		{
			// Four 'Y's are 01111001 10011110 11100111
			name:      "4-symbol word (full 3 bytes) 6-bit ASCII",
			tl:        0x83,
			args:      args{raw: []byte{121, 158, 231}},
			wantChars: []byte("YYYY"),
			wantErr:   false,
		},
		{
			// Five 'Y's are 01111001 10011110 11100111 00111001
			name:      "5-symbol word (not a multiple of 3 bytes) 6-bit ASCII",
			tl:        0x84,
			args:      args{raw: []byte{121, 158, 231, 57}},
			wantChars: []byte("YYYYY"),
			wantErr:   false,
		},
		{
			// Six 'Y's are 01111001 10011110 11100111 01111001 00001110
			name:      "6-symbol word (not a multiple of 3 bytes) 6-bit ASCII",
			tl:        0x85,
			args:      args{raw: []byte{121, 158, 231, 121, 14}},
			wantChars: []byte("YYYYYY"),
			wantErr:   false,
		},
		{
			// Seven 'Y's are 01111001 10011110 11100111 01111001 10011110 00000011
			name:      "7-symbol word (empty 8th char) 6-bit ASCII",
			tl:        0x86,
			args:      args{raw: []byte{121, 158, 231, 121, 158, 3}},
			wantChars: []byte("YYYYYYY "),
			wantErr:   false,
		},
		{
			name:      "valid 6-bit ASCII",
			tl:        0x8a,
			args:      args{raw: []byte{57, 100, 143, 34, 69, 89, 82, 13, 73, 16}},
			wantChars: []byte{89, 48, 86, 67, 66, 52, 52, 54, 50, 85, 48, 50, 48}, // Y0VCB4462U020
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotChars, err := tt.tl.Chars(tt.args.raw)

			if (err != nil) != tt.wantErr {
				t.Errorf("TypeLength.Chars() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(gotChars, tt.wantChars) {
				t.Errorf("TypeLength.Chars() = %#v (%s), want %#v (%s)", gotChars, string(gotChars), tt.wantChars, string(tt.wantChars))
			}
		})
	}
}
