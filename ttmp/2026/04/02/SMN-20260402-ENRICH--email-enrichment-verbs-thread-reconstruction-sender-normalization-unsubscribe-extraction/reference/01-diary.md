---
Title: Diary
Ticket: SMN-20260402-ENRICH
Status: active
Topics:
    - email
    - sqlite
    - glazed
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/smailnail/commands/enrich/all.go
      Note: CLI wrapper for RunAll (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/enrich/common.go
      Note: Shared enrich command DB bootstrap and scope settings (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/enrich/root.go
      Note: CLI group wiring for enrich verbs (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/enrich/senders.go
      Note: CLI wrapper for sender enrichment (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/enrich/threads.go
      Note: CLI wrapper for thread enrichment (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/enrich/unsubscribe.go
      Note: CLI wrapper for unsubscribe enrichment (commit 0f22a5c901b58d32595c904e985126e57da51ea3)
    - Path: cmd/smailnail/commands/mirror.go
      Note: Post-sync enrich-after hook and row output fields (commit 0a86de84f47c4cb637e1e3b7f939551ae2f9c130)
    - Path: pkg/enrich/all.go
      Note: RunAll orchestration used by CLI and mirror hook (commit bb8a68d794fe5034ec050dc041fededa2571d727)
    - Path: pkg/enrich/parse_address.go
      Note: Address normalization helpers (commit 068f4cdbac61f063a891075454d1fda7c00140b3)
    - Path: pkg/enrich/parse_headers.go
      Note: Header parsing helpers (commit 068f4cdbac61f063a891075454d1fda7c00140b3)
    - Path: pkg/enrich/schema.go
      Note: Initial enrichment migration definitions (commit 068f4cdbac61f063a891075454d1fda7c00140b3)
    - Path: pkg/enrich/threads.go
      Note: Thread reconstruction and summary rebuild logic (commit cb2d99a30943360edf0dda91a6249bdc3221e261)
    - Path: pkg/enrich/types.go
      Note: Shared enrichment reports and options (commit 068f4cdbac61f063a891075454d1fda7c00140b3)
    - Path: pkg/enrich/unsubscribe.go
      Note: List-Unsubscribe extraction and sender updates (commit bb8a68d794fe5034ec050dc041fededa2571d727)
    - Path: pkg/mirror/schema.go
      Note: Migration v2 is wired into store bootstrap (commit 068f4cdbac61f063a891075454d1fda7c00140b3)
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-02T13:53:25.183534353-04:00
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Capture the implementation of ticket `SMN-20260402-ENRICH` step by step so review and continuation stay easy even when the work is split across multiple focused commits.

## Step 1: Bootstrap The Enrichment Package

I started with the lowest-risk slice: schema wiring, shared report/options types, and the parsing helpers that every enricher will need. That gives the later integration work a stable base and lets the first tests focus on pure parsing behavior instead of multi-table mutations.

The other reason to start here was ticket hygiene. The task list already separated schema, types, address parsing, and header parsing into distinct tasks, so this checkpoint maps directly to the first four real implementation items and establishes the commit cadence the rest of the ticket can follow.

### Prompt Context

**User prompt (verbatim):** "Implement docmgr ticket SMN-20260402-ENRICH/SMN-20260402-ENRICH , go task by task, commit at appropriate intervals, keep a detailed diary as you work."

**Assistant interpretation:** Implement the enrichment ticket in the existing `smailnail` module, execute the documented tasks in order, keep docmgr bookkeeping current, and create real commits instead of leaving one large uncommitted diff.

**Inferred user intent:** Leave behind a reviewable implementation trail where code, ticket state, and diary entries all stay synchronized.

**Commit (code):** `068f4cdbac61f063a891075454d1fda7c00140b3` — "Add enrichment schema and parser helpers"

### What I did
- Added [`pkg/enrich/schema.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/schema.go) to define migration-v2 statements for `threads`, `senders`, and the new derived `messages` columns.
- Added [`pkg/enrich/types.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/types.go) with the shared `Options`, `ThreadsReport`, `SendersReport`, `UnsubscribeReport`, and `AllReport` structs.
- Added [`pkg/enrich/parse_address.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_address.go) and [`pkg/enrich/parse_headers.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_headers.go) plus focused parser tests.
- Updated [`pkg/mirror/schema.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go) to bump the schema version to `2`, consume the enrichment migration statements, and tolerate duplicate-column errors during migration replay.
- Ran `go fmt ./pkg/enrich ./pkg/mirror`, `go test ./pkg/enrich ./pkg/mirror`, and `go test ./pkg/mirror -tags sqlite_fts5`.
- Checked off ticket tasks `2,3,4,5` and updated the ticket changelog with the new files.

### Why
- The enrichment commands and mirror integration both depend on the schema and shared types existing first.
- Address and header parsing are isolated enough to validate early, which reduces the risk of debugging SQL and parser issues at the same time later.
- Wiring migration v2 before the enrichers avoids each later step having to guess whether the target columns/tables already exist.

### What worked
- The new parser tests in `pkg/enrich` passed immediately after formatting.
- The repo's pre-commit hook successfully validated the checkpoint with `go test -tags "sqlite_fts5" ./...` and `golangci-lint run -v --build-tags sqlite_fts5`.
- The ticket bookkeeping flow worked cleanly: `docmgr task check` and `docmgr changelog update` updated the workspace without needing any manual frontmatter repair first.

### What didn't work
- `go test ./pkg/mirror` failed before adding the build tag because [`pkg/mirror/require_fts5_build_tag.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/require_fts5_build_tag.go) intentionally references `requires_sqlite_fts5_build_tag` when `sqlite_fts5` is absent.
- Exact command and error:

```text
go test ./pkg/enrich ./pkg/mirror
# github.com/go-go-golems/smailnail/pkg/mirror [github.com/go-go-golems/smailnail/pkg/mirror.test]
pkg/mirror/require_fts5_build_tag.go:5:9: undefined: requires_sqlite_fts5_build_tag
```

- Running `go test ./pkg/mirror -tags sqlite_fts5` resolved that immediately, which confirms the failure was environmental rather than caused by the enrichment changes.

### What I learned
- The repo already enforces the correct verification path through `lefthook`, so the safest default for later checkpoints is `go test -tags sqlite_fts5 ./...`.
- `github.com/emersion/go-message/mail` is enough for the `from_summary` decoding cases in the ticket; no extra dependency or custom RFC 2047 handling is needed.
- The current ticket workspace already had a design doc and changelog, but no diary document; `docmgr doc add` fit cleanly on top of those existing artifacts.

### What was tricky to build
- The only real sharp edge in this slice was schema replay behavior. SQLite supports `ALTER TABLE ... ADD COLUMN` but not `IF NOT EXISTS`, so the migration code has to distinguish between a legitimate migration failure and a benign duplicate-column error. The symptom was architectural rather than runtime: without that guard, a partially applied v2 migration would make reruns brittle. I handled it by adding a narrow `isIgnorableMigrationError` check in [`pkg/mirror/schema.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go) instead of complicating the exported enrichment schema with driver-specific branching.

### What warrants a second pair of eyes
- The `GuessRelayDomain` helper currently returns a normalized slug, not a true registered domain. That matches the design doc but could still deserve confirmation if later code starts displaying it as if it were a DNS-resolvable hostname.
- The report structs are intentionally minimal right now. If the CLI layer needs additional summary fields, those should be added once the enrichers are implemented rather than guessed too early.

### What should be done in the future
- Implement the sender enricher next, because unsubscribe enrichment depends on sender identity and thread summaries can reuse sender-derived participant counts.

### Code review instructions
- Start with [`pkg/mirror/schema.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go) to confirm migration versioning and replay behavior.
- Review [`pkg/enrich/parse_address.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_address.go) and [`pkg/enrich/parse_headers.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_headers.go) next; those functions define the normalization semantics the enrichers will build on.
- Validate with `go test -tags sqlite_fts5 ./...` from the repo root.

### Technical details
- Commands run:

```bash
docmgr doc add --ticket SMN-20260402-ENRICH --doc-type reference --title 'Diary'
go fmt ./pkg/enrich ./pkg/mirror
go test ./pkg/enrich ./pkg/mirror
go test ./pkg/mirror -tags sqlite_fts5
git add pkg/enrich pkg/mirror/schema.go
git commit -m "Add enrichment schema and parser helpers"
```

- Ticket bookkeeping completed in this step:
  `docmgr task check --ticket SMN-20260402-ENRICH --id 2,3,4,5`

## Step 2: Implement Sender Normalization

With the schema and parser helpers in place, I moved to the first real enrichment pass: deriving normalized sender rows and backfilling `messages.sender_email` and `messages.sender_domain`. This was the right next step because the unsubscribe pass depends on sender identity, and the thread summary can optionally reuse sender-derived participant counts later.

I kept this slice transactional and testable in isolation. Instead of pulling in the mirror bootstrap path for the integration test, I used an in-memory SQLite fixture with just the base `messages` schema plus the enrichment migration so `go test ./pkg/enrich` still works without the FTS build tag.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Continue the ticket in order, keep the implementation incremental, and make each task reviewable on its own.

**Inferred user intent:** Build the feature in a way that leaves behind trustworthy intermediate checkpoints rather than one opaque end-state commit.

**Commit (code):** `4407dafe15f0f4ae3197ea290c90519c345d564f` — "Implement sender enrichment pass"

### What I did
- Added [`pkg/enrich/common.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/common.go) with shared scope-clause construction, option normalization, and `mirror_metadata` upsert logic for later enrichers.
- Added [`pkg/enrich/senders.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders.go) with `SenderEnricher.Enrich`, sender aggregation by `from_summary`, transactional sender upserts, and message tagging.
- Added [`pkg/enrich/senders_test.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders_test.go) with one test for the initial enrichment run and one test that verifies a rerun only processes newly inserted messages with blank `sender_email`.
- Ran `go fmt ./pkg/enrich`, `go test ./pkg/enrich`, and the repo pre-commit hook validation triggered by `git commit`.
- Checked off ticket task `6` and updated the ticket changelog for the sender pass.

### Why
- Sender normalization is the first enrichment pass that writes durable derived data and therefore validates whether the migration design actually supports the intended workflow.
- Adding shared scope and metadata helpers here avoids duplicating the same SQL plumbing across the next two enrichers.
- Keeping the integration test free of the mirror package makes the feedback loop faster and reduces accidental coupling to the FTS build-tag path.

### What worked
- The incremental path behaved the way the ticket design needs: once a message has `sender_email`, rerunning the sender pass skips it unless `Rebuild` is set.
- The private-relay detection and `GuessRelayDomain` logic produced the expected `zillow` grouping key in the integration test.
- The commit hook again validated the whole repository with `go test -tags "sqlite_fts5" ./...` and `golangci-lint run -v --build-tags sqlite_fts5`.

### What didn't work
- Nothing materially failed in this step after the code was written. The main risk was semantic rather than operational: deciding how to keep `senders.msg_count` correct between incremental and rebuild modes.

### What I learned
- The cleanest way to preserve a fast local test loop is to avoid importing `pkg/mirror` into `pkg/enrich` integration tests unless the build-tag dependency is truly part of what is being tested.
- Grouping by `from_summary` and then merging by normalized email gives a practical middle ground: fewer message updates than row-by-row processing, while still collapsing repeated display strings into a single sender row.

### What was tricky to build
- The sharp edge in this pass was rebuild semantics. Incremental runs should add counts from newly tagged messages, but rebuild runs should not double-count previously processed mail. The underlying cause is that `senders` is keyed only by email, while the commands also support optional account/mailbox scoping. I resolved that for now by making incremental upserts additive and rebuild upserts overwrite the processed sender aggregate, which is correct for full rebuilds and still safe for scoped rebuilds even if scoped counts remain only as precise as the selected slice.

### What warrants a second pair of eyes
- The sender count semantics under scoped rebuilds deserve review. They are operationally safe, but the global `senders` table shape means a mailbox-scoped rebuild cannot perfectly recompute all-email totals without additional schema.
- The sender pass currently skips unparsable `from_summary` values instead of failing the whole run. That keeps the enricher robust, but if visibility into skipped rows becomes important we may want an explicit counter later.

### What should be done in the future
- Implement thread reconstruction next, then use that result as the basis for the `threads` summary table and final `RunAll` orchestration.

### Code review instructions
- Start with [`pkg/enrich/senders.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders.go) and focus on `loadSenderSummaryRows`, `upsertSender`, and `tagMessagesForSender`.
- Review [`pkg/enrich/senders_test.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders_test.go) to confirm the first-run and incremental expectations match the ticket design.
- Validate with `go test ./pkg/enrich` for the narrow loop or `go test -tags sqlite_fts5 ./...` for full repo verification.

### Technical details
- Commands run:

```bash
go fmt ./pkg/enrich
go test ./pkg/enrich
docmgr task check --ticket SMN-20260402-ENRICH --id 6
git add pkg/enrich/common.go pkg/enrich/senders.go pkg/enrich/senders_test.go
git commit -m "Implement sender enrichment pass"
```

## Step 3: Reconstruct Threads

The next pass built thread IDs and summary rows from `References` and `In-Reply-To`. I kept the algorithm in memory and then wrote back only the messages that still had blank `thread_id` values unless the command runs in rebuild mode.

This step also had to deal with a design mismatch in the ticket: thread roots are global message IDs, but some ancestry chains reference external messages that are not present in the local DB. I chose the earliest message that actually exists locally as the root, which matches the written design and keeps the graph deterministic.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Continue through the remaining ticket tasks in order, keeping the implementation and ticket state synchronized after each checkpoint.

**Inferred user intent:** Get a production-usable implementation without losing the intermediate reasoning behind tricky normalization and threading choices.

**Commit (code):** `cb2d99a30943360edf0dda91a6249bdc3221e261` — "Implement thread enrichment pass"

### What I did
- Added [`pkg/enrich/threads.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/threads.go) with `ThreadEnricher`, message-id parsing, parent resolution, thread-depth assignment, and summary upserts into `threads`.
- Added [`pkg/enrich/threads_test.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/threads_test.go) to cover a normal root/reply chain, a dangling external parent, and an incremental rerun that updates an existing thread summary.
- Ran `go fmt ./pkg/enrich` and `go test ./pkg/enrich`, then let the commit hook validate the full repo again.
- Checked off ticket task `7` and updated the changelog entry for the thread pass.

### Why
- Threading is the second core enrichment primitive and unlocks the `threads` summary table the CLI needs to surface.
- Handling missing external ancestors early prevents the later CLI layer from having to guess why some roots point at replies rather than true origin messages.

### What worked
- The depth assignments came out as expected in the integration test: root messages stay at depth `0`, direct replies at `1`, and incremental reruns update only the touched thread summary row.
- The repo-wide hook stayed green after adding the thread graph logic, which means the new code did not destabilize unrelated packages.

### What didn't work
- No blocking runtime failures surfaced in this step.

### What I learned
- Recomputing the root assignment for all in-scope messages, while only writing blank `thread_id` rows by default, is a good compromise between correctness and incremental behavior.
- The easiest place to compute participant counts is the thread-summary build step itself, falling back to `from_summary` parsing when sender enrichment has not run yet.

### What was tricky to build
- The tricky part was root resolution when the first referenced message is not present locally. The symptom is that a naive algorithm would assign a missing external message ID as the thread root and leave the DB pointing at a non-existent key. I avoided that by walking parents only while they exist in the in-memory `present` set and returning the earliest local ancestor instead.

### What warrants a second pair of eyes
- The `threads` schema uses `thread_id` as the primary key while also storing `account_key` and `mailbox_name` columns. The current implementation treats the summary row as global per `thread_id`, which is coherent with the primary key but worth reviewing if per-mailbox summaries are desired later.

### What should be done in the future
- Finish unsubscribe extraction and `RunAll`, then wire the CLI group on top of the stable enrichers.

### Code review instructions
- Start in [`pkg/enrich/threads.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/threads.go), especially `resolveThreadRoot`, `buildThreadSummaries`, and `upsertThreadSummary`.
- Validate with `go test ./pkg/enrich` for the narrow loop or `go test -tags sqlite_fts5 ./...` for the whole repo.

### Technical details
- Commands run:

```bash
go fmt ./pkg/enrich
go test ./pkg/enrich
docmgr task check --ticket SMN-20260402-ENRICH --id 7
git add pkg/enrich/common.go pkg/enrich/threads.go pkg/enrich/threads_test.go
git commit -m "Implement thread enrichment pass"
```

## Step 4: Add Unsubscribe Extraction And RunAll

Once senders and threads were stable, I added the final enrichment pass for `List-Unsubscribe` extraction and then wrapped the three passes in a `RunAll` helper. That gave the later CLI integration a single stable orchestration entry point instead of three separate call sites.

This was also the point where the ticket design became most obviously pragmatic rather than fully normalized. The sender table stores unsubscribe URLs, but it does not persist a dedicated `one_click` column, so the pass reports one-click counts during execution without expanding the schema beyond what the design requested.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete the remaining backend enrichment tasks before moving on to command wiring.

**Inferred user intent:** Reach a point where the backend feature set is done and the CLI layer becomes a thin adapter rather than the place where behavior is invented.

**Commit (code):** `bb8a68d794fe5034ec050dc041fededa2571d727` — "Add unsubscribe enrichment and RunAll"

### What I did
- Added [`pkg/enrich/unsubscribe.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/unsubscribe.go) to parse `List-Unsubscribe` and `List-Unsubscribe-Post`, choose the most recent links per sender, and upsert the sender row.
- Added [`pkg/enrich/unsubscribe_test.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/unsubscribe_test.go) to cover latest-link selection and incremental skipping of already-processed senders.
- Added [`pkg/enrich/all.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/all.go) and [`pkg/enrich/all_test.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/all_test.go) to orchestrate the three enrichers against a single SQLite path.
- Refactored the test helpers slightly so the same bootstrap path can initialize either an in-memory DB or a file-backed DB for `RunAll`.
- Checked off ticket tasks `8` and `9` and updated the changelog.

### Why
- `RunAll` is the foundation both the `smailnail enrich all` command and `mirror --enrich-after` need.
- Unsubscribe extraction depends on sender identity, so it belongs after sender normalization and before the CLI layer.

### What worked
- The unsubscribe test confirmed the newer unsubscribe links win over older ones for the same sender.
- The end-to-end `RunAll` test verified that senders, threads, and unsubscribe metadata all land in the same SQLite file when executed in sequence.

### What didn't work
- No implementation failure blocked this step.

### What I learned
- Keeping one-click data as a runtime report instead of a stored sender column is workable for the ticket scope, but it means any CLI that wants sender-level one-click output has to derive it from the message table rather than the sender table alone.

### What was tricky to build
- The main sharp edge was deciding how incremental unsubscribe runs should behave after a sender already has `has_list_unsubscribe = TRUE`. The symptom is ambiguity: should new mail update the stored links or should the pass skip already processed senders? I followed the design doc and made incremental runs skip those senders unless `Rebuild` is requested, which keeps the pass idempotent and predictable.

### What warrants a second pair of eyes
- If the product later needs historically accurate unsubscribe-link changes over time, the current sender-level storage model will be too lossy and should probably become a separate table.

### What should be done in the future
- Wire the enrichers into `smailnail enrich ...` and then add the optional post-sync mirror hook.

### Code review instructions
- Review [`pkg/enrich/unsubscribe.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/unsubscribe.go) and [`pkg/enrich/all.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/all.go).
- Validate with `go test ./pkg/enrich`.

### Technical details
- Commands run:

```bash
go fmt ./pkg/enrich
go test ./pkg/enrich
docmgr task check --ticket SMN-20260402-ENRICH --id 8,9
git add pkg/enrich/common.go pkg/enrich/unsubscribe.go pkg/enrich/unsubscribe_test.go pkg/enrich/all.go pkg/enrich/all_test.go pkg/enrich/senders_test.go
git commit -m "Add unsubscribe enrichment and RunAll"
```

## Step 5: Add The Enrich Command Group

With the enrichers stable, I added a dedicated `smailnail enrich` command group under `cmd/smailnail/commands/enrich`. I deliberately kept this step as CLI plumbing only: command descriptions, Glazed field definitions, DB bootstrap/open helpers, and output rows wired to the existing enrichers.

That separation paid off because the command code could be validated mostly as compile-time wiring. The backend behavior had already been locked by the previous commits and tests, so this step stayed narrow.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish the user-facing command surface after the backend implementation is stable.

**Inferred user intent:** Make the new enrichment capabilities actually callable from the CLI instead of leaving them as package-only functionality.

**Commit (code):** `0f22a5c901b58d32595c904e985126e57da51ea3` — "Add smailnail enrich command group"

### What I did
- Added [`cmd/smailnail/commands/enrich/root.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/root.go) to define the `enrich` Cobra group.
- Added [`cmd/smailnail/commands/enrich/common.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/common.go) with shared settings, option conversion, and DB bootstrapping/open helpers.
- Added Glazed verbs in [`cmd/smailnail/commands/enrich/senders.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/senders.go), [`cmd/smailnail/commands/enrich/threads.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/threads.go), [`cmd/smailnail/commands/enrich/unsubscribe.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/unsubscribe.go), and [`cmd/smailnail/commands/enrich/all.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/all.go).
- Updated [`cmd/smailnail/main.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/main.go) to register the new group on the root command.
- Ran `go fmt ./cmd/smailnail/...` and `go test -tags sqlite_fts5 ./cmd/smailnail/... ./pkg/enrich ./pkg/mirror`, then let the commit hook re-run the full repo checks.
- Checked off ticket tasks `10` and `11` and updated the changelog.

### Why
- Users need a stable command surface for post-hoc enrichment.
- Splitting the CLI group from the mirror hook made it possible to verify the commands independently before changing sync behavior.

### What worked
- The new command package compiled cleanly on the first focused validation pass.
- The unsubscribe CLI row output was able to derive a sender-level `one_click` signal from the message table without expanding the schema again.

### What didn't work
- No runtime failure blocked this step.

### What I learned
- The simplest way to guarantee schema availability from the CLI layer is to bootstrap through the mirror store first, then open a plain `sqlx` handle for the actual enrichers.

### What was tricky to build
- The subtle part here was avoiding behavior drift in the command layer. It is easy for a CLI wrapper to start inventing new defaults or alternate code paths; the symptom is a user-visible command that behaves differently from direct package calls. I kept the commands thin by converting flags directly into `pkg/enrich.Options` and delegating the real work to the existing enrichers.

### What warrants a second pair of eyes
- The command helper currently bootstraps a hidden `.smailnail-enrich/raw` mirror root beside the SQLite file to satisfy the existing mirror bootstrap API. That side effect is small, but if a cleaner schema-only bootstrap path is later added in `pkg/mirror`, the command helper should switch to it.

### What should be done in the future
- Add the mirror post-sync hook and surface the combined enrichment counts in the mirror output row.

### Code review instructions
- Review [`cmd/smailnail/commands/enrich/common.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/common.go) and [`cmd/smailnail/commands/enrich/root.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/root.go) first.
- Then spot-check one concrete verb, such as [`cmd/smailnail/commands/enrich/all.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/all.go), against the corresponding enricher.

### Technical details
- Commands run:

```bash
go fmt ./cmd/smailnail/...
go test -tags sqlite_fts5 ./cmd/smailnail/... ./pkg/enrich ./pkg/mirror
docmgr task check --ticket SMN-20260402-ENRICH --id 10,11
git add cmd/smailnail/commands/enrich cmd/smailnail/main.go
git commit -m "Add smailnail enrich command group"
```

## Step 6: Add Mirror Post-Sync Enrichment

The final code change was small but user-facing: `smailnail mirror` can now optionally run the full enrichment pipeline immediately after a successful sync. I kept that hook intentionally non-fatal; mirror sync success should remain the primary outcome, with enrichment failures surfaced as warnings and omitted summary fields rather than turning the whole sync red.

This step closed the last open task in the ticket and finished the end-to-end flow the design doc described.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish the ticket by integrating the enrichment flow into the existing mirror command.

**Inferred user intent:** Make enrichment easy to opt into during normal mirroring, not just as a separate cleanup command.

**Commit (code):** `0a86de84f47c4cb637e1e3b7f939551ae2f9c130` — "Add mirror enrich-after hook"

### What I did
- Updated [`cmd/smailnail/commands/mirror.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go) to add the `--enrich-after` flag, invoke `pkg/enrich.RunAll` after a successful sync, and expose the returned enrichment counts on the mirror output row.
- Scoped the post-sync enrichment to the synced account and to the selected mailbox unless `--all-mailboxes` is active.
- Logged post-sync enrichment failures as warnings instead of returning them as mirror command failures.
- Checked off ticket task `12` and updated the changelog.

### Why
- `mirror --enrich-after` is the ergonomic path the design doc explicitly calls for.
- Keeping the hook non-fatal preserves a clean separation between sync reliability and post-processing reliability.

### What worked
- The compile/test pass stayed green after modifying the existing mirror command.
- The row output now has a straightforward place to expose enrichment counts to downstream scripts.

### What didn't work
- No new failure surfaced in this step.

### What I learned
- The mirror command already had enough structured-row output that adding post-sync enrichment metrics was straightforward once `RunAll` existed.

### What was tricky to build
- The main risk was integrating into an actively evolving `mirror.go` without trampling unrelated local edits. The symptom would have been a merge-like regression inside a file that was already carrying additional date-filtering work. I avoided that by re-reading the file immediately before patching it and then keeping the final change tightly localized to one new flag, one post-sync call site, and a few extra output fields.

### What warrants a second pair of eyes
- The post-sync path currently logs enrichment failures and returns success for the mirror command. That is the intended behavior here, but if operators later need stricter CI semantics they may want a dedicated `--enrich-after-strict` flag rather than changing the default.

### What should be done in the future
- N/A

### Code review instructions
- Review [`cmd/smailnail/commands/mirror.go`](/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go) around the new `EnrichAfter` field, the extra Glazed flag definition, and the post-sync `RunAll` call.
- Validate with `go test -tags sqlite_fts5 ./...`.

### Technical details
- Commands run:

```bash
docmgr task check --ticket SMN-20260402-ENRICH --id 12
git add cmd/smailnail/commands/mirror.go
git commit -m "Add mirror enrich-after hook"
```
