---
Title: "Screen Specifications and Widget Hierarchy (v2)"
Ticket: SMN-20260403-ANNOTATION-UI
Status: active
Topics:
    - frontend
    - annotations
    - sqlite
    - ux-design
    - react
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Domain types"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx:Existing SPA"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx:Reference QueryEditor"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/cmd/go-minitrace/cmds/serve/handlers_queries.go:File-based query management"
ExternalSources: []
Summary: "v2: ASCII wireframes + YAML widget hierarchies. File-based SQL queries, no agent API, no fragments table."
LastUpdated: 2026-04-03T13:00:00.000000000-04:00
WhatFor: "Implementable screen specs with layout, widget nesting, and component props"
WhenToUse: ""
---

# Screen Specifications v2

> **Changes from v1:** Removed saved_queries/query_fragments DB tables — queries are `.sql` files on disk. Removed fragments sidebar section. Removed agent-facing API endpoints. Added `--preset-dir` / `--query-dir` configuration. Simplified sidebar to match go-minitrace exactly.

YAML DSL convention:
- `widget:` — React component name
- `parts:` — named sub-regions (`data-part` attributes)
- `children:` — nested widgets
- `props:` — key props/callbacks
- `data:` — data shape

---

## Screen 1: App Shell

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✉ smailnail    [Accounts] [Mailbox] [Rules] [Annotations] [Query]  │
│                                                    🔍 Search...  👤 │
├──────────────────────────────────────────────────────────────────────┤
│                        <page content>                                │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AppShell
parts:
  header:
    widget: AppHeader
    children:
      - widget: Logo
      - widget: NavTabs
        props:
          tabs:
            - { label: Accounts, path: /accounts }
            - { label: Mailbox, path: /mailbox }
            - { label: Rules, path: /rules }
            - { label: Annotations, path: /annotations }
            - { label: Query, path: /query }
      - widget: GlobalSearch
      - widget: UserBadge
  content:
    widget: RouterOutlet
```

---

## Screen 2: Review Queue

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations                                                          │
│ [*Review*] [Senders] [Threads] [Messages] [Domains]                 │
│ [Groups] [Agent Runs] [Logs]                                         │
├──────────────────────────────────────────────────────────────────────┤
│ Target: [All ▾]  Tag: [All ▾]  Source: [All ▾]  Run: [___]          │
│ Since: [____]  Before: [____]                                        │
│                                                                      │
│ 247 to review · 189 agent · 58 heuristic                            │
│ Top tags: newsletter(72) bulk-sender(45) important(23) ignore(18)   │
│                                                                      │
│ [☑ Select All]  [✓ Approve]  [✗ Dismiss]  [↺ Reset]                │
│                                                                      │
│  ☐ │ Type    │ Target               │ Tag        │ Note      │ Src  │
│ ───┼─────────┼──────────────────────┼────────────┼───────────┼───── │
│  ☐ │ 🧑 sndr │ news@example.com     │ newsletter │ High vo…  │ 🤖  │
│  ☐ │ 🧑 sndr │ alerts@github.com    │ important  │ CI noti…  │ 🤖  │
│  ☐ │ 📧 msg  │ <abc123@mail.exam>   │ bulk       │ Market…   │ ⚙️  │
│  ☐ │ 🔗 thrd │ Re: Q1 Planning      │ important  │ 12-msg…   │ 🤖  │
│  …                                                                   │
│                                                                      │
│ ┌─ Expanded Detail ──────────────────────────────────────────────┐  │
│ │ Annotation abc-def-123                                          │  │
│ │ Target: sender / news@example.com                               │  │
│ │ Tag: newsletter    State: 🟡 to_review                         │  │
│ │ Source: 🤖 agent · triage-pass-1 · run-42                      │  │
│ │                                                                  │  │
│ │ Note:                                                            │  │
│ │   High volume sender (342 msgs). Has list-unsubscribe header.  │  │
│ │   Recommend muting or creating a filter rule.                   │  │
│ │                                                                  │  │
│ │ Other annotations on this target (2):                           │  │
│ │   • bulk-sender (🤖 run-41) 🟢 reviewed                       │  │
│ │   • marketing  (⚙️ heurist.) ⚫ dismissed                      │  │
│ │                                                                  │  │
│ │ [View Sender →]  [✓ Approve]  [✗ Dismiss]                      │  │
│ └──────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: ReviewQueue
parts:
  filters:
    widget: ReviewFilters
    props: { onFilterChange: "(filters) => void" }
  summary:
    widget: ReviewSummary
    data: { totalToReview: number, bySourceKind: "Record<string,number>", topTags: "{tag,count}[]" }
  batchBar:
    widget: BatchActionBar
    props: { selectedCount: number, onApprove, onDismiss, onReset, onSelectAll }
  table:
    widget: AnnotationTable
    props: { annotations: "Annotation[]", selected: "Set<string>", expandedId: "string|null" }
    children:
      - widget: AnnotationRow
      - widget: AnnotationDetail
        props: { annotation, relatedAnnotations, onNavigateTarget, onApprove, onDismiss }
```

