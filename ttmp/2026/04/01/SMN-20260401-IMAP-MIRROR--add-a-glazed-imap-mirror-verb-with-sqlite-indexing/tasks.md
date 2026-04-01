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
- [x] Add mirror-specific flags for `--sqlite-path`, `--mirror-root`, `--batch-size`, `--all-mailboxes`, `--search-mode`, `--print-plan`, and `--reset-mailbox-state`
- [x] Register the new command in `cmd/smailnail/main.go`
- [x] Add CLI help text and examples in the command long description

## Phase 2: Local Mirror Store Bootstrap

- [x] Create `pkg/mirror/types.go` for store state, reports, and parsed message view structs
- [x] Create `pkg/mirror/store.go` for opening the local SQLite DB with `sqlx`
- [x] Create `pkg/mirror/schema.go` for mirror-specific migrations and version tracking
- [x] Add bootstrap tables for metadata, mailbox sync state, and mirrored messages
- [x] Add indexes for mailbox/UID lookup and basic search fields
- [x] Add optional FTS5 bootstrap with graceful fallback metadata

## Phase 3: File Store And Canonical Raw Message Persistence

- [ ] Create `pkg/mirror/files.go` for raw message storage helpers
- [ ] Define the on-disk layout under `<mirror-root>/raw/<account-key>/<mailbox>/<uidvalidity>/<uid>.eml`
- [ ] Add deterministic mailbox/account slugging
- [ ] Add SHA-256 calculation for raw message content
- [ ] Make raw writes idempotent and crash-safe

## Phase 4: Incremental IMAP Sync Engine

- [ ] Create `pkg/mirror/service.go` for orchestration
- [ ] Connect with `pkg/mailruntime.Connect`
- [ ] Support syncing one mailbox or enumerating all mailboxes
- [ ] Select mailboxes read-only during sync
- [ ] Load and update per-mailbox sync state keyed by mailbox and `UIDVALIDITY`
- [ ] Detect `UIDVALIDITY` resets and reset the local snapshot cleanly
- [ ] Search for new UIDs using `pkg/mailruntime.IMAPClient.Search`
- [ ] Fetch messages in bounded UID batches
- [ ] Persist raw messages and upsert metadata inside transactional batch commits

## Phase 5: MIME Parsing And Searchable Projection

- [ ] Create `pkg/mirror/parser.go`
- [ ] Parse downloaded RFC 822 messages with `github.com/emersion/go-message/mail`
- [ ] Extract headers, message-id, address summaries, plain text, HTML, and attachment presence
- [ ] Build a stable `search_text` projection
- [ ] Persist parsed data into the SQLite mirror schema
- [ ] Add fallback basic search support even when FTS5 is unavailable

## Phase 6: Reconciliation And Reporting

- [ ] Add summary reporting rows from the mirror command
- [ ] Report mirrored mailbox count, fetched message count, written files, and DB path
- [x] Add `--print-plan` dry-run behavior
- [ ] Add `--reset-mailbox-state` behavior
- [ ] Add optional tombstoning for missing remote messages after full scans

## Phase 7: Validation And Documentation

- [ ] Add unit tests for schema bootstrap and migration upgrades
- [ ] Add unit tests for parser behavior and search-text generation
- [ ] Add tests for mailbox sync-state transitions and `UIDVALIDITY` reset handling
- [ ] Add an end-to-end smoke test against the maintained Docker IMAP fixture
- [ ] Update `README.md` and `cmd/smailnail/README.md` with mirror usage examples
- [ ] Update the diary after each meaningful implementation slice
- [ ] Commit focused slices separately with clear messages
