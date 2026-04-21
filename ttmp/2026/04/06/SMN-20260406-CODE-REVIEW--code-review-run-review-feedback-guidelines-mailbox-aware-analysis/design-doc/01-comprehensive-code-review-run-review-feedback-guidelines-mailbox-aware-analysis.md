---
Title: "Comprehensive Code Review: Run Review Feedback, Guidelines & Mailbox-aware Analysis"
Ticket: SMN-20260406-CODE-REVIEW
Status: active
Topics:
    - code-review
    - annotations
    - backend
    - frontend
    - sqlite
    - react
    - workflow
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotate/repository_feedback.go
      Note: Primary backend repository — feedback/guidelines CRUD and transactional review operations
    - Path: pkg/annotate/types.go
      Note: All Go domain types for feedback, guidelines, and review actions
    - Path: pkg/annotate/schema.go
      Note: SQLite schema migrations V3+V4
    - Path: pkg/annotationui/handlers_annotations.go
      Note: Review handlers with extended comment/guideline support
    - Path: pkg/annotationui/handlers_feedback.go
      Note: Standalone feedback/guideline/run-guideline HTTP handlers
    - Path: pkg/annotationui/types_feedback.go
      Note: HTTP request/response types and conversion helpers
    - Path: pkg/annotationui/server.go
      Note: Route registration and HTTP server wiring
    - Path: ui/src/api/annotations.ts
      Note: RTK Query contract with all new endpoints
    - Path: ui/src/types/reviewFeedback.ts
      Note: TypeScript feedback domain types
    - Path: ui/src/types/reviewGuideline.ts
      Note: TypeScript guideline domain types
    - Path: ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx
      Note: Shared feedback entry dialog
    - Path: ui/src/components/AnnotationTable/AnnotationTable.tsx
      Note: Memoized table with per-row dismiss-explain
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Batch reject-explain integration
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Run-level guideline and feedback sections
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Per-row dismiss-explain on sender detail
Summary: "Detailed code review of the Run Review Feedback, Guidelines, and Mailbox-aware Analysis feature branch (SMN-20260403-RUN-REVIEW) against origin/main. Written for a new intern to understand the entire system — what it does, how it is structured, where the code lives, what is clean, what is confusing, what is deprecated, what is unused, and what should change."
LastUpdated: 2026-04-06T14:00:00-04:00
WhatFor: "Give this document to a new intern before they touch any code in this feature branch. It explains the system, reviews every layer, and flags issues."
WhenToUse: "Use when onboarding onto the annotation UI codebase, reviewing the RUN-REVIEW branch, or planning follow-up work."
---

# Comprehensive Code Review: Run Review Feedback, Guidelines & Mailbox-aware Analysis

## 1. Executive Summary

This document is a detailed code review of the `SMN-20260403-RUN-REVIEW` feature branch in the **smailnail** project. The branch adds three major capabilities to the existing annotation review UI:

1. **Review Feedback** — Reviewers can now reject or dismiss annotations with structured explanations (comments, reject requests, clarification requests). Feedback records live in their own SQLite tables and are linked to annotation targets.

2. **Review Guidelines** — Reusable, editable policy documents stored in SQLite. Guidelines can be linked to agent runs, providing context for what rules were active during a triage session.

3. **Mailbox-aware Context** — The IMAP mailbox name (`INBOX`, `Sent`, `Billing`, etc.) is now surfaced in message previews, sender detail pages, and feedback payloads, so reviewers can reason about which mailbox a message came from.

The diff spans **80 files** with approximately **15,879 lines added** and **793 lines removed**. It touches four architectural layers:

- **Schema** (SQLite DDL in `pkg/annotate/schema.go`)
- **Repository** (Go CRUD + transactional operations in `pkg/annotate/repository_feedback.go`)
- **HTTP Handlers** (Go JSON API in `pkg/annotationui/handlers_*.go`)
- **Frontend** (React + TypeScript + RTK Query + MUI in `ui/src/`)

### High-level verdict

The implementation is **structurally sound** and follows the existing codebase patterns consistently. The most impressive aspect is the transactional repository layer — review-state updates, feedback creation, and guideline linking happen inside a single SQL transaction, which was a specific recovery from an earlier broken attempt.

However, there are several categories of issues that this review will flag:

- **Dead code and unused state** — Redux slice fields that no component reads
- **Naming inconsistencies** — `ReviewCommentDrawer` is actually a Dialog, not a Drawer
- **Deprecated / half-wired features** — `ReviewCommentInline` exists but is never imported
- **Missing backend validation** — `UpdateReviewFeedback` only updates `status` but the type suggests more
- **Mock-only features** — `GuidelineLinkedRuns` always renders an empty list
- **Duplicated SQL patterns** — feedback creation is duplicated inside and outside transactions
- **Enrich command regression** — the `enrichSettings` embedding was flattened, changing the struct layout
- **Performance concern** — `ListReviewFeedback` does N+1 target queries

Each of these will be explained in detail with file references, pseudocode, and recommended fixes.

## 2. What Is This System? (For the New Intern)

### 2.1 The Big Picture

**smailnail** is an email triage tool. It does three things:

1. **Mirror** — Connects to an IMAP server, downloads messages into a local SQLite database, organized by mailbox (INBOX, Sent, Archive, etc.).
2. **Analyze** — An LLM agent examines the mirrored messages and produces **annotations** — structured claims like "this sender is a newsletter" or "this thread is promotional."
3. **Review** — A human reviewer browses the annotations in a web UI, approves or dismisses them, and provides feedback to improve future agent runs.

The code you are reviewing adds a major upgrade to step 3. Previously, the reviewer could only approve or dismiss annotations — a binary toggle with no explanation. Now the reviewer can:

- Leave **structured feedback** explaining *why* something was dismissed
- Create **reusable guidelines** that the agent should follow in future runs
- See **which mailbox** a message came from, which matters because the same sender can behave differently in INBOX vs. Archive

### 2.2 Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        smailnail system                         │
│                                                                 │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │  IMAP    │───▶│  SQLite DB   │◀───│  annotation-ui       │   │
│  │  Mirror  │    │              │    │  (Go HTTP + React)    │   │
│  └──────────┘    │  messages    │    │                       │   │
│                  │  senders     │    │  ┌─────────────────┐  │   │
│  ┌──────────┐    │  annotations │    │  │  React SPA      │  │   │
│  │  LLM     │───▶│  logs        │    │  │  (RTK Query)    │  │   │
│  │  Agent   │    │  groups      │    │  │                 │  │   │
│  └──────────┘    │              │    │  │  Review Queue   │  │   │
│                  │  NEW TABLES: │    │  │  Run Detail     │  │   │
│                  │  feedback    │    │  │  Guidelines     │  │   │
│                  │  guidelines  │    │  │  Sender Detail  │  │   │
│                  │  run_links   │    │  └─────────────────┘  │   │
│                  └──────────────┘    └──────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 The Data Flow (How a Review Happens Now)

Here is the complete flow when a reviewer dismisses an annotation with feedback:

```
1. Reviewer clicks "Dismiss & Explain" on an annotation row
   │
   ▼
2. React opens ReviewCommentDrawer (a Dialog modal)
   │  User fills in: feedbackKind, title, body, optional guidelineIds
   │
   ▼
3. React calls RTK Query mutation: reviewAnnotation({ id, reviewState: "dismissed", comment: {...}, guidelineIds: [...] })
   │
   ▼
4. RTK Query sends PATCH /api/annotations/{id}/review
   │  Body: { reviewState: "dismissed", comment: { feedbackKind, title, bodyMarkdown }, guidelineIds: [...] }
   │
   ▼
5. Go handler: handleReviewAnnotation() in handlers_annotations.go
   │  Parses JSON, calls annotations.ReviewAnnotationWithArtifacts()
   │
   ▼
6. Repository: ReviewAnnotationWithArtifacts() in repository_feedback.go
   │  Begins SQL transaction
   │  ├── Updates annotations.review_state = "dismissed"
   │  ├── If comment provided: INSERT INTO review_feedback + review_feedback_targets
   │  ├── For each guidelineId: INSERT INTO run_guideline_links
   │  └── Commits (or rolls back on ANY error)
   │
   ▼
7. Response returns the updated annotation
   │
   ▼
8. RTK Query invalidates "Annotations", "Runs", "Feedback" cache tags
   │  → All list views automatically refetch
   │
   ▼
9. UI updates: annotation row shows "dismissed", RunFeedbackSection shows new feedback card
```

This is the single most important flow in the entire feature. The key design decision is that steps 6a–6c are **transactional** — they all succeed or all fail together.

### 2.4 Key Files Map

If you are new to this codebase, here is where everything lives, organized by layer:

```
smailnail/
├── pkg/
│   ├── annotate/                          # DOMAIN LAYER (pure Go + SQLite)
│   │   ├── schema.go                      #   SQLite DDL (table definitions)
│   │   ├── types.go                       #   Go structs for all domain objects
│   │   ├── repository.go                  #   Original annotation CRUD
│   │   └── repository_feedback.go         #   ★ NEW: feedback/guideline CRUD + transactions
│   │
│   ├── annotationui/                      # HTTP LAYER (Go net/http + JSON)
│   │   ├── server.go                      #   Route registration, HTTP helpers
│   │   ├── handlers_annotations.go        #   ★ MODIFIED: review handlers now call transactional repo
│   │   ├── handlers_feedback.go           #   ★ NEW: standalone feedback/guideline endpoints
│   │   ├── handlers_senders.go            #   ★ MODIFIED: added mailbox_name to query
│   │   ├── types.go                       #   HTTP DTOs (MessagePreview, SenderDetail)
│   │   └── types_feedback.go              #   ★ NEW: request/response types + conversion helpers
│   │
│   └── mirror/                            # IMAP mirror (pre-existing, not changed)
│       ├── schema.go                      #   Contains messages.mailbox_name column
│       └── service.go                     #   Populates mailbox_name during sync
│
├── ui/                                    # FRONTEND LAYER (React + TypeScript)
│   └── src/
│       ├── api/
│       │   └── annotations.ts             #   ★ MODIFIED: 12 new RTK Query endpoints
│       ├── types/
│       │   ├── annotations.ts             #   ★ MODIFIED: added mailboxName to MessagePreview
│       │   ├── reviewFeedback.ts          #   ★ NEW: ReviewFeedback, FeedbackKind, etc.
│       │   └── reviewGuideline.ts         #   ★ NEW: ReviewGuideline, GuidelineScope, etc.
│       ├── store/
│       │   └── annotationUiSlice.ts       #   ★ MODIFIED: new Redux state fields
│       ├── components/
│       │   ├── shared/                    #   ★ NEW badges: MailboxBadge, FeedbackKindBadge, etc.
│       │   ├── ReviewFeedback/            #   ★ NEW: drawer/dialog, FeedbackCard, GuidelinePicker, etc.
│       │   ├── RunGuideline/              #   ★ NEW: GuidelineCard, RunGuidelineSection
│       │   ├── Guidelines/                #   ★ NEW: GuidelineForm, SummaryCard, LinkedRuns
│       │   └── AnnotationTable/           #   ★ MODIFIED: memoization + onDismissExplain
│       ├── pages/
│       │   ├── ReviewQueuePage.tsx        #   ★ MODIFIED: batch reject-explain
│       │   ├── RunDetailPage.tsx          #   ★ MODIFIED: guideline section + feedback section
│       │   ├── SenderDetailPage.tsx       #   ★ MODIFIED: per-row dismiss-explain
│       │   ├── GuidelinesListPage.tsx     #   ★ NEW
│       │   └── GuidelineDetailPage.tsx    #   ★ NEW
│       └── mocks/
│           ├── annotations.ts             #   ★ NEW: mock feedback + guidelines data
│           └── handlers.ts                #   ★ NEW: MSW handlers for all new endpoints
│
├── cmd/smailnail/commands/enrich/         # ★ MODIFIED: struct flattening (side effect)
│   ├── all.go
│   ├── senders.go
│   ├── threads.go
│   └── unsubscribe.go
│
└── ttmp/2026/04/03/SMN-20260403-RUN-REVIEW--.../   # Design docs + diary (not code)
```

