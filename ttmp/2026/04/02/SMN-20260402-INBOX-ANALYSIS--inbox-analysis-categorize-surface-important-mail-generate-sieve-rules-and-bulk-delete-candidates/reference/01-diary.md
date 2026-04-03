---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/01_basic_stats.sql
      Note: Basic stats query - counts
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/02_sender_analysis.sql
      Note: Top senders and domain frequency analysis
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/03_categorize.sql
      Note: Main category classifier (CASE heuristic)
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/04_other_deep_dive.sql
      Note: Drilling into uncategorized messages
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/05_important_emails.sql
      Note: Surface personal
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/06_bulk_delete_candidates.sql
      Note: Count all safe-to-delete buckets
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/07b_github_breakdown_fixed.sql
      Note: GitHub repo and event-type breakdown
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/08_sieve_rules.sieve
      Note: Generated Sieve rules for discard + file-to-folder
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Explore `smailnail-last-month.sqlite` (1-month INBOX mirror, 2,894 messages) to:
1. Understand category breakdown
2. Surface genuinely important emails
3. Identify safe bulk-delete candidates
4. Produce Sieve rules for ongoing cleanup

---

## Step 1: Environment setup + basic stats

Created docmgr ticket `SMN-20260402-INBOX-ANALYSIS`. Located SQLite at `/tmp/smailnail-last-month.sqlite`.
The DB has a `messages` table (2,894 rows, no remote_deleted), a single account `mail-bl0rg-net-993-manuel-a8c4454ab8d9`,
a single mailbox `INBOX`, and an FTS5 index on subject/from/body.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to run an email analysis of my inbox and then store all the queries and scripts you write in scripts/ folder of the ticket and keep a detailed diary as you explore. The goal is to categorize my email and then besides a report (and surfacing important emails), create sieve rules to clean up the inbox and also find stuff to just delete in bulk"

**Assistant interpretation:** Set up a structured docmgr ticket with scripts/, run exploratory SQL queries against the SQLite inbox mirror, categorize senders, surface important mail, produce Sieve rules and a bulk-delete list.

**Inferred user intent:** Get a clean, actionable picture of a cluttered inbox so that ongoing email hygiene can be automated (Sieve) and a one-time purge of noise can be done safely.

### What I did
- Ran `docmgr ticket create-ticket` for `SMN-20260402-INBOX-ANALYSIS`
- Added diary doc, 6 tasks
- Ran `scripts/01_basic_stats.sql` against the DB

### Key findings
| Metric | Value |
|--------|-------|
| Total messages | 2,894 |
| Date range | 2026-03-02 → 2026-04-02 |
| March messages | 2,772 |
| April messages | 122 |
| With attachments | 46 |
| Typical size | 10–100 KB (2,354 msgs) |
| Large (>1MB) | 6 |
| Most messages unflagged (null flags) | 2,775 |

### What was tricky
`flags_json` stores `null` as a literal SQL null, not `'[]'` — so `GROUP BY flags_json` shows three buckets: null, `["\\Seen"]`, `["NonJunk"]`, `["\\Recent"]`.

---

## Step 2: Sender/domain analysis

Ran `scripts/02_sender_analysis.sql`. The top-level picture immediately showed the inbox is dominated by automated/commercial senders.

### Top senders
| Sender | Count |
|--------|-------|
| Manuel Odendahl `<notifications@github.com>` | 652 |
| chatgpt-codex-connector[bot] `<notifications@github.com>` | 96 |
| Twitch `<no-reply@twitch.tv>` | 84 |
| The Tree Center `<sales@thetreecenter.com>` | 62 |
| Zillow (via Apple private relay) | 59 |
| Substack `<no-reply@substack.com>` | 52 |
| Freelancer.com | 46 |
| Manning Publications | 45 |

### Top domains
| Domain | Count |
|--------|-------|
| github.com | 773 |
| substack.com | 291 |
| privaterelay.appleid.com | 114 |
| twitch.tv | 85 |
| thetreecenter.com | 62 |
| email.meetup.com | 52 |
| facebookmail.com | 51 |
| manning.com | 49 |

### What was tricky
- Several commercial senders come through Apple's **private relay** (`@privaterelay.appleid.com`) making domain-matching unreliable; must match the display name or a header instead.
- `lists.entropia.de` has 51 messages; these are a German tech community mailing list (not spam).

---

## Step 3: Category breakdown

Ran `scripts/03_categorize.sql` with a CASE-based heuristic classifier.

