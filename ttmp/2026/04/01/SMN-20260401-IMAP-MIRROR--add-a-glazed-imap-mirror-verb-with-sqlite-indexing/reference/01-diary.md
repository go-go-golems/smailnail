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
    - Path: smailnail/pkg/mirror/schema.go
      Note: Added mirror schema bootstrap and FTS detection in commit 1d9578a08372607e77e4de17bb95a1b75522568d
    - Path: smailnail/pkg/mirror/service.go
      Note: Added incremental IMAP sync orchestration in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
    - Path: smailnail/pkg/mirror/service_test.go
      Note: Added incremental sync and UIDVALIDITY regression tests in commit 9b0afe7a06542be44f8ae87f397c446232ec8efb
    - Path: smailnail/pkg/mirror/store.go
      Note: Added local mirror store bootstrap in commit 1d9578a08372607e77e4de17bb95a1b75522568d
    - Path: smailnail/pkg/mirror/store_test.go
      Note: Added initial mirror schema tests in commit 1d9578a08372607e77e4de17bb95a1b75522568d
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
LastUpdated: 2026-04-01T20:40:00-04:00
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
