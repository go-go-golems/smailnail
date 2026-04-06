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
LastUpdated: 2026-04-04T18:35:00-04:00
WhatFor: "Preserve design reasoning and discovery notes."
WhenToUse: "Use when reviewing why the ticket was scoped this way."
---

# Diary

## Goal

Capture the full implementation journey for ticket SMN-20260403-RUN-REVIEW: adding review feedback, guidelines management, and mailbox-aware analysis to the smailnail annotation UI.

---

## Step 1: Research & ticket scoping (2026-04-03)

This step established the problem space. The current UI only supports approve/dismiss state transitions with no reviewer comment payload. Mailbox data exists in storage but isn't surfaced. The investigation confirmed what was missing and what patterns already exist to build on.

### What I did
- Re-read the original annotation UI backend spec to anchor in the existing sqlite architecture.
- Confirmed `ui/src/api/annotations.ts` has no reviewer comment payload — only state transitions.
- Confirmed review queue and run detail pages have approve/dismiss but no text-entry flow.
- Confirmed `messages.mailbox_name` exists in the mirror storage schema; the gap is product surfacing, not storage.
- Confirmed sender detail already joins annotation + message-preview data — a strong pattern reference.
- Added ticket docs, related-file links, and task breakdowns.
- Ran `docmgr doctor --ticket SMN-20260403-RUN-REVIEW` (passed).
- Attempted reMarkable upload — failed with `dial tcp [2600:1901:0:4019::]:443: connect: network is unreachable`.

### Why
- Needed to understand the existing architecture before designing new features.
- Separating reviewer feedback from agent logs enables future queries to distinguish human correction from agent narration.

### What worked
- The existing sender-detail join pattern provided a clear blueprint for the feedback endpoints.
- `docmgr doctor` caught no issues — the ticket structure was sound.

### What didn't work
- reMarkable upload failed due to IPv6 network unreachability.

### What I learned
- The codebase has a consistent pattern: types → RTK endpoints → MSW mocks → component → Storybook stories. Following this order minimizes compile errors.

### What was tricky to build
- N/A (research step).

### What warrants a second pair of eyes
- N/A (no code changed).

### What should be done in the future
- Retry reMarkable upload once network is available.

### Code review instructions
- Review ticket docs at `ttmp/2026/04/03/SMN-20260403-RUN-REVIEW--agent-run-review-requests-guidelines-and-mailbox-aware-analysis-workflow/`.

---

## Step 2: UI design document & phased task list (2026-04-04)

Before writing any component code, produced a 1126-line UI design doc with ASCII wireframes and YAML widget DSL for 7 screens, then a 913-line phased task list with 17 phases and 217 tasks. This upfront design work kept implementation focused and avoided mid-stream redesigns.

**Commit (docs):** `4ff7bb1` — "docs(SMN-20260403-RUN-REVIEW): add UI design doc and phased task breakdown"

### Prompt Context

**User prompt (verbatim):** (from earlier session, user asked to create design doc and task list)

**Assistant interpretation:** Create a detailed UI design document with ASCII wireframes and React widget pseudo-DSL in YAML, then a phased task list for implementation.

**Inferred user intent:** Have a concrete blueprint before writing code so implementation can proceed phase-by-phase without ambiguity.

### What I did
- Created `design/02-ui-design-review-feedback-guidelines-mailbox.md` with 7 screens, ASCII wireframes, YAML widget DSL.
- Created `tasks.md` with 17 phases, 217 tasks.
- Uploaded doc bundle to reMarkable at `/ai/2026/04/04/SMN-20260403-RUN-REVIEW/`.

### Why
- Frontend work benefits from upfront visual design — wireframes prevent debates during implementation.
- Phased tasks keep commits focused and `tsc --noEmit` passing after every phase.

### What worked
- YAML widget DSL made it easy to translate wireframes directly into React component props.
- reMarkable upload succeeded this time.

### What didn't work
- N/A.

### What I learned
- The wireframe→YAML→React pipeline is efficient. Each widget DSL entry maps to one component directory with `parts.ts`, stories, and barrel exports.

### What was tricky to build
- N/A (design step).

### What warrants a second pair of eyes
- N/A (docs only).

### What should be done in the future
- Update the design doc if implementation reveals wireframe gaps.

### Code review instructions
- Read `design/02-ui-design-review-feedback-guidelines-mailbox.md` and `tasks.md` in the ticket directory.

---

## Step 3: TypeScript types & RTK Query contract

**Commit (code):** `baefb8f` — "feat(ui): Phase 3 — TypeScript types & RTK Query contract for feedback/guidelines"

