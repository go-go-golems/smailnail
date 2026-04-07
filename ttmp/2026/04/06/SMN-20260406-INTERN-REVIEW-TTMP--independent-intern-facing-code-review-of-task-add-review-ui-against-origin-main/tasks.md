# Tasks

## Done

- [x] Create an independent ticket under `smailnail/ttmp`
- [x] Map the branch architecture and changed subsystems
- [x] Compare `task/add-review-ui` against `origin/main`
- [x] Identify confusing, incomplete, unused, and contract-drift areas
- [x] Write an intern-facing design-doc report with prose, diagrams, API references, pseudocode, and file references
- [x] Write a chronological diary with commands, failures, and validation notes
- [x] Run focused validation commands for backend and frontend review surfaces
- [x] Revise the report after the later meta-review of the intern ticket and label imported findings
- [x] Re-upload the updated review bundle to reMarkable
- [x] Follow-up finding 3: align Go/TypeScript feedback and guideline contracts via shared protobuf codegen (`SMN-20260406-CONTRACT-CODEGEN`)

## Completed Follow-up Work

### Phase 1 — Ship finding 5 fully: guideline detail linked runs

- [x] Add repository support to list run summaries linked to a guideline ID
- [x] Add backend endpoint `GET /api/review-guidelines/{id}/runs`
- [x] Add frontend RTK Query hook for guideline-linked runs
- [x] Load linked runs on `GuidelineDetailPage` instead of hard-coding `runs={[]}`
- [x] Update mocks and any affected stories so the linked-runs section is exercised realistically
- [x] Add focused backend coverage for the new guideline-runs endpoint
- [x] Validate phase 1 (`go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1`, `cd ui && pnpm run check`)
- [x] Commit phase 1 as a focused linked-runs implementation change

### Phase 2 — Address finding 9 by wiring or cleaning dead review UI state

- [x] Audit `annotationUiSlice` review-queue state against the live `ReviewQueuePage`
- [x] Remove dead review-queue slice fields/actions that are no longer wired (`filterType`, `filterSource`, `filterRunId`, `commentDrawerOpen`, `filterMailbox`) unless a live caller still needs them
- [x] Remove fake guideline list UI state that advertises unavailable data (`linkedRunCount={0}`) until the backend exposes a real count
- [x] Re-run frontend validation for the cleanup slice (`cd ui && pnpm run check`)
- [x] Commit phase 2 as a focused review-UI cleanup change

## Current Follow-up Plan

### Phase 3 — Fix finding 1: make the Review Queue actually behave like a queue

- [x] Update `ReviewQueuePage` to query only pending-review annotations (`reviewState=to_review`)
- [x] Make review-queue tag counts derive from the same pending-review population instead of all annotations
- [x] Update review-queue stories/mocks affected by the queue-only semantics
- [x] Validate phase 3 (`cd ui && pnpm run check`)
- [x] Commit phase 3 as a focused Review Queue semantics change

### Phase 4 — Fix finding 2: add `scopeKind` filtering for run feedback

- [x] Extend backend feedback list filtering with `scopeKind`
- [x] Extend frontend feedback filter types and RTK Query usage with `scopeKind`
- [x] Update `RunDetailPage` to request only run-scoped feedback
- [x] Update MSW mocks / stories to respect `scopeKind` filtering
- [x] Add focused backend coverage for scope-filtered feedback listing
- [x] Validate phase 4 (`go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1`, `cd ui && pnpm run check`)
- [x] Commit phase 4 together with phase 5, because the pre-commit full-repo test path stashes unstaged changes and the new audit test depends on the same handler files

### Phase 5 — Fix finding 4: make audit metadata real

- [x] Add a request-scoped review actor helper for the annotation UI handlers
- [x] Populate `CreatedBy` in feedback creation, guideline creation, review-with-artifacts, and batch-review-with-artifacts handlers
- [x] Populate `LinkedBy` in explicit run-guideline link handlers
- [x] Add focused backend coverage proving created/linked audit fields are no longer empty
- [x] Validate phase 5 (`go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1`)
- [x] Commit phase 5 together with phase 4 as one focused feedback-integrity / audit-metadata change

### Phase 6 — Ticket hygiene and handoff for findings 1/2/4

- [x] Update the ticket diary with exact commands, validation, and any pitfalls from phases 3-5
- [x] Update the ticket changelog and index status for the implemented follow-up work
- [x] Relate changed code files to the ticket docs with `docmgr`
- [x] Run `docmgr doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30`
- [x] Commit the ticket-doc updates for findings 1/2/4

### Phase 7 — Fix finding 6: await guideline-link mutations and surface failures

- [x] Make `RunGuidelineSection` await link mutations before closing the picker
- [x] Preserve selection / keep the picker open when a link operation fails
- [x] Surface link errors in the run-guideline UI instead of failing silently
- [x] Make `GuidelineDetailPage` await create-and-link flows before navigating back to the run page
- [x] Surface create/link failures in the guideline detail flow, including the “guideline created but run link failed” case
- [x] Update any affected stories or mock handlers so the async link flow still reflects the real API shape
- [x] Validate phase 7 (`cd ui && pnpm run check`)
- [x] Commit phase 7 as a focused async guideline-link flow fix

### Phase 8 — Ticket hygiene and handoff for finding 6

- [x] Update the ticket diary/changelog/index/tasks for the finding-6 work
- [x] Relate newly changed files to the ticket docs with `docmgr`
- [x] Run `docmgr doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30`
- [ ] Commit the ticket-doc updates for finding 6

## Explicitly Deferred For Now

- [ ] Finding 7: settle the `/` vs `/annotations` transitional app architecture
- [ ] Finding 8: broader feedback/guideline performance/test cleanup beyond the targeted coverage needed for phase 4
- [ ] Package-manager / embed-asset policy cleanup outside the targeted review-queue state cleanup
