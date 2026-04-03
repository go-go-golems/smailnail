---
Title: "UX Design Brief for External Designers"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - ux-design
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Self-contained brief giving external UX designers full context — problem, users, data model, constraints — without prescribing solutions."
LastUpdated: 2026-04-03T13:00:00.000000000-04:00
WhatFor: "Hand to UX designers so they can produce their own design proposals"
WhenToUse: ""
---

# UX Design Brief: Email Annotation Review Interface

## What This Document Is

This is a self-contained brief for UX designers. It explains the situation, the problem, the data model, the users, and the constraints. It does **not** prescribe screens, layouts, or interaction patterns — that is what we're asking you to design.

---

## 1. The Product

**Smailnail** is a personal email management tool that mirrors IMAP mailboxes into a local SQLite database and then runs analysis passes over the data. It is a developer tool — a CLI + web UI used by a single technical operator who manages their own email.

The tool has three layers:

1. **Mirror** — syncs emails from IMAP servers into a SQLite database with full-text search
2. **Enrichment** — derives structure from raw email: thread reconstruction, sender normalization, unsubscribe link extraction
3. **Annotation** — LLM agents and heuristic scripts annotate emails and senders with tags, notes, and groupings

The existing web UI handles IMAP account management, mailbox browsing, and mail processing rules. We need to extend it with an interface for **reviewing what the annotation layer has produced**.

## 2. The Problem

LLM agents run triage passes over the email corpus. A single pass might produce **hundreds of annotations** — tagging senders as newsletters, flagging important threads, grouping bulk senders, noting marketing patterns. The agent also writes **log entries** explaining its reasoning.

Today there is no way to visually review this work. The operator must use command-line queries and raw SQL to inspect what the agent did, then approve or dismiss annotations one at a time. This is too slow.

We need an interface that lets the operator:

- See what changed since the last time they looked
- Understand the agent's reasoning (via log entries)
- Approve correct annotations quickly (ideally in bulk)
- Dismiss wrong ones
- Drill into the underlying email data when something looks off
- Run ad-hoc SQL queries when the structured views aren't enough

## 3. The Data Model

All data lives in a single SQLite file. Here are the entities and their relationships.

### 3.1 Messages

The core email record. ~50K–200K rows in a typical installation.

| Field | Description |
|---|---|
| `id` | Auto-increment integer primary key |
| `account_key` | Which email account (e.g. `gmail-993-user-abc123`) |
| `mailbox_name` | IMAP folder (e.g. `INBOX`, `Archive`) |
| `uid` | IMAP unique ID within the mailbox |
| `message_id` | RFC 5322 Message-ID header |
| `internal_date` | Server timestamp |
| `sent_date` | Parsed Date header |
| `subject` | Subject line |
| `from_summary` | Display-friendly "Name <email>" |
| `to_summary` | Same for To |
| `cc_summary` | Same for Cc |
| `size_bytes` | Message size |
| `flags_json` | JSON array of IMAP flags (Seen, Flagged, etc.) |
| `body_text` | Extracted plain text body |
| `body_html` | Extracted HTML body |
| `has_attachments` | Boolean |
| `thread_id` | (enriched) Thread identifier |
| `sender_email` | (enriched) Normalized sender email |
| `sender_domain` | (enriched) Sender's domain |

Messages are also indexed in a **full-text search table** (SQLite FTS5) covering subject, from, to, cc, body text, and body HTML.

### 3.2 Threads

Reconstructed conversation threads. Derived from Message-ID / In-Reply-To / References headers.

| Field | Description |
|---|---|
| `thread_id` | Primary key, assigned during enrichment |
| `subject` | Normalized thread subject |
| `message_count` | How many messages in the thread |
| `participant_count` | Distinct senders |
| `first_sent_date` / `last_sent_date` | Thread date range |

### 3.3 Senders

Normalized sender profiles, one per unique email address.

