---
Title: 'Phased Implementation Tasks — Frontend + Backend'
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - annotations
    - frontend
    - backend
    - sqlite
    - react
    - storybook
    - msw
DocType: tasks
Intent: long-term
Owners:
    - manuel
Summary: Complete phased task breakdown for implementing review feedback, guidelines, and mailbox-aware context. Each task is atomic, testable, and references the design doc.
LastUpdated: 2026-04-04T16:00:00-04:00
WhatFor: Use as the execution checklist. Check off tasks as you go.
WhenToUse: Open this before starting any implementation work on this ticket.
---

# Phased Implementation Tasks

> **Reference docs:**
> - Design guide: `design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md`
> - UI design: `design/02-ui-design-review-feedback-guidelines-mailbox.md`
> - System map: `reference/01-current-system-map.md`
>
> **Conventions:**
> - Every new component gets a `parts.ts`, `data-widget`/`data-part` attributes, and Storybook stories.
> - Every new API endpoint gets MSW handlers and mock data.
> - Stories cover: default, empty, loading, error, and variant states.
> - Components are presentational; state comes from RTK Query + Redux slice.
> - `parts.ts` is the single source of truth for `data-part` names per widget directory.

---

## Phase 0 — Prerequisites & Orientation

- [ ] 0.1 Read `design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md` end-to-end
- [ ] 0.2 Read `design/02-ui-design-review-feedback-guidelines-mailbox.md` end-to-end
- [ ] 0.3 Read `reference/01-current-system-map.md` and open each listed file in your editor
- [ ] 0.4 Run `storybook dev` and confirm all existing stories pass with no errors
- [ ] 0.5 Run `tsc --noEmit` and confirm the project compiles cleanly

---

## Phase 1 — Schema, Go Types & Repository (Backend)

### 1A. Schema migrations

- [ ] 1A.1 Add `review_feedback` table to `pkg/annotate/schema.go`
  - Fields: `id`, `scope_kind`, `agent_run_id`, `mailbox_name`, `feedback_kind`, `status`, `title`, `body_markdown`, `created_by`, `created_at`, `updated_at`
  - Add to `CreateTables()` or a new migration function
  - Write test: table created, insert + select round-trip

- [ ] 1A.2 Add `review_feedback_targets` table to `pkg/annotate/schema.go`
  - Fields: `feedback_id`, `target_type`, `target_id` (composite PK)
  - Write test: insert targets, query by feedback_id

- [ ] 1A.3 Add `review_guidelines` table to `pkg/annotate/schema.go`
  - Fields: `id`, `slug` (unique), `title`, `scope_kind`, `status`, `priority`, `body_markdown`, `created_by`, `created_at`, `updated_at`
  - Write test: table created, insert with duplicate slug fails

- [ ] 1A.4 Add `run_guideline_links` table to `pkg/annotate/schema.go`
  - Fields: `agent_run_id`, `guideline_id`, `linked_by`, `linked_at` (composite PK on run+guideline)
  - Write test: insert link, duplicate insert ignored or errors

### 1B. Go types

- [ ] 1B.1 Add `ReviewFeedback`, `FeedbackTarget`, `CreateFeedbackInput` structs to `pkg/annotate/types.go`
  - Scope enums: `ScopeAnnotation`, `ScopeSelection`, `ScopeRun`, `ScopeGuideline`
  - Kind enums: `FeedbackComment`, `FeedbackRejectRequest`, `FeedbackGuidelineRequest`, `FeedbackClarification`
  - Status enums: `StatusOpen`, `StatusAcknowledged`, `StatusResolved`, `StatusArchived`

- [ ] 1B.2 Add `ReviewGuideline`, `CreateGuidelineInput`, `UpdateGuidelineInput` structs to `pkg/annotate/types.go`
  - Scope enums: `ScopeGlobal`, `ScopeMailbox`, `ScopeSender`, `ScopeDomain`, `ScopeWorkflow`
  - Status enums: `GuidelineActive`, `GuidelineArchived`, `GuidelineDraft`

- [ ] 1B.3 Add `RunGuidelineLink` struct to `pkg/annotate/types.go`

### 1C. Repository methods

- [ ] 1C.1 `CreateReviewFeedback(ctx, input) → (ReviewFeedback, error)`
  - Insert into `review_feedback`
  - Insert targets into `review_feedback_targets`
  - Return complete record
  - Test: create + get round-trip, targets correct

- [ ] 1C.2 `GetReviewFeedback(ctx, id) → (ReviewFeedback, error)`
  - Fetch feedback + join targets
  - Test: found, not-found

- [ ] 1C.3 `ListReviewFeedback(ctx, filter) → ([]ReviewFeedback, error)`
  - Filter by: `agentRunId`, `status`, `feedbackKind`, `mailboxName`
  - Test: filter by run, filter by status

- [ ] 1C.4 `UpdateReviewFeedbackStatus(ctx, id, status) → error`
  - Update status + `updated_at`
  - Test: open→acknowledged, acknowledged→resolved

- [ ] 1C.5 `CreateGuideline(ctx, input) → (ReviewGuideline, error)`
  - Insert, enforce unique slug
  - Test: create, duplicate-slug error

- [ ] 1C.6 `GetGuideline(ctx, id) → (ReviewGuideline, error)`
  - Test: found, not-found

- [ ] 1C.7 `GetGuidelineBySlug(ctx, slug) → (ReviewGuideline, error)`
  - Test: found by slug

