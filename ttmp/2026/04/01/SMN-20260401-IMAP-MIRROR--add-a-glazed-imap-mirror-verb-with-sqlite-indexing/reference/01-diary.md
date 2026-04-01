---
Title: Diary
Ticket: SMN-20260401-IMAP-MIRROR
Status: active
Topics:
    - imap
    - sqlite
    - glazed
    - email
    - cli
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: smailnail/cmd/smailnail/commands/fetch_mail.go
      Note: |-
        Current CLI command that shaped the recommended Glazed command structure
        Existing CLI command that shaped the recommended Glazed structure
    - Path: smailnail/cmd/smailnail/commands/mirror.go
      Note: |-
        Implemented the mirror Glazed command in commit 1d9578a08372607e77e4de17bb95a1b75522568d
        Extended the mirror command to run sync and report aggregate counts in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
        Removed the dead runtime search-mode split in commit 215920ddf1ec71cbee377ff6624615e861a1acf8
        Added full-mailbox reconcile and tombstone reporting in commit f0aa4292d39d1da6240f2ec66ef068e28a7ae534
    - Path: smailnail/cmd/smailnail/main.go
      Note: Registered the new mirror command in commit 1d9578a08372607e77e4de17bb95a1b75522568d
    - Path: smailnail/docker-compose.local.yml
      Note: Local Docker Dovecot fixture used for the real mirror smoke validation
    - Path: smailnail/pkg/mailruntime/imap_client.go
      Note: |-
        Existing UID-based IMAP runtime identified as the recommended sync foundation
        Runtime identified as the best sync foundation during research
        Fixed insecure TLS support required by the Docker Dovecot mirror smoke in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
    - Path: smailnail/pkg/mirror/files.go
      Note: Added raw-message pathing
    - Path: smailnail/pkg/mirror/parser.go
      Note: |-
        Added raw-message MIME parsing and search-text projection in commit f30a4c432200b77456cb116f4443477c4d8759e3
        Added normalized parsed-header projection in commit bb97160ae5d9bd89af0233f2bf9bda6ba46fc2af
    - Path: smailnail/pkg/mirror/parser_test.go
      Note: |-
        Added parser coverage and record-projection tests in commit f30a4c432200b77456cb116f4443477c4d8759e3
        Extended parser coverage for normalized parsed headers in commit bb97160ae5d9bd89af0233f2bf9bda6ba46fc2af
    - Path: smailnail/pkg/mirror/require_fts5_build_tag.go
      Note: Added compile-time SQLite FTS5 build-tag enforcement in commit d2bed23557ada03540fbf4fc4e1f393df9fdfcbb
    - Path: smailnail/pkg/mirror/schema.go
      Note: |-
        Added mirror schema bootstrap and FTS detection in commit 1d9578a08372607e77e4de17bb95a1b75522568d
        Simplified bootstrap to an unconditional FTS-backed path in commit 215920ddf1ec71cbee377ff6624615e861a1acf8
    - Path: smailnail/pkg/mirror/service.go
      Note: |-
        Added incremental IMAP sync orchestration in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
        Switched mirrored rows to prefer normalized parsed headers in commit bb97160ae5d9bd89af0233f2bf9bda6ba46fc2af
        Added full-mailbox reconcile and `remote_deleted` updates in commit f0aa4292d39d1da6240f2ec66ef068e28a7ae534
    - Path: smailnail/pkg/mirror/service_test.go
      Note: |-
        Added incremental sync and UIDVALIDITY regression tests in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
        Added reconcile tombstone and restore tests in commit f0aa4292d39d1da6240f2ec66ef068e28a7ae534
    - Path: smailnail/pkg/mirror/store.go
      Note: |-
        Added local mirror store bootstrap in commit 1d9578a08372607e77e4de17bb95a1b75522568d
        Removed runtime search-mode selection from store bootstrap in commit 215920ddf1ec71cbee377ff6624615e861a1acf8
    - Path: smailnail/pkg/mirror/store_test.go
      Note: Added initial mirror schema tests in commit 1d9578a08372607e77e4de17bb95a1b75522568d
    - Path: smailnail/Makefile
      Note: Updated tagged build, test, lint, and install targets in commit d2bed23557ada03540fbf4fc4e1f393df9fdfcbb
    - Path: smailnail/scripts/docker-imap-smoke.sh
      Note: Updated the Docker IMAP smoke to invoke `smailnail` with the required SQLite FTS5 build tag in commit d2bed23557ada03540fbf4fc4e1f393df9fdfcbb
    - Path: smailnail/pkg/smailnaild/accounts/service.go
      Note: |-
        Existing hosted account and mailbox preview flow reviewed for integration boundaries
        Hosted account and preview flow reviewed during the analysis
    - Path: smailnail/pkg/smailnaild/db.go
      Note: |-
        Existing SQLite migration pattern reviewed for reuse
        SQLite migration pattern examined during the investigation
ExternalSources: []
Summary: Chronological research diary for the IMAP mirror and SQLite indexing design ticket.
LastUpdated: 2026-04-01T16:37:00-04:00
WhatFor: Record what was investigated, what decisions were made, what failed, and how to review the resulting design docs.
WhenToUse: Use this diary when continuing the ticket or reviewing how the design was assembled.
---





# Diary

## Goal

Capture the research and documentation work behind the IMAP mirror and SQLite indexing design so a later engineer can see the exact investigation path, decisions, commands, and review trail.

## Step 1: Create The Ticket And Map The Architecture

This step established the ticket scaffold, gathered evidence from the codebase, and turned that evidence into a concrete design recommendation. The work stayed intentionally documentation-first: no code was added to the runtime yet, because the goal of this turn was to create a detailed intern-oriented analysis and implementation guide.

The most important conclusion from this step is that the requested mirror/index feature should be a new local CLI subsystem built on `pkg/mailruntime`, not an extension of the current `fetch-mail` preview flow and not new tables mixed into the hosted application schema. That conclusion came directly from reading the existing command wiring, IMAP runtime, hosted SQLite schema, account services, and MCP identity resolution code.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket with docmgr and analyze smailnail because I want to add a verb (with glazed, building upon the existing functionality as much as possible) to mirror / download IMAP servers. Part of that mirroring is to also import the mails into sqlite to make them easily searchable. 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

Consider all possible options and designs and improvements and things that can be made."

**Assistant interpretation:** Create a new `docmgr` ticket, perform an evidence-backed architecture review of `smailnail`, write a detailed intern-facing design and implementation guide for a new IMAP mirroring/search-index feature, keep the work documented in a diary, and deliver the bundle to reMarkable.

**Inferred user intent:** The user wants implementation-ready design work before coding, with enough depth that an intern could build the feature safely and with a strong understanding of the existing system.

### What I did

- Created the new ticket workspace:
  - `docmgr ticket create-ticket --ticket SMN-20260401-IMAP-MIRROR --title "Add a glazed IMAP mirror verb with SQLite indexing" --topics imap,sqlite,glazed,email,cli`
- Added the primary design document:
  - `docmgr doc add --ticket SMN-20260401-IMAP-MIRROR --doc-type design-doc --title "Intern guide: designing an IMAP mirror verb with SQLite indexing"`
- Added the diary document from the correct workspace root after normalizing the `docmgr` cwd issue:
  - `docmgr doc add --ticket SMN-20260401-IMAP-MIRROR --doc-type reference --title "Diary"`
- Read and analyzed the main code surfaces relevant to the feature:
  - `cmd/smailnail/main.go`
  - `cmd/smailnail/commands/fetch_mail.go`
  - `cmd/smailnail/commands/mail_rules.go`
  - `pkg/imap/layer.go`
  - `pkg/dsl/processor.go`
  - `pkg/dsl/actions.go`
  - `pkg/mailruntime/imap_client.go`
  - `pkg/services/smailnailjs/service.go`
  - `pkg/js/modules/smailnail/module.go`
  - `pkg/smailnaild/db.go`
  - `pkg/smailnaild/accounts/service.go`
  - `pkg/mcp/imapjs/identity_middleware.go`
  - `pkg/smailnaild/http.go`
