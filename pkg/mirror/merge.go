package mirror

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type MergeOptions struct {
	InputRoot                 string
	OutputSQLitePath          string
	OutputMirrorRoot          string
	ShardGlob                 string
	DryRun                    bool
	CopyRaw                   bool
	FailOnMissingRaw          bool
	RebuildFTS                bool
	RebuildSyncState          bool
	AllowOverwriteDestination bool
	EnrichAfter               bool
}

type MergeShardInfo struct {
	Name           string                  `json:"name"`
	Root           string                  `json:"root"`
	SQLitePath     string                  `json:"sqlitePath"`
	RawRoot        string                  `json:"rawRoot"`
	SchemaVersion  int                     `json:"schemaVersion"`
	MessageCount   int                     `json:"messageCount"`
	AccountKeys    []string                `json:"accountKeys"`
	Mailboxes      []string                `json:"mailboxes"`
	MailboxStates  []MergeShardMailboxInfo `json:"mailboxStates"`
	MissingRawRoot bool                    `json:"missingRawRoot"`
}

type MergeReport struct {
	Status                    string           `json:"status"`
	InputRoot                 string           `json:"inputRoot"`
	OutputSQLitePath          string           `json:"outputSQLitePath"`
	OutputMirrorRoot          string           `json:"outputMirrorRoot"`
	ShardGlob                 string           `json:"shardGlob"`
	DryRun                    bool             `json:"dryRun"`
	CopyRaw                   bool             `json:"copyRaw"`
	FailOnMissingRaw          bool             `json:"failOnMissingRaw"`
	RebuildFTS                bool             `json:"rebuildFTS"`
	RebuildSyncState          bool             `json:"rebuildSyncState"`
	AllowOverwriteDestination bool             `json:"allowOverwriteDestination"`
	EnrichAfter               bool             `json:"enrichAfter"`
	ShardsDiscovered          int              `json:"shardsDiscovered"`
	ShardsMerged              int              `json:"shardsMerged"`
	MessagesScanned           int              `json:"messagesScanned"`
	MessagesInserted          int              `json:"messagesInserted"`
	MessagesUpdated           int              `json:"messagesUpdated"`
	RawFilesCopied            int              `json:"rawFilesCopied"`
	RawFilesReused            int              `json:"rawFilesReused"`
	RawFilesMissing           int              `json:"rawFilesMissing"`
	RawConflicts              int              `json:"rawConflicts"`
	UIDValidityConflicts      int              `json:"uidValidityConflicts"`
	MailboxStatesRebuilt      int              `json:"mailboxStatesRebuilt"`
	FTSRowsRebuilt            int              `json:"ftsRowsRebuilt"`
	Warnings                  []string         `json:"warnings"`
	Shards                    []MergeShardInfo `json:"shards"`
}

type MergeService struct {
	now func() time.Time
}

type MergeShardMailboxInfo struct {
	AccountKey  string `json:"accountKey"`
	MailboxName string `json:"mailboxName"`
	UIDValidity uint32 `json:"uidValidity"`
}

func NewMergeService() *MergeService {
	return &MergeService{
		now: time.Now,
	}
}

