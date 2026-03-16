package smailnaild

import (
	"net/http"
	"strings"
)

const (
	UserIDHeader     = "X-Smailnail-User-ID"
	DefaultDevUserID = "local-user"
)

type UserResolver interface {
	ResolveUserID(r *http.Request) string
}

type HeaderUserResolver struct {
	DefaultUserID string
}

func (r HeaderUserResolver) ResolveUserID(req *http.Request) string {
	if req != nil {
		userID := strings.TrimSpace(req.Header.Get(UserIDHeader))
		if userID != "" {
			return userID
		}
	}

	if userID := strings.TrimSpace(r.DefaultUserID); userID != "" {
		return userID
	}

	return DefaultDevUserID
}
