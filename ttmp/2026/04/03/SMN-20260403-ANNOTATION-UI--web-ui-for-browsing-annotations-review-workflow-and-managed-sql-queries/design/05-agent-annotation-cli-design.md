---
Title: "Agent Annotation CLI Design"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - cli
    - annotations
    - sqlite
    - agents
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Domain types (Annotation, TargetGroup, AnnotationLog, source kinds, review states)"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/repository.go:Repository CRUD that CLI commands call"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/smailnail/commands/annotate/root.go:Existing annotate command tree"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/smailnail/commands/annotate/annotation_add.go:Current annotation add command"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/smailnail/commands/annotate/log_add.go:Current log add command"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/smailnail/commands/annotate/common.go:Shared DB open/bootstrap helper"
ExternalSources: []
Summary: "Design for a CLI tool that LLM agents use to annotate email targets, with built-in run tracking, mandatory logging, and composite workflows."
LastUpdated: 2026-04-03T13:00:00.000000000-04:00
WhatFor: "Give agents a structured, guardrailed interface for creating annotations instead of raw SQL"
WhenToUse: ""
---

# Agent Annotation CLI Design

## 1. Problem

LLM agents that triage email need to write annotations into the smailnail mirror database. Today they can:

1. Use the existing `smailnail annotate annotation add` command (verbose, per-annotation, no run tracking)
2. Write raw SQL against the SQLite file (no validation, no logging, easy to forget required fields)

Neither approach enforces the workflow discipline we need:
- Every agent session should be tracked as a **run** with a unique ID
- Every run should produce at least one **log entry** explaining what was done
- Annotations should be created with consistent source metadata
- Group creation + member addition should be a single atomic operation
- Bulk annotation of many targets with the same tag should be a one-liner

## 2. Design Goals

1. **Run-centric workflow.** Every agent session starts by declaring a run. All subsequent annotations inherit the run ID.
2. **Mandatory logging.** The `end-run` command refuses to complete unless at least one log entry exists.
3. **Bulk-first.** Most commands accept multiple targets in one invocation.
4. **Glazed output.** All commands emit structured rows via Glazed, so agents can parse JSON output to chain commands.
5. **Backwards-compatible.** Existing `smailnail annotate annotation add` etc. still work. New commands are additions, not replacements.

## 3. Command Tree

```
smailnail annotate
├── annotation
│   ├── add          (existing — unchanged)
│   ├── list         (existing — unchanged)
│   └── review       (existing — unchanged)
├── group
│   ├── create       (existing — unchanged)
│   ├── list         (existing — unchanged)
│   ├── members      (existing — unchanged)
│   └── add-target   (existing — unchanged)
├── log
│   ├── add          (existing — unchanged)
│   ├── list         (existing — unchanged)
│   ├── targets      (existing — unchanged)
│   └── link-target  (existing — unchanged)
│
├── run              (NEW — agent workflow commands)
│   ├── start
│   ├── end
│   ├── status
│   └── list
│
├── tag              (NEW — bulk annotation shorthand)
│   ├── add
│   └── remove
│
├── group-with       (NEW — atomic group creation)
│
└── triage           (NEW — composite high-level verbs)
    ├── senders
    └── messages
```

## 4. Command Specifications

### 4.1 `smailnail annotate run start`

Begin a tracked agent run. Prints the run ID to stdout.

```
smailnail annotate run start \
  --sqlite-path smailnail-mirror.sqlite \
  --source-label "triage-pass-1" \
  --created-by "triage-agent"
```

**Behavior:**
- Generates a UUID run ID
- Creates a log entry: kind=`run-start`, title=`"Run started: {source-label}"`, body=`"Agent run initiated by {created-by}"`
- Returns: `{ "run_id": "run-abc-123", "started_at": "..." }`

**Implementation:** The run ID is just a string convention — there is no `runs` table. A "run" is the set of annotations and logs sharing the same `agent_run_id`. The start command creates the first log entry so the run is discoverable.

### 4.2 `smailnail annotate run end`

End a tracked agent run. Validates that logging requirements are met.

```
smailnail annotate run end \
  --sqlite-path smailnail-mirror.sqlite \
  --run-id "run-abc-123" \
  --summary "Annotated 134 senders, 32 threads. Created 2 groups."
```

**Behavior:**
- Counts annotations + log entries for this run ID
- If no user-created log entries exist (only the auto-generated `run-start`), **fails with an error:** `"Run has no log entries. Use 'smailnail annotate log add --agent-run-id run-abc-123' to document what was done."`
- Creates a final log entry: kind=`run-end`, title=`"Run completed"`, body=`{summary}`
- Returns: `{ "run_id": "...", "annotations_created": 189, "logs_created": 4, "groups_created": 2 }`

