---
Title: Web UI for browsing annotations, review workflow, and managed SQL queries
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - sqlite
    - ux-design
    - react
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../Downloads/smailnail-review-ui.jsx
      Note: Imported JSX design sketch from external designer
    - Path: pkg/annotate/repository.go
      Note: CRUD repository with filter-based listing
    - Path: pkg/annotate/schema.go
      Note: Schema V3 — annotations
    - Path: pkg/annotate/types.go
      Note: Annotation
    - Path: pkg/enrich/schema.go
      Note: Threads and senders enrichment tables
    - Path: pkg/mirror/schema.go
      Note: Messages table
    - Path: pkg/smailnaild/http.go
      Note: Existing HTTP server to extend with annotation/query APIs
    - Path: ui/src/App.tsx
      Note: Existing React SPA shell to extend
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-03T07:51:48.0766413-04:00
WhatFor: ""
WhenToUse: ""
---



# Web UI for browsing annotations, review workflow, and managed SQL queries

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- frontend
- annotations
- sqlite
- ux-design
- react

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
