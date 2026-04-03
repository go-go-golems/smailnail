---
Title: Analysis, design, and implementation plan for a mirror shard merge verb
Ticket: SMN-20260402-MIRROR-MERGE
Status: active
Topics:
    - mirror
    - sqlite
    - backfill
    - cli
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/smailnail/commands/merge_mirror_shards.go
      Note: Initial Glazed command implementation for merge-mirror-shards
    - Path: cmd/smailnail/commands/mirror.go
      Note: Reference command shape and Glazed field/output patterns for the future merge verb
    - Path: cmd/smailnail/main.go
      Note: Register the new merge verb alongside mirror and enrich commands
    - Path: pkg/mirror/files.go
      Note: Relative raw-path contract and raw message write semantics
    - Path: pkg/mirror/merge.go
      Note: Dry-run merge service
    - Path: pkg/mirror/merge_test.go
      Note: Initial tests for shard discovery
    - Path: pkg/mirror/schema.go
      Note: Mirror schema
    - Path: pkg/mirror/service.go
      Note: Current mirror sync lifecycle
    - Path: pkg/mirror/store.go
      Note: Destination bootstrap path for schema and mirror-root creation
    - Path: ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/run-last-24-months-backfill.sh
      Note: Current month-sharded producer workflow that the merge verb must consume
ExternalSources: []
Summary: Detailed intern-facing design guide for a Go verb that merges month-sharded mirror databases and raw-message trees into one durable local mirror.
LastUpdated: 2026-04-02T16:14:59.716157724-04:00
WhatFor: Explain how to add a first-class smailnail command that merges month-sharded mirror databases and raw-message trees into one local mirror.
WhenToUse: Use when implementing, reviewing, or extending a merge verb for parallel mirror backfills.
---



# Analysis, design, and implementation plan for a mirror shard merge verb

## Executive Summary

`smailnail mirror` can now create durable local mirror shards, and the ticket scripts can backfill many month slices in parallel, but the system still lacks a first-class way to consolidate those shards into one usable local mirror. Today, a 24-month backfill produces one shard directory per month, each with its own `mirror.sqlite` and `raw/` tree. That is good for parallel downloading, but it is not yet a good end-state for local search, incremental resync, or further enrichment.

The recommended next feature is a Go verb, tentatively named `smailnail merge-mirror-shards`, backed by a dedicated merge service in `pkg/mirror`. The verb should discover shard directories, validate that they are mergeable, bootstrap a fresh destination mirror, copy canonical `messages` rows and raw `.eml` files into that destination, rebuild `mailbox_sync_state`, rebuild `messages_fts`, and leave derived enrichment tables to a separate enrichment pass. The implementation should optimize first for correctness, determinism, and testability, not for clever SQL shortcuts.

This document is written for a new intern. It explains the current system, the data model, why shell scripts are currently enough for sharding but not enough for merging, the exact design choices to make, where to change code, and how to validate the finished behavior.

## Problem Statement

The project now supports parallel month-sharded backfills. The script at `ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/run-last-24-months-backfill.sh` writes one directory per shard and launches up to six parallel workers. Each worker runs `smailnail mirror --since-date ... --before-date ...` into its own shard-local SQLite database and raw-message tree. That workflow works, but it stops at a fragmented set of artifacts rather than producing one consolidated mirror.

For the user, the missing capability is: “I ran a parallel backfill. Now give me one SQLite database and one mirror root I can keep using.” That capability is harder than a naive SQL copy because the local mirror is made of several coordinated pieces:

- canonical message rows in `messages`,
- canonical raw `.eml` files under the mirror root,
- mutable checkpoint state in `mailbox_sync_state`,
- standalone FTS state in `messages_fts`,
- schema migrations that also include enrichment tables.

The merge verb therefore needs to answer several non-trivial questions:

- How do we discover valid shards?
- How do we detect incompatible inputs?
- Which tables are canonical and should be copied, and which tables are derived and should be rebuilt?
- How do we keep `raw_path` valid after moving raw files into a new destination root?
- How do we deal with duplicate rows, overlapping shards, and UIDVALIDITY conflicts?
- How do we ensure the merged result remains incremental-friendly for future `smailnail mirror` runs?

