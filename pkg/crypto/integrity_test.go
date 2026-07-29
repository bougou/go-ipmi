package crypto

import (
	"bytes"
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestIntegrityPadLen(t *testing.T) {
	// header(12) + payload(5) + padLen(1) + next(1) = 19 → need 1 pad to reach 20
	if got := types.IntegrityPadLen(12, 5); got != 1 {
		t.Fatalf("IntegrityPadLen(12,5)=%d, want 1", got)
	}
	// 12+4+1+1 = 18 → need 2
	if got := types.IntegrityPadLen(12, 4); got != 2 {
		t.Fatalf("IntegrityPadLen(12,4)=%d, want 2", got)
	}
	// 12+2+1+1 = 16 → already aligned
	if got := types.IntegrityPadLen(12, 2); got != 0 {
		t.Fatalf("IntegrityPadLen(12,2)=%d, want 0", got)
	}
}

func TestIntegrityAuthCodeLen(t *testing.T) {
	cases := []struct {
		alg    types.IntegrityAlg
		want   int
		wantOK bool
	}{
		{types.IntegrityAlg_None, 0, true},
		{types.IntegrityAlg_HMAC_SHA1_96, 12, true},
		{types.IntegrityAlg_HMAC_MD5_128, 16, true},
		{types.IntegrityAlg_MD5_128, 16, true},
		{types.IntegrityAlg_HMAC_SHA256_128, 16, true},
		{types.IntegrityAlg(0xff), 0, false},
	}
	for _, tc := range cases {
		got, ok := IntegrityAuthCodeLen(tc.alg)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("IntegrityAuthCodeLen(%d)=(%d,%v), want (%d,%v)", tc.alg, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestSessionIntegrityAuthCode_HMACRoundTrip(t *testing.T) {
	k1 := []byte("0123456789abcdefghij")
	input := []byte{0x06, 0x00, 0x00, 0x00, 0x11, 0x22, 0x33, 0x44, 0x01, 0x00, 0x00, 0x00, 0xaa, 0xbb}

	for _, alg := range []types.IntegrityAlg{
		types.IntegrityAlg_HMAC_SHA1_96,
		types.IntegrityAlg_HMAC_SHA256_128,
		types.IntegrityAlg_HMAC_MD5_128,
	} {
		code, err := SessionIntegrityAuthCode(alg, input, k1, "")
		if err != nil {
			t.Fatalf("alg %d: %v", alg, err)
		}
		wantLen, ok := IntegrityAuthCodeLen(alg)
		if !ok || len(code) != wantLen {
			t.Fatalf("alg %d: len=%d, want %d", alg, len(code), wantLen)
		}
		again, err := SessionIntegrityAuthCode(alg, input, k1, "")
		if err != nil || !bytes.Equal(code, again) {
			t.Fatalf("alg %d: not deterministic", alg)
		}
	}
}
