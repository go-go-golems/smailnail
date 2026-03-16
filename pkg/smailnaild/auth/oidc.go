package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
	EndSessionEndpoint    string `json:"end_session_endpoint"`
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
		log.Error().Err(err).Msg("Failed to create OIDC auth state")
		http.Error(w, "failed to create auth state", http.StatusInternalServerError)
		return
	}
	nonce, err := a.newRandom()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create OIDC auth nonce")
		http.Error(w, "failed to create auth nonce", http.StatusInternalServerError)
		return
	}

	setShortLivedCookie(w, authStateCookieName, state)
	setShortLivedCookie(w, authNonceCookieName, nonce)

	url := a.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
	log.Debug().
		Str("path", r.URL.Path).
		Str("issuer", a.discovery.Issuer).
		Str("client_id", a.settings.OIDCClientID).
		Str("redirect_url", a.oauthConfig.RedirectURL).
		Str("state", state).
		Str("nonce", nonce).
		Str("auth_url", url).
		Msg("Starting hosted OIDC login flow")
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *OIDCAuthenticator) HandleCallback(w http.ResponseWriter, r *http.Request) {
	queryState := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if queryState == "" || code == "" {
		log.Debug().
			Str("path", r.URL.Path).
			Str("query_state", queryState).
			Bool("has_code", code != "").
			Msg("OIDC callback is missing required parameters")
		http.Error(w, "missing oauth callback parameters", http.StatusBadRequest)
		return
	}

	stateCookie, err := r.Cookie(authStateCookieName)
	if err != nil || strings.TrimSpace(stateCookie.Value) == "" || stateCookie.Value != queryState {
		cookieValue := ""
		if err == nil && stateCookie != nil {
			cookieValue = stateCookie.Value
		}
		log.Debug().
			Err(err).
			Str("path", r.URL.Path).
			Str("query_state", queryState).
			Str("cookie_state", cookieValue).
			Str("cookie_name", authStateCookieName).
			Msg("OIDC callback state validation failed")
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	nonceCookie, err := r.Cookie(authNonceCookieName)
	if err != nil || strings.TrimSpace(nonceCookie.Value) == "" {
		log.Debug().
			Err(err).
			Str("path", r.URL.Path).
			Str("cookie_name", authNonceCookieName).
			Msg("OIDC callback nonce cookie is missing")
		http.Error(w, "missing oauth nonce", http.StatusBadRequest)
		return
	}
	log.Debug().
		Str("path", r.URL.Path).
		Str("query_state", queryState).
		Str("nonce", nonceCookie.Value).
		Msg("OIDC callback passed state and nonce cookie checks")
	clearCookie(w, authStateCookieName)
	clearCookie(w, authNonceCookieName)

	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, a.httpClient)
	token, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Error().
			Err(err).
			Str("issuer", a.discovery.Issuer).
			Str("client_id", a.settings.OIDCClientID).
			Str("redirect_url", a.oauthConfig.RedirectURL).
			Msg("OIDC token exchange failed")
		http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusBadGateway)
		return
	}
	log.Debug().
		Str("client_id", a.settings.OIDCClientID).
		Time("token_expiry", token.Expiry).
		Msg("OIDC token exchange succeeded")

	rawIDToken, _ := token.Extra("id_token").(string)
	if strings.TrimSpace(rawIDToken) == "" {
		log.Debug().Msg("OIDC token response is missing id_token")
		http.Error(w, "missing id_token in token response", http.StatusBadGateway)
		return
	}

	claims, err := a.verifyIDToken(ctx, rawIDToken, nonceCookie.Value)
	if err != nil {
		log.Error().
			Err(err).
			Str("issuer", a.discovery.Issuer).
			Str("client_id", a.settings.OIDCClientID).
			Msg("OIDC id_token verification failed")
		http.Error(w, fmt.Sprintf("id_token verification failed: %v", err), http.StatusBadGateway)
		return
	}
	log.Debug().
		Str("issuer", claims.Issuer).
		Str("subject", claims.Subject).
		Str("email", claims.Email).
		Str("preferred_username", claims.PreferredUsername).
		Msg("OIDC id_token verified")

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
		log.Error().
			Err(err).
			Str("issuer", claims.Issuer).
			Str("subject", claims.Subject).
			Msg("OIDC user resolution failed")
		http.Error(w, fmt.Sprintf("user resolution failed: %v", err), http.StatusInternalServerError)
		return
	}
	log.Debug().
		Str("user_id", resolved.User.ID).
		Str("issuer", claims.Issuer).
		Str("subject", claims.Subject).
		Msg("OIDC callback resolved local user")

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
		log.Error().
			Err(err).
			Str("user_id", resolved.User.ID).
			Str("session_id", sessionID).
			Msg("OIDC session creation failed")
		http.Error(w, fmt.Sprintf("session creation failed: %v", err), http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, a.settings.SessionCookieName, sessionID, expiry)
	log.Debug().
		Str("user_id", resolved.User.ID).
		Str("session_id", sessionID).
		Time("expires_at", expiry).
		Str("redirect_to", a.postLoginPath).
		Msg("OIDC callback created hosted session")
	http.Redirect(w, r, a.postLoginPath, http.StatusSeeOther)
}

