package mirror

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/go-go-golems/smailnail/pkg/mailruntime"
)

type SyncOptions struct {
	Server            string
	Port              int
	Username          string
	Password          string
	Insecure          bool
	Mailbox           string
	AllMailboxes      bool
	MirrorRoot        string
	BatchSize         int
	ResetMailboxState bool
}

type imapSession interface {
	List(pattern string) ([]mailruntime.MailboxInfo, error)
	Status(name string) (*mailruntime.MailboxStatus, error)
	SelectMailbox(name string, readOnly bool) (*imap.SelectData, error)
	UnselectMailbox() error
	Search(criteria *mailruntime.SearchCriteria) ([]imap.UID, error)
	Fetch(uids []imap.UID, fields []mailruntime.FetchField) ([]*mailruntime.FetchedMessage, error)
	Logout() error
}

type dialIMAPFunc func(ctx context.Context, opts mailruntime.IMAPOptions) (imapSession, error)

type Service struct {
	store *Store
	dial  dialIMAPFunc
	now   func() time.Time
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
		dial: func(ctx context.Context, opts mailruntime.IMAPOptions) (imapSession, error) {
			return mailruntime.Connect(ctx, opts)
		},
		now: time.Now,
	}
}

func (s *Service) Sync(ctx context.Context, opts SyncOptions) (*SyncReport, error) {
	if s == nil || s.store == nil || s.store.db == nil {
		return nil, fmt.Errorf("mirror service store is not initialized")
	}

	normalized := normalizeSyncOptions(opts)
	if err := validateSyncOptions(normalized); err != nil {
		return nil, err
	}

	session, err := s.dial(ctx, mailruntime.IMAPOptions{
		Host:     normalized.Server,
		Port:     normalized.Port,
		TLS:      true,
		Insecure: normalized.Insecure,
		Username: normalized.Username,
		Password: normalized.Password,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Logout()
	}()

	accountKey := AccountKey(normalized.Server, normalized.Port, normalized.Username)
	mailboxes, err := s.resolveMailboxes(session, normalized)
	if err != nil {
		return nil, err
	}

	report := &SyncReport{
		AccountKey:       accountKey,
		MailboxesPlanned: len(mailboxes),
	}
	for _, mailboxName := range mailboxes {
		mailboxReport, err := s.syncMailbox(ctx, session, accountKey, mailboxName, normalized)
		if err != nil {
			return nil, errors.Wrapf(err, "sync mailbox %s", mailboxName)
		}
		report.Mailboxes = append(report.Mailboxes, *mailboxReport)
		report.MailboxesSynced++
		report.MessagesFetched += mailboxReport.FetchedMessages
		report.MessagesStored += mailboxReport.StoredMessages
		report.RawFilesWritten += mailboxReport.RawFilesWritten
		report.ReusedFileWrites += mailboxReport.ReusedFileWrites
	}

	return report, nil
}

func normalizeSyncOptions(opts SyncOptions) SyncOptions {
	if opts.Port == 0 {
		opts.Port = 993
	}
	if strings.TrimSpace(opts.Mailbox) == "" {
		opts.Mailbox = "INBOX"
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}
	if strings.TrimSpace(opts.MirrorRoot) == "" {
		opts.MirrorRoot = DefaultMirrorRoot
	}
	return opts
}

func validateSyncOptions(opts SyncOptions) error {
	if strings.TrimSpace(opts.Server) == "" {
		return fmt.Errorf("server is required")
	}
	if strings.TrimSpace(opts.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if strings.TrimSpace(opts.Password) == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (s *Service) resolveMailboxes(session imapSession, opts SyncOptions) ([]string, error) {
	if !opts.AllMailboxes {
		return []string{opts.Mailbox}, nil
	}

	mailboxes, err := session.List("*")
	if err != nil {
		return nil, errors.Wrap(err, "list mailboxes")
	}

	ret := make([]string, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		if mailbox.Name == "" || mailboxHasNoSelect(mailbox.Flags) {
			continue
		}
		ret = append(ret, mailbox.Name)
	}

	sort.Strings(ret)
	return ret, nil
}

func mailboxHasNoSelect(flags []string) bool {
	for _, flag := range flags {
		if strings.EqualFold(flag, "\\Noselect") {
			return true
		}
	}
	return false
}

func (s *Service) syncMailbox(
	ctx context.Context,
	session imapSession,
	accountKey, mailboxName string,
	opts SyncOptions,
) (*MailboxSyncResult, error) {
	status, err := session.Status(mailboxName)
	if err != nil {
		return nil, errors.Wrap(err, "read mailbox status")
	}
	if status == nil {
		return nil, fmt.Errorf("mailbox status for %s is nil", mailboxName)
	}

	state, err := loadMailboxSyncState(ctx, s.store.db, accountKey, mailboxName)
	if err != nil {
		return nil, err
	}

	resetApplied := false
	if opts.ResetMailboxState {
		if err := resetMailboxState(ctx, s.store.db, accountKey, mailboxName); err != nil {
			return nil, err
		}
		state = nil
		resetApplied = true
	}
	if state != nil && state.UIDValidity != 0 && status.UIDValidity != 0 && state.UIDValidity != status.UIDValidity {
		if err := resetMailboxState(ctx, s.store.db, accountKey, mailboxName); err != nil {
			return nil, err
		}
		state = nil
		resetApplied = true
	}

	previousHighUID := uint32(0)
	if state != nil && (status.UIDValidity == 0 || state.UIDValidity == status.UIDValidity) {
		previousHighUID = state.HighestUID
	}

	if _, err := session.SelectMailbox(mailboxName, true); err != nil {
		return nil, errors.Wrap(err, "select mailbox")
	}
	defer func() {
		_ = session.UnselectMailbox()
	}()

	report := &MailboxSyncResult{
		MailboxName:     mailboxName,
		UIDValidity:     status.UIDValidity,
		UIDNext:         status.UIDNext,
		PreviousHighUID: previousHighUID,
		HighestUID:      previousHighUID,
		ResetApplied:    resetApplied,
	}

	searchCriteria, shouldSearch := newUIDSearchCriteria(previousHighUID, status.UIDNext)
	if !shouldSearch {
		syncTime := s.now().UTC()
		if err := upsertMailboxSyncState(ctx, s.store.db, MailboxSyncState{
			AccountKey:  accountKey,
			MailboxName: mailboxName,
			UIDValidity: status.UIDValidity,
			HighestUID:  previousHighUID,
			LastUIDNext: status.UIDNext,
			LastSyncAt:  &syncTime,
			Status:      "active",
		}); err != nil {
			return nil, err
		}
		return report, nil
	}

	uids, err := session.Search(searchCriteria)
	if err != nil {
		return nil, errors.Wrap(err, "search mailbox UIDs")
	}
	sort.Slice(uids, func(i, j int) bool {
		return uids[i] < uids[j]
	})

	if len(uids) == 0 {
		syncTime := s.now().UTC()
		if err := upsertMailboxSyncState(ctx, s.store.db, MailboxSyncState{
			AccountKey:  accountKey,
			MailboxName: mailboxName,
			UIDValidity: status.UIDValidity,
			HighestUID:  previousHighUID,
			LastUIDNext: status.UIDNext,
			LastSyncAt:  &syncTime,
			Status:      "active",
		}); err != nil {
			return nil, err
		}
		return report, nil
	}

	fetchFields := []mailruntime.FetchField{
		mailruntime.FetchUID,
		mailruntime.FetchFlags,
		mailruntime.FetchInternalDate,
		mailruntime.FetchSize,
		mailruntime.FetchEnvelope,
		mailruntime.FetchHeaders,
		mailruntime.FetchBodyText,
		mailruntime.FetchBodyRaw,
		mailruntime.FetchAttachments,
	}

	for _, batch := range batchUIDs(uids, opts.BatchSize) {
		msgs, err := session.Fetch(batch, fetchFields)
		if err != nil {
			return nil, errors.Wrap(err, "fetch message batch")
		}
		if err := s.persistBatch(ctx, accountKey, mailboxName, status, report, msgs, opts.MirrorRoot); err != nil {
			return nil, err
		}
	}

	return report, nil
}

func newUIDSearchCriteria(previousHighUID, uidNext uint32) (*mailruntime.SearchCriteria, bool) {
	if previousHighUID == 0 {
		return &mailruntime.SearchCriteria{All: true}, true
	}
	if uidNext != 0 && previousHighUID+1 >= uidNext {
		return nil, false
	}

	uidSet := imap.UIDSet{}
	stop := imap.UID(uidNext - 1)
	uidSet.AddRange(imap.UID(previousHighUID+1), stop)
	return &mailruntime.SearchCriteria{
		All: true,
		UID: &uidSet,
	}, true
}

func batchUIDs(uids []imap.UID, batchSize int) [][]imap.UID {
	if batchSize <= 0 || len(uids) == 0 {
		return nil
	}

	ret := make([][]imap.UID, 0, (len(uids)+batchSize-1)/batchSize)
	for start := 0; start < len(uids); start += batchSize {
		end := start + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		ret = append(ret, uids[start:end])
	}
	return ret
}

func (s *Service) persistBatch(
	ctx context.Context,
	accountKey, mailboxName string,
	status *mailruntime.MailboxStatus,
	report *MailboxSyncResult,
	msgs []*mailruntime.FetchedMessage,
	mirrorRoot string,
) error {
	if len(msgs) == 0 {
		return nil
	}

	now := s.now().UTC()
	tx, err := s.store.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin mirror batch transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	highestUID := report.HighestUID
	for _, msg := range msgs {
		rawResult, err := WriteRawMessage(mirrorRoot, accountKey, mailboxName, status.UIDValidity, msg.UID, msg.BodyRaw)
		if err != nil {
			return err
		}

		record, err := buildMessageRecord(accountKey, mailboxName, status.UIDValidity, msg, rawResult, now)
		if err != nil {
			return err
		}
		if err := upsertMessageRecord(ctx, tx, record); err != nil {
			return err
		}

		report.FetchedMessages++
		report.StoredMessages++
		if rawResult.Reused {
			report.ReusedFileWrites++
		} else {
			report.RawFilesWritten++
		}
		if msg.UID > highestUID {
			highestUID = msg.UID
		}
	}

	if err := upsertMailboxSyncState(ctx, tx, MailboxSyncState{
		AccountKey:  accountKey,
		MailboxName: mailboxName,
		UIDValidity: status.UIDValidity,
		HighestUID:  highestUID,
		LastUIDNext: status.UIDNext,
		LastSyncAt:  &now,
		Status:      "active",
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit mirror batch transaction")
	}

	report.HighestUID = highestUID
	return nil
}

func buildMessageRecord(
	accountKey, mailboxName string,
	uidValidity uint32,
	msg *mailruntime.FetchedMessage,
	rawResult *RawMessageResult,
	now time.Time,
) (MessageRecord, error) {
	flagsJSON, err := marshalJSON(msg.Flags, "flags")
	if err != nil {
		return MessageRecord{}, err
	}
	headersJSON, err := marshalJSON(msg.Headers, "headers")
	if err != nil {
		return MessageRecord{}, err
	}
	partsJSON, err := marshalJSON(msg.Attachments, "attachments")
	if err != nil {
		return MessageRecord{}, err
	}

	record := MessageRecord{
		AccountKey:     accountKey,
		MailboxName:    mailboxName,
		UIDValidity:    uidValidity,
		UID:            msg.UID,
		MessageID:      envelopeString(msg, func(e *mailruntime.MessageEnvelope) string { return e.MessageID }),
		InternalDate:   msg.InternalDate,
		SentDate:       envelopeString(msg, func(e *mailruntime.MessageEnvelope) string { return e.Date }),
		Subject:        envelopeString(msg, func(e *mailruntime.MessageEnvelope) string { return e.Subject }),
		FromSummary:    envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.From }),
		ToSummary:      envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.To }),
		CCSummary:      envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.CC }),
		SizeBytes:      msg.Size,
		FlagsJSON:      flagsJSON,
		HeadersJSON:    headersJSON,
		PartsJSON:      partsJSON,
		BodyText:       msg.BodyText,
		BodyHTML:       msg.BodyHTML,
		SearchText:     buildSearchText(msg),
		RawPath:        rawResult.Path,
		RawSHA256:      rawResult.SHA256,
		HasAttachments: len(msg.Attachments) > 0,
		RemoteDeleted:  false,
		FirstSeenAt:    &now,
		LastSyncedAt:   &now,
	}
	return record, nil
}

