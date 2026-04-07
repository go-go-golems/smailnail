---
Title: Analysis and implementation guide for annotation UI consistency and artifact visibility
Ticket: SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
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
    - Path: cmd/smailnail/commands/sqlite/serve.go
      Note: Entrypoint showing the sqlite annotation server lifecycle and subsystem boundary
    - Path: pkg/annotate/repository_feedback.go
      Note: Repository behavior for feedback listing
    - Path: pkg/annotationui/handlers_feedback.go
      Note: Feedback and guideline endpoints that define the currently available review artifact queries
    - Path: pkg/annotationui/handlers_senders.go
      Note: Backend sender detail contract that currently omits feedback and guideline artifacts
    - Path: pkg/annotationui/server.go
      Note: Route registration and server composition for all annotation review APIs
    - Path: proto/smailnail/annotationui/v1/annotation.proto
      Note: Wire contract showing which sender and run detail fields are currently exposed
    - Path: proto/smailnail/annotationui/v1/review.proto
      Note: Wire contract for feedback
    - Path: ui/src/App.tsx
      Note: Frontend route topology for the annotation surfaces discussed in the design
    - Path: ui/src/api/annotations.ts
      Note: RTK Query endpoint and tag model that drives cache invalidation behavior
    - Path: ui/src/mocks/handlers.ts
      Note: MSW mock layer whose statefulness currently under-models cross-view refresh behavior
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Reference page already using composed artifact queries
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Key example of missing feedback and guideline artifact composition
ExternalSources: []
Summary: Evidence-backed analysis and phased implementation guide for making annotation review views consistent across cache invalidation, artifact visibility, and Storybook coverage.
LastUpdated: 2026-04-07T10:52:00-04:00
WhatFor: Give a new intern enough architecture context and a concrete implementation plan to make the review UI refresh correctly and display feedback/guideline artifacts consistently across views.
WhenToUse: Read this before changing annotation UI data fetching, cache invalidation, sender/run/guideline detail pages, or Storybook/MSW handlers for review workflows.
---


# Analysis and implementation guide for annotation UI consistency and artifact visibility

## Executive Summary

The annotation UI now has the right raw building blocks for review work: a sqlite-backed HTTP server, shared protobuf wire contracts, RTK Query hooks, and distinct pages for runs, senders, guidelines, groups, and the review queue. The core persistence path is not the main problem anymore. Review actions are being written to the database, run-guideline links are persisted, and the review/guideline schema is present on current databases.

The bigger problem is cross-view consistency. Different pages are built from different combinations of denormalized detail payloads and ad hoc side queries, while the frontend cache invalidation strategy is only partially aligned with those query shapes. As a result, a user can perform a correct action and still see stale data, or create feedback/guideline relationships that exist in the database but are not visible in the page where the user expects them.

The most important observations are:

1. The sqlite review server is a dedicated subsystem separate from `smailnaild`, so the annotation UI is responsible for its own full-stack consistency (`cmd/smailnail/commands/sqlite/serve.go:66-116`).
2. The backend already exposes a meaningful set of artifact endpoints for annotations, runs, feedback, guidelines, and run-guideline links (`pkg/annotationui/server.go:149-189`), but not all view-level needs are modeled explicitly.
3. The frontend router shows several detail surfaces (`/annotations/review`, `/annotations/runs/:runId`, `/annotations/senders/:email`, `/annotations/guidelines/:guidelineId`) that each expect a different composition of review artifacts (`ui/src/App.tsx:167-190`).
4. RTK Query currently uses broad string tags (`"Annotations"`, `"Runs"`, `"Senders"`, `"Feedback"`, `"Guidelines"`) rather than entity-scoped tags, which makes consistency possible but fragile (`ui/src/api/annotations.ts:47-260`).
5. The run detail page already composes several artifact queries explicitly, while the sender detail page does not. That asymmetry is the clearest example of the current architectural drift (`ui/src/pages/RunDetailPage.tsx:23-224`, `ui/src/pages/SenderDetailPage.tsx:23-224`).
6. Storybook/MSW coverage is useful but does not yet reliably simulate mutation-driven refreshes across views, because mock annotations are not maintained as a mutable shared state the same way feedback and guidelines are (`ui/src/mocks/handlers.ts:35-69`, `ui/src/mocks/handlers.ts:168-343`).

