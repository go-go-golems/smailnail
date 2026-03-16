package smailnaild

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	hostedauth "github.com/go-go-golems/smailnail/pkg/smailnaild/auth"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type fakeOIDCProvider struct {
	server   *httptest.Server
	clientID string
	signer   jose.Signer
	keySet   jose.JSONWebKeySet
}

func newFakeOIDCProvider(t *testing.T, clientID string) *fakeOIDCProvider {
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

	provider := &fakeOIDCProvider{
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
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                 provider.server.URL,
			"authorization_endpoint": provider.server.URL + "/authorize",
			"token_endpoint":         provider.server.URL + "/token",
			"jwks_uri":               provider.server.URL + "/jwks",
			"end_session_endpoint":   provider.server.URL + "/logout",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, provider.keySet)
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		nonce := strings.TrimSpace(r.Form.Get("nonce"))
		_ = nonce
		idToken := provider.mustSignIDToken(t, r.Form.Get("code"))
		writeJSON(w, http.StatusOK, map[string]any{
			"access_token": "access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     idToken,
		})
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	provider.server = httptest.NewServer(mux)
	t.Cleanup(provider.server.Close)

	return provider
}

func (p *fakeOIDCProvider) mustSignIDToken(t *testing.T, code string) string {
	t.Helper()

	parts := strings.SplitN(code, ":", 2)
	nonce := ""
	if len(parts) == 2 {
		nonce = parts[1]
	}

	now := time.Now().UTC()
	raw, err := josejwt.Signed(p.signer).Claims(struct {
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
			Subject:   "user-subject-1",
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
	return raw
}

func TestOIDCLoginCallbackAndSessionMe(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	provider := newFakeOIDCProvider(t, "smailnail-web")
	settings := &hostedauth.Settings{
		Mode:              hostedauth.AuthModeOIDC,
		SessionCookieName: hostedauth.DefaultSessionCookieName,
		OIDCIssuerURL:     provider.server.URL,
		OIDCClientID:      "smailnail-web",
		OIDCClientSecret:  "",
		OIDCRedirectURL:   "http://smailnail.test/auth/callback",
		OIDCScopes:        []string{"openid", "profile", "email"},
	}

	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo)
	webAuth, err := hostedauth.NewOIDCAuthenticator(context.Background(), settings, identityRepo, identityService)
	if err != nil {
		t.Fatalf("NewOIDCAuthenticator() error = %v", err)
	}

	handler := NewHandler(HandlerOptions{
		DB:           db,
		DBInfo:       DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC),
		UserResolver: SessionUserResolver{Repo: identityRepo, CookieName: hostedauth.DefaultSessionCookieName},
		WebAuth:      webAuth,
	})

	loginReq := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusFound {
		t.Fatalf("login status = %d body=%s", loginRec.Code, loginRec.Body.String())
	}

	loginLocation, err := url.Parse(loginRec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse redirect location: %v", err)
	}
	if loginLocation.String() == "" || !strings.HasPrefix(loginLocation.String(), provider.server.URL+"/authorize") {
		t.Fatalf("unexpected auth redirect location: %s", loginLocation.String())
	}

	stateCookie := cookieByName(loginRec.Result().Cookies(), "smailnail_auth_state")
	nonceCookie := cookieByName(loginRec.Result().Cookies(), "smailnail_auth_nonce")
	if stateCookie == nil || nonceCookie == nil {
		t.Fatalf("missing auth cookies: state=%v nonce=%v", stateCookie, nonceCookie)
	}

	callbackURL := "/auth/callback?state=" + url.QueryEscape(stateCookie.Value) + "&code=" + url.QueryEscape("test-code:"+nonceCookie.Value)
	callbackReq := httptest.NewRequest(http.MethodGet, callbackURL, nil)
	callbackReq.AddCookie(stateCookie)
	callbackReq.AddCookie(nonceCookie)
	callbackRec := httptest.NewRecorder()
	handler.ServeHTTP(callbackRec, callbackReq)

	if callbackRec.Code != http.StatusSeeOther {
		t.Fatalf("callback status = %d body=%s", callbackRec.Code, callbackRec.Body.String())
	}

	sessionCookie := cookieByName(callbackRec.Result().Cookies(), hostedauth.DefaultSessionCookieName)
	if sessionCookie == nil || strings.TrimSpace(sessionCookie.Value) == "" {
		t.Fatalf("missing session cookie: %+v", callbackRec.Result().Cookies())
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
		t.Fatalf("failed to decode me response: %v", err)
	}
	if mePayload.Data.PrimaryEmail != "intern@example.com" {
		t.Fatalf("unexpected me payload: %+v", mePayload.Data)
	}

	logoutReq := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusSeeOther {
		t.Fatalf("logout status = %d body=%s", logoutRec.Code, logoutRec.Body.String())
	}
	logoutLocation, err := url.Parse(logoutRec.Header().Get("Location"))
	if err != nil {
		t.Fatalf("failed to parse logout redirect: %v", err)
	}
	if logoutLocation.Scheme != "http" || logoutLocation.Host != strings.TrimPrefix(provider.server.URL, "http://") || logoutLocation.Path != "/logout" {
		t.Fatalf("unexpected logout redirect location: %s", logoutLocation.String())
	}
	if got := logoutLocation.Query().Get("client_id"); got != "smailnail-web" {
		t.Fatalf("logout client_id = %q", got)
	}
	if got := logoutLocation.Query().Get("post_logout_redirect_uri"); got != "http://smailnail.test/" {
		t.Fatalf("logout post_logout_redirect_uri = %q", got)
	}

	postLogoutReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	postLogoutReq.AddCookie(sessionCookie)
	postLogoutRec := httptest.NewRecorder()
	handler.ServeHTTP(postLogoutRec, postLogoutReq)

	if postLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("post logout status = %d body=%s", postLogoutRec.Code, postLogoutRec.Body.String())
	}
}

func cookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
