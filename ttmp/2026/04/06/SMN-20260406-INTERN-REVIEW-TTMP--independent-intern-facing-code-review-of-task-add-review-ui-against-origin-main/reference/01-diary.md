---
Title: Diary
Ticket: SMN-20260406-INTERN-REVIEW-TTMP
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
    - Path: pkg/annotate/schema.go
      Note: Diary context for splitting annotation v3 tables from review v4 tables
    - Path: pkg/annotationui/server.go
      Note: Diary context for request-scoped review actor helper
    - Path: pkg/annotationui/server_test.go
      Note: |-
        Validation context and existing server test coverage baseline
        Focused endpoint coverage for guideline-linked runs
    - Path: pkg/mirror/schema.go
      Note: |-
        Diary context for the sqlite schema-version split that upgrades legacy review databases
        Diary context for the schema-version bug fix that unblocks guideline creation on legacy DBs
    - Path: pkg/mirror/store_test.go
      Note: |-
        Regression coverage for upgrading legacy schema-version-3 sqlite databases
        Diary context for the copied-db regression and upgrade validation
    - Path: ttmp/2026/04/06/SMN-20260406-REVIEW-REVIEW-TTMP--meta-review-of-the-intern-code-review-for-task-add-review-ui/design-doc/01-review-of-the-intern-s-code-review-for-the-review-ui-branch.md
      Note: Diary records the revision step after the later meta-review
    - Path: ui/package.json
      Note: Validation and package-manager notes recorded in the diary
    - Path: ui/src/components/Guidelines/GuidelineSummaryCard.tsx
      Note: Cleanup of fake linked run count prop usage
    - Path: ui/src/pages/GuidelineDetailPage.tsx
      Note: Diary context for awaited create-and-link flow and partial-success error handling
    - Path: ui/src/pages/GuidelinesListPage.tsx
      Note: Removal of fake linkedRunCount wiring from the live guideline list
    - Path: ui/src/pages/stories/ReviewQueuePage.stories.tsx
      Note: Diary context for pending-only Review Queue story updates
    - Path: ui/src/types/reviewFeedback.ts
      Note: |-
        Diary evidence for TS/Go contract drift
        Diary context for scopeKind frontend feedback filter support
ExternalSources: []
Summary: Chronological diary of the independent code review, including the initial docmgr/root mistakes, investigation commands, validation runs, and handoff notes.
LastUpdated: 2026-04-07T10:20:00Z
WhatFor: Capture how the review was produced so another engineer can reproduce or continue it.
WhenToUse: Read this when you want the exact commands, errors, and reasoning behind the review report.
---







# Diary

## Goal

This diary records how I produced an independent code review of `smailnail` against `origin/main`, with a focus on the `task/add-review-ui` branch. It captures the commands I ran, the mistakes I made while creating the docmgr workspace, the evidence-gathering process, and the validation steps I used before writing the final report.

## Step 1: Create an independent ticket in the right docmgr root

I first needed a ticket workspace that was clearly separate from the already-existing review work. My initial instinct was correct at a high level—create a fresh ticket and do not read the existing review docs—but I made two operational mistakes: I started one level above the actual git repo, and I initially pointed `docmgr --root` at the repo root instead of the `smailnail/ttmp` docs root the user explicitly asked for.

That mattered because the task was not just “write a report somewhere”; it was specifically “create a new docmgr ticket in `smailnail/ttmp`.” I corrected that by creating a second, properly rooted ticket under `smailnail/ttmp` and then doing all actual documentation work there.

### Prompt Context

**User prompt (verbatim):** "Create a detailed code review of smailnail against origin/main. 

Create a new docmgr ticket in smailnail/ttmp (docmgr --root ...) and then reate a review report that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

                                                                                                                     
Look at a confusing code, deprecated code, unused code, unclear naming, etc..."

**Assistant interpretation:** Create a brand-new docmgr ticket under the repo’s `ttmp` workspace, independently review the branch against `origin/main`, write a very detailed intern-facing report plus a diary, validate it, and upload the result to reMarkable.

**Inferred user intent:** Produce a stand-alone onboarding-quality code review that does not piggyback on or depend on an existing review ticket.

### What I did
- Loaded the `ticket-research-docmgr-remarkable`, `docmgr`, `remarkable-upload`, `diary`, and `code-quality-review-cleanup` skills and their reference docs.
- Checked the working directory and discovered I was initially one level above the real git repo.
- Inspected existing docmgr status and noticed an existing `SMN-20260406-CODE-REVIEW` ticket, so I chose a different ticket name for my own work.
- First created `SMN-20260406-INTERN-REVIEW` with `docmgr --root <repo-root>`.
- Realized that did not satisfy the user’s “in `smailnail/ttmp`” requirement.
- Created the correct ticket as `SMN-20260406-INTERN-REVIEW-TTMP` with `docmgr --root /.../smailnail/ttmp`.

### Why
- The ticket had to be independent.
- The root had to be the actual docs workspace requested by the user.
- I wanted the final review to have a clear audit trail instead of being mixed into a previously existing review thread.

### What worked
- `docmgr --root /.../smailnail/ttmp` created the correct ticket structure.
- The unique ticket ID avoided ambiguity with the pre-existing review ticket.

### What didn't work
- Running git commands from the workspace parent failed:

```bash
pwd && git rev-parse --show-toplevel
```

Error:

```text
fatal: not a git repository (or any of the parent directories): .git
```