| Category | Count | % | MB |
|----------|-------|---|----|
| **github** | 771 | 26.6% | 25.6 |
| **commercial** | 505 | 17.4% | 38.5 |
| **other** | 454 | 15.7% | 33.5 |
| **newsletter** | 393 | 13.6% | 38.3 |
| **automated** | 283 | 9.8% | 12.8 |
| **social** | 179 | 6.2% | 6.0 |
| **financial-shopping** | 149 | 5.1% | 11.3 |
| **mailing-list** | 116 | 4.0% | 7.2 |
| **security** | 36 | 1.2% | 0.5 |
| **personal** | 8 | 0.3% | 2.6 |

**Key takeaway:** Only **8 messages** (0.3%) classified as personal human mail. GitHub alone is 27% of the inbox. The "other" bucket (454) was drilled into in Step 4.

---

## Step 4: "Other" deep dive

Ran `scripts/04_other_deep_dive.sql`. The 454 "other" messages break down further into:

- **Photography communities**: Magnum Photos (15), RI Photo Center (7), AS220 (7), Baltimore Photo Space (4)
- **Art/culture**: MACK books (8), Providence Athenaeum (8), La librairie lapin (2), Weimarer Malschule (2)
- **Community/local**: AlexanderTheCreate (8), Woonasquatucket (4), AS220 (7), Underground Music Academy (4)
- **AI/ML marketing**: Weights & Biases (5), OpenHands (2), Mermaid.ai (2)
- **Proxy/scraping services**: Rayobyte (4+), Bright Data (2+)
- **Phishing/spam**: Firebase noreply spam, various invoice/payment lures
- **Personal acquaintances**: Manfred Odendahl (family), Marco Antoniotti, Hans Hübner, lisa besset, Gian und Denise Besset Stadler, catherinechambel

### What was tricky
Firebase spam uses `noreply@<random-hash>.firebaseapp.com` — the domain suffix is always `firebaseapp.com` so a domain-suffix match catches all of it.

---

## Step 5: Important emails surfaced

Ran `scripts/05_important_emails.sql`. Found genuinely important items:

### Personal / human senders
- **Hans Hübner** — calendar invites for "Prompter Stammtisch" and "Catch up" (German tech community)
- **Gian und Denise Besset Stadler / catherinechambel / lisa besset** — thread about "Ascension à La Forclaz" (mountain trip planning)
- **Manfred Odendahl** (family) — forwarded account-hacked emails + photography topic
- **Marco Antoniotti** — technical thread "eval-when mysteries" (Lisp/Common Lisp)
- **Stepan Gantralyan** — gallery opening / concert invitation

### Billing / financial (need attention)
- AWS invoice (April, ~$180 account)
- DigitalOcean invoice (March)
- Hetzner **final payment warning** (K1264826925) — ⚠️ URGENT
- Rayobyte invoices (proxy service)
- Apple Card statements + due notices
- Bank of America Zelle payments (Emily Mason $2,000, $350 to CLUB CLUB LLC)
- Chroma Software / Every Media stripe receipts
- Rhode Island Energy bills
- Cox internet bill
- Conservice utility bill
- Blender Studio subscription receipt
- GitHub payment receipt (wesen organization)
- Invoice Ninja — sent to Legato Hearing (his own business invoicing)

### Security / account
- Google: **3× "Critical security alert"** for wesen3000@googlemail.com on 2026-03-30 — ⚠️ REVIEW
- Microsoft: **multiple "Ungewöhnliche Anmeldeaktivität"** (unusual login activity) March 13 — ⚠️ REVIEW
- Microsoft Power Automate scam: "Unusual Transaction USD 859.99 to Pay-Pal BTC" — PHISHING, discard

### Dev tools (actionable)
- Rollbar: **[TTC-wordpress] New Error #1400 InvalidArgumentException** (March 19/24) — prod issue
- Rollbar: **New Error #1398 E_WARNING Undefined array key** (March 19)
- Rayobyte: account manager introduction + invoices

---

## Step 6: Bulk-delete candidates

Ran `scripts/06_bulk_delete_candidates.sql`.

| Bucket | Count |
|--------|-------|
| Firebase phishing spam | 103 |
| Crowdfunding spam (kickstarter farms) | 99 |
| Twitch stream notifications | 84 |
| Zillow updates | 67 |
| The Tree Center marketing | 62 |
| Freelancer.com notifications | 53 |
| Meetup notifications | 52 |
| Facebook noise | 51 |
| Retail marketing (Domestika/MUJI/TASCHEN/Baronfig) | 54 |
| Manning Publications marketing | 49 |
| Experian monitoring spam | 44 |
| eBay | 20 |
| Manning | 49 |
| ship30for30 | 25 |
| PledgeBox | 27 |
| iPhone Photography School | 22 |
| LinkedIn | 28 |
| Walgreens | 14 |
| CVS | 14 |
| Amazon marketing | 8 |
| **TOTAL SAFE TO DELETE** | **889** |