This step laid the data foundation. All new types and API endpoints were defined before any UI code, so components could import from a stable contract. The key insight is that reviewer feedback and guidelines are independent domain objects linked through run-guideline join records.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Proceed with Phase 3 from the task list — create TypeScript types and RTK Query endpoints.

**Inferred user intent:** Build the data layer first so subsequent widget phases have types and hooks to import.

### What I did
- Created `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ui/src/types/reviewFeedback.ts` with `ReviewFeedback`, `FeedbackKind`, `FeedbackStatus`, `ReviewFeedbackTarget`.
- Created `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ui/src/types/reviewGuideline.ts` with `ReviewGuideline`, `GuidelineScope`, `GuidelinePriority`.
- Extended `MessagePreview` with optional `mailboxName` field.
- Extended `AnnotationFilter` with `mailboxName` field.
- Added 10 RTK Query endpoints: feedback CRUD, guidelines CRUD, run-guideline links.
- Extended `reviewAnnotation`/`batchReview` payloads with `comment`, `guidelineIds`, `mailboxName`.
- Added cache tags `Feedback` + `Guidelines`.
- Updated mock messages with `mailboxName`.

### Why
- Types-first ensures compile safety across all later phases.
- RTK Query endpoints with cache tags enable automatic refetching when feedback/guidelines mutate.

### What worked
- `tsc --noEmit` passed clean immediately.
- RTK Query's `providesTags`/`invalidatesTags` pattern means components don't need manual refetch logic.

### What didn't work
- N/A.

### What I learned
- RTK Query `keepUnusedDataFor: 5` on rarely-changing endpoints (guidelines) reduces unnecessary refetches.

### What was tricky to build
- The `useListReviewFeedbackQuery` takes a filter object `{ agentRunId?: string; targetId?: string }` rather than a plain string — needed to match this pattern to RTK Query's `query` function signature.

### What warrants a second pair of eyes
- Verify that `invalidatesTags` on `createReviewFeedback` correctly triggers refetch of `listReviewFeedback` queries that use different filter params.

### What should be done in the future
- When the real backend replaces MSW mocks, verify the endpoint paths match the RTK Query `query` URLs exactly.

### Code review instructions
- Start in `ui/src/types/reviewFeedback.ts` and `ui/src/types/reviewGuideline.ts` for types.
- Then `ui/src/api/annotations.ts` — search for `// ── Review Feedback` and `// ── Review Guidelines`.
- Validate: `cd ui && npx tsc --noEmit`

### Technical details

New endpoints:
```
GET  /feedback?agentRunId=          → useListReviewFeedbackQuery
GET  /feedback/:id                  → useGetReviewFeedbackQuery
POST /feedback                      → useCreateReviewFeedbackMutation
PATCH /feedback/:id                 → useUpdateReviewFeedbackMutation
GET  /guidelines                    → useListGuidelinesQuery
GET  /guidelines/:id                → useGetGuidelineQuery
POST /guidelines                    → useCreateGuidelineMutation
PATCH /guidelines/:id               → useUpdateGuidelineMutation
GET  /runs/:runId/guidelines        → useGetRunGuidelinesQuery
POST /runs/:runId/guidelines        → useLinkGuidelineToRunMutation
DELETE /runs/:runId/guidelines/:id  → useUnlinkGuidelineFromRunMutation
```

---

## Step 4: MSW mock data & handlers

**Commit (code):** `bbb82f5` — "feat(ui): Phase 4 — MSW mock data & handlers for feedback/guidelines"

With the RTK contract in place, added mock data and MSW v2 handlers so Storybook stories and local dev can exercise the full CRUD lifecycle. The key design decision was placing mutable state (`runGuidelineLinks` Map) outside the handlers array — MSW handlers are an array literal, so you can't declare `const` inside it.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phase 4 — create MSW mock data and handlers for all new endpoints.

**Inferred user intent:** Enable Storybook and local dev to work against realistic mock data before a real backend exists.

### What I did
- Created `mockFeedback` (4 items) and `mockGuidelines` (4 items) arrays in `ui/src/mocks/annotations.ts`.
- Added MSW v2 handlers in `ui/src/mocks/handlers.ts` for feedback CRUD, guidelines CRUD, run-guideline links.
- Used mutable `runGuidelineLinks` Map outside the handlers array for POST/DELETE mutability.

### Why
- Storybook stories need working API calls to render loading/success/error states.
- Mock handlers let us test the full feedback lifecycle (create → list → acknowledge → resolve) without a backend.

