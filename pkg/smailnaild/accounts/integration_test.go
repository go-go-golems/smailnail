package accounts_test

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
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestServiceAgainstLocalDovecot(t *testing.T) {
	fixture := requireDovecotFixture(t)
	acquireDovecotLock(t)

	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := smailnaild.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	service := accounts.NewService(accounts.NewRepository(db), testSecretConfig())

	account, err := service.Create(t.Context(), "alice", accounts.CreateInput{
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
		t.Fatalf("Create() error = %v", err)
	}

	subject := fmt.Sprintf("smailnaild account integration %d", time.Now().UnixNano())
	appendFixtureMessage(t, fixture, subject)

	testResult, err := service.RunTest(t.Context(), "alice", account.ID, accounts.TestInput{Mode: accounts.TestModeReadOnly})
	if err != nil {
		t.Fatalf("RunTest() error = %v", err)
	}
	if !testResult.Success {
		t.Fatalf("expected successful test result, got %+v", testResult)
	}

	mailboxes, err := service.ListMailboxes(t.Context(), "alice", account.ID)
	if err != nil {
		t.Fatalf("ListMailboxes() error = %v", err)
	}
	if len(mailboxes) == 0 {
		t.Fatal("expected at least one mailbox")
	}

	messages, mailbox, err := service.ListMessages(t.Context(), "alice", account.ID, accounts.ListMessagesInput{
		Query: subject,
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if mailbox != fixture.mailbox {
		t.Fatalf("mailbox = %q", mailbox)
	}
	if len(messages) == 0 {
		t.Fatalf("expected seeded message %q to be returned", subject)
	}

	message, resolvedMailbox, err := service.GetMessage(t.Context(), "alice", account.ID, fixture.mailbox, messages[0].UID)
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}
	if resolvedMailbox != fixture.mailbox {
		t.Fatalf("resolved mailbox = %q", resolvedMailbox)
	}
	if !strings.Contains(message.Subject, subject) {
		t.Fatalf("message subject = %q", message.Subject)
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

	raw := []byte(fmt.Sprintf("From: Seeder <seed@example.com>\r\nTo: User A <a@testcot>\r\nSubject: %s\r\nDate: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nHosted test body for %s\r\n",
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