The recommended direction is not a giant rewrite. Instead, the annotation UI should adopt a clear, documented “view-model query layer” policy:

- each page should declare which artifacts it owns and must display;
- each displayed artifact should have an explicit query source;
- each mutation should invalidate the exact query families whose visible state it changes;
- entity/detail pages should stop relying on implicit side effects of other queries;
- Storybook should mirror those runtime relationships with stateful MSW scenarios.

## Problem Statement and Scope

The user asked for a new ticket and a large consistency pass across the annotation system. The pass must check that views are invalidating, refetching, and rendering the right information, and that artifacts such as feedback and linked guidelines appear wherever the relevant items are shown.

In practical terms, this ticket is about four closely related questions:

1. **Cache consistency:** after a review action, which pages should refresh, and do they actually refresh?
2. **Artifact visibility:** when review feedback or guideline links are created, in which views should those artifacts be visible?
3. **API/view alignment:** do the existing backend endpoints expose enough information to satisfy those visibility rules without hidden assumptions?
4. **Storybook truthfulness:** do stories and MSW handlers illustrate those cross-view invariants, or do they give a false sense of correctness?

### In scope

- sqlite annotation UI backend routes and the data they return;
- RTK Query endpoint design, tag provisioning, and invalidation;
- page-level composition for review queue, run detail, sender detail, and guideline detail;
- feedback and guideline visibility for annotations, runs, senders, and guidelines;
- Storybook stories and MSW handlers used to illustrate cross-view review behavior.

### Out of scope

- replacing RTK Query with a different data library;
- redesigning the annotation schema or switching away from protobuf contracts;
- reworking hosted `smailnaild` account/rule/mailbox pages;
- changing the domain meaning of feedback/guidelines themselves.

## Current-State Architecture

## 1. Runtime topology

The annotation UI is not served by the hosted app backend. It has its own sqlite-oriented command entrypoint, which:

1. opens the mirror store,
2. bootstraps schema,
3. opens a raw sqlite handle,
4. creates the annotation HTTP server,
5. serves both API routes and the SPA (`cmd/smailnail/commands/sqlite/serve.go:66-116`).

That matters because all consistency work here is local to the annotation UI subsystem. There is no external orchestrator reconciling frontend and backend behavior.

### High-level flow

```text
+---------------------------+        +-----------------------------+
| Browser / React SPA       |        | sqlite DB                   |
|                           |        |                             |
|  Pages                    |        | annotations                 |
|  RTK Query hooks          | <----> | review_feedback             |
|  Storybook + MSW          |        | review_feedback_targets     |
+------------+--------------+        | review_guidelines           |
             |                       | run_guideline_links         |
             v                       +-----------------------------+
+---------------------------+
| annotationui HTTP server  |
|                           |
| /api/annotations          |
| /api/annotation-runs      |
| /api/mirror/senders       |
| /api/review-feedback      |
| /api/review-guidelines    |
+---------------------------+
```

## 2. Backend route families

The server registers these major route groups (`pkg/annotationui/server.go:149-189`):

- annotations
- groups
- logs
- runs
- senders
- review feedback
- review guidelines
- run-guideline links
- query editor

This is already a good separation by concept. The problem is that the page/view contract is stronger than the route grouping. Some views need *composed artifacts* that do not map one-to-one to one entity family.

## 3. Frontend route families

The SPA router exposes the following annotation surfaces (`ui/src/App.tsx:167-190`):

- `/annotations` → dashboard
- `/annotations/review` → review queue
- `/annotations/runs` and `/annotations/runs/:runId`
- `/annotations/senders` and `/annotations/senders/:email`
- `/annotations/groups`
- `/annotations/guidelines`, `/new`, `/:guidelineId`
- `/query`

This means the UI has at least four meaningful review contexts:

1. queue-oriented review,
2. run-oriented review,
3. sender-oriented review,
4. guideline-oriented reference/detail.

Each context exposes different user expectations for what “related artifacts” means.

## 4. RTK Query endpoint model

The frontend defines a single `annotationsApi` slice with tag families:

- `Annotations`
- `Groups`
- `Logs`
- `Runs`
- `Senders`
- `Queries`
- `Feedback`
- `Guidelines`

and endpoint hooks for list/detail/mutation operations (`ui/src/api/annotations.ts:47-260`).

This is coherent as a first pass, but it has two structural limitations:

