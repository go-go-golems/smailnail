---
Title: 'Intern guide: designing an IMAP mirror verb with SQLite indexing'
Ticket: SMN-20260401-IMAP-MIRROR
Status: active
Topics:
    - imap
    - sqlite
    - glazed
    - email
    - cli
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: smailnail/cmd/smailnail/commands/fetch_mail.go
      Note: |-
        Existing glazed IMAP command that shows the current CLI construction style
        Existing Glazed IMAP command style and transient fetch flow
    - Path: smailnail/cmd/smailnail/main.go
      Note: |-
        Root command registration point for the new mirror verb
        Root CLI registration point for the new mirror verb
    - Path: smailnail/pkg/dsl/actions.go
      Note: Current mail action layer for copy, move, delete, export, and flag mutation
    - Path: smailnail/pkg/dsl/processor.go
      Note: |-
        Current fetch pipeline and its sequence-number-oriented execution model
        Current DSL fetch pipeline and sequence-number-oriented pagination
    - Path: smailnail/pkg/imap/layer.go
      Note: |-
        Reusable IMAP Glazed section and low-level direct connection helper
        Reusable IMAP Glazed section and direct connection settings
    - Path: smailnail/pkg/js/modules/smailnail/module.go
      Note: Current JavaScript API surface, useful for understanding how richer IMAP operations are exposed today
    - Path: smailnail/pkg/mailruntime/imap_client.go
      Note: |-
        UID-oriented IMAP runtime that should be the foundation for incremental mirroring
        Recommended UID-based sync foundation for mirroring
    - Path: smailnail/pkg/mcp/imapjs/identity_middleware.go
      Note: |-
        Security boundary for resolving stored accounts inside MCP execution
        MCP stored-account security boundary that should remain separate from local mirror state
    - Path: smailnail/pkg/services/smailnailjs/service.go
      Note: |-
        Existing service abstraction over `mailruntime`, account resolution, and session lifecycle
        Existing higher-level runtime wrapper around mailruntime
    - Path: smailnail/pkg/smailnaild/accounts/service.go
      Note: |-
        Existing hosted account resolution, mailbox preview, and detail-fetch flows
        Hosted account resolution and mailbox preview implementation boundary
    - Path: smailnail/pkg/smailnaild/db.go
      Note: |-
        Existing SQLite bootstrap pattern and schema migration style inside this repository
        Existing SQLite bootstrap and migration pattern
    - Path: smailnail/pkg/smailnaild/http.go
      Note: Hosted API shape and current account/message read surfaces
ExternalSources: []
Summary: Detailed design and implementation guide for adding a local IMAP mirror verb that downloads messages and imports searchable data into SQLite while reusing smailnail's existing Glazed, IMAP, and account-resolution building blocks.
LastUpdated: 2026-04-01T17:55:00-04:00
WhatFor: Explain the current system, evaluate options, recommend an architecture, and provide a phased plan for implementing a robust local mirror and search index.
WhenToUse: Use this guide before implementing `smailnail mirror`, reviewing the design, or onboarding a new engineer to the feature.
---


# Intern guide: designing an IMAP mirror verb with SQLite indexing

## Executive Summary

`smailnail` already has the key raw ingredients for this feature, but not yet the exact workflow the user wants. The repository already contains Glazed-based CLI verbs, a mature IMAP rule engine, a newer UID-oriented `mailruntime` package, a JavaScript service layer built on that runtime, and SQLite bootstrapping code in the hosted backend. What it does not have is a durable local mirror primitive that can download mailboxes incrementally and turn them into an offline searchable store.

The most important design decision is this: the new mirror feature should be a new local CLI subsystem, not an extension of the hosted application database and not a disguised variation of `fetch-mail`. `fetch-mail` is optimized for transient output rows and DSL-generated previews. The hosted SQLite schema is optimized for accounts, tests, users, rules, and web sessions. A mirror/index feature needs a third thing: a dedicated local storage model with durable sync state, per-mailbox `UIDVALIDITY`, raw message persistence, and searchable projections.

The recommended architecture is:

1. Add a new Glazed verb, `smailnail mirror`.
2. Reuse the existing IMAP section and general command style from the current CLI.
3. Base the sync engine on `pkg/mailruntime/imap_client.go`, because it already uses UID-oriented search/fetch operations and supports raw message fetches, flag reads, copy/move/delete, append, and capability inspection.
4. Add a dedicated local mirror store package with its own SQLite bootstrap and migrations.
5. Persist raw `.eml` files to disk and import structured/searchable projections into SQLite.
6. Use `go-message/mail` to parse downloaded raw messages into searchable plain-text/HTML/header/address fields.
7. Keep the feature read-only against the remote server by default.

That gives the user a real mirror, not just a report. It also keeps the implementation aligned with the repository's existing structure and avoids the mistake of overloading the hosted application schema with local-cache concerns.

## Problem Statement And Scope