- Verified the repository dependency set and existing SQLite usage through:
  - `smailnail/go.mod`
  - `pkg/smailnaild/db_test.go`
- Wrote and updated the ticket artifacts:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - the main design guide
  - this diary

### Why

- The user explicitly asked for a new ticket and a detailed design guide before implementation.
- The design depended on understanding whether the current codebase already had:
  - a reusable Glazed command pattern,
  - a durable IMAP runtime,
  - an existing SQLite message schema,
  - a safe account-resolution path.
- The answer turned out to be mixed:
  - yes for Glazed and IMAP runtime,
  - no for a dedicated local message-index schema.

### What worked

- `docmgr status --summary-only` immediately showed the correct workspace root and confirmed this repository is already configured for ticketed documentation work.
- The repository structure was clear enough to map the relevant surfaces quickly with `rg --files smailnail` and targeted `sed`/`nl -ba` reads.
- The newer `pkg/mailruntime` layer was exactly the kind of reusable building block the feature needs. It already exposes UID-oriented search/fetch and raw-message operations.
- The hosted backend code provided a useful reference for SQLite schema bootstrapping and stored-account boundaries without forcing the mirror design to reuse the hosted schema itself.

### What didn't work

- One `docmgr` command initially failed because I ran it from a different cwd than the earlier ticket-creation command:

```text
Command: docmgr doc add --ticket SMN-20260401-IMAP-MIRROR --doc-type reference --title "Diary"
Error: failed to find ticket directory: ticket not found: SMN-20260401-IMAP-MIRROR
```

- The underlying cause was workspace-config selection. Running `docmgr` from `/home/manuel/workspaces/2026-04-01/smailnail-sqlite` picked up the correct root config consistently, so I normalized all subsequent `docmgr` work there.

### What I learned

- `smailnail` currently has three distinct IMAP-related layers:
  - a direct CLI/DSL fetch path,
  - a richer UID-based `mailruntime` path,
  - a hosted stored-account path.
- The CLI path is Glazed-native already, so the new verb should fit cleanly into the existing command registration model.
- The only durable SQLite schema in the repository today is the hosted application schema. There is no local message-mirror schema yet.
- The repo already depends on `go-message`, but it does not yet contain a dedicated reusable inbound MIME parser for downloaded raw messages.

### What was tricky to build

- The main tricky part was not technical implementation but architecture separation. There were three superficially plausible directions:
  - stretch `fetch-mail`,
  - route the feature through the JS module,
  - or add a new local subsystem.

The first two look cheaper at a glance, but both blur important boundaries. `fetch-mail` is a preview/output command, and the JS module is an integration surface, not the clearest implementation substrate for a durable local mirror. The evidence pass through `pkg/dsl/processor.go` and `pkg/mailruntime/imap_client.go` made the difference clear: the DSL path is sequence-number and output oriented, while `mailruntime` is UID oriented and therefore a better sync foundation.

### What warrants a second pair of eyes

- Whether the first implementation should support `--account-id` and hosted stored-account reuse, or stay credential-flag based in v1.
- Whether FTS5 should be `auto`, `required`, or deferred.
- Whether the canonical local artifact should be raw `.eml` files plus SQLite metadata, or a pure SQLite/BLOB design.
- Whether full-mailbox reconcile and tombstoning should ship in v1 or immediately after.

### What should be done in the future

- Run `docmgr doctor --ticket SMN-20260401-IMAP-MIRROR --stale-after 30`.
- Relate the key source files to the design doc and diary using `docmgr doc relate`.
- Dry-run and then complete the reMarkable bundle upload.
- If implementation starts immediately after review, follow the phases in the design guide:
  - new command,
  - store bootstrap,
  - UID-based sync loop,
  - raw-message persistence,
  - parser/import,
  - search and tests.

### Code review instructions

- Start with the design guide in `design-doc/01-intern-guide-designing-an-imap-mirror-verb-with-sqlite-indexing.md`.
- Check the core evidence anchors:
  - `cmd/smailnail/main.go`
  - `cmd/smailnail/commands/fetch_mail.go`
  - `pkg/dsl/processor.go`
  - `pkg/mailruntime/imap_client.go`
  - `pkg/smailnaild/db.go`
  - `pkg/smailnaild/accounts/service.go`
- Review the option analysis carefully. The most important review question is whether the recommended architecture draws the right line between local mirror state and hosted app state.
- Validate the documentation with:
  - `docmgr status --summary-only`
  - `docmgr doctor --ticket SMN-20260401-IMAP-MIRROR --stale-after 30`

### Technical details

- Evidence-gathering commands used in this step included:
  - `docmgr status --summary-only`
  - `docmgr ticket list`
  - `rg --files smailnail`
  - `rg -n "sqlite|database/sql|go-sqlite3|go-imap|go-message" smailnail -S`
  - `nl -ba <file> | sed -n '<range>'`
- The design guide also references external package documentation for:
  - `github.com/emersion/go-imap/v2/imapclient`
  - `github.com/emersion/go-message/mail`
  - SQLite FTS5

## Related

- [Design Guide](../design-doc/01-intern-guide-designing-an-imap-mirror-verb-with-sqlite-indexing.md)

## Step 2: Break The Design Into Implementation Tasks

After the design doc was in place, the next step was to turn the recommendation into an execution checklist that could actually drive implementation. The existing task list still described the project at a phase level, but not yet at the granularity needed to work step by step and commit focused slices.

I rewrote the implementation section of the task list into seven concrete phases: command scaffold, mirror-store bootstrap, raw-file persistence, incremental sync, MIME parsing/search projection, reconciliation/reporting, and validation/documentation. That gives the implementation work a stable ordering and makes it much easier to keep the diary and commits aligned with the ticket.

### Prompt Context

**User prompt (verbatim):** "alright, create detailed tasks in the ticket, and then work on them one by one, committing at appropriate intervals, keeping a diary."

**Assistant interpretation:** Expand the ticket task list into finer-grained implementation steps, then start executing those steps with focused commits and diary maintenance.

**Inferred user intent:** The user wants the design turned into an actual engineering work plan and wants implementation to proceed in disciplined, reviewable slices.

### What I did

- Expanded `tasks.md` from a short phase list into a detailed execution checklist.
- Broke the work into seven phases and added concrete file/package tasks under each phase.
- Ordered the tasks so the first implementation slice is command scaffold plus local store bootstrap, which is a clean first commit boundary.

### Why

- A detailed task list makes it possible to work one slice at a time without losing the larger architecture.
- It also creates natural commit boundaries for the implementation work the user asked for.

### What worked

- The design guide already had enough structure that the task breakdown could follow it directly.
- The existing repository boundaries made the package/file split fairly obvious.

### What didn't work

- N/A

### What I learned

- The most useful task split is not by “feature idea” but by implementation boundary:
  - command wiring,
  - store bootstrap,
  - sync loop,
  - parsing/import,
  - search/reporting/tests.

### What was tricky to build

- The main judgment call was how much to specify up front versus letting implementation discover the shape. I chose to lock down the package/file ownership and functional ordering, but leave some lower-level schema details flexible inside the implementation phases.

### What warrants a second pair of eyes

- Whether the Phase 2 and Phase 3 split should remain separate, or whether store/bootstrap and file persistence should land in one commit.

### What should be done in the future

- Commit the task-list and diary update as its own docs-focused change.
- Start Phase 1 and Phase 2 implementation immediately after that commit.

### Code review instructions

- Review [tasks.md](../tasks.md) and confirm the ordering makes sense for implementation and review.
- Ensure the task list still matches the architecture described in the design guide.

### Technical details

- Files changed in this step:
  - `tasks.md`
  - `reference/01-diary.md`