The `--summary` flag is **required**. If omitted, the command fails.

### 4.3 `smailnail annotate run status`

Show current run stats.

```
smailnail annotate run status \
  --sqlite-path smailnail-mirror.sqlite \
  --run-id "run-abc-123"
```

Returns:
```json
{
  "run_id": "run-abc-123",
  "source_label": "triage-pass-1",
  "annotations": 87,
  "logs": 2,
  "groups": 1,
  "started_at": "2026-04-02T14:23:00Z",
  "latest_at": "2026-04-02T14:27:00Z",
  "by_tag": { "newsletter": 42, "bulk-sender": 25, "important": 20 },
  "by_target_type": { "sender": 65, "thread": 12, "message": 10 }
}
```

### 4.4 `smailnail annotate run list`

List all known runs (aggregated from annotations/logs).

```
smailnail annotate run list \
  --sqlite-path smailnail-mirror.sqlite \
  --limit 20
```

Glazed table output with: run_id, source_label, annotation_count, log_count, first_at, last_at.

### 4.5 `smailnail annotate tag add`

Bulk-annotate multiple targets with the same tag.

```
smailnail annotate tag add \
  --sqlite-path smailnail-mirror.sqlite \
  --run-id "run-abc-123" \
  --source-label "triage-pass-1" \
  --created-by "triage-agent" \
  --tag "newsletter" \
  --note "High volume sender with list-unsubscribe header" \
  --target "sender:news@example.com" \
  --target "sender:digest@weekly.io" \
  --target "sender:updates@saas-tool.com"
```

**Target format:** `{target_type}:{target_id}` — colon-separated.

**Behavior:**
- Creates one annotation per target, all with the same tag/note/source metadata
- All inherit `--run-id` as their `agent_run_id`
- Default `source_kind` is `agent` when `--run-id` is provided
- Returns one Glazed row per annotation created

**From stdin:** For large batches, targets can be piped:
```
echo -e "sender:a@x.com\nsender:b@y.com" | \
  smailnail annotate tag add \
    --sqlite-path smailnail-mirror.sqlite \
    --run-id "run-abc-123" \
    --tag "newsletter" \
    --targets-from-stdin
```

### 4.6 `smailnail annotate tag remove`

Remove annotations matching a tag from specified targets.

```
smailnail annotate tag remove \
  --sqlite-path smailnail-mirror.sqlite \
  --tag "newsletter" \
  --target "sender:news@example.com"
```

**Behavior:** Deletes annotations where `tag = ? AND target_type = ? AND target_id = ?`. Returns deleted count. (This requires adding a `DeleteAnnotation` method to the repository.)

### 4.7 `smailnail annotate group-with`

Create a group and add members atomically.

```
smailnail annotate group-with \
  --sqlite-path smailnail-mirror.sqlite \
  --run-id "run-abc-123" \
  --source-label "triage-pass-1" \
  --created-by "triage-agent" \
  --name "Possible newsletters" \
  --description "Senders with newsletter-like patterns." \
  --member "sender:news@example.com" \
  --member "sender:digest@weekly.io" \
  --member "sender:updates@saas-tool.com"
```

**Behavior:**
- Creates the group
- Adds all members in a single transaction
- Returns: `{ "group_id": "...", "name": "...", "members_added": 3 }`

**From stdin:**
```
smailnail annotate tag add ... --output-format json | \
  jq -r '.target_type + ":" + .target_id' | \
  smailnail annotate group-with \
    --name "Newsletters" \
    --members-from-stdin \
    ...
```

### 4.8 `smailnail annotate triage senders`

High-level composite command for sender triage. Wraps a complete run.

```
smailnail annotate triage senders \
  --sqlite-path smailnail-mirror.sqlite \
  --source-label "newsletter-scan" \
  --created-by "triage-agent" \
  --tag "newsletter" \
  --where-sql "SELECT email FROM senders WHERE has_list_unsubscribe = 1 AND msg_count > 50" \
  --note "High-volume sender with list-unsubscribe header" \
  --group-name "Possible newsletters" \
  --group-description "Auto-detected newsletter senders for review" \
  --log-title "Newsletter sender scan" \
  --log-body "Scanned senders with list-unsubscribe and msg_count > 50."
```

**Behavior (all in one transaction):**
1. Auto-generates a run ID and creates a `run-start` log
2. Executes the `--where-sql` query to get target IDs
3. Creates one annotation per result with the given tag/note
4. If `--group-name` is provided, creates a group and adds all results as members
5. Creates a log entry with the given title/body, linked to all annotated targets
6. Creates a `run-end` log with summary stats
7. Returns full run summary