The requested feature is a new verb, built with Glazed and reusing as much of the existing codebase as possible, that can:

- connect to one or more IMAP mailboxes,
- download messages in a durable way,
- keep track of what has already been mirrored,
- import message data into SQLite so mail becomes easy to search locally,
- and do all of that in a way that is understandable to a new intern who did not build the current system.

This is not the same thing as the current `fetch-mail` command. `fetch-mail` emits Glazed rows for the current invocation. It does not create a durable store, does not preserve remote sync state, does not maintain a local mailbox model, and does not define a schema for offline search.

This is also not the same thing as the hosted application. The hosted backend already uses SQLite by default, but that SQLite database is application state for accounts, tests, rules, identities, and web sessions, not a local content-addressed mailbox mirror.

### In scope

- a new local Glazed command, likely `smailnail mirror`,
- incremental IMAP download based on durable sync state,
- raw message persistence,
- SQLite metadata and search import,
- intern-level documentation of the existing architecture and the new design,
- concrete file-level implementation guidance.

### Out of scope for the first slice

- replacing the existing YAML DSL,
- turning the hosted web app into a full mail client,
- bidirectional sync with local edits pushed back to IMAP,
- OAuth token refresh flows for new providers,
- attachment extraction into a second document store,
- Sieve synchronization as part of the first mirror implementation.

## Reader Orientation: What `smailnail` Is Today

`smailnail` is not one program. It is a small mail platform with multiple execution surfaces:

- a CLI for YAML-rule execution and direct fetches,
- a hosted backend with encrypted account storage and mailbox previews,
- an MCP server that executes JavaScript against the embedded `require("smailnail")` module,
- and a newer reusable IMAP runtime that the JS layer already builds on.

At a top level, the current system looks like this:

```text
user
 |
 +--> smailnail CLI
 |     |
 |     +--> fetch-mail / mail-rules Glazed commands
 |            |
 |            +--> DSL rule building and DSL fetch pipeline
 |
 +--> smailnaild hosted backend
 |     |
 |     +--> encrypted IMAP account storage in app DB
 |     +--> mailbox preview and message detail endpoints
 |
 +--> smailnail-imap-mcp
       |
       +--> JS runtime
              |
              +--> smailnailjs service
                     |
                     +--> mailruntime IMAP client
                     +--> stored-account resolver (when hosted/MCP-backed)
```

### Current CLI root

The root CLI currently registers two Glazed-backed verbs, `mail-rules` and `fetch-mail`, in `cmd/smailnail/main.go:18-79`. The Cobra root exists mostly to host Glazed-generated commands and help wiring. This is the correct place to register the new mirror verb.

### Current `fetch-mail` command

`cmd/smailnail/commands/fetch_mail.go:28-57` defines a settings struct that combines ad hoc search flags with embedded `imap.IMAPSettings`. `cmd/smailnail/commands/fetch_mail.go:59-203` builds a Glazed command description. `cmd/smailnail/commands/fetch_mail.go:206-442` decodes flags, builds a DSL rule, connects directly to IMAP, runs `rule.FetchMessages`, and prints rows.

This existing command is important because it shows the command-construction style to reuse. It is also important because it shows what not to overload: the command is designed around output rows, not around durable sync state.

### Current IMAP settings section

`pkg/imap/layer.go:13-88` defines `IMAPSettings`, `IMAPSectionSlug`, `NewIMAPSection()`, and `ConnectToIMAPServer()`. This gives us a reusable Glazed section for server, port, username, password, mailbox, and `--insecure`. It is the easiest piece to reuse directly in the new verb.

### Current DSL pipeline

The DSL fetch engine in `pkg/dsl/processor.go:49-320` is a multi-stage search/fetch/MIME-content pipeline. It:

1. builds IMAP search criteria,
2. executes search,
3. paginates using returned sequence numbers,
4. fetches metadata and body structure,
5. determines required MIME sections,
6. fetches MIME content in batches,
7. returns `EmailMessage` values.

The DSL action layer in `pkg/dsl/actions.go:16-238` performs copy, move, delete, flag mutation, and export after a fetch.

This subsystem is mature and worth preserving, but it is optimized for rule execution and preview. It is not a durable sync engine because it is centered on result sets and presentation, not on mailbox-local sync state.

### Current `mailruntime` subsystem

`pkg/mailruntime/imap_client.go:20-128` creates a higher-level `IMAPClient` wrapper around `go-imap/v2`. `pkg/mailruntime/imap_client.go:139-215` exposes LIST, STATUS, SELECT, and UNSELECT. `pkg/mailruntime/imap_client.go:218-360` exposes UID-oriented search, fetch, store, move, copy, delete, expunge, and append. This is the strongest existing foundation for a mirror feature.

The most important detail is that `IMAPClient.Search` uses `UIDSearch` in `pkg/mailruntime/imap_client.go:247-253`, and the fetch path takes UIDs in `pkg/mailruntime/imap_client.go:272-301`. Durable mirroring should be UID-oriented, not sequence-number-oriented, because sequence numbers are session-relative and unstable across expunges.

