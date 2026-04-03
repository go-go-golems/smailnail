---
Title: Agent Annotation Handbook
Slug: smailnail-agent-annotation-handbook
Short: How agents should create annotations, logs, groups, and agent runs when triaging email in the smailnail mirror database.
Topics:
- sqlite
- email
- agent
- annotation
- workflow
Commands:
- annotate
- annotate annotation add
- annotate annotation list
- annotate group create
- annotate group add-target
- annotate log add
- annotate log link-target
Flags:
- sqlite-path
- target-type
- target-id
- tag
- note
- source-kind
- source-label
- agent-run-id
- created-by
- review-state
- log-kind
- title
- body
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This handbook tells agents how to produce clean, auditable annotation work in the smailnail mirror database. Follow it exactly. Humans review your work through the annotation UI, and sloppy metadata makes that review painful.

The goal is that every piece of agent work is traceable: who did it, when, why, what it touched, and how to undo or extend it.

## Before You Start: Generate An Agent Run ID

Every annotation session needs a unique `agent_run_id`. Generate it before your first write and pass it to every command in the session.

Format: `{source-label}-{YYYYMMDD}` or `{source-label}-{YYYYMMDD}-{seq}` if you run multiple passes per day.

```bash
export DB=/path/to/mirror.sqlite
export AGENT_RUN_ID="triage-agent-v2-20260403"
export SOURCE_LABEL="triage-agent-v2"
export CREATED_BY="pi-agent"
```

Use these variables consistently in every command. Never leave `--agent-run-id` empty. If you forget, you lose the ability to query "what did this run do?" and your work becomes an orphaned mess in the database.

## Step 1: Log The Run Start

Before annotating anything, create a log entry that records what you are about to do, why, and what strategy you chose.

```bash
START_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind decision \
  --title "Run started — triage batch" \
  --body "Starting triage run for 23 unclassified senders. Using volume heuristics + content analysis. Will check message counts, List-Unsubscribe headers, subject line patterns, and body structure to classify each sender." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Link the decision to the account being triaged
smailnail annotate log link-target \
  --sqlite-path "$DB" \
  --log-id "$START_LOG_ID" \
  --target-type account \
  --target-id manuel \
  --output json > /dev/null 2>&1
```

Save the returned log `id`. You will link targets to it later.

Use `--log-kind` to communicate intent:

| log-kind | When to use | Frequency |
|---|---|---|
| `decision` | Starting a run, choosing a strategy, switching approach mid-run, explaining why you skipped something | 1-3 per run |
| `reasoning` | Per-target classification reasoning — why you tagged this specific sender/message | **One per target** (this is the most important kind) |
| `summary` | End-of-run summary with counts, outcomes, and what to review | 1 per run |
| `note` | Observations during investigation — interesting findings, data quality issues, things that surprised you | As many as needed |

## Step 2: Create Annotations With Full Metadata

Every annotation must include all metadata fields. Never rely on defaults.

```bash
smailnail annotate annotation add \
  --sqlite-path "$DB" \
  --target-type sender \
  --target-id "news@techcrunch.com" \
  --tag "newsletter" \
  --note "Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: TechCrunch Daily - {date}. Classified as newsletter." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

### Required fields checklist

| Field | What to set | Never |
|---|---|---|
| `--target-type` | `sender`, `domain`, `message`, `thread`, `mailbox`, or `account` | Leave empty |
| `--target-id` | The stable identifier (email address, domain, message id, thread id) | Use an unstable or ambiguous id |
| `--tag` | A short, consistent tag from the agreed taxonomy | Invent new tags without documenting them |
| `--note` | Your reasoning: what you observed, why you chose this tag, key evidence | Leave empty or write "classified" |
| `--source-kind` | `agent` for automated work, `heuristic` for rule-based | Use `human` (that is for human-created annotations) |
| `--source-label` | Your agent name and version, e.g., `triage-agent-v2` | Leave empty |
| `--agent-run-id` | The run ID you generated in step 0 | Leave empty |
| `--created-by` | Your identity, e.g., `pi-agent` | Leave empty |

### Writing good notes

The note is the most important field for reviewers. Write it as if explaining your classification to a skeptical human who cannot see the raw data.

Good note:
> Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: `TechCrunch Daily - {date}`. Classified as **newsletter**.

Bad note:
> newsletter

Good note:
> GitHub notifications sender. 312 messages in archive. Mix of PR reviews, issue updates, and CI notifications. Tagged as **notification** rather than bulk-sender because content is personalized.

Bad note:
> lots of github emails

### Multiple tags on the same target

You can and should add multiple annotations to the same target when it has multiple roles. Each annotation is its own row with its own review state.

```bash
# Kryzak is both a personal contact AND a lawyer
smailnail annotate annotation add ... --target-id "kryzak@yahoo.com" --tag "personal" --note "..." ...
smailnail annotate annotation add ... --target-id "kryzak@yahoo.com" --tag "important/legal" --note "..." ...
```

### Idempotency

Before adding an annotation, check if one already exists with the same tag on the same target. Do not create duplicates.

```bash
existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
if [ "$existing" -gt 0 ]; then
  echo "SKIP: $email already has tag $tag"
  return
