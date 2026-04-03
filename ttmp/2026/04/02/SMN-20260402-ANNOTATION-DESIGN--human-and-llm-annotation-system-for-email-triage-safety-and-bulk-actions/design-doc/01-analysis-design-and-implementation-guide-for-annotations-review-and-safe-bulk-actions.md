---
Title: Analysis, design, and implementation guide for annotations, review, and safe bulk actions
Ticket: SMN-20260402-ANNOTATION-DESIGN
Status: active
Topics:
    - email
    - sqlite
    - cli
    - backend
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/enrich/schema.go
      Note: Current enrichment-owned sender and thread schema additions
    - Path: pkg/mirror/schema.go
      Note: Mirror schema bootstrap and future schema v3 insertion point
    - Path: pkg/smailnaild/db.go
      Note: Current application DB schema used to contrast mirror-vs-app storage
    - Path: pkg/smailnaild/http.go
      Note: Current HTTP route surface that will need annotation and review endpoints
    - Path: pkg/smailnaild/rules/service.go
      Note: Existing service pattern for preview-first workflows
    - Path: pkg/smailnaild/rules/types.go
      Note: Dry-run record types used as a design analogue for proposal batches
    - Path: ui/src/api/types.ts
      Note: Current frontend API contracts that need annotation additions
    - Path: ui/src/features/mailbox/MailboxExplorer.tsx
      Note: Current browse-only mailbox UI that needs annotation panels
    - Path: ui/src/features/mailbox/mailboxSlice.ts
      Note: Current mailbox state model with no annotation or review state
ExternalSources: []
Summary: Recommended architecture and rollout plan for a safety-first annotation system that supports human notes, LLM suggestions, and reviewable bulk action proposals.
LastUpdated: 2026-04-02T14:32:00.053746761-04:00
WhatFor: Use this document to implement annotation storage, conflict resolution, review workflows, and safe bulk-triage tooling on top of smailnail mirrors.
WhenToUse: Read this before adding annotation tables, review APIs, proposal runners, or UI surfaces that combine human and model judgments.
---


# Analysis, design, and implementation guide for annotations, review, and safe bulk actions

## Executive Summary

`smailnail` already has two important foundations:

- a mirror database that stores normalized email content and enrichment outputs such as senders and threads
- an application server and UI that already model user-owned records, draft state, and dry-run workflows for rules

The missing piece is a durable annotation layer. Right now there is no first-class place to record:

- "this sender is important"
- "this sender looks like spam but do not delete automatically"
- "this domain is usually promotional"
- "this message needs follow-up"
- "this thread should never be bulk-archived"
- "the model suggested deleting these 800 messages, but the human approved only 120"

The recommended design adds a new annotation subsystem in the mirror database and exposes it through CLI, HTTP, and UI surfaces. The system is built around three hard requirements:

1. Human judgment overrides model judgment.
2. Destructive bulk actions must flow through reviewable proposal batches, not direct LLM execution.
3. Annotations must work across several entity types, not only individual messages.

The recommended implementation introduces:

- a generic target model for messages, threads, senders, domains, mailboxes, accounts, and future entities
- append-only observations from humans, LLMs, and heuristics
- explicit decisions that determine the effective state seen by downstream automation
- analysis runs that preserve prompt, model, scope, and provenance
- proposal batches and proposal items for archive, move, flag, mute, unsubscribe, or delete recommendations
- review queues, locks, and safety labels that block destructive operations unless a human explicitly approves them

For an intern, the easiest mental model is:

- the mirror database stores the email corpus
- enrichment adds derived structure like `senders` and `threads`
- annotations add judgments and notes
- decisions decide what counts
- proposals package risky actions for review
- executors only act on approved proposals

## Problem Statement

