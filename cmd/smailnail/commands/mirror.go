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
	"github.com/go-go-golems/smailnail/pkg/imap"
	"github.com/go-go-golems/smailnail/pkg/mirror"
)

type MirrorCommand struct {
	*cmds.CommandDescription
}

type MirrorSettings struct {
	SQLitePath            string `glazed:"sqlite-path"`
	MirrorRoot            string `glazed:"mirror-root"`
	BatchSize             int    `glazed:"batch-size"`
	MaxMessages           int    `glazed:"max-messages"`
	SinceDays             int    `glazed:"since-days"`
	AllMailboxes          bool   `glazed:"all-mailboxes"`
	MailboxPattern        string `glazed:"mailbox-pattern"`
	ExcludeMailboxPattern string `glazed:"exclude-mailbox-pattern"`
	PrintPlan             bool   `glazed:"print-plan"`
	ReconcileFull         bool   `glazed:"reconcile-full-mailbox"`
	ResetMailboxState     bool   `glazed:"reset-mailbox-state"`

	imap.IMAPSettings
}

func NewMirrorCommand() (*MirrorCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	imapSection, err := imap.NewIMAPSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP section: %w", err)
	}

	mirrorSection, err := schema.NewSection(
		"mirror",
		"Local Mirror Settings",
		schema.WithFields(
			fields.New(
				"sqlite-path",
				fields.TypeString,
				fields.WithHelp("SQLite path for the local mirror store"),
				fields.WithDefault(mirror.DefaultSQLiteDBPath),
			),
			fields.New(
				"mirror-root",
				fields.TypeString,
				fields.WithHelp("Root directory for local raw message storage"),
				fields.WithDefault(mirror.DefaultMirrorRoot),
			),
			fields.New(
				"batch-size",
				fields.TypeInteger,
				fields.WithHelp("Maximum number of UIDs to fetch in each sync batch"),
				fields.WithDefault(100),
			),
			fields.New(
				"max-messages",
				fields.TypeInteger,
				fields.WithHelp("Maximum number of messages to fetch across the whole sync run (0 means no limit)"),
				fields.WithDefault(0),
			),
			fields.New(
				"since-days",
				fields.TypeInteger,
				fields.WithHelp("Only fetch messages whose IMAP date is within the last N days (0 means no date limit)"),
				fields.WithDefault(0),
			),
			fields.New(
				"all-mailboxes",
				fields.TypeBool,
				fields.WithHelp("Mirror all listed mailboxes instead of only the selected mailbox"),
				fields.WithDefault(false),
			),
			fields.New(
				"mailbox-pattern",
				fields.TypeString,
				fields.WithHelp("Only mirror mailboxes whose names match this glob pattern when --all-mailboxes is enabled"),
			),
			fields.New(
				"exclude-mailbox-pattern",
				fields.TypeString,
				fields.WithHelp("Skip mailboxes whose names match this glob pattern when --all-mailboxes is enabled"),
			),
			fields.New(
				"print-plan",
				fields.TypeBool,
				fields.WithHelp("Print the mirror plan without mutating local storage"),
				fields.WithDefault(false),
			),
			fields.New(
				"reconcile-full-mailbox",
				fields.TypeBool,
				fields.WithHelp("Search the full mailbox after sync and mark missing local rows as remote_deleted"),
				fields.WithDefault(false),
			),
			fields.New(
				"reset-mailbox-state",
				fields.TypeBool,
				fields.WithHelp("Reset stored local mailbox sync state before syncing"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mirror section: %w", err)
	}

	return &MirrorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"mirror",
			cmds.WithShort("Bootstrap and run a local IMAP mirror"),
			cmds.WithLong(`Bootstrap a local mirror workspace for IMAP mail.

This command prepares the local SQLite store and raw-message mirror layout that
later sync phases use for durable mailbox downloads.

Examples:
  smailnail mirror --server imap.example.com --username user --password secret --mailbox INBOX
  smailnail mirror --server imap.example.com --username user --password secret --mailbox INBOX --max-messages 100
  smailnail mirror --server imap.example.com --username user --password secret --mailbox INBOX --since-days 30
  smailnail mirror --all-mailboxes --mailbox-pattern 'Archive/*'
  smailnail mirror --all-mailboxes --sqlite-path ./mail.db --mirror-root ./mail-mirror
  smailnail mirror --mailbox Archive --reconcile-full-mailbox
  smailnail mirror --print-plan`),
			cmds.WithSections(glazedSection, imapSection, mirrorSection),
		),
	}, nil
}