### What worked
- `tsc --noEmit` passed clean.
- MSW v2's `http.get/post/patch/delete` API is clean and type-safe with the mock data.

### What didn't work
- N/A.

### What I learned
- Mutable state outside the handlers array is the MSW pattern for simulating server-side mutations. The handlers close over the mutable reference.

### What was tricky to build
- The `runGuidelineLinks` Map needed to be initialized with default links from `mockGuidelines` so that `GET /runs/:runId/guidelines` returns meaningful data on first load.

### What warrants a second pair of eyes
- Verify that `runGuidelineLinks.delete()` in the DELETE handler correctly triggers a cache invalidation refetch in consuming components.

### What should be done in the future
- When replacing MSW with real API calls, remove the mutable `runGuidelineLinks` Map entirely — it's purely mock scaffolding.

### Code review instructions
- Start in `ui/src/mocks/annotations.ts` — search for `mockFeedback` and `mockGuidelines`.
- Then `ui/src/mocks/handlers.ts` — search for `/feedback` and `/guidelines` and `/runs/:runId/guidelines`.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 5: Shared badge widgets

**Commit (code):** `330886b` — "feat(ui): Phase 5 — shared badge widgets (MailboxBadge, FeedbackKind, FeedbackStatus, GuidelineScope)"

Four small, reusable badge components that every later widget depends on. Building these first means FeedbackCard, GuidelineCard, and tables can import consistent, styled badges. Each follows the project pattern: `parts.ts` namespace, barrel export, Storybook stories with default/variant/empty states.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phase 5 — create shared badge components.

**Inferred user intent:** Produce the atomic UI primitives that all feedback/guideline widgets will consume.

### What I did
- Created `ui/src/components/shared/MailboxBadge.tsx` — chip/inline variant, icon-per-mailbox, color-coded.
- Created `ui/src/components/shared/FeedbackKindBadge.tsx` — color-coded chip for comment/reject_request/guideline_request.
- Created `ui/src/components/shared/FeedbackStatusBadge.tsx` — chip for open/acknowledged/resolved.
- Created `ui/src/components/shared/GuidelineScopeBadge.tsx` — chip with icon for global/mailbox/pattern/run.
- Added all entries to `shared/parts.ts` and barrel exports from `shared/index.ts`.
- Created Storybook stories for each with default, variant, and empty states.

### Why
- Badges are the visual language for the new domain concepts (feedback kinds, guideline scopes, mailbox identity).
- Centralizing them in `shared/` prevents duplicate styling and inconsistent color choices.

### What worked
- All four badges compiled and rendered in Storybook on first try.
- The `parts.ts` namespace pattern keeps `data-part` attributes consistent.

### What didn't work
- N/A.

### What I learned
- MUI `Chip` with `icon` prop and `size="small"` is the right primitive for status badges in a data-dense UI.

### What was tricky to build
- `MailboxBadge` needed per-mailbox icons (INBOX → InboxIcon, SENT → SendIcon, etc.) which required a mapping object keyed by mailbox name.

### What warrants a second pair of eyes
- Verify the color palette is accessible in dark theme (contrast ratios on `warning.main`, `info.main`, etc.).

### What should be done in the future
- If mailbox names expand beyond INBOX/SENT/DRAFTS/SPAM/TRASH, the icon mapping in `MailboxBadge` needs updating.

### Code review instructions
- Start in `ui/src/components/shared/` — read the four badge files.
- Check `ui/src/components/shared/parts.ts` for namespace entries.
- Check `ui/src/components/shared/stories/` for Storybook stories.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 6: ReviewFeedback widget directory

**Commit (code):** `04ed683` — "feat(ui): Phases 6+8 — ReviewFeedback widget directory (GuidelinePicker, CommentDrawer, FeedbackCard, etc.)"

This was the largest single commit. Created six components in `components/ReviewFeedback/` that together form the feedback creation and display system. Combined Phases 6 and 8 because FeedbackCard and RunFeedbackSection were needed simultaneously for barrel exports to work.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phases 6+8 — create all ReviewFeedback components (GuidelinePicker, ReviewCommentDrawer, ReviewCommentInline, FeedbackCard, RunFeedbackSection, GuidelineLinkPicker).

**Inferred user intent:** Build the complete feedback creation/display widget set so RunDetailPage can integrate it.