## Step 3: Land The Mirror Command And Store Bootstrap

This step turned the ticket from design-only into executable code. I implemented the first two planned phases together because they form one clean boundary: a new Glazed `mirror` verb that can print its plan or bootstrap a local mirror database and file root, without yet attempting any IMAP synchronization.

The resulting slice is intentionally narrow but production-shaped. The command is registered in the real CLI, the local store can create its schema and metadata, FTS5 capability is detected with graceful fallback, and the ticket now has a concrete starting point for the next slice: incremental mailbox sync built on `pkg/mailruntime`.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Start implementing the mirror feature in small, reviewable slices, commit each slice cleanly, and keep the ticket diary aligned with the code.

**Inferred user intent:** The user wants the planned IMAP mirror built incrementally, with disciplined commit boundaries and enough written traceability that later work can continue without re-discovery.

**Commit (code):** `1d9578a08372607e77e4de17bb95a1b75522568d` — `feat(mirror): scaffold local mirror store`

### What I did

- Added the new mirror command in `cmd/smailnail/commands/mirror.go`.
- Defined `MirrorSettings` with Glazed tags for:
  - IMAP connection settings via the reused IMAP section
  - local SQLite path
  - local mirror root
  - batch size
  - search mode
  - plan/reset/all-mailboxes controls
- Registered the command in `cmd/smailnail/main.go`.
- Added the first local mirror package files:
  - `pkg/mirror/types.go`
  - `pkg/mirror/schema.go`
  - `pkg/mirror/store.go`
  - `pkg/mirror/store_test.go`
- Implemented schema bootstrap for:
  - `mirror_metadata`
  - `mailbox_sync_state`
  - `messages`
- Added core indexes for mailbox/UID and date/message-id lookup.
- Added FTS5 bootstrap detection with `basic`, `fts-auto`, and `fts-required` behavior.
- Added a command dry-run path with `--print-plan`.
- Validated the slice with:
  - `go test ./cmd/smailnail ./pkg/mirror`
  - `go run ./cmd/smailnail mirror --print-plan --output json`
  - `go run ./cmd/smailnail mirror --sqlite-path "$db" --mirror-root "$root" --output json`
- Committed the code after the repository pre-commit hooks completed successfully.

### Why

- The command scaffold and store bootstrap are the minimum useful foundation for all later phases.
- Landing them first keeps the next sync implementation focused on IMAP behavior instead of mixing schema/CLI churn into the same commit.
- Reusing the existing Glazed and IMAP-section patterns keeps the new verb consistent with the rest of the CLI.

### What worked

- The mirror command integrated cleanly with the existing `cmd/smailnail/main.go` command registration flow.
- The local SQLite bootstrap worked both in unit tests and in a real `go run` invocation against a temporary directory.
- FTS5 detection degraded cleanly to `fts_status = "unavailable"` during runtime validation, which matches the design goal for portable local mirrors.
- The repo pre-commit pipeline passed, including full `go test ./...` and `golangci-lint run -v`.

### What didn't work

- My first attempt to inspect the IMAP/runtime layout used the wrong guessed filenames:

```text
sed: can't read /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/imap/imap.go: No such file or directory
sed: can't read /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mailruntime/runtime.go: No such file or directory
sed: can't read /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mailruntime/store.go: No such file or directory
```

- The fix was straightforward: use `rg --files smailnail/pkg/mailruntime` and read the real files:
  - `pkg/imap/layer.go`
  - `pkg/mailruntime/imap_client.go`
  - `pkg/mailruntime/types.go`

### What I learned

- `pkg/mailruntime.IMAPClient` already has the right primitives for the next phase:
  - `List`
  - `Status`
  - `SelectMailbox`
  - `Search`
  - `Fetch`
  - `FetchRaw`
- The first mirror slice did not need to talk to IMAP at all. Keeping that separation made it easier to validate the local schema and CLI behavior independently.
- FTS5 availability cannot be assumed, so persisting the capability result in mirror metadata is useful for later reporting and search-path selection.

### What was tricky to build

- The main tricky part was choosing the exact boundary for the first commit. It would have been easy to immediately start mixing in sync logic, but that would have tangled together command-shape decisions, schema design, filesystem layout, and IMAP behavior. I deliberately stopped at a locally verifiable boundary: a command that can either describe its plan or bootstrap durable local state.

- The other sharp edge was deciding how strict the FTS behavior should be. The implementation keeps three explicit modes instead of silently doing “best effort” all the time. That makes later operational behavior clearer: local search can remain functional in basic mode, but `fts-required` can fail fast when someone explicitly depends on FTS-backed indexing.

### What warrants a second pair of eyes

- Whether the initial `messages` table fields are the right minimal set before the parser/import phase expands them.
- Whether the mirror command should already return a richer bootstrap report schema, or wait until mailbox sync exists.
- Whether FTS5 metadata should track more than availability/status in v1.

### What should be done in the future

- Implement Phase 3 and Phase 4:
  - raw `.eml` file persistence
  - mailbox/account slugging
  - incremental UID-based sync using `pkg/mailruntime`
- Add transaction-scoped upserts so each fetch batch lands atomically in both filesystem and SQLite metadata.
- Expand reporting once real mailbox sync statistics exist.

### Code review instructions

- Start with:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/store.go`
  - `pkg/mirror/schema.go`
- Then check wiring in:
  - `cmd/smailnail/main.go`
- Validate with:
  - `go test ./cmd/smailnail ./pkg/mirror`
  - `go run ./cmd/smailnail mirror --print-plan --output json`
  - `tmpdir=$(mktemp -d) && db="$tmpdir/mirror.sqlite" && root="$tmpdir/root" && go run ./cmd/smailnail mirror --sqlite-path "$db" --mirror-root "$root" --output json`

### Technical details

- Runtime validation output for `--print-plan`:

```json
[
  {
    "all_mailboxes": false,
    "batch_size": 100,
    "fts_available": false,
    "fts_status": "",
    "mirror_root": "smailnail-mirror",
    "reset_mailbox_state": false,
    "schema_version": 0,
    "search_mode": "fts-auto",
    "selected_mailbox": "INBOX",
    "sqlite_driver": "sqlite3",
    "sqlite_path": "smailnail-mirror.sqlite",
    "status": "plan"
  }
]
```

- Runtime validation output for actual bootstrap:

```json
[
  {
    "all_mailboxes": false,
    "batch_size": 100,
    "fts_available": false,
    "fts_status": "unavailable",
    "mirror_root": "/tmp/tmp.K3SGSSmFjk/root",
    "reset_mailbox_state": false,
    "schema_version": 1,
    "search_mode": "fts-auto",
    "selected_mailbox": "INBOX",
    "sqlite_driver": "sqlite3",
    "sqlite_path": "/tmp/tmp.K3SGSSmFjk/mirror.sqlite",
    "status": "bootstrapped"
  }
]
```

- Files changed in the code commit:
  - `cmd/smailnail/main.go`
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/types.go`
  - `pkg/mirror/schema.go`
  - `pkg/mirror/store.go`
  - `pkg/mirror/store_test.go`

## Step 4: Add Incremental Raw-Message Sync And Verify It Against Docker Dovecot

This step added the first real mirroring behavior: connect to IMAP through `pkg/mailruntime`, discover new UIDs, fetch raw messages in bounded batches, persist `.eml` files under the mirror root, and upsert a durable mailbox/message snapshot into SQLite. It also tightened the command output so the CLI now reports account key, mailbox counts, message counts, and raw-file write counts after a sync.

The most important part of this step was not just getting the sync loop to work in unit tests, but validating it against the Docker Compose Dovecot fixture the repo already uses elsewhere. That real-server check surfaced one subtle incremental-search bug that the unit tests did not catch at first, and fixing that bug before commit substantially increased confidence in the current mirror foundation.

### Prompt Context

**User prompt (verbatim):** "alright, create detailed tasks in the ticket, and then work on them one by one, committing at appropriate intervals, keeping a diary."