- Reusing a duplicate ticket ID caused docmgr ambiguity when adding docs:

```text
Error: failed to find ticket directory: ambiguous ticket index doc for SMN-20260406-CODE-REVIEW (got 2)
```

- Using `docmgr --root /.../smailnail` created a ticket outside the intended `ttmp` docs root, which was not what the user asked for.

### What I learned
- In this repo, `docmgr --root` needs to point at `smailnail/ttmp` if I want the physical documents to live there.
- Using a unique ticket ID immediately is safer when the workspace already contains similarly named review tickets.

### What was tricky to build
- The tricky part was not the code review itself; it was making sure the documentation workspace matched the user’s requested location while avoiding collisions with existing tickets.
- The `.ttmp.yaml` file exists at repo root, but `docmgr` will still behave differently depending on which root path I pass. That can be easy to misread when you are moving quickly.

### What warrants a second pair of eyes
- The accidentally created repo-root ticket is unrelated to the requested deliverable and should be ignored. If someone later cleans up ticket hygiene, they may want to decide whether to archive or remove that stray workspace.

### What should be done in the future
- Prefer `docmgr --root /abs/path/to/repo/ttmp` immediately when the user explicitly names the docs root.

### Code review instructions
- Confirm the real deliverable lives under:
  - `smailnail/ttmp/2026/04/06/SMN-20260406-INTERN-REVIEW-TTMP--independent-intern-facing-code-review-of-task-add-review-ui-against-origin-main`
- Ignore the earlier accidental ticket outside `ttmp`.

### Technical details
- Ticket created:

```bash
cd smailnail
ROOT=/home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ttmp
TICKET=SMN-20260406-INTERN-REVIEW-TTMP

docmgr --root "$ROOT" ticket create-ticket --ticket "$TICKET" --title "Independent intern-facing code review of task/add-review-ui against origin/main" --topics annotations,backend,frontend,sqlite,workflow
docmgr --root "$ROOT" doc add --ticket "$TICKET" --doc-type design-doc --title "Intern guide and independent code review of the review UI branch"
docmgr --root "$ROOT" doc add --ticket "$TICKET" --doc-type reference --title "Diary"
```

## Step 2: Inventory the branch and map the review architecture

Once the ticket root was correct, I switched to evidence gathering. I wanted the report to work for a new intern, so I did not just scan changed files; I first reconstructed the high-level system: what smailnail is at the repo level, where the review UI lives, how the sqlite server is started, which files define the domain model, and how the frontend talks to the backend.

I then focused on the branch diff itself. The diff was heavily concentrated in `pkg/annotate`, `pkg/annotationui`, and `ui/`, which told me the branch was mostly a review-workflow branch rather than a general mailbox/retrieval change.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Build an evidence-based architecture map first, then identify confusing, unused, deprecated, or unclear code in the changed review UI slice.

**Inferred user intent:** Make the report teach the system before critiquing it, so a new intern can both understand and improve the code.