The rest of this guide exists to answer those questions in a way that is implementable and testable.

## Scope And Non-Goals

### In scope

- A new top-level Glazed verb in `smailnail` for merging shard mirrors.
- Discovery of shard directories under an input root.
- Validation and preflight reporting.
- Copying and deduplicating canonical message rows.
- Copying and validating raw `.eml` files.
- Recomputing `mailbox_sync_state`.
- Rebuilding `messages_fts`.
- Emitting a structured report that an operator can inspect in JSON or table form.

### Explicit non-goals for v1

- Replacing the existing shell scripts that create parallel shards.
- Performing the parallel backfill itself inside the CLI.
- Merging enrichment tables directly.
- Supporting in-place merge into one existing shard database.
- Supporting arbitrary “merge whatever SQLite databases I point at” with no directory conventions.

## Terms And Mental Model

Before touching code, define the main terms precisely.

- **Mirror**: the local durable representation of a mailbox/account, made of one SQLite DB plus one mirror root containing raw `.eml` files.
- **Shard**: one partial mirror produced by a bounded sync, usually one month, stored in one directory with its own `mirror.sqlite` and `raw/`.
- **Destination mirror**: the fresh output mirror produced by the new merge verb.
- **Canonical persisted state**: data that should be copied directly because it is the source of truth. In this system, that means `messages` rows and raw `.eml` files.
- **Derived state**: data that can be regenerated from canonical persisted state. In this system, that means `messages_fts`, `mailbox_sync_state`, and enrichment outputs.
- **Checkpoint**: the state needed so future incremental syncs know where to resume. Here that is `mailbox_sync_state`.

If you remember nothing else from this design, remember this split:

```text
Copy directly:
  - messages
  - raw/*.eml

Rebuild deliberately:
  - mailbox_sync_state
  - messages_fts
  - enrichment tables
```

## Current-State Architecture

This section is evidence-based. Every major claim below maps back to an existing file.

### 1. The current CLI surface is the `mirror` verb

The existing mirror command is defined in `cmd/smailnail/commands/mirror.go`. Its settings struct already includes IMAP connection parameters, local output paths, batching controls, mailbox filters, date bounds, and reconcile/reset options (`cmd/smailnail/commands/mirror.go:24-42`). The command builds a Glazed section named `mirror` with fields such as `sqlite-path`, `mirror-root`, `batch-size`, `since-date`, `before-date`, `reconcile-full-mailbox`, and `enrich-after` (`cmd/smailnail/commands/mirror.go:55-145`).

Operationally, the command does three things:

1. open a local store,
2. bootstrap schema and mirror root,
3. call `mirror.NewService(store).Sync(...)`,
4. optionally run enrichment after sync,
5. emit one structured output row (`cmd/smailnail/commands/mirror.go:188-330`).

That matters because the new merge command should follow the same command shape:

- Glazed settings struct,
- store bootstrap,
- service object in `pkg/mirror`,
- one structured result row.

### 2. The current mirror sync service downloads mail and writes canonical local state

`pkg/mirror/service.go` is the heart of the producer side. `SyncOptions` includes the network parameters and all local sync-scoping knobs (`pkg/mirror/service.go:21-40`). `Service.Sync` normalizes and validates options, opens the IMAP session, resolves mailboxes, loops through them, and aggregates a `SyncReport` (`pkg/mirror/service.go:70-198`).

The per-mailbox flow is:

1. get mailbox status,
2. load or reset `mailbox_sync_state`,
3. select the mailbox,
4. search UIDs, respecting date bounds and max-message limits,
5. fetch messages in batches,
6. persist each batch,
7. optionally reconcile a full mailbox snapshot,
8. upsert a final `mailbox_sync_state` row (`pkg/mirror/service.go:312-556`).

The persistence path is especially important for merge design:

- `persistBatch` writes raw `.eml` files before writing rows (`pkg/mirror/service.go:642-712`).
- `buildMessageRecord` assembles the row payload, using raw RFC 822 parsing to normalize headers, address summaries, parts, text, and HTML when parsing succeeds (`pkg/mirror/service.go:714-806`).
- `upsertMessageRecord` upserts by `(account_key, mailbox_name, uidvalidity, uid)` (`pkg/mirror/service.go:1025-1082`).

