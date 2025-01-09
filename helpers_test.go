package ipmi

import (
	"testing"
)

func Test_onesComplement(t *testing.T) {
	// https://en.wikipedia.org/wiki/Ones%27_complement
	tests := []struct {
		name     string
		input    uint32
		bitSize  uint8
		expected int32
	}{
		{"test1", 127, 8, 127},
		{"test2", 126, 8, 126},
		{"test3", 2, 8, 2},
		{"test4", 1, 8, 1},
		{"test5", 0, 8, 0},
		{"test6", 255, 8, -0},
		{"test7", 254, 8, -1},
		{"test8", 253, 8, -2},
		{"test9", 129, 8, -126},
		{"test10", 128, 8, -127},
	}

	for _, test := range tests {
		got := onesComplement(test.input, test.bitSize)
		if got != test.expected {
			t.Errorf("test %s not matched, got: %d, expected: %d", test.name, got, test.expected)
		}
	}
}

func Test_twosComplement(t *testing.T) {
	// https://en.wikipedia.org/wiki/Two%27s_complement
	tests := []struct {
		name     string
		input    uint32
		bitSize  uint8
		expected int32
	}{
		{"test1", 0, 8, 0},
		{"test2", 1, 8, 1},
		{"test3", 2, 8, 2},
		{"test4", 126, 8, 126},
		{"test5", 127, 8, 127},
		{"test6", 128, 8, -128},
		{"test7", 129, 8, -127},
		{"test8", 130, 8, -126},
		{"test9", 254, 8, -2},
		{"test10", 255, 8, -1},
	}

	for _, test := range tests {
		got := twosComplement(test.input, test.bitSize)
		if got != test.expected {
			t.Errorf("test %s not matched, got: %d, expected: %d", test.name, got, test.expected)
		}
	}
}

func Test_onesComplementEncode(t *testing.T) {
	// https://en.wikipedia.org/wiki/Ones%27_complement
	tests := []struct {
		name     string
		expected uint32
		bitSize  uint8
		input    int32
	}{
		{"test1", 127, 8, 127},
		{"test2", 126, 8, 126},
		{"test3", 2, 8, 2},
		{"test4", 1, 8, 1},
		{"test5", 0, 8, 0},
		// {"test6", 255, 8, -0},
		{"test7", 254, 8, -1},
		{"test8", 253, 8, -2},
		{"test9", 129, 8, -126},
		{"test10", 128, 8, -127},
	}

	for _, test := range tests {
		got := onesComplementEncode(test.input, test.bitSize)
		if got != test.expected {
			t.Errorf("test %s not matched, got: %d, expected: %d", test.name, got, test.expected)
		}
	}
}

func Test_twosComplementEncode(t *testing.T) {
	// https://en.wikipedia.org/wiki/Two%27s_complement
	tests := []struct {
		name     string
		expected uint32
		bitSize  uint8
		input    int32
	}{
		{"test1", 0, 8, 0},
		{"test2", 1, 8, 1},
		{"test3", 2, 8, 2},
		{"test4", 126, 8, 126},
		{"test5", 127, 8, 127},
		{"test6", 128, 8, -128},
		{"test7", 129, 8, -127},
		{"test8", 130, 8, -126},
		{"test9", 254, 8, -2},
		{"test10", 255, 8, -1},
	}

	for _, test := range tests {
		got := twosComplementEncode(test.input, test.bitSize)
		if got != test.expected {
			t.Errorf("test %s not matched, got: %d, expected: %d", test.name, got, test.expected)
		}
	}
}

func Test_unpackBytes(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
		input    []uint8
		offset   int
		length   int
	}{
		{"test1", []byte{0x03, 0x04}, []uint8{0x01, 0x02, 0x03, 0x04}, 2, 2},
		{"test2", []byte{0x04, 0x05, 0x06, 0x07}, []uint8{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, 3, 4},
		{"test3", []byte{}, []uint8{}, 0, 0},
		{"test4", []byte{0x01}, []uint8{0x01}, 0, 1},
	}

	for _, tt := range tests {
		got, _, err := unpackBytes(tt.input, tt.offset, tt.length)
		if err != nil {
			t.Errorf("test (%s) unpackBytes failed, err: %s", tt.name, err)
			return
		}
		if !isByteSliceEqual(got, tt.expected) {
			t.Errorf("test %s not matched, got: %v, expected: %v", tt.name, got, tt.expected)
			return
		}
	}
}
