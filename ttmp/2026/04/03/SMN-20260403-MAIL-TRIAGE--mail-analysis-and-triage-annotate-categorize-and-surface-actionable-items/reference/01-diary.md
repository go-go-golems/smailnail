---
title: Diary
doc_type: reference
status: active
intent: long-term
topics:
  - email
  - annotations
  - triage
  - sqlite
  - workflow
owners:
  - manuel
ticket: SMN-20260403-MAIL-TRIAGE
---

# Diary

## Goal

Record the step-by-step journey of analyzing Manuel's 2-year email archive (32,912 messages), designing a triage/annotation workflow, and executing it using `smailnail annotate` to surface actionable and interesting mail while suppressing noise.

## Step 1: Explore the database and understand the mailbox shape

Explored the smailnail SQLite mirror database at `~/smailnail/smailnail-last-24-months-merged.sqlite` to understand what we're working with. The database is well-enriched — threads are reconstructed (29,305 threads), senders are normalized (2,187 unique senders), and body text/HTML is available for nearly all messages. Zero annotations exist yet, so this is a clean start.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to do mail analysis on my mailbox and design an overall workflow for reviewing mail and surfacing good things. You have a sqlite database of the last 2 years of my email in ~/smailnail, and you can use `smailnail` the binary to annotate the data..."

**Assistant interpretation:** Explore the email database, understand its structure and content, design a triage/annotation workflow using `smailnail annotate`, and present a plan before executing.

**Inferred user intent:** Manuel wants a "secretary" that organizes his inbox — categorizing senders, suppressing noise, surfacing important/interesting mail, and leaving durable annotations in the DB that he can inspect and review.

### What I did

1. Explored the database schema: `messages`, `senders`, `threads`, `annotations`, `annotation_logs`, `target_groups`, `target_group_members`
2. Ran volume queries: 32,912 messages, May 2024 – Mar 2026, all in INBOX
3. Analyzed top senders, domains, thread sizes
4. Categorized mail by domain type
5. Explored the `smailnail annotate` CLI and its playbook documentation

### Key findings

**Volume breakdown (approximate):**
| Category | Count | % |
|---|---|---|
| GitHub notifications | 9,153 | 28% |
| Newsletters (substack, beehiiv, every.to, readwise) | 3,661 | 11% |
| Commerce/retail (amazon, thetreecenter, walgreens, cvs) | 2,094 | 6% |
| Social platforms (twitch, soundcloud, bandcamp, meetup, facebook) | 1,644 | 5% |
| Personal email (gmail, privaterelay, icloud, etc.) | 1,352 | 4% |
| Mailing lists (entropia) | 420 | 1% |
| Financial (paypal, bank of america) | 292 | 1% |
| Other (remaining ~14K) | ~14,296 | 43% |

**GitHub sub-breakdown (9,153 msgs):**
- Thread replies: 4,145
- CI failures: 4,003
- Other: 889
- Dependabot: 101
- Issues: 15

**Thread structure:**
- 28,405 single-message threads (mostly newsletters, notifications)
- 688 threads with 2-5 msgs
- 199 threads with 6-20 msgs
- 13 threads with 20+ msgs (largest: 47 msgs, work PRs)

