package secrets

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	key := make([]byte, requiredEncryptionBytes)
	for i := range key {
		key[i] = byte(i + 1)
	}
	value := base64.StdEncoding.EncodeToString(key)

	t.Setenv(EncryptionKeyEnv, value)

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}
	if config.KeyID != DefaultEncryptionKeyID {
		t.Fatalf("unexpected key id %q", config.KeyID)
	}
	if len(config.Key) != requiredEncryptionBytes {
		t.Fatalf("expected %d-byte key, got %d", requiredEncryptionBytes, len(config.Key))
	}
}

func TestLoadConfigFromEnvRejectsMissingValue(t *testing.T) {
	t.Setenv(EncryptionKeyEnv, "")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatal("expected missing env var to fail")
	}
}

func TestLoadConfigFromEnvRejectsShortKey(t *testing.T) {
	t.Setenv(EncryptionKeyEnv, base64.StdEncoding.EncodeToString([]byte("short")))

	_, err := LoadConfigFromEnv()
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

func TestLoadConfigFromEnvUsesCurrentProcessEnv(t *testing.T) {
	key := make([]byte, requiredEncryptionBytes)
	for i := range key {
		key[i] = 0x42
	}
	value := base64.StdEncoding.EncodeToString(key)

	original, ok := os.LookupEnv(EncryptionKeyEnv)
	defer func() {
		if ok {
			_ = os.Setenv(EncryptionKeyEnv, original)
		} else {
			_ = os.Unsetenv(EncryptionKeyEnv)
		}
	}()

	if err := os.Setenv(EncryptionKeyEnv, value); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}

	config, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("LoadConfigFromEnv() error = %v", err)
	}
	if got := base64.StdEncoding.EncodeToString(config.Key); got != value {
		t.Fatalf("expected key %q, got %q", value, got)
	}
}
