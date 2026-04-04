---
Title: "Frontend Implementation Plan — MUI + RTK Query + Storybook"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - react
    - ux-design
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/sources/smailnail-review-ui.jsx:Imported JSX design sketch — dark-theme annotation review UI with inline styles"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/theme/theme.ts:go-minitrace MUI dark theme — shared aesthetic baseline"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/api/minitrace.ts:go-minitrace RTK Query API slice pattern"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/store/store.ts:go-minitrace Redux store setup"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx:QueryEditor widget — port target for smailnail"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QuerySidebar.tsx:QuerySidebar widget — reusable as-is"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/ResultsTable.tsx:ResultsTable widget — port target"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/SqlEditor.tsx:CodeMirror SQL editor widget"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/SessionBrowser/SessionBrowser.tsx:SessionBrowser — reference for filterable data table pattern"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/shared/ActiveBadge.tsx:Shared badge component pattern"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx:Existing smailnail SPA shell to extend"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Backend annotation types"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/http.go:Existing HTTP handler to extend with API routes"
ExternalSources: []
Summary: "Concrete implementation plan translating the JSX design sketch into production MUI components with RTK Query, Storybook stories, and a shared-package extraction path."
LastUpdated: 2026-04-03T14:00:00.000000000-04:00
WhatFor: "Hand to frontend developers as a self-contained implementation guide with widget decomposition, API contracts, and task ordering"
WhenToUse: ""
---

# Frontend Implementation Plan

## 1. Purpose

This document translates the imported JSX design sketch (`smailnail-review-ui.jsx`) into a production implementation plan using the same architecture as `go-minitrace/web`:

- **MUI (Material UI)** for all components — no inline style objects
- **RTK Query** for server state — API layer with cache tags and invalidation
- **Redux Toolkit** for UI state (selected rows, filter state, expanded panels)
- **Storybook** for every widget — default, themed, loading, empty, error states
- **`data-widget` / `data-part`** attribute convention for stable test/theme hooks
- **Shared package extraction path** — widgets reusable across go-minitrace and smailnail

## 2. Design Sketch Analysis

The imported JSX (`smailnail-review-ui.jsx`) is a single-file prototype with:

### What it demonstrates
- **Dark theme** with amber/slate palette matching go-minitrace's "data observatory" aesthetic
- **Sidebar nav** with sections (Overview, Review, Browse, Tools) and pending-count badge
- **8 views:** Dashboard, Review Queue, Agent Runs, Run Detail, Senders, Sender Detail, Groups, SQL Workbench
- **Tag chips** with per-tag color schemes (newsletter=blue, notification=orange, important=green, etc.)
- **Review badges** (to_review=amber, reviewed=green, dismissed=red)
- **Batch selection** with checkbox multi-select and bulk approve/dismiss
- **Source badges** with monospace code styling
- **Stat boxes** for dashboard and detail headers
- **Agent log timeline** in Run Detail
- **SQL workbench** with preset sidebar, textarea editor, mock results table

### What needs to change for production
1. **Inline styles → MUI `sx` prop + theme tokens.** The sketch uses a global `S` object with raw CSS. These map cleanly to MUI's `sx` prop with theme references (`palette.primary.main`, etc.)
2. **Mock data → RTK Query hooks.** All `ANNOTATIONS`, `SENDERS`, `GROUPS`, `LOGS`, `AGENT_RUNS` constants become API queries.
3. **`useState` navigation → react-router.** The sketch uses `useState("dashboard")` for routing. Production uses `react-router` with URL-based navigation.
4. **`textarea` → CodeMirror.** The SQL editor textarea becomes the CodeMirror 6 `SqlEditor` widget (port from go-minitrace).
5. **Tag colors → theme-integrated tag palette.** The sketch hardcodes tag colors; production derives them from a configurable map in the MUI theme.
6. **No Storybook stories.** Every widget needs stories.

## 3. Theme

Extend go-minitrace's MUI theme. The sketch's `palette` object maps almost 1:1:

