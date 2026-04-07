# Changelog

## 2026-04-07

- Created ticket `SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP` under `smailnail/ttmp`
- Wrote the primary design doc `design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`
- Wrote the investigation diary `reference/01-investigation-diary.md`
- Documented the current runtime/query/page/story architecture relevant to cross-view consistency
- Documented the main current-state gaps around broad cache tags, uneven page composition, missing target-scoped feedback lookup, sender-visible guideline ambiguity, and Storybook/MSW state drift
- Added a phased implementation plan for backend query support, frontend view composition, invalidation policy, and Storybook coverage
- Added a dedicated artifact/query/invalidation matrix reference document so later implementation slices can be checked against one page-by-page source of truth
- Expanded `tasks.md` into detailed execution steps covering documentation, backend, frontend, cache, Storybook, and final handoff slices
- Related the key runtime/query/page/story files to the ticket docs with `docmgr doc relate`
- Ran `docmgr doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30` and got a clean report
- Uploaded the ticket bundle to reMarkable as `SMN-20260407 Annotation UI Consistency Pass.pdf` under `/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP`
