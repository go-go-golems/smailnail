---
Title: MVP fast path for annotations, groups, and logs
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
      Note: Current sender and thread target surfaces already present in the mirror
    - Path: pkg/mirror/schema.go
      Note: Mirror schema migration entrypoint for the MVP annotation tables
    - Path: pkg/smailnaild/http.go
      Note: HTTP route surface to extend with lightweight annotation
    - Path: pkg/smailnaild/rules/service.go
      Note: Existing preview-first service pattern that can inform simple annotation APIs
    - Path: ui/src/api/types.ts
      Note: Frontend contracts to extend for annotations
    - Path: ui/src/features/mailbox/MailboxExplorer.tsx
      Note: Message and mailbox UI surface for a first annotation sidebar or review panel
    - Path: ui/src/features/mailbox/mailboxSlice.ts
      Note: State model to extend with to_review and annotation state
ExternalSources: []
Summary: A minimal annotation design for getting started quickly with generic targets, tags, notes, review state, groups, and linkable log entries.
LastUpdated: 2026-04-02T20:25:00-04:00
WhatFor: Use this as the implementation target for the first annotation MVP instead of the larger long-term design.
WhenToUse: Read this when the goal is to ship a working annotation system quickly and defer complex decision engines or proposal workflows.
---


# MVP fast path for annotations, groups, and logs

## Why This Exists

The larger annotation design is useful for long-term direction, but it is heavier than necessary for getting real usage quickly. For the first iteration, the system should optimize for:

- being easy to implement
- being easy to inspect with SQL
- being easy to evolve later
- capturing real human and agent behavior now

This MVP deliberately avoids a separate decisions layer, proposal engine, or annotation schema catalog. The goal is to start storing real review data immediately and learn from it.

## Core Idea

Use one generic target reference:

- `target_type`
- `target_id`

Then add only four concepts:

1. annotations on targets
2. groups of targets
3. log entries
4. links from log entries to one or more targets

That gives us enough to do all of the following:

- tag a sender as important
- tag a message as suspicious
- leave a free-form note on a thread
- mark something `to_review` or `reviewed`
- record which agent run created an annotation
- create a group like "April possible newsletters"
- attach one log entry to many messages or senders

## Minimal Data Model

### 1. Annotations

This is the main table.

```sql
CREATE TABLE annotations (
    id TEXT PRIMARY KEY,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    tag TEXT NOT NULL DEFAULT '',
    note_markdown TEXT NOT NULL DEFAULT '',
    source_kind TEXT NOT NULL DEFAULT 'human',
    source_label TEXT NOT NULL DEFAULT '',
    agent_run_id TEXT NOT NULL DEFAULT '',
    review_state TEXT NOT NULL DEFAULT 'to_review',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_annotations_target
    ON annotations(target_type, target_id);

CREATE INDEX idx_annotations_review_state
    ON annotations(review_state, created_at);

CREATE INDEX idx_annotations_tag
    ON annotations(tag, created_at);
```

Interpretation:

- `tag` is the structured label for now
- `note_markdown` is the free-form explanation
- `agent_run_id` records provenance when an agent created it
- `review_state` is the only workflow state in v1

Recommended `review_state` values:

- `to_review`
- `reviewed`
- `dismissed`

Recommended `source_kind` values:

- `human`
- `agent`
- `heuristic`
- `import`

This table is intentionally permissive. A row can be:

- tag-only
- note-only
- tag plus note

### 2. Groups

Groups are named sets of targets.

```sql
CREATE TABLE target_groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    source_kind TEXT NOT NULL DEFAULT 'human',
    source_label TEXT NOT NULL DEFAULT '',
    agent_run_id TEXT NOT NULL DEFAULT '',
    review_state TEXT NOT NULL DEFAULT 'to_review',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE target_group_members (
    group_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, target_type, target_id)
);

CREATE INDEX idx_target_group_members_target
    ON target_group_members(target_type, target_id);
```

Examples:

- `Likely newsletters to review`
- `Possible spam cluster from April 2`
- `Important people and institutions`
- `Messages to archive after human review`

Groups are useful because analysis often starts as a cluster, not a final judgment.

### 3. Log Entries

Log entries record activity or reasoning and can be linked to many targets.

```sql
CREATE TABLE annotation_logs (
    id TEXT PRIMARY KEY,
    log_kind TEXT NOT NULL DEFAULT 'note',
    title TEXT NOT NULL DEFAULT '',
    body_markdown TEXT NOT NULL DEFAULT '',
    source_kind TEXT NOT NULL DEFAULT 'human',
    source_label TEXT NOT NULL DEFAULT '',
    agent_run_id TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE annotation_log_targets (
    log_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    PRIMARY KEY (log_id, target_type, target_id)
);

CREATE INDEX idx_annotation_log_targets_target
    ON annotation_log_targets(target_type, target_id);
```

This is the answer to "1 log entry, n things".

Examples:

- one agent run says "these 42 senders look promotional because of list-unsubscribe and low reply density"
- one human review note says "do not bulk-delete this cluster until we separate financial alerts from marketing"
- one debugging log says "GitHub notifications are noisy but still important"

### 4. Optional Link from Logs to Groups

If you want one log entry to refer to an entire group directly, add this small join table:

```sql
CREATE TABLE annotation_log_groups (
    log_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    PRIMARY KEY (log_id, group_id)
);
```

This is optional for the first implementation. You can start by linking logs to individual targets only.

## Target Types for MVP

Do not overexpand the first release. Start with:

- `message`
- `thread`
- `sender`
- `domain`
- `mailbox`
- `account`

Recommended `target_id` values:

- `message`: `messages.id` as text
- `thread`: `threads.thread_id`
- `sender`: normalized email
- `domain`: normalized domain
- `mailbox`: `account_key || '::' || mailbox_name`
- `account`: `account_key`

## Suggested Tags

Keep tags free-form for now, but start with a recommended vocabulary in docs and UI suggestions.

Good first tags:

- `important`
- `suspected-spam`
- `newsletter`
- `promotional`
- `needs-followup`
- `do-not-delete`
- `trusted-sender`
- `watch`
- `review-later`

The key point is that tags are not enforced yet. That keeps the MVP simple.

## Review Model

The only workflow state we need at first is whether the item still needs human attention.

Recommended behavior:

- any annotation created by an agent defaults to `to_review`
- any annotation created by a human may default to `reviewed`
- groups created by an agent default to `to_review`
- log entries do not need review state unless you decide later that they should

This is intentionally shallow. The MVP is about collecting human judgment and agent suggestions in one place, not solving every conflict.

## What We Are Explicitly Not Building Yet

- no separate decision table
- no effective-state resolver
- no proposal batches
- no safe bulk-action executor
- no tag registry table
- no precedence engine between humans and agents

If a human wants to override an agent in the MVP, they simply add a reviewed annotation or update the existing one. We can formalize that later after we have real usage data.

## Recommended Queries

### List unresolved agent annotations

```sql
SELECT *
FROM annotations
WHERE source_kind = 'agent'
  AND review_state = 'to_review'
ORDER BY created_at DESC;
```

### Show all annotations for a sender

```sql
SELECT *
FROM annotations
WHERE target_type = 'sender'
  AND target_id = 'notifications@github.com'
ORDER BY created_at DESC;
```

### Show all targets in a group

```sql
SELECT *
FROM target_group_members
WHERE group_id = ?;
```

### Show logs linked to a message

```sql
SELECT l.*
FROM annotation_logs l
JOIN annotation_log_targets lt ON lt.log_id = l.id
WHERE lt.target_type = 'message'
  AND lt.target_id = ?;
```

## Example Workflow

### Agent-driven review pass

1. Agent scans recent messages.
2. Agent creates:
   - one group called `Possible newsletters from 2026-04-02`
   - one annotation per sender with tag `newsletter`
   - one log entry describing its reasoning
3. All agent-created annotations and groups are marked `to_review`.
4. Human opens the queue and:
   - changes some rows to `reviewed`
   - edits a few tags
   - adds `do-not-delete` notes to exceptions

### Human note-taking workflow

1. Human opens a message or sender.
2. Human adds:
   - tag `important`
   - note `Personal contact, never bulk archive`
   - review state `reviewed`
3. Future agent runs can still add notes, but they do not erase the human entry.

## API Sketch

Minimal HTTP endpoints:

- `GET /api/annotations`
- `POST /api/annotations`
- `PATCH /api/annotations/{id}`
- `GET /api/groups`
- `POST /api/groups`
- `POST /api/groups/{id}/members`
- `GET /api/logs`
- `POST /api/logs`
- `POST /api/logs/{id}/targets`

Minimal CLI commands:

- `smailnail annotate add`
- `smailnail annotate list`
- `smailnail annotate update`
- `smailnail annotate group create`
- `smailnail annotate group add-target`
- `smailnail annotate log add`
- `smailnail annotate log link-target`

## Implementation Order

Build this in the smallest useful slices.

1. Add schema migration with `annotations`.
2. Add repository methods and CLI for annotation CRUD.
3. Add `target_groups` and `target_group_members`.
4. Add `annotation_logs` and `annotation_log_targets`.
5. Add basic list/filter endpoints in HTTP.
6. Add a simple UI sidebar and a `to_review` queue.

That is enough for real usage.

## Why This MVP Is Good Enough

This MVP captures the most important things immediately:

- what was tagged
- what note was left
- who or what created it
- whether a human still needs to review it
- how to organize targets into groups
- how to attach one reasoning log to many targets

It also keeps the escape hatch open for later:

- groups can become proposal batches later
- logs can become analysis runs later
- annotations can split into observations and decisions later

## Recommended Next Step After MVP

Once this is in use for a bit, inspect:

- which tags actually get used
- whether humans want explicit override semantics
- whether groups are mostly exploratory or action-oriented
- whether logs need structured fields beyond markdown

Only then decide whether to add the heavier architecture.
