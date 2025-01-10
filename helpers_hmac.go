package ipmi

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
)

// generate_hmac generates message authentication code.
// Currently supported algorithms are: "md5", "sha1", "sha256"
//
// hmac, hash-based message authentication code
// mac, message authentication code
// md, message digest
func generate_hmac(alg string, data []byte, key []byte) ([]byte, error) {
	switch alg {
	case "md5":
		h := hmac.New(md5.New, key)
		_, err := h.Write(data)
		if err != nil {
			return nil, fmt.Errorf("hmac md5 failed, err: %w", err)
		}
		return h.Sum(nil), nil

	case "sha1":
		h := hmac.New(sha1.New, key)
		_, err := h.Write(data)
		if err != nil {
			return nil, fmt.Errorf("hmac sha1 failed, err: %w", err)
		}
		return h.Sum(nil), nil

	case "sha256":
		h := hmac.New(sha256.New, key)
		_, err := h.Write(data)
		if err != nil {
			return nil, fmt.Errorf("hmac sha256 failed, err: %w", err)
		}
		return h.Sum(nil), nil

	default:
		return nil, fmt.Errorf("not support for hmac algorithm %s", alg)
	}
}

func generate_auth_hmac(authAlg interface{}, data []byte, key []byte) ([]byte, error) {
	algorithm := ""
	switch authAlg.(type) {
	case AuthAlg:
		switch authAlg {
		case AuthAlgRAKP_HMAC_SHA1:
			algorithm = "sha1"
		case AuthAlgRAKP_HMAC_MD5:
			algorithm = "md5"
		case AuthAlgRAKP_HMAC_SHA256:
			algorithm = "sha256"
		default:
			return nil, fmt.Errorf("not support for authentication algorithm %x", authAlg)
		}
	case IntegrityAlg:
		switch authAlg {
		case IntegrityAlg_HMAC_SHA1_96:
			algorithm = "sha1"
		case IntegrityAlg_HMAC_MD5_128:
			algorithm = "md5"
		case IntegrityAlg_HMAC_SHA256_128:
			algorithm = "sha256"
		default:
			return nil, fmt.Errorf("not support for integrity algorithm %x", authAlg)
		}
	}

	if len(algorithm) == 0 {
		return []byte{}, nil
	} else {
		return generate_hmac(algorithm, data, key)
	}
}

// The plainText must be already padded.
// The cipherKey length is either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.
func encryptAES(plainText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
	if len(plainText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("input plainText must be multiple of aes block size (16)")
	}

	l := len(cipherKey)
	if l != 16 && l != 24 && l != 32 {
		return nil, fmt.Errorf("cipherKey length must be either 16, 24, 32")
	}

	cipherBlock, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}

	cipherText := make([]byte, len(plainText))

	mode := cipher.NewCBCEncrypter(cipherBlock, iv)
	mode.CryptBlocks(cipherText, plainText)

	return cipherText, nil
}

func decryptAES(cipherText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
	l := len(cipherKey)
	if l != 16 && l != 24 && l != 32 {
		return nil, fmt.Errorf("cipherKey length must be either 16, 24, 32")
	}

	cipherBlock, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}

	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(cipherBlock, iv)
	mode.CryptBlocks(plainText, cipherText)

	return plainText, nil
}

func encryptRC4(plainText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
	rc4Cipher, err := rc4.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}

	cipherText := make([]byte, len(plainText))
	rc4Cipher.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

func decryptRC4(cipherText []byte, cipherKey []byte, iv []byte) ([]byte, error) {
	rc4Cipher, err := rc4.NewCipher(cipherKey)
	if err != nil {
		return nil, fmt.Errorf("NewCipher failed, err: %w", err)
	}

	plainText := make([]byte, len(cipherText))
	rc4Cipher.XORKeyStream(plainText, cipherText)
	return plainText, nil
}
