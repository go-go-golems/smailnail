package rules_test

import (
	"testing"

	"github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestServiceCreateUpdateDeleteAndValidation(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := smailnaild.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), &secrets.Config{
		KeyID: secrets.DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	})
	account, err := accountService.Create(t.Context(), "alice", accounts.CreateInput{
		Label:          "Work",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "alice@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       accounts.AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create account error = %v", err)
	}

	service := rules.NewService(rules.NewRepository(db), accountService)

	record, err := service.Create(t.Context(), "alice", rules.CreateInput{
		IMAPAccountID: account.ID,
		Name:          "Invoice triage",
		Description:   "Find invoices",
		RuleYAML: `name: placeholder
description: placeholder
search:
  subject_contains: invoice
output:
  format: json
  limit: 20
  fields:
    - uid
    - subject
`,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if record.Name != "Invoice triage" {
		t.Fatalf("record.Name = %q", record.Name)
	}

	listed, err := service.List(t.Context(), "alice")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("len(List()) = %d", len(listed))
	}

	newName := "Invoice triage updated"
	updated, err := service.Update(t.Context(), "alice", record.ID, rules.UpdateInput{Name: &newName})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("updated.Name = %q", updated.Name)
	}

	if _, err := service.Create(t.Context(), "alice", rules.CreateInput{
		IMAPAccountID: account.ID,
		RuleYAML:      `name: broken`,
	}); err == nil {
		t.Fatal("expected validation error for invalid rule YAML")
	}

	if err := service.Delete(t.Context(), "alice", record.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(t.Context(), "alice", record.ID); err != rules.ErrNotFound {
		t.Fatalf("Get() after delete error = %v, want %v", err, rules.ErrNotFound)
	}
}
