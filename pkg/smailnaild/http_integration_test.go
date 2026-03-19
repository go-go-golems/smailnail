package smailnaild

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestHostedHTTPFlowAgainstLocalDovecot(t *testing.T) {
	fixture := requireHostedDovecotFixture(t)
	acquireHostedDovecotLock(t)

	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), &secrets.Config{
		KeyID: secrets.DefaultEncryptionKeyID,
		Key:   []byte("0123456789abcdef0123456789abcdef"),
	})
	ruleService := rules.NewService(rules.NewRepository(db), accountService)

	server := httptest.NewServer(NewHandler(HandlerOptions{
		DB:           db,
		DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Now().UTC(),
		UserResolver: HeaderUserResolver{DefaultUserID: "local-user"},
		AccountAPI:   accountService,
		RuleAPI:      ruleService,
	}))
	defer server.Close()

	subject := fmt.Sprintf("smailnaild http integration %d", time.Now().UnixNano())
	appendHostedFixtureMessage(t, fixture, subject)

	accountID := createHostedAccount(t, server.URL, fixture)
	postJSON(t, server.URL+"/api/accounts/"+accountID+"/test", `{}`)

	var mailboxesResponse struct {
		Data []accounts.MailboxInfo `json:"data"`
	}
	getJSON(t, server.URL+"/api/accounts/"+accountID+"/mailboxes", &mailboxesResponse)
	if len(mailboxesResponse.Data) == 0 {
		t.Fatal("expected mailbox list")
	}

	var messagesResponse struct {
		Data []accounts.MessageView `json:"data"`
	}
	getJSON(t, server.URL+"/api/accounts/"+accountID+"/messages?query="+url.QueryEscape(subject), &messagesResponse)
	if len(messagesResponse.Data) == 0 {
		t.Fatalf("expected seeded message %q in message preview", subject)
	}

	ruleID := createHostedRule(t, server.URL, accountID, subject)

	var dryRunResponse struct {
		Data rules.DryRunResult `json:"data"`
	}
	postJSONInto(t, server.URL+"/api/rules/"+ruleID+"/dry-run", `{"imapAccountId":"`+accountID+`"}`, &dryRunResponse)
	if dryRunResponse.Data.MatchedCount == 0 {
		t.Fatalf("expected dry run to match seeded message")
	}
}

type hostedFixture struct {
	server   string
	port     int
	username string
	password string
	mailbox  string
	insecure bool
}

func requireHostedDovecotFixture(t *testing.T) hostedFixture {
	t.Helper()
	if os.Getenv("SMAILNAILD_DOVECOT_TEST") != "1" {
		t.Skip("set SMAILNAILD_DOVECOT_TEST=1 to run hosted HTTP integration tests")
	}

	return hostedFixture{
		server:   hostedGetenvDefault("SMAILNAILD_DOVECOT_SERVER", "localhost"),
		port:     hostedGetenvIntDefault("SMAILNAILD_DOVECOT_PORT", 993),
		username: hostedGetenvDefault("SMAILNAILD_DOVECOT_USERNAME", "a"),
		password: hostedGetenvDefault("SMAILNAILD_DOVECOT_PASSWORD", "pass"),
		mailbox:  hostedGetenvDefault("SMAILNAILD_DOVECOT_MAILBOX", "INBOX"),
		insecure: true,
	}
}

func acquireHostedDovecotLock(t *testing.T) {
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

func appendHostedFixtureMessage(t *testing.T, fixture hostedFixture, subject string) {
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

	raw := []byte(fmt.Sprintf("From: Seeder <seed@example.com>\r\nTo: User A <a@testcot>\r\nSubject: %s\r\nDate: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nHTTP hosted flow body for %s\r\n",
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

func createHostedAccount(t *testing.T, baseURL string, fixture hostedFixture) string {
	t.Helper()

	var response struct {
		Data accounts.Account `json:"data"`
	}
	postJSONInto(t, baseURL+"/api/accounts", fmt.Sprintf(`{
		"label":"Local Dovecot",
		"providerHint":"local",
		"server":"%s",
		"port":%d,
		"username":"%s",
		"password":"%s",
		"mailboxDefault":"%s",
		"insecure":true,
		"authKind":"password"
	}`, fixture.server, fixture.port, fixture.username, fixture.password, fixture.mailbox), &response)
	return response.Data.ID
}

func createHostedRule(t *testing.T, baseURL, accountID, subject string) string {
	t.Helper()

	body := fmt.Sprintf(`{
		"imapAccountId":"%s",
		"ruleYaml":"name: Hosted dry run\ndescription: Hosted dry run\nsearch:\n  subject_contains: %s\noutput:\n  format: json\n  limit: 10\n  fields:\n    - uid\n    - subject\n    - from\nactions:\n  move_to: Archive\n"
	}`, accountID, subject)

	var response struct {
		Data rules.RuleRecord `json:"data"`
	}
	postJSONInto(t, baseURL+"/api/rules", body, &response)
	return response.Data.ID
}

func postJSON(t *testing.T, url, body string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 300 {
		t.Fatalf("POST %s returned %d", url, res.StatusCode)
	}
}

func postJSONInto(t *testing.T, url, body string, target any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 300 {
		t.Fatalf("POST %s returned %d", url, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
}

func getJSON(t *testing.T, url string, target any) {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s error = %v", url, err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 300 {
		t.Fatalf("GET %s returned %d", url, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
}

func hostedGetenvDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func hostedGetenvIntDefault(key string, fallback int) int {
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
