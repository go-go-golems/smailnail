package imapjs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestMountedHandlersCanBeServedByHostedHandler(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	mcpMux := http.NewServeMux()
	if err := MountHTTPHandlers(mcpMux, MountedOptions{
		Transport: "streamable_http",
		Auth:      HostedSettings{Enabled: true}.AuthOptions(""),
		DB:        db,
	}); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}

	handler := hostedapp.NewHandler(hostedapp.HandlerOptions{
		DB:         db,
		DBInfo:     hostedapp.DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:  time.Date(2026, 3, 16, 17, 0, 0, 0, time.UTC),
		MCPHandler: mcpMux,
	})

	apiReq := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	apiRec := httptest.NewRecorder()
	handler.ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusOK {
		t.Fatalf("api/info status = %d body=%s", apiRec.Code, apiRec.Body.String())
	}

	mcpReq := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	mcpRec := httptest.NewRecorder()
	handler.ServeHTTP(mcpRec, mcpReq)
	if mcpRec.Code == http.StatusNotFound {
		t.Fatalf("mcp route fell through to not found")
	}
}

func TestMountedHandlersKeepWebAndMCPAuthSeparated(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := hostedapp.BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	provider := newFakeWebOIDCProvider(t, "smailnail-mcp")
	mcpMux := http.NewServeMux()
	if err := MountHTTPHandlers(mcpMux, MountedOptions{
		Transport: "streamable_http",
		Auth: HostedSettings{
			Enabled:         true,
			AuthMode:        "external_oidc",
			AuthResourceURL: "https://smailnail.example.com/mcp",
			OIDCIssuerURL:   provider.server.URL,
		}.AuthOptions(""),
		DB: db,
	}); err != nil {
		t.Fatalf("MountHTTPHandlers() error = %v", err)
	}

	handler := hostedapp.NewHandler(hostedapp.HandlerOptions{
		DB:           db,
		DBInfo:       hostedapp.DatabaseInfo{Driver: "sqlite3", Target: ":memory:", Mode: "structured"},
		StartedAt:    time.Date(2026, 3, 16, 17, 0, 0, 0, time.UTC),
		UserResolver: hostedapp.SessionUserResolver{Repo: identity.NewRepository(db), CookieName: "smailnail_session"},
		MCPHandler:   mcpMux,
	})

	metadataReq := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil)
	metadataRec := httptest.NewRecorder()
	handler.ServeHTTP(metadataRec, metadataReq)
	if metadataRec.Code != http.StatusOK {
		t.Fatalf("metadata status = %d body=%s", metadataRec.Code, metadataRec.Body.String())
	}

	var metadata map[string]any
	if err := json.Unmarshal(metadataRec.Body.Bytes(), &metadata); err != nil {
		t.Fatalf("failed to decode metadata: %v", err)
	}
	if _, ok := metadata["authorization_servers"]; !ok {
		t.Fatalf("metadata missing authorization_servers: %#v", metadata)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusUnauthorized {
		t.Fatalf("api/me status = %d body=%s", meRec.Code, meRec.Body.String())
	}

	mcpReq := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	mcpRec := httptest.NewRecorder()
	handler.ServeHTTP(mcpRec, mcpReq)
	if mcpRec.Code != http.StatusUnauthorized {
		t.Fatalf("mcp status = %d body=%s", mcpRec.Code, mcpRec.Body.String())
	}
	if got := mcpRec.Header().Get("WWW-Authenticate"); got == "" {
		t.Fatalf("missing WWW-Authenticate header")
	}
}
