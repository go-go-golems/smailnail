__doc__("parseRule", {
  summary: "Parse a YAML rule string into a JavaScript object view of the smailnail DSL rule.",
  concepts: ["rule-parsing", "yaml-dsl"],
  params: [
    { name: "yamlString", type: "string", description: "Raw YAML containing a smailnail rule." }
  ],
  returns: { type: "Rule", description: "Structured smailnail rule data." },
  related: ["buildRule"],
  tags: ["dsl", "yaml", "rules"]
})

doc`
---
symbol: parseRule
---

Use \`parseRule\` when a script already has YAML rule content and needs the normalized object form.
`

__doc__("buildRule", {
  summary: "Construct a DSL rule from JavaScript-friendly search and output options.",
  concepts: ["rule-building", "search-config"],
  params: [
    { name: "options", type: "BuildRuleOptions", description: "Lower-camel JavaScript options for search and output fields." }
  ],
  returns: { type: "Rule", description: "Structured smailnail rule data ready for execution or serialization." },
  related: ["parseRule", "newService"],
  tags: ["dsl", "rules", "builder"]
})

doc`
---
symbol: buildRule
---

\`buildRule\` is the main ergonomic entrypoint for rule construction from JavaScript.
`

__doc__("buildSieveScript", {
  summary: "Build a Sieve script string from the fluent builder DSL.",
  concepts: ["sieve", "script-builder"],
  params: [
    { name: "builderFn", type: "function", description: "Callback that receives the Sieve builder object." }
  ],
  returns: { type: "string", description: "Rendered Sieve script." },
  related: ["connectSieve", "putScript", "check"],
  tags: ["sieve", "builder", "scripts"]
})

doc`
---
symbol: buildSieveScript
---

Use \`buildSieveScript\` to construct ManageSieve-compatible script text before uploading or validating it.
`

__doc__("newService", {
  summary: "Create a service object exposing rule helpers, IMAP operations, and Sieve operations.",
  concepts: ["service-construction", "imap-runtime", "sieve-runtime"],
  params: [],
  returns: { type: "Service", description: "Service object with rule, IMAP, and Sieve helpers." },
  related: ["connect", "connectSieve"],
  tags: ["service", "runtime"]
})

doc`
---
symbol: newService
---

Create a service when a script needs live IMAP or ManageSieve access in addition to rule helpers.
`

__doc__("connect", {
  summary: "Open an IMAP session using ConnectOptions and return a mailbox automation session.",
  concepts: ["imap-connection", "session-lifecycle"],
  params: [
    { name: "options", type: "ConnectOptions", description: "Server, credentials, mailbox, and TLS settings." }
  ],
  returns: { type: "IMAPSession", description: "Session object for mailbox automation." },
  related: ["newService", "close", "connectSieve"],
  tags: ["imap", "connection", "session"]
})

doc`
---
symbol: connect
---

\`connect\` opens an IMAP session, selects the requested mailbox, and returns the richer mailbox automation object.
`

__doc__("connectSieve", {
  summary: "Open a ManageSieve session using SieveConnectOptions.",
  concepts: ["sieve-connection", "session-lifecycle"],
  params: [
    { name: "options", type: "SieveConnectOptions", description: "Server, credentials, and optional account-derived settings." }
  ],
  returns: { type: "SieveSession", description: "Session object for server-side Sieve script management." },
  related: ["buildSieveScript", "close"],
  tags: ["sieve", "connection", "session"]
})

doc`
---
symbol: connectSieve
---

\`connectSieve\` opens a ManageSieve session for listing, validating, uploading, activating, and deleting scripts.
`

__doc__("close", {
  summary: "Close an IMAP or Sieve session.",
  concepts: ["session-lifecycle", "cleanup"],
  params: [],
  returns: { type: "void", description: "No return value." },
  related: ["connect", "connectSieve"],
  tags: ["session", "cleanup"]
})

doc`
---
symbol: close
---

Always call \`close\` when a script is finished with an IMAP or Sieve session.
`

__doc__("capabilities", {
  summary: "Return the capability map or capability object for the current remote session.",
  concepts: ["capabilities", "server-discovery"],
  params: [],
  returns: { type: "object", description: "Capability information from the remote IMAP or Sieve server." },
  related: ["connect", "connectSieve"],
  tags: ["imap", "sieve", "capabilities"]
})

doc`
---
symbol: capabilities
---

Use \`capabilities\` to inspect which protocol extensions the connected server advertises.
`

__doc__("list", {
  summary: "List IMAP mailboxes matching an optional pattern.",
  concepts: ["mailboxes", "listing"],
  params: [
    { name: "pattern", type: "string", description: "Optional IMAP LIST pattern; defaults to *." }
  ],
  returns: { type: "MailboxInfo[]", description: "Mailbox descriptors with name, flags, and delimiter." },
  related: ["status", "selectMailbox"],
  tags: ["imap", "mailboxes"]
})