The mailbox mirror now contains enough structure to support higher-level triage, but not enough policy to act safely. The current mirror schema in `pkg/mirror/schema.go:24-79` stores `messages` and mirror metadata. The enrichment migration in `pkg/enrich/schema.go:5-39` adds `thread_id`, `sender_email`, `sender_domain`, plus `threads` and `senders`. That gives us useful nouns, but no review system.

The evidence from `/tmp/smailnail-parallel-a.sqlite` shows why this matters:

- there are `2929` messages in the sampled mirror
- there are `533` normalized senders
- there are `2880` threads, of which `2869` are singletons
- only `390` messages even have reply headers
- `237` senders expose list-unsubscribe metadata
- a few domains dominate volume, including `github.com` with `702` messages and `substack.com` with `280`

That mailbox shape tells us several things.

First, message-by-message review alone will be too expensive. Many useful decisions are really about senders, domains, or newsletters. Second, threads are helpful but not sufficient as the only unit, because most mail in this sample is singleton mail. Third, bulk operations will be tempting because the corpus contains large volumes of recurring notifications, newsletters, alerts, and likely low-value mail. Fourth, any design that lets a model directly delete mail is unsafe, because a false positive on an important sender or a weird-but-legitimate invoice would be costly.

The system therefore needs to preserve nuance instead of collapsing everything into a single spam bit.

## Existing System Walkthrough

An intern should understand the current architecture before adding anything.

### Mirror Database

The mirror schema lives in `pkg/mirror/schema.go`.

- `messages` is the core corpus table.
- `mirror_metadata` tracks schema version and FTS availability.
- schema version `2` delegates to `pkg/enrich` for sender and thread additions.

Important message fields today:

- `id` is the local stable row id and should become the preferred message annotation key
- `(account_key, mailbox_name, uidvalidity, uid)` identifies the original IMAP placement
- `message_id` is useful but not unique enough to be the only key
- `headers_json`, `body_text`, `body_html`, and `search_text` carry analysis context
- `thread_id`, `sender_email`, and `sender_domain` are enrichment outputs

### Enrichment Layer

The enrichment migration in `pkg/enrich/schema.go` adds:

- `messages.thread_id`
- `messages.thread_depth`
- `messages.sender_email`
- `messages.sender_domain`
- `threads`
- `senders`

This is the right level for annotation targets because it already normalizes:

- messages
- senders
- domains
- threads

### Application Database

The server-side application DB lives in `pkg/smailnaild/db.go`. Its schema is separate from the mirror DB and currently contains user, account, session, and rule tables. This is relevant because we need to decide where annotations live.

The current app DB is a better fit for:

- user identities
- auth/session state
- server config
- shared metadata unrelated to mirrored message corpora

The current mirror DB is a better fit for:

- annotations that join directly to messages, senders, and threads
- offline or CLI-driven review work
- model-analysis runs over a concrete mirrored corpus

### HTTP and UI Surfaces

The current HTTP server in `pkg/smailnaild/http.go:25-197` exposes:

- account CRUD
- mailbox listing
- message listing/detail
- rules CRUD
- rules dry-run

The current mailbox UI in `ui/src/features/mailbox/MailboxExplorer.tsx:21-162` and `ui/src/features/mailbox/mailboxSlice.ts:5-158` is browse-only. It can load mailboxes, list messages, and show a message detail view, but it has no annotation state, no review queue, and no notion of sender or thread judgments.

The current API types in `ui/src/api/types.ts:1-193` also stop at accounts, messages, and rules. There is no annotation contract yet.

### Rules as a Design Analogy

The rules service is worth studying because it already models draft and dry-run behavior.

Relevant files:

- `pkg/smailnaild/rules/types.go:5-30`
- `pkg/smailnaild/rules/service.go:18-252`

Important ideas to borrow:

- user-owned records
- normalized service inputs
- explicit dry-run mode
- persisted run history
- sample rows returned to the UI for preview

The annotation subsystem should mimic that safety posture, but on a broader target model.

## Design Goals

