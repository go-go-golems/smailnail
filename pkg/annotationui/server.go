package annotationui

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type ServerOptions struct {
	Host       string
	Port       int
	DB         *sqlx.DB
	DBInfo     DatabaseInfo
	QueryDirs  []string
	PresetDirs []string
	PublicFS   fs.FS
}

type HandlerOptions struct {
	DB         *sqlx.DB
	DBInfo     DatabaseInfo
	StartedAt  time.Time
	QueryDirs  []string
	PresetDirs []string
	PublicFS   fs.FS
}

type appHandler struct {
	db          *sqlx.DB
	dbInfo      DatabaseInfo
	startedAt   time.Time
	annotations *annotate.Repository
	queryDirs   []string
	presetDirs  []string
	publicFS    fs.FS
}

const defaultAnnotationUIActor = "local-reviewer"

func NewHTTPServer(options ServerOptions) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf("%s:%d", options.Host, options.Port),
		Handler: NewHandler(HandlerOptions{
			DB:         options.DB,
			DBInfo:     options.DBInfo,
			StartedAt:  time.Now().UTC(),
			QueryDirs:  options.QueryDirs,
			PresetDirs: options.PresetDirs,
			PublicFS:   options.PublicFS,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func RunServer(ctx context.Context, server *http.Server) error {
	if server == nil {
		return errors.New("http server is nil")
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		log.Info().
			Str("address", server.Addr).
			Msg("Starting sqlite annotation server")
		err := server.ListenAndServe()
		if err == nil || stderrors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.Wrap(err, "listen and serve sqlite annotation server")
	})
	group.Go(func() error {
		<-groupCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !stderrors.Is(err, http.ErrServerClosed) {
			return errors.Wrap(err, "shutdown sqlite annotation server")
		}
		return nil
	})

	return group.Wait()
}

func NewHandler(options HandlerOptions) http.Handler {
	h := &appHandler{
		db:          options.DB,
		dbInfo:      options.DBInfo,
		startedAt:   options.StartedAt,
		annotations: annotate.NewRepository(options.DB),
		queryDirs:   normalizeDirList(options.QueryDirs),
		presetDirs:  normalizeDirList(options.PresetDirs),
		publicFS:    options.PublicFS,
	}

	mux := http.NewServeMux()
	h.registerHealthRoutes(mux)
	h.registerAPIRoutes(mux)
	h.registerStaticRoutes(mux)

	return withRequestDebugLogging(mux)
}

func (h *appHandler) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if h.db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not-ready",
				"error":  "database is nil",
			})
			return
		}
		if err := h.db.PingContext(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not-ready",
				"error":  err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {
		writeProtoJSON(w, http.StatusOK, &annotationuiv1.InfoResponse{
			Service:   "smailnail-sqlite",
			Version:   "dev",
			StartedAt: formatProtoTime(h.startedAt),
			Database:  databaseInfoToProto(h.dbInfo),
		})
	})
}