func (s *MergeService) Merge(ctx context.Context, opts MergeOptions) (*MergeReport, error) {
	if s == nil {
		return nil, fmt.Errorf("merge service is nil")
	}

	normalized := normalizeMergeOptions(opts)
	if err := validateMergeOptions(normalized); err != nil {
		return nil, err
	}

	shards, err := discoverMergeShards(normalized.InputRoot, normalized.ShardGlob)
	if err != nil {
		return nil, err
	}

	report := &MergeReport{
		Status:                    "plan",
		InputRoot:                 normalized.InputRoot,
		OutputSQLitePath:          normalized.OutputSQLitePath,
		OutputMirrorRoot:          normalized.OutputMirrorRoot,
		ShardGlob:                 normalized.ShardGlob,
		DryRun:                    normalized.DryRun,
		CopyRaw:                   normalized.CopyRaw,
		FailOnMissingRaw:          normalized.FailOnMissingRaw,
		RebuildFTS:                normalized.RebuildFTS,
		RebuildSyncState:          normalized.RebuildSyncState,
		AllowOverwriteDestination: normalized.AllowOverwriteDestination,
		EnrichAfter:               normalized.EnrichAfter,
		Shards:                    make([]MergeShardInfo, 0, len(shards)),
	}

	for _, shard := range shards {
		info, err := inspectMergeShard(ctx, shard)
		if err != nil {
			return nil, errors.Wrapf(err, "inspect shard %s", shard.Name)
		}
		report.Shards = append(report.Shards, info)
		report.ShardsDiscovered++
		report.MessagesScanned += info.MessageCount
		if info.MissingRawRoot {
			report.Warnings = append(report.Warnings, fmt.Sprintf("shard %s is missing raw root %s", info.Name, info.RawRoot))
		}
	}

	if err := validateMergeShards(report.Shards); err != nil {
		return nil, err
	}

	if normalized.DryRun {
		return report, nil
	}

	if !normalized.AllowOverwriteDestination {
		if exists, err := pathExists(normalized.OutputSQLitePath); err != nil {
			return nil, errors.Wrap(err, "check output-sqlite path")
		} else if exists {
			return nil, errors.Errorf("output-sqlite path already exists: %s", normalized.OutputSQLitePath)
		}
		if exists, err := pathExists(normalized.OutputMirrorRoot); err != nil {
			return nil, errors.Wrap(err, "check output-mirror-root path")
		} else if exists {
			return nil, errors.Errorf("output-mirror-root path already exists: %s", normalized.OutputMirrorRoot)
		}
	}

	store, err := OpenStore(normalized.OutputSQLitePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = store.Close()
	}()

	if _, err := store.Bootstrap(ctx, normalized.OutputMirrorRoot); err != nil {
		return nil, err
	}

	for _, shard := range report.Shards {
		if err := s.mergeShard(ctx, store.db, shard, normalized, report); err != nil {
			return nil, errors.Wrapf(err, "merge shard %s", shard.Name)
		}
		report.ShardsMerged++
	}

	if normalized.RebuildSyncState {
		rebuilt, err := rebuildMailboxSyncStates(ctx, store.db, s.now().UTC())
		if err != nil {
			return nil, err
		}
		report.MailboxStatesRebuilt = rebuilt
	}

	if normalized.RebuildFTS {
		reloaded, err := rebuildMessagesFTS(ctx, store.db)
		if err != nil {
			return nil, err
		}
		report.FTSRowsRebuilt = reloaded
	}

	report.Status = "merged"
	return report, nil
}

func normalizeMergeOptions(opts MergeOptions) MergeOptions {
	opts.InputRoot = strings.TrimSpace(opts.InputRoot)
	opts.OutputSQLitePath = strings.TrimSpace(opts.OutputSQLitePath)
	opts.OutputMirrorRoot = strings.TrimSpace(opts.OutputMirrorRoot)
	opts.ShardGlob = strings.TrimSpace(opts.ShardGlob)
	return opts
}

func validateMergeOptions(opts MergeOptions) error {
	if opts.InputRoot == "" {
		return errors.New("input-root is required")
	}
	if opts.OutputSQLitePath == "" {
		return errors.New("output-sqlite is required")
	}
	if opts.OutputMirrorRoot == "" {
		return errors.New("output-mirror-root is required")
	}
	if opts.ShardGlob != "" {
		if _, err := filepath.Match(opts.ShardGlob, "2026-04"); err != nil {
			return errors.Wrap(err, "invalid shard-glob")
		}
	}
	return nil
}

func discoverMergeShards(inputRoot, shardGlob string) ([]MergeShardInfo, error) {
	entries, err := os.ReadDir(inputRoot)
	if err != nil {
		return nil, errors.Wrap(err, "read input-root")
	}

	ret := make([]MergeShardInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if shardGlob != "" {
			matched, err := filepath.Match(shardGlob, name)
			if err != nil {
				return nil, errors.Wrap(err, "match shard-glob")
			}
			if !matched {
				continue
			}
		}

		root := filepath.Join(inputRoot, name)
		sqlitePath := filepath.Join(root, "mirror.sqlite")
		if _, err := os.Stat(sqlitePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, errors.Wrapf(err, "stat shard sqlite %s", sqlitePath)
		}

		ret = append(ret, MergeShardInfo{
			Name:       name,
			Root:       root,
			SQLitePath: sqlitePath,
			RawRoot:    filepath.Join(root, "raw"),
		})
	}

	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	if len(ret) == 0 {
		return nil, errors.Errorf("no mergeable shards found under %s", inputRoot)
	}
	return ret, nil
}

