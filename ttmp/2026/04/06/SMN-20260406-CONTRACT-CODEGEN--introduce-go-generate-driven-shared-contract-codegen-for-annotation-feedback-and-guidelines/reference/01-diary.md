---
Title: Diary
Ticket: SMN-20260406-CONTRACT-CODEGEN
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - workflow
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: buf.yaml
      Note: Diary records the lint/configuration fix for the proto-rooted module
    - Path: pkg/annotationui/server_test.go
      Note: Diary records the added contract tests
    - Path: pkg/doc/annotationui-contract-codegen-playbook.md
      Note: Diary step for the reusable Glazed help-style playbook
    - Path: proto/smailnail/annotationui/v1/annotation.proto
      Note: Diary step for phase-2 annotation UI contract expansion
    - Path: ui/src/mocks/handlers.ts
      Note: Diary records the mock/stories contract migration
ExternalSources: []
Summary: Step-by-step diary for introducing a go-generate driven shared contract codegen flow for annotation review feedback and guidelines.
LastUpdated: 2026-04-06T21:20:00Z
WhatFor: Capture what was implemented, how generation was wired, what broke during validation, and how it was fixed.
WhenToUse: Read this when reviewing or continuing the shared contract codegen work.
---



# Diary

## Goal

This diary records the implementation of shared contract codegen for the annotation review feedback/guideline slice. The intent was to remove the most obvious Go/TypeScript contract drift while keeping the migration narrow enough to finish safely in one ticket.

## Step 1: Define the scope and add generation plumbing

I started by writing the implementation plan and then constraining the scope to the wire layer: feedback entities, guideline entities, run-guideline linking, and the review-action request payloads that embed review comments and guideline IDs. I intentionally did not try to convert the full annotation API in one pass.

The important design decision in this step was to use protobuf as the shared IDL, but to keep workflow/status/kind fields as strings in v1. That preserved the current REST JSON shape and made the migration tractable. I also followed the user's instruction to use the Go tool / `go generate` approach by adding a repo-local Go generator command that runs `buf generate`.

### Prompt Context

**User prompt (verbatim):** "alright, create a new ticket, write implementation plan, then add tasks, then execute the tasks one by one, and then commit and keep a diary."

**Assistant interpretation:** Create a new docmgr ticket, define a scoped implementation plan, track tasks explicitly, implement them in sequence, validate, commit, and document the process.

**Inferred user intent:** Turn the earlier design discussion into real code and a documented workflow, not just advice.

### What I did
- Created ticket `SMN-20260406-CONTRACT-CODEGEN` under `smailnail/ttmp`.
- Wrote the implementation plan design doc.
- Added explicit tasks.
- Added:
  - `buf.yaml`
  - `buf.gen.yaml`
  - `proto/smailnail/annotationui/v1/review.proto`
  - `cmd/generate-annotationui-contracts/main.go`
  - `pkg/annotationui/generate.go`
- Ran:

```bash
cd smailnail
go generate ./pkg/annotationui
```

### Why
- The current feedback/guideline wire contract was duplicated in too many places.
- `go generate` makes regeneration discoverable and repo-local.
- `ts-proto` plain interfaces are easier to adopt in a React + RTK Query app than runtime-branded message objects.

### What worked
- Buf generation succeeded through the Go generator command.
- The generated TS output was plain interfaces, which fit the frontend nicely.

### What didn't work
- N/A in this step.

### What I learned
- The Go-command wrapper around Buf gives the right ergonomics for this repo.
- Keeping v1 values as strings is a pragmatic compromise; it unifies field names and shapes without forcing a wire-format redesign.

### What was tricky to build
- The tricky part was choosing the boundary: generated wire layer only, not generated repository/domain types.

### What warrants a second pair of eyes
- The long-term decision about whether v2 should move workflow/status fields from strings to enums.

### What should be done in the future
- If the pattern works well, expand generation to the rest of the annotation UI wire layer incrementally.

### Code review instructions
- Start with:
  - `proto/smailnail/annotationui/v1/review.proto`
  - `cmd/generate-annotationui-contracts/main.go`
  - `pkg/annotationui/generate.go`

### Technical details
- The generator command walks upward to find `go.mod` and then runs `buf generate` from the repo root.

## Step 2: Generate code and migrate the backend HTTP layer

Once the schema existed, I generated the Go and TS outputs and then migrated the backend HTTP layer to use generated request/response types for the feedback/guideline endpoints. I also switched the review-action request bodies for `PATCH /api/annotations/{id}/review` and `POST /api/annotations/batch-review` to generated contract types, since those requests carry the same review-comment/guideline-link structures.

The key implementation detail here was `protojson`. Generated Go protobuf structs use snake_case `json` tags, so using the standard library JSON codec would have broken the existing camelCase API shape. `protojson` preserved the expected field names while still letting the handlers use generated types.

### What I did
- Generated and committed:
  - `pkg/gen/smailnail/annotationui/v1/review.pb.go`
  - `ui/src/gen/smailnail/annotationui/v1/review.ts`
