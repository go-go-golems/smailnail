__package__({
  name: "smailnail",
  title: "smailnail IMAP JavaScript API",
  category: "imap",
  description: "JavaScript API for rule construction, IMAP mailbox automation, and ManageSieve scripting through the smailnail Go services."
})

doc`
---
package: smailnail
---

The \`smailnail\` package exposes a small JavaScript surface over the Go IMAP tooling.

Use it when you want to:

- build or parse rule YAML from JavaScript,
- open IMAP sessions with search, fetch, flag, move, copy, delete, expunge, and append operations,
- open ManageSieve sessions for script management,
- build Sieve scripts with a fluent builder,
- call the same service layer used by the CLI code,
- drive the module through a dedicated MCP server.
`