func inspectMergeShard(ctx context.Context, shard MergeShardInfo) (MergeShardInfo, error) {
	db, err := sqlx.Open("sqlite3", shard.SQLitePath)
	if err != nil {
		return MergeShardInfo{}, errors.Wrap(err, "open shard sqlite")
	}
	defer func() {
		_ = db.Close()
	}()

	version, err := schemaVersion(ctx, db)
	if err != nil {
		return MergeShardInfo{}, errors.Wrap(err, "load shard schema version")
	}

	var messageCount int
	if err := db.GetContext(ctx, &messageCount, `SELECT COUNT(*) FROM messages`); err != nil {
		return MergeShardInfo{}, errors.Wrap(err, "count shard messages")
	}

	accountRows := []string{}
	if err := db.SelectContext(ctx, &accountRows, `SELECT DISTINCT account_key FROM messages ORDER BY account_key`); err != nil {
		return MergeShardInfo{}, errors.Wrap(err, "load shard account keys")
	}

	type mailboxUIDValidityRow struct {
		AccountKey  string `db:"account_key"`
		MailboxName string `db:"mailbox_name"`
		UIDValidity uint32 `db:"uidvalidity"`
	}
	rows := []mailboxUIDValidityRow{}
	if err := db.SelectContext(ctx, &rows, `SELECT DISTINCT account_key, mailbox_name, uidvalidity FROM messages ORDER BY account_key, mailbox_name, uidvalidity`); err != nil {
		return MergeShardInfo{}, errors.Wrap(err, "load shard mailbox uidvalidity rows")
	}

	mailboxes := make([]string, 0)
	mailboxStates := make([]MergeShardMailboxInfo, 0, len(rows))
	seenMailboxes := map[string]struct{}{}
	for _, row := range rows {
		if _, ok := seenMailboxes[row.MailboxName]; !ok {
			mailboxes = append(mailboxes, row.MailboxName)
			seenMailboxes[row.MailboxName] = struct{}{}
		}
		mailboxStates = append(mailboxStates, MergeShardMailboxInfo(row))
	}

	missingRawRoot := false
	if _, err := os.Stat(shard.RawRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			missingRawRoot = true
		} else {
			return MergeShardInfo{}, errors.Wrap(err, "stat shard raw root")
		}
	}

	shard.SchemaVersion = version
	shard.MessageCount = messageCount
	shard.AccountKeys = accountRows
	shard.Mailboxes = mailboxes
	shard.MailboxStates = mailboxStates
	shard.MissingRawRoot = missingRawRoot
	return shard, nil
}

func validateMergeShards(shards []MergeShardInfo) error {
	type mailboxKey struct {
		AccountKey  string
		MailboxName string
	}

	seen := map[mailboxKey]uint32{}
	for _, shard := range shards {
		if shard.SchemaVersion > currentSchemaVersion {
			return errors.Errorf("shard %s schema version %d is newer than supported version %d", shard.Name, shard.SchemaVersion, currentSchemaVersion)
		}
		for _, state := range shard.MailboxStates {
			key := mailboxKey{AccountKey: state.AccountKey, MailboxName: state.MailboxName}
			if previous, ok := seen[key]; ok && previous != state.UIDValidity {
				return errors.Errorf(
					"mailbox %s for account %s has conflicting uidvalidity values across shards (%d and %d)",
					state.MailboxName,
					state.AccountKey,
					previous,
					state.UIDValidity,
				)
			}
			seen[key] = state.UIDValidity
		}
	}
	return nil
}

