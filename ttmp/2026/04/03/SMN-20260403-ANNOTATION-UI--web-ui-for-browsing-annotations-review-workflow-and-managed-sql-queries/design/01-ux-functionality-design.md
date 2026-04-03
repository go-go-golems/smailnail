---
Title: UX Functionality Design
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
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/http.go:Existing HTTP server with accounts, messages, rules API"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx:Existing React SPA shell"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx:Reference: SQL editor with sidebar, CodeMirror, results table"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/SessionBrowser/SessionBrowser.tsx:Reference: filterable table with summary stats"
ExternalSources: []
Summary: "Complete UX design for the smailnail annotation review UI, managed SQL queries, and reusable fragments."
LastUpdated: 2026-04-03T12:00:00.000000000-04:00
WhatFor: "Define the full user experience for reviewing LLM-generated annotations, browsing targets, managing SQL queries against the mirror database, and composing reusable query fragments"
WhenToUse: ""
---

# UX Functionality Design: Annotation Review & Query UI

## 1. Context and Problem

Smailnail's mirror database stores mirrored IMAP messages alongside enrichment data (threads, senders, unsubscribe metadata) and LLM-generated annotations. The annotation system has four core entities:

- **Annotations** — tags and notes attached to targets (messages, senders, threads, domains, mailboxes)
- **Target Groups** — named collections of targets (e.g. "Possible newsletters")
- **Annotation Logs** — timestamped narrative entries from agents or humans documenting what was done and why
- **Messages/Senders/Threads** — the underlying email data being annotated

Currently, annotations are created and queried exclusively through CLI commands (`smailnail annotate annotation add/list/review`, `smailnail annotate group create/list`, etc.) and via raw SQL queries. There is no visual interface for:

1. Reviewing what LLM agents have annotated
2. Approving, dismissing, or editing annotations
3. Browsing targets grouped by annotation patterns
4. Running and saving ad-hoc SQL queries against the mirror database
5. Building and sharing reusable query fragments

The existing smailnail web UI (React SPA, served by `smailnaild`) has account management, mailbox browsing, and rule management. This design extends it with annotation review and SQL query capabilities.

## 2. Users and Personas

### Primary: The Email Operator (Manuel)

- Runs LLM triage passes that produce hundreds of annotations per run
- Needs to quickly scan what the agent did, approve correct annotations, dismiss wrong ones
- Wants to drill from a high-level summary down to individual messages
- Writes ad-hoc SQL to investigate patterns, then saves useful queries for reuse
- Switches frequently between reviewing annotations and exploring data with SQL

### Secondary: Future LLM Agents (via API)

- Consume the same API endpoints to read/write annotations programmatically
- No direct UI interaction, but the UI must show their work clearly

## 3. Design Principles

1. **Review-first, not creation-first.** The primary flow is reviewing what agents produced, not manually creating annotations. Creation is a secondary action.
2. **Drill-down navigation.** Every aggregate view (sender summary, tag cloud, group list) should click-through to the underlying targets and annotations.
3. **Batch operations.** Reviewing annotations one at a time is too slow. Support multi-select + bulk approve/dismiss.
4. **SQL as a first-class citizen.** Power users live in SQL. The query editor should feel native, not bolted-on. Follow the go-minitrace pattern: sidebar with presets/saved queries, CodeMirror editor, results table with export.
5. **Consistent visual language.** Reuse the MUI component library and data-widget/data-part attribute patterns from go-minitrace for Storybook-ready widgets.

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
    ├── Preset Queries
    ├── Saved Queries
    └── Fragments Library