### What I did
- Created `GuidelinePicker.tsx` — checkbox list of guidelines for linking to feedback.
- Created `ReviewCommentDrawer.tsx` — MUI Drawer supporting batch/single/run modes, with guideline picker and feedback kind selector.
- Created `ReviewCommentInline.tsx` — compact dismiss-with-reason form.
- Created `FeedbackCard.tsx` — single feedback display with badges, body, acknowledge/resolve actions.
- Created `RunFeedbackSection.tsx` — section wrapper showing feedback list for a run, with "Add Feedback" button.
- Created `GuidelineLinkPicker.tsx` — dialog modal for linking existing guidelines to a run.
- Added `parts.ts` namespace, barrel exports, and Storybook stories for all six.

### Why
- The drawer pattern (vs modal) keeps context visible — reviewers can see selected items while writing feedback.
- "Just Dismiss" fast path must not slow down power users — `ReviewCommentDrawer` supports a skip-comment mode.

### What worked
- `tsc --noEmit` passed clean.
- Barrel exports in `index.ts` mean consumers import from `components/ReviewFeedback` without knowing internal structure.

### What didn't work
- N/A.

### What I learned
- The `ReviewCommentDrawer`'s three modes (batch/single/run) share the same form UI but differ in which mutation they call and what they close. Using a `mode` prop keeps the component unified without branching logic.

### What was tricky to build
- `GuidelineLinkPicker` needs to fetch all guidelines and exclude already-linked ones. The component calls `useListGuidelinesQuery` internally and filters against the `guidelines` prop (already-linked IDs).

### What warrants a second pair of eyes
- `FeedbackCard` action buttons (acknowledge/resolve) call `useUpdateReviewFeedbackMutation` directly — verify the status transition logic matches the backend state machine (open → acknowledged → resolved, no backwards transitions).

### What should be done in the future
- Add `feedbackKind: "guideline_request"` handling — currently the UI creates these but there's no special rendering or auto-guideline-creation flow.

### Code review instructions
- Start in `ui/src/components/ReviewFeedback/index.ts` for the barrel exports.
- Read `ReviewCommentDrawer.tsx` first (most complex), then `FeedbackCard.tsx`, then `RunFeedbackSection.tsx`.
- Check `ui/src/components/ReviewFeedback/stories/` for Storybook stories.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 7: ReviewQueuePage batch reject drawer

**Commit (code):** `f6c8a9d` — "feat(ui): Phase 7 — enhance ReviewQueuePage with Reject & Explain drawer"

Extended the existing review queue with a "Reject & Explain" action that opens the `ReviewCommentDrawer`. This preserves the power-user "Just Dismiss" fast path while adding the option for richer rejection feedback. The drawer keeps the selected annotations visible, so the reviewer doesn't lose context.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phase 7 — wire the ReviewCommentDrawer into ReviewQueuePage for batch reject.

**Inferred user intent:** Enable batch rejection with optional explanation without breaking the existing fast dismiss flow.

### What I did
- Extended `BatchActionBar` with optional `onRejectExplain` callback prop.
- Added `commentDrawerOpen` state to `ReviewQueuePage`.
- Wired `ReviewCommentDrawer` for batch reject flow — "Reject & Explain" opens the drawer.
- "Just Dismiss" fast path still works without opening the drawer.

### Why
- Not every dismissal needs a reason, but when it does, the drawer provides the form without navigating away.

### What worked
- `tsc --noEmit` passed clean.
- The optional `onRejectExplain` prop means `BatchActionBar` works unchanged in pages that don't need the drawer.

### What didn't work
- N/A.

### What I learned
- Optional callback props are the cleanest way to extend existing components without breaking consumers.

### What was tricky to build
- The drawer's `onSubmit` calls `batchReview` with `reviewState: "dismissed"` plus the comment and guideline IDs from the drawer form. Getting the payload shape right required matching the extended `batchReview` mutation type from Step 3.

### What warrants a second pair of eyes
- Verify that the drawer closes and selection resets after batch reject submission — state cleanup is easy to miss.

### What should be done in the future
- Add a toast/snackbar confirmation after batch reject completes.

### Code review instructions
- Start in `ui/src/components/shared/BatchActionBar.tsx` — look for `onRejectExplain` prop.
- Then `ui/src/pages/ReviewQueuePage.tsx` — search for `commentDrawerOpen`.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 8: RunGuideline widget directory

**Commit (code):** `31e567f` — "feat(ui): Phase 9 — RunGuideline widget directory (GuidelineCard, RunGuidelineSection)"

Created the guideline display and linking components for the run detail page. `GuidelineCard` is a compact card showing a guideline with scope badge, priority, and optional unlink action. `RunGuidelineSection` wraps the list and provides "Link Existing" + "Create New" buttons, integrating the `GuidelineLinkPicker` modal from Step 6.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phase 9 — create RunGuideline components.

