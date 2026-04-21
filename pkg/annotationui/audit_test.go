package annotationui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandlerPopulatesAuditMetadata(t *testing.T) {
	db, queryDir, presetDir := openAnnotationUITestDB(t)
	fixture := seedAnnotationUITestData(t, db)

	handler := NewHandler(HandlerOptions{
		DB:         db,
		DBInfo:     DatabaseInfo{Driver: "sqlite3", Target: "test.sqlite", Mode: "mirror"},
		StartedAt:  time.Date(2026, 4, 3, 14, 0, 0, 0, time.UTC),
		QueryDirs:  []string{queryDir},
		PresetDirs: []string{presetDir},
		PublicFS: fstest.MapFS{
			"index.html": {Data: []byte("<!doctype html><html><body><div id=\"root\"></div></body></html>")},
		},
	})

	actor := "reviewer-1"

	feedbackCreateRec := performProtoRequestAsUser(t, handler, http.MethodPost, "/api/review-feedback", &annotationuiv1.CreateFeedbackRequest{
		ScopeKind:    annotate.FeedbackScopeRun,
		AgentRunId:   "run-42",
		FeedbackKind: annotate.FeedbackKindComment,
		Title:        "Audited feedback",
		BodyMarkdown: "Created with actor header.",
	}, actor)
	if feedbackCreateRec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", feedbackCreateRec.Code, feedbackCreateRec.Body.String())
	}
	var createdFeedback annotationuiv1.ReviewFeedback
	decodeProtoJSONResponse(t, feedbackCreateRec, &createdFeedback)
	if createdFeedback.CreatedBy != actor {
		t.Fatalf("feedback createdBy = %q", createdFeedback.CreatedBy)
	}

	guidelineCreateRec := performProtoRequestAsUser(t, handler, http.MethodPost, "/api/review-guidelines", &annotationuiv1.CreateGuidelineRequest{
		Slug:         "audit-guideline",
		Title:        "Audit guideline",
		ScopeKind:    annotate.GuidelineScopeWorkflow,
		BodyMarkdown: "Created with actor header.",
	}, actor)
	if guidelineCreateRec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", guidelineCreateRec.Code, guidelineCreateRec.Body.String())
	}
	var createdGuideline annotationuiv1.ReviewGuideline
	decodeProtoJSONResponse(t, guidelineCreateRec, &createdGuideline)
	if createdGuideline.CreatedBy != actor {
		t.Fatalf("guideline createdBy = %q", createdGuideline.CreatedBy)
	}

	reviewRec := performProtoRequestAsUser(t, handler, http.MethodPatch, "/api/annotations/"+fixture.AnnotationOneID+"/review", &annotationuiv1.ReviewAnnotationRequest{
		ReviewState: annotate.ReviewStateDismissed,
		Comment: &annotationuiv1.ReviewComment{
			FeedbackKind: annotate.FeedbackKindRejectRequest,
			Title:        "Needs review",
			BodyMarkdown: "Dismissed by actor header.",
		},
	}, actor)
	if reviewRec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", reviewRec.Code, reviewRec.Body.String())
	}

	feedbackListRec := performRequest(t, handler, http.MethodGet, "/api/review-feedback?agentRunId=run-42&feedbackKind=reject_request", "")
	if feedbackListRec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", feedbackListRec.Code, feedbackListRec.Body.String())
	}
	var feedbackList annotationuiv1.ReviewFeedbackListResponse
	decodeProtoJSONResponse(t, feedbackListRec, &feedbackList)
	foundActorFeedback := false
	for _, item := range feedbackList.Items {
		if item.Title == "Needs review" {
			foundActorFeedback = true
			if item.CreatedBy != actor {
				t.Fatalf("artifact feedback createdBy = %q", item.CreatedBy)
			}
		}
	}
	if !foundActorFeedback {
		t.Fatalf("expected review artifact feedback from actor %q", actor)
	}

	linkRec := performProtoRequestAsUser(t, handler, http.MethodPost, "/api/annotation-runs/run-42/guidelines", &annotationuiv1.LinkRunGuidelineRequest{GuidelineId: createdGuideline.Id}, actor)
	if linkRec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", linkRec.Code, linkRec.Body.String())
	}

	var linkedBy string
	query := db.Rebind(`SELECT linked_by FROM run_guideline_links WHERE agent_run_id = ? AND guideline_id = ?`)
	if err := db.Get(&linkedBy, query, "run-42", createdGuideline.Id); err != nil {
		t.Fatalf("query linked_by error = %v", err)
	}
	if linkedBy != actor {
		t.Fatalf("linkedBy = %q", linkedBy)
	}
}

func performRequestWithHeaders(t *testing.T, handler http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
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
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func performProtoRequestAsUser(t *testing.T, handler http.Handler, method, path string, msg proto.Message, user string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := protojson.Marshal(msg)
	if err != nil {
		t.Fatalf("protojson.Marshal() error = %v", err)
	}
	return performRequestWithHeaders(t, handler, method, path, string(body), map[string]string{
		"X-Smailnail-User": user,
	})
}
