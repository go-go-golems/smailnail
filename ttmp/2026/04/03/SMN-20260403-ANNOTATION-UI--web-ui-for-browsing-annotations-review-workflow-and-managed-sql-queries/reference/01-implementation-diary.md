---
Title: Implementation Diary
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - react
    - ux-design
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../go-minitrace/web/src/api/minitrace.ts:go-minitrace RTK Query pattern — followed for annotations API
    - Path: ../../../../../../../go-minitrace/web/src/store/store.ts:go-minitrace Redux store — structural reference
    - Path: ../../../../../../../go-minitrace/web/src/theme/theme.ts:go-minitrace theme — aesthetic reference baseline
    - Path: pkg/annotate/types.go:Go backend annotation types — source of truth for TS types
    - Path: ui/.storybook/main.ts:Storybook 8.6 config (commit 7515544)
    - Path: ui/.storybook/preview.tsx:Storybook decorator — MUI ThemeProvider + CssBaseline (commit 7515544)
    - Path: ui/src/App.tsx
      Note: Router + layout (commit 9319d86)
    - Path: ui/src/App.tsx:Root App — BrowserRouter with legacy shell + annotation routes (commit 9319d86)
    - Path: ui/src/api/annotations.ts
      Note: RTK Query API slice (commit 9319d86)
    - Path: ui/src/api/annotations.ts:RTK Query API slice — all endpoints + cache tags (commit 9319d86)
    - Path: ui/src/components/AnnotationTable/AnnotationDetail.tsx:Collapsible detail — full note, related annotations (commit 7cbb343)
    - Path: ui/src/components/AnnotationTable/AnnotationRow.tsx:Table row — checkbox, TargetLink, TagChip, SourceBadge, actions (commit 7cbb343)
    - Path: ui/src/components/AnnotationTable/AnnotationTable.tsx
      Note: Composing table widget (commit 7cbb343)
    - Path: ui/src/components/AnnotationTable/AnnotationTable.tsx:Composing table with select-all, expand, empty state (commit 7cbb343)
    - Path: ui/src/components/AppLayout/AnnotationLayout.tsx:Layout shell — sidebar + Outlet (commit 9319d86)
    - Path: ui/src/components/AppLayout/AnnotationSidebar.tsx:Sidebar nav — Overview, Review, Browse, Tools (commit 9319d86)
    - Path: ui/src/components/shared/index.ts
      Note: 10 shared widgets barrel (commit 7515544)
    - Path: ui/src/components/shared/index.ts:Shared widget barrel — 10 widgets (commit 7515544)
    - Path: ui/src/components/shared/parts.ts:data-part attribute names for shared widgets (commit 7515544)
    - Path: ui/src/mocks/annotations.ts
      Note: Mock data for Storybook (commit 7515544)
    - Path: ui/src/mocks/annotations.ts:Realistic mock data for Storybook (commit 7515544)
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Review queue page wiring (commit 7cbb343)
    - Path: ui/src/pages/ReviewQueuePage.tsx:Review Queue page — wires filters, batch actions, RTK Query (commit 7cbb343)
    - Path: ui/src/store/annotationUiSlice.ts
      Note: Redux UI state (commit 9319d86)
    - Path: ui/src/store/annotationUiSlice.ts:Redux UI state for review queue + query editor (commit 9319d86)
    - Path: ui/src/store/index.ts:Redux store — wired annotationsApi + annotationUi (commit 9319d86)
    - Path: ui/src/theme/tagColors.ts
      Note: Tag color map (commit 7515544)
    - Path: ui/src/theme/tagColors.ts:Per-tag color map with fallback (commit 7515544)
    - Path: ui/src/theme/theme.ts
      Note: MUI dark theme (commit 7515544)
    - Path: ui/src/theme/theme.ts:MUI dark theme — data-observatory aesthetic (commit 7515544)
    - Path: ui/src/types/annotations.ts
      Note: TS types matching Go backend (commit 7515544)
    - Path: ui/src/types/annotations.ts:TypeScript types matching Go backend (commit 7515544)
