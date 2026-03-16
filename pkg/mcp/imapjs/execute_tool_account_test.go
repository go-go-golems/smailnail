package imapjs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
)

type testDialer struct {
	gotOpts smailnailjs.ConnectOptions
}

type testSession struct {
	mailbox string
}

func (d *testDialer) Dial(_ context.Context, opts smailnailjs.ConnectOptions) (smailnailjs.Session, error) {
	d.gotOpts = opts
	mailbox := opts.Mailbox
	if mailbox == "" {
		mailbox = "INBOX"
	}
	return &testSession{mailbox: mailbox}, nil
}

func (s *testSession) Mailbox() string {
	return s.mailbox
}

func (s *testSession) Close() {}

func TestExecuteIMAPJSHandlerConnectsUsingStoredAccount(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	secretConfig, err := secrets.LoadConfigFromSettings(&secrets.Settings{
		KeyBase64: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		KeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}

	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	resolved, err := identityService.ResolveOrProvisionUser(context.Background(), identity.ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "subject-1",
		ProviderKind:      identity.ProviderKindOIDC,
		ClientID:          "smailnail-web",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Intern Example",
	})
	if err != nil {
		t.Fatalf("ResolveOrProvisionUser() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), secretConfig)
	account, err := accountService.Create(context.Background(), resolved.User.ID, accounts.CreateInput{
		Label:          "Work",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "user@example.com",
		Password:       "secret",
		MailboxDefault: "Archive",
		AuthKind:       accounts.AuthKindPassword,
		MCPEnabled:     true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	runtime := newSharedIdentityRuntimeWithServices(db, identityService, accountService)
	dialer := &testDialer{}
	ctx := withDialer(
		embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
			Issuer:            "https://auth.example.com/realms/smailnail",
			Subject:           "subject-1",
			ClientID:          "smailnail-mcp",
			Email:             "intern@example.com",
			EmailVerified:     true,
			PreferredUsername: "intern",
			DisplayName:       "Intern Example",
		}),
		dialer,
	)

	result, err := runtime.middleware()(executeIMAPJSHandler)(ctx, map[string]interface{}{
		"code": `
const smailnail = require("smailnail");
const svc = smailnail.newService();
const session = svc.connect({ accountId: "` + account.ID + `" });
session.mailbox;
`,
	})
	if err != nil {
		t.Fatalf("executeIMAPJSHandler error = %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got %#v", result)
	}

	var decoded ExecuteIMAPJSResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.Value != "Archive" {
		t.Fatalf("decoded.Value = %#v, want Archive", decoded.Value)
	}
	if dialer.gotOpts.Username != "user@example.com" || dialer.gotOpts.Password != "secret" {
		t.Fatalf("unexpected dialer opts: %+v", dialer.gotOpts)
	}
}

func TestExecuteIMAPJSHandlerRejectsCrossUserStoredAccount(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	secretConfig, err := secrets.LoadConfigFromSettings(&secrets.Settings{
		KeyBase64: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		KeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}

	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	otherUser, err := identityService.ResolveOrProvisionUser(context.Background(), identity.ExternalPrincipal{
		Issuer:            "https://auth.example.com/realms/smailnail",
		Subject:           "other-subject",
		ProviderKind:      identity.ProviderKindOIDC,
		ClientID:          "smailnail-web",
		Email:             "other@example.com",
		EmailVerified:     true,
		PreferredUsername: "other",
		DisplayName:       "Other User",
	})
	if err != nil {
		t.Fatalf("ResolveOrProvisionUser() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), secretConfig)
	account, err := accountService.Create(context.Background(), otherUser.User.ID, accounts.CreateInput{
		Label:          "Other",
		Server:         "imap.example.com",
		Port:           993,
		Username:       "other@example.com",
		Password:       "secret",
		MailboxDefault: "INBOX",
		AuthKind:       accounts.AuthKindPassword,
		MCPEnabled:     true,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	runtime := newSharedIdentityRuntimeWithServices(db, identityService, accountService)
	result, err := runtime.middleware()(executeIMAPJSHandler)(
		embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
			Issuer:            "https://auth.example.com/realms/smailnail",
			Subject:           "subject-1",
			ClientID:          "smailnail-mcp",
			Email:             "intern@example.com",
			EmailVerified:     true,
			PreferredUsername: "intern",
			DisplayName:       "Intern Example",
		}),
		map[string]interface{}{
			"code": `
const smailnail = require("smailnail");
const svc = smailnail.newService();
svc.connect({ accountId: "` + account.ID + `" });
`,
		},
	)
	if err != nil {
		t.Fatalf("executeIMAPJSHandler error = %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result")
	}
}
