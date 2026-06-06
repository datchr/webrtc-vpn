// Package shared contains encryption utilities
package shared

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Cipher wraps AES-256-GCM encryption
type Cipher struct {
	key [32]byte
}

// NewCipher creates a new cipher from base64-encoded key
func NewCipher(keyB64 string) (*Cipher, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	c := &Cipher{}
	copy(c.key[:], key)
	return c, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (c *Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	return plaintext, err
}

// GenerateKey generates a random 32-byte key and returns it as base64
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
