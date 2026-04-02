package mirror

import (
	"context"
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

	return nil, errors.New("merge execution is not implemented yet; rerun with --dry-run while implementation is in progress")
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
