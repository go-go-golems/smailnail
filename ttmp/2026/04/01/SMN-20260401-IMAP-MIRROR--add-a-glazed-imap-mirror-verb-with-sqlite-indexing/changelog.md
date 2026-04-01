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


## 2026-04-01

Step 5: parsed raw RFC 822 messages into body_text, body_html, parts_json, and search_text, and validated multipart HTML mirroring against Docker Dovecot (commit f30a4c432200b77456cb116f4443477c4d8759e3)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/parser.go — New raw-message parser and search-text normalization
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/parser_test.go — New parser and projection tests
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/service.go — Message records now prefer parsed raw-message fields


## 2026-04-01

Step 6: required SQLite FTS5 build tags at compile time, updated tagged validation entry points, and reran the Docker Dovecot smoke with the tagged `smailnail` CLI (commit d2bed23557ada03540fbf4fc4e1f393df9fdfcbb)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/require_fts5_build_tag.go — Compile-time guard that fails untagged mirror builds
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/Makefile — Default tagged build, test, lint, and install targets
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/scripts/docker-imap-smoke.sh — Docker smoke now runs `smailnail` with the required build tag


## 2026-04-01

Step 7: removed the runtime FTS fallback and deleted the `--search-mode` split so mirror bootstrap now assumes a single FTS-backed contract (commit 215920ddf1ec71cbee377ff6624615e861a1acf8)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go — Mirror command no longer exposes a dead `--search-mode` flag
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go — Schema bootstrap now requires FTS table creation unconditionally
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/store.go — Store bootstrap no longer accepts a runtime search-mode selector


## 2026-04-01

Step 8: normalized parsed header projections so `headers_json` now prefers semantic raw-message values for addresses and `Message-Id` (commit bb97160ae5d9bd89af0233f2bf9bda6ba46fc2af)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/parser.go — Raw parser now emits canonicalized address summaries and normalized parsed headers
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/service.go — Mirrored rows now prefer normalized parsed header maps over fetched header maps
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/parser_test.go — Parser and record-projection tests now cover normalized header output


## 2026-04-01

Step 9: added opt-in full-mailbox reconciliation so mirror can tombstone rows as `remote_deleted` after remote deletes or expunges (commit f0aa4292d39d1da6240f2ec66ef068e28a7ae534)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go — Mirror command now exposes `--reconcile-full-mailbox` and reports tombstone counters
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/service.go — Full-mailbox reconcile and `remote_deleted` updates
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/service_test.go — New tombstone and restore reconcile coverage


## 2026-04-01

Step 10: wired embedded Glazed help pages into the CLI and added operator-facing mirror help entries for overview, first sync, and maintenance workflows (commit 00c5faeabd7aabca28a10c21fdf88082964a2b65)

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/main.go — Root command now loads embedded help sections before registering the Glazed help system
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/docs/embed.go — New embedded help loader for `cmd/smailnail/docs`
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/docs/mirror-overview.md — General-topic help entry for mirror architecture and flags
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/docs/mirror-first-sync-tutorial.md — Tutorial help entry for the first mirror run
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/docs/mirror-maintenance.md — Maintenance and reconcile help entry for existing mirrors
