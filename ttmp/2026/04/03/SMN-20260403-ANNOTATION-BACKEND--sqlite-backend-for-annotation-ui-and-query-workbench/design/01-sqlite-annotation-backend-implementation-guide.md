---
Title: SQLite annotation backend implementation guide
Ticket: SMN-20260403-ANNOTATION-BACKEND
Status: active
Topics:
    - backend
    - annotations
    - sqlite
    - api
    - cli
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/smailnail/main.go
      Note: Root CLI wiring where the sqlite command group will be registered
    - Path: cmd/smailnaild/commands/serve.go
      Note: Existing hosted server used only as a pattern reference
    - Path: pkg/annotate/repository.go
      Note: Repository layer that already contains most annotation CRUD primitives
    - Path: ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md
      Note: Source handoff spec being corrected for the sqlite server architecture
    - Path: ui/src/api/annotations.ts
      Note: Frontend RTK Query contract that the sqlite backend must satisfy
    - Path: ui/src/mocks/handlers.ts
      Note: Reference filtering and composition logic for the API handlers
    - Path: ui/src/types/annotations.ts
      Note: Canonical TypeScript response shapes for the backend
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-03T09:43:51.032449805-04:00
WhatFor: Correct the annotation UI backend handoff and drive implementation of a sqlite-backed web server under the smailnail CLI
WhenToUse: Use when implementing or reviewing the sqlite-backed annotation/query backend and the smailnail sqlite serve command
---


# SQLite annotation backend implementation guide

## Goal

Implement the annotation UI backend as a separate sqlite-backed server under `smailnail sqlite serve`, not inside `smailnaild`.

## Spec Audit

The source design doc in `SMN-20260403-ANNOTATION-UI/design/08-backend-api-specification-for-annotation-ui.md` is strong on endpoint shapes, but it is wrong about the host process. It assumes the work should be added to `pkg/smailnaild/http.go` and even shows `go run ./cmd/smailnail serve --dev`, while the current repository only has `cmd/smailnaild serve` and `smailnaild` is scoped around hosted user/account/rule management.

For this ticket, the frontend contract remains valid, but the hosting architecture changes:

- Keep the UI response shapes exactly aligned with `ui/src/types/annotations.ts` and `ui/src/api/annotations.ts`
- Reuse the existing mirror sqlite database and annotation/enrichment tables
- Reuse the built SPA assets from `pkg/smailnaild/web/embed/public`
- Do not add this feature to `smailnaild`
- Add a new `smailnail sqlite serve` command and a dedicated sqlite-web backend package

## Chosen Architecture

### Runtime boundary

- Binary: `smailnail`
- Command path: `smailnail sqlite serve`
- Database: mirror sqlite database such as `smailnail-mirror.sqlite`
- HTTP surface: annotation API, query API, sender browser API, health/info routes, SPA/static assets
- Not included: hosted auth, hosted account CRUD, rules CRUD, MCP, credential storage

### Package split

- `pkg/annotate`: keep repository-centric CRUD plus add run aggregation and batch review support
- `pkg/annotationsui` or equivalent sqlite-web package: HTTP handler, sender/query support, request/response helpers
- `cmd/smailnail/commands/sqlite`: cobra/glazed command group for sqlite-specific verbs, beginning with `serve`
- `pkg/smailnaild/web`: continue to provide the built SPA files

### UI serving strategy

The built frontend still contains the legacy `/` shell for account/rule pages, which is not meaningful in sqlite-only mode. The sqlite server should therefore:

- Redirect `/` to `/annotations`
- Serve SPA fallback for `/annotations`, `/annotations/*`, `/query`, and `/query/*`
- Serve `/assets/*` and other static files from the existing embedded/disk UI build
- Avoid exposing fake or partial hosted-account endpoints just to satisfy the legacy shell

## Contract Decisions

### JSON envelope

Use bare JSON for the new sqlite annotation endpoints. The frontend RTK Query layer already expects bare arrays/objects, and changing the frontend just to accommodate `smailnaild` envelope conventions would be unnecessary churn.

### Error shape

Use small direct JSON payloads:

- Not found: `{"error":"not-found","message":"..."}`
- Validation/query errors: `{"message":"..."}`

This matches the spirit of the UI mocks closely enough without importing the hosted `smailnaild` error envelope.

### Query execution safety

The query workbench must be read-only. Prefer SQLite read-only enforcement where practical and also reject obviously mutating SQL statements up front. The implementation should allow CTEs that begin with `WITH`, but should reject statements whose effective first keyword is one of:

- `INSERT`
- `UPDATE`
- `DELETE`
- `DROP`
- `ALTER`
- `CREATE`
- `REPLACE`
- `VACUUM`
- `ATTACH`
- `DETACH`
- `PRAGMA` when used as a mutating write

## Endpoint Scope

Implement the 16 endpoints already consumed by the frontend:

1. `GET /api/annotations`
2. `GET /api/annotations/{id}`
3. `PATCH /api/annotations/{id}/review`
4. `POST /api/annotations/batch-review`
5. `GET /api/annotation-groups`
6. `GET /api/annotation-groups/{id}`
7. `GET /api/annotation-logs`
8. `GET /api/annotation-logs/{id}`
9. `GET /api/annotation-runs`
10. `GET /api/annotation-runs/{id}`
11. `GET /api/mirror/senders`
12. `GET /api/mirror/senders/{email}`
13. `POST /api/query/execute`
14. `GET /api/query/presets`
15. `GET /api/query/saved`
16. `POST /api/query/saved`

## Implementation Order

### Slice 1: Repository and response types

- Add `AgentRunID` to `annotate.ListAnnotationsFilter`
- Add run/group detail types
- Add `BatchUpdateReviewState`
- Add run aggregation/detail repository methods
- Add repository tests for the new methods

### Slice 2: SQLite HTTP server foundation

- Create the sqlite-web package and server options
- Add health/info routes
- Add JSON/error helpers
- Add root redirect and SPA/static serving
- Add `smailnail sqlite serve`

### Slice 3: Annotation/group/log/run handlers

- Implement the first 10 endpoints
- Keep response JSON bare
- Match the frontend mocks for filtering and detail composition
- Add handler tests

### Slice 4: Senders and query workbench

- Add sender list/detail queries against `senders`, `messages`, and `annotations`
- Add preset/saved query loading and saving
- Add read-only SQL execution
- Add tests around query safety and filesystem query persistence

### Slice 5: Validation and ticket cleanup

- Run focused package tests, then `go test ./...`
- Smoke the server via `tmux`
- Update tasks, changelog, related files, and diary entries after each slice
- Commit in focused checkpoints

## Review Notes

The highest-risk areas are:

- Reusing the existing built SPA without accidentally serving the legacy hosted shell at `/`
- Dynamic SQL row scanning for the query editor
- Correct aggregation for run summaries
- Sender detail joins that pull logs indirectly through matching `agent_run_id`

## Validation Commands

```bash
go test ./pkg/annotate ./pkg/annotationsui/... ./cmd/smailnail/commands/sqlite/...
go test ./...
go run ./cmd/smailnail sqlite serve --sqlite-path ./smailnail-mirror.sqlite --listen-port 8080
curl -s localhost:8080/api/annotations | jq '.[0]'
curl -s localhost:8080/api/annotation-runs | jq '.[0]'
curl -s -X POST localhost:8080/api/query/execute -H 'content-type: application/json' \
  -d '{"sql":"SELECT 1 AS value"}' | jq '.'
```