| Sketch token | MUI theme path | Value |
|---|---|---|
| `palette.bg` | `palette.background.default` | `#0d1117` (use go-minitrace) |
| `palette.bgRaised` | `palette.background.paper` | `#161b22` (use go-minitrace) |
| `palette.accent` | `palette.primary.main` | `#f5a623` (go-minitrace amber) |
| `palette.green` | `palette.success.main` | `#3fb950` (go-minitrace) |
| `palette.red` | `palette.error.main` | `#f85149` (go-minitrace) |
| `palette.blue` | `palette.info.main` | `#58a6ff` (go-minitrace) |
| `palette.orange` | `palette.warning.main` | `#d29922` (go-minitrace) |
| `palette.purple` | `palette.secondary.main` | `#a78bfa` (smailnail addition) |
| `palette.text` | `palette.text.primary` | `#e6edf3` |
| `palette.textMuted` | `palette.text.secondary` | `#8b949e` |
| `palette.border` | `palette.divider` | `#30363d` |

**Tag color map** — added to theme as a custom key:

```typescript
// theme/tagColors.ts
export const tagColors: Record<string, { bg: string; fg: string; border: string }> = {
  newsletter:       { bg: "#1e293b", fg: "#60a5fa", border: "#334155" },
  notification:     { bg: "#1c1917", fg: "#fb923c", border: "#3f3224" },
  "bulk-sender":    { bg: "#1a1625", fg: "#a78bfa", border: "#2e2544" },
  important:        { bg: "#14231a", fg: "#4ade80", border: "#1e3a28" },
  transactional:    { bg: "#1a1a14", fg: "#d4a053", border: "#33301e" },
  ignore:           { bg: "#1c1115", fg: "#f87171", border: "#3b1a22" },
  "action-required": { bg: "#231414", fg: "#f87171", border: "#3b1a1a" },
};
// fallback: palette.action.selected bg, text.secondary fg
```

## 4. Widget Inventory

Every widget listed below gets:
- A `.tsx` component file
- A `.stories.tsx` Storybook file
- `data-widget` on the root element (for page-level widgets) or `data-part` (for sub-widgets)
- Props-only interface (no internal data fetching — pages compose widgets with RTK Query hooks)

### 4.1 Shared Widgets (candidates for cross-project package)

These widgets are generic enough to share between go-minitrace and smailnail. They belong in `ui/src/components/shared/` initially, with the plan to extract into a `@smailnail/ui-shared` package.

| Widget | `data-widget` / `data-part` | Sketch source | Description | MUI components used |
|---|---|---|---|---|
| **TagChip** | `data-part="tag-chip"` | `Tag()` | Colored chip from tag color map | `Chip` |
| **ReviewStateBadge** | `data-part="review-badge"` | `ReviewBadge()` | Yellow/green/gray chip | `Chip` |
| **SourceBadge** | `data-part="source-badge"` | `S.code` spans | Monospace badge showing source_label | `Chip` with `sx={{ fontFamily: "monospace" }}` |
| **TargetLink** | `data-part="target-link"` | `TargetLink()` | Type icon + monospace ID, clickable | `Link`, `Typography` |
| **StatBox** | `data-part="stat-box"` | `StatBox()` | Large number + label, colored | `Box`, `Typography` |
| **ReviewProgressBar** | `data-part="review-progress"` | progress bar in Agent Runs | Thin bar showing reviewed/pending/dismissed ratio | `LinearProgress` or custom `Box` |
| **BatchActionBar** | `data-part="batch-bar"` | checkbox + Approve/Dismiss buttons | Select-all checkbox, count, action buttons | `Checkbox`, `Button`, `Stack` |
| **CountSummaryBar** | `data-part="count-summary"` | "247 to review · 189 agent · 58 heuristic" | Inline stats with chip-like counts | `Typography`, `Chip` |
| **FilterPillBar** | `data-part="filter-pills"` | tag/type filter pills | Clickable pills for filter dimensions | `ToggleButtonGroup` or custom `Chip` group |
| **MarkdownRenderer** | `data-part="markdown-body"` | `whiteSpace: "pre-wrap"` text | Renders markdown to HTML (for log bodies) | `Box` + `react-markdown` |

**Widgets already in go-minitrace to port/share:**

| Widget | Source | Action |
|---|---|---|
| **SqlEditor** | `go-minitrace/web/src/components/QueryEditor/SqlEditor.tsx` | Copy into shared, no changes |
| **QuerySidebar** | `go-minitrace/web/src/components/QueryEditor/QuerySidebar.tsx` | Copy into shared, no changes |
| **ResultsTable** | `go-minitrace/web/src/components/QueryEditor/ResultsTable.tsx` | Copy, add annotation-aware ID clicking |
| **QueryEditor** | `go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx` | Copy, minor adaptation for smailnail API |

