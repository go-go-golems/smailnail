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
    - Path: pkg/annotate/repository.go
      Note: Repository slice implemented in commit 4bda44f
    - Path: pkg/annotationui/server.go
      Note: Server slice implemented in commit 9a7345a
    - Path: pkg/annotationui/server_test.go
      Note: Handler contract coverage used during validation
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

**Commit (code):** 2d53809 — "docs: scaffold sqlite annotation backend ticket"

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

### Step 2: Extend the annotation repository for backend aggregation

The next slice stayed entirely in `pkg/annotate` so the HTTP layer could build on stable repository primitives. I added `AgentRunID` filtering to annotation listing, batch review updates, and aggregated run summary/detail queries, then covered those additions with focused repository tests.

This was the right seam because the backend API needs those capabilities whether the caller is HTTP, CLI, or future automation. Landing them separately kept the server implementation smaller and made the resulting API handlers mostly about request parsing and JSON composition instead of bespoke SQL.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build the backend incrementally and commit at logical checkpoints, starting with the repository foundation.

**Inferred user intent:** Reduce implementation risk by landing reusable backend primitives before the server layer.

**Commit (code):** 4bda44f — "annotate: add run summaries and batch review support"

### What I did

- Added `AgentRunID` to `annotate.ListAnnotationsFilter`
- Added `annotate.GroupDetail`, `annotate.AgentRunSummary`, and `annotate.AgentRunDetail`
- Added `Repository.BatchUpdateReviewState`
- Added `Repository.ListRuns` and `Repository.GetRunDetail`
- Added repository tests for run filtering, batch review, and run aggregation/detail
- Formatted the package and validated it with the required build tag

### Why

- The API contract depends on these repository capabilities
- Keeping these changes out of the HTTP package avoided mixing transport concerns with SQL behavior

### What worked

- The existing repository already had the right CRUD shape, so the missing features fit naturally
- Focused package tests made it easy to validate the aggregation behavior before the server existed

### What didn't work

- Running `go test ./pkg/annotate` without tags failed because the repo intentionally requires SQLite FTS5 support:
  `pkg/mirror/require_fts5_build_tag.go:5:9: undefined: requires_sqlite_fts5_build_tag`
- The fix was to run the repository tests as `go test -tags sqlite_fts5 ./pkg/annotate`

### What I learned

- The repo’s mirror/bootstrap path is consistently built around the `sqlite_fts5` tag, so runtime smoke commands need the same tag discipline as tests
- The run summary shape works cleanly as a repository-level aggregation rather than a handler-only composition

### What was tricky to build

- The main sharp edge was SQLite aggregate scanning for `MIN(created_at)` and `MAX(created_at)`. The existing entity types use `time.Time`, but the aggregate rows are safer as string fields because SQLite returns expression results differently from table-backed timestamp columns. That kept the JSON shape aligned with the frontend without fighting the driver’s scanning rules.

### What warrants a second pair of eyes

- The semantics of run aggregation when there are logs/groups without annotations for a run
- Whether repository-level review-state validation should be tightened later rather than living only at the handler boundary

### What should be done in the future

- Build the sqlite HTTP server on top of these repository methods

### Code review instructions

- Start in `pkg/annotate/types.go`
- Then review `pkg/annotate/repository.go` for the new SQL paths
- Validate behavior with `go test -tags sqlite_fts5 ./pkg/annotate`

### Technical details

- Test command: `go test -tags sqlite_fts5 ./pkg/annotate`
- Full validation command used during commit hook: `go test -tags sqlite_fts5 ./...`

### Step 3: Implement the sqlite annotation UI server and CLI verb

The main implementation slice created a dedicated `pkg/annotationui` package plus a new `smailnail sqlite serve` command. This server exposes the full bare-JSON annotation API, sender browser endpoints, read-only query workbench endpoints, embedded preset SQL, and SPA serving with a root redirect to `/annotations`.

The key design decision here was to reuse the already-built UI assets from `pkg/smailnaild/web` without reusing `smailnaild` itself. That preserved the frontend work while keeping the runtime boundary correct: sqlite browse mode lives in `smailnail`, and hosted credential management remains in `smailnaild`.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the backend implementation end to end, including the new CLI command, HTTP server, tests, smoke validation, and ticket updates.

