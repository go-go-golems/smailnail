---
Title: Screen Specifications and Widget Hierarchy
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
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/types.go:Annotation, TargetGroup, AnnotationLog domain types"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/annotate/schema.go:Schema V3 tables"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/ui/src/App.tsx:Existing React SPA to extend"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QueryEditor.tsx:QueryEditor reference widget"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/SessionBrowser/SessionBrowser.tsx:SessionBrowser reference widget"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/QuerySidebar.tsx:QuerySidebar reference widget"
    - "/home/manuel/code/wesen/corporate-headquarters/go-minitrace/web/src/components/QueryEditor/ResultsTable.tsx:ResultsTable reference widget"
ExternalSources: []
Summary: "Concrete screen wireframes (ASCII) and YAML widget hierarchies for every screen in the annotation review and query editor UI."
LastUpdated: 2026-04-03T12:00:00.000000000-04:00
WhatFor: "Provide implementable screen specifications with exact layout, widget nesting, and component props"
WhenToUse: ""
---

# Screen Specifications and Widget Hierarchy

This document specifies every screen in the annotation review and query editor UI. Each screen has:
1. An ASCII wireframe showing visual layout
2. A YAML DSL describing the widget hierarchy

The YAML DSL convention:
- `widget:` names a React component (PascalCase)
- `parts:` lists named sub-regions (maps to `data-part` attributes)
- `children:` lists nested widgets
- `props:` lists key props/callbacks
- `data:` describes the data shape flowing into the widget

---

## Screen 1: App Shell (Extended)

The existing app shell is extended with two new top-level nav items.

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✉ smailnail    [Accounts] [Mailbox] [Rules] [Annotations] [Query]  │
│                                                    🔍 Search...  👤 │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│                        <page content>                                │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AppShell
parts:
  header:
    widget: AppHeader
    children:
      - widget: Logo
        props: { onClick: "navigate(/)" }
      - widget: NavTabs
        props:
          tabs:
            - { label: Accounts, path: /accounts }
            - { label: Mailbox, path: /mailbox }
            - { label: Rules, path: /rules }
            - { label: Annotations, path: /annotations }
            - { label: Query, path: /query }
      - widget: GlobalSearch
        props: { onSearch: "navigate(/search?q=...)" }
      - widget: UserBadge
  content:
    widget: RouterOutlet
