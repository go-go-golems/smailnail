---
Title: Investigation diary
Ticket: SMN-20260402-MIRROR-MERGE
Status: active
Topics:
    - mirror
    - sqlite
    - backfill
    - cli
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/smailnail/commands/mirror.go
      Note: Diary evidence for current mirror command behavior and operator contract
    - Path: pkg/mirror/files.go
      Note: Diary evidence for raw file naming and SHA semantics
    - Path: pkg/mirror/schema.go
      Note: Diary evidence for schema versioning and derived-state tables
    - Path: pkg/mirror/service.go
      Note: Diary evidence for sync flow
    - Path: ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/dashboard-last-24-months-backfill.sh
      Note: Diary evidence for the shard directory contract observed by operators
ExternalSources: []
Summary: Chronological research log for the merge-shards verb design, including evidence gathered from the current mirror command, service, schema, and scripts.
LastUpdated: 2026-04-02T16:14:59.780360632-04:00
WhatFor: Chronological record of the analysis work used to design the mirror shard merge verb.
WhenToUse: Use when continuing the merge-verb design or reviewing why the proposed architecture was chosen.
---


# Investigation diary

## Goal

Capture the evidence, reasoning, and concrete commands behind the design for a `smailnail` merge-shards verb. This diary is written so a future engineer can understand what was inspected, what constraints were discovered, and why the proposed implementation favors a correctness-first Go service over a shell-only or SQL-only merge.

## Step 1: Create the ticket and map the current mirror architecture

I started by creating a new ticket workspace specifically for the merge-shards problem instead of extending the earlier IMAP mirror ticket further. The previous ticket already documented how month-sharded backfills are produced, but this follow-up work needed its own scope because the merge step is a separate product capability with different failure modes, especially around raw-file integrity, checkpoint reconstruction, and derived index rebuilds.

The main work in this step was evidence gathering. I inspected the current `mirror` verb, the mirror service, the on-disk raw-message layout, the schema bootstrap path, and the month-sharded backfill scripts. That inspection made the central architectural point clear: the missing merge is not just a `sqlite3 ATTACH` convenience script. It has to reconcile database rows, raw `.eml` files, `mailbox_sync_state`, and the standalone `messages_fts` table in one coherent operation.

### Prompt Context

**User prompt (verbatim):** "Alright, create a new docmgr ticket and create a detailed analysis / implementation plan. 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new docmgr ticket dedicated to the merge-shards verb, gather evidence from the current mirror system, and write a detailed intern-facing design/implementation guide plus diary, then upload the bundle to reMarkable.

**Inferred user intent:** Turn the earlier ad hoc discussion about merging month-sharded mirror databases into a formal, actionable design package that another engineer could implement safely.

**Commit (code):** N/A

### What I did
- Created a new ticket workspace with `docmgr ticket create-ticket --ticket SMN-20260402-MIRROR-MERGE --title "Add a mirror shard merge verb for month-sharded backfills" --topics mirror,sqlite,backfill,cli`.
- Added a design doc and diary with `docmgr doc add`.
- Inspected `cmd/smailnail/commands/mirror.go` to understand the current Glazed command surface and output contract.
- Inspected `pkg/mirror/service.go` to understand how sync currently writes rows, raw files, and sync state.
- Inspected `pkg/mirror/schema.go`, `pkg/mirror/store.go`, `pkg/mirror/files.go`, and `pkg/mirror/types.go`.
- Inspected the existing backfill scripts in `ttmp/2026/04/01/.../scripts/run-last-24-months-backfill.sh` and `dashboard-last-24-months-backfill.sh`.

### Why
- The merge verb must fit the current mirror model instead of inventing a parallel data model.
- The main design decisions depend on observed invariants, especially:
  - `raw_path` is relative under the mirror root.
  - `mailbox_sync_state` is keyed by `(account_key, mailbox_name)` rather than by shard.
  - `messages_fts` is created at bootstrap and must be rebuilt deliberately.
  - month sharding is currently orchestrated by scripts, not by a Go command.

### What worked
- The current mirror code is structured cleanly enough to support a merge verb in the same package family.
- The earlier sharding work already created the exact artifact layout the merge verb needs to consume: one shard directory with `mirror.sqlite`, `raw/`, `result.json`, and logs.
- The evidence showed a clear separation of concerns:
  - `mirror` handles fetching and persistence.
  - scripts handle orchestration.
  - the missing product feature is the consolidation step.

### What didn't work
- N/A. This step was analysis-only and did not involve code changes or runtime debugging.

### What I learned
- The current schema version is `2`, and schema bootstrap already pulls in enrichment migrations via `enrich.SchemaMigrationV2Statements()`.
- The merge verb should not blindly copy enrichment tables from shards. Those tables are derived state and should instead be rebuilt or rerun after merge.
- `mailbox_sync_state` cannot safely be copied shard-by-shard because the merged destination needs one coherent checkpoint per `(account_key, mailbox_name)`.

