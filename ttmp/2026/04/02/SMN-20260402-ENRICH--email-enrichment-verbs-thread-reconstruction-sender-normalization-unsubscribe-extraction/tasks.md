# Tasks

## TODO

- [ ] Add tasks here

- [x] Add migration v2 schema (pkg/enrich/schema.go + wire into pkg/mirror/schema.go)
- [x] Implement pkg/enrich/types.go — Options, report types
- [x] Implement parse_address.go + tests (from_summary parser, domain, private relay)
- [x] Implement parse_headers.go + tests (GetHeader, ParseListUnsubscribe)
- [ ] Implement SenderEnricher + integration test
- [ ] Implement ThreadEnricher + integration test
- [ ] Implement UnsubscribeEnricher + integration test
- [ ] Implement RunAll (pkg/enrich/all.go)
- [ ] Implement cmd/smailnail/commands/enrich/ Glazed verbs + group root
- [ ] Wire enrich group into smailnail main.go
- [ ] Add --enrich-after flag to mirror command