This is the uniqueness contract the merge verb must preserve.

### 3. The schema tells us what is canonical and what is derived

The schema is bootstrapped in `pkg/mirror/schema.go`.

Observed facts:

- Schema version is currently `2` (`pkg/mirror/schema.go:14-17`).
- `messages` contains the core mirrored row model, including `raw_path`, `raw_sha256`, `remote_deleted`, and timestamps (`pkg/mirror/schema.go:39-66`).
- `mailbox_sync_state` is keyed by `(account_key, mailbox_name)` and stores one `uidvalidity`, one `highest_uid`, one `last_uidnext`, and one `last_sync_at` (`pkg/mirror/schema.go:29-38`).
- `messages_fts` is a separate virtual table created by `bootstrapFTS`, not an automatically maintained shadow table (`pkg/mirror/schema.go:136-156`).

Two consequences follow immediately:

1. `mailbox_sync_state` is not shard-native state. It is a destination checkpoint that must be recomputed for the merged output.
2. `messages_fts` is derived state that must be rebuilt after all rows are merged.

Also note that schema migration version `2` delegates to `enrich.SchemaMigrationV2Statements()` (`pkg/mirror/schema.go:75-78`). That means enrichment tables are part of the same DB schema, but not necessarily part of what the merge should copy directly.

### 4. Raw files are stored under a relative path contract

The raw-file contract is in `pkg/mirror/files.go`.

`RawMessagePath` returns:

```text
raw/<accountKey>/<mailboxSlug>/<uidvalidity>/<uid>.eml
```

(`pkg/mirror/files.go:23-31`)

`WriteRawMessage` then joins that relative path under the configured mirror root, writes atomically via a temp file + rename, and stores the SHA-256 of the contents (`pkg/mirror/files.go:33-92`).

This matters for merge because:

- `messages.raw_path` stores the relative path, not an absolute path.
- A merged destination must therefore physically contain the copied raw files at the same relative path under its own `mirror-root`.
- If two shards contain the same relative path but different bytes, that is a real integrity conflict, not a cosmetic mismatch.

### 5. The sharding workflow is currently script-driven

The live parallel backfill manager script lives at:

`ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/run-last-24-months-backfill.sh`

What it does:

- computes monthly `[since-date, before-date)` shard ranges (`:40-52`),
- creates one shard directory per month (`:55-73`),
- runs `go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror ... --since-date ... --before-date ... --sqlite-path <shard>/mirror.sqlite --mirror-root <shard>/raw` (`:74-86`),
- tracks state files such as `state`, `exit_code`, `started_at`, and `finished_at` (`:65-95`),
- runs up to six workers in parallel (`:100-147`).

The dashboard script reads the same directory layout and shows rows from `messages` in each shard-local SQLite database (`dashboard-last-24-months-backfill.sh:33-85`).

This is good news for the merge verb. The input layout already exists and is stable enough to target.

## Gap Analysis

The current system is missing exactly one core user-facing step: consolidation.

### What exists today

- bounded mirror sync into one local SQLite DB and one raw root,
- date-ranged month sharding,
- multi-process backfill orchestration,
- dashboard monitoring for shard runs.

### What does not exist yet

- one command that consumes those shard directories and produces a single destination mirror.

### Why a shell script is not enough anymore

A shell script could copy SQLite rows and raw files, but that would leave too many subtle failure modes under-tested:

- overlap and dedupe behavior,
- raw-path conflicts,
- missing raw files,
- UIDVALIDITY conflicts,
- checkpoint reconstruction,
- FTS rebuild correctness,
- future integration with enrichment.

These are data-model concerns, not mere operator orchestration concerns. That is why the merge belongs in Go.

## Proposed Solution

Implement a new first-class Glazed verb:

```text
smailnail merge-mirror-shards \
  --input-root /tmp/smailnail-last-24-months-backfill \
  --output-sqlite /tmp/smailnail-last-24-months-merged.sqlite \
  --output-mirror-root /tmp/smailnail-last-24-months-merged
```

