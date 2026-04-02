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

