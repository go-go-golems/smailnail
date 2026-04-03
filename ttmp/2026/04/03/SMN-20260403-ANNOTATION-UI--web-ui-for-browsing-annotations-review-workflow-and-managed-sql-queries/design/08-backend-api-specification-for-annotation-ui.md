---
Title: "Backend API Specification for Annotation UI"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - backend
    - api
    - annotations
    - go
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/api/annotations.ts:RTK Query API slice — the frontend contract these endpoints must satisfy"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/types/annotations.ts:TypeScript response shapes — JSON must match these exactly"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/mocks/handlers.ts:MSW mock handlers — reference implementation of filtering and response assembly"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/repository.go:Existing annotation repository — has most CRUD methods already"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Go annotation types — already have correct JSON tags"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/schema.go:SQLite schema for annotations, target_groups, annotation_logs"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/enrich/senders.go:Sender enricher — source of senders table data"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/enrich/schema.go:Senders table schema (email, domain, msg_count, has_list_unsubscribe)"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/http.go:Existing HTTP handler — add new routes in registerAPIRoutes"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/serve/handlers_queries.go:go-minitrace query handler — reference for preset/saved query file pattern"
ExternalSources: []
Summary: "Complete backend API specification for the annotation review UI. Lists every endpoint the frontend calls, with exact URL paths, HTTP methods, query parameters, request/response JSON shapes, SQL queries, and implementation notes. Designed as a self-contained handoff document for a backend engineer."
LastUpdated: 2026-04-03T18:00:00.000000000-04:00
WhatFor: "Hand to a backend engineer to implement all API endpoints needed by the annotation review UI"
WhenToUse: ""
---

# Backend API Specification for Annotation Review UI

## 1. Context

The annotation review UI (React + MUI + RTK Query) is complete. It calls **16 API endpoints** under `/api/`. The frontend is built, tested in Storybook with MSW mock handlers, and ready to connect to real endpoints. This document specifies every endpoint the backend must implement.

### Existing infrastructure
- **HTTP handler**: `pkg/smailnaild/http.go` — add new routes in `registerAPIRoutes()`
- **Annotation repository**: `pkg/annotate/repository.go` — already has `CreateAnnotation`, `GetAnnotation`, `ListAnnotations`, `UpdateAnnotationReviewState`, `CreateGroup`, `GetGroup`, `ListGroups`, `AddGroupMember`, `ListGroupMembers`, `CreateLog`, `GetLog`, `ListLogs`, `LinkLogTarget`, `ListLogTargets`
- **Senders table**: `pkg/enrich/schema.go` — `senders(email, display_name, domain, msg_count, first_seen_date, last_seen_date, has_list_unsubscribe, ...)`
- **Messages table**: `pkg/mirror/schema.go` — `messages(uid, subject, date, size, sender_email, sender_domain, ...)`
- **Response convention**: existing endpoints use `writeDataJSON(w, status, data, meta)` which wraps in `{"data": ..., "meta": ...}`

### Important: JSON field naming
The Go types in `pkg/annotate/types.go` already have JSON tags using **camelCase** (e.g., `json:"targetType"`, `json:"reviewState"`). The frontend expects camelCase. **Do not change these tags.**

### Important: Response envelope
The existing API uses `writeDataJSON` which wraps responses in `{"data": ..., "meta": ...}`. However, the frontend's RTK Query layer expects **bare JSON arrays/objects** (no envelope). Choose one approach and be consistent:

**Option A (recommended):** New annotation endpoints return bare JSON (matching the MSW mocks). Write a `writeJSONDirect(w, status, payload)` helper.

**Option B:** Keep the envelope and adjust the RTK Query `baseQuery` to unwrap `.data`. This requires a one-line frontend change.

## 2. Endpoint Reference

### 2.1 Annotations

#### `GET /api/annotations`

List annotations with optional filters. All filters are AND-combined.

