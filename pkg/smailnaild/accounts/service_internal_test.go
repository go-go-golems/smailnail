package accounts

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestCreateKeepsExistingDefaultWhenInsertFails(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := bootstrapRetryTestDB(db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := NewRepository(db)
	service := NewService(repo, testRetrySecretConfig())
	service.newID = func() string { return "duplicate-id" }

	first, err := service.Create(context.Background(), "alice", CreateInput{
		Label:          "Primary",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
		IsDefault:      true,
	})
	if err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}
	if !first.IsDefault {
		t.Fatalf("expected first account to be default")
	}

	_, err = service.Create(context.Background(), "alice", CreateInput{
		Label:          "Secondary",
		Server:         "imap-backup.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
		IsDefault:      true,
	})
	if err == nil {
		t.Fatal("expected duplicate create to fail")
	}

	accounts, err := repo.ListByUser(context.Background(), "alice")
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("len(ListByUser()) = %d, want 1", len(accounts))
	}
	if !accounts[0].IsDefault {
		t.Fatalf("expected existing default account to remain default after failed create")
	}
}

func TestDeleteDoesNotTouchAnotherUsersTests(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := bootstrapRetryTestDB(db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := NewRepository(db)
	service := NewService(repo, testRetrySecretConfig())

	alice, err := service.Create(context.Background(), "alice", CreateInput{
		Label:          "Alice",
		Server:         "imap.alice.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create(alice) error = %v", err)
	}

	bob, err := service.Create(context.Background(), "bob", CreateInput{
		Label:          "Bob",
		Server:         "imap.bob.example.com",
		Port:           993,
		Username:       "bob@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create(bob) error = %v", err)
	}

	if err := repo.CreateTest(context.Background(), &AccountTest{
		ID:            "bob-test",
		IMAPAccountID: bob.ID,
		TestMode:      TestModeReadOnly,
		Success:       true,
		TCPOK:         true,
		LoginOK:       true,
		CreatedAt:     service.now(),
	}); err != nil {
		t.Fatalf("CreateTest() error = %v", err)
	}

	err = service.Delete(context.Background(), "alice", bob.ID)
	if err != ErrNotFound {
		t.Fatalf("Delete() error = %v, want %v", err, ErrNotFound)
	}

	latest, err := repo.LatestTestByAccount(context.Background(), bob.ID)
	if err != nil {
		t.Fatalf("LatestTestByAccount() error = %v", err)
	}
	if latest == nil {
		t.Fatal("expected bob's latest test to remain after alice deletion attempt")
	}

	if _, err := repo.GetByID(context.Background(), "alice", alice.ID); err != nil {
		t.Fatalf("GetByID(alice) error = %v", err)
	}
	if _, err := repo.GetByID(context.Background(), "bob", bob.ID); err != nil {
		t.Fatalf("GetByID(bob) error = %v", err)
	}
}
