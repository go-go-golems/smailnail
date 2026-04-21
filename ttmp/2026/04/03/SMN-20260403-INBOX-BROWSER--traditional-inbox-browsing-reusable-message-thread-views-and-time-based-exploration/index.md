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
      Note: Sender browsing handlers (COMPLETED)
    - Path: pkg/enrich/schema.go
      Note: Thread and sender enrichment schema (COMPLETED)
    - Path: pkg/enrich/threads.go
      Note: Thread reconstruction logic (COMPLETED)
    - Path: pkg/mirror/schema.go
      Note: Core message storage and FTS support
    - Path: pkg/smailnaild/http.go
      Note: Account/mailbox/message API (COMPLETED)
    - Path: ui/src/features/mailbox/MailboxExplorer.tsx
      Note: MailboxExplorer with MessageList/MessageDetail (COMPLETED)
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Sender detail with recent messages (COMPLETED)
    - Path: ui/src/pages/SendersPage.tsx
      Note: Current sender list UX
    - Path: ui/src/types/annotations.ts
      Note: Current frontend DTO constraints
ExternalSources:
    - /home/manuel/code/wesen/corporate-headquarters/smailnail/ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md
Summary: Ticket for inbox-style message/thread browsing. Progress: ✅ Backend sender endpoints, mailbox/message browsing, and schema infrastructure completed. Remaining: thread API, timeline endpoint, reusable thread views, time-based exploration, and unified filter model.
LastUpdated: 2026-04-08T00:00:00-04:00
WhatFor: Use this ticket to design and implement traditional inbox browsing on top of the sqlite mirror and enrich tables.
WhenToUse: Use when adding message/thread browsing, sender filtering, or time-based exploration features.
---


# Traditional inbox browsing, reusable message/thread views, and time-based exploration

This ticket covers inbox-style message/thread browsing. **Progress (2026-04-08):** Backend sender endpoints, mailbox/message browsing, and MailboxExplorer components are implemented. Thread data exists in the schema but thread API endpoints and views are not yet built. Time-based exploration and unified filter model remain.

The design guide is [design/01-inbox-browser-and-time-based-exploration-guide.md](./design/01-inbox-browser-and-time-based-exploration-guide.md) and the updated task list is in [tasks.md](./tasks.md).

The detailed guide is [design/01-inbox-browser-and-time-based-exploration-guide.md](./design/01-inbox-browser-and-time-based-exploration-guide.md). It explains the current mirror/enrichment system, the product goals, proposed backend APIs, reusable frontend view models, search affordances, and an implementation sequence.

Use [reference/01-current-system-map.md](./reference/01-current-system-map.md) for codebase orientation and [reference/02-diary.md](./reference/02-diary.md) for the reasoning that shaped the ticket.

The task list is in [tasks.md](./tasks.md), and the ticket-level decision log is in [changelog.md](./changelog.md).