- Let humans and LLMs attach notes and structured labels to more than one entity type.
- Preserve provenance for every claim: who or what said it, why, when, and with what confidence.
- Make human decisions authoritative and explicit.
- Support batch review and bulk-action proposal workflows.
- Keep raw suggestions separate from effective state.
- Make destructive actions opt-in, reviewable, and auditable.
- Support both CLI-first and web-UI-first workflows.
- Keep the data model extensible enough to add URLs, attachments, or contacts later.

## Non-Goals

- Do not let the first version execute destructive IMAP actions automatically from raw model output.
- Do not attempt perfect global identity resolution for every sender alias in v1.
- Do not build a giant ontology before basic note-taking and review work.
- Do not add a backward-compatibility abstraction layer unless implementation work proves it is necessary.

## Recommended Architecture

The recommended architecture is a layered system.

```text
IMAP -> mirror sync -> messages
                    -> enrich -> senders / threads / sender_domain
                    -> analyze -> observations
                    -> review  -> decisions
                    -> propose -> proposal_batches / proposal_items
                    -> apply   -> only approved actions
```

The core separation is:

- observations are not decisions
- decisions are not actions
- actions are not executed unless approved

That separation is what makes "human final word" enforceable in code instead of just aspiration.

## Recommended Storage Decision

Store annotation and proposal data in the mirror database, not the existing application database.

### Why the Mirror DB Is the Better Primary Home

- The annotation targets already live there: `messages`, `senders`, and `threads`.
- Batch analysis is naturally a SQL-heavy operation over mirrored content.
- CLI workflows can operate directly on one mirror file without needing the web app.
- Ticket scripts and ad hoc investigations become simpler because all joins stay local.
- Mirror DB portability matters. You can copy a single `.sqlite` file and preserve both the corpus and the judgments made against it.

### Why Not the App DB

- The current app DB does not contain mirrored message rows.
- Joining app-level annotations back to mirror entities would require synthetic foreign keys and fragile compound identifiers.
- A design that spans two databases from day one increases complexity before we even know the review workflow.

### Recommended Split

- Mirror DB owns corpus-local annotations, decisions, analysis runs, and proposal batches.
- App DB can later own user accounts, auth, and possibly user preferences for views or reviewer identity mapping.

## Target Model

The system should start with these target types:

- `message`
- `thread`
- `sender`
- `domain`
- `mailbox`
- `account`

Future-safe but optional target types:

- `url`
- `unsubscribe_endpoint`
- `attachment`
- `contact_cluster`

Use one generic target encoding:

- `target_type TEXT`
- `target_id TEXT`

Recommended `target_id` formats:

- message: the `messages.id` integer encoded as text
- thread: `threads.thread_id`
- sender: normalized email address
- domain: normalized domain string
- mailbox: `account_key + "::" + mailbox_name`
- account: `account_key`

This is simpler than creating six separate annotation tables and keeps the system extensible.

## Vocabulary Model

Do not make the first version a pure free-form note system. That will devolve into string soup and make review automation weak. Also do not make it so rigid that every new label requires a migration.

Use a schema catalog table for annotation keys.

Recommended table:

- `annotation_schema`

Suggested columns:

- `annotation_key TEXT PRIMARY KEY`
- `label TEXT NOT NULL`
- `description TEXT NOT NULL`
- `target_types_json TEXT NOT NULL`
- `value_type TEXT NOT NULL`
- `allowed_values_json TEXT NOT NULL DEFAULT '[]'`
- `danger_level TEXT NOT NULL DEFAULT 'low'`
- `human_review_required BOOLEAN NOT NULL DEFAULT FALSE`
- `supports_notes BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
- `updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`

Examples of annotation keys for v1:

- `triage.importance`
- `triage.spam_likelihood`
- `triage.newsletter`
- `triage.needs_followup`
- `safety.do_not_delete`
- `sender.relationship`
- `sender.trust_level`
- `sender.allow_unsubscribe`
- `notes.general`

Examples of allowed values:

- `triage.importance`: `critical`, `high`, `normal`, `low`, `unknown`
- `triage.spam_likelihood`: `none`, `low`, `medium`, `high`, `unknown`
- `sender.relationship`: `self`, `personal`, `work`, `community`, `vendor`, `unknown`
- `sender.trust_level`: `trusted`, `watch`, `suspicious`, `blocked`, `unknown`

## Core Data Model

The recommended minimal schema has five groups of tables.

### 1. Observations

Observations are raw claims or notes from a human, an LLM, or a heuristic.

Recommended table:

- `annotation_observations`

Suggested columns:

- `id TEXT PRIMARY KEY`
- `target_type TEXT NOT NULL`
- `target_id TEXT NOT NULL`
- `annotation_key TEXT NOT NULL`
- `value_json TEXT NOT NULL`
- `note_markdown TEXT NOT NULL DEFAULT ''`
- `source_kind TEXT NOT NULL`
- `source_label TEXT NOT NULL DEFAULT ''`
- `author_id TEXT NOT NULL DEFAULT ''`
- `analysis_run_id TEXT NOT NULL DEFAULT ''`
- `confidence REAL`
- `status TEXT NOT NULL DEFAULT 'active'`
- `evidence_json TEXT NOT NULL DEFAULT '{}'`
- `supersedes_id TEXT NOT NULL DEFAULT ''`
- `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`

Important `source_kind` values:

- `human`
- `llm`
- `heuristic`
- `import`
- `system`

Important `status` values:

- `active`
- `withdrawn`
- `superseded`

### 2. Decisions

Decisions are the governing state. These are what automation reads.

Recommended table:

- `annotation_decisions`

Suggested columns:

- `id TEXT PRIMARY KEY`
- `target_type TEXT NOT NULL`
- `target_id TEXT NOT NULL`
- `annotation_key TEXT NOT NULL`
- `decision_kind TEXT NOT NULL`
- `resolved_value_json TEXT NOT NULL`
- `source_observation_id TEXT NOT NULL DEFAULT ''`
- `decided_by TEXT NOT NULL`
- `rationale_markdown TEXT NOT NULL DEFAULT ''`
- `locked BOOLEAN NOT NULL DEFAULT FALSE`
- `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`

Important `decision_kind` values:

- `accepted`
- `rejected`
- `overridden`
- `set`
- `cleared`

The most important invariant is:

- destructive automation must consult decisions, not observations

### 3. Analysis Runs

Analysis runs preserve provenance for model-generated suggestions.

Recommended table:

- `analysis_runs`

Suggested columns:

- `id TEXT PRIMARY KEY`
- `runner_kind TEXT NOT NULL`
- `model_name TEXT NOT NULL DEFAULT ''`
- `prompt_version TEXT NOT NULL DEFAULT ''`
- `scope_json TEXT NOT NULL`
- `parameters_json TEXT NOT NULL DEFAULT '{}'`
- `status TEXT NOT NULL`
- `started_at TIMESTAMP NOT NULL`
- `completed_at TIMESTAMP`
- `summary_json TEXT NOT NULL DEFAULT '{}'`

This table is critical for reproducibility. Without it, you will not know which prompt or model produced a risky proposal.

### 4. Proposal Batches

Proposal batches group actions generated by a human or model for review.

Recommended tables:

- `proposal_batches`
- `proposal_items`

Suggested `proposal_batches` columns:

- `id TEXT PRIMARY KEY`
- `title TEXT NOT NULL`
- `proposal_kind TEXT NOT NULL`
- `created_by TEXT NOT NULL`
- `analysis_run_id TEXT NOT NULL DEFAULT ''`
- `scope_json TEXT NOT NULL DEFAULT '{}'`
- `status TEXT NOT NULL`
- `created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
- `reviewed_at TIMESTAMP`
- `applied_at TIMESTAMP`

