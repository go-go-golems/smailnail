---
Title: "Current System Map For Run Review Ticket"
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/api/annotations.ts"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/types/annotations.ts"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/repository.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotationui/server.go"
ExternalSources: []
Summary: "Short orientation document mapping the current code paths that matter for reviewer feedback and guideline features."
LastUpdated: 2026-04-03T12:15:00-04:00
WhatFor: "Use this as the shortest path to codebase orientation before implementation."
WhenToUse: "Use at the start of implementation or code review."
---

# Current System Map

## Frontend

- `ui/src/api/annotations.ts`
  - Defines the current API surface
  - Review mutations only carry `reviewState`
- `ui/src/types/annotations.ts`
  - Defines current annotation, run, sender, and message preview DTOs
  - `MessagePreview` currently lacks `mailboxName`
- `ui/src/pages/ReviewQueuePage.tsx`
  - Current single-item and batch review interaction
- `ui/src/pages/RunDetailPage.tsx`
  - Current run aggregation page with approve-all path
- `ui/src/pages/SenderDetailPage.tsx`
  - Example of mixing annotation data, logs, and recent message previews

## Backend

- `pkg/annotationui/server.go`
  - Registers routes and serves the sqlite SPA
- `pkg/annotationui/handlers_senders.go`
  - Good reference for composite endpoint assembly
- `pkg/annotate/repository.go`
  - Existing source of truth for annotation CRUD and run aggregation
- `pkg/annotate/schema.go`
  - Current annotation-related schema

## Mirror And Enrichment

- `pkg/mirror/schema.go`
  - `messages.mailbox_name`
  - `messages.search_text`
- `pkg/mirror/service.go`
  - Populates mirror records from IMAP
- `pkg/enrich/schema.go`
  - Adds `thread_id`, `sender_email`, `sender_domain`, and the `threads` / `senders` tables

## Key Gaps

- No first-class review feedback records
- No first-class guidelines table
- No run-guideline linking model
- No clear mailbox-aware review contract at the UI/API level
