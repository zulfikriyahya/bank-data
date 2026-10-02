package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

func getKey() ([]byte, error) {
	keyB64 := os.Getenv("ENCRYPTION_KEY")
	if keyB64 == "" {
		return nil, errors.New("ENCRYPTION_KEY belum diatur di .env")
	}
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, errors.New("ENCRYPTION_KEY tidak valid (harus base64)")
	}
	if len(key) != 32 {
		return nil, errors.New("ENCRYPTION_KEY harus 32 byte (generate pakai: openssl rand -base64 32)")
	}
	return key, nil
}

// Encrypt - enkripsi plaintext, hasil berupa base64 string siap disimpan ke DB
func Encrypt(plaintext string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt - kebalikan dari Encrypt
func Decrypt(encoded string) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("format data terenkripsi tidak valid")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("data terenkripsi rusak")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("gagal dekripsi, kemungkinan ENCRYPTION_KEY berubah")
	}

	return string(plaintext), nil
}
