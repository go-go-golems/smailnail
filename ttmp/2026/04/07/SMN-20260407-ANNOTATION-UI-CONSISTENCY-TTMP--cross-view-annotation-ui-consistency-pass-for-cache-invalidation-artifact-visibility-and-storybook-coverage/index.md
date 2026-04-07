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
    - Path: pkg/annotationui/handlers_senders.go
      Note: Ticket now includes sender-visible guideline endpoint work
    - Path: pkg/doc/annotationui-review-consistency-playbook.md
      Note: Ticket now links to the durable repo playbook created from this pass
    - Path: ui/src/api/annotations.ts
      Note: Ticket centers on the annotation query and invalidation layer
    - Path: ui/src/components/AnnotationTable/AnnotationDetail.tsx
      Note: Ticket now treats expanded annotation detail as the place for both feedback and linked-guideline visibility
    - Path: ui/src/components/SenderProfile/SenderGuidelinePanel.tsx
      Note: Ticket now includes sender artifact rendering work
    - Path: ui/src/mocks/handlers.ts
      Note: Ticket includes a Storybook/MSW truthfulness pass
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Ticket now includes the post-rollout queue dismiss-and-explain parity fix
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Ticket uses run detail as the reference composed-artifact view
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Ticket centers on sender-view artifact visibility and refresh behavior
ExternalSources: []
Summary: Ticket workspace for the broad annotation UI consistency pass covering cache invalidation, explicit artifact visibility across views, and stronger Storybook/MSW proof of cross-view behavior.
LastUpdated: 2026-04-07T12:35:00-04:00
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
- Artifact/query/invalidation matrix: **written and updated after implementation** in `reference/02-artifact-query-and-invalidation-matrix.md`
- Phase 1 inventory/planning work: **completed** (`1a57036`)
- Phase 2 backend contract/read-model work: **completed** (`81f67cd`)
- Phase 3 frontend artifact-visibility work: **completed** (`a684383`)
- Phase 4 cache/tag audit: **completed**, with broad family tags deliberately retained for now
- Phase 5 Storybook/MSW truthfulness work: **completed** (`571cede`)
- Repo playbook/help entry: **added** in `pkg/doc/annotationui-review-consistency-playbook.md`
- Post-rollout UX parity follow-up: **completed** (`b7a3f74`) restoring the queue dismiss-and-explain bubble and showing linked guidelines inside expanded annotation detail
- `docmgr doctor`: **passed**
- reMarkable upload: **completed and refreshed after implementation** (`/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP`)

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
- Feedback listing is now target-addressable, which lets pages ask for annotation-scoped feedback explicitly instead of hiding it.
- Sender detail now treats guidelines and annotation feedback as explicit artifact surfaces rather than hoping the base sender payload is enough.
- Broad RTK Query family tags are still acceptable for this slice, but only because the relevant mounted detail queries now provide those tag families consistently.
- Storybook/MSW now models mutable annotations, feedback, and guideline links well enough to exercise mutation-driven refresh behavior more honestly.
- Expanded annotation detail is now the canonical place to inspect both annotation-scoped feedback and any run-linked guidelines relevant to the currently opened item.

## Structure

- `design-doc/` — main architecture and implementation guide
- `reference/` — investigation diary and future operational notes
- `scripts/` — ticket-local tooling for validation or data setup if implementation begins
- `playbooks/` — reserved for future execution runbooks
- `archive/` — reserved for superseded artifacts