1. tags are broad, not entity-scoped;
2. several pages require view-specific composition rules that are not encoded anywhere except in the page component.

## 5. Artifact creation rules already present in the repository

The repository shows that review actions can create more than a state change.

### Annotation review with artifacts

`ReviewAnnotationWithArtifacts(...)` performs all of the following in one transaction (`pkg/annotate/repository_feedback.go:432-487`):

1. updates the annotation review state;
2. optionally creates annotation-scoped feedback;
3. optionally links one or more guidelines to the annotation’s run.

### Batch review with artifacts

`BatchReviewWithArtifacts(...)` similarly (`pkg/annotate/repository_feedback.go:489-549`):

1. batch-updates review state;
2. optionally creates selection-scoped feedback;
3. optionally links guidelines to a run.

That transactional model is good. It means persistence is already treating review state, feedback, and run-guideline links as related artifacts. The view layer should therefore do the same.

## 6. Current page composition differs sharply by view

### Run detail page

`RunDetailPage` composes multiple queries (`ui/src/pages/RunDetailPage.tsx:23-224`):

- `useGetRunQuery(runId)`
- `useGetRunGuidelinesQuery(runId)`
- `useListReviewFeedbackQuery({ agentRunId: runId, scopeKind: "run" })`

This page already models the idea that the base run detail payload is not enough by itself.

### Sender detail page

`SenderDetailPage` does not do that (`ui/src/pages/SenderDetailPage.tsx:23-224`). It fetches only:

- `useGetSenderQuery(email)`

and then derives UI sections only from:

- `sender.annotations`
- `sender.logs`
- `sender.recentMessages`

It does not fetch sender-visible feedback or sender-applicable guidelines.

### Guideline detail page

`GuidelineDetailPage` also composes additional queries (`ui/src/pages/GuidelineDetailPage.tsx:23-176`):

- `useGetGuidelineQuery(guidelineId)`
- `useGetGuidelineRunsQuery(guidelineId)`
- `useLinkGuidelineToRunMutation()` during create-and-link flow

Again, this page already behaves like a composed artifact view.

## 7. Sender detail backend contract omits feedback and guidelines entirely

The `SenderDetail` protobuf contract currently contains (`proto/smailnail/annotationui/v1/annotation.proto:174-187`):

- sender identity fields,
- annotation count and tags,
- first/last seen,
- repeated annotations,
- repeated logs,
- repeated recent messages.

It does **not** contain:

- review feedback,
- run-guideline links,
- guidelines applicable to the sender through linked runs.

That omission is mirrored in the backend handler (`pkg/annotationui/handlers_senders.go:108-199`), which populates only:

- sender basics,
- sender annotations,
- sender logs,
- recent messages.

So the current system is not merely “forgetting to render” sender feedback/guideline artifacts. For guidelines, the sender detail API does not even expose them.

## 8. Feedback listing is not target-addressable yet

The feedback list filter currently supports only (`pkg/annotate/types.go:217-224`, `pkg/annotate/repository_feedback.go:83-129`, `pkg/annotationui/handlers_feedback.go:13-34`, `ui/src/types/reviewFeedback.ts:69-75`):

- `scopeKind`
- `agentRunId`
- `status`
- `feedbackKind`
- `mailboxName`
- `limit`

It does **not** support filtering by:

- `targetType`
- `targetId`

This is the main backend reason the sender page cannot show annotation-specific feedback in a principled way. The repository already stores feedback targets, but the listing API cannot ask for “feedback attached to annotation X” or “feedback touching sender Y”.

## 9. Storybook/MSW currently under-models cross-view coherence

The Storybook/MSW layer does some good things:

- feedback and guideline collections are mutable (`ui/src/mocks/handlers.ts:25-31`, `ui/src/mocks/handlers.ts:168-343`),
- run-guideline links are modeled as mutable shared state (`ui/src/mocks/handlers.ts:25-28`, `ui/src/mocks/handlers.ts:327-343`).

But annotations are still effectively static:

- `/api/annotations/:id/review` returns a modified annotation payload, but does not update shared annotation state (`ui/src/mocks/handlers.ts:60-65`),
- `/api/annotation-runs/:id` recomputes from `mockAnnotations` rather than a mutable annotation store (`ui/src/mocks/handlers.ts:105-112`),
- `/api/mirror/senders/:email` also recomputes from `mockAnnotations` (`ui/src/mocks/handlers.ts:119-135`).