### 4.2 Annotation-Specific Widgets

| Widget | `data-widget` | Sketch source | Description | Key props |
|---|---|---|---|---|
| **AnnotationTable** | `data-widget="annotation-table"` | `ReviewQueueView` table | Selectable table of annotations with inline expand | `annotations`, `selected`, `expandedId`, `onToggle`, `onExpand`, `onReview` |
| **AnnotationRow** | `data-part="annotation-row"` | `<tr>` in ReviewQueueView | Single table row: checkbox, target link, tag, note, source, actions | `annotation`, `isSelected`, `onToggle`, `onApprove`, `onDismiss` |
| **AnnotationDetail** | `data-part="annotation-detail"` | Not in sketch (from v2 spec) | Expanded inline panel: full note, related annotations, navigation | `annotation`, `relatedAnnotations`, `onNavigate`, `onReview` |
| **RunTimeline** | `data-widget="run-timeline"` | Log timeline in `RunDetailView` | Chronological log entries with time column, kind badge, body, linked targets | `logs`, `onNavigateTarget` |
| **GroupCard** | `data-widget="group-card"` | `GroupsView` card | Group header + description + member list | `group`, `members`, `onNavigateTarget` |
| **SenderProfileCard** | `data-widget="sender-profile"` | `SenderDetailView` header | Stat boxes row: messages, domain, dates, unsubscribe status | `sender` |
| **SenderAnnotationList** | `data-part="sender-annotations"` | Annotations table in SenderDetail | Annotations for a single sender with review actions | `annotations`, `onReview` |
| **AgentReasoningPanel** | `data-part="agent-reasoning"` | "Agent Reasoning" card in SenderDetail | Related log entries with title + body + timestamps | `logs` |
| **MessagePreviewTable** | `data-part="message-preview"` | "Recent Messages" table in SenderDetail | Compact message list: date, subject, size | `messages`, `onSelect` |
| **DashboardStatGrid** | `data-widget="dashboard-stats"` | Dashboard stat cards row | 6-column stat grid | `stats: {label, value, color}[]` |
| **LatestRunBanner** | `data-part="latest-run-banner"` | Amber banner in Dashboard | Highlighted CTA for latest unreviewed run | `run`, `pendingCount`, `onReview`, `onInspect` |
| **RecentActivityList** | `data-part="recent-activity"` | "Recent Agent Activity" in Dashboard | Compact log timeline for dashboard | `logs` |

### 4.3 Page Components

Pages compose widgets with RTK Query hooks. Each page is a route target.

| Page | Route | Sketch source | Widgets composed |
|---|---|---|---|
| **DashboardPage** | `/annotations` | `DashboardView` | DashboardStatGrid, LatestRunBanner, RecentActivityList |
| **ReviewQueuePage** | `/annotations/review` | `ReviewQueueView` | FilterPillBar, CountSummaryBar, BatchActionBar, AnnotationTable |
| **AgentRunsPage** | `/annotations/runs` | `AgentRunsView` | MUI Table with ReviewProgressBar per row |
| **RunDetailPage** | `/annotations/runs/:runId` | `RunDetailView` | StatBox row, BatchActionBar, RunTimeline, GroupCard[], AnnotationTable |
| **SendersPage** | `/annotations/senders` | `SendersView` | MUI Table with TagChip, TargetLink, ReviewProgressBar |
| **SenderDetailPage** | `/annotations/senders/:email` | `SenderDetailView` | SenderProfileCard, SenderAnnotationList, AgentReasoningPanel, MessagePreviewTable |
| **GroupsPage** | `/annotations/groups` | `GroupsView` | GroupCard[] |
| **QueryEditorPage** | `/query` | `SqlWorkbenchView` | QueryEditor (ported from go-minitrace) |

## 5. RTK Query API Slice

