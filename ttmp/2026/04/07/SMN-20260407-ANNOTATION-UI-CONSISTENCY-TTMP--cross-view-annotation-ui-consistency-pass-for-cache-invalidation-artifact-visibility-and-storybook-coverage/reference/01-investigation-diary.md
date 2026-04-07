---
Title: Investigation diary
Ticket: SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
    - workflow
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotate/repository_feedback.go
      Note: Diary records the backend sender-artifact read-model implementation and its test fix
    - Path: pkg/annotationui/handlers_senders.go
      Note: Diary context for current sender detail payload composition
    - Path: pkg/doc/annotationui-review-consistency-playbook.md
      Note: Diary records the durable repo playbook added at handoff time
    - Path: ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage/design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md
      Note: Primary deliverable recorded in the diary
    - Path: ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage/reference/02-artifact-query-and-invalidation-matrix.md
      Note: Canonical page-by-page matrix added during phase-1 planning
    - Path: ui/src/api/annotations.ts
      Note: Diary context for query/tag observations gathered during the investigation
    - Path: ui/src/components/AnnotationTable/AnnotationDetail.tsx
      Note: Diary Step 12 records inline rendering of run-linked guidelines beside annotation feedback
    - Path: ui/src/components/AnnotationTable/AnnotationTable.tsx
      Note: Diary Step 12 records the new guideline slot for shared expanded annotation detail
    - Path: ui/src/mocks/handlers.ts
      Note: |-
        Diary context for Storybook/MSW mutation-state limitations
        Diary records the move to mutable Storybook annotation state
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Diary Step 12 records restoring the per-row dismiss-and-explain action in the review queue
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Diary context for the run-detail composed-query baseline
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: |-
        Diary context for the sender-detail artifact visibility gap
        Diary records explicit sender artifact queries and the expandedId hook ordering fix
    - Path: ui/src/pages/stories/ReviewQueuePage.stories.tsx
      Note: Diary Step 12 records the queue story copy update for the restored dismiss-and-explain action
    - Path: ui/src/pages/stories/SenderDetailPage.stories.tsx
      Note: Diary Step 12 records sender story guidance for inline linked-guideline visibility
ExternalSources: []
Summary: Chronological diary for creating the annotation UI consistency ticket, gathering evidence across backend/frontend/storybook layers, and producing the design/implementation guide plus reMarkable delivery.
LastUpdated: 2026-04-07T12:35:00-04:00
WhatFor: Capture the exact investigation path, commands, reasoning, and delivery steps for the annotation UI consistency pass.
WhenToUse: Read this when continuing the ticket, reviewing how the analysis was assembled, or reproducing the upload/validation workflow.
---




# Diary

## Goal

This diary records how I created the new annotation UI consistency ticket, gathered evidence about cache invalidation and artifact visibility, and wrote the design/implementation guide for a future broad cleanup pass. The intended reader is a new intern or follow-up engineer who needs both repository orientation and the exact command trail used to assemble the deliverable.

## Step 1: Create a dedicated ticket for the consistency pass

The first task was to turn the user’s emerging concern into a properly tracked workstream. The issue was no longer a single bug about one page failing to refresh; it had become a cross-view consistency problem spanning routes, backend queries, frontend caching, and Storybook truthfulness.

I therefore created a new ticket instead of extending the earlier intern-review ticket again. That keeps the new work legible as a dedicated architecture/implementation pass rather than as another small follow-up finding.

### Prompt Context

**User prompt (verbatim):** "Create reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new docmgr ticket under `smailnail/ttmp`, produce a thorough intern-facing architecture/design/implementation guide about annotation UI consistency, keep a diary, and upload the resulting ticket bundle to reMarkable.

**Inferred user intent:** Convert the observed refresh/visibility bugs into a well-scoped, durable work package that another engineer can execute confidently.

### What I did
- Loaded the `ticket-research-docmgr-remarkable`, `docmgr`, `diary`, and `remarkable-upload` skills and references.
- Created a fresh ticket under `smailnail/ttmp`:
  - `SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP`
- Added:
  - a primary design doc,
  - an investigation diary.

### Why
- The user asked for a dedicated ticket and a full implementation guide, not just an informal summary in chat.
- The issue crosses enough subsystems that it deserves its own index/tasks/changelog and later implementation diary.

### What worked
- `docmgr --root "$PWD/ttmp" ticket create-ticket ...` created the expected workspace and standard files.
- Adding the design doc and diary up front made it easy to structure the later analysis.

### What didn't work
- A quick `find ttmp/2026/04/07` check before creation naturally failed because the date directory did not exist yet:

```text
find: ‘ttmp/2026/04/07’: No such file or directory
```

This was not a real blocker; it just confirmed I was creating a brand-new workspace rather than updating an existing one.

### What I learned
- The consistency pass is broad enough that it should not be hidden inside the earlier review-finding ticket history.

### What was tricky to build
- The only subtle part here was choosing the right scope: “cache invalidation bug” was too narrow, but “redesign the annotation system” would have been too broad. The ticket needed to explicitly include invalidation, fetching, artifact visibility, and Storybook coverage as one coherent pass.

### What warrants a second pair of eyes
- The eventual implementation should confirm whether sender-visible guidelines are intended to be grouped by run or summarized more minimally. That product/UI choice affects the backend query design.

### What should be done in the future
- Keep follow-up implementation work inside this ticket so the analysis and execution history stay together.

### Code review instructions
- Start at the ticket root:
  - `ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage`
- Then read:
  - `design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`
  - `reference/01-investigation-diary.md`

### Technical details
- Ticket creation commands:

```bash
cd smailnail
ROOT="$PWD/ttmp"
TICKET=SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP

docmgr --root "$ROOT" ticket create-ticket \
  --ticket "$TICKET" \
  --title "Cross-view annotation UI consistency pass for cache invalidation, artifact visibility, and Storybook coverage" \
  --topics annotations,backend,frontend,sqlite,workflow

docmgr --root "$ROOT" doc add --ticket "$TICKET" \
  --doc-type design-doc \
  --title "Analysis and implementation guide for annotation UI consistency and artifact visibility"

docmgr --root "$ROOT" doc add --ticket "$TICKET" \
  --doc-type reference \
  --title "Investigation diary"
```

## Step 2: Reconstruct the current annotation UI architecture before proposing fixes

Before writing recommendations, I needed to re-map how the sqlite review UI is wired today. The user wanted a document that would work for a new intern, so I treated this as an orientation exercise first and a critique second.

I focused on the runtime boundaries, the registered routes, the RTK Query slice, the key pages where the user had observed stale or missing state, and the Storybook/MSW layer that should eventually illustrate these relationships. I also pulled line-numbered snippets so the design doc could cite concrete evidence instead of relying on hand-wavy recollection.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build an evidence-backed architecture map of the annotation subsystem before recommending a broad consistency pass.

**Inferred user intent:** Make the resulting design doc teach the system clearly, not merely complain about symptoms.

### What I did
- Read and line-anchored the sqlite server entrypoint:
  - `cmd/smailnail/commands/sqlite/serve.go`
- Read and line-anchored route registration:
  - `pkg/annotationui/server.go`
- Read and line-anchored the frontend router:
  - `ui/src/App.tsx`
- Read and line-anchored the RTK Query slice:
  - `ui/src/api/annotations.ts`
- Read the core page implementations most relevant to the reported problems:
  - `ui/src/pages/RunDetailPage.tsx`
  - `ui/src/pages/SenderDetailPage.tsx`
  - `ui/src/pages/GuidelineDetailPage.tsx`
  - `ui/src/pages/ReviewQueuePage.tsx`
- Read the sender handler and feedback handlers:
  - `pkg/annotationui/handlers_senders.go`
  - `pkg/annotationui/handlers_feedback.go`
