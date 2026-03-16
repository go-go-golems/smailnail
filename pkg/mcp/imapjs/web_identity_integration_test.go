package imapjs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	hostedauth "github.com/go-go-golems/smailnail/pkg/smailnaild/auth"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/secrets"
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type fakeWebOIDCProvider struct {
	server   *httptest.Server
	clientID string
	signer   jose.Signer
	keySet   jose.JSONWebKeySet
}

func newFakeWebOIDCProvider(t *testing.T, clientID string) *fakeWebOIDCProvider {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key: jose.JSONWebKey{
			Key:       privateKey,
			KeyID:     "test-key",
			Algorithm: string(jose.RS256),
			Use:       "sig",
		},
	}, nil)
	if err != nil {
		t.Fatalf("jose.NewSigner() error = %v", err)
	}

	provider := &fakeWebOIDCProvider{
		clientID: clientID,
		signer:   signer,
		keySet: jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-key",
				Algorithm: string(jose.RS256),
				Use:       "sig",
			}},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(w, http.StatusOK, map[string]any{
			"issuer":                 provider.server.URL,
			"authorization_endpoint": provider.server.URL + "/authorize",
			"token_endpoint":         provider.server.URL + "/token",
			"jwks_uri":               provider.server.URL + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(w, http.StatusOK, provider.keySet)
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		idToken := provider.mustSignIDToken(t, r.FormValue("code"))
		writeTestJSON(w, http.StatusOK, map[string]any{
			"access_token": "access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     idToken,
		})
	})

	provider.server = httptest.NewServer(mux)
	t.Cleanup(provider.server.Close)

	return provider
}

func (p *fakeWebOIDCProvider) mustSignIDToken(t *testing.T, code string) string {
	t.Helper()

	parts := strings.SplitN(code, ":", 2)
	nonce := ""
	if len(parts) == 2 {
		nonce = parts[1]
	}

	now := time.Now().UTC()
	token, err := josejwt.Signed(p.signer).Claims(struct {
		josejwt.Claims
		Nonce             string `json:"nonce"`
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
		Picture           string `json:"picture"`
	}{
		Claims: josejwt.Claims{
			Issuer:    p.server.URL,
			Subject:   "subject-1",
			Audience:  josejwt.Audience{p.clientID},
			IssuedAt:  josejwt.NewNumericDate(now),
			Expiry:    josejwt.NewNumericDate(now.Add(1 * time.Hour)),
			NotBefore: josejwt.NewNumericDate(now.Add(-1 * time.Minute)),
		},
		Nonce:             nonce,
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		Name:              "Intern Example",
		Picture:           "https://example.com/avatar.png",
	}).CompactSerialize()
	if err != nil {
		t.Fatalf("CompactSerialize() error = %v", err)
	}
	return token
}

func TestBrowserOIDCAndMCPIdentityResolveSameLocalUser(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	provider := newFakeWebOIDCProvider(t, "smailnail-web")
	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	webAuth, err := hostedauth.NewOIDCAuthenticator(context.Background(), &hostedauth.Settings{
		Mode:              hostedauth.AuthModeOIDC,
		SessionCookieName: hostedauth.DefaultSessionCookieName,
		OIDCIssuerURL:     provider.server.URL,
		OIDCClientID:      "smailnail-web",
		OIDCRedirectURL:   "http://smailnail.test/auth/callback",
		OIDCScopes:        []string{"openid", "profile", "email"},
	}, identityRepo, identityService)
	if err != nil {
		t.Fatalf("NewOIDCAuthenticator() error = %v", err)
	}

	handler := hostedapp.NewHandler(hostedapp.HandlerOptions{
		DB:           db,
		DBInfo:       hostedapp.DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC),
		UserResolver: hostedapp.SessionUserResolver{Repo: identityRepo, CookieName: hostedauth.DefaultSessionCookieName},
		WebAuth:      webAuth,
	})

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	stateCookie := testCookieByName(loginRec.Result().Cookies(), "smailnail_auth_state")
	nonceCookie := testCookieByName(loginRec.Result().Cookies(), "smailnail_auth_nonce")
	if stateCookie == nil || nonceCookie == nil {
		t.Fatalf("missing auth cookies")
	}

	callbackReq := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+url.QueryEscape(stateCookie.Value)+"&code="+url.QueryEscape("code:"+nonceCookie.Value), nil)
	callbackReq.AddCookie(stateCookie)
	callbackReq.AddCookie(nonceCookie)
	callbackRec := httptest.NewRecorder()
	handler.ServeHTTP(callbackRec, callbackReq)

	sessionCookie := testCookieByName(callbackRec.Result().Cookies(), hostedauth.DefaultSessionCookieName)
	if sessionCookie == nil {
		t.Fatalf("missing session cookie")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.AddCookie(sessionCookie)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d body=%s", meRec.Code, meRec.Body.String())
	}

	var mePayload struct {
		Data identity.User `json:"data"`
	}
	if err := json.Unmarshal(meRec.Body.Bytes(), &mePayload); err != nil {
		t.Fatalf("unmarshal me response: %v", err)
	}

	runtime := newSharedIdentityRuntimeWithDB(db)
	middleware := runtime.middleware()
	ctx := embeddable.WithAuthPrincipal(context.Background(), embeddable.AuthPrincipal{
		Issuer:            provider.server.URL,
		Subject:           "subject-1",
		ClientID:          "smailnail-mcp",
		Email:             "intern@example.com",
		EmailVerified:     true,
		PreferredUsername: "intern",
		DisplayName:       "Intern Example",
		AvatarURL:         "https://example.com/avatar.png",
		Scopes:            []string{"openid", "profile", "email"},
	})

	_, err = middleware(func(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
		resolved, ok := ResolvedIdentityFromContext(ctx)
		if !ok {
			t.Fatalf("expected resolved identity in context")
		}
		if resolved.User.ID != mePayload.Data.ID {
			t.Fatalf("resolved MCP user id = %q, want %q", resolved.User.ID, mePayload.Data.ID)
		}
		return newJSONToolResult(map[string]any{"userID": resolved.User.ID})
	})(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("middleware handler error = %v", err)
	}
}

