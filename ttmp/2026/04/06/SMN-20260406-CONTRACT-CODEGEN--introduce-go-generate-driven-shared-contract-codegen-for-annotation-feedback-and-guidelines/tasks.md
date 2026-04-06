# Tasks

## Done

- [x] Add shared proto schema and go-generate plumbing for review feedback/guidelines
- [x] Generate Go and TypeScript contract code and commit generated outputs
- [x] Migrate backend annotationui handlers to generated review contract types
- [x] Migrate frontend RTK Query types, wrappers, mocks, and stories to generated review contract types
- [x] Validate, update diary/docs, and create a focused git commit
- [x] Extend shared codegen to the rest of the annotation UI wire layer (annotations, groups, logs, runs, senders, query endpoints)
- [x] Add a Glazed help-style playbook in `pkg/doc/` for future contract-codegen work
- [x] Write the repo-wide wire-contract unification spec for all frontend/backend DTOs
- [x] Update the contract-codegen playbook with repo-wide conventions and hosted-API envelope guidance
- [x] Introduce shared protobuf contracts for the hosted web API (`/api/info`, `/api/me`, `/api/accounts/*`, `/api/rules/*`)
- [x] Migrate backend `pkg/smailnaild/http.go` to generated request/response contracts and `protojson`
- [x] Migrate frontend `ui/src/api/client.ts` and `ui/src/api/types.ts` to generated hosted-API contracts
- [x] Update hosted API tests to validate generated wire shapes
- [x] Validate, diary, and create a focused follow-up commit for the hosted API slice