**Inferred user intent:** Have a usable sqlite-backed backend that the annotation UI can talk to immediately, without polluting the hosted server.

**Commit (code):** 9a7345a — "sqlite: serve the annotation ui from mirror db"

### What I did

- Added `pkg/annotationui` with:
  - health/info routes
  - annotation/group/log/run handlers
  - sender list/detail handlers
  - query execute/presets/saved handlers
  - SPA/static serving with `/` redirecting to `/annotations`
  - query file loading/saving helpers
  - embedded preset SQL files
- Added `pkg/annotationui/server_test.go` to exercise the frontend contract through the real handler
- Added `cmd/smailnail/commands/sqlite/root.go` and `serve.go`
- Wired the new sqlite command group into `cmd/smailnail/main.go`
- Ran `go test -tags sqlite_fts5 ./...`
- Smoked `go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve --sqlite-path ./smailnail-mirror.sqlite --listen-port 18080` in `tmux`

### Why

- The frontend already expects these endpoints and response shapes
- A dedicated sqlite package cleanly separates mirror browsing from hosted auth/account workflows
- Handler tests catch route/shape regressions that compile-only checks miss

### What worked

- The full tagged test suite passed before and during the commit hook
- The handler tests exercised the actual JSON shapes, filtering, query persistence, and SPA behavior
- The runtime smoke validated the actual CLI path and server process, not just unit tests

### What didn't work

- The first `go run ./cmd/smailnail sqlite serve ...` tmux smoke exited immediately because the repo requires FTS5 build tags for sqlite-backed packages
- Rerunning with `go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve ...` fixed it
- The initial handler tests exposed three real bugs:
  - root redirect matching too broadly and redirecting `/annotations`
  - a bad batch-review test fixture selection
  - an over-constrained sender-list assertion
- Those were fixed directly rather than weakening coverage

### What I learned

- Reusing `pkg/smailnaild/web.PublicFS` is practical as long as the sqlite server owns the routing behavior and redirects `/` away from the legacy hosted shell
- The query workbench can stay file-based and lightweight without importing the larger go-minitrace saved-query model

### What was tricky to build

- The trickiest part was serving the existing SPA without accidentally exposing the wrong application entrypoint. The built frontend still knows about the legacy hosted shell at `/`, so the sqlite server has to treat `/` specially and fall back to `index.html` only for the annotation/query routes. The failing test around `/annotations` confirmed that a naive `GET /` redirect registration in `ServeMux` was too broad and had to be collapsed into the catch-all static handler with an exact-path check.

### What warrants a second pair of eyes

- Query read-only enforcement is intentionally conservative and keyword-based; it is safe for the current UI but worth revisiting if the workbench becomes more sophisticated
- The sender aggregation query uses `GROUP_CONCAT(DISTINCT ...)`; if future tags contain commas, the serialization strategy should be revisited

### What should be done in the future

- Add update/delete support for saved queries only if the frontend starts consuming it
- Consider a sqlite-specific frontend shell later so `/` no longer depends on redirect behavior

### Code review instructions

- Start in `pkg/annotationui/server.go` for the route boundary and SPA handling
- Then review `pkg/annotationui/handlers_annotations.go`, `pkg/annotationui/handlers_senders.go`, and `pkg/annotationui/handlers_query.go`
- Review the CLI wiring in `cmd/smailnail/commands/sqlite/serve.go` and `cmd/smailnail/main.go`
- Validate with:
  - `go test -tags sqlite_fts5 ./pkg/annotationui`
  - `go test -tags sqlite_fts5 ./...`
  - `go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve --sqlite-path ./smailnail-mirror.sqlite --listen-port 18080`

### Technical details

- Runtime smoke responses observed:
  - `GET /healthz` → `{"status":"ok"}`
  - `GET /api/info` → service `smailnail-sqlite`, driver `sqlite3`, mode `mirror`
  - `GET /api/query/presets` returned the embedded annotation/mirror preset set
  - `POST /api/query/execute` with `SELECT COUNT(*) AS count FROM annotations` returned a one-row result

## Related

- `../design/01-sqlite-annotation-backend-implementation-guide.md`
