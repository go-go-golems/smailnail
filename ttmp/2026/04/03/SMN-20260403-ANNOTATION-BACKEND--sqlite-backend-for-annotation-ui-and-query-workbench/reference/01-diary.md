---
Title: Diary
Ticket: SMN-20260403-ANNOTATION-BACKEND
Status: active
Topics:
    - backend
    - annotations
    - sqlite
    - api
    - cli
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ttmp/2026/04/03/SMN-20260403-ANNOTATION-BACKEND--sqlite-backend-for-annotation-ui-and-query-workbench/design/01-sqlite-annotation-backend-implementation-guide.md
      Note: Primary design document for the implementation steps recorded here
    - Path: ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md
      Note: Original spec being audited and implemented against the codebase
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-03T09:43:51.079308477-04:00
WhatFor: Track implementation steps, failures, commit checkpoints, and review guidance for the sqlite annotation backend work
WhenToUse: Use while implementing the sqlite backend ticket and when reviewing the resulting commits
---


# Diary

## Goal

Track the implementation of the sqlite-backed annotation UI backend, including the design correction from `smailnaild` to `smailnail sqlite serve`.

## Context

The frontend annotation UI already exists and expects 16 bare-JSON endpoints. The handoff spec captures those endpoint contracts, but points implementation at the wrong binary and package. This diary records the corrected architecture, the implementation slices, and the validation performed after each slice.

## Quick Reference

### Step 1: Create backend ticket and correct the architecture

The first move was to create a new backend-focused docmgr ticket so the implementation could be tracked separately from the frontend ticket. The main design correction was architectural, not API-level: the backend spec describes the right endpoint contract but puts it in `smailnaild`, which is the hosted credentials/rules server. The codebase needs a sqlite-backed server attached to the `smailnail` CLI instead.

This step created the ticket skeleton, audited the existing backend spec against the current repository layout, and established the implementation plan around a new `smailnail sqlite serve` command plus a dedicated sqlite-web handler package. That avoids overloading `smailnaild` with mirror-database browsing concerns.

### Prompt Context

**User prompt (verbatim):** "Work on the docmgr ticket and read ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md . Add it as a serve verb on the smailnail sqlite binary, this is separate from the main smailnaild which is about managing user credentials.

Create a new docmgr ticket for imlementing the backend, double check the provided document and create your own analysis / implamantation guide document in the new ticket, add a detailed list of tasks to the new ticket and then implement them one by one, committing at appropriate intervals, and keeping a detailed diary."

**Assistant interpretation:** Create a backend implementation ticket, audit the provided API design against the actual codebase, write a corrected implementation guide, then implement the sqlite-backed annotation/query server incrementally with commits and diary updates.

**Inferred user intent:** Land the annotation UI backend in the right runtime boundary and leave behind clean ticketing, task tracking, and reviewable implementation history.

**Commit (code):** N/A

### What I did

- Read the backend API spec and the frontend RTK Query/types/mock handlers
- Audited the current command tree and confirmed there is no existing `smailnail sqlite serve`
- Confirmed the only current HTTP server is `smailnaild`, which is scoped to hosted auth/accounts/rules
- Created ticket `SMN-20260403-ANNOTATION-BACKEND`
- Created the design doc and diary for the new ticket
- Wrote the corrected architecture and implementation order into the backend guide

### Why

- The endpoint contract should be preserved, but the hosting boundary in the handoff spec is wrong for this repository
- A separate ticket keeps backend implementation history and tasks isolated from the already-completed frontend work

### What worked

- The frontend contract was easy to confirm from `ui/src/api/annotations.ts`, `ui/src/types/annotations.ts`, and `ui/src/mocks/handlers.ts`
- The mismatch between the handoff spec and the actual binaries was concrete and easy to prove from the repo layout

### What didn't work

- N/A

### What I learned

- The existing built SPA assets can likely be reused from `pkg/smailnaild/web/embed/public`, even though the sqlite server will live elsewhere
- The frontend expects bare JSON and would be awkward to adapt back to the hosted `writeDataJSON` envelope

### What was tricky to build

- The tricky part at this stage was deciding whether to reuse `pkg/smailnaild/http.go` or build a new server package. Reusing it would have dragged in hosted concepts that do not belong in sqlite browse mode and would also leave the legacy `/` shell pointing at irrelevant account/rule flows. The solution is to keep the UI assets but give them a sqlite-specific server boundary and routing behavior.

### What warrants a second pair of eyes

- The exact package name and placement for the sqlite HTTP server once implementation starts
- Whether the current built SPA should be served as-is with a root redirect, or whether the frontend should later gain a sqlite-only shell

### What should be done in the future

- Implement the repository extensions and the sqlite server slices
- Keep the diary updated after each commit checkpoint

### Code review instructions

- Start with the corrected implementation guide in this ticket
- Compare it against the source design doc from `SMN-20260403-ANNOTATION-UI`
- Validate that the planned command boundary is `smailnail sqlite serve`, not `smailnaild`

### Technical details

- Source design doc: `ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md`
- Frontend contract: `ui/src/api/annotations.ts`, `ui/src/types/annotations.ts`, `ui/src/mocks/handlers.ts`
- Existing hosted server: `cmd/smailnaild/commands/serve.go`, `pkg/smailnaild/http.go`

## Usage Examples

Use this diary as the review trail while the backend is implemented. Each later step should add the code commit hash, exact test commands, and any failures encountered.

## Related

- `../design/01-sqlite-annotation-backend-implementation-guide.md`
