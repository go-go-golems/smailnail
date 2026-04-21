---
Title: "Current System Map For Inbox Browser Ticket (Updated 2026-04-08)"
Ticket: SMN-20260403-INBOX-BROWSER
Status: active
Topics:
    - email
    - sqlite
    - frontend
    - backend
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/schema.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/enrich/schema.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/enrich/threads.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotationui/handlers_senders.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/http.go"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/pages/SenderDetailPage.tsx"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/features/mailbox/MailboxExplorer.tsx"
ExternalSources: []
Summary: "Updated codebase orientation document reflecting implementation progress as of 2026-04-08."
LastUpdated: 2026-04-08T00:00:00-04:00
WhatFor: "Use this to orient implementation of inbox-style browsing."
WhenToUse: "Use at the start of implementation or code review."
---

# Current System Map (Updated 2026-04-08)

## Overview

The system has **two separate API surfaces** that handle message data:

1. **smailnaild** (`:8580`) — Account/mailbox/message browsing via IMAP (live or mirrored)
2. **annotationui** (`:8080`) — SQLite-based annotation UI with sender browsing

Both use the same underlying SQLite mirror database but expose different slices of functionality.

## Data Layer

### Mirror Schema (`pkg/mirror/schema.go`)

- Canonical mirrored message storage
- Includes `mailbox_name` and `search_text`
- Includes `messages_fts` for FTS5-backed searching

### Enrichment Schema (`pkg/enrich/schema.go`) ✅ COMPLETED

- `messages.thread_id` — Thread identifier for each message
- `messages.thread_depth` — Depth in thread hierarchy
- `threads` table — Thread summaries with:
  - thread_id, subject, mailbox_name
  - message_count, participant_count
  - first_sent_date, last_sent_date
- `senders` table — Sender profiles with:
  - email, display_name, domain
  - msg_count, first_seen_date, last_seen_date
  - unsubscribe_mailto, unsubscribe_http

### Thread Reconstruction (`pkg/enrich/threads.go`) ✅ COMPLETED

- `ThreadEnricher` reconstructs thread structure from message IDs
- Writes thread summaries to `threads` table
- Groups messages by thread_id

## Backend APIs

### smailnaild HTTP API (`pkg/smailnaild/http.go`) ✅ COMPLETED

Routes for account-based message browsing:

```
GET  /api/accounts
GET  /api/accounts/{id}
POST /api/accounts
PATCH /api/accounts/{id}
DELETE /api/accounts/{id}
GET  /api/accounts/{id}/mailboxes     ✅
GET  /api/accounts/{id}/messages       ✅
GET  /api/accounts/{id}/messages/{uid} ✅
```

Message browsing is account-centric (works with IMAP directly or mirror).

`ListMessagesInput` supports:
- `mailbox`, `limit`, `offset`
- `query` (search), `unreadOnly`
- `includeContent`, `contentType`

### annotationui HTTP API (`pkg/annotationui/server.go`) ✅ COMPLETED

Routes for SQLite annotation browsing:

```
GET  /api/mirror/senders              ✅
GET  /api/mirror/senders/{email}      ✅
GET  /api/mirror/senders/{email}/guidelines ✅
```

Sender endpoints return annotation counts, recent messages (limited to 20), annotations, and logs.

## Frontend Components

### Message Browsing (`ui/src/features/mailbox/`) ✅ COMPLETED

- `MailboxExplorer.tsx` — Main explorer with sidebar + list + detail layout
- `MailboxSidebar.tsx` — Mailbox navigation sidebar
- `MessageList.tsx` — Reusable message list with pagination
- `MessageDetail.tsx` — Message detail view with MIME structure

Uses Redux slice `mailboxSlice` with:
- `fetchMailboxes`, `fetchMessages`, `fetchMessageDetail` thunks
- Pagination via offset/limit

### Sender Detail (`ui/src/pages/SenderDetailPage.tsx`) ✅ COMPLETED

- Sender profile card with stats
- Annotations table
- Agent reasoning panel
- Recent messages (limited to 20)
- Guideline panel

Uses `useGetSenderQuery` and `useGetSenderGuidelinesQuery` hooks.

### API Client (`ui/src/api/client.ts`) ✅ COMPLETED

Methods for smailnaild:
- `listAccounts`, `getAccount`, `createAccount`, etc.
- `listMailboxes(accountId)`
- `listMessages(accountId, params)`
- `getMessage(accountId, mailbox, uid)`

## Key Gaps (What's Still Needed)

1. **No thread browsing API**
   - Threads data exists in schema but no `/api/mirror/threads` endpoint
   - No thread list/detail frontend views

2. **No timeline endpoint**
   - No `/api/mirror/timeline` for time-based aggregation

3. **Limited sender message browsing**
   - Sender detail only shows 20 recent messages
   - No pagination or full-text search for sender messages

4. **No reusable embeddable views**
   - MessageList/MessageDetail exist but tied to account-based API
   - No views that work across both smailnaild and annotationui contexts

5. **No unified filter model**
   - Different query params across endpoints
   - No shared time-range filter across messages/annotations/threads

6. **No annotation counts in message list**
   - Message rows don't show annotation counts
   - Can't filter messages by annotation state

## Architecture Notes

### Two API Surfaces

| Aspect | smailnaild | annotationui |
|--------|-----------|--------------|
| Port | 8580 | 8080 |
| Source | IMAP (live or mirror) | SQLite |
| Focus | Account-centric browsing | Annotation-centric browsing |
| Message list | Full pagination | 20 recent only |
| Thread support | IMAP-native (no thread_id) | SQLite (has thread_id) |
| Search | IMAP SEARCH | FTS available but not wired |

### Thread Data Flow

```
IMAP Mirror → enrich.ThreadEnricher → threads table
                                   ↓
                              messages table
                              (thread_id, thread_depth)

annotationui API → reads threads table → (no endpoint yet!)
```

### Next Implementation Steps

1. **Add `/api/mirror/threads` endpoints** to annotationui
   - Thread list with annotation counts
   - Thread detail with all messages
   - Thread search by subject/participants

2. **Add `/api/mirror/timeline` endpoint**
   - Time-bucketed aggregation
   - Works across messages and annotations

3. **Extend sender message browsing**
   - Pagination beyond 20
   - Full-text search via messages_fts

4. **Create reusable thread views**
   - ThreadList, ThreadDetail components
   - Embeddable in sender detail, tag detail, etc.

5. **Add shared filter model**
   - Time range, sender, tag, annotation state
   - Reusable across all views
