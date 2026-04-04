# Tasks

## TODO

- [ ] Add tasks here

- [ ] Fix: Add DELETE FROM messages_fts to resetMailboxState (service.go:868)
- [ ] Test: Assert FTS row count after UIDVALIDITY-triggered reset + re-sync (no orphan rows)
- [ ] Optional: Consider merging upsertMessageFTS into upsertMessageRecord for safety
- [ ] Close GitHub issue with explanation: main sync path was already correct, resetMailboxState gap fixed