### What I did
- Ran `git diff --stat origin/main...HEAD` and `git diff --name-only origin/main...HEAD`.
- Read `README.md`, `go.mod`, `cmd/smailnail/main.go`, `cmd/smailnaild/main.go`, `ui/package.json`.
- Read the review-specific backend files:
  - `pkg/annotate/types.go`
  - `pkg/annotate/schema.go`
  - `pkg/annotate/repository.go`
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotationui/server.go`
  - `pkg/annotationui/handlers_annotations.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/types_feedback.go`
- Read the key frontend files:
  - `ui/src/App.tsx`
  - `ui/src/api/annotations.ts`
  - `ui/src/pages/ReviewQueuePage.tsx`
  - `ui/src/pages/RunDetailPage.tsx`
  - `ui/src/pages/GuidelinesListPage.tsx`
  - `ui/src/pages/GuidelineDetailPage.tsx`
  - several components around review feedback and guideline linking
- Pulled line-numbered excerpts with `nl -ba ... | sed -n ...` for evidence-backed citations.
- Ran targeted `rg` searches for TODO/FIXME/deprecation markers and for test coverage of the new feedback/guideline symbols.

### Why
- The intern-facing report needed more than a bug list; it needed a system map.
- The user specifically asked for prose, bullet points, pseudocode, diagrams, API references, and file references.
- The most reliable way to produce that is to walk the code in the same order the runtime uses it.

### What worked
- The diff split was clear enough to identify the feature center quickly:
  - review backend tables/repository methods,
  - HTTP handlers,
  - RTK Query hooks,
  - review/guideline pages and components.
- The files were generally readable enough to reconstruct the end-to-end flow without guessing.

### What didn't work
- Searching for explicit TODO/FIXME/deprecated markers in the reviewed files did not produce useful results:

```bash
rg -n "TODO|FIXME|deprecated|Deprecated|unused|XXX" pkg/annotate pkg/annotationui ui/src cmd/smailnail/commands/enrich README.md pkg/doc -S
```

This was not a failure in the code; it just meant I had to identify confusing or unfinished areas from behavior and contract mismatches rather than from inline markers.

### What I learned
- The branch’s strongest idea is the review-feedback/guideline loop, not any one UI component.
- The clearest architecture seams are:
  - schema/model in `pkg/annotate`,
  - transport in `pkg/annotationui`,
  - typed client in `ui/src/api/annotations.ts`,
  - route/page composition in `ui/src/pages`.
- Several of the most important issues are semantic mismatches, not syntax or typecheck failures.

### What was tricky to build
- The hard part of this step was separating “feature incompleteness” from “broken implementation.” Some things are intentionally early-stage, but even early-stage code can mislead the user or the next engineer if the naming or UI implies more completeness than the implementation really has.

### What warrants a second pair of eyes
- The boundary between `smailnaild` and the sqlite review server is conceptually confusing for first-time readers. It is worth having a maintainer confirm whether that separation is intended to remain long-term.

### What should be done in the future
- Add a short architecture note near the review server entrypoint explaining why this slice is served from `smailnail sqlite serve` instead of `smailnaild`.

### Code review instructions
- Start at `README.md`, then `cmd/smailnail/commands/sqlite/serve.go`, then `pkg/annotationui/server.go`, then `ui/src/api/annotations.ts`.
- That order mirrors the runtime layering.

### Technical details
- Diff summary command:

```bash
cd smailnail
git diff --stat origin/main...HEAD
```

- File counts command:

```bash
echo FILES=$(git diff --name-only origin/main...HEAD | wc -l)
echo BACKEND=$(git diff --name-only origin/main...HEAD | rg '^(pkg/annotate|pkg/annotationui|cmd/smailnail/commands/enrich|pkg/doc)' | wc -l)
echo FRONTEND=$(git diff --name-only origin/main...HEAD | rg '^ui/' | wc -l)
```

## Step 3: Validate, write the review, and prepare delivery

After the architecture and findings were clear, I ran focused validation commands and then wrote the actual design-doc report and this diary. I deliberately validated both the backend and frontend in the narrow scope I had reviewed, because the final handoff should distinguish between "I think this is wrong" and "this actually fails to build or typecheck."

The most interesting validation result was that a plain `go test` fails before it even reaches the reviewed packages unless the repository is built with the required SQLite FTS5 tag. That is not a new branch bug, but it is absolutely the kind of operational detail a new intern must know when verifying this area.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the review like an engineer, not just a note-taker: validate the slice, document exact commands, and make the handoff reproducible.

**Inferred user intent:** Leave behind a deliverable another person can trust and independently verify.

### What I did
- Ran backend tests without tags first, observed the build-tag failure, then reran with `-tags sqlite_fts5`.
- Ran `pnpm run check` in `ui/`.
- Wrote the design doc with:
  - executive summary,
  - system map,
  - API references,
  - diagrams,
  - detailed issues with cleanup sketches,
  - phased implementation plan,
  - testing strategy.
- Wrote this diary with exact commands and failures.

### Why
- The review needed concrete evidence, not just opinions.
- The user asked for a very detailed new-intern-friendly report, which requires explicit verification notes.

### What worked
- `go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1` passed.
- `pnpm run check` passed.
- The report structure mapped cleanly onto the evidence gathered in Step 2.

### What didn't work
- Running backend tests without the required build tag failed:

```bash
cd smailnail
go test ./pkg/annotate ./pkg/annotationui -count=1
```

Error:

```text
# github.com/go-go-golems/smailnail/pkg/mirror
pkg/mirror/require_fts5_build_tag.go:5:9: undefined: requires_sqlite_fts5_build_tag
FAIL	github.com/go-go-golems/smailnail/pkg/annotate [build failed]
FAIL	github.com/go-go-golems/smailnail/pkg/annotationui [build failed]
FAIL
```

### What I learned
- The branch is type-safe enough to pass `pnpm run check`, which means many of the real problems are behavioral and contract-level rather than TS compiler failures.
- The repository-level SQLite FTS5 build tag is a real onboarding sharp edge and worth explicitly documenting in a review meant for new contributors.

### What was tricky to build
- The trickiest part of writing the report was balancing onboarding clarity against code-review sharpness. A new intern needs the big picture first, but the user also asked for concrete cleanup around confusing and unclear code. I solved that by front-loading architecture sections and then using issue templates with specific file/line references and cleanup sketches.

### What warrants a second pair of eyes
- The semantic issues around feedback scope, review-queue meaning, and audit metadata are high-value enough that I would want a maintainer to confirm intended behavior before someone starts implementing fixes blindly.

### What should be done in the future
- Add the validation commands from this diary into a small review playbook in the ticket if this kind of branch review becomes common.

### Code review instructions
- Reproduce validation with:

```bash
cd /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1
cd ui && pnpm run check
```

- Read the final report first:
  - `ttmp/2026/04/06/SMN-20260406-INTERN-REVIEW-TTMP--independent-intern-facing-code-review-of-task-add-review-ui-against-origin-main/design-doc/01-intern-guide-and-independent-code-review-of-the-review-ui-branch.md`

### Technical details
- Successful validation commands:

```bash
cd smailnail
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1