- Read the review repository methods and filter types:
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotate/types.go`
- Read the protobuf contracts and Storybook/MSW files:
  - `proto/smailnail/annotationui/v1/annotation.proto`
  - `proto/smailnail/annotationui/v1/review.proto`
  - `ui/src/mocks/handlers.ts`
  - `ui/src/pages/stories/RunDetailPage.stories.tsx`
  - `ui/src/pages/stories/SenderDetailPage.stories.tsx`

### Why
- The analysis needed to explain not just what was broken, but how data is supposed to move from repository to handler to wire contract to RTK Query hook to page section.
- The architectural asymmetry between run detail and sender detail was one of the main insights, and it only becomes obvious after reading both pages and their upstream handlers together.

### What worked
- The repo is structured cleanly enough that the main consistency seams are visible with a handful of entrypoint and page files.
- The combination of backend handler reading and frontend query reading was enough to explain why some artifacts persist but remain invisible.

### What didn't work
- There was no single “artifact visibility matrix” in the repo to consult. I had to infer the current behavior by comparing backend data availability against what the pages actually query and render.

### What I learned
- The system is already half-way to a strong architecture: run detail and guideline detail are composed from explicit subqueries, which is a good pattern.
- The main consistency debt is that other views—especially sender detail—have not yet been brought up to the same standard.

### What was tricky to build
- The tricky part was distinguishing three different failure modes:
  1. data not persisted,
  2. data persisted but not fetched,
  3. data fetched but not invalidated/refreshed correctly.

The user-reported issues covered all three at different moments of the earlier conversation, so it was important not to blur them together.

### What warrants a second pair of eyes
- The desired UX for sender-visible guidelines needs confirmation before implementation. There is a genuine product decision hidden inside that technical question.

### What should be done in the future
- Preserve the line-anchored evidence style for later implementation docs so follow-up changes can be reviewed against the same architecture map.

### Code review instructions
- To reproduce the architecture map, read in this order:
  1. `cmd/smailnail/commands/sqlite/serve.go`
  2. `pkg/annotationui/server.go`
  3. `ui/src/App.tsx`
  4. `ui/src/api/annotations.ts`
  5. `ui/src/pages/RunDetailPage.tsx`
  6. `ui/src/pages/SenderDetailPage.tsx`

### Technical details
- Core evidence-gathering commands:

```bash
cd smailnail
nl -ba cmd/smailnail/commands/sqlite/serve.go | sed -n '1,170p'
nl -ba pkg/annotationui/server.go | sed -n '1,240p'
nl -ba ui/src/App.tsx | sed -n '150,210p'
nl -ba ui/src/api/annotations.ts | sed -n '1,260p'
nl -ba ui/src/pages/RunDetailPage.tsx | sed -n '1,220p'
nl -ba ui/src/pages/SenderDetailPage.tsx | sed -n '1,230p'
nl -ba pkg/annotationui/handlers_senders.go | sed -n '90,240p'
nl -ba pkg/annotationui/handlers_feedback.go | sed -n '1,280p'
nl -ba pkg/annotate/repository_feedback.go | sed -n '340,560p'
nl -ba ui/src/mocks/handlers.ts | sed -n '1,420p'
```

## Step 3: Synthesize the design and implementation guide

Once the evidence was collected, I translated it into a long-form design doc aimed at a new intern. I intentionally structured the doc around runtime topology, page semantics, artifact visibility, query invalidation, and Storybook truthfulness so it would explain both the current system and the rationale for the proposed pass.

The core recommendation was to adopt an explicit “view-model query layer” rule: each page should declare which artifacts it shows, which query powers each artifact section, and which mutations must invalidate those sections. That framing made it possible to discuss sender detail, run detail, and guideline detail using one consistent mental model.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Turn the architecture map into a detailed, implementation-focused guide with pseudocode, diagrams, references, and phased execution guidance.

**Inferred user intent:** Leave behind a document that a new engineer can actually execute from, not just admire.

### What I did
- Wrote the primary design doc with:
  - executive summary,
  - current-state architecture,
  - gap analysis,
  - proposed solution,
  - API sketches,
  - pseudocode,
  - phased implementation plan,
  - validation strategy,
  - risks/open questions,
  - references.
- Updated the design-doc frontmatter summary/use fields.
- Used file-backed references throughout rather than speculative recommendations.

### Why
- The user explicitly asked for a guide with prose, bullet points, pseudocode, diagrams, API references, and file references.
- This kind of ticket needs a design doc that is readable by someone who was not present for the debugging conversation.

### What worked
- The route/query/page asymmetry was the best organizing principle for the document.
- The missing target-scoped feedback filter turned out to be a particularly clear architectural hinge for explaining why sender pages cannot yet display annotation feedback cleanly.

### What didn't work
- The repo does not yet have a stable answer for sender-guideline UX, so the design doc had to surface that as an open question instead of pretending it was already settled.

### What I learned
- A large class of these bugs become easier to explain when described as “persistence semantics are richer than page semantics.” That framing connects backend transactional behavior and frontend visibility gaps cleanly.

### What was tricky to build
- The difficult part was balancing near-term practicality against a tempting larger refactor. It would be easy to prescribe a grand generic artifact framework, but the repo is not ready for that. The design doc therefore recommends a phased pass with explicit view queries first and finer-grained tags later.

### What warrants a second pair of eyes
- The proposed sender-guideline endpoint/contract shape should be reviewed by whoever owns the intended sender-page UX.
- The decision about when to introduce entity-scoped RTK Query tags versus broad tags should be reviewed by whoever maintains the frontend state conventions.

### What should be done in the future
- When implementation starts, keep the artifact visibility matrix current inside the ticket and later promote it into a repo-level playbook/help entry.

### Code review instructions
- Review the design doc sections in this order:
  1. Executive Summary
  2. Current-State Architecture
  3. Gap Analysis
  4. Proposed Solution
  5. Phased Implementation Plan

### Technical details
- Main authored document:

```text
ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage/design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md
```

## Step 4: Update ticket bookkeeping, validate docs, and prepare reMarkable delivery

After the writing pass, I updated the ticket scaffolding so the workspace itself became a useful handoff artifact instead of a folder containing only one strong design doc. That meant filling in the ticket index, replacing the placeholder tasks with a phased plan, and recording the work in the changelog and diary.

I then related the key code files to the ticket docs and ran docmgr validation before preparing the upload bundle.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the ticket like a durable research deliverable: clean bookkeeping, clear task phases, and docmgr hygiene before upload.

**Inferred user intent:** Make the ticket navigable and trustworthy for the next engineer.

### What I did
- Updated:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - this diary
- Related the key runtime/query/page/story files to the design doc and diary.
- Ran `docmgr doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30`.

### Why
- A design doc alone is not enough; the ticket also needs a clear execution checklist and discoverable related files.

### What worked
- The ticket structure maps naturally onto the intended handoff: design doc for strategy, tasks for phases, diary for provenance.
- `docmgr doctor` is a good last-mile check that the metadata and relations are sane.

### What didn't work
- No major blockage here. The main work was simply replacing the generated placeholders with real content.

### What I learned
- The generated ticket scaffolding is intentionally light. It needs meaningful manual curation before it becomes useful for a larger engineering pass.

### What was tricky to build
- The only tricky part was choosing task phases that are implementation-ready without prematurely deciding every API detail. I used “contract alignment → backend support → frontend composition → Storybook truthfulness → validation” as the main sequence.

### What warrants a second pair of eyes
- Future execution should check whether the implementation is still best tracked in one ticket or whether code-heavy sub-slices should spin out to child tickets.

### What should be done in the future
- If implementation work starts, add commit-backed diary steps rather than overwriting the current analytical diary.

### Code review instructions
- Check the ticket bookkeeping files:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - `reference/01-investigation-diary.md`

### Technical details
- Planned validation command:

```bash
cd smailnail
docmgr --root "$PWD/ttmp" doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30
```

## Step 5: Upload the ticket bundle to reMarkable

The final deliverable is not just the local ticket but also a reMarkable-friendly bundle. The intended bundle should contain the index, design doc, diary, tasks, and changelog so the reader gets a navigable PDF with a table of contents instead of a fragmented upload.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Publish the ticket bundle to reMarkable after local validation, using the safe dry-run-first workflow.

**Inferred user intent:** Make the analysis easy to review away from the development machine.

### What I did
- Prepared a bundle upload plan using the ticket documents.
- Ran the standard sequence:
  1. `remarquee status`
  2. `remarquee cloud account --non-interactive`
  3. bundle upload dry-run
  4. real bundle upload
  5. remote listing verification
- Uploaded a bundled PDF containing the index, design doc, diary, tasks, and changelog.

### Why
- The upload is part of the user’s requested handoff workflow, not an optional extra.

### What worked
- Bundled upload is the right shape here because the design doc and diary are meant to be read together.
- `remarquee` was already authenticated and the dry-run/real upload flow completed without requiring any workaround.

### What didn't work
- No toolchain failure here; the upload path worked on the first attempt.

### What I learned
- For tickets like this, a single bundle with a ToC is substantially better than a handful of separate PDFs.

### What was tricky to build
- The only tricky part is making sure the bundle order matches the desired reading order. Using the index first, then design doc, then diary, then tasks/changelog, keeps the PDF coherent.

### What warrants a second pair of eyes
- If the reMarkable upload toolchain reports auth or LaTeX issues, those should be recorded exactly rather than hand-waved.

### What should be done in the future
- If this becomes a common workflow, consider a ticket-local upload helper script in `scripts/` to standardize bundle order.

### Code review instructions
- Verify that the upload contains:
  - ticket index,
  - main design doc,
  - diary,
  - tasks,
  - changelog.

### Technical details
- Remote destination:

```text
/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP
```

- Upload/verification commands:

```bash
cd smailnail
BASE="$PWD/ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage"
INDEX="$BASE/index.md"
DESIGN="$BASE/design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md"
DIARY="$BASE/reference/01-investigation-diary.md"
TASKS="$BASE/tasks.md"
CHANGELOG="$BASE/changelog.md"
REMOTE="/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP"