doc`
---
symbol: list
---

\`list\` surfaces mailbox discovery from the current IMAP session.
`

__doc__("status", {
  summary: "Fetch IMAP STATUS information for a mailbox.",
  concepts: ["mailboxes", "status"],
  params: [
    { name: "name", type: "string", description: "Mailbox name." }
  ],
  returns: { type: "MailboxStatus", description: "Mailbox counts and UID metadata." },
  related: ["list", "selectMailbox"],
  tags: ["imap", "status"]
})

doc`
---
symbol: status
---

\`status\` is useful for inspection before selecting or processing a mailbox.
`

__doc__("selectMailbox", {
  summary: "Switch the active mailbox for the current IMAP session.",
  concepts: ["mailboxes", "selection"],
  params: [
    { name: "name", type: "string", description: "Mailbox name to select." },
    { name: "options", type: "object", description: "Optional selection settings such as readOnly." }
  ],
  returns: { type: "MailboxSelection", description: "Selection metadata for the newly active mailbox." },
  related: ["status", "search", "fetch"],
  tags: ["imap", "mailboxes", "selection"]
})

doc`
---
symbol: selectMailbox
---

\`selectMailbox\` changes the session's active mailbox and updates the session's \`mailbox\` property.
`

__doc__("search", {
  summary: "Run an IMAP search against the current mailbox and return matching UIDs.",
  concepts: ["search", "querying"],
  params: [
    { name: "criteria", type: "object", description: "Search criteria object with flags, text, dates, headers, or UID constraints." }
  ],
  returns: { type: "number[]", description: "Matching message UIDs." },
  related: ["fetch", "addFlags", "move"],
  tags: ["imap", "search"]
})

doc`
---
symbol: search
---

\`search\` is the main entrypoint for mailbox filtering before fetch or mutation work.
`

__doc__("fetch", {
  summary: "Fetch message summaries or bodies by UID.",
  concepts: ["fetch", "messages"],
  params: [
    { name: "uids", type: "number[]", description: "UIDs to fetch." },
    { name: "fields", type: "string[]", description: "Requested fetch fields such as uid, flags, envelope, or body.text." }
  ],
  returns: { type: "FetchedMessage[]", description: "Structured message records." },
  related: ["search", "append"],
  tags: ["imap", "fetch", "messages"]
})

doc`
---
symbol: fetch
---

\`fetch\` returns JSON-shaped message objects keyed by the request fields.
`

