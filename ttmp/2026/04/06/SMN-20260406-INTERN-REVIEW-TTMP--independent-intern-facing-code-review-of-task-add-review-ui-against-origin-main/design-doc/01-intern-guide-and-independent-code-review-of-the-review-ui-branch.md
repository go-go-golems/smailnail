---
Title: Intern guide and independent code review of the review UI branch
Ticket: SMN-20260406-INTERN-REVIEW-TTMP
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
    - workflow
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: README.md
      Note: Repo-level orientation and product surfaces used in the intern guide
    - Path: cmd/smailnail/commands/sqlite/serve.go
      Note: Entry point for the sqlite-backed review server
    - Path: pkg/annotate/repository_feedback.go
      Note: Core transactional repository logic and several key review findings
    - Path: pkg/annotate/types.go
      Note: Review data model
    - Path: pkg/annotationui/handlers_annotations.go
      Note: Single and batch review handlers for annotations
    - Path: pkg/annotationui/handlers_feedback.go
      Note: Feedback
    - Path: pkg/annotationui/server.go
      Note: HTTP route registration
    - Path: ttmp/2026/04/06/SMN-20260406-CODE-REVIEW--code-review-run-review-feedback-guidelines-mailbox-aware-analysis/design-doc/01-comprehensive-code-review-run-review-feedback-guidelines-mailbox-aware-analysis.md
      Note: Later reviewed to import a small set of validated findings with provenance labels
    - Path: ui/src/api/annotations.ts
      Note: Central RTK Query contract for the review UI
    - Path: ui/src/components/Guidelines/GuidelineSummaryCard.tsx
      Note: Added after reviewing the intern ticket for fake linkedRunCount analysis
    - Path: ui/src/components/ReviewFeedback/ReviewCommentInline.tsx
      Note: Added after reviewing the intern ticket as a validated dead-code finding
    - Path: ui/src/mocks/handlers.ts
      Note: Added after reviewing the intern ticket for Storybook/MSW drift and non-persistent create handlers
    - Path: ui/src/pages/GuidelineDetailPage.tsx
      Note: Evidence for placeholder linked-runs UI
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Evidence for review-queue semantics issues
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Evidence for run detail workflow and feedback-scope leakage
    - Path: ui/src/pages/stories/GuidelineDetailPage.stories.tsx
      Note: Added after reviewing the intern ticket for incorrect story endpoint overrides
    - Path: ui/src/pages/stories/GuidelinesListPage.stories.tsx
      Note: Added after reviewing the intern ticket for incorrect story endpoint overrides
    - Path: ui/src/store/annotationUiSlice.ts
      Note: Evidence for unused review state
ExternalSources: []
Summary: Evidence-backed intern guide to the review UI branch, including architecture mapping, API flow, strengths, concrete cleanup recommendations against origin/main, and a short revision note incorporating additional validated findings surfaced by a later meta-review of the intern review.
LastUpdated: 2026-04-06T20:55:00Z
WhatFor: Onboard a new intern to smailnail's annotation/review slice and capture an independent code review of the task/add-review-ui branch.
WhenToUse: Read this before modifying the annotation UI, review feedback, guideline linking, or the sqlite-backed review server.
---



# Intern guide and independent code review of the review UI branch

## Executive Summary

This branch is a substantial annotation-review feature expansion on top of smailnail’s SQLite mirror tooling. Compared with `origin/main`, it adds a new review-feedback data model, reusable review guidelines, run-to-guideline linking, several new API routes in the SQLite-backed review server, and a large React/MUI frontend slice for reviewing agent output. The work is directionally good: the backend schema is explicit, the repository layer is readable, UI stories and MSW mocks were added, and the main review actions are wrapped in database transactions.

For a new intern, the most important mental model is this: **smailnail is not “just an email UI.”** It is a repository with several tools, but this branch is mainly about one specific sub-system: a local SQLite-backed annotation workbench where an agent proposes labels or groups, and a human reviewer accepts, dismisses, comments, and turns repeated feedback into reusable guidelines. The relevant runtime path is:

```text
Browser (React/MUI)
  -> RTK Query hooks in ui/src/api/annotations.ts
  -> /api routes in pkg/annotationui/server.go
  -> HTTP handlers in pkg/annotationui/*.go
  -> annotate.Repository in pkg/annotate/*.go
  -> SQLite mirror / annotation tables
```

The highest-value review findings are:

1. **The “Review Queue” page is not actually a queue yet**; it loads all annotations instead of only pending review items.
2. **Run-level feedback is mixed with annotation/selection feedback** because neither the API nor the page filters by `scopeKind`.
3. **The TypeScript and Go feedback contracts have already drifted** (`targetIds` vs `targets`, update payload shape mismatch).
4. **Audit fields exist in the schema but are not populated by the handlers**, so `created_by` / `linked_by` are effectively empty for the new feature set.
5. **The guideline detail page advertises linked runs, but the implementation always renders an empty list.**
6. **The branch still shows transitional architecture and tooling confusion**: a legacy shell remains in the app while the embedded server redirects `/` to `/annotations`, and the UI declares `pnpm` while a large `package-lock.json` was added.

