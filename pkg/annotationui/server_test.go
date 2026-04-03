package annotationui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	"github.com/go-go-golems/smailnail/pkg/mirror"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type annotationUITestFixture struct {
	AnnotationOneID   string
	AnnotationTwoID   string
	AnnotationThreeID string
	GroupID           string
	LogID             string
}

func TestHandlerServesAnnotationAPIAndSPA(t *testing.T) {
	db, queryDir, presetDir := openAnnotationUITestDB(t)
	fixture := seedAnnotationUITestData(t, db)

	_ = os.WriteFile(filepath.Join(queryDir, "initial.sql"), []byte("-- Existing saved query\nSELECT 1 AS value;\n"), 0o644)
	if err := os.MkdirAll(filepath.Join(presetDir, "custom"), 0o755); err != nil {
		t.Fatalf("MkdirAll(presetDir) error = %v", err)
	}
	_ = os.WriteFile(filepath.Join(presetDir, "custom", "sender-count.sql"), []byte("-- Count senders\nSELECT COUNT(*) AS count FROM senders;\n"), 0o644)

	handler := NewHandler(HandlerOptions{
		DB:         db,
		DBInfo:     DatabaseInfo{Driver: "sqlite3", Target: "test.sqlite", Mode: "mirror"},
		StartedAt:  time.Date(2026, 4, 3, 14, 0, 0, 0, time.UTC),
		QueryDirs:  []string{queryDir},
		PresetDirs: []string{presetDir},
		PublicFS: fstest.MapFS{
			"index.html":           {Data: []byte("<!doctype html><html><body><div id=\"root\"></div></body></html>")},
			"assets/app.js":        {Data: []byte("console.log('app');")},
			"assets/app.css":       {Data: []byte("body { color: black; }")},
			"favicon.ico":          {Data: []byte("ico")},
			"mockServiceWorker.js": {Data: []byte("// worker")},
		},
	})

	t.Run("root redirects to annotations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if location := rec.Header().Get("Location"); location != "/annotations" {
			t.Fatalf("location = %q", location)
		}
	})

	t.Run("annotations route serves spa index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/annotations", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "<div id=\"root\"></div>") {
			t.Fatalf("unexpected body: %s", rec.Body.String())
		}
	})

	t.Run("list annotations supports run filter", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/annotations?agentRunId=run-42", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload []annotate.Annotation
		decodeJSONResponse(t, rec, &payload)
		if len(payload) != 2 {
			t.Fatalf("expected 2 annotations, got %d", len(payload))
		}
	})

	t.Run("review annotation updates state", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodPatch, "/api/annotations/"+fixture.AnnotationOneID+"/review", `{"reviewState":"reviewed"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload annotate.Annotation
		decodeJSONResponse(t, rec, &payload)
		if payload.ReviewState != annotate.ReviewStateReviewed {
			t.Fatalf("reviewState = %q", payload.ReviewState)
		}
	})

	t.Run("batch review updates multiple annotations", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodPost, "/api/annotations/batch-review", `{"ids":["`+fixture.AnnotationOneID+`","`+fixture.AnnotationTwoID+`"],"reviewState":"dismissed"}`)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		check := performRequest(t, handler, http.MethodGet, "/api/annotations?agentRunId=run-42", "")
		var payload []annotate.Annotation
		decodeJSONResponse(t, check, &payload)
		foundDismissed := 0
		for _, annotation := range payload {
			if annotation.ReviewState == annotate.ReviewStateDismissed {
				foundDismissed++
			}
		}
		if foundDismissed != 2 {
			t.Fatalf("expected 2 dismissed annotations, got %d", foundDismissed)
		}
	})

	t.Run("group detail includes members", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/annotation-groups/"+fixture.GroupID, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload annotate.GroupDetail
		decodeJSONResponse(t, rec, &payload)
		if len(payload.Members) != 2 {
			t.Fatalf("expected 2 members, got %d", len(payload.Members))
		}
	})

	t.Run("logs list supports run filter", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/annotation-logs?agentRunId=run-42", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload []annotate.AnnotationLog
		decodeJSONResponse(t, rec, &payload)
		if len(payload) != 1 {
			t.Fatalf("expected 1 log, got %d", len(payload))
		}
	})

	t.Run("run detail aggregates annotations logs and groups", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/annotation-runs/run-42", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload annotate.AgentRunDetail
		decodeJSONResponse(t, rec, &payload)
		if payload.RunID != "run-42" {
			t.Fatalf("runID = %q", payload.RunID)
		}
		if len(payload.Annotations) != 2 || len(payload.Logs) != 1 || len(payload.Groups) != 1 {
			t.Fatalf("unexpected detail sizes: annotations=%d logs=%d groups=%d", len(payload.Annotations), len(payload.Logs), len(payload.Groups))
		}
	})

	t.Run("list senders returns annotation counts and tags", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/mirror/senders?hasAnnotations=true&tag=newsletter", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload []SenderRow
		decodeJSONResponse(t, rec, &payload)
		if len(payload) != 1 {
			t.Fatalf("expected 1 sender, got %d", len(payload))
		}
		if payload[0].Email != "news@example.com" {
			t.Fatalf("email = %q", payload[0].Email)
		}
		if len(payload[0].Tags) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(payload[0].Tags))
		}
	})

	t.Run("sender detail includes annotations logs and messages", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/mirror/senders/news%40example.com", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload SenderDetail
		decodeJSONResponse(t, rec, &payload)
		if len(payload.Annotations) != 2 {
			t.Fatalf("expected 2 annotations, got %d", len(payload.Annotations))
		}
		if len(payload.Logs) != 1 {
			t.Fatalf("expected 1 log, got %d", len(payload.Logs))
		}
		if len(payload.RecentMessages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(payload.RecentMessages))
		}
	})

	t.Run("execute query returns rows", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodPost, "/api/query/execute", `{"sql":"SELECT tag, COUNT(*) AS count FROM annotations GROUP BY tag ORDER BY tag ASC"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload QueryResult
		decodeJSONResponse(t, rec, &payload)
		if len(payload.Columns) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(payload.Columns))
		}
		if payload.RowCount == 0 {
			t.Fatalf("expected rows, got 0")
		}
	})

	t.Run("execute query rejects writes", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodPost, "/api/query/execute", `{"sql":"UPDATE annotations SET tag = 'x'"}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "read-only") && !strings.Contains(rec.Body.String(), "allowed") {
			t.Fatalf("unexpected body: %s", rec.Body.String())
		}
	})

	t.Run("presets merge embedded and external queries", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/query/presets", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}

		var payload []SavedQuery
		decodeJSONResponse(t, rec, &payload)
		if len(payload) < 5 {
			t.Fatalf("expected embedded plus external presets, got %d", len(payload))
		}
		if !containsQueryNamed(payload, "sender-count") {
			t.Fatalf("expected external preset sender-count")
		}
	})

	t.Run("saved queries list and create", func(t *testing.T) {
		rec := performRequest(t, handler, http.MethodGet, "/api/query/saved", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		var initial []SavedQuery
		decodeJSONResponse(t, rec, &initial)
		if len(initial) != 1 {
			t.Fatalf("expected 1 initial saved query, got %d", len(initial))
		}

		create := performRequest(t, handler, http.MethodPost, "/api/query/saved", `{"name":"my-senders","folder":"custom","description":"Custom sender analysis","sql":"SELECT email FROM senders ORDER BY email"}`)
		if create.Code != http.StatusCreated {
			t.Fatalf("status = %d body=%s", create.Code, create.Body.String())
		}

		var created SavedQuery
		decodeJSONResponse(t, create, &created)
		if created.Name != "my-senders" || created.Folder != "custom" {
			t.Fatalf("unexpected created query: %#v", created)
		}

		after := performRequest(t, handler, http.MethodGet, "/api/query/saved", "")
		var saved []SavedQuery
		decodeJSONResponse(t, after, &saved)
		if len(saved) != 2 {
			t.Fatalf("expected 2 saved queries, got %d", len(saved))
		}
	})
}

func openAnnotationUITestDB(t *testing.T) (*sqlx.DB, string, string) {
	t.Helper()

	root := t.TempDir()
	path := filepath.Join(root, "mirror.sqlite")
	store, err := mirror.OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	if _, err := store.Bootstrap(context.Background(), filepath.Join(root, "mirror-root")); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	db := sqlx.MustOpen("sqlite3", path)
	t.Cleanup(func() { _ = db.Close() })

	queryDir := filepath.Join(root, "queries")
	presetDir := filepath.Join(root, "presets")
	if err := os.MkdirAll(queryDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(queryDir) error = %v", err)
	}
	if err := os.MkdirAll(presetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(presetDir) error = %v", err)
	}

	return db, queryDir, presetDir
}

func seedAnnotationUITestData(t *testing.T, db *sqlx.DB) annotationUITestFixture {
	t.Helper()

	repo := annotate.NewRepository(db)

	ann1, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:   "sender",
		TargetID:     "news@example.com",
		Tag:          "newsletter",
		NoteMarkdown: "Daily newsletter.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-agent-v1",
		AgentRunID:   "run-42",
		CreatedBy:    "system",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation(ann-1) error = %v", err)
	}
	ann2, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:   "sender",
		TargetID:     "news@example.com",
		Tag:          "important",
		NoteMarkdown: "Actually useful.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-agent-v1",
		AgentRunID:   "run-42",
		CreatedBy:    "system",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation(ann-2) error = %v", err)
	}
	ann3, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:   "sender",
		TargetID:     "other@example.com",
		Tag:          "notification",
		NoteMarkdown: "Elsewhere.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-agent-v2",
		AgentRunID:   "run-99",
		CreatedBy:    "system",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation(ann-3) error = %v", err)
	}
	if _, err := repo.UpdateAnnotationReviewState(context.Background(), ann1.ID, annotate.ReviewStateReviewed); err != nil {
		t.Fatalf("UpdateAnnotationReviewState() error = %v", err)
	}

	group, err := repo.CreateGroup(context.Background(), annotate.CreateGroupInput{
		Name:        "News senders",
		Description: "Grouped by agent",
		SourceKind:  annotate.SourceKindAgent,
		SourceLabel: "triage-agent-v1",
		AgentRunID:  "run-42",
		CreatedBy:   "system",
	})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if err := repo.AddGroupMember(context.Background(), annotate.AddGroupMemberInput{
		GroupID:    group.ID,
		TargetType: "sender",
		TargetID:   "news@example.com",
	}); err != nil {
		t.Fatalf("AddGroupMember(1) error = %v", err)
	}
	if err := repo.AddGroupMember(context.Background(), annotate.AddGroupMemberInput{
		GroupID:    group.ID,
		TargetType: "sender",
		TargetID:   "alerts@example.com",
	}); err != nil {
		t.Fatalf("AddGroupMember(2) error = %v", err)
	}

	logEntry, err := repo.CreateLog(context.Background(), annotate.CreateLogInput{
		Title:        "Run summary",
		BodyMarkdown: "Detected newsletters.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-agent-v1",
		AgentRunID:   "run-42",
		CreatedBy:    "system",
	})
	if err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}
	if err := repo.LinkLogTarget(context.Background(), annotate.LinkLogTargetInput{
		LogID:      logEntry.ID,
		TargetType: "sender",
		TargetID:   "news@example.com",
	}); err != nil {
		t.Fatalf("LinkLogTarget() error = %v", err)
	}

	mustExec(t, db, `INSERT INTO senders (
		email, display_name, domain, msg_count, first_seen_date, last_seen_date, has_list_unsubscribe
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"news@example.com", "News Daily", "example.com", 47, "2025-01-15T00:00:00Z", "2026-04-01T08:00:00Z", true)
	mustExec(t, db, `INSERT INTO senders (
		email, display_name, domain, msg_count, first_seen_date, last_seen_date, has_list_unsubscribe
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"other@example.com", "Other Sender", "example.com", 12, "2025-02-01T00:00:00Z", "2026-04-02T08:00:00Z", false)

	mustExec(t, db, `INSERT INTO messages (
		account_key, mailbox_name, uidvalidity, uid, message_id, internal_date, sent_date, subject,
		from_summary, to_summary, cc_summary, size_bytes, flags_json, headers_json, parts_json,
		body_text, body_html, search_text, raw_path, raw_sha256, has_attachments, remote_deleted,
		sender_email, sender_domain
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"acct", "INBOX", 1, 1001, "msg-1", "2026-04-01T08:00:00Z", "2026-04-01T08:00:00Z", "Daily digest",
		"News Daily <news@example.com>", "user@example.com", "", 45320, "[]", "{}", "[]",
		"body", "", "body", "/tmp/raw-1.eml", "sha1", false, false, "news@example.com", "example.com")
	mustExec(t, db, `INSERT INTO messages (
		account_key, mailbox_name, uidvalidity, uid, message_id, internal_date, sent_date, subject,
		from_summary, to_summary, cc_summary, size_bytes, flags_json, headers_json, parts_json,
		body_text, body_html, search_text, raw_path, raw_sha256, has_attachments, remote_deleted,
		sender_email, sender_domain
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"acct", "INBOX", 1, 1002, "msg-2", "2026-03-31T08:00:00Z", "2026-03-31T08:00:00Z", "Previous digest",
		"News Daily <news@example.com>", "user@example.com", "", 35320, "[]", "{}", "[]",
		"body", "", "body", "/tmp/raw-2.eml", "sha2", false, false, "news@example.com", "example.com")

	return annotationUITestFixture{
		AnnotationOneID:   ann1.ID,
		AnnotationTwoID:   ann2.ID,
		AnnotationThreeID: ann3.ID,
		GroupID:           group.ID,
		LogID:             logEntry.ID,
	}
}

func mustExec(t *testing.T, db *sqlx.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("Exec(%s) error = %v", query, err)
	}
}

func performRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(dest); err != nil {
		t.Fatalf("Decode() error = %v body=%s", err, rec.Body.String())
	}
}

func containsQueryNamed(queries []SavedQuery, name string) bool {
	for _, query := range queries {
		if query.Name == name {
			return true
		}
	}
	return false
}
