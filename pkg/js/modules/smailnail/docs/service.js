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

Use \`parseRule\` when a script already has a YAML rule definition and needs the normalized object form.

This is useful for validation, inspection, and transformation before handing the rule to other code.
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

\`buildRule\` is the main ergonomic entrypoint for JavaScript callers.

It accepts the lower-camel option keys exposed by \`pkg/services/smailnailjs\`, validates the resulting rule, and returns the same object view used elsewhere in the module.
`

__doc__("newService", {
  summary: "Create a service object exposing rule helpers and IMAP connection methods.",
  concepts: ["service-construction", "imap-runtime"],
  params: [],
  returns: { type: "Service", description: "Service object with parseRule, buildRule, and connect methods." },
  related: ["connect", "close"],
  tags: ["service", "runtime"]
})

doc`
---
symbol: newService
---

Create a service when a script needs both rule helpers and live IMAP session access.

The returned object is the main stateful surface in the JavaScript API.
`

__doc__("connect", {
  summary: "Open an IMAP session using ConnectOptions and return a session wrapper.",
  concepts: ["imap-connection", "session-lifecycle"],
  params: [
    { name: "options", type: "ConnectOptions", description: "Server, credentials, mailbox, and TLS settings." }
  ],
  returns: { type: "Session", description: "Session object with mailbox metadata and a close method." },
  related: ["newService", "close"],
  tags: ["imap", "connection", "session"]
})

doc`
---
symbol: connect
---

\`connect\` validates connection options, dials the IMAP server, selects the mailbox, and returns a lightweight session wrapper.

Always call \`close\` on the returned session once the script is done with the mailbox.
`

__doc__("close", {
  summary: "Close an IMAP session returned by connect.",
  concepts: ["session-lifecycle", "cleanup"],
  params: [],
  returns: { type: "void", description: "No return value." },
  related: ["connect"],
  tags: ["imap", "session", "cleanup"]
})

doc`
---
symbol: close
---

\`close\` releases the underlying IMAP client owned by the session wrapper.

This should be the last operation in any script that opens a mailbox session.
`