---

## Screen 3: Senders Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Senders                                                │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter by domain, email...    Has annotations: [All ▾]           │
│ 1,247 senders · 842 annotated · 312 to review                       │
│                                                                      │
│  Email                  │ Domain       │ Msgs │ Ann │ Tags    │ Rev │
│ ────────────────────────┼──────────────┼──────┼─────┼─────────┼─────│
│  news@example.com       │ example.com  │  342 │   3 │ NL, Blk │ ██░│
│  alerts@github.com      │ github.com   │ 1205 │   2 │ Imp     │ ███│
│  no-reply@amazon.com    │ amazon.com   │   87 │   1 │ Commer  │ █░░│
│  john@colleague.org     │ colleague.or │   45 │   0 │ —       │  — │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: SendersBrowser
parts:
  filters:
    widget: SenderFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { total: number, annotated: number, toReview: number }
  table:
    widget: SenderTable
    props: { senders: "SenderRow[]", onSelectSender: "(email) => void" }
```

---

## Screen 4: Sender Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Senders    news@example.com                                        │
├──────────────────────────────────────────────────────────────────────┤
│ ┌─ Profile ───────────────────────────────────────────────────────┐ │
│ │ Email: news@example.com    Display: Example News                │ │
│ │ Domain: example.com        Messages: 342                        │ │
│ │ First seen: 2025-01-15     Last: 2026-04-02                    │ │
│ │ Private relay: No          Unsubscribe: mailto:… [HTTP link]   │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ─── Annotations (3) ─────────────────────────────────────────────── │
│  Tag         │ Note                    │ Source      │ State         │
│ ─────────────┼─────────────────────────┼─────────────┼────────────── │
│  newsletter  │ High volume sender...   │ 🤖 run-42  │ 🟡 review    │
│  bulk-sender │ Consistent list-unsub…  │ 🤖 run-41  │ 🟢 ok        │
│  marketing   │ Promotional content     │ ⚙️ heurist.│ ⚫ dismiss    │
│                                                                      │
│ ─── Groups (1) ──────────────────────────────────────────────────── │
│  • Possible newsletters (12 members) → [View group]                 │
│                                                                      │
│ ─── Recent Messages ─────────────────────────────────────────────── │
│  Date       │ Subject                           │ Size  │ Flags     │
│ ────────────┼───────────────────────────────────┼───────┼────────── │
│  2026-04-02 │ Your weekly digest                │ 45KB  │ Seen      │
│  2026-03-26 │ Your weekly digest                │ 42KB  │ Seen      │
│  [Load more...]                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: SenderDetail
props: { email: string, onBack: "() => void" }
parts:
  profile:
    widget: SenderProfileCard
    data: { sender: SenderRecord }
  annotations:
    widget: TargetAnnotationList
    props: { targetType: sender, targetId: "$email", onReviewChange }
  groups:
    widget: TargetGroupMemberships
    props: { targetType: sender, targetId: "$email", onNavigateGroup }
  messages:
    widget: SenderMessageList
    props: { senderEmail: "$email", limit: 10, onSelectMessage, onLoadMore }
```