ExternalSources: []
Summary: Step-by-step implementation diary for the annotation review UI frontend.
LastUpdated: 2026-04-03T16:00:00-04:00
WhatFor: Track implementation decisions, failures, and review instructions for the annotation UI frontend work.
WhenToUse: ""
---


# Diary

## Goal

Capture the step-by-step implementation of the smailnail annotation review UI — a React/MUI/RTK Query frontend following go-minitrace's architecture. Track what worked, what was tricky, and what needs review attention.

## Step 1: Foundation — Theme, Shared Widgets, Storybook, Types, Mocks

Set up the entire Sprint 1 foundation in one batch: MUI dark theme, 10 shared widgets each with Storybook stories, TypeScript types matching the Go backend, mock data, and Storybook infrastructure. This unlocked all subsequent work by providing the building blocks every page and widget depends on.

The existing smailnail UI uses Bootstrap with a hand-rolled CSS theme (`data-widget`/`data-part` conventions). Rather than migrating the existing pages to MUI, the annotation UI lives in its own route tree (`/annotations/*`, `/query`) with its own MUI ThemeProvider. The legacy Bootstrap shell stays untouched at `/`.

### Prompt Context

**User prompt (verbatim):** "Work on docmgr ticket SMN-20260403-ANNOTATION-UI. Work task by task, commit at appropriate intervals, keep an implementation diary. Use storybook, rtk-query, material-ui, react, modular css (see skill)."

**Assistant interpretation:** Implement the annotation review UI frontend, following the task list in the ticket, using MUI + RTK Query + Storybook. Commit incrementally and maintain an implementation diary.

**Inferred user intent:** Produce a working, well-documented frontend that a developer can review via Storybook, with clean commit history and a diary tracking decisions and progress.

**Commit (code):** 7515544 — "ui: add MUI theme, shared widgets, Storybook, types, and mocks"

### What I did
- Installed MUI (`@mui/material`, `@mui/icons-material`, `@emotion/react`, `@emotion/styled`), Storybook 8.6 (`storybook`, `@storybook/react`, `@storybook/react-vite`, `@storybook/addon-essentials`, `@storybook/blocks`, `@storybook/test`), `react-router-dom`, `react-markdown`
- Created `ui/.storybook/main.ts` and `preview.tsx` with MUI ThemeProvider decorator
- Created `ui/src/theme/theme.ts` — MUI dark theme matching go-minitrace's "data-observatory" aesthetic
- Created `ui/src/theme/tagColors.ts` — per-tag color map (10 named tags + fallback)
- Created `ui/src/types/annotations.ts` — TypeScript interfaces for all backend types (Annotation, TargetGroup, AnnotationLog, GroupMember, AgentRunSummary, etc.) plus filter types and query editor types
- Created `ui/src/mocks/annotations.ts` — 7 annotations, 4 logs, 3 runs, 2 groups, 5 senders, 3 presets, 1 query result
- Built 10 shared widgets, each with a `data-part` attribute and Storybook stories:
  - `TagChip` — MUI Chip with tag color lookup
  - `ReviewStateBadge` — warning/success/default chip per review state
  - `SourceBadge` — monospace chip with source_kind icon
  - `TargetLink` — type icon + monospace link
  - `StatBox` — large value + label, configurable color
  - `ReviewProgressBar` — thin segmented bar (reviewed/pending/dismissed)
  - `BatchActionBar` — select-all checkbox + Approve/Dismiss/Reset buttons
  - `FilterPillBar` — clickable pill group with optional counts
  - `CountSummaryBar` — inline stats (e.g. "247 to review · 189 agent")
  - `MarkdownRenderer` — `react-markdown` with MUI-styled code/list/link elements