This makes it hard for Storybook to catch the class of issue the user observed in the live app: “the mutation succeeded, but the page or neighboring view did not visibly update.”

## Gap Analysis

## Gap 1: persistence semantics are richer than page semantics

The repository treats review state, annotation feedback, and run-guideline links as one transactional family, but many pages display only a subset of those artifacts.

### Example

A dismiss-with-comment action on an annotation can create:

- a dismissed annotation,
- annotation-scoped feedback,
- run-level guideline links.

Today:

- the run page can show the run-level guideline links,
- the sender page can show the annotation state,
- but neither page necessarily shows the full artifact set created by that action.

## Gap 2: some views are composed, others are denormalized

The run page and guideline page each compose multiple artifact queries. The sender page does not.

This creates inconsistent engineering expectations:

- on one page, missing data means “add a query”;
- on another page, missing data looks like “the server forgot it” because the detail payload is denormalized.

For a new intern, this inconsistency is confusing and likely to produce accidental omissions.

## Gap 3: RTK Query invalidation is broad but not policy-driven

The current tag strategy can work, but it has no documented ownership model. The repo recently needed fixes so mutations would refresh sender and run detail views. That indicates the system was relying on implicit knowledge rather than a declared rule set.

The missing rule is something like:

> If an action changes the user-visible contents of a view, the query powering that view must either provide an invalidated tag or be updated manually.

That rule is currently inferred, not codified.

## Gap 4: sender-facing artifacts are underspecified

The system has a clear model for run-visible guidelines, but not yet for sender-visible guidelines.

Possible interpretations include:

1. show guidelines linked to any run that produced sender-targeting annotations;
2. group guidelines by run on the sender page;
3. show only guidelines when there is a single run;
4. keep sender pages annotation-centric and link out to runs instead.

The code currently contains no explicit contract answering that question.

## Gap 5: Storybook cannot reliably prove cross-query coherence

Because annotation mutation state is not shared across handlers, the stories under-represent the real challenge. Existing page stories cover loading/not-found/default scenarios well enough, but they do not communicate how review state, feedback, and guideline links should ripple across views.

## Proposed Solution

## 1. Adopt a “view-model query layer” rule

For each routed page, explicitly document:

1. its base entity,
2. its visible artifact sections,
3. the query responsible for each section,
4. the mutations that must invalidate those sections.

### Proposed policy

```text
Page = base entity query + explicit artifact subqueries + explicit invalidation map
```

Do not rely on “whatever happened to be bundled into the detail payload” unless that bundling is deliberate and documented.

## 2. Define an artifact visibility matrix

The repo needs one canonical matrix that says which artifacts belong on which page.

### Recommended v1 matrix

| View | Must show | Should show | Optional / later |
| --- | --- | --- | --- |
| Review Queue | pending annotations, review state changes, comment/guideline review affordances | immediate removal/recount after review | linked artifact preview |
| Run Detail | run stats, annotations, run-level feedback, linked guidelines | selection feedback summary | annotation feedback summary |
| Sender Detail | sender annotations, annotation review states, annotation-scoped feedback for visible annotations | run-linked guidelines grouped by run | mailbox-scoped guidelines |
| Guideline Detail | guideline fields, linked runs | recent feedback mentioning guideline | reverse-linked senders |
| Annotation detail/expanded row | annotation metadata, review state | annotation-scoped feedback | applicable guidelines |

This matrix should live in the ticket docs and later in a permanent playbook/help page.

## 3. Extend feedback listing to support target-based queries

The cleanest backend improvement is to make feedback queryable by target.

### Proposed filter additions

Add to Go + HTTP + TS:

```text
ListFeedbackFilter:
  scopeKind?
  agentRunId?
  status?
  feedbackKind?
  mailboxName?
  targetType?
  targetId?
  limit?
```

### Query semantics

- when `targetType` and `targetId` are present, join through `review_feedback_targets`;
- continue to return full feedback objects with attached targets;
- allow combining target filters with `scopeKind` or `agentRunId` for tighter page queries.

### Why this is the right first step

This exposes information the data model already has, without forcing sender detail to become a giant bespoke DTO.

## 4. Introduce sender-guideline visibility as an explicit backend-backed query

Sender pages need a clear answer for guideline visibility. The cleanest page contract is not to infer on the client by N+1 per-run fetches, but to add a focused backend query.

