package client

import (
	"testing"

	"github.com/bougou/go-ipmi/pkg/types"
)

func TestDecryptPayloadXRC4LaterPacket(t *testing.T) {
	c := &Client{session: &session{v20: v20{
		cryptAlg: types.CryptAlg_xRC4_128,
		k2:       make([]byte, 16),
	}}}

	first, err := c.encryptPayload([]byte("hello"), nil)
	if err != nil {
		t.Fatalf("encrypt first: %v", err)
	}
	plain, err := c.decryptPayload(first)
	if err != nil {
		t.Fatalf("decrypt first: %v", err)
	}
	if string(plain) != "hello" {
		t.Fatalf("first plaintext = %q", plain)
	}

	// Later packets carry a 4-byte offset, not another IV. "world" encrypts
	// to 9 bytes, which used to panic on data[20:].
	second, err := c.encryptPayload([]byte("world"), nil)
	if err != nil {
		t.Fatalf("encrypt second: %v", err)
	}
	if len(second) >= 20 {
		t.Fatalf("second packet length = %d, want a short later packet", len(second))
	}
	plain, err = c.decryptPayload(second)
	if err != nil {
		t.Fatalf("decrypt second: %v", err)
	}
	if string(plain) != "world" {
		t.Fatalf("second plaintext = %q", plain)
	}

	if _, err := c.decryptPayload([]byte{0, 0, 0, 0, 1, 2, 3, 4}); err == nil {
		t.Fatal("short first xRC4 packet returned success")
	}
	// Non-zero offset, shorter than 20: must not panic.
	if _, err := c.decryptPayload([]byte{0, 0, 0, 1, 9, 8, 7, 6}); err != nil {
		t.Fatalf("short later packet: %v", err)
	}
}