```

---

## Screen 2: Review Queue (Annotations Landing)

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations                                                          │
│ [Review Queue] [Senders] [Threads] [Messages] [Domains]             │
│ [Groups] [Agent Runs] [Logs]                                         │
├──────────────────────────────────────────────────────────────────────┤
│ ┌─ Filters ───────────────────────────────────────────────────────┐ │
│ │ Target: [All ▾]  Tag: [All ▾]  Source: [All ▾]  Run: [___]     │ │
│ │ Since: [____]  Before: [____]                                    │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ 247 to review · 189 agent · 58 heuristic                            │
│ Top tags: newsletter(72) bulk-sender(45) important(23) ignore(18)   │
│                                                                      │
│ ┌─ Batch ─────────────────────────────────────────┐                 │
│ │ [☑ Select All]  [✓ Approve]  [✗ Dismiss]  [↺ Reset]             │ │
│ └─────────────────────────────────────────────────┘                 │
│                                                                      │
│  ☐ │ Type    │ Target ID                │ Tag        │ Note    │ Src│
│ ───┼─────────┼──────────────────────────┼────────────┼─────────┼────│
│  ☐ │ 🧑sender│ news@example.com         │ newsletter │ High vo…│ 🤖│
│  ☐ │ 🧑sender│ alerts@github.com        │ important  │ CI noti…│ 🤖│
│  ☐ │ 📧msg   │ <abc123@mail.example>    │ bulk       │ Marketi…│ ⚙️│
│  ☐ │ 🔗thread│ Re: Q1 Planning          │ important  │ 12-msg …│ 🤖│
│  …  │         │                          │            │         │    │
│                                                                      │
│ ┌─ Expanded Detail (when row clicked) ────────────────────────────┐ │
│ │ Annotation: abc-def-123                                          │ │
│ │ Target: sender / news@example.com                                │ │
│ │ Tag: newsletter    Review: 🟡 to_review                         │ │
│ │ Source: 🤖 agent · triage-pass-1 · run-42                       │ │
│ │ Note:                                                             │ │
│ │   High volume sender (342 msgs). Has list-unsubscribe header.   │ │
│ │   Recommend muting or creating a filter rule.                    │ │
│ │                                                                   │ │
│ │ Other annotations on this target (2):                            │ │
│ │   • bulk-sender (agent, run-41)  🟢 reviewed                    │ │
│ │   • marketing (heuristic)        ⚫ dismissed                    │ │
│ │                                                                   │ │
│ │ [View Sender Profile →]  [✓ Approve]  [✗ Dismiss]               │ │
│ └──────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AnnotationsPage
parts:
  tabs:
    widget: AnnotationTabs
    props:
      tabs:
        - { label: Review Queue, path: /annotations }
        - { label: Senders, path: /annotations/senders }
        - { label: Threads, path: /annotations/threads }
        - { label: Messages, path: /annotations/messages }
        - { label: Domains, path: /annotations/domains }
        - { label: Groups, path: /annotations/groups }
        - { label: Agent Runs, path: /annotations/runs }
        - { label: Logs, path: /annotations/logs }
  content:
    widget: ReviewQueue
    parts:
      filters:
        widget: ReviewFilters
        props:
          onFilterChange: "(filters) => void"
        data:
          targetTypes: ["sender","thread","message","domain","mailbox"]
          sourceKinds: ["agent","human","heuristic","import"]
          tags: "string[]  # populated from distinct tag values"
      summary:
        widget: ReviewSummary
        data:
          totalToReview: number
          bySourceKind: "Record<string, number>"
          topTags: "Array<{tag: string, count: number}>"
      batchBar:
        widget: BatchActionBar
        props:
          selectedCount: number
          onSelectAll: "() => void"
          onApprove: "() => void"
          onDismiss: "() => void"
          onReset: "() => void"
      table:
        widget: AnnotationTable
        props:
          annotations: "Annotation[]"
          selected: "Set<string>"
          onToggleSelect: "(id: string) => void"
          onRowClick: "(id: string) => void"
          expandedId: "string | null"
        children:
          - widget: AnnotationRow
            props:
              annotation: Annotation
              isSelected: boolean
              isExpanded: boolean
          - widget: AnnotationDetail
            props:
              annotation: Annotation
              relatedAnnotations: "Annotation[]"
              onNavigateTarget: "(type, id) => void"
              onApprove: "() => void"
              onDismiss: "() => void"
```

---

## Screen 3: Senders Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Senders                                                │
│ [Review Queue] [*Senders*] [Threads] [Messages] [Domains] ...       │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter by domain, email...    Has annotations: [All ▾]           │
│                                                                      │
│ 1,247 senders · 842 annotated · 312 to review                       │
│                                                                      │
│  Email                  │ Domain       │ Msgs │ Ann │ Tags    │ Rev │
│ ────────────────────────┼──────────────┼──────┼─────┼─────────┼─────│
│  news@example.com       │ example.com  │  342 │   3 │ NL, Blk │ ██░│
│  alerts@github.com      │ github.com   │ 1205 │   2 │ Imp     │ ███│
│  no-reply@amazon.com    │ amazon.com   │   87 │   1 │ Commer  │ █░░│
│  john@colleague.org     │ colleague.or │   45 │   0 │ —       │  — │
│  …                      │              │      │     │         │     │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: SendersBrowser
parts:
  filters:
    widget: SenderFilters
    props:
      onFilterChange: "(filters) => void"
    data:
      domains: "string[]  # top domains"
  summary:
    widget: CountSummary
    data: { total: number, annotated: number, toReview: number }
  table:
    widget: SenderTable
    props:
      senders: "SenderRow[]"
      onSelectSender: "(email: string) => void"
    data:
      SenderRow:
        email: string
        displayName: string
        domain: string
        msgCount: number
        annotationCount: number
        topTags: "string[]"
        reviewProgress: "{reviewed: n, toReview: n, dismissed: n}"
