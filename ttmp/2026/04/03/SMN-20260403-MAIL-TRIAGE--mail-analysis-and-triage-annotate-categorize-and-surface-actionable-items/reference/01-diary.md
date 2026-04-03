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
