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
LastUpdated: 2026-04-03T12:15:00-04:00
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
