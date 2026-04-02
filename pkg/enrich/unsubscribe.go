package enrich

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type UnsubscribeEnricher struct{}

type unsubscribeMessageRow struct {
	ID          int64  `db:"id"`
	SentDate    string `db:"sent_date"`
	FromSummary string `db:"from_summary"`
	SenderEmail string `db:"sender_email"`
	HeadersJSON string `db:"headers_json"`
}

type unsubscribeRecord struct {
	Email              string
	DisplayName        string
	Domain             string
	Mailto             string
	HTTPURL            string
	OneClick           bool
	IsPrivateRelay     bool
	RelayDisplayDomain string
	SentDate           string
}

func (e *UnsubscribeEnricher) Enrich(ctx context.Context, db *sqlx.DB, opts Options) (UnsubscribeReport, error) {
	start := time.Now()
	opts = normalizeOptions(opts)

	if db == nil {
		return UnsubscribeReport{}, errors.New("unsubscribe enricher database is nil")
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return UnsubscribeReport{}, errors.Wrap(err, "begin unsubscribe enrichment transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows, err := loadUnsubscribeRows(ctx, tx, opts)
	if err != nil {
		return UnsubscribeReport{}, err
	}

	records := map[string]unsubscribeRecord{}
	for _, row := range rows {
		listHeader := GetHeader(row.HeadersJSON, "List-Unsubscribe")
		postHeader := GetHeader(row.HeadersJSON, "List-Unsubscribe-Post")
		mailto, httpURL, oneClick := ParseListUnsubscribe(listHeader, postHeader)
		if mailto == "" && httpURL == "" {
			continue
		}

		email := strings.TrimSpace(row.SenderEmail)
		displayName := ""
		if email == "" {
			parsedEmail, parsedDisplay, err := ParseFromSummary(row.FromSummary)
			if err != nil || parsedEmail == "" {
				continue
			}
			email = parsedEmail
			displayName = parsedDisplay
		} else {
			_, parsedDisplay, err := ParseFromSummary(row.FromSummary)
			if err == nil {
				displayName = parsedDisplay
			}
		}

		domain := ExtractDomain(email)
		record := unsubscribeRecord{
			Email:          email,
			DisplayName:    displayName,
			Domain:         domain,
			Mailto:         mailto,
			HTTPURL:        httpURL,
			OneClick:       oneClick,
			IsPrivateRelay: IsPrivateRelay(domain),
			SentDate:       row.SentDate,
		}
		if record.IsPrivateRelay {
			record.RelayDisplayDomain = GuessRelayDomain(displayName)
		}

		existing, ok := records[email]
		if !ok || record.SentDate >= existing.SentDate {
			records[email] = record
		}
	}

	report := UnsubscribeReport{}
	for _, record := range records {
		if !opts.Rebuild {
			hasList, err := senderHasListUnsubscribe(ctx, tx, record.Email)
			if err != nil {
				return UnsubscribeReport{}, err
			}
			if hasList {
				continue
			}
		}

		report.SendersWithUnsubscribe++
		if record.Mailto != "" {
			report.MailtoLinks++
		}
		if record.HTTPURL != "" {
			report.HTTPLinks++
		}
		if record.OneClick {
			report.OneClickLinks++
		}

		if opts.DryRun {
			continue
		}
		if err := upsertSenderUnsubscribe(ctx, tx, record); err != nil {
			return UnsubscribeReport{}, err
		}
	}

	if opts.DryRun {
		report.ElapsedMS = time.Since(start).Milliseconds()
		return report, nil
	}

	if err := setMetadataValue(ctx, tx, metadataKeyEnrichUnsubAt, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return UnsubscribeReport{}, err
	}

	if err := tx.Commit(); err != nil {
		return UnsubscribeReport{}, errors.Wrap(err, "commit unsubscribe enrichment transaction")
	}

	report.ElapsedMS = time.Since(start).Milliseconds()
	return report, nil
}

func loadUnsubscribeRows(ctx context.Context, tx *sqlx.Tx, opts Options) ([]unsubscribeMessageRow, error) {
	whereClause, args := buildMessageScopeClause(opts, "")
	query := `SELECT
		id,
		sent_date,
		from_summary,
		sender_email,
		headers_json
	FROM messages
	WHERE headers_json LIKE '%List-Unsubscribe%'
		AND ` + whereClause + `
	ORDER BY sent_date DESC, id DESC`

	rows := []unsubscribeMessageRow{}
	if err := tx.SelectContext(ctx, &rows, tx.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "load unsubscribe rows")
	}
	return rows, nil
}

func senderHasListUnsubscribe(ctx context.Context, tx *sqlx.Tx, email string) (bool, error) {
	var hasList bool
	query := tx.Rebind(`SELECT has_list_unsubscribe FROM senders WHERE email = ?`)
	if err := tx.GetContext(ctx, &hasList, query, email); err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return false, nil
		}
		return false, errors.Wrapf(err, "load unsubscribe status for sender %s", email)
	}
	return hasList, nil
}

func upsertSenderUnsubscribe(ctx context.Context, tx *sqlx.Tx, record unsubscribeRecord) error {
	query := tx.Rebind(`INSERT INTO senders (
		email,
		display_name,
		domain,
		is_private_relay,
		relay_display_domain,
		unsubscribe_mailto,
		unsubscribe_http,
		has_list_unsubscribe,
		last_synced_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, CURRENT_TIMESTAMP)
	ON CONFLICT(email) DO UPDATE SET
		display_name = CASE
			WHEN senders.display_name = '' AND excluded.display_name != '' THEN excluded.display_name
			ELSE senders.display_name
		END,
		domain = CASE
			WHEN senders.domain = '' THEN excluded.domain
			ELSE senders.domain
		END,
		is_private_relay = CASE
			WHEN senders.is_private_relay THEN TRUE
			ELSE excluded.is_private_relay
		END,
		relay_display_domain = CASE
			WHEN senders.relay_display_domain = '' AND excluded.relay_display_domain != '' THEN excluded.relay_display_domain
			ELSE senders.relay_display_domain
		END,
		unsubscribe_mailto = CASE
			WHEN excluded.unsubscribe_mailto != '' THEN excluded.unsubscribe_mailto
			ELSE senders.unsubscribe_mailto
		END,
		unsubscribe_http = CASE
			WHEN excluded.unsubscribe_http != '' THEN excluded.unsubscribe_http
			ELSE senders.unsubscribe_http
		END,
		has_list_unsubscribe = TRUE,
		last_synced_at = CURRENT_TIMESTAMP`)

	if _, err := tx.ExecContext(
		ctx,
		query,
		record.Email,
		record.DisplayName,
		record.Domain,
		record.IsPrivateRelay,
		record.RelayDisplayDomain,
		record.Mailto,
		record.HTTPURL,
	); err != nil {
		return errors.Wrapf(err, "upsert unsubscribe data for sender %s", record.Email)
	}

	return nil
}
