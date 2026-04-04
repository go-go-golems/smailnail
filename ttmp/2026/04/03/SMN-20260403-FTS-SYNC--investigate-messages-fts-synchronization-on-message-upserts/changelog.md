# Changelog

## 2026-04-03

- Initial workspace created


## 2026-04-03

Investigation complete: GitHub issue claiming messages_fts not synced during mirror upserts is a false positive for the main sync path (service.go:678-683 correctly pairs upsertMessageRecord + upsertMessageFTS). However, found a real bug: resetMailboxState (service.go:868) deletes messages without cleaning up messages_fts, leaving orphaned FTS rows after UIDVALIDITY changes or manual resets. Report and implementation plan written.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/service.go — resetMailboxState missing FTS cleanup (L877)

