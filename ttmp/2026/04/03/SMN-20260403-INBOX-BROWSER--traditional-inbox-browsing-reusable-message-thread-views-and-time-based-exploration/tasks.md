---
Title: "SMN-20260403-INBOX-BROWSER Tasks"
Ticket: SMN-20260403-INBOX-BROWSER
Status: active
Topics:
    - email
    - frontend
    - backend
    - sqlite
    - search
    - ux-design
DocType: tasks
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/annotationui/handlers_senders.go
      Note: Sender browsing handlers (COMPLETED)
    - Path: pkg/enrich/schema.go
      Note: Thread/sender schema (COMPLETED)
    - Path: pkg/enrich/threads.go
      Note: ThreadEnricher implementation (COMPLETED)
    - Path: ui/src/features/mailbox/MailboxExplorer.tsx
      Note: MailboxExplorer with MessageList/MessageDetail (COMPLETED)
    - Path: ui/src/pages/SenderDetailPage.tsx
      Note: Sender detail with recent messages (COMPLETED)
ExternalSources: []
Summary: Implementation checklist for inbox browsing ticket, updated 2026-04-08 to reflect current state.
LastUpdated: 2026-04-08T00:00:00-04:00
WhatFor: Track progress on inbox browsing implementation
WhenToUse: Use to track and update task status during implementation
---

# Tasks

## ✅ Completed

### Backend Query Surface

- [x] **Add list/detail endpoints for messages** — Implemented in smailnaild: `/api/accounts/{id}/mailboxes`, `/api/accounts/{id}/messages`, `/api/accounts/{id}/messages/{uid}`
- [x] **Extend sender endpoints with search and richer filters** — Implemented in annotationui: `/api/mirror/senders`, `/api/mirror/senders/{email}`

### Frontend View Primitives

- [x] **Create an embeddable message list view** — Implemented: `ui/src/features/mailbox/MessageList.tsx`
- [x] **Create an embeddable message detail view** — Implemented: `ui/src/features/mailbox/MessageDetail.tsx`
- [x] **Create mailbox sidebar and explorer** — Implemented: `ui/src/features/mailbox/MailboxExplorer.tsx`, `MailboxSidebar.tsx`

### Data Infrastructure

- [x] **Thread reconstruction logic** — Implemented: `pkg/enrich/threads.go` (ThreadEnricher)
- [x] **Thread and sender enrichment schema** — Implemented: `pkg/enrich/schema.go` (threads table, thread_id/thread_depth on messages, senders table)

## 🔄 In Progress

### Backend Query Surface

- [ ] **Add list/detail endpoints for threads** — NOT IMPLEMENTED: No `/api/mirror/threads` endpoints yet
- [ ] **Add shared filter parsing for mailbox, sender, tag, annotation state, and time range** — PARTIAL: smailnaild has some query params but no unified filter model
- [ ] **Add endpoints or query options for timeline-based aggregation** — NOT IMPLEMENTED: No `/api/mirror/timeline`

### Frontend View Primitives

- [ ] **Create an embeddable thread list view** — NOT IMPLEMENTED
- [ ] **Create an embeddable thread detail view** — NOT IMPLEMENTED
- [ ] **Create a sender-scoped message view** — PARTIAL: SenderDetailPage shows recentMessages but limited to 20, no pagination
- [ ] **Create a tag-scoped message view** — NOT IMPLEMENTED
- [ ] **Create a time-range exploration view for messages and annotations** — NOT IMPLEMENTED

### Search And Filtering

- [ ] **Add sender search text input and domain/mailbox filters** — PARTIAL: annotationui senders endpoint has domain/tag filters but no full-text search
- [ ] **Add message search using FTS-backed search text where available** — PARTIAL: smailnaild has `query` param but FTS not wired in annotationui
- [ ] **Add thread search by subject, participants, and annotation state** — NOT IMPLEMENTED
- [ ] **Add time-range filters that work consistently across message, sender, thread, and annotation contexts** — NOT IMPLEMENTED
- [ ] **Add filter chips or summary state so embedded views stay understandable when reused** — NOT IMPLEMENTED

### Annotation Integration

- [ ] **Show annotation counts in message/thread list rows where useful** — NOT IMPLEMENTED
- [ ] **Allow browsing all messages for a sender from sender detail** — PARTIAL: Shows 20 recent messages only
- [ ] **Allow browsing all messages for a tag from an annotation/tag context** — NOT IMPLEMENTED
- [ ] **Allow viewing annotations and messages together over a time period** — NOT IMPLEMENTED

### Validation

- [ ] **Add backend tests for list/detail/search behavior** — TODO
- [ ] **Add frontend tests for reusable view composition** — TODO
- [ ] **Add a manual playbook for browsing sender, tag, and time-slice contexts** — NOT IMPLEMENTED

### Documentation

- [ ] **Keep the diary updated during implementation** — PARTIAL: Updated 2026-04-08
- [ ] **Relate core backend and frontend files to the focused design doc** — PARTIAL: Updated in design doc
- [ ] **Run `docmgr doctor --ticket SMN-20260403-INBOX-BROWSER`** — PENDING
- [ ] **Upload the document bundle to reMarkable after the guide is finalized** — PENDING

## 📋 Implementation Notes (2026-04-08)

### What Exists

1. **Two separate API surfaces:**
   - `smailnaild` (`:8580`): Account/mailbox/message browsing via IMAP (live or mirrored)
   - `annotationui` (`:8080`): SQLite-based annotation UI with sender browsing

2. **Message browsing is account-centric:**
   - Works with IMAP directly or mirror
   - `MessageView` includes UID, subject, from/to addresses, date, flags, size, mimeParts
   - No thread_id in the API response currently

3. **Thread data exists but not exposed:**
   - `pkg/enrich/schema.go`: threads table with thread_id, subject, mailbox_name, message_count, participant_count, first/last_sent_date
   - `pkg/enrich/threads.go`: ThreadEnricher reconstructs threads
   - But no `/api/mirror/threads` endpoint

4. **Sender detail is annotation-centric:**
   - Shows sender info, annotations, logs, recentMessages (20 max)
   - No pagination for messages
   - No full-text search on messages

### What's Needed

1. **Thread API** (`/api/mirror/threads`, `/api/mirror/threads/{threadId}`)
2. **Timeline endpoint** for time-based aggregation
3. **Extended message browsing** with:
   - Pagination beyond 20 messages per sender
   - FTS-backed search in annotationui
   - Annotation counts in list rows
4. **Reusable embeddable views** that work in multiple contexts
5. **Time-range filters** that work consistently across all views

### Architecture Considerations

- The `MailboxExplorer` is already reusable but tied to the smailnaild account-based API
- Need to consider whether thread browsing should live in smailnaild (IMAP-native) or annotationui (SQLite)
- Time-based exploration should probably use the annotationui SQLite layer for consistency with annotations
