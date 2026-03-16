package smailnaild

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
)

const (
	UserIDHeader     = "X-Smailnail-User-ID"
	DefaultDevUserID = "local-user"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type UserResolver interface {
	ResolveUserID(r *http.Request) (string, error)
}

type HeaderUserResolver struct {
	DefaultUserID string
}

type SessionUserResolver struct {
	Repo       *identity.Repository
	CookieName string
}

func (r HeaderUserResolver) ResolveUserID(req *http.Request) (string, error) {
	if req != nil {
		userID := strings.TrimSpace(req.Header.Get(UserIDHeader))
		if userID != "" {
			return userID, nil
		}
	}

	if userID := strings.TrimSpace(r.DefaultUserID); userID != "" {
		return userID, nil
	}

	return "", ErrUnauthenticated
}

func (r SessionUserResolver) ResolveUserID(req *http.Request) (string, error) {
	if req == nil || r.Repo == nil {
		return "", ErrUnauthenticated
	}

	cookieName := strings.TrimSpace(r.CookieName)
	if cookieName == "" {
		cookieName = "smailnail_session"
	}

	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return "", ErrUnauthenticated
	}
	if strings.TrimSpace(cookie.Value) == "" {
		return "", ErrUnauthenticated
	}

	session, err := r.Repo.GetSessionByID(req.Context(), cookie.Value)
	if err != nil {
		return "", ErrUnauthenticated
	}
	if !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		return "", ErrUnauthenticated
	}

	userID := strings.TrimSpace(session.UserID)
	if userID == "" {
		return "", ErrUnauthenticated
	}
	return userID, nil
}
