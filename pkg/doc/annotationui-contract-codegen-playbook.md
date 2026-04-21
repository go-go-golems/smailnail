---
Title: Annotation UI Contract Codegen Playbook
Slug: annotationui-contract-codegen-playbook
Short: Add or extend shared protobuf-based Go and TypeScript wire contracts for the annotation UI using Buf, go generate, backend mappers, and frontend wrapper types.
Topics:
- annotations
- backend
- frontend
- workflow
- glazed
Commands:
- go generate
- buf generate
- buf lint
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This playbook explains how to add new HTTP wire contracts to the shared protobuf code generation pipeline used in smailnail. It started in the annotation UI, but the same workflow also applies to the hosted web API and future UI-visible wire surfaces. The goal is not to replace repository or database structs wholesale. The goal is to make the wire layer explicit, generated, and shared between Go and TypeScript so that backend handlers, frontend API code, mocks, and tests stop drifting apart.

## Why Use The Shared Contract Workflow

The annotation UI has several places where the same payload shape can drift:

- backend handlers decode or encode JSON by hand,
- frontend API code declares parallel TypeScript interfaces,
- Storybook/MSW mocks quietly keep old field names alive,
- tests validate the old shape until a runtime integration finally breaks.

The protobuf + Buf + `go generate` workflow fixes that by giving the project one source of truth for the wire contract. Generated Go structs anchor the HTTP boundary, and generated TypeScript interfaces anchor the frontend API layer. Small handwritten mapper files still belong in the design, because domain and storage structs should stay local to the backend.

## What This Playbook Covers

Use this workflow when you need to:

- add a new annotation UI endpoint payload,
- add a new hosted web API payload (`/api/me`, `/api/accounts/*`, `/api/rules/*`, etc.),
- migrate an existing handwritten JSON DTO to generated code,
- add a new list wrapper response with `items`,
- add a hosted-API response envelope with `data` + `meta`,
- add a new request body for a POST or PATCH endpoint,
- extend frontend wrapper types without duplicating the raw contract shape.

The project currently uses two response-envelope conventions:

- annotation UI endpoints generally return wrapper responses with `items`
- hosted web API endpoints generally return wrapper responses with `data` and optional `meta`

Both conventions are acceptable as long as they are expressed explicitly in protobuf and preserved consistently through backend handlers, frontend clients, mocks, and tests.

Do not use this workflow to:

- replace `pkg/annotate` repository structs everywhere,
- introduce gRPC,
- push generated protobuf types deep into persistence code.

## How This Aligns With The Generic Protobuf Go/TS Skill

This playbook is intentionally aligned with the general `protobuf-go-ts-schema-exchange` approach:

- schema-first protobuf contracts,
- Buf v2 generation,
- Go and TypeScript generated from the same schema,
- `protojson` at the Go HTTP boundary,
- generated protobuf types kept at the transport edge with handwritten mapper code below,
- explicit validation after schema changes.

The main repo-specific difference is the TypeScript generator strategy.

The generic skill assumes a runtime-oriented TypeScript flow using `protoc-gen-es`, `@bufbuild/protobuf`, and `fromJson(...)`. This repo currently uses `ts-proto` instead, with plain generated TypeScript interfaces and without generated runtime JSON helpers:

- `buf.gen.yaml` uses `buf.build/community/stephenh-ts-proto`
- `outputEncodeMethods=false`
- `outputJsonMethods=false`
- `outputClientImpl=false`

That means the frontend usually consumes generated shapes as types and keeps JSON transport logic in the existing API layer, instead of decoding wire payloads with a protobuf runtime in the browser.

Because of that choice:

- keep using `protojson` in Go so HTTP JSON matches the proto schema,
- keep treating generated TS output as the contract source of truth,
- do not add parallel handwritten TS DTOs unless you are creating a very small wrapper alias for UI ergonomics,
- revisit the generator choice only if a future UI surface truly needs runtime protobuf JSON decoding or encoding in TypeScript.

## Prerequisites

- `buf` installed and available on `PATH`
- Go toolchain available
- frontend dependencies installed in `ui/`
- familiarity with the current generated files under:
  - `pkg/gen/smailnail/annotationui/v1/`
  - `pkg/gen/smailnail/app/v1/`
  - `ui/src/gen/smailnail/annotationui/v1/`
  - `ui/src/gen/smailnail/app/v1/`

Before editing anything, verify the generation toolchain works:

```bash
cd smailnail
buf lint
go generate ./pkg/annotationui ./pkg/smailnaild
```

