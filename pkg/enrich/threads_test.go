package enrich

import "testing"

type threadRow struct {
	ThreadID         string `db:"thread_id"`
	MessageCount     int    `db:"message_count"`
	ParticipantCount int    `db:"participant_count"`
}

func TestThreadEnricherBuildsThreadsAndDepths(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<root@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<reply@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Jane Doe <jane@example.com>",
		HeadersJSON: `{"References":"<root@example.com>","In-Reply-To":"<root@example.com>"}`,
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         3,
		MessageID:   "<dangling@example.com>",
		SentDate:    "2026-04-03T10:00:00Z",
		FromSummary: "List <list@example.com>",
		HeadersJSON: `{"References":"<missing@example.com>","In-Reply-To":"<missing@example.com>"}`,
	})

	enricher := &ThreadEnricher{}
	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("Enrich() error = %v", err)
	}
	if report.MessagesProcessed != 3 || report.ThreadsCreated != 2 || report.ThreadsUpdated != 0 {
		t.Fatalf("unexpected thread report: %+v", report)
	}

	var rootThreadID string
	var rootDepth int
	if err := db.Get(&rootThreadID, `SELECT thread_id FROM messages WHERE message_id = ?`, "<root@example.com>"); err != nil {
		t.Fatalf("load root thread_id: %v", err)
	}
	if rootThreadID != "<root@example.com>" {
		t.Fatalf("expected root thread id, got %q", rootThreadID)
	}

	if err := db.Get(&rootDepth, `SELECT thread_depth FROM messages WHERE message_id = ?`, "<reply@example.com>"); err != nil {
		t.Fatalf("load reply depth: %v", err)
	}
	if rootDepth != 1 {
		t.Fatalf("expected reply depth 1, got %d", rootDepth)
	}

	var summary threadRow
	if err := db.Get(&summary, `SELECT thread_id, message_count, participant_count FROM threads WHERE thread_id = ?`, "<root@example.com>"); err != nil {
		t.Fatalf("load thread summary: %v", err)
	}
	if summary.MessageCount != 2 || summary.ParticipantCount != 2 {
		t.Fatalf("unexpected thread summary: %+v", summary)
	}
}

func TestThreadEnricherIncrementalRunUpdatesTouchedThread(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<root@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<reply@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Jane Doe <jane@example.com>",
		HeadersJSON: `{"References":"<root@example.com>","In-Reply-To":"<root@example.com>"}`,
	})

	enricher := &ThreadEnricher{}
	if _, err := enricher.Enrich(t.Context(), db, Options{}); err != nil {
		t.Fatalf("first Enrich() error = %v", err)
	}

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         3,
		MessageID:   "<reply-2@example.com>",
		SentDate:    "2026-04-03T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
		HeadersJSON: `{"References":"<root@example.com> <reply@example.com>","In-Reply-To":"<reply@example.com>"}`,
	})

	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("second Enrich() error = %v", err)
	}
	if report.MessagesProcessed != 1 || report.ThreadsCreated != 0 || report.ThreadsUpdated != 1 {
		t.Fatalf("unexpected incremental thread report: %+v", report)
	}

	var summary threadRow
	if err := db.Get(&summary, `SELECT thread_id, message_count, participant_count FROM threads WHERE thread_id = ?`, "<root@example.com>"); err != nil {
		t.Fatalf("load updated summary: %v", err)
	}
	if summary.MessageCount != 3 {
		t.Fatalf("expected updated message count 3, got %+v", summary)
	}
}