**Assistant interpretation:** Continue the mirror implementation in focused slices, keep the ticket diary current, and use the task list as the execution order.

**Inferred user intent:** The user wants steady forward implementation with commit hygiene, documented reasoning, and enough verification that the work is trustworthy before moving to the next phase.

**Commit (code):** `9b0afe7a06542be44f8ae87f397c446232ec8efb` — `feat(mirror): add incremental raw-message sync`

### What I did

- Added `pkg/mirror/files.go`:
  - deterministic account/mailbox slugging
  - raw mirror path generation under `raw/<account-key>/<mailbox-slug>/<uidvalidity>/<uid>.eml`
  - SHA-256 hashing
  - idempotent atomic raw writes through temp-file + rename
- Added `pkg/mirror/service.go`:
  - `SyncOptions`
  - `Service`
  - mailbox enumeration for single-mailbox and `--all-mailboxes`
  - per-mailbox state loading and updates
  - `UIDVALIDITY` reset handling
  - bounded UID batching
  - transactional message/state upserts
- Extended `cmd/smailnail/commands/mirror.go` so the command now:
  - bootstraps the local store
  - runs sync
  - reports mailbox/message/raw-file counters
- Extended `pkg/mirror/types.go` with sync/result structs.
- Added tests:
  - `pkg/mirror/files_test.go`
  - `pkg/mirror/service_test.go`
- Fixed `pkg/mailruntime.Connect` and `pkg/services/smailnailjs/service.go` so `insecure` means TLS with skipped verification, which is required for the local self-signed Dovecot fixture.
- Verified locally with:
  - `go test ./pkg/mirror ./cmd/smailnail ./pkg/mailruntime ./pkg/services/smailnailjs`
  - repo pre-commit `go test ./...`
  - repo pre-commit `golangci-lint run -v`
- Verified against the Docker Compose IMAP fixture with:
  - `docker compose -f docker-compose.local.yml up -d dovecot`
  - `go run ./cmd/imap-tests store-text-message ...`
  - `go run ./cmd/smailnail mirror ...`
  - `docker compose -f docker-compose.local.yml down`

### Why

- Raw message persistence and incremental UID sync are the first point where the feature becomes materially useful.
- The design specifically called for building on `pkg/mailruntime`, so this slice turns that recommendation into real code rather than creating a separate IMAP stack.
- Verifying against Docker Dovecot was important because IMAP range semantics and self-signed TLS behavior are easy to get subtly wrong even when unit tests are green.

### What worked

- The raw file layout landed cleanly and produces deterministic paths keyed by account, mailbox, `UIDVALIDITY`, and UID.
- The SQLite state/model updates worked in unit tests for:
  - first sync
  - incremental second sync
  - `UIDVALIDITY` reset replacement
- The local Dovecot smoke succeeded end to end:
  - first mirror run imported the existing mailbox snapshot
  - second run performed a true no-op
  - after appending one more fixture message, the next run fetched exactly one new message
- The insecure TLS fix made the `pkg/mailruntime` path compatible with the local test fixture instead of forcing plain IMAP semantics.

### What didn't work

- My first attempt to seed the fixture used `imap-tests store-text-message` without required sender/recipient fields:

```text
Field from is required
Error: Field from is required
exit status 1
```

- The fix was simply to rerun with:
  - `--from seed@example.com`
  - `--to a@test.local`

- A more interesting failure showed up during the first Docker incremental smoke. The second mirror run fetched one already-mirrored message again when it should have fetched zero. Inspecting `/tmp/smailnail-mirror-smoke/mirror.sqlite` showed the stored state was:

```text
127-0-0-1-993-a-d1e840edfa83|INBOX|1775068041|2|3
1|test mail 1775068041 to a|...
2|Mirror smoke 2026-04-01T20:15Z|...
```

- The root cause was the incremental search window. I initially used a dynamic IMAP range equivalent to `UID 3:*`. With `UIDNEXT = 3`, `*` resolves to the current maximum UID, and the resulting reversed range still matches the existing last message. The fix was to bound the search explicitly to `previousHighUID+1 .. UIDNEXT-1` and skip SEARCH entirely when that window is empty.

- One commit attempt also failed on lint because an updated test file had not been restaged, leaving the staged snapshot with two `copy` variable names that trigger the `predeclared` linter. Restaging the file fixed that cleanly.

### What I learned

- The IMAP incremental boundary should use `UIDNEXT` as an upper bound, not `*`, when searching for “newer than highest mirrored UID”.
- The local Docker Dovecot fixture is valuable for catching exactly these protocol-edge issues; the fake test session was not enough by itself.
- The existing `pkg/mailruntime` layer was close to reusable as-is, but the insecure TLS semantics needed correction for real local testing.

### What was tricky to build

- The hardest part was IMAP UID-window semantics. On paper, “search for UIDs greater than the highest mirrored UID” sounds trivial. In practice, the `*` token means “current highest assigned UID”, and IMAP ranges are not invalidated just because the left side is greater than the right side. That created a subtle replay bug right at the boundary where the mirror had already caught up. The resolution was to use mailbox status data more carefully and construct a bounded range from `highest+1` to `uidNext-1`.

- The other tricky part was defining what “reset local snapshot cleanly” should mean for this slice. I chose to make the SQLite snapshot authoritative for reset behavior now, while keeping raw files namespaced by `UIDVALIDITY`. That avoids mixed generations in the DB while leaving room to add more aggressive raw-file cleanup later if we decide it is worth the extra risk/complexity.

### What warrants a second pair of eyes

- Whether the current reset strategy should also prune old raw-file directories immediately, or whether leaving older `UIDVALIDITY` directories on disk is the right v1 tradeoff.
- Whether the command should already expose per-mailbox rows instead of only an aggregate summary row.
- Whether `pkg/mailruntime.Fetch` plus `FetchBodyRaw` remains the right medium-term primitive once the MIME parser/import phase lands.

### What should be done in the future

- Implement Phase 5:
  - add `pkg/mirror/parser.go`
  - parse RFC 822 properly with `go-message/mail`
  - populate richer text/HTML/attachment/search projections
- Add optional tombstoning/reconcile behavior for full scans.
- Add a dedicated end-to-end test that drives the maintained Docker Dovecot fixture automatically, not just the manual smoke path used in this step.

### Code review instructions

- Start with:
  - `pkg/mirror/service.go`
  - `pkg/mirror/files.go`
  - `cmd/smailnail/commands/mirror.go`
- Then review the support changes:
  - `pkg/mailruntime/imap_client.go`
  - `pkg/services/smailnailjs/service.go`
- Validate with:
  - `go test ./pkg/mirror ./cmd/smailnail ./pkg/mailruntime ./pkg/services/smailnailjs`
  - `docker compose -f docker-compose.local.yml up -d dovecot`
  - `go run ./cmd/imap-tests store-text-message --server 127.0.0.1 --port 993 --username a --password pass --mailbox INBOX --insecure --from seed@example.com --to a@test.local --subject "Mirror smoke ..." --body "Mirror smoke body ..." --output json`
  - `go run ./cmd/smailnail mirror --server 127.0.0.1 --port 993 --username a --password pass --mailbox INBOX --insecure --sqlite-path /tmp/smailnail-mirror-smoke/mirror.sqlite --mirror-root /tmp/smailnail-mirror-smoke/root --batch-size 25 --output json`
  - rerun the mirror command once to confirm `messages_fetched = 0`
  - append one more message and rerun to confirm `messages_fetched = 1`

### Technical details

- Real Docker fixture outputs observed during this step included:

```json
[
  {
    "account_key": "127-0-0-1-993-a-d1e840edfa83",
    "mailboxes_planned": 1,
    "mailboxes_synced": 1,
    "messages_fetched": 2,
    "messages_stored": 2,
    "raw_files_written": 2,
    "reused_file_writes": 0,
    "status": "synced"
  }
]
```

