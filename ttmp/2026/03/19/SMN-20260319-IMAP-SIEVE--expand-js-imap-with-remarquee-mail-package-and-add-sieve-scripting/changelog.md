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


## 2026-03-19 - Ticket bookkeeping synced to implementation (commit bc4baaa)

Updated the task matrix, diary, changelog, and related-file metadata so the ticket now reflects the landed implementation commits and the current follow-up boundary for hosted-account Sieve settings.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/tasks.md — Execution status updated after the implementation commits
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/reference/01-diary.md — Detailed implementation diary with failures, commands, and commit hashes
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/changelog.md — Commit-indexed implementation milestones
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/.ttmp.yaml — Repo-local docmgr workspace configuration now tracked in git
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/vocabulary.yaml — Repo-local topic vocabulary tracked with the ticket workspace


## 2026-03-19 - Starter documentation and example expansion (commit c6fd9c6)

Expanded the ticket guide with an explicit explanation of the remaining hosted-account Sieve follow-up and added a much larger starter example set so new users can copy working IMAP and Sieve patterns immediately.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ttmp/2026/03/19/SMN-20260319-IMAP-SIEVE--expand-js-imap-with-remarquee-mail-package-and-add-sieve-scripting/design-doc/01-intern-guide-expanding-js-imap-and-adding-a-sieve-scripting-layer.md — Open-item explanation plus quickstart cookbook
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/docs/examples.js — Expanded starter example set for IMAP and Sieve flows


## 2026-03-19 - PR review and CI cleanup (commit 91dd372)

Addressed the open PR review findings and the hosted-app CI failures by making account default/deletion mutations transactional, scoping account-test deletion by owning user, hardening OIDC cookie security, fixing unsafe JS-to-UID integer conversions, using IPv6-safe ManageSieve dialing, preserving shutdown context values without `context.Background()`, and making `go generate` reuse committed frontend assets when package managers are unavailable in CI.

### Related Files

- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/repository.go — Transactional create/update/delete logic and tenant-safe test deletion
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service.go — Service now routes default-account mutations through transactional repository helpers
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/accounts/service_internal_test.go — Regression tests for failed default reassignment and cross-tenant delete safety
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/oidc.go — Secure-cookie handling now tracks HTTPS transport/redirect configuration
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/auth/oidc_test.go — Cookie security regression coverage
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/js/modules/smailnail/js_helpers.go — Bounded JS integer conversion for UID-style arrays
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/imap_client.go — Bounded UID parsing for JS array inputs
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/mailruntime/sieve_client.go — IPv6-safe ManageSieve dialing
- /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/pkg/smailnaild/web/generate_build.go — CI-friendly frontend-build fallback when `pnpm` is unavailable
