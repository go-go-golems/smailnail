package secrets

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

const (
	EncryptionKeyEnv        = "SMAILNAILD_ENCRYPTION_KEY"
	DefaultEncryptionKeyID  = "env:smailnaild-encryption-key"
	requiredEncryptionBytes = 32
)

type Config struct {
	KeyID string
	Key   []byte
}

func LoadConfigFromEnv() (*Config, error) {
	raw := strings.TrimSpace(os.Getenv(EncryptionKeyEnv))
	if raw == "" {
		return nil, fmt.Errorf("%s is required", EncryptionKeyEnv)
	}

	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", EncryptionKeyEnv, err)
	}
	if len(key) != requiredEncryptionBytes {
		return nil, fmt.Errorf("%s must decode to %d bytes, got %d", EncryptionKeyEnv, requiredEncryptionBytes, len(key))
	}

	return &Config{
		KeyID: DefaultEncryptionKeyID,
		Key:   key,
	}, nil
}