```json
[
  {
    "account_key": "127-0-0-1-993-a-d1e840edfa83",
    "mailboxes_planned": 1,
    "mailboxes_synced": 1,
    "messages_fetched": 0,
    "messages_stored": 0,
    "raw_files_written": 0,
    "reused_file_writes": 0,
    "status": "synced"
  }
]
```

```json
[
  {
    "account_key": "127-0-0-1-993-a-d1e840edfa83",
    "mailboxes_planned": 1,
    "mailboxes_synced": 1,
    "messages_fetched": 1,
    "messages_stored": 1,
    "raw_files_written": 1,
    "reused_file_writes": 0,
    "status": "synced"
  }
]
```

- Files changed in the code commit:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mailruntime/imap_client.go`
  - `pkg/services/smailnailjs/service.go`
  - `pkg/mirror/types.go`
  - `pkg/mirror/files.go`
  - `pkg/mirror/files_test.go`
  - `pkg/mirror/service.go`
  - `pkg/mirror/service_test.go`

## Step 5: Parse Raw RFC 822 Messages Into Searchable Fields

This step moved the mirror from “metadata plus raw files” to “metadata plus a searchable content projection.” Instead of relying only on envelope data and the ad hoc `FetchBodyText` field, the mirror now parses the stored raw RFC 822 message bytes with `go-message/mail` and uses that parsed view to populate `body_text`, `body_html`, attachment metadata, and a normalized `search_text` field.

The key value in this slice is that it validates the design direction of raw-message-first mirroring. Because the raw `.eml` bytes are already persisted, the parser can be improved independently of the sync transport. This step used that property directly and verified it on the local Docker Dovecot fixture with a multipart HTML message.

### Prompt Context

**User prompt (verbatim):** "alright, create detailed tasks in the ticket, and then work on them one by one, committing at appropriate intervals, keeping a diary."

**Assistant interpretation:** Continue through the planned implementation phases and document each code slice with the same discipline as the earlier ones.

**Inferred user intent:** The user wants the mirror to keep becoming more implementation-complete without losing commit hygiene, test coverage, or the reasoning trail in the ticket.

**Commit (code):** `f30a4c432200b77456cb116f4443477c4d8759e3` — `feat(mirror): parse raw messages into search fields`

### What I did

- Added `pkg/mirror/parser.go`:
  - `ParseMessage(raw []byte)`
  - parsed message subject, message-id, sent date, addresses
  - extracted inline `text/plain` and `text/html` parts
  - detected attachment parts and captured parsed part metadata
  - normalized search-text generation
  - simple HTML stripping for search indexing
- Updated `pkg/mirror/service.go`:
  - `buildMessageRecord` now prefers parsed raw-message values when parsing succeeds
  - parsed part metadata now replaces the earlier attachment-only `parts_json` projection
- Added `pkg/mirror/parser_test.go`:
  - multipart alternative parsing coverage
  - record-building coverage that proves parsed output overrides fallback envelope/body fields
- Verified with:
  - `go test ./pkg/mirror`
  - `go test ./pkg/mirror ./cmd/smailnail`
  - repo pre-commit `go test ./...`
  - repo pre-commit `golangci-lint run -v`
- Verified against Docker Dovecot with a multipart HTML message:
  - `docker compose -f docker-compose.local.yml up -d dovecot`
  - `go run ./cmd/imap-tests store-html-message ...`
  - `go run ./cmd/smailnail mirror ...`
  - `sqlite3 /tmp/smailnail-mirror-parser-smoke/mirror.sqlite "select subject, substr(body_text,1,40), body_html != '', instr(search_text,'Mirror HTML smoke') > 0 ..."`
  - `docker compose -f docker-compose.local.yml down`

### Why

- The mirror needs searchable message content, not just raw bytes and envelope metadata.
- Parsing from stored `.eml` files keeps the sync loop simpler and makes future parser improvements local and deterministic.
- HTML-message validation against Docker Dovecot was necessary because parser tests alone do not prove that the mirrored raw bytes coming from IMAP are the same shape the parser expects.

### What worked

- The parser successfully extracted both plain text and HTML from multipart/alternative messages in unit tests.
- `buildMessageRecord` now upgrades the SQLite projection with parsed subject/body/attachment/search fields when raw parsing succeeds.
- The Docker HTML smoke verified the end-to-end path:
  - message stored through the fixture
  - raw message downloaded by `smailnail mirror`
  - SQLite row contained:
    - the expected subject
    - non-empty `body_text`
    - non-empty `body_html`
    - `search_text` containing stripped HTML text

### What didn't work

- N/A

### What I learned

- The raw-first architecture pays off quickly: once the `.eml` bytes are durable, parser improvements become purely local transformations over already mirrored data.
- Even a simple HTML stripping pass is enough to make `search_text` materially better for a first search-oriented projection.
- It is still useful to keep the fetched IMAP envelope/body data as a fallback path so malformed raw messages do not block the mirror.

### What was tricky to build

- The main tricky part was deciding how aggressive the parser should be in v1. It would be easy to start normalizing every header and MIME nuance immediately, but that would slow the implementation and make the slice harder to verify. I kept the parser focused on the fields the schema already cares about: subject, date, address summaries, plain text, HTML, attachment presence, and a deterministic search projection.

- Another subtle point was making parsed data override fallback IMAP fetch data without making parsing mandatory. The solution was to parse opportunistically inside `buildMessageRecord` and only replace individual fields when the parser yields a stronger value. That keeps the mirror resilient while still benefiting from richer parsing whenever the raw message is well-formed.

### What warrants a second pair of eyes

- Whether the current HTML stripping is sufficient for search-text generation, or whether a more HTML-aware normalization pass is needed sooner.
- Whether attachment metadata in `parts_json` should distinguish inline related parts from true user-visible attachments more explicitly.
- Whether the parser should also become the source of truth for `headers_json` in a later slice instead of continuing to reuse fetched header maps.

### What should be done in the future

- Finish the remaining Phase 5 gap:
  - decide whether `headers_json` should come from the parser rather than IMAP fetch headers
- Add search/query behavior that can use `search_text` even when FTS5 is unavailable.
- Add a smoke case for attachment-heavy multipart messages, not just HTML alternatives.

### Code review instructions

- Start with:
  - `pkg/mirror/parser.go`
  - `pkg/mirror/parser_test.go`
- Then inspect integration points in:
  - `pkg/mirror/service.go`
- Validate with:
  - `go test ./pkg/mirror`
  - `docker compose -f docker-compose.local.yml up -d dovecot`
  - `go run ./cmd/imap-tests store-html-message --server 127.0.0.1 --port 993 --username a --password pass --mailbox INBOX --insecure --from parser@example.com --to a@test.local --subject "Mirror HTML smoke ..." --text-body "Plain html smoke body" --html-body '<html>...</html>' --output json`
  - `go run ./cmd/smailnail mirror --server 127.0.0.1 --port 993 --username a --password pass --mailbox INBOX --insecure --sqlite-path /tmp/smailnail-mirror-parser-smoke/mirror.sqlite --mirror-root /tmp/smailnail-mirror-parser-smoke/root --batch-size 25 --output json`
  - inspect the mirrored row with `sqlite3`

### Technical details

- The SQLite spot-check used for the real-fixture validation returned:

```text
Mirror HTML smoke 2026-04-01T20:50Z|Plain html smoke body|1|1
```

- Interpreted left to right, that row means:
  - the expected subject was mirrored
  - `body_text` started with the expected plain-text alternative
  - `body_html` was non-empty
  - `search_text` contained stripped HTML content matching `Mirror HTML smoke`

- Files changed in the code commit:
  - `pkg/mirror/parser.go`
  - `pkg/mirror/parser_test.go`
  - `pkg/mirror/service.go`

## Step 9: Add Full-Mailbox Reconciliation And Tombstoning

This step implemented the remaining mirror-side deletion story. The mirror was already durable for new and updated messages, but `remote_deleted` was just a stored column with no behavior behind it. The new reconcile path makes that column meaningful by comparing the local mailbox snapshot to a full server UID scan and updating rows that disappeared remotely.

The implementation is deliberately opt-in. Normal incremental sync still only fetches new UIDs, but `--reconcile-full-mailbox` now triggers an additional full mailbox UID search after sync and marks locally mirrored rows as `remote_deleted` when the server no longer reports them. That keeps the default path fast while giving operators a conservative mirror mode when they explicitly want remote deletion tracking.

### Prompt Context

**User prompt (verbatim):** "go ahead."

**Assistant interpretation:** Continue with the next remaining mirror implementation task instead of stopping after the previous slice.

**Inferred user intent:** The user wants the remaining major mirror gaps closed with the same pattern of implementation, validation, commits, and diary updates.

**Commit (code):** `f0aa4292d39d1da6240f2ec66ef068e28a7ae534` — `feat(mirror): add full-mailbox reconciliation`

### What I did

- Added `ReconcileFull` to mirror settings and sync options.
- Added the `--reconcile-full-mailbox` flag to `cmd/smailnail/commands/mirror.go`.
- Extended the mirror report types with:
  - `tombstonedMessages`
  - `restoredMessages`
  - `reconcileApplied`
- Updated `pkg/mirror/service.go`:
  - mailbox sync now can run a full post-sync UID search
  - local message rows for the current mailbox snapshot are compared to the remote UID set
  - rows missing on the server are marked `remote_deleted = true`
  - previously tombstoned rows still present on the server are restored to `remote_deleted = false`
  - reconcile runs even when there are no new UIDs, which is necessary for delete-only mailbox changes
- Added unit coverage in `pkg/mirror/service_test.go` for:
  - tombstoning a missing remote message
  - restoring a previously tombstoned message that is still present remotely
- Verified with:
  - `go test -tags sqlite_fts5 ./pkg/mirror -run 'TestServiceSync(ReconcileTombstonesMissingMessages|ReconcileRestoresPresentMessages)|TestServiceSyncPersistsIncrementalMessages|TestServiceSyncResetsOnUIDValidityChange'`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - Docker compose reconcile smoke using:
    - two seeded messages
    - a temporary `mail-rules` delete action against one subject
    - a second mirror run with `--reconcile-full-mailbox`
    - a SQLite check of `remote_deleted`

### Why

- The mirror needed a way to represent remote deletes without deleting local raw `.eml` files or rows by default.
- Reconciliation had to be explicit because a full mailbox UID scan is more expensive than the normal incremental sync path.
- The `remote_deleted` column only becomes useful once there is a concrete rule for setting and clearing it.

### What worked

- Both new reconcile unit tests passed.
- Full tagged repo tests and lint passed after the reconcile path was added.
- The repo-local compose-backed smoke produced the expected SQLite result after deleting one message remotely and rerunning the mirror with reconciliation:

```text
Mirror Reconcile Delete 1775075189|1
Mirror Reconcile Keep 1775075189|0
```

- That proves the deleted message was tombstoned while the surviving message remained active.

### What didn't work

- The first reconcile smoke attempt failed because the temporary delete rule used an invalid DSL `output` block with no fields:

```text
Error: error parsing rule file: invalid output config: at least one output field is required
exit status 1
```

- I corrected the rule to include a minimal `uid` field and reran the smoke successfully.

### What I learned

- Reconciliation must run even when there are no new UIDs, otherwise a mailbox that only lost messages never updates local tombstones.
- A conservative mirror mode works well as an explicit opt-in feature: it captures remote deletes without making local raw-message retention lossy by default.

### What was tricky to build

- The key subtlety was control flow, not SQL. The earlier sync logic returned early when `UIDNEXT` showed there were no new messages. That is correct for incremental fetches but wrong for delete-only mailbox changes. I had to refactor the “finalize mailbox sync” path so it can still reconcile and then persist sync-state timestamps even when no fetch batch runs.

- The other tricky point was restore semantics. A row can already be tombstoned from an earlier reconcile, and the next full scan should clear `remote_deleted` if that UID is still present remotely. Without that second branch, the column would only ever move in one direction and become stale.

### What warrants a second pair of eyes

- Whether the `--reconcile-full-mailbox` name is the best long-term operator-facing flag or whether a shorter alias such as `--tombstone-missing` is still worth adding later.
- Whether future large-mailbox optimization is needed for reconcile updates if mailbox UID sets get significantly larger.

### What should be done in the future

- Decide whether reconcile should eventually become a scheduled/default maintenance mode rather than a fully manual flag.
- Consider adding stricter mirror cleanup modes that physically remove rows and raw files after tombstoning, but only as an explicit separate mode.

### Code review instructions

- Start with:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/service.go`
  - `pkg/mirror/service_test.go`
