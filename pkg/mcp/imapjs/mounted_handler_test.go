package imapjs

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hostedapp "github.com/go-go-golems/smailnail/pkg/smailnaild"
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