### Current JS service and JS module

`pkg/services/smailnailjs/service.go:88-119` defines a rich `Session` interface for mailbox list/status/select/search/fetch/mutation/append, and `pkg/services/smailnailjs/service.go:301-344` shows how that service can either use raw credentials or resolve stored accounts by `accountId`. The real implementation delegates to `mailruntime` in `pkg/services/smailnailjs/service.go:359-555`.

`pkg/js/modules/smailnail/module.go:50-115` exports the top-level JS module. `pkg/js/modules/smailnail/module.go:117-295` exposes mailbox selection, search, fetch, flag changes, move/copy/delete, expunge, append, and close to JavaScript callers.

This matters because it proves the repository already has a richer mailbox runtime than the CLI exposes today. The new mirror verb does not need to invent a fresh IMAP abstraction. It should reuse the same underlying runtime directly in Go instead of routing through JS.

### Current hosted backend and app database

`pkg/smailnaild/db.go:16-173` defines the default SQLite application database and a migration framework. The existing schema in `pkg/smailnaild/db.go:191-260` creates tables for `imap_accounts`, account test history, rules, rule runs, users, identities, and sessions.

`pkg/smailnaild/accounts/service.go:199-329` shows how hosted accounts are created, encrypted, and resolved. `pkg/smailnaild/accounts/service.go:387-453` exposes list/detail mailbox reads by opening a selected mailbox and then using the DSL fetch path. `pkg/smailnaild/http.go:25-44` and `pkg/smailnaild/http.go:180-197` show the hosted API contract for accounts, mailbox previews, message detail, and rules.

This subsystem matters for two reasons:

1. it proves the repo already contains a mature SQLite migration pattern and encrypted credential storage pattern,
2. it also makes clear that the hosted DB schema is an app-state database, not a local mailbox mirror.

### Current MCP identity and stored-account boundary

`pkg/mcp/imapjs/identity_middleware.go:69-105` boots a shared app DB when MCP is configured with a DSN and encryption key. `pkg/mcp/imapjs/identity_middleware.go:108-149` resolves the authenticated principal to a local user, and `pkg/mcp/imapjs/identity_middleware.go:197-221` resolves a stored IMAP account into connect options only when that account is enabled for MCP.

This is the correct security boundary for hosted or MCP-backed credential reuse. It is not the right persistence boundary for a local mailbox mirror database.

## Current-State Observations That Matter For The Mirror Design

### Observation 1: the CLI is already Glazed-first

The new feature should follow the existing pattern in `cmd/smailnail/commands/fetch_mail.go:59-203`, not create a separate ad hoc Cobra implementation. The repository already uses `cmds.NewCommandDescription`, embedded Glazed sections, and `cli.BuildCobraCommandFromCommand(...)` in `cmd/smailnail/main.go:40-67`.

### Observation 2: the DSL fetch engine is feature-rich, but its sync model is not the right one for mirroring

The DSL fetch path works by searching, receiving sequence numbers, and then doing pagination and content fetches around those sequence numbers in `pkg/dsl/processor.go:77-231`. That is fine for previews and one-off rule executions, but it is not ideal for an incremental durable mirror because:

- sequence numbers shift,
- offset-based pagination is not a durable sync cursor,
- the code is optimized for output shaping, not for durable sync checkpoints.

### Observation 3: `mailruntime` is already UID-oriented

`pkg/mailruntime/imap_client.go:247-253` uses `UID SEARCH`, and `pkg/mailruntime/imap_client.go:272-360` fetches, stores, deletes, expunges, and appends by UID. This is the right base for sync state such as:

- highest mirrored UID per mailbox,
- mailbox `UIDVALIDITY`,
- tombstoning or pruning,
- restart-safe incremental sync.

### Observation 4: there is no reusable inbound MIME parsing package yet

The repository already depends on `github.com/emersion/go-message v0.18.2` and already uses it for message generation and address parsing, as shown in `go.mod:8-17` and `pkg/mailutil/addresses.go:9-21`. But there is not yet a dedicated package that takes downloaded raw RFC 822 bytes and produces normalized searchable message content for indexing.

That missing parser package is one of the real implementation tasks for this feature.

### Observation 5: SQLite already exists in the repo, but only for the hosted app

The default hosted database path is `smailnaild.sqlite` in `pkg/smailnaild/db.go:16-20`, and tests in `pkg/smailnaild/db_test.go` show that SQLite bootstrap is already normal here. That makes SQLite a natural fit for the mirror feature as well. But it should be a separate schema, not new tables stuffed into the hosted application database unless there is a deliberate future plan to unify them.

## Gap Analysis

The feature request requires capabilities the current codebase does not yet provide as one coherent flow.

### What exists already

- Glazed command infrastructure.
- Shared IMAP connection flags.
- Direct CLI IMAP access.
- Rich UID-oriented runtime.
- Hosted account resolution and encrypted secrets.
- Existing SQLite bootstrap patterns.
- Existing mail search/fetch output models.