func (s *MergeService) mergeShard(
	ctx context.Context,
	dest *sqlx.DB,
	shard MergeShardInfo,
	opts MergeOptions,
	report *MergeReport,
) error {
	source, err := sqlx.Open("sqlite3", shard.SQLitePath)
	if err != nil {
		return errors.Wrap(err, "open shard sqlite for merge")
	}
	defer func() {
		_ = source.Close()
	}()

	records, err := loadShardMessageRecords(ctx, source)
	if err != nil {
		return err
	}

	tx, err := dest.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin merge transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	for _, record := range records {
		if opts.CopyRaw {
			rawResult, err := ensureMergedRawFile(shard, opts.OutputMirrorRoot, record, opts.FailOnMissingRaw)
			if err != nil {
				return err
			}
			report.RawFilesCopied += rawResult.Copied
			report.RawFilesReused += rawResult.Reused
			report.RawFilesMissing += rawResult.Missing
			report.RawConflicts += rawResult.Conflicts
			report.Warnings = append(report.Warnings, rawResult.Warnings...)
		}

		existed, err := destinationMessageExists(ctx, tx, record)
		if err != nil {
			return err
		}
		if _, err := upsertMessageRecord(ctx, tx, record); err != nil {
			return err
		}
		if existed {
			report.MessagesUpdated++
		} else {
			report.MessagesInserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit merge transaction")
	}
	return nil
}

func loadShardMessageRecords(ctx context.Context, db *sqlx.DB) ([]MessageRecord, error) {
	records := []MessageRecord{}
	if err := db.SelectContext(
		ctx,
		&records,
		`SELECT id, account_key, mailbox_name, uidvalidity, uid, message_id, internal_date, sent_date, subject,
		        from_summary, to_summary, cc_summary, size_bytes, flags_json, headers_json, parts_json,
		        body_text, body_html, search_text, raw_path, raw_sha256, has_attachments, remote_deleted,
		        first_seen_at, last_synced_at
		   FROM messages
		  ORDER BY account_key, mailbox_name, uidvalidity, uid`,
	); err != nil {
		return nil, errors.Wrap(err, "load shard message records")
	}
	return records, nil
}

type rawMergeResult struct {
	Copied    int
	Reused    int
	Missing   int
	Conflicts int
	Warnings  []string
}

func ensureMergedRawFile(shard MergeShardInfo, outputMirrorRoot string, record MessageRecord, failOnMissingRaw bool) (rawMergeResult, error) {
	srcPath, srcRaw, err := loadShardRawBytes(shard, record.RawPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			warning := fmt.Sprintf("missing raw file for shard=%s raw_path=%s", shard.Name, record.RawPath)
			if failOnMissingRaw {
				return rawMergeResult{}, errors.New(warning)
			}
			return rawMergeResult{
				Missing:  1,
				Warnings: []string{warning},
			}, nil
		}
		return rawMergeResult{}, errors.Wrap(err, "read shard raw file")
	}

	sourceSHA := sha256Hex(srcRaw)
	if record.RawSHA256 != "" && !strings.EqualFold(record.RawSHA256, sourceSHA) {
		return rawMergeResult{}, errors.Errorf(
			"raw file sha mismatch for shard=%s path=%s: row=%s file=%s",
			shard.Name,
			srcPath,
			record.RawSHA256,
			sourceSHA,
		)
	}

	destPath := filepath.Join(outputMirrorRoot, record.RawPath)
	if existingRaw, err := os.ReadFile(destPath); err == nil {
		if sha256Hex(existingRaw) != sourceSHA {
			return rawMergeResult{Conflicts: 1}, errors.Errorf(
				"destination raw file conflict for %s",
				destPath,
			)
		}
		return rawMergeResult{Reused: 1}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return rawMergeResult{}, errors.Wrap(err, "read destination raw file")
	}

	rawResult, err := WriteRawMessage(outputMirrorRoot, record.AccountKey, record.MailboxName, record.UIDValidity, record.UID, srcRaw)
	if err != nil {
		return rawMergeResult{}, err
	}
	if rawResult.Path != record.RawPath {
		return rawMergeResult{}, errors.Errorf(
			"raw path mismatch for mailbox=%s uid=%d: expected %s got %s",
			record.MailboxName,
			record.UID,
			record.RawPath,
			rawResult.Path,
		)
	}

	return rawMergeResult{Copied: 1}, nil
}

func loadShardRawBytes(shard MergeShardInfo, rawPath string) (string, []byte, error) {
	var firstMissing string
	for _, candidate := range shardRawCandidates(shard, rawPath) {
		srcRaw, err := os.ReadFile(candidate)
		if err == nil {
			return candidate, srcRaw, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			if firstMissing == "" {
				firstMissing = candidate
			}
			continue
		}
		return "", nil, err
	}
	if firstMissing == "" {
		firstMissing = filepath.Join(shard.Root, filepath.FromSlash(rawPath))
	}
	return firstMissing, nil, os.ErrNotExist
}