remarquee status
remarquee cloud account --non-interactive
remarquee upload bundle --dry-run "$INDEX" "$DESIGN" "$DIARY" "$TASKS" "$CHANGELOG" \
  --name "SMN-20260407 Annotation UI Consistency Pass" \
  --remote-dir "$REMOTE" \
  --toc-depth 2
remarquee upload bundle "$INDEX" "$DESIGN" "$DIARY" "$TASKS" "$CHANGELOG" \
  --name "SMN-20260407 Annotation UI Consistency Pass" \
  --remote-dir "$REMOTE" \
  --toc-depth 2
remarquee cloud ls "$REMOTE" --long --non-interactive
```

- Verification result:

```text
[f]    SMN-20260407 Annotation UI Consistency Pass
```

## Step 6: Turn the high-level plan into a page-by-page execution matrix

Once the first design doc was in place, the next useful step was to make the work operational. The original task list was too coarse to drive multi-commit implementation cleanly, and the design doc itself was too narrative to serve as the one place I could quickly check “which page owns which artifacts?” or “which mutation is supposed to refresh which view?”

I fixed that by adding a dedicated matrix reference document and then rewriting `tasks.md` into a more detailed phase-by-phase execution checklist. That gives the ticket both a strategic document and a day-to-day implementation control surface.

### Prompt Context

**User prompt (verbatim):** "Add detailed tasks to the ticket. Then work on each task one by one, committing at appropriate intervals, checking off tasks as you go, keeping a detailed diary"

**Assistant interpretation:** Expand the ticket into a concrete execution plan, record the current planning step in the diary, and then start implementing the phases in order with disciplined commits.

**Inferred user intent:** Turn the new ticket from a design-only artifact into a working implementation tracker that can be executed incrementally.

### What I did
- Added a new ticket reference document:
  - `reference/02-artifact-query-and-invalidation-matrix.md`
- Recorded, page by page:
  - visible artifact sections,
  - current query sources,
  - mutations that should refresh them,
  - current status,
  - missing implementation deltas.
- Rewrote `tasks.md` into a more detailed phased checklist with explicit commit and validation checkpoints.
- Updated `index.md` and `changelog.md` so the ticket now advertises the matrix doc and notes that phase 1 planning is complete in documentation form.

### Why
- The ticket needed a stable source of truth for the cross-view consistency rules before code changes begin.
- A detailed task list is necessary if the implementation is going to land in focused commits without losing track of what has already been validated.

### What worked
- The page/mutation matrix made the sender-detail gaps much easier to reason about than the narrative design doc alone.
- Turning each implementation phase into explicit validation + commit steps makes the upcoming execution loop much easier to follow.

### What didn't work
- Nothing failed technically here, but the exercise confirmed that the repo still has no existing built-in artifact visibility matrix; it had to be created from scratch.

### What I learned
- A single cross-view consistency ticket really needs two doc layers:
  1. a narrative architecture/design guide,
  2. a terse matrix/checklist for implementation.

### What was tricky to build
- The tricky part was deciding how much detail belonged in the tasks versus the matrix. I ended up putting page/query/mutation truth in the matrix and using `tasks.md` for executable slices with validation/commit boundaries.

### What warrants a second pair of eyes
- Before phase 2 code lands, it would be good to confirm that the proposed sender-visible guideline policy in the matrix matches the intended product behavior.

### What should be done in the future
- Keep the matrix current as implementation changes the page/query ownership model; do not let it become stale architecture fiction.

### Code review instructions
- Review these docs together:
  - `tasks.md`
  - `reference/02-artifact-query-and-invalidation-matrix.md`
  - `design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`

### Technical details
- Matrix doc creation command:

```bash
cd smailnail
docmgr --root "$PWD/ttmp" doc add \
  --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP \
  --doc-type reference \
  --title "Artifact query and invalidation matrix"