### What is missing

- a dedicated `mirror` verb,
- durable sync state per mailbox,
- a mirror-specific schema,
- raw message persistence strategy,
- a reusable inbound parser for downloaded mail,
- a searchable local message index,
- mirror-specific tests and smoke coverage.

### What must not be conflated

- preview/report output vs durable synchronization,
- hosted app state vs local mailbox content,
- sequence-number pagination vs restart-safe UID checkpoints,
- search index tables vs canonical raw message storage.

## Design Goals

1. Build on existing functionality as much as possible.
2. Keep the first version read-only against the remote server.
3. Make the mirror restart-safe and idempotent.
4. Make the local store rebuildable from raw data.
5. Keep the code understandable to an intern.
6. Avoid adding compatibility adapters unless they are strictly necessary.

## Non-Goals

1. Do not turn `fetch-mail` into a mirror command.
2. Do not bolt mail-content tables onto the hosted app database in v1.
3. Do not introduce a second hidden credential store.
4. Do not assume advanced IMAP extensions like QRESYNC or CONDSTORE are always available.

## Options Considered

### Option A: extend `fetch-mail` with `--sqlite-path` and `--download`

This would reuse the current Glazed command and flag surface. It is tempting because the command already knows how to search and fetch messages.

Why it is attractive:

- lowest initial code churn,
- reuses the CLI surface users already know,
- reuses the DSL rule-building code directly.

Why it is not recommended:

- it overloads a row-output command with durable sync responsibilities,
- its control flow is built around transient rows, not sync checkpoints,
- it encourages sequence-number and offset thinking instead of UID checkpoints,
- it will make the command harder to understand and maintain.

Conclusion: reject as the primary design. Reuse pieces of `fetch-mail`, but not the command itself.

### Option B: drive the mirror through the JS module

Because the JS module already exposes `search`, `fetch`, `list`, `status`, `append`, `move`, `copy`, `delete`, and `expunge`, the CLI could theoretically run JavaScript to perform the mirror logic.

Why it is attractive:

- very flexible,
- reuses the existing exported IMAP runtime surface,
- could later align CLI and MCP scripting stories.

Why it is not recommended for v1:

- adds an unnecessary execution layer for a Go-native CLI feature,
- makes sync-state and parser code harder to type-check and test,
- obscures control flow for an intern,
- complicates error handling around durable storage and migrations.

Conclusion: keep JS as a consumer of the runtime, not the core implementation path.

### Option C: store mirrored mail in the hosted application database

This option would reuse the existing SQLite bootstrap and put mail mirror tables into the same schema as `imap_accounts`, `rules`, and `users`.

Why it is attractive:

- one DB per environment,
- existing migration pattern is already present,
- future hosted search UI could read the same data.

Why it is not recommended for the first local mirror feature:

- the lifecycle is different: local mirror data can grow large and be disposable,
- the ownership semantics are different,
- the risk of mixing app metadata and content blobs is high,
- it makes future hosted-vs-local separation harder, not easier.

Conclusion: reject for v1. Reuse the migration pattern, not the schema.

### Option D: mirror only to SQLite BLOBs

In this approach, full raw messages are stored as `BLOB` columns in SQLite and all search data lives in the same DB.

Why it is attractive:

- one artifact to copy and back up,
- simple operational story,
- easy transactional integrity.

Why it is only partially attractive:

- the DB can grow very quickly,
- reindexing becomes more awkward,
- inspecting raw messages outside the DB is harder,
- large attachments can make the DB unwieldy.

Conclusion: acceptable for a prototype, but not the best long-term mirror shape.

### Option E: mirror to raw `.eml` files plus a dedicated SQLite index

In this approach, the raw `.eml` file is the canonical local copy, and SQLite stores structured metadata and search fields pointing at that raw file.

Why it is attractive:

- true downloadable mirror, not just a cache,
- easy reindexing if schema changes,
- easier debugging and manual inspection,
- safer handling of schema evolution,
- keeps the search DB smaller than a raw-BLOB-only design.

Tradeoffs:

- two storage layers instead of one,
- requires path management and disk layout design,
- must ensure writes are idempotent and crash-safe.

Conclusion: this is the recommended design.

## Recommended Architecture

### High-level recommendation

Implement a new local mirror subsystem with three main layers:

```text
smailnail mirror (Glazed command)
        |
        v
mirror orchestrator
        |
        +--> IMAP source adapter (built on pkg/mailruntime)
        +--> raw file store (.eml mirror)
        +--> SQLite mirror store (metadata + search)
        +--> MIME parser/importer (go-message/mail)
```

### Proposed package layout

One reasonable write set is:

```text
cmd/smailnail/commands/mirror.go
pkg/mirror/types.go
pkg/mirror/service.go
pkg/mirror/store.go
pkg/mirror/schema.go
pkg/mirror/parser.go
pkg/mirror/files.go
pkg/mirror/service_test.go
pkg/mirror/store_test.go
```

