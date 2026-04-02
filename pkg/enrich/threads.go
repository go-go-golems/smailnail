package enrich

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var messageIDPattern = regexp.MustCompile(`<[^>]+>`)

type ThreadEnricher struct{}

type threadMessageRow struct {
	ID          int64  `db:"id"`
	AccountKey  string `db:"account_key"`
	MailboxName string `db:"mailbox_name"`
	MessageID   string `db:"message_id"`
	Subject     string `db:"subject"`
	SentDate    string `db:"sent_date"`
	FromSummary string `db:"from_summary"`
	HeadersJSON string `db:"headers_json"`
	SenderEmail string `db:"sender_email"`
	ThreadID    string `db:"thread_id"`
	ThreadDepth int    `db:"thread_depth"`
}

type threadAssignment struct {
	ThreadID string
	Depth    int
}

type threadSummary struct {
	ThreadID         string
	Subject          string
	AccountKey       string
	MailboxName      string
	MessageCount     int
	ParticipantCount int
	FirstSentDate    string
	LastSentDate     string
}

func (e *ThreadEnricher) Enrich(ctx context.Context, db *sqlx.DB, opts Options) (ThreadsReport, error) {
	start := time.Now()
	opts = normalizeOptions(opts)

	if db == nil {
		return ThreadsReport{}, errors.New("thread enricher database is nil")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return ThreadsReport{}, errors.Wrap(err, "begin thread enrichment transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows, err := loadThreadRows(ctx, tx, opts)
	if err != nil {
		return ThreadsReport{}, err
	}

	present := map[string]struct{}{}
	parentOf := map[string]string{}
	assignments := map[int64]threadAssignment{}
	rowByMessageID := map[string]threadMessageRow{}
	touchedThreadIDs := map[string]struct{}{}

	for _, row := range rows {
		if row.MessageID == "" {
			continue
		}
		present[row.MessageID] = struct{}{}
		if _, ok := rowByMessageID[row.MessageID]; !ok {
			rowByMessageID[row.MessageID] = row
		}
		parentOf[row.MessageID] = resolveParentMessageID(row.HeadersJSON)
	}

	targetRows := make([]threadMessageRow, 0, len(rows))
	for _, row := range rows {
		if row.MessageID == "" {
			continue
		}
		if !opts.Rebuild && row.ThreadID != "" {
			continue
		}
		targetRows = append(targetRows, row)

		root, depth := resolveThreadRoot(row.MessageID, parentOf, present)
		assignments[row.ID] = threadAssignment{
			ThreadID: root,
			Depth:    depth,
		}
		if root != "" {
			touchedThreadIDs[root] = struct{}{}
		}
	}

	allAssignments := map[string]threadAssignment{}
	for _, row := range rows {
		if row.MessageID == "" {
			continue
		}
		root, depth := resolveThreadRoot(row.MessageID, parentOf, present)
		allAssignments[row.MessageID] = threadAssignment{
			ThreadID: root,
			Depth:    depth,
		}
	}

	report := ThreadsReport{
		MessagesProcessed: len(targetRows),
	}

	for threadID := range touchedThreadIDs {
		exists, err := threadExists(ctx, tx, threadID)
		if err != nil {
			return ThreadsReport{}, err
		}
		if exists {
			report.ThreadsUpdated++
		} else {
			report.ThreadsCreated++
		}
	}

	if opts.DryRun {
		report.ElapsedMS = time.Since(start).Milliseconds()
		return report, nil
	}

	for _, row := range targetRows {
		assignment := assignments[row.ID]
		query := tx.Rebind(`UPDATE messages SET thread_id = ?, thread_depth = ? WHERE id = ?`)
		if _, err := tx.ExecContext(ctx, query, assignment.ThreadID, assignment.Depth, row.ID); err != nil {
			return ThreadsReport{}, errors.Wrapf(err, "update thread assignment for message %d", row.ID)
		}
	}

	summaries := buildThreadSummaries(rows, allAssignments)
	for threadID := range touchedThreadIDs {
		summary, ok := summaries[threadID]
		if !ok {
			continue
		}
		if err := upsertThreadSummary(ctx, tx, summary); err != nil {
			return ThreadsReport{}, err
		}
	}

	if err := setMetadataValue(ctx, tx, metadataKeyEnrichThreadsAt, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return ThreadsReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return ThreadsReport{}, errors.Wrap(err, "commit thread enrichment transaction")
	}

	report.ElapsedMS = time.Since(start).Milliseconds()
	return report, nil
}

func loadThreadRows(ctx context.Context, tx *sqlx.Tx, opts Options) ([]threadMessageRow, error) {
	whereClause, args := buildMessageScopeClause(opts, "")
	query := `SELECT
		id,
		account_key,
		mailbox_name,
		message_id,
		subject,
		sent_date,
		from_summary,
		headers_json,
		sender_email,
		thread_id,
		thread_depth
	FROM messages
	WHERE ` + whereClause + `
	ORDER BY id`

	rows := []threadMessageRow{}
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "load thread rows")
	}
	return rows, nil
}