```

## Step 7: Commit the planning docs and matrix as the phase-1 checkpoint

Once the matrix doc and detailed tasks existed, the ticket finally had enough structure to support implementation without guesswork. That was the right moment for a focused docs-only checkpoint commit: the work had moved from “analysis exists somewhere in chat and partial docs” to “the ticket itself now describes the execution plan.”

### Prompt Context

**User prompt (verbatim):** (see Step 6)

**Assistant interpretation:** Freeze the planning layer before starting code so the later implementation commits can be reviewed against a stable task/matrix baseline.

**Inferred user intent:** Make the implementation history understandable by separating planning from code changes.

**Commit (code):** `1a57036` — `Docs: plan annotation UI consistency pass`

### What I did
- Added the matrix doc to the ticket.
- Expanded the task list into detailed execution slices.
- Updated the ticket index/changelog/diary enough to make the ticket self-contained.
- Committed the new ticket workspace and planning docs.

### Why
- This pass is too broad to execute safely without a clean planning checkpoint.

### What worked
- A docs-only checkpoint made it much easier to reason about the next backend and frontend slices independently.

### What didn't work
- N/A; this was a straightforward documentation checkpoint.

### What I learned
- The matrix doc is just as important as the narrative design doc for day-to-day implementation control.

### What was tricky to build
- The main challenge was deciding what had to be committed before code began. The answer was: enough documentation that later code reviews can ask “did this change match the matrix and tasks?”

### What warrants a second pair of eyes
- Reviewers should confirm that the matrix doc and task phases are consistent with the design doc rather than subtly drifting.

### What should be done in the future
- Keep future doc updates tightly synchronized with code commits so the ticket remains an execution tracker, not a stale plan.

### Code review instructions
- Review commit `1a57036` together with:
  - `tasks.md`
  - `reference/02-artifact-query-and-invalidation-matrix.md`
  - `design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`

### Technical details
- Commit command:

```bash
cd smailnail
git add ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage
git commit -m "Docs: plan annotation UI consistency pass"
```

## Step 8: Add backend read models for target-scoped feedback and sender-visible guidelines

The first code slice was the backend because sender-visible artifacts could not be rendered cleanly without backend support. The biggest backend gaps were exactly the ones identified in the design doc: feedback could not be listed by target, and sender detail had no endpoint for the run-linked guidelines relevant to a sender.

I treated this as one focused contract/read-model change: extend the list feedback filter, add sender-guideline wire messages plus endpoint support, and cover the new behavior with server tests.

### Prompt Context

**User prompt (verbatim):** (see Step 6)

**Assistant interpretation:** Execute phase 2 by giving the backend enough read-model support that sender detail and expanded annotation views can show the artifacts users are already creating.

**Inferred user intent:** Stop hiding persisted feedback/guideline relationships behind missing backend query surfaces.

**Commit (code):** `81f67cd` — `AnnotationUI: add sender artifact read models`

### What I did
- Extended `annotate.ListFeedbackFilter` with `TargetType` and `TargetID`.
- Updated `Repository.ListReviewFeedback(...)` to join through `review_feedback_targets` when target filters are present.
- Threaded the new query params through `pkg/annotationui/handlers_feedback.go`.
- Added `annotate.SenderGuidelineGroup` and `Repository.ListSenderGuidelineGroups(...)`.
- Added `GET /api/mirror/senders/{email}/guidelines`.
- Extended `proto/smailnail/annotationui/v1/annotation.proto` with `SenderGuidelineGroup` and `SenderGuidelineListResponse`.
- Regenerated the Go/TS contract outputs.
- Added focused handler coverage for target-filtered feedback and sender-guideline grouping in `pkg/annotationui/server_test.go`.

### Why
- The sender page needed real backend-backed artifact queries, not client-side guessing.
- Target-scoped feedback lookup is the minimum contract needed to display annotation feedback naturally.

### What worked
- The new target filters fit cleanly into the existing feedback listing path.
- Grouping sender guidelines by run gave the sender page a more comprehensible response shape than a flat list would have.
- The generated wire contract extension landed cleanly with the existing protobuf/`go generate` setup.

### What didn't work
- My first sender-guideline endpoint test assumed there would be exactly two linked guidelines, but earlier test setup had already linked another guideline to the same run. The focused test failed with:

```text
--- FAIL: TestHandlerServesAnnotationAPIAndSPA/sender_guideline_endpoint_groups_linked_guidelines_by_run (0.01s)
    server_test.go:513: expected 2 guidelines, got 4
```

- I fixed that by changing the assertion to require the two newly created guideline slugs to be present rather than assuming the run had no pre-existing links.

### What I learned
- Handler tests in this file share one seeded DB fixture across subtests, so endpoint tests should assert on the specific data they create instead of assuming isolation from earlier subtests.

### What was tricky to build
- The subtle design decision was response shape: sender-visible guidelines really do want run grouping, because the relationship is run-level even though the user is looking at a sender page.
- The second tricky part was keeping the new backend contract focused enough that it supported the page without turning sender detail into a giant bespoke DTO.

### What warrants a second pair of eyes
- Reviewers should confirm that `SenderGuidelineGroup` is the right long-term wire shape and that a future richer sender artifact endpoint is still unnecessary.

### What should be done in the future
- If the sender page later needs run stats alongside grouped guidelines, consider extending the grouped response rather than flattening it.

### Code review instructions
- Start with:
  - `pkg/annotate/types.go`
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/handlers_senders.go`
  - `proto/smailnail/annotationui/v1/annotation.proto`
  - `pkg/annotationui/server_test.go`

### Technical details
- Validation commands:

```bash
cd smailnail
gofmt -w pkg/annotate/types.go pkg/annotate/repository_feedback.go pkg/annotationui/contracts_annotation.go pkg/annotationui/handlers_feedback.go pkg/annotationui/handlers_senders.go pkg/annotationui/server.go pkg/annotationui/server_test.go
buf lint
go generate ./pkg/annotationui
go test -tags sqlite_fts5 ./pkg/annotationui ./pkg/annotate -count=1
```

## Step 9: Wire sender/run pages to the new artifact queries and surface annotation feedback inline

After the backend read models existed, the next step was to make the UI use them. The most important frontend outcome was that sender detail no longer acts as a thin wrapper around `getSender(...)`; it now explicitly loads sender-visible guideline groups and target-filtered annotation feedback.

I also reused the same annotation feedback display pattern in run detail by extending the expanded annotation detail panel instead of inventing a second, unrelated annotation feedback UI.

### Prompt Context

**User prompt (verbatim):** (see Step 6)

