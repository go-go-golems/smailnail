package commands

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	enrichpkg "github.com/go-go-golems/smailnail/pkg/enrich"
	"github.com/go-go-golems/smailnail/pkg/mirror"
	"github.com/rs/zerolog/log"
)

type MergeMirrorShardsCommand struct {
	*cmds.CommandDescription
}

type MergeMirrorShardsSettings struct {
	InputRoot                 string `glazed:"input-root"`
	OutputSQLitePath          string `glazed:"output-sqlite"`
	OutputMirrorRoot          string `glazed:"output-mirror-root"`
	ShardGlob                 string `glazed:"shard-glob"`
	DryRun                    bool   `glazed:"dry-run"`
	CopyRaw                   bool   `glazed:"copy-raw"`
	FailOnMissingRaw          bool   `glazed:"fail-on-missing-raw"`
	RebuildFTS                bool   `glazed:"rebuild-fts"`
	RebuildSyncState          bool   `glazed:"rebuild-sync-state"`
	AllowOverwriteDestination bool   `glazed:"allow-overwrite-destination"`
	EnrichAfter               bool   `glazed:"enrich-after"`
}

func NewMergeMirrorShardsCommand() (*MergeMirrorShardsCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	mergeSection, err := schema.NewSection(
		"merge-mirror",
		"Merge Mirror Shards",
		schema.WithFields(
			fields.New("input-root", fields.TypeString, fields.WithHelp("Root directory containing shard subdirectories with mirror.sqlite and raw/")),
			fields.New("output-sqlite", fields.TypeString, fields.WithHelp("Destination SQLite path for the merged mirror")),
			fields.New("output-mirror-root", fields.TypeString, fields.WithHelp("Destination mirror root containing merged raw/ files")),
			fields.New("shard-glob", fields.TypeString, fields.WithHelp("Optional glob used to filter shard directory names under --input-root")),
			fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Inspect shards and print the merge plan without mutating the destination")),
			fields.New("copy-raw", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Copy shard raw .eml files into the destination mirror root")),
			fields.New("fail-on-missing-raw", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Fail instead of warning when a shard row points at a missing raw message file")),
			fields.New("rebuild-fts", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Rebuild messages_fts after the merge completes")),
			fields.New("rebuild-sync-state", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Rebuild mailbox_sync_state after the merge completes")),
			fields.New("allow-overwrite-destination", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Allow writing into an existing destination path instead of requiring a fresh output")),
			fields.New("enrich-after", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Run sender, thread, and unsubscribe enrichment after a successful merge")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge section: %w", err)
	}

	return &MergeMirrorShardsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"merge-mirror-shards",
			cmds.WithShort("Merge month-sharded local mirror databases into one destination mirror"),
			cmds.WithLong(`Inspect and merge shard-local mirror databases produced by parallel backfills.

This command is designed for month-sharded mirror roots where each child directory
contains its own mirror.sqlite and raw/ tree.

Examples:
  smailnail merge-mirror-shards --input-root /tmp/backfill --output-sqlite /tmp/merged.sqlite --output-mirror-root /tmp/merged-root --dry-run
  smailnail merge-mirror-shards --input-root /tmp/backfill --output-sqlite /tmp/merged.sqlite --output-mirror-root /tmp/merged-root --shard-glob '2026-*'
  smailnail merge-mirror-shards --input-root /tmp/backfill --output-sqlite /tmp/merged.sqlite --output-mirror-root /tmp/merged-root --enrich-after`),
			cmds.WithSections(glazedSection, mergeSection),
		),
	}, nil
}

func (c *MergeMirrorShardsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &MergeMirrorShardsSettings{}
	if err := parsedValues.DecodeSectionInto("merge-mirror", settings); err != nil {
		return err
	}

	service := mirror.NewMergeService()
	report, err := service.Merge(ctx, mirror.MergeOptions{
		InputRoot:                 settings.InputRoot,
		OutputSQLitePath:          settings.OutputSQLitePath,
		OutputMirrorRoot:          settings.OutputMirrorRoot,
		ShardGlob:                 settings.ShardGlob,
		DryRun:                    settings.DryRun,
		CopyRaw:                   settings.CopyRaw,
		FailOnMissingRaw:          settings.FailOnMissingRaw,
		RebuildFTS:                settings.RebuildFTS,
		RebuildSyncState:          settings.RebuildSyncState,
		AllowOverwriteDestination: settings.AllowOverwriteDestination,
		EnrichAfter:               settings.EnrichAfter,
	})
	if err != nil {
		return err
	}

	var enrichReport *enrichpkg.AllReport
	if !settings.DryRun && settings.EnrichAfter {
		enrichment, err := enrichpkg.RunAll(ctx, settings.OutputSQLitePath, enrichpkg.Options{})
		if err != nil {
			log.Warn().Err(err).Str("sqlite_path", settings.OutputSQLitePath).Msg("post-merge enrichment failed")
		} else {
			enrichReport = &enrichment
		}
	}

	row := types.NewRow()
	row.Set("status", report.Status)
	row.Set("input_root", report.InputRoot)
	row.Set("output_sqlite_path", report.OutputSQLitePath)
	row.Set("output_mirror_root", report.OutputMirrorRoot)
	row.Set("shard_glob", report.ShardGlob)
	row.Set("dry_run", report.DryRun)
	row.Set("copy_raw", report.CopyRaw)
	row.Set("fail_on_missing_raw", report.FailOnMissingRaw)
	row.Set("rebuild_fts", report.RebuildFTS)
	row.Set("rebuild_sync_state", report.RebuildSyncState)
	row.Set("allow_overwrite_destination", report.AllowOverwriteDestination)
	row.Set("enrich_after", report.EnrichAfter)
	row.Set("shards_discovered", report.ShardsDiscovered)
	row.Set("shards_merged", report.ShardsMerged)
	row.Set("messages_scanned", report.MessagesScanned)
	row.Set("messages_inserted", report.MessagesInserted)
	row.Set("messages_updated", report.MessagesUpdated)
	row.Set("raw_files_copied", report.RawFilesCopied)
	row.Set("raw_files_reused", report.RawFilesReused)
	row.Set("raw_files_missing", report.RawFilesMissing)
	row.Set("raw_conflicts", report.RawConflicts)
	row.Set("uidvalidity_conflicts", report.UIDValidityConflicts)
	row.Set("mailbox_states_rebuilt", report.MailboxStatesRebuilt)
	row.Set("fts_rows_rebuilt", report.FTSRowsRebuilt)
	row.Set("warnings", report.Warnings)
	row.Set("shards", report.Shards)

	if enrichReport != nil {
		row.Set("enrich_senders_created", enrichReport.Senders.SendersCreated)
		row.Set("enrich_senders_updated", enrichReport.Senders.SendersUpdated)
		row.Set("enrich_messages_tagged", enrichReport.Senders.MessagesTagged)
		row.Set("enrich_threads_processed", enrichReport.Threads.MessagesProcessed)
		row.Set("enrich_threads_created", enrichReport.Threads.ThreadsCreated)
		row.Set("enrich_threads_updated", enrichReport.Threads.ThreadsUpdated)
		row.Set("enrich_senders_with_unsubscribe", enrichReport.Unsubscribe.SendersWithUnsubscribe)
	}

	return gp.AddRow(ctx, row)
}
