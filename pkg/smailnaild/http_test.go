package smailnaild

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewHandlerHealthAndInfo(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(t.Context(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	startedAt := time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)
	handler := NewHandler(db, DatabaseInfo{
		Driver: "sqlite3",
		Target: ":memory:",
		Mode:   "structured",
	}, startedAt)

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", healthRec.Code)
	}

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyRec := httptest.NewRecorder()
	handler.ServeHTTP(readyRec, readyReq)
	if readyRec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d", readyRec.Code)
	}

	infoReq := httptest.NewRequest(http.MethodGet, "/api/info", nil)
	infoRec := httptest.NewRecorder()
	handler.ServeHTTP(infoRec, infoReq)
	if infoRec.Code != http.StatusOK {
		t.Fatalf("api/info status = %d", infoRec.Code)
	}

	var payload infoResponse
	if err := json.Unmarshal(infoRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode info response: %v", err)
	}
	if payload.Service != "smailnaild" {
		t.Fatalf("unexpected service: %q", payload.Service)
	}
	if payload.Database.Driver != "sqlite3" {
		t.Fatalf("unexpected database driver: %q", payload.Database.Driver)
	}
	if !payload.StartedAt.Equal(startedAt) {
		t.Fatalf("unexpected startedAt: %v", payload.StartedAt)
	}
}
