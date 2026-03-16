package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	defaultOIDCHTTPTimeout = 10 * time.Second
	defaultJWKSRefresh     = 5 * time.Minute
	authStateCookieName    = "smailnail_auth_state"
	authNonceCookieName    = "smailnail_auth_nonce"
)

type WebHandler interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleCallback(w http.ResponseWriter, r *http.Request)
	HandleLogout(w http.ResponseWriter, r *http.Request)
}

type oidcDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	JWKSURI               string `json:"jwks_uri"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

type idTokenClaims struct {
	jwt.Claims
	Nonce             string `json:"nonce,omitempty"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Name              string `json:"name,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

type OIDCAuthenticator struct {
	settings      *Settings
	identityRepo  *identity.Repository
	identitySvc   *identity.Service
	httpClient    *http.Client
	discovery     oidcDiscoveryDocument
	oauthConfig   oauth2.Config
	jwks          *jwksCache
	now           func() time.Time
	newID         func() string
	newRandom     func() (string, error)
	postLoginPath string
}

func NewOIDCAuthenticator(
	ctx context.Context,
	settings *Settings,
	identityRepo *identity.Repository,
	identitySvc *identity.Service,
) (*OIDCAuthenticator, error) {
	return newOIDCAuthenticatorWithClient(ctx, settings, identityRepo, identitySvc, &http.Client{Timeout: defaultOIDCHTTPTimeout})
}

func newOIDCAuthenticatorWithClient(
	ctx context.Context,
	settings *Settings,
	identityRepo *identity.Repository,
	identitySvc *identity.Service,
	httpClient *http.Client,
) (*OIDCAuthenticator, error) {
	if settings == nil {
		return nil, fmt.Errorf("oidc settings are nil")
	}
	if identityRepo == nil || identitySvc == nil {
		return nil, fmt.Errorf("identity dependencies are required")
	}
	if strings.TrimSpace(settings.OIDCIssuerURL) == "" {
		return nil, fmt.Errorf("oidc issuer url is required")
	}
	if strings.TrimSpace(settings.OIDCClientID) == "" {
		return nil, fmt.Errorf("oidc client id is required")
	}
	if strings.TrimSpace(settings.OIDCRedirectURL) == "" {
		return nil, fmt.Errorf("oidc redirect url is required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultOIDCHTTPTimeout}
	}

	discoveryURL := strings.TrimRight(settings.OIDCIssuerURL, "/") + "/.well-known/openid-configuration"
	discovery, err := fetchOIDCDiscovery(ctx, httpClient, discoveryURL)
	if err != nil {
		return nil, err
	}
	if discovery.Issuer == "" || discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" || discovery.JWKSURI == "" {
		return nil, fmt.Errorf("oidc discovery document is incomplete")
	}

	scopes := append([]string(nil), settings.OIDCScopes...)
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}

	return &OIDCAuthenticator{
		settings:     settings,
		identityRepo: identityRepo,
		identitySvc:  identitySvc,
		httpClient:   httpClient,
		discovery:    discovery,
		oauthConfig: oauth2.Config{
			ClientID:     settings.OIDCClientID,
			ClientSecret: settings.OIDCClientSecret,
			RedirectURL:  settings.OIDCRedirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  discovery.AuthorizationEndpoint,
				TokenURL: discovery.TokenEndpoint,
			},
		},
		jwks: &jwksCache{
			client:          httpClient,
			jwksURI:         discovery.JWKSURI,
			refreshInterval: defaultJWKSRefresh,
		},
		now:           func() time.Time { return time.Now().UTC() },
		newID:         uuid.NewString,
		newRandom:     randomToken,
		postLoginPath: "/",
	}, nil
}

func (a *OIDCAuthenticator) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := a.newRandom()
	if err != nil {
		http.Error(w, "failed to create auth state", http.StatusInternalServerError)
		return
	}
	nonce, err := a.newRandom()
	if err != nil {
		http.Error(w, "failed to create auth nonce", http.StatusInternalServerError)
		return
	}

	setShortLivedCookie(w, authStateCookieName, state)
	setShortLivedCookie(w, authNonceCookieName, nonce)

	url := a.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *OIDCAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) {
	queryState := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if queryState == "" || code == "" {
		http.Error(w, "missing oauth callback parameters", http.StatusBadRequest)
		return
	}

	stateCookie, err := r.Cookie(authStateCookieName)
	if err != nil || strings.TrimSpace(stateCookie.Value) == "" || stateCookie.Value != queryState {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	nonceCookie, err := r.Cookie(authNonceCookieName)
	if err != nil || strings.TrimSpace(nonceCookie.Value) == "" {
		http.Error(w, "missing oauth nonce", http.StatusBadRequest)
		return
	}
	clearCookie(w, authStateCookieName)
	clearCookie(w, authNonceCookieName)

	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, a.httpClient)
	token, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusBadGateway)
		return
	}

	rawIDToken, _ := token.Extra("id_token").(string)
	if strings.TrimSpace(rawIDToken) == "" {
		http.Error(w, "missing id_token in token response", http.StatusBadGateway)
		return
	}

	claims, err := a.verifyIDToken(ctx, rawIDToken, nonceCookie.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("id_token verification failed: %v", err), http.StatusBadGateway)
		return
	}

	resolved, err := a.identitySvc.ResolveOrProvisionUser(ctx, identity.ExternalPrincipal{
		Issuer:            claims.Issuer,
		Subject:           claims.Subject,
		ProviderKind:      identity.ProviderKindOIDC,
		ClientID:          a.settings.OIDCClientID,
		Email:             claims.Email,
		EmailVerified:     claims.EmailVerified,
		PreferredUsername: claims.PreferredUsername,
		DisplayName:       claims.Name,
		AvatarURL:         claims.Picture,
		Scopes:            append([]string(nil), a.oauthConfig.Scopes...),
		Claims: map[string]any{
			"email":              claims.Email,
			"email_verified":     claims.EmailVerified,
			"preferred_username": claims.PreferredUsername,
			"name":               claims.Name,
			"picture":            claims.Picture,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("user resolution failed: %v", err), http.StatusInternalServerError)
		return
	}

	expiry := token.Expiry
	if expiry.IsZero() {
		expiry = a.now().Add(24 * time.Hour)
	}
	sessionID := a.newID()
	if err := a.identityRepo.CreateSession(ctx, &identity.WebSession{
		ID:         sessionID,
		UserID:     resolved.User.ID,
		Issuer:     claims.Issuer,
		Subject:    claims.Subject,
		ExpiresAt:  expiry.UTC(),
		CreatedAt:  a.now(),
		LastSeenAt: a.now(),
	}); err != nil {
		http.Error(w, fmt.Sprintf("session creation failed: %v", err), http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, a.settings.SessionCookieName, sessionID, expiry)
	http.Redirect(w, r, a.postLoginPath, http.StatusSeeOther)
}

func (a *OIDCAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookieName := a.settings.SessionCookieName
	if cookieName == "" {
		cookieName = DefaultSessionCookieName
	}
	if cookie, err := r.Cookie(cookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		_ = a.identityRepo.DeleteSession(r.Context(), cookie.Value)
	}
	clearCookie(w, cookieName)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *OIDCAuthenticator) verifyIDToken(ctx context.Context, rawIDToken, expectedNonce string) (*idTokenClaims, error) {
	parsed, err := jwt.ParseSigned(rawIDToken)
	if err != nil {
		return nil, err
	}

	kid := ""
	if len(parsed.Headers) > 0 {
		kid = parsed.Headers[0].KeyID
	}

	keys, err := a.jwks.Keys(ctx, kid, false)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, key := range keys {
		var claims idTokenClaims
		if err := parsed.Claims(key.Key, &claims); err != nil {
			lastErr = err
			continue
		}

		expected := jwt.Expected{
			Issuer:   a.discovery.Issuer,
			Audience: jwt.Audience{a.settings.OIDCClientID},
			Time:     a.now(),
		}
		if err := claims.Validate(expected); err != nil {
			lastErr = err
			continue
		}
		if strings.TrimSpace(expectedNonce) != "" && claims.Nonce != expectedNonce {
			lastErr = fmt.Errorf("unexpected nonce")
			continue
		}
		return &claims, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no valid jwks keys")
	}
	return nil, lastErr
}

type jwksCache struct {
	client          *http.Client
	jwksURI         string
	refreshInterval time.Duration
	lastFetched     time.Time
	keySet          jose.JSONWebKeySet
}

func (c *jwksCache) Keys(ctx context.Context, kid string, forceRefresh bool) ([]jose.JSONWebKey, error) {
	if forceRefresh || len(c.keySet.Keys) == 0 || time.Since(c.lastFetched) >= c.refreshInterval {
		if err := c.refresh(ctx); err != nil {
			return nil, err
		}
	}

	if kid != "" {
		if keys := c.keySet.Key(kid); len(keys) > 0 {
			return keys, nil
		}
	}
	if len(c.keySet.Keys) == 0 {
		return nil, fmt.Errorf("jwks key set is empty")
	}
	return append([]jose.JSONWebKey(nil), c.keySet.Keys...), nil
}

func (c *jwksCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.jwksURI, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks request failed with status %d", resp.StatusCode)
	}

	var keySet jose.JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return err
	}
	c.keySet = keySet
	c.lastFetched = time.Now()
	return nil
}

func fetchOIDCDiscovery(ctx context.Context, client *http.Client, discoveryURL string) (oidcDiscoveryDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return oidcDiscoveryDocument{}, fmt.Errorf("oidc discovery request failed with status %d", resp.StatusCode)
	}

	var doc oidcDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return oidcDiscoveryDocument{}, err
	}
	return doc, nil
}

func setShortLivedCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((10 * time.Minute).Seconds()),
	})
}

func setSessionCookie(w http.ResponseWriter, name, value string, expiresAt time.Time) {
	if strings.TrimSpace(name) == "" {
		name = DefaultSessionCookieName
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt.UTC(),
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func randomToken() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