- [ ] 1C.8 `ListGuidelines(ctx, filter) → ([]ReviewGuideline, error)`
  - Filter by: `status`, `scopeKind`, `search` (title/slug/body LIKE)
  - Test: filter by status, search by title substring

- [ ] 1C.9 `UpdateGuideline(ctx, id, input) → (ReviewGuideline, error)`
  - Partial update: only set non-zero fields
  - Test: update title, update status to archived

- [ ] 1C.10 `LinkGuidelineToRun(ctx, runID, guidelineID, linkedBy) → error`
  - Insert into `run_guideline_links`
  - Test: link, duplicate link idempotent

- [ ] 1C.11 `UnlinkGuidelineFromRun(ctx, runID, guidelineID) → error`
  - Delete from `run_guideline_links`
  - Test: unlink, unlink non-existent no error

- [ ] 1C.12 `ListRunGuidelines(ctx, runID) → ([]ReviewGuideline, error)`
  - Join `run_guideline_links` → `review_guidelines`
  - Test: returns linked guidelines only

- [ ] 1C.13 `CountGuidelineRunLinks(ctx, guidelineID) → (int, error)`
  - Count how many runs link to a guideline
  - Test: zero, multiple

- [ ] 1C.14 Extend `BatchUpdateReviewState` to optionally create feedback + link guidelines in one transaction
  - Accept optional `Comment *ReviewCommentInput` and `GuidelineIDs []string`
  - If comment provided: create feedback + targets for each annotation ID
  - If guidelineIDs provided: link each to the run
  - Test: batch review with comment, batch review without comment (backward compat)

---

## Phase 2 — Backend HTTP Handlers

### 2A. Request/response types

- [ ] 2A.1 Add HTTP request/response structs to `pkg/annotationui/types.go` (or a new file)
  - `ReviewFeedbackResponse` (JSON)
  - `CreateFeedbackRequest` (JSON)
  - `UpdateFeedbackRequest` (JSON)
  - `ListFeedbackParams` (query params)
  - `GuidelineResponse` (JSON)
  - `CreateGuidelineRequest` (JSON)
  - `UpdateGuidelineRequest` (JSON)
  - `ListGuidelinesParams` (query params)
  - `LinkGuidelineRequest` (JSON)
  - `ReviewCommentInput` (embedded in extended review payloads)
  - `ExtendedBatchReviewRequest` (adds comment + guidelineIDs + agentRunID + mailboxName)
  - `ExtendedReviewRequest` (adds comment + guidelineIDs + mailboxName)

### 2B. Extend existing review handlers

- [ ] 2B.1 Extend `handleReviewAnnotation` to accept optional comment/guideline IDs
  - Parse `comment` and `guidelineIds` from request body
  - If present: create feedback + link guidelines in the same transaction
  - Keep backward compat: if fields missing, behave exactly as before
  - Test: review with comment, review without comment

- [ ] 2B.2 Extend `handleBatchReview` to accept optional comment/guideline IDs
  - Same pattern as 2B.1 but for batch
  - If comment: create one feedback record targeting all annotation IDs
  - Test: batch with comment creates one feedback with N targets

### 2C. New feedback endpoints

- [ ] 2C.1 `GET /api/review-feedback` — list feedback with filters
  - Query params: `agentRunId`, `status`, `feedbackKind`
  - Returns `[]ReviewFeedbackResponse`
  - Test: empty list, filtered list

- [ ] 2C.2 `POST /api/review-feedback` — create standalone feedback
  - For run-level and general feedback (not tied to a review action)
  - Returns `ReviewFeedbackResponse` with 201
  - Test: create, validation errors

- [ ] 2C.3 `GET /api/review-feedback/:id` — get single feedback
  - Returns `ReviewFeedbackResponse` with targets
  - Test: found, not found

- [ ] 2C.4 `PATCH /api/review-feedback/:id` — update feedback status
  - Accepts `{ status: "acknowledged" | "resolved" | "archived" }`
  - Returns updated `ReviewFeedbackResponse`
  - Test: status transitions

### 2D. New guideline endpoints

- [ ] 2D.1 `GET /api/review-guidelines` — list guidelines
  - Query params: `status`, `scopeKind`, `search`
  - Returns `[]GuidelineResponse`
  - Test: filter by status, search

- [ ] 2D.2 `POST /api/review-guidelines` — create guideline
  - Returns `GuidelineResponse` with 201
  - Test: create, duplicate slug → 409

- [ ] 2D.3 `GET /api/review-guidelines/:id` — get single guideline
  - Test: found, not found

- [ ] 2D.4 `PATCH /api/review-guidelines/:id` — update guideline
  - Partial update
  - Test: update title, update status

### 2E. Run-guideline link endpoints

- [ ] 2E.1 `GET /api/annotation-runs/:id/guidelines` — list guidelines linked to a run
  - Returns `[]GuidelineResponse`
  - Test: run with 0, 1, 3 guidelines

- [ ] 2E.2 `POST /api/annotation-runs/:id/guidelines` — link a guideline
  - Body: `{ guidelineId: string }`
  - Test: link, link already-linked (idempotent)

- [ ] 2E.3 `DELETE /api/annotation-runs/:id/guidelines/:guidelineId` — unlink
  - Returns 204
  - Test: unlink, unlink non-existent

### 2F. Mailbox in existing endpoints

- [ ] 2F.1 Add `mailboxName` to sender detail response (recent messages)
  - Extend `MessagePreview` JSON to include `mailboxName` from `messages.mailbox_name`
  - Test: response includes mailbox