```

---

## Screen 4: Sender Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Senders    news@example.com                                        │
├──────────────────────────────────────────────────────────────────────┤
│ ┌─ Profile ───────────────────────────────────────────────────────┐ │
│ │ Email: news@example.com                                          │ │
│ │ Display: Example News   Domain: example.com                      │ │
│ │ Messages: 342    First seen: 2025-01-15   Last: 2026-04-02      │ │
│ │ Private relay: No    Unsubscribe: mailto:unsub@ex... [HTTP link]│ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ─── Annotations (3) ────────────────────────────────────────────── │
│                                                                      │
│  Tag         │ Note                     │ Source       │ State       │
│ ─────────────┼──────────────────────────┼──────────────┼─────────── │
│  newsletter  │ High volume sender...    │ 🤖 run-42   │ 🟡 review  │
│  bulk-sender │ Consistent list-unsub... │ 🤖 run-41   │ 🟢 ok      │
│  marketing   │ Promotional content      │ ⚙️ heurist. │ ⚫ dismiss  │
│                                                                      │
│ ─── Groups (1) ─────────────────────────────────────────────────── │
│  • Possible newsletters (12 members) → [View group]                 │
│                                                                      │
│ ─── Recent Messages ────────────────────────────────────────────── │
│  Date       │ Subject                           │ Size  │ Flags     │
│ ────────────┼───────────────────────────────────┼───────┼────────── │
│  2026-04-02 │ Your weekly digest                │ 45KB  │ Seen      │
│  2026-03-26 │ Your weekly digest                │ 42KB  │ Seen      │
│  2026-03-19 │ Breaking: Product update          │ 18KB  │ Flagged   │
│  [Load more...]                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: SenderDetail
props:
  email: string
  onBack: "() => void"
parts:
  profile:
    widget: SenderProfileCard
    data:
      sender: "SenderRecord"  # from senders table
  annotations:
    widget: TargetAnnotationList
    props:
      targetType: sender
      targetId: "$email"
      onReviewChange: "(id, state) => void"
  groups:
    widget: TargetGroupMemberships
    props:
      targetType: sender
      targetId: "$email"
      onNavigateGroup: "(groupId) => void"
  messages:
    widget: SenderMessageList
    props:
      senderEmail: "$email"
      limit: 10
      onSelectMessage: "(msgId) => void"
      onLoadMore: "() => void"
```

---

## Screen 5: Groups List

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Groups                                                 │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter groups...   Review: [All ▾]   Source: [All ▾]             │
│                                                                      │
│ 18 groups · 12 to review                                             │
│                                                                      │
│  Name                   │ Members │ Source     │ Review   │ Created  │
│ ────────────────────────┼─────────┼───────────┼──────────┼───────── │
│  Possible newsletters   │      12 │ 🤖 run-42 │ 🟡       │ Apr 2   │
│  CI notification noise  │       5 │ 🤖 run-42 │ 🟡       │ Apr 2   │
│  VIP senders            │       8 │ 👤 manuel │ 🟢       │ Mar 28  │
│  Bulk commercial        │      23 │ 🤖 run-41 │ 🟢       │ Mar 25  │
│  …                      │         │           │          │          │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: GroupsBrowser
parts:
  filters:
    widget: GroupFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { total: number, toReview: number }
  table:
    widget: GroupTable
    props:
      groups: "GroupRow[]"
      onSelectGroup: "(id) => void"
    data:
      GroupRow:
        id: string
        name: string
        description: string
        memberCount: number
        sourceKind: string
        sourceLabel: string
        reviewState: string
        createdAt: string
