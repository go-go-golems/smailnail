---
Title: Human and LLM annotation system for email triage, safety, and bulk actions
Ticket: SMN-20260402-ANNOTATION-DESIGN
Status: active
Topics:
    - email
    - sqlite
    - cli
    - backend
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Design ticket for adding human and LLM annotations, review queues, and safe bulk-action workflows on top of the mirrored and enriched email corpus."
LastUpdated: 2026-04-02T14:32:00.000751436-04:00
WhatFor: "Use this ticket to understand the recommended architecture, data model, workflows, and rollout plan for a safety-first annotation and bulk-triage subsystem."
WhenToUse: "Read this before implementing annotation storage, LLM review tooling, or any bulk action mechanism that must respect human override and non-destructive review."
---

# Human and LLM annotation system for email triage, safety, and bulk actions

## Overview

This ticket designs an annotation subsystem for `smailnail` that lets both humans and LLMs attach structured judgments, free-form notes, and safe action proposals to messages, threads, senders, domains, mailboxes, and accounts. The design is intentionally safety-first: LLMs can suggest, score, and cluster, but humans retain final authority over any effective judgment that can influence bulk actions.

The recommended design is centered on four layers:

- annotation targets, such as a message, sender, thread, or domain
- observations, which are claims or notes produced by a human, LLM, or heuristic
- decisions, which resolve conflicting observations into the effective state
- proposal batches, which package destructive or high-impact actions for human review

The detailed design is in `design-doc/01-analysis-design-and-implementation-guide-for-annotations-review-and-safe-bulk-actions.md`, with supporting investigation notes in `reference/01-investigation-diary.md`.

There is also a lighter-weight implementation path in `design-doc/02-mvp-fast-path-annotations-groups-and-logs.md`. That document is the recommended starting point if the immediate goal is to ship a practical annotation system quickly without building the full long-term architecture first.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- email
- sqlite
- cli
- backend

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