func envelopeString(msg *mailruntime.FetchedMessage, getter func(*mailruntime.MessageEnvelope) string) string {
	if msg.Envelope == nil {
		return ""
	}
	return getter(msg.Envelope)
}

func envelopeJoin(msg *mailruntime.FetchedMessage, getter func(*mailruntime.MessageEnvelope) []string) string {
	if msg.Envelope == nil {
		return ""
	}
	return strings.Join(getter(msg.Envelope), ", ")
}

func buildSearchText(msg *mailruntime.FetchedMessage) string {
	parts := []string{
		envelopeString(msg, func(e *mailruntime.MessageEnvelope) string { return e.Subject }),
		envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.From }),
		envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.To }),
		envelopeJoin(msg, func(e *mailruntime.MessageEnvelope) []string { return e.CC }),
		msg.BodyText,
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "\n")
}

func marshalJSON(v interface{}, label string) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrapf(err, "marshal %s json", label)
	}
	return string(raw), nil
}

func loadMailboxSyncState(ctx context.Context, db *sqlx.DB, accountKey, mailboxName string) (*MailboxSyncState, error) {
	var state MailboxSyncState
	err := db.GetContext(ctx, &state, `SELECT account_key, mailbox_name, uidvalidity, highest_uid, last_uidnext, last_sync_at, status
		FROM mailbox_sync_state
		WHERE account_key = ? AND mailbox_name = ?`, accountKey, mailboxName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "load mailbox sync state")
	}
	return &state, nil
}

