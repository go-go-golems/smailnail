package accounts

import (
	"context"
	"errors"
	"testing"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestRunTestRetriesTransientProbeFailure(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := bootstrapRetryTestDB(db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	service := NewService(NewRepository(db), testRetrySecretConfig())
	account, err := service.Create(context.Background(), "alice", CreateInput{
		Label:          "Retry account",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	attempts := 0
	service.runReadOnlyProbe = func(_ *ConnectionDetails) (*readOnlyProbeResult, string, error) {
		attempts++
		if attempts == 1 {
			return &readOnlyProbeResult{
				TCPOK:   true,
				LoginOK: true,
				Details: map[string]any{"mailbox": "INBOX"},
			}, "select", errors.New("use of closed network connection")
		}
		return &readOnlyProbeResult{
			TCPOK:           true,
			LoginOK:         true,
			MailboxSelectOK: true,
			ListOK:          true,
			SampleFetchOK:   true,
			Details: map[string]any{
				"mailbox":         "INBOX",
				"selectedMailbox": "INBOX",
			},
		}, "", nil
	}

	result, err := service.RunTest(context.Background(), "alice", account.ID, TestInput{Mode: TestModeReadOnly})
	if err != nil {
		t.Fatalf("RunTest() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if !result.Success {
		t.Fatalf("expected success after retry, got %+v", result)
	}
	if result.ErrorCode != "" || result.ErrorMessage != "" {
		t.Fatalf("expected cleared error state, got code=%q message=%q", result.ErrorCode, result.ErrorMessage)
	}
}

func TestRunTestDoesNotRetryNonTransientProbeFailure(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := bootstrapRetryTestDB(db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	service := NewService(NewRepository(db), testRetrySecretConfig())
	account, err := service.Create(context.Background(), "alice", CreateInput{
		Label:          "No retry account",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	attempts := 0
	service.runReadOnlyProbe = func(_ *ConnectionDetails) (*readOnlyProbeResult, string, error) {
		attempts++
		return &readOnlyProbeResult{
			TCPOK:   true,
			LoginOK: false,
			Details: map[string]any{"mailbox": "INBOX"},
		}, "login", errors.New("authentication failed")
	}

	result, err := service.RunTest(context.Background(), "alice", account.ID, TestInput{Mode: TestModeReadOnly})
	if err != nil {
		t.Fatalf("RunTest() error = %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if result.Success {
		t.Fatalf("expected failed result, got %+v", result)
	}
	if result.ErrorCode != "account-test-login-failed" {
		t.Fatalf("errorCode = %q", result.ErrorCode)
	}
}

func testRetrySecretConfig() *secrets.Config {
	return &secrets.Config{
		KeyID: secrets.DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	}
}

func bootstrapRetryTestDB(db *sqlx.DB) error {
	schema := []string{
		`CREATE TABLE imap_accounts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			label TEXT NOT NULL,
			provider_hint TEXT NOT NULL DEFAULT '',
			server TEXT NOT NULL,
			port INTEGER NOT NULL,
			username TEXT NOT NULL,
			mailbox_default TEXT NOT NULL DEFAULT 'INBOX',
			insecure BOOLEAN NOT NULL DEFAULT FALSE,
			auth_kind TEXT NOT NULL DEFAULT 'password',
			secret_ciphertext TEXT NOT NULL,
			secret_nonce TEXT NOT NULL,
			secret_key_id TEXT NOT NULL,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			mcp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE imap_account_tests (
			id TEXT PRIMARY KEY,
			imap_account_id TEXT NOT NULL,
			test_mode TEXT NOT NULL,
			success BOOLEAN NOT NULL DEFAULT FALSE,
			tcp_ok BOOLEAN NOT NULL DEFAULT FALSE,
			login_ok BOOLEAN NOT NULL DEFAULT FALSE,
			mailbox_select_ok BOOLEAN NOT NULL DEFAULT FALSE,
			list_ok BOOLEAN NOT NULL DEFAULT FALSE,
			sample_fetch_ok BOOLEAN NOT NULL DEFAULT FALSE,
			write_probe_ok BOOLEAN,
			warning_code TEXT NOT NULL DEFAULT '',
			error_code TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT '',
			details_json TEXT NOT NULL DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