- Added backend helpers:
  - `pkg/annotationui/protojson.go`
  - `pkg/annotationui/contracts_review.go`
- Rewrote:
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/handlers_annotations.go`
- Removed now-obsolete `pkg/annotationui/types_feedback.go`.
- Added backend tests in `pkg/annotationui/server_test.go` covering generated-contract request/response flows.

### Why
- The backend had to become a real consumer of the generated contract, otherwise the TS side would still be “generated against a spec the server is not actually using.”
- The tests were important because this migration changed response shapes for list endpoints and request decoding behavior for multiple handlers.

### What worked
- `protojson` made the generated Go types compatible with the existing camelCase JSON surface.
- The server tests verified both standalone feedback/guideline endpoints and review-action artifact creation through generated request bodies.

### What didn't work
- My first validation pass with `buf lint` failed because Buf expected the package path to be relative to the module root, not `proto/...` directly.

Command:

```bash
cd smailnail
buf lint
```

Error:

```text
proto/smailnail/annotationui/v1/review.proto:3:1:Files with package "smailnail.annotationui.v1" must be within a directory "smailnail/annotationui/v1" relative to root but were in directory "proto/smailnail/annotationui/v1".
```

Fix:
- changed `buf.yaml` to use a v2 module rooted at `proto`:

```yaml
version: v2
modules:
  - path: proto
    name: buf.build/local/smailnail
