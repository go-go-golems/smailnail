---
Title: Annotate And Query The Mirror SQLite DB
Slug: smailnail-annotate-sqlite-playbook
Short: Use `smailnail annotate` to add annotations, organize targets into groups, attach logs, and query everything back out of the local mirror database.
Topics:
- sqlite
- glazed
- cli
- email
Commands:
- annotate
- annotate annotation add
- annotate annotation list
- annotate annotation review
- annotate group create
- annotate group list
- annotate group add-target
- annotate group members
- annotate log add
- annotate log list
- annotate log link-target
- annotate log targets
Flags:
- sqlite-path
- target-type
- target-id
- tag
- note
- review-state
- source-kind
- source-label
- agent-run-id
- created-by
- group-id
- log-id
- limit
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This playbook shows how to use the SQLite-backed annotation workflow that now lives in `smailnail annotate`. The goal is to work against a local mirror database, not the live IMAP server. That distinction matters because annotations, groups, and logs are stored in the mirror SQLite file and can be inspected, copied, and versioned independently of any current IMAP session.

## Why Use The Annotation CLI

Use `smailnail annotate` when you want to leave durable judgment in the mirror database itself. The annotation system is intentionally simple in this first version:

- annotations attach a `tag` and optional free-form note to a generic target,
- groups collect multiple targets under one name,
- logs record one note that can be linked to many targets,
- review state tells you whether a row still needs human attention.

That is enough to support practical workflows such as:

- mark `notifications@github.com` as important but noisy,
- collect likely newsletters into one review group,
- attach one agent-run explanation to many senders,
- query unresolved agent-created rows later.

## What Lives In SQLite

When you point the command at a mirror DB, the CLI bootstraps schema version 3 if needed and uses these tables:

- `annotations`
- `target_groups`
- `target_group_members`
- `annotation_logs`
- `annotation_log_targets`

All of those tables are keyed by generic targets:

- `target_type`
- `target_id`

The intended first target types are:

- `message`
- `thread`
- `sender`
- `domain`
- `mailbox`
- `account`

The main advantage of this model is that you do not need a separate command family for every entity shape. One command surface works across all mirrored entity types.

## Prerequisites

- A build that includes `sqlite_fts5`
- A mirror DB created by `smailnail mirror` or otherwise bootstrapped through the mirror store
- A stable SQLite path you intend to keep using

If you are experimenting, copy the DB first so you can throw the copy away:

```bash
cp /tmp/smailnail-parallel-a.sqlite /tmp/smailnail-annotate-demo.sqlite
```

## Step 1: Pick The SQLite File

Every command in this playbook takes `--sqlite-path`. Use the same file repeatedly or you will spread annotations across multiple DBs and then wonder why later queries look empty.

Example path used below:

```bash
export DB=/tmp/smailnail-annotate-demo.sqlite
```

## Step 2: Add An Annotation

Start with the smallest useful operation: attach a tag and note to one target.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation add \
  --sqlite-path "$DB" \
  --target-type sender \
  --target-id notifications@github.com \
  --tag important \
  --note "Still important despite high volume" \
  --source-kind human \
  --created-by manuel
```

What this does:

- creates one row in `annotations`,
- defaults review state to `reviewed` for a human-created row,
- prints the inserted row through Glazed.

The fields matter:

- `--target-type` tells the system which entity family the id belongs to
- `--target-id` is the stable key within that family
- `--tag` is the quick label
- `--note` carries the nuance that the tag cannot express by itself

## Step 3: Query Annotations Back Out

Once rows exist, query them by target, tag, review state, or source kind.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation list \
  --sqlite-path "$DB" \
  --target-type sender \
  --target-id notifications@github.com
```

This is the primary inspection command when you want to answer:

- what annotations already exist for this sender?
- did the agent already flag this thread?
- are there unresolved rows left to review?

To list only rows that still need attention:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation list \
  --sqlite-path "$DB" \
  --review-state to_review \
  --source-kind agent
```

That query is the first usable review queue in the MVP.

## Step 4: Update Review State

The first version does not have a separate decision engine. Review state is the lightweight workflow marker that tells you whether a row still needs human attention.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation review \
  --sqlite-path "$DB" \
  --id 4cc62d7e-e52c-4f49-95b8-bdea24d6af19 \
  --review-state dismissed
```

Use cases:

- move an agent-created row from `to_review` to `reviewed`
- dismiss a suggestion without deleting the historical row

Recommended meanings:

- `to_review`: an agent or heuristic produced it and a human still needs to inspect it
- `reviewed`: a human has looked at it and wants it to remain part of the record
- `dismissed`: the row stays in history but should not drive the active queue anymore

## Step 5: Create A Group

Groups are the right tool when analysis starts as a cluster rather than a final judgment.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate group create \
  --sqlite-path "$DB" \
  --name "Possible newsletters" \
  --description "Review these senders before muting them." \
  --source-kind agent \
  --agent-run-id smoke-run-1
```

Use groups when you want to say:

- these 40 senders belong together for review
- this set of messages came from one analysis pass
- this cluster needs another pass before it turns into a stronger label

Because groups have their own `review_state`, they can also function as a coarse review queue.

## Step 6: Add Targets To A Group

After the group exists, attach one or more targets to it.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate group add-target \
  --sqlite-path "$DB" \
  --group-id cf87e5b4-2dca-4a27-a381-f946529883c2 \
  --target-type sender \
  --target-id hello@readwise.io
```