func resolveParentMessageID(headersJSON string) string {
	references := parseMessageIDList(GetHeader(headersJSON, "References"))
	if len(references) > 0 {
		return references[len(references)-1]
	}
	return normalizeMessageIDValue(GetHeader(headersJSON, "In-Reply-To"))
}

func parseMessageIDList(raw string) []string {
	matches := messageIDPattern.FindAllString(raw, -1)
	if len(matches) > 0 {
		ret := make([]string, 0, len(matches))
		for _, match := range matches {
			if normalized := normalizeMessageIDValue(match); normalized != "" {
				ret = append(ret, normalized)
			}
		}
		return ret
	}

	fields := strings.Fields(raw)
	ret := make([]string, 0, len(fields))
	for _, field := range fields {
		if normalized := normalizeMessageIDValue(field); normalized != "" {
			ret = append(ret, normalized)
		}
	}
	return ret
}

func normalizeMessageIDValue(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "<>")
	if raw == "" {
		return ""
	}
	return "<" + raw + ">"
}

func resolveThreadRoot(messageID string, parentOf map[string]string, present map[string]struct{}) (string, int) {
	current := messageID
	depth := 0
	visited := map[string]bool{current: true}

	for {
		parent := parentOf[current]
		if parent == "" {
			return current, depth
		}
		if visited[parent] {
			return current, depth
		}
		if _, ok := present[parent]; !ok {
			return current, depth
		}

		visited[parent] = true
		current = parent
		depth++
	}
}

func threadExists(ctx context.Context, tx *sqlx.Tx, threadID string) (bool, error) {
	var count int
	query := tx.Rebind(`SELECT COUNT(1) FROM threads WHERE thread_id = ?`)
	if err := tx.GetContext(ctx, &count, query, threadID); err != nil {
		return false, errors.Wrapf(err, "check thread %s", threadID)
	}
	return count > 0, nil
}

func buildThreadSummaries(rows []threadMessageRow, assignments map[string]threadAssignment) map[string]threadSummary {
	summaries := map[string]threadSummary{}
	participants := map[string]map[string]struct{}{}

	for _, row := range rows {
		if row.MessageID == "" {
			continue
		}
		assignment, ok := assignments[row.MessageID]
		if !ok || assignment.ThreadID == "" {
			continue
		}

		summary := summaries[assignment.ThreadID]
		if summary.ThreadID == "" {
			summary = threadSummary{
				ThreadID:      assignment.ThreadID,
				Subject:       row.Subject,
				AccountKey:    row.AccountKey,
				MailboxName:   row.MailboxName,
				FirstSentDate: row.SentDate,
				LastSentDate:  row.SentDate,
			}
		}
		summary.MessageCount++
		if summary.Subject == "" && row.Subject != "" {
			summary.Subject = row.Subject
		}
		if summary.FirstSentDate == "" || (row.SentDate != "" && row.SentDate < summary.FirstSentDate) {
			summary.FirstSentDate = row.SentDate
		}
		if row.SentDate > summary.LastSentDate {
			summary.LastSentDate = row.SentDate
		}

		participant := strings.TrimSpace(row.SenderEmail)
		if participant == "" {
			email, _, err := ParseFromSummary(row.FromSummary)
			if err == nil {
				participant = email
			}
		}
		if participant != "" {
			if _, ok := participants[assignment.ThreadID]; !ok {
				participants[assignment.ThreadID] = map[string]struct{}{}
			}
			participants[assignment.ThreadID][participant] = struct{}{}
			summary.ParticipantCount = len(participants[assignment.ThreadID])
		}

		summaries[assignment.ThreadID] = summary
	}

	return summaries
}

func upsertThreadSummary(ctx context.Context, tx *sqlx.Tx, summary threadSummary) error {
	query := tx.Rebind(`INSERT INTO threads (
		thread_id,
		subject,
		account_key,
		mailbox_name,
		message_count,
		participant_count,
		first_sent_date,
		last_sent_date,
		last_rebuilt_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(thread_id) DO UPDATE SET
		subject = excluded.subject,
		account_key = excluded.account_key,
		mailbox_name = excluded.mailbox_name,
		message_count = excluded.message_count,
		participant_count = excluded.participant_count,
		first_sent_date = excluded.first_sent_date,
		last_sent_date = excluded.last_sent_date,
		last_rebuilt_at = CURRENT_TIMESTAMP`)
	if _, err := tx.ExecContext(
		ctx,
		query,
		summary.ThreadID,
		summary.Subject,
		summary.AccountKey,
		summary.MailboxName,
		summary.MessageCount,
		summary.ParticipantCount,
		summary.FirstSentDate,
		summary.LastSentDate,
	); err != nil {
		return errors.Wrapf(err, "upsert thread summary %s", summary.ThreadID)
	}

	return nil
}