### 2.4 Data Flow — From Click to SQLite

When a reviewer clicks "Dismiss & Explain" on an annotation, here is exactly what happens:

```
1. User clicks dismiss icon in AnnotationRow
      ↓
2. AnnotationRow calls onDismissExplain(annotation.id)
      ↓
3. Page component sets commentAnnotation state
      ↓
4. ReviewCommentDrawer (a MUI Dialog) opens
      ↓
5. User fills: feedbackKind, title, bodyMarkdown, picks guidelineIds
      ↓
6. User clicks "Dismiss & Explain"
      ↓
7. ReviewCommentDrawer calls onSubmit(payload)
      ↓
8. Page component calls reviewAnnotation mutation:
      RTK Query → PATCH /api/annotations/:id/review
      Body: { reviewState: "dismissed", comment: {...}, guidelineIds: [...] }
      ↓
9. Go handler: handleReviewAnnotation (handlers_annotations.go)
      Decodes JSON into reviewRequest struct
      Validates reviewState
      Calls annotations.ReviewAnnotationWithArtifacts(...)
      ↓
10. Go repository: ReviewAnnotationWithArtifacts (repository_feedback.go)
      Begins SQL transaction
      UPDATE annotations SET review_state = 'dismissed' WHERE id = ?
      SELECT * FROM annotations WHERE id = ? (to get agent_run_id)
      INSERT INTO review_feedback (...) VALUES (...)  (if comment present)
      INSERT INTO review_feedback_targets (...) VALUES (...)
      INSERT INTO run_guideline_links (...) VALUES (...) (for each guidelineId)
      COMMIT
      ↓
11. Response: updated Annotation JSON
      ↓
12. RTK Query invalidates "Annotations", "Runs", "Feedback" cache tags
      ↓
13. UI re-renders with new data
```

### 2.5 Key Files to Read First

Read these in order before touching any code:

| # | File | Why |
|---|------|-----|
| 1 | `pkg/annotate/schema.go` | Understand the tables first |
| 2 | `pkg/annotate/types.go` | Understand the domain objects |
| 3 | `pkg/annotate/repository_feedback.go` | Understand the SQL operations |
| 4 | `pkg/annotationui/server.go` | See how routes are wired |
| 5 | `pkg/annotationui/handlers_annotations.go` | See the review handlers |
| 6 | `pkg/annotationui/handlers_feedback.go` | See the feedback/guideline handlers |
| 7 | `pkg/annotationui/types_feedback.go` | See the HTTP DTO layer |
| 8 | `ui/src/types/reviewFeedback.ts` | TypeScript mirror of Go types |
| 9 | `ui/src/types/reviewGuideline.ts` | TypeScript mirror of Go types |
| 10 | `ui/src/api/annotations.ts` | RTK Query contract |
| 11 | `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx` | The main feedback entry UI |
| 12 | `ui/src/components/AnnotationTable/AnnotationTable.tsx` | The table with memoization |

## 3. Backend Code Review

### 3.1 Schema Layer — `pkg/annotate/schema.go`

**What it does:** Defines SQLite table creation statements. The new V4 migration adds four tables:

- `review_feedback` — Stores reviewer-authored feedback records
- `review_feedback_targets` — Links feedback to annotation targets
- `review_guidelines` — Stores reusable review policy documents
- `run_guideline_links` — Join table linking guidelines to agent runs

**What is clean:**

- The `CREATE TABLE IF NOT EXISTS` pattern is consistent with existing V3 migration
- Indexes cover the primary query patterns: `(agent_run_id, created_at)`, `(status, created_at)`, `(status, priority DESC)`
- `ON CONFLICT` handling in `run_guideline_links` uses `DO NOTHING` (idempotent upsert)

**Issues found:**

1. **No migration versioning or rollback mechanism.** The schema is applied as raw SQL statements returned by a function. There is no migration version tracking table (e.g., `schema_migrations`). If `SchemaMigrationV4Statements()` is applied twice, the `IF NOT EXISTS` guards prevent errors, but there is no way to know which migrations have already been applied. This is fragile for production use.

   ```sql
   -- Recommended: Add a migrations tracking table
   CREATE TABLE IF NOT EXISTS schema_migrations (
       version TEXT PRIMARY KEY,
       applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   );
   ```

2. **Missing indexes on `review_feedback_targets`.** The table has an index on `(target_type, target_id)` which is good for "find feedback targeting annotation X." But there is no index on `feedback_id` alone. The `listFeedbackTargets` query filters by `feedback_id`, so this would benefit from an index:

   ```sql
   CREATE INDEX IF NOT EXISTS idx_review_feedback_targets_feedback
       ON review_feedback_targets(feedback_id);
   ```

   Currently the query `WHERE feedback_id = ? ORDER BY target_type, target_id` does a full table scan if no index on `feedback_id`.

3. **`review_guidelines` full-text search is via `LIKE`.** The `ListGuidelines` repository method uses `WHERE title LIKE ? OR slug LIKE ? OR body_markdown LIKE ?` for search. This is a prefix-wildcard-suffix pattern (`%term%`) which prevents SQLite from using indexes. For small datasets this is fine, but it a design doc it should note this as a future FTS5 improvement target.

### 3.2 Types Layer — `pkg/annotate/types.go`

**What it does:** Defines all Go domain types. The new types added are:

- `ReviewFeedback`, `FeedbackTarget` — Structured reviewer feedback
- `ReviewGuideline` — Reusable policy document
- `RunGuidelineLink` — Join record
- `ReviewCommentInput` — Embedded comment within a review action
- `ReviewAnnotationActionInput`, `BatchReviewActionInput` — Combined review+comment+guideline inputs
- Various string constants for scope kinds, feedback kinds, statuses, guideline scopes

**What is clean:**

- Clear separation between domain types (`types.go`) and HTTP DTOs (`types_feedback.go`)
- String constants prevent typos in magic strings
- `db` and `json` struct tags are consistent with existing patterns
- `UpdateGuidelineInput` uses `*string` pointers (nullable) to distinguish "not provided" from "set to empty"

**Issues found:**

1. **`FeedbackTarget.FeedbackID` is serialized as `feedbackId` but targets are always loaded through the parent.** The `FeedbackID` field on `FeedbackTarget` is technically redundant when targets are loaded via `listFeedbackTargets(ctx, feedbackID)` — the caller already knows the feedback ID. This is not a bug (the field correctly maps the DB column), but it is slightly wasteful serialization since every target in a feedback object will have the same `feedbackId`.

2. **No `BodyMarkdown` field on `UpdateFeedbackInput`.** The type only allows updating `Status`:

   ```go
   type UpdateFeedbackInput struct {
       Status string
   }
   ```

   But the HTTP handler (`handleUpdateFeedback`) also mentions `bodyMarkdown` in its validation comment and the frontend `UpdateFeedbackRequest` type includes it. The backend silently ignores any body update. This is a missing feature — if a reviewer wants to edit the text of their feedback, the API contract says they can but the implementation does not honor it.

   **Recommendation:** Either add `BodyMarkdown *string` to `UpdateFeedbackInput` and wire it through, or remove `bodyMarkdown` from the HTTP request type and the frontend type.

3. **`defaultReviewState` function is called but not defined in the reviewed files.** The `updateAnnotationReviewStateTx` function calls `defaultReviewState(reviewState, "")` which is presumably defined in the existing `repository.go`. This is fine but worth noting — the helper lives outside the diff.

### 3.3 Repository Layer — `pkg/annotate/repository_feedback.go`

This is the most complex and most important file in the branch. It contains ~622 lines.

**What it does:**

- CRUD for `review_feedback` (create, get, list, update)
- CRUD for `review_guidelines` (create, get, list, update)
- Run-guideline link operations (link, unlink, list)
- **Transactional combined operations** — `ReviewAnnotationWithArtifacts` and `BatchReviewWithArtifacts`
- Internal transaction helpers (`getAnnotationTx`, `updateAnnotationReviewStateTx`, `batchUpdateReviewStateTx`, `createReviewFeedbackTx`, `linkGuidelineToRunTx`, `inferSingleRunIDForAnnotationsTx`)

**What is clean:**

- The transactional pattern is correct: begin tx → defer rollback → do work → commit. The deferred rollback is a no-op after commit.
- `inferSingleRunIDForAnnotationsTx` is a smart solution for batch guideline linking — it refuses to guess if annotations span multiple runs.
- `isDuplicateKeyError` handles both SQLite (`UNIQUE constraint failed`) and Postgres (`duplicate key`) error strings.
- Input validation is thorough — empty strings are trimmed and rejected.

**Issues found:**

1. **⚠️ Duplicated SQL — feedback creation appears twice.** The `CreateReviewFeedback` method and the `createReviewFeedbackTx` method contain almost identical SQL INSERT logic. The non-transactional version:

   ```go
   // CreateReviewFeedback — standalone, manages own tx
   tx, err := r.db.BeginTxx(ctx, nil)
   // ... insert feedback + targets ...
   tx.Commit()
   ```

   The transactional version:

   ```go
   // createReviewFeedbackTx — receives tx from caller
   // ... insert feedback + targets ... (same SQL, different signature)
   ```

   This means any change to the feedback INSERT columns must be made in two places. **Recommendation:** Extract a `createReviewFeedbackCore(ctx, db Executor, input) (string, error)` that accepts either a `sqlx.DB` or `sqlx.Tx` using a common interface.

2. **⚠️ N+1 query in `ListReviewFeedback`.** After loading feedback records, the code loops and calls `listFeedbackTargets` for each one:

   ```go
   for i := range feedbacks {
       targets, err := r.listFeedbackTargets(ctx, feedbacks[i].ID)
       feedbacks[i].Targets = targets
   }
   ```

   This issues one additional SELECT per feedback record. For 100 feedback items, that is 101 queries. **Recommendation:** Batch-load targets using `WHERE feedback_id IN (?)` with `sqlx.In`, then group by feedback ID in Go.