```

---

## Screen 6: Group Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Groups    Possible newsletters                      🟡 to_review  │
├──────────────────────────────────────────────────────────────────────┤
│ Source: 🤖 agent · triage-pass-1 · run-42                           │
│ Created: 2026-04-02 14:23                                            │
│                                                                      │
│ Description:                                                         │
│ ┌──────────────────────────────────────────────────────────────────┐│
│ │ Senders that exhibit newsletter-like patterns: high volume,      ││
│ │ consistent list-unsubscribe headers, and bulk-send cadence.      ││
│ │ Review before creating mute rules.                               ││
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
│  …      │                          │           │                    │
│                                                                      │
│ ─── Related Logs ───────────────────────────────────────────────── │
│  2026-04-02 │ Initial review pass │ 🤖 run-42                      │
│    Grouped likely newsletters based on list-unsubscribe and         │
│    high volume. See annotation log for detailed reasoning.          │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: GroupDetail
props:
  groupId: string
  onBack: "() => void"
parts:
  header:
    widget: GroupHeader
    data:
      group: TargetGroup
    props:
      onApprove: "() => void"
      onDismiss: "() => void"
  description:
    widget: MarkdownRenderer
    data: { markdown: "group.description" }
  members:
    widget: GroupMemberList
    props:
      members: "GroupMember[]"
      onNavigateTarget: "(type, id) => void"
      onRemoveMember: "(type, id) => void"
  logs:
    widget: RelatedLogList
    props:
      logs: "AnnotationLog[]"
      onSelectLog: "(logId) => void"
```

---

## Screen 7: Agent Runs

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Agent Runs                                             │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter...                                                        │
│                                                                      │
│ 8 agent runs · 412 annotations · 247 to review                      │
│                                                                      │
│  Run ID   │ Source Label    │ Ann │ Logs │ First      │ Rev Progress │
│ ──────────┼────────────────┼─────┼──────┼────────────┼───────────── │
│  run-42   │ triage-pass-1  │ 189 │    3 │ Apr 2 14:2 │ ██████░░░░  │
│  run-41   │ triage-pass-1  │ 145 │    2 │ Mar 25 09: │ ██████████  │
│  run-40   │ sender-scan    │  58 │    1 │ Mar 20 11: │ ████████░░  │
│  run-39   │ thread-review  │  20 │    1 │ Mar 18 16: │ ██████████  │
│  …        │                │     │      │            │              │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AgentRunsBrowser
parts:
  filters:
    widget: RunFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { runs: number, annotations: number, toReview: number }
  table:
    widget: AgentRunTable
    props:
      runs: "AgentRunRow[]"
      onSelectRun: "(runId) => void"
    data:
      AgentRunRow:
        runId: string
        sourceLabel: string
        annotationCount: number
        logCount: number
        firstTimestamp: string
        lastTimestamp: string
        reviewProgress: "{reviewed: n, toReview: n, dismissed: n}"
