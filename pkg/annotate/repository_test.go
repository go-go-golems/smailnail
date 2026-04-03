package annotate_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	"github.com/go-go-golems/smailnail/pkg/mirror"
	"github.com/jmoiron/sqlx"
)

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "mirror.sqlite")
	store, err := mirror.OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	root := filepath.Join(t.TempDir(), "mirror-root")
	if _, err := store.Bootstrap(context.Background(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	db := sqlx.MustOpen("sqlite3", path)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestRepositoryAnnotationCRUD(t *testing.T) {
	db := openTestDB(t)
	repo := annotate.NewRepository(db)

	created, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:   "sender",
		TargetID:     "notifications@github.com",
		Tag:          "important",
		NoteMarkdown: "Still noisy, but do not ignore.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-pass-1",
		AgentRunID:   "run-123",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation() error = %v", err)
	}
	if created.ReviewState != annotate.ReviewStateToReview {
		t.Fatalf("expected default review state %q, got %q", annotate.ReviewStateToReview, created.ReviewState)
	}

	listed, err := repo.ListAnnotations(context.Background(), annotate.ListAnnotationsFilter{
		TargetType: "sender",
		TargetID:   "notifications@github.com",
	})
	if err != nil {
		t.Fatalf("ListAnnotations() error = %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(listed))
	}

	updated, err := repo.UpdateAnnotationReviewState(context.Background(), created.ID, annotate.ReviewStateReviewed)
	if err != nil {
		t.Fatalf("UpdateAnnotationReviewState() error = %v", err)
	}
	if updated.ReviewState != annotate.ReviewStateReviewed {
		t.Fatalf("expected review state %q, got %q", annotate.ReviewStateReviewed, updated.ReviewState)
	}
}

func TestRepositoryGroupsAndLogs(t *testing.T) {
	db := openTestDB(t)
	repo := annotate.NewRepository(db)

	group, err := repo.CreateGroup(context.Background(), annotate.CreateGroupInput{
		Name:        "Possible newsletters",
		Description: "Review these senders before muting them.",
		SourceKind:  annotate.SourceKindAgent,
		AgentRunID:  "run-42",
	})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if err := repo.AddGroupMember(context.Background(), annotate.AddGroupMemberInput{
		GroupID:    group.ID,
		TargetType: "sender",
		TargetID:   "newsletter@example.com",
	}); err != nil {
		t.Fatalf("AddGroupMember() error = %v", err)
	}

	members, err := repo.ListGroupMembers(context.Background(), group.ID)
	if err != nil {
		t.Fatalf("ListGroupMembers() error = %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 group member, got %d", len(members))
	}

	logEntry, err := repo.CreateLog(context.Background(), annotate.CreateLogInput{
		Title:        "Initial review pass",
		BodyMarkdown: "Grouped likely newsletters based on list-unsubscribe and high volume.",
		SourceKind:   annotate.SourceKindAgent,
		AgentRunID:   "run-42",
	})
	if err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}
	if err := repo.LinkLogTarget(context.Background(), annotate.LinkLogTargetInput{
		LogID:      logEntry.ID,
		TargetType: "sender",
		TargetID:   "newsletter@example.com",
	}); err != nil {
		t.Fatalf("LinkLogTarget() error = %v", err)
	}

	targets, err := repo.ListLogTargets(context.Background(), logEntry.ID)
	if err != nil {
		t.Fatalf("ListLogTargets() error = %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 log target, got %d", len(targets))
	}
}

func TestRepositoryListAnnotationsFiltersByAgentRunID(t *testing.T) {
	db := openTestDB(t)
	repo := annotate.NewRepository(db)

	_, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType: "sender",
		TargetID:   "one@example.com",
		Tag:        "newsletter",
		SourceKind: annotate.SourceKindAgent,
		AgentRunID: "run-1",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation(run-1) error = %v", err)
	}

	_, err = repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType: "sender",
		TargetID:   "two@example.com",
		Tag:        "newsletter",
		SourceKind: annotate.SourceKindAgent,
		AgentRunID: "run-2",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation(run-2) error = %v", err)
	}

	annotations, err := repo.ListAnnotations(context.Background(), annotate.ListAnnotationsFilter{
		AgentRunID: "run-2",
	})
	if err != nil {
		t.Fatalf("ListAnnotations() error = %v", err)
	}
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].AgentRunID != "run-2" {
		t.Fatalf("expected agent run id run-2, got %q", annotations[0].AgentRunID)
	}
}