### Recommended endpoint

```text
GET /api/mirror/senders/{email}/guidelines
```

### Suggested response shape

Either:

```text
message SenderGuidelineLink {
  string run_id;
  ReviewGuideline guideline;
}
message SenderGuidelineListResponse {
  repeated SenderGuidelineLink items;
}
```

or a grouped response:

```text
message SenderRunGuidelines {
  string run_id;
  repeated ReviewGuideline guidelines;
}
message SenderGuidelineListResponse {
  repeated SenderRunGuidelines items;
}
```

The grouped response is probably better for the sender page because one sender may appear in multiple runs.

## 5. Keep run detail as the reference page for composed review artifacts

The run page is closest to the desired architecture already. It should become the reference implementation for future pages:

- base detail query for the entity,
- explicit subqueries for adjacent artifacts,
- narrow filters for the right scope.

The sender page should be brought up to that standard instead of trying to overload `SenderDetail` with every possible relationship.

## 6. Replace broad cache assumptions with entity-aware tags over time

The current broad tags are acceptable for a first cleanup pass, but the long-term shape should be entity-aware.

### Current

```text
invalidatesTags: ["Annotations", "Runs", "Feedback", "Senders"]
```

### Recommended evolution

```text
providesTags: (result, error, id) => [{ type: "Runs", id }, { type: "Annotations", id: "LIST" }]
invalidatesTags: (result, error, arg) => [
  { type: "Annotations", id: arg.id },
  { type: "Runs", id: inferredRunId },
  { type: "Senders", id: inferredSenderEmail },
  { type: "Feedback", id: "LIST" },
]
```

This should not be phase 1 unless the team wants to absorb a larger RTK Query refactor. But it should be the target architecture.

## 7. Make Storybook/MSW stateful across the same artifact boundaries as production

Storybook should be used to prove these invariants:

1. reviewing an annotation changes the annotation row state;
2. the parent run page counters update;
3. sender detail updates when the sender annotation changes;
4. creating annotation feedback makes that feedback visible in the appropriate section;
5. linking guidelines makes them visible in run and sender contexts where policy says they belong.

To do that, MSW must maintain mutable stores for:

- annotations,
- feedback,
- guidelines,
- run-guideline links.

## Proposed Architecture and API Sketches

## A. Page composition model

```text
ReviewQueuePage
  -> listAnnotations({ reviewState: "to_review", ... })
  -> batch/single review mutations
  -> invalidates queue + runs + senders + feedback

RunDetailPage
  -> getRun(runId)
  -> getRunGuidelines(runId)
  -> listReviewFeedback({ agentRunId: runId, scopeKind: "run" })

SenderDetailPage
  -> getSender(email)
  -> listReviewFeedback({ targetType: "annotation", targetId: <visible annotation ids> ... })
     OR sender-scoped artifact endpoint(s)
  -> getSenderGuidelines(email)

GuidelineDetailPage
  -> getGuideline(id)
  -> getGuidelineRuns(id)
```

## B. Backend pseudocode for target-scoped feedback

```go
func (r *Repository) ListReviewFeedback(ctx context.Context, filter ListFeedbackFilter) ([]ReviewFeedback, error) {
    query := `SELECT DISTINCT rf.* FROM review_feedback rf`
    args := []any{}

    if filter.TargetType != "" || filter.TargetID != "" {
        query += ` INNER JOIN review_feedback_targets rft ON rft.feedback_id = rf.id`
    }

    query += ` WHERE 1 = 1`

    if filter.ScopeKind != "" {
        query += ` AND rf.scope_kind = ?`
        args = append(args, filter.ScopeKind)
    }
    if filter.AgentRunID != "" {
        query += ` AND rf.agent_run_id = ?`
        args = append(args, filter.AgentRunID)
    }
    if filter.TargetType != "" {
        query += ` AND rft.target_type = ?`
        args = append(args, filter.TargetType)
    }
    if filter.TargetID != "" {
        query += ` AND rft.target_id = ?`
        args = append(args, filter.TargetID)
    }

    // order, limit, load targets
}
```

## C. Backend pseudocode for sender guidelines

```go
func (h *appHandler) handleListSenderGuidelines(w http.ResponseWriter, r *http.Request) {
    email := r.PathValue("email")

    // 1. find distinct run IDs from annotations where target is this sender
    // 2. join run_guideline_links + review_guidelines
    // 3. return grouped-by-run response
}
```

