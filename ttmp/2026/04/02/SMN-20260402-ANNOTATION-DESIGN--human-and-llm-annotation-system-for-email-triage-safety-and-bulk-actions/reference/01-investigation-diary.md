---
Title: Investigation diary
Ticket: SMN-20260402-ANNOTATION-DESIGN
Status: active
Topics:
    - email
    - sqlite
    - cli
    - backend
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/00-run-all.sh
      Note: Ordered evidence runner for the ticket SQL bundle
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/01-schema-inventory.sql
      Note: Schema and metadata inspection query
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/02-message-and-thread-shape.sql
      Note: Message and thread-shape evidence query
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/03-sender-shape.sql
      Note: Sender distribution evidence query
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/04-risk-and-unsubscribe-shape.sql
      Note: Risk and unsubscribe candidate query
    - Path: ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/05-annotation-targets.sql
      Note: Entity-count and annotation-target discovery query
ExternalSources: []
Summary: Chronological investigation log for the annotation-system design ticket, including repository inspection, mailbox evidence gathering, and design decisions.
LastUpdated: 2026-04-02T14:32:00.106190947-04:00
WhatFor: Use this diary to retrace the design work, rerun the evidence scripts, and understand why the final recommendation looks the way it does.
WhenToUse: Read this when validating the design, onboarding a contributor, or revisiting assumptions about mailbox shape and risk.
---


# Investigation diary

## Goal

Design a new annotation and review subsystem for `smailnail` that lets humans and LLMs annotate messages, senders, threads, and related entities without allowing unsafe bulk actions to bypass human review.

## Prompt Context

The user asked for a new ticket that would:

- design a way for humans and LLMs to annotate emails, email addresses, and other entities
- support batch analysis of incoming emails
- keep humans as the final authority
- handle "important", "suspected spam", and similar judgments with nuance rather than a binary flag
- include a detailed analysis, design, and implementation guide suitable for a new intern
- store evidence-gathering scripts and SQL queries inside the ticket
- upload the result to reMarkable

## What I Inspected

I first inspected the existing code paths that would constrain or enable the design.

Repository surfaces reviewed:

- `pkg/mirror/schema.go`
- `pkg/enrich/schema.go`
- `pkg/smailnaild/db.go`
- `pkg/smailnaild/http.go`
- `pkg/smailnaild/rules/types.go`
- `pkg/smailnaild/rules/service.go`
- `ui/src/api/types.ts`
- `ui/src/features/mailbox/MailboxExplorer.tsx`
- `ui/src/features/mailbox/mailboxSlice.ts`

Key findings:

- the mirror DB already has stable message rows and enrichment-derived sender/thread structures
- the server and UI currently stop at mailbox browsing and rule dry-runs
- there is no annotation schema, no review queue, and no proposal model
- the rules subsystem already demonstrates a useful dry-run pattern

## Scripts Added

The ticket now contains these ordered scripts:

- `scripts/00-run-all.sh`
- `scripts/01-schema-inventory.sql`
- `scripts/02-message-and-thread-shape.sql`
- `scripts/03-sender-shape.sql`
- `scripts/04-risk-and-unsubscribe-shape.sql`
- `scripts/05-annotation-targets.sql`

The intent was to make the design grounded in real mailbox data rather than guesses.

## Commands Run

These are the main commands used during the investigation.

```bash
find ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions -maxdepth 3 -type f | sort
git status --short
nl -ba pkg/mirror/schema.go | sed -n '1,220p'
nl -ba pkg/enrich/schema.go | sed -n '1,260p'
nl -ba pkg/smailnaild/db.go | sed -n '1,260p'
nl -ba pkg/smailnaild/http.go | sed -n '1,320p'
nl -ba pkg/smailnaild/rules/types.go | sed -n '1,260p'
nl -ba pkg/smailnaild/rules/service.go | sed -n '1,320p'
nl -ba ui/src/api/types.ts | sed -n '1,260p'
nl -ba ui/src/features/mailbox/MailboxExplorer.tsx | sed -n '1,260p'
nl -ba ui/src/features/mailbox/mailboxSlice.ts | sed -n '1,320p'
bash ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/00-run-all.sh /tmp/smailnail-parallel-a.sqlite
sqlite3 -header -column /tmp/smailnail-parallel-a.sqlite "SELECT domain, COUNT(*) AS sender_count, SUM(msg_count) AS total_messages FROM senders WHERE domain <> '' GROUP BY domain ORDER BY total_messages DESC LIMIT 20;"
sqlite3 -header -column /tmp/smailnail-parallel-a.sqlite "SELECT account_key, mailbox_name, COUNT(*) AS messages FROM messages GROUP BY account_key, mailbox_name ORDER BY messages DESC LIMIT 20;"
```

## Evidence Snapshot

The most important evidence came from running `scripts/00-run-all.sh` against `/tmp/smailnail-parallel-a.sqlite`.

Mirror status:

- `schema_version = 2`
- `fts5_status = available`
- enrichment timestamps were present for senders, threads, and unsubscribe extraction

