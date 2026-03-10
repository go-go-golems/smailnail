__example__({
  id: "build-rule-basic",
  title: "Build a simple invoice search rule",
  symbols: ["buildRule"],
  concepts: ["rule-building"]
})
function buildRuleBasicExample() {
  const smailnail = require("smailnail")
  return smailnail.buildRule({
    name: "invoice-search",
    subjectContains: "invoice",
    includeContent: true,
    contentType: "text/plain"
  })
}

__example__({
  id: "connect-basic",
  title: "Open and close an IMAP session",
  symbols: ["newService", "connect", "close"],
  concepts: ["imap-connection", "session-lifecycle"]
})
function connectBasicExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    server: "imap.example.com",
    username: "user@example.com",
    password: "secret",
    mailbox: "INBOX"
  })
  session.close()
}
