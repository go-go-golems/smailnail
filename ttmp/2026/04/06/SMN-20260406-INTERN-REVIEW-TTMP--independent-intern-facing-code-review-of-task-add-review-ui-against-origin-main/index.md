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
RelatedFiles: []
ExternalSources: []
Summary: Independent code review ticket for the review UI branch, now also tracking follow-up implementation work for selected findings after the original review.
LastUpdated: 2026-04-06T23:55:00Z
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
- Follow-up finding 3: **completed elsewhere** via shared protobuf contract work
- Follow-up finding 5: **implemented** via guideline-linked-runs backend/frontend wiring
- Follow-up finding 9: **cleanup in progress / partially implemented** via review-queue state cleanup and fake guideline-count removal
- Findings 7 and 8: **explicitly deferred for now**
- Diary: **updated with implementation follow-up steps**
- Validation: **completed for the landed follow-up slices** (`go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1`, `pnpm run check`)
- Delivery to reMarkable: **completed (updated bundle re-uploaded)**

## Main Conclusions

- The branch has a strong product direction around human review of agent-generated annotations.
- The largest correctness issues are semantic rather than compile-time failures.
- The original highest-priority fixes were:
  1. make the Review Queue actually show pending review items,
  2. add `scopeKind` filtering so run-level feedback is not mixed with selection/annotation feedback,
  3. align TypeScript and Go feedback payload contracts,
  4. restore authorship/audit metadata.
- Since the original review, selected follow-up work has started:
  - finding 3 was addressed through shared protobuf wire contracts,
  - finding 5 was shipped properly with a real linked-runs endpoint and UI wiring,
  - part of finding 9 was cleaned by removing dead review-queue state and fake guideline count wiring.

## Structure

- `design-doc/` — main report
- `reference/` — diary and operational notes
- `scripts/` — reserved for ticket-local tooling if more analysis is added later
- `archive/` — reserved for future stale artifacts if the ticket evolves