Suggested `proposal_items` columns:

- `id TEXT PRIMARY KEY`
- `batch_id TEXT NOT NULL`
- `target_type TEXT NOT NULL`
- `target_id TEXT NOT NULL`
- `action_type TEXT NOT NULL`
- `action_params_json TEXT NOT NULL DEFAULT '{}'`
- `reason_markdown TEXT NOT NULL DEFAULT ''`
- `confidence REAL`
- `decision_state TEXT NOT NULL DEFAULT 'pending'`
- `reviewed_by TEXT NOT NULL DEFAULT ''`
- `review_note_markdown TEXT NOT NULL DEFAULT ''`
- `applied_at TIMESTAMP`

Important `decision_state` values:

- `pending`
- `approved`
- `rejected`
- `skipped`
- `applied`
- `failed`

### 5. Optional Materialized Effective State

You have two implementation options for the effective state:

- compute it from the latest decision rows on demand
- materialize it into a cache table or view

Recommended first move:

- create a view named `effective_annotations_vw`

Example conceptual definition:

```sql
SELECT d.target_type,
       d.target_id,
       d.annotation_key,
       d.resolved_value_json,
       d.decided_by,
       d.locked,
       d.created_at
FROM annotation_decisions d
JOIN (
  SELECT target_type, target_id, annotation_key, MAX(created_at) AS max_created_at
  FROM annotation_decisions
  GROUP BY target_type, target_id, annotation_key
) latest
  ON latest.target_type = d.target_type
 AND latest.target_id = d.target_id
 AND latest.annotation_key = d.annotation_key
 AND latest.max_created_at = d.created_at;
```

## Human-Override Model

This is the single most important behavioral rule in the system.

If an LLM says "high spam likelihood" but a human marks the same sender as "trusted", the effective state must resolve to the human decision. The system must not silently merge them or average them.

Recommended precedence:

1. locked human decision
2. most recent human decision
3. explicitly approved machine suggestion adopted by a human
4. machine observation without decision
5. heuristic fallback
6. unknown

In practice, raw machine observations should remain visible in the UI, but they should not control execution.

### Example

```text
sender=notifications@github.com

observation 1: llm says triage.importance=low confidence=0.61
observation 2: human says sender.trust_level=trusted
decision 1: human sets safety.do_not_delete=true locked=true

effective result:
- sender.trust_level = trusted
- safety.do_not_delete = true
- llm importance suggestion remains visible as background evidence only
```

## Safety Model for Bulk Actions

Bulk action safety should be designed before any executor exists.

### Absolute Rules

- No delete, move, or unsubscribe action may execute directly from an LLM observation.
- Every destructive proposal item must be individually reviewable.
- A single human "approve all" is allowed only on a proposal batch that is already materialized and inspectable.
- `safety.do_not_delete=true` blocks delete proposals regardless of model score.
- Human rejection should be sticky enough to prevent the same suggestion from reappearing immediately without changed evidence.

### Recommended Proposal Types

- `archive_messages`
- `move_messages`
- `mark_read`
- `mute_sender`
- `unsubscribe_sender`
- `delete_messages`
- `open_followup_queue`

### Risk Tiers

- `informational`: note-only, no action
- `reversible`: mark read, apply label, add to review queue
- `semi_reversible`: archive or move
- `destructive`: delete, unsubscribe, block sender

Only the first two categories should be in scope for unattended execution in early phases, and only after human-approved decisions exist.

## Suggested Review Workflows

The design should support several human workflows.

### Workflow A: Sender Triage

Use case:

- review high-volume senders and classify them as trusted, watch, suspicious, newsletter, or bulk promo

Why it matters:

- this is the cheapest place to reduce noise across many messages at once

Suggested queue:

- senders ordered by message volume and unresolved proposal count

### Workflow B: Thread Review

Use case:

- inspect a conversation and mark it as important, actionable, or safe to archive

Why it matters:

- useful for real conversations where message-level classification is too granular

### Workflow C: Message Exceptions

Use case:

- override the sender-level rule for one invoice, contract, or important personal note

Why it matters:

- real mail always has exceptions

### Workflow D: Batch Proposal Review

Use case:

- the LLM proposes 300 messages for archive and 40 senders for unsubscribe

Human review steps:

1. inspect batch summary
2. sample affected messages
3. reject whole classes of mistakes
4. approve the subset that is actually safe
5. apply only approved items

## API Design

The API should mirror the storage model and keep review actions explicit.

### HTTP Endpoints

Suggested new endpoints under `/api`:

- `GET /api/annotation-schema`
- `GET /api/review/queue`
- `GET /api/targets/{type}/{id}/annotations`
- `POST /api/targets/{type}/{id}/observations`
- `POST /api/targets/{type}/{id}/decisions`
- `GET /api/analysis-runs`
- `POST /api/analysis-runs`
- `GET /api/proposal-batches`
- `GET /api/proposal-batches/{id}`
- `POST /api/proposal-batches/{id}/approve`
- `POST /api/proposal-batches/{id}/reject`
- `POST /api/proposal-batches/{id}/apply`

Design notes:

- keep approval separate from apply
- keep observation creation separate from decision creation
- make list endpoints filterable by `target_type`, `annotation_key`, `decision_state`, `source_kind`, and date

### CLI Commands

Recommended new command group:

- `smailnail annotate`

Suggested subcommands:

- `smailnail annotate observation add`
- `smailnail annotate decision set`
- `smailnail annotate list --target-type sender --target-id notifications@github.com`
- `smailnail annotate queue list`
- `smailnail annotate analyze run`
- `smailnail annotate proposals list`
- `smailnail annotate proposals review`
- `smailnail annotate proposals apply --batch <id>`

The CLI matters even if the UI exists, because a large amount of early analysis will be done against local SQLite files.

## UI Design

The current mailbox explorer can already display mailbox and message detail, but it has no room for annotation state. The fastest useful UI increment is not a full rewrite. It is an annotation sidebar and review queue added on top of existing list/detail patterns.

Recommended UI modules:

- target badges on message rows
- message detail annotation panel
- sender profile panel
- thread summary panel
- review queue page
- proposal batch review page

Minimal state additions in frontend Redux or RTK Query:

- active review queue filters
- fetched observations and decisions for selected target
- proposal batch summaries and item lists
- optimistic mutation state for approve/reject/set-note flows

Suggested visual hierarchy:

- effective human decision first
- raw machine suggestions second
- provenance and confidence third

That ordering prevents operators from confusing speculation with approved truth.

## Recommended File Layout for Implementation

Back-end packages to add:

- `pkg/annotate/schema.go`
- `pkg/annotate/types.go`
- `pkg/annotate/repository.go`
- `pkg/annotate/service.go`
- `pkg/annotate/effective.go`
- `pkg/annotate/proposals.go`

CLI packages to add:

- `cmd/smailnail/commands/annotate/*.go`

Server-side packages to add or extend:

- `pkg/smailnaild/http.go`
- `pkg/smailnaild/annotations/*`

Frontend files likely to change:

- `ui/src/api/types.ts`
- `ui/src/api/client.ts`
- `ui/src/features/mailbox/*`
- `ui/src/features/annotations/*`

Mirror schema touchpoints:

- `pkg/mirror/schema.go`
- `pkg/enrich/all.go` only if the analysis entrypoint wants to piggyback on enrich flows

## Suggested Schema Migration

The next mirror schema version should be `3`.

Recommended migration steps:

1. create annotation schema catalog table
2. create observation table
3. create decision table
4. create analysis run table
5. create proposal batch and item tables
6. create supporting indexes
7. seed common annotation schema keys

Recommended indexes:

- observations by `(target_type, target_id)`
- observations by `(annotation_key, source_kind, created_at)`
- decisions by `(target_type, target_id, annotation_key, created_at DESC)`
- proposal items by `(batch_id, decision_state)`
- proposal items by `(target_type, target_id)`

## Pseudocode

### Effective-State Resolution

```text
function resolveEffectiveAnnotation(targetType, targetID, annotationKey):
    decisions = fetch decisions for target ordered newest-first
    if decisions contains locked human decision:
        return that decision
    if decisions contains any human decision:
        return newest human decision

    approved = fetch machine observations explicitly adopted by a human
    if approved exists:
        return approved as effective

    observations = fetch active observations ordered newest-first
    if observations contains llm or heuristic suggestion:
        return advisory-only result with effective=false

    return unknown
```

### Proposal Generation

```text
function generateProposalBatch(scope, prompt, model):
    run = create analysis_run(scope, prompt, model)
    candidates = query corpus for scope
    suggestions = call analyzer(candidates)

    batch = create proposal_batch(status="draft")
    for each suggestion in suggestions:
        if blockedByHumanLock(suggestion.target):
            continue
        create proposal_item(
            batch=batch.id,
            target=suggestion.target,
            action=suggestion.action,
            reason=suggestion.reason,
            confidence=suggestion.confidence,
            decision_state="pending"
        )

    mark run completed
    return batch
```

### Proposal Application

```text
function applyProposalBatch(batchID, actor):
    batch = load batch
    require batch.status in ["reviewed", "partially_reviewed"]

    for item in batch.items:
        if item.decision_state != "approved":
            continue
        if blockedByEffectiveAnnotation(item.target, "safety.do_not_delete"):
            mark item failed with reason
            continue
        execute action
        record result

    mark batch applied
```

## Evidence-Based Design Notes

The sample mailbox strongly supports a multi-entity model.

### Why Sender and Domain Annotations Matter

The ticket script results show:

- `github.com` contributes `702` messages
- `substack.com` contributes `280`
- `privaterelay.appleid.com` contributes `114`
- only one mailbox is present in the sample, but there are hundreds of senders

That means sender- and domain-level judgments will do more work than message-only labels. If a human marks `notifications@github.com` as "important but noisy", that should influence hundreds of future messages without redoing manual work.

### Why Thread Annotations Still Matter

Even though most threads are singletons in this sample, the multi-message threads that do exist are obviously high-value conversational units. A thread is often where "needs reply", "active conversation", or "never bulk archive" belongs.

### Why Message Exceptions Matter

Some senders are mixed. A generally low-value sender can still send one password reset, invoice, or human reply that matters. Message-level exceptions are therefore mandatory.

## Alternatives Considered

### Alternative A: Notes Only

Rejected as the main design because:

- unstructured notes are hard to query
- there is no safe way to drive review queues from prose alone
- conflict resolution becomes impossible

### Alternative B: Simple Labels on Messages Only

Rejected because:

- most leverage sits at sender and domain granularity
- thread and sender reasoning gets duplicated across many messages
- bulk review becomes inefficient

### Alternative C: Store Everything in the App DB

Rejected for the first iteration because:

- the mirror corpus is elsewhere
- cross-database joins and ids become awkward
- CLI and offline workflows get worse

### Alternative D: Let the LLM Write Final State Directly

Rejected because:

- it violates the human-final-authority requirement
- it creates unacceptable risk for delete or unsubscribe workflows
- it makes provenance and audit weak

## Implementation Plan

### Phase 1: Storage and CLI Basics

- add mirror schema v3
- implement annotation schema seeding
- add repository and service layer
- add CLI commands to create observations and decisions
- add list and inspect commands

Definition of done:

- a human can attach a note or structured decision to sender, thread, or message
- effective state can be queried from SQLite and CLI

### Phase 2: Analysis Runs and Proposal Batches

