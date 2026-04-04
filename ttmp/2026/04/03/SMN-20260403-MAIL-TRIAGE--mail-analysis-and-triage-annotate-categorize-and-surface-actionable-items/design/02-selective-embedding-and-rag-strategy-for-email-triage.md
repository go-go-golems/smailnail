---
title: Selective Embedding and RAG Strategy for Email Triage
doc_type: design
status: active
intent: long-term
topics:
  - email
  - embeddings
  - rag
  - cross-encoder
  - llm
  - sqlite
owners:
  - manuel
ticket: SMN-20260403-MAIL-TRIAGE
---

# Selective Embedding and RAG Strategy for Email Triage

## The Core Insight: Don't Embed Everything

We have 32,912 messages but 48% are noise. The raw body text is enormous (newsletters average 22K chars, personal+important total ~59M chars, newsletters total ~91M chars). Embedding all of this would be expensive and mostly useless — the signal-to-noise ratio in the embedding index would mirror the inbox itself.

Instead, we build a **tiered embedding architecture** where:

1. **LLM transforms** compress high-value content into embeddable summaries first
2. **Bi-encoder embeddings** index the transformed documents for fast recall
3. **Cross-encoders** rerank results against the actual query for precision
4. We also embed **our own queries, scripts, and reports** so the system can reuse its own tools

## What We Know About the Data

| Slice | Messages | Body text | Notes |
|---|---|---|---|
| Noise (skip entirely) | 15,798 | ~huge | Never embed |
| Newsletter bodies | 4,644 | ~91M chars (~23M tokens) | Too large to embed raw |
| Personal + Important | 815 | ~59M chars (~15M tokens) | High value, many are quote-heavy email threads |
| Work | 636 | unknown | Mostly GitHub PR notifications |
| Community | 646 | unknown | Mailing list + Recurse |
| Other signal | 2,058 | unknown | Services, hobby, financial |
| Subject lines (non-noise) | 8,718 | ~448K chars (~112K tokens) | Cheap to embed directly |
| Sender profiles | 387 annotated | ~200 words each when generated | Very cheap |
| Threads (multi-msg, non-noise) | 145 | Varies | Best candidates for summarization |
| Scripts & queries | 49 files | ~35K chars | Embed for tool reuse |

## The Five Embedding Layers

### Layer 0: Scripts, Queries, and Reports (meta-tools)

**What:** Embed the SQL queries, shell scripts, and generated reports from this triage session.

**Why:** When a future query comes in like "show me my unsubscribe candidates" or "what did my CPA say about taxes?", the system can retrieve the *query archetype* that already answers this, rather than constructing it from scratch. This is essentially a **tool retrieval** layer.

**What to embed:**
- Each `.sql` and `.sh` script with a synthetic description (the comment header + what it does)
- The summary report sections
- The annotation tag vocabulary with descriptions

**Transform before embedding:** Attach a natural-language description to each script:
```
Script: 30-investigate-tax-emails.sql
Description: Find all tax-related emails by searching subjects for keywords
like 'tax', 'CPA', '1099', 'W-2', 'IRS', 'Steuer' and by looking up
specific CPA senders at davemillercpa.com. Returns date, sender, subject.
Use when: user asks about taxes, CPA correspondence, tax filing status.
```

**Volume:** ~49 documents. Negligible cost.

**Cross-encoder use:** When a user asks a question, retrieve top-5 script candidates by bi-encoder, then cross-encode `(query, script_description)` pairs to pick the best tool to run.

### Layer 1: Sender Profiles (who is this?)

**What:** Generate a synthetic "sender profile" document for each annotated sender (387 senders), embedding the profile.

**Why:** Enables queries like "who writes to me about photography?", "which newsletters cover AI?", "who is my CPA?". The profile is a much richer embedding target than just the email address.

**LLM transform:** For each sender, generate a ~100-200 word profile from available metadata:

```markdown
## dave@davemillercpa.com — David Miller CPA LLC
- **Category:** important/tax (CPA, primary tax preparer)
- **Domain:** davemillercpa.com
- **Volume:** 11 messages, Jul 2024 – Oct 2025
- **Relationship:** Professional — Manuel's CPA for personal and equity tax matters
- **Key topics:** 2024 tax returns, equity/taxes advice, stock option exercise questions,
  quarterly investment commentary
- **Related senders:** liam@davemillercpa.com (tax preparer, same firm),
  ginny.white@gmail.com (involved in equity/taxes thread)
- **Action patterns:** Responds to document uploads, sends invoices via QuickBooks,
  follows up on e-filing deadlines
```

**Input for the LLM:** Sender metadata from `senders` table + annotation tag/note + sample of 5-10 recent subject lines from their messages. This is very cheap — maybe 500 tokens input per sender.

**Volume:** 387 profiles × ~150 words = ~58K words. Trivial to embed.