func resetMailboxState(ctx context.Context, db *sqlx.DB, accountKey, mailboxName string) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin mailbox reset transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM messages WHERE account_key = ? AND mailbox_name = ?`, accountKey, mailboxName); err != nil {
		return errors.Wrap(err, "delete mirrored mailbox messages")
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM mailbox_sync_state WHERE account_key = ? AND mailbox_name = ?`, accountKey, mailboxName); err != nil {
		return errors.Wrap(err, "delete mirrored mailbox sync state")
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit mailbox reset transaction")
	}
	return nil
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func upsertMailboxSyncState(ctx context.Context, executor sqlExecutor, state MailboxSyncState) error {
	query := `INSERT INTO mailbox_sync_state (
		account_key, mailbox_name, uidvalidity, highest_uid, last_uidnext, last_sync_at, status
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(account_key, mailbox_name) DO UPDATE SET
		uidvalidity = excluded.uidvalidity,
		highest_uid = excluded.highest_uid,
		last_uidnext = excluded.last_uidnext,
		last_sync_at = excluded.last_sync_at,
		status = excluded.status`
	if _, err := executor.ExecContext(
		ctx,
		query,
		state.AccountKey,
		state.MailboxName,
		state.UIDValidity,
		state.HighestUID,
		state.LastUIDNext,
		state.LastSyncAt,
		state.Status,
	); err != nil {
		return errors.Wrap(err, "upsert mailbox sync state")
	}
	return nil
}

