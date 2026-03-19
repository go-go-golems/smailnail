# Changelog

## 2026-03-19

- Initial workspace created


## 2026-03-19 - Initial analysis and design guide

Configured smailnail as a local docmgr workspace, created the SMN-20260319-IMAP-SIEVE ticket, mapped the current JS/IMAP/MCP/hosted-account architecture, and authored a detailed intern-oriented design and implementation guide for expanding JS IMAP with donor remarquee mail code plus a sieve layer.

### Related Files

- /home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_imap.go — Donor JS IMAP capability source
- /home/manuel/workspaces/2026-03-04/fix-remarquee-oauth-refresh/remarquee/pkg/mail/js_sieve.go — Donor JS sieve capability source
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go — Current runtime surface that motivates the expansion
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/identity_middleware.go — MCP account-resolution path preserved in the design
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go — Current orchestration layer and proposed service growth point
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/db.go — Schema migration anchor for future sieve account fields


## 2026-03-19 - Validation and delivery

Added missing topic vocabulary for the new smailnail docmgr workspace, passed docmgr doctor cleanly, dry-ran the reMarkable bundle upload, published the final PDF bundle, and verified the remote listing under /ai/2026/03/19/SMN-20260319-IMAP-SIEVE.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/design-doc/01-intern-guide-expanding-js-imap-and-adding-a-sieve-scripting-layer.md — Primary deliverable included in the published bundle
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/reference/01-diary.md — Recorded validation and delivery commands and outcomes
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/vocabulary.yaml — Added repository-local topic slugs required by the ticket metadata


## 2026-03-19 - Runtime and service implementation (commit 439258f71f655e80c664583cfe4e8c33041ea76a)

Ported the donor IMAP and ManageSieve runtime into a new `pkg/mailruntime` package, expanded `pkg/services/smailnailjs` to expose richer IMAP session methods and a new Sieve connection flow, and updated the fake-backed service and MCP account-resolution tests to match the new service contract.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/imap_client.go — Donor-derived IMAP runtime port
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/sieve_client.go — Donor-derived ManageSieve runtime port
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service.go — Expanded shared service contract and real dialers
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/services/smailnailjs/service_test.go — Fake-backed IMAP and Sieve service coverage
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/execute_tool_account_test.go — Updated MCP stored-account fake session contract


## 2026-03-19 - JS runtime and documentation expansion (commit e66bd4545660a065f9bea828a9e2a1adf1565536)

Expanded the `smailnail` goja module to expose richer IMAP session methods, ManageSieve session methods, and an offline Sieve script builder; updated the embedded docs/examples to match the new callable surface; and refreshed MCP documentation tests for the new example set.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module.go — Richer IMAP and Sieve goja export layer
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/js_helpers.go — JS argument parsing and JSON-shape normalization
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/sieve_builder.go — Offline Sieve builder DSL
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/module_test.go — goja runtime and documentation parity coverage
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/service.js — Symbol docs for the expanded runtime surface
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js — Updated IMAP and Sieve examples
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mcp/imapjs/docs_tool_test.go — Documentation query expectations for the richer example set