cd smailnail/ui
pnpm run check
```

## Step 4: Revise the review after the later meta-review of the intern ticket

After writing the original report, I later reviewed the separate intern code-review ticket and then came back to my own document to decide whether it should change. The answer was yes, but only in a targeted way. I did not weaken the core semantic findings from the original review; instead I incorporated a small set of additional, source-validated cleanup findings that the intern surfaced and that I agreed were worth preserving.

The key editorial rule in this revision step was: **only add things from the intern review if I revalidated them directly in code**. That prevented the updated report from inheriting the intern review’s weaker prioritization and its incorrect claims. I explicitly labeled the imported additions in the design doc as coming from the later review pass.

### Prompt Context

**User prompt (verbatim):** "go ahead, update it and reupload it. note when things were added from the intern review"

**Assistant interpretation:** Amend the original independent review to incorporate the useful validated findings from the intern review, label those additions clearly, and re-upload the updated bundle.

**Inferred user intent:** Keep the stronger original review as the main document, but let it benefit from useful second-pass discoveries without losing provenance.

### What I did
- Re-read the meta-review conclusions and selected only the intern findings that were worth importing.
- Updated the main design doc to add:
  - a revision note near the top,
  - a new section with added validated findings,
  - implementation-plan updates,
  - references to the intern review and newly relevant files.
- Explicitly labeled imported items with markers like `**[Added after reviewing the intern ticket]**`.
- Prepared the existing bundle for re-upload.

### Why
- The user explicitly asked whether I should update my own review in light of the new findings, and then asked me to do it.
- Some of the intern’s lower-level cleanup findings were genuinely useful and improved the report.
- Provenance mattered: I did not want later readers to confuse the original independent findings with additions inspired by the intern review.

### What worked
- The added items fit naturally into the existing report without changing the top-level conclusions.
- Labeling the imported findings inline makes the revision history clear.

### What didn't work
- N/A

### What I learned
- A second review is most useful when it sharpens and annotates the first one rather than replacing it wholesale.
- Provenance labels matter when combining two review passes; otherwise readers cannot tell which findings were original and which were imported later.

### What was tricky to build
- The tricky part was deciding what *not* to import. The intern review had some valid but low-value sprawl, and some incorrect claims. I only imported the findings that were both correct and meaningfully useful.

### What warrants a second pair of eyes
- If we later turn this report into an implementation plan, a maintainer should still do a final pass on the top-priority ordering after combining both reviews.

### What should be done in the future
- When revising a review after a later audit, add explicit provenance tags rather than silently merging the new material.

### Code review instructions
- In the updated design doc, search for:
  - `Revision note after meta-review of the intern review`
  - `Added after reviewing the intern ticket`
- Those markers show exactly what was imported later.

### Technical details
- Relevant references used during this revision step:
  - `ttmp/2026/04/06/SMN-20260406-CODE-REVIEW--code-review-run-review-feedback-guidelines-mailbox-aware-analysis/design-doc/01-comprehensive-code-review-run-review-feedback-guidelines-mailbox-aware-analysis.md`
  - `ui/src/mocks/handlers.ts`
  - `ui/src/components/ReviewFeedback/ReviewCommentInline.tsx`
  - `ui/src/pages/stories/GuidelinesListPage.stories.tsx`
  - `ui/src/pages/stories/GuidelineDetailPage.stories.tsx`

## Step 5: Turn selected findings into implementation phases and ship guideline linked runs

After the later contract-codegen work closed finding 3, I came back to this original intern-review ticket to treat it as a living follow-up plan rather than a dead report. The user explicitly wanted finding 5 shipped fully, findings 7 and 8 left alone for now, and finding 9 either properly wired or cleaned up.

### Prompt Context

**User prompt (verbatim):** "ok, back to @smailnail/2026/04/06/SMN-20260406-INTERN-REVIEW--independent-intern-facing-code-review-of-task-add-review-ui-against-origin-main/ . we tackled finding 3 with the protobuf, for 5, , add the backend endpoint to hship it fully, . no need to worry about 7 for now. 8 neither. 9. wire things up properly (or clean it).

Add detailed tasks to the ticket, in the different phases, then work on them one by one, committing at appropriate intervals and keepting a detailed diary"

**Assistant interpretation:** Use the intern-review ticket as the execution tracker for selected follow-up findings, add a phased plan first, then implement the agreed slices in order with commits and diary updates.

**Inferred user intent:** Do not reopen the entire review. Instead, execute a narrow, high-value subset of the findings cleanly and document the work like a real follow-up project.

### What I did
- Re-read the original findings around guideline linked runs and dead review UI state.
- Updated `tasks.md` to create three concrete phases:
  - Phase 1: ship guideline linked runs fully,
  - Phase 2: clean or wire dead review UI state,
  - Phase 3: finish ticket hygiene and handoff.
- Implemented Phase 1:
  - added `Repository.ListGuidelineRuns(...)` in `pkg/annotate/repository_feedback.go`
  - added `GET /api/review-guidelines/{id}/runs` in `pkg/annotationui/server.go`
  - added `handleListGuidelineRuns(...)` in `pkg/annotationui/handlers_feedback.go`
  - added frontend hook `useGetGuidelineRunsQuery` in `ui/src/api/annotations.ts`
  - loaded live linked runs in `ui/src/pages/GuidelineDetailPage.tsx`
  - added a loading state to `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx`
  - updated MSW in `ui/src/mocks/handlers.ts`
  - added focused endpoint coverage in `pkg/annotationui/server_test.go`

### Why
- Finding 5 was one of the clearest “placeholder pretending to be a feature” issues in the original report.
- Shipping the linked-runs section fully is more valuable than leaving a polished but empty component in the UI.
- The work was narrow enough to land as a focused phase without pulling in the deferred route-architecture or broader testing findings.

### What worked
- The existing `AgentRunSummary` contract was already sufficient, so I did not need a new protobuf schema for the endpoint.
- The new endpoint fit cleanly into the existing annotation UI contract layer by reusing `annotationuiv1.AgentRunListResponse`.
- The repository query could be implemented as a single SQL summary query instead of an N+1 run-by-run lookup.

### What didn't work
- The first frontend pass tried to use `mode` in the RTK Query `skip` option before `mode` was declared, and TypeScript caught it immediately:

```text
src/pages/GuidelineDetailPage.tsx(37,20): error TS2448: Block-scoped variable 'mode' used before its declaration.
src/pages/GuidelineDetailPage.tsx(37,20): error TS2454: Variable 'mode' is used before being assigned.
```

Fix:
- moved the `useState` declaration for `mode` above the `useGetGuidelineRunsQuery(...)` call.

### What I learned
- A good follow-up implementation ticket can grow naturally out of a strong review ticket if the findings are already evidence-backed and scoped sharply.
- Reusing the existing run-summary contract avoided unnecessary schema churn.

### What was tricky to build
- The main design choice was where to put the new contract. I chose to reuse the existing run-summary response from the broader annotation UI contract rather than creating a guideline-specific wrapper that would only repeat the same shape.
- I also had to avoid mixing this work with unrelated local changes already present in the working tree.

### What warrants a second pair of eyes
- Whether the guideline-runs endpoint should eventually return 404 for unknown guideline IDs rather than an empty list. For now it is consistent with the existing run-guideline list behavior.

### What should be done in the future
- If the guideline list page later needs a true linked-run count, add that explicitly to the guideline list/read model rather than synthesizing fake zero values in the UI.

### Code review instructions
- Start with:
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/server.go`
  - `pkg/annotationui/server_test.go`
  - `ui/src/api/annotations.ts`
  - `ui/src/pages/GuidelineDetailPage.tsx`
  - `ui/src/components/Guidelines/GuidelineLinkedRuns.tsx`