fi
```

## Step 3: Log Your Reasoning Per Target

Create a reasoning log entry for **every target you classify**. This is the most important part of the handbook. Each reasoning entry appears in the annotation UI timeline with its timestamp, author badge, and log-kind chip — it is what makes agent work reviewable.

```bash
# Create the reasoning log
LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind reasoning \
  --title "Sender classification: news@techcrunch.com" \
  --body "Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: \`TechCrunch Daily - {date}\`. Classified as **newsletter**." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Link to the target — NEVER skip this step
smailnail annotate log link-target \
  --sqlite-path "$DB" \
  --log-id "$LOG_ID" \
  --target-type sender \
  --target-id "news@techcrunch.com" \
  --output json > /dev/null 2>&1
```

### The default is one reasoning log per target

This is not optional. Every sender, domain, or message you annotate should have a reasoning log linked to it. The reviewer sees a timeline like:

```
06:29:00  [decision]  Run started — triage batch             🤖 triage-agent-v2
06:30:00  [reasoning] Sender classification: news@tc.com     🤖 triage-agent-v2
06:31:00  [reasoning] Sender classification: noreply@gh.com  🤖 triage-agent-v2
06:32:00  [reasoning] Sender classification: promo@x.com     🤖 triage-agent-v2
06:40:00  [summary]   Run complete — 23 senders classified   🤖 triage-agent-v2
```

Without per-target reasoning entries, the timeline is just two bookend entries and the reviewer has no idea why any individual classification was made.

### What to include in reasoning logs

Every reasoning log body should contain:

1. **What you observed** — message count, subject patterns, body structure, headers
2. **Key evidence** — the specific signal that drove the classification
3. **Your classification** — the tag you chose, bolded
4. **Why this tag and not another** — especially for borderline cases

Examples of good reasoning:

> Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: `TechCrunch Daily - {date}`. Classified as **newsletter**.

> GitHub notifications sender. 312 messages in archive. Mix of PR reviews, issue updates, and CI notifications. Tagged as **notification** rather than bulk-sender because content is personalized.

> Only 3 messages from this sender, all with personal greetings and unique content. Last message was a reply to Manuel's email about camera repair. Classified as **personal**.

> CPA firm — David Miller. 11 messages about 2024 tax returns, stock option exercise questions, invoices via QuickBooks. Tagged as **important/tax** because tax correspondence has real deadlines and financial consequences.

> 55 messages from various senders at costsoldier.com. All subjects are B2B software marketing ("QuickBooks alternatives", "The Ultimate Contractor Software"). No legitimate business relationship. Tagged as **noise/spam**.

### Batch reasoning for obvious groups

When classifying a batch of senders that all match the same pattern (e.g., 14 senders at `costsoldier.com` that are all clearly spam), you may create one reasoning log and link it to all targets:

```bash
BATCH_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind reasoning \
  --title "Batch classification: costsoldier.com — 14 senders" \
  --body "All 14 senders at costsoldier.com send B2B software marketing spam. Same domain, same subject patterns (\`QuickBooks\`, \`Netsuite\`, \`Payroll\`), no business relationship. Classified all as **noise/spam**." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Link to ALL senders in the batch