- add analysis run tracking
- add batch proposal storage
- add dry-run proposal generation from scripts or LLM tooling
- add review commands for approve/reject flows

Definition of done:

- the system can generate a reviewable batch but cannot yet perform destructive actions without explicit apply

### Phase 3: HTTP and UI

- add HTTP endpoints
- extend frontend API contracts
- add annotation panels and review queues
- show effective state and provenance in message detail

Definition of done:

- a human can review and override model suggestions from the web UI

### Phase 4: Safe Executors

- implement reversible actions first
- add action audit rows
- add guardrails for `safety.do_not_delete`, reviewer locks, and already-reviewed items
- only then add destructive executors such as delete or unsubscribe

Definition of done:

- approved proposals can be applied with full audit and safety blocks

## Testing Strategy

Back-end tests:

- schema bootstrap tests
- repository CRUD tests
- effective-resolution tests
- precedence and lock tests
- proposal review and apply tests

Integration tests:

- generate observations from a fake analysis run
- approve and reject subsets of a batch
- confirm blocked items never execute when a safety decision exists

Frontend tests:

- annotation panel rendering
- optimistic review actions
- queue filtering by unresolved state and target type

Manual test scripts:

- use the ordered SQL scripts in this ticket to pick realistic candidate senders and threads
- generate a small proposal batch from local mail
- confirm that human decisions override model suggestions in the UI and CLI

## Suggested First Annotation Keys

Keep the first release small and legible.

- `triage.importance`
- `triage.spam_likelihood`
- `triage.newsletter`
- `triage.needs_followup`
- `safety.do_not_delete`
- `sender.trust_level`
- `sender.relationship`
- `notes.general`

That is enough to unlock meaningful review work without inventing twenty overlapping labels.

## Open Questions

- Should human identity be stored as a free-form string in the mirror DB first, or tied immediately to app users?
- Should proposal application operate only on the mirror corpus initially, or also call live IMAP mutations?
- How should repeated rejected suggestions be suppressed so the review queue does not feel haunted?
- Should `domain` decisions automatically seed default suggestions for unseen senders from that domain, or remain advisory only?
- Is there a need for shareable annotation bundles across multiple mirror files, or is one-file portability enough for now?

## Intern-Facing Implementation Checklist

If you hand this ticket to a new intern, tell them to follow this order.

1. Read `pkg/mirror/schema.go`, `pkg/enrich/schema.go`, `pkg/smailnaild/http.go`, `pkg/smailnaild/rules/service.go`, `ui/src/features/mailbox/MailboxExplorer.tsx`, and `ui/src/api/types.ts`.
2. Run `scripts/00-run-all.sh /tmp/smailnail-parallel-a.sqlite` and read the output.
3. Implement mirror schema v3 with annotation tables and indexes.
4. Add repository tests before building the service layer.
5. Implement observation and decision CRUD in CLI first.
6. Add effective-state resolution and lock precedence tests.
7. Only after the data model is stable, add proposal batches.
8. Only after review works, add any executor that can mutate mail state.

## References

Relevant repository files:

- `pkg/mirror/schema.go`
- `pkg/enrich/schema.go`
- `pkg/smailnaild/db.go`
- `pkg/smailnaild/http.go`
- `pkg/smailnaild/rules/types.go`
- `pkg/smailnaild/rules/service.go`
- `ui/src/api/types.ts`
- `ui/src/features/mailbox/MailboxExplorer.tsx`
- `ui/src/features/mailbox/mailboxSlice.ts`

Relevant ticket artifacts:

- `scripts/00-run-all.sh`
- `scripts/01-schema-inventory.sql`
- `scripts/02-message-and-thread-shape.sql`
- `scripts/03-sender-shape.sql`
- `scripts/04-risk-and-unsubscribe-shape.sql`
- `scripts/05-annotation-targets.sql`
- `reference/01-investigation-diary.md`

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