- [ ] 2F.2 Audit all endpoints that return message previews and add `mailboxName`
  - Check: sender detail, any future message-list endpoints
  - Test: mailbox present in response

- [ ] 2F.3 Register all new routes in `pkg/annotationui/server.go`
  - Add route groups for feedback, guidelines, run-guideline links
  - Test: server starts, routes respond

### 2R. Backend recovery & hardening (added after first incomplete attempt)

- [x] 2R.1 Confirm Vite dev proxy already exists in `ui/vite.config.ts`
  - Verify `/api` and `/auth` proxy to backend target
  - Do not add duplicate proxy logic unless a real gap is found

- [x] 2R.2 Remove duplicate / unused Phase 2 request types and handlers
  - Keep one source of truth for extended review payloads
  - Remove dead `extendedReviewRequest`, `extendedBatchReviewRequest`, and any unused helper handlers/imports
  - Verify `make lint` passes with zero `unused` findings

- [x] 2R.3 Move review-state update + optional feedback creation + optional guideline linking into one repository transaction
  - Add a repository method for single-review actions with artifacts
  - Add a repository method for batch-review actions with artifacts
  - Stop ignoring errors from feedback creation and guideline linking
  - Verify partial failure cannot return success after only the review-state update succeeds

- [x] 2R.4 Rewire `handleReviewAnnotation` and `handleBatchReview` to call the transactional repository methods
  - Preserve backward compatibility when `comment`, `guidelineIds`, `agentRunId`, and `mailboxName` are absent
  - Return proper HTTP errors when artifact creation/linking fails

- [x] 2R.5 Keep standalone feedback/guideline/run-link endpoints, but validate them end-to-end
  - `GET/POST/PATCH /api/review-feedback`
  - `GET/POST/PATCH /api/review-guidelines`
  - `GET/POST/DELETE /api/annotation-runs/:id/guidelines`
  - Verify route registration matches frontend RTK Query paths exactly

- [x] 2R.6 Validate sender mailbox propagation after backend changes
  - `pkg/annotationui/types.go` → `MessagePreview.mailboxName`
  - `pkg/annotationui/handlers_senders.go` query selects `mailbox_name`
  - Verify response shape matches frontend `MessagePreview`

- [x] 2R.7 Run backend validation loop before commit
  - `gofmt -w pkg/annotate/*.go pkg/annotationui/*.go`
  - `go test -tags sqlite_fts5 ./...`
  - `make lint`
  - Only commit Phase 2 when all three succeed

---

## Phase 3 — TypeScript Types & RTK Query Contract

### 3A. New type files

- [ ] 3A.1 Create `ui/src/types/reviewFeedback.ts`
  ```
  ReviewFeedback, FeedbackTarget, CreateFeedbackRequest,
  ReviewCommentDraft, FeedbackKind, FeedbackStatus, FeedbackScopeKind
  ```
  - Export from `types/index.ts` or import directly
  - Verify: `tsc --noEmit` passes

- [ ] 3A.2 Create `ui/src/types/reviewGuideline.ts`
  ```
  ReviewGuideline, CreateGuidelineRequest, UpdateGuidelineRequest,
  GuidelineScopeKind, GuidelineStatus
  ```
  - Export from `types/index.ts` or import directly
  - Verify: `tsc --noEmit` passes

### 3B. Extend existing types

- [ ] 3B.1 Extend `MessagePreview` in `types/annotations.ts` — add `mailboxName: string`
- [ ] 3B.2 Extend `AnnotationFilter` — add `mailboxName?: string`, `feedbackStatus?: string`
- [ ] 3B.3 Verify `tsc --noEmit` still passes after extensions

### 3C. RTK Query endpoints

- [ ] 3C.1 Add new tag types: `"Feedback"`, `"Guidelines"` to `tagTypes` array
- [ ] 3C.2 Extend `reviewAnnotation` mutation payload:
  ```
  { id, reviewState, comment?, guidelineIds?, mailboxName? }
  ```
- [ ] 3C.3 Extend `batchReview` mutation payload:
  ```
  { ids, reviewState, comment?, guidelineIds?, agentRunId?, mailboxName? }
  ```
- [ ] 3C.4 Add `listReviewFeedback` query endpoint → `ReviewFeedback[]`
- [ ] 3C.5 Add `getReviewFeedback` query endpoint → `ReviewFeedback`
- [ ] 3C.6 Add `createReviewFeedback` mutation endpoint → `ReviewFeedback`
- [ ] 3C.7 Add `updateReviewFeedback` mutation endpoint → `ReviewFeedback`
- [ ] 3C.8 Add `listGuidelines` query endpoint → `ReviewGuideline[]`
- [ ] 3C.9 Add `getGuideline` query endpoint → `ReviewGuideline`
- [ ] 3C.10 Add `createGuideline` mutation endpoint → `ReviewGuideline`
- [ ] 3C.11 Add `updateGuideline` mutation endpoint → `ReviewGuideline`
- [ ] 3C.12 Add `getRunGuidelines` query endpoint → `ReviewGuideline[]`
- [ ] 3C.13 Add `linkGuidelineToRun` mutation endpoint → void
- [ ] 3C.14 Add `unlinkGuidelineFromRun` mutation endpoint → void
- [ ] 3C.15 Export all new hooks from `api/annotations.ts`
- [ ] 3C.16 Verify: `tsc --noEmit` passes with all new endpoints

---

## Phase 4 — MSW Mock Data & Handlers

### 4A. Mock data