**Inferred user intent:** Build the widgets that let reviewers see which guidelines apply to a run and manage those links.

### What I did
- Created `ui/src/components/RunGuideline/GuidelineCard.tsx` — compact card with scope badge, status, priority, truncated body, optional unlink button.
- Created `ui/src/components/RunGuideline/RunGuidelineSection.tsx` — section wrapper with linked guideline cards, "Link Existing" + "Create New" buttons.
- Added `parts.ts` namespace, barrel exports, Storybook stories.

### Why
- Reviewers need to see at a glance which guidelines a run is measured against.
- Linking/unlinking should happen inline without navigating to a separate page.

### What worked
- `tsc --noEmit` passed clean.
- `RunGuidelineSection` manages its own `linkGuidelineToRun`/`unlinkGuidelineFromRun` mutations — the consuming page only needs to pass data.

### What didn't work
- N/A.

### What I learned
- The "Create New" button navigates to a guidelines page with a `runId` query param — this avoids embedding a full guideline creation form inside the run detail page.

### What was tricky to build
- `GuidelineCard`'s unlink button calls the parent callback (`onUnlink`) which triggers the mutation in `RunGuidelineSection`. Keeping the mutation at the section level (not the card level) ensures the guideline list refetches correctly.

### What warrants a second pair of eyes
- Verify that unlinking a guideline immediately removes it from the rendered list (RTK Query cache invalidation).

### What should be done in the future
- Add a confirmation dialog before unlinking — currently it's instant.

### Code review instructions
- Start in `ui/src/components/RunGuideline/RunGuidelineSection.tsx`.
- Then `GuidelineCard.tsx`.
- Check `ui/src/components/RunGuideline/stories/`.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 9: RunDetailPage integration

**Commit (code):** `d79e3b2` — "feat(ui): Phase 10 — enhance RunDetailPage with guideline linking and run-level feedback sections"

Integrated `RunGuidelineSection` and `RunFeedbackSection` into the existing run detail page. This step was almost derailed by repeated edit-tool corruption — the exact-text-matching requirement meant that stale file contents produced garbled output. After three failed attempts, restored from git and made four precise edits that compiled clean on first try.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Proceed with Phase 10 — wire RunGuidelineSection and RunFeedbackSection into RunDetailPage.

**Inferred user intent:** Make the run detail page the central hub for reviewing a run — see linked guidelines, view/add feedback, then drill into annotations.

### What I did
- Added `useGetRunGuidelinesQuery` and `useListReviewFeedbackQuery` hooks to `RunDetailPage`.
- Added imports for `RunGuidelineSection` and `RunFeedbackSection`.
- Inserted `<RunGuidelineSection>` between stat boxes and Timeline.
- Inserted `<RunFeedbackSection>` between Timeline and Groups.
- Both components manage their own mutations — no unused hooks needed in the page.

### Why
- The run detail page is the natural place for run-level context: which guidelines apply, what feedback exists.
- Placing guidelines before timeline and feedback after timeline follows the reviewer's mental flow: understand the rules → see what happened → provide feedback.

### What worked
- Final approach: restore from git, make four small targeted edits, verify each with `tsc --noEmit`.
- Both components fetch their own data via hooks passed from the page — clean separation of concerns.

### What didn't work
- **Three failed edit attempts corrupted the file.** The edit tool requires `oldText` to match the file on disk byte-for-byte. I was matching against a stale mental model of the file. Symptoms: duplicate variable declarations, garbled content, syntax errors. Fix: `git checkout HEAD -- ui/src/pages/RunDetailPage.tsx` then re-read the file and match exactly.

### What I learned
- **Always re-read the file immediately before editing.** The file on disk may differ from what you last saw if previous edits were applied or reverted.
- **Make the smallest possible edits.** Four 2–5 line replacements are safer than one 30-line block replacement.

### What was tricky to build
- The edit tool's exact-match requirement when the file has been modified by prior (failed) edits. The file on disk diverges from your mental model, causing a cascade of mismatches.

### What warrants a second pair of eyes
- The placement order (Guidelines → Timeline → Feedback → Groups → Annotations) — verify this matches the reviewer's expected workflow.

### What should be done in the future
- Add RunDetailPage Storybook story that includes the new sections (the existing story only covers the original page).

### Code review instructions
- Start in `ui/src/pages/RunDetailPage.tsx` — verify imports (lines 10–24) and JSX (lines 164–188).
- Run: `cd ui && npx tsc --noEmit`
- Check `git diff 31e567f..d79e3b2 -- ui/src/pages/RunDetailPage.tsx` for the exact changes.

