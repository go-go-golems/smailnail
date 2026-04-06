---
Title: Implementation plan for shared feedback and guideline contract codegen
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
    - Path: buf.gen.yaml
      Note: Code generation targets for Go and ts-proto output
    - Path: buf.yaml
      Note: Buf module configuration rooted at proto for lint/generation
    - Path: cmd/generate-annotationui-contracts/main.go
      Note: Repo-local Go generator command used by go generate
    - Path: pkg/annotationui/contracts_review.go
      Note: Mappings between generated wire types and annotate domain structs
    - Path: pkg/annotationui/generate.go
      Note: go:generate entrypoint for contract generation
    - Path: pkg/annotationui/handlers_annotations.go
      Note: Review-action request decoding migrated to generated contracts
    - Path: pkg/annotationui/handlers_feedback.go
      Note: Backend feedback/guideline handlers migrated to generated contracts
    - Path: pkg/annotationui/protojson.go
      Note: Shared protojson encode/decode helpers for generated wire types
    - Path: pkg/annotationui/server_test.go
      Note: Backend contract tests added for generated request/response flows
    - Path: pkg/gen/smailnail/annotationui/v1/review.pb.go
      Note: Generated Go contract output
    - Path: proto/smailnail/annotationui/v1/review.proto
      Note: Shared IDL introduced for feedback
    - Path: ui/src/api/annotations.ts
      Note: RTK Query contract updated to generated interfaces and wrapper responses
    - Path: ui/src/gen/smailnail/annotationui/v1/review.ts
      Note: Generated TypeScript contract output
    - Path: ui/src/mocks/handlers.ts
      Note: Mocks aligned to generated contract and wrapper list responses
    - Path: ui/src/types/reviewFeedback.ts
      Note: Frontend feedback wrapper types now derive from generated code
    - Path: ui/src/types/reviewGuideline.ts
      Note: Frontend guideline wrapper types now derive from generated code
ExternalSources: []
Summary: Introduce a go-generate driven shared IDL and codegen pipeline for annotation review feedback and guidelines, then migrate the backend HTTP layer and frontend type layer to the generated contract.
LastUpdated: 2026-04-06T21:05:00Z
WhatFor: Remove TS/Go contract drift around review feedback and guidelines.
WhenToUse: Read before implementing or reviewing shared contract codegen for the annotation UI.
---


# Implementation plan for shared feedback and guideline contract codegen

## Executive Summary

The current review-feedback/guideline slice has drift between Go and TypeScript, most visibly in the create-feedback payload (`targets` vs `targetIds`) and partially in update payload expectations. The goal of this ticket is to introduce a **shared IDL plus generated code** for the annotation UI review contracts, driven by **`go generate`** and a repository-local Go generator command.

This implementation is intentionally scoped. It will unify the **wire contract** for:

- review feedback entities and mutations,
- review guideline entities and mutations,
- run-guideline linking,
- review-action request payloads that embed review comments/guideline IDs.

It will **not** try to replace the full annotation domain model or every list/query DTO in one pass.

## Problem Statement

Right now the review feature has at least three layers of type definitions:

1. Go domain/repository structs in `pkg/annotate/types.go`
2. Go HTTP DTOs in `pkg/annotationui/types_feedback.go`
3. TypeScript frontend DTOs in `ui/src/types/reviewFeedback.ts` and `ui/src/types/reviewGuideline.ts`

These layers are close enough to look consistent but far enough apart to drift. The most important live example is:

- TypeScript create feedback request: `targetIds?: string[]`
- Go handler request DTO: `targets: [{ targetType, targetId }]`

This is precisely the sort of mismatch that code generation should prevent.

## Proposed Solution

### IDL choice

Use **protobuf schemas** as the shared IDL, but with a deliberate compatibility choice:

- use **string fields** for review-state / kind / status values for now,
- not protobuf enums yet,
- so that the current JSON wire format stays stable and human-readable.

This keeps the migration low-risk while still giving us one source of truth for field names and payload shapes.

### Codegen choice

Use a **Go-based generator command** invoked by `go generate`, which in turn runs `buf generate`.

