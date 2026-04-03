package enrich

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type SenderEnricher struct{}

type senderSummaryRow struct {
	FromSummary   string `db:"from_summary"`
	MessageCount  int    `db:"message_count"`
	FirstSeenDate string `db:"first_seen_date"`
	LastSeenDate  string `db:"last_seen_date"`
}

type senderUpsertRecord struct {
	Email              string
	DisplayName        string
	Domain             string
	IsPrivateRelay     bool
	RelayDisplayDomain string
	MessageCount       int
	FirstSeenDate      string
	LastSeenDate       string
}

func (e *SenderEnricher) Enrich(ctx context.Context, db *sqlx.DB, opts Options) (SendersReport, error) {
	start := time.Now()
	opts = normalizeOptions(opts)

	if db == nil {
		return SendersReport{}, errors.New("sender enricher database is nil")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return SendersReport{}, errors.Wrap(err, "begin sender enrichment transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if opts.Rebuild {
		if err := clearSenderColumns(ctx, tx, opts); err != nil {
			return SendersReport{}, err
		}
	}

	rows, err := loadSenderSummaryRows(ctx, tx, opts)
	if err != nil {
		return SendersReport{}, err
	}

	report := SendersReport{}
	messageUpdates := make([]senderSummaryRow, 0, len(rows))
	senderRecords := map[string]*senderUpsertRecord{}

	for _, row := range rows {
		email, displayName, err := ParseFromSummary(row.FromSummary)
		if err != nil || email == "" {
			continue
		}

		domain := ExtractDomain(email)
		record := senderRecords[email]
		if record == nil {
			record = &senderUpsertRecord{
				Email:          email,
				DisplayName:    displayName,
				Domain:         domain,
				IsPrivateRelay: IsPrivateRelay(domain),
				FirstSeenDate:  row.FirstSeenDate,
				LastSeenDate:   row.LastSeenDate,
			}
			if record.IsPrivateRelay {
				record.RelayDisplayDomain = GuessRelayDomain(displayName)
			}
			senderRecords[email] = record
		}

		record.MessageCount += row.MessageCount
		if record.DisplayName == "" && displayName != "" {
			record.DisplayName = displayName
		}
		if record.FirstSeenDate == "" || (row.FirstSeenDate != "" && row.FirstSeenDate < record.FirstSeenDate) {
			record.FirstSeenDate = row.FirstSeenDate
		}
		if row.LastSeenDate > record.LastSeenDate {
			record.LastSeenDate = row.LastSeenDate
		}
		if record.RelayDisplayDomain == "" && record.IsPrivateRelay {
			record.RelayDisplayDomain = GuessRelayDomain(displayName)
		}

		messageUpdates = append(messageUpdates, row)
	}

	for _, record := range senderRecords {
		exists, err := senderExists(ctx, tx, record.Email)
		if err != nil {
			return SendersReport{}, err
		}
		if exists {
			report.SendersUpdated++
		} else {
			report.SendersCreated++
		}
		if record.IsPrivateRelay {
			report.PrivateRelayCount++
		}
	}

	if opts.DryRun {
		for _, row := range rows {
			report.MessagesTagged += row.MessageCount
		}
		report.ElapsedMS = time.Since(start).Milliseconds()
		return report, nil
	}

	for _, record := range senderRecords {
		if err := upsertSender(ctx, tx, *record, opts.Rebuild); err != nil {
			return SendersReport{}, err
		}
	}

	for _, row := range messageUpdates {
		email, _, err := ParseFromSummary(row.FromSummary)
		if err != nil || email == "" {
			continue
		}
		domain := ExtractDomain(email)
		tagged, err := tagMessagesForSender(ctx, tx, opts, row.FromSummary, email, domain)
		if err != nil {
			return SendersReport{}, err
		}
		report.MessagesTagged += tagged
	}

	if err := setMetadataValue(ctx, tx, metadataKeyEnrichSendersAt, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return SendersReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return SendersReport{}, errors.Wrap(err, "commit sender enrichment transaction")
	}

	report.ElapsedMS = time.Since(start).Milliseconds()
	return report, nil
}

func loadSenderSummaryRows(ctx context.Context, tx *sqlx.Tx, opts Options) ([]senderSummaryRow, error) {
	whereClause, args := buildMessageScopeClause(opts, "")
	query := `SELECT
		from_summary,
		COUNT(*) AS message_count,
		MIN(NULLIF(sent_date, '')) AS first_seen_date,
		MAX(NULLIF(sent_date, '')) AS last_seen_date
	FROM messages
	WHERE from_summary != ''
		AND ` + whereClause

	if !opts.Rebuild {
		query += ` AND sender_email = ''`
	}

	query += `
	GROUP BY from_summary
	ORDER BY from_summary`

	rows := []senderSummaryRow{}
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "load sender summary rows")
	}

	return rows, nil
}