### Recommended command shape

Suggested required flags:

- `--input-root`
- `--output-sqlite`
- `--output-mirror-root`

Suggested optional flags:

- `--shard-glob`
- `--dry-run`
- `--fail-on-missing-raw`
- `--copy-raw` default `true`
- `--rebuild-fts` default `true`
- `--rebuild-sync-state` default `true`
- `--allow-overwrite-destination` default `false`
- `--enrich-after` default `false`, but supported in v1 so the merged destination can be enriched immediately after a successful merge

### Recommended package layout

Add these files:

- `cmd/smailnail/commands/merge_mirror_shards.go`
- `pkg/mirror/merge.go`
- `pkg/mirror/merge_test.go`

Modify:

- `cmd/smailnail/main.go` to register the new command.

Optional later additions:

- `cmd/smailnail/docs/mirror-merge-overview.md`
- `cmd/smailnail/docs/mirror-merge-tutorial.md`

## Detailed Design

### Design principle 1: merge into a fresh destination, never into an existing shard

The destination should be newly bootstrapped via `mirror.OpenStore(...).Bootstrap(...)`, the same way the existing `mirror` command boots a fresh mirror (`cmd/smailnail/commands/mirror.go:213-239`, `pkg/mirror/store.go:41-60`).

Why:

- rollback is simpler,
- invariants are clearer,
- test setup is cleaner,
- operators avoid accidentally mutating one shard into a half-merged hybrid.

### Design principle 2: treat `messages` and raw `.eml` files as canonical

These artifacts should be copied forward into the destination.

Why:

- they are the durable output of the fetch pipeline,
- they already contain parsed headers, body text/HTML, attachment metadata, and search text,
- raw `.eml` files are the byte-level canonical source and must survive the merge.

### Design principle 3: rebuild derived state instead of trying to merge it

The merge verb should rebuild:

- `mailbox_sync_state`
- `messages_fts`

It should not directly merge:

- enrichment tables

Why:

- `mailbox_sync_state` should reflect the merged destination, not individual shard histories,
- `messages_fts` is standalone and should be repopulated from final `messages`,
- enrichment tables are derivative and can be recomputed from the merged canonical message set.

### Design principle 4: fail fast on structural conflicts

The merge should detect and reject:

- no shards found,
- shard schema too old/new to understand,
- destination already exists unless explicitly allowed,
- duplicate unique keys with conflicting raw hashes,
- multiple `uidvalidity` values for the same `(account_key, mailbox_name)` unless a future explicit override is added,
- missing raw files when strict mode is enabled.

Correctness matters more than squeezing through a dubious merge.

## Proposed API Sketch

The intern should add a separate merge service rather than overloading `Service.Sync`.

```go
type MergeOptions struct {
    InputRoot                string
    OutputSQLitePath         string
    OutputMirrorRoot         string
    ShardGlob                string
    DryRun                   bool
    CopyRaw                  bool
    FailOnMissingRaw         bool
    RebuildFTS               bool
    RebuildSyncState         bool
    AllowOverwriteDestination bool
}

type ShardInfo struct {
    Name            string
    Root            string
    SQLitePath      string
    RawRoot         string
    SchemaVersion   int
    MessageCount    int
    AccountKeys     []string
    Mailboxes       []string
    UIDValidityByMailbox map[string][]uint32
}

type MergeReport struct {
    Status                string
    InputRoot             string
    OutputSQLitePath      string
    OutputMirrorRoot      string
    ShardsDiscovered      int
    ShardsMerged          int
    MessagesScanned       int
    MessagesInserted      int
    MessagesUpdated       int
    RawFilesCopied        int
    RawFilesReused        int
    RawFilesMissing       int
    RawConflicts          int
    UIDValidityConflicts  int
    MailboxStatesRebuilt  int
    FTSRowsRebuilt        int
}

type MergeService struct {
    now func() time.Time
}
```

This should live next to the existing mirror service, not inside the command package.

## Merge Flow

### High-level flow