### 4.9 `smailnail annotate triage messages`

Same pattern as `triage senders`, but targets are messages.

```
smailnail annotate triage messages \
  --sqlite-path smailnail-mirror.sqlite \
  --source-label "large-msg-scan" \
  --created-by "triage-agent" \
  --tag "oversized" \
  --where-sql "SELECT id FROM messages WHERE size_bytes > 10000000" \
  --note "Message exceeds 10MB" \
  --log-title "Large message scan" \
  --log-body "Found messages over 10MB for review."
```

## 5. Run Lifecycle Example

A typical agent session:

```bash
# 1. Start run
RUN_ID=$(smailnail annotate run start \
  --source-label "weekly-triage" \
  --created-by "triage-agent" \
  --output-format json | jq -r .run_id)

# 2. Annotate senders
smailnail annotate tag add \
  --run-id "$RUN_ID" \
  --tag "newsletter" \
  --target "sender:news@example.com" \
  --target "sender:digest@weekly.io" \
  --note "Newsletter pattern detected"

# 3. Create a group
smailnail annotate group-with \
  --run-id "$RUN_ID" \
  --name "Possible newsletters" \
  --description "Review before muting." \
  --member "sender:news@example.com" \
  --member "sender:digest@weekly.io"

# 4. Log what was done (REQUIRED before end-run)
smailnail annotate log add \
  --agent-run-id "$RUN_ID" \
  --source-kind agent \
  --source-label "weekly-triage" \
  --title "Newsletter detection pass" \
  --body "Scanned senders with >50 msgs and list-unsubscribe headers. Found 2."

# 5. Link log to targets
LOG_ID=$(smailnail annotate log list --agent-run-id "$RUN_ID" --limit 1 \
  --output-format json | jq -r '.[0].id')
smailnail annotate log link-target --log-id "$LOG_ID" \
  --target-type sender --target-id "news@example.com"
smailnail annotate log link-target --log-id "$LOG_ID" \
  --target-type sender --target-id "digest@weekly.io"

# 6. End run (fails if no log entries)
smailnail annotate run end \
  --run-id "$RUN_ID" \
  --summary "Annotated 2 senders as newsletters, created 1 group."
```

Or, using the composite triage command (does all of the above in one call):

```bash
smailnail annotate triage senders \
  --source-label "weekly-triage" \
  --created-by "triage-agent" \
  --tag "newsletter" \
  --where-sql "SELECT email FROM senders WHERE has_list_unsubscribe = 1 AND msg_count > 50" \
  --note "Newsletter pattern detected" \
  --group-name "Possible newsletters" \
  --group-description "Review before muting." \
  --log-title "Newsletter detection pass" \
  --log-body "Scanned senders with >50 msgs and list-unsubscribe headers."
```

## 6. Enforcement Rules

| Rule | Enforcement |
|---|---|
| Every annotation should have an `agent_run_id` when created by an agent | `tag add` and `triage` commands auto-set `source_kind=agent` when `--run-id` is provided |
| Every run must have at least one explanatory log | `run end` refuses to complete without one |
| Every run must have a summary | `run end --summary` is required |
| Bulk operations are transactional | `tag add`, `group-with`, `triage` wrap all writes in a single SQLite transaction |

## 7. Repository Changes Needed

| Method | Description |
|---|---|
| `DeleteAnnotationByTagAndTarget(ctx, tag, targetType, targetID)` | For `tag remove` |
| `CreateAnnotationBatch(ctx, []CreateAnnotationInput)` | For bulk `tag add` (single transaction) |
| `CreateGroupWithMembers(ctx, CreateGroupInput, []AddGroupMemberInput)` | For `group-with` (atomic) |
| `CountAnnotationsForRun(ctx, runID)` | For `run status` / `run end` validation |
| `CountLogsForRun(ctx, runID)` | For `run end` validation |
| `AggregateRunStats(ctx, runID)` | For `run status` (by-tag, by-type breakdowns) |
| `ListDistinctRuns(ctx, limit)` | For `run list` |

## 8. Implementation Order

| Phase | Commands | Effort |
|---|---|---|
| **Phase 1** | `run start`, `run end`, `run status`, `run list` | 1 day |
| **Phase 2** | `tag add`, `tag remove` (bulk operations) | 1 day |
| **Phase 3** | `group-with` (atomic create + add members) | 0.5 day |
| **Phase 4** | `triage senders`, `triage messages` (composite) | 1 day |
| **Phase 5** | Repository batch methods, stdin piping | 0.5 day |

Total: ~4 days.