for email in concur@costsoldier.com netsuite@costsoldier.com payroll@costsoldier.com ...; do
  smailnail annotate log link-target \
    --sqlite-path "$DB" \
    --log-id "$BATCH_LOG_ID" \
    --target-type sender \
    --target-id "$email" \
    --output json > /dev/null 2>&1
done
```

But use batch reasoning only when the evidence genuinely is the same for all targets. If you had to think differently about any target in the batch, give it its own reasoning log.

### Investigation notes along the way

During your investigation (before you start annotating), log interesting findings as `note` entries. These are valuable context for reviewers even if they do not directly produce annotations.

```bash
smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind note \
  --title "Finding: 1,523 messages have empty sender_email" \
  --body "All 1,523 empty-sender messages are from chatgpt-codex-connector[bot] via GitHub notifications. The from_summary shows the sender but sender normalization did not extract an email. These messages cannot be annotated by sender — would need message-level or thread-level annotations." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

```bash
smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind note \
  --title "Finding: GitHub CI failures are 12% of entire inbox" \
  --body "4,003 messages match subject patterns 'PR run failed' or 'Run failed'. This is the single largest noise category and a strong candidate for Sieve filtering at the IMAP level." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

### Decision logs when you change strategy mid-run

If you discover something during the run that changes your approach, log it as a `decision`:

```bash
smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind decision \
  --title "Strategy change: adding important/* tags on top of base categories" \
  --body "Realized that volume-based triage misses high-consequence senders. A CPA sending 11 messages is low-volume but critical. Adding a second pass with important/* tags for tax, legal, equity, housing, health, and conference senders. These will be added as additional annotations alongside the base category tags." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

## Step 4: Create Groups For Review Clusters

When you identify a set of targets that belong together for review, create a group. Always include `--agent-run-id`.

```bash
GROUP_ID=$(smailnail annotate group create \
  --sqlite-path "$DB" \
  --name "Unsubscribe Candidates" \
  --description "High-volume senders with List-Unsubscribe headers — candidates for unsubscribing" \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)
```

Then add targets. **Important:** capture the group ID cleanly. Do not let `echo` output contaminate the variable.

```bash
# CORRECT: use --select id to get just the UUID
GROUP_ID=$(smailnail annotate group create ... --select id)

# WRONG: parsing from table output will include echo noise
GROUP_ID=$(smailnail annotate group create ...)  # includes table headers
```

Add members:

```bash
smailnail annotate group add-target \
  --sqlite-path "$DB" \
  --group-id "$GROUP_ID" \
  --target-type sender \
  --target-id "news@techcrunch.com" \
  --output json > /dev/null 2>&1
```

## Step 5: Log The Run Completion

After all annotations are created, log a summary with counts, outcomes, and anything the reviewer should pay attention to. Link it to all targets processed in the run.

```bash
SUMMARY_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind summary \
  --title "Run complete — 23 senders classified" \
  --body "Classified 23 senders:
- 8 newsletters
- 6 notifications
- 5 bulk-sender
- 3 transactional
- 1 important

All annotations created with review_state=to_review.

**Needs human attention:**
- dave@davemillercpa.com has a past-due invoice from Dec 2025
- kryzaklaw@yahoo.com — separation agreement may need follow-up

**Data quality issues found:**
- 1,523 messages have empty sender_email (all GitHub bot notifications)
- Experian alerts tagged financial but are mostly marketing

**Created 1 group:** Unsubscribe Candidates (86 members)" \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Link to ALL targets processed in this run — not just representative ones
sqlite3 "$DB" "
SELECT DISTINCT target_id FROM annotations
WHERE agent_run_id = '$AGENT_RUN_ID' AND target_type = 'sender';" | while read -r email; do
  smailnail annotate log link-target \
    --sqlite-path "$DB" \
    --log-id "$SUMMARY_LOG_ID" \
    --target-type sender \
    --target-id "$email" \
    --output json > /dev/null 2>&1
done
```

The summary body should include:

1. **Counts by tag** — what was classified and how
2. **Items needing human attention** — anything time-sensitive or ambiguous
3. **Data quality issues** — things you noticed but could not fix
4. **Groups created** — what review clusters exist
5. **What was NOT classified** — targets you skipped and why (e.g., "skipped 8,767 messages from senders with <10 messages")