I intentionally wrote this report as an **independent review** and did **not** consult the existing review ticket contents during the original write-up. All primary conclusions above are grounded in direct inspection of the branch diff and source files.

### Revision note after meta-review of the intern review

After finishing the original version of this report, I later reviewed the separate intern ticket `SMN-20260406-CODE-REVIEW` and rechecked its useful claims against source. I am **not** replacing the original analysis, but I am explicitly folding in a few additional findings that the intern surfaced and that I independently validated in code.

**Added after reviewing the intern ticket and re-validating in source:**

1. `ReviewCommentInline` is dead code and should either be removed or explicitly parked for future use.
2. `annotationUiSlice` contains dead review UI state (`commentDrawerOpen`, `filterMailbox`) that is no longer wired to the live page flow.
3. Storybook guideline page stories use `/api/guidelines` overrides while the real app uses `/api/review-guidelines`, so some overrides are dead and the fallback mocks are doing the real work.
4. MSW create handlers return synthetic feedback/guideline objects without persisting them into their backing mock collections, so create-then-list flows are misleading in Storybook/dev.
5. `GuidelineSummaryCard` currently renders `linkedRunCount={0}` everywhere, which advertises data the backend does not yet provide.
6. Feedback creation SQL is duplicated between transactional and non-transactional repository paths, which is real maintenance debt even if it is not a correctness failure today.

Where those additions matter, I call them out explicitly below as **[Added after reviewing the intern ticket]**.

## Scope and Review Method

### Branch and diff surface

I reviewed the current branch against `origin/main` using `git diff origin/main...HEAD` from the `smailnail` repo root.

High-level diff summary:

- **77 changed files** total.
- **58 frontend files** under `ui/`.
- **15 backend files** in `pkg/annotate`, `pkg/annotationui`, small CLI enrich changes, and related docs.
- The biggest new code blocks are in:
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `ui/src/api/annotations.ts`
  - `ui/src/pages/RunDetailPage.tsx`
  - `ui/src/pages/ReviewQueuePage.tsx`
  - `ui/src/pages/GuidelinesListPage.tsx`
  - `ui/src/pages/GuidelineDetailPage.tsx`

### Validation commands

I used these concrete checks while reviewing:

```bash
cd smailnail
git diff --stat origin/main...HEAD
rg -n "CreateReviewFeedback|ListReviewFeedback|CreateGuideline|LinkGuidelineToRun|ReviewAnnotationWithArtifacts|BatchReviewWithArtifacts|review-feedback|review-guidelines|guideline" pkg/annotate pkg/annotationui -g '*_test.go'
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1
cd ui && pnpm run check
```

Results:

- `go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1` ✅
- `pnpm run check` ✅
- The `rg ... -g '*_test.go'` search returned **no hits** for the new feedback/guideline flows, which is itself a review finding.

## What smailnail is, in plain English

### Repository-level view

At the repo root, `README.md` explains that smailnail currently contains several surfaces:

- `smailnail`: CLI for IMAP search/fetch/mirror/enrich workflows.
- `mailgen`: fixture/synthetic email generator.
- `imap-tests`: helper CLI for IMAP fixture setup.
- `smailnaild`: hosted backend for accounts/rules/auth work.
- `smailnail-imap-mcp`: JS/MCP surface.
- A separate SQLite-backed annotation/review UI served by `smailnail sqlite serve`.

For this branch, you can largely ignore the IMAP fetching internals at first and focus on four directories:

| Area | Why it matters | Key files |
|---|---|---|
| Review server entrypoint | starts the SQLite-backed review UI backend | `cmd/smailnail/commands/sqlite/serve.go` |
| Review domain model + persistence | owns tables, repository methods, transactional writes | `pkg/annotate/types.go`, `pkg/annotate/schema.go`, `pkg/annotate/repository.go`, `pkg/annotate/repository_feedback.go` |
| HTTP API layer | translates requests into repository calls | `pkg/annotationui/server.go`, `pkg/annotationui/handlers_annotations.go`, `pkg/annotationui/handlers_feedback.go`, `pkg/annotationui/types_feedback.go` |
| React frontend | routes, pages, components, RTK Query hooks | `ui/src/App.tsx`, `ui/src/api/annotations.ts`, `ui/src/pages/*`, `ui/src/components/*` |

### What this branch adds conceptually

Before this branch, the annotation UI was mostly about inspecting annotations, logs, groups, senders, and runs. This branch adds a **human feedback loop** on top of that.

That loop looks like this:

```text
Agent creates annotations/logs/groups
  -> human sees them in review pages
  -> human approves / dismisses / comments
  -> comments become stored review_feedback rows
  -> repeated lessons can become review_guidelines
  -> guidelines can be linked to an agent run
```

This is a useful product direction because it turns one-off review decisions into reusable policy, instead of leaving them buried in free-form notes.

