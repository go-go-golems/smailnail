# Tasks

## TODO

- [x] Spec audit: confirm endpoint contract against ui/src/api/annotations.ts, ui/src/types/annotations.ts, and ui/src/mocks/handlers.ts
- [x] Create sqlite backend ticket docs: implementation guide, diary, related-file links, and changelog scaffolding
- [ ] Extend pkg/annotate types and repository filters for agentRunId-aware listing
- [ ] Implement batch annotation review updates in pkg/annotate with repository tests
- [ ] Implement agent run summary/detail aggregation in pkg/annotate with repository tests
- [ ] Create a dedicated sqlite annotation web package with JSON helpers, health/info routes, and SPA/root redirect behavior
- [ ] Implement annotation/group/log/run HTTP handlers returning bare JSON and add handler tests
- [ ] Implement sender list/detail queries over senders/messages/annotations/logs and add handler tests
- [ ] Implement read-only SQL execution, preset query loading, and saved-query filesystem persistence with tests
- [ ] Add cmd/smailnail sqlite serve and wire it into the root CLI
- [ ] Smoke the sqlite server in tmux against a mirror sqlite database and verify the key API routes
- [ ] Run go test ./... and close the ticket slice with diary, changelog, task updates, and focused commits