- Validate with:
  - `go test -tags sqlite_fts5 ./pkg/mirror`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - reproduce the compose-backed delete + reconcile smoke and inspect `remote_deleted` in SQLite

### Technical details

- The reconcile smoke used a temporary delete rule with:

```yaml
name: reconcile delete smoke
description: delete one known fixture message
search:
  subject_contains: "<delete subject>"
output:
  format: json
  fields:
    - uid
actions:
  delete: true
```

- After deleting one message remotely and rerunning the mirror with `--reconcile-full-mailbox`, SQLite returned:

```text
Mirror Reconcile Delete 1775075189|1
Mirror Reconcile Keep 1775075189|0
```

- Files changed in the code commit:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/service.go`
  - `pkg/mirror/service_test.go`
  - `pkg/mirror/types.go`

## Step 6: Require SQLite FTS5 At Build Time

This step tightened the mirror from "FTS preferred when present" to "FTS required for supported builds". The main change is a deliberate compile-time failure for untagged `smailnail` builds that reach `pkg/mirror`, so operators and CI fail immediately instead of discovering a missing FTS feature after the binary is already built.

The second half of the step was operational cleanup. Once the build became intentionally strict, every supported entry point needed to agree with that decision. I updated the Makefile, CI workflows, and Docker IMAP smoke path so the required `sqlite_fts5` tag is always present in the normal build and validation flows, then reran the compose-backed smoke against the maintained Dovecot fixture.

### Prompt Context

**User prompt (verbatim):** "good, add that.

then, i don't understand the raw RFC 822 improvement"

**Assistant interpretation:** Enforce the FTS5 requirement in the build itself, then clarify the value of the raw-message parsing improvement.

**Inferred user intent:** The user wants the mirror to depend on FTS5 explicitly instead of pretending it is optional, and they want the design rationale explained in plain language.

**Commit (code):** `d2bed23557ada03540fbf4fc4e1f393df9fdfcbb` — `build(mirror): require sqlite fts5 tags`

### What I did

- Added `pkg/mirror/require_fts5_build_tag.go` with a negative build tag that references an undefined symbol unless `sqlite_fts5` or `fts5` is set.
- Updated `Makefile` so the default `build`, `test`, `lint`, `install`, and `build-embed` paths pass `sqlite_fts5`.
- Updated `.github/workflows/push.yml` and `.github/workflows/lint.yml` to run tagged tests and lint in CI.
- Updated `scripts/docker-imap-smoke.sh` to pass the required build tag for all `smailnail` invocations, with `SMAILNAIL_GO_TAGS` as an override.
- Validated both failure and success paths:
  - `go build ./cmd/smailnail`
  - `go build -tags sqlite_fts5 ./cmd/smailnail`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - `DOCKER_IMAP_FIXTURE_ROOT=/home/manuel/code/others/docker-test-dovecot ./scripts/docker-imap-smoke.sh`

### Why

- The mirror feature depends on SQLite FTS5 for its intended search behavior, so the build should say that directly.
- Compile-time failure is clearer than shipping a binary that only discovers missing FTS support when opening the database.
- Once the build contract changed, the repo needed one consistent operator story for local runs, CI, and Docker-based smoke validation.

### What worked

- Untagged `go build ./cmd/smailnail` now fails immediately in `pkg/mirror/require_fts5_build_tag.go`, which is the intended enforcement point.
- Tagged builds and tests passed cleanly after the Makefile and CI changes.
- The maintained Docker Dovecot smoke still passed after the script was updated to run `go run -tags "$SMAILNAIL_GO_TAGS" ./cmd/smailnail ...`.

### What didn't work

- The first smoke-script run failed before Docker startup because the script assumes a sibling checkout at `../docker-test-dovecot`, but this workspace did not have that layout. The exact output was:

```text
Docker IMAP fixture not found at '/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/../docker-test-dovecot'.
Set DOCKER_IMAP_FIXTURE_ROOT to the docker-test-dovecot checkout.
```

- I reran the smoke with `DOCKER_IMAP_FIXTURE_ROOT=/home/manuel/code/others/docker-test-dovecot`, which completed successfully.

### What I learned

- A compile-time guard is a better contract than a runtime "maybe" when the feature genuinely depends on a compiled SQLite option.
- The repo already had enough central build surfaces that enforcing the tag mostly meant fixing documentation drift, not redesigning the build graph.
- Keeping the smoke script configurable with `SMAILNAIL_GO_TAGS` and `DOCKER_IMAP_FIXTURE_ROOT` makes it easier to validate the same path across multiple local workspace layouts.

### What was tricky to build

- The delicate part was deciding where to enforce the build tag. Enforcing it inside `pkg/mirror` means any binary importing the mirror package fails early unless FTS5 is compiled in. That is stricter than a CLI-only guard, but it matches the architectural reality that the mirror package itself now assumes an FTS-capable SQLite build. I kept the guard minimal so the failure is obvious and does not add any compatibility shim or alternate code path.

- The other tricky piece was avoiding partial enforcement. A build-tag guard is only useful if every routine validation path also adopts the same tag. That is why I updated the Makefile, CI workflows, and Docker smoke script together instead of treating the guard as an isolated code change.

### What warrants a second pair of eyes

- Whether `pkg/mirror` is the right enforcement boundary long term, or whether a future library consumer would want a narrower CLI-only restriction.
- Whether the remaining runtime FTS-detection fallback in `pkg/mirror/schema.go` should now be removed as dead complexity.
- Whether release automation such as Goreleaser should be updated in a follow-up to make the required tag explicit there too.

### What should be done in the future

- Decide whether to remove the remaining runtime FTS fallback code paths now that supported builds require FTS5.
- Review release packaging and cross-compilation assumptions for `go-sqlite3` plus FTS5.
- Keep the Docker smoke script path assumptions documented or make fixture discovery more flexible if workspace layouts continue to vary.

### Code review instructions

- Start with:
  - `pkg/mirror/require_fts5_build_tag.go`
  - `Makefile`
  - `scripts/docker-imap-smoke.sh`
- Then inspect:
  - `.github/workflows/push.yml`
  - `.github/workflows/lint.yml`
- Validate with:
  - `go build ./cmd/smailnail`
  - `go build -tags sqlite_fts5 ./cmd/smailnail`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - `DOCKER_IMAP_FIXTURE_ROOT=/home/manuel/code/others/docker-test-dovecot ./scripts/docker-imap-smoke.sh`

### Technical details

- The compile-time guard file is intentionally tiny:

```go
//go:build !sqlite_fts5 && !fts5

