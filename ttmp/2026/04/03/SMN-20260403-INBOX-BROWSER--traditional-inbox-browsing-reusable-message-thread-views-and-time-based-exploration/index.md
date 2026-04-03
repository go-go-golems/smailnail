---
Title: Traditional inbox browsing, reusable message/thread views, and time-based exploration
Ticket: SMN-20260403-INBOX-BROWSER
Status: active
Topics:
    - email
    - frontend
    - backend
    - sqlite
    - search
    - ux-design
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotationui/handlers_senders.go
      Note: Existing sender browsing handler pattern
    - Path: pkg/enrich/schema.go
      Note: Thread and sender enrichment schema
    - Path: pkg/enrich/threads.go
      Note: Thread reconstruction logic
    - Path: pkg/mirror/schema.go
      Note: Core message storage and FTS support
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Current sender detail and recent message previews
    - Path: ui/src/pages/SendersPage.tsx
      Note: Current sender list UX
    - Path: ui/src/types/annotations.ts
      Note: Current frontend DTO constraints
ExternalSources:
    - /home/manuel/code/wesen/corporate-headquarters/smailnail/ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md
Summary: Functionality-first ticket for expanding the sqlite UI into reusable message/thread browsing flows with sender search, time filters, and email-client-style affordances.
LastUpdated: 2026-04-03T12:15:00-04:00
WhatFor: Use this ticket to design and implement traditional inbox browsing on top of the sqlite mirror and enrich tables.
WhenToUse: Use when adding message/thread browsing, sender filtering, or time-based exploration features.
---


# Traditional inbox browsing, reusable message/thread views, and time-based exploration

This ticket is about making the sqlite UI behave more like an actual mail analysis environment instead of only an annotation browser. The existing UI already exposes senders and recent message previews, which proves the data is present. The missing layer is a reusable browsing model for messages, threads, senders, tags, and time ranges.

The detailed guide is [design/01-inbox-browser-and-time-based-exploration-guide.md](./design/01-inbox-browser-and-time-based-exploration-guide.md). It explains the current mirror/enrichment system, the product goals, proposed backend APIs, reusable frontend view models, search affordances, and an implementation sequence.

Use [reference/01-current-system-map.md](./reference/01-current-system-map.md) for codebase orientation and [reference/02-diary.md](./reference/02-diary.md) for the reasoning that shaped the ticket.

The task list is in [tasks.md](./tasks.md), and the ticket-level decision log is in [changelog.md](./changelog.md).