3. **`defaultString` helper is defined here but belongs in a shared utilities file.** The `defaultString(value, fallback string)` function is a general-purpose helper. It should be in a shared utilities package or at the top of the existing repository file. Currently it lives at the bottom of `repository_feedback.go`, which means other repository files cannot reuse it without import cycles.

4. **`BatchReviewWithArtifacts` returns `error` but `ReviewAnnotationWithArtifacts` returns `(*Annotation, error)`.** The inconsistency means the batch path gives no feedback about which annotations were affected. This is acceptable for now (204 No Content response) but limits future use.

5. **`defer func() { _ = tx.Rollback() }()` is the correct Go pattern** but it relies on the fact that `Rollback` after `Commit` is a no-op. This is worth a code comment for the next reader.

6. **The `CreateReviewFeedback` standalone method is never called.** Looking at the handlers, `handleCreateFeedback` calls `h.annotations.CreateReviewFeedback`. So it IS used — but only from the standalone feedback creation endpoint. The real question is: should the standalone endpoint exist at all, given that feedback is also created inside `ReviewAnnotationWithArtifacts`? The answer is yes — run-level feedback is created independently from review-state changes. This is fine.

7. **`listFeedbackTargets` uses `feedback_id` but the column was defined as `feedback_id` in schema.** This is correct. However, the `FeedbackTarget` struct has `db:"feedback_id"` but the `review_feedback_targets` table has columns `feedback_id, target_type, target_id`. The `FeedbackID` is correctly populated by `sqlx`. This is fine.

### 3.4 HTTP Handler Layer — `pkg/annotationui/handlers_annotations.go`

**What it does:** Extends existing annotation review endpoints to accept optional `comment`, `guidelineIds`, and `mailboxName` fields. Also contains listing/get handlers for annotations, groups, logs, runs.

**What is clean:**

- The `reviewRequest` and `batchReviewRequest` structs correctly separate the extended fields with `json:"comment,omitempty"` — old clients that only send `reviewState` continue to work.
- `toAnnotateReviewComment` correctly null-checks: if both title and body are empty, it returns nil (no feedback created).
- `isValidReviewState` is strict: only `to_review`, `reviewed`, `dismissed`.
- `decodeJSONBody` uses `DisallowUnknownFields()` which means adding new fields will be a coordinated deploy.

**Issues found:**

1. **`decodeJSONBody` with `DisallowUnknownFields` is a forward-compatibility risk.** If the frontend adds a new field before the backend is updated (or vice versa), the request will be rejected. In a monorepo this is manageable, but it is worth noting. The alternative is to silently ignore unknown fields.

2. **`handleReviewAnnotation` returns the annotation but not the feedback.** When a review action creates feedback, the client gets back the updated annotation but has no way to know the feedback ID that was created. The client would need to refetch feedback separately (which the RTK Query cache invalidation handles). This is fine for now.

3. **`handleBatchReview` returns `204 No Content` on success.** This means the client does not know which feedback ID was created. Again, RTK Query invalidation handles the refetch, but a 200 with `{ feedbackId: "..." }` would be more informative.

### 3.5 Feedback & Guideline Handlers — `pkg/annotationui/handlers_feedback.go`

**What it does:** Adds 10 new HTTP endpoints for feedback CRUD, guideline CRUD, and run-guideline link management.

**What is clean:**

- Consistent error handling pattern: validation → repository call → not-found check → response
- `handleLinkRunGuideline` returns the updated guideline list (not just 204), which is friendlier to the frontend
- Slug uniqueness is enforced with 409 Conflict response
- Status validation uses allow-list (`isValidFeedbackStatus`, `isValidGuidelineStatus`)

**Issues found:**

1. **`handleLinkRunGuideline` returns the full guideline list after linking.** This is the only link endpoint that returns a body — `handleUnlinkRunGuideline` returns 204. The inconsistency is not harmful but is worth noting. The frontend `useLinkGuidelineToRunMutation` has `invalidatesTags: ["Guidelines", "Runs"]` which would refetch anyway, so the response body is arguably redundant.

2. **Empty `linkedBy` string passed to `LinkGuidelineToRun`.** The handler calls:

   ```go
   h.annotations.LinkGuidelineToRun(r.Context(),
       r.PathValue("id"),
       req.GuidelineID,
       "",  // ← empty linkedBy
   )
   ```

   The `created_by` / `linked_by` fields are empty strings everywhere because the annotation UI server does not have authentication yet. This is documented as a known gap. When auth is added, these should be populated from the session.

3. **`handleCreateGuideline` silently swallows `createdBy`.** The handler does not pass any `CreatedBy` value to the repository:

   ```go
   g, err := h.annotations.CreateGuideline(r.Context(), annotate.CreateGuidelineInput{
       Slug:         req.Slug,
       Title:        req.Title,
       ScopeKind:    req.ScopeKind,
       BodyMarkdown: req.BodyMarkdown,
       // CreatedBy: missing — defaults to ""
   })
   ```

   Same issue for `handleCreateFeedback` — the repository input has `CreatedBy` but the handler never sets it. This is consistent with the no-auth state, but should be tracked.

### 3.6 HTTP DTO Layer — `pkg/annotationui/types_feedback.go`

**What it does:** Defines request/response types for the HTTP layer and conversion helpers (`feedbackToResponse`, `guidelineToResponse`).

**What is clean:**

- Clear separation: Go domain types (in `pkg/annotate/types.go`) vs HTTP wire types (here)
- Conversion helpers are thorough — timestamps are formatted to ISO 8601
- `isValidFeedbackStatus` and `isValidGuidelineStatus` use switch statements for validation

**Issues found:**

1. **`feedbackTargetJSON` uses different field names than `FeedbackTarget`.** The Go domain type has `db:"feedback_id" json:"feedbackId"` but the HTTP response type has `TargetType` and `TargetID` (no `FeedbackID`). This is intentional — the client does not need the feedback ID because it is implied by the parent object. But it means the same conceptual entity (`FeedbackTarget`) has two different JSON shapes depending on whether you are looking at the domain type or the HTTP response type. This is fine but could confuse a new developer.

2. **`reviewCommentInput` type is defined here AND in `handlers_annotations.go`.** Wait — actually it's only in `types_feedback.go`. The `handlers_annotations.go` file uses `reviewCommentInput` which is the same type. This is correct — `reviewCommentInput` is defined once in `types_feedback.go` and used from `handlers_annotations.go` because they are in the same package. This is fine.

### 3.7 Server Wiring — `pkg/annotationui/server.go`

**What it does:** Registers all HTTP routes on the Go `http.ServeMux`.

**What is clean:**

- Route pattern is consistent with Go 1.22+ `METHOD /path/{param}` syntax
- Routes are organized by section with clear comments
- Health check endpoints are separate from API endpoints

**Issues found:**

1. **No CORS middleware.** The annotation UI server serves both the API and the SPA. In production this is fine (same origin). But in development, the Vite dev server runs on a different port and proxies `/api` requests. This means CORS is handled by the Vite proxy, not by the Go server. If anyone ever runs the frontend on a different domain without the proxy, requests will fail. This is acceptable for now but should be documented.

2. **`handleLinkRunGuideline` uses POST (not PUT) for idempotent linking.** The `ON CONFLICT DO NOTHING` in the repository means POSTing the same link twice is idempotent. REST purists would argue PUT is more semantically correct for idempotent creation, but POST is simpler and the idempotency is handled at the DB level. This is a valid design choice.

### 3.8 Sender Detail Mailbox — `pkg/annotationui/handlers_senders.go`

**What changed:** Added `mailbox_name` to the sender detail message preview query.

**What is clean:**

- Minimal change: one column added to SELECT, one field added to struct
- Uses `COALESCE(NULLIF(sent_date, ''), NULLIF(internal_date, ''))` pattern consistent with existing query

**Issues found:** None. This is a clean, minimal change.

## 4. Frontend Code Review
### 4.1 Type System — `ui/src/types/reviewFeedback.ts` and `reviewGuideline.ts`

**What they do:** Define TypeScript types that mirror the Go backend domain types.

**What is clean:**

- Type unions (`FeedbackKind = "comment" | "reject_request" | ...`) prevent typos
- Request types are separate from entity types (e.g., `CreateFeedbackRequest` vs `ReviewFeedback`)
- `ReviewCommentDraft` is a stripped-down version for inline review actions
- Filter types (`FeedbackFilter`, `GuidelineFilter`) match the query parameters

**Issues found:**

1. **`UpdateFeedbackRequest` includes `bodyMarkdown` but the backend ignores it.**

   ```ts
   // ui/src/types/reviewFeedback.ts
   export interface UpdateFeedbackRequest {
     status?: FeedbackStatus;
     bodyMarkdown?: string;  // ← backend ignores this
   }
   ```

   The Go `updateFeedbackRequest` type has `Status string` only. The frontend type promises the ability to update `bodyMarkdown` but the backend will silently ignore it. This is a **contract mismatch**. Either the backend should honor it, or the frontend type should remove it.

2. **`FeedbackScopeKind` includes `"guideline"` but no UI flow creates guideline-scoped feedback.** The `scopeKind: "guideline"` value exists in both Go and TypeScript, but there is no UI flow that creates feedback with this scope. It is reserved for future use. This is fine but should be documented as unused.

### 4.2 RTK Query Contract — `ui/src/api/annotations.ts`

**What it does:** Defines all API endpoints using RTK Query. Adds 10 new endpoints and extends 2 existing ones.

**What is clean:**

- Cache tags (`Feedback`, `Guidelines`) are correctly provided and invalidated
- `reviewAnnotation` invalidates `["Annotations", "Runs", "Feedback"]` — correct, because dismissing with feedback changes all three
- `batchReview` has the same invalidation — correct
- Hook exports are organized by section with comments
- `getRunGuidelines` provides `["Guidelines", "Runs"]` tags — correct because it joins both entities

**Issues found:**

1. **`linkGuidelineToRun` return type is `void` but the backend returns the updated guideline list.** The RTK mutation type says:

   ```ts
   linkGuidelineToRun: builder.mutation<void, { runId: string; guidelineId: string }>({
   ```

   But the Go handler `handleLinkRunGuideline` returns `200 OK` with a JSON array of guidelines. The frontend discards this. This is not a bug (the cache invalidation handles the refetch), but it means the backend is doing unnecessary work serializing a response that is never used. **Recommendation:** Either change the backend to return `204 No Content` or change the frontend type to capture the response for an optimistic update.

2. **`getGuideline` uses `providesTags: ["Guidelines"]` but should use a specific tag for cache granularity.** Currently:

   ```ts
   getGuideline: builder.query<ReviewGuideline, string>({
     query: (id) => `review-guidelines/${id}`,
     providesTags: ["Guidelines"],
   }),
   ```

   If any guideline is updated, ALL guideline queries refetch. The more RTK-idiomatic pattern is:

   ```ts
   providesTags: (result, error, id) => [{ type: "Guidelines", id }],
   ```

   This would allow invalidating a specific guideline when it is updated, not all of them.

3. **No `keepUnusedDataFor` tuning on guideline endpoints.** Guidelines change rarely. Adding `keepUnusedDataFor: 300` (5 minutes) would reduce refetches when navigating between list and detail pages.