func TestRepositoryBatchUpdateReviewState(t *testing.T) {
	db := openTestDB(t)
	repo := annotate.NewRepository(db)

	created := make([]*annotate.Annotation, 0, 2)
	for _, targetID := range []string{"one@example.com", "two@example.com"} {
		annotation, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
			TargetType: "sender",
			TargetID:   targetID,
			Tag:        "newsletter",
			SourceKind: annotate.SourceKindAgent,
			AgentRunID: "run-batch",
		})
		if err != nil {
			t.Fatalf("CreateAnnotation(%s) error = %v", targetID, err)
		}
		created = append(created, annotation)
	}

	if err := repo.BatchUpdateReviewState(context.Background(), []string{
		created[0].ID,
		created[1].ID,
	}, annotate.ReviewStateReviewed); err != nil {
		t.Fatalf("BatchUpdateReviewState() error = %v", err)
	}

	annotations, err := repo.ListAnnotations(context.Background(), annotate.ListAnnotationsFilter{
		AgentRunID: "run-batch",
	})
	if err != nil {
		t.Fatalf("ListAnnotations() error = %v", err)
	}
	for _, annotation := range annotations {
		if annotation.ReviewState != annotate.ReviewStateReviewed {
			t.Fatalf("expected all annotations reviewed, got %q", annotation.ReviewState)
		}
	}
}

func TestRepositoryListRunsAndGetRunDetail(t *testing.T) {
	db := openTestDB(t)
	repo := annotate.NewRepository(db)

	created, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:  "sender",
		TargetID:    "sender@example.com",
		Tag:         "newsletter",
		SourceKind:  annotate.SourceKindAgent,
		SourceLabel: "triage-agent-v1",
		AgentRunID:  "run-42",
	})
	if err != nil {
		t.Fatalf("CreateAnnotation() error = %v", err)
	}
	if _, err := repo.UpdateAnnotationReviewState(context.Background(), created.ID, annotate.ReviewStateReviewed); err != nil {
		t.Fatalf("UpdateAnnotationReviewState() error = %v", err)
	}

	if _, err := repo.CreateAnnotation(context.Background(), annotate.CreateAnnotationInput{
		TargetType:  "sender",
		TargetID:    "sender-two@example.com",
		Tag:         "notification",
		SourceKind:  annotate.SourceKindAgent,
		SourceLabel: "triage-agent-v1",
		AgentRunID:  "run-42",
	}); err != nil {
		t.Fatalf("CreateAnnotation(second) error = %v", err)
	}

	group, err := repo.CreateGroup(context.Background(), annotate.CreateGroupInput{
		Name:        "Likely newsletters",
		SourceKind:  annotate.SourceKindAgent,
		SourceLabel: "triage-agent-v1",
		AgentRunID:  "run-42",
	})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if err := repo.AddGroupMember(context.Background(), annotate.AddGroupMemberInput{
		GroupID:    group.ID,
		TargetType: "sender",
		TargetID:   "sender@example.com",
	}); err != nil {
		t.Fatalf("AddGroupMember() error = %v", err)
	}

	if _, err := repo.CreateLog(context.Background(), annotate.CreateLogInput{
		Title:        "Run summary",
		BodyMarkdown: "Found likely newsletters.",
		SourceKind:   annotate.SourceKindAgent,
		SourceLabel:  "triage-agent-v1",
		AgentRunID:   "run-42",
	}); err != nil {
		t.Fatalf("CreateLog() error = %v", err)
	}

	runs, err := repo.ListRuns(context.Background())
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].RunID != "run-42" {
		t.Fatalf("expected run id run-42, got %q", runs[0].RunID)
	}
	if runs[0].AnnotationCount != 2 {
		t.Fatalf("expected annotation count 2, got %d", runs[0].AnnotationCount)
	}
	if runs[0].ReviewedCount != 1 {
		t.Fatalf("expected reviewed count 1, got %d", runs[0].ReviewedCount)
	}
	if runs[0].PendingCount != 1 {
		t.Fatalf("expected pending count 1, got %d", runs[0].PendingCount)
	}
	if runs[0].LogCount != 1 {
		t.Fatalf("expected log count 1, got %d", runs[0].LogCount)
	}
	if runs[0].GroupCount != 1 {
		t.Fatalf("expected group count 1, got %d", runs[0].GroupCount)
	}

	detail, err := repo.GetRunDetail(context.Background(), "run-42")
	if err != nil {
		t.Fatalf("GetRunDetail() error = %v", err)
	}
	if len(detail.Annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(detail.Annotations))
	}
	if len(detail.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(detail.Logs))
	}
	if len(detail.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(detail.Groups))
	}
}