- Created `ui/src/components/shared/parts.ts` — single source of truth for all `data-part` names
- Added `storybook-static` to `.gitignore`

### Why
- go-minitrace's architecture (MUI + RTK Query + Redux + Storybook) is proven and the design doc (07) prescribes it
- Building shared widgets first means every page can compose from tested building blocks
- Storybook gives immediate visual verification without needing the Go backend running
- The `data-part` convention from the existing smailnail UI and the react-modular-themable-storybook skill provides stable hooks for testing and theming

### What worked
- Theme ported from go-minitrace with minor additions (purple secondary for smailnail, tag color map) — compiles and renders correctly
- All 10 shared widgets compile clean (`npx tsc --noEmit`) and Storybook builds without errors
- Mock data structures match the Go backend types exactly (cross-referenced `pkg/annotate/types.go`)

### What didn't work
- Storybook version mismatch: initially installed `storybook@10.3.4` (latest) but `@storybook/addon-essentials` resolved to `8.6.14` → peer dependency warnings. Fixed by pinning all packages to `^8.6`.

### What I learned
- Storybook 10 is available but addons ecosystem hasn't caught up yet — pin to 8.6 for now
- The existing UI's `data-widget`/`data-part` convention transfers cleanly to MUI via the `data-part` prop on MUI components

### What was tricky to build
- The tag color map needs careful contrast tuning. Each entry has bg/fg/border tuned for dark backgrounds. Rather than computing contrast ratios dynamically, I hardcoded values from the JSX sketch (which were already tested visually). Unknown tags fall back to neutral gray. This is fragile if new tags are added without updating the map — the fallback handles it gracefully but without distinctive colors.

### What warrants a second pair of eyes
- `tagColors.ts` color values — verify contrast ratios meet WCAG AA for the dark background
- `types/annotations.ts` — verify `ReviewState` and `SourceKind` string unions match Go constants exactly (they do: `to_review`/`reviewed`/`dismissed` and `human`/`agent`/`heuristic`/`import`)

### What should be done in the future
- When shared package extraction happens (Phase 3 in design doc 07), the `shared/` widgets, `theme/`, and `parts.ts` need to move into a `@go-go-golems/ui-shared` package
- The mock data should be replaced with MSW (Mock Service Worker) handlers for more realistic API simulation in Storybook

### Code review instructions
- Start in `ui/src/theme/theme.ts` and `tagColors.ts` — verify palette values match go-minitrace
- Then `ui/src/types/annotations.ts` — verify against `pkg/annotate/types.go`
- Then scan any widget in `ui/src/components/shared/` — each is <50 LOC
- Validate: `cd ui && npx tsc --noEmit && npx storybook build --quiet`

### Technical details
- Storybook preview decorator wraps all stories in `<ThemeProvider theme={theme}><CssBaseline />`
- All widgets are presentational (no hooks, no data fetching) — pages own the data layer
- `parts.ts` is the only place `data-part` strings are defined; widgets import from there


## Step 2: RTK Query API, Redux Store, Routing, Sidebar Navigation

Wired the data layer and navigation infrastructure. Created the RTK Query API slice with all endpoints, the Redux UI state slice, updated the store, added react-router with nested routes under `/annotations/*` and `/query`, and built the sidebar layout shell. This made it possible to navigate between placeholder pages and start building real page components.

The trickiest decision was how to integrate react-router into an app that used `useState` for navigation. Rather than rewriting the legacy accounts/mailbox/rules pages to use react-router, I wrapped them in a `LegacyShell` component mounted at `/` and gave the annotation UI its own `AnnotationLayout` at `/annotations`. An "Annotations" button in the legacy shell navigates to the new UI.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue implementing Sprint 1 tasks — API layer, Redux store, routing.

**Inferred user intent:** Complete the infrastructure so page components can fetch data and be reachable via URL.

**Commit (code):** 9319d86 — "ui: add RTK Query API, Redux store, react-router, sidebar navigation"