- [ ] 4A.1 Add mock feedback data to `mocks/annotations.ts` (or a new `mocks/feedback.ts`):
  ```
  mockFeedback: ReviewFeedback[] — 4 items:
    - open reject_request targeting 3 annotations on run-42
    - resolved comment on run-42
    - acknowledged clarification on run-41
    - open guideline_request on run-41
  ```
- [ ] 4A.2 Add mock guideline data to `mocks/annotations.ts` (or new `mocks/guidelines.ts`):
  ```
  mockGuidelines: ReviewGuideline[] — 4 items:
    - transactional-vs-promotional (workflow, active, pri 50)
    - billing-mail-classification (global, active, pri 30)
    - sender-domain-normalization (sender, draft, pri 0)
    - newsletter-vs-circular (mailbox, archived, pri 20)
  ```
- [ ] 4A.3 Add `mailboxName` to existing `mockMessages` data

### 4B. MSW handlers

- [ ] 4B.1 Add feedback handlers to `mocks/handlers.ts`:
  - `GET /api/review-feedback` — filter by `agentRunId`, `status`, `feedbackKind`
  - `POST /api/review-feedback` — echo back with generated ID
  - `GET /api/review-feedback/:id` — find by ID or 404
  - `PATCH /api/review-feedback/:id` — update status, return updated
- [ ] 4B.2 Add guideline handlers to `mocks/handlers.ts`:
  - `GET /api/review-guidelines` — filter by `status`, `scopeKind`, `search`
  - `POST /api/review-guidelines` — echo back with generated ID, reject duplicate slug
  - `GET /api/review-guidelines/:id` — find by ID or 404
  - `PATCH /api/review-guidelines/:id` — partial update
- [ ] 4B.3 Add run-guideline link handlers:
  - `GET /api/annotation-runs/:id/guidelines` — return linked subset
  - `POST /api/annotation-runs/:id/guidelines` — add link
  - `DELETE /api/annotation-runs/:id/guidelines/:guidelineId` — remove link
- [ ] 4B.4 Extend existing review handlers to accept/echo comment + guidelineIds
- [ ] 4B.5 Verify: all existing stories still load without errors (regression check)

---

## Phase 5 — Shared Badges & UI Primitives

> **Pattern:** Each widget goes in its own file under `components/shared/`,
> gets a `data-part` in `shared/parts.ts`, and a Storybook story.

### 5A. Parts namespace

- [ ] 5A.1 Extend `components/shared/parts.ts`:
  ```ts
  mailboxBadge: "mailbox-badge",
  feedbackStatusBadge: "feedback-status-badge",
  feedbackKindBadge: "feedback-kind-badge",
  guidelineScopeBadge: "guideline-scope-badge",
  ```

### 5B. MailboxBadge

- [ ] 5B.1 Create `components/shared/MailboxBadge.tsx`
  - Props: `mailboxName: string`, `variant?: "chip" | "inline"`
  - Render nothing if `mailboxName === ""`
  - Chip with icon: INBOX→mailbox, Sent→send, Archive→archive, default→folder
  - `data-part={parts.mailboxBadge}`

- [ ] 5B.2 Create `components/shared/stories/MailboxBadge.stories.tsx`
  - Stories: Empty, INBOX, Sent, Archive, Custom, AllVariants row

### 5C. FeedbackKindBadge

- [ ] 5C.1 Create `components/shared/FeedbackKindBadge.tsx`
  - Props: `kind: FeedbackKind`
  - Colors: reject_request→error, comment→info, guideline_request→warning, clarification→default
  - `data-part={parts.feedbackKindBadge}`

- [ ] 5C.2 Create `components/shared/stories/FeedbackKindBadge.stories.tsx`
  - Stories: Each kind, AllKinds row

### 5D. FeedbackStatusBadge

- [ ] 5D.1 Create `components/shared/FeedbackStatusBadge.tsx`
  - Props: `status: FeedbackStatus`
  - Colors: open→warning, acknowledged→info, resolved→success, archived→default
  - `data-part={parts.feedbackStatusBadge}`

- [ ] 5D.2 Create `components/shared/stories/FeedbackStatusBadge.stories.tsx`
  - Stories: Each status, AllStatuses row

### 5E. GuidelineScopeBadge

- [ ] 5E.1 Create `components/shared/GuidelineScopeBadge.tsx`
  - Props: `scopeKind: GuidelineScopeKind`
  - Icons: global→public, mailbox→mail, sender→person, domain→domain, workflow→settings
  - `data-part={parts.guidelineScopeBadge}`

- [ ] 5E.2 Create `components/shared/stories/GuidelineScopeBadge.stories.tsx`
  - Stories: Each scope, AllScopes row

### 5F. Verify

- [ ] 5F.1 Run `tsc --noEmit` — all badge components compile
- [ ] 5F.2 Run `storybook dev` — all badge stories render without errors
- [ ] 5F.3 Export all new badges from `components/shared/index.ts`

---

## Phase 6 — ReviewCommentDrawer & GuidelinePicker

> **New widget directory:** `components/ReviewFeedback/`

### 6A. Parts & barrel

- [ ] 6A.1 Create `components/ReviewFeedback/parts.ts`:
  ```ts
  feedbackPanel: "feedback-panel",
  feedbackList: "feedback-list",
  feedbackCard: "feedback-card",
  feedbackForm: "feedback-form",
  ```
- [ ] 6A.2 Create `components/ReviewFeedback/index.ts` barrel exports

### 6B. GuidelinePicker (inline checklist)

