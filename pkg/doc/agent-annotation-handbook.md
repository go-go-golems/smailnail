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

## Phase 0: Prepare The Database

Before you can annotate senders, the mirror database must have enriched sender and thread data. If the database was just synced, run enrichment first:

```bash
export DB=/path/to/mirror.sqlite

smailnail enrich all --sqlite-path "$DB"
```

This produces three enrichment passes in sequence:

1. **Senders** — normalizes sender emails from `from_summary`, populates the `senders` table with `email`, `display_name`, `domain`, `msg_count`
2. **Threads** — reconstructs message threads from `In-Reply-To`/`References` headers, populates the `threads` table
3. **Unsubscribe** — extracts `List-Unsubscribe` headers into `unsubscribe_mailto`, `unsubscribe_http`, `has_list_unsubscribe` on the `senders`

After enrichment, verify the results:

```bash
sqlite3 "$DB" "
SELECT
  (SELECT COUNT(*) FROM senders) as sender_count,
  (SELECT COUNT(*) FROM messages WHERE sender_email != '') as tagged_msgs,
  (SELECT COUNT(*) FROM messages WHERE sender_email = '')as untagged_msgs,
  (SELECT COUNT(*) FROM threads)as thread_count
FROM senders LIMIT 1;
"
```

### What to do with untagged messages

Messages with empty `sender_email` cannot be annotated by sender. Log these as a finding. Common causes:

- Bot senders with complex `From` headers that normalization misses (e.g., `chatgpt-codex-connector[bot] <notifications@github.com>`)
- Legitimate senders with non-standard address formatting

```bash
sqlite3 "$DB" "SELECT id, subject, from_summary FROM messages WHERE sender_email = '' LIMIT 10;"
```

## Phase 1: Generate An Agent Run ID

Every annotation session needs a unique `agent_run_id`. Generate it before your first write and pass it to every command in the session.

Format: `{source-label}-{YYYYMMDD}` or `{source-label}-{YYYYMMDD}-{seq}` if you run multiple passes per day.

```bash
export DB=/path/to/mirror.sqlite
export AGENT_RUN_ID="triage-agent-v2-20260403"
export SOURCE_LABEL="triage-agent-v2"
export CREATED_BY="pi-agent"
```

Use these variables consistently in every command. Never leave `--agent-run-id` empty. If you forget, you lose the ability to query "what did this run do?" and your work becomes an orphaned mess in the database.

## Phase 2: Investigate Before Annotating

Before writing any annotations, run investigation queries to understand the inbox composition. This phase produces the evidence that drives your classifications.

### Essential investigation queries

Run these in order and save findings as `note` log entries:

```bash
# 1. Overall stats
sqlite3 "$DB" "
SELECT COUNT(*) as total_msgs,
       MIN(internal_date) as earliest,
       MAX(internal_date) as latest
FROM messages;
"

# 2. Top senders by volume
sqlite3 "$DB" -column -header "
SELECT email, display_name, domain, msg_count,
       has_list_unsubscribe, first_seen_date, last_seen_date
FROM senders
ORDER BY msg_count DESC;
"

# 3. Domain distribution
sqlite3 "$DB" -column -header "
SELECT domain, COUNT(*) as sender_cnt, SUM(msg_count) as total_msgs
FROM senders
GROUP BY domain
ORDER BY total_msgs DESC
LIMIT 20;
"

# 4. Messages without sender_email
sqlite3 "$DB" -column -header "
SELECT id, subject, from_summary
FROM messages
WHERE sender_email = '';
"

# 5. Subject patterns (duplicates = bulk mail)
sqlite3 "$DB" -column -header "
SELECT subject, COUNT(*) as cnt
FROM messages
GROUP BY subject
HAVING cnt > 1
ORDER BY cnt DESC
LIMIT 30;
"
```

### Log findings as notes

Record each interesting finding as a `note` log entry:

```bash
smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind note \
  --title "Finding: N messages have empty sender_email" \
  --body "Description of what you found and why it matters." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

## Phase 3: Log The Run Start

Before annotating anything, create a log entry that records what you are about to do, why, and what strategy you chose.

```bash
START_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind decision \
  --title "Run started — triage batch" \
  --body "Starting triage run for N unclassified senders. Using volume heuristics + content analysis. Will check message counts, List-Unsubscribe headers, subject line patterns, and body structure to classify each sender." \
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

## Phase 4: Classify And Annotate Senders

This is the core work. For each sender, produce an annotation AND a reasoning log.

### The annotate_sender helper function

Define this shell function early in your session to avoid repetitive errors:

```bash
annotate_sender() {
  local email="$1"
  local tag="$2"
  local note="$3"
  local reasoning="$4"

  # Idempotency check — always check before inserting
  local existing
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag' AND agent_run_id='$AGENT_RUN_ID';")
  if [ "$existing" -gt 0 ]; then
    echo "SKIP: $email already has tag $tag"
    return
  fi

  # Add annotation
  smailnail annotate annotation add \
    --sqlite-path "$DB" \
    --target-type sender \
    --target-id "$email" \
    --tag "$tag" \
    --note "$note" \
    --source-kind agent \
    --source-label "$SOURCE_LABEL" \
    --agent-run-id "$AGENT_RUN_ID" \
    --created-by "$CREATED_BY" \
    --output json > /dev/null 2>&1

  # Add reasoning log
  local LOG_ID
  LOG_ID=$(smailnail annotate log add \
    --sqlite-path "$DB" \
    --log-kind reasoning \
    --title "Sender classification: $email" \
    --body "$reasoning" \
    --source-kind agent \
    --source-label "$SOURCE_LABEL" \
    --agent-run-id "$AGENT_RUN_ID" \
    --created-by "$CREATED_BY" \
    --select id)

  # Link log to target — NEVER skip this step
  smailnail annotate log link-target \
    --sqlite-path "$DB" \
    --log-id "$LOG_ID" \
    --target-type sender \
    --target-id "$email" \
    --output json > /dev/null 2>&1

  echo "OK: $email -> $tag"
}
```

Usage:

```bash
annotate_sender "news@techcrunch.com" "newsletter/tech" \
  "Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers." \
  "47 messages. Identical HTML structure. Unsubscribe headers. Subject pattern: TechCrunch Daily - {date}. Classified as **newsletter/tech**."
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
| `--agent-run-id` | The run ID you generated in phase 1 | Leave empty |
| `--created-by` | Your identity, e.g., `pi-agent` | Leave empty |

### Writing good notes and reasoning

The note is for the annotation row. The reasoning is for the timeline log. Both matter for reviewers.

Good reasoning:
> 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: `TechCrunch Daily - {date}`. Classified as **newsletter/tech**.

Bad reasoning:
> newsletter

Good reasoning:
> GitHub notifications sender. 312 messages in archive. Mix of PR reviews, issue updates, and CI notifications. Tagged as **work** because GitHub is a primary work tool.

Bad reasoning:
> lots of github emails

Every reasoning log body should contain:

1. **What you observed** — message count, subject patterns, body structure, headers
2. **Key evidence** — the specific signal that drove the classification
3. **Your classification** — the tag you chose, bolded
4. **Why this tag and not another** — especially for borderline cases

### Multiple tags on the same target

You can and should add multiple annotations to the same target when it has multiple roles. Each annotation is its own row with its own review state.

```bash
# Kryzak is both a personal contact AND a lawyer
annotate_sender "kryzak@yahoo.com" "personal" "..." "Personal contact — real human, personal correspondence. Classified as **personal**."
annotate_sender "kryzak@yahoo.com" "important/legal" "..." "Lawyer. Legal correspondence with deadlines. Classified as **important/legal**."
```

### Batch reasoning for obvious groups

When classifying a batch of senders that all match the same pattern (e.g., 14 senders at `costsoldier.com` that are all clearly spam), you may create one reasoning log and link it to all targets:

```bash
BATCH_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind reasoning \
  --title "Batch classification: costsoldier.com — 14 senders" \
  --body "All 14 senders at costsoldier.com send B2B software marketing spam. Same domain, same subject patterns, no business relationship. Classified all as **noise/spam**." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Create annotations individually, but link ONE reasoning log to ALL senders
for email in concur@costsoldier.com netsuite@costsoldier.com payroll@costsoldier.com; do
  # Create annotation per sender
  smailnail annotate annotation add \
    --sqlite-path "$DB" \
    --target-type sender \
    --target-id "$email" \
    --tag "noise/spam" \
    --note "B2B marketing spam from costsoldier.com" \
    --source-kind agent \
    --source-label "$SOURCE_LABEL" \
    --agent-run-id "$AGENT_RUN_ID" \
    --created-by "$CREATED_BY" \
    --output json > /dev/null 2>&1

  # Link shared reasoning to each sender
  smailnail annotate log link-target \
    --sqlite-path "$DB" \
    --log-id "$BATCH_LOG_ID" \
    --target-type sender \
    --target-id "$email" \
    --output json > /dev/null 2>&1
done
```

But use batch reasoning only when the evidence genuinely is the same for all targets. If you had to think differently about any target in the batch, give it its own reasoning log.

### Decision logs when you change strategy mid-run

If you discover something during the run that changes your approach, log it as a `decision`:

```bash
smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind decision \
  --title "Strategy change: adding important/* tags on top of base categories" \
  --body "Realized that volume-based triage misses high-consequence senders. A CPA sending 11 messages is low-volume but critical. Adding a second pass with important/* tags for tax, legal, equity, housing, health, and conference senders." \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --output json > /dev/null 2>&1
