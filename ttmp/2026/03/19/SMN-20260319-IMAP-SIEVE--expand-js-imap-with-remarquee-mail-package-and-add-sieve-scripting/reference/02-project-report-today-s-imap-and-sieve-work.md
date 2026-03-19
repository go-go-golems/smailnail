---
Title: 'Project Report: Today''s IMAP and Sieve Work'
Ticket: SMN-20260319-IMAP-SIEVE
Status: active
Topics:
    - imap
    - javascript
    - sieve
    - email
    - mcp
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/js/modules/smailnail/docs/examples.js
      Note: Starter examples expanded for newcomers
    - Path: pkg/js/modules/smailnail/module.go
      Note: JavaScript surface expansion
    - Path: pkg/mailruntime/imap_client.go
      Note: Shared IMAP runtime port completed today
    - Path: pkg/mailruntime/sieve_client.go
      Note: Shared ManageSieve runtime port completed today
    - Path: pkg/services/smailnailjs/service.go
      Note: Service-layer API expansion and account-backed Sieve wiring
    - Path: ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/reference/01-diary.md
      Note: Detailed diary of the work completed today
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-19T12:47:17.850443843-04:00
WhatFor: ""
WhenToUse: ""
---


# Project Report: Today's IMAP and Sieve Work

## Goal

Capture the work completed today on the `smailnail` IMAP and Sieve expansion so a reader can quickly understand what changed, why it changed, what remains open, and how to start using the new APIs immediately.

## Context

Today’s work turned the original research ticket into a shipped implementation package. The feature started as a design exercise around expanding the JavaScript IMAP runtime with donor code from `remarquee/pkg/mail` and adding a Sieve scripting layer. By the end of the day, the repository had:

- a reusable runtime core in `pkg/mailruntime`,
- a richer service layer in `pkg/services/smailnailjs`,
- an expanded JavaScript module in `pkg/js/modules/smailnail`,
- synced embedded docs and examples,
- a detailed ticket report and diary,
- and a tracked open item for hosted-account Sieve schema support.

The key architectural constraint is that the current Sieve runtime works, but the hosted account model still does not store dedicated Sieve settings. Today’s implementation uses the existing IMAP account as the default source for Sieve host/user/password when `connectSieve({ accountId })` is used.

## Quick Reference

### What shipped today

| Area | Result | Key file |
| --- | --- | --- |
| Shared runtime | Ported donor IMAP and ManageSieve logic into `pkg/mailruntime` | [imap_client.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/imap_client.go), [sieve_client.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/sieve_client.go) |
| Service layer | Added richer IMAP session interfaces and `ConnectSieve` | [service.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go) |
| JS module | Added mailbox automation methods and Sieve session/builder APIs | [module.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go) |
| Docs | Expanded symbol docs and starter examples | [service.js](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js), [examples.js](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js) |
| Ticket metadata | Updated task matrix, diary, changelog, and validation state | [tasks.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/tasks.md), [01-diary.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/reference/01-diary.md), [changelog.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/changelog.md) |

### Commit trail

- `439258f` `Add shared IMAP and Sieve runtime core`
- `e66bd45` `Expand smailnail JS IMAP and Sieve APIs`
- `bc4baaa` `Record IMAP and Sieve implementation ticket progress`
- `c6fd9c6` `Expand starter documentation and examples`
- `886b6d5` `Record starter-doc expansion in ticket`

### Working API examples

```javascript
const smailnail = require("smailnail")
const svc = smailnail.newService()

const session = svc.connect({
  accountId: "acc_123",
  mailbox: "INBOX"
})

try {
  const uids = session.search({ unseen: true, subject: "invoice" })
  const messages = session.fetch(uids, ["uid", "flags", "body.text"])
  session.addFlags(uids, ["\\Seen"])
  session.move(uids, "Processed/Invoices")
  return messages
} finally {
  session.close()
}
```

```javascript
const smailnail = require("smailnail")
const svc = smailnail.newService()

const script = svc.buildSieveScript((s) => {
  s.require(["fileinto"])
  s.if(s.headerContains("Subject", "invoice"), (a) => {
    a.fileInto("Invoices")
    a.stop()
  })
})

const sieve = svc.connectSieve({
  server: "sieve.example.com",
  username: "user@example.com",
  password: "secret"
})

try {
  sieve.check(script)
  sieve.putScript("main", script, { activate: true })
  return sieve.listScripts()
} finally {
  sieve.close()
}
```

### Validation commands

```bash
go test ./pkg/mailruntime ./pkg/services/smailnailjs ./pkg/js/modules/smailnail ./pkg/mcp/imapjs -count=1
docmgr doctor --root ttmp --ticket SMN-20260319-IMAP-SIEVE --stale-after 30
```

## Usage Examples

### 1. Start with a stored account

Use this when you already have an `accountId` and want mailbox automation without handling secrets directly:

```javascript
const session = svc.connect({
  accountId: "acc_123",
  mailbox: "INBOX"
})
```

This path is the default recommendation for MCP and hosted flows because it keeps credentials inside the existing account store.

### 2. Discover and inspect mailboxes

```javascript
const boxes = session.list()
const archive = session.status("Archive")
const selected = session.selectMailbox("Archive", { readOnly: true })
```

Use this when you need to figure out which mailbox to read from or when you want to switch from `INBOX` to another folder.

### 3. Search and fetch messages

```javascript
const uids = session.search({ from: "alerts@example.com", unseen: true })
const messages = session.fetch(uids, ["uid", "flags", "envelope", "body.text"])
```

This is the common flow for scripts that need to inspect message content before deciding what to do next.

### 4. Mutate messages

```javascript
session.addFlags(uids, ["\\Seen"])
session.copy(uids, "Archive/Alerts")
session.delete(uids, { expunge: false })
session.expunge()
```

Use this for the usual mail-automation tasks: mark, copy, move, delete, and clean up deleted mail.

### 5. Append a raw message

```javascript
const uid = session.append(rawRfc822, {
  mailbox: "Drafts",
  flags: ["\\Draft"]
})
```

This is the low-level way to create a message from JavaScript.

### 6. Build and upload a Sieve script

```javascript
const script = svc.buildSieveScript((s) => {
  s.require(["fileinto"])
  s.if(s.headerContains("Subject", "invoice"), (a) => {
    a.fileInto("Invoices")
    a.stop()
  })
})

sieve.check(script)
sieve.putScript("main", script, { activate: true })
```

Use this when the filtering should happen on the server instead of in a client-side rule engine.

## Related

- [Project diary](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/reference/01-diary.md)
- [Task list](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/tasks.md)
- [Design guide](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/design-doc/01-intern-guide-expanding-js-imap-and-adding-a-sieve-scripting-layer.md)
- [JS examples](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js)
