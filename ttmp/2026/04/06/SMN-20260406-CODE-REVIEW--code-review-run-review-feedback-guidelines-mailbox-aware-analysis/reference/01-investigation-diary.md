---
Title: Investigation Diary
Ticket: SMN-20260406-CODE-REVIEW
Status: active
Topics:
    - code-review
    - annotations
    - backend
    - frontend
    - sqlite
    - react
    - workflow
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-06T15:35:23.632112074-04:00
WhatFor: ""
WhenToUse: ""
---

# Investigation Diary

## Goal

<!-- What is the purpose of this reference document? -->

## Context

<!-- Provide background context needed to use this reference -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->

## Step 1: Code Review Investigation

This step performed a comprehensive code review of the entire SMN-20260403-RUN-REVIEW feature branch against origin/main. The review covered 80 files, ~15,879 lines added, spanning the full stack from SQLite schema through Go backend to React/TypeScript frontend.

### What I did
- Read the full `git diff origin/main --stat` to understand the scope
- Read both design documents from the RUN-REVIEW ticket
- Read the full implementation diary (18 steps, ~1100 lines)
- Read every Go file: schema.go, types.go, repository_feedback.go, handlers_annotations.go, handlers_feedback.go, types_feedback.go, server.go, handlers_senders.go
- Read every TypeScript file: annotations.ts (RTK Query), reviewFeedback.ts, reviewGuideline.ts, annotationUiSlice.ts
- Read every component: ReviewCommentDrawer, ReviewCommentInline, FeedbackCard, GuidelinePicker, GuidelineLinkPicker, RunFeedbackSection, RunGuidelineSection, GuidelineCard, GuidelineForm, GuidelineSummaryCard, GuidelineLinkedRuns, MailboxBadge, AnnotationTable, AnnotationRow
- Read every page: ReviewQueuePage, RunDetailPage, SenderDetailPage, GuidelinesListPage, GuidelineDetailPage
- Read the enrich command changes
- Read mock data and MSW handlers
- Created docmgr ticket SMN-20260406-CODE-REVIEW with design doc and diary
- Wrote ~1,400-line comprehensive code review document

### Why
- User requested a detailed code review suitable for a new intern, covering confusing code, deprecated code, unused code, and unclear naming
- The review needed to be stored in a docmgr ticket and uploaded to reMarkable

### What worked
- Reading the design docs first provided the architectural context needed to evaluate whether the implementation matched the intent
- The diary was invaluable for understanding *why* certain decisions were made (e.g., the drawer→dialog conversion, the transaction recovery)
- Reading files in dependency order (schema → types → repository → handlers → API → components → pages) made the review efficient

### What didn't work
- N/A

### What I learned
- The most impactful finding is the transactional repository pattern — it was specifically recovered from a broken attempt and is now the strongest architectural element
- The most common category of issue is "half-implemented features" — code that is structurally correct but wired to empty data sources (GuidelineLinkedRuns always gets [], linkedRunCount is always 0)
- The second most common category is "dead state" — Redux fields that were added for future features but never wired up

### What was tricky to build
- Managing the doc file path carefully — case-sensitive directory names caused issues in earlier attempts

### What warrants a second pair of eyes
- The M1-M3 must-fix items (contract mismatch, dead code, dead Redux state)
- The N+1 query in ListReviewFeedback — is it actually a problem at current data volumes?

### What should be done in the future
- Upload the review doc to reMarkable
- Address M1-M3 before merge
- Address S1-S9 in follow-up tickets

### Code review instructions
- Start with the design doc: `ttmp/2026/04/06/SMN-20260406-CODE-REVIEW--code-review-run-review-feedback-guidelines-mailbox-aware-analysis/design-doc/01-comprehensive-code-review-run-review-feedback-guidelines-mailbox-aware-analysis.md`
