package secrets

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

const (
	EncryptionSectionSlug   = "encryption"
	DefaultEncryptionKeyID  = "app:smailnaild-encryption-key"
	requiredEncryptionBytes = 32
)

type Config struct {
	KeyID string
	Key   []byte
}

type Settings struct {
	KeyBase64 string `glazed:"encryption-key-base64"`
	KeyID     string `glazed:"encryption-key-id"`
}

func NewSection() (schema.Section, error) {
	return schema.NewSection(
		EncryptionSectionSlug,
		"Encryption Settings",
		schema.WithFields(
			fields.New(
				"encryption-key-base64",
				fields.TypeString,
				fields.WithHelp("Base64-encoded 32-byte AES-GCM key used to encrypt stored IMAP passwords"),
			),
			fields.New(
				"encryption-key-id",
				fields.TypeString,
				fields.WithHelp("Logical identifier stored alongside encrypted IMAP passwords"),
				fields.WithDefault(DefaultEncryptionKeyID),
			),
		),
	)
}

func LoadConfigFromParsedValues(parsedValues *values.Values) (*Config, error) {
	if parsedValues == nil {
		return nil, fmt.Errorf("parsed values are nil")
	}

	settings := &Settings{}
	if err := parsedValues.DecodeSectionInto(EncryptionSectionSlug, settings); err != nil {
		return nil, err
	}

	return LoadConfigFromSettings(settings)
}

func LoadConfigFromSettings(settings *Settings) (*Config, error) {
	if settings == nil {
		return nil, fmt.Errorf("encryption settings are nil")
	}

	raw := strings.TrimSpace(settings.KeyBase64)
	if raw == "" {
		return nil, fmt.Errorf("encryption-key-base64 is required")
	}

	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode encryption-key-base64: %w", err)
	}
	if len(key) != requiredEncryptionBytes {
		return nil, fmt.Errorf("encryption-key-base64 must decode to %d bytes, got %d", requiredEncryptionBytes, len(key))
	}

	keyID := strings.TrimSpace(settings.KeyID)
	if keyID == "" {
		keyID = DefaultEncryptionKeyID
	}

	return &Config{
		KeyID: keyID,
		Key:   key,
	}, nil
}
