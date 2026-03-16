package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Envelope struct {
	Ciphertext string
	Nonce      string
	KeyID      string
}

func EncryptString(config *Config, plaintext string) (*Envelope, error) {
	if config == nil {
		return nil, fmt.Errorf("secret config is nil")
	}

	aead, err := newAEAD(config.Key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	return &Envelope{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		KeyID:      config.KeyID,
	}, nil
}

func DecryptString(config *Config, envelope *Envelope) (string, error) {
	if config == nil {
		return "", fmt.Errorf("secret config is nil")
	}
	if envelope == nil {
		return "", fmt.Errorf("secret envelope is nil")
	}

	aead, err := newAEAD(config.Key)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(envelope.Nonce)
	if err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return string(plaintext), nil
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != requiredEncryptionBytes {
		return nil, fmt.Errorf("encryption key must be %d bytes, got %d", requiredEncryptionBytes, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return aead, nil
}