If either command fails before you start, fix that first. Otherwise you will not know whether later failures come from your new schema or from a broken baseline.

## Step 1: Choose The Boundary Carefully

Start by deciding whether the new shape is truly part of the HTTP wire layer. That sounds obvious, but it is the most important architectural decision in this workflow.

Good candidates:

- request bodies sent by the UI,
- response objects rendered directly by the UI,
- list wrapper responses like `{ items: [...] }`,
- small filter/request DTOs that the frontend uses to describe an endpoint.

Bad candidates:

- internal SQL row structs,
- repository input structs that carry domain-only concerns,
- low-level storage rows that are not part of the public HTTP contract.

The rule of thumb is simple: generate at the handler boundary, map below it.

## Step 2: Add Or Extend A Proto File Under `proto/`

Put new schema files under the subsystem-specific `proto/` package path:

- `proto/smailnail/annotationui/v1/` for annotation UI contracts
- `proto/smailnail/app/v1/` for hosted web API contracts

Reuse the matching package and `go_package` for the subsystem you are extending:

```proto
syntax = "proto3";

package smailnail.annotationui.v1;

option go_package = "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1;annotationuiv1";
```

```proto
syntax = "proto3";

package smailnail.app.v1;

option go_package = "github.com/go-go-golems/smailnail/pkg/gen/smailnail/app/v1;appv1";
```

Use separate files when it keeps reviews clear. The current split is a good model:

- `review.proto` for feedback/guideline/review-action payloads
- `annotation.proto` for the broader annotation UI read models and query workbench payloads
- `hosted.proto` for hosted web API info, identity, accounts, message, and rule payloads

Prefer these conventions:

- keep wire status/kind/scope values as `string` in v1 when the REST shape already uses strings,
- use `optional` only when presence matters,
- use wrapper list responses such as `AnnotationListResponse { repeated Annotation items = 1; }`,
- use `google.protobuf.Struct` only for truly dynamic JSON maps such as query result rows.

A representative list wrapper looks like this:

```proto
message AnnotationListResponse {
  repeated Annotation items = 1;
}
```

That wrapper is worth the extra line of schema because it makes backend encoding, frontend transforms, and future pagination changes much easier to manage.

## Step 3: Preserve The Existing JSON Shape Deliberately

The generated contract should usually preserve the existing REST JSON shape unless you are intentionally making a breaking API change.

That means:

- keep camelCase JSON field names by relying on `protojson`,
- keep string timestamps if the API already exposes timestamps as strings,
- keep list responses stable unless you are intentionally migrating callers,
- make any breaking rename explicit in the same ticket for backend, frontend, mocks, and tests.

For example, this project intentionally standardized list endpoints around wrapper responses with `items`. Once that decision is made, update all consumers in the same change set rather than trying to carry both shapes indefinitely.

## Step 4: Choose The Right Response Envelope

Before you regenerate code, decide which response-envelope convention the endpoint belongs to.

Use `items` wrappers when you are working in the annotation UI family and the endpoint is already aligned with that style:

```proto
message AnnotationListResponse {
  repeated Annotation items = 1;
}
```

Use explicit `data` + `meta` wrappers when you are working in the hosted web API family and the frontend already expects that envelope:

```proto
message ListAccountsMeta {
  int32 count = 1;
}

message ListAccountsResponse {
  repeated AccountListItem data = 1;
  ListAccountsMeta meta = 2;
}
```

Do not silently flatten a long-standing hosted endpoint from `data` into bare arrays or `items` unless you are deliberately doing a breaking API migration and updating all callers in the same change.

## Step 5: Regenerate Go And TypeScript Types

After editing the proto files, regenerate everything through the repo-local workflow:

```bash
cd smailnail
go generate ./pkg/annotationui ./pkg/smailnaild
```

If you are touching only one subsystem, you can run just that package's generator entrypoint, but the safe default is to run both.

Do not run ad hoc generator commands from memory unless you are debugging the generator itself. The repo-local `go generate` entrypoint is the supported path and the one future maintainers should trust.

Generated outputs should land here:

- Go: `pkg/gen/smailnail/annotationui/v1/`, `pkg/gen/smailnail/app/v1/`
- TypeScript: `ui/src/gen/smailnail/annotationui/v1/`, `ui/src/gen/smailnail/app/v1/`

The current TypeScript output is generated by `ts-proto` as plain types/interfaces. This repo does not currently rely on generated TS encode/decode helpers or an `@bufbuild/protobuf` runtime in the frontend path.