| Query Param | Type | Description |
|---|---|---|
| `targetType` | string | Filter by target type (`sender`, `domain`, `message`) |
| `targetId` | string | Filter by target ID |
| `tag` | string | Filter by tag (exact match) |
| `reviewState` | string | One of: `to_review`, `reviewed`, `dismissed` |
| `sourceKind` | string | One of: `human`, `agent`, `heuristic`, `import` |
| `agentRunId` | string | Filter by agent run ID |
| `limit` | int | Max results (default: 500) |

**Response:** `200 OK`

```json
[
  {
    "id": "ann-001",
    "targetType": "sender",
    "targetId": "news@techcrunch.com",
    "tag": "newsletter",
    "noteMarkdown": "Regular tech newsletter...",
    "sourceKind": "agent",
    "sourceLabel": "triage-agent-v2",
    "agentRunId": "run-42",
    "reviewState": "to_review",
    "createdBy": "system",
    "createdAt": "2026-04-01T10:30:00Z",
    "updatedAt": "2026-04-01T10:30:00Z"
  }
]
```

**Implementation:** Use `repository.ListAnnotations()`. Extend `ListAnnotationsFilter` to add `AgentRunID` field (currently missing — add it to `types.go` and the SQL builder in `repository.go`).

**SQL sketch:**
```sql
SELECT id, target_type, target_id, tag, note_markdown, source_kind,
       source_label, agent_run_id, review_state, created_by, created_at, updated_at
FROM annotations
WHERE 1=1
  AND (? = '' OR target_type = ?)
  AND (? = '' OR target_id = ?)
  AND (? = '' OR tag = ?)
  AND (? = '' OR review_state = ?)
  AND (? = '' OR source_kind = ?)
  AND (? = '' OR agent_run_id = ?)
ORDER BY created_at DESC
LIMIT ?
```

---

#### `GET /api/annotations/{id}`

Get a single annotation by ID.

**Response:** `200 OK` — single `Annotation` object (same shape as list item)

**404:** `{"error": "not-found", "message": "..."}`

**Implementation:** Use `repository.GetAnnotation(ctx, id)`.

---

#### `PATCH /api/annotations/{id}/review`

Update an annotation's review state.

**Request body:**
```json
{
  "reviewState": "reviewed"
}
```

**Response:** `200 OK` — the updated `Annotation` object

**Implementation:** Use `repository.UpdateAnnotationReviewState(ctx, id, reviewState)`. Validate that `reviewState` is one of `to_review`, `reviewed`, `dismissed`.

---

#### `POST /api/annotations/batch-review`

Batch-update review state for multiple annotations.

**Request body:**
```json
{
  "ids": ["ann-001", "ann-002", "ann-003"],
  "reviewState": "reviewed"
}
```

**Response:** `204 No Content`

**Implementation:** **New method needed** — `repository.BatchUpdateReviewState(ctx, ids []string, reviewState string)`. Execute in a transaction:

```sql
UPDATE annotations
SET review_state = ?, updated_at = CURRENT_TIMESTAMP
WHERE id IN (?, ?, ...)
```

Use `sqlx.In()` to expand the ID list. Validate `reviewState`. Return 400 if `ids` is empty.

---

### 2.2 Target Groups

#### `GET /api/annotation-groups`

List target groups with optional filters.

| Query Param | Type | Description |
|---|---|---|
| `reviewState` | string | Filter by review state |
| `sourceKind` | string | Filter by source kind |
| `limit` | int | Max results (default: 100) |

**Response:** `200 OK` — array of `TargetGroup` objects

**Implementation:** Use `repository.ListGroups()`.

---

#### `GET /api/annotation-groups/{id}`

Get a group with its members.

**Response:** `200 OK`

```json
{
  "id": "grp-001",
  "name": "Tech Newsletter Senders",
  "description": "...",
  "sourceKind": "agent",
  "sourceLabel": "triage-agent-v2",
  "agentRunId": "run-42",
  "reviewState": "to_review",
  "createdBy": "system",
  "createdAt": "2026-04-01T10:35:00Z",
  "updatedAt": "2026-04-01T10:35:00Z",
  "members": [
    {
      "groupId": "grp-001",
      "targetType": "sender",
      "targetId": "news@techcrunch.com",
      "addedAt": "2026-04-01T10:35:00Z"
    }
  ]
}
```