Alternative naming if you want harder boundaries:

```text
pkg/mirror/
pkg/mirrorstore/
pkg/mailparse/
```

I would start with one `pkg/mirror` package and split only if it becomes noisy. For an intern, one package is easier to follow.

### What to reuse directly

- `cmd/smailnail/main.go:18-79` for CLI registration style.
- `pkg/imap/layer.go:24-88` for the shared IMAP section and direct connection flags.
- `pkg/mailruntime/imap_client.go:91-360` for connect, capability discovery, UID search, fetch, and raw fetch.
- `go.mod:8-17` and `pkg/mailutil/addresses.go:9-21` as proof that `go-message/mail` is already in the repo and can support parsing.
- `pkg/smailnaild/db.go:29-173` as the style reference for SQLite bootstrap and migrations.

### What to reuse conceptually, but not directly

- `pkg/dsl/processor.go:49-320` as a reference for batched fetch and MIME-aware content extraction.
- `pkg/smailnaild/accounts/service.go:528-629` as a reference for opening and probing mailboxes cleanly.
- `pkg/services/smailnailjs/service.go:301-555` as a reference for how the higher-level service wraps `mailruntime`.

### What not to reuse as the core sync engine

- the sequence-number and offset-centric loop in `pkg/dsl/processor.go:77-231`,
- the hosted app schema in `pkg/smailnaild/db.go:191-260`,
- the JS module as the execution engine for a Go CLI.

## Mirror Storage Layout

### Filesystem layout

Recommended mirror root:

```text
<mirror-root>/
  mirror.db
  raw/
    <account-key>/
      <mailbox-slug>/
        <uidvalidity>/
          <uid>.eml
```

Why include `uidvalidity` in the path:

- IMAP UIDs are only stable within a mailbox for a given `UIDVALIDITY`.
- If the mailbox is recreated or reset, old UIDs must not collide with new ones.

### SQLite schema

Recommended v1 schema:

```sql
CREATE TABLE mirror_metadata (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mailbox_sync_state (
    account_key TEXT NOT NULL,
    mailbox_name TEXT NOT NULL,
    uidvalidity INTEGER NOT NULL,
    highest_uid INTEGER NOT NULL DEFAULT 0,
    last_uidnext INTEGER NOT NULL DEFAULT 0,
    last_sync_at TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'active',
    PRIMARY KEY (account_key, mailbox_name)
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_key TEXT NOT NULL,
    mailbox_name TEXT NOT NULL,
    uidvalidity INTEGER NOT NULL,
    uid INTEGER NOT NULL,
    message_id TEXT NOT NULL DEFAULT '',
    internal_date TEXT NOT NULL DEFAULT '',
    sent_date TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    from_summary TEXT NOT NULL DEFAULT '',
    to_summary TEXT NOT NULL DEFAULT '',
    cc_summary TEXT NOT NULL DEFAULT '',
    size_bytes INTEGER NOT NULL DEFAULT 0,
    flags_json TEXT NOT NULL DEFAULT '[]',
    headers_json TEXT NOT NULL DEFAULT '{}',
    parts_json TEXT NOT NULL DEFAULT '[]',
    body_text TEXT NOT NULL DEFAULT '',
    body_html TEXT NOT NULL DEFAULT '',
    search_text TEXT NOT NULL DEFAULT '',
    raw_path TEXT NOT NULL,
    raw_sha256 TEXT NOT NULL DEFAULT '',
    has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
    remote_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (account_key, mailbox_name, uidvalidity, uid)
);

CREATE INDEX idx_messages_mailbox_uid
    ON messages(account_key, mailbox_name, uidvalidity, uid);

CREATE INDEX idx_messages_message_id
    ON messages(message_id);

CREATE INDEX idx_messages_dates
    ON messages(internal_date, sent_date);
```

This is intentionally simple. It avoids over-normalization in v1 and optimizes for:

- idempotent upserts,
- direct mailbox/UID lookup,
- offline inspection,
- straightforward text querying.

### Full-text search options

There are three reasonable designs.

#### FTS option 1: no FTS in v1

Use indexed columns and `LIKE`/`instr` queries over subject/from/body snippets.

Pros:

- simplest,
- no dependence on SQLite FTS availability.

Cons:

- weaker search quality,
- slower on large datasets.

#### FTS option 2: optional FTS5 in `auto` mode

Try to create an FTS5 virtual table. If it works, use it. If not, continue with plain SQL search.

Pros:

- good user experience when available,
- graceful fallback.

Cons:

- more branching in tests and code.

#### FTS option 3: require FTS5

Fail fast if FTS5 is unavailable.

Pros:

- one clean code path.

Cons:

- brittle for local builds,
- unnecessary friction for the first implementation.

Recommendation: use `auto` mode. The user asked for mail to be easily searchable, but the safest v1 is to make FTS5 a capability, not a hard requirement.