## D. Frontend pseudocode for sender detail

```tsx
const { data: sender } = useGetSenderQuery(email)
const annotationIds = sender?.annotations.map(a => a.id) ?? []
const { data: senderGuidelines = [] } = useGetSenderGuidelinesQuery(email)
const { data: annotationFeedback = [] } = useListReviewFeedbackQuery({
  targetType: "annotation",
  targetIds: annotationIds, // if API supports multiple ids
})

return (
  <SenderDetailLayout>
    <SenderProfileCard />
    <SenderGuidelinePanel groupedByRun={senderGuidelines} />
    <SenderFeedbackPanel feedback={annotationFeedback} />
    <AnnotationTable annotations={sender.annotations} />
    <AgentReasoningPanel logs={sender.logs} />
    <MessagePreviewTable messages={sender.recentMessages} />
  </SenderDetailLayout>
)
```

If multi-id target queries are too much for phase 1, use a sender-artifact endpoint instead.

## Design Decisions

## Decision 1: prefer explicit subqueries over overstuffed detail DTOs

**Chosen direction:** keep view composition explicit.

**Why:**

- run detail already works this way;
- it scales better than endlessly expanding `SenderDetail` or `AgentRunDetail`;
- it makes invalidation relationships visible in code.

## Decision 2: first close correctness gaps, then optimize tag granularity

**Chosen direction:** make pages refresh and render the right artifacts first; migrate from broad tags to entity-scoped tags afterward.

**Why:** the current user pain is correctness and visibility, not excess refetch volume.

## Decision 3: add Storybook scenarios that prove invariants, not just static states

**Chosen direction:** make MSW shared state realistic enough to show ripple effects.

**Why:** static “default / loading / not found” stories are not sufficient for debugging cache-consistency behavior.

## Decision 4: treat sender-guideline visibility as a product decision encoded in API contracts

**Chosen direction:** add a documented sender-guideline query rather than quietly inferring on the client.

**Why:** a sender may participate in multiple runs; that grouping deserves explicit semantics.

## Alternatives Considered

## Alternative A: keep current endpoints, fix only invalidation

This would help stale refreshes, but it would not solve artifact visibility. Feedback and guidelines that are not queried cannot become visible through invalidation alone.

## Alternative B: expand `SenderDetail` to include all feedback and guidelines

This would be convenient in the short term, but it creates a monolithic DTO with unclear ownership and makes sender detail a dumping ground for cross-cutting review artifacts.

## Alternative C: build one generic “artifact bundle” endpoint for every entity type

This might eventually be valuable, but it is too large for a first pass. The repo does not yet have enough stability in its page semantics to generalize confidently.

## Phased Implementation Plan

## Phase 1 — Inventory and contract alignment

1. Write a permanent artifact visibility matrix covering queue, run, sender, guideline, and annotation surfaces.
2. Enumerate every page-level query and mutation in `ui/src/api/annotations.ts` and each consuming page.
3. Record which tag families each query provides and which mutations invalidate.
4. Identify every visible artifact that currently has no query source.

### Deliverables

- updated ticket docs,
- query/invalidation matrix,
- artifact visibility matrix.

## Phase 2 — Backend support for missing view artifacts

1. Extend feedback filtering to support target-based lookup.
2. Add tests for target-scoped feedback queries.
3. Add sender-guideline listing endpoint (or a sender artifact endpoint if that proves cleaner).
4. Extend protobuf contracts and regenerate Go/TS outputs if new wire messages are introduced.

### Likely files

- `pkg/annotate/types.go`
- `pkg/annotate/repository_feedback.go`
- `pkg/annotationui/handlers_feedback.go`
- `pkg/annotationui/handlers_senders.go`
- `pkg/annotationui/server.go`
- `proto/smailnail/annotationui/v1/*.proto`
- generated Go/TS outputs

## Phase 3 — Frontend composition pass

1. Update `annotationsApi` with any new queries and more deliberate tag ownership.
2. Update `SenderDetailPage` to render feedback and guidelines according to the matrix.
3. Revisit `RunDetailPage`, `ReviewQueuePage`, and `GuidelineDetailPage` for consistency with the same policy.
4. Ensure that approve/dismiss/review-with-comment flows visibly update all mounted affected pages.

### Likely files