Key properties:

- generation is repo-local and reproducible,
- developers do not need to remember the exact `buf generate` incantation,
- CI and local workflow can rely on `go generate ./...`.

### TS plugin choice

Use the `ts-proto` plugin via Buf remote plugins so that TypeScript output is **plain interfaces**, not message-runtime branded objects. That avoids the extra frontend runtime complexity of `@bufbuild/protobuf` message decoding for simple REST JSON payloads.

### Backend integration choice

Use the generated Go protobuf structs at the **HTTP boundary** and serialize them with `protojson`.

That means:

- request decoding for feedback/guideline/review-action payloads uses generated types,
- response encoding for feedback/guideline endpoints uses generated types,
- repository/domain structs remain hand-written and are mapped to/from generated wire structs.

### Frontend integration choice

Frontend API types will move to the generated TS interfaces. Small handwritten wrapper files may remain for:

- query-filter types,
- local UI convenience unions/constants,
- RTK Query list-response unwrapping.

## Why this scope

This is the highest-value slice because it directly addresses the current contract drift without requiring a full conversion of every annotation endpoint.

In scope:

- `review-feedback` endpoints
- `review-guidelines` endpoints
- `annotation-runs/{id}/guidelines` endpoints
- request payloads for `annotations/{id}/review` and `annotations/batch-review`

Out of scope for this ticket:

- converting all annotation/group/log/run responses to generated contracts,
- introducing protobuf enums or timestamps,
- replacing repository/domain model structs,
- switching to gRPC or Connect.

## Architecture Diagram

```text
proto/smailnail/annotationui/v1/review.proto
        |
        v
cmd/generate-annotationui-contracts (Go tool)
        |
        v
    buf generate
      /      \
     v        v
pkg/gen/...   ui/src/gen/...
     |            |
     |            +--> RTK Query + mocks + UI types use generated TS interfaces
     |
     +--> annotationui handlers decode/encode generated Go messages via protojson
                 |
                 v
          annotate.Repository domain types
```

## Files to Add

### IDL and generation

- `buf.yaml`
- `buf.gen.yaml`
- `proto/smailnail/annotationui/v1/review.proto`
- `cmd/generate-annotationui-contracts/main.go`
- `pkg/annotationui/generate.go`

### Generated outputs

- `pkg/gen/smailnail/annotationui/v1/review.pb.go`
- `ui/src/gen/smailnail/annotationui/v1/review.ts`

### Backend integration helpers

Likely one or more of:

- `pkg/annotationui/protojson.go`
- `pkg/annotationui/contracts_review.go`

### Frontend compatibility wrappers

Likely updates to:

- `ui/src/types/reviewFeedback.ts`
- `ui/src/types/reviewGuideline.ts`
- `ui/src/api/annotations.ts`
- `ui/src/mocks/handlers.ts`
- `ui/src/mocks/annotations.ts`

## Proto schema shape

The schema should include at least:

- `FeedbackTarget`
- `ReviewFeedback`
- `ReviewFeedbackListResponse`
- `CreateFeedbackRequest`
- `UpdateFeedbackRequest`
- `ReviewComment`
- `ReviewAnnotationRequest`
- `BatchReviewRequest`
- `ReviewGuideline`
- `ReviewGuidelineListResponse`
- `CreateGuidelineRequest`
- `UpdateGuidelineRequest`
- `LinkRunGuidelineRequest`

Important design note:

- list endpoints will use explicit wrapper messages with `items`,
- so backend encoding can stay protojson-based without inventing custom array serializers.

## Implementation Tasks

### Task 1 — Add shared schema + generator plumbing

Deliverables:

- buf config files
- proto schema
- generator command
- `go:generate` directive

Validation:

```bash
go generate ./pkg/annotationui
```

### Task 2 — Generate and commit Go + TS contract outputs

Deliverables:

- generated Go contract code
- generated TS contract code
- `go.mod` updates for protobuf runtime if needed

Validation:

```bash
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1
cd ui && pnpm run check
```

### Task 3 — Migrate backend HTTP layer to generated contracts