package mirror

var _ = requires_sqlite_fts5_build_tag
```

- The untagged build now fails with:

```text
# github.com/go-go-golems/smailnail/pkg/mirror
pkg/mirror/require_fts5_build_tag.go:5:9: undefined: requires_sqlite_fts5_build_tag
```

- Files changed in the code commit:
  - `pkg/mirror/require_fts5_build_tag.go`
  - `Makefile`
  - `.github/workflows/push.yml`
  - `.github/workflows/lint.yml`
  - `scripts/docker-imap-smoke.sh`

## Step 7: Remove The Remaining Runtime FTS Fallback

This step brought the runtime in line with the build contract introduced in the previous slice. Once the build already requires SQLite FTS5, the mirror no longer benefits from carrying three runtime “search modes” or exposing a `--search-mode` flag that no longer represents a supported operational choice.

The cleanup was small but architectural: the schema bootstrap now always creates the FTS table, the store bootstrap no longer accepts a runtime search-mode selector, and the mirror command now presents a single supported path. I reran the tagged repo checks and the Docker compose IMAP smoke to verify that removing the dead branch did not change normal behavior.

### Prompt Context

**User prompt (verbatim):** "go ahead, add tasks, and then work on them as before."

**Assistant interpretation:** Update the ticket with the next implementation tasks and continue executing them with the same commit and diary discipline as earlier slices.

**Inferred user intent:** The user wants the design cleanup to continue incrementally, with each meaningful slice implemented, tested, committed, and documented in the ticket.

**Commit (code):** `215920ddf1ec71cbee377ff6624615e861a1acf8` — `refactor(mirror): make fts bootstrap mandatory`

### What I did

- Updated the task list to add and then complete the “remove runtime FTS fallback” slice.
- Removed the `--search-mode` field and flag from `cmd/smailnail/commands/mirror.go`.
- Simplified `pkg/mirror/schema.go`:
  - `bootstrapSchema` no longer accepts a search-mode parameter
  - `bootstrapFTS` now always attempts to create the FTS5 virtual table
  - FTS creation failure is always treated as an error
- Simplified `pkg/mirror/store.go` so `Store.Bootstrap` no longer takes a runtime search-mode argument.
- Updated `pkg/mirror/store_test.go` and `pkg/mirror/service_test.go` to use the single supported FTS-backed bootstrap path.
- Updated the implementation guide so it no longer documents `--search-mode` as a supported v1 flag.
- Validated with:
  - `go test -tags sqlite_fts5 ./pkg/mirror ./cmd/smailnail`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - `DOCKER_IMAP_FIXTURE_ROOT=/home/manuel/code/others/docker-test-dovecot ./scripts/docker-imap-smoke.sh`

### Why

- The mirror already fails at build time without FTS5, so keeping runtime “basic” and “auto” branches only creates a misleading API surface.
- Removing the dead selector makes the command help, schema bootstrap, and tests easier to understand for the next implementer.
- This cleanup reduces the number of branches to reason about before adding more substantive behavior like tombstoning and raw-header canonicalization.

### What worked

- The mirror package and CLI compiled cleanly after the `Store.Bootstrap` signature change and flag removal.
- Tagged repo tests and lint still passed without additional changes.
- The Docker compose IMAP smoke still passed, showing that the supported operator path did not depend on the removed runtime search-mode branch.

### What didn't work

- N/A

### What I learned

- The compile-time FTS requirement and the runtime bootstrap path should be treated as one contract. If they disagree, the codebase starts advertising choices that do not really exist.
- The dead branch removal was smaller than it looked because the earlier FTS enforcement slice had already aligned most build surfaces.

### What was tricky to build

- The main subtlety was deciding whether to keep `search_mode` around as a user-configurable setting for later. I removed the flag entirely because it no longer reflects a supported runtime choice, but I kept `search_mode` in the bootstrap report as the fixed value `fts5` so current output remains explicit about the backend in use. That keeps the contract clear without adding a compatibility layer.

- The other small sharp edge was the constructor signature change on `Store.Bootstrap`. That kind of API cleanup is easy to miss in tests, so I used a ripgrep sweep before the full test run to catch stale callsites quickly.

### What warrants a second pair of eyes

- Whether the `search_mode` output field should stay as a fixed informational value or be renamed later to something less choice-shaped, such as `search_backend`.
- Whether the design guide should be updated further to describe the runtime FTS table maintenance strategy now that FTS is fully mandatory.

### What should be done in the future

- Implement full-mailbox reconciliation and tombstoning for remotely deleted messages.
- Make the raw RFC 822 parser the canonical source for stored headers and address summaries.
- Consider renaming `search_mode` output to reduce the impression that multiple runtime modes still exist.

### Code review instructions

- Start with:
  - `pkg/mirror/schema.go`
  - `pkg/mirror/store.go`
  - `cmd/smailnail/commands/mirror.go`
- Then inspect:
  - `pkg/mirror/store_test.go`
  - `pkg/mirror/service_test.go`
- Validate with:
  - `go test -tags sqlite_fts5 ./pkg/mirror ./cmd/smailnail`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - `DOCKER_IMAP_FIXTURE_ROOT=/home/manuel/code/others/docker-test-dovecot ./scripts/docker-imap-smoke.sh`

### Technical details

- The removed runtime split covered three values:
  - `basic`
  - `fts-auto`
  - `fts-required`

- After this slice, the supported runtime path is:
  - build with `sqlite_fts5`
  - bootstrap the mirror schema
  - create `messages_fts`
  - fail immediately if FTS creation does not succeed

- Files changed in the code commit:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/schema.go`
  - `pkg/mirror/store.go`
  - `pkg/mirror/store_test.go`
  - `pkg/mirror/service_test.go`
  - `pkg/mirror/types.go`