```typescript
// ui/src/api/smailnailAnnotations.ts
import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import type { Annotation, TargetGroup, AnnotationLog, ... } from "../types/annotations";

export const annotationsApi = createApi({
  reducerPath: "annotationsApi",
  baseQuery: fetchBaseQuery({ baseUrl: "/api" }),
  tagTypes: ["Annotations", "Groups", "Logs", "Runs", "Senders", "Queries"],
  endpoints: (builder) => ({
    // ── annotations ──────────────────
    listAnnotations: builder.query<Annotation[], AnnotationFilter>({
      query: (filter) => ({ url: "annotations", params: filter }),
      providesTags: ["Annotations"],
    }),
    getAnnotation: builder.query<Annotation, string>({
      query: (id) => `annotations/${id}`,
    }),
    reviewAnnotation: builder.mutation<Annotation, { id: string; reviewState: string }>({
      query: ({ id, reviewState }) => ({
        url: `annotations/${id}/review`,
        method: "PATCH",
        body: { reviewState },
      }),
      invalidatesTags: ["Annotations", "Runs"],
    }),
    batchReview: builder.mutation<void, { ids: string[]; reviewState: string }>({
      query: (body) => ({ url: "annotations/batch-review", method: "POST", body }),
      invalidatesTags: ["Annotations", "Runs"],
    }),

    // ── groups ───────────────────────
    listGroups: builder.query<TargetGroup[], GroupFilter>({
      query: (filter) => ({ url: "annotation-groups", params: filter }),
      providesTags: ["Groups"],
    }),
    getGroup: builder.query<GroupDetail, string>({
      query: (id) => `annotation-groups/${id}`,
    }),

    // ── logs ─────────────────────────
    listLogs: builder.query<AnnotationLog[], LogFilter>({
      query: (filter) => ({ url: "annotation-logs", params: filter }),
      providesTags: ["Logs"],
    }),
    getLog: builder.query<LogDetail, string>({
      query: (id) => `annotation-logs/${id}`,
    }),

    // ── runs (aggregated) ────────────
    listRuns: builder.query<AgentRunSummary[], void>({
      query: () => "annotation-runs",
      providesTags: ["Runs"],
    }),
    getRun: builder.query<AgentRunDetail, string>({
      query: (id) => `annotation-runs/${id}`,
    }),

    // ── senders ──────────────────────
    listSenders: builder.query<SenderRow[], SenderFilter>({
      query: (filter) => ({ url: "mirror/senders", params: filter }),
      providesTags: ["Senders"],
    }),
    getSender: builder.query<SenderDetail, string>({
      query: (email) => `mirror/senders/${encodeURIComponent(email)}`,
    }),

    // ── query editor ─────────────────
    executeQuery: builder.mutation<QueryResult, { sql: string }>({
      query: (body) => ({ url: "query/execute", method: "POST", body }),
    }),
    getPresets: builder.query<SavedQuery[], void>({
      query: () => "query/presets",
    }),
    getSavedQueries: builder.query<SavedQuery[], void>({
      query: () => "query/saved",
      providesTags: ["Queries"],
    }),
    saveQuery: builder.mutation<SavedQuery, SaveQueryInput>({
      query: (body) => ({ url: "query/saved", method: "POST", body }),
      invalidatesTags: ["Queries"],
    }),
  }),
});
```

**Cache invalidation strategy:**
- `reviewAnnotation` and `batchReview` invalidate both `Annotations` and `Runs` tags (run progress bar updates)
- `saveQuery` invalidates `Queries` tag
- Senders list doesn't need invalidation (read-only view)

## 6. Redux UI Slice

```typescript
// ui/src/store/annotationUiSlice.ts
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface AnnotationUiState {
  reviewQueue: {
    selected: string[];       // annotation IDs
    filterTag: string | null;
    filterType: string | null;
    filterSource: string | null;
    filterRunId: string | null;
    expandedId: string | null;
  };
  queryEditor: {
    sql: string;
    activeSourcePath: string | null;
  };
}
```

## 7. Storybook Strategy

Every widget gets stories covering:

| State | Description |
|---|---|
| **Default** | Normal data, no selection |
| **Empty** | Zero items, empty-state message |
| **Loading** | Skeleton placeholders |
| **Error** | Error banner |
| **WithSelection** | Multiple items selected (for BatchActionBar, AnnotationTable) |
| **Expanded** | Detail panel open (for AnnotationTable) |
| **AllReviewed** | Queue cleared, success state |

**Story file convention:** Same as go-minitrace — `stories/WidgetName.stories.tsx` in each widget folder, using `withTheme` decorator.