**Implementation:** Call `repository.GetGroup(ctx, id)` then `repository.ListGroupMembers(ctx, id)`. Compose the `GroupDetail` struct:

```go
type GroupDetail struct {
    TargetGroup
    Members []GroupMember `json:"members"`
}
```

---

### 2.3 Annotation Logs

#### `GET /api/annotation-logs`

| Query Param | Type | Description |
|---|---|---|
| `sourceKind` | string | Filter by source kind |
| `agentRunId` | string | Filter by agent run ID |
| `limit` | int | Max results (default: 200) |

**Response:** `200 OK` — array of `AnnotationLog` objects

**Implementation:** Use `repository.ListLogs()`. The existing `ListLogsFilter` already supports `AgentRunID`.

---

#### `GET /api/annotation-logs/{id}`

**Response:** `200 OK` — single `AnnotationLog` object

**Implementation:** Use `repository.GetLog(ctx, id)`.

---

### 2.4 Agent Runs (Aggregated)

These endpoints don't map to a single table. "Runs" are **aggregations** of annotations/logs/groups by `agent_run_id`.

#### `GET /api/annotation-runs`

List all distinct agent runs with aggregated counts.

**Response:** `200 OK`

```json
[
  {
    "runId": "run-42",
    "sourceLabel": "triage-agent-v2",
    "sourceKind": "agent",
    "annotationCount": 23,
    "pendingCount": 18,
    "reviewedCount": 3,
    "dismissedCount": 2,
    "logCount": 4,
    "groupCount": 2,
    "startedAt": "2026-04-01T10:29:00Z",
    "completedAt": "2026-04-01T10:40:00Z"
  }
]
```

**Implementation:** **New query** — aggregate across annotations, logs, and groups:

```sql
WITH run_annotations AS (
    SELECT
        agent_run_id,
        source_label,
        source_kind,
        COUNT(*) AS annotation_count,
        SUM(CASE WHEN review_state = 'to_review' THEN 1 ELSE 0 END) AS pending_count,
        SUM(CASE WHEN review_state = 'reviewed' THEN 1 ELSE 0 END) AS reviewed_count,
        SUM(CASE WHEN review_state = 'dismissed' THEN 1 ELSE 0 END) AS dismissed_count,
        MIN(created_at) AS started_at,
        MAX(created_at) AS completed_at
    FROM annotations
    WHERE agent_run_id != ''
    GROUP BY agent_run_id, source_label, source_kind
),
run_logs AS (
    SELECT agent_run_id, COUNT(*) AS log_count
    FROM annotation_logs
    WHERE agent_run_id != ''
    GROUP BY agent_run_id
),
run_groups AS (
    SELECT agent_run_id, COUNT(*) AS group_count
    FROM target_groups
    WHERE agent_run_id != ''
    GROUP BY agent_run_id
)
SELECT
    ra.agent_run_id AS run_id,
    ra.source_label,
    ra.source_kind,
    ra.annotation_count,
    ra.pending_count,
    ra.reviewed_count,
    ra.dismissed_count,
    COALESCE(rl.log_count, 0) AS log_count,
    COALESCE(rg.group_count, 0) AS group_count,
    ra.started_at,
    ra.completed_at
FROM run_annotations ra
LEFT JOIN run_logs rl ON rl.agent_run_id = ra.agent_run_id
LEFT JOIN run_groups rg ON rg.agent_run_id = ra.agent_run_id
ORDER BY ra.started_at DESC
```

Add a new method `repository.ListRuns(ctx) ([]AgentRunSummary, error)`.

Define `AgentRunSummary` in Go:

