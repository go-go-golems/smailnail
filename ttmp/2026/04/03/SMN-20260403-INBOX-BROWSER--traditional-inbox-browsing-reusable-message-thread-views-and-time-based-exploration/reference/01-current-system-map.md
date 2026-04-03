---
Title: "Current System Map For Inbox Browser Ticket"
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
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/pages/SenderDetailPage.tsx"
ExternalSources: []
Summary: "Short orientation document mapping the code that already supports message, sender, and thread exploration."
LastUpdated: 2026-04-03T12:15:00-04:00
WhatFor: "Use this to orient implementation of inbox-style browsing."
WhenToUse: "Use at the start of implementation or code review."
---

# Current System Map

## Data

- `pkg/mirror/schema.go`
  - canonical mirrored message storage
  - includes `mailbox_name` and `search_text`
  - includes `messages_fts`
- `pkg/enrich/schema.go`
  - sender and thread schema additions
- `pkg/enrich/threads.go`
  - reconstructs thread structure and writes `threads`

## Backend

- `pkg/annotationui/server.go`
  - current sqlite route registration
- `pkg/annotationui/handlers_senders.go`
  - existing sender list/detail handlers with recent message previews

## Frontend

- `ui/src/pages/SendersPage.tsx`
  - basic sender table
- `ui/src/pages/SenderDetailPage.tsx`
  - sender detail plus recent messages
- `ui/src/types/annotations.ts`
  - current message preview DTO is intentionally small

## Key Gaps

- no direct message browser
- no direct thread browser
- limited sender filtering
- no time-based browsing surface spanning annotations and messages
- no shared filter/view model for embedded browsing contexts