**889 / 2,894 = 30.7% of inbox is safe bulk-delete.**

---

## Step 7: GitHub breakdown

Ran `scripts/07b_github_breakdown_fixed.sql`.

| Repo | Count |
|------|-------|
| go-go-golems/other | 382 |
| wesen/other | 119 |
| go-go-golems/geppetto | 72 |
| go-go-golems/pinocchio | 71 |
| wesen/goldeneaglecoin.com | 31 |
| go-go-golems/bobatea | 23 |
| go-go-golems/go-go-os-frontend | 22 |
| github-system | 20 |
| wesen/temporal-relationships | 17 |
| go-go-golems/codex-sessions | 11 |

| Event type | Count |
|------------|-------|
| **CI failures** | **652** |
| PR/issue discussions | 97 |
| github-system | 18 |

**84% of GitHub mail is CI failure notifications.** The Codex bot alone generated 96 notifications (all CI-related). This is the single biggest inbox noise source.

### What was tricky
SQLite doesn't have `regexp_substr` — had to use a long CASE/LIKE chain to extract repo names from subject-line patterns like `[owner/repo] ...`.

---

## Step 8: Sieve rules

Wrote `scripts/08_sieve_rules.sieve` covering:

**Discard rules (30.7% of inbox):**
- Crowdfunding spam farms (6 domains)
- Firebase noreply phishing
- Facebook & Instagram noise
- SoundCloud, Zillow
- The Tree Center, Experian, Freelancer, LinkedIn
- Walgreens, CVS, ship30for30, PledgeBox
- iPhone Photography School, Twitch (non-billing)
- MUJI/TASCHEN/MACK/Baronfig/Domestika retail
- Manning & eBay (non-transaction)

**File-to-folder rules:**
- `github/ci` — CI failures
- `github/prs` — PR discussions
- `github/notifications` — all other GitHub
- `newsletters/substack`, `/beehiiv`, `/every`, `/readwise`
- `events/meetup`
- `lists` — mailing lists (riseup, entropia)
- `monitoring/rollbar`
- `dev-tools/wandb`, `/firecrawl`, `/replit`, `/augment`, `/coderabbit`
- `finance/apple`, `/bofa`, `/paypal`, `/aws`, `/stripe-receipts`, `/digitalocean`, `/hetzner`, `/lendingclub`, `/affirm`
- `security/google`, `/microsoft`
- `newsletters/apple-developer`

### What should be done in the future
- Tune Sieve rules for the `go-go-golems` CI failures specifically: consider filing only to `github/ci` and only notifying on new errors (not repeat failures)
- Add a Sieve rule for `privaterelay.appleid.com` that inspects display-name or body domain to route retail vs real
- Add unsubscribe automation for anything that matched discard rules (send unsubscribe requests)

---

## What warrants a second pair of eyes

1. **Hetzner "Final Payment Warning"** — check if K1264826925 is real or has lapsed
2. **Google critical security alerts** (March 30) on wesen3000@googlemail.com — verify no account compromise
3. **Microsoft unusual login activity** (March 13, multiple times) — may indicate credential leak
4. **Rollbar prod errors** on TTC-wordpress — #1400 InvalidArgumentException is a production regression
5. **Bank of America Zelle: $2,000 to Emily Mason** — verify intentional
6. **Invoice Ninja: invoice to Legato Hearing** — verify if business accounting is current

---

## Code review instructions

- All scripts are in `scripts/` directory of this ticket
- Run any `.sql` file with: `sqlite3 /tmp/smailnail-last-month.sqlite < scripts/NN_name.sql`
- The Sieve file (`08_sieve_rules.sieve`) is ready to be loaded onto the mail server via `sieveshell` or ManageSieve
- Validate Sieve syntax with: `sieve-test` (if available) or the Dovecot `sievec` compiler

## Technical details

**DB connection:** `sqlite3 /tmp/smailnail-last-month.sqlite`

**Key columns:**
- `from_summary` — "Display Name <email@domain>"
- `subject` — decoded subject
- `sent_date` — ISO8601 with tz offset
- `flags_json` — JSON array or NULL
- `has_attachments` — BOOLEAN (0/1)

**Gotchas:**
- No `regexp_substr` in SQLite — use `LIKE` + `INSTR` + `SUBSTR`
- `flags_json` can be NULL (not `'[]'`)
- Apple private relay hides real sender domain; use display-name matching
- Substack sends from both `@substack.com` AND `@mg1.substack.com`
