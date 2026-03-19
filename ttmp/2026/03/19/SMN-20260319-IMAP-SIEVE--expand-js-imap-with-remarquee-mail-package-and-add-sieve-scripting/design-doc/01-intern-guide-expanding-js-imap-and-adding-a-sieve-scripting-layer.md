---
Title: 'Intern guide: expanding JS IMAP and adding a sieve scripting layer'
Ticket: SMN-20260319-IMAP-SIEVE
Status: active
Topics:
    - imap
    - javascript
    - sieve
    - email
    - mcp
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../../../2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/imap_client.go
      Note: Donor IMAP runtime with mailbox and message operations
    - Path: ../../../../../../../../../2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_imap.go
      Note: Donor JS IMAP API shape to adapt into smailnail
    - Path: ../../../../../../../../../2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_sieve.go
      Note: Donor Sieve JS API and builder surface
    - Path: ../../../../../../../../../2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/sieve_client.go
      Note: Donor ManageSieve client and transport constraints
    - Path: pkg/js/modules/smailnail/module.go
      Note: |-
        Current JS module export surface and minimal session object
        JS runtime now exposes the richer IMAP and Sieve surface described in the guide
    - Path: pkg/mailruntime/imap_client.go
      Note: Implementation now follows the proposed transplant boundary
    - Path: pkg/mailruntime/sieve_client.go
      Note: Implementation now follows the proposed transplant boundary
    - Path: pkg/mcp/imapjs/identity_middleware.go
      Note: Identity-aware stored-account resolution for MCP execution
    - Path: pkg/mcp/imapjs/server.go
      Note: MCP tool surface and transport defaults
    - Path: pkg/services/smailnailjs/service.go
      Note: Current service and stored-account connection path
    - Path: pkg/smailnaild/accounts/service.go
      Note: Encrypted account storage and mailbox browsing service
    - Path: pkg/smailnaild/db.go
      Note: Hosted schema versions and account table migration point
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-19T11:23:39.421313162-04:00
WhatFor: ""
WhenToUse: ""
---



# Intern guide: expanding JS IMAP and adding a sieve scripting layer

## Executive Summary

`smailnail` already has the beginnings of a JavaScript IMAP runtime, but today that runtime is narrow. The exported JavaScript module can parse rules, build rules, and open an IMAP connection, yet the returned session only exposes `mailbox` metadata and `close()`. The hosted application and MCP server are more mature than the JS surface: they already support encrypted account storage, user/account ownership checks, MCP identity resolution, and JS documentation lookup.

The donor package in `remarquee/pkg/mail` is much broader. It already implements a higher-level IMAP client, a mailbox/query/message JavaScript API, a ManageSieve client, and a small Sieve script builder. That makes it the right source of implementation patterns and code, but not something to copy wholesale. `smailnail` has newer architectural constraints around account storage, auth, MCP, docs extraction, and deployment that the donor package does not know about.

The proposed direction is:

1. Introduce a reusable mail runtime core inside `smailnail`, derived from `remarquee/pkg/mail`.
2. Preserve the existing `require("smailnail")` module identity and extend it rather than replacing it with a donor-style global `mail` shim.
3. Expand the JavaScript IMAP session into a mailbox/query/message API.
4. Add a first-class Sieve client and Sieve builder API.
5. Extend the hosted account model so MCP and future UI/API flows can resolve Sieve endpoints securely.
6. Keep the design incremental: do not break the current rule-building APIs while the richer runtime lands.

If implemented in phases, this gives `smailnail` one coherent story across CLI, hosted app, and MCP:

- IMAP automation via a richer JS runtime,
- hosted account-backed execution via MCP identity resolution,
- rule DSL support for existing workflows,
- and Sieve management for server-side filtering.

## Problem Statement And Scope

### What `smailnail` is today

At a repository level, `smailnail` is not just one CLI. It is a collection of mail-focused surfaces:

- `smailnail`: rule DSL and direct fetch CLI ([`cmd/smailnail/main.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/main.go), [`cmd/smailnail/commands/mail_rules.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go), [`cmd/smailnail/commands/fetch_mail.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go))
- `smailnaild`: hosted backend with account CRUD, mailbox browsing, and rule dry-runs ([`pkg/smailnaild/http.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go))
- `smailnail-imap-mcp`: JavaScript execution MCP server ([`cmd/smailnail-imap-mcp/main.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go), [`pkg/mcp/imapjs/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go))
- `pkg/js/modules/smailnail`: reusable JS module exposed through `go-go-goja` ([`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go))
- `pkg/services/smailnailjs`: adapter/service layer used by the JS runtime ([`pkg/services/smailnailjs/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go))
- `pkg/dsl`: the YAML rule engine that drives CLI and hosted rule execution ([`pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go), [`pkg/dsl/search.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/search.go), [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go), [`pkg/dsl/actions.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/actions.go))

### The concrete problem

The user request is not "add IMAP from scratch." The problem is narrower and more architectural:

1. `smailnail` already has a JS surface, but it is too thin to be useful for mailbox automation.
2. `remarquee/pkg/mail` already contains a much richer runtime, but it was built in a different host context.
3. `smailnail` already has a secure hosted-account and MCP identity path that should remain the authority for account resolution.
4. There is no Sieve account or Sieve runtime path in `smailnail` yet.
5. The system needs one intern-readable guide that explains how all these pieces fit together.

### In scope

- Expanding the JS IMAP runtime inside `smailnail`.
- Reusing code and design from `remarquee/pkg/mail`.
- Adding a Sieve scripting and script-management layer.
- Explaining the CLI, hosted app, MCP, docs, auth, and account-storage boundaries.
- Producing a phased implementation plan with file-level guidance.

### Out of scope for the first implementation slice

- Replacing the YAML rule DSL with Sieve.
- Auto-translating all existing rule DSL constructs into Sieve scripts.
- Full UI implementation for Sieve management in the same phase as the runtime port.
- Multi-provider OAuth-specific IMAP/Sieve auth beyond the current password-based account model.

## Reader Orientation: How The Current System Fits Together

### Current top-level flow

```text
CLI user / Hosted user / MCP caller
            |
            v
      smailnail surfaces
      - CLI commands
      - hosted HTTP API
      - MCP executeIMAPJS
            |
            v
   smailnail business layers
   - dsl package
   - services/smailnailjs
   - js/modules/smailnail
   - smailnaild/accounts
   - smailnaild/rules
            |
            v
   go-imap/v2 client + account secrets + SQLite/Postgres app DB
```

### Why this matters for an intern

If you only read the JS module, you will misunderstand the system. The JS module is not the source of truth for accounts, auth, or even most mailbox functionality. The hosted app stores encrypted IMAP credentials. The MCP middleware resolves the authenticated principal to a local user. The JS runtime is then injected with a stored-account resolver so JavaScript can say `accountId` instead of handling credentials directly. Any design that ignores those layers will be wrong.

## Current-State Analysis

### 1. The existing JavaScript surface is intentionally small

The current module exports only three top-level callables and two service/session callables:

- `parseRule`
- `buildRule`
- `newService`
- `connect`
- `close`

Evidence:

- runtime exports are registered in [`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go)
- embedded docs only define those same five symbols in [`pkg/js/modules/smailnail/docs/service.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js)
- the module/export parity test enforces this in [`pkg/js/modules/smailnail/module_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go)

The important constraint is that `connect()` currently returns a session object that exposes only:

- `mailbox`
- `close()`

That behavior is implemented in [`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go), where `newSessionObject` sets `mailbox` and binds `close()`, but nothing like mailbox listing, searching, message fetching, or flag mutation.

### 2. The JS service layer is mostly a rule adapter plus a simple dialer

[`pkg/services/smailnailjs/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go) does three major things today:

1. decode JavaScript-friendly maps into Go structs,
2. build/parse DSL rules,
3. dial IMAP and select one mailbox.

Important observations:

- `BuildRuleOptions` exists and is mapped into `dsl.Rule`, which is why the current JS runtime is good at constructing searches.
- `ConnectOptions` only models IMAP connection details, not Sieve details.
- `Service.Connect` can resolve `accountId` via an injected `StoredAccountResolver`, which is already the correct hook for hosted and MCP execution.
- `RealDialer.Dial` normalizes defaults, creates `imap.IMAPSettings`, dials, logs in, selects a mailbox, and returns a tiny `Session`.

That means the service layer already has a useful dependency-injection shape, but the interface it returns is too weak for the requested feature set.

### 3. The current IMAP transport helper is low-level and TLS-only

[`pkg/imap/layer.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go) provides `IMAPSettings` and `ConnectToIMAPServer()`.

This helper:

- holds `server`, `port`, `username`, `password`, `mailbox`, and `insecure`,
- always uses `imapclient.DialTLS`,
- logs in and returns the raw `*imapclient.Client`.

This file is useful, but it is not a mail runtime abstraction. It is a connection bootstrap helper for CLI-style code.

### 4. The YAML rule engine is already a serious subsystem

This matters because the JS expansion should not accidentally duplicate it.

Key files:

- rule model and validation: [`pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go)
- search translation: [`pkg/dsl/search.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/search.go)
- fetch pipeline: [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go)
- actions: [`pkg/dsl/actions.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/actions.go)

Important architectural facts:

- `Rule.FetchMessages()` already performs a multi-stage fetch pipeline with search, metadata fetch, MIME-part analysis, and batch body fetch.
- `ActionConfig` already covers flags, copy, move, delete, and export.
- CLI and hosted rule dry-runs are built on top of this DSL path.

The JS IMAP expansion should therefore complement this engine, not replace it. The current rule DSL remains the best path for declarative search-and-act workflows. The richer JS runtime should cover interactive or script-driven mailbox automation.

### 5. The hosted backend already owns account storage and mailbox browsing

The hosted side is more mature than the JS module:

- account model: [`pkg/smailnaild/accounts/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/types.go)
- account repository: [`pkg/smailnaild/accounts/repository.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/repository.go)
- account service: [`pkg/smailnaild/accounts/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go)
- DB bootstrap/migrations: [`pkg/smailnaild/db.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)
- HTTP API: [`pkg/smailnaild/http.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)

The current schema version is `6`, and `imap_accounts` already stores:

- ownership (`user_id`),
- connection settings,
- encrypted password fields,
- `is_default`,
- `mcp_enabled`.

This is the correct place to extend account data for Sieve. It would be a design mistake to let the JS runtime invent a second long-lived secret store.

### 6. The MCP server already has the right security shape

Important files:

- MCP tool registration: [`pkg/mcp/imapjs/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go)
- JS execution handler: [`pkg/mcp/imapjs/execute_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go)
- identity middleware: [`pkg/mcp/imapjs/identity_middleware.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/identity_middleware.go)
- execution context helpers: [`pkg/mcp/imapjs/service_context.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/service_context.go)

The middleware already does the hard part:

1. read the authenticated MCP principal,
2. resolve or provision the local user,
3. build a stored-account resolver,
4. inject that resolver into the JS execution context,
5. enforce `MCPEnabled` on the stored account.

This means the richest new IMAP and Sieve APIs can be exposed to MCP callers without changing the outer security model. The new code should plug into this context-injection pattern instead of bypassing it.

### 7. The documentation path is already embedded and queryable

The docs subsystem matters because the user explicitly asked for intern-friendly guidance, and the runtime itself already exposes documentation via MCP.

Key files:

- docs registry: [`pkg/mcp/imapjs/docs_registry.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_registry.go)
- docs tool: [`pkg/mcp/imapjs/docs_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_tool.go)
- embedded package docs: [`pkg/js/modules/smailnail/docs/package.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/package.js), [`pkg/js/modules/smailnail/docs/service.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js), [`pkg/js/modules/smailnail/docs/examples.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js)

The implication is straightforward: any runtime expansion should ship with expanded symbol docs and examples in the same PR. Otherwise the MCP server becomes materially harder to use even if the code works.

## Gap Analysis

### Observed current capability

The current `smailnail` JS layer can:

- parse YAML rule strings,
- build normalized rule objects,
- connect to IMAP by inline credentials or stored `accountId`,
- close the session,
- query embedded docs through a second MCP tool.

### Missing capability relative to the request

The current `smailnail` JS layer cannot directly:

- list mailboxes,
- inspect server capabilities,
- fetch mailbox status,
- search mailboxes interactively,
- construct mailbox-scoped queries,
- fetch messages by UID with field selection,
- mutate message flags from JS,
- move/copy/delete messages from JS,
- append raw RFC822 content from JS,
- fetch message bodies or MIME parts on demand,
- connect to ManageSieve,
- list/get/put/activate/deactivate/rename/delete/check scripts,
- build Sieve scripts from JS.

### Why the donor package matters

The donor package already implements nearly all of those operations. That makes the problem mostly architectural integration and selective porting, not greenfield invention.

## Donor Package Analysis: `remarquee/pkg/mail`

### Core IMAP client

The donor IMAP core lives in:

- [`remarquee/pkg/mail/imap_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/imap_client.go)
- [`remarquee/pkg/mail/types.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/types.go)
- [`remarquee/pkg/mail/errors.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/errors.go)

What it already does:

- capability discovery,
- mailbox `LIST`,
- mailbox `STATUS`,
- mailbox `SELECT` / `UNSELECT`,
- UID search with a richer `SearchCriteria` model,
- field-oriented message fetch,
- flag mutation,
- move/copy/delete/expunge,
- append,
- mailbox create/rename/delete/subscribe/unsubscribe,
- fetch specific MIME body parts.

That is much closer to the runtime requested by the user than the current `smailnailjs.RealDialer`.

### Donor JS IMAP surface

The donor JS adapter in [`remarquee/pkg/mail/js_imap.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_imap.go) exposes a layered runtime:

1. `mail.imap.use(...)`
2. client methods like `capabilities`, `list`, `status`, `mailbox`, `withMailbox`, `batch`
3. mailbox methods like `search`, `query`, `get`, `fetch`, `move`, `copy`, `delete`, `expunge`, `append`
4. query-builder methods like `limit`, `peek`, `fetch`, `uids`, `list`, `each`
5. message methods like `addFlags`, `removeFlags`, `setFlags`, `markSeen`, `markUnseen`, `move`, `copy`, `delete`, `getBody`, `getPart`

This API shape is valuable because it gives JavaScript callers a real working vocabulary instead of one-shot helpers.

### Donor Sieve core

The donor Sieve core lives in [`remarquee/pkg/mail/sieve_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/sieve_client.go).

It already implements:

- greeting/capability parsing,
- AUTHENTICATE PLAIN,
- `LISTSCRIPTS`,
- `GETSCRIPT`,
- `PUTSCRIPT`,
- `SETACTIVE`,
- `DELETESCRIPT`,
- `RENAMESCRIPT`,
- `CHECKSCRIPT`,
- `HAVESPACE`,
- logout.

### Donor Sieve JS layer

[`remarquee/pkg/mail/js_sieve.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_sieve.go) exposes:

- connection-scoped Sieve client methods,
- a string-building Sieve DSL with conditions and actions,
- helper methods for script upload/activation/validation.

This is exactly the kind of runtime the user is asking for when they say "add a sieve scripting layer."

### Donor strengths

- Broad IMAP feature coverage already exists.
- The query/message object model is usable from JS.
- Sieve support is already present.
- Structured `MailError` is reusable.

### Donor risks and constraints

The donor package cannot be copied blindly.

Risk 1: host model mismatch

- donor code creates a global `mail` object and custom `require("mail")` shim.
- `smailnail` already uses `go-go-goja` native modules and embedded docs.
- Therefore we should port behavior, not the donor's host-registration pattern.

Risk 2: account model mismatch

- donor runtime uses named accounts and secret maps in `RuntimeOptions`.
- `smailnail` already has encrypted hosted accounts and MCP identity resolution.
- Therefore the donor account resolution approach should be replaced with `smailnail`'s stored-account resolver pattern.

Risk 3: Sieve transport security is not production-ready as written

- donor `ConnectSieve` uses plain `net.Dial` and then `AUTHENTICATE "PLAIN"`.
- it parses `STARTTLS` capability but does not use it.
- that is acceptable as a prototype reference, not as a final production implementation for `smailnail`.

Risk 4: semantic overlap with the rule DSL

- the donor JS runtime encourages imperative mailbox scripting.
- `smailnail` already has a declarative YAML rule engine.
- the design needs to keep both layers clear instead of muddling them together.

## Proposed Solution

### Design principles

1. Keep `require("smailnail")` as the public JavaScript package name.
2. Preserve the existing `parseRule`, `buildRule`, and `newService` APIs.
3. Make `connect()` richer rather than replacing it.
4. Add explicit Sieve support through the same service object.
5. Reuse donor implementation code where it lowers risk, but keep `smailnail`'s auth, docs, and account model.
6. Keep IMAP rule DSL and Sieve scripting as separate concepts in phase 1.

### Proposed package layout

```text
pkg/
  mailruntime/
    errors.go
    types.go
    imap_client.go
    sieve_client.go
    js_imap_adapter.go
    js_sieve_adapter.go
  services/smailnailjs/
    service.go
    imap_session.go
    sieve_session.go
    views.go
  js/modules/smailnail/
    module.go
    docs/
      package.js
      service.js
      examples.js
  mcp/imapjs/
    execute_tool.go
    docs_registry.go
    docs_tool.go
  smailnaild/accounts/
    types.go
    repository.go
    service.go
  smailnaild/
    db.go
    http.go
```

### Why a new `pkg/mailruntime` package is the cleanest boundary

`smailnailjs.Service` currently mixes:

- rule-building logic,
- connection option decoding,
- a tiny IMAP dialer abstraction.

That is too small for the requested runtime, but it is still the right orchestration layer. The richer IMAP and Sieve mechanics should live below it in a reusable package with no knowledge of MCP or hosted HTTP APIs. That gives us:

- reuse from JS module,
- reuse from MCP,
- future reuse from CLI or hosted rule execution,
- room for tests that do not depend on the JS layer.

### Target runtime architecture

```text
Authenticated user / CLI caller / MCP caller
                |
                v
      smailnail JS module: require("smailnail")
                |
                v
        smailnailjs.Service
        - rule helpers
        - connect IMAP
        - connect Sieve
        - stored-account resolution
                |
                v
           pkg/mailruntime
        - IMAP client wrapper
        - Sieve client wrapper
        - JS object adapters
        - structured mail errors
                |
                v
       go-imap / ManageSieve transport
```

### Proposed JavaScript API shape

The main rule here is: keep the existing names, add power below them.

#### Top level

```javascript
const smailnail = require("smailnail")

smailnail.parseRule(yaml)
smailnail.buildRule(options)
const svc = smailnail.newService()
```

#### IMAP connection

```javascript
const imap = svc.connect({
  accountId: "acc_123",
  mailbox: "INBOX"
})

imap.mailbox
imap.close()
imap.capabilities()
imap.list("*")
imap.status("INBOX")
imap.mailboxHandle("INBOX")
imap.withMailbox("INBOX", { readOnly: true }, (mbox) => {
  return mbox.query({ subject: "invoice" }).limit(10).list()
})
```

#### Mailbox/query/message flow

```javascript
const inbox = imap.mailboxHandle("INBOX")

const messages = inbox
  .query({ from: "alerts@example.com", unseen: true })
  .limit(25)
  .fetch(["uid", "flags", "envelope", "body.text"])
  .list()

messages[0].markSeen()
messages[0].move("Processed")
```

#### Sieve connection

```javascript
const sieve = svc.connectSieve({
  accountId: "acc_123"
})

const script = sieve.build((r) => {
  r.require(["fileinto", "imap4flags"])
  r.if(
    r.headerContains("Subject", "invoice"),
    (a) => {
      a.fileInto("Finance")
      a.stop()
    }
  )
})

sieve.check(script)
sieve.putScript("finance-filter", script, { activate: true })
```

### Naming decision: `mailboxHandle` instead of donor `mailbox`

The donor code uses `imap.mailbox(name)`. That is fine, but in `smailnail` the session already has a `mailbox` property string. Reusing the same name for both a property and a method is possible in some designs but unnecessarily confusing. The clearer shape is:

- `imap.mailbox` for the currently selected mailbox name
- `imap.mailboxHandle(name)` for an object representing mailbox operations

If preserving the donor method name is important, then the property should be renamed to `selectedMailbox`. One of those two changes should happen. Keeping both as `mailbox` would be intern-hostile.

### Account model extension for Sieve

The current `imap_accounts` table is IMAP-only. To support Sieve through the hosted app and MCP, extend it rather than creating a separate account ownership model.

Recommended additional fields:

- `sieve_enabled BOOLEAN NOT NULL DEFAULT FALSE`
- `sieve_server TEXT NOT NULL DEFAULT ''`
- `sieve_port INTEGER NOT NULL DEFAULT 4190`
- `sieve_username TEXT NOT NULL DEFAULT ''`
- `sieve_insecure BOOLEAN NOT NULL DEFAULT FALSE`

Optional later fields if separate auth becomes necessary:

- `sieve_secret_ciphertext`
- `sieve_secret_nonce`
- `sieve_secret_key_id`

Phase 1 recommendation:

- default Sieve credentials to IMAP username/password,
- allow Sieve host/port override,
- do not add separate Sieve secrets until a real provider requires them.

That keeps schema and UI complexity under control while still enabling most installations.

### Security and ownership model

The Sieve path should match the existing IMAP path:

1. hosted app stores encrypted credentials,
2. account is owned by a local `users.id`,
3. MCP principal resolves to local user,
4. stored-account resolver checks ownership and `MCPEnabled`,
5. runtime receives only resolved connection details.

For intern understanding, the key rule is simple:

`JavaScript should not become a second source of truth for account secrets.`

### Documentation model

Every new public symbol should have:

- a runtime export,
- a `__doc__` block,
- at least one example if the symbol is non-trivial,
- test coverage that keeps docs and exports aligned.

This follows the pattern already enforced by [`pkg/js/modules/smailnail/module_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go).

## Detailed API Sketches

### Go-side interfaces

```go
type IMAPSession interface {
    Mailbox() string
    Close()
    Capabilities() map[string]bool
    List(pattern string) ([]MailboxInfo, error)
    Status(name string) (*MailboxStatus, error)
    MailboxHandle(name string) MailboxHandle
}

type MailboxHandle interface {
    Name() string
    Search(criteria SearchCriteria) ([]uint32, error)
    Query(criteria SearchCriteria) Query
    Get(uid uint32) (MessageHandle, error)
    Fetch(uids []uint32, fields []FetchField) ([]*FetchedMessage, error)
    Move(uids []uint32, dest string) error
    Copy(uids []uint32, dest string) error
    Delete(uids []uint32, expunge bool) error
    Append(raw []byte, flags []string, date *time.Time) (uint32, error)
}

type SieveSession interface {
    Close()
    Capabilities() SieveCapabilities
    ListScripts() ([]ScriptInfo, error)
    GetScript(name string) (string, error)
    PutScript(name, body string, activate bool) error
    Activate(name string) error
    Deactivate() error
    DeleteScript(name string) error
    RenameScript(oldName, newName string) error
    CheckScript(body string) error
}
```

### Service-level API

```go
type Service struct {
    imapDialer  IMAPDialer
    sieveDialer SieveDialer
    storedAccountResolver StoredAccountResolver
}

func (s *Service) Connect(ctx context.Context, opts ConnectOptions) (IMAPSession, error)
func (s *Service) ConnectSieve(ctx context.Context, opts SieveConnectOptions) (SieveSession, error)
func (s *Service) ParseRuleMap(yaml string) (map[string]interface{}, error)
func (s *Service) BuildRuleMap(opts BuildRuleOptions) (map[string]interface{}, error)
```

### Stored-account resolver shape

Keep the existing resolver pattern, but broaden the resolved data:

```go
type ResolvedAccountOptions struct {
    IMAP  ConnectOptions
    Sieve SieveConnectOptions
}
```

That makes it possible for one account row to resolve both transport families cleanly.

## Pseudocode And Key Flows

### Flow 1: MCP caller executes JS against a stored account

```text
MCP request
  -> executeIMAPJS
  -> identity middleware resolves local user
  -> middleware injects stored-account resolver into context
  -> JS code calls smailnail.newService().connect({ accountId })
  -> service resolves account -> decrypted credentials
  -> runtime returns rich IMAP session object
  -> script performs mailbox/query/message operations
  -> tool returns exported JSON value
```

Pseudocode:

```go
func executeIMAPJSHandler(ctx context.Context, raw map[string]any) ToolResult {
    req := decode(raw)
    service := buildExecutionService(ctx) // already present
    module := smailnailmodule.NewModuleWithServiceAndContext(ctx, service)
    rt := newRuntimeWith(module)
    value := rt.RunString(req.Code)
    return export(value)
}
```

### Flow 2: `connect({ accountId })` for IMAP

```go
func (s *Service) Connect(ctx context.Context, opts ConnectOptions) (IMAPSession, error) {
    if opts.AccountID != "" {
        resolved := s.storedAccountResolver.ResolveAccountOptions(ctx, opts.AccountID)
        opts = mergeMailboxOverride(resolved.IMAP, opts)
    }
    client := s.imapDialer.Dial(ctx, opts)
    return newRichIMAPSession(client), nil
}
```

### Flow 3: `connectSieve({ accountId })`

```go
func (s *Service) ConnectSieve(ctx context.Context, opts SieveConnectOptions) (SieveSession, error) {
    if opts.AccountID != "" {
        resolved := s.storedAccountResolver.ResolveAccountOptions(ctx, opts.AccountID)
        opts = resolved.Sieve
    }
    return s.sieveDialer.Dial(ctx, opts)
}
```

### Flow 4: Sieve script builder

```javascript
const script = sieve.build((r) => {
  r.require(["fileinto", "imap4flags"])
  r.if(
    r.all(
      r.headerContains("From", "billing@example.com"),
      r.headerContains("Subject", "invoice")
    ),
    (a) => {
      a.fileInto("Finance")
      a.addFlag("\\Seen")
      a.stop()
    }
  )
})
```

This builder should remain explicit string construction. It should not pretend to be a validated AST unless we are willing to build a real AST layer.

## Design Decisions

### Decision 1: port donor behavior into a new internal runtime package

Reason:

- avoids overloading `smailnailjs.Service`,
- preserves host independence,
- keeps room for direct Go tests.

### Decision 2: keep the existing `smailnail` module identity

Reason:

- current MCP docs tooling is already built around that package,
- existing callers and tests already use `require("smailnail")`,
- replacing it with donor `mail` global semantics would be needless churn.

### Decision 3: keep the rule DSL separate from the Sieve scripting layer

Reason:

- the current rule DSL is IMAP search + client-side action execution,
- Sieve is server-side filtering with different semantics,
- automatic translation would be lossy and easy to mis-sell.

### Decision 4: extend account storage before promising account-backed Sieve over MCP

Reason:

- inline Sieve credentials are easy to prototype,
- but real product use through MCP needs persistent secure configuration,
- the hosted app DB already solves ownership and secret storage.

### Decision 5: expand docs and tests in the same implementation phases

Reason:

- this system is exposed through a documentation-aware MCP,
- hidden features are operationally equivalent to missing features for new users.

## Alternatives Considered

### Alternative A: leave the current module alone and add a second module just for mailbox scripting

Rejected because:

- it would split the public story between rule-building and mailbox automation,
- the docs surface would become confusing,
- the existing `newService()` abstraction is already the natural place to grow.

### Alternative B: copy `remarquee/pkg/mail` almost verbatim

Rejected because:

- donor host registration does not match `smailnail`,
- donor account resolution does not match `smailnail`,
- donor Sieve transport security needs improvement,
- a literal transplant would fight the current docs and MCP identity model.

### Alternative C: make Sieve the new canonical rule engine and de-emphasize the IMAP DSL

Rejected for phase 1 because:

- the current DSL already powers CLI and hosted dry-runs,
- Sieve cannot express every existing client-side action workflow,
- that migration would be much larger than the user requested.

## Implementation Plan

### Phase 1: Create the reusable mail runtime core

Goal:

- establish `pkg/mailruntime` by porting and adapting donor code.

Files to add:

- `pkg/mailruntime/errors.go`
- `pkg/mailruntime/types.go`
- `pkg/mailruntime/imap_client.go`
- `pkg/mailruntime/sieve_client.go`

Files to read while implementing:

- donor IMAP core: [`remarquee/pkg/mail/imap_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/imap_client.go)
- donor errors/types: [`remarquee/pkg/mail/errors.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/errors.go), [`remarquee/pkg/mail/types.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/types.go)

Required adaptations:

- rename package and types to fit `smailnail`,
- keep structured errors,
- keep IMAP feature coverage,
- improve Sieve transport handling so TLS/STARTTLS is not silently ignored.

### Phase 2: Refactor `smailnailjs.Service` into an orchestration layer

Goal:

- preserve current rule helpers,
- add richer IMAP and Sieve connect paths.

Files to change:

- [`pkg/services/smailnailjs/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go)
- [`pkg/services/smailnailjs/views.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/views.go)
- add `pkg/services/smailnailjs/sieve.go` or equivalent helper files

Key tasks:

- keep `BuildRuleOptions` and `ParseRuleMap`,
- add `SieveConnectOptions`,
- make stored-account resolution return IMAP plus Sieve details,
- keep test injection points for dialers.

### Phase 3: Expand the JS module and docs

Goal:

- expose the richer runtime through `require("smailnail")`.

Files to change:

- [`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go)
- [`pkg/js/modules/smailnail/module_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go)
- [`pkg/js/modules/smailnail/docs/package.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/package.js)
- [`pkg/js/modules/smailnail/docs/service.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js)
- [`pkg/js/modules/smailnail/docs/examples.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js)

Key tasks:

- preserve old exports,
- add richer IMAP session methods,
- add `connectSieve`,
- add examples for mailbox queries and Sieve scripts,
- update export/doc parity tests.

### Phase 4: Wire MCP execution and docs query to the expanded runtime

Goal:

- make the new runtime usable through MCP without changing the security model.

Files to inspect and possibly change:

- [`pkg/mcp/imapjs/execute_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go)
- [`pkg/mcp/imapjs/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go)
- [`pkg/mcp/imapjs/docs_registry.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_registry.go)
- [`pkg/mcp/imapjs/docs_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_tool.go)
- [`pkg/mcp/imapjs/identity_middleware.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/identity_middleware.go)

Key tasks:

- keep `executeIMAPJS`,
- keep `getIMAPJSDocumentation`,
- inject expanded account resolution for Sieve,
- add tests for account-backed IMAP and account-backed Sieve.

### Phase 5: Extend hosted account storage for Sieve

Goal:

- make Sieve configuration durable and user-owned.

Files to change:

- [`pkg/smailnaild/db.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)
- [`pkg/smailnaild/accounts/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/types.go)
- [`pkg/smailnaild/accounts/repository.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/repository.go)
- [`pkg/smailnaild/accounts/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go)
- [`pkg/smailnaild/http.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)

Key tasks:

- bump schema version,
- add migration for Sieve columns,
- expose create/update/get fields through the account API,
- update service validation and defaulting rules.

### Phase 6: Optional hosted UI support

Goal:

- let a human configure and validate Sieve settings through the web UI.

Likely files:

- `ui/src/...` account form and detail flows
- generated embed assets after UI build

This phase is useful but should not block the core runtime and MCP implementation if time is tight.

## Testing And Validation Strategy

### Unit tests

Add or expand tests for:

- IMAP client wrapper behavior,
- Sieve client parsing and error handling,
- JS export surfaces,
- docs/export parity,
- stored-account resolution,
- `connectSieve({ accountId })` behavior.

Existing useful patterns:

- [`pkg/services/smailnailjs/service_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service_test.go)
- [`pkg/js/modules/smailnail/module_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go)
- [`pkg/mcp/imapjs/execute_tool_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool_test.go)
- [`pkg/mcp/imapjs/execute_tool_account_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool_account_test.go)

### Integration tests

Use the maintained Dovecot fixture for IMAP integration, following the style in:

- [`pkg/smailnaild/accounts/integration_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/integration_test.go)

Recommended new integration tests:

1. account-backed JS IMAP mailbox listing and query,
2. account-backed JS message mutation,
3. account-backed MCP execution path,
4. local ManageSieve integration against a compatible fixture if available.

### Documentation validation

For every new JS symbol:

- add `__doc__`,
- add example coverage where warranted,
- keep docs aligned with exports through the existing parity test.

### Migration validation

Add DB tests covering:

- schema version bump,
- default values for new Sieve columns,
- backward compatibility from schema version `6`.

## Risks, Alternatives, And Open Questions

### Risk: Sieve transport security

The donor code is a strong feature reference but a weak security reference for Sieve. Before exposing Sieve broadly, decide:

- whether to require STARTTLS,
- whether to support insecure/dev-only mode,
- whether some providers require distinct ports or auth semantics.

### Risk: account-schema creep

If separate Sieve credentials are added too early, the account model will become more complex than necessary. Start with shared credentials plus Sieve endpoint overrides unless real provider evidence forces a split.

### Risk: API sprawl

The donor runtime is broad. If every donor method is ported immediately, the first implementation may become hard to review. Prioritize:

1. mailbox list/status/query/fetch,
2. message flag/move/copy/delete operations,
3. Sieve list/get/put/activate/check/build.

Leave less critical helpers for follow-up phases if needed.

### Open question: should any YAML rule subset translate to Sieve?

Recommendation:

- not in phase 1,
- maybe later for a narrow subset such as `subject_contains -> fileinto`.

The semantic mismatch is real enough that auto-translation should be a separate design ticket.

### Open question: should the richer IMAP runtime also power CLI execution directly?

Recommendation:

- not immediately,
- but design `pkg/mailruntime` so CLI code could migrate later if duplication becomes painful.

## Intern Walkthrough: Where To Start Reading The Code

If you are a new intern joining this work, read in this order:

1. Repository overview:
   [`README.md`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md)
2. Current JS module:
   [`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go)
3. Current JS service layer:
   [`pkg/services/smailnailjs/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go)
4. MCP execution and auth:
   [`pkg/mcp/imapjs/execute_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go),
   [`pkg/mcp/imapjs/identity_middleware.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/identity_middleware.go)
5. Hosted account model:
   [`pkg/smailnaild/accounts/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go),
   [`pkg/smailnaild/db.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)
6. Existing rule engine:
   [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go)
7. Donor runtime:
   [`remarquee/pkg/mail/js_imap.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_imap.go),
   [`remarquee/pkg/mail/sieve_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/sieve_client.go),
   [`remarquee/pkg/mail/js_sieve.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_sieve.go)

## References

### Primary `smailnail` sources

- [`README.md`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/README.md)
- [`cmd/smailnail/commands/fetch_mail.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/fetch_mail.go)
- [`cmd/smailnail/commands/mail_rules.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail/commands/mail_rules.go)
- [`pkg/imap/layer.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/imap/layer.go)
- [`pkg/dsl/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/types.go)
- [`pkg/dsl/search.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/search.go)
- [`pkg/dsl/processor.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/processor.go)
- [`pkg/dsl/actions.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/dsl/actions.go)
- [`pkg/services/smailnailjs/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go)
- [`pkg/services/smailnailjs/views.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/views.go)
- [`pkg/js/modules/smailnail/module.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go)
- [`pkg/js/modules/smailnail/docs/service.js`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js)
- [`pkg/js/modules/smailnail/module_test.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go)
- [`pkg/mcp/imapjs/server.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/server.go)
- [`pkg/mcp/imapjs/execute_tool.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool.go)
- [`pkg/mcp/imapjs/identity_middleware.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/identity_middleware.go)
- [`pkg/mcp/imapjs/docs_registry.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_registry.go)
- [`pkg/smailnaild/db.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go)
- [`pkg/smailnaild/accounts/types.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/types.go)
- [`pkg/smailnaild/accounts/repository.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/repository.go)
- [`pkg/smailnaild/accounts/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go)
- [`pkg/smailnaild/rules/service.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/rules/service.go)
- [`pkg/smailnaild/http.go`](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/http.go)

### Donor sources

- [`remarquee/pkg/mail/types.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/types.go)
- [`remarquee/pkg/mail/errors.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/errors.go)
- [`remarquee/pkg/mail/imap_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/imap_client.go)
- [`remarquee/pkg/mail/js_imap.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_imap.go)
- [`remarquee/pkg/mail/sieve_client.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/sieve_client.go)
- [`remarquee/pkg/mail/js_sieve.go`](/home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_sieve.go)