### Technical details

Edit sequence that worked (after git restore):
1. Replace import block (add 4 new imports).
2. Add 2 data-fetching hooks after `useGetRunQuery`.
3. Insert `<RunGuidelineSection>` block before Timeline section.
4. Insert `<RunFeedbackSection>` block before Groups section.

---

## Step 10: Guidelines management pages

**Commit (code):** `3cb43d1` — "feat(ui): Phase 11 — Guidelines management pages"

Created the full guidelines management UI: a `components/Guidelines/` directory with three widgets (GuidelineSummaryCard, GuidelineForm, GuidelineLinkedRuns) plus two new pages (GuidelinesListPage, GuidelineDetailPage). The list page supports status filter chips and search; the detail page supports view/edit/create modes and auto-links new guidelines to runs via a `?runId=` query param.

### Prompt Context

**User prompt (verbatim):** "ok, continue, task by task, committing at appropriate intervals, and keeping a detailed diary as you go"

**Assistant interpretation:** Continue implementing the remaining phases from the task list, committing after each phase, updating the diary after each commit.

**Inferred user intent:** Complete the frontend implementation in a disciplined loop: implement → compile-check → commit → diary-update.

### What I did
- Created `components/Guidelines/parts.ts` and `index.ts` (barrel exports).
- Created `GuidelineSummaryCard.tsx` — displays guideline with scope badge, status chip, priority label, truncated body, linked run count, Edit/Archive/Activate buttons.
- Created `GuidelineForm.tsx` — three modes (view/edit/create) with MUI TextFields, scope select, edit/preview tabs for Markdown body.
- Created `GuidelineLinkedRuns.tsx` — list of run summary rows with annotation/pending counts and navigation.
- Created `pages/GuidelinesListPage.tsx` — uses `useListGuidelinesQuery`, status filter chips (All/Active/Draft/Archived), search field, GuidelineSummaryCard list, "New Guideline" button.
- Created `pages/GuidelineDetailPage.tsx` — detects mode from route (`/new` vs `/:id`), handles create-then-link-to-run flow via `runId` query param.
- Created Storybook stories for all three components and both pages.
- Fixed import path: stories must import from `@storybook/react` (not `@storybook/react-vite`).
- Fixed non-null assertions: `mockGuidelines[0]!` to satisfy strict TypeScript.

### Why
- Guidelines are the reusable policy layer that reviewers reference when providing feedback. They need CRUD management pages.
- The `?runId=` auto-link pattern lets reviewers create a guideline from the run detail page and be returned to the run.

### What worked
- `tsc --noEmit` passed clean after fixing the Storybook import path issue.
- The GuidelineForm edit/preview tabs use MUI Tabs component for a clean UX.

### What didn't work
- First attempt used `import type { Meta, StoryObj } from "@storybook/react-vite"` — this doesn't exist. Fixed to `@storybook/react`.
- `mockGuidelines[0]` returns `ReviewGuideline | undefined` — needed non-null assertion `!` for strict mode.

### What I learned
- The existing codebase consistently uses `@storybook/react` for type imports. Always check existing stories before writing new ones.

### What was tricky to build
- `GuidelineForm`'s three modes share the same JSX but differ in which fields are read-only and which submit payload type they produce (`CreateGuidelineRequest` vs `UpdateGuidelineRequest`). The `canSave` validation differs: create requires slug, edit does not.

### What warrants a second pair of eyes
- `GuidelineDetailPage`'s create-then-link flow: the `createGuideline` mutation returns the new entity, then `linkGuidelineToRun` is called. If the link fails, the guideline exists but isn't linked. Should this be a single server-side operation?

### What should be done in the future
- `GuidelineLinkedRuns` currently receives an empty `runs` array — need a backend endpoint that returns runs linked to a specific guideline.

### Code review instructions
- Start in `ui/src/components/Guidelines/index.ts` for barrel exports.
- Read `GuidelineForm.tsx` (most complex), then `GuidelineSummaryCard.tsx`.
- Check `ui/src/pages/GuidelinesListPage.tsx` and `GuidelineDetailPage.tsx`.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 11: Sidebar, routes & navigation

**Commit (code):** `26cfeed` — "feat(ui): Phase 12 — add Guidelines sidebar entry and routes"

Wired the guidelines pages into the app's navigation: added a "Guidelines" entry with `MenuBookIcon` to the Review section of the sidebar, added three routes to `App.tsx`, and exported the new pages from the barrel. This was a small but critical integration step — without it, the pages exist but can't be reached.

### Prompt Context

