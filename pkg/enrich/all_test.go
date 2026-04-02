package enrich

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func TestRunAllExecutesAllEnrichers(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "mail.sqlite")
	db := openEnrichTestFileDB(t, path)
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<root@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "Marketing <marketing@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:marketing@example.com>","References":"<root@example.com>"}`,
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<reply@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Support <support@example.com>",
		HeadersJSON: `{"In-Reply-To":"<root@example.com>","References":"<root@example.com>"}`,
	})
	_ = db.Close()

	report, err := RunAll(t.Context(), path, Options{})
	if err != nil {
		t.Fatalf("RunAll() error = %v", err)
	}
	if report.Senders.SendersCreated != 2 {
		t.Fatalf("unexpected sender report: %+v", report.Senders)
	}
	if report.Threads.ThreadsCreated != 1 {
		t.Fatalf("unexpected thread report: %+v", report.Threads)
	}
	if report.Unsubscribe.SendersWithUnsubscribe != 1 {
		t.Fatalf("unexpected unsubscribe report: %+v", report.Unsubscribe)
	}

	verifyDB := sqlx.MustOpen("sqlite3", path)
	defer func() { _ = verifyDB.Close() }()

	var threadCount int
	if err := verifyDB.Get(&threadCount, `SELECT COUNT(1) FROM threads`); err != nil {
		t.Fatalf("count threads: %v", err)
	}
	if threadCount != 1 {
		t.Fatalf("expected one thread row, got %d", threadCount)
	}
}