func shardRawCandidates(shard MergeShardInfo, rawPath string) []string {
	candidates := []string{}
	seen := map[string]struct{}{}
	add := func(candidate string) {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	relPath := filepath.FromSlash(rawPath)
	add(filepath.Join(shard.Root, relPath))
	if shard.RawRoot != "" {
		add(filepath.Join(shard.RawRoot, relPath))
		rawPrefix := "raw" + string(filepath.Separator)
		if strings.HasPrefix(relPath, rawPrefix) {
			add(filepath.Join(shard.RawRoot, strings.TrimPrefix(relPath, rawPrefix)))
		}
	}

	return candidates
}

func destinationMessageExists(ctx context.Context, tx *sqlx.Tx, record MessageRecord) (bool, error) {
	var count int
	if err := tx.GetContext(
		ctx,
		&count,
		`SELECT COUNT(*) FROM messages
		  WHERE account_key = ? AND mailbox_name = ? AND uidvalidity = ? AND uid = ?`,
		record.AccountKey,
		record.MailboxName,
		record.UIDValidity,
		record.UID,
	); err != nil {
		return false, errors.Wrap(err, "check existing merged message")
	}
	return count > 0, nil
}

func rebuildMailboxSyncStates(ctx context.Context, db *sqlx.DB, now time.Time) (int, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "begin mailbox sync state rebuild transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM mailbox_sync_state`); err != nil {
		return 0, errors.Wrap(err, "clear mailbox sync state")
	}

	type row struct {
		AccountKey  string         `db:"account_key"`
		MailboxName string         `db:"mailbox_name"`
		UIDValidity uint32         `db:"uidvalidity"`
		HighestUID  uint32         `db:"highest_uid"`
		LastSyncAt  sql.NullString `db:"last_sync_at"`
	}
	rows := []row{}
	if err := tx.SelectContext(
		ctx,
		&rows,
		`SELECT account_key, mailbox_name, uidvalidity, MAX(uid) AS highest_uid, MAX(last_synced_at) AS last_sync_at
		   FROM messages
		  GROUP BY account_key, mailbox_name, uidvalidity
		  ORDER BY account_key, mailbox_name`,
	); err != nil {
		return 0, errors.Wrap(err, "query rebuilt mailbox sync states")
	}

	for _, row := range rows {
		var lastSyncAt *time.Time
		if row.LastSyncAt.Valid && strings.TrimSpace(row.LastSyncAt.String) != "" {
			var parsed time.Time
			var err error
			layouts := []string{
				time.RFC3339Nano,
				time.RFC3339,
				"2006-01-02 15:04:05Z07:00",
				"2006-01-02 15:04:05.999999999Z07:00",
				"2006-01-02 15:04:05",
			}
			for _, layout := range layouts {
				parsed, err = time.Parse(layout, row.LastSyncAt.String)
				if err == nil {
					break
				}
			}
			if err != nil {
				return 0, errors.Wrapf(err, "parse rebuilt last_sync_at %q", row.LastSyncAt.String)
			}
			parsedUTC := parsed.UTC()
			lastSyncAt = &parsedUTC
		} else {
			fallback := now
			lastSyncAt = &fallback
		}
		if err := upsertMailboxSyncState(ctx, tx, MailboxSyncState{
			AccountKey:  row.AccountKey,
			MailboxName: row.MailboxName,
			UIDValidity: row.UIDValidity,
			HighestUID:  row.HighestUID,
			LastUIDNext: row.HighestUID + 1,
			LastSyncAt:  lastSyncAt,
			Status:      "active",
		}); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.Wrap(err, "commit mailbox sync state rebuild transaction")
	}
	return len(rows), nil
}

func rebuildMessagesFTS(ctx context.Context, db *sqlx.DB) (int, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "begin fts rebuild transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM messages_fts`); err != nil {
		return 0, errors.Wrap(err, "clear messages_fts")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO messages_fts (
		rowid, account_key, mailbox_name, subject, from_summary, to_summary, cc_summary, body_text, body_html, search_text
	) SELECT
		id, account_key, mailbox_name, subject, from_summary, to_summary, cc_summary, body_text, body_html, search_text
	  FROM messages`); err != nil {
		return 0, errors.Wrap(err, "rebuild messages_fts")
	}

	var count int
	if err := tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM messages_fts`); err != nil {
		return 0, errors.Wrap(err, "count rebuilt messages_fts rows")
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.Wrap(err, "commit fts rebuild transaction")
	}
	return count, nil
}

func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
