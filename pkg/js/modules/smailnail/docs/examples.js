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
  id: "imap-session-automation",
  title: "Search, fetch, and relabel mail from JavaScript",
  symbols: ["newService", "connect", "search", "fetch", "addFlags", "move", "close"],
  concepts: ["imap-connection", "mailbox-automation"]
})
function imapSessionAutomationExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    server: "imap.example.com",
    username: "user@example.com",
    password: "secret",
    mailbox: "INBOX"
  })

  const uids = session.search({ subject: "invoice", unseen: true })
  const messages = session.fetch(uids, ["uid", "flags", "body.text"])
  session.addFlags(uids, ["\\Seen"])
  session.move(uids, "Processed/Invoices")
  session.close()

  return messages
}

__example__({
  id: "sieve-script-management",
  title: "Build and upload a Sieve script",
  symbols: ["newService", "buildSieveScript", "connectSieve", "putScript", "activate", "check", "close"],
  concepts: ["sieve", "script-builder", "script-management"]
})
function sieveScriptManagementExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const script = service.buildSieveScript((s) => {
    s.require(["fileinto"])
    s.if(s.headerContains("Subject", "invoice"), (a) => {
      a.fileInto("Invoices")
      a.stop()
    })
  })

  const sieve = service.connectSieve({
    server: "sieve.example.com",
    username: "user@example.com",
    password: "secret"
  })
  sieve.check(script)
  sieve.putScript("main", script, { activate: true })
  sieve.close()

  return script
}