func (c *MirrorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &MirrorSettings{}
	if err := parsedValues.DecodeSectionInto("mirror", settings); err != nil {
		return err
	}
	if err := parsedValues.DecodeSectionInto(imap.IMAPSectionSlug, &settings.IMAPSettings); err != nil {
		return err
	}

	report := mirror.BootstrapReport{
		Database: mirror.DatabaseInfo{
			Driver: "sqlite3",
			Path:   settings.SQLitePath,
		},
		MirrorRoot:            settings.MirrorRoot,
		SearchMode:            mirror.SearchModeFTS5,
		PrintPlan:             settings.PrintPlan,
		SelectedMailbox:       settings.Mailbox,
		AllMailboxes:          settings.AllMailboxes,
		MailboxPattern:        settings.MailboxPattern,
		ExcludeMailboxPattern: settings.ExcludeMailboxPattern,
		BatchSize:             settings.BatchSize,
		MaxMessages:           settings.MaxMessages,
		SinceDays:             settings.SinceDays,
		ReconcileFull:         settings.ReconcileFull,
		ResetState:            settings.ResetMailboxState,
	}

	var syncReport *mirror.SyncReport
	if !settings.PrintPlan {
		store, err := mirror.OpenStore(settings.SQLitePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = store.Close()
		}()

		bootstrapped, err := store.Bootstrap(ctx, settings.MirrorRoot)
		if err != nil {
			return err
		}
		report = *bootstrapped
		report.PrintPlan = settings.PrintPlan
		report.SelectedMailbox = settings.Mailbox
		report.AllMailboxes = settings.AllMailboxes
		report.MailboxPattern = settings.MailboxPattern
		report.ExcludeMailboxPattern = settings.ExcludeMailboxPattern
		report.BatchSize = settings.BatchSize
		report.MaxMessages = settings.MaxMessages
		report.SinceDays = settings.SinceDays
		report.ResetState = settings.ResetMailboxState

		service := mirror.NewService(store)
		syncReport, err = service.Sync(ctx, mirror.SyncOptions{
			Server:                settings.Server,
			Port:                  settings.Port,
			Username:              settings.Username,
			Password:              settings.Password,
			Insecure:              settings.Insecure,
			Mailbox:               settings.Mailbox,
			AllMailboxes:          settings.AllMailboxes,
			MailboxPattern:        settings.MailboxPattern,
			ExcludeMailboxPattern: settings.ExcludeMailboxPattern,
			MirrorRoot:            settings.MirrorRoot,
			BatchSize:             settings.BatchSize,
			MaxMessages:           settings.MaxMessages,
			SinceDays:             settings.SinceDays,
			ReconcileFull:         settings.ReconcileFull,
			ResetMailboxState:     settings.ResetMailboxState,
		})
		if err != nil {
			return err
		}
	}

	row := types.NewRow()
	row.Set("status", statusFromPlan(settings.PrintPlan))
	row.Set("sqlite_driver", report.Database.Driver)
	row.Set("sqlite_path", report.Database.Path)
	row.Set("mirror_root", report.MirrorRoot)
	row.Set("search_mode", report.SearchMode)
	row.Set("fts_available", report.FTSAvailable)
	row.Set("fts_status", report.FTSStatus)
	row.Set("schema_version", report.SchemaVersion)
	row.Set("selected_mailbox", report.SelectedMailbox)
	row.Set("all_mailboxes", report.AllMailboxes)
	row.Set("mailbox_pattern", report.MailboxPattern)
	row.Set("exclude_mailbox_pattern", report.ExcludeMailboxPattern)
	row.Set("batch_size", report.BatchSize)
	row.Set("max_messages", report.MaxMessages)
	row.Set("since_days", report.SinceDays)
	row.Set("reconcile_full_mailbox", report.ReconcileFull)
	row.Set("reset_mailbox_state", report.ResetState)
	if syncReport != nil {
		row.Set("account_key", syncReport.AccountKey)
		row.Set("mailboxes_planned", syncReport.MailboxesPlanned)
		row.Set("mailboxes_synced", syncReport.MailboxesSynced)
		row.Set("max_messages_reached", syncReport.MaxMessagesReached)
		row.Set("messages_fetched", syncReport.MessagesFetched)
		row.Set("messages_stored", syncReport.MessagesStored)
		row.Set("raw_files_written", syncReport.RawFilesWritten)
		row.Set("reused_file_writes", syncReport.ReusedFileWrites)
		row.Set("messages_tombstoned", syncReport.TombstonedMessages)
		row.Set("messages_restored", syncReport.RestoredMessages)
	}

	return gp.AddRow(ctx, row)
}

func statusFromPlan(printPlan bool) string {
	if printPlan {
		return "plan"
	}
	return "synced"
}