Then inspect the current membership:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate group members \
  --sqlite-path "$DB" \
  --group-id cf87e5b4-2dca-4a27-a381-f946529883c2
```

This is useful when you want a fast answer to:

- what exactly is in this review cluster?
- did the agent already place this sender into the newsletter bucket?

## Step 7: Create A Log Entry

Logs are the shared explanation layer. One log can later point at multiple targets.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate log add \
  --sqlite-path "$DB" \
  --title "April newsletter pass" \
  --body "Grouped likely newsletters based on list-unsubscribe and high sender volume." \
  --source-kind agent \
  --agent-run-id april-pass-1
```

Logs are better than copying the same note into ten annotation rows when the reasoning is really one shared observation.

## Step 8: Link A Log To Targets

Once the log exists, link it to one or more targets.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate log link-target \
  --sqlite-path "$DB" \
  --log-id 01228f9d-e717-4665-809c-23e25f011742 \
  --target-type sender \
  --target-id hello@readwise.io
```

Inspect the links with:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate log targets \
  --sqlite-path "$DB" \
  --log-id 01228f9d-e717-4665-809c-23e25f011742
```

This is the practical answer to “one log entry, many things.”

## Common Playbook Patterns

### Pattern A: Human Review Note On One Sender

Use this when a sender clearly matters and you do not want later bulk-cleanup work to lose that context.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation add \
  --sqlite-path "$DB" \
  --target-type sender \
  --target-id person@example.com \
  --tag important \
  --note "Personal contact. Never bulk archive without rereading." \
  --source-kind human \
  --created-by manuel
```

### Pattern B: Agent Suggestion Queue

Use this when an agent classifies rows but you still want a human pass afterward.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation add \
  --sqlite-path "$DB" \
  --target-type sender \
  --target-id promo@example.com \
  --tag newsletter \
  --note "Likely promotional sender; low reply expectation." \
  --source-kind agent \
  --source-label triage-agent \
  --agent-run-id batch-2026-04-02
```

Then query unresolved rows:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate annotation list \
  --sqlite-path "$DB" \
  --source-kind agent \
  --review-state to_review
```

### Pattern C: Cluster First, Judgment Later

Use groups when you are not ready to stamp the same tag onto every target yet.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail annotate group create \
  --sqlite-path "$DB" \
  --name "Possible spam cluster" \
  --description "Needs manual pass before tagging or deleting anything." \
  --source-kind agent \
  --agent-run-id batch-2026-04-02
```

This pattern is safer than overcommitting too early.

## Querying Directly In SQLite

The CLI is the intended interface, but direct SQL is useful for bulk inspection.

List unresolved agent annotations:

```sql
SELECT id, target_type, target_id, tag, note_markdown, created_at
FROM annotations
WHERE source_kind = 'agent'
  AND review_state = 'to_review'
ORDER BY created_at DESC;
```

Inspect all annotations for one sender:

```sql
SELECT *
FROM annotations
WHERE target_type = 'sender'
  AND target_id = 'notifications@github.com'
ORDER BY created_at DESC;
```

Inspect one group and its members:

```sql
SELECT g.id, g.name, m.target_type, m.target_id, m.added_at
FROM target_groups g
JOIN target_group_members m ON m.group_id = g.id
WHERE g.id = 'cf87e5b4-2dca-4a27-a381-f946529883c2'
ORDER BY m.added_at DESC;
```

Inspect one log and its linked targets:

```sql
SELECT l.id, l.title, t.target_type, t.target_id
FROM annotation_logs l
JOIN annotation_log_targets t ON t.log_id = l.id
WHERE l.id = '01228f9d-e717-4665-809c-23e25f011742'
ORDER BY t.target_type, t.target_id;
```

Direct SQL matters for two reasons:

- it makes the system easy to audit without special tooling
- it helps you discover which tags and review flows are actually being used before adding more structure

## Failure Modes And Practical Guidance

This first version is intentionally simple, which means some discipline is still manual.

- Use stable `target_id` values. If you annotate senders, always use the normalized sender email.
- Prefer groups when the classification is still fuzzy. Do not rush uncertain clusters into definitive tags.
- Keep human-created notes explicit. Free-form note quality matters more than clever schema in the MVP.
- Do not treat `dismissed` as deletion. It means “keep history, remove from active review.”
- Keep one SQLite path per mirror corpus. Mixing unrelated datasets into one DB makes review less legible.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| `sqlite-path is required` | The command did not receive a DB path and you do not want to rely on defaults | Pass `--sqlite-path /path/to/mirror.sqlite` explicitly |
| The command creates tables unexpectedly | The mirror DB was older than schema v3 | This is expected; the annotation CLI bootstraps missing annotation tables |
| Queries return no rows | You are pointing at a different SQLite file than the one you annotated | Recheck `--sqlite-path` and keep it stable across commands |
| Agent rows never show up as `to_review` | You created them as `source-kind human` or overrode the review state | Use `--source-kind agent` and omit `--review-state` unless you mean to force it |
| Group membership looks empty | The wrong group id was used | Run `smailnail annotate group list --sqlite-path "$DB"` and copy the correct `id` |
| Log links look missing | You created a log but never linked it to targets | Run `smailnail annotate log link-target ...` and then `smailnail annotate log targets ...` |

## See Also

- `smailnail help smailnail-mirror-overview` for the mirror DB model and build-tag requirement
- `smailnail help smailnail-mirror-first-sync` for creating the mirror DB in the first place
- `smailnail annotate --help` for the command tree
- `smailnail annotate annotation --help` for annotation-specific flags