### Technical details
- Validation commands:

```bash
cd smailnail
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1

cd smailnail/ui
pnpm run check
```

## Step 6: Clean finding 9 instead of pretending the dead UI state is still real

With the guideline-linked-runs slice landed, I moved to the narrower cleanup half of the user request: finding 9. I deliberately did not tackle the package-manager or embed-asset policy parts of that finding, because the user asked for the UI state to be wired properly or cleaned, not for a repo-wide tooling detour.

### Prompt Context

**User prompt (verbatim):** same as Step 5

**Assistant interpretation:** Remove the dead review UI state if it is not serving the live page flow, and stop advertising fake guideline list data.

**Inferred user intent:** Prefer honest, simpler UI state over speculative or misleading plumbing.

### What I did
- Re-audited `ui/src/store/annotationUiSlice.ts` against `ui/src/pages/ReviewQueuePage.tsx`.
- Confirmed the live page only uses:
  - `selected`
  - `filterTag`
  - `expandedId`
- Removed dead review-queue slice fields and actions that were no longer wired:
  - `filterType`
  - `filterSource`
  - `filterRunId`
  - `commentDrawerOpen`
  - `filterMailbox`
- Kept `commentDrawerOpen` as local component state in `ReviewQueuePage`, where it is actually used.
- Updated `GuidelineSummaryCard` so `linkedRunCount` is optional rather than mandatory.
- Removed `linkedRunCount={0}` from `GuidelinesListPage`, so the UI no longer pretends to know a count it does not have.
- Re-ran frontend validation with `pnpm run check`.

### Why
- Dead global state is worse than no state; it implies product capabilities and interaction paths that do not exist.
- Hard-coding zero for linked-run counts makes the guideline list look “complete” while still lying about the data source.

### What worked
- The cleanup was mechanically small because the dead Redux fields had no live callers.
- Making `linkedRunCount` optional preserved Storybook/component demo flexibility while letting the real list page stop passing fake data.

### What didn't work
- N/A

### What I learned
- The right cleanup is often to reduce required props and state shape, not to invent more plumbing just to keep an old interface alive.

### What was tricky to build
- The main judgment call was deciding what counted as the “real” owner of state. In this case, the drawer open/close state clearly belongs to the page component, not Redux.

### What warrants a second pair of eyes
- If future review-queue filters are reintroduced, they should come back together with visible controls on the page instead of sleeping in the slice unused.

### What should be done in the future
- If guideline list cards eventually need a linked-run count, add it to the guideline read model or list response explicitly and then thread it through the UI honestly.

### Code review instructions
- Start with:
  - `ui/src/store/annotationUiSlice.ts`
  - `ui/src/pages/ReviewQueuePage.tsx`
  - `ui/src/components/Guidelines/GuidelineSummaryCard.tsx`
  - `ui/src/pages/GuidelinesListPage.tsx`

### Technical details
- Validation command:

```bash
cd smailnail/ui
pnpm run check
```

- Focused commits created during this follow-up:
  - `5192aef` — `AnnotationUI: add guideline linked runs endpoint`
  - `4250dee` — `AnnotationUI: clean dead review UI state`

- Ticket hygiene command:

```bash
cd smailnail
docmgr --root /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ttmp doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30
```

## Step 7: Fix finding 1 by making Review Queue actually pending-only

After the earlier findings 5 and 9 follow-up work, I moved to the next high-priority semantic issue from the original report: the Review Queue was still querying all annotations instead of only pending-review items.

### Prompt Context

**User prompt (verbatim):** "ok what about 1,2 4? Add tasks for those and then work on them in the same fashion"

**Assistant interpretation:** Add new phased tasks for findings 1, 2, and 4, then implement them with the same ticket/diary/commit discipline used for findings 5 and 9.

**Inferred user intent:** Continue treating the review ticket as a living implementation plan and land the remaining high-value findings incrementally.