### What was tricky to build
- The trickiest part of the design was not the SQL insert itself. It was identifying which parts of the current mirror are source-of-truth data versus derived state. The raw `.eml` files and `messages` rows are canonical persisted artifacts; `mailbox_sync_state`, `messages_fts`, and enrichment tables are derived or checkpoint state. That distinction drives the merge order and the failure model.

### What warrants a second pair of eyes
- The UIDVALIDITY conflict policy. The current schema lets `messages` contain multiple `uidvalidity` values for one mailbox, but `mailbox_sync_state` only stores one current `uidvalidity`, which makes merge behavior ambiguous if shards span a mailbox reset.
- Whether the first implementation should support merging previously enriched shards or explicitly require rerunning enrichment afterward.

### What should be done in the future
- Implement the merge verb as a new Glazed command and a dedicated `pkg/mirror` merge service.
- Add integration tests that build monthly shards against the Docker Compose IMAP fixture and then merge them.
- Add help pages once the verb exists.

### Code review instructions
- Start with the design doc and compare its claims against:
  - `cmd/smailnail/commands/mirror.go`
  - `pkg/mirror/service.go`
  - `pkg/mirror/schema.go`
  - `pkg/mirror/files.go`
  - `ttmp/2026/04/01/.../scripts/run-last-24-months-backfill.sh`
- Validate the evidence trail with:
  - `nl -ba cmd/smailnail/commands/mirror.go | sed -n '1,360p'`
  - `nl -ba pkg/mirror/service.go | sed -n '1,1120p'`
  - `nl -ba pkg/mirror/schema.go | sed -n '1,220p'`
  - `nl -ba pkg/mirror/files.go | sed -n '1,120p'`

### Technical details
- Ticket creation:
```bash
docmgr ticket create-ticket \
  --ticket SMN-20260402-MIRROR-MERGE \
  --title "Add a mirror shard merge verb for month-sharded backfills" \
  --topics mirror,sqlite,backfill,cli

docmgr doc add --ticket SMN-20260402-MIRROR-MERGE \
  --doc-type design-doc \
  --title "Analysis, design, and implementation plan for a mirror shard merge verb"

docmgr doc add --ticket SMN-20260402-MIRROR-MERGE \
  --doc-type reference \
  --title "Investigation diary"
```

- Most important evidence locations:
  - `cmd/smailnail/commands/mirror.go:24-42`, `:55-145`, `:188-330`
  - `pkg/mirror/service.go:21-40`, `:70-198`, `:312-556`, `:642-1082`
  - `pkg/mirror/schema.go:24-79`, `:82-156`
  - `pkg/mirror/files.go:23-92`
  - `ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/scripts/run-last-24-months-backfill.sh:45-147`

## Step 2: Write the design guide, relate files, validate, and package the bundle

With the evidence in hand, I wrote the detailed design guide as an implementation handoff for a new intern. The guide explains the current mirror stack, defines the merge problem precisely, recommends a verb/API shape, and walks through the merge algorithm phase by phase, including validation strategy, conflict policy, and test planning.

I also prepared the ticket bookkeeping for delivery: the tasks checklist, changelog, file relations, and upload workflow. The goal of this step was not just to produce a document, but to make the ticket continuation-friendly and reviewable.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the ticket package so the design is stored, validated, and delivered.

**Inferred user intent:** Have a durable, shareable design artifact that can drive implementation and can be read on reMarkable.

**Commit (code):** N/A

### What I did
- Wrote the primary design guide with sections for scope, architecture, gaps, proposed APIs, merge flow, pseudocode, phased implementation, tests, alternatives, risks, and references.
- Updated `index.md`, `tasks.md`, and `changelog.md`.
- Related the key mirror files and sharding scripts to the design doc and diary with `docmgr doc relate`.
- Ran `docmgr doctor --ticket SMN-20260402-MIRROR-MERGE --stale-after 30`.
- Prepared a bundled reMarkable upload under `/ai/2026/04/02/SMN-20260402-MIRROR-MERGE`.

### Why
- The design doc is the main deliverable, but the ticket metadata and relations are what make it searchable and maintainable.
- The upload bundle matters because the user explicitly asked to store the work in the ticket and upload it to reMarkable.

### What worked
- The ticket structure created by `docmgr` fit the workflow well.
- The current mirror code was evidence-rich enough to support a concrete implementation plan without guessing.
- The design naturally decomposed into clear phases: command surface, preflight discovery, merge execution, checkpoint rebuild, FTS rebuild, and test hardening.

### What didn't work
- N/A. There were no code-level bugs in this documentation step.

### What I learned
- The strongest reason to implement the merge in Go is not CLI aesthetics. It is the need to coordinate SQLite rows, file copies, integrity checks, and derived-state rebuilds in one testable unit.
- The strongest reason not to promote the parallel runner into the CLI yet is that it remains operator policy rather than core data-model logic.

### What was tricky to build
- The trickiest writing problem was deciding how much of the earlier month-sharding workflow to re-explain. The right balance was to explain enough for the intern to understand the producer side of shard data, but keep the main focus on the merge semantics and the boundary between canonical persisted artifacts and rebuildable derived state.

