package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptAES(t *testing.T) {
	cases := []struct {
		name            string
		ipmiRequestBody []byte
		iv              []byte
		key             []byte
		expected        []byte
	}{
		{
			name: "test1",
			ipmiRequestBody: []byte{
				0x20, 0x18, 0xc8, 0x81, 0x04, 0x3b, 0x04, 0x3c, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x07,
			},
			iv: []byte{
				0x00, 0xdf, 0x5e, 0x2a, 0x2b, 0x37, 0x2e, 0x80, 0x7a, 0xe5, 0x5b, 0x0c, 0x37, 0x3c, 0x37, 0x69,
			},
			key: []byte{
				0x12, 0x0e, 0x6b, 0x20, 0xe1, 0xe5, 0x2d, 0x13, 0xa0, 0x4a, 0x2b, 0xb8, 0x3d, 0x0d, 0x38, 0xa1,
			},
			expected: []byte{
				0x47, 0x9c, 0x2f, 0x65, 0xfb, 0x59, 0x75, 0x19, 0x71, 0xa2, 0x96, 0xa3, 0x77, 0x15, 0x55, 0x69,
			},
		},
	}

	for _, v := range cases {
		got, err := EncryptAES(v.ipmiRequestBody, v.key, v.iv)
		if err != nil {
			t.Errorf("%s: EncryptAES failed, err: %s", v.name, err)
		}
		if !bytes.Equal(got, v.expected) {
			t.Errorf("%s: got does not match expected, got: %v, want: %v", v.name, got, v.expected)
		}
	}
}

func TestAESRoundTrip(t *testing.T) {
	iv := []byte("1234567890123456")
	cipherKey := []byte("12345678901234567890123456789012")
	plainText := []byte("abcdefghijklmnopqrstuvwxyzABCDEF")

	cipherText, err := EncryptAES(plainText, cipherKey, iv)
	if err != nil {
		t.Fatal(err)
	}

	got, err := DecryptAES(cipherText, cipherKey, iv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plainText) {
		t.Fatalf("AES round-trip mismatch\nwant %q\n got %q", plainText, got)
	}
}

func TestRC4RoundTrip(t *testing.T) {
	iv := []byte("1234567890123456")
	cipherKey := []byte("12345678901234567890123456789012")
	plainText := []byte("abcdefghijklmnopqrstuvwxyzABCDEF")

	cipherText, err := EncryptRC4(plainText, cipherKey, iv)
	if err != nil {
		t.Fatal(err)
	}

	got, err := DecryptRC4(cipherText, cipherKey, iv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plainText) {
		t.Fatalf("RC4 round-trip mismatch\nwant %q\n got %q", plainText, got)
	}
}
