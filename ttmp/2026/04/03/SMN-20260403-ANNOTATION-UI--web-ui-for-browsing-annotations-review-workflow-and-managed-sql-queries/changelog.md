# Changelog

## 2026-04-03

- Initial workspace created


## 2026-04-03

Completed UX design and screen specifications. Two documents: (1) UX Functionality Design covering all features, data model, API endpoints, navigation flow, and incremental delivery plan. (2) Screen Specifications with 15 ASCII wireframes and YAML widget hierarchy DSL for every screen, plus shared widget catalog and routing/file-structure plan. Modeled query editor on go-minitrace's existing QueryEditor + QuerySidebar + ResultsTable components.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx — Reference implementation for query editor design
- /home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go — Domain types analyzed for UI data shapes


## 2026-04-03

Created v2 documents: (1) UX Functionality Design v2 — removed agent API, switched to file-based SQL queries following go-minitrace pattern. (2) Screen Specifications v2 — updated wireframes and YAML hierarchy, removed fragments table and saved_queries table. (3) Agent Annotation CLI Design — new doc specifying run-centric CLI with mandatory logging, bulk tag add, atomic group-with, and composite triage commands. (4) UX Design Brief for External Designers — problem/data/constraints without prescribing solutions.


## 2026-04-03

Imported JSX design sketch (smailnail-review-ui.jsx) and created comprehensive frontend implementation plan (07). Analyzed sketch's 8 views and mapped all inline styles to MUI equivalents. Decomposed into 12 shared widgets, 12 annotation-specific widgets, and 8 page components across 7 sprints. Created 45 detailed tasks covering theme setup, widget building with Storybook stories, RTK Query API slice, Redux UI state, go-minitrace query editor port, and polish. Documented shared-package extraction path for @go-go-golems/ui-shared.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/sources/smailnail-review-ui.jsx — Imported JSX design sketch analyzed for implementation


## 2026-04-03

Step 1: Foundation — MUI theme, 10 shared widgets with Storybook stories, TypeScript types, mock data, Storybook 8.6 setup (commit 7515544)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/components/shared/index.ts — 10 shared widgets
- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/theme/theme.ts — MUI dark theme


## 2026-04-03

Step 2: RTK Query API slice (16 endpoints), Redux annotationUiSlice, react-router routes, AnnotationLayout + sidebar navigation (commit 9319d86)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx — Router + routes
- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/api/annotations.ts — RTK Query API


## 2026-04-03

Step 3: AnnotationTable widget (AnnotationRow + AnnotationDetail + AnnotationTable), ReviewQueuePage with filters + batch actions + RTK Query wiring (commit 7cbb343)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/components/AnnotationTable/AnnotationTable.tsx — Composing table widget
- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/pages/ReviewQueuePage.tsx — Review queue page


## 2026-04-03

Step 4: MSW setup + Sprint 3 — AgentRunsPage, RunTimeline, GroupCard, RunDetailPage, page-level MSW stories for ReviewQueue/AgentRuns/RunDetail (commit f026771)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/mocks/handlers.ts — MSW handlers
- /home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/pages/RunDetailPage.tsx — Run detail page