func clearSenderColumns(ctx context.Context, tx *sqlx.Tx, opts Options) error {
	whereClause, args := buildMessageScopeClause(opts, "")
	query := `UPDATE messages
	SET sender_email = '', sender_domain = ''
	WHERE ` + whereClause

	if _, err := tx.ExecContext(ctx, tx.Rebind(query), args...); err != nil {
		return errors.Wrap(err, "clear sender columns")
	}

	return nil
}

func senderExists(ctx context.Context, tx *sqlx.Tx, email string) (bool, error) {
	var count int
	query := tx.Rebind(`SELECT COUNT(1) FROM senders WHERE email = ?`)
	if err := tx.GetContext(ctx, &count, query, email); err != nil {
		return false, errors.Wrapf(err, "check sender %s", email)
	}
	return count > 0, nil
}

func upsertSender(ctx context.Context, tx *sqlx.Tx, record senderUpsertRecord, rebuild bool) error {
	msgCountExpr := "senders.msg_count + excluded.msg_count"
	if rebuild {
		msgCountExpr = "excluded.msg_count"
	}

	query := tx.Rebind(`INSERT INTO senders (
		email,
		display_name,
		domain,
		is_private_relay,
		relay_display_domain,
		msg_count,
		first_seen_date,
		last_seen_date,
		last_synced_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(email) DO UPDATE SET
		display_name = CASE
			WHEN excluded.display_name != '' THEN excluded.display_name
			ELSE senders.display_name
		END,
		domain = excluded.domain,
		is_private_relay = excluded.is_private_relay,
		relay_display_domain = CASE
			WHEN excluded.relay_display_domain != '' THEN excluded.relay_display_domain
			ELSE senders.relay_display_domain
		END,
		msg_count = ` + msgCountExpr + `,
		first_seen_date = CASE
			WHEN senders.first_seen_date = '' THEN excluded.first_seen_date
			WHEN excluded.first_seen_date = '' THEN senders.first_seen_date
			WHEN excluded.first_seen_date < senders.first_seen_date THEN excluded.first_seen_date
			ELSE senders.first_seen_date
		END,
		last_seen_date = CASE
			WHEN excluded.last_seen_date > senders.last_seen_date THEN excluded.last_seen_date
			ELSE senders.last_seen_date
		END,
		last_synced_at = CURRENT_TIMESTAMP`)

	if _, err := tx.ExecContext(
		ctx,
		query,
		record.Email,
		record.DisplayName,
		record.Domain,
		record.IsPrivateRelay,
		record.RelayDisplayDomain,
		record.MessageCount,
		record.FirstSeenDate,
		record.LastSeenDate,
	); err != nil {
		return errors.Wrapf(err, "upsert sender %s", record.Email)
	}

	return nil
}

func tagMessagesForSender(ctx context.Context, tx *sqlx.Tx, opts Options, fromSummary, email, domain string) (int, error) {
	whereClause, args := buildMessageScopeClause(opts, "")
	query := `UPDATE messages
	SET sender_email = ?, sender_domain = ?
	WHERE from_summary = ?
		AND ` + whereClause

	params := []any{email, domain, fromSummary}
	params = append(params, args...)
	if !opts.Rebuild {
		query += ` AND sender_email = ''`
	}

	result, err := tx.ExecContext(ctx, tx.Rebind(query), params...)
	if err != nil {
		return 0, errors.Wrapf(err, "tag messages for sender %s", email)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrapf(err, "read updated rows for sender %s", email)
	}

	return int(affected), nil
}