| Field | Description |
|---|---|
| `email` | Primary key |
| `display_name` | Common display name |
| `domain` | Email domain |
| `is_private_relay` | Apple/iCloud relay address |
| `msg_count` | Total messages from this sender |
| `first_seen_date` / `last_seen_date` | Activity range |
| `unsubscribe_mailto` | Extracted List-Unsubscribe mailto link |
| `unsubscribe_http` | Extracted List-Unsubscribe HTTP link |
| `has_list_unsubscribe` | Whether the sender includes unsubscribe headers |

### 3.4 Annotations

Tags and notes attached to any **target** in the system.

| Field | Description |
|---|---|
| `id` | UUID |
| `target_type` | What kind of thing is annotated: `message`, `sender`, `thread`, `domain`, `mailbox`, `account` |
| `target_id` | Identifier within the target type (e.g. an email address for senders, a message ID for messages) |
| `tag` | Short free-form label (e.g. `newsletter`, `important`, `bulk-sender`, `ignore`) |
| `note_markdown` | Longer explanation in Markdown |
| `source_kind` | Who created this: `human`, `agent`, `heuristic`, `import` |
| `source_label` | Specific source name (e.g. `triage-pass-1`, `newsletter-scanner`) |
| `agent_run_id` | Groups annotations by agent session (e.g. `run-42`) |
| `review_state` | **`to_review`**, **`reviewed`**, or **`dismissed`** |
| `created_by` | Free-form creator identifier |
| `created_at` / `updated_at` | Timestamps |

**Review state** is the central workflow lever:
- `to_review` — default for agent-created annotations; needs human eyes
- `reviewed` — human confirmed the annotation is correct
- `dismissed` — human decided the annotation is wrong

### 3.5 Target Groups

Named collections of targets. Agents create groups to organize related items (e.g. "Possible newsletters" containing 12 sender emails).

| Field | Description |
|---|---|
| `id` | UUID |
| `name` | Human-readable group name |
| `description` | Markdown explanation of what this group represents |
| `source_kind` / `source_label` / `agent_run_id` | Same provenance tracking as annotations |
| `review_state` | Same tri-state as annotations |

Groups have **members** (a join table):
| Field | Description |
|---|---|
| `group_id` | → Target Group |
| `target_type` | e.g. `sender` |
| `target_id` | e.g. `news@example.com` |

### 3.6 Annotation Logs

Narrative entries that agents write to explain what they did and why. Think of these as the agent's diary.

| Field | Description |
|---|---|
| `id` | UUID |
| `log_kind` | Type of entry (e.g. `note`, `run-start`, `run-end`) |
| `title` | Short summary |
| `body_markdown` | Detailed explanation in Markdown |
| `source_kind` / `source_label` / `agent_run_id` | Provenance |
| `created_by` / `created_at` | Timestamps |

Logs have **linked targets** (a join table):
| Field | Description |
|---|---|
| `log_id` | → Annotation Log |
| `target_type` | e.g. `sender` |
| `target_id` | e.g. `news@example.com` |

### 3.7 Entity Relationship Summary

```
Messages ──────── Threads (via thread_id)
    │                 │
    │                 │
    ▼                 ▼
 [can be annotated targets]
    │                 │
    ▼                 ▼
Senders ─────── Domains (via sender_domain)
    │                 │
    │                 │
    ▼                 ▼
 [can be annotated targets]
    │
    ▼
Annotations ◄──── agent_run_id ────► Annotation Logs
    │                                     │
    ▼                                     ▼
Target Groups ◄── members ──►       Log Targets
    (groups collect targets)     (logs reference targets)
```

Everything connects through the `target_type` + `target_id` pair. An annotation on a sender has `target_type = "sender"` and `target_id = "news@example.com"`. The same targeting convention is used for group members and log links.

## 4. The User

**One person.** Manuel, a software developer who operates his own email infrastructure. He has:

- ~100K mirrored messages across 3 accounts
- ~1,200 distinct senders
- ~8,400 threads
- Agent triage runs that produce 50–300 annotations per run, roughly weekly

He is comfortable with SQL and command-line tools. The web UI is for when he wants a visual overview and needs to make review decisions quickly. He does not want to context-switch between terminals — the UI should be self-sufficient for the review workflow.