- `ui/src/api/annotations.ts`
- `ui/src/pages/SenderDetailPage.tsx`
- `ui/src/pages/RunDetailPage.tsx`
- `ui/src/pages/ReviewQueuePage.tsx`
- new presentational panels/components under `ui/src/components/*`

## Phase 4 — Storybook / MSW truthfulness pass

1. Replace static annotation mocks with mutable shared annotation state.
2. Add stories for:
   - run detail after dismiss-with-feedback,
   - sender detail with annotation feedback present,
   - sender detail with multi-run linked guidelines,
   - review queue refetch/removal semantics,
   - guideline detail showing linked runs updated after linking.
3. Ensure stories use the same wrapper response shapes as production.

### Likely files

- `ui/src/mocks/handlers.ts`
- `ui/src/mocks/annotations.ts`
- `ui/src/pages/stories/*.stories.tsx`
- component-level stories where helpful

## Phase 5 — Validation and handoff

1. Run backend tests for repository and handlers.
2. Run frontend typecheck and Storybook smoke review.
3. Update the ticket diary with findings and edge cases.
4. Capture a durable playbook section for future annotation view additions.

## Testing and Validation Strategy

## Backend validation

- targeted repository tests for target-based feedback filters;
- handler tests for sender guideline listing;
- existing annotation UI tests for review/guideline flows;
- sqlite bootstrap still passing after any proto/handler changes.

### Commands

```bash
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui ./pkg/mirror -count=1
```

## Frontend validation

- `pnpm run check` for type safety;
- manual local smoke runs against a seeded sqlite DB;
- story review for mutation-driven cross-view scenarios.

### Commands

```bash
cd ui
pnpm run check
pnpm run storybook
```

## Suggested manual smoke checklist

1. dismiss annotation from run detail;
2. verify run counters and annotation row update immediately;
3. navigate to sender detail and verify the same annotation state;
4. verify annotation-scoped feedback is visible in sender context;
5. link guidelines during dismiss flow;
6. verify run detail and sender detail both expose the relationship according to the visibility matrix;
7. verify guideline detail shows the linked run.

## Risks, Alternatives, and Open Questions

## Risks

1. **Page-semantic drift:** sender-visible guidelines could become confusing if multi-run grouping is not designed clearly.
2. **Over-invalidating broad tags:** broad cache invalidation may cause extra refetches while the system remains string-tag-based.
3. **Storybook false confidence:** if MSW state remains partially static, stories will still under-report consistency bugs.

## Open questions

1. Should sender pages show guidelines grouped by run, or only surface a summary with links out to run detail?
2. Should annotation-scoped feedback appear directly in `AnnotationTable` expanded rows, in a separate sender/run panel, or both?
3. Is there value in adding entity-scoped RTK Query tags now, or should that wait until the visibility matrix is implemented?
4. Should a future generic artifact endpoint exist, or should the repo keep page-specific subqueries indefinitely?

## References

### Core runtime and routes

- `cmd/smailnail/commands/sqlite/serve.go:66-116`
- `pkg/annotationui/server.go:149-189`
- `ui/src/App.tsx:167-190`

### Frontend query layer

- `ui/src/api/annotations.ts:47-260`
- `ui/src/types/reviewFeedback.ts:69-75`

### Pages

- `ui/src/pages/RunDetailPage.tsx:23-224`
- `ui/src/pages/SenderDetailPage.tsx:23-224`
- `ui/src/pages/GuidelineDetailPage.tsx:23-176`
- `ui/src/pages/ReviewQueuePage.tsx:1-120`

### Backend handlers and repository

- `pkg/annotationui/handlers_senders.go:108-199`
- `pkg/annotationui/handlers_feedback.go:13-245`
- `pkg/annotate/types.go:217-224`
- `pkg/annotate/repository_feedback.go:83-129`
- `pkg/annotate/repository_feedback.go:344-549`

### Contracts and Storybook/MSW

- `proto/smailnail/annotationui/v1/annotation.proto:174-187`
- `proto/smailnail/annotationui/v1/review.proto:17-110`
- `ui/src/mocks/handlers.ts:35-69`
- `ui/src/mocks/handlers.ts:105-135`
- `ui/src/mocks/handlers.ts:168-343`
- `ui/src/pages/stories/RunDetailPage.stories.tsx:1-109`
- `ui/src/pages/stories/SenderDetailPage.stories.tsx:1-100`