- [ ] 6B.1 Create `components/ReviewFeedback/GuidelinePicker.tsx`
  - Props: `selectedIds: string[]`, `onToggle: (id: string) => void`
  - Fetches guidelines via `useListGuidelinesQuery({ status: "active" })`
  - Renders a compact checklist with slug + title + priority
  - `data-part="guideline-picker"`
  - Shows loading/empty states

- [ ] 6B.2 Create `components/ReviewFeedback/stories/GuidelinePicker.stories.tsx`
  - Stories: Default (several guidelines), Empty (no guidelines), Loading, OneSelected
  - Use MSW handlers from Phase 4

### 6C. ReviewCommentDrawer

- [ ] 6C.1 Create `components/ReviewFeedback/ReviewCommentDrawer.tsx`
  - Props:
    ```
    open: boolean
    mode: "single" | "batch" | "run"
    targetCount: number
    agentRunId?: string
    mailboxName?: string
    onSubmit: (payload) => void
    onCancel: () => void
    ```
  - Children: feedbackKind select, title textfield, bodyMarkdown textarea, GuidelinePicker, mailboxName display, Cancel + Submit
  - `data-part="comment-drawer"`
  - Collapse/Slide transition
  - Submit button label varies: "Reject N Items" (batch), "Dismiss & Explain" (single), "Submit Feedback" (run)

- [ ] 6C.2 Create `components/ReviewFeedback/stories/ReviewCommentDrawer.stories.tsx`
  - Stories:
    - Closed
    - OpenBatchMode (3 items selected)
    - OpenSingleMode
    - OpenRunMode
    - WithGuidelinePreSelected
  - Wrap in `withStore` decorator for RTK Query

### 6D. ReviewCommentInline

- [ ] 6D.1 Create `components/ReviewFeedback/ReviewCommentInline.tsx`
  - Props:
    ```
    open: boolean
    annotationId: string
    agentRunId: string
    mailboxName?: string
    onSubmit: (reviewState, comment?, guidelineIds?) => void
    onJustDismiss: () => void
    onCancel: () => void
    ```
  - Simpler than drawer: compact form, no GuidelinePicker initially (just an "Attach Guideline" expandable)
  - Two action buttons: "Just Dismiss" (fast path) + "Dismiss & Explain"
  - `data-part="review-comment-inline"`

- [ ] 6D.2 Create `components/ReviewFeedback/stories/ReviewCommentInline.stories.tsx`
  - Stories: Closed, Open, WithGuidelineExpanded

### 6E. Verify

- [ ] 6E.1 `tsc --noEmit` passes
- [ ] 6E.2 All new stories render in Storybook
- [ ] 6E.3 MSW intercepts guideline + feedback endpoints correctly

---

## Phase 7 — Enhance Review Queue Page

### 7A. Extend BatchActionBar

- [ ] 7A.1 Add `onRejectExplain?: () => void` prop to `BatchActionBar`
- [ ] 7A.2 Add "Reject & Explain" button (appears when `onRejectExplain` is provided + has selection)
  - Color: error, icon: `CancelIcon` + `CommentIcon`
  - Positioned after Dismiss button
- [ ] 7A.3 Update `BatchActionBar.stories.tsx`: add story with RejectExplain button

### 7B. Wire ReviewQueuePage

- [ ] 7B.1 Add local state `commentDrawerOpen: boolean` to `ReviewQueuePage`
- [ ] 7B.2 Add `handleRejectExplain` callback → opens drawer
- [ ] 7B.3 Render `ReviewCommentDrawer` below `AnnotationTable`:
  - `mode="batch"`, `targetCount={selected.length}`
  - `onSubmit` → call `batchReview` with extended payload, close drawer, clear selection
  - `onCancel` → close drawer
- [ ] 7B.4 Update `ReviewQueuePage.stories.tsx`:
  - Default story still works (no drawer)
  - Add `BatchRejectWithComment` story: drawer open, 3 selected

### 7C. Extend AnnotationDetail with inline comment

- [ ] 7C.1 Add `onDismissWithComment?: (id: string) => void` prop to `AnnotationDetail`
- [ ] 7C.2 When dismiss is triggered: show `ReviewCommentInline` below the detail content
  - "Just Dismiss" → call `reviewAnnotation({ id, reviewState: "dismissed" })` (fast path)
  - "Dismiss & Explain" → call `reviewAnnotation({ id, reviewState: "dismissed", comment, guidelineIds })`
- [ ] 7C.3 Update `AnnotationTable.stories.tsx`: add story with inline comment open

### 7D. Verify

- [ ] 7D.1 Fast approve still works (click ✓, no drawer/modal)
- [ ] 7D.2 Fast dismiss still works (via "Just Dismiss" in inline panel)
- [ ] 7D.3 Dismiss with comment creates feedback via extended review mutation
- [ ] 7D.4 Batch reject with comment creates one feedback for all selected items
- [ ] 7D.5 All ReviewQueuePage stories render without errors
- [ ] 7D.6 `tsc --noEmit` passes

---

## Phase 8 — FeedbackCard & RunFeedbackSection

### 8A. FeedbackCard

- [ ] 8A.1 Create `components/ReviewFeedback/FeedbackCard.tsx`
  - Props:
    ```
    feedback: ReviewFeedback
    onAcknowledge?: () => void
    onResolve?: () => void
    compact?: boolean
    ```
  - Renders: FeedbackKindBadge, FeedbackStatusBadge, createdBy, createdAt, title, MarkdownRenderer(body), target count badge, action buttons
  - `data-part={parts.feedbackCard}`

