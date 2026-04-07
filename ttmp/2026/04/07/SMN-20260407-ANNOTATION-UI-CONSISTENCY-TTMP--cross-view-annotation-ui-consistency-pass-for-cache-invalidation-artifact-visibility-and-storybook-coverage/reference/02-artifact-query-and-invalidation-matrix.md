---
Title: Artifact query and invalidation matrix
Ticket: SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
    - workflow
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ui/src/api/annotations.ts
      Note: Matrix anchors page-level query ownership and tag invalidation responsibilities
    - Path: ui/src/mocks/handlers.ts
      Note: Matrix captures current Storybook/MSW statefulness limitations
    - Path: ui/src/pages/GuidelineDetailPage.tsx
      Note: Matrix captures guideline-detail linked-run visibility rules
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Matrix captures queue-specific artifact and refresh expectations
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Matrix captures the current composed-query reference implementation
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Matrix captures the missing sender feedback/guideline artifact surfaces
ExternalSources: []
Summary: Canonical matrix for the annotation UI pages, their visible artifact sections, the queries that power those sections, and the mutations that are expected to refresh them.
LastUpdated: 2026-04-07T11:15:00-04:00
WhatFor: Give implementation work a concrete source of truth for which page owns which artifacts and what invalidation/refetch behavior each review mutation must preserve.
WhenToUse: Consult this before changing page composition, backend read models, or RTK Query invalidation for annotation review workflows.
---


# Artifact query and invalidation matrix

## Goal

This reference turns the high-level design doc into a concrete checklist. For each important annotation UI surface, it records:

1. what the user can see,
2. which current query powers that section,
3. which current mutation changes it,
4. whether the current system already refreshes that section correctly,
5. what implementation work is still needed.

## Context

The annotation UI currently mixes two patterns:

- **denormalized detail payloads** (for example `SenderDetail` includes annotations/logs/messages), and
- **composed subqueries** (for example `RunDetailPage` separately requests run guidelines and run feedback).

That inconsistency is the source of much of the current confusion. This matrix exists so implementation work can proceed against a stable page-by-page contract rather than against ad hoc intuition.

## Quick Reference

## A. Route-level surface map

| Route | Main page | Base entity | Primary user intent |
| --- | --- | --- | --- |
| `/annotations/review` | `ReviewQueuePage` | pending annotations | triage and review the queue |
| `/annotations/runs/:runId` | `RunDetailPage` | agent run | inspect one run and review its outputs |
| `/annotations/senders/:email` | `SenderDetailPage` | sender | inspect everything known about one sender |
| `/annotations/guidelines/:guidelineId` | `GuidelineDetailPage` | review guideline | inspect one reusable guideline and its linked runs |
| `/annotations/guidelines/new` | `GuidelineDetailPage` create mode | new guideline | create guideline and optionally link it to a run |

## B. Page artifact matrix

### 1. Review Queue

| Visible section | Current query source | Current mutation(s) that affect it | Current status | Notes |
| --- | --- | --- | --- | --- |
| queue annotations | `useListAnnotationsQuery({ reviewState: "to_review", ... })` | `reviewAnnotation`, `batchReview` | mostly correct | pending-only semantics were already fixed |
| tag counts | second `useListAnnotationsQuery({ reviewState: "to_review" })` | `reviewAnnotation`, `batchReview` | mostly correct | relies on `Annotations` invalidation |
| selection actions | local Redux/UI state | `batchReview` | correct enough | no server artifact query |
| inline guideline/comment affordance | mutation payload only | `batchReview`, `reviewAnnotation` | partly correct | creates artifacts but queue page does not surface them afterward |

### 2. Run Detail

| Visible section | Current query source | Current mutation(s) that affect it | Current status | Notes |
| --- | --- | --- | --- | --- |
| run stats + annotations + logs + groups | `useGetRunQuery(runId)` | `reviewAnnotation`, `batchReview` | fixed recently | now refreshes because `getRun` provides `Runs` |
| linked guidelines | `useGetRunGuidelinesQuery(runId)` | `linkGuidelineToRun`, `unlinkGuidelineFromRun`, review mutations with `guidelineIds` | correct for run scope | good reference pattern |
| run-level feedback | `useListReviewFeedbackQuery({ agentRunId: runId, scopeKind: "run" })` | `createReviewFeedback`, `updateReviewFeedback`, batch/single review if they create run feedback | correct for run scope | intentionally excludes annotation-scoped feedback |
| annotation-expanded related annotations | derived from `run.annotations` | `reviewAnnotation`, `batchReview` | correct if base query refreshes | no annotation feedback displayed yet |