func (h *appHandler) registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/annotations", h.handleListAnnotations)
	mux.HandleFunc("GET /api/annotations/{id}", h.handleGetAnnotation)
	mux.HandleFunc("PATCH /api/annotations/{id}/review", h.handleReviewAnnotation)
	mux.HandleFunc("POST /api/annotations/batch-review", h.handleBatchReview)

	mux.HandleFunc("GET /api/annotation-groups", h.handleListGroups)
	mux.HandleFunc("GET /api/annotation-groups/{id}", h.handleGetGroup)

	mux.HandleFunc("GET /api/annotation-logs", h.handleListLogs)
	mux.HandleFunc("GET /api/annotation-logs/{id}", h.handleGetLog)

	mux.HandleFunc("GET /api/annotation-runs", h.handleListRuns)
	mux.HandleFunc("GET /api/annotation-runs/{id}", h.handleGetRun)

	mux.HandleFunc("GET /api/mirror/senders", h.handleListSenders)
	mux.HandleFunc("GET /api/mirror/senders/{email}", h.handleGetSender)
	mux.HandleFunc("GET /api/mirror/senders/{email}/guidelines", h.handleListSenderGuidelines)

	// ── Review feedback ────────────────────────────────────────────
	mux.HandleFunc("GET /api/review-feedback", h.handleListFeedback)
	mux.HandleFunc("GET /api/review-feedback/{id}", h.handleGetFeedback)
	mux.HandleFunc("POST /api/review-feedback", h.handleCreateFeedback)
	mux.HandleFunc("PATCH /api/review-feedback/{id}", h.handleUpdateFeedback)

	// ── Review guidelines ──────────────────────────────────────────────
	mux.HandleFunc("GET /api/review-guidelines", h.handleListGuidelines)
	mux.HandleFunc("GET /api/review-guidelines/{id}", h.handleGetGuideline)
	mux.HandleFunc("GET /api/review-guidelines/{id}/runs", h.handleListGuidelineRuns)
	mux.HandleFunc("POST /api/review-guidelines", h.handleCreateGuideline)
	mux.HandleFunc("PATCH /api/review-guidelines/{id}", h.handleUpdateGuideline)

	// ── Run-guideline links ────────────────────────────────────────────
	mux.HandleFunc("GET /api/annotation-runs/{id}/guidelines", h.handleListRunGuidelines)
	mux.HandleFunc("POST /api/annotation-runs/{id}/guidelines", h.handleLinkRunGuideline)
	mux.HandleFunc("DELETE /api/annotation-runs/{id}/guidelines/{guidelineId}", h.handleUnlinkRunGuideline)

	mux.HandleFunc("POST /api/query/execute", h.handleExecuteQuery)
	mux.HandleFunc("GET /api/query/presets", h.handleGetPresets)
	mux.HandleFunc("GET /api/query/saved", h.handleGetSavedQueries)
	mux.HandleFunc("POST /api/query/saved", h.handleSaveQuery)
}

func (h *appHandler) registerStaticRoutes(mux *http.ServeMux) {
	if h.publicFS == nil {
		return
	}
	if _, err := h.publicFS.Open("index.html"); err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(h.publicFS))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			http.NotFound(w, r)
			return
		}

		if r.URL.Path == "/" {
			http.Redirect(w, r, "/annotations", http.StatusTemporaryRedirect)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if f, err := h.publicFS.Open(path); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/annotations") || strings.HasPrefix(r.URL.Path, "/query") {
			h.serveIndex(w)
			return
		}

		http.NotFound(w, r)
	}))
}

func (h *appHandler) serveIndex(w http.ResponseWriter) {
	index, err := h.publicFS.Open("index.html")
	if err != nil {
		http.Error(w, "index not found", http.StatusNotFound)
		return
	}
	defer func() { _ = index.Close() }()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, index)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeMessageError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"message": message})
}

func writeNotFound(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusNotFound, map[string]string{
		"error":   "not-found",
		"message": message,
	})
}

func requestReviewActor(r *http.Request) string {
	if r == nil {
		return defaultAnnotationUIActor
	}
	for _, header := range []string{"X-Smailnail-User", "X-Forwarded-User", "X-Remote-User", "X-User"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return defaultAnnotationUIActor
}

func parseLimitQuery(r *http.Request, key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return defaultValue, nil
	}
	var value int
	if _, err := fmt.Sscanf(raw, "%d", &value); err != nil {
		return 0, errors.Wrapf(err, "parse %s", key)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return value, nil
}

func normalizeDirList(dirs []string) []string {
	ret := make([]string, 0, len(dirs))
	seen := map[string]struct{}{}
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		ret = append(ret, dir)
	}
	return ret
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	cause := errors.Cause(err)
	if cause != nil && strings.Contains(strings.ToLower(cause.Error()), "no rows") {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

func withRequestDebugLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" || r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("SQLite request started")
		next.ServeHTTP(w, r)
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Msg("SQLite request completed")
	})
}
