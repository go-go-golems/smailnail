# Changelog

## 2026-04-02

- Initial workspace created
- Added a detailed design and implementation guide for a first-class mirror shard merge verb, covering current architecture, merge invariants, API design, phased implementation, and testing strategy.
- Added an investigation diary documenting the evidence gathered from the mirror command, mirror service, schema, raw-message layout, and month-sharded backfill scripts.
- Related the key mirror files to the ticket and prepared the bundle for reMarkable delivery.
- Locked the v1 product decisions for implementation: support `--enrich-after`, use root-based shard discovery only, and warn by default for missing raw source files.
- Implemented the first merge slice: a new `merge-mirror-shards` command, root-based shard discovery, shard inspection, dry-run reporting, and initial tests (`a6a9099`).
- Implemented shard merge execution, including destination bootstrap, row upserts, raw-file copy/reuse with warning-by-default missing raw handling, mailbox sync state rebuild, FTS rebuild, and post-merge enrichment support (`4acad07`).
