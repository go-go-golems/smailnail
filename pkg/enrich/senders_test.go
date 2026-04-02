package enrich

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

type senderRow struct {
	Email              string `db:"email"`
	DisplayName        string `db:"display_name"`
	Domain             string `db:"domain"`
	IsPrivateRelay     bool   `db:"is_private_relay"`
	RelayDisplayDomain string `db:"relay_display_domain"`
	MsgCount           int    `db:"msg_count"`
}

func TestSenderEnricherEnrichesAndTagsMessages(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<m1@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "Zillow <relay@privaterelay.appleid.com>",
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<m2@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Zillow <relay@privaterelay.appleid.com>",
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         3,
		MessageID:   "<m3@example.com>",
		SentDate:    "2026-04-03T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
	})

	enricher := &SenderEnricher{}
	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("Enrich() error = %v", err)
	}
	if report.SendersCreated != 2 || report.SendersUpdated != 0 {
		t.Fatalf("unexpected sender report: %+v", report)
	}
	if report.MessagesTagged != 3 {
		t.Fatalf("expected 3 tagged messages, got %+v", report)
	}
	if report.PrivateRelayCount != 1 {
		t.Fatalf("expected 1 private relay sender, got %+v", report)
	}

	var relay senderRow
	if err := db.Get(&relay, `SELECT email, display_name, domain, is_private_relay, relay_display_domain, msg_count FROM senders WHERE email = ?`, "relay@privaterelay.appleid.com"); err != nil {
		t.Fatalf("load relay sender: %v", err)
	}
	if !relay.IsPrivateRelay {
		t.Fatalf("expected relay sender row: %+v", relay)
	}
	if relay.RelayDisplayDomain != "zillow" {
		t.Fatalf("expected relay display slug %q, got %+v", "zillow", relay)
	}
	if relay.MsgCount != 2 {
		t.Fatalf("expected relay count 2, got %+v", relay)
	}

	var taggedCount int
	if err := db.Get(&taggedCount, `SELECT COUNT(1) FROM messages WHERE sender_email != ''`); err != nil {
		t.Fatalf("count tagged messages: %v", err)
	}
	if taggedCount != 3 {
		t.Fatalf("expected 3 tagged messages in DB, got %d", taggedCount)
	}
}

func TestSenderEnricherIncrementalRunOnlyTagsNewMessages(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<m1@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
	})

	enricher := &SenderEnricher{}
	if _, err := enricher.Enrich(t.Context(), db, Options{}); err != nil {
		t.Fatalf("first Enrich() error = %v", err)
	}

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<m2@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "John Doe <john@example.com>",
	})

	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("second Enrich() error = %v", err)
	}
	if report.SendersCreated != 0 || report.SendersUpdated != 1 {
		t.Fatalf("unexpected incremental report: %+v", report)
	}
	if report.MessagesTagged != 1 {
		t.Fatalf("expected only the new message to be tagged, got %+v", report)
	}

	var msgCount int
	if err := db.Get(&msgCount, `SELECT msg_count FROM senders WHERE email = ?`, "john@example.com"); err != nil {
		t.Fatalf("load sender count: %v", err)
	}
	if msgCount != 2 {
		t.Fatalf("expected sender msg_count 2 after incremental run, got %d", msgCount)
	}
}

type testMessageRow struct {
	AccountKey  string
	MailboxName string
	UID         int
	MessageID   string
	SentDate    string
	FromSummary string
	HeadersJSON string
}

func openEnrichTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db := sqlx.MustOpen("sqlite3", ":memory:")
	statements := []string{
		`CREATE TABLE mirror_metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_key TEXT NOT NULL,
			mailbox_name TEXT NOT NULL,
			uidvalidity INTEGER NOT NULL DEFAULT 1,
			uid INTEGER NOT NULL,
			message_id TEXT NOT NULL DEFAULT '',
			internal_date TEXT NOT NULL DEFAULT '',
			sent_date TEXT NOT NULL DEFAULT '',
			subject TEXT NOT NULL DEFAULT '',
			from_summary TEXT NOT NULL DEFAULT '',
			to_summary TEXT NOT NULL DEFAULT '',
			cc_summary TEXT NOT NULL DEFAULT '',
			size_bytes INTEGER NOT NULL DEFAULT 0,
			flags_json TEXT NOT NULL DEFAULT '[]',
			headers_json TEXT NOT NULL DEFAULT '{}',
			parts_json TEXT NOT NULL DEFAULT '[]',
			body_text TEXT NOT NULL DEFAULT '',
			body_html TEXT NOT NULL DEFAULT '',
			search_text TEXT NOT NULL DEFAULT '',
			raw_path TEXT NOT NULL,
			raw_sha256 TEXT NOT NULL DEFAULT '',
			has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
			remote_deleted BOOLEAN NOT NULL DEFAULT FALSE,
			first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (account_key, mailbox_name, uidvalidity, uid)
		)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("exec test schema: %v", err)
		}
	}
	for _, statement := range SchemaMigrationV2Statements() {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("exec enrich schema: %v", err)
		}
	}

	return db
}

func insertTestMessage(t *testing.T, db *sqlx.DB, row testMessageRow) {
	t.Helper()

	if row.HeadersJSON == "" {
		row.HeadersJSON = "{}"
	}

	if _, err := db.Exec(
		`INSERT INTO messages (
			account_key,
			mailbox_name,
			uidvalidity,
			uid,
			message_id,
			sent_date,
			from_summary,
			headers_json,
			raw_path
		) VALUES (?, ?, 1, ?, ?, ?, ?, ?, ?)`,
		row.AccountKey,
		row.MailboxName,
		row.UID,
		row.MessageID,
		row.SentDate,
		row.FromSummary,
		row.HeadersJSON,
		"/tmp/test-message.eml",
	); err != nil {
		t.Fatalf("insert test message: %v", err)
	}
}