### What I did
- Updated `tasks.md` to add new phases for findings 1, 2, and 4.
- Changed `ReviewQueuePage` so the main annotation query always includes `reviewState: "to_review"`.
- Changed the queue’s tag-count query to derive from the same pending-review population rather than all annotations in the database.
- Updated `ui/src/pages/stories/ReviewQueuePage.stories.tsx` so the queue stories return wrapper responses and respect the queue semantics more honestly.

### Why
- The page is literally named “Review Queue,” so loading reviewed and dismissed rows by default was one of the most misleading semantics in the UI.
- Fixing the queue query at the page boundary was much cheaper and clearer than trying to explain the mismatch away in docs.

### What worked
- The backend already supported `reviewState` filtering, so the fix was mostly frontend query hygiene.
- The MSW handler already understood `reviewState`, which kept the runtime mock path consistent once the page started sending that filter.

### What didn't work
- Some of the existing stories were still returning raw arrays instead of the wrapper response shape expected by the live API layer. I fixed that while touching the queue stories so they stopped silently depending on outdated mock behavior.

### What I learned
- Semantic bugs are often “cheap” to fix once the contract layer is already explicit.

### What was tricky to build
- The main judgment call was whether tag-count pills should continue to represent all annotations or only queue-eligible annotations. I chose the latter because the page is specifically a review queue, not a generic annotation browser.

### Code review instructions
- Start with:
  - `ui/src/pages/ReviewQueuePage.tsx`
  - `ui/src/pages/stories/ReviewQueuePage.stories.tsx`

### Technical details
- Validation command:

```bash
cd smailnail/ui
pnpm run check
```

- Focused commit:
  - `8fc85a4` — `AnnotationUI: make review queue pending-only`

## Step 8: Fix finding 2 and finding 4 together because the full-repo pre-commit path forced them to land as one slice

The next two findings were conceptually separate but mechanically adjacent: finding 2 (run feedback should filter by `scopeKind`) and finding 4 (audit metadata should actually be populated). I initially attempted to stage and commit the scope-filtering slice by itself, but the repository’s pre-commit hook runs full-repo tests and stashes unstaged changes. That mattered because I had already added a new audit-focused test in the working tree, and the hook temporarily hid the matching handler changes.

### What I tried first
I tried to isolate the scope-filtering change into its own commit.

### What didn't work
The isolated commit attempt failed during pre-commit because the hook stashed the unstaged audit-handler changes while still running the new audit test, which then failed against the older staged code.

The key failure looked like this:

```text
--- FAIL: TestHandlerPopulatesAuditMetadata (0.06s)
    audit_test.go:47: feedback createdBy = ""
```

That was not a product bug in the final implementation; it was a staging/verification artifact caused by trying to split two touching handler changes while the hook tested the full repo state.

### Fix / approach change
- I kept finding 1 as its own commit.
- I then combined findings 2 and 4 into one focused backend/frontend integrity slice so the hook-tested tree was internally consistent.

### What I did
For **finding 2**:
- Extended `annotate.ListFeedbackFilter` with `ScopeKind`.
- Added `scopeKind` filtering in `Repository.ListReviewFeedback(...)`.
- Threaded `scopeKind` through `handleListFeedback(...)`.
- Extended frontend `FeedbackFilter` and updated `RunDetailPage` to call:
  - `useListReviewFeedbackQuery({ agentRunId: runId, scopeKind: "run" })`
- Updated MSW feedback handlers to honor `scopeKind`.
- Added focused backend coverage proving `GET /api/review-feedback?...&scopeKind=run` excludes annotation-scoped feedback.

For **finding 4**:
- Added a request-scoped actor helper in the annotation UI HTTP layer.
- Populated `CreatedBy` in:
  - feedback creation,
  - guideline creation,
  - review-with-artifacts,
  - batch-review-with-artifacts.
- Populated `LinkedBy` in explicit run-guideline link handlers.
- Added a dedicated backend audit test that verifies authored feedback, authored guidelines, review-created feedback artifacts, and explicit guideline links all carry the request actor.

### Why
- Finding 2 fixes a real semantics bug in the run detail page.
- Finding 4 makes the schema’s authorship fields honest instead of decorative.
- Landing them together avoided fragile hook behavior and produced a cleaner commit boundary around “feedback integrity + audit metadata.”

### What worked
- The backend and frontend changes for `scopeKind` were small and aligned naturally.
- The request-scoped actor helper gave the local annotation UI a practical short-term answer without forcing a broader auth redesign.
- The dedicated audit test now protects the new authorship behavior from regressing silently.

### What I learned
- In repos with full-tree pre-commit validation, the technically “smallest” commit is not always the best commit if the hook stashes related changes and tests the rest of the tree.

### What was tricky to build
- The trickiest part was not the code; it was the sequencing. I had to adapt the commit plan to the repository’s verification model rather than forcing a too-granular split.

### Code review instructions
- Start with:
  - `pkg/annotate/types.go`
  - `pkg/annotate/repository_feedback.go`
  - `pkg/annotationui/handlers_feedback.go`
  - `pkg/annotationui/handlers_annotations.go`
  - `pkg/annotationui/server.go`
  - `pkg/annotationui/server_test.go`
  - `pkg/annotationui/audit_test.go`
  - `ui/src/types/reviewFeedback.ts`
  - `ui/src/pages/RunDetailPage.tsx`
  - `ui/src/mocks/handlers.ts`

### Technical details
- Validation commands:

```bash
cd smailnail
go test -tags sqlite_fts5 ./pkg/annotate ./pkg/annotationui -count=1

cd smailnail/ui
pnpm run check
```