**Assistant interpretation:** Execute phase 3 by making the pages actually render the artifacts the backend now exposes, while keeping the rendering model understandable and reusable.

**Inferred user intent:** Let users see the feedback/guideline artifacts they created in the same environments where the affected annotations and senders are shown.

**Commit (code):** `a684383` — `AnnotationUI: show sender review artifacts`

### What I did
- Extended the TS wrapper types for:
  - `FeedbackFilter` target fields,
  - `SenderGuidelineGroup` responses.
- Added `useGetSenderGuidelinesQuery(...)`.
- Updated `SenderDetailPage` to:
  - load sender-guideline groups,
  - load annotation-scoped feedback for the currently expanded annotation,
  - render a new `SenderGuidelinePanel`.
- Updated `RunDetailPage` to show annotation-scoped feedback for the currently expanded annotation.
- Extended `AnnotationTable` / `AnnotationDetail` so expanded annotations can render attached feedback via existing `FeedbackCard` components.
- Added the new sender-guideline panel component and exported it through `components/SenderProfile`.
- Updated MSW handler filtering so target-scoped feedback queries behave correctly during story/dev use.

### Why
- Backend query support is only useful if the page explicitly asks for and renders the artifacts.
- Showing annotation feedback inside the expanded annotation detail keeps the UI model simple: feedback is attached where the annotation is already being inspected in detail.

### What worked
- Reusing `FeedbackCard` inside `AnnotationDetail` avoided inventing a second feedback rendering style.
- The sender-guideline panel gives sender pages a clear, grouped-by-run representation of linked guidelines.
- Run detail and sender detail now share the same expanded-annotation feedback pattern, which reduces future drift.

### What didn't work
- My first compile attempt failed because I introduced the new `expandedId`-dependent query in `SenderDetailPage` before the `useState` declaration. TypeScript reported:

```text
src/pages/SenderDetailPage.tsx(38,17): error TS2448: Block-scoped variable 'expandedId' used before its declaration.
src/pages/SenderDetailPage.tsx(38,17): error TS2454: Variable 'expandedId' is used before being assigned.
src/pages/SenderDetailPage.tsx(40,14): error TS2448: Block-scoped variable 'expandedId' used before its declaration.
src/pages/SenderDetailPage.tsx(40,14): error TS2454: Variable 'expandedId' is used before being assigned.
```

- I fixed that by moving the local state declarations above the new RTK Query hook calls.

### What I learned
- The expanded-annotation pattern is a good boundary for targeted artifact queries because only one annotation detail is visible at a time.
- A sender page can remain conceptually clean even with extra queries, as long as each query powers an explicit section.

### What was tricky to build
- The main sharp edge was preserving the generic `AnnotationTable` abstraction while still letting page-specific artifact queries feed the expanded row.
- The solution was to add a narrow `getFeedback(annotation)` callback rather than hard-code annotation feedback knowledge into the table itself.

### What warrants a second pair of eyes
- Reviewers should look closely at whether expanded annotation feedback belongs above or below the existing related-annotations section and whether the sender-guideline panel hierarchy is visually clear enough.

### What should be done in the future
- If annotation-scoped feedback becomes more common, consider adding counts or badges at the row level so users know feedback exists before expanding.

### Code review instructions
- Start with:
  - `ui/src/api/annotations.ts`
  - `ui/src/pages/SenderDetailPage.tsx`
  - `ui/src/pages/RunDetailPage.tsx`
  - `ui/src/components/AnnotationTable/AnnotationTable.tsx`
  - `ui/src/components/AnnotationTable/AnnotationDetail.tsx`
  - `ui/src/components/SenderProfile/SenderGuidelinePanel.tsx`

### Technical details
- Validation command:

```bash
cd smailnail/ui
pnpm run check
```

## Step 10: Make Storybook/MSW stateful enough to demonstrate refresh behavior honestly

The final code slice was about proof, not product behavior. The live app problems that motivated this ticket were largely about cross-query consistency, so Storybook needed to model mutation ripple effects rather than frozen before/after snapshots.

I updated the MSW layer to own mutable annotation state, taught it to recompute run and sender read models from that state, and extended the page stories with explicit descriptions that tell reviewers what interactions should now remain coordinated.

### Prompt Context

**User prompt (verbatim):** (see Step 6)

**Assistant interpretation:** Execute the Storybook/MSW phase so the repo has a practical demonstration environment for the same refresh and artifact-visibility invariants the live app now depends on.

**Inferred user intent:** Make the behavior easy to review and regress-test visually rather than rediscovering these bugs only against a real sqlite DB.

**Commit (code):** `571cede` — `Storybook: model annotation artifact refresh flows`

### What I did
- Introduced mutable annotation state into `ui/src/mocks/handlers.ts`.
- Updated annotation review and batch-review handlers to mutate annotations and to create feedback / guideline-link side effects that mirror the live backend more closely.
- Recomputed run detail and run list responses from mutable annotation state.
- Recomputed sender detail and sender-guideline responses from mutable annotation state.
- Added stateful story variants/descriptions for run detail, sender detail, and review queue.
- Ran `pnpm run build-storybook` to make sure the stories compile into a static Storybook build.

### Why
- Static handlers were giving false confidence: one request could look updated while follow-up queries still read from frozen fixtures.
- This ticket is fundamentally about cross-view coordination, so Storybook had to prove that coordination rather than sidestep it.

### What worked
- Mutable annotation state plus derived run/sender read models gives Storybook a much more honest representation of the production data flow.
- The build passed, which means the stories and handlers are at least syntactically and compositionally consistent.

### What didn't work
- Storybook build itself succeeded, but it emitted non-blocking warnings about large chunks and eval usage inside Storybook runtime code. Those warnings did not indicate a repo-specific failure.

