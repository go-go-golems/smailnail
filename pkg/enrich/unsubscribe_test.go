package enrich

import "testing"

type unsubscribeRow struct {
	Email              string `db:"email"`
	UnsubscribeMailto  string `db:"unsubscribe_mailto"`
	UnsubscribeHTTP    string `db:"unsubscribe_http"`
	HasListUnsubscribe bool   `db:"has_list_unsubscribe"`
}

func TestUnsubscribeEnricherExtractsLatestLinks(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<m1@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "Marketing <marketing@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:old@example.com>, <https://old.example.com/unsub>"}`,
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<m2@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Marketing <marketing@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:new@example.com>, <https://new.example.com/unsub>","List-Unsubscribe-Post":"List-Unsubscribe=One-Click"}`,
	})
	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         3,
		MessageID:   "<m3@example.com>",
		SentDate:    "2026-04-03T10:00:00Z",
		FromSummary: "Alerts <alerts@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:alerts@example.com>"}`,
	})

	enricher := &UnsubscribeEnricher{}
	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("Enrich() error = %v", err)
	}
	if report.SendersWithUnsubscribe != 2 || report.MailtoLinks != 2 || report.HTTPLinks != 1 || report.OneClickLinks != 1 {
		t.Fatalf("unexpected unsubscribe report: %+v", report)
	}

	var marketing unsubscribeRow
	if err := db.Get(&marketing, `SELECT email, unsubscribe_mailto, unsubscribe_http, has_list_unsubscribe FROM senders WHERE email = ?`, "marketing@example.com"); err != nil {
		t.Fatalf("load marketing sender: %v", err)
	}
	if marketing.UnsubscribeMailto != "mailto:new@example.com" || marketing.UnsubscribeHTTP != "https://new.example.com/unsub" || !marketing.HasListUnsubscribe {
		t.Fatalf("unexpected marketing sender row: %+v", marketing)
	}
}

func TestUnsubscribeEnricherSkipsSendersAlreadyProcessedOnIncrementalRun(t *testing.T) {
	t.Parallel()

	db := openEnrichTestDB(t)
	defer func() { _ = db.Close() }()

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         1,
		MessageID:   "<m1@example.com>",
		SentDate:    "2026-04-01T10:00:00Z",
		FromSummary: "Marketing <marketing@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:marketing@example.com>"}`,
	})

	enricher := &UnsubscribeEnricher{}
	if _, err := enricher.Enrich(t.Context(), db, Options{}); err != nil {
		t.Fatalf("first Enrich() error = %v", err)
	}

	insertTestMessage(t, db, testMessageRow{
		AccountKey:  "acct-1",
		MailboxName: "INBOX",
		UID:         2,
		MessageID:   "<m2@example.com>",
		SentDate:    "2026-04-02T10:00:00Z",
		FromSummary: "Alerts <alerts@example.com>",
		HeadersJSON: `{"List-Unsubscribe":"<mailto:alerts@example.com>"}`,
	})

	report, err := enricher.Enrich(t.Context(), db, Options{})
	if err != nil {
		t.Fatalf("second Enrich() error = %v", err)
	}
	if report.SendersWithUnsubscribe != 1 || report.MailtoLinks != 1 || report.HTTPLinks != 0 {
		t.Fatalf("unexpected incremental unsubscribe report: %+v", report)
	}
}