### 4.3 Shared Badges — `ui/src/components/shared/`

Four new badge components: `MailboxBadge`, `FeedbackKindBadge`, `FeedbackStatusBadge`, `GuidelineScopeBadge`.

**What is clean:**

- Each follows the same pattern: MUI `Chip`, `size="small"`, color-coded by value, `data-part` attribute
- Icon mapping in `MailboxBadge` is clean: a static object keyed by mailbox name
- All have Storybook stories with default, variant, and "all values" stories
- Empty string mailbox renders nothing (correct behavior)

**Issues found:**

1. **`GuidelineScopeBadge` has an icon mapping for `"run"` that does not exist in the `GuidelineScopeKind` type.** Looking at the type:

   ```ts
   export type GuidelineScopeKind = "global" | "mailbox" | "sender" | "domain" | "workflow";
   ```

   There is no `"run"` value. But the badge component may have been written to handle it defensively. This is a minor inconsistency — the badge handles more values than the type allows, which is good defensive programming but slightly misleading.

2. **Badge color choices may have insufficient contrast in dark theme.** For example, `FeedbackKindBadge` uses `"warning"` for `reject_request`. In MUI dark theme, warning chips can have low contrast against dark backgrounds. This should be verified with an accessibility audit.

### 4.4 ReviewCommentDrawer — `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx`

**What it does:** The primary feedback entry UI. Despite the name "Drawer", it is a MUI `Dialog` (modal). It supports three modes: `single`, `batch`, and `run`.

**What is clean:**

- Form state resets on open via `useEffect`
- `feedbackKind` defaults differently per mode (`"comment"` for run, `"reject_request"` for single/batch)
- Guideline attachment is expandable (Collapse) so it doesn't clutter the simple case
- `title.trim()` is required before submit, `bodyMarkdown` is optional — correct UX
- The `onSubmit` callback returns a flat object, keeping the component decoupled from API types

**⚠️ Issues found:**

1. **⚠️ Naming: "Drawer" but it is a Dialog.** The file is named `ReviewCommentDrawer.tsx`, the data-part is `commentDrawer`, and the component is called `ReviewCommentDrawer`. But it renders a `<Dialog>`. The diary explains this was originally a Drawer that was converted to a Dialog during Step 18, but the name was not updated. **Recommendation:** Rename to `ReviewCommentDialog.tsx` and update all imports and references. The current name will confuse future developers.

2. **`useEffect` with `open` and `mode` dependencies but references `resetForm`.** The effect:

   ```ts
   useEffect(() => {
     if (open) { resetForm(); }
   }, [open, mode]);
   ```

   `resetForm` is not in the dependency array. This works because `resetForm` is a plain function that reads state via closures (not a stable useCallback reference). React's exhaustive deps lint rule would flag this. **Recommendation:** Move the reset logic inline or add `resetForm` to deps (after wrapping it in `useCallback`).

3. **`agentRunId` prop is accepted but never used inside the component.** The `ReviewCommentDrawerProps` interface includes `agentRunId?: string` but the component body never references it. This prop exists in the type but does nothing. The `agentRunId` is used at the caller level (the pages pass it to the mutation directly). **Recommendation:** Remove `agentRunId` from `ReviewCommentDrawerProps` since it is unused.

4. **`targetCount` is used only for the submit button label, not for any logic.** The `targetCount` prop appears in:

   ```ts
   const submitLabel = mode === "batch" ? `Reject ${targetCount} Items` : ...
   ```

   This is fine. It is a display prop. But when `mode === "batch"` and `targetCount === 0`, the button would say "Reject 0 Items". The parent should guard against this (and `ReviewQueuePage` does — the drawer only opens when items are selected).

5. **`guidelinesExpanded` state defaults to `false` and resets on every open.** This means the user must re-expand the guidelines section every time they open the dialog. For power users who always attach guidelines, this is a minor annoyance. Consider persisting this preference.

DOCFILE="ttmp/2026/04/06/SMN-20260406-CODE-REVIEW--code-review-run-review-feedback-guidelines-mailbox-aware-analysis/design-doc/01-comprehensive-code-review-run-review-feedback-guidelines-mailbox-aware-analysis.md"
echo "Chunk 8 done (annotation table)"

### 4.5 AnnotationTable — `ui/src/components/AnnotationTable/AnnotationTable.tsx`

**What it does:** Renders the main annotation list with selection, expansion, review actions, and an optional per-row "Dismiss & Explain" button. The key architectural feature is `AnnotationTableItem`, a `React.memo`-wrapped component that prevents unnecessary rerenders when the user toggles checkboxes.

**What is clean:**

- `selectedSet` is a `useMemo(() => new Set(selected))` — O(1) lookups instead of O(N) `Array.includes()`
- `expandedRelated` is computed only for the single expanded annotation, not for every row
- The memo comparator checks stable visual state (data identity, boolean flags, callback identity) rather than doing deep equality
- `EMPTY_RELATED` constant avoids creating a new empty array on every render
- Only the expanded row mounts `AnnotationDetail` — collapsed rows render nothing

**Issues found:**

1. **The memo comparator compares callback identity (`prev.onToggleSelect === next.onToggleSelect`).** This means the memo is only effective if the parent stabilizes these callbacks with `useCallback`. The `ReviewQueuePage`, `RunDetailPage`, and `SenderDetailPage` all do stabilize their callbacks, so this works. But if a future page passes inline arrow functions, the memo breaks silently. **Recommendation:** Add a code comment above the comparator explaining this requirement.

2. **`columnCount` is a constant (`COLUMN_COUNT = 8`) passed as a prop through `AnnotationTableItem`.** This is harmless but unnecessary — it could be imported directly in `AnnotationDetail`. Passing it as a prop adds it to the memo comparator for no benefit.

3. **`AnnotationRow` is NOT memoized.** Only `AnnotationTableItem` (the row+detail pair) is wrapped in `React.memo`. Since `AnnotationRow` receives new arrow-function props (`() => onApprove(annotation.id)`) on every render, it rerenders every time even though the visual output may not change. For large tables, memoizing `AnnotationRow` individually would be more effective than memoizing the outer wrapper.

### 4.6 AnnotationRow — `ui/src/components/AnnotationTable/AnnotationRow.tsx`

**What it does:** Renders a single annotation table row with checkbox, target, tag, note, source, status, date, and action buttons (approve, dismiss, optional dismiss-explain).

**Issues found:**

1. **New `onDismissExplain` prop is optional but the button renders unconditionally when the prop is provided.** The `SpeakerNotesIcon` button appears on every row in sender detail and run detail pages. This is correct behavior — the button is the entry point for per-row dismiss-with-feedback. No issue here.

2. **The row creates inline arrow functions for every callback prop:**

   ```ts
   onToggleSelect={() => onToggleSelect(annotation.id)}
   ```

   This means the parent `AnnotationTableItem` memo comparator MUST compare `onToggleSelect` identity (which it does). But within the row, every child component receives a new function reference on every render. For MUI components this is fine (they are lightweight), but it is worth noting.

### 4.7 ReviewCommentInline — `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx`

**⚠️ DEAD CODE — This component is never imported or used anywhere in the application.**

The file exists, compiles cleanly, and has Storybook stories, but no page or component imports it. It was designed for an inline dismiss-with-feedback form (embedded below the annotation row), but the actual implementation uses the modal `ReviewCommentDrawer` instead.

**Recommendation:** Either delete this file and its stories, or add a code comment explaining it is reserved for future use. Currently it is dead code that adds confusion.

### 4.8 GuidelinePicker — `ui/src/components/ReviewFeedback/GuidelinePicker.tsx`

**What it does:** A checkbox list of active guidelines, used inside the `ReviewCommentDrawer` dialog.

**Issues found:**

1. **Calls `useListGuidelinesQuery({ status: "active" })` internally.** This means every instance of `GuidelinePicker` triggers a separate API call. If the dialog is mounted in multiple pages simultaneously, this could cause duplicate requests. In practice only one dialog is open at a time, so this is fine. But RTK Query's caching means if the guidelines were already fetched, it returns the cached data — so the duplication risk is low.

2. **No loading or error state handling.** If the guidelines query fails, the picker silently shows nothing. **Recommendation:** Add a small error message or retry button.

### 4.9 FeedbackCard — `ui/src/components/ReviewFeedback/FeedbackCard.tsx`

**What it does:** Displays a single feedback record with badges (kind, status), title, body, metadata, and acknowledge/resolve action buttons.

**Issues found:**

1. **Acknowledge/Resolve buttons call `useUpdateReviewFeedbackMutation` directly inside the card.** This means the component owns its own mutation logic, which couples it to the API layer. An alternative pattern would be to pass `onAcknowledge` and `onResolve` callbacks from the parent. The current approach works and is simpler, but makes the card harder to reuse in contexts where the mutation should be batched or handled differently.

2. **No "archive" button.** The feedback status enum includes `"archived"` but the card only shows "Acknowledge" and "Resolve". The archive transition has no UI entry point. This is acceptable for v1 (archiving is a cleanup operation that can be done via the API directly).

### 4.10 RunFeedbackSection — `ui/src/components/ReviewFeedback/RunFeedbackSection.tsx`

**What it does:** A section wrapper that shows a list of `FeedbackCard` items for a specific run, with an "Add Run Feedback" button.

**Issues found:**

1. **The "Add Run Feedback" button opens a `ReviewCommentDrawer` in `mode="run"`, but the `onSubmit` callback creates feedback via `createReviewFeedback` (a standalone endpoint), NOT via the review-annotation endpoint.** This is correct — run-level feedback is independent from annotation review state. But it is worth noting that there are two code paths for feedback creation:
   - **Path A:** Review action → `reviewAnnotation` mutation → backend creates feedback inside transaction
   - **Path B:** Run-level → `createReviewFeedback` mutation → backend creates feedback standalone

   Both paths write to the same `review_feedback` table but through different API endpoints. This is fine architecturally.

### 4.11 GuidelineLinkPicker — `ui/src/components/ReviewFeedback/GuidelineLinkPicker.tsx`

**What it does:** A modal dialog for searching and selecting guidelines to link to a run. Used from `RunGuidelineSection`.

**Issues found:**

1. **Local search filter duplicates server-side search.** The component fetches ALL active guidelines via `useListGuidelinesQuery({ status: "active" })` and then filters them client-side with a text input. The server-side `search` query parameter on the guidelines endpoint is not used. For small datasets this is fine, but for hundreds of guidelines, fetching all of them and filtering client-side is wasteful. **Recommendation:** Pass the search text to the RTK query filter so the backend handles pagination and search.

