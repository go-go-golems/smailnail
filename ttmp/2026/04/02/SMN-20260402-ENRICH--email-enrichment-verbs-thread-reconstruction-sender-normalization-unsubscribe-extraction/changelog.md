# Changelog

## 2026-04-02

- Initial workspace created


## 2026-04-02

Initial analysis and full design doc written: schema migration v2, ThreadEnricher, SenderEnricher, UnsubscribeEnricher, enrich command group, --enrich-after on mirror. 11 implementation tasks created.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/ttmp/2026/04/02/SMN-20260402-ENRICH--email-enrichment-verbs-thread-reconstruction-sender-normalization-unsubscribe-extraction/design-doc/01-analysis-design-and-implementation-plan.md — Full analysis


## 2026-04-02

Implemented enrichment schema v2, shared report/options types, and parser helpers for from_summary and List-Unsubscribe.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_address.go — Sender parsing and relay helpers
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/parse_headers.go — Header JSON and List-Unsubscribe parsing
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/schema.go — Schema migration statements for enrichment tables and columns
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/types.go — Shared enrich options and report structs
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/mirror/schema.go — Wire migration v2 into mirror bootstrap


## 2026-04-02

Implemented SenderEnricher with transactional sender upserts, message tagging, sender metadata watermarking, and incremental integration tests.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/common.go — Shared scope and metadata helpers for enrichers
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders.go — SenderEnricher implementation and SQL updates
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/senders_test.go — Integration coverage for first-run and incremental sender enrichment


## 2026-04-02

Implemented ThreadEnricher with parent-chain reconstruction, thread summary upserts, and incremental integration tests.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/threads.go — Thread graph resolution and threads summary writes
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/threads_test.go — Integration coverage for root resolution and incremental updates


## 2026-04-02

Implemented unsubscribe enrichment and RunAll orchestration with integration tests covering latest-link selection and full pass execution.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/all.go — RunAll orchestration across senders
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/all_test.go — Integration coverage for RunAll end-to-end execution
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/unsubscribe.go — List-Unsubscribe extraction and sender upserts
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/pkg/enrich/unsubscribe_test.go — Integration coverage for unsubscribe extraction and incremental skipping


## 2026-04-02

Added the smailnail enrich command group with senders, threads, unsubscribe, and all Glazed verbs, and registered it in the CLI root.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/all.go — Glazed all-in-one enrichment verb
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/common.go — Shared command settings and DB bootstrap helper
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/root.go — Cobra group for enrich verbs
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/senders.go — Glazed sender enrichment verb
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/threads.go — Glazed thread enrichment verb
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/enrich/unsubscribe.go — Glazed unsubscribe enrichment verb
- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/main.go — Register enrich group in smailnail root command


## 2026-04-02

Added mirror --enrich-after to run the full enrichment pipeline after a successful sync and expose the summary counts in mirror output.

### Related Files

- /home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail/cmd/smailnail/commands/mirror.go — Optional post-sync enrichment hook and output fields

