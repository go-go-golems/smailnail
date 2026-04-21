---
Title: "Investigation Diary For Inbox Browser Ticket"
Ticket: SMN-20260403-INBOX-BROWSER
Status: active
Topics:
    - diary
    - email
    - sqlite
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Chronological notes captured while opening the inbox browser ticket."
LastUpdated: 2026-04-08T00:00:00-04:00
WhatFor: "Preserve why the ticket is scoped around functionality and reusable view affordances."
WhenToUse: "Use when reviewing or extending the ticket scope."
---

# Investigation Diary

## 2026-04-03

- Confirmed that the sqlite mirror already contains the raw ingredients for inbox browsing: messages, FTS search text, sender enrichment, and thread enrichment.
- Confirmed that the current UI only exposes a sender-centric slice of that data and does not yet provide first-class message or thread browsing.
- Noted that the user explicitly wants embeddable views rather than one fixed screen, so the ticket is framed around reusable list/detail/filter primitives.
- Decided to keep the document focused on UX intent and functional affordances instead of committing to visual design or exact layout.
- Identified time-based exploration as a cross-cutting concern that should ideally share one filter model rather than being bolted onto every page differently.
- Added detailed ticket docs, related-file links, and task breakdowns, then ran `docmgr doctor --ticket SMN-20260403-INBOX-BROWSER`, which passed.
- Attempted reMarkable bundle upload via `remarquee upload bundle ... --remote-dir /ai/2026/04/03/SMN-20260403-INBOX-BROWSER --toc-depth 2` after a successful dry-run. The live upload failed while creating the remote directory with `cannot parse rootIndex, cant parse line '543873959500', wrong number of fields 1`, so the ticket docs are ready locally but not confirmed on the device.

## 2026-04-08 — Codebase Review

Reviewed ticket against current codebase. Key findings:

### What's Implemented

1. **Backend sender endpoints** — `pkg/annotationui/handlers_senders.go`
   - `GET /api/mirror/senders` with domain/tag/hasAnnotations filters
   - `GET /api/mirror/senders/{email}` with recentMessages (20 max), annotations, logs
   - `GET /api/mirror/senders/{email}/guidelines`

2. **Mailbox/message browsing** — `pkg/smailnaild/http.go` + `pkg/smailnaild/accounts/`
   - `GET /api/accounts/{id}/mailboxes`
   - `GET /api/accounts/{id}/messages` with pagination, query, unreadOnly
   - `GET /api/accounts/{id}/messages/{uid}`

3. **Frontend mailbox explorer** — `ui/src/features/mailbox/`
   - `MailboxExplorer.tsx` — sidebar + list + detail layout
   - `MessageList.tsx` — reusable list with pagination
   - `MessageDetail.tsx` — message detail with MIME structure
   - Redux slice with fetchMailboxes/fetchMessages/fetchMessageDetail

4. **Schema infrastructure** — `pkg/enrich/schema.go`, `pkg/enrich/threads.go`
   - threads table with thread_id, subject, message_count, participant_count, etc.
   - ThreadEnricher reconstructs thread structure
   - senders table with display_name, domain, msg_count, etc.

### What's Missing

1. **Thread API endpoints** — No `/api/mirror/threads` routes
   - Need thread list with annotation counts
   - Need thread detail with all messages
   - Need thread search by subject/participants

2. **Timeline endpoint** — No `/api/mirror/timeline`
   - Time-bucketed aggregation across messages and annotations

3. **Extended sender browsing**
   - Pagination beyond 20 messages
   - Full-text search via messages_fts
   - Annotation counts in message rows

4. **Thread views**
   - ThreadList component
   - ThreadDetail component
   - Embeddable in sender/tag/time contexts

5. **Shared filter model**
   - Time range filters across all contexts
   - Filter chips for embedded views

### Architecture Notes

- Two API surfaces: smailnaild (IMAP-centric) vs annotationui (SQLite-centric)
- Thread data exists in SQLite but not exposed via API
- Message browsing works but tied to account-based model in smailnaild
- Consider whether threads should live in annotationui (SQLite) for consistency with annotations