## Why `mailruntime` Should Drive The Sync Loop

The core mirror loop should use `pkg/mailruntime/imap_client.go`, not the DSL fetch pipeline, for three reasons.

### Reason 1: UID-based search is the right sync primitive

`pkg/mailruntime/imap_client.go:247-253` performs `UID SEARCH`. That means incremental sync can naturally say:

- load last mirrored UID for this mailbox,
- search for UIDs above that checkpoint,
- fetch only those UIDs.

That is much closer to a mirror than `offset` and `limit`.

### Reason 2: `mailruntime` already supports raw fetch

`pkg/mailruntime/imap_client.go:272-301` can fetch `FetchBodyRaw`, and `pkg/mailruntime/imap_client.go:344-360` already supports append. For downloading, raw message bytes are the canonical thing to persist. The mirror should store that canonical raw form and derive search fields from it.

### Reason 3: `mailruntime` already has mailbox capability/state helpers

`pkg/mailruntime/imap_client.go:139-215` supports LIST, STATUS, SELECT, and UNSELECT. Those are the exact operations needed to:

- enumerate mailboxes,
- inspect `UIDNEXT` and message counts,
- detect mailbox resets,
- select mailboxes read-only during sync.

## Proposed Command Surface

Recommended command name:

```text
smailnail mirror
```

### Recommended flags

Reuse the existing IMAP section, then add a mirror-specific section.

Recommended v1 flags:

- `--sqlite-path`: path to the mirror SQLite DB, default `./smailnail-mirror.sqlite`
- `--mirror-root`: root directory for raw message files, default `./smailnail-mirror`
- `--mailbox`: one mailbox to mirror, default `INBOX`
- `--all-mailboxes`: enumerate and mirror all selectable mailboxes
- `--batch-size`: number of UIDs to fetch per batch, default `100`
- `--max-messages`: optional cap for test runs
- `--download-raw`: default `true`
- `--import-sqlite`: default `true`
- `--search-mode`: `basic|fts-auto|fts-required`
- `--print-plan`: show what would happen without downloading
- `--reset-mailbox-state`: clear stored state for the selected mailbox before syncing
- `--tombstone-missing`: mark remotely missing messages as deleted in the local DB after a full mailbox scan

### Optional future flags

- `--account-id`: reuse stored hosted credentials when explicitly wired to the app DB
- `--mailbox-include` / `--mailbox-exclude`
- `--since`
- `--stop-after-uid`
- `--write-json-export`
- `--attachment-policy`

## Sync Semantics

### Recommended default behavior

- connect using IMAP in read-only mode for the selected mailbox,
- inspect `UIDVALIDITY`, `UIDNEXT`, and message count,
- resume from stored highest UID if `UIDVALIDITY` matches,
- fetch new messages in UID batches,
- persist raw `.eml`,
- parse and upsert normalized search data,
- update sync state only after the batch is durable.

### Handling `UIDVALIDITY`

This is non-negotiable. If the stored `UIDVALIDITY` differs from the server's current `UIDVALIDITY`, the existing mailbox checkpoint is invalid.

Recommended behavior:

1. mark old rows for that mailbox snapshot as superseded,
2. create a fresh sync state row with the new `UIDVALIDITY`,
3. write new raw files under a new `uidvalidity` directory,
4. do not silently merge old and new UID spaces.

### Handling deletions and expunges

There are two reasonable policies.

#### Conservative default

- keep local copies of already-downloaded mail,
- mark rows as `remote_deleted = true` when a full reconcile detects they are gone remotely,
- do not delete raw files automatically.

#### Strict mirror mode

- remove local rows and files when the remote message is gone.

Recommendation: conservative default, strict optional mode later. The user asked for download and searchability; data loss should not be the default.

## Proposed Runtime Flow

### Initial full sync

```text
load config
open/create mirror.db
bootstrap schema
connect to IMAP
choose mailbox set
for each mailbox:
  STATUS or SELECT mailbox
  read UIDVALIDITY and UIDNEXT
  load stored sync state
  if missing or UIDVALIDITY changed:
    initialize fresh state
  determine UID range to fetch
  fetch messages in batches
  write raw .eml files
  parse MIME into structured fields
  upsert rows into SQLite
  advance highest_uid checkpoint
close IMAP
```

### Incremental sync pseudocode

