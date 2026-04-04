---
Title: "UX Functionality Design (v2)"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - sqlite
    - ux-design
    - react
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Annotation, TargetGroup, AnnotationLog domain types"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/schema.go:Schema V3 — annotations, target_groups, annotation_logs tables"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/repository.go:CRUD repository for annotations, groups, and logs"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/schema.go:Messages table, FTS5, enrich columns (threads, senders)"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/enrich/schema.go:Threads and senders enrichment tables"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/http.go:Existing HTTP server to extend"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx:Existing React SPA shell"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/pkg/query/assets.go:Reference: go:embed preset SQL + ResolvePresetSQL"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/serve/handlers_queries.go:Reference: loadSQLDirs, file-based saved queries, preset/query dir pattern"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/serve/serve.go:Reference: --preset-dir and --query-dir CLI flags"
ExternalSources: []
Summary: "v2: Simplified design — agents use CLI (not API), SQL queries stored as files (not database rows), following go-minitrace patterns."
LastUpdated: 2026-04-03T13:00:00.000000000-04:00
WhatFor: "Define the user experience for reviewing LLM annotations and running SQL queries, with file-based query management"
WhenToUse: ""
---

# UX Functionality Design v2

> **Changes from v1:** (1) Removed API-for-agents design — agents write to the SQLite database directly via a CLI tool (see separate CLI design doc). (2) Replaced database-stored saved queries and fragments with file-based SQL management, following go-minitrace's `--preset-dir` / `--query-dir` pattern.

## 1. Context and Problem

Smailnail's mirror database stores mirrored IMAP messages alongside enrichment data (threads, senders, unsubscribe metadata) and LLM-generated annotations. The annotation system has four core entities:

- **Annotations** — tags and notes attached to targets (messages, senders, threads, domains, mailboxes)
- **Target Groups** — named collections of targets (e.g. "Possible newsletters")
- **Annotation Logs** — timestamped narrative entries from agents or humans documenting what was done and why
- **Messages/Senders/Threads** — the underlying email data being annotated

Currently, annotations are created via the CLI (`smailnail annotate`) and queried via raw SQL. There is no visual interface for:

1. Reviewing what LLM agents have annotated
2. Approving, dismissing, or editing annotations in bulk
3. Browsing targets grouped by annotation patterns
4. Running and saving ad-hoc SQL queries against the mirror database

### How agents interact

Agents do **not** use a web API. They interact with the mirror database in two ways:

1. **CLI tool** — `smailnail annotate` subcommands (see CLI design doc). This is the recommended path because the CLI enforces log entries, validates inputs, and maintains agent run tracking.
2. **Direct SQL** — Agents that need raw database access can write SQL against the SQLite file. This is acceptable for read-heavy analysis but discouraged for writes because it bypasses log enforcement.

The web UI is exclusively for **human review** — browsing what agents produced and making review decisions.

## 2. Users and Personas

### Primary: The Email Operator (Manuel)

- Runs LLM triage passes that produce hundreds of annotations per run
- Needs to quickly scan what the agent did, approve correct annotations, dismiss wrong ones
- Wants to drill from a high-level summary down to individual messages
- Writes ad-hoc SQL to investigate patterns, then saves useful queries as `.sql` files for reuse

## 3. Design Principles

1. **Review-first.** The primary flow is reviewing what agents produced, not manually creating annotations.
2. **Drill-down navigation.** Every aggregate view clicks through to underlying targets and annotations.
3. **Batch operations.** Multi-select + bulk approve/dismiss.
4. **SQL as files.** Queries live on disk as `.sql` files, organized in directories. Preset queries ship embedded in the binary. Saved queries are written to a `--query-dir` on the filesystem. No database tables for queries.
5. **Consistent with go-minitrace.** Reuse the same MUI component library, QueryEditor/QuerySidebar/ResultsTable widget architecture, and file-based query management pattern.

## 4. Information Architecture

```
smailnail (existing)
├── Accounts          (existing)
├── Mailbox Explorer  (existing)
├── Rules             (existing)
├── Annotations       (NEW)
│   ├── Review Queue
│   ├── By Target Type
│   │   ├── Senders
│   │   ├── Threads
│   │   ├── Messages
│   │   └── Domains
│   ├── Groups
│   ├── Agent Runs
│   └── Logs
└── Query Editor      (NEW)
    ├── Editor + Results
    ├── Preset Queries (embedded .sql files)
    └── Saved Queries (filesystem .sql files)
```

## 5. Feature Specifications

### 5.1 Review Queue

**Purpose:** Landing page for annotations. Shows all annotations in `to_review` state, newest first.

**Behavior:**
- Filterable by: target type, tag, source kind, source label, agent run ID, date range
- Each row: target type badge, target ID (linked), tag (chip), note (truncated), source badge, age
- Multi-select checkboxes for batch operations
- Batch toolbar: ✓ Approve (→ reviewed), ✗ Dismiss (→ dismissed), ↺ Reset (→ to_review)
- Clicking a row expands inline detail: full annotation, target context, related annotations on same target
- Target ID links navigate to appropriate target detail view