- [ ] 8A.2 Create `components/ReviewFeedback/stories/FeedbackCard.stories.tsx`
  - Stories: OpenRejectRequest, ResolvedComment, AcknowledgedClarification, Compact mode, AllStates row

### 8B. RunFeedbackSection

- [ ] 8B.1 Create `components/ReviewFeedback/RunFeedbackSection.tsx`
  - Props:
    ```
    runId: string
    feedback: ReviewFeedback[]
    onCreateFeedback: () => void
    onUpdateStatus: (id: string, status: string) => void
    ```
  - Renders: section header + "Add Run Feedback" button + FeedbackCard list
  - Empty state: "No feedback yet" message
  - `data-part="feedback-section"`

- [ ] 8B.2 Create `components/ReviewFeedback/stories/RunFeedbackSection.stories.tsx`
  - Stories: Empty, MultipleFeedback, WithStatusTransitions

### 8C. Verify

- [ ] 8C.1 `tsc --noEmit` passes
- [ ] 8C.2 FeedbackCard stories render correctly
- [ ] 8C.3 RunFeedbackSection stories render with MSW data

---

## Phase 9 — RunGuidelineSection & GuidelineLinkPicker

> **New widget directory:** `components/RunGuideline/`

### 9A. Parts & barrel

- [ ] 9A.1 Create `components/RunGuideline/parts.ts`:
  ```ts
  runGuidelineSection: "run-guideline-section",
  guidelineCard: "guideline-card",
  ```
- [ ] 9A.2 Create `components/RunGuideline/index.ts` barrel exports

### 9B. GuidelineLinkPicker modal

- [ ] 9B.1 Create `components/ReviewFeedback/GuidelineLinkPicker.tsx`
  - MUI `Dialog` modal
  - Props:
    ```
    open: boolean
    runId: string
    alreadyLinkedIds: string[]
    onLink: (guidelineIds: string[]) => void
    onClose: () => void
    ```
  - Fetches `useListGuidelinesQuery({ status: "active" })`
  - Local search field filters by title/slug
  - Checklist with guideline cards
  - Submit button: "Link N Guidelines"
  - `data-part="guideline-link-picker"`

- [ ] 9B.2 Create `components/ReviewFeedback/stories/GuidelineLinkPicker.stories.tsx`
  - Stories: EmptySearch, WithResults, WithSelection, AlreadyLinkedExcluded

### 9C. GuidelineCard (compact, for run section)

- [ ] 9C.1 Create `components/RunGuideline/GuidelineCard.tsx`
  - Props:
    ```
    guideline: ReviewGuideline
    onUnlink?: () => void
    compact?: boolean
    ```
  - Renders: title + slug, GuidelineScopeBadge, status badge, priority, truncated body, Unlink button
  - `data-part={parts.guidelineCard}`

- [ ] 9C.2 Create `components/RunGuideline/stories/GuidelineCard.stories.tsx`
  - Stories: Active, Archived, Draft, WithUnlinkButton

### 9D. RunGuidelineSection

- [ ] 9D.1 Create `components/RunGuideline/RunGuidelineSection.tsx`
  - Props:
    ```
    runId: string
    guidelines: ReviewGuideline[]
    onLink: (guidelineId: string) => void
    onUnlink: (guidelineId: string) => void
    onCreateAndLink: () => void
    ```
  - Renders: section header + GuidelineCard[] + action buttons
  - "Link Existing Guideline" → opens GuidelineLinkPicker
  - "Create New Guideline for This Run" → calls `onCreateAndLink`
  - Empty state: "No guidelines linked to this run"

- [ ] 9D.2 Create `components/RunGuideline/stories/RunGuidelineSection.stories.tsx`
  - Stories: Empty, OneGuideline, MultipleGuidelines

### 9E. Verify

- [ ] 9E.1 `tsc --noEmit` passes
- [ ] 9E.2 All RunGuideline stories render
- [ ] 9E.3 GuidelineLinkPicker opens, searches, and submits correctly in Storybook

---

## Phase 10 — Enhance Run Detail Page

### 10A. Wire RunGuidelineSection

- [ ] 10A.1 Add `useGetRunGuidelinesQuery(runId)` to `RunDetailPage`
- [ ] 10A.2 Add `useLinkGuidelineToRunMutation()` and `useUnlinkGuidelineFromRunMutation()`
- [ ] 10A.3 Render `RunGuidelineSection` between StatBox row and RunTimeline
  - `onCreateAndLink` → navigate to `/annotations/guidelines/new?runId=${runId}`
- [ ] 10A.4 Wire GuidelineLinkPicker: open state, onLink → call `linkGuidelineToRun` for each ID

### 10B. Wire RunFeedbackSection

- [ ] 10B.1 Add `useListReviewFeedbackQuery({ agentRunId: runId })` to `RunDetailPage`
- [ ] 10B.2 Add `useCreateReviewFeedbackMutation()` and `useUpdateReviewFeedbackMutation()`
- [ ] 10B.3 Render `RunFeedbackSection` between RunTimeline and Groups
- [ ] 10B.4 "Add Run Feedback" → open ReviewCommentDrawer with `mode="run"`
- [ ] 10B.5 Status transitions (Acknowledge/Resolve) → call `updateReviewFeedback`

### 10C. Extend batch actions on run annotations

