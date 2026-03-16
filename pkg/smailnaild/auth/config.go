package auth

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

const (
	AuthSectionSlug          = "auth"
	AuthModeDev              = "dev"
	AuthModeSession          = "session"
	AuthModeOIDC             = "oidc"
	DefaultSessionCookieName = "smailnail_session"
)

type Settings struct {
	Mode              string   `glazed:"auth-mode"`
	DevUserID         string   `glazed:"auth-dev-user-id"`
	SessionCookieName string   `glazed:"auth-session-cookie-name"`
	OIDCIssuerURL     string   `glazed:"oidc-issuer-url"`
	OIDCClientID      string   `glazed:"oidc-client-id"`
	OIDCClientSecret  string   `glazed:"oidc-client-secret"`
	OIDCRedirectURL   string   `glazed:"oidc-redirect-url"`
	OIDCScopes        []string `glazed:"oidc-scopes"`
}

func NewSection() (schema.Section, error) {
	return schema.NewSection(
		AuthSectionSlug,
		"Authentication Settings",
		schema.WithFields(
			fields.New(
				"auth-mode",
				fields.TypeChoice,
				fields.WithHelp("Authentication mode for the hosted API"),
				fields.WithChoices(AuthModeDev, AuthModeSession, AuthModeOIDC),
				fields.WithDefault(AuthModeDev),
			),
			fields.New(
				"auth-dev-user-id",
				fields.TypeString,
				fields.WithHelp("Development fallback user ID used only in auth-mode=dev"),
				fields.WithDefault("local-user"),
			),
			fields.New(
				"auth-session-cookie-name",
				fields.TypeString,
				fields.WithHelp("Cookie name used for hosted smailnail sessions"),
				fields.WithDefault(DefaultSessionCookieName),
			),
			fields.New(
				"oidc-issuer-url",
				fields.TypeString,
				fields.WithHelp("OIDC issuer URL for the hosted web application"),
			),
			fields.New(
				"oidc-client-id",
				fields.TypeString,
				fields.WithHelp("OIDC client ID for the hosted web application"),
			),
			fields.New(
				"oidc-client-secret",
				fields.TypeString,
				fields.WithHelp("OIDC client secret for confidential web clients"),
			),
			fields.New(
				"oidc-redirect-url",
				fields.TypeString,
				fields.WithHelp("Callback URL used by the hosted web application"),
			),
			fields.New(
				"oidc-scopes",
				fields.TypeStringList,
				fields.WithHelp("Scopes requested during hosted web login"),
				fields.WithDefault([]string{"openid", "profile", "email"}),
			),
		),
	)
}

func LoadSettingsFromParsedValues(parsedValues *values.Values) (*Settings, error) {
	if parsedValues == nil {
		return nil, fmt.Errorf("parsed values are nil")
	}

	settings := &Settings{}
	if err := parsedValues.DecodeSectionInto(AuthSectionSlug, settings); err != nil {
		return nil, err
	}

	settings.Mode = normalizeMode(settings.Mode)
	settings.DevUserID = strings.TrimSpace(settings.DevUserID)
	if settings.DevUserID == "" {
		settings.DevUserID = "local-user"
	}
	settings.SessionCookieName = strings.TrimSpace(settings.SessionCookieName)
	if settings.SessionCookieName == "" {
		settings.SessionCookieName = DefaultSessionCookieName
	}
	settings.OIDCIssuerURL = strings.TrimSpace(settings.OIDCIssuerURL)
	settings.OIDCClientID = strings.TrimSpace(settings.OIDCClientID)
	settings.OIDCClientSecret = strings.TrimSpace(settings.OIDCClientSecret)
	settings.OIDCRedirectURL = strings.TrimSpace(settings.OIDCRedirectURL)

	return settings, nil
}

func normalizeMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case AuthModeSession:
		return AuthModeSession
	case AuthModeOIDC:
		return AuthModeOIDC
	default:
		return AuthModeDev
	}
}
