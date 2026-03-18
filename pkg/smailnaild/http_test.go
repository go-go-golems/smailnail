package smailnaild

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type fakeAccountAPI struct {
	lastUserID         string
	lastCreateInput    accounts.CreateInput
	lastUpdateInput    accounts.UpdateInput
	lastListInput      accounts.ListMessagesInput
	lastMessageMailbox string
	lastMessageUID     uint32
}

func (f *fakeAccountAPI) List(_ context.Context, userID string) ([]accounts.AccountListItem, error) {
	f.lastUserID = userID
	return []accounts.AccountListItem{
		{Account: accounts.Account{ID: "acc-1", Label: "Local"}},
	}, nil
}

func (f *fakeAccountAPI) Create(_ context.Context, userID string, input accounts.CreateInput) (*accounts.Account, error) {
	f.lastUserID = userID
	f.lastCreateInput = input
	return &accounts.Account{ID: "acc-1", Label: input.Label, Username: input.Username}, nil
}

func (f *fakeAccountAPI) Get(_ context.Context, userID, accountID string) (*accounts.Account, error) {
	f.lastUserID = userID
	return &accounts.Account{ID: accountID, Label: "Local"}, nil
}

func (f *fakeAccountAPI) Update(_ context.Context, userID, accountID string, input accounts.UpdateInput) (*accounts.Account, error) {
	f.lastUserID = userID
	f.lastUpdateInput = input
	return &accounts.Account{ID: accountID, Label: derefString(input.Label, "Local")}, nil
}

func (f *fakeAccountAPI) Delete(_ context.Context, userID, accountID string) error {
	f.lastUserID = userID
	return nil
}

func (f *fakeAccountAPI) RunTest(_ context.Context, userID, accountID string, input accounts.TestInput) (*accounts.TestResult, error) {
	f.lastUserID = userID
	return &accounts.TestResult{
		ID:            "test-1",
		IMAPAccountID: accountID,
		TestMode:      input.Mode,
		Success:       true,
		Details:       map[string]any{"mailbox": "INBOX"},
	}, nil
}

func (f *fakeAccountAPI) ListMailboxes(_ context.Context, userID, accountID string) ([]accounts.MailboxInfo, error) {
	f.lastUserID = userID
	return []accounts.MailboxInfo{{Name: "INBOX", Path: "INBOX"}}, nil
}

func (f *fakeAccountAPI) ListMessages(_ context.Context, userID, accountID string, input accounts.ListMessagesInput) ([]accounts.MessageView, string, error) {
	f.lastUserID = userID
	f.lastListInput = input
	return []accounts.MessageView{{UID: 42, Subject: "Invoice", TotalCount: 1}}, "INBOX", nil
}

func (f *fakeAccountAPI) GetMessage(_ context.Context, userID, accountID, mailbox string, uid uint32) (*accounts.MessageView, string, error) {
	f.lastUserID = userID
	f.lastMessageMailbox = mailbox
	f.lastMessageUID = uid
	return &accounts.MessageView{UID: uid, Subject: "Invoice"}, mailbox, nil
}

type fakeRuleAPI struct {
	lastUserID      string
	lastCreateInput rules.CreateInput
	lastDryRunInput rules.DryRunInput
}

func (f *fakeRuleAPI) List(_ context.Context, userID string) ([]rules.RuleRecord, error) {
	f.lastUserID = userID
	return []rules.RuleRecord{{ID: "rule-1", Name: "Invoices", Status: "draft"}}, nil
}

func (f *fakeRuleAPI) Create(_ context.Context, userID string, input rules.CreateInput) (*rules.RuleRecord, error) {
	f.lastUserID = userID
	f.lastCreateInput = input
	return &rules.RuleRecord{ID: "rule-1", Name: "Invoices", Status: "draft", RuleYAML: input.RuleYAML}, nil
}

func (f *fakeRuleAPI) Get(_ context.Context, userID, ruleID string) (*rules.RuleRecord, error) {
	f.lastUserID = userID
	return &rules.RuleRecord{ID: ruleID, Name: "Invoices", Status: "draft"}, nil
}

func (f *fakeRuleAPI) Update(_ context.Context, userID, ruleID string, input rules.UpdateInput) (*rules.RuleRecord, error) {
	f.lastUserID = userID
	return &rules.RuleRecord{ID: ruleID, Name: derefString(input.Name, "Invoices"), Status: "draft"}, nil
}

func (f *fakeRuleAPI) Delete(_ context.Context, userID, ruleID string) error {
	f.lastUserID = userID
	return nil
}

func (f *fakeRuleAPI) DryRun(_ context.Context, userID, ruleID string, input rules.DryRunInput) (*rules.DryRunResult, error) {
	f.lastUserID = userID
	f.lastDryRunInput = input
	return &rules.DryRunResult{
		RuleID:        ruleID,
		IMAPAccountID: input.IMAPAccountID,
		MatchedCount:  1,
		ActionPlan:    map[string]any{"moveTo": "Archive"},
		CreatedAt:     time.Date(2026, 3, 16, 10, 14, 0, 0, time.UTC),
	}, nil
}

func TestNewHandlerHealthAndInfo(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	startedAt := time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)
	handler := NewHandler(HandlerOptions{
		DB:        db,
		DBInfo:    DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt: startedAt,
	})

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", healthRec.Code)
	}

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyRec := httptest.NewRecorder()
	handler.ServeHTTP(readyRec, readyReq)
	if readyRec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d", readyRec.Code)
	}

	infoReq := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	infoRec := httptest.NewRecorder()
	handler.ServeHTTP(infoRec, infoReq)
	if infoRec.Code != http.StatusOK {
		t.Fatalf("api/info status = %d", infoRec.Code)
	}

	var payload infoResponse
	if err := json.Unmarshal(infoRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode info response: %v", err)
	}
	if payload.Service != "smailnaild" {
		t.Fatalf("unexpected service: %q", payload.Service)
	}
	if payload.Database.Driver != "sqlite3" {
		t.Fatalf("unexpected database driver: %q", payload.Database.Driver)
	}
	if !payload.StartedAt.Equal(startedAt) {
		t.Fatalf("unexpected startedAt: %v", payload.StartedAt)
	}
}

