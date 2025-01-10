package ipmi

import (
	"fmt"
	"testing"
)

func Test_encryptAES(t *testing.T) {
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
		{
			name: "test2",
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
		got, err := encryptAES(v.ipmiRequestBody, v.key, v.iv)
		if err != nil {
			t.Errorf("encryptAES failed, err: %w", err)
		}
		if !isByteSliceEqual(got, v.expected) {
			t.Errorf("got does not match expected, got: %v, want: %v", got, v.expected)
		}
	}
}

func Test_AES_Decrypt(t *testing.T) {
	ivStr := "1234567890123456"
	cipherKey := "12345678901234567890123456789012"

	plainText := "abcdefghijklmnopqrstuvwxyzABCDEF"
	cipherText, err := encryptAES([]byte(plainText), []byte(cipherKey), []byte(ivStr))
	if err != nil {
		t.Error(err)
	}

	p, err := decryptAES([]byte(cipherText), []byte(cipherKey), []byte(ivStr))
	if err != nil {
		t.Error(err)
	}

	if len(plainText) != len(p) {
		t.Error("Not equal")
	}

	for k, v := range []byte(plainText) {
		if v != p[k] {
			t.Errorf("not equal at %d", k)
		}
	}
}

func Test_RC4_Encrypt(t *testing.T) {
	ivStr := "1234567890123456"
	cipherKey := "12345678901234567890123456789012"

	plainText := "abcdefghijklmnopqrstuvwxyzABCDEF"
	cipherText, err := encryptRC4([]byte(plainText), []byte(cipherKey), []byte(ivStr))
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(cipherText))
	fmt.Println(cipherText)
}

func Test_RC4_Decrypt(t *testing.T) {
	ivStr := "1234567890123456"
	cipherKey := "12345678901234567890123456789012"

	plainText := "abcdefghijklmnopqrstuvwxyzABCDEF"
	cipherText, err := encryptRC4([]byte(plainText), []byte(cipherKey), []byte(ivStr))
	if err != nil {
		t.Error(err)
	}

	p, err := decryptRC4([]byte(cipherText), []byte(cipherKey), []byte(ivStr))
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(p))

	if len(plainText) != len(p) {
		t.Error("Not equal")
	}

	for k, v := range []byte(plainText) {
		if v != p[k] {
			t.Errorf("not equal at %d", k)
		}
	}
}