---

## Screen 5: Groups List

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Groups                                                 │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter...   Review: [All ▾]   Source: [All ▾]                    │
│ 18 groups · 12 to review                                             │
│                                                                      │
│  Name                   │ Members │ Source     │ Review   │ Created  │
│ ────────────────────────┼─────────┼───────────┼──────────┼───────── │
│  Possible newsletters   │      12 │ 🤖 run-42 │ 🟡       │ Apr 2   │
│  CI notification noise  │       5 │ 🤖 run-42 │ 🟡       │ Apr 2   │
│  VIP senders            │       8 │ 👤 manuel │ 🟢       │ Mar 28  │
│  Bulk commercial        │      23 │ 🤖 run-41 │ 🟢       │ Mar 25  │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: GroupsBrowser
parts:
  filters: { widget: GroupFilters }
  summary: { widget: CountSummary, data: { total, toReview } }
  table:
    widget: GroupTable
    props: { groups: "GroupRow[]", onSelectGroup: "(id) => void" }
```

---

## Screen 6: Group Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Groups    Possible newsletters                      🟡 to_review  │
├──────────────────────────────────────────────────────────────────────┤
│ Source: 🤖 agent · triage-pass-1 · run-42    Created: Apr 2 14:23  │
│                                                                      │
│ Description:                                                         │
│ ┌──────────────────────────────────────────────────────────────────┐│
│ │ Senders that exhibit newsletter-like patterns: high volume,      ││
│ │ consistent list-unsubscribe headers, and bulk-send cadence.      ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ [✓ Approve Group]  [✗ Dismiss Group]                                │
│                                                                      │
│ ─── Members (12) ───────────────────────────────────────────────── │
│  Type   │ ID                       │ Added     │ Actions            │
│ ────────┼──────────────────────────┼───────────┼────────────────── │
│  sender │ news@example.com         │ Apr 2     │ [View] [Remove]   │
│  sender │ digest@weekly.io         │ Apr 2     │ [View] [Remove]   │
│  sender │ updates@saas-tool.com    │ Apr 2     │ [View] [Remove]   │
│                                                                      │
│ ─── Related Logs ───────────────────────────────────────────────── │
│  Apr 2 14:23 │ Initial review pass │ 🤖 run-42                     │
│    Grouped likely newsletters based on list-unsubscribe and high…  │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: GroupDetail
props: { groupId: string, onBack }
parts:
  header: { widget: GroupHeader, data: { group: TargetGroup }, props: { onApprove, onDismiss } }
  description: { widget: MarkdownRenderer, data: { markdown: "group.description" } }
  members:
    widget: GroupMemberList
    props: { members: "GroupMember[]", onNavigateTarget, onRemoveMember }
  logs:
    widget: RelatedLogList
    props: { logs: "AnnotationLog[]", onSelectLog }
```

---

## Screen 7: Agent Runs

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Agent Runs                                             │
├──────────────────────────────────────────────────────────────────────┤
│ 8 agent runs · 412 annotations · 247 to review                      │
│                                                                      │
│  Run ID   │ Source Label    │ Ann │ Logs │ First      │ Rev Progress │
│ ──────────┼────────────────┼─────┼──────┼────────────┼───────────── │
│  run-42   │ triage-pass-1  │ 189 │    3 │ Apr 2 14:2 │ ██████░░░░  │
│  run-41   │ triage-pass-1  │ 145 │    2 │ Mar 25 09: │ ██████████  │
│  run-40   │ sender-scan    │  58 │    1 │ Mar 20 11: │ ████████░░  │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AgentRunsBrowser
parts:
  summary: { widget: CountSummary, data: { runs, annotations, toReview } }
  table:
    widget: AgentRunTable
    props: { runs: "AgentRunRow[]", onSelectRun }
