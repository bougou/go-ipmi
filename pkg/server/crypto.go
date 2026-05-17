package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// encryptAES encrypts plainText (must be block-aligned) with AES-CBC.
// cipherKey must be 16, 24, or 32 bytes.
func encryptAES(plainText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
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

// decryptAES decrypts cipherText with AES-CBC.
func decryptAES(cipherText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
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

// randomBytes returns n cryptographically random bytes.
func randomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return b
}