```text
discover shards
  -> inspect shard metadata
  -> validate compatibility
  -> bootstrap fresh destination
  -> merge canonical rows + raw files
  -> rebuild mailbox_sync_state
  -> rebuild messages_fts
  -> optionally trigger enrichment later
  -> emit report
```

### Suggested runtime diagram

```text
┌────────────────────────┐
│ merge-mirror-shards    │
│ Glazed command         │
└───────────┬────────────┘
            │
            v
┌────────────────────────┐
│ MergeService           │
│ pkg/mirror/merge.go    │
└───────────┬────────────┘
            │
            ├──────────── discover shard dirs
            │
            ├──────────── inspect shard DB metadata
            │
            ├──────────── bootstrap destination store/root
            │
            ├──────────── copy messages rows
            │
            ├──────────── copy raw/*.eml files
            │
            ├──────────── rebuild mailbox_sync_state
            │
            └──────────── rebuild messages_fts
```

### Suggested data-flow diagram

```text
input-root/
  2024-05/
    mirror.sqlite
    raw/...
  2024-06/
    mirror.sqlite
    raw/...
  ...

           merge
             │
             v

destination/
  merged.sqlite
  raw/...
```

## Recommended Merge Algorithm

This is the most important section for the implementation.

### Phase A: Discover shard candidates

Expected shard directory contract:

- `<input-root>/<shard-name>/mirror.sqlite`
- `<input-root>/<shard-name>/raw/`

Discovery algorithm:

1. list direct children under `input-root`,
2. filter by `shard-glob` if present,
3. keep only directories containing `mirror.sqlite`,
4. sort by shard name for deterministic behavior.

Pseudocode:

```go
func discoverShards(inputRoot, shardGlob string) ([]ShardInfo, error) {
    entries := readDir(inputRoot)
    candidates := []ShardInfo{}
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }
        if shardGlob != "" && !matches(shardGlob, entry.Name()) {
            continue
        }
        sqlitePath := filepath.Join(inputRoot, entry.Name(), "mirror.sqlite")
        if !exists(sqlitePath) {
            continue
        }
        candidates = append(candidates, ShardInfo{
            Name: entry.Name(),
            Root: filepath.Join(inputRoot, entry.Name()),
            SQLitePath: sqlitePath,
            RawRoot: filepath.Join(inputRoot, entry.Name(), "raw"),
        })
    }
    sortByName(candidates)
    if len(candidates) == 0 {
        return nil, errors.New("no mergeable shards found")
    }
    return candidates, nil
}
```

### Phase B: Inspect each shard before touching the destination

For each shard:

- open SQLite read-only if convenient,
- read schema version from `mirror_metadata`,
- count `messages`,
- list distinct `(account_key, mailbox_name, uidvalidity)`,
- verify that the shard DB is structurally understandable,
- optionally spot-check that the raw root exists.

Why this preflight matters:

- operators should see incompatibilities before any destination is mutated,
- the merge report can show exactly what will be merged,
- conflict checks need global context.

### Phase C: Validate mergeability

Recommended validation rules:

1. all shards must have a supported schema version,
2. each shard must contain at least one message or be explicitly tolerated,
3. if the same `(account_key, mailbox_name)` appears with more than one `uidvalidity` across all shards, fail in v1,
4. destination paths must not already exist unless overwrite is explicitly allowed.

UIDVALIDITY conflict handling deserves emphasis.

Why fail on mixed UIDVALIDITY in v1:

- `messages` can technically store multiple `uidvalidity` generations,
- but `mailbox_sync_state` stores only one current generation per mailbox,
- the merged destination would otherwise not have an unambiguous checkpoint for future incremental syncs.

This is one of the most important correctness rules in the whole design.

### Phase D: Bootstrap the destination

Do this exactly the way the mirror command already does:

```go
store, err := mirror.OpenStore(opts.OutputSQLitePath)
report, err := store.Bootstrap(ctx, opts.OutputMirrorRoot)
```

That gives the destination:

- schema tables,
- FTS table,
- enrichment tables,
- `raw/` root directory.

Even though the merge will later rebuild FTS, the destination still needs the table structure up front.

### Phase E: Merge canonical message rows and raw files

