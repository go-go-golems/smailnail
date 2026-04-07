---
Title: Independent intern-facing code review of task/add-review-ui against origin/main
Ticket: SMN-20260406-INTERN-REVIEW-TTMP
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
    - workflow
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotate/repository_feedback.go
      Note: Follow-up implementation for guideline-linked runs repository query
    - Path: pkg/annotationui/audit_test.go
      Note: Focused audit metadata coverage for finding 4 follow-up work
    - Path: pkg/annotationui/handlers_feedback.go
      Note: Follow-up implementation for guideline-linked runs HTTP endpoint
    - Path: pkg/annotationui/server.go
      Note: Route registration for guideline-linked runs endpoint
    - Path: pkg/mirror/schema.go
      Note: |-
        Follow-up fix for legacy sqlite databases stuck on schema_version 3 without review tables
        Legacy sqlite review databases now auto-upgrade from schema version 3 to 4 on bootstrap
    - Path: pkg/mirror/store_test.go
      Note: |-
        Regression coverage for upgrading legacy schema-version-3 sqlite databases
        Regression coverage for the legacy schema-version-3 upgrade path
    - Path: ui/src/components/ReviewFeedback/GuidelineLinkPicker.tsx
      Note: Finding 6 follow-up async picker behavior now waits for link completion before clearing selection
    - Path: ui/src/components/RunGuideline/RunGuidelineSection.tsx
      Note: Finding 6 follow-up link/unlink error handling and awaited mutation flow
    - Path: ui/src/components/RunGuideline/stories/RunGuidelineSection.stories.tsx
      Note: Finding 6 follow-up story updates for async guideline-link wrapper responses
    - Path: ui/src/pages/GuidelineDetailPage.tsx
      Note: Frontend detail-page wiring for live linked runs
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Finding 1 follow-up implementation making Review Queue query only pending-review items
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Finding 2 follow-up implementation filtering run feedback by scopeKind
    - Path: ui/src/store/annotationUiSlice.ts
      Note: Cleanup of dead review-queue Redux state from finding 9
ExternalSources: []
Summary: Independent code review ticket for the review UI branch, now also tracking and documenting the targeted follow-up implementation work executed from the review findings.
LastUpdated: 2026-04-07T10:20:00Z
WhatFor: Track and publish an independent code review of the task/add-review-ui branch and the targeted follow-up work executed from that review.
WhenToUse: Start here to find the main report, diary, validation notes, and follow-up tasks.
---





# Independent intern-facing code review of task/add-review-ui against origin/main

## Overview

This ticket contains an independent code review of the `task/add-review-ui` branch against `origin/main`, written for a new intern who needs both orientation and actionable critique. The review focuses on the new SQLite-backed annotation/review workflow: review feedback, reusable guidelines, run-to-guideline links, and the frontend pages/components built around them.

I intentionally did **not** use the existing review ticket contents as source material. This workspace stands on direct source inspection, diff inspection, and targeted validation commands only.

## Primary Documents

- [Design doc: intern guide and independent code review](./design-doc/01-intern-guide-and-independent-code-review-of-the-review-ui-branch.md)
- [Diary](./reference/01-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Current Status

- Ticket status: **active**
- Review report: **written and later revised after meta-review of the intern ticket**
- Follow-up finding 1: **implemented** by making Review Queue queries pending-only
- Follow-up finding 2: **implemented** by adding `scopeKind` filtering for run feedback
- Follow-up finding 3: **completed elsewhere** via shared protobuf contract work
- Follow-up finding 4: **implemented** by populating review audit metadata through handlers
- Follow-up finding 5: **implemented** via guideline-linked-runs backend/frontend wiring
- Follow-up finding 6: **implemented** by awaiting guideline-link flows and surfacing failures in the UI
- Follow-up finding 9: **implemented as targeted cleanup** via review-queue state cleanup and fake guideline-count removal
- Follow-up infra fix: **implemented** by splitting sqlite schema bootstrapping into versions 3 and 4 so legacy `schema_version=3` mirror DBs upgrade to the review tables automatically
- Findings 7 and 8: **explicitly deferred for now**
- Diary: **updated with implementation follow-up steps**
- Validation: **completed for the landed follow-up slices** (`go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1`, `pnpm run check`, full pre-commit repo `go test ./...`, `golangci-lint`)
- Delivery to reMarkable: **completed (updated bundle re-uploaded)**

## Main Conclusions

- The branch has a strong product direction around human review of agent-generated annotations.
- The largest correctness issues are semantic rather than compile-time failures.
- The original highest-priority fixes were:
  1. make the Review Queue actually show pending review items,
  2. add `scopeKind` filtering so run-level feedback is not mixed with selection/annotation feedback,
  3. align TypeScript and Go feedback payload contracts,
  4. restore authorship/audit metadata.
- Since the original review, selected follow-up work has now been implemented:
  - finding 1 was fixed by making Review Queue queries pending-only,
  - finding 2 was fixed by adding `scopeKind` filtering so run feedback is no longer mixed with annotation/selection feedback,
  - finding 3 was addressed through shared protobuf wire contracts,
  - finding 4 was fixed by populating review authorship/link metadata at the handler boundary,
  - finding 5 was shipped properly with a real linked-runs endpoint and UI wiring,
  - finding 6 was fixed by awaiting guideline-link mutations and surfacing failures instead of navigating away early,
  - part of finding 9 was cleaned by removing dead review-queue state and fake guideline count wiring,
  - and the sqlite mirror bootstrap path was fixed so older databases already marked as schema version 3 now upgrade to the newer review/guideline tables instead of failing guideline creation with missing-table errors.

## Structure

- `design-doc/` — main report
- `reference/` — diary and operational notes
- `scripts/` — reserved for ticket-local tooling if more analysis is added later
- `archive/` — reserved for future stale artifacts if the ticket evolves