```

---

## Screen 8: Agent Run Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Agent Runs    run-42                                               │
├──────────────────────────────────────────────────────────────────────┤
│ Source: triage-pass-1   Annotations: 189   Logs: 3                   │
│ Period: 2026-04-02 14:23 → 14:31                                     │
│ Progress: ████░░░░░░ 42% reviewed                                    │
│                                                                      │
│ [✓ Approve All Remaining]  [✗ Dismiss All Remaining]                │
│                                                                      │
│ ─── By Tag ─────────────────────────────────────────────────────── │
│  newsletter: 72  │  bulk-sender: 45  │  important: 23  │  other: 49│
│                                                                      │
│ ─── By Target Type ─────────────────────────────────────────────── │
│  sender: 134  │  thread: 32  │  message: 18  │  domain: 5          │
│                                                                      │
│ ─── Timeline ───────────────────────────────────────────────────── │
│  14:23:01 │ 🤖 Created group "Possible newsletters"                │
│  14:23:01 │ 📝 Log: "Initial review pass"                          │
│  14:23:02 │ 🏷  sender/news@example.com → newsletter               │
│  14:23:02 │ 🏷  sender/digest@weekly.io → newsletter               │
│  14:23:03 │ 🏷  sender/alerts@github.com → important               │
│  …         │                                                         │
│  14:31:00 │ 📝 Log: "Summary: 189 annotations across 4 types"      │
│                                                                      │
│ ─── Logs ───────────────────────────────────────────────────────── │
│  14:23 │ Initial review pass                                        │
│    Grouped likely newsletters based on list-unsubscribe...          │
│  14:27 │ Sender normalization notes                                 │
│    Merged 12 private relay addresses into 8 canonical senders...    │
│  14:31 │ Summary: 189 annotations across 4 types                   │
│    Annotated 134 senders, 32 threads, 18 messages, 5 domains...    │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: AgentRunDetail
props:
  runId: string
  onBack: "() => void"
parts:
  header:
    widget: RunHeader
    data:
      run: AgentRunSummary
    props:
      onApproveAll: "() => void"
      onDismissAll: "() => void"
  breakdowns:
    children:
      - widget: TagBreakdownBar
        data: "Array<{tag: string, count: number}>"
      - widget: TargetTypeBreakdownBar
        data: "Array<{type: string, count: number}>"
  timeline:
    widget: RunTimeline
    props:
      events: "RunEvent[]"
      onNavigateTarget: "(type, id) => void"
      onNavigateLog: "(logId) => void"
    data:
      RunEvent:
        timestamp: string
        kind: "annotation | log | group"
        targetType: string
        targetId: string
        tag: string
        description: string
  logs:
    widget: RunLogList
    props:
      logs: "AnnotationLog[]"
      onSelectLog: "(logId) => void"
```

---

## Screen 9: Annotation Logs

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Logs                                                   │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter logs...   Source: [All ▾]   Run: [____]                   │
│                                                                      │
│ 23 log entries                                                       │
│                                                                      │
│  Time          │ Kind │ Title                    │ Src    │ Targets  │
│ ───────────────┼──────┼──────────────────────────┼────────┼───────── │
│  Apr 2 14:31   │ note │ Summary: 189 annotations │ 🤖     │ 5       │
│  Apr 2 14:27   │ note │ Sender normalization not │ 🤖     │ 12      │
│  Apr 2 14:23   │ note │ Initial review pass      │ 🤖     │ 8       │
│  Mar 25 09:15  │ note │ Bulk triage complete     │ 🤖     │ 15      │
│  …             │      │                          │        │          │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: LogsBrowser
parts:
  filters:
    widget: LogFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { total: number }
  table:
    widget: LogTable
    props:
      logs: "AnnotationLog[]"
      onSelectLog: "(logId) => void"
```

---

## Screen 10: Log Detail

```
┌──────────────────────────────────────────────────────────────────────┐
│ ← Logs    Initial review pass                                        │
├──────────────────────────────────────────────────────────────────────┤
│ Kind: note   Source: 🤖 agent · triage-pass-1 · run-42              │
│ Created: 2026-04-02 14:23   By: triage-agent                        │
│                                                                      │
│ ─── Body ───────────────────────────────────────────────────────── │
│ ┌──────────────────────────────────────────────────────────────────┐│
│ │ Grouped likely newsletters based on list-unsubscribe and high    ││
│ │ volume (>50 messages in the last 90 days). Created one group     ││
│ │ "Possible newsletters" with 12 senders.                          ││
│ │                                                                   ││
│ │ ## Criteria used                                                  ││
│ │ - `has_list_unsubscribe = true`                                   ││
│ │ - `msg_count > 50`                                                ││
│ │ - Domain is not in the protected list                             ││
│ │                                                                   ││
│ │ ## Recommendations                                                ││
│ │ Review each sender in the group. Some may be transactional       ││
│ │ (e.g. shipping notifications) rather than true newsletters.      ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ ─── Linked Targets (8) ────────────────────────────────────────── │
│  sender │ news@example.com      → [View]                            │
│  sender │ digest@weekly.io      → [View]                            │
│  sender │ updates@saas-tool.com → [View]                            │
│  …                                                                   │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: LogDetail
props:
  logId: string
  onBack: "() => void"
