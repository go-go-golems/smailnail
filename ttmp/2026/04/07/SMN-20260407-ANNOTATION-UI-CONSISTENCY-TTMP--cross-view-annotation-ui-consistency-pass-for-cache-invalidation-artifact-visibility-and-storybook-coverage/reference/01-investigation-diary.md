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
    - Path: pkg/annotationui/handlers_senders.go
      Note: Diary context for current sender detail payload composition
    - Path: ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage/design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md
      Note: Primary deliverable recorded in the diary
    - Path: ttmp/2026/04/07/SMN-20260407-ANNOTATION-UI-CONSISTENCY-TTMP--cross-view-annotation-ui-consistency-pass-for-cache-invalidation-artifact-visibility-and-storybook-coverage/reference/02-artifact-query-and-invalidation-matrix.md
      Note: Canonical page-by-page matrix added during phase-1 planning
    - Path: ui/src/api/annotations.ts
      Note: Diary context for query/tag observations gathered during the investigation
    - Path: ui/src/mocks/handlers.ts
      Note: Diary context for Storybook/MSW mutation-state limitations
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Diary context for the run-detail composed-query baseline
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Diary context for the sender-detail artifact visibility gap
ExternalSources: []
Summary: Chronological diary for creating the annotation UI consistency ticket, gathering evidence across backend/frontend/storybook layers, and producing the design/implementation guide plus reMarkable delivery.
LastUpdated: 2026-04-07T11:15:00-04:00
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

## Related

- Design doc: `../design-doc/01-analysis-and-implementation-guide-for-annotation-ui-consistency-and-artifact-visibility.md`
- Matrix: `./02-artifact-query-and-invalidation-matrix.md`
