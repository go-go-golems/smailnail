---
Title: "Investigation Diary For Run Review Ticket"
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - diary
    - annotations
    - sqlite
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Chronological notes captured while opening the ticket and shaping the implementation plan."
LastUpdated: 2026-04-04T17:22:00-04:00
WhatFor: "Preserve design reasoning and discovery notes."
WhenToUse: "Use when reviewing why the ticket was scoped this way."
---

# Investigation Diary

## 2026-04-03

- Re-read the original annotation UI backend spec to anchor this ticket in the existing sqlite architecture rather than the older `smailnaild` assumptions.
- Confirmed that the current React API contract only supports state transitions for review. There is no reviewer comment payload in `ui/src/api/annotations.ts`.
- Confirmed that the current review queue and run detail pages expose approve/dismiss actions but no text-entry flow for reviewer correction requests.
- Confirmed that mailbox is already present in the mirror storage schema as `messages.mailbox_name`; the problem is not missing storage but missing end-to-end product surfacing and provenance guidance.
- Confirmed that sender detail already joins annotation and message-preview data, which makes it a strong pattern reference for future review-feedback endpoints.
- Decided that reviewer feedback should be modeled separately from agent/system logs so future queries can distinguish human correction from agent narration.
- Decided to keep this ticket functionality-first and explicitly avoid committing to pixel-level screen design.
- Added detailed ticket docs, related-file links, and task breakdowns, then ran `docmgr doctor --ticket SMN-20260403-RUN-REVIEW`, which passed.
- Attempted reMarkable bundle upload via `remarquee upload bundle ... --remote-dir /ai/2026/04/03/SMN-20260403-RUN-REVIEW --toc-depth 2` after a successful dry-run. The live upload failed with `dial tcp [2600:1901:0:4019::]:443: connect: network is unreachable`, so the ticket docs are ready locally but not confirmed on the device.

## 2026-04-04 — Frontend Implementation

### Phase 3 — TypeScript types & RTK Query contract (commit `baefb8f`)

- Created `types/reviewFeedback.ts` with `ReviewFeedback`, `FeedbackKind`, `FeedbackStatus`, `ReviewFeedbackTarget` types.
- Created `types/reviewGuideline.ts` with `ReviewGuideline`, `GuidelineScope`, `GuidelinePriority` types.
- Extended `MessagePreview` in `types/annotations.ts` with optional `mailboxName` field.
- Extended `AnnotationFilter` with `mailboxName` field.
- Added 10 RTK Query endpoints to `api/annotations.ts`: feedback CRUD (`listReviewFeedback`, `getReviewFeedback`, `createReviewFeedback`, `updateReviewFeedback`), guidelines CRUD (`listGuidelines`, `getGuideline`, `createGuideline`, `updateGuideline`), run-guideline links (`getRunGuidelines`, `linkGuidelineToRun`, `unlinkGuidelineFromRun`).
- Extended `reviewAnnotation` and `batchReview` mutation payloads with optional `comment`, `guidelineIds`, `mailboxName`.
- Added new cache tags: `Feedback`, `Guidelines`.
- Updated mock messages in `mocks/annotations.ts` with `mailboxName` field.
- `tsc --noEmit` passed clean.

### Phase 4 — MSW mock data & handlers (commit `bbb82f5`)

- Created `mockFeedback` (4 items) and `mockGuidelines` (4 items) arrays in `mocks/annotations.ts`.
- Added MSW v2 handlers for all new endpoints: feedback CRUD, guidelines CRUD, run-guideline links (GET/POST/DELETE).
- Used mutable `runGuidelineLinks` Map outside the handlers array so POST/DELETE can mutate it.
- `tsc --noEmit` passed clean.

### Phase 5 — Shared badge widgets (commit `330886b`)

- Created `MailboxBadge` — chip/inline variant, icon-per-mailbox, color-coded.
- Created `FeedbackKindBadge` — color-coded chip for comment/reject_request/guideline_request.
- Created `FeedbackStatusBadge` — chip for open/acknowledged/resolved.
- Created `GuidelineScopeBadge` — chip with icon for global/mailbox/pattern/run.
- All with `parts.ts` entries, barrel exports from `shared/index.ts`, Storybook stories with default/variant/empty states.
- `tsc --noEmit` passed clean.

### Phases 6+8 — ReviewFeedback widget directory (commit `04ed683`)

- Created `components/ReviewFeedback/` directory with:
  - `GuidelinePicker` — checkbox list of guidelines for linking to feedback.
  - `ReviewCommentDrawer` — MUI Drawer supporting batch/single/run modes, with guideline picker and feedback kind selector.
  - `ReviewCommentInline` — compact dismiss-with-reason form.
  - `FeedbackCard` — displays single feedback with badges, body, acknowledge/resolve actions.
  - `RunFeedbackSection` — section wrapper showing feedback list for a run, with "Add Feedback" button.
  - `GuidelineLinkPicker` — dialog modal for linking existing guidelines to a run.
- All with `parts.ts` namespace, barrel exports, Storybook stories.
- `tsc --noEmit` passed clean.

### Phase 7 — ReviewQueuePage batch reject drawer (commit `f6c8a9d`)

- Extended `BatchActionBar` with optional `onRejectExplain` callback prop.
- Wired `ReviewQueuePage` with `commentDrawerOpen` state and `ReviewCommentDrawer` for batch reject flow.
- "Reject & Explain" button opens drawer; "Just Dismiss" fast path still available.
- `tsc --noEmit` passed clean.

### Phase 9 — RunGuideline widget directory (commit `31e567f`)

- Created `components/RunGuideline/` directory with:
  - `GuidelineCard` — compact card showing guideline with scope badge, status, priority, truncated body, optional unlink button.
  - `RunGuidelineSection` — section wrapper with linked guideline cards, "Link Existing" + "Create New" buttons, integrates `GuidelineLinkPicker` modal.
- All with `parts.ts` namespace, barrel exports, Storybook stories.
- `tsc --noEmit` passed clean.

### Phase 10 — RunDetailPage integration (commit `d79e3b2`)

- Added `useGetRunGuidelinesQuery` and `useListReviewFeedbackQuery` hooks to `RunDetailPage`.
- Added imports for `RunGuidelineSection` and `RunFeedbackSection`.
- Inserted `<RunGuidelineSection>` between stat boxes and Timeline sections.
- Inserted `<RunFeedbackSection>` between Timeline and Groups sections.
- Both components manage their own mutations internally — no unused hooks needed in the page.
- **What went wrong**: Multiple failed edit attempts corrupted the file (duplicate declarations, garbled content). Had to restore from git (`git checkout HEAD --`) and redo. The edit tool's exact-match requirement means oldText must be byte-perfect against the file on disk.
- `tsc --noEmit` passed clean.