## Step 8: Normalize Parsed Header Projections

This step finished the semantic-header direction for the mirror. The raw `.eml` file remains the exact archived message source, but the convenient SQLite projection now prefers normalized values derived from parsing that raw message instead of mixing fetched IMAP header maps with parsed body fields.

The key behavioral change is in `headers_json`. When raw parsing succeeds, the mirror now stores a normalized parsed header map for important fields like `From`, `To`, `Cc`, `Date`, `Subject`, and `Message-Id`, aligned with the explicit message columns and the search projection. That removes formatting noise such as quoted display-name variants and double-bracketed message IDs.

### Prompt Context

**User prompt (verbatim):** "ok, 2"

**Assistant interpretation:** Proceed with normalizing parsed addresses and headers into a stable summary/projection format instead of keeping raw RFC-style formatting in SQLite.

**Inferred user intent:** The user wants the raw message to remain the source of truth while keeping the searchable SQLite projection deterministic and human-readable.

**Commit (code):** `bb97160ae5d9bd89af0233f2bf9bda6ba46fc2af` — `feat(mirror): normalize parsed header projections`

### What I did

- Updated `pkg/mirror/parser.go`:
  - added parsed raw-header extraction using `net/mail`
  - canonicalized header keys with MIME-style capitalization
  - normalized display-name summaries into `Name <address>` form
  - normalized `Message-Id` into a single bracketed semantic value
  - built a normalized parsed header map that prefers semantic parsed values for key headers
- Updated `pkg/mirror/service.go` so `buildMessageRecord` replaces fetched `headers_json` with the normalized parsed header map whenever raw parsing succeeds.
- Updated `pkg/mirror/parser_test.go`:
  - added direct assertions for normalized parsed `From` and `Message-Id`
  - verified `buildMessageRecord` stores parsed normalized headers instead of fallback fetched headers
- Verified with:
  - `go test -tags sqlite_fts5 ./pkg/mirror -run 'TestParseMessageMultipartAlternative|TestBuildMessageRecordUsesParsedProjection'`
  - `go test -tags sqlite_fts5 ./pkg/mirror ./cmd/smailnail`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - Docker Dovecot smoke using the repo-local compose fixture plus a SQLite spot-check:
    - `go run ./cmd/imap-tests store-text-message ...`
    - `go run -tags sqlite_fts5 ./cmd/smailnail mirror ...`
    - `sqlite3 ... "select subject || '|' || json_extract(headers_json, '$.From') || '|' || json_extract(headers_json, '$.\"Message-Id\"') ..."`

### Why

- The exact wire-format source is already preserved in the mirrored `.eml` file, so `headers_json` should optimize for stable semantics, not byte-faithful reproduction.
- A normalized header projection is a better fit for SQLite consumers, search, and tests than a lossy pseudo-raw header dump.
- This keeps the row internally consistent: the explicit columns and `headers_json` now come from the same parsed raw source.

### What worked

- Parser tests now pass with normalized address and `Message-Id` expectations.
- Full tagged repo tests and lint passed after the change.
- The repo-local Docker Dovecot smoke produced the expected normalized `From` header in SQLite:

```text
Mirror Header Canonicalization 1775074690|Header Parser <parser@example.com>|
```

- That confirms the parsed-header path is active end to end for mirrored rows.

### What didn't work

- The first end-to-end Docker attempt failed because an earlier external Dovecot fixture I had started in this session was still holding ports `993` and `4190`.
- After stopping `docker-test-dovecot-testcot-1`, the next attempt raced Dovecot startup and the initial IMAP operation returned `EOF`.
- I reran the smoke with a stricter readiness wait against `imap-tests list-mailboxes`, after which the mirror smoke completed successfully.
- The SQLite smoke row showed an empty `Message-Id` field because `imap-tests store-text-message` does not appear to inject a `Message-ID` header. That means end-to-end `Message-Id` canonicalization is still covered by unit tests rather than by this specific Docker fixture path.

### What I learned

- The cleanest mirror contract is:
  - raw `.eml` for exact source
  - normalized parsed fields for SQLite projection
- `headers_json` becomes much more useful once it is treated as a semantic projection rather than an attempt at a pseudo-raw dump.
- The local Docker fixture is good for proving the parser path is wired, but not every semantic header is guaranteed to exist in every generated fixture message.

### What was tricky to build

- The main sharp edge was that two independent normalizations had to line up:
  - parsed address summaries should lose formatting noise like added quotes
  - `Message-Id` needed to avoid double-bracketing
  It was not enough to normalize only the explicit columns; `headers_json` had to be built from the same semantic view or the row would remain internally inconsistent.

- Another subtle point was key capitalization. The raw header parser can surface keys in various cases, so canonicalizing them with MIME-style capitalization avoids silent duplication between forms like `Message-ID` and `Message-Id` inside the stored JSON map.

### What warrants a second pair of eyes

- Whether `headers_json` should eventually preserve repeated headers in a richer shape instead of a flat `map[string]string`.
- Whether the fixture generators should explicitly set `Message-ID` so end-to-end smoke coverage can assert the normalized `Message-Id` value as well.

### What should be done in the future

- Implement full-mailbox reconciliation and tombstoning for remotely deleted messages.
- Consider adding a higher-fidelity structured header/address representation if future consumers need more than summary strings.
- Add a deterministic smoke case that injects an explicit `Message-ID` so Docker-based validation covers that path too.

### Code review instructions

- Start with:
  - `pkg/mirror/parser.go`
  - `pkg/mirror/service.go`
  - `pkg/mirror/parser_test.go`
- Validate with:
  - `go test -tags sqlite_fts5 ./pkg/mirror -run 'TestParseMessageMultipartAlternative|TestBuildMessageRecordUsesParsedProjection'`
  - `go test -tags sqlite_fts5 ./...`
  - `golangci-lint run -v --build-tags sqlite_fts5`
  - run the repo-local compose Dovecot smoke and inspect the mirrored SQLite row

### Technical details

- The canonical address summary format is now:
  - `Name <user@example.com>` when a display name exists
  - `user@example.com` when it does not

- `Message-Id` normalization trims whitespace and angle brackets, then stores a single bracketed form:

```go
func normalizeMessageID(messageID string) string {
    messageID = strings.TrimSpace(messageID)
    messageID = strings.Trim(messageID, "<>")
    if messageID == "" {
        return ""
    }
    return "<" + messageID + ">"
}
```

- Files changed in the code commit:
  - `pkg/mirror/parser.go`
  - `pkg/mirror/parser_test.go`
  - `pkg/mirror/service.go`