parts:
  header:
    widget: LogHeader
    data: { log: AnnotationLog }
  body:
    widget: MarkdownRenderer
    data: { markdown: "log.bodyMarkdown" }
  targets:
    widget: LinkedTargetList
    props:
      targets: "LogTarget[]"
      onNavigateTarget: "(type, id) => void"
```

---

## Screen 11: Query Editor

This is the central SQL workbench, closely modeled on go-minitrace's QueryEditor.

```
┌──────────────────────────────────────────────────────────────────────┐
│ ✉ smailnail    [Accounts] [Mailbox] [Rules] [Annotations] [*Query*]│
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
│   ⭐ CI  │  50 rows · 3 cols · 12ms      [⬇ CSV] [⬇ JSON]         │
│          │                                                           │
│ FRAGMENT │  # │ email               │ msg_count │ annotation_count  │
│ 📁 gener │ ───┼─────────────────────┼───────────┼────────────────── │
│   🧩 act │  1 │ news@example.com    │       342 │                3  │
│   🧩 FTS │  2 │ alerts@github.com   │      1205 │                2  │
│   🧩 ann │  3 │ no-reply@amazon.com │        87 │                1  │
│          │  4 │ promo@shop.io       │        65 │                1  │
│          │  … │                     │           │                   │
│          │                                                           │
│          │ [🏷 Annotate Selected]  [📁 Create Group from Results]   │
└──────────┴───────────────────────────────────────────────────────────┘
```

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
          fragments: "QueryFragment[]"
          onSelectQuery: "(q, kind) => void"
          onInsertFragment: "(fragment) => void"
        children:
          - widget: QuerySection
            props: { title: Presets }
            children:
              - widget: QueryFolderGroup
                children:
                  - widget: QueryItem
                    props: { icon: "📄", name: string, tooltip: string }
          - widget: QuerySection
            props: { title: Saved }
            children:
              - widget: QueryFolderGroup
                children:
                  - widget: QueryItem
                    props: { icon: "⭐", name: string }
          - widget: QuerySection
            props: { title: Fragments }
            children:
              - widget: FragmentFolderGroup
                children:
                  - widget: FragmentItem
                    props: { icon: "🧩", name: string, tooltip: string }
      editor:
        widget: SqlEditorPane
        children:
          - widget: SqlEditor
            props:
              value: string
              onChange: "(sql) => void"
              onExecute: "() => void"
          - widget: EditorToolbar
            props:
              onRun: "() => void"
              onSave: "() => void"
              isLoading: boolean
      results:
        widget: ResultsPane
        children:
          - widget: ResultsHeader
            props:
              rowCount: number
              colCount: number
              durationMs: number
              onExportCsv: "() => void"
              onExportJson: "() => void"
          - widget: ResultsTable
            props:
              result: QueryResult
              onClickId: "(col, value) => void"
          - widget: ResultsActions
            props:
              onAnnotateSelected: "() => void"
              onCreateGroup: "() => void"
              hasAnnotatableRows: boolean
```

---