```go
type AgentRunSummary struct {
    RunID           string `db:"run_id" json:"runId"`
    SourceLabel     string `db:"source_label" json:"sourceLabel"`
    SourceKind      string `db:"source_kind" json:"sourceKind"`
    AnnotationCount int    `db:"annotation_count" json:"annotationCount"`
    PendingCount    int    `db:"pending_count" json:"pendingCount"`
    ReviewedCount   int    `db:"reviewed_count" json:"reviewedCount"`
    DismissedCount  int    `db:"dismissed_count" json:"dismissedCount"`
    LogCount        int    `db:"log_count" json:"logCount"`
    GroupCount      int    `db:"group_count" json:"groupCount"`
    StartedAt       string `db:"started_at" json:"startedAt"`
    CompletedAt     string `db:"completed_at" json:"completedAt"`
}
```

---

#### `GET /api/annotation-runs/{id}`

Get a single run with its annotations, logs, and groups.

**Response:** `200 OK`

```json
{
  "runId": "run-42",
  "sourceLabel": "triage-agent-v2",
  "sourceKind": "agent",
  "annotationCount": 23,
  "pendingCount": 18,
  "reviewedCount": 3,
  "dismissedCount": 2,
  "logCount": 4,
  "groupCount": 2,
  "startedAt": "2026-04-01T10:29:00Z",
  "completedAt": "2026-04-01T10:40:00Z",
  "annotations": [ ... ],
  "logs": [ ... ],
  "groups": [ ... ]
}
```

**Implementation:** Compose from existing queries:
1. `ListAnnotations(ctx, {AgentRunID: id})`
2. `ListLogs(ctx, {AgentRunID: id})`
3. Groups: `SELECT * FROM target_groups WHERE agent_run_id = ?`
4. Compute summary counts from the annotations list
5. Return `AgentRunDetail` struct

---

### 2.5 Senders (from mirror/enrich)

#### `GET /api/mirror/senders`

List senders with annotation data.

| Query Param | Type | Description |
|---|---|---|
| `domain` | string | Filter by domain |
| `hasAnnotations` | bool | Only senders with annotations |
| `tag` | string | Only senders with this tag in annotations |
| `limit` | int | Max results (default: 200) |

**Response:** `200 OK`

```json
[
  {
    "email": "news@techcrunch.com",
    "displayName": "TechCrunch Daily",
    "domain": "techcrunch.com",
    "messageCount": 47,
    "annotationCount": 1,
    "tags": ["newsletter"],
    "hasUnsubscribe": true
  }
]
```

**Implementation:** **New query** joining `senders` and `annotations`:

```sql
SELECT
    s.email,
    s.display_name,
    s.domain,
    s.msg_count AS message_count,
    COUNT(DISTINCT a.id) AS annotation_count,
    s.has_list_unsubscribe AS has_unsubscribe
FROM senders s
LEFT JOIN annotations a ON a.target_type = 'sender' AND a.target_id = s.email
WHERE 1=1
  AND (? = '' OR s.domain = ?)
  AND (? = '' OR a.tag = ?)
GROUP BY s.email
HAVING (? = FALSE OR annotation_count > 0)
ORDER BY s.msg_count DESC
LIMIT ?
```

For the `tags` array: run a second query or use `GROUP_CONCAT`:

```sql
SELECT DISTINCT a.tag
FROM annotations a
WHERE a.target_type = 'sender' AND a.target_id = ?
```

Or use `GROUP_CONCAT(DISTINCT a.tag)` in the main query and split in Go.

Define the Go response type:

```go
type SenderRow struct {
    Email           string   `json:"email"`
    DisplayName     string   `json:"displayName"`
    Domain          string   `json:"domain"`
    MessageCount    int      `json:"messageCount"`
    AnnotationCount int      `json:"annotationCount"`
    Tags            []string `json:"tags"`
    HasUnsubscribe  bool     `json:"hasUnsubscribe"`
}
```

---

#### `GET /api/mirror/senders/{email}`

Get a single sender with annotations, related logs, and recent messages.