**Cross-encoder use:** After bi-encoder recall of candidate senders, cross-encode `(query, full_profile)` to rank. Useful when the query is ambiguous ("who helps me with money stuff?" should rank CPA > Affirm > PayPal).

### Layer 2: Thread Summaries and Newsletter Digests (what was discussed?)

This is the most impactful layer. Instead of embedding raw email bodies (which are full of quoted replies, signatures, HTML artifacts, and boilerplate), we **summarize first**.

#### 2a: Thread Summaries

**What:** For the 145 multi-message non-noise threads, generate a structured summary.

**LLM transform:** Feed the full thread (chronologically) to an LLM and produce:

```markdown
## Thread: equity/taxes advice
- **Participants:** ginny.white@gmail.com, dave@davemillercpa.com, liam@davemillercpa.com
- **Duration:** Jul 23, 2024 – Sep 18, 2024 (15 messages)
- **Topic:** Advice on exercising Formlabs stock options — tax implications,
  timing of exercise, cost basis questions, interaction with CPA
- **Key decisions/outcomes:** Dave provided analysis of tax scenarios for
  exercising options. Discussed ISO vs NSO treatment. Liam followed up with
  specific document requests.
- **Action items (if any):** Upload tax documents to CPA portal
- **Emotional tone:** Professional, collaborative, some urgency around exercise windows
- **Tags:** important/tax, important/equity, personal
```

**Input cost:** The equity/taxes thread is 82K chars (~20K tokens). A summary would compress it to ~200 words. This is a 100:1 compression ratio.

**Volume:** 145 threads. Largest is 47 messages. Most are 3-5 messages. Total LLM input cost: maybe 500K-1M tokens across all threads. Output: ~30K words of summaries.

**What NOT to summarize:** Single-message threads (28,405 of them). For those, we use Layer 3 (subject lines).

#### 2b: Newsletter Digests

**What:** For each newsletter sender, generate a **rolling topic digest** instead of embedding every issue.

**Why:** Gary Marcus sent 179 issues. Embedding all 179 separately would create a noisy index. Instead, produce one document per sender summarizing the topics they cover, with links to standout issues.

**LLM transform:** Feed the last 10-20 subject lines + first 500 chars of body to produce:

```markdown
## Newsletter Digest: simonw@substack.com — Simon Willison's Weblog
- **Frequency:** Weekly, 76 issues over 22 months
- **Core topics:** Python, AI/LLM tooling, SQLite, datasette, open source,
  prompt engineering, web development
- **Standout issues:**
  - "Two new Showboat tools: Chartroom and datasette-showboat" (Mar 2026)
  - "Everything I built with Claude Artifacts" (2025)
- **Relevance to Manuel:** High — overlaps with Go/SQLite work, LLM tooling,
  open source CLI design
```

**Volume:** 99 newsletter senders × ~200 words = ~20K words. Input cost: ~50K tokens (just subject lines + snippets).

**Alternative approach:** Instead of one digest per sender, create **monthly topic clusters** across all newsletters. "In October 2025, your newsletters covered: AI agents (pragmaticengineer, swyx, thursdai), labor/tech (bloodinthemachine), formal methods (hillelwayne)..." This would be ~24 documents for 24 months.

### Layer 3: Subject Lines (fast topic lookup)

**What:** Embed subject lines directly, but only for non-noise messages.

**Why:** Subject lines are cheap (~112K tokens for all 8,718 non-noise messages), already in natural language, and answer the most common query: "did I get an email about X?"

**Transform before embedding:** Minimal — just prepend the sender category tag:

```
[personal] Re: Examples of Method Combinations, Multiple Dispatch, MOP
[important/tax] RE: Questions About Tax Year 2024
[newsletter/tech] Two new Showboat tools: Chartroom and datasette-showboat
```

This tag prefix helps the embedding model distinguish "tax" in a newsletter headline from "tax" in actual CPA correspondence.

**Volume:** 8,718 embeddings. At ~$0.10/M tokens with a cheap embedding model, this costs about $0.01.

**Cross-encoder use:** After bi-encoder retrieves top-50 subject lines, cross-encode `(query, "[tag] subject")` pairs to rerank. The cross-encoder will understand that "what did my lawyer say?" should rank `[important/legal] separation agreement` above `[newsletter/culture] Will you donate $5 to investigate Trump's illegal...`.

### Layer 4: Full Message Bodies (rare, on-demand)

**What:** Embed individual message bodies only for high-value messages that pass a filter.

**When:** Only for:
- Messages from `important/*` senders (471 msgs)
- Personal correspondence (363 msgs)
- Messages flagged `action-required` by a future LLM pass

**Transform before embedding:** Strip quoted replies, signatures, HTML boilerplate, and legalese. For email threads, extract only the **new content** in each message (the part above the `>` quoted text). This alone can reduce token count by 60-80%.

