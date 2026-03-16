package smailnaild

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/jmoiron/sqlx"
)

type AccountAPI interface {
	List(ctx context.Context, userID string) ([]accounts.AccountListItem, error)
	Create(ctx context.Context, userID string, input accounts.CreateInput) (*accounts.Account, error)
	Get(ctx context.Context, userID, accountID string) (*accounts.Account, error)
	Update(ctx context.Context, userID, accountID string, input accounts.UpdateInput) (*accounts.Account, error)
	Delete(ctx context.Context, userID, accountID string) error
	RunTest(ctx context.Context, userID, accountID string, input accounts.TestInput) (*accounts.TestResult, error)
	ListMailboxes(ctx context.Context, userID, accountID string) ([]accounts.MailboxInfo, error)
	ListMessages(ctx context.Context, userID, accountID string, input accounts.ListMessagesInput) ([]accounts.MessageView, string, error)
	GetMessage(ctx context.Context, userID, accountID, mailbox string, uid uint32) (*accounts.MessageView, string, error)
}

type RuleAPI interface {
	List(ctx context.Context, userID string) ([]rules.RuleRecord, error)
	Create(ctx context.Context, userID string, input rules.CreateInput) (*rules.RuleRecord, error)
	Get(ctx context.Context, userID, ruleID string) (*rules.RuleRecord, error)
	Update(ctx context.Context, userID, ruleID string, input rules.UpdateInput) (*rules.RuleRecord, error)
	Delete(ctx context.Context, userID, ruleID string) error
	DryRun(ctx context.Context, userID, ruleID string, input rules.DryRunInput) (*rules.DryRunResult, error)
}

type ServerOptions struct {
	Host         string
	Port         int
	DB           *sqlx.DB
	DBInfo       DatabaseInfo
	UserResolver UserResolver
	AccountAPI   AccountAPI
	RuleAPI      RuleAPI
}

type HandlerOptions struct {
	DB           *sqlx.DB
	DBInfo       DatabaseInfo
	StartedAt    time.Time
	UserResolver UserResolver
	AccountAPI   AccountAPI
	RuleAPI      RuleAPI
}

type infoResponse struct {
	Service   string       `json:"service"`
	Version   string       `json:"version"`
	StartedAt time.Time    `json:"startedAt"`
	Database  DatabaseInfo `json:"database"`
}

type appHandler struct {
	db           *sqlx.DB
	dbInfo       DatabaseInfo
	startedAt    time.Time
	userResolver UserResolver
	accounts     AccountAPI
	rules        RuleAPI
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func NewHTTPServer(options ServerOptions) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf("%s:%d", options.Host, options.Port),
		Handler: NewHandler(HandlerOptions{
			DB:           options.DB,
			DBInfo:       options.DBInfo,
			StartedAt:    time.Now().UTC(),
			UserResolver: options.UserResolver,
			AccountAPI:   options.AccountAPI,
			RuleAPI:      options.RuleAPI,
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func NewHandler(options HandlerOptions) http.Handler {
	h := &appHandler{
		db:           options.DB,
		dbInfo:       options.DBInfo,
		startedAt:    options.StartedAt,
		userResolver: options.UserResolver,
		accounts:     options.AccountAPI,
		rules:        options.RuleAPI,
	}
	if h.userResolver == nil {
		h.userResolver = HeaderUserResolver{}
	}

	mux := http.NewServeMux()
	h.registerHealthRoutes(mux)
	h.registerAPIRoutes(mux)
	return mux
}

func (h *appHandler) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := PingDatabase(r.Context(), h.db); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not-ready",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, infoResponse{
			Service:   "smailnaild",
			Version:   "dev",
			StartedAt: h.startedAt,
			Database:  h.dbInfo,
		})
	})
}

func (h *appHandler) registerAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/accounts", h.handleListAccounts)
	mux.HandleFunc("POST /api/accounts", h.handleCreateAccount)
	mux.HandleFunc("GET /api/accounts/{id}", h.handleGetAccount)
	mux.HandleFunc("PATCH /api/accounts/{id}", h.handleUpdateAccount)
	mux.HandleFunc("DELETE /api/accounts/{id}", h.handleDeleteAccount)
	mux.HandleFunc("POST /api/accounts/{id}/test", h.handleTestAccount)
	mux.HandleFunc("GET /api/accounts/{id}/mailboxes", h.handleListMailboxes)
	mux.HandleFunc("GET /api/accounts/{id}/messages", h.handleListMessages)
	mux.HandleFunc("GET /api/accounts/{id}/messages/{uid}", h.handleGetMessage)

	mux.HandleFunc("GET /api/rules", h.handleListRules)
	mux.HandleFunc("POST /api/rules", h.handleCreateRule)
	mux.HandleFunc("GET /api/rules/{id}", h.handleGetRule)
	mux.HandleFunc("PATCH /api/rules/{id}", h.handleUpdateRule)
	mux.HandleFunc("DELETE /api/rules/{id}", h.handleDeleteRule)
	mux.HandleFunc("POST /api/rules/{id}/dry-run", h.handleDryRunRule)
}

func (h *appHandler) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	if h.accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "accounts-unavailable", "Account API is not configured.", nil)
		return
	}

	items, err := h.accounts.List(r.Context(), h.userID(r))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, items, map[string]any{"count": len(items)})
}

func (h *appHandler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if h.accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "accounts-unavailable", "Account API is not configured.", nil)
		return
	}

	var input accounts.CreateInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	account, err := h.accounts.Create(r.Context(), h.userID(r), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusCreated, account, nil)
}

func (h *appHandler) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	account, err := h.accounts.Get(r.Context(), h.userID(r), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, account, nil)
}

func (h *appHandler) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	var input accounts.UpdateInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	account, err := h.accounts.Update(r.Context(), h.userID(r), r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, account, nil)
}