2. **`alreadyLinkedIds` prop is used to disable already-linked guidelines.** The component disables checkboxes for guidelines already linked to the run. The disabled items are still visible and checkable (the checkbox just doesn't toggle). **Recommendation:** Consider hiding already-linked guidelines or moving them to a separate "Already Linked" section.

### 4.12 GuidelineForm — `ui/src/components/Guidelines/GuidelineForm.tsx`

**What it does:** A three-mode form (view / edit / create) for guideline management. Handles slug, title, scope, status, priority, and markdown body with a preview tab.

**Issues found:**

1. **Edit/Preview tab pattern is clean.** The body field uses MUI `Tabs` to switch between editing raw markdown and viewing rendered output. The preview renders inside a `MarkdownRenderer` component (from the existing codebase). This is a good UX pattern.

2. **`canSave` validation differs between modes but is not visually communicated.** In create mode, slug is required. In edit mode, it is not. The submit button is disabled when `canSave` is false, but there is no tooltip or helper text explaining *why* it is disabled. A user might type a title and wonder why "Save" is still greyed out (because they forgot the slug in create mode).

3. **The form does not validate slug format.** The slug field accepts any string including spaces, special characters, and uppercase letters. Conventional slugs are lowercase-with-hyphens. **Recommendation:** Either auto-normalize the slug (replace spaces with hyphens, lowercase) or add a validation hint.

### 4.13 GuidelineSummaryCard — `ui/src/components/Guidelines/GuidelineSummaryCard.tsx`

**What it does:** A summary card used in the guidelines list page. Shows title, body preview (truncated), scope badge, status, priority, linked run count, and action buttons.

**Issues found:**

1. **`linkedRunCount` prop is always `0` in practice.** The `GuidelinesListPage` passes `linkedRunCount={0}` for every guideline because there is no backend endpoint that returns "how many runs are linked to this guideline." The card reserves space for this information but it is always zero. **Recommendation:** Either add a backend endpoint for linked-run counts or remove the prop and show "N/A" for now.

### 4.14 GuidelineLinkedRuns — `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx`

**⚠️ ALWAYS EMPTY — This component is used in `GuidelineDetailPage` but always receives `runs={[]}`.**

The `GuidelineDetailPage` hard-codes:
```tsx
<GuidelineLinkedRuns runs={[]} onNavigateRun={...} />
```

There is no backend endpoint for "get all runs linked to guideline X." The component itself is well-built (it renders a list of run summary rows with annotation counts), but it has never shown data. **Recommendation:** Add a `GET /api/review-guidelines/{id}/runs` endpoint, or remove this section from the detail page until the backend is ready.

### 4.15 Page Components

#### ReviewQueuePage — `ui/src/pages/ReviewQueuePage.tsx`

**What it does:** The main review queue. Shows filter pills, count summary, batch action bar, annotation table, and the reject-explain dialog.

**What is clean:**

- All handlers are stabilized with `useCallback`
- `toggleExpandedId` action from Redux keeps the expand handler stable
- `handleGetRelated` is memoized with `useCallback`
- The dialog state uses local `useState` (not Redux) which is correct for ephemeral UI state

**Issues found:**

1. **⚠️ `commentDrawerOpen` state exists in BOTH local `useState` AND Redux.** The page uses:
   ```ts
   const [commentDrawerOpen, setCommentDrawerOpen] = useState(false);
   ```
   But the Redux slice also has `commentDrawerOpen: boolean` with `openCommentDrawer` / `closeCommentDrawer` actions. The Redux state is never read or written by this page. The `openCommentDrawer` and `closeCommentDrawer` actions are exported but never dispatched anywhere.

   **This is dead Redux state.** Either:
   - Migrate the page to use Redux for drawer state (as the diary suggests), or
   - Remove the Redux fields and actions

2. **`filterMailbox` Redux field is also unused.** The slice has `filterMailbox: string | null` and `setFilterMailbox` action, but no component reads or writes this state. It was added "for future mailbox filter pills" (per the diary) but the pills were never implemented.

3. **`handleCommentSubmit` calls `batchReview` with `reviewState: "dismissed"` but does not pass `agentRunId`.** This means the backend's `BatchReviewWithArtifacts` cannot infer a run ID from the selected annotations (it tries, but only if `guidelineIds` are present). If the user submits a batch reject-explain with no guidelineIds, the feedback is created without a run link. This is technically correct (feedback does not require a run), but it means the feedback won't appear in the run's feedback section.

#### RunDetailPage — `ui/src/pages/RunDetailPage.tsx`

**What it does:** Shows a single agent run with linked guidelines, timeline, run-level feedback, groups, and annotations.

**What is clean:**

- `useGetRunGuidelinesQuery` and `useListReviewFeedbackQuery` are skipped when `runId` is undefined — correct
- `RunGuidelineSection` is placed before timeline, `RunFeedbackSection` after — matches the reviewer's mental flow
- Per-row dismiss-explain opens the `ReviewCommentDrawer` in single mode

**Issues found:**

1. **`handleDismissExplain` does `annotations.find(item => item.id === id)` on every call.** This is an O(N) linear scan. For small lists this is fine, but a `Map<string, Annotation>` would be more efficient. This is a minor performance concern.

2. **The page fetches `linkedGuidelines` and `feedback` with separate hooks, which means two API calls on mount.** These could potentially be combined into a single "run detail with relations" endpoint, but the current approach is cleaner for caching granularity.

3. **`handleCommentSubmit` sends `guidelineIds` but no `agentRunId` on the review payload.** The backend gets the run ID from the annotation itself (via `ReviewAnnotationWithArtifacts`). This is correct — the annotation already has `agentRunId`.

#### SenderDetailPage — `ui/src/pages/SenderDetailPage.tsx`

**What it does:** Shows sender profile, recent messages (with mailbox column), and annotations.

**What is clean:**

- `senderMailboxName` is computed with `useMemo` — extracts unique mailbox names from recent messages and returns one only if all messages share the same mailbox. This avoids showing a misleading single-mailbox badge when the sender appears in multiple mailboxes.
- Per-row dismiss-explain works the same as RunDetailPage

**Issues found:**

1. **`senderMailboxName` filters with `mailboxName.length > 0` which treats empty strings as "no mailbox".** But the backend defaults `mailbox_name` to `""` for messages where it is unknown. This is correct behavior — unknown mailbox should not be treated as a distinct mailbox name.

2. **The page does not fetch guidelines or feedback.** The sender detail page shows annotations but not feedback or guidelines. This means a reviewer cannot see feedback that was created for annotations on this sender. This is a feature gap, not a bug.

#### GuidelinesListPage — `ui/src/pages/GuidelinesListPage.tsx`

**Issues found:**

1. **Client-side search duplicates the server-side search.** The page fetches guidelines with the `status` filter passed to the query, but then does a second client-side `useMemo` filter on `search`. The RTK Query endpoint already accepts `search` as a query parameter. The page should pass `search` to the query instead of filtering client-side. Currently the page fetches ALL guidelines matching the status filter, then filters locally. For large guideline sets this is wasteful.

   **Current (wasteful):**
   ```ts
   const { data: guidelines } = useListGuidelinesQuery({
     status: statusFilter === "all" ? undefined : statusFilter,
     search: search || undefined,  // ← this IS passed to the query
   });
   const filtered = useMemo(() => {
     if (!search) return guidelines;
     // ... client-side filter on title/slug/body
   }, [guidelines, search]);
   ```

   Wait — the search IS passed to the query. And then it is ALSO filtered client-side. This means the search is applied twice: once by the backend and once by the frontend. With MSW mocks, the backend search works. With the real backend, this would also work. But the double-filter is redundant. **Recommendation:** Remove the client-side `filtered` useMemo since the backend already handles search.

#### GuidelineDetailPage — `ui/src/pages/GuidelineDetailPage.tsx`

**Issues found:**

1. **Create-then-link race condition.** The `handleSave` in create mode:
   ```ts
   void createGuideline(payload).then((result) => {
     if ("data" in result && result.data) {
       if (runIdParam) {
         void linkGuidelineToRun({ runId: runIdParam, guidelineId: result.data.id });
         navigate(`/annotations/runs/${runIdParam}`);
       }
     }
   });
   ```
   The `linkGuidelineToRun` call is fire-and-forget (`void`). If the link fails, the user is already navigated back to the run detail page and the guideline exists but is not linked. **Recommendation:** Await the link call and handle errors.

2. **The `isNew` detection uses `guidelineId === "new"`.** This works because the route `/annotations/guidelines/new` maps to `:guidelineId = "new"`. But it is a magic string comparison. If someone creates a guideline with ID "new", it would be treated as a new-guideline form. This is extremely unlikely but worth noting.

### 4.16 Redux Slice — `ui/src/store/annotationUiSlice.ts`

**What changed:** Added three new fields and actions to the `ReviewQueueState`:
- `commentDrawerOpen: boolean` + `openCommentDrawer` / `closeCommentDrawer`
- `filterMailbox: string | null` + `setFilterMailbox`
- `toggleExpandedId` action (replaces `setExpandedId` toggle logic in the page)

**⚠️ Dead state analysis:**

| Redux Field | Written by | Read by | Status |
|-------------|-----------|---------|--------|
| `commentDrawerOpen` | `openCommentDrawer`, `closeCommentDrawer` | **Nobody** | **⚠️ DEAD** — pages use local `useState` instead |
| `filterMailbox` | `setFilterMailbox` | **Nobody** | **⚠️ DEAD** — no filter pills use it |
| `expandedId` (existing) | `setExpandedId`, `toggleExpandedId` | `ReviewQueuePage` | ✅ Live |
| `toggleExpandedId` action | `ReviewQueuePage` | `ReviewQueuePage` | ✅ Live |

**Recommendation:** Remove `commentDrawerOpen`, `openCommentDrawer`, `closeCommentDrawer`, `filterMailbox`, and `setFilterMailbox` from the slice. Add them back when the features that need them are actually implemented.

### 4.17 Mock Layer — `ui/src/mocks/annotations.ts` and `handlers.ts`

**What they do:** MSW v2 mock data and handlers for Storybook and local dev. Contains 4 mock feedback items, 4 mock guidelines, and handlers for all new endpoints.

**What is clean:**

- `runGuidelineLinks` Map is correctly placed outside the handlers array (MSW handlers are a static array literal, mutable state must be in closure scope)
- Mock data is realistic — uses real-world examples (billing classification, transactional vs promotional)
- Handlers correctly simulate CRUD operations including 404, 409, and mutation

**Issues found:**

1. **Mock handlers do not persist mutations across requests (except run-guideline links).** When `POST /api/review-guidelines` creates a new guideline, it returns a synthetic response but does not add it to `mockGuidelines`. A subsequent `GET /api/review-guidelines` will not include the newly created guideline. This means the "create then list" flow is broken in Storybook. **Recommendation:** Use a mutable array (like `runGuidelineLinks`) for mock data that should survive mutations.

2. **`mockGuidelines[0]!` uses non-null assertion in Storybook stories.** This is required because TypeScript cannot prove the array access is safe. It is acceptable for mock data but would cause a runtime error if the mock array were ever emptied.

3. **`http.get("/api/review-guidelines")` handler uses the correct URL pattern** but some Storybook stories override it with `http.get("/api/guidelines")` (missing `review-` prefix). This means those Storybook stories' MSW handlers never match, and the fallback handler (from `handlers.ts`) is used instead. This works by accident — the fallback handler returns the correct data. But the Storybook-specific override is dead code. Found in `GuidelinesListPage.stories.tsx` and `GuidelineDetailPage.stories.tsx`.

### 4.18 Routing — `ui/src/App.tsx` and `AnnotationSidebar.tsx`

**What changed:** Three new routes and one new sidebar entry.

**Issues found:**

1. **Route order is correct.** `guidelines/new` is registered before `guidelines/:guidelineId`, so "new" does not match the parameterized route. This was explicitly noted in the diary as a learning.

2. **Sidebar `isActive` check uses `location.pathname.startsWith(item.path)`.** This means `/annotations/guidelines` correctly highlights "Guidelines", and `/annotations/guidelines/guideline-001` also highlights it. This is the desired behavior — the sidebar entry stays active on child routes.

## 5. Enrich Command Changes — `cmd/smailnail/commands/enrich/`

### What changed

The four enrich command files (`all.go`, `senders.go`, `threads.go`, `unsubscribe.go`) were modified to **flatten the embedded `enrichSettings` struct**. Previously, each command's settings struct embedded `enrichSettings`:

```go
// BEFORE
type allSettings struct {
    enrichSettings  // embedded
}
```

Now they declare the fields explicitly:

```go
// AFTER
type allSettings struct {
    SQLitePath string `glazed:"sqlite-path"`
    AccountKey string `glazed:"account-key"`
    Mailbox    string `glazed:"mailbox"`
    Rebuild    bool   `glazed:"rebuild"`
    DryRun     bool   `glazed:"dry-run"`
}
```

And the call site now wraps them back:

```go
// AFTER
report, err := enrichpkg.RunAll(ctx, settings.SQLitePath, toOptions(enrichSettings{
    SQLitePath: settings.SQLitePath,
    AccountKey: settings.AccountKey,
    Mailbox:    settings.Mailbox,
    Rebuild:    settings.Rebuild,
    DryRun:     settings.DryRun,
}))
```

### Why this happened

The Glazed CLI framework's `DecodeSectionInto` does not correctly handle embedded structs when decoding from a `values.Values` map. The embedded struct's fields are not visible at the right section path. Flattening the fields makes them directly accessible to the decoder.

### ⚠️ Issue: This is a maintenance hazard

The five fields are now duplicated across four command files. If `enrichSettings` gains a new field, it must be added to **all four** settings structs AND all four call sites. This is error-prone.

**Recommendation:** Instead of duplicating the fields, fix the Glazed decoder to support embedded structs, or use a helper function:

```go
// helper that extracts shared fields from any struct that has them
func enrichOptsFromFlags(sqlitePath, accountKey, mailbox string, rebuild, dryRun bool) enrichSettings {
    return enrichSettings{
        SQLitePath: sqlitePath,
        AccountKey: accountKey,
        Mailbox:    mailbox,
        Rebuild:    rebuild,
        DryRun:     dryRun,
    }
}
```

This at least reduces the duplication to field names (not the full struct-to-struct mapping).

## 6. Issues Summary — Categorized by Severity

### 🔴 Must Fix (Before Merge)

| # | Location | Issue | Why it matters |
|---|----------|-------|----------------|
| M1 | `pkg/annotate/types.go` | `UpdateFeedbackInput` only has `Status` but frontend `UpdateFeedbackRequest` also sends `bodyMarkdown` | Contract mismatch — client promises a feature the server ignores |
| M2 | `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx` | Dead code — never imported | Confusing for future developers; removes to reduce noise |
| M3 | `ui/src/store/annotationUiSlice.ts` | `commentDrawerOpen` + `filterMailbox` + their actions are dead Redux state | Bloats the slice, confuses future developers about what is actually used |

### 🟡 Should Fix (Near-Term)

| # | Location | Issue | Why it matters |
|---|----------|-------|----------------|
| S1 | `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx` | File named "Drawer" but renders a Dialog | Naming confusion — every developer will wonder where the drawer is |
| S2 | `pkg/annotate/repository_feedback.go` | Duplicated feedback INSERT SQL (tx and non-tx versions) | Maintenance burden — any column change must be made in two places |
| S3 | `pkg/annotate/repository_feedback.go` | N+1 query in `ListReviewFeedback` | Performance degrades linearly with feedback count |
| S4 | `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx` | `agentRunId` prop is accepted but never used inside the component | Misleading API surface |
| S5 | `ui/src/pages/GuidelineDetailPage.tsx` | Create-then-link is fire-and-forget; navigation happens before link completes | Race condition — guideline exists but is not linked if link fails |
| S6 | `cmd/smailnail/commands/enrich/*.go` | Shared settings fields duplicated across 4 files | Maintenance hazard — new field requires changes in 8+ places |
| S7 | `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx` | Always receives `runs={[]}` — no backend endpoint exists | Shows "No runs linked" permanently — misleading UX |
| S8 | `ui/src/components/Guidelines/GuidelineSummaryCard.tsx` | `linkedRunCount` is always 0 — no backend data source | Wasted rendering |
| S9 | `ui/src/pages/GuidelinesListPage.tsx` | Client-side `filtered` useMemo duplicates backend search filter | Redundant computation |

### 🟢 Nice to Have (Future)

| # | Location | Issue | Why it matters |
|---|----------|-------|----------------|
| N1 | `pkg/annotate/schema.go` | No migration versioning or rollback | Risk of double-application in production |
| N2 | `pkg/annotate/schema.go` | Missing index on `review_feedback_targets(feedback_id)` | Full table scan on per-feedback target lookup |
| N3 | `pkg/annotate/schema.go` | Full-text search uses `LIKE %term%` (no index usage) | Acceptable for small data; future FTS5 target |
| N4 | `ui/src/api/annotations.ts` | `getGuideline` provides coarse `["Guidelines"]` tag instead of per-ID | Any guideline update invalidates all guideline queries |
| N5 | `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx` | `useEffect` dependency array missing `resetForm` | React exhaustive-deps lint warning |
| N6 | `ui/src/components/Guidelines/GuidelineForm.tsx` | Slug field has no format validation | Users can create slugs with spaces/uppercase |
| N7 | `ui/src/mocks/handlers.ts` | Mock mutations not persisted (except run-guideline links) | Create-then-list flow broken in Storybook |
| N8 | `ui/src/pages/stories/GuidelineDetailPage.stories.tsx` | Storybook MSW override uses wrong URL path (`/api/guidelines` vs `/api/review-guidelines`) | Dead override code |
| N9 | `pkg/annotationui/handlers_feedback.go` | `CreatedBy` / `linkedBy` always empty (no auth) | Will need wiring when authentication is added |
| N10 | `ui/src/components/AnnotationTable/AnnotationTable.tsx` | Memo comparator compares callback identity — requires parent to use `useCallback` | Undocumented constraint; breaking it silently degrades performance |
| N11 | `ui/src/components/ReviewFeedback/GuidelinePicker.tsx` | No loading/error state | Silent empty render on fetch failure |
| N12 | `ui/src/components/ReviewFeedback/FeedbackCard.tsx` | Acknowledge/Resolve call mutation directly | Harder to reuse card in non-standard contexts |

## 7. API Reference — Complete Endpoint Map

All new and modified HTTP endpoints:

### Modified Endpoints

```
PATCH /api/annotations/{id}/review
  Before: { reviewState: string }
  After:  { reviewState: string, comment?: { feedbackKind, title, bodyMarkdown }, guidelineIds?: string[], mailboxName?: string }
  Returns: Annotation (updated)
  Backend: handleReviewAnnotation → ReviewAnnotationWithArtifacts (transactional)

POST /api/annotations/batch-review
  Before: { ids: string[], reviewState: string }
  After:  { ids: string[], reviewState: string, agentRunId?: string, comment?: {...}, guidelineIds?: string[], mailboxName?: string }
  Returns: 204 No Content
  Backend: handleBatchReview → BatchReviewWithArtifacts (transactional)
```

### New Feedback Endpoints

```
GET    /api/review-feedback            → ReviewFeedback[]
  Query params: agentRunId, status, feedbackKind, mailboxName, limit
  Backend: handleListFeedback → ListReviewFeedback

GET    /api/review-feedback/{id}       → ReviewFeedback
  Backend: handleGetFeedback → GetReviewFeedback

POST   /api/review-feedback            → ReviewFeedback (201)
  Body: { scopeKind, agentRunId?, mailboxName?, feedbackKind, title, bodyMarkdown, targets? }
  Backend: handleCreateFeedback → CreateReviewFeedback

PATCH  /api/review-feedback/{id}       → ReviewFeedback
  Body: { status }  (NOTE: bodyMarkdown is accepted but ignored by backend)
  Backend: handleUpdateFeedback → UpdateReviewFeedback
```

### New Guideline Endpoints

```
GET    /api/review-guidelines          → ReviewGuideline[]
  Query params: status, scopeKind, search, limit
  Backend: handleListGuidelines → ListGuidelines

GET    /api/review-guidelines/{id}     → ReviewGuideline
  Backend: handleGetGuideline → GetGuideline

POST   /api/review-guidelines          → ReviewGuideline (201)
  Body: { slug, title, scopeKind, bodyMarkdown }
  Backend: handleCreateGuideline → CreateGuideline

PATCH  /api/review-guidelines/{id}     → ReviewGuideline
  Body: { title?, scopeKind?, status?, priority?, bodyMarkdown? }
  Backend: handleUpdateGuideline → UpdateGuideline
```

### New Run-Guideline Link Endpoints

```
GET    /api/annotation-runs/{id}/guidelines          → ReviewGuideline[]
  Backend: handleListRunGuidelines → ListRunGuidelines

POST   /api/annotation-runs/{id}/guidelines          → ReviewGuideline[] (200, NOT 204)
  Body: { guidelineId }
  Backend: handleLinkRunGuideline → LinkGuidelineToRun + ListRunGuidelines

DELETE /api/annotation-runs/{id}/guidelines/{guidelineId} → 204
  Backend: handleUnlinkRunGuideline → UnlinkGuidelineFromRun
```

### Modified Sender Endpoint (mailbox exposure)

```
GET /api/mirror/senders/{email}
  Response change: MessagePreview now includes mailboxName field
  SQL change: added mailbox_name to SELECT
```

## 8. SQLite Schema Reference — New Tables

### review_feedback

```sql
CREATE TABLE review_feedback (
    id TEXT PRIMARY KEY,                    -- UUID
    scope_kind TEXT NOT NULL DEFAULT 'selection',  -- annotation|selection|run|guideline
    agent_run_id TEXT NOT NULL DEFAULT '',
    mailbox_name TEXT NOT NULL DEFAULT '',
    feedback_kind TEXT NOT NULL DEFAULT 'comment',  -- comment|reject_request|guideline_request|clarification
    status TEXT NOT NULL DEFAULT 'open',    -- open|acknowledged|resolved|archived
    title TEXT NOT NULL DEFAULT '',
    body_markdown TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Indexes: (agent_run_id, created_at), (status, created_at)
```

### review_feedback_targets

```sql
CREATE TABLE review_feedback_targets (
    feedback_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    PRIMARY KEY (feedback_id, target_type, target_id)
);
-- Index: (target_type, target_id)
-- MISSING: (feedback_id) -- see issue in section 3.1
```

### review_guidelines

```sql
CREATE TABLE review_guidelines (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    scope_kind TEXT NOT NULL DEFAULT 'global',  -- global|mailbox|sender|domain|workflow
    status TEXT NOT NULL DEFAULT 'active',       -- active|archived|draft
    priority INTEGER NOT NULL DEFAULT 0,
    body_markdown TEXT NOT NULL DEFAULT '',
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Indexes: (status, priority DESC), (slug)
```

### run_guideline_links

```sql
CREATE TABLE run_guideline_links (
    agent_run_id TEXT NOT NULL,
    guideline_id TEXT NOT NULL,
    linked_by TEXT NOT NULL DEFAULT '',
    linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (agent_run_id, guideline_id)
);
-- No additional indexes (PK covers lookups by run+guideline)
```

## 9. Key Pseudocode — Transactional Review

This is the most critical code path in the feature. Here is the transactional review pseudocode explained for an intern:

```go
func ReviewAnnotationWithArtifacts(input) {
    // 1. BEGIN TRANSACTION
    tx = db.Begin()
    defer tx.Rollback()  // no-op after commit

    // 2. UPDATE the annotation's review state
    //    "to_review" → "reviewed" or "dismissed"
    tx.Exec("UPDATE annotations SET review_state = ? WHERE id = ?",
        input.ReviewState, input.AnnotationID)

    // 3. READ the annotation to get its agent_run_id
    //    (needed to link guidelines to the correct run)
    annotation = tx.Query("SELECT * FROM annotations WHERE id = ?",
        input.AnnotationID)

    // 4. IF the reviewer left a comment, CREATE a feedback record
    if input.Comment != nil && input.Comment.BodyMarkdown != "" {
        feedbackID = newUUID()
        tx.Exec(`INSERT INTO review_feedback (id, scope_kind, agent_run_id,
                 mailbox_name, feedback_kind, status, title, body_markdown)
                 VALUES (?, 'annotation', ?, ?, ?, 'open', ?, ?)`,
            feedbackID, annotation.AgentRunID, input.MailboxName,
            input.Comment.FeedbackKind, input.Comment.Title,
            input.Comment.BodyMarkdown)

        // 4b. Link the feedback to the annotation target
        tx.Exec(`INSERT INTO review_feedback_targets
                 (feedback_id, target_type, target_id) VALUES (?, 'annotation', ?)`,
            feedbackID, annotation.ID)
    }

    // 5. FOR each guideline the reviewer attached, LINK it to the run
    for _, guidelineID := range input.GuidelineIDs {
        if annotation.AgentRunID == "" {
            return ERROR("cannot link guideline: annotation has no run")
        }
        tx.Exec(`INSERT INTO run_guideline_links
                 (agent_run_id, guideline_id, linked_by)
                 VALUES (?, ?, ?)
                 ON CONFLICT DO NOTHING`,
            annotation.AgentRunID, guidelineID, input.CreatedBy)
    }

    // 6. COMMIT — all or nothing
    tx.Commit()

    // 7. Return the updated annotation
    return repository.GetAnnotation(input.AnnotationID)
}
```

**Why this transaction matters:** Without it, a partial failure would leave the annotation in "dismissed" state but with no feedback record explaining why. The reviewer would think they left feedback, but the agent would have no record of it. The transaction ensures that either everything happens (state change + feedback + guideline links) or nothing happens (annotation stays in its original state).

**The batch version is similar but:**
- Updates multiple annotations at once with `WHERE id IN (?)`
- Creates one feedback record targeting all selected annotations
- Requires either an explicit `agentRunId` or infers one from the selected annotations (refuses if they span multiple runs)

## 10. Naming Issues & Confusing Code

This section catalogs every naming inconsistency, confusing pattern, or misleading artifact found during review.

### 10.1 "Drawer" vs "Dialog" Naming

| File / Export | What it actually renders | What the name says |
|---------------|------------------------|-------------------|
| `ReviewCommentDrawer.tsx` | `<Dialog>` | "Drawer" |
| `data-part: "comment-drawer"` | — | "drawer" |
| `ReviewCommentInline.tsx` | `<Collapse>` inline form | "Inline" (accurate, but dead code) |
| `ReviewCommentInlineProps` | — | Uses "Inline" suffix correctly |

**Recommendation:** Rename `ReviewCommentDrawer` → `ReviewCommentDialog`. Update:
- File name
- Component name
- Export name
- `data-part` value in `parts.ts`
- All 6 import sites across pages
- Storybook story titles

### 10.2 "List" vs "Get" Endpoint Naming

The API follows a consistent pattern:
- `listGuidelines` (plural) → returns array
- `getGuideline` (singular) → returns single item

But in the Go handler names:
- `handleListGuidelines` → correct
- `handleGetGuideline` → correct
- `handleListRunGuidelines` → correct (returns array from a join)
- `handleLinkRunGuideline` → singular (links one guideline)

This is actually consistent. No issue here.

### 10.3 `scopeKind` vs `scope_kind` vs `ScopeKind`

Same concept, three naming conventions:
- **Go domain types:** `ScopeKind` (PascalCase struct field)
- **Go DB columns:** `scope_kind` (snake_case via `db` tag)
- **Go JSON responses:** `scopeKind` (camelCase via `json` tag)
- **TypeScript types:** `scopeKind` (camelCase, matches JSON)
- **SQL DDL:** `scope_kind` (snake_case column name)
- **HTTP query params:** `scopeKind` (camelCase, matches JSON)

This is the standard Go+TypeScript convention and is correct. No issue.

### 10.4 `FeedbackScopeSelection` vs `"selection"`

The Go constants use long names:
```go
FeedbackScopeSelection  = "selection"
FeedbackScopeAnnotation = "annotation"
```

But the TypeScript types use string unions:
```ts
type FeedbackScopeKind = "annotation" | "selection" | "run" | "guideline";
```

The TypeScript approach is idiomatic for TS. The Go approach is idiomatic for Go. They agree on the actual string values. No issue.

### 10.5 `ReviewCommentInput` appears in both Go packages

- `annotate.ReviewCommentInput` — domain layer (in `pkg/annotate/types.go`)
- `reviewCommentInput` — HTTP DTO layer (in `pkg/annotationui/types_feedback.go`)

They have the same fields (`feedbackKind`, `title`, `bodyMarkdown`) but different Go types. The `toAnnotateReviewComment` function converts between them. This is correct layered architecture — the HTTP layer has its own types and converts to/from domain types. No issue, but a new developer might be confused by the name similarity.

### 10.6 Misleading `linkedRunCount` Prop

```ts
<GuidelineSummaryCard linkedRunCount={0} ... />
```

This prop suggests the component shows how many runs use this guideline, but it is always 0 because the data is unavailable. The component renders "0 runs linked" for every guideline, which is misleading. **Recommendation:** Either hide the count when data is unavailable, or remove the prop entirely.

### 10.7 `GuidelineScopeBadge` Handles More Values Than The Type Allows

The TypeScript `GuidelineScopeKind` type defines 5 values: `"global" | "mailbox" | "sender" | "domain" | "workflow"`. The Go `GuidelineScope*` constants define the same 5 values. But the `GuidelineScopeBadge` component has icon mappings for additional values (like `"run"`) that do not exist in either type system. This is defensive but potentially confusing.

## 11. Deprecated, Unused & Dead Code

### 11.1 Dead Code

| File | Component | Status | Evidence |
|------|-----------|--------|----------|
| `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx` | `ReviewCommentInline` | **Never imported** | `grep -r "ReviewCommentInline" ui/src/` returns only the file itself and its stories |
| `ui/src/store/annotationUiSlice.ts` | `commentDrawerOpen` state | **Never read** | Pages use local `useState` |
| `ui/src/store/annotationUiSlice.ts` | `openCommentDrawer` action | **Never dispatched** | `grep -r "openCommentDrawer" ui/src/` returns only the export |
| `ui/src/store/annotationUiSlice.ts` | `closeCommentDrawer` action | **Never dispatched** | Same as above |
| `ui/src/store/annotationUiSlice.ts` | `filterMailbox` state | **Never read** | No component accesses it |
| `ui/src/store/annotationUiSlice.ts` | `setFilterMailbox` action | **Never dispatched** | Same as above |

### 11.2 Half-Implemented Features

| Feature | What exists | What is missing |
|---------|-------------|-----------------|
| Guideline linked runs display | `GuidelineLinkedRuns` component | Backend endpoint `GET /guidelines/{id}/runs` |
| Guideline linked run count | `linkedRunCount` prop on `GuidelineSummaryCard` | Backend data source |
| Feedback body editing | `bodyMarkdown` in `UpdateFeedbackRequest` (TypeScript) | Backend `UpdateFeedbackInput` ignores it |
| Mailbox filter pills in review queue | `filterMailbox` Redux state | UI filter component |
| `feedbackKind: "guideline_request"` | Value exists in both type systems | No special rendering or auto-guideline-creation flow |
| `scopeKind: "guideline"` | Value exists in both type systems | No UI creates guideline-scoped feedback |
| `scopeKind: "domain"` on guidelines | Value exists in both type systems | No guideline is scoped to a domain |

### 11.3 Deprecated Patterns

| Pattern | Where | What replaced it |
|---------|-------|-----------------|
| `enrichSettings` embedding | `cmd/smailnail/commands/enrich/*.go` | Flattened field declarations |
| `setExpandedId` with inline toggle | `ReviewQueuePage.tsx` | `toggleExpandedId` action |

The `enrichSettings` embedding was removed because the Glazed `DecodeSectionInto` function could not handle embedded structs. The old pattern is visible in the git history but is no longer present in the codebase. This is not dead code — it is removed code. Mentioned for historical context.

## 12. Performance Review

### 12.1 Backend Performance

| Operation | Complexity | Issue |
|-----------|-----------|-------|
| `CreateReviewFeedback` | O(T) where T = target count | ✅ Single transaction, one INSERT per target |
| `ListReviewFeedback` | O(F × T) — N+1 targets | ⚠️ Issues one query per feedback to load targets |
| `ReviewAnnotationWithArtifacts` | O(G) where G = guideline count | ✅ Single transaction |
| `BatchReviewWithArtifacts` | O(N + T + G) | ✅ Single transaction |
| `ListGuidelines` with search | O(N) full table scan | ⚠️ LIKE %term% prevents index use |
| `ListRunGuidelines` | O(G log G) | ✅ JOIN + ORDER BY, indexed |

**The N+1 in `ListReviewFeedback` is the most impactful backend performance issue.** For 50 feedback items with 3 targets each, this is 51 SQL queries. The fix is a batch target loader:

```sql
-- Instead of N queries:
SELECT * FROM review_feedback_targets WHERE feedback_id = ?
-- Use 1 query:
SELECT * FROM review_feedback_targets WHERE feedback_id IN (?, ?, ?, ...)
```

Then group results in Go using a map.

### 12.2 Frontend Performance

| Component | Issue | Severity |
|-----------|-------|----------|
| `AnnotationTable` | ✅ Memoized via `React.memo` with custom comparator | Resolved in this branch |
| `AnnotationRow` | Not individually memoized | 🟢 Minor — lightweight render |
| `GuidelinesListPage` | Double filtering (server + client) | 🟡 Redundant work |
| `GuidelineLinkPicker` | Client-side search over full guideline list | 🟢 Fine for small datasets |
| `ReviewCommentDrawer` | Form resets via `useEffect` on every open | ✅ Correct pattern |
| `RunDetailPage` | 3 RTK queries on mount (run + guidelines + feedback) | 🟢 Acceptable — parallel |
| `SenderDetailPage` | `senderMailboxName` recomputed via `useMemo` on every render | ✅ Memoized correctly |

**The annotation table memoization work (Steps 16-17) is the most impactful frontend performance improvement.** Before this change, selecting a checkbox triggered a full table rerender with related-annotation computation for every row. After the change, only the changed row rerenders.

## 13. Testing Gaps

### 13.1 Backend Tests — What is Missing

| Area | What should be tested | Current state |
|------|----------------------|---------------|
| `ReviewAnnotationWithArtifacts` rollback | Feedback creation fails after review-state update | **Not tested** |
| `BatchReviewWithArtifacts` rollback | Guideline linking fails after feedback creation | **Not tested** |
| Multi-run batch guideline linking | Batch with annotations from 2+ runs and guidelineIds | **Not tested** (documented in diary) |
| `UpdateReviewFeedback` with bodyMarkdown | Body update is silently ignored | **Not tested** (not implemented) |
| `CreateGuideline` slug uniqueness | Duplicate slug returns 409 | **Not tested** |
| `ListReviewFeedback` with all filters | AgentRunID + status + feedbackKind + mailboxName | **Not tested** |

### 13.2 Frontend Tests — What is Missing

| Area | What should be tested | Current state |
|------|----------------------|---------------|
| `ReviewCommentDrawer` form submission | Correct payload shape | Storybook only (no unit test) |
| `ReviewQueuePage` batch reject-explain | Drawer opens, submits, closes | Not tested |
| `GuidelineDetailPage` create-then-link | Navigation after create with runId | Not tested |
| `AnnotationTable` memoization | Selection does not trigger full rerender | Not tested (manual verification only) |
| RTK Query cache invalidation | Review with comment triggers feedback refetch | Not tested |

### 13.3 What IS Tested

- All new components have Storybook stories (visual testing)
- MSW handlers exercise the full CRUD lifecycle in Storybook
- Backend compiles and lints clean (`go test`, `make lint`)
- TypeScript compiles clean (`tsc --noEmit`)

**The testing strategy relies on Storybook for component verification and manual testing for integration flows.** This is acceptable for an MVP but should be supplemented with automated integration tests before the feature reaches production.

## 14. Security Considerations

1. **No authentication on the annotation UI server.** All `created_by` and `linked_by` fields are empty strings. Anyone with network access can create feedback, modify guidelines, or change review state. **Mitigation:** The server is designed to run locally (localhost-only) and is separate from the production `smailnaild`. When this is deployed for multi-user use, authentication must be added.

2. **SQL injection is prevented by parameterized queries.** All repository methods use `?` placeholders with `sqlx` binding. The `LIKE` pattern in `ListGuidelines` uses parameterized `"%"+search+"%"`. No string concatenation is used in SQL.

3. **No rate limiting on API endpoints.** The annotation UI server has no rate limiting. A malicious or buggy client could flood the feedback or guideline endpoints. **Mitigation:** Acceptable for localhost-only use.

4. **`DisallowUnknownFields` in JSON decoder prevents unexpected fields.** This is a positive security measure — it prevents clients from injecting extra JSON fields that might confuse the application.

5. **No input length limits.** The `title` and `bodyMarkdown` fields accept arbitrarily long strings. A malicious client could create multi-megabyte feedback records. **Mitigation:** For localhost-only use this is low risk. **Recommendation:** Add `len(title) > 500` and `len(bodyMarkdown) > 10000` validations.

## 15. Architecture Assessment

### 15.1 What the Codebase Does Well

- **Layered architecture with clear boundaries.** Domain types (`pkg/annotate/types.go`) are independent of HTTP DTOs (`pkg/annotationui/types_feedback.go`). Repository methods accept domain types and return domain types. Handlers translate between domain types and HTTP types. This three-layer pattern (handler → domain → storage) is consistently applied.

- **Transactional correctness.** The `ReviewAnnotationWithArtifacts` and `BatchReviewWithArtifacts` methods are the most important contribution of this branch. They correctly wrap multi-step operations in SQL transactions, ensuring that review-state changes, feedback creation, and guideline linking are atomic. The diary records that this was specifically a recovery from an earlier broken attempt that used best-effort side effects — the correction was the right call.

- **Consistent RTK Query patterns.** Every new endpoint follows the same pattern: type imports → query/mutation definition → cache tags → hook exports. The `providesTags` / `invalidatesTags` strategy is well-thought-out: review actions invalidate `["Annotations", "Runs", "Feedback"]` because they touch all three.

- **MSW mock coverage.** Every new endpoint has a mock handler, and every new component has Storybook stories. This means a new developer can explore the entire feature visually without running the backend.

- **Incremental delivery.** The diary records 18 steps with clean compile checks after each one. The phased approach (types → API → mocks → badges → widgets → pages → integration → performance → UX polish) minimized rework.

### 15.2 What Could Be Better

- **Testing discipline.** There are no automated tests beyond compilation checks. The diary mentions "validate with `go test` and `make lint`" but no test cases were written for the new repository methods or handlers. For a transaction-critical code path, this is a gap. **Recommendation:** At minimum, add tests for `ReviewAnnotationWithArtifacts` rollback scenarios.

- **Dead code accumulation.** The Redux slice has unused state. The `ReviewCommentInline` component is never imported. The `GuidelineLinkedRuns` component always renders empty. Each was created with good intentions (future-proofing, design iteration) but the net effect is noise. **Recommendation:** Delete unused code immediately. Git history preserves the intent.

- **Duplicated code patterns.** The feedback INSERT SQL is duplicated in transactional and non-transactional methods. The enrich settings fields are duplicated across four command files. The GuidelinesListPage does both server-side and client-side search. Each duplication has a pragmatic reason (the code works, the deadline was met) but each will cause maintenance headaches.

- **Naming accuracy.** The "Drawer" that is actually a Dialog is the most visible example, but the pattern extends to types and state fields that suggest features that do not exist (`filterMailbox`, `bodyMarkdown` in update types).

## 16. Recommended Next Steps (Priority Order)

1. **Fix M1: Align `UpdateFeedbackInput`** — Either add `BodyMarkdown *string` to the Go type and wire it through the repository, or remove `bodyMarkdown` from the frontend type. This is a contract mismatch that will confuse API consumers.

2. **Fix M2: Delete `ReviewCommentInline.tsx`** — It is dead code. Delete the file, its Storybook stories, and its barrel export.

3. **Fix M3: Clean up Redux slice** — Remove `commentDrawerOpen`, `openCommentDrawer`, `closeCommentDrawer`, `filterMailbox`, `setFilterMailbox`. These are unused.

4. **Fix S1: Rename `ReviewCommentDrawer` → `ReviewCommentDialog`** — Update file, component, exports, imports, parts.ts, and stories.

5. **Fix S2: Extract shared feedback INSERT SQL** — Create `createReviewFeedbackCore(ctx, executor, input) (string, error)` that accepts a common interface (`sqlx.Ext`) and is called by both the transactional and non-transactional methods.

6. **Fix S3: Batch-load feedback targets** — Replace the N+1 `listFeedbackTargets` loop in `ListReviewFeedback` with a single `WHERE feedback_id IN (?)` query.

7. **Fix S5: Await guideline link in `GuidelineDetailPage.handleSave`** — The fire-and-forget `void linkGuidelineToRun(...)` should be awaited with error handling.

8. **Add integration tests** — At minimum: test `ReviewAnnotationWithArtifacts` success path, rollback path, and `BatchReviewWithArtifacts` multi-run rejection.

9. **Remove client-side search duplication** — `GuidelinesListPage` should rely on the backend `search` parameter and not double-filter.

10. **Add `CreatedBy` population when auth is available** — Track this as a follow-up ticket.

## 17. References

### Design Documents (in the SMN-20260403-RUN-REVIEW ticket)

- `design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md` — Full implementation guide
- `design/02-ui-design-review-feedback-guidelines-mailbox.md` — UI wireframes and widget DSL
- `reference/02-diary.md` — Implementation diary (18 steps)

### Key Files Reviewed

| Layer | File | Lines Changed |
|-------|------|--------------|
| Schema | `pkg/annotate/schema.go` | +62 |
| Types | `pkg/annotate/types.go` | +152 |
| Repository | `pkg/annotate/repository_feedback.go` | +622 |
| Handlers | `pkg/annotationui/handlers_annotations.go` | +44 (modified) |
| Handlers | `pkg/annotationui/handlers_feedback.go` | +270 |
| DTOs | `pkg/annotationui/types_feedback.go` | +145 |
| Routes | `pkg/annotationui/server.go` | +17 |
| Sender | `pkg/annotationui/handlers_senders.go` | +3 |
| API | `ui/src/api/annotations.ts` | +140 |
| Types | `ui/src/types/reviewFeedback.ts` | +78 |
| Types | `ui/src/types/reviewGuideline.ts` | +58 |
| Store | `ui/src/store/annotationUiSlice.ts` | +21 |
| Table | `ui/src/components/AnnotationTable/AnnotationTable.tsx` | +144 |
| Row | `ui/src/components/AnnotationTable/AnnotationRow.tsx` | +13 |
| Dialog | `ui/src/components/ReviewFeedback/ReviewCommentDrawer.tsx` | +212 |
| Inline | `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx` | +156 (dead) |
| Card | `ui/src/components/ReviewFeedback/FeedbackCard.tsx` | +104 |
| Picker | `ui/src/components/ReviewFeedback/GuidelinePicker.tsx` | +71 |
| LinkPicker | `ui/src/components/ReviewFeedback/GuidelineLinkPicker.tsx` | +156 |
| Section | `ui/src/components/ReviewFeedback/RunFeedbackSection.tsx` | +105 |
| RunGuide | `ui/src/components/RunGuideline/RunGuidelineSection.tsx` | +102 |
| RunCard | `ui/src/components/RunGuideline/GuidelineCard.tsx` | +86 |
| Form | `ui/src/components/Guidelines/GuidelineForm.tsx` | +212 |
| Summary | `ui/src/components/Guidelines/GuidelineSummaryCard.tsx` | +133 |
| Linked | `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx` | +82 |
| Badges | `ui/src/components/shared/` (4 files) | +176 |
| Pages | `ui/src/pages/` (5 files modified/created) | +578 |
| Mocks | `ui/src/mocks/` (2 files) | +310 |
| Enrich | `cmd/smailnail/commands/enrich/` (4 files) | +62 |

**Total: ~15,879 lines added, ~793 lines removed across 80 files.**
