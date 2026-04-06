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
