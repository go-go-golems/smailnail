package smailnaild

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/rs/zerolog/log"
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
			log.Debug().
				Str("path", req.URL.Path).
				Str("resolver", "header").
				Str("user_id", userID).
				Msg("Resolved hosted user from request header")
			return userID, nil
		}
	}

	if userID := strings.TrimSpace(r.DefaultUserID); userID != "" {
		log.Debug().
			Str("path", requestPath(req)).
			Str("resolver", "dev-default").
			Str("user_id", userID).
			Msg("Resolved hosted user from development fallback")
		return userID, nil
	}

	log.Debug().
		Str("path", requestPath(req)).
		Str("resolver", "header").
		Msg("No hosted user could be resolved from header or fallback")
	return "", ErrUnauthenticated
}

func (r SessionUserResolver) ResolveUserID(req *http.Request) (string, error) {
	if req == nil || r.Repo == nil {
		log.Debug().
			Str("resolver", "session").
			Msg("Session resolver is missing request or repository")
		return "", ErrUnauthenticated
	}

	cookieName := strings.TrimSpace(r.CookieName)
	if cookieName == "" {
		cookieName = "smailnail_session"
	}

	cookie, err := req.Cookie(cookieName)
	if err != nil {
		log.Debug().
			Str("path", req.URL.Path).
			Str("resolver", "session").
			Str("cookie_name", cookieName).
			Msg("Session cookie not found")
		return "", ErrUnauthenticated
	}
	if strings.TrimSpace(cookie.Value) == "" {
		log.Debug().
			Str("path", req.URL.Path).
			Str("resolver", "session").
			Str("cookie_name", cookieName).
			Msg("Session cookie is empty")
		return "", ErrUnauthenticated
	}

	session, err := r.Repo.GetSessionByID(req.Context(), cookie.Value)
	if err != nil {
		log.Debug().
			Err(err).
			Str("path", req.URL.Path).
			Str("resolver", "session").
			Str("cookie_name", cookieName).
			Msg("Session lookup failed")
		return "", ErrUnauthenticated
	}
	if !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		log.Debug().
			Str("path", req.URL.Path).
			Str("resolver", "session").
			Time("expired_at", session.ExpiresAt).
			Msg("Session has expired")
		return "", ErrUnauthenticated
	}

	userID := strings.TrimSpace(session.UserID)
	if userID == "" {
		log.Debug().
			Str("path", req.URL.Path).
			Str("resolver", "session").
			Msg("Resolved session is missing user id")
		return "", ErrUnauthenticated
	}
	log.Debug().
		Str("path", req.URL.Path).
		Str("resolver", "session").
		Str("user_id", userID).
		Str("session_id", session.ID).
		Msg("Resolved hosted user from session")
	return userID, nil
}

func requestPath(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Path
}
