# Changelog

## 2026-04-06

- Created ticket `SMN-20260406-CONTRACT-CODEGEN` under `smailnail/ttmp`
- Wrote the implementation plan for shared feedback/guideline contract codegen
- Added a shared protobuf schema for the annotation review feedback/guideline wire layer
- Added `buf.yaml` / `buf.gen.yaml` and a repo-local Go generator command
- Wired generation into `go generate` via `pkg/annotationui/generate.go`
- Generated Go contracts into `pkg/gen/...` and TS contracts into `ui/src/gen/...`
- Migrated backend feedback/guideline handlers and review-action request decoding to generated contract types
- Migrated frontend type wrappers, RTK Query contracts, mocks, and stories to the generated contract
- Added backend tests for generated-contract request/response flows
- Validated generation, Go tests, UI typecheck, and Buf lint
- Created focused commit `AnnotationUI: add shared feedback contract codegen`
- Extended shared codegen from the review slice to the rest of the annotation UI wire layer
- Added `proto/smailnail/annotationui/v1/annotation.proto` for annotations, groups, logs, runs, senders, info, and query payloads
- Migrated backend list/detail/query handlers to generated annotation contract types + `protojson`
- Migrated frontend annotation types, RTK Query list unwrapping, and MSW mocks to the generated annotation contract
- Added `pkg/doc/annotationui-contract-codegen-playbook.md` as the operator/developer playbook for future protobuf contract additions
- Added a repo-wide wire-contract unification specification covering all frontend/backend DTO surfaces and the hosted web API rollout plan
- Updated the playbook with repo-wide response-envelope guidance (`items` vs `data` + `meta`) and hosted-API migration conventions