This is the core merge loop.

Recommended v1 choice:

- implement the merge row-by-row in Go,
- use transactions in batches per shard,
- reuse `MessageRecord` and a dedicated destination upsert helper,
- do not start with an `ATTACH`-heavy SQL bulk merge.

Why this is the best v1 choice:

- easier to reason about,
- easier to unit test,
- easier to combine with raw-file copying and hash verification,
- easier to surface per-record conflicts cleanly.

Per message row:

1. read the row from the shard,
2. confirm the raw file exists under `<shard>/raw/<record.RawPath>` if `CopyRaw` is enabled,
3. if the destination raw file does not exist, copy it,
4. if it exists:
   - if SHA matches, reuse it,
   - if SHA differs, fail with a conflict,
5. upsert the message row into destination `messages`,
6. track whether the row was inserted or updated.

Recommended raw copy rule:

```text
same relative path + same SHA   -> reuse
same relative path + different SHA -> hard error
missing shard raw file + strict mode -> hard error
missing shard raw file + non-strict mode -> count warning, continue
```

Suggested pseudocode:

```go
for _, shard := range shards {
    rows := loadShardMessages(shard.SQLitePath)
    tx := dest.Begin()
    for _, record := range rows {
        report.MessagesScanned++

        if opts.CopyRaw {
            src := filepath.Join(shard.RawRoot, record.RawPath)
            dst := filepath.Join(opts.OutputMirrorRoot, record.RawPath)
            copyResult, err := ensureRawFile(src, dst, record.RawSHA256, opts.FailOnMissingRaw)
            if err != nil {
                return err
            }
            updateRawCounters(report, copyResult)
        }

        action, err := upsertMergedMessage(ctx, tx, record)
        if err != nil {
            return err
        }
        updateMessageCounters(report, action)
    }
    tx.Commit()
    report.ShardsMerged++
}
```

### Phase F: Rebuild `mailbox_sync_state`

Do not copy shard checkpoint rows directly.

Instead, rebuild by grouping destination `messages`:

- group by `(account_key, mailbox_name)`,
- read the single allowed `uidvalidity`,
- compute `highest_uid = MAX(uid)`,
- compute `last_uidnext = highest_uid + 1` as a best-effort local checkpoint,
- compute `last_sync_at = MAX(last_synced_at)`,
- set `status = 'active'`.

Why reconstruct this way:

- it aligns the checkpoint with the merged destination contents,
- it avoids carrying per-shard partial state into a supposedly complete merged mirror.

Important note:

`last_uidnext = highest_uid + 1` is an approximation, not the authoritative remote `UIDNEXT`. That is acceptable for a merge baseline because the next actual sync will fetch mailbox status from IMAP and refresh the checkpoint. What matters most is having a sane `highest_uid` and a current `uidvalidity`.

### Phase G: Rebuild `messages_fts`

Observed behavior:

- `messages_fts` is created at bootstrap in `pkg/mirror/schema.go:136-156`.
- The current upsert path in `pkg/mirror/service.go:1025-1082` does not directly update FTS.

Therefore the merge service should rebuild the FTS table explicitly after all rows are in place.

Recommended rebuild flow:

1. `DELETE FROM messages_fts`
2. `INSERT INTO messages_fts(rowid, account_key, mailbox_name, subject, from_summary, to_summary, cc_summary, body_text, body_html, search_text) SELECT ... FROM messages`

Why use `rowid = messages.id`:

- it keeps FTS rows aligned with the primary table ids,
- it makes later search joins predictable.

### Phase H: Derived enrichment state

Because schema version `2` includes enrichment migrations, the destination DB will already have those tables after bootstrap. However, v1 of merge should not attempt to merge them.

Recommendation:

- leave enrichment tables empty or as bootstrapped defaults before the merge completes,
- support `--enrich-after` in v1 by running `enrich.RunAll(...)` against the merged destination after canonical rows, mailbox state, and FTS have been rebuilt,
- keep enrichment opt-in so the base merge path remains predictable and easier to debug.

## Why The Merge Should Be A Go Verb, Not Just A Script

This deserves its own explicit argument because it is the main product-boundary decision.