**User prompt (verbatim):** (same as Step 10)

### What I did
- Added `MenuBookIcon` import and "Guidelines" nav item to `reviewItems` in `AnnotationSidebar.tsx`.
- Added `GuidelinesListPage` and `GuidelineDetailPage` exports to `pages/index.ts`.
- Added three routes to `App.tsx`: `guidelines`, `guidelines/new`, `guidelines/:guidelineId`.
- Imported new page components in `App.tsx`.

### Why
- Pages without routes are unreachable. The sidebar entry provides discoverability.

### What worked
- Three small targeted edits, all compiled clean on first try.

### What didn't work
- N/A.

### What I learned
- Route order matters: `guidelines/new` must come before `guidelines/:guidelineId` to avoid `:guidelineId` matching "new".

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- Verify that the sidebar's `isActive` check (`location.pathname.startsWith(item.path)`) correctly highlights "Guidelines" when on `/annotations/guidelines/guideline-001`.

### What should be done in the future
- N/A.

### Code review instructions
- `git diff d79e3b2..26cfeed` — three files changed (sidebar, barrel, routes).
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 12: Mailbox context integration (MessagePreviewTable)

**Commit (code):** `d2f25df` — "feat(ui): Phase 13 — add MailboxBadge column to MessagePreviewTable"

Added a Mailbox column to the message preview table in the sender detail page. The `MailboxBadge` component (created in Phase 5) renders a compact inline badge showing which mailbox each message belongs to. Skipped the AnnotationTable and AnnotationDetail mailbox tasks because the `Annotation` type doesn't carry `mailboxName` — that field only exists on `MessagePreview`.

### Prompt Context

**User prompt (verbatim):** (same as Step 10)

### What I did
- Added `MailboxBadge` import to `MessagePreviewTable.tsx`.
- Added "Mailbox" column header after "Subject".
- Added table cell rendering `MailboxBadge` (inline variant) when `msg.mailboxName` is truthy, otherwise a dash.

### Why
- Reviewers need to see which mailbox a message belongs to when reviewing sender-level context.

### What worked
- Clean compile, single targeted edit.

### What didn't work
- Phase 13 tasks 13A/13B (mailbox in AnnotationTable/AnnotationDetail) can't be implemented because `Annotation` doesn't have `mailboxName`. Only `MessagePreview` does.

### What I learned
- The data model doesn't have mailbox provenance on annotations — it's on messages. Future backend work would need to denormalize or join to support annotation-level mailbox filtering.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- Add `mailboxName` to `Annotation` type (backend change) to enable annotation-level mailbox filtering and badges.

### Code review instructions
- `git diff 26cfeed..d2f25df` — one file changed.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 13: Redux slice enhancements

**Commit (code):** `473439e` — "feat(ui): Phase 14 — add commentDrawerOpen and filterMailbox to annotationUiSlice"

Added two new fields to the `ReviewQueueState` slice: `commentDrawerOpen` (boolean) and `filterMailbox` (string | null), with corresponding actions. These enable the ReviewQueuePage to use Redux for drawer state (instead of local `useState`) and support future mailbox filtering.

### Prompt Context

**User prompt (verbatim):** (same as Step 10)

### What I did
- Added `commentDrawerOpen: boolean` and `filterMailbox: string | null` to `ReviewQueueState`.
- Added `openCommentDrawer`, `closeCommentDrawer`, `setFilterMailbox` reducers.
- Exported new actions from the slice.
- Initial values: `commentDrawerOpen: false`, `filterMailbox: null`.

### Why
- Drawer state in Redux enables cross-component coordination (e.g., BatchActionBar opens drawer, ReviewCommentDrawer reads state).
- `filterMailbox` prepares the state for future mailbox filter pills.

### What worked
- Clean compile. Backward-compatible — existing consumers unaffected.

### What didn't work
- N/A.

### What I learned
- Adding optional state to an existing Redux slice is safe as long as defaults match the old behavior.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- Verify that `ReviewQueuePage` still uses local `useState` for `commentDrawerOpen` — it should be migrated to the Redux action in a follow-up.

### What should be done in the future
- Migrate `ReviewQueuePage`'s local `commentDrawerOpen` state to use `useAppSelector(s => s.annotationUi.reviewQueue.commentDrawerOpen)` and dispatch `openCommentDrawer`/`closeCommentDrawer`.

### Code review instructions
- `git diff d2f25df..473439e` — one file changed.
- Validate: `cd ui && npx tsc --noEmit`

---

## Step 14: Backend Phase 2 recovery analysis and execution plan