## Current-State Architecture

### 1. Server bootstrap

The review UI is served by `smailnail sqlite serve`, not by `smailnaild`.

`cmd/smailnail/commands/sqlite/serve.go` does three important things:

1. bootstraps the mirror store,
2. opens the SQLite DB with `sqlx`,
3. starts `annotationui.NewHTTPServer(...)` and serves embedded frontend assets from `pkg/smailnaild/web`.

Important evidence:

- `cmd/smailnail/commands/sqlite/serve.go` describes this as a server separate from `smailnaild`.
- It passes `web.PublicFS` into the annotation UI server so the same process serves both API and SPA assets.

Pseudocode:

```go
func Run(...) {
  bootstrapMirror(sqlitePath, mirrorRoot)
  db := sqlx.Open("sqlite3", sqlitePath)
  server := annotationui.NewHTTPServer({ DB: db, PublicFS: web.PublicFS, ... })
  return annotationui.RunServer(ctx, server)
}
```

### 2. API registration and SPA serving

`pkg/annotationui/server.go` is the HTTP wiring hub.

Important routes now include:

- `GET /api/annotations`
- `PATCH /api/annotations/{id}/review`
- `POST /api/annotations/batch-review`
- `GET/POST/PATCH /api/review-feedback...`
- `GET/POST/PATCH /api/review-guidelines...`
- `GET/POST/DELETE /api/annotation-runs/{id}/guidelines...`

The same server also serves the SPA for `/annotations` and `/query` routes, and explicitly redirects `/` to `/annotations` (`pkg/annotationui/server.go:203-223`).

That is one source of architectural confusion, because the frontend app still contains a separate legacy root shell at `/` (see `ui/src/App.tsx:165-190`). More on that in the findings section.

### 3. Review data model

The domain model additions live in `pkg/annotate/types.go:161-279` and the schema in `pkg/annotate/schema.go`.

The core new entities are:

- `ReviewFeedback`
  - scope (`annotation`, `selection`, `run`, `guideline`)
  - kind (`comment`, `reject_request`, `guideline_request`, `clarification`)
  - status (`open`, `acknowledged`, `resolved`, `archived`)
  - optional targets via `review_feedback_targets`
- `ReviewGuideline`
  - reusable instruction/policy with `slug`, `scopeKind`, `status`, `priority`
- `RunGuidelineLink`
  - join table linking a run to guidelines

Schema additions are nicely isolated in `SchemaMigrationV4Statements()`.

### 4. Repository behavior

`pkg/annotate/repository_feedback.go` is the persistence center for the new feature set.

Good news first:

- `CreateReviewFeedback` wraps feedback + target inserts in a transaction.
- `ReviewAnnotationWithArtifacts` wraps review-state updates, optional feedback creation, and optional guideline links in one transaction.
- `BatchReviewWithArtifacts` does the same for multi-annotation review.

That is the right overall direction. A reviewer action should not half-succeed.

Here is the effective shape of single-item review with artifacts:

```text
PATCH /api/annotations/{id}/review
  -> handleReviewAnnotation()
    -> Repository.ReviewAnnotationWithArtifacts()
      -> update annotation.review_state
      -> reload annotation
      -> if comment present: insert review_feedback + targets
      -> if guidelineIds present: link guidelines to annotation.AgentRunID
      -> commit transaction
```

And in pseudocode:

```go
tx := db.Begin()
update annotations set review_state = ... where id = ...
annotation := select * from annotations where id = ...
if comment != nil {
  insert into review_feedback (...)
  insert into review_feedback_targets (...annotation id...)
}
for guidelineID in guidelineIDs {
  insert into run_guideline_links (annotation.AgentRunID, guidelineID)
}
tx.Commit()
```

### 5. Frontend state and routing

The annotation UI frontend is now mostly a routed MUI app.

Relevant routes from `ui/src/App.tsx:169-184`:

- `/annotations` → dashboard
- `/annotations/review` → review queue
- `/annotations/runs` → run list
- `/annotations/runs/:runId` → run detail
- `/annotations/guidelines` → guideline list
- `/annotations/guidelines/new` and `/:guidelineId` → guideline create/detail
- `/query` → SQL workbench

Network access is centralized in `ui/src/api/annotations.ts`, which is a strong choice. It gives the branch a single API contract file instead of scattered `fetch()` calls.

### 6. Frontend review workflow

The main interaction points are:

- `ReviewQueuePage` for batch review of annotations.
- `RunDetailPage` for reviewing a single agent run in context.
- `RunFeedbackSection` for adding run feedback.
- `ReviewCommentDrawer` for attaching explanation/guidelines to a dismissal.
- `GuidelinesListPage` and `GuidelineDetailPage` for guideline lifecycle.
- `RunGuidelineSection` for linking guidelines to a run.

UI flow diagram:

```text
ReviewQueuePage / RunDetailPage
  -> useReviewAnnotationMutation / useBatchReviewMutation
  -> POST/PATCH /api/annotations...
  -> repository transaction
  -> invalidate RTK Query tags
  -> rerender queue/run/guideline feedback views
```

## API Quick Reference for Interns

These are the API surfaces you will touch first when changing this branch.

| Route | Purpose | Handler | Repository method |
|---|---|---|---|
| `GET /api/annotations` | list annotations | `handleListAnnotations` | `ListAnnotations` |
| `PATCH /api/annotations/{id}/review` | single review action | `handleReviewAnnotation` | `ReviewAnnotationWithArtifacts` |
| `POST /api/annotations/batch-review` | batch review action | `handleBatchReview` | `BatchReviewWithArtifacts` |
| `GET /api/review-feedback` | list feedback | `handleListFeedback` | `ListReviewFeedback` |
| `POST /api/review-feedback` | create feedback | `handleCreateFeedback` | `CreateReviewFeedback` |
| `PATCH /api/review-feedback/{id}` | update feedback status | `handleUpdateFeedback` | `UpdateReviewFeedback` |
| `GET /api/review-guidelines` | list guidelines | `handleListGuidelines` | `ListGuidelines` |
| `POST /api/review-guidelines` | create guideline | `handleCreateGuideline` | `CreateGuideline` |
| `PATCH /api/review-guidelines/{id}` | update guideline | `handleUpdateGuideline` | `UpdateGuideline` |
| `GET /api/annotation-runs/{id}/guidelines` | list linked guidelines for a run | `handleListRunGuidelines` | `ListRunGuidelines` |
| `POST /api/annotation-runs/{id}/guidelines` | link guideline to run | `handleLinkRunGuideline` | `LinkGuidelineToRun` |
| `DELETE /api/annotation-runs/{id}/guidelines/{guidelineId}` | unlink guideline | `handleUnlinkRunGuideline` | `UnlinkGuidelineFromRun` |

## Positive Findings

### A. The branch has a coherent product idea

The review-feedback and review-guideline concepts fit naturally on top of annotations and runs. This is not a random UI expansion; it is a meaningful second-order workflow for turning one-off human review into reusable policy.

### B. Transaction boundaries are mostly in the right place

`ReviewAnnotationWithArtifacts` and `BatchReviewWithArtifacts` are exactly the kinds of operations that benefit from a repository-owned transaction. The code reflects that.

### C. API usage is centralized on the frontend

`ui/src/api/annotations.ts` is a good consolidation point. A new engineer can find most frontend/backend coupling in one file instead of searching the whole UI tree.

### D. Storybook/MSW investment is real

The branch adds many stories and mock handlers. That is a good sign for UI iteration and should make future refactors safer if the team keeps them aligned with real contracts.

## Detailed Findings

---

### 1. High: the “Review Queue” is not actually restricted to pending review items

**Problem**

The page named “Review Queue” fetches all annotations, optionally filtered by tag, but not by `reviewState=to_review`. That makes the page semantically misleading and weakens every downstream action, because batch controls operate on a mixed set of reviewed, dismissed, and pending items.

**Where to look**

- `ui/src/pages/ReviewQueuePage.tsx:36-38`
- `ui/src/pages/ReviewQueuePage.tsx:75-91`
- `pkg/annotationui/handlers_annotations.go:33-40`

**Evidence**

The page loads data with:

```ts
const { data: annotations = [], isLoading } = useListAnnotationsQuery(
  filterTag ? { tag: filterTag } : {},
);
```

There is no `reviewState: "to_review"` in the query.

**Why it matters**

- Reviewers see already-reviewed items in a place that implies pending work.
- Counts and UI language become inconsistent.
- “Select all” can include items that no longer need review.
- Interns will misread the intended product behavior and may build more code on the wrong assumption.

**Cleanup sketch**

```ts
const baseFilter = {
  reviewState: "to_review",
  ...(filterTag ? { tag: filterTag } : {}),
};
useListAnnotationsQuery(baseFilter);
```

If the team wants an “All annotations” browse view too, create a separate page instead of weakening the queue concept.

---

### 2. High: run detail mixes run-level feedback with annotation/selection feedback

**Problem**

`RunDetailPage` asks for feedback by `agentRunId` only and passes all returned items to a component labeled “Run-Level Feedback”. The backend filter type also lacks `scopeKind`, so there is no clean way to request just run-scoped feedback.

**Where to look**

- `ui/src/pages/RunDetailPage.tsx:31-34`
- `ui/src/components/ReviewFeedback/RunFeedbackSection.tsx:45-81`
- `pkg/annotate/types.go:217-223`
- `pkg/annotate/repository_feedback.go:83-124`
- `pkg/annotationui/handlers_feedback.go:19-25`

**Evidence**

- Frontend request:
  - `useListReviewFeedbackQuery({ agentRunId: runId })`
- Display label:
  - `Run-Level Feedback ({feedback.length})`
- Backend list filter has fields for `AgentRunID`, `Status`, `FeedbackKind`, `MailboxName`, but **not** `ScopeKind`.

