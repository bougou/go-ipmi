package md2

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// RFC 1319 test vectors.
func TestMD2_RFC1319_Vectors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "8350e5a3e24c153df2275c9f80692773"},
		{"a", "32ec01ec4a6dac72c0ab96fb34c0b5d1"},
		{"abc", "da853b0d3f88d99b30283a69e6ded6bb"},
		{"message digest", "ab4f496bfb2a530b219ff33031fe06b0"},
		{"abcdefghijklmnopqrstuvwxyz", "4e8ddff3650292ab5a4108c3aa47940b"},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", "da33def2a42df13975352846c30338cd"},
		{"12345678901234567890123456789012345678901234567890123456789012345678901234567890", "d5976f79d83d3a0dc9806c3c66f3efd8"},
	}
	for _, tc := range tests {
		h := New()
		h.Write([]byte(tc.input))
		got := hex.EncodeToString(h.Sum(nil))
		if got != tc.want {
			t.Errorf("MD2(%q) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

// TestMD2_SumWithoutWrite demonstrates the bug: calling Sum(input) on a
// fresh hasher WITHOUT Write does NOT hash input. It computes MD2("") and
// appends it to input — input bytes never enter the hash state.
func TestMD2_SumWithoutWrite_IsBroken(t *testing.T) {
	payload := []byte("this data is NOT hashed")

	// Correct usage.
	good := New()
	good.Write(payload)
	correct := good.Sum(nil)

	// Broken usage: the pattern found in the original auth code.
	broken := New().Sum(payload)
	broken = broken[:Size] // original code sliced to 16 bytes

	// Correct hash should differ from the first 16 bytes of payload.
	if bytes.Equal(correct, payload[:Size]) {
		t.Errorf("correct MD2 hash happens to equal first 16 bytes of input? extremely unlikely")
	}

	// Broken "hash" is just the first 16 bytes of input (plus MD2("") appended),
	// so after [:Size] it's the cleartext input prefix.
	if !bytes.Equal(broken, payload[:Size]) {
		t.Errorf("New().Sum(payload)[:16] = %x, want payload[:16] = %x — "+
			"expected Sum to NOT hash the argument", broken, payload[:Size])
	}

	// Hard proof: broken == payload[:16] means the auth code was cleartext password.
	t.Logf("correct MD2(input) = %x", correct)
	t.Logf("broken New().Sum(input)[:16] = %x (cleartext input prefix)", broken)
	t.Logf("MD2(empty) appended to input: New().Sum(input) full = %x", New().Sum(payload))

	// Also confirm: New().Sum(nil) with no Write = MD2("").
	emptyHash := New().Sum(nil)
	if hex.EncodeToString(emptyHash) != "8350e5a3e24c153df2275c9f80692773" {
		t.Errorf("MD2('') = %x, want 8350e5a3e24c153df2275c9f80692773", emptyHash)
	}
}

// TestMD2_SameInputDifferentWrite confirms that after Write+Sum, the hasher
// state is reset correctly (Sum copies and finalizes, leaving original reusable).
func TestMD2_SameInputDifferentWrite(t *testing.T) {
	a := []byte("hello")
	b := []byte("world")

	h := New()
	h.Write(a)
	sumA := h.Sum(nil)

	h.Write(b)
	sumB := h.Sum(nil)

	if bytes.Equal(sumA, sumB) {
		t.Errorf("MD2('hello') and MD2('helloworld') should differ")
	}
	t.Logf("MD2('hello')          = %x", sumA)
	t.Logf("MD2('hello')+MD2('world') = %x", sumB)
}
