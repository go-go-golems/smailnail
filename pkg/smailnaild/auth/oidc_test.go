package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieSecureFlagTracksTransport(t *testing.T) {
	authenticator := &OIDCAuthenticator{}
	authenticator.oauthConfig.RedirectURL = "https://smailnail.example.com/auth/callback"

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://smailnail.example.com/auth/login", nil)

	setShortLivedCookie(rec, authStateCookieName, "state", authenticator.shouldUseSecureCookies(req))
	cookie := cookieByName(rec.Result().Cookies(), authStateCookieName)
	if cookie == nil {
		t.Fatal("expected auth state cookie")
	}
	if !cookie.Secure {
		t.Fatal("expected cookie to be secure for https redirect URLs")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "http://smailnail.example.com/auth/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	setShortLivedCookie(rec, authNonceCookieName, "nonce", authenticator.shouldUseSecureCookies(req))
	cookie = cookieByName(rec.Result().Cookies(), authNonceCookieName)
	if cookie == nil {
		t.Fatal("expected auth nonce cookie")
	}
	if !cookie.Secure {
		t.Fatal("expected cookie to be secure behind an https proxy")
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