### What I did
- Created `ui/src/api/annotations.ts` — RTK Query API slice with 16 endpoints across 6 tag types (Annotations, Groups, Logs, Runs, Senders, Queries)
- Created `ui/src/store/annotationUiSlice.ts` — Redux slice for review queue state (selected IDs, filter by tag/type/source/runId, expandedId) and query editor state (sql, activeSourcePath)
- Updated `ui/src/store/index.ts` — wired `annotationsApi.reducer` + `annotationsApi.middleware` + `annotationUiReducer`
- Created `AnnotationLayout` — flex layout with sidebar + `<Outlet />`
- Created `AnnotationSidebar` — 4 sections (Overview, Review, Browse, Tools) with `ListItemButton` + active state from `useLocation()`
- Created 8 placeholder pages in `ui/src/pages/`
- Rewrote `App.tsx` — `BrowserRouter` with nested routes, `LegacyShell` for existing pages at `/`

### Why
- RTK Query cache tags (`invalidatesTags` on mutations) ensure the review queue and run progress bars update atomically after batch review operations
- Sidebar active state from `useLocation()` means bookmarkable URLs work correctly
- Keeping legacy shell at `/` avoids touching the Bootstrap pages

### What worked
- Clean separation: legacy Bootstrap UI at `/`, MUI annotation UI at `/annotations/*`
- RTK Query API mirrors the Go backend endpoint structure from `pkg/smailnaild/http.go`
- TypeScript compiles clean; Storybook still builds with all existing stories

### What didn't work
- N/A — this step was straightforward

### What I learned
- `react-router-dom` v7's `<Outlet />` pattern makes nested layouts clean — each route level renders its own chrome
- RTK Query's `invalidatesTags` is the right abstraction for "batch review updates the run progress bar" — both `Annotations` and `Runs` tags get invalidated by `batchReview`

### What was tricky to build
- The sidebar's active-state detection needs to handle exact match for `/annotations` (dashboard) vs prefix match for all other routes. I used `location.pathname === "/annotations"` for the dashboard item and `location.pathname.startsWith(item.path)` for all others. This works because no route is a prefix of another (e.g., `/annotations/runs` doesn't collide with `/annotations/review`).

### What warrants a second pair of eyes
- API endpoint URLs in `annotations.ts` — these must match whatever the Go backend serves. The backend API doesn't exist yet (task 9–11), so these are forward-declared contracts.
- The `BrowserRouter` setup assumes the Go backend serves `index.html` for all unknown paths (SPA fallback). The existing vite proxy config handles this in dev, but production needs the Go handler to do the same.

### What should be done in the future
- Backend API endpoints (tasks 9–11) need to match the URL paths declared in `api/annotations.ts`
- Consider adding a `NavigationContext` or breadcrumb provider for deep navigation (Sender → Annotation detail → etc.)

### Code review instructions
- Start in `ui/src/api/annotations.ts` — verify endpoint URLs and cache tag strategy
- Then `ui/src/store/index.ts` — verify middleware is wired
- Then `ui/src/App.tsx` — verify route tree structure
- Validate: `cd ui && npx tsc --noEmit`


## Step 3: AnnotationTable Widget and ReviewQueuePage

Built the core review workflow UI: a three-widget composition (AnnotationRow + AnnotationDetail + AnnotationTable) and the ReviewQueuePage that wires everything together with RTK Query hooks and Redux state.

AnnotationRow is a table row that composes TagChip, TargetLink, SourceBadge, and ReviewStateBadge from the shared widgets. AnnotationDetail is a collapsible inline panel that expands below a row showing the full markdown note, metadata, and related annotations on the same target. AnnotationTable composes both with Fragment-based row+detail pairs, select-all logic, and empty state. The ReviewQueuePage connects FilterPillBar, CountSummaryBar, BatchActionBar, and AnnotationTable with RTK Query for data and Redux for UI state.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue to Sprint 2 — build the review queue.

