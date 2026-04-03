# Tasks

## Current Phase

- [ ] Review the detailed guide in `design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md`
- [ ] Confirm whether mailbox should be represented as `mailbox`, `mailboxName`, or both at the UI/API boundary

## Schema And Repository

- [ ] Add a schema migration for reviewer feedback records
- [ ] Add a schema migration for reviewer feedback target links
- [ ] Add a schema migration for review guidelines
- [ ] Add a schema migration for run-to-guideline links
- [ ] Decide whether guideline versioning is required in the first iteration or whether `updated_at` plus append-only changelog is sufficient
- [ ] Extend Go types in `pkg/annotate/types.go` for review feedback, guideline records, and run-guideline links
- [ ] Extend `pkg/annotate/repository.go` with CRUD and listing methods for reviewer feedback
- [ ] Extend `pkg/annotate/repository.go` with CRUD and listing methods for review guidelines
- [ ] Add repository helpers for linking and unlinking guidelines to agent runs
- [ ] Add repository filters for mailbox-aware queries where appropriate

## Backend API

- [ ] Add review-feedback HTTP endpoints under the sqlite annotation server
- [ ] Extend single-item review mutation to accept optional reviewer comment/guideline references
- [ ] Extend batch review mutation to accept optional reviewer comment/guideline references
- [ ] Add guideline CRUD endpoints
- [ ] Add run-guideline linking endpoints
- [ ] Add response fields for mailbox-aware message previews
- [ ] Audit existing sender/run/detail endpoints and include mailbox context where it materially affects review decisions
- [ ] Add tests covering create, edit, link, filter, and batch-review-with-comment flows

## Frontend UX

- [ ] Add a reviewer text entry flow to the review queue
- [ ] Add a reviewer text entry flow to the run detail page
- [ ] Support review comments for one selected annotation
- [ ] Support review comments for multi-select batch actions
- [ ] Support a "reject and explain why" path that does not require leaving the review context
- [ ] Add a guidelines management surface in the annotation UI
- [ ] Add a run detail affordance for linked guidelines
- [ ] Display mailbox context in relevant review tables and detail cards
- [ ] Add loading, empty, and error states for the new review/guideline features

## Import And Provenance

- [ ] Trace every import/mirror/enrichment path that constructs message-facing DTOs and confirm mailbox preservation
- [ ] Ensure imported annotations or review artifacts can carry mailbox context when the source knows it
- [ ] Document mailbox provenance rules so future imports stay consistent

## Validation

- [ ] Add repository tests for new schema and methods
- [ ] Add handler tests for the new HTTP routes
- [ ] Add frontend tests for review-comment entry and guideline editing
- [ ] Add an end-to-end manual validation playbook for reviewer-to-agent revision loops

## Documentation

- [ ] Keep the ticket diary updated as implementation progresses
- [ ] Relate the most relevant code files to the focused design doc with `docmgr doc relate`
- [ ] Run `docmgr doctor --ticket SMN-20260403-RUN-REVIEW`
- [ ] Upload the document bundle to reMarkable after the guide is finalized
