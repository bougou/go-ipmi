package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// EncryptAES encrypts plainText (must be block-aligned) with AES-CBC.
// cipherKey must be 16, 24, or 32 bytes (AES-128/192/256).
//
// Spec: v2.0§13.29 Encryption with AES.
func EncryptAES(plainText, cipherKey, iv []byte) ([]byte, error) {
	if len(plainText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("input plainText must be multiple of aes block size (16)")
	}
	l := len(cipherKey)
	if l != 16 && l != 24 && l != 32 {
		return nil, fmt.Errorf("cipherKey length must be either 16, 24, 32")
	}
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}
	cipherText := make([]byte, len(plainText))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

// DecryptAES decrypts cipherText with AES-CBC.
func DecryptAES(cipherText, cipherKey, iv []byte) ([]byte, error) {
	l := len(cipherKey)
	if l != 16 && l != 24 && l != 32 {
		return nil, fmt.Errorf("cipherKey length must be either 16, 24, 32")
	}
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}
	plainText := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plainText, cipherText)
	return plainText, nil
}

// PadAES pads plain to a multiple of 16 bytes per IPMI 2.0 spec §13.29:
//
//	plain || pad bytes (1,2,3,...) || pad-length byte
func PadAES(plain []byte) []byte {
	padLen := 16 - (len(plain)+1)%16
	if padLen == 16 {
		padLen = 0
	}
	padded := make([]byte, len(plain)+padLen+1)
	copy(padded, plain)
	for i := 0; i < padLen; i++ {
		padded[len(plain)+i] = byte(i + 1)
	}
	padded[len(plain)+padLen] = byte(padLen)
	return padded
}

// UnpadAES strips IPMI §13.29 AES padding from a decrypted block.
func UnpadAES(padded []byte) ([]byte, error) {
	if len(padded) == 0 {
		return nil, fmt.Errorf("decrypted payload is empty")
	}
	padLen := int(padded[len(padded)-1])
	if padLen >= len(padded) {
		return nil, fmt.Errorf("invalid AES pad length %d for %d-byte block", padLen, len(padded))
	}
	return padded[:len(padded)-1-padLen], nil
}

// EncryptAESPayload encrypts an IPMI confidential payload with AES-CBC-128.
// k2 is the session K2 key; the first 16 bytes are used as the cipher key.
// iv must be 16 bytes (caller supplies or uses RandomBytes).
// Wire format: IV(16) || ciphertext.
func EncryptAESPayload(plain, k2, iv []byte) ([]byte, error) {
	if len(k2) < 16 {
		return nil, fmt.Errorf("k2 too short for AES-128")
	}
	if len(iv) != 16 {
		return nil, fmt.Errorf("iv must be 16 bytes")
	}
	encrypted, err := EncryptAES(PadAES(plain), k2[:16], iv)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 16+len(encrypted))
	copy(out[:16], iv)
	copy(out[16:], encrypted)
	return out, nil
}

// DecryptAESPayload decrypts an AES-CBC-128 confidential payload and strips
// §13.29 padding. Wire format: IV(16) || ciphertext.
func DecryptAESPayload(cipherText, k2 []byte) ([]byte, error) {
	if len(cipherText) < 16 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	if len(k2) < 16 {
		return nil, fmt.Errorf("k2 too short for AES-128")
	}
	padded, err := DecryptAES(cipherText[16:], k2[:16], cipherText[:16])
	if err != nil {
		return nil, err
	}
	return UnpadAES(padded)
}
