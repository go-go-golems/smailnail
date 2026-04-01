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
    - Path: smailnail/pkg/mailruntime/imap_client.go
      Note: |-
        Existing UID-based IMAP runtime identified as the recommended sync foundation
        Runtime identified as the best sync foundation during research
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
LastUpdated: 2026-04-01T17:55:00-04:00
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
