package backend

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const aesGCMTokenPrefix = "aesgcm:v1:"

func sealSecretAESGCM(key []byte, plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if len(key) != 32 {
		return "", errors.New("invalid AES-256-GCM key length")
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
	sealed := gcm.Seal(nonce, nonce, []byte(plain), []byte("AppleVault.keychain.v1"))
	return aesGCMTokenPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func openSecretAESGCM(key []byte, encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	if !strings.HasPrefix(encoded, aesGCMTokenPrefix) {
		return "", errors.New("unsupported protected secret format")
	}
	if len(key) != 32 {
		return "", errors.New("invalid AES-256-GCM key length")
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, aesGCMTokenPrefix))
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
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("protected secret payload too short")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, []byte("AppleVault.keychain.v1"))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