**Response:** `200 OK`

```json
{
  "email": "news@techcrunch.com",
  "displayName": "TechCrunch Daily",
  "domain": "techcrunch.com",
  "messageCount": 47,
  "annotationCount": 1,
  "tags": ["newsletter"],
  "hasUnsubscribe": true,
  "firstSeen": "2025-01-15T00:00:00Z",
  "lastSeen": "2026-04-01T08:00:00Z",
  "annotations": [ ... ],
  "logs": [ ... ],
  "recentMessages": [
    {
      "uid": 1001,
      "subject": "TechCrunch Daily - April 1, 2026",
      "date": "2026-04-01T08:00:00Z",
      "sizeBytes": 45320
    }
  ]
}
```

**Implementation:** Compose from:
1. `SELECT * FROM senders WHERE email = ?`
2. `ListAnnotations(ctx, {TargetType: "sender", TargetID: email})`
3. Find related logs: `SELECT * FROM annotation_logs WHERE agent_run_id IN (SELECT DISTINCT agent_run_id FROM annotations WHERE target_type = 'sender' AND target_id = ?)`
4. Recent messages: `SELECT uid, subject, date, size FROM messages WHERE sender_email = ? ORDER BY date DESC LIMIT 20`

The `{email}` path parameter is URL-encoded by the frontend (`encodeURIComponent`). Use `r.PathValue("email")` which Go's net/http already decodes.

---

### 2.6 Query Editor

#### `POST /api/query/execute`

Execute an arbitrary SQL query against the SQLite database (read-only).

**Request body:**
```json
{
  "sql": "SELECT tag, COUNT(*) as count FROM annotations GROUP BY tag ORDER BY count DESC"
}
```

**Response:** `200 OK`

```json
{
  "columns": ["tag", "count"],
  "rows": [
    {"tag": "newsletter", "count": 8},
    {"tag": "notification", "count": 6}
  ],
  "durationMs": 12,
  "rowCount": 5
}
```

**Error response:** `400 Bad Request`

```json
{
  "message": "Error: no such column: foo"
}
```

**Implementation:** Follow go-minitrace's pattern (`handlers_queries.go`):
1. Open a read-only connection or use `BEGIN READONLY` transaction
2. Execute with `sqlx.QueryxContext`
3. Scan rows dynamically into `[]map[string]any`
4. Capture column names from `rows.Columns()`
5. Measure duration
6. **CRITICAL: Enforce read-only** — reject queries starting with `INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `CREATE`. Or use a read-only SQLite connection (`?mode=ro`).

---

#### `GET /api/query/presets`

List preset SQL queries from `go:embed` assets or `--preset-dir`.

**Response:** `200 OK`

```json
[
  {
    "name": "annotations-by-tag",
    "folder": "annotations",
    "description": "Count annotations grouped by tag",
    "sql": "SELECT tag, COUNT(*) as count..."
  }
]
```

**Implementation:** Follow go-minitrace's `pkg/query/` pattern:
1. Embed `.sql` files in a Go `embed.FS` with `//go:embed queries/*.sql`
2. Parse frontmatter from each file (name, folder, description in YAML/comment header)
3. Allow `--preset-dir` flag for additional directories
4. Return merged list

Store preset files in `queries/` directory:
```
queries/
  annotations/
    by-tag.sql
    pending-review.sql
  mirror/
    sender-volume.sql
    top-domains.sql
```

---

#### `GET /api/query/saved`

List user-saved queries from `--query-dir` filesystem directory.

**Response:** `200 OK` — array of `SavedQuery` (same shape as presets)

**Implementation:** Scan the `--query-dir` directory for `.sql` files. Same format as presets.

---

#### `POST /api/query/saved`

Save a new query to the `--query-dir` filesystem directory.

**Request body:**
```json
{
  "name": "my-senders",
  "folder": "custom",
  "description": "Custom sender analysis",
  "sql": "SELECT ..."
}
```

**Response:** `201 Created` — the saved `SavedQuery` object

