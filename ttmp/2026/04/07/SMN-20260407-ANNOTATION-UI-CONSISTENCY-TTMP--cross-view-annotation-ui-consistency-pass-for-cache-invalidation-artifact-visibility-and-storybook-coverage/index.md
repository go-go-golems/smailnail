---
Title: Cross-view annotation UI consistency pass for cache invalidation, artifact visibility, and Storybook coverage
Ticket: SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
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
    - Path: ui/src/api/annotations.ts
      Note: Ticket centers on the annotation query and invalidation layer
    - Path: ui/src/mocks/handlers.ts
      Note: Ticket includes a Storybook/MSW truthfulness pass
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Ticket uses run detail as the reference composed-artifact view
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Ticket centers on sender-view artifact visibility and refresh behavior
ExternalSources: []
Summary: Ticket workspace for the broad annotation UI consistency pass covering cache invalidation, explicit artifact visibility across views, and stronger Storybook/MSW proof of cross-view behavior.
LastUpdated: 2026-04-07T10:52:00-04:00
WhatFor: Track the analysis and future implementation work needed to make annotation review artifacts refresh and appear consistently across run, sender, queue, and guideline views.
WhenToUse: Start here to understand the ticket scope, current deliverables, phased tasks, and where the detailed implementation guide lives.
---


# Cross-view annotation UI consistency pass for cache invalidation, artifact visibility, and Storybook coverage

## Overview

This ticket captures a broad consistency pass for the sqlite-backed annotation UI. The immediate trigger was a series of user-visible issues where review actions persisted correctly but did not reliably refresh all affected views, and where persisted artifacts such as annotation feedback or run-linked guidelines were not visible in all of the places a reviewer would reasonably expect.

The core thesis of this ticket is that the annotation system needs an explicit cross-view data policy. Each page should clearly declare which artifacts it shows, which queries power those artifacts, and which mutations must invalidate those queries. This ticket documents the current state, proposes that policy, and lays out a phased implementation plan.

## Primary Documents

- [Design doc: analysis and implementation guide](./design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md)
- [Reference: artifact query and invalidation matrix](./reference/02-artifact-query-and-invalidation-matrix.md)
- [Investigation diary](./reference/01-investigation-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Current Status

- Ticket status: **active**
- New ticket workspace: **created**
- Detailed analysis/design/implementation guide: **written**
- Investigation diary: **written**
- Ticket tasks/changelog/index: **updated**
- Artifact/query/invalidation matrix: **written** in `reference/02-artifact-query-and-invalidation-matrix.md`
- Phase 1 inventory/planning work: **completed in docs, awaiting focused docs commit**
- `docmgr doctor`: **passed**
- reMarkable upload: **completed** (`/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP`)

## Scope Summary

This ticket covers:

1. RTK Query invalidation and refresh behavior across annotation views,
2. explicit visibility rules for feedback and linked guidelines,
3. backend query support for missing artifact surfaces,
4. Storybook/MSW coverage that can prove cross-view consistency,
5. intern-friendly documentation for the whole pass.

## Key Conclusions So Far

- The sqlite annotation server is a distinct subsystem and owns its own frontend/backend consistency contract.
- Run detail and guideline detail already use composed queries; sender detail does not, which is one source of drift.
- Review actions already persist richer artifacts than some pages display.
- Feedback listing is not yet target-addressable, which blocks some sender/annotation-centric feedback views.
- Storybook/MSW currently models mutable feedback and guidelines better than mutable annotations, so it under-tests refresh behavior.

## Structure

- `design-doc/` — main architecture and implementation guide
- `reference/` — investigation diary and future operational notes
- `scripts/` — ticket-local tooling for validation or data setup if implementation begins
- `playbooks/` — reserved for future execution runbooks
- `archive/` — reserved for superseded artifacts
