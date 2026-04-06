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
Summary: Implemented shared contract codegen for the annotation review feedback/guideline wire layer, driven by go generate and a repo-local Go generator command.
LastUpdated: 2026-04-06T21:20:00Z
WhatFor: Track the shared IDL/codegen implementation and resulting contract migration work.
WhenToUse: Start here to find the implementation plan, diary, task list, and validation notes.
---

# Introduce go-generate driven shared contract codegen for annotation feedback and guidelines

## Overview

This ticket implements a shared IDL and code generation pipeline for the annotation review feedback/guideline slice. The goal was to remove Go/TypeScript drift in the HTTP wire contract while keeping the current REST shape recognizable and integrating generation into the normal `go generate` workflow.

## Primary Documents

- [Implementation plan](./design-doc/01-implementation-plan-for-shared-feedback-and-guideline-contract-codegen.md)
- [Diary](./reference/01-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Current Status

- Schema/codegen plumbing: **done**
- Generated Go/TS contract outputs: **done**
- Backend migration to generated wire types: **done**
- Frontend migration to generated wire types: **done**
- Validation: **done**
- Focused git commit: **done** (`AnnotationUI: add shared feedback contract codegen`)

## Main Outcomes

- Added protobuf IDL for feedback/guideline/review-action payloads
- Added protobuf IDL for the rest of the annotation UI wire layer (annotations, groups, logs, runs, senders, info, query payloads)
- Added Go-command-driven `go generate` workflow for contract generation
- Generated and committed Go + TS contract code
- Standardized create-feedback payloads on `targets`
- Standardized list endpoints across the annotation UI contract on wrapper responses with `items`
- Switched backend feedback/guideline endpoints to generated wire types + `protojson`
- Switched backend annotation/list/detail/query endpoints to generated wire types + `protojson`
- Switched frontend type layer and RTK Query usage to generated contract types
- Updated mocks and stories to the same shared contract
- Added a `pkg/doc` playbook for future contract-codegen work