```go
func SyncMailbox(ctx context.Context, client *mailruntime.IMAPClient, store *Store, mailbox string, opts SyncOptions) error {
    state, err := store.LoadMailboxState(mailbox)
    if err != nil { return err }

    selection, err := client.SelectMailbox(mailbox, true)
    if err != nil { return err }

    currentUIDValidity := selection.UIDValidity
    if state.Exists && state.UIDValidity != currentUIDValidity {
        if err := store.ResetMailboxSnapshot(mailbox, currentUIDValidity); err != nil {
            return err
        }
        state = NewMailboxState(mailbox, currentUIDValidity)
    }

    startUID := state.HighestUID + 1
    candidateUIDs, err := client.Search(&mailruntime.SearchCriteria{
        UID: uidRange(startUID, 0),
    })
    if err != nil { return err }

    for _, batch := range chunkUIDs(candidateUIDs, opts.BatchSize) {
        msgs, err := client.Fetch(batch, []mailruntime.FetchField{
            mailruntime.FetchUID,
            mailruntime.FetchFlags,
            mailruntime.FetchInternalDate,
            mailruntime.FetchSize,
            mailruntime.FetchEnvelope,
            mailruntime.FetchHeaders,
            mailruntime.FetchBodyRaw,
            mailruntime.FetchAttachments,
        })
        if err != nil { return err }

        tx, err := store.BeginTx(ctx)
        if err != nil { return err }

        for _, msg := range msgs {
            rawPath, sha, err := WriteRawMessage(opts.RawRoot, mailbox, currentUIDValidity, msg.UID, msg.BodyRaw)
            if err != nil { tx.Rollback(); return err }

            parsed, err := ParseRawMessage(msg.BodyRaw)
            if err != nil { tx.Rollback(); return err }

            if err := tx.UpsertMessage(mailbox, currentUIDValidity, msg, parsed, rawPath, sha); err != nil {
                tx.Rollback()
                return err
            }
        }

        maxUID := batch[len(batch)-1]
        if err := tx.UpdateMailboxState(mailbox, currentUIDValidity, maxUID, selection.UIDNext); err != nil {
            tx.Rollback()
            return err
        }
        if err := tx.Commit(); err != nil { return err }
    }

    return nil
}
```

## Parsing Strategy

### Recommended parser approach

Use the raw RFC 822 message as input and parse it with `github.com/emersion/go-message/mail`.

Outputs to derive:

- canonical headers map,
- `Message-ID`,
- normalized from/to/cc summaries,
- plain-text body,
- HTML body,
- attachment metadata,
- a pre-concatenated `search_text` field.

### Why not rely only on the existing DSL MIME extraction

The DSL processor is good at fetching and projecting output fields during a live IMAP request, but the mirror/index path needs:

- a parser that can run against already-downloaded raw messages,
- a parser that can be rerun without touching the network,
- a parser whose output schema is owned by the local mirror package.

That parser does not exist yet. It should be added rather than forcing the mirror to keep reusing the live DSL fetch path after download.

## API Sketches

### Mirror service

```go
type SyncOptions struct {
    Mailboxes         []string
    BatchSize         int
    MaxMessages       int
    DownloadRaw       bool
    ImportSQLite      bool
    SearchMode        string
    TombstoneMissing  bool
}

type Service struct {
    client *mailruntime.IMAPClient
    store  *Store
    files  *FileStore
    parser *Parser
}

func NewService(client *mailruntime.IMAPClient, store *Store, files *FileStore, parser *Parser) *Service
func (s *Service) Sync(ctx context.Context, opts SyncOptions) (*SyncReport, error)
func (s *Service) SyncMailbox(ctx context.Context, mailbox string, opts SyncOptions) (*MailboxReport, error)
```

### Store bootstrap

```go
type Store struct {
    db *sqlx.DB
}

func OpenStore(path string) (*Store, error)
func (s *Store) Bootstrap(ctx context.Context) error
func (s *Store) LoadMailboxState(ctx context.Context, accountKey, mailbox string) (*MailboxState, error)
func (s *Store) UpsertMessage(ctx context.Context, input UpsertMessageInput) error
func (s *Store) UpdateMailboxState(ctx context.Context, state MailboxState) error
```

### Parser

```go
type ParsedMessage struct {
    MessageID      string
    Headers        map[string]string
    FromSummary    string
    ToSummary      string
    CCSummary      string
    BodyText       string
    BodyHTML       string
    SearchText     string
    HasAttachments bool
    Parts          []PartView
}

func ParseRawMessage(raw []byte) (*ParsedMessage, error)
```

## Testing Strategy

### Unit tests

- schema bootstrap creates required tables and indexes,
- state reset on `UIDVALIDITY` change works,
- upsert is idempotent for the same `(mailbox, uidvalidity, uid)`,
- parser extracts headers, addresses, and MIME text correctly,
- `search_text` generation is deterministic.

### Integration tests

- incremental sync starting from empty state,
- rerunning sync downloads nothing new when no new UIDs exist,
- mailbox reset with different `UIDVALIDITY`,
- missing raw directory is created automatically,
- FTS auto mode falls back cleanly if virtual table creation fails.

### End-to-end smoke

Use the maintained Docker IMAP fixture referenced in the repo `README.md`. The smoke flow should:

1. create or seed a mailbox,
2. run `smailnail mirror`,
3. inspect SQLite rows and raw file count,
4. rerun sync and confirm idempotence,
5. add a new message and confirm incremental pickup.

## Implementation Plan

### Phase 1: CLI and store bootstrap

