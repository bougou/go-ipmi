package server

import (
	"bytes"
	"testing"
)

// TestAesPadRoundTrip verifies the IPMI 2.0 AES-CBC-128 payload padding format
// (spec §13.29): plain || pad bytes (1,2,3,...) || pad-length byte, total
// length a multiple of 16.  decryptPayload must strip exactly the pad bytes
// and the pad-length byte to recover the original payload.
func TestAesPadRoundTrip(t *testing.T) {
	k2 := bytes.Repeat([]byte{0x42}, 32)

	cases := []int{
		0, 1, 7, 13, 15, 16, 17, 31, 32, 33,
	}
	for _, n := range cases {
		plain := make([]byte, n)
		for i := range plain {
			plain[i] = byte(i)
		}

		ct, err := encryptPayload(plain, k2)
		if err != nil {
			t.Fatalf("n=%d: encryptPayload failed: %v", n, err)
		}

		got, err := decryptPayload(ct, k2)
		if err != nil {
			t.Fatalf("n=%d: decryptPayload failed: %v", n, err)
		}
		if !bytes.Equal(got, plain) {
			t.Fatalf("n=%d: round-trip mismatch\nwant % x\n got % x", n, plain, got)
		}
	}
}

// TestDecryptPayload_StripsPadding asserts that decryptPayload removes the
// trailing pad bytes and pad-length byte rather than returning the full
// decrypted block.  This is the regression guard for the boot-options handlers
// that validate exact request lengths.
func TestDecryptPayload_StripsPadding(t *testing.T) {
	k2 := bytes.Repeat([]byte{0x07}, 32)

	// A 13-byte IPMI message pads to 16 bytes (padLen=2 + pad-length byte).
	plain := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}
	ct, err := encryptPayload(plain, k2)
	if err != nil {
		t.Fatalf("encryptPayload failed: %v", err)
	}

	got, err := decryptPayload(ct, k2)
	if err != nil {
		t.Fatalf("decryptPayload failed: %v", err)
	}
	if len(got) != len(plain) {
		t.Fatalf("decryptPayload did not strip padding: got %d bytes, want %d", len(got), len(plain))
	}
}
