---
Title: Add a mirror shard merge verb for month-sharded backfills
Ticket: SMN-20260402-MIRROR-MERGE
Status: active
Topics:
    - mirror
    - sqlite
    - backfill
    - cli
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Ticket for designing a first-class smailnail verb that merges month-sharded mirror databases and raw-message trees into one local mirror."
LastUpdated: 2026-04-02T16:14:59.682149727-04:00
WhatFor: "Track analysis and implementation planning for a first-class smailnail verb that merges month-sharded mirror databases and raw-message trees into one usable local mirror."
WhenToUse: "Use when designing or implementing the merge step that follows parallel month-sharded IMAP backfills."
---

# Add a mirror shard merge verb for month-sharded backfills

## Overview

This ticket captures the design and implementation plan for a new `smailnail` verb that merges multiple shard-local mirror databases into a single destination mirror. The current month-sharded backfill workflow writes one `mirror.sqlite` and one `raw/` tree per shard; the missing product capability is a safe, testable, first-class merge step that can convert those shards into one incremental-friendly local mirror.

The primary deliverable is an intern-facing design guide that explains the current mirror architecture, the constraints created by the existing schema and raw-file layout, the recommended merge algorithm, and the implementation/testing plan. The guide is intentionally detailed enough that a new engineer can implement the verb without having to rediscover core invariants.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Primary Design Doc**: [design-doc/01-analysis-design-and-implementation-plan-for-a-mirror-shard-merge-verb.md](./design-doc/01-analysis-design-and-implementation-plan-for-a-mirror-shard-merge-verb.md)
- **Diary**: [reference/01-investigation-diary.md](./reference/01-investigation-diary.md)

## Status

Current status: **active**

## Topics

- mirror
- sqlite
- backfill
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
