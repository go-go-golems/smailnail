---
Title: Add a glazed IMAP mirror verb with SQLite indexing
Ticket: SMN-20260401-IMAP-MIRROR
Status: active
Topics:
    - imap
    - sqlite
    - glazed
    - email
    - cli
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/smailnail/main.go
      Note: Current CLI root where the new glazed verb will be registered
    - Path: cmd/smailnail/commands/fetch_mail.go
      Note: Existing glazed IMAP verb that reuses the DSL builder and output processor
    - Path: pkg/mailruntime/imap_client.go
      Note: UID-oriented IMAP runtime that should drive durable mirror sync loops
    - Path: pkg/services/smailnailjs/service.go
      Note: Existing service abstraction that already wraps mailruntime and stored-account resolution
    - Path: pkg/smailnaild/db.go
      Note: Existing SQLite bootstrap pattern and proof that the repo already carries sqlx/sqlite schema migration utilities
    - Path: pkg/smailnaild/accounts/service.go
      Note: Current hosted mailbox preview flow and encrypted account-resolution logic
ExternalSources: []
Summary: Research ticket for adding a new glazed mirror verb that downloads IMAP mailboxes and imports searchable message data into SQLite without overloading the hosted application schema.
LastUpdated: 2026-04-01T17:55:00-04:00
WhatFor: Detailed architecture, tradeoff, and implementation guidance for an intern who will add a local mirror/indexing workflow to smailnail.
WhenToUse: Use this ticket when implementing or reviewing the local IMAP mirroring and SQLite search feature.
---

# Add a glazed IMAP mirror verb with SQLite indexing

## Overview

This ticket captures the design work for a new `smailnail mirror` verb. The goal is not only to download mail from IMAP, but to do it in a way that produces a durable local mirror and a searchable SQLite index that can be queried offline.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- [Design Guide](./design-doc/01-intern-guide-designing-an-imap-mirror-verb-with-sqlite-indexing.md)
- [Diary](./reference/01-diary.md)

## Status

Current status: **active**

## Topics

- imap
- sqlite
- glazed
- email
- cli

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
