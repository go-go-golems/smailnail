---
Title: Compact expanded annotation row follow-up
Ticket: SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
Status: active
Topics:
    - annotations
    - frontend
    - ux
    - storybook
DocType: design-doc
Intent: implementation
Owners:
    - manuel
RelatedFiles:
    - Path: ui/package.json
      Note: Confirms Storybook scripts already exist for visual validation
    - Path: ui/src/components/AnnotationTable/AnnotationDetail.tsx
      Note: Primary renderer that will need the compact expanded-row layout adjustment
    - Path: ui/src/components/ReviewFeedback/FeedbackCard.tsx
      Note: Candidate for reuse or replacement in the compact feedback summary design
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Queue page owns the expanded-row experience that this follow-up targets
    - Path: ui/src/pages/stories/ReviewQueuePage.stories.tsx
      Note: Primary Storybook page scenario to update once the compact layout lands
ExternalSources: []
Summary: Small implementation note for tightening the expanded annotation detail row so review feedback remains visible while linked guidelines move inline as compact chips.
LastUpdated: 2026-04-07T13:05:00-04:00
WhatFor: ""
WhenToUse: ""
---


# Compact expanded annotation row follow-up

## Goal

Tighten the expanded annotation detail row used in the review queue so it remains readable but consumes substantially less vertical space.

## Why this follow-up exists

The current expanded row now shows the right artifacts, but it is still visually tall because it stacks:

1. metadata,
2. the full note block,
3. a dedicated review feedback block,
4. a dedicated linked-guidelines block,
5. related annotations.

The desired direction is to keep the **artifact visibility** we just restored while reducing the amount of card chrome and vertical separation.

## Desired review-queue layout

```text
REVIEW QUEUE

┌────┬──────────────────────┬─────────────────┬──────────────────────────────────────────────┬──────────────────────────────┬────────────┬────────────┬──────────────────────┐
│ [] │ Target               │ Tag             │ Note                                         │ Source                       │ Status     │ Date       │ Actions              │
├────┼──────────────────────┼─────────────────┼──────────────────────────────────────────────┼──────────────────────────────┼────────────┼────────────┼──────────────────────┤
│ ☑  │ tango@example.com    │ promo-review-2  │ Synthetic pending review annotation 2-02...  │ Backfill review smoke test B │ Dismissed  │ 2026-04-07 │ ✓  ✕  💬  ▴         │
│    ├──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│    │ [promo-review-2] [Backfill review smoke test B] [Dismissed]  Created 4/7/2026, 10:02:09 AM · Updated 4/7/2026, 3:52:33 PM                          │
│    │ Synthetic pending review annotation 2-02 for tango@example.com                                                                              │
│    │                                                                                                                                              │
│    │ REVIEW FEEDBACK (1)  [Reject Request] [Open]  local-reviewer · 4/7/2026, 3:52:33 PM  (Guidelines: [foo] [foo2])                            │
│    │ meh — not good                                                                                                                               │
│    │                                                                                                                                              │
│    │ OTHER ANNOTATIONS ON THIS TARGET (1)                                                                                                          │
│    │ [promo-review-1]  Synthetic pending review annotation 2-01...                                                                 [Reviewed]   │
├────┼──────────────────────┼─────────────────┼──────────────────────────────────────────────┼──────────────────────────────┼────────────┼────────────┼──────────────────────┤
│ ☐  │ uniform@example.com  │ promo-review-2  │ Synthetic pending review annotation 2-03...  │ Backfill review smoke test B │ To Review  │ 2026-04-07 │ ✓  ✕  💬  ▾         │
└────┴──────────────────────┴─────────────────┴──────────────────────────────────────────────┴──────────────────────────────┴────────────┴────────────┴──────────────────────┘
```

## Non-goals

- Do not hide review feedback behind a second click.
- Do not move linked guidelines back into a separate tall section.
- Do not add a dedicated sender-link row inside the expanded detail.
- Do not redesign the table columns themselves in this slice.

## Proposed implementation

### 1. Remove the dedicated sender-link row

Delete the inline `View sender` action from expanded annotation detail. The target cell in the main table row is already navigable enough for this queue context.

### 2. Collapse feedback + guidelines into one compact feedback summary block

Keep the section label `REVIEW FEEDBACK (N)`, but render the feedback item in a compact layout:

- line 1: kind/status/author/date plus inline guideline chips
- line 2: title/body summary (`meh — not good` style)

If there are multiple feedback entries, each entry can still use the same two-line compact summary block instead of the current larger card format.

### 3. Render linked guidelines as chips, not cards

For this expanded-row context, guidelines should render as lightweight chips or inline tokens:

```text
(Guidelines: [foo] [foo2])
```

The chip text should prefer the guideline slug. Clicking a chip should still navigate to the guideline detail page.

### 4. Keep related annotations as the only separate section below feedback

The related-annotations list can stay as a separate compact section because it represents a different artifact type and benefits from visual separation.

## Likely file touch points

- `ui/src/components/AnnotationTable/AnnotationDetail.tsx`
- `ui/src/components/ReviewFeedback/FeedbackCard.tsx` or a new smaller queue/detail-specific feedback summary renderer
- `ui/src/pages/ReviewQueuePage.tsx`
- possibly `ui/src/pages/RunDetailPage.tsx`
- possibly `ui/src/pages/SenderDetailPage.tsx`

The key design choice is whether to:

1. add a compact mode to `FeedbackCard`, or
2. add a small `AnnotationFeedbackSummary` renderer specifically for expanded annotation detail.

I currently prefer **a dedicated compact summary renderer** if the current `FeedbackCard` API starts fighting the layout, because the queue/detail display goal is significantly denser than the existing card design.

## Storybook status

Yes, the repo already has Storybook.

Relevant commands:

```bash
cd smailnail/ui
pnpm run storybook
pnpm run build-storybook
```

Relevant existing stories:

- `ui/src/pages/stories/ReviewQueuePage.stories.tsx`
- `ui/src/pages/stories/RunDetailPage.stories.tsx`
- `ui/src/pages/stories/SenderDetailPage.stories.tsx`
- `ui/src/components/AnnotationTable/stories/AnnotationTable.stories.tsx`
- `ui/src/components/ReviewFeedback/stories/FeedbackCard.stories.tsx`

## Recommended validation

- Review the queue page in Storybook after the change.
- Expand one dismissed item that has both feedback and linked guidelines.
- Confirm the expanded row:
  - has no sender-link row,
  - shows guideline chips inline with the feedback metadata,
  - keeps the actual feedback content on the line below,
  - still preserves related annotations.

## Suggested implementation task wording

> Tighten the expanded annotation detail row in the review queue by removing the sender-link row and rendering linked guidelines as inline chips inside the compact review-feedback summary, with the actual feedback text on the following line.