**Implementation:** Write to `{query-dir}/{folder}/{name}.sql` with a comment header containing the description. Sanitize `name` and `folder` (no `..`, no absolute paths).

---

## 3. Route Registration

Add to `registerAPIRoutes()` in `pkg/smailnaild/http.go`:

```go
func (h *appHandler) registerAPIRoutes(mux *http.ServeMux) {
    // ... existing account/rule routes ...

    // Annotations
    mux.HandleFunc("GET /api/annotations", h.handleListAnnotations)
    mux.HandleFunc("GET /api/annotations/{id}", h.handleGetAnnotation)
    mux.HandleFunc("PATCH /api/annotations/{id}/review", h.handleReviewAnnotation)
    mux.HandleFunc("POST /api/annotations/batch-review", h.handleBatchReview)

    // Groups
    mux.HandleFunc("GET /api/annotation-groups", h.handleListGroups)
    mux.HandleFunc("GET /api/annotation-groups/{id}", h.handleGetGroup)

    // Logs
    mux.HandleFunc("GET /api/annotation-logs", h.handleListLogs)
    mux.HandleFunc("GET /api/annotation-logs/{id}", h.handleGetLog)

    // Runs (aggregated)
    mux.HandleFunc("GET /api/annotation-runs", h.handleListRuns)
    mux.HandleFunc("GET /api/annotation-runs/{id}", h.handleGetRun)

    // Senders
    mux.HandleFunc("GET /api/mirror/senders", h.handleListSenders)
    mux.HandleFunc("GET /api/mirror/senders/{email}", h.handleGetSender)

    // Query editor
    mux.HandleFunc("POST /api/query/execute", h.handleExecuteQuery)
    mux.HandleFunc("GET /api/query/presets", h.handleGetPresets)
    mux.HandleFunc("GET /api/query/saved", h.handleGetSavedQueries)
    mux.HandleFunc("POST /api/query/saved", h.handleSaveQuery)
}
```

## 4. New Go Types Needed

Add to `pkg/annotate/types.go`:

```go
// AgentRunSummary is an aggregated view — not stored directly.
type AgentRunSummary struct {
    RunID           string `db:"run_id" json:"runId"`
    SourceLabel     string `db:"source_label" json:"sourceLabel"`
    SourceKind      string `db:"source_kind" json:"sourceKind"`
    AnnotationCount int    `db:"annotation_count" json:"annotationCount"`
    PendingCount    int    `db:"pending_count" json:"pendingCount"`
    ReviewedCount   int    `db:"reviewed_count" json:"reviewedCount"`
    DismissedCount  int    `db:"dismissed_count" json:"dismissedCount"`
    LogCount        int    `db:"log_count" json:"logCount"`
    GroupCount      int    `db:"group_count" json:"groupCount"`
    StartedAt       string `db:"started_at" json:"startedAt"`
    CompletedAt     string `db:"completed_at" json:"completedAt"`
}

type AgentRunDetail struct {
    AgentRunSummary
    Annotations []Annotation   `json:"annotations"`
    Logs        []AnnotationLog `json:"logs"`
    Groups      []TargetGroup   `json:"groups"`
}

type GroupDetail struct {
    TargetGroup
    Members []GroupMember `json:"members"`
}
```

Add to `pkg/annotate/types.go` — update `ListAnnotationsFilter`:

```go
type ListAnnotationsFilter struct {
    TargetType  string
    TargetID    string
    Tag         string
    ReviewState string
    SourceKind  string
    AgentRunID  string  // ← ADD THIS
    Limit       int
}
```

Create new types in `pkg/smailnaild/` or a new `pkg/browse/` package:

```go
type SenderRow struct {
    Email           string   `json:"email"`
    DisplayName     string   `json:"displayName"`
    Domain          string   `json:"domain"`
    MessageCount    int      `json:"messageCount"`
    AnnotationCount int      `json:"annotationCount"`
    Tags            []string `json:"tags"`
    HasUnsubscribe  bool     `json:"hasUnsubscribe"`
}

type SenderDetail struct {
    SenderRow
    FirstSeen      string           `json:"firstSeen"`
    LastSeen        string           `json:"lastSeen"`
    Annotations    []annotate.Annotation   `json:"annotations"`
    Logs           []annotate.AnnotationLog `json:"logs"`
    RecentMessages []MessagePreview  `json:"recentMessages"`
}

type MessagePreview struct {
    UID       uint32 `json:"uid"`
    Subject   string `json:"subject"`
    Date      string `json:"date"`
    SizeBytes int    `json:"sizeBytes"`
}
```

## 5. New Repository Methods Needed

| Method | Signature | Notes |
|---|---|---|
| `BatchUpdateReviewState` | `(ctx, ids []string, reviewState string) error` | Use `sqlx.In()` for the ID list |
| `ListRuns` | `(ctx) ([]AgentRunSummary, error)` | Aggregation query from §2.4 |
| `GetRunDetail` | `(ctx, runID string) (*AgentRunDetail, error)` | Compose from ListAnnotations + ListLogs + groups query |

## 6. Dependencies / New Packages

| Package | Purpose |
|---|---|
| `pkg/annotate/repository.go` | Add `BatchUpdateReviewState`, `ListRuns`, `GetRunDetail` |
| `pkg/smailnaild/handlers_annotations.go` | New file — annotation/group/log/run HTTP handlers |
| `pkg/smailnaild/handlers_senders.go` | New file — sender list/detail HTTP handlers |
| `pkg/smailnaild/handlers_query.go` | New file — query execute/presets/saved handlers |
| `queries/` | New directory — embedded SQL presets |

## 7. Implementation Order

1. **`ListAnnotationsFilter.AgentRunID`** — add field to types.go, update SQL builder in repository.go
2. **`BatchUpdateReviewState`** — new repository method
3. **`ListRuns` / `GetRunDetail`** — new repository methods with aggregation query
4. **`handlers_annotations.go`** — 4 annotation endpoints + 2 group endpoints + 2 log endpoints + 2 run endpoints
5. **`handlers_senders.go`** — 2 sender endpoints (cross-DB join between senders and annotations)
6. **`handlers_query.go`** — 4 query endpoints (follow go-minitrace pattern)
7. **Route registration** in `http.go`
8. **Preset SQL files** in `queries/` directory
9. **Integration test** — start server, hit each endpoint, verify response shapes match TypeScript types

## 8. Testing Strategy

### Verify against MSW mock handlers
The MSW handlers in `ui/src/mocks/handlers.ts` are the reference implementation. Each Go handler should produce output matching the MSW handler for the same input. The mock handlers show:
- What query parameters are supported
- How filtering works
- What the response shape looks like
- How composed endpoints (runs, sender detail) assemble their responses

### Quick smoke test
```bash
# Start the dev server
go run ./cmd/smailnail serve --dev

# Hit each endpoint
curl -s localhost:8080/api/annotations | jq '.[:2]'
curl -s localhost:8080/api/annotations?tag=newsletter | jq 'length'
curl -s localhost:8080/api/annotation-runs | jq '.[0]'
curl -s localhost:8080/api/annotation-runs/run-42 | jq '.annotations | length'
curl -s localhost:8080/api/mirror/senders | jq '.[0].tags'
curl -s "localhost:8080/api/mirror/senders/news@techcrunch.com" | jq '.recentMessages | length'
curl -s -X POST localhost:8080/api/query/execute -d '{"sql":"SELECT 1+1 as result"}' | jq '.'
```

### Frontend integration test
```bash
# In one terminal:
go run ./cmd/smailnail serve --dev

# In another:
cd ui && pnpm dev

# Navigate to http://localhost:5050/annotations
# Verify: Dashboard shows stats, Review Queue shows annotations, filter pills work,
# batch approve works, Agent Runs shows progress bars, Sender Detail loads, SQL Workbench runs queries
```
