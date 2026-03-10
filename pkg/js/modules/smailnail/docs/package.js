__package__({
  name: "smailnail",
  title: "smailnail IMAP JavaScript API",
  category: "imap",
  description: "JavaScript API for rule construction and IMAP session handling through the smailnail Go services."
})

doc`
---
package: smailnail
---

The \`smailnail\` package exposes a small JavaScript surface over the Go IMAP tooling.

Use it when you want to:

- build or parse rule YAML from JavaScript,
- open IMAP sessions with validated options,
- call the same service layer used by the CLI code,
- drive the module through a dedicated MCP server.
`