### What I learned
- The most important part of Storybook truthfulness here is not fancy play functions; it is shared mutable state plus derived follow-up queries.

### What was tricky to build
- The sharp edge was balancing realism against scope. I wanted the handlers to model the important artifact side effects without turning the story layer into a second backend implementation.
- The compromise was to mirror the high-value behaviors: annotation state changes, annotation feedback creation, guideline linking, and derived run/sender reads.

### What warrants a second pair of eyes
- Reviewers should confirm that the Storybook descriptions are specific enough to guide manual interaction and that the mutable-state handlers do not drift too far from the backend semantics over time.

### What should be done in the future
- If the team wants even stronger visual regression protection, add play-function assertions for queue shrinkage and artifact panel updates.

### Code review instructions
- Start with:
  - `ui/src/mocks/handlers.ts`
  - `ui/src/pages/stories/RunDetailPage.stories.tsx`
  - `ui/src/pages/stories/SenderDetailPage.stories.tsx`
  - `ui/src/pages/stories/ReviewQueuePage.stories.tsx`

### Technical details
- Validation commands:

```bash
cd smailnail/ui
pnpm run check
pnpm run build-storybook
rm -rf storybook-static
```

## Step 11: Capture the durable playbook and update ticket bookkeeping for handoff

With the main implementation slices landed, the remaining work is to turn the ticket into a reliable handoff bundle again. I added a durable repo playbook in `pkg/doc` so future changes can reuse the same page/query/invalidation policy without reading this whole ticket first.

### Prompt Context

**User prompt (verbatim):** (see Step 6)

**Assistant interpretation:** Finish the implementation pass by capturing the enduring rules in repo docs and by updating the ticket bookkeeping with the new commits, validations, and remaining decisions.

**Inferred user intent:** Leave behind a maintainable system and a continuation-friendly paper trail, not just a stack of commits.

### What I did
- Added `pkg/doc/annotationui-review-consistency-playbook.md`.
- Verified the help page renders through:
  - `go run -tags sqlite_fts5 ./cmd/smailnail help annotationui-review-consistency-playbook`
- Updated the design doc, matrix doc, task list, changelog, index, and this diary with the landed implementation slices and commit hashes.
- Refreshed the `docmgr` file relations for the new backend/frontend/storybook/playbook files.
- Re-ran `docmgr doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30`.
- Re-uploaded the ticket bundle to reMarkable with the implementation updates included.

### Why
- The ticket design doc is excellent for one deep dive, but the repo also needs a shorter enduring playbook for future incremental changes.

### What worked
- The help page loaded cleanly through the existing Glazed help path, which means the policy is now discoverable outside the ticket.
- `docmgr doctor` passed cleanly after the new relations were added.
- The reMarkable bundle refresh upload succeeded with `--force`, so the device/cloud copy now matches the implemented state instead of the earlier design-only version.

### What didn't work
- N/A; the help page rendered successfully on the first pass.

### What I learned
- A repo playbook is the right landing zone for the “broad family tags for now, explicit artifact ownership always” policy.

### What was tricky to build
- The subtle part was choosing the right level of abstraction: the playbook needed to be durable enough for future use without simply duplicating the entire ticket design doc.

### What warrants a second pair of eyes
- Reviewers should confirm that the playbook says enough about sender-visible guidelines and annotation feedback without over-committing the repo to one future UI presentation forever.

### What should be done in the future
- If the repo later adopts entity-scoped tags broadly, update the playbook and the matrix doc together.

### Code review instructions
- Start with:
  - `pkg/doc/annotationui-review-consistency-playbook.md`
  - then compare it against `reference/02-artifact-query-and-invalidation-matrix.md`

### Technical details
- Help verification command:

```bash
cd smailnail
go run -tags sqlite_fts5 ./cmd/smailnail help annotationui-review-consistency-playbook
```

- Final bookkeeping / upload commands:

```bash
cd smailnail
docmgr --root "$PWD/ttmp" doctor --ticket SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP --stale-after 30

BASE="$PWD/ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage"
INDEX="$BASE/index.md"
DESIGN="$BASE/design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md"
MATRIX="$BASE/reference/02-artifact-query-and-invalidation-matrix.md"
DIARY="$BASE/reference/01-investigation-diary.md"
TASKS="$BASE/tasks.md"
CHANGELOG="$BASE/changelog.md"
REMOTE="/ai/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP"

remarquee upload bundle --dry-run "$INDEX" "$DESIGN" "$MATRIX" "$DIARY" "$TASKS" "$CHANGELOG" \
  --name "SMN-20260407 Annotation UI Consistency Pass" \
  --remote-dir "$REMOTE" \
  --toc-depth 2
remarquee upload bundle "$INDEX" "$DESIGN" "$MATRIX" "$DIARY" "$TASKS" "$CHANGELOG" \
  --name "SMN-20260407 Annotation UI Consistency Pass" \
  --remote-dir "$REMOTE" \
  --toc-depth 2 \
  --force
remarquee cloud ls "$REMOTE" --long --non-interactive
```

## Step 12: Restore queue dismiss-and-explain parity and show linked guidelines inside annotation detail

This follow-up started from two concrete visual regressions the user caught in live use: the review queue rows no longer exposed the orange dismiss-and-explain bubble that sender detail already had, and expanded annotation detail showed comment feedback but not the linked guidelines that had been attached during review. The screenshots were valuable because they narrowed the problem from “something feels inconsistent” to two specific missing affordances.

The actual fixes stayed small once the problem was localized. The queue bug was a page-composition omission: `ReviewQueuePage` never passed `onDismissExplain` to `AnnotationTable`, so the row action was never rendered there. The guideline bug was a read-model/rendering omission: `AnnotationDetail` had grown support for feedback cards, but no parallel slot for run-linked guidelines, even though the sender and run pages already had enough data to supply them.

