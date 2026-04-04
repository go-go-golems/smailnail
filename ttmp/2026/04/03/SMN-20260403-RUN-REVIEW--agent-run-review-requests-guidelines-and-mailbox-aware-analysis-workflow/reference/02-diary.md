---
Title: "Investigation Diary For Run Review Ticket"
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - diary
    - annotations
    - sqlite
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Chronological notes captured while opening the ticket and shaping the implementation plan."
LastUpdated: 2026-04-03T12:15:00-04:00
WhatFor: "Preserve design reasoning and discovery notes."
WhenToUse: "Use when reviewing why the ticket was scoped this way."
---

# Investigation Diary

## 2026-04-03

- Re-read the original annotation UI backend spec to anchor this ticket in the existing sqlite architecture rather than the older `smailnaild` assumptions.
- Confirmed that the current React API contract only supports state transitions for review. There is no reviewer comment payload in `ui/src/api/annotations.ts`.
- Confirmed that the current review queue and run detail pages expose approve/dismiss actions but no text-entry flow for reviewer correction requests.
- Confirmed that mailbox is already present in the mirror storage schema as `messages.mailbox_name`; the problem is not missing storage but missing end-to-end product surfacing and provenance guidance.
- Confirmed that sender detail already joins annotation and message-preview data, which makes it a strong pattern reference for future review-feedback endpoints.
- Decided that reviewer feedback should be modeled separately from agent/system logs so future queries can distinguish human correction from agent narration.
- Decided to keep this ticket functionality-first and explicitly avoid committing to pixel-level screen design.
- Added detailed ticket docs, related-file links, and task breakdowns, then ran `docmgr doctor --ticket SMN-20260403-RUN-REVIEW`, which passed.
- Attempted reMarkable bundle upload via `remarquee upload bundle ... --remote-dir /ai/2026/04/03/SMN-20260403-RUN-REVIEW --toc-depth 2` after a successful dry-run. The live upload failed with `dial tcp [2600:1901:0:4019::]:443: connect: network is unreachable`, so the ticket docs are ready locally but not confirmed on the device.
