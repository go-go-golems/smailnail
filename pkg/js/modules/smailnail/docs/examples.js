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
  id: "connect-with-account-id",
  title: "Open IMAP using a stored account",
  symbols: ["newService", "connect", "close"],
  concepts: ["imap-connection", "account-resolution"]
})
function connectWithAccountIDExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    accountId: "acc_123",
    mailbox: "INBOX"
  })

  try {
    return session.capabilities()
  } finally {
    session.close()
  }
}

__example__({
  id: "mailbox-discovery",
  title: "List mailboxes and inspect status",
  symbols: ["newService", "connect", "list", "status", "selectMailbox", "close"],
  concepts: ["mailboxes", "status"]
})
function mailboxDiscoveryExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    server: "imap.example.com",
    username: "user@example.com",
    password: "secret",
    mailbox: "INBOX"
  })

  try {
    const boxes = session.list()
    const archiveStatus = session.status("Archive")
    const selected = session.selectMailbox("Archive", { readOnly: true })
    return { boxes, archiveStatus, selected }
  } finally {
    session.close()
  }
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
  id: "copy-delete-expunge",
  title: "Copy, delete, and expunge messages",
  symbols: ["newService", "connect", "search", "copy", "delete", "expunge", "close"],
  concepts: ["message-mutation", "deletion"]
})
function copyDeleteExpungeExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    server: "imap.example.com",
    username: "user@example.com",
    password: "secret",
    mailbox: "INBOX"
  })

  try {
    const uids = session.search({ from: "alerts@example.com", seen: true })
    session.copy(uids, "Archive/Alerts")
    session.delete(uids, { expunge: false })
    session.expunge()
    return { copied: uids.length }
  } finally {
    session.close()
  }
}

__example__({
  id: "append-message",
  title: "Append a raw RFC822 message",
  symbols: ["newService", "connect", "append", "close"],
  concepts: ["message-creation", "mailboxes"]
})
function appendMessageExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const session = service.connect({
    server: "imap.example.com",
    username: "user@example.com",
    password: "secret",
    mailbox: "Drafts"
  })

  try {
    const raw = [
      "From: user@example.com",
      "To: user@example.com",
      "Subject: Draft from smailnail",
      "",
      "Hello from the append API."
    ].join("\\r\\n")

    return session.append(raw, {
      mailbox: "Drafts",
      flags: ["\\Draft"]
    })
  } finally {
    session.close()
  }
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

__example__({
  id: "sieve-list-and-read",
  title: "List and read server-side Sieve scripts",
  symbols: ["newService", "connectSieve", "listScripts", "getScript", "close"],
  concepts: ["sieve", "script-management"]
})
function sieveListAndReadExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const sieve = service.connectSieve({
    accountId: "acc_123"
  })

  try {
    const scripts = sieve.listScripts()
    const first = scripts[0]
    return {
      scripts,
      source: first ? sieve.getScript(first.name) : ""
    }
  } finally {
    sieve.close()
  }
}

__example__({
  id: "sieve-vacation-rule",
  title: "Build a vacation responder script",
  symbols: ["newService", "buildSieveScript"],
  concepts: ["sieve", "script-builder"]
})
function sieveVacationRuleExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  return service.buildSieveScript((s) => {
    s.require(["vacation"])
    s.if(s.true(), (a) => {
      a.vacation({
        days: 7,
        subject: "Out of office",
        message: "I am away and will respond when I return."
      })
      a.stop()
    })
  })
}

__example__({
  id: "sieve-rename-delete",
  title: "Rename and delete a Sieve script",
  symbols: ["newService", "connectSieve", "renameScript", "deleteScript", "close"],
  concepts: ["sieve", "script-management"]
})
function sieveRenameDeleteExample() {
  const smailnail = require("smailnail")
  const service = smailnail.newService()
  const sieve = service.connectSieve({
    server: "sieve.example.com",
    username: "user@example.com",
    password: "secret"
  })

  try {
    sieve.renameScript("main", "main-old")
    sieve.deleteScript("main-old")
  } finally {
    sieve.close()
  }
}