### Prompt Context

**User prompt (verbatim):** "we are missing the dismiss and explain comment bubble icon for reviewing an annotation in the review queue, as is show in the senders view: /tmp/pi-clipboard-b67b3b4d-21d4-41fd-accf-7edac36e9337.png

I am also not seeing attached guidelines attached to items, say like here in papa@example.com /tmp/pi-clipboard-09561c28-b92a-451b-b2d7-c8ba917b7069.png

Analyze, and then fix. keep a diary, and commit"

**Assistant interpretation:** Compare the queue and sender annotation-row affordances, confirm the missing guideline visibility from the screenshot, then fix the page composition so both actions and attached artifacts appear where the user expects.

**Inferred user intent:** Make single-annotation review behavior visually consistent across queue/run/sender views and stop hiding linked review artifacts behind page-level panels.

**Commit (code):** `b7a3f74` — `AnnotationUI: restore queue comment action and show item guidelines`

### What I did
- Used image analysis on both screenshots to confirm:
  - the orange dismiss-and-explain bubble exists in sender rows between dismiss and expand,
  - the `papa@example.com` annotation detail showed review feedback but no visible linked guidelines.
- Updated `ui/src/pages/ReviewQueuePage.tsx` to:
  - restore per-row dismiss-and-explain via `onDismissExplain`,
  - open a single-item `ReviewCommentDrawer`,
  - load run-linked guidelines for the expanded row so the queue detail can show them inline.
- Extended `ui/src/components/AnnotationTable/AnnotationTable.tsx` and `ui/src/components/AnnotationTable/AnnotationDetail.tsx` with a `getGuidelines(...)` path and a compact linked-guidelines section.
- Updated `ui/src/pages/SenderDetailPage.tsx` to map the expanded annotation to the correct sender-guideline group and render those guidelines inside the annotation detail.
- Updated `ui/src/pages/RunDetailPage.tsx` to pass run-linked guidelines into expanded annotation detail.
- Refreshed the relevant Storybook story descriptions for queue and sender detail.
- Ran:

```bash
cd smailnail/ui
pnpm run check
pnpm run build-storybook
rm -rf storybook-static
```

### Why
- The queue and sender pages should not disagree on whether a single annotation can be dismissed with an explanation.
- If a reviewer attaches guidelines while reviewing an item, the most natural place to rediscover them is inside that item’s expanded detail, not only in a page-level run/sender summary section.

### What worked
- The queue bubble fix was simple once I confirmed the row icon is rendered purely from whether `onDismissExplain` is provided.
- Adding a general `getGuidelines(...)` hook to `AnnotationTable` avoided yet another page-specific annotation-detail fork.
- Existing run and sender guideline query results were already sufficient; the missing piece was composition, not a new backend endpoint.
- Both TypeScript validation and Storybook static build passed after the UI changes.

### What didn't work
- There was no implementation failure once the issue was localized. The main cost was diagnosis time: the queue bug looked like an icon/styling issue at first, but the real cause was a missing page prop.

### What I learned
- A reusable table/detail component needs parallel extension points for every artifact class we expect to show consistently. Adding feedback without adding guidelines created a subtle asymmetry that only showed up in real usage.
- Screenshot-driven bug reports are especially effective for page-composition regressions because they expose what the user expects to see, not just what the data model technically contains.

### What was tricky to build
- The tricky part was preserving the abstraction boundary of `AnnotationTable`. I wanted guideline visibility to be consistent without turning the table into a query owner. The solution was to mirror the existing `getFeedback(...)` pattern with a new `getGuidelines(...)` callback so each page still owns its data loading while the shared detail renderer owns presentation.
- The second subtle point was semantics: the guidelines visible on an annotation are run-linked guidelines relevant to that annotation’s run, not annotation-target rows stored separately in the database. The UI copy therefore says “Linked guidelines for this run” to avoid implying a narrower storage model than the backend actually has.

### What warrants a second pair of eyes
- Reviewers should confirm the wording “Linked guidelines for this run” is clear enough on sender detail, where the user may mentally associate the guideline with the selected annotation rather than the run.
- Reviewers should also confirm that restoring the queue’s single-item comment drawer does not create any confusing overlap with the existing batch reject/explain drawer.

### What should be done in the future
- If reviewers want stronger provenance, add lightweight badges or helper text distinguishing annotation-scoped feedback from run-scoped linked guidelines inside expanded detail.

### Code review instructions
- Start with:
  - `ui/src/pages/ReviewQueuePage.tsx`
  - `ui/src/components/AnnotationTable/AnnotationTable.tsx`
  - `ui/src/components/AnnotationTable/AnnotationDetail.tsx`
  - `ui/src/pages/SenderDetailPage.tsx`
  - `ui/src/pages/RunDetailPage.tsx`
- Validate by:
  - expanding a queue item and confirming the comment bubble appears again,
  - opening the single-item queue drawer and verifying it still submits a dismiss+comment mutation,
  - expanding a sender/run annotation that belongs to a run with linked guidelines and confirming the compact guideline cards appear under the feedback section.

### Technical details
- Screenshot analysis inputs:

```text
/tmp/pi-clipboard-b67b3b4d-21d4-41fd-accf-7edac36e9337.png
/tmp/pi-clipboard-09561c28-b92a-451b-b2d7-c8ba917b7069.png
```

- Validation commands:

```bash
cd smailnail/ui
pnpm run check
pnpm run build-storybook
rm -rf storybook-static
```

## Related

- Design doc: `../design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`
- Matrix: `./02-artifact-query-and-invalidation-matrix.md`