**Mock data:** Create `ui/src/mocks/annotations.ts` with the mock objects from the JSX sketch (they're well-structured and realistic).

## 8. Shared Package Extraction Path

Widgets that are identical between go-minitrace and smailnail should eventually live in a shared package. The plan:

### Phase 1: Copy (now)
Copy these widgets from go-minitrace into `ui/src/components/shared/`:
- `SqlEditor` — CodeMirror 6 wrapper
- `QuerySidebar` — folder-grouped query list
- `ResultsTable` — sortable, exportable results with cell expansion
- `QueryEditor` — three-panel editor layout

### Phase 2: Align (during implementation)
As we build smailnail widgets, keep the following conventions identical to go-minitrace:
- `data-widget` / `data-part` attribute naming
- MUI `sx` prop patterns (no CSS modules, no styled-components)
- Props interfaces use the same naming conventions
- Theme structure compatible (both extend the same base theme)

### Phase 3: Extract (follow-up ticket)
Create `@go-go-golems/ui-shared` package containing:
- Theme definition (the shared dark MUI theme)
- Query editor widgets (SqlEditor, QuerySidebar, ResultsTable, QueryEditor)
- Shared atoms (TagChip, SourceBadge, ReviewStateBadge, StatBox, etc.)
- Storybook config and decorator

Both go-minitrace and smailnail depend on this package. Each app provides its own theme overrides and API layer.

## 9. Directory Structure

```
ui/src/
├── api/
│   ├── smailnail.ts           # existing account/rules API
│   └── annotations.ts          # NEW: RTK Query annotation + query API
├── store/
│   ├── index.ts
│   ├── store.ts                 # add annotationsApi reducer + middleware
│   └── annotationUiSlice.ts     # NEW: UI state for review queue + query editor
├── theme/
│   ├── theme.ts                 # extended MUI theme (match go-minitrace + tag colors)
│   └── tagColors.ts             # tag → color mapping
├── types/
│   └── annotations.ts           # TypeScript types matching backend
├── mocks/
│   └── annotations.ts           # mock data from JSX sketch
├── components/
│   ├── shared/                  # ← shared-package candidates
│   │   ├── TagChip.tsx
│   │   ├── TagChip.stories.tsx
│   │   ├── ReviewStateBadge.tsx
│   │   ├── ReviewStateBadge.stories.tsx
│   │   ├── SourceBadge.tsx
│   │   ├── SourceBadge.stories.tsx
│   │   ├── TargetLink.tsx
│   │   ├── TargetLink.stories.tsx
│   │   ├── StatBox.tsx
│   │   ├── StatBox.stories.tsx
│   │   ├── ReviewProgressBar.tsx
│   │   ├── ReviewProgressBar.stories.tsx
│   │   ├── BatchActionBar.tsx
│   │   ├── BatchActionBar.stories.tsx
│   │   ├── CountSummaryBar.tsx
│   │   ├── FilterPillBar.tsx
│   │   ├── MarkdownRenderer.tsx
│   │   └── index.ts
│   ├── QueryEditor/             # ← ported from go-minitrace
│   │   ├── QueryEditor.tsx
│   │   ├── QuerySidebar.tsx
│   │   ├── SqlEditor.tsx
│   │   ├── ResultsTable.tsx
│   │   ├── index.ts
│   │   └── stories/
│   │       ├── QueryEditor.stories.tsx
│   │       ├── ResultsTable.stories.tsx
│   │       └── QuerySidebar.stories.tsx
│   ├── AnnotationTable/
│   │   ├── AnnotationTable.tsx
│   │   ├── AnnotationRow.tsx
│   │   ├── AnnotationDetail.tsx
│   │   ├── index.ts
│   │   └── stories/
│   │       └── AnnotationTable.stories.tsx
│   ├── RunTimeline/
│   │   ├── RunTimeline.tsx
│   │   ├── index.ts
│   │   └── stories/
│   │       └── RunTimeline.stories.tsx
│   ├── GroupCard/
│   │   ├── GroupCard.tsx
│   │   ├── index.ts
│   │   └── stories/
│   │       └── GroupCard.stories.tsx
│   ├── SenderProfile/
│   │   ├── SenderProfileCard.tsx
│   │   ├── SenderAnnotationList.tsx
│   │   ├── AgentReasoningPanel.tsx
│   │   ├── MessagePreviewTable.tsx
│   │   ├── index.ts
│   │   └── stories/
│   │       └── SenderProfile.stories.tsx
│   └── Dashboard/
│       ├── DashboardStatGrid.tsx
│       ├── LatestRunBanner.tsx
│       ├── RecentActivityList.tsx
│       ├── index.ts
│       └── stories/
│           └── Dashboard.stories.tsx
├── pages/
│   ├── DashboardPage.tsx
│   ├── ReviewQueuePage.tsx
│   ├── AgentRunsPage.tsx
│   ├── RunDetailPage.tsx
│   ├── SendersPage.tsx
│   ├── SenderDetailPage.tsx
│   ├── GroupsPage.tsx
│   └── QueryEditorPage.tsx
└── App.tsx                      # extended with annotation + query routes
```

## 10. Sketch-to-MUI Mapping Reference

For developers translating the JSX sketch, here's how the sketch's styling patterns map to MUI:

| Sketch pattern | MUI equivalent |
|---|---|
| `style={S.card}` | `<Paper data-part="card">` |
| `style={S.cardHead}` | `<Box sx={{ display: "flex", p: 1.5, borderBottom: 1, borderColor: "divider" }}>` |
| `style={S.cardTitle}` | `<Typography variant="overline">` |
| `style={S.table}` | `<Table size="small" stickyHeader>` |
| `style={S.th}` | `<TableCell>` (MUI theme override handles styling) |
| `style={S.td}` | `<TableCell>` |
| `style={S.tag(type)}` | `<TagChip tag={type} />` (custom widget) |
| `style={S.reviewBadge(state)}` | `<ReviewStateBadge state={state} />` |
| `style={S.btn(variant)}` | `<Button variant="outlined" size="small" color={...}>` |
| `style={S.code}` | `<Chip size="small" sx={{ fontFamily: "monospace" }} />` |
| `style={S.mono}` | `sx={{ fontFamily: "monospace" }}` |
| `style={S.link}` | `<Link component="button" sx={{ fontFamily: "monospace" }}>` |
| `style={S.pill}` | `<Chip variant="outlined" size="small">` |
| `style={S.navItem(active)}` | `<ListItemButton selected={active}>` |
| `style={S.navBadge}` | `<Badge badgeContent={count}>` or `<Chip size="small" color="primary">` |
| `style={S.sidebar}` | `<Box sx={{ width: 220, borderRight: 1, borderColor: "divider" }}>` |
| `style={S.emptyState}` | `<Box sx={{ textAlign: "center", p: 5, color: "text.secondary" }}>` |

## 11. Implementation Order

Ordered by dependency (shared widgets first, then consumers):

### Sprint 1: Foundation (3 days)
1. Set up theme (`theme.ts`, `tagColors.ts`) — extend go-minitrace theme
2. Build all shared widgets with Storybook stories
3. Set up RTK Query API slice with mock data
4. Set up Redux store with annotationUiSlice
5. Set up routing in App.tsx

### Sprint 2: Review Queue (3 days)
6. Build AnnotationTable + AnnotationRow + AnnotationDetail
7. Build ReviewQueuePage (compose filters + batch bar + table)
8. Wire batch review mutation with cache invalidation
9. Storybook stories for all states

### Sprint 3: Agent Runs (2 days)
10. Build AgentRunsPage with MUI Table + ReviewProgressBar
11. Build RunDetailPage: stats, RunTimeline, GroupCard[], AnnotationTable
12. Storybook stories

### Sprint 4: Senders & Groups (2 days)
13. Build SendersPage with MUI Table
14. Build SenderDetailPage: SenderProfileCard, SenderAnnotationList, AgentReasoningPanel, MessagePreviewTable
15. Build GroupsPage with GroupCard
16. Storybook stories

### Sprint 5: Dashboard (1 day)
17. Build DashboardPage: DashboardStatGrid, LatestRunBanner, RecentActivityList

### Sprint 6: Query Editor (2 days)
18. Port QueryEditor, QuerySidebar, SqlEditor, ResultsTable from go-minitrace
19. Build QueryEditorPage with smailnail API hooks
20. Storybook stories

### Sprint 7: Polish (1 day)
21. Keyboard shortcuts (j/k navigation, x select, a/d review, Ctrl+Enter run)
22. Breadcrumb navigation
23. Sidebar pending-count badge (live from RTK Query cache)
24. Empty states and loading skeletons
