---
Title: Repo-wide wire contract unification spec
Ticket: SMN-20260406-CONTRACT-CODEGEN
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - workflow
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/doc/annotationui-contract-codegen-playbook.md
      Note: Repo-wide playbook updated with hosted API envelope conventions
    - Path: pkg/smailnaild/contracts_hosted.go
      Note: Mapper helpers between hosted domain/service structs and generated protobuf wire types
    - Path: pkg/smailnaild/http.go
      Note: Hosted HTTP handlers migrated to generated protobuf wire contracts
    - Path: pkg/smailnaild/protojson.go
      Note: Hosted API protojson encode/decode helpers and typed error envelope writer
    - Path: proto/smailnail/app/v1/hosted.proto
      Note: Hosted web API contract schema for current user
    - Path: ui/src/api/client.ts
      Note: Frontend hosted API client updated to generated request/response contracts
    - Path: ui/src/api/types.ts
      Note: Frontend hosted API wrapper types derived from generated contracts
ExternalSources: []
Summary: Specify how all frontend/backend wire DTOs in smailnail should converge on protobuf-defined shared contracts, generated for Go and TypeScript, while keeping domain/storage structs handwritten.
LastUpdated: 2026-04-06T22:40:00Z
WhatFor: Provide the target architecture, conventions, rollout plan, and implementation scope for repo-wide DTO unification.
WhenToUse: Read before extending shared contract codegen beyond the current annotation UI and hosted web API slices.
---


# Repo-wide wire contract unification spec

## Executive summary

Smailnail should converge on a single pattern for every real frontend/backend boundary: define wire contracts in protobuf, generate Go and TypeScript code, decode/encode JSON with `protojson` at HTTP boundaries, and keep handwritten mapping code between generated wire types and internal domain/storage types.

This specification deliberately does **not** recommend generating every struct in the repository. The target is **wire-contract unification**, not wholesale replacement of domain models, repository inputs, SQL rows, or service-internal state.

## Architectural goal

The repository should have one consistent answer to the question: “Where does this API payload shape come from?”

The answer should be:

1. The shape is defined in `proto/...`.
2. Go types are generated under `pkg/gen/...`.
3. TypeScript interfaces are generated under `ui/src/gen/...`.
4. Backend handlers map internal structs to generated wire types and use `protojson`.
5. Frontend wrapper types narrow or adapt generated types for ergonomics, but do not become the source of truth.
6. Mocks, stories, and tests follow the generated contract as well.

## Scope

### In scope

- HTTP request bodies
- HTTP response bodies
- list/detail DTOs
- query/filter DTOs used by frontend and handlers
- consistent error response envelopes
- future websocket/SSE payloads if they become part of the UI contract

### Out of scope

- database row structs
- repository input/output types used only internally
- service-private helper structs
- direct replacement of `pkg/annotate`, `pkg/smailnaild/accounts`, or `pkg/smailnaild/rules` domain structs with generated protobuf structs

## Current state after the first implementation waves

Already converted:

- annotation UI review contracts
- annotation UI read/query contracts

Still to complete for full repo-wide unification at the time this spec was written:

- hosted web API contracts for:
  - `/api/info`
  - `/api/me`
  - `/api/accounts/*`
  - `/api/rules/*`
- shared error/envelope conventions across both API surfaces
- future non-HTTP streamed/event payloads if introduced later

## Target proto layout

Recommended long-term layout:

```text
proto/
  smailnail/
    annotationui/
      v1/
        review.proto
        annotation.proto
    app/
      v1/
        hosted.proto
        common.proto
```

The initial implementation can keep common messages duplicated if that lowers migration risk, but the long-term direction should be to share genuinely common wire concepts once they stabilize.

## Generation layout

Generated outputs should remain:

```text
pkg/gen/smailnail/.../v1/*.pb.go
ui/src/gen/smailnail/.../v1/*.ts
```

The generation path should continue to be driven by:

```bash
go generate ./pkg/annotationui
```

and, once the hosted API slice is added, by a matching repo-local generator entrypoint for that surface as well, or a small shared `go generate` target if the repository wants a single generation command.

## Contract conventions

### 1. Transport boundary only

Generated protobuf types belong at the transport edge only. They should not replace domain/storage structs in services or repositories.

### 2. Wrapper list responses

List endpoints should prefer wrapper messages with `items`, even if the current API returned bare arrays before.

Example:

```proto
message RuleListResponse {
  repeated RuleRecord data = 1;
  RuleListMeta meta = 2;
}
```

For annotation UI endpoints, `items` is already the chosen convention. For the hosted API, preserving the current `data` + `meta` envelope is acceptable because the frontend already depends on it.

### 3. Preserve v1 JSON shape when practical

For existing APIs, prefer preserving:

- camelCase field names
- string status values
- string timestamps
- existing envelope layout where it is already deeply integrated

### 4. `protojson` everywhere for generated JSON payloads

All generated HTTP request/response handling should use `protojson`, not `encoding/json`, to avoid snake_case/camelCase drift.

### 5. Strings vs enums

Use strings for status/scope/kind fields in existing v1 APIs unless there is a compelling reason to break the shape. Consider enums only for new greenfield surfaces or a future v2.

### 6. Dynamic objects

Use `google.protobuf.Struct` only where truly dynamic JSON is unavoidable, such as query result rows or action-plan payloads.

### 7. Frontend wrapper types are allowed, but secondary

Frontend wrapper files may narrow generated `string` fields into literal unions for component ergonomics, but they should not re-handwrite the full object shape.

## Hosted API target shape

The hosted API currently uses a success envelope:

- `data`
- optional `meta`

and an error envelope:

- `error.code`
- `error.message`
- optional `error.details`

That shape should be preserved in v1, but expressed as generated protobuf messages. Because protobuf does not provide ergonomic generics for this pattern, each endpoint family should use explicit response messages.

Examples:

- `ListAccountsResponse { repeated AccountListItem data = 1; ListAccountsMeta meta = 2; }`
- `GetCurrentUserResponse { CurrentUser data = 1; }`
- `ListMessagesResponse { repeated MessageView data = 1; ListMessagesMeta meta = 2; }`
- `RuleResponse { RuleRecord data = 1; }`
- `DryRunRuleResponse { DryRunResult data = 1; }`
- `ErrorResponse { ApiError error = 1; }`

## Rollout plan

### Phase 1 — Annotation UI review slice
Done.

### Phase 2 — Annotation UI read/query slice
Done.

### Phase 3 — Hosted web API slice
Implement shared protobuf contracts for:

- current user
- accounts CRUD/test
- mailboxes/messages views
- rules CRUD
- rule dry-run
- shared info response
- shared error envelope

### Phase 4 — Common contract consolidation
After the hosted migration lands, evaluate whether common messages should move into a shared `common.proto` without creating churn too early.

### Phase 5 — Streaming/event surfaces
If the repo introduces websocket/SSE event payloads, model them in protobuf rather than ad hoc JSON.

## Impact by layer

### Backend

- handlers become uniform in request decode / response encode style
- mapping code increases slightly but becomes much clearer
- error responses become consistent and typed

### Frontend

- API methods import generated request/response types
- wrapper types become thinner and more stable
- mock data and story data gain a single contract source of truth

### Tests

- server tests decode responses into generated proto messages
- integration tests validate real wire shape through `protojson`

## Risks and mitigations

| Risk | Why it matters | Mitigation |
|------|----------------|------------|
| Massive one-shot refactor | Too much churn, hard to review | Convert subsystem by subsystem with commits at stable boundaries |
| Over-generating internal structs | Pollutes service/repository code with transport concerns | Keep mapping files at the handler boundary |
| Envelope inconsistency between subsystems | Frontend complexity and surprise | Document and preserve existing hosted `data/meta` vs annotation `items` conventions explicitly |
| Query/event dynamic payloads undermine type safety | Too much `Struct` defeats schema-first design | Restrict `Struct` use to genuinely dynamic payloads |
| Frontend wrappers reintroduce drift | Generated code stops being authoritative | Limit wrappers to narrow unions and ergonomic aliases only |

## Definition of done for repo-wide unification

The repo reaches the intended state when:

- every web UI-visible HTTP payload is defined in protobuf
- generated Go and TS outputs are committed and used
- backend handlers encode/decode generated wire messages via `protojson`
- frontend API/types use generated contracts as the source of truth
- mocks/stories/tests align with the generated contract
- domain/repository structs remain handwritten and transport-agnostic

## Review checklist

When reviewing future contract-unification work, verify:

- schema lives in `proto/...`
- generated outputs are updated
- handler request/response logic uses generated types
- frontend types derive from generated types
- mocks/stories were updated too
- tests decode generated responses rather than ad hoc JSON structs