- Combined focused commit:
  - `991e0f7` — `AnnotationUI: scope run feedback and populate audit fields`

## Step 9: Refresh the ticket docs and handoff state after findings 1, 2, and 4

After the code landed, I updated the ticket docs again so the intern-review workspace remained a trustworthy source of truth rather than a stale historical artifact.

### What I did
- Marked phases 3, 4, and 5 complete in `tasks.md`.
- Updated `changelog.md` with the new commits and behavior changes.
- Updated `index.md` so it now reflects that findings 1, 2, 4, 5, and 9 are implemented, with 3 completed in the separate protobuf contract ticket.
- Related the newly changed code files to the ticket docs via `docmgr`.
- Ran `docmgr doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30`.

### Why
- The user explicitly asked for phased ticket tracking and a detailed diary, so the docs needed to stay synchronized with the implementation work.

### Technical details
- Ticket doctor command:

```bash
cd smailnail
docmgr --root /home/manuel/workspaces/2026-04-03/js-repl-smailnail/smailnail/ttmp doctor --ticket SMN-20260406-INTERN-REVIEW-TTMP --stale-after 30
```

## Step 10: Fix finding 6 by awaiting guideline-link flows and surfacing failures

The next deferred item the user explicitly brought back into scope was finding 6. This was a frontend flow problem rather than a backend contract problem: the UI fired guideline-link mutations and then immediately closed or navigated away, which could silently hide failures.

### Prompt Context

**User prompt (verbatim):** "Ok, address that one too."

**Assistant interpretation:** Take the remaining async-linking semantics issue and fix it with the same phase/commit/doc discipline as the earlier findings.

### What I did
- Updated `tasks.md` to add a dedicated phase for finding 6.
- Changed `GuidelineLinkPicker` so `onLink` can be async and so the dialog waits for the callback before clearing local selection.
- Changed `RunGuidelineSection` to:
  - await all selected link mutations with `Promise.all(...unwrap())`,
  - keep the picker open on failure,
  - show an inline error alert instead of silently failing,
  - await unlink operations too and surface unlink failures in the same area.
- Changed `GuidelineDetailPage` to:
  - await guideline creation,
  - await run-link creation before navigating back to the run page,
  - handle the partial-success case where guideline creation succeeds but run-link creation fails by routing to the new guideline detail page with an explicit error message.
- Updated the `RunGuidelineSection` Storybook handlers so they still return wrapper-response shapes matching the real API contract.

### Why
- The original problem was a classic async UI lie: the UI acted like success had already happened.
- This kind of issue is especially painful in review tooling because the user is making deliberate, high-value curation decisions and expects those links to stick.

### What worked
- `unwrap()` made the mutation success/failure path much clearer than the old fire-and-forget usage.
- Keeping the picker open on failure means users do not lose context when a request fails.
- The guideline-create flow now handles the awkward partial-success case much more honestly.

### What didn't work
- No hard implementation blocker here; the main work was carefully deciding how to handle the “guideline created, but link failed” edge case without leaving the user stranded on `/new`.

### What I learned
- Async UI correctness often needs a small amount of explicit error state, not just `await`.
- The best failure UX here was not “stay on `/new` forever”; it was “navigate to the created guideline and clearly say that the run link failed.”

### What was tricky to build
- The trickiest part was making the picker preserve user intent on failure instead of clearing local selection too early.
- The second tricky part was the create-and-link partial success case, because there are really two operations with different failure modes.

### Code review instructions
- Start with:
  - `ui/src/components/ReviewFeedback/GuidelineLinkPicker.tsx`
  - `ui/src/components/RunGuideline/RunGuidelineSection.tsx`
  - `ui/src/pages/GuidelineDetailPage.tsx`
  - `ui/src/components/RunGuideline/stories/RunGuidelineSection.stories.tsx`

### Technical details
- Validation command:

```bash
cd smailnail/ui
pnpm run check
```

- Focused commit:
  - `0897d2b` — `AnnotationUI: await guideline link flows`

## Step 11: Refresh ticket docs after finding 6

After the code landed, I updated the ticket again so the follow-up history remained chronological and explicit.

### What I did
- Marked the new finding-6 phase complete in `tasks.md`.
- Updated `changelog.md`, `index.md`, and this diary.
- Related the new async-link flow files to the ticket docs with `docmgr`.
- Re-ran `docmgr doctor` for the ticket.

## Step 12: Fix legacy sqlite schema upgrades for review tables

While testing the review UI against a long-lived merged sqlite database, the user hit a real persistence bug that had nothing to do with React state or request payloads: guideline creation failed with `insert review guideline: no such table: review_guidelines`. I traced that back to mirror schema bootstrapping rather than the guideline handlers themselves.

The root cause was versioning drift inside the sqlite bootstrap path. Older databases had already been stamped `schema_version=3` before the review/guideline tables were folded into the version-3 annotation migration helper, so bootstrap would skip reapplying that migration and never create the newer review tables.

### Prompt Context

