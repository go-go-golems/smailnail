package smailnaild

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"io/fs"

	appv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/app/v1"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/accounts"
	hostedauth "github.com/go-go-golems/smailnail/pkg/smailnaild/auth"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/identity"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/rules"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/web"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
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
	WebAuth      hostedauth.WebHandler
	MCPHandler   http.Handler
	PublicFS     fs.FS
}

type HandlerOptions struct {
	DB           *sqlx.DB
	DBInfo       DatabaseInfo
	StartedAt    time.Time
	UserResolver UserResolver
	AccountAPI   AccountAPI
	RuleAPI      RuleAPI
	WebAuth      hostedauth.WebHandler
	MCPHandler   http.Handler
	PublicFS     fs.FS
}

type appHandler struct {
	db           *sqlx.DB
	dbInfo       DatabaseInfo
	startedAt    time.Time
	userResolver UserResolver
	identityRepo *identity.Repository
	accounts     AccountAPI
	rules        RuleAPI
	webAuth      hostedauth.WebHandler
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
			WebAuth:      options.WebAuth,
			MCPHandler:   options.MCPHandler,
			PublicFS:     options.PublicFS,
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
		identityRepo: identity.NewRepository(options.DB),
		accounts:     options.AccountAPI,
		rules:        options.RuleAPI,
		webAuth:      options.WebAuth,
	}
	if h.userResolver == nil {
		h.userResolver = HeaderUserResolver{DefaultUserID: DefaultDevUserID}
	}

	mux := http.NewServeMux()
	h.registerHealthRoutes(mux)
	h.registerAPIRoutes(mux)
	if options.MCPHandler != nil {
		mux.Handle("/.well-known/oauth-protected-resource", options.MCPHandler)
		mux.Handle("/mcp", options.MCPHandler)
		mux.Handle("/mcp/", options.MCPHandler)
	}

	// SPA handler must be registered last so API routes take precedence
	web.RegisterSPA(mux, options.PublicFS, web.SPAOptions{APIPrefix: "/api"})

	return withRequestDebugLogging(mux)
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
		writeProtoJSON(w, http.StatusOK, &appv1.InfoResponse{
			Service:   "smailnaild",
			Version:   "dev",
			StartedAt: formatHostedProtoTime(h.startedAt),
			Database:  databaseInfoToProto(h.dbInfo),
		})
	})
	mux.HandleFunc("GET /api/me", h.handleGetMe)
	if h.webAuth != nil {
		mux.HandleFunc("GET /auth/login", h.webAuth.HandleLogin)
		mux.HandleFunc("GET /auth/callback", h.webAuth.HandleCallback)
		mux.HandleFunc("GET /auth/logout", h.webAuth.HandleLogout)
	}
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

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	items, err := h.accounts.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.ListAccountsResponse{
		Data: accountsListToProto(items),
		Meta: &appv1.ListAccountsMeta{Count: int32(len(items))},
	})
}

func (h *appHandler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if h.accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "accounts-unavailable", "Account API is not configured.", nil)
		return
	}

	req := &appv1.CreateAccountRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	input := createAccountRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	account, err := h.accounts.Create(r.Context(), userID, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusCreated, &appv1.AccountResponse{Data: accountToProto(account)})
}

func (h *appHandler) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	account, err := h.accounts.Get(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.AccountResponse{Data: accountToProto(account)})
}

func (h *appHandler) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	req := &appv1.UpdateAccountRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	input := updateAccountRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	account, err := h.accounts.Update(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.AccountResponse{Data: accountToProto(account)})
}

func (h *appHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	if err := h.accounts.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *appHandler) handleTestAccount(w http.ResponseWriter, r *http.Request) {
	req := &appv1.TestAccountRequest{}
	if !decodeProtoJSONBodyAllowEmpty(w, r, req) {
		return
	}
	input := testAccountRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	result, err := h.accounts.RunTest(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	payload, err := testResultToProto(result)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal-error", err.Error(), nil)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.TestAccountResponse{Data: payload})
}

func (h *appHandler) handleListMailboxes(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	mailboxes, err := h.accounts.ListMailboxes(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.ListMailboxesResponse{
		Data: mailboxesToProto(mailboxes),
		Meta: &appv1.ListMailboxesMeta{Count: int32(len(mailboxes))},
	})
}

func (h *appHandler) handleListMessages(w http.ResponseWriter, r *http.Request) {
	req, err := parseListMessagesRequest(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-query", err.Error(), nil)
		return
	}
	input := listMessagesRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	messages, mailbox, err := h.accounts.ListMessages(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	totalCount := 0
	if len(messages) > 0 {
		totalCount = int(messages[0].TotalCount)
	}

	writeProtoJSON(w, http.StatusOK, &appv1.ListMessagesResponse{
		Data: messagesToProto(messages),
		Meta: &appv1.ListMessagesMeta{
			Mailbox:    mailbox,
			Count:      int32(len(messages)),
			Limit:      int32(effectiveLimit(input.Limit)),
			Offset:     int32(max(input.Offset, 0)),
			TotalCount: int32(totalCount),
		},
	})
}

func (h *appHandler) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	uid64, err := strconv.ParseUint(r.PathValue("uid"), 10, 32)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid-uid", "Message UID must be an unsigned integer.", nil)
		return
	}

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	message, mailbox, err := h.accounts.GetMessage(
		r.Context(),
		userID,
		r.PathValue("id"),
		r.URL.Query().Get("mailbox"),
		uint32(uid64),
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.GetMessageResponse{
		Data: messageViewToProto(*message),
		Meta: &appv1.GetMessageMeta{Mailbox: mailbox},
	})
}