Files:

- `cmd/smailnail/main.go`
- `cmd/smailnail/commands/mirror.go`
- `pkg/mirror/store.go`
- `pkg/mirror/schema.go`

Work:

1. Add the new command and Glazed sections.
2. Add a dedicated SQLite bootstrap with schema versioning.
3. Add a minimal report output summarizing mirrored mailboxes, fetched messages, and DB path.

### Phase 2: UID-based mirror loop

Files:

- `pkg/mirror/service.go`
- `pkg/mirror/types.go`

Work:

1. Open IMAP with `mailruntime.Connect(...)`.
2. Enumerate mailboxes or use the selected mailbox.
3. Load and update mailbox sync state.
4. Fetch messages in UID batches with raw body enabled.

### Phase 3: raw file persistence and parsing

Files:

- `pkg/mirror/files.go`
- `pkg/mirror/parser.go`

Work:

1. Persist raw `.eml` files in a `uidvalidity`-aware path.
2. Parse RFC 822 into searchable fields.
3. Persist normalized metadata to SQLite.

### Phase 4: search and reconciliation

Files:

- `pkg/mirror/store.go`
- `pkg/mirror/service.go`

Work:

1. Add optional FTS5 bootstrap.
2. Add tombstoning for missing remote messages during full reconcile.
3. Add CLI output that summarizes index stats and deltas.

### Phase 5: tests and docs

Files:

- `pkg/mirror/*_test.go`
- `README.md`
- `cmd/smailnail/README.md`

Work:

1. Add unit and integration coverage.
2. Document the new verb.
3. Add a smoke script if needed.

## Key Risks And Sharp Edges

### Risk 1: confusing preview semantics with sync semantics

If the implementation is built by stretching `fetch-mail`, it will inherit the wrong mental model. A mirror is durable state plus checkpoints, not just output rows.

### Risk 2: relying on sequence numbers

Sequence numbers are fine inside one live request, but they are not durable identity keys. Use UIDs plus `UIDVALIDITY`.

### Risk 3: indexing too much too early

If v1 tries to fully normalize every address, every MIME part, every attachment body, and every thread relation, the implementation will slow down. Keep v1 simple: raw files plus a flat searchable projection.

### Risk 4: assuming FTS5 availability

SQLite search design must tolerate the possibility that FTS5 is unavailable in a given build or environment. Design for fallback.

### Risk 5: storing secrets in the mirror DB

The mirror database should store mirrored mail data and sync state, not passwords. Credentials should come from the command invocation or an explicitly integrated account resolver path.

## Future Improvements

- add `--account-id` support that explicitly reads hosted stored accounts when a user opts into that integration,
- add remote change reconciliation based on `UIDNEXT`, `STATUS`, and later `CONDSTORE`/`QRESYNC` when available,
- add attachment extraction policies,
- add richer local query verbs against the mirror DB,
- surface mirror status in `smailnaild` later if local and hosted stories intentionally converge,
- add background sync or IDLE-based live updates as a separate feature.

## Open Questions

1. Should v1 mirror only one mailbox by default, or enumerate all mailboxes when `--mailbox` is omitted?
2. Should the canonical local artifact be a raw `.eml` file plus SQLite metadata, or should SQLite BLOB mode also be supported behind a flag?
3. Is FTS5 availability consistent enough in the target environments to make `fts-auto` worthwhile immediately?
4. Do you want the first implementation to support hosted stored accounts via `--account-id`, or should it stay credential-flag-based in v1?
5. Should deleted remote messages be tombstoned by default or only when `--tombstone-missing` is set?

## References

### Internal code references

- `cmd/smailnail/main.go:18-79`
- `cmd/smailnail/commands/fetch_mail.go:59-203`
- `cmd/smailnail/commands/fetch_mail.go:206-460`
- `pkg/imap/layer.go:13-88`
- `pkg/dsl/processor.go:49-320`
- `pkg/dsl/actions.go:16-238`
- `pkg/mailruntime/imap_client.go:91-360`
- `pkg/services/smailnailjs/service.go:88-119`
- `pkg/services/smailnailjs/service.go:301-555`
- `pkg/js/modules/smailnail/module.go:50-295`
- `pkg/smailnaild/db.go:16-260`
- `pkg/smailnaild/accounts/service.go:199-329`
- `pkg/smailnaild/accounts/service.go:387-835`
- `pkg/mcp/imapjs/identity_middleware.go:69-221`
- `pkg/smailnaild/http.go:25-197`
- `pkg/mailutil/addresses.go:9-21`
- `go.mod:5-17`

### External API references

- `github.com/emersion/go-imap/v2/imapclient` package docs: https://pkg.go.dev/github.com/emersion/go-imap/v2/imapclient
- `github.com/emersion/go-message/mail` package docs: https://pkg.go.dev/github.com/emersion/go-message/mail
- SQLite FTS5 documentation: https://www.sqlite.org/fts5.html

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
