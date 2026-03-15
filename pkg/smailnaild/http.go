package smailnaild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type ServerOptions struct {
	Host   string
	Port   int
	DB     *sqlx.DB
	DBInfo DatabaseInfo
}

type infoResponse struct {
	Service   string       `json:"service"`
	Version   string       `json:"version"`
	StartedAt time.Time    `json:"startedAt"`
	Database  DatabaseInfo `json:"database"`
}

func NewHTTPServer(options ServerOptions) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", options.Host, options.Port),
		Handler:           NewHandler(options.DB, options.DBInfo, time.Now().UTC()),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func NewHandler(db *sqlx.DB, dbInfo DatabaseInfo, startedAt time.Time) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := PingDatabase(r.Context(), db); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not-ready",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, infoResponse{
			Service:   "smailnaild",
			Version:   "dev",
			StartedAt: startedAt,
			Database:  dbInfo,
		})
	})

	return mux
}

func ShutdownServer(ctx context.Context, server *http.Server) error {
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
