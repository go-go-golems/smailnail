package secrets

import (
	"encoding/base64"
	"testing"
)

func TestLoadConfigFromSettings(t *testing.T) {
	key := make([]byte, requiredEncryptionBytes)
	for i := range key {
		key[i] = byte(i + 1)
	}
	value := base64.StdEncoding.EncodeToString(key)

	config, err := LoadConfigFromSettings(&Settings{
		KeyBase64: value,
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}
	if config.KeyID != DefaultEncryptionKeyID {
		t.Fatalf("unexpected key id %q", config.KeyID)
	}
	if len(config.Key) != requiredEncryptionBytes {
		t.Fatalf("expected %d-byte key, got %d", requiredEncryptionBytes, len(config.Key))
	}
}

func TestLoadConfigFromSettingsRejectsMissingValue(t *testing.T) {
	_, err := LoadConfigFromSettings(&Settings{})
	if err == nil {
		t.Fatal("expected missing encryption key to fail")
	}
}

func TestLoadConfigFromSettingsRejectsShortKey(t *testing.T) {
	_, err := LoadConfigFromSettings(&Settings{
		KeyBase64: base64.StdEncoding.EncodeToString([]byte("short")),
	})
	if err == nil {
		t.Fatal("expected short key to fail")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	config := &Config{
		KeyID: DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	}

	envelope, err := EncryptString(config, "app-password-123")
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}
	if envelope.Ciphertext == "" || envelope.Nonce == "" {
		t.Fatal("expected ciphertext and nonce to be populated")
	}

	plaintext, err := DecryptString(config, envelope)
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}
	if plaintext != "app-password-123" {
		t.Fatalf("unexpected plaintext %q", plaintext)
	}
}

func TestDecryptRejectsCorruptCiphertext(t *testing.T) {
	config := &Config{
		KeyID: DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	}

	envelope, err := EncryptString(config, "secret")
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}
	envelope.Ciphertext = base64.StdEncoding.EncodeToString([]byte("corrupt"))

	_, err = DecryptString(config, envelope)
	if err == nil {
		t.Fatal("expected corrupt ciphertext to fail")
	}
}

func TestEncryptRejectsNilConfig(t *testing.T) {
	_, err := EncryptString(nil, "secret")
	if err == nil {
		t.Fatal("expected nil config to fail")
	}
}

func TestLoadConfigFromSettingsUsesExplicitKeyID(t *testing.T) {
	key := make([]byte, requiredEncryptionBytes)
	for i := range key {
		key[i] = 0x42
	}
	value := base64.StdEncoding.EncodeToString(key)

	config, err := LoadConfigFromSettings(&Settings{
		KeyBase64: value,
		KeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}
	if got := base64.StdEncoding.EncodeToString(config.Key); got != value {
		t.Fatalf("expected key %q, got %q", value, got)
	}
	if config.KeyID != "test-key" {
		t.Fatalf("expected key id %q, got %q", "test-key", config.KeyID)
	}
}