func upsertMessageRecord(ctx context.Context, tx *sqlx.Tx, record MessageRecord) error {
	query := `INSERT INTO messages (
		account_key, mailbox_name, uidvalidity, uid, message_id, internal_date, sent_date, subject,
		from_summary, to_summary, cc_summary, size_bytes, flags_json, headers_json, parts_json,
		body_text, body_html, search_text, raw_path, raw_sha256, has_attachments, remote_deleted,
		first_seen_at, last_synced_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(account_key, mailbox_name, uidvalidity, uid) DO UPDATE SET
		message_id = excluded.message_id,
		internal_date = excluded.internal_date,
		sent_date = excluded.sent_date,
		subject = excluded.subject,
		from_summary = excluded.from_summary,
		to_summary = excluded.to_summary,
		cc_summary = excluded.cc_summary,
		size_bytes = excluded.size_bytes,
		flags_json = excluded.flags_json,
		headers_json = excluded.headers_json,
		parts_json = excluded.parts_json,
		body_text = excluded.body_text,
		body_html = excluded.body_html,
		search_text = excluded.search_text,
		raw_path = excluded.raw_path,
		raw_sha256 = excluded.raw_sha256,
		has_attachments = excluded.has_attachments,
		remote_deleted = excluded.remote_deleted,
		last_synced_at = excluded.last_synced_at`
	if _, err := tx.ExecContext(
		ctx,
		query,
		record.AccountKey,
		record.MailboxName,
		record.UIDValidity,
		record.UID,
		record.MessageID,
		record.InternalDate,
		record.SentDate,
		record.Subject,
		record.FromSummary,
		record.ToSummary,
		record.CCSummary,
		record.SizeBytes,
		record.FlagsJSON,
		record.HeadersJSON,
		record.PartsJSON,
		record.BodyText,
		record.BodyHTML,
		record.SearchText,
		record.RawPath,
		record.RawSHA256,
		record.HasAttachments,
		record.RemoteDeleted,
		record.FirstSeenAt,
		record.LastSyncedAt,
	); err != nil {
		return errors.Wrap(err, "upsert mirrored message")
	}
	return nil
}