If you add `google.protobuf.Struct` or other well-known types, expect extra generated TypeScript support files such as `ui/src/gen/google/protobuf/struct.ts`.

## Step 6: Update Backend Handler Mappers

Generated messages should not leak straight into repository code. Add or extend mapper helpers in `pkg/annotationui/contracts_*.go`.

Typical responsibilities for these mapper files:

- convert repository/domain structs into generated response messages,
- convert generated request messages into repository input structs,
- normalize timestamps into wire strings,
- handle dynamic values such as `structpb.Struct` creation for query rows.

A minimal example looks like this:

```go
func annotateAnnotationToProto(annotation *annotate.Annotation) *annotationuiv1.Annotation {
    if annotation == nil {
        return nil
    }

    return &annotationuiv1.Annotation{
        Id:           annotation.ID,
        TargetType:   annotation.TargetType,
        TargetId:     annotation.TargetID,
        Tag:          annotation.Tag,
        NoteMarkdown: annotation.NoteMarkdown,
        SourceKind:   annotation.SourceKind,
        SourceLabel:  annotation.SourceLabel,
        AgentRunId:   annotation.AgentRunID,
        ReviewState:  annotation.ReviewState,
        CreatedBy:    annotation.CreatedBy,
        CreatedAt:    formatProtoTime(annotation.CreatedAt),
        UpdatedAt:    formatProtoTime(annotation.UpdatedAt),
    }
}
```

The key idea is that protobuf belongs at the transport edge, not at the center of the application.

## Step 7: Use `protojson` At The HTTP Boundary

When a handler reads or writes generated protobuf messages, use the shared helpers in `pkg/annotationui/protojson.go`.

Use:

- `decodeProtoJSONBody(...)` for POST/PATCH request bodies
- `writeProtoJSON(...)` for generated responses

Do not switch back to `encoding/json` for generated protobuf payloads. The generated Go structs carry snake_case `json` tags, while `protojson` emits the correct camelCase field names from the proto descriptors.

A typical handler shape is:

```go
req := &annotationuiv1.ExecuteQueryRequest{}
if !decodeProtoJSONBody(w, r, req) {
    return
}

payload, err := queryResultToProto(result)
if err != nil {
    writeMessageError(w, http.StatusInternalServerError, err.Error())
    return
}

writeProtoJSON(w, http.StatusOK, payload)
```

For GET endpoints, parse query params manually and then map them into generated request/filter messages if that improves consistency.

## Step 8: Update Frontend Wrapper Types Instead Of Re-Handwriting DTOs

Do not scatter raw generated types through every React component unless they already fit perfectly. The preferred pattern in this repo is:

1. import generated interfaces into `ui/src/types/*.ts`
2. create small wrapper aliases that narrow string unions for UI ergonomics
3. keep API/mocks/components importing the wrapper types

That preserves one contract source of truth while keeping component code readable.

Example pattern:

```ts
import type { Annotation as GeneratedAnnotation } from "../gen/smailnail/annotationui/v1/annotation";

export type Annotation = Omit<GeneratedAnnotation, "sourceKind" | "reviewState"> & {
  sourceKind: SourceKind;
  reviewState: ReviewState;
};
```

This is the right compromise when the wire contract says `string` but the UI benefits from narrower literal unions.

## Step 9: Update Frontend API Clients, RTK Query, And Mock Handlers Together

Whenever a response shape changes, update the relevant frontend API layer in the same pass:

- `ui/src/api/annotations.ts` for annotation UI RTK Query endpoints
- `ui/src/api/client.ts` and `ui/src/api/types.ts` for the hosted web API
- `ui/src/mocks/handlers.ts`
- any stories or fixtures that bypass the main API layer

This matters especially for wrapper list responses. When the backend returns `{ items: [...] }`, the API layer should usually unwrap it with `transformResponse`, and the MSW mocks must return the same wrapper.

Example:

```ts
listAnnotations: builder.query<Annotation[], AnnotationFilter>({
  query: (filter) => ({ url: "annotations", params: filter }),
  transformResponse: (response: AnnotationListResponse) => response.items,
})
```

If you forget to update mocks, Storybook or local dev can continue “working” while production code and tests silently diverge.

## Step 10: Add Or Update Backend Contract Tests

Every migrated endpoint should have a backend test that validates the generated contract shape. The current contract tests live in:

- `pkg/annotationui/server_test.go`

Use `decodeProtoJSONResponse(...)` for generated responses. That ensures the handler emits payloads that are actually compatible with the proto schema, not just vaguely similar JSON.

Good test targets include:

- list endpoints returning wrapper responses,
- detail endpoints returning generated messages,
- request-body endpoints decoded through `protojson`,
- query endpoints with dynamic rows,
- end-to-end artifact creation when a request fans out into feedback/guideline side effects.

## Step 11: Validate The Whole Contract Loop

Run the full validation loop before committing:

```bash
cd smailnail
buf lint
go generate ./pkg/annotationui ./pkg/smailnaild
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate ./pkg/smailnaild -count=1

cd ui
pnpm run check
```

If the repo hook runs broader checks on commit, let it. The explicit validation above gives you fast feedback, while the commit hook catches repo-wide fallout.

## Step 12: Document The Migration

When the contract surface changes, update the active ticket diary and changelog so future work can answer:

- what endpoints moved to generated types,
- what wire shape changed,
- what helper files now own the mapping,
- what validation commands passed,
- what subtle issue had to be fixed.

This is especially important when introducing wrapper responses or dynamic `Struct` rows, because those are easy to rediscover the hard way.

## Recommended File Checklist

For a typical contract extension, expect to touch some subset of these files:

| Area | Typical Files |
|------|---------------|
| Schema | `proto/smailnail/annotationui/v1/*.proto`, `proto/smailnail/app/v1/*.proto` |
| Generation | `buf.yaml`, `buf.gen.yaml`, repo-local `go generate` entrypoints |
| Go mappers | `pkg/annotationui/contracts_*.go`, `pkg/smailnaild/contracts_*.go` |
| Go handlers | `pkg/annotationui/handlers_*.go`, `pkg/annotationui/server.go`, `pkg/smailnaild/http.go` |
| Generated Go | `pkg/gen/smailnail/.../*.pb.go` |
| TS wrappers | `ui/src/types/*.ts`, `ui/src/api/types.ts` |
| TS API | `ui/src/api/annotations.ts`, `ui/src/api/client.ts` |
| Generated TS | `ui/src/gen/smailnail/.../*.ts` |
| Mocks/Stories | `ui/src/mocks/*.ts`, affected story files |
| Validation | `pkg/annotationui/server_test.go`, `pkg/smailnaild/http_test.go`, integration tests as needed |

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `buf lint` says files are in the wrong directory for the package | Buf module root does not match the `proto/` directory layout | Keep `buf.yaml` rooted at `proto` so `smailnail.annotationui.v1` maps to `proto/smailnail/annotationui/v1/` |
| Generated Go imports point at the wrong package path | `option go_package` does not match the committed generated directory | Set `go_package` to `github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1;annotationuiv1` |
| JSON fields come out snake_case instead of camelCase | Handler used `encoding/json` on generated protobuf structs | Use `writeProtoJSON` / `decodeProtoJSONBody` instead |
| Frontend compiles but Storybook/MSW behavior is wrong | Mock handlers still return the old handwritten shape | Update `ui/src/mocks/handlers.ts` in the same change as the API layer |
| UI complains that `string` is not assignable to a narrower union | Generated interfaces expose raw `string` fields | Narrow those fields in `ui/src/types/*.ts` wrapper aliases instead of re-handwriting the full DTO |
| Query result rows fail to convert | Row values include types `structpb` cannot serialize directly | Normalize rows first, then convert maps with `structpb.NewStruct` |
| `int64` or large counters become awkward in TS | The generic protobuf Go/TS flow often expects runtime decoding into `bigint`, but this repo currently uses `ts-proto` plain-type output | Prefer string IDs and existing stable wire shapes in v1; if true 64-bit numeric semantics become important, document the behavior explicitly and reassess TS generator/runtime choices |
| GET filter types drift even though POST bodies are generated | Query params are still parsed ad hoc in multiple places | Add generated filter/request messages and map query params into them in handlers and frontend wrapper types |

## See Also

- [Annotate And Query The Mirror SQLite DB](./annotate-sqlite-playbook.md) — operational playbook for the underlying annotation data model and CLI workflow.
- `proto/smailnail/annotationui/v1/review.proto` — example of the first contract slice migrated to generated code.
- `proto/smailnail/annotationui/v1/annotation.proto` — example of the broader annotation UI read/query contract slice.
- `proto/smailnail/app/v1/*.proto` — hosted web API contract slices once that migration is in place.
- `pkg/annotationui/protojson.go` — shared HTTP encode/decode helpers for generated protobuf JSON.
- `pkg/annotationui/contracts_annotation.go` and `pkg/annotationui/contracts_review.go` — mapper pattern for keeping protobuf at the transport edge.