**Why it matters**

The branch already distinguishes scopes in the model (`annotation`, `selection`, `run`, `guideline`). If the UI collapses those scopes back together, the model’s clarity is lost. A reviewer trying to inspect run-level commentary can instead see per-annotation rejection requests and selection-level notes.

**Cleanup sketch**

Add scope filtering end-to-end.

```go
// pkg/annotate/types.go
type ListFeedbackFilter struct {
  AgentRunID string
  ScopeKind  string
  Status     string
  ...
}
```

```go
if filter.ScopeKind != "" {
  query += ` AND scope_kind = ?`
}
```

```ts
useListReviewFeedbackQuery({ agentRunId: runId, scopeKind: "run" })
```

If you also want “all feedback for this run”, expose that as a second section with explicit naming.

---

### 3. High: the TypeScript and Go feedback contracts are already drifting apart

**Problem**

The frontend feedback request types do not match the Go request/response layer.

**Where to look**

- `ui/src/types/reviewFeedback.ts:50-63`
- `pkg/annotationui/types_feedback.go:31-43`
- `ui/src/api/annotations.ts:145-168`

**Evidence**

Mismatch 1: create payload shape

- TypeScript says:

```ts
interface CreateFeedbackRequest {
  ...
  targetIds?: string[];
}
```

- Go handler expects:

```go
type createFeedbackRequest struct {
  ...
  Targets []feedbackTargetJSON `json:"targets"`
}
```

Mismatch 2: update payload shape

- TypeScript says `UpdateFeedbackRequest` may include `bodyMarkdown`.
- Go handler/repository only accept `status`.

**Why it matters**

This is the kind of bug that hides well in Storybook or mock-backed development and only surfaces later in real integration. It also teaches new contributors the wrong API contract.

**Cleanup sketch**

Pick one canonical contract and make both layers identical.

Recommended minimal fix:

```ts
interface CreateFeedbackRequest {
  scopeKind: FeedbackScopeKind;
  agentRunId?: string;
  mailboxName?: string;
  feedbackKind: FeedbackKind;
  title: string;
  bodyMarkdown: string;
  targets?: { targetType: string; targetId: string }[];
}

interface UpdateFeedbackRequest {
  status?: FeedbackStatus;
}
```

Then update MSW mocks to follow the same shape.

---

### 4. High: audit metadata is modeled but not actually populated

**Problem**

The schema and repository methods carry `created_by` / `linked_by`, but the handlers do not pass a user identity through. As a result, the new feedback and guideline actions lose authorship.

**Where to look**

- `pkg/annotate/types.go:201-209`, `235-252`
- `pkg/annotate/repository_feedback.go:29-42`, `188-199`, `307-315`, `375-384`, `441-449`
- `pkg/annotationui/handlers_feedback.go:72-80`, `166-171`, `237-241`
- `pkg/annotationui/handlers_annotations.go:72-78`, `103-110`

**Evidence**

Examples:

- `CreateFeedbackInput` includes `CreatedBy`.
- `CreateGuidelineInput` includes `CreatedBy`.
- `ReviewAnnotationActionInput` and `BatchReviewActionInput` include `CreatedBy`.
- But the HTTP handlers do not populate those fields.
- `handleLinkRunGuideline` literally passes `""` as `linkedBy`.

**Why it matters**

This feature is specifically about human review and policy formation. Missing authorship metadata makes it harder to:

- trace who created a guideline,
- distinguish human reviewer feedback from imported/system feedback,
- debug policy churn later.

It also makes the schema look more complete than the actual product behavior.

**Cleanup sketch**

Short-term:

```go
createdBy := currentUserIDFromRequest(r) // or temporary fallback
```

Longer-term:

- introduce a small request-scoped identity helper,
- populate `CreatedBy` / `LinkedBy` in every review/guideline handler,
- show authorship in the UI consistently.

---

### 5. Medium: guideline detail has a dead “linked runs” section

**Problem**

The detail page renders a linked-runs component but always passes `runs={[]}`. There is no API query for “which runs are linked to this guideline”.

**Where to look**

- `ui/src/pages/GuidelineDetailPage.tsx:129-138`
- `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx:15-27`
- `ui/src/api/annotations.ts`

**Evidence**

The component call is:

```tsx
<GuidelineLinkedRuns runs={[]} ... />
```

So the UI promises a capability that does not exist yet.

**Why it matters**

This is confusing for users and for new contributors. The presence of a fully named component implies the underlying feature exists, but it is just a placeholder.

**Cleanup sketch**

Two valid options:

1. **Ship it fully**
   - add backend endpoint `GET /api/review-guidelines/{id}/runs`
   - add RTK Query hook
   - load the runs on the detail page
2. **Hide it until ready**
   - remove the section for now
   - add a TODO in ticket docs, not in production UI

---

### 6. Medium: asynchronous guideline-link flows ignore failure and navigate away early

**Problem**