func (h *appHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	if err := h.accounts.Delete(r.Context(), h.userID(r), r.PathValue("id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *appHandler) handleTestAccount(w http.ResponseWriter, r *http.Request) {
	var input accounts.TestInput
	if !decodeJSONBodyAllowEmpty(w, r, &input) {
		return
	}

	result, err := h.accounts.RunTest(r.Context(), h.userID(r), r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, result, nil)
}

func (h *appHandler) handleListMailboxes(w http.ResponseWriter, r *http.Request) {
	mailboxes, err := h.accounts.ListMailboxes(r.Context(), h.userID(r), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, mailboxes, map[string]any{"count": len(mailboxes)})
}

func (h *appHandler) handleListMessages(w http.ResponseWriter, r *http.Request) {
	input, err := parseListMessagesInput(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-query", err.Error(), nil)
		return
	}

	messages, mailbox, err := h.accounts.ListMessages(r.Context(), h.userID(r), r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	totalCount := 0
	if len(messages) > 0 {
		totalCount = int(messages[0].TotalCount)
	}

	writeDataJSON(w, http.StatusOK, messages, map[string]any{
		"mailbox":    mailbox,
		"count":      len(messages),
		"limit":      effectiveLimit(input.Limit),
		"offset":     max(input.Offset, 0),
		"totalCount": totalCount,
	})
}

func (h *appHandler) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	uid64, err := strconv.ParseUint(r.PathValue("uid"), 10, 32)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-uid", "Message UID must be an unsigned integer.", nil)
		return
	}

	message, mailbox, err := h.accounts.GetMessage(
		r.Context(),
		h.userID(r),
		r.PathValue("id"),
		r.URL.Query().Get("mailbox"),
		uint32(uid64),
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, message, map[string]any{"mailbox": mailbox})
}

func (h *appHandler) handleListRules(w http.ResponseWriter, r *http.Request) {
	items, err := h.rules.List(r.Context(), h.userID(r))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, items, map[string]any{"count": len(items)})
}

func (h *appHandler) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	var input rules.CreateInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	record, err := h.rules.Create(r.Context(), h.userID(r), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusCreated, record, nil)
}

func (h *appHandler) handleGetRule(w http.ResponseWriter, r *http.Request) {
	record, err := h.rules.Get(r.Context(), h.userID(r), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, record, nil)
}

func (h *appHandler) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	var input rules.UpdateInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	record, err := h.rules.Update(r.Context(), h.userID(r), r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, record, nil)
}

func (h *appHandler) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	if err := h.rules.Delete(r.Context(), h.userID(r), r.PathValue("id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *appHandler) handleDryRunRule(w http.ResponseWriter, r *http.Request) {
	var input rules.DryRunInput
	if !decodeJSONBodyAllowEmpty(w, r, &input) {
		return
	}

	result, err := h.rules.DryRun(r.Context(), h.userID(r), r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeDataJSON(w, http.StatusOK, result, nil)
}

func (h *appHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, accounts.ErrNotFound), errors.Is(err, rules.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not-found", err.Error(), nil)
	case errors.Is(err, accounts.ErrValidation), errors.Is(err, rules.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "validation-error", err.Error(), nil)
	case errors.Is(err, accounts.ErrIMAP):
		writeAPIError(w, http.StatusBadGateway, "imap-error", err.Error(), nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, "internal-error", err.Error(), nil)
	}
}

func (h *appHandler) userID(r *http.Request) string {
	return h.userResolver.ResolveUserID(r)
}

func parseListMessagesInput(r *http.Request) (accounts.ListMessagesInput, error) {
	query := r.URL.Query()
	limit, err := parseOptionalInt(query.Get("limit"), 20)
	if err != nil {
		return accounts.ListMessagesInput{}, fmt.Errorf("limit must be a valid integer")
	}
	offset, err := parseOptionalInt(query.Get("offset"), 0)
	if err != nil {
		return accounts.ListMessagesInput{}, fmt.Errorf("offset must be a valid integer")
	}

	return accounts.ListMessagesInput{
		Mailbox:        strings.TrimSpace(query.Get("mailbox")),
		Limit:          limit,
		Offset:         offset,
		Query:          strings.TrimSpace(query.Get("query")),
		UnreadOnly:     parseBoolQuery(query.Get("unread_only")),
		IncludeContent: parseBoolQuery(query.Get("include_content")),
		ContentType:    strings.TrimSpace(query.Get("content_type")),
	}, nil
}

func parseOptionalInt(raw string, defaultValue int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func parseBoolQuery(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func effectiveLimit(limit int) int {
	switch {
	case limit <= 0:
		return 20
	case limit > 100:
		return 100
	default:
		return limit
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Body == nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", "Request body is required.", nil)
		return false
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-body", fmt.Sprintf("Invalid JSON body: %v", err), nil)
		return false
	}
	return true
}

func decodeJSONBodyAllowEmpty(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Body == nil {
		return true
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, context.Canceled) {
			writeAPIError(w, http.StatusBadRequest, "invalid-body", "Request body could not be read.", nil)
			return false
		}
		if errors.Is(err, io.EOF) {
			return true
		}
		writeAPIError(w, http.StatusBadRequest, "invalid-body", fmt.Sprintf("Invalid JSON body: %v", err), nil)
		return false
	}
	return true
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

func writeDataJSON(w http.ResponseWriter, statusCode int, data any, meta map[string]any) {
	payload := map[string]any{"data": data}
	if len(meta) > 0 {
		payload["meta"] = meta
	}
	writeJSON(w, statusCode, payload)
}

func writeAPIError(w http.ResponseWriter, statusCode int, code, message string, details map[string]any) {
	writeJSON(w, statusCode, errorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