```

---

## Screen 8: Agent Run Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Agent Runs    run-42                                               │
├──────────────────────────────────────────────────────────────────────┤
│ Source: triage-pass-1   Ann: 189   Logs: 3                           │
│ Period: Apr 2 14:23 → 14:31    Progress: ████░░░░░░ 42%             │
│                                                                      │
│ [✓ Approve All Remaining]  [✗ Dismiss All Remaining]                │
│                                                                      │
│ By Tag: newsletter(72) bulk-sender(45) important(23) other(49)      │
│ By Type: sender(134) thread(32) message(18) domain(5)               │
│                                                                      │
│ ─── Timeline ───────────────────────────────────────────────────── │
│  14:23:01 │ 🤖 Created group "Possible newsletters"                │
│  14:23:01 │ 📝 Log: "Initial review pass"                          │
│  14:23:02 │ 🏷  sender/news@example.com → newsletter               │
│  14:23:02 │ 🏷  sender/digest@weekly.io → newsletter               │
│  14:23:03 │ 🏷  sender/alerts@github.com → important               │
│  …                                                                   │
│  14:31:00 │ 📝 Log: "Summary: 189 annotations across 4 types"      │
│                                                                      │
│ ─── Logs ───────────────────────────────────────────────────────── │
│  14:23 │ Initial review pass                                        │
│    Grouped likely newsletters based on list-unsubscribe...          │
│  14:31 │ Summary: 189 annotations across 4 types                   │
│    Annotated 134 senders, 32 threads, 18 messages, 5 domains...    │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AgentRunDetail
props: { runId: string, onBack }
parts:
  header:
    widget: RunHeader
    props: { onApproveAll, onDismissAll }
    data: { run: AgentRunSummary }
  breakdowns:
    children:
      - widget: TagBreakdownBar
      - widget: TargetTypeBreakdownBar
  timeline:
    widget: RunTimeline
    props: { events: "RunEvent[]", onNavigateTarget, onNavigateLog }
  logs:
    widget: RunLogList
    props: { logs: "AnnotationLog[]", onSelectLog }
```

---

## Screen 9: Logs List

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Logs                                                   │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter...   Source: [All ▾]   Run: [____]                        │
│ 23 log entries                                                       │
│                                                                      │
│  Time          │ Kind │ Title                    │ Src    │ Targets  │
│ ───────────────┼──────┼──────────────────────────┼────────┼───────── │
│  Apr 2 14:31   │ note │ Summary: 189 annotations │ 🤖     │ 5       │
│  Apr 2 14:27   │ note │ Sender normalization not │ 🤖     │ 12      │
│  Apr 2 14:23   │ note │ Initial review pass      │ 🤖     │ 8       │
│  Mar 25 09:15  │ note │ Bulk triage complete     │ 🤖     │ 15      │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: LogsBrowser
parts:
  filters: { widget: LogFilters }
  summary: { widget: CountSummary }
  table: { widget: LogTable, props: { logs: "AnnotationLog[]", onSelectLog } }
```

---

## Screen 10: Log Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Logs    Initial review pass                                        │
├──────────────────────────────────────────────────────────────────────┤
│ Kind: note   Source: 🤖 agent · triage-pass-1 · run-42              │
│ Created: Apr 2 14:23   By: triage-agent                              │
│                                                                      │
│ ┌──────────────────────────────────────────────────────────────────┐│
│ │ Grouped likely newsletters based on list-unsubscribe and high    ││
│ │ volume (>50 messages in the last 90 days).                       ││
│ │                                                                   ││
│ │ ## Criteria used                                                  ││
│ │ - `has_list_unsubscribe = true`                                   ││
│ │ - `msg_count > 50`                                                ││
│ │                                                                   ││
│ │ ## Recommendations                                                ││
│ │ Review each sender in the group. Some may be transactional       ││
│ │ rather than true newsletters.                                     ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ ─── Linked Targets (8) ────────────────────────────────────────── │
│  sender │ news@example.com      → [View]                            │
│  sender │ digest@weekly.io      → [View]                            │
│  sender │ updates@saas-tool.com → [View]                            │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: LogDetail
props: { logId: string, onBack }
parts:
  header: { widget: LogHeader, data: { log: AnnotationLog } }
  body: { widget: MarkdownRenderer, data: { markdown: "log.bodyMarkdown" } }
  targets: { widget: LinkedTargetList, props: { targets: "LogTarget[]", onNavigateTarget } }
```