Several frontend mutations are fired with `void ...` and the UI closes/navigates immediately, even though the link operation may fail.

**Where to look**

- `ui/src/components/RunGuideline/RunGuidelineSection.tsx:31-36`
- `ui/src/pages/GuidelineDetailPage.tsx:42-58`

**Evidence**

- `RunGuidelineSection` loops over IDs and calls `void linkGuidelineToRun(...)`, then closes the picker immediately.
- `GuidelineDetailPage` creates a guideline, then if `runIdParam` exists, fires `void linkGuidelineToRun(...)` and immediately navigates back to the run page.

**Why it matters**

These are classic “optimistic but untracked” flows.

Possible symptoms:

- link silently fails and the user lands on the run page assuming it succeeded,
- partial success when linking multiple guidelines,
- harder-to-debug race conditions during slower networks or future auth failures.

**Cleanup sketch**

```ts
const created = await createGuideline(payload).unwrap();
if (runIdParam) {
  await linkGuidelineToRun({ runId: runIdParam, guidelineId: created.id }).unwrap();
  navigate(`/annotations/runs/${runIdParam}`);
}
```

Similarly, batch link existing guidelines with `Promise.all` and surface per-item failures.

---

### 7. Medium: frontend and embedded-server architecture are in an awkward transitional state

**Problem**

The app still includes a legacy Bootstrap shell at `/`, but the embedded annotation server redirects `/` to `/annotations`. That means the actual user experience depends on how the UI is hosted, not only on the React route table.

**Where to look**

- `ui/src/App.tsx:29-40`, `165-190`
- `pkg/annotationui/server.go:203-223`
- `pkg/annotationui/server_test.go` includes a test expecting `/` to redirect to `/annotations`

**Why it matters**

This is a source of onboarding confusion:

- reading the frontend alone suggests there is a meaningful root app at `/`,
- reading the embedded server suggests `/annotations` is the real entrypoint,
- both are “true” depending on runtime.

That is exactly the kind of transitional ambiguity that grows more expensive over time.

**Cleanup sketch**

Choose one of these explicitly:

1. **Annotation UI is the main app now**
   - remove/deprecate the legacy shell
   - make `/` consistently serve the same routed app
2. **Legacy shell remains first-class**
   - stop redirecting `/` in the embedded server
   - teach the SPA handler to serve the true root app instead

Document the decision in README and route comments.

---

### 8. Medium: `ListReviewFeedback` uses an N+1 query pattern and the new feature has no direct tests

**Problem**

The repository lists feedback rows, then fetches targets in a loop, one feedback item at a time. At the same time, I found no dedicated `_test.go` coverage for the new feedback/guideline methods or routes.

**Where to look**

- `pkg/annotate/repository_feedback.go:110-122`
- command: `rg -n "CreateReviewFeedback|ListReviewFeedback|CreateGuideline|LinkGuidelineToRun|ReviewAnnotationWithArtifacts|BatchReviewWithArtifacts|review-feedback|review-guidelines|guideline" pkg/annotate pkg/annotationui -g '*_test.go'`

**Why it matters**

The N+1 pattern is acceptable for tiny datasets, but this is literally a review surface where per-run or per-mailbox history can grow. And because there are no focused tests for the new endpoints, future refactors could easily break the contract without warning.

**Cleanup sketch**

- Add targeted tests for:
  - create/list/update feedback,
  - create/update guideline,
  - link/unlink guideline to run,
  - review-with-artifacts transaction behavior.
- Consider a joined query or bulk target prefetch when feedback volume grows.

Example bulk load direction:

```go
feedbacks := select * from review_feedback where ...
ids := collect(feedback.ID)
targets := select * from review_feedback_targets where feedback_id in (ids)
group targets by feedback_id
attach grouped targets in memory
```

---

### 9. Low-to-Medium: there is dead or confusing UI/tooling state around this feature

**Problem**

There are several smaller signs of unfinished integration work:

1. the Redux slice keeps review filters/state that are not used by the page,
2. the UI declares `pnpm` as the package manager but now also carries `package-lock.json`,
3. the working tree currently contains generated embed/public asset churn.

**Where to look**

- `ui/src/store/annotationUiSlice.ts:3-12`, `61-88`
- `ui/src/pages/ReviewQueuePage.tsx:32-38` only consumes `selected`, `filterTag`, `expandedId`
- `ui/package.json:6`
- `ui/package-lock.json` and `ui/pnpm-lock.yaml`
- `git status --short pkg/smailnaild/web/embed/public ui/package-lock.json ui/pnpm-lock.yaml`

**Why it matters**

None of these items is a product-breaking bug by itself, but together they increase cognitive noise for a new engineer:

- Which filters are real?
- Which package manager should I use?
- Are embedded assets source-controlled, generated, or both?

**Cleanup sketch**

- delete unused slice fields or wire them into real controls,
- standardize on **one** JS package manager,
- define a clear policy for `pkg/smailnaild/web/embed/public` (generated in CI, committed, ignored, or copied during release only).

