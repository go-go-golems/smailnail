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