func TestNewHandlerSkipsReadyzDebugLogs(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	var buf bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&buf).Level(zerolog.DebugLevel)
	defer func() {
		log.Logger = previousLogger
	}()

	handler := NewHandler(HandlerOptions{
		DB:        db,
		DBInfo:    DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt: time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d", rec.Code)
	}
	if strings.Contains(buf.String(), "/readyz") || strings.Contains(buf.String(), "Hosted request started") || strings.Contains(buf.String(), "Hosted request completed") {
		t.Fatalf("expected /readyz to be excluded from hosted debug logs, got logs: %s", buf.String())
	}
}

func TestNewHandlerAccountAndRuleRoutes(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	accountAPI := &fakeAccountAPI{}
	ruleAPI := &fakeRuleAPI{}
	handler := NewHandler(HandlerOptions{
		DB:           db,
		DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
		UserResolver: HeaderUserResolver{DefaultUserID: "local-user"},
		AccountAPI:   accountAPI,
		RuleAPI:      ruleAPI,
	})

	t.Run("list accounts uses header user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
		req.Header.Set(UserIDHeader, "alice")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if accountAPI.lastUserID != "alice" {
			t.Fatalf("userID = %q", accountAPI.lastUserID)
		}
	})

	t.Run("create account decodes json", func(t *testing.T) {
		body := bytes.NewBufferString(`{"label":"Work","server":"localhost","port":993,"username":"a","password":"pass","mailboxDefault":"INBOX","authKind":"password"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/accounts", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if accountAPI.lastCreateInput.Label != "Work" {
			t.Fatalf("last label = %q", accountAPI.lastCreateInput.Label)
		}
	})

	t.Run("list messages parses query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/accounts/acc-1/messages?mailbox=Archive&limit=5&offset=2&query=invoice&unread_only=1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if accountAPI.lastListInput.Mailbox != "Archive" || accountAPI.lastListInput.Limit != 5 || accountAPI.lastListInput.Offset != 2 || !accountAPI.lastListInput.UnreadOnly {
			t.Fatalf("unexpected list input: %+v", accountAPI.lastListInput)
		}
	})

	t.Run("get message parses uid and mailbox", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/accounts/acc-1/messages/42?mailbox=INBOX", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if accountAPI.lastMessageUID != 42 || accountAPI.lastMessageMailbox != "INBOX" {
			t.Fatalf("unexpected message lookup: uid=%d mailbox=%q", accountAPI.lastMessageUID, accountAPI.lastMessageMailbox)
		}
	})

	t.Run("dry run rule decodes json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/rules/rule-1/dry-run", bytes.NewBufferString(`{"imapAccountId":"acc-1"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if ruleAPI.lastDryRunInput.IMAPAccountID != "acc-1" {
			t.Fatalf("unexpected dry-run input: %+v", ruleAPI.lastDryRunInput)
		}
	})

	t.Run("missing auth returns unauthorized", func(t *testing.T) {
		protected := NewHandler(HandlerOptions{
			DB:           db,
			DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
			StartedAt:    time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
			UserResolver: HeaderUserResolver{},
			AccountAPI:   accountAPI,
			RuleAPI:      ruleAPI,
		})

		req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
	})
}

func TestNewHandlerMeRouteWithSessionCookie(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := identity.NewRepository(db)
	if err := repo.CreateUser(t.Context(), &identity.User{
		ID:           "user-1",
		PrimaryEmail: "intern@example.com",
		DisplayName:  "Intern",
		AvatarURL:    "https://example.com/avatar.png",
	}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if err := repo.CreateSession(t.Context(), &identity.WebSession{
		ID:         "session-1",
		UserID:     "user-1",
		Issuer:     "https://auth.example.com/realms/smailnail",
		Subject:    "abc123",
		ExpiresAt:  time.Now().UTC().Add(1 * time.Hour),
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	handler := NewHandler(HandlerOptions{
		DB:           db,
		DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
		UserResolver: SessionUserResolver{Repo: repo, CookieName: "smailnail_session"},
		AccountAPI:   &fakeAccountAPI{},
		RuleAPI:      &fakeRuleAPI{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "smailnail_session", Value: "session-1"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Data identity.User `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode me response: %v", err)
	}
	if payload.Data.ID != "user-1" {
		t.Fatalf("user id = %q", payload.Data.ID)
	}
}

func TestNewHandlerMeRouteRejectsExpiredSession(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	repo := identity.NewRepository(db)
	if err := repo.CreateUser(t.Context(), &identity.User{
		ID:           "user-1",
		PrimaryEmail: "intern@example.com",
		DisplayName:  "Intern",
	}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if err := repo.CreateSession(t.Context(), &identity.WebSession{
		ID:         "expired-session",
		UserID:     "user-1",
		Issuer:     "https://auth.example.com/realms/smailnail",
		Subject:    "abc123",
		ExpiresAt:  time.Now().UTC().Add(-1 * time.Minute),
		CreatedAt:  time.Now().UTC().Add(-2 * time.Hour),
		LastSeenAt: time.Now().UTC().Add(-10 * time.Minute),
	}); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	handler := NewHandler(HandlerOptions{
		DB:           db,
		DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
		UserResolver: SessionUserResolver{Repo: repo, CookieName: "smailnail_session"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "smailnail_session", Value: "expired-session"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func derefString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}