---

## Screen 11: Query Editor

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✉ smailnail  [Accounts] [Mailbox] [Rules] [Annotations] [*Query*]  │
├──────────┬───────────────────────────────────────────────────────────┤
│ PRESETS  │ ┌─────────────────────────────────────────────────────┐  │
│          │ │ SELECT s.email, s.msg_count,                        │  │
│ 📁 annot │ │   COUNT(a.id) AS annotation_count                  │  │
│   📄 rev │ │ FROM senders s                                      │  │
│   📄 by- │ │ LEFT JOIN annotations a                             │  │
│ 📁 sende │ │   ON a.target_type = 'sender'                      │  │
│   📄 top │ │   AND a.target_id = s.email                        │  │
│   📄 NL  │ │ WHERE s.msg_count > 50                              │  │
│ 📁 threa │ │ GROUP BY s.email                                    │  │
│   📄 lon │ │ ORDER BY annotation_count DESC                      │  │
│ 📁 msgs  │ │ LIMIT 50;                                           │  │
│   📄 siz │ └─────────────────────────────────────────────────────┘  │
│          │ [▶ Run]  [💾 Save]  Ctrl+Enter to run                    │
│ SAVED    │                                                           │
│ 📁 my-qu │ ─────────────────────────────────────────────────────── │
│   ⭐ NL  │                                                           │
│   ⭐ CI  │  50 rows · 3 cols · 12ms          [⬇ CSV] [⬇ JSON]     │
│          │                                                           │
│          │  # │ email               │ msg_count │ annotation_count  │
│          │ ───┼─────────────────────┼───────────┼────────────────── │
│          │  1 │ news@example.com    │       342 │                3  │
│          │  2 │ alerts@github.com   │      1205 │                2  │
│          │  3 │ no-reply@amazon.com │        87 │                1  │
│          │  … │                     │           │                   │
│          │                                                           │
│          │ ⚠ Source: my-queries/NL-senders.sql (file changed) [↻]  │
└──────────┴───────────────────────────────────────────────────────────┘
```

The sidebar shows two sections: **Presets** (read-only, from `go:embed` + `--preset-dir`) and **Saved** (read-write, from `--query-dir`). No fragments section — reusable SQL snippets are just saved queries in a well-named folder.

Each `.sql` file's first `-- ` comment line becomes the sidebar tooltip description.

```yaml
widget: QueryEditorPage
children:
  - widget: QueryEditor
    parts:
      sidebar:
        widget: QuerySidebar
        props:
          presets: "SavedQuery[]"
          savedQueries: "SavedQuery[]"
          onSelect: "(q, kind) => void"
        children:
          - widget: QuerySection
            props: { title: Presets }
            children:
              - widget: QueryFolderGroup
                children:
                  - widget: QueryItem
                    props: { icon: "📄", name, tooltip, readonly: true }
          - widget: QuerySection
            props: { title: Saved }
            children:
              - widget: QueryFolderGroup
                children:
                  - widget: QueryItem
                    props: { icon: "⭐", name, tooltip, readonly: false }
      editor:
        widget: SqlEditorPane
        children:
          - widget: SourceStatusBanner
            props:
              label: string
              path: string
              missing: boolean
              externalUpdateAvailable: boolean
              onReload: "() => void"
          - widget: SqlEditor
            props: { value: string, onChange, onExecute }
          - widget: EditorToolbar
            props: { onRun, onSave, isLoading }
      results:
        widget: ResultsPane
        children:
          - widget: ResultsHeader
            props: { rowCount, colCount, durationMs, onExportCsv, onExportJson }
          - widget: ResultsTable
            props: { result: QueryResult, onClickId }
