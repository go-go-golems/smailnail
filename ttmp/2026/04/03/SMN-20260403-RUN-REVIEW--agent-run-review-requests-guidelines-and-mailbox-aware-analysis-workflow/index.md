---
Title: Agent-run review requests, guidelines, and mailbox-aware analysis workflow
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - sqlite
    - workflow
    - email
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotate/repository.go
      Note: Current annotation repository surface
    - Path: pkg/annotate/schema.go
      Note: Current annotation schema baseline
    - Path: pkg/annotationui/server.go
      Note: Current sqlite server and route registration
    - Path: pkg/mirror/schema.go
      Note: Storage-level mailbox support
    - Path: pkg/mirror/service.go
      Note: Mailbox provenance during mirror sync
    - Path: ui/src/api/annotations.ts
      Note: Current frontend API contract for review and run pages
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Current reviewer interaction surface
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Current run detail interaction surface
ExternalSources:
    - /home/manuel/code/wesen/corporate-headquarters/smailnail/ttmp/2026/04/03/SMN-20260403-ANNOTATION-UI--web-ui-for-browsing-annotations-review-workflow-and-managed-sql-queries/design/08-backend-api-specification-for-annotation-ui.md
Summary: Functionality-first ticket for adding reviewer-authored requests/comments, DB-backed review guidelines linked to runs, and mailbox-aware review context across the sqlite annotation UI.
LastUpdated: 2026-04-03T12:15:00-04:00
WhatFor: Use this ticket to implement the next layer of human-in-the-loop review functionality on top of the sqlite annotation UI.
WhenToUse: Use when extending run review, rejection workflows, reviewer guidance capture, or mailbox-aware analysis context.
---


# Agent-run review requests, guidelines, and mailbox-aware analysis workflow

This ticket is about turning the current annotation UI from a binary approval surface into a real reviewer-to-agent workflow. Today the reviewer can approve or dismiss annotations, but cannot explain why, cannot issue a structured correction request, cannot maintain DB-backed review instructions, and cannot reliably use mailbox as a first-class part of analysis context in the UI contract.

The primary implementation guide is [design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md](./design/01-agent-run-review-guidelines-and-mailbox-implementation-guide.md). That document is written for a new intern and explains the current system, the product intent, the proposed schema and API changes, the UX affordances, and an implementation sequence.

Use [reference/01-current-system-map.md](./reference/01-current-system-map.md) to understand which existing packages already contain the needed concepts, and [reference/02-diary.md](./reference/02-diary.md) for the investigation narrative and design decisions made while opening this ticket.

The actionable checklist lives in [tasks.md](./tasks.md). The changelog for ticket-level decisions lives in [changelog.md](./changelog.md).
