---
Title: Investigate messages_fts synchronization on message upserts
Ticket: SMN-20260403-FTS-SYNC
Status: active
Topics:
    - mirror
    - sqlite
    - fts
    - bug-investigation
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/mirror/merge.go
      Note: Merge path skips per-row FTS
    - Path: pkg/mirror/schema.go
      Note: Standalone FTS5 table bootstrap (no content= sync) at L142
    - Path: pkg/mirror/service.go
      Note: Main sync path (correct FTS at L678-683)
    - Path: pkg/mirror/service_test.go
      Note: Existing FTS sync test passes; UIDVALIDITY reset test missing FTS assertion
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-03T07:40:59.955093417-04:00
WhatFor: ""
WhenToUse: ""
---


# Investigate messages_fts synchronization on message upserts

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- mirror
- sqlite
- fts
- bug-investigation

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
