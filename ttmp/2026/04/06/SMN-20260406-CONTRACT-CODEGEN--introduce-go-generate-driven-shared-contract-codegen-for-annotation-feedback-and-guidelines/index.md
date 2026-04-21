---
Title: Introduce go-generate driven shared contract codegen for annotation feedback and guidelines
Ticket: SMN-20260406-CONTRACT-CODEGEN
Status: active
Topics:
    - annotations
    - backend
    - frontend
    - workflow
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Implemented shared contract codegen for the annotation UI wire layer and the hosted web API wire layer, driven by go generate and repo-local generator entrypoints.
LastUpdated: 2026-04-06T23:10:00Z
WhatFor: Track the shared IDL/codegen implementation, repo-wide contract spec, and resulting migration work.
WhenToUse: Start here to find the implementation plan, diary, task list, and validation notes.
---

# Introduce go-generate driven shared contract codegen for annotation feedback and guidelines

## Overview

This ticket implements a shared IDL and code generation pipeline first for the annotation UI wire layer and then for the hosted web API wire layer. The goal was to remove Go/TypeScript drift in the HTTP contract while keeping existing REST JSON shapes recognizable and integrating generation into the normal `go generate` workflow.

## Primary Documents

- [Implementation plan](./design-doc/01-implementation-plan-for-shared-feedback-and-guideline-contract-codegen.md)
- [Repo-wide wire contract unification spec](./design-doc/02-repo-wide-wire-contract-unification-spec.md)
- [Diary](./reference/01-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Current Status

- Schema/codegen plumbing: **done**
- Generated Go/TS contract outputs: **done**
- Backend migration to generated wire types: **done**
- Frontend migration to generated wire types: **done**
- Validation: **done**
- Focused implementation commits: **done**

## Main Outcomes

- Added protobuf IDL for feedback/guideline/review-action payloads
- Added protobuf IDL for the rest of the annotation UI wire layer (annotations, groups, logs, runs, senders, info, query payloads)
- Added protobuf IDL for the hosted web API (`/api/info`, `/api/me`, `/api/accounts/*`, `/api/rules/*`)
- Added Go-command-driven `go generate` workflow for contract generation
- Generated and committed Go + TS contract code across both API surfaces
- Standardized create-feedback payloads on `targets`
- Standardized list endpoints across the annotation UI contract on wrapper responses with `items`
- Switched backend feedback/guideline endpoints to generated wire types + `protojson`
- Switched backend annotation/list/detail/query endpoints to generated wire types + `protojson`
- Switched frontend type layer and RTK Query usage to generated contract types
- Switched hosted frontend client/types to generated contract types
- Updated mocks and stories to the same shared contract
- Added a repo-wide wire-contract unification spec
- Added a `pkg/doc` playbook for future contract-codegen work