## 5. The SQL Query Workbench

In addition to the structured annotation review, the operator needs a **SQL query editor** embedded in the web UI. This is for exploration that doesn't fit the predefined views — pattern analysis, one-off investigations, ad-hoc reports.

The query system:
- Executes SQL against the same SQLite database
- Ships with **preset queries** (read-only `.sql` files embedded in the application binary)
- Supports **saved queries** (read-write `.sql` files on the local filesystem)
- Returns tabular results that can be sorted, exported to CSV/JSON, and clicked through to detail views

Query files use a simple convention: the first `-- ` comment in the file becomes the description shown in the UI.

```sql
-- Top senders by message count with annotation coverage
SELECT s.email, s.domain, s.msg_count,
  COUNT(a.id) AS annotations
FROM senders s
LEFT JOIN annotations a ON a.target_type = 'sender' AND a.target_id = s.email
GROUP BY s.email
ORDER BY s.msg_count DESC
LIMIT 50;
```

## 6. Scale and Performance Expectations

| Entity | Typical Count |
|---|---|
| Messages | 50K – 200K |
| Senders | 500 – 2,000 |
| Threads | 5K – 20K |
| Annotations | 500 – 5,000 |
| Target groups | 10 – 50 |
| Annotation logs | 20 – 100 |
| Agent runs | 5 – 30 |

All data is local SQLite. Queries are fast (single-digit milliseconds for indexed lookups, 10–100ms for aggregations). The UI does not need to worry about network latency or pagination for most views, though message lists benefit from pagination given their size.

## 7. Constraints

| Constraint | Detail |
|---|---|
| **Single user** | No multi-user features, no permissions, no collaboration. The UI serves one person. |
| **Local data** | Everything is in a SQLite file on the operator's machine. No cloud, no sync. |
| **Existing UI** | There is already a React SPA with account management, mailbox browsing, and rules. The new features must integrate into it. |
| **Agent interaction is CLI-only** | Agents do not use the web UI. They create annotations via a CLI tool. The web UI is read-mostly: it reads annotations and changes review states, but does not need full CRUD for all annotation entities. |
| **SQL files, not database** | Saved queries and presets are `.sql` files on the filesystem, not rows in a database table. |

## 8. What We're Asking You to Design

Design the interface for:

1. **Annotation review** — How does the operator see what agents have produced, understand it, and make review decisions efficiently? This is the core problem.

2. **Target browsing** — How does the operator navigate the annotated entities (senders, threads, messages, domains) and see their annotation state? How do detail views work?

3. **Groups and logs** — How does the operator interact with the groupings and narrative logs that agents create?

4. **Agent run inspection** — How does the operator see what a specific agent run did, judge its quality, and batch-accept or batch-reject its output?

5. **SQL query workbench** — How does the operator run ad-hoc queries, manage saved query files, and connect query results back to the annotation workflow?

6. **Navigation** — How do all these pieces fit together? How does the operator move between reviewing annotations, exploring data, and running queries?

## 9. Deliverables

We'd like to see:

- **Screen designs** (wireframes, mockups, or prototypes) for the key workflows
- **Navigation model** — how the pieces connect, what the primary and secondary flows are
- **Widget/component breakdown** — what are the reusable pieces
- **Interaction patterns** for batch operations, drill-down, and cross-linking

## 10. Helpful Context (Not Requirements)

These are observations that might inform your design, not constraints:

- The review queue is the most time-sensitive view — the operator wants to clear it quickly after each agent run
- Agents annotate senders more often than individual messages (senders are higher-leverage targets)
- The `agent_run_id` is the natural grouping for "what changed" — it bundles everything an agent did in one session
- Annotation logs are the agent's explanation of *why* it made the choices it did — surfacing these alongside the annotations they explain makes review much faster
- The operator sometimes wants to go from "this annotation looks wrong" → "show me the actual emails" → "oh, I see, the agent was right" — the drill-down path from annotation to underlying data matters
- SQL is a fallback for everything the structured views don't cover — it should feel like a natural part of the tool, not a hidden developer feature