**Notable senders:**
- `notifications@github.com` dominates with 8,886 msgs
- 1,523 messages have empty sender_email (parsing issue — from_summary shows they're mostly chatgpt-codex-connector bot GitHub notifications)
- 93 Substack newsletter senders, 939 senders with List-Unsubscribe headers
- Active personal correspondents: ginny.white, perryholser, robokid13, erinmperfect, gleitman.sam, fabianfaedrich1

**Enrichment status:** Complete — threads, senders, unsubscribe links all populated.

### What worked
- All queries ran fast against the SQLite DB
- The enrichment is solid — sender normalization and thread reconstruction are done
- The annotation CLI has clean semantics for target types (message, thread, sender, domain, mailbox, account)

### What I learned
- The mailbox is heavily skewed toward automated/notification mail (~60%+ is GitHub + newsletters + commerce)
- Personal correspondence is a small but high-value fraction
- The empty sender_email rows (1,523) are mostly GitHub bot notifications that didn't get normalized
- Work threads (team-mento PRs) are the largest multi-message threads

### What was tricky to build
N/A — this was pure exploration.

### What warrants a second pair of eyes
- The 1,523 empty-sender messages — should these be enriched/fixed before annotating?
- Whether the annotation taxonomy I propose (Step 2) maps well to Manuel's actual priorities

### What should be done in the future
- Decide on the annotation taxonomy and execute the triage plan

## Step 2: Execute 4-phase triage plan

Executed the full triage workflow: sender categorization (scripts 01–05), review group creation (script 06), message surfacing (script 07), and summary report generation. All annotations stored in the SQLite DB with `source_label=mail-triage-v1` and `created_by=pi-agent` for traceability.

### Prompt Context

**User prompt (verbatim):** "I think it's good, create the doc"

**Assistant interpretation:** Proceed with the proposed 4-phase plan — categorize senders, create groups, surface actionable messages, generate summary report.

**Inferred user intent:** Manuel approved the plan and wants the annotations executed and the summary report created.

### What I did

1. **Script 01** — Tagged noise senders: CI (2 senders, 9,120 msgs), marketing (12), spam (6 domains → 95 individual senders), social notifications (7), transactional (9)
2. **Script 02** — Tagged newsletter senders: tech (23), culture (8), creative (9)
3. **Script 03** — Tagged personal correspondents (27), work (7), community (5), financial (2), services (10), hobby (6), plus more spam domains (25 senders)
4. **Script 04** — Sweep of remaining high-volume senders: Apple Private Relay services (4), more hobby/creative (3), community (3), services (5), financial (2), tech newsletters (5), social noise (4), more newsletters (5)
5. **Script 05** — Final sweep: more Facebook noise (3), services/commerce (6), newsletters (5), hobby (4), remaining Substack (38), remaining Manning (3), remaining Gmail personal (25), remaining Facebook (8), remaining Apple (12), misc (8)
6. **Script 06** — Created 5 review groups: Unsubscribe Candidates (86 members), Valuable Newsletters (97), Personal Contacts (52), Work Senders (8), Hobby & Creative (25)
7. **Script 07** — Surfaced recent personal emails (21 items), multi-message personal threads (33), work threads (12), community messages (30), financial alerts (20)
8. **Summary report** — Generated comprehensive inbox summary report with action items, newsletter inventory, unsubscribe candidates, and SQL query reference

### Why
- Establish a durable sender-level taxonomy in the DB that makes future triage incremental
- Surface the most actionable items (personal replies, financial alerts) so Manuel can focus on what matters
- Identify unsubscribe candidates to reduce future noise

### What worked
- Shell scripts with idempotent `annotate_sender()` functions work well — can be re-run safely
- The `smailnail annotate` CLI is clean and fast
- Annotation logs in the DB provide a good audit trail
- 73.3% coverage with 416 annotations is a solid first pass

### What didn't work
- Some Substack newsletters got tagged `newsletter/tech` when they're really culture/philosophy (e.g., slavoj+philosophy, juliaserano). The heuristic was too coarse.
- Newsletter issues from the last 30 days showed empty because the most recent data is March 2026 (data stops at 2026-03-03)

### What I learned
- Nearly half the inbox (48%) is pure noise that could be filtered or unsubscribed
- Personal email is only 1% of volume but highest value
- Apple Private Relay senders need special handling — the obfuscated email addresses are hard to categorize without domain extraction from the relay prefix
- The 1,523 empty-sender messages are all from GitHub bot notifications

### What was tricky to build
- Maintaining idempotency across multiple script runs — each `annotate_sender()` checks for existing annotations before adding
- Categorizing the long tail of 1,305 domains — focused on senders with 10+ messages and accepted that the remaining 27% (8,767 messages) are one-off or low-volume senders

### What warrants a second pair of eyes
- Many Substack senders were bulk-tagged as `newsletter/tech` — Manuel should review whether some should be `newsletter/culture` or `newsletter/creative`
- The personal correspondents list was built from gmail/icloud/yahoo/hotmail senders with 2+ messages — some may be strangers, some may be missing
- Experian alerts were tagged `financial` but they're mostly marketing; consider moving to `noise/marketing`

### What should be done in the future
- Refine Substack newsletter subcategories based on Manuel's actual reading habits
- Handle the 1,523 empty-sender messages (GitHub bot notifications)
- Build Sieve rules based on the noise annotations to auto-file future messages
- Set up periodic re-runs to catch new senders
- Consider LLM-based content analysis for message-level annotations (interesting vs. skip)

## Step 3: Add high-importance sender categories

Manuel pointed out that the initial triage missed critical categories: tax/CPA correspondence, lawyer emails, stock option/equity inquiries, housing/rent, health, and conference/speaking invitations. These are low-volume but extremely high-priority.

### Prompt Context

**User prompt (verbatim):** "what about important emails like about taxes, or lawyer or work stuff or offers for talks and other things?"

**Assistant interpretation:** The initial triage focused on volume-based noise/signal but missed importance-based categories. Search for and tag high-importance senders regardless of volume.

**Inferred user intent:** A good secretary doesn't just sort by volume — they prioritize by consequence. Tax deadlines, lawyer correspondence, stock option windows, and speaking invitations require timely attention.

### What I did

1. Searched for tax-related messages — found 55 messages from davemillercpa.com (CPA), plus QuickBooks/Intuit invoices
2. Searched for lawyer correspondence — found kryzaklaw (separation agreement, 19 msgs) and wrightfamilylawgroup (retainer, workshops, 11 msgs)
3. Searched for equity/stock options — found 32 messages across Hiive, EquityZen, EquityBee about Formlabs holdings
4. Found work admin: admoove.com (bl0rg compensation), carta.com (equity management)
5. Found housing: Bilt Rewards (rent autopay), Bozzuto (property management)
6. Found health: Revive Therapeutic, Prosper Health, Phoenix Rising Centers
7. Found conferences: AI Engineer summits, MLOps Community Coding Agents Conference, PyData Boston CFP, Maven courses
8. Created script 08 and annotated 30 senders with `important/*` tags
9. Updated summary report with high-importance section

### Why
- Volume-based triage misses high-consequence, low-volume senders
- These categories have real deadlines and financial/legal implications
- A secretary's #1 job is ensuring nothing critical falls through the cracks

### What worked
- Subject-line keyword search found the key senders quickly
- The `important/*` tag namespace separates these from the original taxonomy without conflicting
- Some senders (kryzaklaw, davemillercpa) now have both a `personal` tag AND an `important/*` tag, which is correct

### What I learned
- The CPA thread "Questions About Tax Year 2024" ran for 10 months (Mar 2025 – Dec 2025) with ~30 exchanges — this is clearly high-priority
- There's a past-due CPA invoice from Dec 2025 that may still need attention
- Formlabs equity: EquityZen keeps asking to confirm holdings, EquityBee asked about stock options — these may have time-sensitive windows
- Conference invitations from ai.engineer and mlops.community are legitimate speaking/attendance opportunities

### What was tricky to build
- Distinguishing genuine legal correspondence from noise that happens to mention "legal" (PayPal terms changes, Intercept articles about legality)
- The PyData CFP messages had empty sender_email so they can't be tagged by sender — they'd need message-level annotations

### What warrants a second pair of eyes
- The past-due CPA invoice (Dec 2025) — is this resolved?
- Formlabs equity confirmation requests from EquityZen — do these need action?
- Is the kryzaklaw separation agreement finalized? Last message was Oct 2025.

### What should be done in the future
- Add message-level `action-required` annotations for the most time-sensitive items
- Build an "importance dashboard" that surfaces these categories first
- Consider a recurring triage pass that checks for new messages from important/* senders

## Step 4: Design selective embedding and RAG strategy

Designed a 5-layer embedding architecture that avoids indexing everything. The key insight is to use LLM transforms to compress content before embedding, and cross-encoders to rerank results. The most novel aspect is embedding our own scripts and queries as retrievable tools (Layer 0).

### Prompt Context

**User prompt (verbatim):** "if you were to enrich this data with a vector embedding approach, but approach it such that we don't have to index everything, what are the things you would target? also think about using cross encoders, and also doing transformations with an LLM before doing embeddings, for example to match queries better."

Followed by: "you could also compute embeddings over bigger groups of emails or just their subjects or like the results of certain queries too, for example. or even embeddings on queries and scripts so we can reuse them."

**Assistant interpretation:** Design a cost-effective embedding/RAG strategy that selectively indexes the most valuable content, uses LLM transforms to compress before embedding, leverages cross-encoders for precision, and treats scripts/queries/reports as embeddable artifacts for tool reuse.

**Inferred user intent:** Manuel wants a practical retrieval system that can answer natural-language questions about his email without the cost and noise of embedding everything. He also sees the value of embedding the investigative artifacts (queries, scripts) so future sessions can reuse them.

### What I did

1. Analyzed body text sizes across categories (scripts 35-36, 42)
2. Estimated token budgets: newsletters ~23M tokens raw, personal+important ~15M tokens raw, subject lines only ~112K tokens
3. Identified 145 multi-message non-noise threads as prime summarization candidates (script 47)
4. Catalogued 5 recurring query archetypes from the session (script 48)
5. Discovered the sender_email index problem makes joins 5-minute waits, but CTE pre-aggregation solves it (script 45b: 0.7s vs timeout)
6. Wrote comprehensive design doc with 5 layers, 6 LLM transforms, 3 cross-encoder use cases, cost estimates, and implementation order
7. Saved all investigation queries as numbered `.sql` files (09-48)

### Key numbers

| Approach | Documents | Tokens | Cost |
|---|---|---|---|
| Naive (embed everything) | 32,912 | ~592M | ~$59.20 |
| Selective (our design) | ~10,232 | ~3.2M | ~$0.27 |
| LLM transform cost | - | ~1.1M | ~$0.48 |
| **Total selective** | **~10,232** | **~4.3M** | **~$0.75** |

200x cheaper than naive embedding.

### Why
- Naive embedding is wasteful: 48% of messages are noise, newsletter bodies are 22K chars avg but only carry ~200 words of unique topic info
- LLM summarization before embedding gives better retrieval quality AND lower cost
- Embedding scripts/queries enables a "tool retrieval" pattern where the system finds and runs a pre-built query instead of searching embeddings
- Cross-encoders at specific points (tool selection, subject reranking, cluster validation) add precision where bi-encoders are weak

### What worked
- The CTE pre-aggregation pattern for avoiding the sender_email index problem — proven in 45b with 0.7s vs timeout
- Token budget estimation gave clear evidence that selective embedding is 200x cheaper
- The 5-layer architecture maps cleanly to the annotation taxonomy we already built

### What I learned
- The DB is 5.3GB, mostly due to body_text/html stored inline — even a "scan" touches all that data
- Newsletter bodies are wildly variable: 292 chars to 6.7M chars (probably an HTML monster)
- Sender profiles are extremely cheap to generate (~200 words each) and would be the highest-value embedding per token
- Thread summaries are the highest compression ratio: the equity/taxes thread is 82K chars that compress to ~200 words (400:1)

### What was tricky to build
- The monthly digest query kept timing out because it hit the sender_email scan problem. Had to restructure as a CTE that pre-aggregates by sender_email first.
- Distinguishing newsletter "issues" from newsletter "updates" — Simon Willison sends one per day (76 issues), Every.to sends multiple per day (426 msgs over 362 days)

### What warrants a second pair of eyes
- Layer 4 (full body embedding) might not be needed at all if thread summaries + subject lines provide enough recall
- The query archetype embedding (Layer 0) is novel — should validate that it actually helps retrieval vs just using a keyword lookup
- Newsletter digest granularity: per-sender vs per-month cross-sender clusters

### What should be done in the future
- Implement Layer 3 first (subject line embedding) as proof of concept
- Add the sender_email index to make all category queries instant
- Build a prototype retrieval pipeline and test against real questions
- Evaluate embedding model choices (local vs API)