- [ ] 10C.1 Add local `commentDrawerOpen` state
- [ ] 10C.2 Replace simple "Approve All" with enhanced batch bar:
  - "Approve All" (fast, no comment)
  - "Reject & Explain" (opens drawer in batch mode)
- [ ] 10C.3 Wire batch review through extended `batchReview` mutation

### 10D. Update stories

- [ ] 10D.1 Update `RunDetailPage.stories.tsx`:
  - Default: now shows RunGuidelineSection + RunFeedbackSection
  - Add `RunWithGuidelines`: 2 linked guidelines
  - Add `RunWithFeedback`: 3 feedback items, mixed statuses
  - Add `RunWithInlineComment`: annotation detail expanded with comment inline
  - Loading/NotFound stories still work

### 10E. Verify

- [ ] 10E.1 `tsc --noEmit` passes
- [ ] 10E.2 All RunDetailPage stories render
- [ ] 10E.3 Guideline link/unlink triggers correct mutations
- [ ] 10E.4 Feedback create/status-update works in stories
- [ ] 10E.5 Existing "Approve All" still works as before

---

## Phase 11 — Guidelines Management Pages

> **New widget directory:** `components/Guidelines/`

### 11A. Parts & barrel

- [ ] 11A.1 Create `components/Guidelines/parts.ts`:
  ```ts
  guidelineList: "guideline-list",
  guidelineSummaryCard: "guideline-summary-card",
  guidelineEditor: "guideline-editor",
  guidelineLinkedRuns: "guideline-linked-runs",
  ```
- [ ] 11A.2 Create `components/Guidelines/index.ts` barrel exports

### 11B. GuidelineSummaryCard

- [ ] 11B.1 Create `components/Guidelines/GuidelineSummaryCard.tsx`
  - Props:
    ```
    guideline: ReviewGuideline
    linkedRunCount: number
    onEdit: () => void
    onArchive?: () => void
    onActivate?: () => void
    ```
  - Renders: title + slug, truncated body (2 lines), GuidelineScopeBadge, status badge, priority, linked run count, dates, action buttons
  - `data-part={parts.guidelineSummaryCard}`

- [ ] 11B.2 Create `components/Guidelines/stories/GuidelineSummaryCard.stories.tsx`
  - Stories: Active, Archived, Draft, AllStates row

### 11C. GuidelineForm

- [ ] 11C.1 Create `components/Guidelines/GuidelineForm.tsx`
  - Props:
    ```
    guideline?: ReviewGuideline
    mode: "view" | "edit" | "create"
    onSave: (payload) => void
    onCancel: () => void
    ```
  - View mode: read-only metadata + MarkdownRenderer for body
  - Edit/Create mode: textfields, selects, textarea, live preview via MarkdownRenderer
  - Slug field: editable in create, read-only in edit
  - `data-part={parts.guidelineEditor}`

- [ ] 11C.2 Create `components/Guidelines/stories/GuidelineForm.stories.tsx`
  - Stories: ViewMode, EditMode, CreateMode, CreateWithPreview

### 11D. GuidelineLinkedRuns

- [ ] 11D.1 Create `components/Guidelines/GuidelineLinkedRuns.tsx`
  - Props: `runs: AgentRunSummary[]`
  - Renders: list of run summary rows (clickable → run detail page)
  - `data-part={parts.guidelineLinkedRuns}`

- [ ] 11D.2 Create `components/Guidelines/stories/GuidelineLinkedRuns.stories.tsx`
  - Stories: MultipleRuns, SingleRun, Empty

### 11E. GuidelinesListPage

- [ ] 11E.1 Create `pages/GuidelinesListPage.tsx`
  - Uses `useListGuidelinesQuery()` with status/search params
  - Filter pills: All, Active, Archived, Draft
  - Search field
  - CountSummaryBar
  - GuidelineSummaryCard list
  - "New Guideline" button → navigate to `/annotations/guidelines/new`
  - `data-widget="guidelines-list-page"`

- [ ] 11E.2 Create `pages/stories/GuidelinesListPage.stories.tsx`
  - Stories: Default (mixed statuses), OnlyActive, Empty, SearchFiltered
  - Full MSW handlers for guidelines + feedback endpoints
  - Wrap in `withAll("/annotations/guidelines")`

### 11F. GuidelineDetailPage

- [ ] 11F.1 Create `pages/GuidelineDetailPage.tsx`
  - Detect mode from route: `/new` → create, `/:id` → view/edit
  - View mode: GuidelineForm (view) + GuidelineLinkedRuns + feedback list
  - Edit mode: GuidelineForm (edit) with save/cancel
  - Create mode: GuidelineForm (create), optionally with `?runId=` param
  - After create+link: navigate back to run detail if `runId` param present
  - `data-widget="guideline-detail-page"`

- [ ] 11F.2 Create `pages/stories/GuidelineDetailPage.stories.tsx`
  - Stories: ViewMode (with linked runs), EditMode, CreateMode, CreateFromRun
  - Wrap in `withAll("/annotations/guidelines/guideline-001", "/annotations/guidelines/:guidelineId")`

### 11G. Verify

- [ ] 11G.1 `tsc --noEmit` passes
- [ ] 11G.2 All guideline stories render
- [ ] 11G.3 Create guideline form validates required fields
- [ ] 11G.4 Edit → save updates via mutation
- [ ] 11G.5 Linked runs display correctly

---

## Phase 12 — Sidebar, Routes & Navigation

### 12A. Sidebar