```

---

## Screen 12: Messages Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Messages                                               │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Full-text search...   Mailbox: [All ▾]  Has ann: [All ▾]        │
│ 48,231 messages · 218 annotated                                      │
│                                                                      │
│  Date       │ From              │ Subject              │ Ann│ Tags   │
│ ────────────┼───────────────────┼──────────────────────┼────┼─────── │
│  2026-04-02 │ news@example.com  │ Your weekly digest   │  1 │ NL     │
│  2026-04-02 │ alerts@github.com │ [go-minitrace] PR #4 │  0 │ —      │
│  2026-04-01 │ john@colleague    │ Re: Q1 Planning      │  2 │ Imp    │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: MessagesBrowser
parts:
  filters: { widget: MessageFilters, props: { onSearch, onFilterChange } }
  summary: { widget: CountSummary, data: { total, annotated } }
  table: { widget: MessageTable, props: { messages: "MessageRow[]", onSelectMessage } }
```

---

## Screen 13: Message Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Messages    Re: Q1 Planning                                       │
├──────────────────────────────────────────────────────────────────────┤
│ ┌─ Headers ───────────────────────────────────────────────────────┐ │
│ │ From: john@colleague.org         Date: 2026-04-01 15:30        │ │
│ │ To: manuel@example.com           Mailbox: INBOX                │ │
│ │ Cc: team@colleague.org           Thread: (5 msgs) [View →]    │ │
│ │ Message-ID: <abc123@colleague>   Size: 12KB                    │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ┌─ Body ──────────────────────────────────────────────────────────┐ │
│ │ Hi Manuel,                                                       │ │
│ │                                                                   │ │
│ │ I've updated the Q1 planning doc with the latest numbers.        │ │
│ │ Please review before the meeting on Friday.                      │ │
│ └──────────────────────────────────────────────────────────────────┘ │
│ [Text ▾] [HTML] [Raw]                                                │
│                                                                      │
│ ─── Annotations (2) ────────────────────────────────────────────── │
│  important  │ Key planning thread    │ 🤖 run-42 │ 🟡               │
│  action-req │ Contains review request│ ⚙️ heurist│ 🟢               │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: MessageDetail
props: { messageId: number, onBack }
parts:
  headers: { widget: MessageHeaders, props: { onNavigateThread, onNavigateSender } }
  body: { widget: MessageBody, props: { viewMode: "text|html|raw", onChangeViewMode } }
  annotations: { widget: TargetAnnotationList, props: { targetType: message, targetId } }
```

---

## Screen 14: Threads Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Threads                                                │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter...   Account: [All ▾]                                     │
│ 8,421 threads · 32 annotated                                        │
│                                                                      │
│  Subject              │ Msgs │ Participants │ Date Range    │ Ann   │
│ ──────────────────────┼──────┼──────────────┼───────────────┼────── │
│  Re: Q1 Planning      │   12 │            5 │ Mar 1 - Apr 2 │    3 │
│  Weekly standup notes  │    8 │            3 │ Mar 15 - Apr 1│    0 │
│  [go-minitrace] PR #4 │    4 │            2 │ Mar 28 - Mar 3│    1 │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: ThreadsBrowser
parts:
  filters: { widget: ThreadFilters }
  summary: { widget: CountSummary, data: { total, annotated } }
  table: { widget: ThreadTable, props: { threads: "ThreadRow[]", onSelectThread } }
```

---

## Screen 15: Domains Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Domains                                                │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter...                                                        │
│ 342 domains · 89 annotated senders                                   │
│                                                                      │
│  Domain            │ Senders │ Messages │ Ann. Senders │ Top Tag     │
│ ───────────────────┼─────────┼──────────┼──────────────┼──────────── │
│  github.com        │       8 │     3420 │            3 │ important   │
│  example.com       │      12 │      890 │            5 │ newsletter  │
│  amazon.com        │       4 │      340 │            2 │ commercial  │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: DomainsBrowser
parts:
  filters: { widget: DomainFilters }
  summary: { widget: CountSummary, data: { total, annotatedSenders } }
  table: { widget: DomainTable, props: { domains: "DomainRow[]", onSelectDomain } }
