package rules_test

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestDryRunAgainstLocalDovecot(t *testing.T) {
	fixture := requireDovecotFixture(t)
	acquireDovecotLock(t)

	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := smailnaild.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), &secrets.Config{
		KeyID: secrets.DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	})
	ruleService := rules.NewService(rules.NewRepository(db), accountService)

	account, err := accountService.Create(t.Context(), "alice", accounts.CreateInput{
		Label:          "Local Dovecot",
		ProviderHint:   "local",
		Server:         fixture.server,
		Port:           fixture.port,
		Username:       fixture.username,
		Password:       fixture.password,
		MailboxDefault: fixture.mailbox,
		Insecure:       fixture.insecure,
		AuthKind:       accounts.AuthKindPassword,
	})
	if err != nil {
		t.Fatalf("Create account error = %v", err)
	}

	subject := fmt.Sprintf("smailnaild rule integration %d", time.Now().UnixNano())
	appendFixtureMessage(t, fixture, subject)

	record, err := ruleService.Create(t.Context(), "alice", rules.CreateInput{
		IMAPAccountID: account.ID,
		RuleYAML: fmt.Sprintf(`name: Local dry run
description: Local dry run
search:
  subject_contains: %q
output:
  format: json
  limit: 10
  fields:
    - uid
    - subject
    - from
    - date
actions:
  move_to: Archive
`, subject),
	})
	if err != nil {
		t.Fatalf("Create rule error = %v", err)
	}

	result, err := ruleService.DryRun(t.Context(), "alice", record.ID, rules.DryRunInput{})
	if err != nil {
		t.Fatalf("DryRun() error = %v", err)
	}
	if result.MatchedCount == 0 {
		t.Fatalf("expected dry run to match seeded message")
	}
	if result.ActionPlan["moveTo"] != "Archive" {
		t.Fatalf("unexpected action plan: %+v", result.ActionPlan)
	}
	if len(result.SampleRows) == 0 || !strings.Contains(result.SampleRows[0].Subject, subject) {
		t.Fatalf("unexpected sample rows: %+v", result.SampleRows)
	}
}

type dovecotFixture struct {
	server   string
	port     int
	username string
	password string
	mailbox  string
	insecure bool
}

func requireDovecotFixture(t *testing.T) dovecotFixture {
	t.Helper()
	if os.Getenv("SMAILNAILD_DOVECOT_TEST") != "1" {
		t.Skip("set SMAILNAILD_DOVECOT_TEST=1 to run local Dovecot integration tests")
	}

	return dovecotFixture{
		server:   getenvDefault("SMAILNAILD_DOVECOT_SERVER", "localhost"),
		port:     getenvIntDefault("SMAILNAILD_DOVECOT_PORT", 993),
		username: getenvDefault("SMAILNAILD_DOVECOT_USERNAME", "a"),
		password: getenvDefault("SMAILNAILD_DOVECOT_PASSWORD", "pass"),
		mailbox:  getenvDefault("SMAILNAILD_DOVECOT_MAILBOX", "INBOX"),
		insecure: true,
	}
}

func acquireDovecotLock(t *testing.T) {
	t.Helper()

	lockPath := filepath.Join(os.TempDir(), "smailnaild-dovecot-fixture.lock")
	deadline := time.Now().Add(30 * time.Second)

	for {
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			t.Cleanup(func() {
				_ = lockFile.Close()
				_ = os.Remove(lockPath)
			})
			return
		}
		if !os.IsExist(err) {
			t.Fatalf("failed to acquire Dovecot lock: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out acquiring Dovecot lock %s", lockPath)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func appendFixtureMessage(t *testing.T, fixture dovecotFixture, subject string) {
	t.Helper()

	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			// #nosec G402 -- explicit integration test against a local self-signed fixture.
			InsecureSkipVerify: fixture.insecure,
		},
	}

	client, err := imapclient.DialTLS(fmt.Sprintf("%s:%d", fixture.server, fixture.port), options)
	if err != nil {
		t.Fatalf("DialTLS() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Login(fixture.username, fixture.password).Wait(); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	raw := []byte(fmt.Sprintf("From: Seeder <seed@example.com>\r\nTo: User A <a@testcot>\r\nSubject: %s\r\nDate: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nRule dry run body for %s\r\n",
		subject,
		time.Now().Format(time.RFC1123Z),
		subject,
	))

	appendCmd := client.Append(fixture.mailbox, int64(len(raw)), nil)
	if _, err := appendCmd.Write(raw); err != nil {
		t.Fatalf("Append.Write() error = %v", err)
	}
	if err := appendCmd.Close(); err != nil {
		t.Fatalf("Append.Close() error = %v", err)
	}
	if _, err := appendCmd.Wait(); err != nil {
		t.Fatalf("Append.Wait() error = %v", err)
	}
}

func getenvDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvIntDefault(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
