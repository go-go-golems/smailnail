---
Title: SQLite backend for annotation UI and query workbench
Ticket: SMN-20260403-ANNOTATION-BACKEND
Status: active
Topics:
    - backend
    - annotations
    - sqlite
    - api
    - cli
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-03T09:43:50.960588076-04:00
WhatFor: "Track implementation of the sqlite-backed annotation UI backend and query workbench server"
WhenToUse: "Use when reviewing backend work for the annotation UI or finding the implementation plan, diary, and task checklist"
---

# SQLite backend for annotation UI and query workbench

## Overview

This ticket implements the backend for the new annotation UI against the mirror sqlite database. The main architectural correction is that this work belongs under `smailnail sqlite serve`, not `smailnaild`, because it is about browsing annotations, senders, and read-only SQL over a local mirror rather than managing hosted user credentials and rules.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- backend
- annotations
- sqlite
- api
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