- [ ] 12A.1 Add `Guidelines` entry to `reviewItems` in `AnnotationSidebar.tsx`
  - Icon: `MenuBookIcon` (or `RuleIcon`)
  - Path: `/annotations/guidelines`

### 12B. Routes

- [ ] 12B.1 Add guideline routes to `App.tsx`:
  ```tsx
  <Route path="guidelines" element={<GuidelinesListPage />} />
  <Route path="guidelines/new" element={<GuidelineDetailPage />} />
  <Route path="guidelines/:guidelineId" element={<GuidelineDetailPage />} />
  ```

### 12C. Verify

- [ ] 12C.1 Sidebar shows Guidelines under Review section
- [ ] 12C.2 Clicking Guidelines navigates to list page
- [ ] 12C.3 Clicking a guideline card navigates to detail page
- [ ] 12C.4 "New Guideline" button navigates to create page
- [ ] 12C.5 All existing routes still work (Dashboard, Review, Runs, Senders, Groups, Query)

---

## Phase 13 — Mailbox Context Integration

### 13A. AnnotationTable

- [ ] 13A.1 Add optional `Mailbox` column to `AnnotationTable` (after Source column)
  - Only shown when `annotations` span multiple `mailboxName` values (or always if annotation type has it)
  - Use `MailboxBadge` in compact mode
- [ ] 13A.2 Update `AnnotationTable.stories.tsx`: add story with mixed mailboxes

### 13B. AnnotationDetail

- [ ] 13B.1 Show `MailboxBadge` in the header metadata row of `AnnotationDetail`
  - After `ReviewStateBadge`, only if `mailboxName` is non-empty
- [ ] 13B.2 Update detail story with mailbox badge

### 13C. MessagePreviewTable

- [ ] 13C.1 Add `Mailbox` column to `MessagePreviewTable` (after Subject)
  - Uses `MailboxBadge`
- [ ] 13C.2 Update `SenderProfile.stories.tsx` with mailbox data in messages

### 13D. Review queue filters

- [ ] 13D.1 Add mailbox filter pills to `ReviewQueuePage`
  - Compute unique mailbox names from annotations
  - Only show when >1 distinct mailbox in current view
  - Filter via `AnnotationFilter.mailboxName`
- [ ] 13D.2 Update `ReviewQueuePage.stories.tsx` with mixed-mailbox data

### 13E. Verify

- [ ] 13E.1 `tsc --noEmit` passes
- [ ] 13E.2 MailboxBadge appears in tables when data has `mailboxName`
- [ ] 13E.3 MailboxBadge hidden when `mailboxName === ""`
- [ ] 13E.4 Mailbox filter pills appear only when annotations span multiple mailboxes
- [ ] 13E.5 All stories still render after mailbox integration

---

## Phase 14 — Redux Slice Enhancements

### 14A. annotationUiSlice

- [ ] 14A.1 Add `commentDrawerOpen: boolean` to `ReviewQueueState`
- [ ] 14A.2 Add `openCommentDrawer` / `closeCommentDrawer` actions
- [ ] 14A.3 Add `filterMailbox: string | null` to `ReviewQueueState`
- [ ] 14A.4 Add `setFilterMailbox` action
- [ ] 14A.5 Verify existing review queue stories still work (backward compat)

---

## Phase 15 — End-to-End Verification

### 15A. Manual walkthrough

- [ ] 15A.1 Start annotation UI server, open in browser
- [ ] 15A.2 Open Review Queue → select 3 items → "Reject & Explain" → fill comment → submit → verify feedback appears
- [ ] 15A.3 Open single annotation → dismiss with "Just Dismiss" → verify fast path
- [ ] 15A.4 Open single annotation → "Dismiss & Explain" → fill comment → submit
- [ ] 15A.5 Open Run Detail → verify linked guidelines section shows
- [ ] 15A.6 Link a guideline to a run → verify it appears
- [ ] 15A.7 Unlink a guideline → verify it disappears
- [ ] 15A.8 Create a new guideline from run context → verify it's linked
- [ ] 15A.9 Add run-level feedback → verify it appears
- [ ] 15A.10 Acknowledge feedback → verify status badge updates
- [ ] 15A.11 Resolve feedback → verify status badge updates
- [ ] 15A.12 Open Guidelines list → filter by status → search → verify
- [ ] 15A.13 Open a guideline → verify linked runs display
- [ ] 15A.14 Edit a guideline → save → verify updated
- [ ] 15A.15 Archive a guideline → verify it disappears from active list
- [ ] 15A.16 Check mailbox badges appear in annotation table (when data has mailbox)
- [ ] 15A.17 Check mailbox filter pills in review queue

### 15B. Storybook smoke test

- [ ] 15B.1 Run `storybook build` → confirm no build errors
- [ ] 15B.2 Open static storybook → click through all new stories
- [ ] 15B.3 Verify no console errors in any new story

### 15C. Type safety

- [ ] 15C.1 `tsc --noEmit` passes with zero errors
- [ ] 15C.2 No `any` types introduced (grep for `: any` in new files)

---

## Phase 16 — Documentation & Cleanup

- [ ] 16.1 Update `reference/02-diary.md` with implementation narrative
- [ ] 16.2 Run `docmgr doc relate` for all new/modified UI files
- [ ] 16.3 Run `docmgr doctor --ticket SMN-20260403-RUN-REVIEW`
- [ ] 16.4 Upload final doc bundle to reMarkable
- [ ] 16.5 Review all `data-part` attributes are documented in `parts.ts` files
- [ ] 16.6 Verify all new components have Storybook stories (no bare widgets without stories)