**User prompt (verbatim):** "I have another issue, I am running  go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve --sqlite-path /home/manuel/smailnail/smailnail-last-24-months-merged.sqlite (and pnpm dev on :5050), and for example creating guidelines I get: insert review guideline: no such table: review_guidelines HTTP 500 on await fetch(\"http://10.10.10.1:5050/api/review-guidelines\", {
    \"credentials\": \"omit\",
    \"headers\": {
        \"User-Agent\": \"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:148.0) Gecko/20100101 Firefox/148.0\",
        \"Accept\": \"application/json\",
        \"Accept-Language\": \"en-US,en;q=0.9\",
        \"content-type\": \"application/json\",
        \"Priority\": \"u=0\"
    },
    \"referrer\": \"http://10.10.10.1:5050/annotations/guidelines/new?runId=mail-triage-v1-20260403\",
    \"body\": \"{\\\"slug\\\":\\\"foobar\\\",\\\"title\\\":\\\"Foobar\\\",\\\"scopeKind\\\":\\\"global\\\",\\\"bodyMarkdown\\\":\\\"kjsdflakjsdhflkasjdf\\\"}\",
    \"method\": \"POST\",
    \"mode\": \"cors\"
});

---

Is it because migrations are not applied / we don't have migrations?"

**Assistant interpretation:** Diagnose why a long-lived sqlite mirror database still lacks the review tables and land the real schema-versioning fix rather than a one-off manual workaround.

**Inferred user intent:** Make existing local review/test databases self-heal on startup so guideline creation works without manual SQL patching.

**Commit (code):** `252c9fc` — `Mirror: upgrade legacy review schema to v4`

### What I did
- Verified the failing merged sqlite DB existed, reported `schema_version=3`, and still had no `review_guidelines` / `review_feedback` / `run_guideline_links` tables.
- Read the sqlite server bootstrap path in `cmd/smailnail/commands/sqlite/serve.go`, `pkg/mirror/store.go`, and `pkg/mirror/schema.go`.
- Split the mirror schema migrations so:
  - version 3 now applies only `annotate.SchemaMigrationV3CoreStatements()`;
  - version 4 applies `annotate.SchemaMigrationV4Statements()`.
- Bumped `currentSchemaVersion` to `4` and taught `schemaVersion(...)` to parse version `4`.
- Simplified `annotate.SchemaMigrationV3Statements()` so its name again matches its contents instead of silently returning v4 tables too.
- Added a regression test in `pkg/mirror/store_test.go` that seeds a legacy version-3 database and proves bootstrap upgrades it to version 4 with the review tables present.
- Ran a copied-db smoke test by copying `/home/manuel/smailnail/smailnail-last-24-months-merged.sqlite` into repo-local `tmp/`, bootstrapping it through `mirror.OpenStore(...).Bootstrap(...)`, and verifying the copied DB reached schema version 4 and gained the missing review tables.

### Why
- The bug was not that migrations were absent; it was that the schema version number did not advance when new review tables were added to the migration helper.
- Without a version bump, any long-lived database already marked version 3 would silently skip the new review schema forever.

### What worked
- The version split cleanly expresses the historical reality: annotation MVP tables are schema version 3; review/guideline tables are schema version 4.
- The regression test protects the exact failure mode the user hit.
- The copied-db smoke test confirmed that the real class of local database now upgrades automatically.

### What didn't work
- My first smoke-test attempt used `timeout 8s go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve ...` against a copied DB and then immediately inspected it with `sqlite3`. That left the database locked when I probed it:

```text
Error: in prepare, database is locked (5)
```

- I switched to a tiny one-off Go bootstrap program that opened the copied DB, ran `mirror.OpenStore(...).Bootstrap(...)`, and exited cleanly before inspecting the schema.

### What I learned
- Embedding newer DDL into an old migration helper is harmless for fresh databases but dangerous for already-versioned databases.
- A focused upgrade test is worth far more here than relying on clean-database bootstrap tests alone.

### What was tricky to build
- The subtle part was recognizing that `annotate.SchemaMigrationV3Statements()` had become historically inaccurate. Fresh DBs still worked, which hid the fact that older DBs already stamped at version 3 would never see the newer tables.
- The other tricky part was validating the real-world case without mutating the user’s actual database, which is why I used a copied DB for the smoke test.

### What warrants a second pair of eyes
- The merge/bootstrap path in `pkg/mirror/merge.go` also keys behavior off `currentSchemaVersion`; reviewers should confirm that moving the repo to schema version 4 is acceptable for all existing merge workflows.

### What should be done in the future
- Add one more focused integration-style test around `cmd/smailnail sqlite serve` or a higher-level bootstrap wrapper if this path changes again.

### Code review instructions
- Start with:
  - `pkg/mirror/schema.go`
  - `pkg/mirror/store_test.go`
  - `pkg/annotate/schema.go`
- Validate with:
  - `go test -tags sqlite_fts5 ./pkg/mirror -count=1`

### Technical details
- Diagnosis commands:

```bash
cd smailnail
sqlite3 /home/manuel/smailnail/smailnail-last-24-months-merged.sqlite "SELECT value FROM mirror_metadata WHERE key='schema_version';"
sqlite3 /home/manuel/smailnail/smailnail-last-24-months-merged.sqlite ".tables" | tr ' ' '\n' | rg '^review_|^run_guideline_links$'
```

- Validation commands:

```bash
gofmt -w pkg/mirror/schema.go pkg/mirror/store_test.go pkg/annotate/schema.go
go test -tags sqlite_fts5 ./pkg/mirror -count=1
```

- Copied-db smoke test result after the fix:

```text
schema=4 fts=true status=available
schema_version:
value
4
tables:
review_guidelines
run_guideline_links
review_feedback
review_feedback_targets
```

## Related

- Design doc: `../design-doc/01-intern-guide-and-independent-code-review-of-the-review-ui-branch.md`