Before changing more backend code, I stopped to analyze the half-finished Phase 2 worktree and the exact failures from the earlier attempt. The immediate problem was not just a lint error: the implementation drifted away from the design intent by splitting review-state updates from optional feedback/guideline artifacts and then ignoring errors from those later operations.

This step documents the recovery plan before implementation so the rest of the work can proceed in a controlled loop: clean up the duplicated handler/type surface, move review-with-artifacts into transactional repository methods, rewire handlers to call those methods, validate with `gofmt`, `go test`, and `make lint`, and only then commit the backend Phase 2 changes.

### Prompt Context

**User prompt (verbatim):** "Ok, add detailed tasks for these steps to the ticket, then do them task by task. commit at appropriate intervals, and keep your diary. In fact, enter a detailed diary entry now b efore working so that keep track of what is happening."

**Assistant interpretation:** Record the current backend recovery situation in the ticket docs first, then execute the backend cleanup and completion work one task at a time with disciplined commits.

**Inferred user intent:** Prevent more thrashing by explicitly documenting the failure state and recovery plan before continuing implementation.

### What I did
- Inspected the current dirty backend files:
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/handlers_annotations.go`
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/handlers_feedback.go`
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/handlers_senders.go`
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/server.go`
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/types.go`
  - `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotationui/types_feedback.go`
- Confirmed the already-committed backend state: `15de236` (Phase 1 only).
- Confirmed the Vite proxy was already present in `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ui/vite.config.ts` for `/api` and `/auth`.
- Added a new `2R. Backend recovery & hardening` subsection to `tasks.md` to track the recovery work explicitly.

### Why
- The backend worktree already contained useful code, but it was in an unstable intermediate state. Continuing to pile on changes without first documenting the failure modes would make it easier to repeat the same mistakes.
- The design guide explicitly requires review-state update + optional feedback creation + optional guideline linking to succeed or fail together when triggered by one user action.

### What worked
- The inspection clarified that the Vite proxy work was already complete and did not need additional changes.
- The route additions, mailbox propagation, and standalone feedback/guideline handlers are directionally correct and worth preserving.

### What didn't work
- The earlier Phase 2 attempt introduced redundant request types and helper handlers, which caused `unused` lint failures.
- More importantly, `handleReviewAnnotation` and `handleBatchReview` were extended by updating review state first and then doing optional feedback/guideline work with ignored errors (`_ = ...`). That violates the transactional shape described in the design doc.
- The worktree therefore became "compilable in parts" but not ready to commit safely.

### What I learned
- The real bug was semantic, not cosmetic. The lint failures were just a symptom. Even if the earlier code had been cleaned enough to compile, it still would have returned success for partially-applied review actions.
- The correct recovery path is to push the combined operation into the repository layer, not to keep layering conditional logic in handlers.

### What was tricky to build
- The sharp edge here is that the frontend contract and the design doc both treat review + comment + guideline links as one logical action, but the old repository surface only had `UpdateAnnotationReviewState` and `BatchUpdateReviewState`. Extending handlers directly is tempting, but it creates ordering and error-propagation problems immediately.

### What warrants a second pair of eyes
- The transactional repository API shape for single-review vs batch-review artifact creation. The implementation should be reviewed for correctness around failure handling and backward compatibility, not just successful-path behavior.

### What should be done in the future
- After Phase 2 is complete, add repository and handler tests that explicitly simulate failure in feedback creation or guideline linking and verify that review state is not committed independently.

### Code review instructions
- Start with the new recovery checklist in `/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ttmp/2026/04/03/SMN-20260403-RUN-REVIEW--agent-run-review-requests-guidelines-and-mailbox-aware-analysis-workflow/tasks.md` under `2R. Backend recovery & hardening`.
- Compare the dirty worktree against commit `15de236` to understand what Phase 2 is trying to add.
- Review the design section `11. Pseudocode For A Combined Batch Review Flow` in `design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md` before reviewing the repository refactor.

### Technical details
- Existing committed backend baseline: `15de236` — schema/types/repository only.
- Current dirty worktree already contains:
  - standalone feedback handlers
  - standalone guideline handlers
  - run-guideline link handlers
  - sender-detail mailbox propagation
  - non-transactional extensions to review handlers
- Recovery plan:
  1. remove duplicate / unused request types and helper handlers
  2. add transactional repository methods for single + batch review with artifacts
  3. rewire review handlers to call those repository methods
  4. keep standalone endpoints and sender mailbox support
  5. validate with `gofmt`, `go test -tags sqlite_fts5 ./...`, and `make lint`
  6. commit Phase 2 only after all three succeed