### Keep as scripts

- parallel worker orchestration,
- tmux dashboard,
- policies like “last 24 months with 6 workers”.

### Promote to Go

- shard discovery contract,
- merge validation,
- copy/upsert semantics,
- raw-file integrity rules,
- checkpoint rebuild,
- FTS rebuild,
- structured operator report.

The easiest way to remember the boundary is:

```text
Scripts answer: how do we fan out work?
Go verb answers: what does a correct merged mirror mean?
```

## Alternatives Considered

### Alternative 1: keep merge as an ad hoc shell script

Rejected for v1.

Why:

- hard to test,
- hard to express integrity rules,
- hard to reuse,
- easy to accidentally merge into a half-invalid destination.

### Alternative 2: implement merge purely with SQLite `ATTACH` and `INSERT ... SELECT`

Reasonable later optimization, but not the best first implementation.

Pros:

- potentially faster for very large merges,
- less Go iteration overhead.

Cons:

- harder to interleave with raw-file validation/copying,
- harder to surface useful per-conflict diagnostics,
- harder for a new engineer to reason about.

Recommendation:

- start with Go-driven row iteration,
- optimize with bulk SQL only if profiling later shows it is necessary.

### Alternative 3: merge into one existing shard DB in place

Rejected.

Why:

- confusing rollback story,
- easy to leave the shard half-mutated,
- destination semantics become ambiguous,
- harder to tell operators what is safe.

### Alternative 4: redesign backfill to write into one shared SQLite DB directly

Rejected for now.

Why:

- it changes the parallel download architecture,
- concurrent writes to one SQLite DB would create new locking and coordination issues,
- it is much riskier than adding a safe merge step.

## File-Level Implementation Plan

This section is the handoff checklist for the intern.

### Phase 1: Command scaffolding

Files:

- new `cmd/smailnail/commands/merge_mirror_shards.go`
- modify `cmd/smailnail/main.go`

Tasks:

1. define a `MergeMirrorShardsCommand`,
2. add a settings struct with Glazed tags,
3. create a `mirror-merge` section with fields,
4. instantiate the merge service,
5. output one row using `types.NewRow()`,
6. register the command in `main.go`.

Review target:

- mirror command style in `cmd/smailnail/commands/mirror.go:44-173`, `:175-330`.

### Phase 2: Merge service skeleton

Files:

- new `pkg/mirror/merge.go`
- new `pkg/mirror/merge_test.go`

Tasks:

1. define `MergeOptions`, `ShardInfo`, and `MergeReport`,
2. implement shard discovery,
3. implement shard inspection,
4. implement destination bootstrap,
5. return a dry-run report without mutation.

### Phase 3: Canonical row and raw-file merge

Files:

- `pkg/mirror/merge.go`
- maybe a helper file such as `pkg/mirror/merge_files.go` if it grows too large

Tasks:

1. implement streaming or batch row reads from shard DBs,
2. implement raw-file copy/reuse/conflict logic,
3. implement destination upsert helper,
4. add counters for inserted/updated/reused/copied/conflict cases.

### Phase 4: Rebuild derived state

Files:

- `pkg/mirror/merge.go`

Tasks:

1. add `rebuildMailboxSyncState(...)`,
2. add `rebuildMessagesFTS(...)`,
3. document enrichment follow-up behavior in the report and help.

### Phase 5: Test hardening

Files:

- `pkg/mirror/merge_test.go`

Tasks:

1. test shard discovery,
2. test empty input root,
3. test overlapping duplicate rows with same raw SHA,
4. test overlapping duplicate rows with different raw SHA,
5. test missing raw files in strict and non-strict mode,
6. test UIDVALIDITY conflict failure,
7. test mailbox-sync-state reconstruction,
8. test FTS rebuild row counts,
9. test end-to-end merge of two or more small shard fixtures.

### Phase 6: Docs and operator polish

Files:

- `cmd/smailnail/docs/...`
- `README.md`
- `cmd/smailnail/README.md`

Tasks:

1. add help pages,
2. add examples that follow the current backfill scripts,
3. document post-merge enrichment workflow,
4. document the dry-run mode and integrity flags.