func (h *appHandler) handleListRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	items, err := h.rules.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.ListRulesResponse{
		Data: rulesToProto(items),
		Meta: &appv1.ListRulesMeta{Count: int32(len(items))},
	})
}

func (h *appHandler) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	req := &appv1.CreateRuleRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	input := createRuleRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	record, err := h.rules.Create(r.Context(), userID, input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusCreated, &appv1.RuleResponse{Data: ruleRecordToProto(record)})
}

func (h *appHandler) handleGetRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	record, err := h.rules.Get(r.Context(), userID, r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.RuleResponse{Data: ruleRecordToProto(record)})
}

func (h *appHandler) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	req := &appv1.UpdateRuleRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	input := updateRuleRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	record, err := h.rules.Update(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.RuleResponse{Data: ruleRecordToProto(record)})
}

func (h *appHandler) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	if err := h.rules.Delete(r.Context(), userID, r.PathValue("id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *appHandler) handleDryRunRule(w http.ResponseWriter, r *http.Request) {
	req := &appv1.DryRunRuleRequest{}
	if !decodeProtoJSONBodyAllowEmpty(w, r, req) {
		return
	}
	input := dryRunRuleRequestToDomain(req)

	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	result, err := h.rules.DryRun(r.Context(), userID, r.PathValue("id"), input)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	payload, err := dryRunResultToProto(result)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal-error", err.Error(), nil)
		return
	}
	writeProtoJSON(w, http.StatusOK, &appv1.DryRunRuleResponse{Data: payload})
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

func (h *appHandler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireUserID(w, r)
	if !ok {
		return
	}

	if h.identityRepo == nil {
		writeProtoJSON(w, http.StatusOK, &appv1.CurrentUserResponse{Data: &appv1.CurrentUser{Id: userID}})
		return
	}

	user, err := h.identityRepo.GetUserByID(r.Context(), userID)
	if errors.Is(err, identity.ErrNotFound) {
		writeProtoJSON(w, http.StatusOK, &appv1.CurrentUserResponse{Data: &appv1.CurrentUser{Id: userID}})
		return
	}
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeProtoJSON(w, http.StatusOK, &appv1.CurrentUserResponse{Data: currentUserToProto(user)})
}

func (h *appHandler) userID(r *http.Request) (string, error) {
	return h.userResolver.ResolveUserID(r)
}

func (h *appHandler) requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, err := h.userID(r)
	if err != nil || strings.TrimSpace(userID) == "" {
		log.Debug().
			Err(err).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Hosted request is unauthenticated")
		writeAPIError(w, http.StatusUnauthorized, "unauthenticated", "Authentication is required.", nil)
		return "", false
	}
	log.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("user_id", userID).
		Msg("Hosted request resolved authenticated user")
	return userID, true
}

type debugResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *debugResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *debugResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func withRequestDebugLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldSkipHostedRequestDebugLogging(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		wrapped := &debugResponseWriter{ResponseWriter: w}
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Bool("has_cookie", len(r.Cookies()) > 0).
			Msg("Hosted request started")
		next.ServeHTTP(wrapped, r)
		status := wrapped.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Int("bytes", wrapped.bytes).
			Dur("duration", time.Since(start)).
			Msg("Hosted request completed")
	})
}

func shouldSkipHostedRequestDebugLogging(r *http.Request) bool {
	return r.Method == http.MethodGet && r.URL.Path == "/readyz"
}

func parseListMessagesRequest(r *http.Request) (*appv1.ListMessagesRequest, error) {
	query := r.URL.Query()
	limit32, err := parseOptionalInt32(query.Get("limit"), 20)
	if err != nil {
		return nil, fmt.Errorf("limit must be a valid integer")
	}
	offset32, err := parseOptionalInt32(query.Get("offset"), 0)
	if err != nil {
		return nil, fmt.Errorf("offset must be a valid integer")
	}
	queryText := strings.TrimSpace(query.Get("query"))
	contentType := strings.TrimSpace(query.Get("content_type"))
	unreadOnly := parseBoolQuery(query.Get("unread_only"))
	includeContent := parseBoolQuery(query.Get("include_content"))

	return &appv1.ListMessagesRequest{
		Mailbox:        strings.TrimSpace(query.Get("mailbox")),
		Limit:          &limit32,
		Offset:         &offset32,
		Query:          optionalStringPointer(queryText),
		UnreadOnly:     &unreadOnly,
		IncludeContent: &includeContent,
		ContentType:    optionalStringPointer(contentType),
	}, nil
}

func parseOptionalInt32(raw string, defaultValue int32) (int32, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("value must be non-negative")
	}
	if value > math.MaxInt32 {
		value = math.MaxInt32
	}
	return int32(value), nil
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

func writeAPIError(w http.ResponseWriter, statusCode int, code, message string, details map[string]any) {
	writeProtoAPIError(w, statusCode, code, message, details)
}
