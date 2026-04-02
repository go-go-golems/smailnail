# Tasks

## Ticket Setup And Research

- [x] Create the `SMN-20260401-IMAP-MIRROR` ticket workspace
- [x] Create the primary design doc and diary documents
- [x] Map the current CLI, DSL, `mailruntime`, JS module, hosted account, and MCP identity architecture with file-backed evidence
- [x] Compare the options for implementing a mirror verb without adding unnecessary compatibility layers
- [x] Write the intern-oriented analysis and implementation guide with prose, diagrams, pseudocode, API references, and file references
- [x] Validate the ticket with `docmgr doctor`
- [x] Dry-run and complete the reMarkable bundle upload

## Recommended Implementation Phases

## Phase 1: Command Scaffold

- [x] Add `MirrorCommand` in `cmd/smailnail/commands/mirror.go`
- [x] Define `MirrorSettings` with Glazed tags for IMAP, local storage, and sync controls
- [x] Reuse `pkg/imap.NewIMAPSection()` in the new command
- [x] Add mirror-specific flags for `--sqlite-path`, `--mirror-root`, `--batch-size`, `--all-mailboxes`, `--print-plan`, and `--reset-mailbox-state`
- [x] Register the new command in `cmd/smailnail/main.go`
- [x] Add CLI help text and examples in the command long description

## Phase 2: Local Mirror Store Bootstrap

- [x] Create `pkg/mirror/types.go` for store state, reports, and parsed message view structs
- [x] Create `pkg/mirror/store.go` for opening the local SQLite DB with `sqlx`
- [x] Create `pkg/mirror/schema.go` for mirror-specific migrations and version tracking
- [x] Add bootstrap tables for metadata, mailbox sync state, and mirrored messages
- [x] Add indexes for mailbox/UID lookup and basic search fields
- [x] Add FTS5 bootstrap for searchable local mirror metadata
- [x] Require `sqlite_fts5` or `fts5` build tags so mirror builds fail fast when SQLite FTS5 is not compiled in
- [x] Remove the remaining runtime FTS fallback branches and make FTS-backed bootstrap the only supported path

## Phase 3: File Store And Canonical Raw Message Persistence

- [x] Create `pkg/mirror/files.go` for raw message storage helpers
- [x] Define the on-disk layout under `<mirror-root>/raw/<account-key>/<mailbox>/<uidvalidity>/<uid>.eml`
- [x] Add deterministic mailbox/account slugging
- [x] Add SHA-256 calculation for raw message content
- [x] Make raw writes idempotent and crash-safe

## Phase 4: Incremental IMAP Sync Engine

- [x] Create `pkg/mirror/service.go` for orchestration
- [x] Connect with `pkg/mailruntime.Connect`
- [x] Support syncing one mailbox or enumerating all mailboxes
- [x] Select mailboxes read-only during sync
- [x] Load and update per-mailbox sync state keyed by mailbox and `UIDVALIDITY`
- [x] Detect `UIDVALIDITY` resets and reset the local snapshot cleanly
- [x] Search for new UIDs using `pkg/mailruntime.IMAPClient.Search`
- [x] Fetch messages in bounded UID batches
- [x] Persist raw messages and upsert metadata inside transactional batch commits

## Phase 5: MIME Parsing And Searchable Projection

- [x] Create `pkg/mirror/parser.go`
- [x] Parse downloaded RFC 822 messages with `github.com/emersion/go-message/mail`
- [x] Extract headers, message-id, address summaries, plain text, HTML, and attachment presence
- [x] Build a stable `search_text` projection
- [x] Persist parsed data into the SQLite mirror schema
- [x] Make raw RFC 822 parsing the canonical source for stored headers and address summaries

## Phase 6: Reconciliation And Reporting

- [x] Add summary reporting rows from the mirror command
- [x] Report mirrored mailbox count, fetched message count, written files, and DB path
- [x] Add `--print-plan` dry-run behavior
- [x] Add `--reset-mailbox-state` behavior
- [x] Add optional tombstoning for missing remote messages after full scans
- [x] Add a full-mailbox reconciliation mode that marks locally mirrored rows as remotely deleted when the server no longer reports them
- [x] Add root logging flags and progress-oriented mirror sync logs so long-running runs no longer look idle

## Phase 6B: Sync Scope And Safety Controls

- [x] Add `--max-messages` so first syncs can stop after a bounded number of fetched messages
- [x] Add `--since-days` so first syncs can limit IMAP search to recent mail only
- [x] Add `--mailbox-pattern` so `--all-mailboxes` can be narrowed to matching mailbox names
- [x] Add `--exclude-mailbox-pattern` so `--all-mailboxes` can skip noisy mailboxes like Trash or Spam
- [x] Add `--stop-on-error` so multi-mailbox syncs can continue after one mailbox fails when desired
- [x] Extend sync reporting rows to surface the new scope-control settings and partial-error behavior
- [x] Add targeted unit coverage for new message-limit, date-limit, mailbox-filter, and stop-on-error behavior

## Phase 7: Validation And Documentation

- [x] Add unit tests for schema bootstrap and migration upgrades
- [x] Add unit tests for parser behavior and search-text generation
- [x] Add tests for mailbox sync-state transitions and `UIDVALIDITY` reset handling
- [x] Add an end-to-end smoke test against the maintained Docker IMAP fixture
- [x] Update CI, smoke scripts, and operator docs to build `smailnail` with `sqlite_fts5`
- [x] Update `README.md` and `cmd/smailnail/README.md` with mirror usage examples
- [x] Add embedded Glazed help entries for mirror overview, first sync, and maintenance workflows
- [x] Update the embedded mirror help pages and README examples for the new sync-scope and partial-failure flags
- [x] Update the diary after each meaningful implementation slice
- [x] Commit focused slices separately with clear messages
