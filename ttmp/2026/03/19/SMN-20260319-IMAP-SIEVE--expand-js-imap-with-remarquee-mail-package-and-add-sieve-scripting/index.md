---
Title: Expand JS IMAP with remarquee mail package and add sieve scripting
Ticket: SMN-20260319-IMAP-SIEVE
Status: active
Topics:
    - imap
    - javascript
    - sieve
    - email
    - mcp
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/js/modules/smailnail/module.go
      Note: Existing JS module that will be expanded
    - Path: pkg/mcp/imapjs/identity_middleware.go
      Note: Current secure MCP account-resolution path
    - Path: pkg/services/smailnailjs/service.go
      Note: Current service layer and future orchestration point
    - Path: pkg/smailnaild/db.go
      Note: Schema migration anchor for future sieve fields
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-19T11:23:39.356564035-04:00
WhatFor: ""
WhenToUse: ""
---



# Expand JS IMAP with remarquee mail package and add sieve scripting

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- imap
- javascript
- sieve
- email
- mcp

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
