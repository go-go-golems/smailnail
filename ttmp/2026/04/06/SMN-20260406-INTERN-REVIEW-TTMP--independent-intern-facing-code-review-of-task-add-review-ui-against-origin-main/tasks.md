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

## Current Follow-up Plan

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

### Phase 3 — Ticket hygiene and handoff

- [x] Update the ticket diary with exact commands, validation, and any pitfalls from phases 1-2
- [x] Update the ticket changelog and index status for the implemented follow-up work
- [x] Relate changed code files to the ticket docs with `docmgr`
- [x] Run `docmgr doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30`

## Explicitly Deferred For Now

- [ ] Finding 7: settle the `/` vs `/annotations` transitional app architecture
- [ ] Finding 8: broader feedback/guideline performance/test cleanup beyond the targeted coverage needed for phase 1
- [ ] Finding 6: await guideline-link mutations and surface failures more explicitly in the UI
- [ ] Package-manager / embed-asset policy cleanup outside the targeted review-queue state cleanup
