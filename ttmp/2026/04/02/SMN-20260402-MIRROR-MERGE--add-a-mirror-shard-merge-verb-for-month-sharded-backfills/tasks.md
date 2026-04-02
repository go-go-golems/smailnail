# Tasks

## TODO

- [x] Create the ticket workspace and seed the core design-doc/reference documents.
- [x] Inspect the current mirror command, mirror service, schema, raw-file layout, and month-sharded backfill scripts.
- [x] Write a detailed intern-facing analysis, design, and implementation guide for a merge-shards verb.
- [x] Record the investigation in the ticket diary and capture the major design decisions.
- [x] Relate the most important implementation files to the design doc and diary.
- [x] Run `docmgr doctor` for the ticket and fix any documentation hygiene issues.
- [x] Upload the ticket bundle to reMarkable and verify the remote listing.

## Implementation

- [x] Lock the v1 product decisions in the design docs: `--enrich-after` on merge, missing raw files warn by default, root-based shard discovery only.
- [x] Add the `merge-mirror-shards` Glazed command and wire it into `cmd/smailnail/main.go`.
- [x] Implement shard discovery and dry-run inspection reporting in `pkg/mirror`.
- [ ] Implement canonical message-row merge and destination upsert behavior.
- [ ] Implement raw-file copy/reuse logic with warning-by-default handling for missing source files.
- [ ] Rebuild `mailbox_sync_state` from merged destination rows.
- [ ] Rebuild `messages_fts` from the merged `messages` table.
- [ ] Add optional `--enrich-after` support for the merge verb.
- [ ] Add focused unit tests for discovery, conflicts, raw warnings, state rebuild, FTS rebuild, and enrich-after behavior.
- [ ] Validate with targeted `go test` runs and at least one end-to-end local shard merge smoke.
- [ ] Update the ticket diary, changelog, and implementation docs after each completed slice.