### 3. Sender Detail

| Visible section | Current query source | Current mutation(s) that affect it | Current status | Notes |
| --- | --- | --- | --- | --- |
| sender profile basics | `useGetSenderQuery(email)` | any mutation that changes sender-visible counts/tags | fixed recently for refresh | `getSender` now provides `Senders` |
| sender annotations | `useGetSenderQuery(email)` | `reviewAnnotation`, `batchReview` | fixed recently for refresh | state persists and now refetches |
| sender logs | `useGetSenderQuery(email)` | none in normal review flows | okay | read-only in current scope |
| recent messages | `useGetSenderQuery(email)` | none in review flows | okay | read-only in current scope |
| annotation-scoped feedback for sender-visible annotations | **missing dedicated query** | `reviewAnnotation` with comment | missing | persisted but not shown |
| run-linked guidelines relevant to this sender | **missing dedicated query** | `reviewAnnotation` / `batchReview` with guideline IDs, `linkGuidelineToRun` | missing | persisted but not shown |

### 4. Guideline Detail

| Visible section | Current query source | Current mutation(s) that affect it | Current status | Notes |
| --- | --- | --- | --- | --- |
| guideline fields | `useGetGuidelineQuery(id)` | `createGuideline`, `updateGuideline` | correct | straightforward detail query |
| linked runs | `useGetGuidelineRunsQuery(id)` | `linkGuidelineToRun`, `unlinkGuidelineFromRun`, create-and-link flow | mostly correct | async-link flow was fixed recently |
| create-and-link partial-success errors | local route state | `createGuideline` + `linkGuidelineToRun` | correct enough | now surfaces flash error |

## C. Mutation → expected refresh matrix

| Mutation | Writes | Views that must visibly refresh | Current status |
| --- | --- | --- | --- |
| `reviewAnnotation` without artifacts | annotation review state | review queue, run detail, sender detail, run list counters | now mostly correct |
| `reviewAnnotation` with comment | annotation review state + annotation feedback | all above + annotation feedback surface for the affected annotation | missing annotation feedback surface |
| `reviewAnnotation` with guideline IDs | annotation review state + run-guideline links | all above + run detail guidelines + sender-visible guideline surface + guideline detail linked runs | sender-visible guideline surface missing |
| `batchReview` without artifacts | multiple annotation review states | review queue, run detail, sender detail, run list counters | now mostly correct |
| `batchReview` with comment | multiple states + selection feedback | same plus whatever page owns selection feedback display | currently run-level only in a limited sense |
| `batchReview` with guideline IDs | multiple states + run-guideline links | same plus run/sender/guideline artifact surfaces | sender surface missing |
| `createReviewFeedback` | review feedback row | whichever page lists that feedback scope | correct for run feedback, incomplete for target-based surfaces |
| `updateReviewFeedback` | feedback status | whichever page lists that feedback scope | correct where query exists |
| `createGuideline` | guideline row | guideline list/detail | correct |
| `linkGuidelineToRun` | run-guideline link | run detail, guideline detail, sender-visible guideline surfaces | sender surface missing |
| `unlinkGuidelineFromRun` | remove run-guideline link | same as above | sender surface missing |

## D. Required implementation deltas

### Sender detail must gain explicit artifact queries

The sender page should no longer rely on the base `SenderDetail` payload alone. It needs explicit artifact surfaces for:

1. feedback attached to sender-visible annotations,
2. guidelines linked to runs that produced sender-visible annotations.

### Feedback listing must become target-addressable

The repository stores feedback targets, but the list API currently cannot filter by `targetType` / `targetId`. That is the minimum backend enhancement needed to display annotation-scoped feedback naturally.

### Storybook must own mutable annotation state

Storybook/MSW can currently simulate mutable feedback and guidelines better than mutable annotations. That means it under-tests the class of bugs that motivated this ticket.

## Usage Examples

### Example 1: evaluating a review-state mutation

When changing `reviewAnnotation`, ask:

1. does it alter annotation row state?
2. does it alter run aggregates?
3. does it alter sender detail visible state?
4. does it create feedback?
5. does it create run-guideline links?

Then verify every affected page either:

- provides an invalidated tag, or
- is updated manually.

### Example 2: evaluating a new sender artifact section

When adding sender-visible guidelines, do **not** just render a new panel. First answer:

1. What is the backend query?
2. Is it grouped by run?
3. Which mutation invalidates it?
4. Which Storybook scenario proves it updates after a dismiss-with-guidelines flow?

## Related

- Design doc: `../design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`
- Diary: `./01-investigation-diary.md`