func TestBrowserCreatedAccountCanBeUsedThroughMCP(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	provider := newFakeWebOIDCProvider(t, "smailnail-web")
	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	webAuth, err := hostedauth.NewOIDCAuthenticator(context.Background(), &hostedauth.Settings{
		Mode:              hostedauth.AuthModeOIDC,
		SessionCookieName: hostedauth.DefaultSessionCookieName,
		OIDCIssuerURL:     provider.server.URL,
		OIDCClientID:      "smailnail-web",
		OIDCRedirectURL:   "http://smailnail.test/auth/callback",
		OIDCScopes:        []string{"openid", "profile", "email"},
	}, identityRepo, identityService)
	if err != nil {
		t.Fatalf("NewOIDCAuthenticator() error = %v", err)
	}

	handler := hostedapp.NewHandler(hostedapp.HandlerOptions{
		DB:           db,
		DBInfo:       hostedapp.DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC),
		UserResolver: hostedapp.SessionUserResolver{Repo: identityRepo, CookieName: hostedauth.DefaultSessionCookieName},
		WebAuth:      webAuth,
	})

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	stateCookie := testCookieByName(loginRec.Result().Cookies(), "smailnail_auth_state")
	nonceCookie := testCookieByName(loginRec.Result().Cookies(), "smailnail_auth_nonce")
	callbackReq := httptest.NewRequest(http.MethodGet, "/auth/callback?state="+url.QueryEscape(stateCookie.Value)+"&code="+url.QueryEscape("code:"+nonceCookie.Value), nil)
	callbackReq.AddCookie(stateCookie)
	callbackReq.AddCookie(nonceCookie)
	callbackRec := httptest.NewRecorder()
	handler.ServeHTTP(callbackRec, callbackReq)

	sessionCookie := testCookieByName(callbackRec.Result().Cookies(), hostedauth.DefaultSessionCookieName)
	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.AddCookie(sessionCookie)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)

	var mePayload struct {
		Data identity.User `json:"data"`
	}
	if err := json.Unmarshal(meRec.Body.Bytes(), &mePayload); err != nil {
		t.Fatalf("unmarshal me response: %v", err)
	}

	secretConfig, err := secrets.LoadConfigFromSettings(&secrets.Settings{
		KeyBase64: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		KeyID:     "test-key",
	})
	if err != nil {
		t.Fatalf("LoadConfigFromSettings() error = %v", err)
	}

	accountService := accounts.NewService(accounts.NewRepository(db), secretConfig)
	account, err := accountService.Create(context.Background(), mePayload.Data.ID, accounts.CreateInput{
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
			Issuer:            provider.server.URL,
			Subject:           "subject-1",
			ClientID:          "smailnail-mcp",
			Email:             "intern@example.com",
			EmailVerified:     true,
			PreferredUsername: "intern",
			DisplayName:       "Intern Example",
			AvatarURL:         "https://example.com/avatar.png",
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
	if dialer.gotOpts.Username != "user@example.com" || dialer.gotOpts.Password != "secret" {
		t.Fatalf("unexpected dialer opts: %+v", dialer.gotOpts)
	}
}

func writeTestJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func testCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
