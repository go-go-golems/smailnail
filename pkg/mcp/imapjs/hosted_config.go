package imapjs

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
)

const HostedSectionSlug = "mcp"

type HostedSettings struct {
	Enabled            bool     `glazed:"mcp-enabled"`
	Transport          string   `glazed:"mcp-transport"`
	AuthMode           string   `glazed:"mcp-auth-mode"`
	AuthResourceURL    string   `glazed:"mcp-auth-resource-url"`
	OIDCIssuerURL      string   `glazed:"mcp-oidc-issuer-url"`
	OIDCDiscoveryURL   string   `glazed:"mcp-oidc-discovery-url"`
	OIDCAudience       string   `glazed:"mcp-oidc-audience"`
	OIDCRequiredScopes []string `glazed:"mcp-oidc-required-scopes"`
}

func NewHostedSection() (schema.Section, error) {
	return schema.NewSection(
		HostedSectionSlug,
		"Hosted MCP Settings",
		schema.WithFields(
			fields.New(
				"mcp-enabled",
				fields.TypeBool,
				fields.WithHelp("Expose the MCP HTTP endpoints from the hosted smailnaild server"),
				fields.WithDefault(true),
			),
			fields.New(
				"mcp-transport",
				fields.TypeChoice,
				fields.WithHelp("HTTP transport to expose for the hosted MCP surface"),
				fields.WithChoices("streamable_http", "sse"),
				fields.WithDefault("streamable_http"),
			),
			fields.New(
				"mcp-auth-mode",
				fields.TypeChoice,
				fields.WithHelp("Authentication mode for the hosted MCP surface"),
				fields.WithChoices(string(embeddable.AuthModeNone), string(embeddable.AuthModeExternalOIDC)),
				fields.WithDefault(string(embeddable.AuthModeNone)),
			),
			fields.New(
				"mcp-auth-resource-url",
				fields.TypeString,
				fields.WithHelp("Public MCP resource URL advertised to OAuth clients"),
			),
			fields.New(
				"mcp-oidc-issuer-url",
				fields.TypeString,
				fields.WithHelp("OIDC issuer used for hosted MCP bearer-token validation"),
			),
			fields.New(
				"mcp-oidc-discovery-url",
				fields.TypeString,
				fields.WithHelp("Optional override for MCP OIDC discovery"),
			),
			fields.New(
				"mcp-oidc-audience",
				fields.TypeString,
				fields.WithHelp("Required audience for hosted MCP bearer tokens"),
			),
			fields.New(
				"mcp-oidc-required-scopes",
				fields.TypeStringList,
				fields.WithHelp("Required scopes for hosted MCP bearer tokens"),
			),
		),
	)
}

func LoadHostedSettingsFromParsedValues(parsedValues *values.Values) (*HostedSettings, error) {
	if parsedValues == nil {
		return nil, fmt.Errorf("parsed values are nil")
	}

	settings := &HostedSettings{}
	if err := parsedValues.DecodeSectionInto(HostedSectionSlug, settings); err != nil {
		return nil, err
	}

	settings.Transport = normalizeTransport(settings.Transport)
	settings.AuthMode = normalizeMCPAuthMode(settings.AuthMode)
	settings.AuthResourceURL = strings.TrimSpace(settings.AuthResourceURL)
	settings.OIDCIssuerURL = strings.TrimSpace(settings.OIDCIssuerURL)
	settings.OIDCDiscoveryURL = strings.TrimSpace(settings.OIDCDiscoveryURL)
	settings.OIDCAudience = strings.TrimSpace(settings.OIDCAudience)

	return settings, nil
}

func (s HostedSettings) AuthOptions(defaultIssuer string) embeddable.AuthOptions {
	authMode := embeddable.AuthMode(s.AuthMode)
	issuerURL := s.OIDCIssuerURL
	if issuerURL == "" {
		issuerURL = strings.TrimSpace(defaultIssuer)
	}
	return embeddable.AuthOptions{
		Mode:        authMode,
		ResourceURL: s.AuthResourceURL,
		External: embeddable.ExternalOIDCOptions{
			IssuerURL:      issuerURL,
			DiscoveryURL:   s.OIDCDiscoveryURL,
			Audience:       s.OIDCAudience,
			RequiredScopes: append([]string(nil), s.OIDCRequiredScopes...),
		},
	}
}

func normalizeTransport(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "sse":
		return "sse"
	default:
		return "streamable_http"
	}
}

func normalizeMCPAuthMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case string(embeddable.AuthModeExternalOIDC):
		return string(embeddable.AuthModeExternalOIDC)
	default:
		return string(embeddable.AuthModeNone)
	}
}