```

---

## Shared Widgets

```yaml
widget: SourceBadge
props: { sourceKind: "agent|human|heuristic|import", sourceLabel: "string?" }

widget: ReviewStateChip
props: { state: "to_review|reviewed|dismissed", onClick: "()=>void?" }

widget: ReviewProgressBar
props: { reviewed: number, toReview: number, dismissed: number }

widget: TargetAnnotationList  # used in SenderDetail, MessageDetail, ThreadDetail
props: { targetType, targetId, annotations: "Annotation[]", onReviewChange }

widget: CountSummary  # used in every browser
props: { items: "{label,count}[]" }

widget: BatchActionBar  # used in ReviewQueue, AgentRunDetail
props: { selectedCount, onSelectAll, onApprove, onDismiss, onReset }

widget: MarkdownRenderer
props: { markdown: string, maxHeight: "number?" }

widget: LinkedTargetList  # used in LogDetail, GroupDetail
props: { targets: "{type,id}[]", onNavigateTarget }
```

---

## Routing

```yaml
routes:
  /accounts: AccountSetupPage
  /mailbox: MailboxExplorer
  /rules: RulesPage
  /annotations: ReviewQueue
  /annotations/senders: SendersBrowser
  /annotations/senders/:email: SenderDetail
  /annotations/threads: ThreadsBrowser
  /annotations/threads/:id: ThreadDetail
  /annotations/messages: MessagesBrowser
  /annotations/messages/:id: MessageDetail
  /annotations/domains: DomainsBrowser
  /annotations/groups: GroupsBrowser
  /annotations/groups/:id: GroupDetail
  /annotations/runs: AgentRunsBrowser
  /annotations/runs/:id: AgentRunDetail
  /annotations/logs: LogsBrowser
  /annotations/logs/:id: LogDetail
  /query: QueryEditorPage
```

---

## File Structure

```
ui/src/features/annotations/
├── AnnotationsPage.tsx
├── ReviewQueue.tsx
├── ReviewFilters.tsx
├── AnnotationTable.tsx
├── AnnotationDetail.tsx
├── SendersBrowser.tsx
├── SenderDetail.tsx
├── ThreadsBrowser.tsx
├── MessagesBrowser.tsx
├── MessageDetail.tsx
├── DomainsBrowser.tsx
├── GroupsBrowser.tsx
├── GroupDetail.tsx
├── AgentRunsBrowser.tsx
├── AgentRunDetail.tsx
├── LogsBrowser.tsx
├── LogDetail.tsx
└── shared/
    ├── SourceBadge.tsx
    ├── ReviewStateChip.tsx
    ├── ReviewProgressBar.tsx
    ├── TargetAnnotationList.tsx
    ├── BatchActionBar.tsx
    ├── LinkedTargetList.tsx
    └── CountSummary.tsx

ui/src/features/query/
├── QueryEditorPage.tsx
├── QueryEditor.tsx
├── QuerySidebar.tsx
├── SqlEditor.tsx
├── ResultsTable.tsx
└── SourceStatusBanner.tsx

pkg/query/
├── presets/
│   ├── annotations/
│   │   ├── review-queue-counts.sql
│   │   ├── by-tag.sql
│   │   └── by-source.sql
│   ├── senders/
│   │   ├── top-senders.sql
│   │   └── newsletter-candidates.sql
│   ├── threads/
│   │   └── longest-threads.sql
│   └── messages/
│       ├── size-distribution.sql
│       └── fts-search.sql
├── assets.go       # go:embed presets/*.sql
└── engine.go       # LoadSQLDirs, ResolveSQL, RunQuery
```