**Header counters:**
- Total to_review count, breakdown by source_kind, top 5 tags

### 5.2 Target-Type Browsers

Four sub-views, one per target type:

#### Senders Browser
- Table of senders from the `senders` enrichment table
- Columns: email, display name, domain, message count, annotation count, top tags, review progress bar
- Click → Sender Detail: metadata, annotations, recent messages, group memberships

#### Threads Browser
- Table from the `threads` enrichment table
- Columns: subject, account/mailbox, message count, participant count, date range, annotation count
- Click → Thread Detail: messages in thread order with inline annotations

#### Messages Browser
- FTS5 full-text search bar for subject, body, from/to
- Columns: date, from, subject, mailbox, annotation count, tags
- Click → Message Detail: headers, body preview (text/html/raw tabs), annotations

#### Domains Browser
- Aggregated: domain, sender count, message count, annotated sender count
- Click → filtered Senders Browser for that domain

### 5.3 Groups

- List with: name, description, member count, source badge, review state
- Click → Group Detail: metadata, rendered markdown description, member list (each linked), related logs
- Actions: change review state, add/remove members

### 5.4 Agent Runs

- List of distinct `agent_run_id` values aggregated from annotations + logs
- Each row: run ID, source label, annotation count, log count, timestamps, review progress
- Click → Agent Run Detail: tag/type breakdowns, chronological timeline, batch approve/dismiss all

### 5.5 Annotation Logs

- Chronological list of log entries
- Each: timestamp, title, source badge, kind, linked targets count
- Click → Log Detail: rendered markdown body, linked targets (each clickable)

### 5.6 Query Editor

**Purpose:** Run ad-hoc SQL against the mirror database. File-based query management.

**Layout:** Three-panel, same as go-minitrace:
- **Left sidebar (240px):** Preset queries + saved queries, organized by folder
- **Top pane:** CodeMirror SQL editor with syntax highlighting and Ctrl+Enter to run
- **Bottom pane:** Results table with sort, export (CSV/JSON), clickable IDs

#### Preset Queries (read-only, embedded)

Preset SQL files are embedded in the Go binary via `go:embed`, exactly like go-minitrace. They ship with the application and cannot be modified by users.

```
pkg/query/presets/
├── annotations/
│   ├── review-queue-counts.sql
│   ├── annotations-by-tag.sql
│   ├── annotations-by-source.sql
│   └── annotations-by-run.sql
├── senders/
│   ├── top-senders.sql
│   ├── newsletter-candidates.sql
│   ├── private-relay-senders.sql
│   └── unsubscribe-coverage.sql
├── threads/
│   ├── longest-threads.sql
│   ├── most-participated.sql
│   └── recent-activity.sql
├── messages/
│   ├── size-distribution.sql
│   ├── attachment-types.sql
│   └── fts-search.sql
└── enrichment/
    ├── sender-domain-breakdown.sql
    └── thread-depth-distribution.sql
```

Each SQL file uses the go-minitrace comment convention:
```sql
-- review-queue-counts.sql
-- Count annotations by review state, tag, and source kind.
SELECT
  review_state,
  tag,
  source_kind,
  COUNT(*) AS count
FROM annotations
GROUP BY review_state, tag, source_kind
ORDER BY count DESC;
```

The first `-- ` comment line becomes the description shown in the sidebar tooltip.

#### Saved Queries (read-write, filesystem)

Saved queries live on disk in directories specified via `--query-dir` flags (default: `./queries`). The server watches these directories and serves their contents. New queries are saved to the first `--query-dir`.

```
queries/
├── my-analysis/
│   ├── noisy-senders.sql
│   └── annotation-coverage.sql
└── team-shared/
    └── weekly-review-status.sql
```

Users create, edit, rename, and delete saved queries through the UI. The UI hits filesystem-backed API endpoints (same as go-minitrace):

| Method | Path | Description |
|---|---|---|
| GET | `/api/query/presets` | List preset queries (embedded) |
| GET | `/api/query/saved` | List saved queries (filesystem) |
| POST | `/api/query/saved` | Create saved query (writes .sql file) |
| PUT | `/api/query/saved/{path}` | Update saved query (overwrites .sql file) |
| DELETE | `/api/query/saved/{path}` | Delete saved query (removes .sql file) |
| POST | `/api/query/execute` | Execute SQL, return columns + rows + timing |

**Source status banner:** When a saved query is loaded, a banner shows the file path. If the file changes on disk (another editor, git pull), the banner offers a "Reload file" button — same as go-minitrace.

**Results → Annotation flow:** When query results contain `target_type` + `target_id` columns, the results table shows action buttons: "Annotate selected", "Create group from results". Clickable IDs navigate to detail views.

### 5.7 Cross-Cutting Features

#### Source Badges
- 🤖 **agent** — blue, shows source_label on hover
- 👤 **human** — green
- ⚙️ **heuristic** — orange
- 📥 **import** — gray