Deliverables:

- generated request decoding for feedback/guideline endpoints
- generated protojson response writing for feedback/guideline endpoints
- generated request decoding for review-action payloads
- mapping helpers between generated wire types and domain structs

Validation:

- targeted server tests for feedback/guideline endpoints
- `go test -tags sqlite_fts5 ./pkg/annotationui -count=1`

### Task 4 — Migrate frontend API/mocks/types to generated contracts

Deliverables:

- RTK Query types switched to generated interfaces
- create-feedback payload corrected to `targets`
- list wrappers unwrapped in `transformResponse`
- mocks/stories aligned with generated contract

Validation:

```bash
cd ui && pnpm run check
```

### Task 5 — Validate end-to-end, update docs, and commit

Deliverables:

- updated ticket docs
- updated diary with commands and any failures
- focused git commit with code + docs

Validation:

```bash
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1
cd ui && pnpm run check
```

## Implementation Outcome

This plan has now been executed in the scope described above.

Delivered in code:

- Added `proto/smailnail/annotationui/v1/review.proto`
- Added `buf.yaml` / `buf.gen.yaml`
- Added a repo-local generator command: `cmd/generate-annotationui-contracts`
- Added `pkg/annotationui/generate.go` so generation runs via `go generate`
- Generated and committed:
  - Go protobuf contracts under `pkg/gen/smailnail/annotationui/v1/`
  - TypeScript interfaces under `ui/src/gen/smailnail/annotationui/v1/`
- Migrated the backend feedback/guideline handlers to generated contract types and `protojson`
- Migrated review-action request payload decoding (`annotations/{id}/review`, `annotations/batch-review`) to generated contract types
- Migrated the frontend feedback/guideline type layer and RTK Query contract to generated TypeScript interfaces
- Fixed the create-feedback contract drift by standardizing on `targets`
- Updated mocks and story overrides to match the generated contract and wrapper list responses
- Added backend tests covering generated-contract request/response flows

Notable deliberate choices:

1. **String-valued workflow fields were kept as strings** in the schema to preserve the current REST JSON shape.
2. **List endpoints now return explicit wrapper messages** with `items`, and the frontend unwraps them with `transformResponse`.
3. **Repository/domain structs remain hand-written**; only the HTTP wire layer was code-generated in this ticket.

## Alternatives Considered

### 1. Keep handwritten DTOs and just fix the mismatch

Rejected because it fixes today’s bug but not the structural drift problem.

### 2. Use protobuf enums and timestamps immediately

Deferred because it would either:

- change the JSON wire format too much, or
- require a larger mapping layer than is justified for this first pass.

### 3. Use `@bufbuild/protobuf` message-runtime TS types

Rejected for this ticket because the frontend uses plain REST JSON and RTK Query. `ts-proto` plain interfaces are easier to adopt incrementally.

### 4. Convert the entire annotation API at once

Rejected as too large. This ticket focuses on the feedback/guideline slice where the drift already exists.

## Risks

1. Wrapper list responses (`items`) will require coordinated backend + frontend updates.
2. `protojson` output must be used consistently; stdlib `encoding/json` would emit snake_case for generated Go structs.
3. Generated TS interfaces will not automatically give us literal-string enum safety while we keep the wire format as strings.
4. Existing mocks and tests may break in subtle ways because of wrapper response shapes.

## Review Notes

A reviewer of this ticket should confirm:

1. generation is reproducible through `go generate`,
2. the TS and Go code are actually driven by the same `.proto`,
3. create-feedback now uses `targets` consistently everywhere,
4. list responses are consistently wrapped and unwrapped,
5. no unrelated branch noise was staged into the commit.

## References

- `pkg/annotationui/handlers_feedback.go`
- `pkg/annotationui/handlers_annotations.go`
- `pkg/annotationui/server.go`
- `pkg/annotate/types.go`
- `pkg/annotate/repository_feedback.go`
- `ui/src/api/annotations.ts`
- `ui/src/types/reviewFeedback.ts`
- `ui/src/types/reviewGuideline.ts`