```

## 5. Feature Specifications

### 5.1 Review Queue

**Purpose:** The landing page for the annotation section. Shows all annotations in `to_review` state, newest first.

**Behavior:**
- Filterable by: target type, tag, source kind, source label, agent run ID, date range
- Each row shows: target type badge, target ID (linked), tag (chip), note (truncated), source badge, age
- Multi-select with checkboxes for batch operations
- Batch actions toolbar: ✓ Approve (→ reviewed), ✗ Dismiss (→ dismissed), ↺ Reset (→ to_review)
- Clicking a row expands an inline detail panel showing the full annotation, target context, and related annotations on the same target
- The target ID is a link: clicking it navigates to the appropriate target detail view (sender profile, message detail, thread view)

**Counters in header:**
- Total to_review count
- Breakdown by source_kind (agent vs heuristic)
- Breakdown by tag (top 5 tags)

### 5.2 Target-Type Browsers

Four sub-views, one per target type, each following the same pattern:

#### 5.2.1 Senders Browser

**Purpose:** Browse senders with their annotation summaries.

**Behavior:**
- Table of senders from the `senders` enrichment table
- Columns: email, display name, domain, message count, annotation count, latest tag, review progress (bar showing reviewed/to_review/dismissed)
- Filter bar: domain, message count range, has annotations, review state
- Click a sender row → Sender Detail view
- Sender Detail shows:
  - Sender metadata (email, domain, private relay status, unsubscribe links)
  - All annotations on this sender
  - Recent messages from this sender (link to message detail)
  - Groups this sender belongs to

#### 5.2.2 Threads Browser

- Table of threads from the `threads` enrichment table
- Columns: subject, account/mailbox, message count, participant count, date range, annotation count
- Click → Thread Detail showing messages in thread order with inline annotations

#### 5.2.3 Messages Browser

- Table with columns: date, from, subject, mailbox, annotation count, tags
- Uses FTS5 for full-text search bar
- Click → Message Detail showing headers, body preview, all annotations, link to raw message

#### 5.2.4 Domains Browser

- Aggregated view: domain, sender count, message count, annotation count
- Click → filtered Senders Browser for that domain

### 5.3 Groups

**Purpose:** Browse and manage target groups created by agents or humans.

**Behavior:**
- List of groups with: name, description, member count, source badge, review state
- Click a group → Group Detail:
  - Group metadata and description (rendered markdown)
  - Member list (target type + target ID, each linked)
  - Annotations on the group itself
  - Logs referencing this group's members
- Actions: change review state, add/remove members, edit description

### 5.4 Agent Runs

**Purpose:** See what each agent run did at a glance.

**Behavior:**
- List of distinct `agent_run_id` values, aggregated from annotations + logs
- Each row: run ID, source label, annotation count, log count, first/last timestamp, review progress
- Click → Agent Run Detail:
  - Timeline of all annotations created in this run (chronological)
  - All logs from this run
  - Batch approve/dismiss all annotations from this run
  - Statistics: annotations by tag, by target type

### 5.5 Annotation Logs

**Purpose:** Browse the narrative log entries that agents produce to explain their reasoning.

**Behavior:**
- Chronological list of log entries
- Each entry: timestamp, title, source badge, log kind, linked targets count
- Click → Log Detail:
  - Full markdown body (rendered)
  - Linked targets (each a clickable link)
  - Related annotations (annotations on the same targets created in the same agent run)

### 5.6 Query Editor

**Purpose:** Run ad-hoc SQL against the mirror database. Modeled closely on go-minitrace's QueryEditor.

**Layout:** Three-panel:
- **Left sidebar (240px):** Preset queries organized by folder, saved queries, fragments library
- **Top pane:** CodeMirror SQL editor with syntax highlighting and Ctrl+Enter to run
- **Bottom pane:** Results table with sort, export (CSV/JSON), clickable IDs

**Preset queries** are read-only SQL files shipped with the application, organized by category:
- `annotations/` — review queue counts, annotations by tag, by source
- `senders/` — top senders, newsletter candidates, private relay senders
- `threads/` — longest threads, most-participated, recent activity
- `messages/` — size distribution, attachment types, FTS search
- `enrichment/` — unsubscribe coverage, sender domain breakdown

**Saved queries** are user-created, stored in the hosted database (new table):
```sql
CREATE TABLE IF NOT EXISTS saved_queries (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    folder TEXT NOT NULL DEFAULT 'ungrouped',
    description TEXT NOT NULL DEFAULT '',
    sql_text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Fragments Library** is a special section of the sidebar for reusable SQL fragments:
```sql
CREATE TABLE IF NOT EXISTS query_fragments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    description TEXT NOT NULL DEFAULT '',
    sql_text TEXT NOT NULL,
    placeholder_names TEXT NOT NULL DEFAULT '[]',  -- JSON array of placeholder names
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Fragments are building blocks, not complete queries. Example fragments:
- "Active senders (>N messages)" → `SELECT email, msg_count FROM senders WHERE msg_count > $min_count`
- "Annotations by run" → `SELECT * FROM annotations WHERE agent_run_id = $run_id`
- "FTS match" → `SELECT id, subject, from_summary FROM messages JOIN messages_fts ON messages.id = messages_fts.rowid WHERE messages_fts MATCH $query`

**Fragment insertion:** Clicking a fragment in the sidebar inserts it at the cursor position in the editor. Placeholders (`$name`) are highlighted for the user to fill in.

**Results → Annotation flow:** When query results contain recognized column patterns (`target_type` + `target_id`, or `id` from annotations), the results table shows action buttons:
- "Annotate selected" — opens a bulk annotation dialog
- "Create group from results" — creates a target group from the result set
- Clickable IDs navigate to the corresponding detail views

### 5.7 Cross-Cutting Features

#### 5.7.1 Source Badges

All annotation-producing entities show their source as a colored badge:
- 🤖 **agent** — blue badge, shows source_label on hover
- 👤 **human** — green badge
- ⚙️ **heuristic** — orange badge
- 📥 **import** — gray badge

#### 5.7.2 Review State Chips

Consistent styling across all views:
- 🟡 **to_review** — yellow chip, bold
- 🟢 **reviewed** — green chip, muted
- ⚫ **dismissed** — gray chip, strikethrough on text

#### 5.7.3 Keyboard Shortcuts

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

#### 5.7.4 Global Search

A search bar in the app header that performs:
1. FTS5 search across messages (subject, body, from/to)
2. Annotation text search (tag + note)
3. Sender email/domain search
4. Group name search

Results are grouped by type with counts, click-through to detail views.

## 6. Navigation Flow

```
App Shell
├── [Nav] Annotations → Review Queue
│         │
│         ├── Filter/select annotations
│         ├── Batch approve/dismiss
│         ├── Click annotation row → Inline detail
│         │     └── Click target ID → Target detail
│         │
│         ├── [Tab] Senders → Senders table
│         │     └── Click row → Sender detail
│         │           ├── Annotations list
│         │           ├── Messages list (linked)
│         │           └── Groups membership
│         │
│         ├── [Tab] Threads → Threads table
│         │     └── Click row → Thread detail (messages + annotations)
│         │
│         ├── [Tab] Messages → Messages table (FTS)
│         │     └── Click row → Message detail
│         │
│         ├── [Tab] Domains → Domains aggregate
│         │     └── Click row → filtered Senders
│         │
│         ├── [Tab] Groups → Groups list
│         │     └── Click group → Group detail (members + logs)
│         │
│         ├── [Tab] Agent Runs → Runs list
│         │     └── Click run → Run detail (timeline + batch actions)
│         │
│         └── [Tab] Logs → Logs list
│               └── Click log → Log detail (markdown + linked targets)
│
└── [Nav] Query → Query Editor
          ├── [Sidebar] Presets / Saved / Fragments
          ├── [Editor] SQL input + Run/Save
          └── [Results] Table with export + annotation actions
```

## 7. API Endpoints Required

### Annotation API (new)

| Method | Path | Description |
|---|---|---|
| GET | `/api/annotations` | List with filters (target_type, tag, review_state, source_kind, agent_run_id, limit, offset) |
| GET | `/api/annotations/:id` | Get single annotation |
| PATCH | `/api/annotations/:id/review` | Update review state |
| POST | `/api/annotations/batch-review` | Batch update review states |
| POST | `/api/annotations` | Create annotation |
| GET | `/api/annotation-groups` | List groups with filters |
| GET | `/api/annotation-groups/:id` | Get group with members |
| POST | `/api/annotation-groups` | Create group |
| POST | `/api/annotation-groups/:id/members` | Add member |
| DELETE | `/api/annotation-groups/:id/members` | Remove member |
| GET | `/api/annotation-logs` | List logs |
| GET | `/api/annotation-logs/:id` | Get log with linked targets |
| POST | `/api/annotation-logs` | Create log |
| GET | `/api/annotation-runs` | List distinct agent runs (aggregated) |
| GET | `/api/annotation-runs/:runId` | Get run detail (annotations + logs for run) |

### Query API (new)

| Method | Path | Description |
|---|---|---|
| POST | `/api/query/execute` | Execute SQL, return columns + rows + timing |
| GET | `/api/query/presets` | List preset queries |
| GET | `/api/query/saved` | List saved queries for user |
| POST | `/api/query/saved` | Save query |
| PATCH | `/api/query/saved/:id` | Update saved query |
| DELETE | `/api/query/saved/:id` | Delete saved query |
| GET | `/api/query/fragments` | List fragments for user |
| POST | `/api/query/fragments` | Create fragment |
| PATCH | `/api/query/fragments/:id` | Update fragment |
| DELETE | `/api/query/fragments/:id` | Delete fragment |

### Target Browser APIs (new)

| Method | Path | Description |
|---|---|---|
| GET | `/api/mirror/senders` | Senders list with annotation counts |
| GET | `/api/mirror/senders/:email` | Sender detail with annotations |
| GET | `/api/mirror/threads` | Threads list with annotation counts |
| GET | `/api/mirror/threads/:id` | Thread detail with messages |
| GET | `/api/mirror/messages` | Messages list (FTS support) with annotation counts |
| GET | `/api/mirror/messages/:id` | Message detail with annotations |
| GET | `/api/mirror/domains` | Domains aggregate |

## 8. Technology Decisions

| Choice | Rationale |
|---|---|
| **React + MUI** | Consistent with go-minitrace; MUI provides the dense data-table components needed |
| **RTK Query** | Already used in smailnail UI (via Redux Toolkit); handles caching and cache invalidation on review state changes |
| **CodeMirror 6** | Same as go-minitrace query editor; SQL language mode + syntax highlighting |
| **react-router** | Same routing pattern as go-minitrace |
| **Vite** | Already the smailnail UI bundler |
| **go:embed** | Existing production build pattern — SPA assets embedded into Go binary |

## 9. Incremental Delivery Plan

| Phase | Scope | Effort |
|---|---|---|
| **Phase 1** | Review Queue + annotation list/detail + batch review actions | 3-4 days |
| **Phase 2** | Query Editor (port from go-minitrace) + presets + saved queries | 2-3 days |
| **Phase 3** | Senders browser + sender detail + domain aggregation | 2 days |
| **Phase 4** | Groups + Agent Runs + Logs views | 2 days |
| **Phase 5** | Threads/Messages browsers + FTS integration | 2 days |
| **Phase 6** | Fragments library + cross-navigation from query results | 1-2 days |
| **Phase 7** | Global search + keyboard shortcuts + polish | 1-2 days |

Total estimated: ~2-3 weeks of focused development.