**Volume:** ~834 messages after stripping. Maybe 2-3M tokens total. Still manageable.

**Cross-encoder use:** This is where cross-encoders matter most. A query like "what documents does my CPA need?" should retrieve the specific message where Liam says "I do not see it. Which folder did you put it in?" — the bi-encoder will recall several CPA messages, but the cross-encoder scores the most relevant one.

## The Retrieval Pipeline

```
User query: "what's the status of my tax filing?"
                │
                ▼
   ┌─────────────────────────┐
   │  Layer 0: Tool Retrieval │ ── bi-encoder → "30-investigate-tax-emails.sql"
   │  (scripts & queries)     │    cross-encoder confirms relevance
   └─────────┬───────────────┘    → can RUN the query directly
             │
             ▼
   ┌─────────────────────────┐
   │  Layer 1: Sender Lookup  │ ── bi-encoder → dave@davemillercpa.com profile
   │  (sender profiles)       │    → identifies WHO to look at
   └─────────┬───────────────┘
             │
             ▼
   ┌─────────────────────────┐
   │  Layer 2: Thread Summary │ ── bi-encoder → "equity/taxes advice" summary
   │  + Newsletter Digest     │    → WHAT was discussed
   └─────────┬───────────────┘
             │
             ▼
   ┌─────────────────────────┐
   │  Layer 3: Subject Lines  │ ── bi-encoder → top-50 tax-related subjects
   │  (cheap, broad recall)   │    cross-encoder reranks → top-5
   └─────────┬───────────────┘
             │
             ▼
   ┌─────────────────────────┐
   │  Layer 4: Full Bodies    │ ── only fetched for the top-5 messages
   │  (on-demand, expensive)  │    cross-encoder scores against query
   └─────────────────────────┘
             │
             ▼
   ┌─────────────────────────┐
   │  LLM Answer Generation   │ ── synthesizes from all layers
   └─────────────────────────┘
```

The key insight is that **each layer filters progressively**: Layer 0 tells you which tool to use, Layer 1 tells you who to look at, Layer 2 tells you what was discussed, Layer 3 finds specific messages, and Layer 4 provides the actual content. You never need to embed or search everything.

## Cross-Encoder Strategy

Cross-encoders score `(query, document)` pairs directly — much more accurate than bi-encoders but too expensive to run over thousands of documents. We use them at three points:

### 1. Tool Selection (Layer 0)

After bi-encoder retrieves 5-10 candidate scripts, cross-encode to pick the best one. This decides whether to run a SQL query, show a report section, or fall through to embedding search.

### 2. Subject Line Reranking (Layer 3)

Bi-encoder retrieves top-50 subject lines. Cross-encoder reranks to top-5. This is the highest-leverage use because subject lines are short (fast cross-encoding) and the bi-encoder's recall is broad but imprecise.

**Example:** Query "Formlabs stock" → bi-encoder returns subjects about EquityZen, Hiive, EquityBee, plus false positives like "Domestika online courses". Cross-encoder correctly ranks EquityZen confirmation requests above the false positives.

### 3. Duplicate/Cluster Validation

Cross-encode pairs of newsletter senders to find redundant subscriptions: is `swyx@substack.com` covering the same topics as `frontierai@substack.com`? Score `(sender_A_digest, sender_B_digest)` to find overlap.

## LLM Transformation Strategies

### Transform 1: Thread → Structured Summary
**Input:** Full thread messages in chronological order
**Output:** Participants, duration, topic, key decisions, action items, tone
**Model:** Any good instruction model (Claude Haiku, GPT-4o-mini)
**Cost:** ~500K tokens input across 145 threads

### Transform 2: Newsletter → Topic Digest
**Input:** Last 20 subject lines + first 500 chars of most recent issues
**Output:** Core topics, frequency, standout issues, relevance assessment
**Model:** Same as above
**Cost:** ~50K tokens input across 99 senders

### Transform 3: Sender → Profile
**Input:** Sender metadata + annotation tag/note + 10 recent subjects
**Output:** Structured sender profile with relationship context
**Model:** Same
**Cost:** ~200K tokens across 387 senders