## Additional validated findings added after reviewing the intern ticket

This section was added after the later meta-review of the intern’s review ticket. The items below were **not** in the first version of my report, but I independently rechecked them and think they are worth preserving.

### 10. [Added after reviewing the intern ticket] Dead `ReviewCommentInline` component

**Problem**

`ui/src/components/ReviewFeedback/ReviewCommentInline.tsx` exists, is exported, and has Storybook presence, but it is not used by the actual application flow. The live product uses the modal dialog path instead.

**Where to look**

- `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx`
- `ui/src/components/ReviewFeedback/index.ts`
- `rg -n "ReviewCommentInline" ui/src -S`

**Why it matters**

This is exactly the kind of dead branch-local UI experiment that confuses a new engineer. It suggests there are two supported feedback-entry patterns when there is really one.

**Cleanup sketch**

- remove the component and its stories if it is truly abandoned, or
- add a short comment saying it is intentionally parked for a future inline-edit variant.

### 11. [Added after reviewing the intern ticket] Storybook/MSW guideline endpoints drift from the real API

**Problem**

The real app uses `/api/review-guidelines`, but some guidelines page stories override `/api/guidelines`. Those overrides do not match the live endpoint contract, which means the fallback MSW handlers happen to keep the stories working.

**Where to look**

- `ui/src/pages/stories/GuidelinesListPage.stories.tsx`
- `ui/src/pages/stories/GuidelineDetailPage.stories.tsx`
- `ui/src/mocks/handlers.ts`

**Why it matters**

This is not production-breakage, but it weakens Storybook as a trustworthy development environment. A reader of the story file can think they are testing one endpoint while the fallback mock layer is actually serving another.

**Cleanup sketch**

- change story overrides to `/api/review-guidelines...`,
- keep Storybook URL contracts identical to the real RTK Query endpoints.

### 12. [Added after reviewing the intern ticket] MSW create handlers do not persist created entities into mock state

**Problem**

The MSW create handlers for review feedback and guidelines synthesize a new object for the POST response but do not push that object into the backing mock arrays. That means create-then-list flows in Storybook/dev are not realistic.

**Where to look**

- `ui/src/mocks/handlers.ts`
- guideline create path around the `POST /api/review-guidelines` handler
- feedback create path around the `POST /api/review-feedback` handler

**Why it matters**

This can hide integration assumptions. A developer may think “create succeeded and I should now see it in the list” while the mock environment silently drops the new item on the next GET.

**Cleanup sketch**

Use mutable arrays or an explicit mock-state store for create/update/delete endpoints, the same way the run-guideline link mock already maintains mutable state.

### 13. [Added after reviewing the intern ticket] `linkedRunCount` and linked-runs UI should be treated as honest placeholders, not real data

**Problem**

I had already flagged the empty linked-runs section on the guideline detail page, but the intern review also surfaced the related list-page issue: `GuidelineSummaryCard` is being passed `linkedRunCount={0}` everywhere.

**Where to look**

- `ui/src/pages/GuidelinesListPage.tsx`
- `ui/src/components/Guidelines/GuidelineSummaryCard.tsx`
- `ui/src/pages/GuidelineDetailPage.tsx`

**Why it matters**

This is not just “unfinished.” It actively teaches the wrong thing about the system’s maturity. The UI looks like it has run-link analytics when it really has placeholders.

**Cleanup sketch**

- either hide the count until the backend provides it,
- or mark it explicitly as unavailable / not yet implemented.

### 14. [Added after reviewing the intern ticket] duplicated feedback insert logic is worth refactoring once semantics are stable

**Problem**

`CreateReviewFeedback` and `createReviewFeedbackTx` duplicate near-identical feedback/target insert behavior.

**Where to look**

- `pkg/annotate/repository_feedback.go`

**Why it matters**

This is not a top-tier bug, but it is real maintenance debt. Any future column change, validation tweak, or audit-field addition must be remembered in both places.

**Cleanup sketch**

Factor the common insert logic behind a shared helper that accepts an executor/transaction boundary, but only after the higher-priority semantic/API issues are resolved.

## Smaller Notes on the enrich CLI changes

This branch also touches:

- `cmd/smailnail/commands/enrich/all.go`
- `cmd/smailnail/commands/enrich/senders.go`
- `cmd/smailnail/commands/enrich/threads.go`
- `cmd/smailnail/commands/enrich/unsubscribe.go`

These look like incremental usability additions rather than architectural pivots:

- more explicit dry-run behavior,
- optional richer row output for sender/unsubscribe results,
- all-in-one enrichment aggregation.

I did not find a major correctness concern in that smaller slice. The branch’s real review weight is the annotation/review UI work.

## Recommended Implementation Plan

### Phase 1: fix semantic correctness first

1. Make `ReviewQueuePage` query only `reviewState=to_review`.
2. Add `scopeKind` filtering to feedback list API and use it in `RunDetailPage`.
3. Align TS and Go feedback payload types exactly.
4. Await guideline-link mutations before closing UI or navigating.