```

After that, `buf lint` passed.

### What I learned
- If the repo wants a `proto/` directory but also wants Buf’s package/path rules to stay happy, the module path has to be rooted there.
- Wrapper list responses (`items`) are the easiest way to keep backend protojson encoding simple.

### What was tricky to build
- The trickiest part was choosing where to stop: generated contract at the handler boundary, hand-written domain structs below that.
- That split avoids dragging protobuf details into repository code.

### What warrants a second pair of eyes
- The decision to drop `bodyMarkdown` from `UpdateFeedbackRequest` in the shared schema rather than trying to add backend editing support in this ticket.

### What should be done in the future
- If feedback body editing becomes desirable, add it to the repository/domain layer first and then regenerate the shared contract.

### Code review instructions
- Review these files together:
  - `pkg/annotationui/protojson.go`
  - `pkg/annotationui/contracts_review.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/handlers_annotations.go`
  - `pkg/annotationui/server_test.go`

### Technical details
- Validation commands from this step:

```bash
cd smailnail
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1
buf lint
```

## Step 3: Migrate the frontend type layer, RTK Query, mocks, and stories

After the backend was using generated contracts, I migrated the frontend. The main idea was to let generated interfaces own the shared field layout while keeping a small hand-written compatibility layer for query filters and UI-oriented string unions. That kept the UI ergonomic without reintroducing duplicated request/response shapes.

This step also fixed the original payload drift directly: create-feedback now uses `targets`, and list endpoints now unwrap wrapper responses with `transformResponse`. I also fixed Storybook guideline endpoint drift while I was already touching the same contract surface.

**Commit (code + docs):** `AnnotationUI: add shared feedback contract codegen`

### What I did
- Rewrote:
  - `ui/src/types/reviewFeedback.ts`
  - `ui/src/types/reviewGuideline.ts`
  - `ui/src/api/annotations.ts`
  - `ui/src/mocks/handlers.ts`
  - `ui/src/pages/stories/GuidelinesListPage.stories.tsx`
  - `ui/src/pages/stories/GuidelineDetailPage.stories.tsx`
- Switched list endpoints to wrapper-response unwrapping.
- Switched create-feedback mock handling from `targetIds` to `targets`.
- Made mutable mock feedback/guideline collections persist create/update operations.
- Created the focused git commit after full validation and let the repo pre-commit hook rerun `go test ./...` and `golangci-lint`.

### Why
- The frontend is where the original drift was easiest to see.
- Migrating the wrappers, mocks, and stories together prevented “types changed but dev environment still lies” problems.

### What worked
- The generated TS interfaces fit well once I narrowed certain workflow/status fields back into local literal-union helper types for UI ergonomics.
- `pnpm run check` passed after the migration.

### What didn't work
- The first TS pass failed because generated contract fields are plain strings, while some UI components were typed against narrower unions (`FeedbackKind`, `GuidelineScopeKind`, etc.).

That showed up as `string is not assignable to ...` errors in badge and form components. I fixed it by making the local wrapper files use generated interfaces for shared shapes, but narrowing specific fields in the exported aliases that the UI consumes.

### What I learned
- There is a useful middle ground between “all handwritten types” and “raw generated types everywhere.”
- Generated interfaces can own the contract while small local wrappers preserve UI-friendly typing where needed.

### What was tricky to build
- The trickiest part was keeping the frontend compile clean without giving up on the generated source of truth.
- The second trickiest part was not forgetting mocks and stories; they are part of the contract surface too.

### What warrants a second pair of eyes
- Whether the project wants to keep the local narrowed wrapper types long-term, or eventually move fully to generated string fields in UI components.

### What should be done in the future
- If more endpoints are converted, create a small contract-conventions note for when to use raw generated types vs narrowed wrapper exports.

### Code review instructions
- Start with:
  - `ui/src/gen/smailnail/annotationui/v1/review.ts`
  - `ui/src/types/reviewFeedback.ts`
  - `ui/src/types/reviewGuideline.ts`
  - `ui/src/api/annotations.ts`
  - `ui/src/mocks/handlers.ts`

### Technical details
- Validation command:

```bash
cd smailnail/ui
pnpm run check
```

## Step 4: Extend the generated contract to the rest of the annotation UI wire layer and write the reusable playbook

After finishing the review slice, I extended the same pattern to the broader annotation UI surface: annotations, groups, logs, runs, senders, the `/api/info` response, and the query workbench endpoints. I also wrote a Glazed help-style playbook in `pkg/doc` so the next contract migration does not have to reverse-engineer the workflow from this ticket.

The most important design choice in this phase was to keep the backend/domain split intact even while broadening the generated contract. Repository and SQL-facing structs stay handwritten. Generated protobuf messages now own the HTTP payloads, and mapper helpers translate between the two sides.

### What I did
- Added `proto/smailnail/annotationui/v1/annotation.proto`.
- Regenerated:
  - `pkg/gen/smailnail/annotationui/v1/annotation.pb.go`
  - `ui/src/gen/smailnail/annotationui/v1/annotation.ts`
  - `ui/src/gen/google/protobuf/struct.ts`
- Added backend mapper helpers in `pkg/annotationui/contracts_annotation.go`.
- Migrated backend handlers:
  - `pkg/annotationui/handlers_annotations.go`
  - `pkg/annotationui/handlers_senders.go`
  - `pkg/annotationui/handlers_query.go`
  - `pkg/annotationui/server.go`
- Migrated frontend annotation wrapper types in `ui/src/types/annotations.ts`.
- Updated `ui/src/api/annotations.ts` to unwrap generated wrapper list responses for annotations, groups, logs, runs, senders, presets, and saved queries.
- Updated `ui/src/mocks/handlers.ts` so MSW serves the same generated-contract shapes.
- Added `pkg/doc/annotationui-contract-codegen-playbook.md` using the Glazed help writing guidelines.
- Extended backend tests in `pkg/annotationui/server_test.go` to decode the broader generated contract via `protojson`.

### Why
- The first phase solved the most obvious review-feedback drift, but the rest of the annotation UI still depended on handwritten parallel DTOs.
- Extending the generated contract gives the annotation UI one consistent wire-contract story rather than one generated island surrounded by handwritten JSON.
- The playbook makes the workflow repeatable and reviewable for the next slice.

### What worked
- The broader contract migration fit the same pattern as the review slice: generated protobuf messages at the HTTP edge, handwritten domain structs below.
- Wrapper list responses with `items` kept frontend RTK Query integration straightforward.
- `google.protobuf.Struct` worked well for query result rows while preserving the JSON shape as an array of plain objects.
- The generated TypeScript output for query rows (`{ [key: string]: any }[]`) was good enough to wrap into the existing `Record<string, unknown>[]` UI type.

### What didn't work
- Nothing fundamentally blocked this phase, but it reinforced that query-result payloads are the oddest part of the contract because they carry dynamic row maps instead of stable fixed fields.
- That dynamic shape required an explicit backend conversion step through `structpb.NewStruct`.

### What I learned
- Once the protojson + mapper pattern is in place, extending the contract to adjacent endpoints is much more mechanical and much less risky.
- Dynamic JSON rows are still manageable in the shared contract, but they deserve to stay isolated to the query-specific messages.

### What was tricky to build
- The tricky part was migrating list endpoints and mocks together so the app, tests, and Storybook all agreed on wrapper responses.
- The second tricky part was broadening the ticket without letting the backend start returning raw repository structs again in “just one more” handler.

### What warrants a second pair of eyes
- Whether the project wants every list endpoint in this subsystem to stay on wrapper responses permanently, or whether some older endpoints should eventually be versioned if external consumers exist.
- Whether future query-result evolution should keep `Struct` rows or introduce stronger typed result envelopes for known workbench queries.

### What should be done in the future
- Split future schema additions into more focused proto files whenever a new slice becomes large enough to deserve its own review boundary.
- Add pagination-specific wrapper metadata if any list endpoint grows beyond the current simple `items` shape.

### Code review instructions
- Start with:
  - `proto/smailnail/annotationui/v1/annotation.proto`
  - `pkg/annotationui/contracts_annotation.go`
  - `pkg/annotationui/handlers_annotations.go`
  - `pkg/annotationui/handlers_senders.go`
  - `pkg/annotationui/handlers_query.go`
  - `ui/src/types/annotations.ts`
  - `ui/src/api/annotations.ts`
  - `pkg/doc/annotationui-contract-codegen-playbook.md`

### Technical details
- Validation commands from this step:

```bash
cd smailnail
buf lint
go generate ./pkg/annotationui
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1

cd ui
pnpm run check
```

## Related

- Design doc: `../design-doc/01-implementation-plan-for-shared-feedback-and-guideline-contract-codegen.md`
