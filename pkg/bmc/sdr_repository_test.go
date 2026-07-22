package bmc

import "testing"

func TestEncodeSDRRepoFreeSpace(t *testing.T) {
	cases := []struct {
		free int
		want uint16
	}{
		{0, 0},
		{-1, 0},
		{1, 1},
		{0xFFFD, 0xFFFD},
		{0xFFFE, 0xFFFE},
		{0xFFFF, 0xFFFE},
		{64 * 1024, 0xFFFE}, // v2.0§33.9: FFFEh = 64KB-2 or more
	}
	for _, tc := range cases {
		if got := encodeSDRRepoFreeSpace(tc.free); got != tc.want {
			t.Fatalf("encodeSDRRepoFreeSpace(%d): want %#04x got %#04x", tc.free, tc.want, got)
		}
	}
}