## The Complete Agent Run Structure

A well-formed agent run produces a dense, reviewable timeline in the database. For a run that classifies 23 senders, you should see roughly 28+ log entries:

```
Agent Run: triage-agent-v2-20260403
│
├── Log [decision]: "Run started — triage batch"
│   body: "Starting triage run for 23 unclassified senders. Using volume
│          heuristics + content analysis."
│   └── linked to: account:manuel
│
├── Log [note]: "Finding: 48% of inbox is automated noise"
│   body: "GitHub CI failures alone are 4,003 messages (12%). Newsletters
│          are 11%. Commerce/transactional 6%."
│   └── linked to: account:manuel
│
├── Annotation: sender:news@techcrunch.com → tag:newsletter
│   └── Log [reasoning]: "Sender classification: news@techcrunch.com"
│       body: "47 messages. Identical HTML structure. Unsubscribe headers.
│              Subject pattern: TechCrunch Daily - {date}. → newsletter."
│       └── linked to: sender:news@techcrunch.com
│
├── Annotation: sender:noreply@github.com → tag:notification
│   └── Log [reasoning]: "Sender classification: noreply@github.com"
│       body: "312 messages. Mix of PR reviews, issue updates, CI. Tagged
│              notification not bulk-sender — content is personalized."
│       └── linked to: sender:noreply@github.com
│
├── Annotations: 14 senders at costsoldier.com → tag:noise/spam
│   └── Log [reasoning]: "Batch classification: costsoldier.com — 14 senders"
│       body: "All B2B software marketing spam. Same patterns."
│       └── linked to: all 14 senders at costsoldier.com
│
├── Log [decision]: "Strategy change: adding important/* tags"
│   body: "Realized volume-based triage misses high-consequence senders.
│          Adding second pass with important/* for tax, legal, equity."
│   └── linked to: account:manuel
│
├── Annotation: sender:dave@davemillercpa.com → tag:important/tax
│   └── Log [reasoning]: "Sender classification: dave@davemillercpa.com"
│       body: "CPA firm. 11 messages about 2024 tax returns, stock option
│              exercise questions. Invoices via QuickBooks. Past-due invoice
│              from Dec 2025. → important/tax."
│       └── linked to: sender:dave@davemillercpa.com
│
├── Group: "Unsubscribe Candidates" (86 members)
│   └── Log [reasoning]: "Group creation: Unsubscribe Candidates"
│       body: "Created group of 86 noise senders that have List-Unsubscribe
│              headers. All tagged noise/*. Ready for human review before
│              actually unsubscribing."
│       └── linked to: all 86 members
│
└── Log [summary]: "Run complete — 23 senders classified"
    body: "Classified 23 senders:
           • 8 newsletters
           • 6 notifications
           • 5 bulk-sender
           • 3 transactional
           • 1 important
           All annotations created with review_state=to_review."
    └── linked to: all 23 senders
```

The rule of thumb: **if you made a decision or observed something interesting, it should be a log entry**. A reviewer scrolling the timeline should be able to reconstruct your entire reasoning process without looking at any external notes.

Every row has the same `agent_run_id`, `source_label`, and `created_by`. A reviewer can query the entire run with:

```sql
-- Full timeline for this run
SELECT l.created_at, l.log_kind, l.title, l.body_markdown
FROM annotation_logs l
WHERE l.agent_run_id = 'triage-agent-v2-20260403'
ORDER BY l.created_at;

-- What annotations were created?
SELECT tag, COUNT(*) FROM annotations
WHERE agent_run_id = 'triage-agent-v2-20260403'
GROUP BY tag ORDER BY COUNT(*) DESC;

-- What does a specific target's timeline look like?
SELECT l.created_at, l.log_kind, l.title, l.body_markdown
FROM annotation_logs l
JOIN annotation_log_targets lt ON lt.log_id = l.id
WHERE lt.target_type = 'sender'
  AND lt.target_id = 'dave@davemillercpa.com'
ORDER BY l.created_at;
```

## Querying Your Own Run

After your run, verify it looks correct:

```sql
-- What did this run create?
SELECT tag, COUNT(*) as cnt FROM annotations
WHERE agent_run_id = 'triage-agent-v2-20260403'
GROUP BY tag ORDER BY cnt DESC;

-- What logs were written?
SELECT id, log_kind, title, created_at FROM annotation_logs
WHERE agent_run_id = 'triage-agent-v2-20260403'
ORDER BY created_at;

-- Are all logs linked to targets?
SELECT l.title, COUNT(lt.target_id) as linked_targets
FROM annotation_logs l
LEFT JOIN annotation_log_targets lt ON lt.log_id = l.id
WHERE l.agent_run_id = 'triage-agent-v2-20260403'
GROUP BY l.id;

-- Any logs with zero links? (bad — fix these)
SELECT l.id, l.title FROM annotation_logs l
WHERE l.agent_run_id = 'triage-agent-v2-20260403'
  AND NOT EXISTS (SELECT 1 FROM annotation_log_targets lt WHERE lt.log_id = l.id);
```

## Tag Taxonomy

Use the established tag hierarchy. Do not invent new top-level categories without documenting them.

| Tag | When to use |
|---|---|
| `noise/ci` | CI failure notifications, automated build alerts |
| `noise/marketing` | Marketing, sales, promotions |
| `noise/transactional` | Order confirmations, shipping, receipts |
| `noise/social-notif` | Social platform notifications |
| `noise/spam` | Outright junk |
| `newsletter/tech` | Tech newsletters worth reading |
| `newsletter/culture` | Non-tech culture newsletters |
| `newsletter/creative` | Music, photography, film, art |
| `personal` | Real humans writing directly |
| `work` | Work-related (employer, work tools) |
| `community` | Mailing lists, hackerspaces, meetups |
| `financial` | Banking, payments, credit |
| `services` | Services the user uses (hosting, apps, etc.) |
| `hobby` | Personal interests (3D printing, gaming, fitness) |
| `important/tax` | Tax/CPA correspondence |
| `important/legal` | Lawyer, legal notices, contracts |
| `important/equity` | Stock options, equity platforms |
| `important/work-admin` | Compensation, HR, equity management |
| `important/housing` | Rent, lease, property management |
| `important/health` | Medical, therapy, health services |
| `important/conferences` | Conference invitations, CFPs, speaking |

If you need a new tag, document it in the summary log with a clear definition and examples.

## Saving Investigation Queries

Save every SQL query and shell command you run as a numbered file in the ticket's `scripts/` directory. Use the format:

```
00-08: action scripts (.sh) — scripts that modify the database
09+:   investigation queries (.sql) — read-only queries for analysis
```

Name pattern: `{NN}-{verb}-{what}.{sql|sh}`

Examples:
- `09-investigate-schema.sql`
- `15-investigate-recent-messages.sql`
- `30-investigate-tax-emails.sql`
- `01-categorize-noise-senders.sh`

This creates a full audit trail that a human (or future agent) can replay step by step.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Annotations have empty `agent_run_id` | You forgot to pass `--agent-run-id` | Backfill with `UPDATE annotations SET agent_run_id='...' WHERE source_label='...' AND agent_run_id='';` |
| Log entries exist but `annotation_log_targets` is empty | You created logs but never linked them | Run `smailnail annotate log link-target` for each log |
| Group membership has corrupted `group_id` values | Script output (echo) was captured into the group_id variable | Use `--select id` to capture just the UUID, keep echo separate |
| Duplicate annotations on same target+tag | Missing idempotency check | Always check before inserting: `SELECT COUNT(*) FROM annotations WHERE target_type=? AND target_id=? AND tag=?` |
| Review state is `reviewed` instead of `to_review` for agent work | Used `--source-kind human` or defaulted | Always use `--source-kind agent` for automated work; the system defaults agent rows to `to_review` |
| Cannot query run because multiple runs share the same run ID | Reused the run ID across sessions | Always include the date (and sequence number if needed) in the run ID |

## See Also

- `smailnail help smailnail-annotate-sqlite-playbook` for the basic annotation workflow
- `smailnail help smailnail-mirror-overview` for the mirror DB model
- `smailnail annotate --help` for the command tree
