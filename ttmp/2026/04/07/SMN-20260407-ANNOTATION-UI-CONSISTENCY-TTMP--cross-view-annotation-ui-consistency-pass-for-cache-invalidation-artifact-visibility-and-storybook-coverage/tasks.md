# Tasks

## Done

- [x] Create a new ticket under `smailnail/ttmp` for the annotation UI consistency pass
- [x] Write a detailed analysis/design/implementation guide for a new intern
- [x] Write a chronological investigation diary for the ticket
- [x] Identify the key runtime, query, page, and Storybook seams involved in the consistency problem

## Planned Implementation Phases

### Phase 1 — Inventory the current artifact/query/invalidation matrix

- [x] Enumerate every routed annotation page and the artifact sections each one displays
- [x] Map each visible artifact section to its exact backend query / RTK Query hook
- [x] Map each review/guideline mutation to the queries and views it should refresh
- [x] Record the matrix in the ticket docs as the canonical source of truth
- [x] Add a dedicated reference doc for the matrix so later implementation commits can update one stable source of truth
- [x] Commit phase 1 as a focused documentation / planning checkpoint

### Phase 2 — Add backend support for missing artifact surfaces

- [x] Extend `annotate.ListFeedbackFilter` with target-addressable fields for feedback lookup
- [x] Update `Repository.ListReviewFeedback(...)` to join through `review_feedback_targets` when target filters are present
- [x] Thread the new filter fields through `pkg/annotationui/handlers_feedback.go`
- [x] Add focused repository and/or handler coverage proving annotation-targeted feedback can be listed correctly
- [x] Add an explicit sender-guideline read model and HTTP endpoint for sender-visible linked guidelines
- [x] Extend the annotation UI protobuf contract and generated Go/TS outputs for the sender-guideline response
- [x] Add focused backend coverage for the sender-guideline endpoint
- [x] Validate phase 2 (`buf lint`, `go generate ./pkg/annotationui`, `go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1`)
- [x] Commit phase 2 as a focused backend contract/read-model change

### Phase 3 — Align frontend view composition with artifact ownership

- [x] Expose the new feedback target filters and sender-guideline query in `ui/src/api/annotations.ts`
- [x] Add UI types/wrappers for the sender-guideline response shape
- [x] Update `SenderDetailPage` to load sender-visible guidelines explicitly
- [x] Update annotation detail rendering so annotation-scoped feedback is visible in the environments where those annotations are shown
- [x] Revisit `RunDetailPage`, `ReviewQueuePage`, and `GuidelineDetailPage` to confirm their current artifact composition still matches the matrix after the new sender work lands
- [x] Document any intentional link-out behavior where a page points to a richer artifact view instead of rendering inline
- [x] Validate phase 3 (`cd ui && pnpm run check`)
- [x] Commit phase 3 as a focused frontend artifact-visibility change

### Phase 4 — Tighten cache invalidation and query-tag ownership

- [x] Audit every query for `providesTags` and every mutation for `invalidatesTags` against the matrix doc
- [x] Close any remaining gaps where a mounted detail view can stay stale after a successful mutation
- [x] Decide and document whether this ticket stops at broad family tags or migrates selected high-risk views to entity-scoped tags
- [x] Add regression coverage or explicit Storybook notes for every stale-view bug fixed in this phase
- [x] Validate phase 4 (`cd ui && pnpm run check` and any focused backend/frontend tests needed for cache ownership changes)
- [x] Commit phase 4 as a focused cache-coordination fix

### Phase 5 — Make Storybook/MSW prove the same invariants as production

- [x] Replace static annotation mock behavior with mutable shared annotation state in MSW
- [x] Make run-detail MSW handlers recompute from that mutable state rather than frozen `mockAnnotations`
- [x] Make sender-detail MSW handlers reflect annotation feedback and sender-visible guideline relationships
- [x] Add sender-detail stories that include visible annotation feedback and linked-guideline scenarios
- [x] Add run-detail and review-queue stories that demonstrate mutation-driven refresh expectations
- [x] Ensure Storybook handlers use the same response envelopes and relationship semantics as the live backend
- [x] Validate phase 5 (`cd ui && pnpm run check` plus manual Storybook review)
- [x] Commit phase 5 as a focused Storybook/MSW consistency pass

### Phase 6 — Validation, handoff, and durable playbook capture

- [x] Run the focused backend test suite for repository/handler additions
- [x] Run frontend typecheck and manually review the key Storybook scenarios
- [x] Update the design doc, matrix doc, changelog, and diary with implementation outcomes and validation evidence
- [x] Relate any newly added code files to the ticket docs with `docmgr doc relate`
- [x] Run `docmgr doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30`
- [x] Promote the final artifact-visibility/query-invalidation policy into a durable repo playbook/help entry if the implementation settles cleanly
- [x] Upload an updated ticket bundle to reMarkable
- [x] Commit phase 6 as the ticket hygiene / handoff checkpoint

## Post-rollout follow-up

- [ ] Tighten the expanded annotation detail row in the review queue by removing the sender-link row and rendering linked guidelines as inline chips inside the compact review-feedback summary, with the actual feedback text on the following line
- [ ] Decide whether the compact feedback line should reuse `FeedbackCard` with a dense mode or introduce a smaller expanded-row-specific summary renderer
- [ ] Update the relevant queue/run/sender Storybook stories to demonstrate the compact expanded-row layout once implemented

## Delivery / Documentation

- [x] Relate key files to the ticket docs with `docmgr doc relate`
- [x] Run `docmgr doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30`
- [x] Upload the ticket bundle to reMarkable