Corpus counts:

- `messages = 2929`
- `senders = 533`
- `threads = 2880`

Coverage shape:

- `remote_deleted_messages = 0`
- `attachment_messages = 47`
- `threaded_messages = 2915`
- `messages_with_reply_headers = 390`
- `messages_without_sender_email = 135`
- `messages_without_thread_id = 14`

Thread shape:

- `2869` singleton threads
- `11` multi-message threads
- largest observed thread size was `8`

High-volume domains:

- `github.com = 702`
- `substack.com = 280`
- `privaterelay.appleid.com = 114`
- `twitch.tv = 86`

Mailbox shape:

- sampled DB currently contains one mailbox: `mail-bl0rg-net-993-manuel-a8c4454ab8d9 / INBOX`

## What This Evidence Changed

This evidence directly shaped the design.

### Decision 1: Support More Than Messages

Reason:

- sender and domain decisions will eliminate much more noise than message-only annotations

Evidence:

- hundreds of messages are concentrated in a small number of senders and domains

### Decision 2: Separate Observations From Decisions

Reason:

- the user explicitly wants varying degrees of suspicion and human final authority

Consequence:

- raw LLM output cannot become final state
- explicit human decisions must exist as a distinct table or layer

### Decision 3: Use Proposal Batches for Risky Bulk Work

Reason:

- the corpus obviously contains many messages that look newsletter-like or alert-like
- the same corpus also contains important operational and conversational mail

Consequence:

- the design treats archive, move, unsubscribe, and delete as proposal items to review, not direct actions

### Decision 4: Store Corpus-Local Annotations in the Mirror DB

Reason:

- the mirror DB already contains the entities being annotated
- local SQL analysis and CLI workflows are easier if everything stays in one SQLite file

## What Worked

- The enriched mirror DB provided enough structure to reason about sender-, thread-, and domain-level annotation needs.
- The rules subsystem provided a clean analogy for dry-run and preview behavior.
- The ordered ticket scripts kept the evidence reproducible and easy to rerun.

## What Did Not Exist Yet

- There is no current annotation subsystem in the repo.
- There are no annotation API contracts in `ui/src/api/types.ts`.
- There is no annotation state in `ui/src/features/mailbox/mailboxSlice.ts`.
- There is no server route family for annotation, review, or proposal handling in `pkg/smailnaild/http.go`.

## Design Recommendation in One Sentence

Build a mirror-DB-first annotation subsystem with generic targets, append-only observations, explicit human decisions, analysis-run provenance, and reviewable proposal batches that are the only path to risky bulk actions.

## Quick Reference

Recommended first target types:

- `message`
- `thread`
- `sender`
- `domain`
- `mailbox`
- `account`

Recommended first annotation keys:

- `triage.importance`
- `triage.spam_likelihood`
- `triage.newsletter`
- `triage.needs_followup`
- `safety.do_not_delete`
- `sender.trust_level`
- `sender.relationship`
- `notes.general`

Recommended first tables:

- `annotation_schema`
- `annotation_observations`
- `annotation_decisions`
- `analysis_runs`
- `proposal_batches`
- `proposal_items`

## Suggested Rerun Procedure

If someone wants to validate the evidence again:

```bash
bash ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/00-run-all.sh /tmp/smailnail-parallel-a.sqlite
```

If they want to inspect only likely annotation targets:

```bash
sqlite3 /tmp/smailnail-parallel-a.sqlite < ttmp/2026/04/02/SMN-20260402-ANNOTATION-DESIGN--human-and-llm-annotation-system-for-email-triage-safety-and-bulk-actions/scripts/05-annotation-targets.sql
```

## Related

- `design-doc/01-analysis-design-and-implementation-guide-for-annotations-review-and-safe-bulk-actions.md`
- `design-doc/02-mvp-fast-path-annotations-groups-and-logs.md`
- `scripts/00-run-all.sh`
- `scripts/05-annotation-targets.sql`

## Implementation Log

### 2026-04-02 20:xx Eastern: narrowed scope to SQLite plus CLI MVP

I updated the ticket task list to match the new request:

- no HTTP API work
- no UI work
- only mirror SQLite schema and CLI verbs

This matters because the previous task list still reflected the broader architecture ticket.

### 2026-04-02 20:xx Eastern: added schema v3 and repository layer

I implemented the first code slice:

- added `pkg/annotate/schema.go`
- added `pkg/annotate/types.go`
- added `pkg/annotate/repository.go`
- added `pkg/annotate/repository_test.go`
- updated `pkg/mirror/schema.go` to bootstrap schema version `3`
- updated `pkg/mirror/store_test.go` to assert the new tables exist

What this slice includes:

- `annotations`
- `target_groups`
- `target_group_members`
- `annotation_logs`
- `annotation_log_targets`

Repository methods added:

- create/get/list annotations
- update annotation review state
- create/get/list groups
- add/list group members
- create/get/list logs
- link/list log targets

Validation command for this slice:

```bash
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/mirror
```

Issue encountered:

- the first test run used `package annotate` in `repository_test.go` while importing `pkg/mirror`, which created an import cycle because `mirror` now imports `annotate`

Fix:

- switched the test file to `package annotate_test`
- reran the package tests with the required `sqlite_fts5` build tag

Commit:

- `b18fbb0` `Add annotation schema and repository MVP`

### 2026-04-02 20:xx Eastern: added SQLite-only CLI verbs

I implemented the CLI layer under `cmd/smailnail/commands/annotate` and wired it into `cmd/smailnail/main.go`.

Command tree added:

- `smailnail annotate annotation add`
- `smailnail annotate annotation list`
- `smailnail annotate annotation review`
- `smailnail annotate group create`
- `smailnail annotate group list`
- `smailnail annotate group add-target`
- `smailnail annotate group members`
- `smailnail annotate log add`
- `smailnail annotate log list`
- `smailnail annotate log link-target`
- `smailnail annotate log targets`

Files added in this slice:

- `cmd/smailnail/commands/annotate/common.go`
- `cmd/smailnail/commands/annotate/root.go`
- `cmd/smailnail/commands/annotate/annotation_root.go`
- `cmd/smailnail/commands/annotate/annotation_add.go`
- `cmd/smailnail/commands/annotate/annotation_list.go`
- `cmd/smailnail/commands/annotate/annotation_review.go`
- `cmd/smailnail/commands/annotate/group_root.go`
- `cmd/smailnail/commands/annotate/group_create.go`
- `cmd/smailnail/commands/annotate/group_list.go`
- `cmd/smailnail/commands/annotate/group_add_target.go`
- `cmd/smailnail/commands/annotate/group_members.go`
- `cmd/smailnail/commands/annotate/log_root.go`
- `cmd/smailnail/commands/annotate/log_add.go`
- `cmd/smailnail/commands/annotate/log_list.go`
- `cmd/smailnail/commands/annotate/log_link_target.go`
- `cmd/smailnail/commands/annotate/log_targets.go`

Focused verification command for this slice:

```bash
go test -tags sqlite_fts5 ./cmd/smailnail ./cmd/smailnail/commands/annotate ./pkg/annotate ./pkg/mirror
```

Smoke-test DB:

```bash
cp /tmp/smailnail-parallel-a.sqlite /tmp/smailnail-parallel-a.annotate-smoke.sqlite
```

Smoke-test commands that succeeded:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation add \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --target-type sender \
  --target-id notifications@github.com \
  --tag important \
  --note 'Still important despite high volume' \
  --source-kind human \
  --created-by manuel

go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation review \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --id 4cc62d7e-e52c-4f49-95b8-bdea24d6af19 \
  --review-state dismissed

go run -tags sqlite_fts5 ./cmd/smailnail annotate group create \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --name 'Possible newsletters' \
  --description 'Smoke test group' \
  --source-kind agent \
  --agent-run-id smoke-run-1

go run -tags sqlite_fts5 ./cmd/smailnail annotate group add-target \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --group-id cf87e5b4-2dca-4a27-a381-f946529883c2 \
  --target-type sender \
  --target-id hello@readwise.io

go run -tags sqlite_fts5 ./cmd/smailnail annotate log add \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --title 'Smoke test pass' \
  --body 'Testing annotation CLI against copied mirror DB.' \
  --source-kind agent \
  --agent-run-id smoke-run-1

go run -tags sqlite_fts5 ./cmd/smailnail annotate log link-target \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --log-id 01228f9d-e717-4665-809c-23e25f011742 \
  --target-type sender \
  --target-id hello@readwise.io
```

Lookup commands that succeeded:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation list \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --target-type sender \
  --target-id notifications@github.com

go run -tags sqlite_fts5 ./cmd/smailnail annotate group members \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --group-id cf87e5b4-2dca-4a27-a381-f946529883c2

go run -tags sqlite_fts5 ./cmd/smailnail annotate log targets \
  --sqlite-path /tmp/smailnail-parallel-a.annotate-smoke.sqlite \
  --log-id 01228f9d-e717-4665-809c-23e25f011742
```

Issue encountered:

- the first compile pass failed because `common.go` still contained a dead helper using a `types.Row` API shape that did not exist in that context

Fix:

- removed the unused helper
- reran `gofmt`
- reran the focused package tests

Commit:

- `6d849cf` `Add annotate CLI verbs for annotations groups and logs`

### 2026-04-02 20:xx Eastern: finalized the ticket bundle

After the code commits, I staged and committed the remaining ticket files so the work is reviewable in git rather than only present in the worktree.

Files tracked in this final ticket-only pass:

- `README.md`
- `index.md`
- both design docs
- all `scripts/` files
- updated `tasks.md`
- updated `changelog.md`
- updated `reference/01-investigation-diary.md`

Validation command for the final ticket state:

```bash
docmgr doctor --ticket SMN-20260402-ANNOTATION-DESIGN --stale-after 30
```

Result:

- all checks passed

Commit:

- `9484adb` `Track annotation ticket docs and investigation bundle`