### What warrants a second pair of eyes
- The proposed conflict resolution rule for duplicate rows and duplicate raw files.
- The exact v1 contract for missing raw files: warning, hard failure, or operator-controlled flag.
- Whether `--enrich-after` should be in scope for the first implementation or documented as a follow-up.

### What should be done in the future
- After implementation, add:
  - help pages for the new verb,
  - operator examples showing merge after a 24-month shard run,
  - integration tests that verify the merged DB is incrementally syncable.

### Code review instructions
- Read the design doc top to bottom once before reviewing implementation details.
- Verify that every major recommendation has evidence in the referenced code.
- After implementation exists, validate with:
  - a temp-dir unit test harness for merge service logic,
  - a shard merge of synthetic fixture DBs,
  - a real shard merge built from Docker Compose IMAP fixture month runs.

### Technical details
- Validation command:
```bash
docmgr doctor --ticket SMN-20260402-MIRROR-MERGE --stale-after 30
```

- Planned reMarkable bundle shape:
```bash
remarquee upload bundle \
  ttmp/2026/04/02/SMN-20260402-MIRROR-MERGE--add-a-mirror-shard-merge-verb-for-month-sharded-backfills/design-doc/01-analysis-design-and-implementation-plan-for-a-mirror-shard-merge-verb.md \
  ttmp/2026/04/02/SMN-20260402-MIRROR-MERGE--add-a-mirror-shard-merge-verb-for-month-sharded-backfills/reference/01-investigation-diary.md \
  --name "SMN-20260402-MIRROR-MERGE" \
  --remote-dir "/ai/2026/04/02/SMN-20260402-MIRROR-MERGE" \
  --toc-depth 2
```

## Step 3: Convert the design into an implementation checklist and lock the remaining v1 decisions

Once the research bundle was done, the next step was to convert the ticket into an actual implementation workspace. I replaced the placeholder task list with a staged implementation checklist and updated the design recommendations to match the product choices you clarified after reading the first version of the ticket.

The three clarified decisions were important because they directly affect command shape and error handling. The merge command should support `--enrich-after` in v1, missing raw source files should warn by default instead of hard-failing, and shard discovery should stay root-based in v1 rather than accepting arbitrary explicit shard paths.

### Prompt Context

**User prompt (verbatim):** "--enrich-after is great. can you add that flag to the normal download too? missing raw should warn. 

I don't understand the last point"

**Assistant interpretation:** Confirm the current mirror command already has `--enrich-after`, explain the shard discovery choice more concretely, and update the merge design to lock in the requested defaults.

**Inferred user intent:** Reduce ambiguity in the design before implementation starts, especially around operator ergonomics and failure behavior.

**Commit (code):** pending

### What I did
- Confirmed that the existing `smailnail mirror` command already supports `--enrich-after`.
- Reframed the shard-discovery decision as a simple v1 `--input-root` scan, without explicit `--shard` flags.
- Expanded the ticket `tasks.md` with an implementation checklist.
- Updated the design guide so v1 now explicitly includes `--enrich-after`, warning-by-default missing raw handling, and root-based shard discovery only.
- Updated `changelog.md` to record those product decisions.

### Why
- These decisions are not cosmetic. They control:
  - whether the command surface needs enrichment reporting from day one,
  - whether the merge path is strict or forgiving when a shard is slightly incomplete,
  - whether the first implementation stays simple by targeting the directory layout that the current scripts already produce.

### What worked
- The existing `mirror` command already had the right enrichment pattern to copy.
- Root-based discovery turned out to be the clearest operator story because it matches the existing shard producer scripts exactly.
- Warning-by-default for missing raw files fits the user’s operational preference without complicating the first implementation too much.

### What didn't work
- N/A. This step was a design and ticket-management step, not a code-debugging step.

### What I learned
- The merge verb’s product contract is now narrow enough to implement directly:
  - one input root,
  - one destination DB,
  - one destination mirror root,
  - optional enrichment after merge,
  - warnings for missing raw files rather than immediate failure.

### What was tricky to build
- The subtle part was making sure “warn by default for missing raw files” did not silently undermine the rest of the design. The right framing is that missing raw files are tolerated and counted in the report, but genuine raw-file conflicts, such as the same relative path with a different SHA, should still remain hard errors.

### What warrants a second pair of eyes
- Whether warning-by-default for missing raw files should still be paired with an explicit strict flag from the start.
- Whether the first merge report should include detailed per-shard warning breakdowns or just aggregate counts.

### What should be done in the future
- Implement the command and service to match these now-locked decisions.

### Code review instructions
- Check that the design doc now consistently reflects the clarified decisions in:
  - command shape,
  - enrichment section,
  - open questions / recommendations.
- Check that `tasks.md` is now implementation-oriented rather than research-oriented.

### Technical details
- The current normal download command already exposes `enrich-after` in `cmd/smailnail/commands/mirror.go`.