func (a *OIDCAuthenticator) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookieName := a.settings.SessionCookieName
	if cookieName == "" {
		cookieName = DefaultSessionCookieName
	}
	if cookie, err := r.Cookie(cookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		log.Debug().
			Str("session_id", cookie.Value).
			Str("cookie_name", cookieName).
			Msg("Deleting hosted session during logout")
		_ = a.identityRepo.DeleteSession(r.Context(), cookie.Value)
	}
	clearCookie(w, cookieName)
	redirectURL := a.buildLogoutRedirectURL()
	log.Debug().
		Str("cookie_name", cookieName).
		Str("redirect_to", redirectURL).
		Msg("Completed hosted logout")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (a *OIDCAuthenticator) buildLogoutRedirectURL() string {
	if strings.TrimSpace(a.discovery.EndSessionEndpoint) == "" {
		return a.postLoginPath
	}

	endSessionURL, err := url.Parse(a.discovery.EndSessionEndpoint)
	if err != nil {
		log.Debug().
			Err(err).
			Str("end_session_endpoint", a.discovery.EndSessionEndpoint).
			Msg("Failed to parse OIDC end-session endpoint, falling back to post-login path")
		return a.postLoginPath
	}

	postLogoutURL, err := derivePostLogoutRedirectURL(a.oauthConfig.RedirectURL)
	if err != nil {
		log.Debug().
			Err(err).
			Str("redirect_url", a.oauthConfig.RedirectURL).
			Msg("Failed to derive post-logout redirect URL, falling back to post-login path")
		return a.postLoginPath
	}

	query := endSessionURL.Query()
	query.Set("post_logout_redirect_uri", postLogoutURL)
	if strings.TrimSpace(a.settings.OIDCClientID) != "" {
		query.Set("client_id", a.settings.OIDCClientID)
	}
	endSessionURL.RawQuery = query.Encode()
	return endSessionURL.String()
}

func derivePostLogoutRedirectURL(redirectURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(redirectURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("redirect url must include scheme and host")
	}
	parsed.Path = "/"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
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
	log.Debug().
		Str("kid", kid).
		Str("issuer", a.discovery.Issuer).
		Msg("Verifying OIDC id_token against JWKS")

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
	log.Debug().
		Str("discovery_url", discoveryURL).
		Msg("Fetching OIDC discovery document")
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
	log.Debug().
		Str("issuer", doc.Issuer).
		Str("authorization_endpoint", doc.AuthorizationEndpoint).
		Str("token_endpoint", doc.TokenEndpoint).
		Str("jwks_uri", doc.JWKSURI).
		Msg("Loaded OIDC discovery document")
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