## Test Strategy

### Unit tests

Use temp directories and temp SQLite DBs.

Scenarios:

- discover shards under a temp input root,
- merge disjoint shard databases,
- merge duplicate rows with matching raw hashes,
- detect raw conflicts,
- detect UIDVALIDITY conflicts,
- verify rebuilt `mailbox_sync_state`,
- verify `messages_fts` row count equals message row count.

### Integration tests in-repo

Build small synthetic shard DBs by:

1. bootstrapping a temp store,
2. inserting `MessageRecord` rows,
3. creating raw files under matching `raw_path`s,
4. invoking the merge service.

This gives fast, deterministic tests.

### Docker Compose IMAP fixture tests

When appropriate, test against the repo’s Docker Compose IMAP server flow because that is the closest thing to a production-like producer path.

Suggested end-to-end scenario:

1. run two or three month-bounded `smailnail mirror` commands against the Docker fixture into separate shard directories,
2. run the merge verb,
3. verify:
   - destination message count equals sum of shard rows,
   - destination raw files exist,
   - destination FTS can answer a known text query,
   - a follow-up `smailnail mirror` against the merged DB behaves incrementally.

This is the highest-signal integration test because it exercises both the producer and the merger.

## Risks, Sharp Edges, And Recommended Policies

### Risk 1: UIDVALIDITY conflicts

Recommendation:

- fail in v1,
- explain why clearly in the error,
- add an explicit override only later if there is a real need.

### Risk 2: raw-file mismatches

Recommendation:

- hard-fail when the same relative path maps to different SHA values.

### Risk 3: merging derived enrichment tables

Recommendation:

- do not do it in v1,
- rerun enrichment after merge instead.

### Risk 4: operator confusion about destination overwrite

Recommendation:

- default to refusing existing destination paths,
- allow explicit overwrite only behind a deliberate flag.

### Risk 5: FTS correctness drift

Recommendation:

- rebuild FTS from scratch every time,
- do not try to incrementally patch FTS during merge until tests prove it necessary.

## Open Questions

These are the questions worth resolving before or during implementation review.

1. Should the report include per-shard details similar to `SyncReport.Mailboxes`, or is aggregate reporting enough initially?

My recommendation:

- v1: support `--enrich-after`,
- v1: missing raw files warn by default and are counted in the report,
- v1: root-based discovery only,
- v1: aggregate report plus a short per-shard summary.

## Concrete Implementation Checklist For The Intern

1. Read:
   - `cmd/smailnail/commands/mirror.go`
   - `pkg/mirror/service.go`
   - `pkg/mirror/schema.go`
   - `pkg/mirror/files.go`
   - `pkg/mirror/store.go`
   - `pkg/mirror/types.go`
2. Add a new command file that follows the same Glazed command style as `mirror`.
3. Implement a `MergeService` in `pkg/mirror`.
4. Build dry-run discovery and inspection first.
5. Add destination bootstrap.
6. Add canonical row + raw-file merge logic.
7. Add `mailbox_sync_state` rebuild.
8. Add `messages_fts` rebuild.
9. Add unit tests for all conflict and success paths.
10. Only after tests are green, add user docs and examples.

Do not start by trying to optimize the SQL path. Get the semantics correct first.

## References

### Key code files

- `cmd/smailnail/main.go`
- `cmd/smailnail/commands/mirror.go`
- `pkg/mirror/service.go`
- `pkg/mirror/schema.go`
- `pkg/mirror/store.go`
- `pkg/mirror/files.go`
- `pkg/mirror/types.go`

### Key script references

- `ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/run-last-24-months-backfill.sh`
- `ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/dashboard-last-24-months-backfill.sh`

### Most relevant line anchors

- `cmd/smailnail/commands/mirror.go:24-42`, `55-145`, `188-330`
- `pkg/mirror/service.go:21-40`, `70-198`, `312-556`, `642-1082`
- `pkg/mirror/schema.go:24-79`, `82-156`
- `pkg/mirror/files.go:23-92`
- `pkg/mirror/store.go:20-60`
- `pkg/mirror/types.go:19-125`