## Screen 12: Messages Browser (with FTS)

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Messages                                               │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Full-text search...   Mailbox: [All ▾]  Has ann: [All ▾]        │
│                                                                      │
│ 48,231 messages · 218 annotated                                      │
│                                                                      │
│  Date       │ From              │ Subject              │ Ann│ Tags   │
│ ────────────┼───────────────────┼──────────────────────┼────┼─────── │
│  2026-04-02 │ news@example.com  │ Your weekly digest   │  1 │ NL     │
│  2026-04-02 │ alerts@github.com │ [go-minitrace] PR #4 │  0 │ —      │
│  2026-04-01 │ john@colleague    │ Re: Q1 Planning      │  2 │ Imp    │
│  …          │                   │                      │    │        │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: MessagesBrowser
parts:
  filters:
    widget: MessageFilters
    props:
      onSearch: "(query) => void"
      onFilterChange: "(f) => void"
    data:
      mailboxes: "string[]"
  summary:
    widget: CountSummary
    data: { total: number, annotated: number }
  table:
    widget: MessageTable
    props:
      messages: "MessageRow[]"
      onSelectMessage: "(id) => void"
    data:
      MessageRow:
        id: number
        date: string
        from: string
        subject: string
        mailbox: string
        annotationCount: number
        topTags: "string[]"
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
│ │ Cc: team@colleague.org           Thread: (5 messages)  [View →]│ │
│ │ Message-ID: <abc123@colleague>   Size: 12KB                    │ │
│ │ Flags: Seen, Flagged                                            │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│ ─── Body ───────────────────────────────────────────────────────── │
│ ┌──────────────────────────────────────────────────────────────────┐│
│ │ Hi Manuel,                                                       ││
│ │                                                                   ││
│ │ I've updated the Q1 planning doc with the latest numbers.        ││
│ │ Please review before the meeting on Friday.                      ││
│ │                                                                   ││
│ │ Best,                                                             ││
│ │ John                                                              ││
│ └──────────────────────────────────────────────────────────────────┘│
│ [Text ▾] [HTML] [Raw]                                                │
│                                                                      │
│ ─── Annotations (2) ────────────────────────────────────────────── │
│  important │ Key planning thread, escalate  │ 🤖 run-42 │ 🟡       │
│  action-req│ Contains explicit review reque │ ⚙️ heurist│ 🟢       │
│                                                                      │
│ [+ Add Annotation]                                                   │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: MessageDetail
props:
  messageId: number
  onBack: "() => void"
parts:
  headers:
    widget: MessageHeaders
    data: { message: MessageRecord }
    props:
      onNavigateThread: "(threadId) => void"
      onNavigateSender: "(email) => void"
  body:
    widget: MessageBody
    data: { bodyText: string, bodyHtml: string, rawPath: string }
    props:
      viewMode: "text | html | raw"
      onChangeViewMode: "(mode) => void"
  annotations:
    widget: TargetAnnotationList
    props:
      targetType: message
      targetId: "$messageId"
      onReviewChange: "(id, state) => void"
      onAddAnnotation: "() => void"
```

---

## Screen 14: Threads Browser

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Threads                                                │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter threads...   Account: [All ▾]                             │
│                                                                      │
│ 8,421 threads · 32 annotated                                        │
│                                                                      │
│  Subject              │ Msgs │ Participants │ Date Range    │ Ann   │
│ ──────────────────────┼──────┼──────────────┼───────────────┼────── │
│  Re: Q1 Planning      │   12 │            5 │ Mar 1 - Apr 2 │    3 │
│  Weekly standup notes  │    8 │            3 │ Mar 15 - Apr 1│    0 │
│  [go-minitrace] PR #4 │    4 │            2 │ Mar 28 - Mar 3│    1 │
│  …                    │      │              │               │       │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: ThreadsBrowser
parts:
  filters:
    widget: ThreadFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { total: number, annotated: number }
  table:
    widget: ThreadTable
    props:
      threads: "ThreadRow[]"
      onSelectThread: "(threadId) => void"
    data:
      ThreadRow:
        threadId: string
        subject: string
        messageCount: number
        participantCount: number
        firstDate: string
        lastDate: string
        annotationCount: number
```

---

## Screen 15: Domains Aggregate

```
┌──────────────────────────────────────────────────────────────────────┐
│ Annotations > Domains                                                │
├──────────────────────────────────────────────────────────────────────┤
│ 🔍 Filter domains...                                                │
│                                                                      │
│ 342 domains · 89 annotated senders                                   │
│                                                                      │
│  Domain            │ Senders │ Messages │ Ann. Senders │ Top Tag     │
│ ───────────────────┼─────────┼──────────┼──────────────┼──────────── │
│  github.com        │       8 │     3420 │            3 │ important   │
│  example.com       │      12 │      890 │            5 │ newsletter  │
│  amazon.com        │       4 │      340 │            2 │ commercial  │
│  …                 │         │          │              │             │
│                                                                      │
│ Click row → filtered Senders view for that domain                   │
└──────────────────────────────────────────────────────────────────────┘
```