#### Review State Chips
- 🟡 **to_review** — yellow, bold
- 🟢 **reviewed** — green, muted
- ⚫ **dismissed** — gray, strikethrough

#### Keyboard Shortcuts

| Shortcut | Context | Action |
|---|---|---|
| `j` / `k` | Any list view | Move selection down/up |
| `x` | Any list view | Toggle row selection |
| `a` | Review queue | Approve selected |
| `d` | Review queue | Dismiss selected |
| `Enter` | Any list view | Open detail |
| `Esc` | Detail panel | Close detail |
| `Ctrl+Enter` | Query editor | Execute query |
| `Ctrl+S` | Query editor | Save query |
| `/` | Anywhere | Focus search/filter bar |

#### Global Search
FTS5 search across messages + annotation text search + sender/group name search. Results grouped by type.

## 6. Navigation Flow

```
App Shell
├── [Nav] Annotations → Review Queue
│         ├── Click annotation → Inline detail → Click target → Target detail
│         ├── [Tab] Senders → table → Sender detail (annotations, messages, groups)
│         ├── [Tab] Threads → table → Thread detail (messages + annotations)
│         ├── [Tab] Messages → FTS table → Message detail (headers, body, annotations)
│         ├── [Tab] Domains → aggregate → filtered Senders
│         ├── [Tab] Groups → list → Group detail (members + logs)
│         ├── [Tab] Agent Runs → list → Run detail (timeline + batch actions)
│         └── [Tab] Logs → list → Log detail (markdown + linked targets)
│
└── [Nav] Query → Query Editor
          ├── [Sidebar] Presets (embedded .sql) / Saved (filesystem .sql)
          ├── [Editor] SQL input + Run/Save
          └── [Results] Table with export + annotation actions
```

## 7. API Endpoints

### Annotation API (read + review only — agents use CLI)

| Method | Path | Description |
|---|---|---|
| GET | `/api/annotations` | List with filters |
| GET | `/api/annotations/:id` | Get single |
| PATCH | `/api/annotations/:id/review` | Update review state |
| POST | `/api/annotations/batch-review` | Batch update review states |
| GET | `/api/annotation-groups` | List groups |
| GET | `/api/annotation-groups/:id` | Get group with members |
| GET | `/api/annotation-logs` | List logs |
| GET | `/api/annotation-logs/:id` | Get log with targets |
| GET | `/api/annotation-runs` | List distinct agent runs |
| GET | `/api/annotation-runs/:runId` | Get run detail |

### Query API (file-based)

| Method | Path | Description |
|---|---|---|
| POST | `/api/query/execute` | Execute SQL |
| GET | `/api/query/presets` | List embedded preset queries |
| GET | `/api/query/saved` | List saved queries from --query-dir |
| POST | `/api/query/saved` | Save query (creates .sql file) |
| PUT | `/api/query/saved/{path}` | Update saved query |
| DELETE | `/api/query/saved/{path}` | Delete saved query |

### Target Browser APIs

| Method | Path | Description |
|---|---|---|
| GET | `/api/mirror/senders` | Senders list with annotation counts |
| GET | `/api/mirror/senders/:email` | Sender detail with annotations |
| GET | `/api/mirror/threads` | Threads list with annotation counts |
| GET | `/api/mirror/threads/:id` | Thread detail with messages |
| GET | `/api/mirror/messages` | Messages list (FTS) with annotation counts |
| GET | `/api/mirror/messages/:id` | Message detail with annotations |
| GET | `/api/mirror/domains` | Domains aggregate |

## 8. Technology Decisions

| Choice | Rationale |
|---|---|
| **React + MUI** | Consistent with go-minitrace |
| **RTK Query** | Already in smailnail UI |
| **CodeMirror 6** | Same as go-minitrace query editor |
| **File-based SQL** | go-minitrace pattern: `go:embed` for presets, filesystem dirs for saved queries. No DB tables for queries. |
| **CLI for agents** | Agents call `smailnail annotate` commands. No web API needed for writes. |
| **Vite + go:embed** | Existing smailnail build pattern |

## 9. Server Configuration

```
smailnaild serve \
  --preset-dir ./presets/custom \
  --query-dir ./queries \
  --query-dir ./queries/team-shared
```

- `--preset-dir` — additional read-only SQL directories (beyond the embedded presets)
- `--query-dir` — writable directories for user-saved queries; first dir receives new saves
- Defaults: no extra preset dirs, `./queries` as the default query dir

## 10. Incremental Delivery Plan

| Phase | Scope | Effort |
|---|---|---|
| **Phase 1** | Review Queue + annotation list/detail + batch review | 3-4 days |
| **Phase 2** | Query Editor (port from go-minitrace) + embedded presets + file-based saved queries | 2-3 days |
| **Phase 3** | Senders browser + sender detail + domain aggregation | 2 days |
| **Phase 4** | Groups + Agent Runs + Logs views | 2 days |
| **Phase 5** | Threads/Messages browsers + FTS integration | 2 days |
| **Phase 6** | Global search + keyboard shortcuts + polish | 1-2 days |

Total estimated: ~2 weeks.