__doc__("addFlags", {
  summary: "Add flags to one or more messages.",
  concepts: ["flags", "message-mutation"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "flags", type: "string[]", description: "Flags to add." },
    { name: "options", type: "object", description: "Optional mutation settings such as silent." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["removeFlags", "setFlags"],
  tags: ["imap", "flags"]
})

doc`
---
symbol: addFlags
---

\`addFlags\` performs an IMAP STORE +FLAGS operation.
`

__doc__("removeFlags", {
  summary: "Remove flags from one or more messages.",
  concepts: ["flags", "message-mutation"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "flags", type: "string[]", description: "Flags to remove." },
    { name: "options", type: "object", description: "Optional mutation settings such as silent." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["addFlags", "setFlags"],
  tags: ["imap", "flags"]
})

doc`
---
symbol: removeFlags
---

\`removeFlags\` performs an IMAP STORE -FLAGS operation.
`

__doc__("setFlags", {
  summary: "Replace the flags on one or more messages.",
  concepts: ["flags", "message-mutation"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "flags", type: "string[]", description: "Complete flag set." },
    { name: "options", type: "object", description: "Optional mutation settings such as silent." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["addFlags", "removeFlags"],
  tags: ["imap", "flags"]
})

doc`
---
symbol: setFlags
---

\`setFlags\` performs an IMAP STORE FLAGS operation.
`

__doc__("move", {
  summary: "Move one or more messages to another mailbox.",
  concepts: ["message-mutation", "mailboxes"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "destination", type: "string", description: "Target mailbox." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["copy", "delete"],
  tags: ["imap", "move"]
})

doc`
---
symbol: move
---

\`move\` relocates messages to another mailbox.
`

__doc__("copy", {
  summary: "Copy one or more messages to another mailbox.",
  concepts: ["message-mutation", "mailboxes"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "destination", type: "string", description: "Target mailbox." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["move", "delete"],
  tags: ["imap", "copy"]
})

doc`
---
symbol: copy
---

\`copy\` duplicates messages into another mailbox without removing the originals.
`

__doc__("delete", {
  summary: "Mark one or more messages as deleted and optionally expunge them.",
  concepts: ["message-mutation", "deletion"],
  params: [
    { name: "uids", type: "number[]", description: "Message UIDs." },
    { name: "options", type: "object", description: "Optional delete settings such as expunge." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["expunge", "move"],
  tags: ["imap", "delete"]
})

doc`
---
symbol: delete
---

\`delete\` adds the \`\\Deleted\` flag and can immediately expunge the affected messages.
`

__doc__("expunge", {
  summary: "Expunge deleted messages from the current mailbox.",
  concepts: ["deletion", "cleanup"],
  params: [
    { name: "uids", type: "number[]", description: "Optional UID list when the server supports UID EXPUNGE." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["delete"],
  tags: ["imap", "expunge"]
})

doc`
---
symbol: expunge
---

\`expunge\` removes messages already marked deleted.
`

__doc__("append", {
  summary: "Append a raw RFC822 message to a mailbox.",
  concepts: ["message-creation", "mailboxes"],
  params: [
    { name: "content", type: "string", description: "Raw RFC822 message bytes represented as a string." },
    { name: "options", type: "object", description: "Optional append settings such as mailbox, flags, and date." }
  ],
  returns: { type: "number", description: "The appended message UID when the server returns one." },
  related: ["fetch"],
  tags: ["imap", "append"]
})

doc`
---
symbol: append
---

\`append\` is the low-level message creation API for the current IMAP session.
`

__doc__("listScripts", {
  summary: "List Sieve scripts currently stored on the server.",
  concepts: ["sieve", "script-management"],
  params: [],
  returns: { type: "ScriptInfo[]", description: "Script names plus active-state metadata." },
  related: ["getScript", "putScript"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: listScripts
---

\`listScripts\` shows which scripts exist and which one is active.
`

__doc__("getScript", {
  summary: "Fetch the source of a Sieve script by name.",
  concepts: ["sieve", "script-management"],
  params: [
    { name: "name", type: "string", description: "Script name." }
  ],
  returns: { type: "string", description: "Script source." },
  related: ["listScripts", "putScript"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: getScript
---

\`getScript\` retrieves the raw source of a stored Sieve script.
`

__doc__("putScript", {
  summary: "Upload a Sieve script to the server and optionally activate it.",
  concepts: ["sieve", "script-management"],
  params: [
    { name: "name", type: "string", description: "Script name." },
    { name: "content", type: "string", description: "Script source." },
    { name: "options", type: "object", description: "Optional upload settings such as activate." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["buildSieveScript", "activate", "check"],
  tags: ["sieve", "scripts", "upload"]
})

doc`
---
symbol: putScript
---

\`putScript\` uploads a script to the ManageSieve server.
`

__doc__("activate", {
  summary: "Activate a named Sieve script.",
  concepts: ["sieve", "script-management"],
  params: [
    { name: "name", type: "string", description: "Script name." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["deactivate", "putScript"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: activate
---

\`activate\` makes the named Sieve script the active server-side script.
`

__doc__("deactivate", {
  summary: "Deactivate the currently active Sieve script.",
  concepts: ["sieve", "script-management"],
  params: [],
  returns: { type: "void", description: "No return value." },
  related: ["activate"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: deactivate
---

\`deactivate\` clears the active script selection when the server allows it.
`

__doc__("deleteScript", {
  summary: "Delete a named Sieve script.",
  concepts: ["sieve", "script-management"],
  params: [
    { name: "name", type: "string", description: "Script name." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["renameScript", "putScript"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: deleteScript
---

\`deleteScript\` removes a stored Sieve script from the server.
`

__doc__("renameScript", {
  summary: "Rename a stored Sieve script.",
  concepts: ["sieve", "script-management"],
  params: [
    { name: "oldName", type: "string", description: "Existing script name." },
    { name: "newName", type: "string", description: "New script name." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["deleteScript", "putScript"],
  tags: ["sieve", "scripts"]
})

doc`
---
symbol: renameScript
---

\`renameScript\` changes the remote name of an existing Sieve script.
`

__doc__("check", {
  summary: "Validate Sieve script content without storing it.",
  concepts: ["sieve", "validation"],
  params: [
    { name: "content", type: "string", description: "Script source." }
  ],
  returns: { type: "void", description: "No return value." },
  related: ["buildSieveScript", "putScript"],
  tags: ["sieve", "validation"]
})

doc`
---
symbol: check
---

\`check\` validates syntax and server acceptance of a Sieve script.
`

__doc__("haveSpace", {
  summary: "Ask the server whether a script of the given size can be stored.",
  concepts: ["sieve", "validation"],
  params: [
    { name: "name", type: "string", description: "Script name." },
    { name: "sizeBytes", type: "number", description: "Script size in bytes." }
  ],
  returns: { type: "boolean", description: "True when the server reports enough storage space." },
  related: ["putScript", "check"],
  tags: ["sieve", "capacity"]
})

doc`
---
symbol: haveSpace
---

\`haveSpace\` is the capacity preflight helper for Sieve uploads.
`