```yaml
widget: DomainsBrowser
parts:
  filters:
    widget: DomainFilters
    props: { onFilterChange: "(f) => void" }
  summary:
    widget: CountSummary
    data: { total: number, annotatedSenders: number }
  table:
    widget: DomainTable
    props:
      domains: "DomainRow[]"
      onSelectDomain: "(domain) => void"
    data:
      DomainRow:
        domain: string
        senderCount: number
        messageCount: number
        annotatedSenderCount: number
        topTag: string
```

---

## Shared / Reusable Widgets

These widgets appear across multiple screens:

```yaml
# Shared widgets used in multiple screens

widget: SourceBadge
props:
  sourceKind: "agent | human | heuristic | import"
  sourceLabel: "string | undefined"
  tooltip: boolean

widget: ReviewStateChip
props:
  state: "to_review | reviewed | dismissed"
  onClick: "() => void | undefined"

widget: ReviewProgressBar
props:
  reviewed: number
  toReview: number
  dismissed: number
  compact: boolean

widget: TargetAnnotationList
# Reused in: SenderDetail, MessageDetail, ThreadDetail
props:
  targetType: string
  targetId: string
  annotations: "Annotation[]"
  onReviewChange: "(id, newState) => void"
  onAddAnnotation: "() => void"

widget: CountSummary
# Reused in: every browser view
props:
  items: "Array<{label: string, count: number}>"

widget: BatchActionBar
# Reused in: ReviewQueue, AgentRunDetail
props:
  selectedCount: number
  onSelectAll: "() => void"
  onApprove: "() => void"
  onDismiss: "() => void"
  onReset: "() => void"

widget: MarkdownRenderer
props:
  markdown: string
  maxHeight: "number | undefined"

widget: LinkedTargetList
# Reused in: LogDetail, GroupDetail
props:
  targets: "Array<{type: string, id: string}>"
  onNavigateTarget: "(type, id) => void"
```

---

## Routing Summary

```yaml
routes:
  /accounts: AccountSetupPage  # existing
  /mailbox: MailboxExplorer    # existing
  /rules: RulesPage            # existing

  /annotations: AnnotationsPage
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

## Component File Structure

```
ui/src/features/annotations/
├── AnnotationsPage.tsx          # Tab container
├── ReviewQueue.tsx              # Screen 2
├── ReviewFilters.tsx
├── AnnotationTable.tsx
├── AnnotationRow.tsx
├── AnnotationDetail.tsx
├── SendersBrowser.tsx           # Screen 3
├── SenderDetail.tsx             # Screen 4
├── SenderProfileCard.tsx
├── ThreadsBrowser.tsx           # Screen 14
├── MessagesBrowser.tsx          # Screen 12
├── MessageDetail.tsx            # Screen 13
├── DomainsBrowser.tsx           # Screen 15
├── GroupsBrowser.tsx            # Screen 5
├── GroupDetail.tsx              # Screen 6
├── AgentRunsBrowser.tsx         # Screen 7
├── AgentRunDetail.tsx           # Screen 8
├── LogsBrowser.tsx              # Screen 9
├── LogDetail.tsx                # Screen 10
└── shared/
    ├── SourceBadge.tsx
    ├── ReviewStateChip.tsx
    ├── ReviewProgressBar.tsx
    ├── TargetAnnotationList.tsx
    ├── BatchActionBar.tsx
    ├── LinkedTargetList.tsx
    └── CountSummary.tsx

ui/src/features/query/
├── QueryEditorPage.tsx          # Screen 11
├── QueryEditor.tsx
├── QuerySidebar.tsx
├── SqlEditor.tsx
├── ResultsTable.tsx
├── ResultsActions.tsx
├── FragmentItem.tsx
└── types.ts
```