### Phase 2: restore auditability

1. Add a request-scoped current-user helper.
2. Populate `CreatedBy` and `LinkedBy` in all handlers.
3. Surface authorship more consistently in the UI.

### Phase 3: finish or remove half-built UI surfaces

1. Implement guideline-to-runs detail, or remove the placeholder section.
2. Hide or replace fake `linkedRunCount` data until the backend actually provides it. **[Added after reviewing the intern ticket]**
3. Remove dead `ReviewCommentInline`, or explicitly park it as a future pattern. **[Added after reviewing the intern ticket]**
4. Decide whether the legacy shell still matters.
5. Simplify route behavior so `/` means one thing.

### Phase 4: reduce maintenance drag

1. Remove unused Redux review state. **[Expanded after reviewing the intern ticket]**
2. Fix Storybook guideline endpoint drift and make MSW create handlers persist created entities. **[Added after reviewing the intern ticket]**
3. Refactor duplicated feedback insert logic after the data contract stabilizes. **[Added after reviewing the intern ticket]**
4. Choose pnpm or npm, not both.
5. Add focused tests for feedback/guidelines/transactional review flows.
6. Decide how generated frontend embed assets should be handled.

## Testing Strategy

### What I verified now

```bash
cd smailnail
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1
cd ui
pnpm run check
```

### What should be added next

#### Backend tests

- `TestCreateReviewFeedback`
- `TestListReviewFeedbackFiltersByScope`
- `TestReviewAnnotationWithArtifactsCreatesFeedbackAndLinksGuidelines`
- `TestBatchReviewWithArtifactsRejectsMixedRunGuidelineLinking`
- `TestCreateGuidelineConflictOnDuplicateSlug`
- `TestListGuidelineRuns` (if endpoint is added)

#### Frontend tests or Storybook interaction tests

- Review Queue shows only pending items.
- Run detail run-level feedback excludes annotation-level feedback.
- Create guideline from run waits for link success before navigation.
- Failed guideline link shows visible error state.

## Risks and Open Questions

### Risks

- If the team keeps adding UI on top of the current mixed-scope feedback API, subtle semantic bugs will multiply.
- Missing authorship metadata will become harder to retrofit after data already exists in production databases.
- Transitional route architecture increases the chance of “works in dev, behaves differently in embedded mode” bugs.

### Open questions

1. Is the annotation UI intended to replace the legacy root shell, or coexist with it long-term?
2. Should guidelines apply only to runs, or should they also attach directly to mailbox/sender/domain scopes at runtime?
3. Is feedback expected to be queried primarily by run, by scope, or by mailbox? The API design should follow the dominant use case.
4. Should frontend contract types be generated from a shared schema to prevent the drift already visible here?

## References

### Core repo orientation

- `README.md`
- `cmd/smailnail/main.go`
- `cmd/smailnail/commands/sqlite/serve.go`

### Backend review system

- `pkg/annotate/types.go`
- `pkg/annotate/schema.go`
- `pkg/annotate/repository.go`
- `pkg/annotate/repository_feedback.go`
- `pkg/annotationui/server.go`
- `pkg/annotationui/handlers_annotations.go`
- `pkg/annotationui/handlers_feedback.go`
- `pkg/annotationui/types_feedback.go`

### Frontend review system

- `ui/src/App.tsx`
- `ui/src/api/annotations.ts`
- `ui/src/pages/ReviewQueuePage.tsx`
- `ui/src/pages/RunDetailPage.tsx`
- `ui/src/pages/GuidelinesListPage.tsx`
- `ui/src/pages/GuidelineDetailPage.tsx`
- `ui/src/components/ReviewFeedback/RunFeedbackSection.tsx`
- `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx` **[Added after reviewing the intern ticket]**
- `ui/src/components/RunGuideline/RunGuidelineSection.tsx`
- `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx`
- `ui/src/components/Guidelines/GuidelineSummaryCard.tsx` **[Added after reviewing the intern ticket]**
- `ui/src/store/annotationUiSlice.ts`
- `ui/src/types/reviewFeedback.ts`
- `ui/src/types/reviewGuideline.ts`
- `ui/src/mocks/handlers.ts` **[Added after reviewing the intern ticket]**
- `ui/src/pages/stories/GuidelinesListPage.stories.tsx` **[Added after reviewing the intern ticket]**
- `ui/src/pages/stories/GuidelineDetailPage.stories.tsx` **[Added after reviewing the intern ticket]**

### Related review references

- `ttmp/2026/04/06/SMN-20260406-CODE-REVIEW--code-review-run-review-feedback-guidelines-mailbox-aware-analysis/design-doc/01-comprehensive-code-review-run-review-feedback-guidelines-mailbox-aware-analysis.md` **[Reviewed later; used only to pull in additional findings that were revalidated in source]**

### Validation and test references

- `pkg/annotationui/server_test.go`
- `ui/package.json`