**Inferred user intent:** Get the primary workflow (review annotations) functional end-to-end in the UI.

**Commit (code):** 7cbb343 — "ui: add AnnotationTable widget and ReviewQueuePage"

### What I did
- Created `ui/src/components/AnnotationTable/parts.ts` — data-part names for table/row/detail
- Created `AnnotationRow.tsx` — selectable table row: checkbox, TargetLink, TagChip, truncated note, SourceBadge, ReviewStateBadge, date, action buttons (approve/dismiss/expand)
- Created `AnnotationDetail.tsx` — Collapse-wrapped panel: full MarkdownRenderer, metadata stack, navigate-to-target button, related annotations on same target
- Created `AnnotationTable.tsx` — composing widget: sticky header, select-all checkbox, Fragment-based row+detail pairs, empty state
- Created `stories/AnnotationTable.stories.tsx` — 6 stories: Default, Empty, WithSelection, WithExpanded, WithRelated, Interactive (local useState demo)
- Rewrote `ReviewQueuePage.tsx` — wires useListAnnotationsQuery, useBatchReviewMutation, useReviewAnnotationMutation, Redux selectors/dispatchers, computed tag counts for FilterPillBar, computed summary counts for CountSummaryBar, navigate-to-sender handler

### Why
- The review queue is the primary workflow — most user time is spent here
- Fragment-based row+detail pattern avoids wrapper divs that break MUI Table semantics
- `getRelated` callback lets AnnotationTable find sibling annotations without owning the data

### What worked
- Composing from the shared widgets built in Step 1 — TagChip, SourceBadge, etc. slot in cleanly with no modifications
- TypeScript compiles clean; Storybook builds with all 6 new stories
- The Interactive story (local useState) validates the select/expand/deselect flow without needing Redux

### What didn't work
- N/A

### What I learned
- Fragment-based row+detail pairs in MUI Table require careful `borderBottom` handling on the detail row — when collapsed, the detail's `TableCell` needs `borderBottom: "none"` to avoid a double border
- `useMemo` for tag counts and summary items keeps re-renders manageable when annotations list is large

### What was tricky to build
- The click handling in AnnotationRow requires `stopPropagation` on the checkbox and action buttons cells to prevent the row click (expand toggle) from firing when clicking those controls. Without this, clicking "Approve" would also toggle the expansion state.
- AnnotationDetail's `colSpan={columnCount}` needs to match the table's actual column count exactly. I passed it as a prop rather than hardcoding because the table header might change. But the parent AnnotationTable does hardcode `COLUMN_COUNT = 8`, so it's a single place to update.

### What warrants a second pair of eyes
- `ReviewQueuePage` fires `batchReview` then `clearSelected` synchronously — if the mutation fails, the selection is already cleared and the user loses their selection. Consider optimistic update with rollback, or only clearing on success.
- The `onNavigateTarget` handler only handles `targetType === "sender"`. Other types (domain, message, group) are silently ignored — this needs handling as those pages get built.

### What should be done in the future
- Add optimistic updates to `batchReview` and `reviewAnnotation` mutations (currently fires and waits for cache invalidation)
- Handle all target types in `onNavigateTarget`
- Add keyboard navigation (j/k for row focus, x for select, a/d for approve/dismiss — task 53)

### Code review instructions
- Start in `ui/src/components/AnnotationTable/AnnotationTable.tsx` — see the Fragment-based composition pattern
- Then `AnnotationRow.tsx` — see click handling and shared widget composition
- Then `pages/ReviewQueuePage.tsx` — see RTK Query + Redux wiring
- Validate: `cd ui && npx tsc --noEmit && npx storybook build --quiet`
- Visual: `cd ui && npx storybook dev -p 6006` → navigate to "Annotations/AnnotationTable" → check Interactive story
