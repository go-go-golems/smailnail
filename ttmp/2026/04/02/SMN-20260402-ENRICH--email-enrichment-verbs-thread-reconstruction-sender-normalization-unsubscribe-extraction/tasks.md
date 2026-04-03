# Tasks

## TODO

- [x] Add migration v2 schema (pkg/enrich/schema.go + wire into pkg/mirror/schema.go)
- [x] Implement pkg/enrich/types.go — Options, report types
- [x] Implement parse_address.go + tests (from_summary parser, domain, private relay)
- [x] Implement parse_headers.go + tests (GetHeader, ParseListUnsubscribe)
- [x] Implement SenderEnricher + integration test
- [x] Implement ThreadEnricher + integration test
- [x] Implement UnsubscribeEnricher + integration test
- [x] Implement RunAll (pkg/enrich/all.go)
- [x] Implement cmd/smailnail/commands/enrich/ Glazed verbs + group root
- [x] Wire enrich group into smailnail main.go
- [x] Add --enrich-after flag to mirror command
