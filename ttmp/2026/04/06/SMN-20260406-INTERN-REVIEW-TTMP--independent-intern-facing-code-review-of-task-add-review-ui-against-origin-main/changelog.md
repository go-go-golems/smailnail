# Changelog

## 2026-04-06

- Created independent ticket `SMN-20260406-INTERN-REVIEW-TTMP` under `smailnail/ttmp`
- Investigated the `task/add-review-ui` branch against `origin/main`
- Mapped the SQLite review server, annotation repository, and React review frontend
- Validated the reviewed slice with:
  - `go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1`
  - `cd ui && pnpm run check`
- Wrote the main intern-facing report in `design-doc/01-intern-guide-and-independent-code-review-of-the-review-ui-branch.md`
- Wrote the chronological diary in `reference/01-diary.md`
- Documented the main review findings around queue semantics, feedback scope, API contract drift, missing audit metadata, placeholder guideline-run UX, transitional routing, missing focused tests, and tooling hygiene
- Revised the report after the later meta-review of the intern review ticket and explicitly labeled imported findings from that second pass
- Added validated follow-up findings around dead `ReviewCommentInline`, dead Redux review state, Storybook guideline endpoint drift, non-persistent MSW create handlers, fake `linkedRunCount`, and duplicated feedback insert logic
- Re-uploaded the updated bundle to reMarkable
- Added a phased follow-up implementation plan inside this ticket for findings 5 and 9, with findings 7 and 8 explicitly deferred for now
- Implemented guideline-linked runs fully by adding repository support, backend endpoint `GET /api/review-guidelines/{id}/runs`, frontend query wiring, detail-page loading, mock support, and focused backend coverage
- Created focused linked-runs implementation commit `AnnotationUI: add guideline linked runs endpoint`
- Cleaned dead review-queue Redux state that was no longer wired to `ReviewQueuePage`
- Removed fake `linkedRunCount={0}` wiring from the guideline list page until the backend exposes real count data

## 2026-04-06

Completed independent intern-facing code review of task/add-review-ui against origin/main; documented architecture, findings, validation, and delivery.

### Related Files

- /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/pkg/annotate/repository_feedback.go — Main backend evidence for feedback/guideline review flows
- /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ttmp/2026/04/06/SMN-20260406-INTERN-REVIEW-TTMP--independent-intern-facing-code-review-of-task-add-review-ui-against-origin-main/design-doc/01-intern-guide-and-independent-code-review-of-the-review-ui-branch.md — Primary deliverable
- /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ui/src/pages/ReviewQueuePage.tsx — Main frontend evidence for review-queue semantics

