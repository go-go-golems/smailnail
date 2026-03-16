package accounts_test

import (
	"encoding/base64"
	"testing"

	"github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestServiceCreateListUpdateAndOwnership(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := smailnaild.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	service := accounts.NewService(accounts.NewRepository(db), testSecretConfig())

	account, err := service.Create(t.Context(), "alice", accounts.CreateInput{
		Label:          "Work",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       accounts.AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !account.IsDefault {
		t.Fatalf("expected first account to be default")
	}

	listed, err := service.List(t.Context(), "alice")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(List()) = %d", len(listed))
	}
	if listed[0].LatestTest != nil {
		t.Fatalf("expected no latest test, got %+v", listed[0].LatestTest)
	}

	if _, err := service.Get(t.Context(), "bob", account.ID); err != accounts.ErrNotFound {
		t.Fatalf("Get() error = %v, want %v", err, accounts.ErrNotFound)
	}

	newLabel := "Work Updated"
	newPassword := "secret-2"
	updated, err := service.Update(t.Context(), "alice", account.ID, accounts.UpdateInput{
		Label:    &newLabel,
		Password: &newPassword,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Label != newLabel {
		t.Fatalf("updated label = %q", updated.Label)
	}

	connection, err := service.ResolveConnection(t.Context(), "alice", account.ID)
	if err != nil {
		t.Fatalf("ResolveConnection() error = %v", err)
	}
	if connection.Password != newPassword {
		t.Fatalf("resolved password = %q", connection.Password)
	}

	if err := service.Delete(t.Context(), "alice", account.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(t.Context(), "alice", account.ID); err != accounts.ErrNotFound {
		t.Fatalf("Get() after delete error = %v, want %v", err, accounts.ErrNotFound)
	}
}

func testSecretConfig() *secrets.Config {
	key := []byte("0123456789abcdef0123456789abcdef")
	return &secrets.Config{
		KeyID: secrets.DefaultEncryptionKeyID,
		Key:   key,
	}
}

func TestTestSecretConfigIsBase64Compatible(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString(testSecretConfig().Key)
	if raw == "" {
		t.Fatal("expected non-empty base64 key")
	}
}
