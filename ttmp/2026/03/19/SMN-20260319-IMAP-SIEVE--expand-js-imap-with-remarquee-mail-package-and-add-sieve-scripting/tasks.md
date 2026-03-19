# Tasks

## Ticket Setup And Research

- [x] Configure `smailnail` as a local `docmgr` workspace and create the ticket scaffold
- [x] Create the primary design doc and the diary document
- [x] Map the current `smailnail` IMAP, JS, MCP, and hosted-account architecture with file-backed evidence
- [x] Compare the current implementation with `remarquee/pkg/mail` IMAP and Sieve capabilities
- [x] Write the intern-oriented analysis/design/implementation guide with prose, diagrams, pseudocode, API sketches, and file references
- [x] Relate the key source files to the ticket documents and update changelog/index metadata
- [x] Validate the ticket with `docmgr doctor`
- [x] Dry-run and complete the reMarkable bundle upload

## Phase 1: Granular Implementation Planning

- [x] Replace the research checklist with a granular engineering task list that can be executed step by step
- [x] Capture a clean baseline for the current JS module and MCP tests before implementation
- [x] Keep the task list synchronized with real implementation progress as commits land

## Phase 2: Shared Runtime Port

- [x] Create `pkg/mailruntime` as the transplant boundary for donor mail code
- [x] Port shared mail types and structured error types from `remarquee/pkg/mail`
- [x] Port the IMAP client wrapper and supporting fetch/search helpers from `remarquee/pkg/mail`
- [x] Port the ManageSieve client wrapper and protocol helpers from `remarquee/pkg/mail`
- [x] Adjust imports, logging, and package names so the runtime compiles cleanly inside `smailnail`
- [ ] Add focused unit tests for any new parser/helper logic that can be verified without network access

## Phase 3: Service-Layer Expansion

- [x] Expand `pkg/services/smailnailjs.ConnectOptions` and normalization around the richer runtime
- [x] Introduce richer IMAP session interfaces for list, status, search, fetch, flag mutation, append, move/copy/delete, and expunge operations
- [x] Add Sieve connection options, session interfaces, and service wiring
- [x] Preserve stored-account resolution for IMAP connections while keeping the new Sieve path explicit
- [x] Add service-layer tests for the richer IMAP and Sieve flows using fakes

## Phase 4: JavaScript Module Expansion

- [x] Expand `require("smailnail")` exports without breaking `parseRule`, `buildRule`, or `newService`
- [x] Add richer IMAP session methods to the JS runtime surface
- [x] Add Sieve connection and session methods to the JS runtime surface
- [x] Add an offline Sieve script builder surface derived from the donor builder layer
- [x] Keep runtime naming coherent so mailbox properties and mailbox operation helpers are not ambiguous
- [x] Add JS-module tests that exercise the new IMAP and Sieve surfaces through goja

## Phase 5: Embedded Docs And Examples

- [x] Update package docs to describe both IMAP and Sieve capabilities
- [x] Document every newly exported runtime symbol so doc extraction matches the runtime surface
- [x] Add worked examples for richer IMAP mailbox automation
- [x] Add worked examples for Sieve script construction and server-side script management
- [x] Re-run the documented-symbol parity tests and fix any mismatches

## Phase 6: Account And MCP Follow-Through

- [x] Decide whether Sieve account settings will ship in this change or remain a documented follow-up
- [ ] If shipping now, extend hosted account types, repository SQL, service validation, and tests for Sieve settings
- [x] Review whether the MCP `imapjs` runtime should expose the new Sieve and mailbox features immediately or behind follow-up docs/tooling

## Phase 7: Validation, Diary, And Commits

- [x] Run focused Go tests for `pkg/mailruntime`, `pkg/services/smailnailjs`, `pkg/js/modules/smailnail`, and `pkg/mcp/imapjs`
- [x] Address PR review findings and failing hosted-app build/security checks
- [x] Commit the shared runtime port in a focused commit
- [x] Commit the service/module/runtime-surface expansion in a focused commit
- [x] Commit the ticket bookkeeping and diary updates in a focused commit
- [x] Update the diary with what changed, what failed, the exact test commands, and review guidance
- [x] Update `changelog.md` with commit hashes and implementation milestones
- [x] Re-run `docmgr doctor --root ttmp --ticket SMN-20260319-IMAP-SIEVE --stale-after 30`