```

## Phase 5: Create Groups For Review Clusters

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

**Important:** capture the group ID cleanly with `--select id`. Do not let `echo` output contaminate the variable.

Add members programmatically:

```bash
# Example: add all noise/* senders that have unsubscribe headers
sqlite3 "$DB" "
SELECT DISTINCT s.email
FROM senders s
JOIN annotations a ON a.target_id = s.email AND a.target_type = 'sender'
WHERE s.has_list_unsubscribe = 1
  AND a.tag LIKE 'noise/%'
  AND a.agent_run_id = '$AGENT_RUN_ID';
" | while read -r email; do
  smailnail annotate group add-target \
    --sqlite-path "$DB" \
    --group-id "$GROUP_ID" \
    --target-type sender \
    --target-id "$email" \
    --output json > /dev/null 2>&1
done
```

## Phase 6: Log The Run Completion

After all annotations are created, log a summary with counts, outcomes, and anything the reviewer should pay attention to. Link it to all targets processed in the run.

```bash
# Get tag counts for the summary body
sqlite3 "$DB" -column -header "
SELECT tag, COUNT(*) as cnt FROM annotations
WHERE agent_run_id = '$AGENT_RUN_ID'
GROUP BY tag ORDER BY cnt DESC;
"

SUMMARY_LOG_ID=$(smailnail annotate log add \
  --sqlite-path "$DB" \
  --log-kind summary \
  --title "Run complete — N senders classified" \
  --body "Classified N senders:
- tag1: X senders
- tag2: Y senders
...

All annotations created with review_state=to_review.

**Needs human attention:**
- specific items that need review

**Data quality issues found:**
- messages with empty sender_email

**Created groups:** Group Name (N members)" \
  --source-kind agent \
  --source-label "$SOURCE_LABEL" \
  --agent-run-id "$AGENT_RUN_ID" \
  --created-by "$CREATED_BY" \
  --select id)

# Link to ALL targets processed in this run
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

# Also link to the account
smailnail annotate log link-target \
  --sqlite-path "$DB" \
  --log-id "$SUMMARY_LOG_ID" \
  --target-type account \
  --target-id manuel \
  --output json > /dev/null 2>&1
```

The summary body should include:

1. **Counts by tag** — what was classified and how
2. **Items needing human attention** — anything time-sensitive or ambiguous
3. **Data quality issues** — things you noticed but could not fix
4. **Groups created** — what review clusters exist
5. **What was NOT classified** — targets you skipped and why (e.g., "skipped messages with empty sender_email")

## Phase 7: Verify The Run

After your run, verify it looks correct. Run all four verification queries:

```bash
echo "=== Tag counts ==="
sqlite3 "$DB" -column -header "
SELECT tag, COUNT(*) as cnt FROM annotations
WHERE agent_run_id = '$AGENT_RUN_ID'
GROUP BY tag ORDER BY cnt DESC;"

echo "=== Log entries ==="
sqlite3 "$DB" -column -header "
SELECT log_kind, COUNT(*) as cnt FROM annotation_logs
WHERE agent_run_id = '$AGENT_RUN_ID'
GROUP BY log_kind ORDER BY cnt DESC;"

echo "=== Any logs with zero links? (must be empty) ==="
sqlite3 "$DB" -column -header "
SELECT l.id, l.title FROM annotation_logs l
WHERE l.agent_run_id = '$AGENT_RUN_ID'
  AND NOT EXISTS (SELECT 1 FROM annotation_log_targets lt WHERE lt.log_id = l.id);"

echo "=== Group membership ==="
sqlite3 "$DB" -column -header "
SELECT g.name, COUNT(gm.target_id) as member_count
FROM target_groups g
LEFT JOIN target_group_members gm ON gm.group_id = g.id
WHERE g.agent_run_id = '$AGENT_RUN_ID'
GROUP BY g.id;"
```

Expected verification results:
- Tag counts: should show all categories you classified
- Log entries: should show ~4 log kinds (decision, note, reasoning, summary) with reasoning count matching annotation count
- Zero-link logs: **must be empty** — if any appear, link them to targets immediately
- Group membership: should show created groups with expected member counts

## The Complete Agent Run Structure

A well-formed agent run produces a dense, reviewable timeline in the database. For a run that classifies 91 senders, you should see roughly 95+ log entries:

```
Agent Run: triage-agent-v2-20260403
│
├── Log [decision]: "Run started — triage batch"
│   body: "Starting triage run for 91 unclassified senders. Using volume
│          heuristics + content analysis."
│   └── linked to: account:manuel
│
├── Log [note]: "Finding: 23 messages have empty sender_email"
     body: "All are from chatgpt-codex-connector[bot]. Cannot annotate by sender."
│   └── linked to: account:manuel
│
├── Annotation: sender:notifications@github.com → tag:work
│   └── Log [reasoning]: "Sender classification: notifications@github.com"
│       body: "234 messages (58.5% of total). All GitHub notifications.
│              Tagged as **work** because GitHub is a primary work tool."
│       └── linked to: sender:notifications@github.com
│
├── ... (one reasoning log per sender) ...
│
├── Group: "Unsubscribe Candidates" (15 members)
│   └── linked to: all 15 noise/* senders with List-Unsubscribe
│
└── Log [summary]: "Run complete — 91 senders classified"
    body: "Classified 91 senders:
           • noise/marketing: 16
 services: 15, community: 13
 newsletter/tech: 11. ..."
    └── linked to: all 91 senders + account:manuel
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
  AND lt.target_id = 'notifications@github.com'
ORDER BY l.created_at;
```

## Tag Taxonomy

Use the established tag hierarchy. Do not invent new top-level categories without documenting them.

| Tag | When to use |
|---|---|
| `noise/ci` | CI failure notifications, automated build alerts |
| `noise/marketing` | Marketing, sales, promotions |
| `noise/transactional` | Order confirmations, shipping, receipts |
| `noise/social-notif` | Social platform notifications |
| `noise/spam` | Outright junk, unsolicited cold outreach |
| `newsletter/tech` | Tech newsletters worth reading |
| `newsletter/culture` | Non-tech culture newsletters |
| `newsletter/creative` | Music, photography, film, art |
| `personal` | Real humans writing directly |
| `work` | Work-related (employer, work tools) |
| `community` | Mailing lists, hackerspaces, meetups |
| `financial` | Banking, payments, credit, billing |
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

### Classification decision guide

When classifying a sender, use these signals in priority order:

1. **Is it a real human writing to Manuel personally?** → `personal` (look for personal greetings, replies to Manuel's emails, unique content)
2. **Is it work tooling/infrastructure?** → `work` (GitHub, Slack, Rollbar, CI systems the user actively uses for work)
3. **Does it have financial/legal/tax consequences?** → `important/*` or `financial` (CPA, lawyer, bank, payment processor, billing)
4. **Is it a service Manuel actively uses?** → `services` (hosting providers, app subscriptions, domain registrars, tools with active accounts)
5. **Is it a community/mailing list Manuel participates in?** → `community` (mailing lists with `pro-leave` in unsubscribe, hackerspaces, meetups, professional groups)
6. **Is it a hobby/interest?** → `hobby` (fitness, gaming, photography galleries, maker spaces)
7. **Is it a newsletter Manuel subscribed to?** → `newsletter/*` (has List-Unsubscribe, consistent format, regular cadence)
8. **Is it automated noise?** → `noise/*`:
   - `noise/social-notif` — social platform notifications (Facebook, Instagram, Twitch, SoundCloud)
   - `noise/marketing` — B2B outreach, retail promos, product announcements
   - `noise/transactional` — order confirmations, shipping updates
   - `noise/ci` — CI/build failure notifications
   - `noise/spam` — outright junk, unsolicited email from unknown senders

**Borderline cases:**

| Sender type | Tag as | Not |
|---|---|---|
| Credit monitoring service (Experian) | `financial` | `noise/marketing` (even if some messages are marketing, the service has financial relevance) |
| Photography gallery announcements | `hobby` | `newsletter/creative` (it's a personal interest, not a newsletter) |
| Event platform (Meetup) for a tech group | `community` | `noise/social-notif` (the user chose to join the group) |
| Amazon order confirmation | `noise/transactional` | `services` (the user doesn't need to track these) |
| Notification from a service the user pays for | `services` | `noise/*` (active paid service) |

## Saving Investigation Queries

Save every SQL query and shell command you run as a numbered file in the ticket's `scripts/` directory. Use the format:

```
00-08: action scripts (.sh) — scripts that modify the database
09+:   investigation queries (.sql) — read-only queries for analysis
```

Name pattern: `{NN}-{verb}-{what}.{sql|sh}`

Examples:
- `09-investigate-schema.sql`
- `15-investigate-top-senders.sql`
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
| `sqlite-path is required` but you passed the flag | Embedded struct fields not decoded by glazed | Check that settings structs use direct fields with `glazed` tags, not embedded anonymous structs |
| Senders table is empty | Enrichment was not run | Run `smailnail enrich all --sqlite-path "$DB"` before annotating |

## See Also

- `smailnail help smailnail-annotate-sqlite-playbook` for the basic annotation workflow
- `smailnail help smailnail-mirror-overview` for the mirror DB model
- `smailnail annotate --help` for the command tree
