---
Title: Annotation UI Review Consistency Playbook
Slug: annotationui-review-consistency-playbook
Short: Keep the sqlite annotation UI consistent across pages by giving every visible artifact an explicit query owner, an invalidation path, and Storybook/MSW coverage.
Topics:
- annotations
- sqlite
- frontend
- backend
- glazed
Commands:
- sqlite serve
- help
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This playbook captures the practical rules for keeping the sqlite-backed annotation UI consistent across review queue, run detail, sender detail, and guideline detail surfaces.

The short version is:

1. every visible artifact on a page must come from an explicit query,
2. every mutation that changes that artifact must invalidate a tag the query provides,
3. Storybook/MSW should model the same mutation ripple effects so the behavior is easy to review before it reaches a real sqlite DB.

## Why This Playbook Exists

The annotation UI is not a thin client over one normalized REST resource. A single user action can change multiple visible artifacts at once.

Examples:

- dismissing one annotation changes the annotation row state,
- dismissing with a comment also creates annotation-scoped feedback,
- dismissing with guideline IDs also links one or more guidelines to the annotation's run,
- batch review changes queue membership, run counters, and possibly feedback/guideline surfaces.

If the page/query/invalidation contract is not explicit, the result is a familiar class of bugs:

- writes succeed but pages stay stale,
- artifacts exist in sqlite but are invisible in the page where the user expects them,
- Storybook looks correct because mocks are static even though the live app drifts after mutations.

## Core Rule: Page = Base Query + Artifact Queries + Invalidation Map

Treat each routed page as a composition boundary.

For each page, answer these three questions before you change anything:

1. What is the base entity?
2. Which extra review artifacts are visible here?
3. Which mutations must refresh those artifacts?

Represent that mentally like this:

```text
Page = base entity query
     + explicit artifact subqueries
     + explicit invalidation map
```

Do not rely on “the detail DTO happens to include enough related state” unless that choice is deliberate and documented.

## Page Ownership Rules

## Review Queue

The queue owns:

- pending annotations,
- pending-only counts and filters,
- review affordances for approving, dismissing, and batch-reviewing.

The queue does **not** need to render every downstream artifact inline, but it must refresh immediately when items leave the pending set.

## Run Detail

Run detail is the reference composed page.

It owns:

- run summary counters,
- annotations belonging to the run,
- run-linked guidelines,
- run-scoped feedback,
- optionally annotation-scoped feedback for whichever annotation is currently expanded.

If another page is unsure how to compose artifacts, copy the run-detail pattern rather than inflating a detail DTO blindly.

## Sender Detail

Sender detail owns:

- sender profile basics,
- sender annotations,
- sender logs and recent messages,
- annotation-scoped feedback for sender-visible annotations,
- run-linked guidelines that are relevant to that sender (typically grouped by run).

Sender detail should not assume that the base sender payload contains every review artifact. It is allowed—and often preferable—to use explicit sender-focused subqueries.

## Guideline Detail

Guideline detail owns:

- guideline fields,
- linked runs,
- create/edit/link flows and their failure states.

If a mutation changes which runs are linked to a guideline, this page must refresh without a manual reload.

## Cache / Invalidation Rules

The repo currently uses broad RTK Query tag families:

- `Annotations`
- `Runs`
- `Senders`
- `Feedback`
- `Guidelines`
- `Groups`
- `Logs`
- `Queries`

That is acceptable for now, but only if every mounted detail query actually **provides** one of the families that the relevant mutations invalidate.

### Minimum checklist for a new or changed query

When you add or modify a query:

- decide which tag family owns it,
- add `providesTags`,
- check whether existing mutations already invalidate that family,
- if not, update those mutations.

### Minimum checklist for a new or changed mutation

When you add or modify a mutation:

- list every page that should visibly change after success,
- identify which query powers that page section,
- make sure the query provides a tag the mutation invalidates.

If the query does not provide a tag, the mutation effectively has no refresh path for that page.

## Artifact Visibility Rules

Two rules matter especially for annotation review work.

### 1. Feedback should be shown where the target is shown

If feedback targets an annotation, then any page that prominently displays that annotation should have a clear way to expose that feedback.

This does not always mean “render the full feedback card inline with no interaction.” It can mean:

- show feedback inside the expanded annotation detail,
- show a summary count with an affordance to expand,
- or link out to a richer feedback view.

But it should not mean “persist it and make it invisible.”

### 2. Guidelines linked to runs should be shown where the run relationship matters

If a guideline is linked to a run, then:

- run detail should show it,
- guideline detail should show the linked run,
- sender detail should show sender-relevant linked guidelines when the sender participates in that run.

Again, grouping matters. On sender detail, guidelines are usually easiest to understand when grouped by run.

## Storybook / MSW Rules

Static mocks are not enough for this subsystem.

Storybook should make it easy to verify these claims:

- dismissing an annotation changes the row state,
- queue membership changes when an annotation leaves `to_review`,
- run counts update after review actions,
- annotation-scoped feedback appears when created,
- run-linked guidelines appear where the page policy says they belong.

To make those claims believable, MSW should keep mutable shared state for:

- annotations,
- feedback,
- guidelines,
- run-guideline links.

If Storybook returns a one-off updated response but the follow-up queries still read from frozen fixtures, it is not modeling the real app.

## Practical Review Checklist

Use this checklist before you call a new review-surface change “done.”

### Backend

- Does the backend expose a query for every artifact the page claims to show?
- If feedback is target-scoped, can the list endpoint filter by target?
- If sender pages need run-linked guidelines, is there a sender-focused read model or endpoint for them?

### Frontend

- Does each displayed artifact come from an explicit query hook?
- Do those hooks provide tags?
- Do all relevant mutations invalidate those tags?
- If a page intentionally links out instead of rendering inline, is that choice obvious in the code and docs?

### Storybook

- Is there at least one story demonstrating the page with the artifact visible?
- Is there at least one stateful story or handler path showing the mutation ripple effect?
- Do handlers reuse the same response shapes as production (`items`, wrapped lists, etc.)?

## Recommended Validation Commands

```bash
buf lint
go generate ./pkg/annotationui
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1
cd ui && pnpm run check
cd ui && pnpm run build-storybook
```

Remove `ui/storybook-static` afterward if you do not intend to keep the build output locally.

## When To Revisit This Playbook

Update this playbook when:

- a new annotation detail page is added,
- a new artifact type becomes visible in the review UI,
- cache invalidation strategy changes substantially,
- the repo moves from broad tag families to entity-scoped tags.
