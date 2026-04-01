# Changelog

## 2026-04-01

- Initial workspace created
- Completed an evidence-backed architecture review of the current CLI, DSL, `mailruntime`, JS module, hosted account services, and SQLite bootstrapping surfaces relevant to IMAP mirroring
- Wrote the primary design guide and implementation plan for a new glazed IMAP mirror verb with local SQLite indexing

## 2026-04-01

Validated the ticket with docmgr doctor, added missing topic vocabulary, and uploaded the documentation bundle to /ai/2026/04/01/SMN-20260401-IMAP-MIRROR on reMarkable

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/design-doc/01-intern-guide-designing-an-imap-mirror-verb-with-sqlite-indexing.md — Primary design guide validated and delivered
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/ttmp/2026/04/01/SMN-20260401-IMAP-MIRROR--add-a-glazed-imap-mirror-verb-with-sqlite-indexing/reference/01-diary.md — Diary updated and included in the uploaded bundle


## 2026-04-01

Step 3: landed the mirror command and local store bootstrap (commit 1d9578a08372607e77e4de17bb95a1b75522568d)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go — New mirror command scaffold and report output
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go — New mirror schema
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/store.go — New local SQLite store bootstrap


## 2026-04-01

Step 4: added incremental raw-message sync, fixed the UIDNEXT search boundary, and validated against Docker Dovecot (commit 9b0afe7a06542be44f8ae87f397c446232ec8efb)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go — Mirror command now runs sync and reports aggregate counters
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mailruntime/imap_client.go — TLS insecure support for local self-signed IMAP fixtures
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/files.go — Raw-message file layout and idempotent writes
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/service.go — Incremental UID sync