### Transform 4: Subject → Expanded Query Target
**Input:** Raw subject line + sender category
**Output:** Enriched subject with context: `"[important/tax] RE: Questions About Tax Year 2024 — CPA follow-up on tax document uploads for 2024 personal and equity tax filing"`
**When:** Only for important/* and personal senders (~500 subjects)
**Cost:** ~50K tokens

### Transform 5: Query → Embeddable Description
**Input:** SQL query + comment header
**Output:** Natural language description of what the query does and when to use it
**Cost:** ~10K tokens across 49 scripts

### Transform 6: Monthly Digest → Embedding Target
**Input:** All signal messages from one month (subjects + sender tags)
**Output:** "In February 2026, Manuel received 37 important/personal emails including CPA follow-ups, a method dispatch discussion with achambers and fahree, SOITS Alumni thread updates, and a Zettelkasten course offer..."
**Cost:** ~100K tokens across 24 months

**Total LLM transform cost:** ~910K tokens input, ~200K tokens output. At Claude Haiku pricing (~$0.25/M input, $1.25/M output), this is roughly **$0.48 total**. Extremely cheap.

## Embedding Budget Summary

| Layer | Documents | Est. tokens to embed | Embedding cost ($0.10/M) |
|---|---|---|---|
| 0: Scripts/queries | 49 | ~10K | $0.001 |
| 1: Sender profiles | 387 | ~60K | $0.006 |
| 2a: Thread summaries | 145 | ~30K | $0.003 |
| 2b: Newsletter digests | 99 (or 24 monthly) | ~20K | $0.002 |
| 3: Subject lines | 8,718 | ~112K | $0.011 |
| 4: Full bodies (stripped) | 834 | ~2-3M | $0.25 |
| **Total** | **~10,232** | **~2.2-3.2M** | **~$0.27** |

Compare with naive "embed everything": 32,912 messages × avg ~18K tokens = ~592M tokens = **$59.20**. Our selective approach is **200x cheaper**.

## Embedding the Queries and Scripts Themselves

This is the most novel aspect. We treat our own investigation artifacts as first-class retrievable documents:

### What to embed:
1. **SQL queries** with their natural-language descriptions — so future sessions can find "the query that shows tax correspondence" without writing SQL from scratch
2. **Shell scripts** with their annotation methodology — so the categorization can be rerun or extended
3. **Report sections** from the inbox summary — so "give me the unsubscribe candidates" can retrieve the pre-computed list
4. **Query archetypes** — the 5 patterns we identified (48-investigate-query-reuse-patterns.sql):
   - "Recent messages from [category]"
   - "Top senders in [category]"
   - "Messages about [topic]"
   - "Threads with real conversations"
   - "Monthly breakdown by category"

### How this enables query reuse:

```
User: "who should I unsubscribe from?"
  → Layer 0 retrieves: "06-create-review-groups.sh" + "Unsubscribe Candidates" group
  → System runs: SELECT target_id FROM target_group_members WHERE group_id = '<unsub_gid>'
  → Returns pre-computed list with no embedding search needed
```

```
User: "what was that thread about stock options?"
  → Layer 0 retrieves: "32-investigate-equity-emails.sql"
  → Layer 2 retrieves: thread summary for "equity/taxes advice"
  → System can both run the SQL AND return the summary
```

## Performance Note: The sender_email Index Problem

During this session we discovered that `messages.sender_email` has no index, making any join between `messages` and `annotations` require a full table scan of the 5.3GB database (~5 minutes). This affects the embedding pipeline too — generating sender profiles or filtering messages by category requires this join.

**Recommendation:** Add `CREATE INDEX idx_messages_sender_email ON messages(sender_email);` before running the embedding pipeline. This would make category-filtered queries instant instead of 5-minute waits.

See `scripts/39-investigate-slow-query-analysis.sql` for the full analysis.

**Workaround (already proven):** Use CTEs to pre-aggregate by `sender_email` first, then join to annotations. The query in `45b-investigate-monthly-digest-fast.sql` runs in 0.7s instead of timing out, using exactly this pattern.

## Implementation Order

1. **Layer 3 first** (subject lines) — cheapest, broadest coverage, immediate value
2. **Layer 0 next** (scripts/queries) — enables tool reuse in future sessions
3. **Layer 1** (sender profiles) — LLM transform + embed, enables "who" queries
4. **Layer 2a** (thread summaries) — highest value per token, enables "what happened" queries
5. **Layer 2b** (newsletter digests) — enables "what do I read about X?" queries
6. **Layer 4 last** (full bodies) — only after the other layers prove useful

## Open Design Questions

1. **Embedding model choice:** `text-embedding-3-small` (OpenAI, cheap, good) vs `nomic-embed-text` (local, free) vs `voyage-3` (best for retrieval)?

2. **Cross-encoder model:** `cross-encoder/ms-marco-MiniLM-L-6-v2` (fast, local) vs API-based (better quality, higher latency)?

3. **Storage:** Embed into the same SQLite DB (via `sqlite-vec` or similar) or a separate vector store (Chroma, Qdrant)?

4. **Incremental updates:** When new mail arrives, which layers need updating? Subject lines (Layer 3) are append-only. Sender profiles (Layer 1) need regeneration. Thread summaries (Layer 2a) need regeneration if the thread grows.

5. **Monthly digest granularity:** One digest per newsletter sender, or cross-sender monthly topic clusters? The latter is better for "what was the AI discourse in October?" but harder to maintain.
