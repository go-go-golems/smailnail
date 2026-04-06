---
Title: 'UI Design Document: Review Feedback, Guidelines & Mailbox-Aware Screens'
Ticket: SMN-20260403-RUN-REVIEW
Status: active
Topics:
    - annotations
    - frontend
    - design
    - react
    - widgets
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ui/src/api/annotations.ts
      Note: RTK Query contract to extend with feedback/guideline endpoints
    - Path: ui/src/types/annotations.ts
      Note: TypeScript types to extend with feedback, guideline, mailbox types
    - Path: ui/src/pages/ReviewQueuePage.tsx
      Note: Primary screen getting review-comment affordance
    - Path: ui/src/pages/RunDetailPage.tsx
      Note: Run detail getting guideline linking and run-level feedback
    - Path: ui/src/components/shared/BatchActionBar.tsx
      Note: Batch bar to extend with comment affordance
    - Path: ui/src/components/shared/parts.ts
      Note: Shared data-part namespace
    - Path: ui/src/components/AppLayout/AnnotationSidebar.tsx
      Note: Sidebar getting new Guidelines nav entry
Summary: Complete UI design for review feedback entry, guidelines management, run-guideline linking, and mailbox-aware context across all annotation review screens. Screen-by-screen ASCII renders and React widget pseudo-DSL.
LastUpdated: 2026-04-04T00:00:00-04:00
WhatFor: Use this document as the implementation blueprint for all frontend changes in this ticket.
WhenToUse: Open this before writing any component code for review feedback, guidelines, or mailbox context.
---

# UI Design Document: Review Feedback, Guidelines & Mailbox-Aware Screens

## 1. Overview

This document specifies every screen, widget, and interaction for the frontend half of SMN-20260403-RUN-REVIEW. It covers:

- **Review Feedback Panel** — inline comment/revision-request entry on single and batch review actions
- **Guidelines Management** — list, create, edit, and archive reusable review policies
- **Run-Guideline Linking** — attach/detach guidelines to agent runs
- **Mailbox Context** — expose `mailboxName` in tables, badges, filters, and feedback payloads
- **Sidebar & Navigation** — new nav entries and route structure

All designs follow the existing MUI dark-theme widget pattern with `data-widget`/`data-part` attributes, `parts.ts` namespaces, and Storybook stories.

---

## 2. Widget Tree & New Types

### 2.1 New TypeScript types

```yaml
# types/reviewFeedback.ts
ReviewFeedback:
  id: string
  scopeKind: "annotation" | "selection" | "run" | "guideline"
  agentRunId: string
  mailboxName: string
  feedbackKind: "comment" | "reject_request" | "guideline_request" | "clarification"
  status: "open" | "acknowledged" | "resolved" | "archived"
  title: string
  bodyMarkdown: string
  createdBy: string
  createdAt: string
  updatedAt: string
  targets: FeedbackTarget[]

FeedbackTarget:
  targetType: string
  targetId: string

CreateFeedbackRequest:
  scopeKind: string
  agentRunId?: string
  mailboxName?: string
  feedbackKind: string
  title: string
  bodyMarkdown: string
  targetIds?: string[]

ReviewCommentDraft:
  feedbackKind: string
  title: string
  bodyMarkdown: string

# types/reviewGuideline.ts
ReviewGuideline:
  id: string
  slug: string
  title: string
  scopeKind: "global" | "mailbox" | "sender" | "domain" | "workflow"
  status: "active" | "archived" | "draft"
  priority: number
  bodyMarkdown: string
  createdBy: string
  createdAt: string
  updatedAt: string

CreateGuidelineRequest:
  slug: string
  title: string
  scopeKind: string
  bodyMarkdown: string

UpdateGuidelineRequest:
  title?: string
  scopeKind?: string
  status?: string
  priority?: number
  bodyMarkdown?: string

# Extended types
MessagePreview (extended):
  mailboxName: string  # NEW

AnnotationFilter (extended):
  mailboxName?: string
  feedbackStatus?: string

ReviewPayload (extended):
  id: string
  reviewState: string
  comment?: ReviewCommentDraft
  guidelineIds?: string[]
  mailboxName?: string

BatchReviewPayload (extended):
  ids: string[]
  reviewState: string
  comment?: ReviewCommentDraft
  guidelineIds?: string[]
  agentRunId?: string
  mailboxName?: string
```

### 2.2 New RTK Query endpoints

```yaml
# api/annotations.ts — additions
reviewAnnotation (extended):
  payload: ReviewPayload

batchReview (extended):
  payload: BatchReviewPayload

listReviewFeedback:
  query: { agentRunId?: string, status?: string, feedbackKind?: string }
  response: ReviewFeedback[]

getReviewFeedback:
  query: string (id)
  response: ReviewFeedback

createReviewFeedback:
  payload: CreateFeedbackRequest
  response: ReviewFeedback

updateReviewFeedback:
  payload: { id: string, status?: string, bodyMarkdown?: string }
  response: ReviewFeedback

listGuidelines:
  query: { status?: string, scopeKind?: string, search?: string }
  response: ReviewGuideline[]

getGuideline:
  query: string (id)
  response: ReviewGuideline

createGuideline:
  payload: CreateGuidelineRequest
  response: ReviewGuideline

updateGuideline:
  payload: { id: string } & UpdateGuidelineRequest
  response: ReviewGuideline

getRunGuidelines:
  query: string (runId)
  response: ReviewGuideline[]

linkGuidelineToRun:
  payload: { runId: string, guidelineId: string }
  response: void

unlinkGuidelineFromRun:
  payload: { runId: string, guidelineId: string }
  response: void
```

### 2.3 New `data-part` namespace additions

```yaml
# components/shared/parts.ts additions
mailboxBadge: mailbox-badge
feedbackStatusBadge: feedback-status-badge
feedbackKindBadge: feedback-kind-badge
guidelineScopeBadge: guideline-scope-badge

# components/ReviewFeedback/parts.ts
feedbackPanel: feedback-panel
feedbackList: feedback-list
feedbackCard: feedback-card
feedbackForm: feedback-form

# components/Guidelines/parts.ts
guidelineList: guideline-list
guidelineCard: guideline-card
guidelineEditor: guideline-editor
guidelineLinkPicker: guideline-link-picker

# components/shared/ReviewCommentDrawer/parts.ts
commentDrawer: comment-drawer
commentForm: comment-form
guidelinePicker: guideline-picker
```

---

## 3. Screen Designs

### 3.1 Sidebar — Updated Navigation

```
┌──────────────────────┐
│  OVERVIEW            │
│  ▸ Dashboard         │
│                      │
│  REVIEW              │
│  ▸ Review Queue      │
│  ▸ Agent Runs        │
│  ▸ Guidelines     ✦  │  ← NEW
│                      │
│  BROWSE              │
│  ▸ Senders           │
│  ▸ Groups            │
│                      │
│  TOOLS               │
│  ▸ SQL Workbench     │
└──────────────────────┘
```

```yaml
AnnotationSidebar:
  modification: add Guidelines entry to reviewItems
  new entry:
    label: "Guidelines"
    icon: MenuBookIcon
    path: /annotations/guidelines

App.tsx new routes:
  - path: /annotations/guidelines
    element: GuidelinesListPage
  - path: /annotations/guidelines/new
    element: GuidelineDetailPage
  - path: /annotations/guidelines/:guidelineId
    element: GuidelineDetailPage
```

---

### 3.2 Review Queue Page — Enhanced with Comment Drawer

The review queue keeps its fast approve/dismiss actions and adds an optional "Reject & Explain" expansion.

**Default state (no drawer):**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Review Queue                                                                │
│                                                                             │
│ [newsletter] [transactional] [promotional] [bulk]                           │
│ to review: 23  ·  agent: 45  ·  heuristic: 12                              │
│                                                                             │
│ ☐ All  67 items                                     [Approve] [Dismiss ▾]  │
│ ─────────────────────────────────────────────────────────────────────────── │
│ ☐ │ sender │ newsletter │ "Likely promo..." │ agent │ ⬤ to_review │ Apr 3 │
│ ☐ │ sender │ receipt    │ "Transactional"   │ agent │ ⬤ to_review │ Apr 3 │
│ ☐ │ sender │ bulk       │ "Marketing blast" │ agent │ ⬤ to_review │ Apr 3 │
│    │                                             │ [✓] [✗]              │     │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Batch reject with drawer expanded (3 selected):**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Review Queue                                                                │
│                                                                             │
│ [newsletter] [transactional] [promotional] [bulk]                           │
│ to review: 23  ·  agent: 45  ·  heuristic: 12                              │
│                                                                             │
│ ☑ 3 of 67 selected               [Approve] [Reject & Explain] [Reset]     │
│ ─────────────────────────────────────────────────────────────────────────── │
│ ☑ │ sender │ newsletter │ "Likely promo..." │ agent │ ⬤ to_review │ Apr 3 │
│ ☑ │ sender │ receipt    │ "Transactional"   │ agent │ ⬤ to_review │ Apr 3 │
│ ☑ │ sender │ bulk       │ "Marketing blast" │ agent │ ⬤ to_review │ Apr 3 │
│ ─────────────────────────────────────────────────────────────────────────── │
│ ┌─── Reject & Explain ────────────────────────────────────────────────────┐ │
│ │                                                                         │ │
│ │  Kind:  [▼ reject_request  ]                                            │ │
│ │  Title: [Wrong sender categorization_____________________________]       │ │
│ │                                                                         │ │
│ │  ┌─ Explanation ──────────────────────────────────────────────────────┐ │ │
│ │  │ These messages are transactional receipts, not newsletters.        │ │ │
│ │  │ The agent treated invoice notifications as marketing mail.         │ │ │
│ │  └───────────────────────────────────────────────────────────────────┘ │ │
│ │                                                                         │ │
│ │  Attach Guidelines:                                                     │ │
│ │    [+ Add Guideline]                                                    │ │
│ │    ┌─────────────────────────────────────────────┐                      │ │
│ │    │ ☑ transactional-vs-promotional (priority 50)│                      │ │
│ │    │ ☐ billing-mail-classification               │                      │ │
│ │    └─────────────────────────────────────────────┘                      │ │
│ │                                                                         │ │
│ │  Mailbox: INBOX                                                         │ │
│ │                                                                         │ │
│ │                            [Cancel]  [Reject 3 Items]                   │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

```yaml
ReviewQueuePage (enhanced):
  data-widget: review-queue-page
  children:
    - Title
    - FilterPillBar
    - CountSummaryBar
    - BatchActionBar (enhanced):
        new prop: onRejectExplain: () => void
        existing: onApprove, onDismiss, onReset
    - AnnotationTable (unchanged externally)
    - ReviewCommentDrawer:  # NEW
        data-widget: comment-drawer
        props:
          - open: boolean
          - mode: "single" | "batch"
          - targetCount: number
          - agentRunId?: string
          - mailboxName?: string
          - onSubmit: (payload: BatchReviewPayload) => void
          - onCancel: () => void
        children:
          - feedbackKind select
          - title textfield
          - bodyMarkdown textarea
          - GuidelinePicker:
              data-part: guideline-picker
              props:
                - selectedIds: string[]
                - onToggle: (id) => void
              fetches: useListGuidelinesQuery({ status: "active" })
          - mailboxName display (read-only)
          - Cancel + Submit buttons
```

---

### 3.3 Single Annotation Dismiss — Inline Comment

When a reviewer clicks the dismiss icon on a single annotation, an inline expansion appears below the annotation detail row offering both fast dismiss and comment paths.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ☐ │ sender │ newsletter │ "Likely promo..." │ agent │ ⬤ to_review │ Apr 3 │
│    │ (expanded detail...)                                           │       │
│    │ ┌── Dismiss with Feedback ────────────────────────────────────┐       │
│    │ │                                                             │       │
│    │ │  Kind:  [▼ reject_request  ]                                │       │
│    │ │  Title: [Misclassified sender______________________]        │       │
│    │ │  Note:                                                      │       │
│    │ │  ┌──────────────────────────────────────────────────────┐   │       │
│    │ │  │ This is a transactional sender, not a newsletter.   │   │       │
│    │ │  │ Please re-classify on next run.                     │   │       │
│    │ │  └──────────────────────────────────────────────────────┘   │       │
│    │ │                                                             │       │
│    │ │  [+ Attach Guideline]                                       │       │
│    │ │                                                             │       │
│    │ │                    [Just Dismiss] [Dismiss & Explain]       │       │
│    │ └─────────────────────────────────────────────────────────────┘       │
│ ☐ │ sender │ receipt    │ "Transactional"   │ agent │ ⬤ to_review │ Apr 3 │
└─────────────────────────────────────────────────────────────────────────────┘
```

```yaml
AnnotationDetail (enhanced):
  new child: ReviewCommentInline
  props:
    - open: boolean
    - annotationId: string
    - agentRunId: string
    - mailboxName?: string
    - onSubmit: (reviewState, comment?, guidelineIds?) => void
    - onJustDismiss: () => void
    - onCancel: () => void
  children:
    - feedbackKind select
    - title textfield
    - bodyMarkdown textarea
    - GuidelinePicker (compact)
    - "Just Dismiss" button (fast path)
    - "Dismiss & Explain" button (with comment)
```

---

### 3.4 Run Detail Page — Enhanced with Guidelines & Run-Level Feedback

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ← All Runs          Run: run-2026-04-03-triage                   [Approve] │
│                                                                             │
│  Total: 67  │  Pending: 23  │  Reviewed: 32  │  Dismissed: 12             │
│                                                                             │
│ ─── Linked Guidelines ──────────────────────────────────────────────────── │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ 📘 transactional-vs-promotional     workflow   ● active   pri: 50      │ │
│ │   If the primary purpose is a receipt or confirmation, do not tag      │ │
│ │   it as newsletter.                                                    │ │
│ │                                                [Unlink]                │ │
│ ├─────────────────────────────────────────────────────────────────────────┤ │
│ │ 📘 billing-mail-classification      global    ● active   pri: 30      │ │
│ │   Billing and invoice emails should be categorized separately          │ │
│ │   from promotional mail regardless of sender domain.                   │ │
│ │                                                [Unlink]                │ │
│ ├─────────────────────────────────────────────────────────────────────────┤ │
│ │ [+ Link Existing Guideline] [Create New Guideline for This Run]        │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ ─── Agent Log (5 entries) ──────────────────────────────────────────────── │
│ 14:23:05 ● reasoning   Analyzing sender patterns...                        │
│ 14:23:12 ● decision    Classified 45 senders as newsletter                 │
│                                                                             │
│ ─── Run-Level Feedback ─────────────────────────────────────────────────── │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ [✏️ Add Run Feedback]                                                   │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│ ┌─ Feedback #001 ─────────────────────────────────────────────────────────┐ │
│ │ ⚠ reject_request   │ ● open   │ manuel   │ Apr 3, 14:30               │ │
│ │ Misclassified financial messages                                        │ │
│ │ Please separate invoices and receipts from promotional newsletters.     │ │
│ │ Targets: 3 annotations                    [Acknowledge] [Resolve]       │ │
│ ├─ Feedback #002 ─────────────────────────────────────────────────────────┤ │
│ │ 💬 comment         │ ✓ resolved │ manuel  │ Apr 3, 14:15               │ │
│ │ Good thread reconstruction                                              │ │
│ │ The agent correctly identified conversation threads across mailboxes.   │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ ─── Target Groups (2) ──────────────────────────────────────────────────── │
│ (existing GroupCards)                                                       │
│                                                                             │
│ ─── Annotations (67) ──────────────────────────────────────────────────── │
│ (existing AnnotationTable with enhanced inline comment on dismiss)          │
└─────────────────────────────────────────────────────────────────────────────┘
```

```yaml
RunDetailPage (enhanced):
  data-widget: run-detail-page
  children:
    - Header (existing, enhanced)
    - StatBox row (existing)
    - RunGuidelineSection:  # NEW
        data-widget: run-guideline-section
        props:
          - runId: string
          - guidelines: ReviewGuideline[]
          - onLink: (guidelineId) => void
          - onUnlink: (guidelineId) => void
          - onCreateAndLink: () => void
        children:
          - GuidelineCard[] (compact):
              props:
                - guideline: ReviewGuideline
                - onUnlink: () => void
          - action buttons:
              - LinkGuidelineButton → GuidelineLinkPicker modal
              - CreateGuidelineButton → nav to /annotations/guidelines/new?runId=...
    - RunTimeline (existing)
    - RunFeedbackSection:  # NEW
        data-widget: run-feedback-section
        props:
          - runId: string
          - feedback: ReviewFeedback[]
          - onCreateFeedback: () => void
          - onUpdateStatus: (id, status) => void
        children:
          - AddFeedbackButton → ReviewCommentDrawer (mode="run")
          - FeedbackCard[]:
              props:
                - feedback: ReviewFeedback
                - onAcknowledge / onResolve
              children:
                - FeedbackKindBadge
                - FeedbackStatusBadge
                - createdBy + createdAt
                - title
                - bodyMarkdown (MarkdownRenderer)
                - target count
                - status action buttons
    - GroupCard[] (existing)
    - AnnotationTable (existing, enhanced)
```

---

### 3.5 Guidelines List Page — New Screen

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ← Dashboard          Review Guidelines                [+ New Guideline]    │
│                                                                             │
│ [All] [Active] [Archived] [Draft]          Search: [________________] 🔍   │
│                                                                             │
│ 12 guidelines  ·  8 active  ·  2 archived  ·  2 draft                     │
│                                                                             │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ 📘 transactional-vs-promotional                                        │ │
│ │   Separate transactional mail from promotional mail                    │ │
│ │   workflow  │  ● active  │  pri: 50  │  2 runs linked                 │ │
│ │   Updated Apr 3  │  Created Apr 1                                     │ │
│ │                                              [Edit] [Archive]          │ │
│ ├─────────────────────────────────────────────────────────────────────────┤ │
│ │ 📘 billing-mail-classification                                         │ │
│ │   Billing and invoice emails should be categorized separately          │ │
│ │   global   │  ● active  │  pri: 30  │  1 run linked                   │ │
│ │   Updated Apr 2  │  Created Mar 28                                    │ │
│ │                                              [Edit] [Archive]          │ │
│ ├─────────────────────────────────────────────────────────────────────────┤ │
│ │ 📘 sender-domain-normalization                                         │ │
│ │   When the same sender appears from multiple domains...                │ │
│ │   sender   │  ○ draft   │  pri: 0   │  0 runs linked                  │ │
│ │                                              [Edit] [Activate]         │ │
│ ├─────────────────────────────────────────────────────────────────────────┤ │
│ │ 📘 newsletter-vs-circular                                              │ │
│ │   Circular emails from community groups are not commercial newsletters │ │
│ │   mailbox  │  ✕ archived│  pri: 20  │  3 runs linked                  │ │
│ │                                              [Edit] [Reactivate]       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

```yaml
GuidelinesListPage:
  data-widget: guidelines-list-page
  route: /annotations/guidelines
  children:
    - PageHeader:
        title: "Review Guidelines"
        actions: [NewGuidelineButton]
    - FilterBar:
        - StatusFilterPills: [all, active, archived, draft]
        - SearchField: filters title + slug + body
    - CountSummaryBar
    - GuidelineList:
        children:
          - GuidelineSummaryCard[]:
              props:
                - guideline: ReviewGuideline
                - linkedRunCount: number
                - onEdit / onArchive / onActivate
              children:
                - title + slug
                - body (truncated 2 lines)
                - GuidelineScopeBadge
                - status badge
                - priority
                - linked run count
                - dates
                - action buttons
```

---

### 3.6 Guideline Detail / Editor Page

**View mode:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ← Guidelines     transactional-vs-promotional              [Edit]          │
│                                                                             │
│ ┌─ Metadata ─────────────────────────────────────────────────────────────┐  │
│ │ Slug:      transactional-vs-promotional                                │  │
│ │ Scope:     workflow                                                    │  │
│ │ Status:    ● active                                                    │  │
│ │ Priority:  50                                                          │  │
│ │ Created:   Apr 1, 2026 by manuel                                      │  │
│ │ Updated:   Apr 3, 2026                                                │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│ ┌─ Body ─────────────────────────────────────────────────────────────────┐  │
│ │  # Separate Transactional Mail from Promotional Mail                    │  │
│ │                                                                         │  │
│ │  If the primary purpose of an email is to deliver a receipt,           │  │
│ │  confirmation, or account action notification, do not tag it as         │  │
│ │  a newsletter regardless of the sender's domain or unsubscribe          │  │
│ │  header presence.                                                       │  │
│ │                                                                         │  │
│ │  Key signals:                                                           │  │
│ │  - Subject contains "receipt", "confirmation", "invoice"               │  │
│ │  - Sender is a known billing platform (stripe, shopify, etc.)          │  │
│ │  - Email has no promotional imagery or marketing links                  │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│ ─── Linked Runs (2) ──────────────────────────────────────────────────────  │
│ ┌────────────────────────────────────────────────────────────────────────┐  │
│ │ run-2026-04-03-triage    67 annotations   23 pending   Apr 3          │  │
│ │ run-2026-04-01-triage    45 annotations   0 pending    Apr 1          │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│ ─── Feedback Referencing This Guideline ──────────────────────────────────  │
│ (compact FeedbackCard list)                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Edit mode:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ← Guidelines     Edit: transactional-vs-promotional       [Save] [Cancel] │
│                                                                             │
│ Slug:     [transactional-vs-promotional    ]  (read-only)                   │
│ Title:    [Separate Transactional Mail from Promotional Mail_____________]  │
│ Scope:    [▼ workflow     ]                                                │
│ Status:   [▼ active       ]                                                │
│ Priority: [50]                                                             │
│                                                                             │
│ ┌─ Body (Markdown) ─────────────────────────────────────────────────────┐  │
│ │ # Separate Transactional Mail from Promotional Mail                    │  │
│ │                                                                         │  │
│ │ If the primary purpose of an email is to deliver a receipt, ...        │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│ ┌─ Preview ─────────────────────────────────────────────────────────────┐  │
│ │ (live rendered markdown via MarkdownRenderer)                          │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

```yaml
GuidelineDetailPage:
  data-widget: guideline-detail-page
  route: /annotations/guidelines/:guidelineId
  modes: [view, edit, create]
  children:
    - PageHeader with mode-dependent actions
    - GuidelineForm:
        props:
          - guideline?: ReviewGuideline
          - mode: view | edit | create
          - onSave / onCancel
        children:
          - slug (ro in edit, editable in create)
          - title textfield
          - scopeKind select
          - status select
          - priority input
          - bodyMarkdown textarea
          - MarkdownRenderer preview
    - GuidelineLinkedRuns (view only):
        props: runs: AgentRunSummary[]
    - GuidelineFeedbackList (view only):
        props: feedback: ReviewFeedback[]
```

---

### 3.7 Guideline Link Picker — Modal

Used from RunDetailPage to search and select guidelines to link.

```
┌─── Link Guideline to Run ────────────────────────────────────────────────┐
│                                                                           │
│  Search: [_________________________________________] 🔍                   │
│                                                                           │
│  ☐ transactional-vs-promotional   workflow   active   pri 50             │
│  ☐ billing-mail-classification    global    active   pri 30             │
│  ☑ sender-domain-normalization    sender    draft    pri 0              │
│  ☐ newsletter-vs-circular         mailbox   archived pri 20             │
│  ☐ unsubscribe-detection          workflow  active   pri 40             │
│                                                                           │
│                                           [Cancel] [Link 1 Guideline]    │
└───────────────────────────────────────────────────────────────────────────┘
```

```yaml
GuidelineLinkPicker:
  data-part: guideline-link-picker
  component: Dialog (MUI modal)
  props:
    - open: boolean
    - runId: string
    - alreadyLinkedIds: string[]
    - onLink: (guidelineIds: string[]) => void
    - onClose: () => void
  children:
    - SearchField (local filter on title/slug)
    - Checklist of active guidelines
    - Cancel + Link buttons
  data source: useListGuidelinesQuery({ status: "active" })
```

---

## 4. Mailbox Context — Cross-Cutting

### 4.1 MailboxBadge — New Shared Widget

```
 ┌──────────┐  ┌──────────┐  ┌──────────┐
 │ 📬 INBOX │  │ 📁 Bills │  │ 📂 Archive│
 └──────────┘  └──────────┘  └──────────┘
```

```yaml
MailboxBadge:
  data-part: mailbox-badge
  props:
    - mailboxName: string
    - variant: "chip" | "inline"
  behavior:
    - empty string → renders nothing
    - known mailboxes get icon (INBOX→mailbox, Sent→send, Archive→archive)
    - unknown mailboxes get folder icon
```

### 4.2 Where MailboxBadge appears

```yaml
placements:
  - AnnotationTable: new column after "Source" (compact, only if mailboxName differs across visible rows)
  - AnnotationDetail: in the header metadata row, after ReviewStateBadge
  - MessagePreviewTable: new "Mailbox" column after "Subject"
  - SenderDetailPage.recentMessages: same as MessagePreviewTable
  - ReviewCommentDrawer: read-only display, derived from selection
  - FeedbackCard: if feedback.mailboxName is non-empty
```

### 4.3 Mailbox filter in Review Queue

```yaml
FilterPillBar (enhanced):
  new prop: mailboxPills
  behavior:
    - when annotations span multiple mailboxes, show mailbox filter pills
    - active mailbox filter applied via AnnotationFilter.mailboxName
```

---

## 5. Feedback Status Badges — New Shared Widgets

### 5.1 FeedbackKindBadge

```
  ⚠ reject_request    💬 comment    📘 guideline_request    ❓ clarification
```

```yaml
FeedbackKindBadge:
  data-part: feedback-kind-badge
  props:
    - kind: string
  colors:
    reject_request: error
    comment: info
    guideline_request: warning
    clarification: default
```

### 5.2 FeedbackStatusBadge

```
  ● open    ◐ acknowledged    ✓ resolved    ✕ archived
```

```yaml
FeedbackStatusBadge:
  data-part: feedback-status-badge
  props:
    - status: string
  colors:
    open: warning
    acknowledged: info
    resolved: success
    archived: default
```

### 5.3 GuidelineScopeBadge

```
  🌐 global    📬 mailbox    👤 sender    🏢 domain    ⚙ workflow
```

```yaml
GuidelineScopeBadge:
  data-part: guideline-scope-badge
  props:
    - scopeKind: string
```

---

## 6. Component File Map

```
ui/src/
├── types/
│   ├── annotations.ts          # EXTEND: MessagePreview, Filter types
│   ├── reviewFeedback.ts       # NEW
│   └── reviewGuideline.ts      # NEW
│
├── api/
│   └── annotations.ts          # EXTEND: new endpoints + extended payloads
│
├── components/
│   ├── shared/
│   │   ├── parts.ts            # EXTEND: new data-part names
│   │   ├── MailboxBadge.tsx    # NEW
│   │   ├── FeedbackKindBadge.tsx  # NEW
│   │   ├── FeedbackStatusBadge.tsx # NEW
│   │   ├── GuidelineScopeBadge.tsx # NEW
│   │   ├── BatchActionBar.tsx  # EXTEND: add onRejectExplain
│   │   └── stories/
│   │       ├── MailboxBadge.stories.tsx       # NEW
│   │       ├── FeedbackKindBadge.stories.tsx  # NEW
│   │       ├── FeedbackStatusBadge.stories.tsx # NEW
│   │       └── GuidelineScopeBadge.stories.tsx # NEW
│   │
│   ├── AnnotationTable/
│   │   ├── AnnotationDetail.tsx  # EXTEND: add ReviewCommentInline
│   │   └── AnnotationRow.tsx     # EXTEND: MailboxBadge column
│   │
│   ├── ReviewFeedback/           # NEW directory
│   │   ├── parts.ts
│   │   ├── ReviewCommentDrawer.tsx
│   │   ├── ReviewCommentInline.tsx
│   │   ├── FeedbackCard.tsx
│   │   ├── RunFeedbackSection.tsx
│   │   ├── GuidelinePicker.tsx
│   │   ├── GuidelineLinkPicker.tsx
│   │   └── stories/
│   │       ├── ReviewCommentDrawer.stories.tsx
│   │       ├── FeedbackCard.stories.tsx
│   │       ├── RunFeedbackSection.stories.tsx
│   │       └── GuidelinePicker.stories.tsx
│   │
│   ├── Guidelines/               # NEW directory
│   │   ├── parts.ts
│   │   ├── GuidelineSummaryCard.tsx
│   │   ├── GuidelineForm.tsx
│   │   ├── GuidelineLinkedRuns.tsx
│   │   └│   │       ├── GuidelineSummaryCard.stories.tsx
│   │       └── GuidelineForm.stories.tsx
│   │
│   ├── RunGuideline/              # NEW directory
│   │   ├── parts.ts
│   │   ├── RunGuidelineSection.tsx
│   │   └── stories/
│   │       └── RunGuidelineSection.stories.tsx
│   │
│   └── AppLayout/
│       └── AnnotationSidebar.tsx  # EXTEND: add Guidelines nav entry
│
├── pages/
│   ├── ReviewQueuePage.tsx       # EXTEND: add ReviewCommentDrawer
│   ├── RunDetailPage.tsx          # EXTEND: add RunGuidelineSection + RunFeedbackSection
│   ├── GuidelinesListPage.tsx    # NEW
│   ├── GuidelineDetailPage.tsx   # NEW
│   └── stories/
│       ├── GuidelinesListPage.stories.tsx    # NEW
│       └── GuidelineDetailPage.stories.tsx   # NEW
│
└── App.tsx                       # EXTEND: new guideline routes
```

---

## 7. Interaction Flows

### 7.1 Fast Approve (no comment)

```
Reviewer clicks ✓ on annotation row
  → reviewAnnotation({ id, reviewState: "reviewed" })
  → annotation state updates in-place
  → no drawer, no modal
```

### 7.2 Fast Dismiss (no comment)

```
Reviewer clicks ✗ on annotation row
  → inline ReviewCommentInline appears (collapsed)
  → Reviewer clicks "Just Dismiss"
  → reviewAnnotation({ id, reviewState: "dismissed" })
  → inline panel closes
```

### 7.3 Dismiss with Explanation (single)

```
Reviewer clicks ✗ on annotation row
  → inline ReviewCommentInline appears
  → Reviewer fills kind, title, body, optionally picks guidelines
  → Reviewer clicks "Dismiss & Explain"
  → reviewAnnotation({
      id,
      reviewState: "dismissed",
      comment: { feedbackKind, title, bodyMarkdown },
      guidelineIds: [...],
      mailboxName: "INBOX"
    })
  → backend creates review_feedback + links guideline to run
  → annotation state updates, feedback appears in RunFeedbackSection
```

### 7.4 Batch Reject & Explain

```
Reviewer selects 3 annotations
  → clicks "Reject & Explain" in BatchActionBar
  → ReviewCommentDrawer slides open below batch bar
  → Reviewer fills kind, title, body, picks guidelines
  → clicks "Reject 3 Items"
  → batchReview({
      ids: [a1, a2, a3],
      reviewState: "dismissed",
      comment: { ... },
      guidelineIds: [...],
      mailboxName: "INBOX"
    })
  → backend creates one feedback targeting all 3 annotations
  → drawer closes, selection clears, annotations update
```

### 7.5 Link Guideline to Run

```
Reviewer on RunDetailPage
  → clicks "+ Link Existing Guideline" in RunGuidelineSection
  → GuidelineLinkPicker modal opens (list of active guidelines)
  → Reviewer checks one or more guidelines
  → clicks "Link N Guidelines"
  → linkGuidelineToRun called for each
  → modal closes, RunGuidelineSection refreshes
```

### 7.6 Create Guideline from Run Context

```
Reviewer on RunDetailPage
  → clicks "Create New Guideline for This Run"
  → navigates to /annotations/guidelines/new?runId=run-42
  → GuidelineDetailPage opens in create mode
  → reviewer fills slug, title, body, scope
  → clicks "Create"
  → createGuideline + linkGuidelineToRun called
  → navigates back to run detail
```

### 7.7 Add Run-Level Feedback

```
Reviewer on RunDetailPage
  → clicks "Add Run Feedback" in RunFeedbackSection
  → ReviewCommentDrawer opens (mode="run")
  → fills kind, title, body
  → clicks "Submit"
  → createReviewFeedback({ scopeKind: "run", agentRunId, ... })
  → drawer closes, feedback card appears in section
```

---

## 8. State Management

No new Redux slices needed. The existing `annotationUiSlice` manages review-queue selection and filter state. All new data flows through RTK Query cache:

```yaml
cache_tags:
  - ReviewFeedback  # NEW — invalidated on create/update
  - Guidelines      # NEW — invalidated on create/update/link/unlink
  - Annotations     # existing — invalidated on review with comment
  - Runs            # existing — invalidated on guideline link

optimistic_updates:
  - single review state toggle: existing pattern
  - batch review: existing pattern
  - no optimistic update for feedback creation (too complex for v1)
```

---

## 9. Storybook Coverage Plan

Every new widget gets a Storybook story. Stories cover:

```yaml
stories_to_create:
  - MailboxBadge:
      - empty mailboxName
      - INBOX
      - custom mailbox name
      - "Sent" variant
  - FeedbackKindBadge:
      - each kind with color
  - FeedbackStatusBadge:
      - each status with color
  - GuidelineScopeBadge:
      - each scope with icon
  - ReviewCommentDrawer:
      - closed state
      - open batch mode
      - open single mode
      - with pre-filled guideline
  - FeedbackCard:
      - open reject_request
      - resolved comment
      - with targets
  - RunFeedbackSection:
      - empty (no feedback)
      - multiple feedback items
      - with status transitions
  - GuidelinePicker:
      - no guidelines selected
      - one selected
  - GuidelineLinkPicker:
      - empty search
      - filtered
      - with selection
  - GuidelineSummaryCard:
      - active guideline
      - archived guideline
      - draft guideline
  - GuidelineForm:
      - create mode
      - edit mode
      - view mode
  - RunGuidelineSection:
      - no linked guidelines
      - two linked guidelines
  - GuidelinesListPage:
      - mixed status guidelines
      - empty state
      - filtered to active
  - GuidelineDetailPage:
      - view mode with linked runs
      - edit mode
      - create mode
```

---

## 10. Implementation Sequence (Frontend Only)

```yaml
phase_1_types_and_api:
  - add types/reviewFeedback.ts
  - add types/reviewGuideline.ts
  - extend types/annotations.ts (MessagePreview, Filter)
  - extend api/annotations.ts (new endpoints, extended payloads)

phase_2_shared_badges:
  - MailboxBadge
  - FeedbackKindBadge
  - FeedbackStatusBadge
  - GuidelineScopeBadge
  - extend shared/parts.ts
  - all badge stories

phase_3_review_comment:
  - ReviewCommentDrawer
  - ReviewCommentInline
  - GuidelinePicker
  - extend BatchActionBar with onRejectExplain
  - extend ReviewQueuePage with drawer
  - extend AnnotationDetail with inline comment
  - stories for all above

phase_4_run_guidelines:
  - GuidelineLinkPicker modal
  - RunGuidelineSection
  - RunFeedbackSection
  - FeedbackCard
  - extend RunDetailPage
  - stories for all above

phase_5_guideline_pages:
  - GuidelineSummaryCard
  - GuidelineForm
  - GuidelineLinkedRuns
  - GuidelinesListPage
  - GuidelineDetailPage
  - extend AnnotationSidebar
  - extend App.tsx routes
  - stories for all above

phase_6_mailbox_integration:
  - add MailboxBadge to AnnotationTable, AnnotationDetail, MessagePreviewTable
  - add mailbox filter pills to ReviewQueuePage
  - verify all API payloads include mailboxName where specified
```

---

## 11. Design Decisions

| Decision | Rationale |
|----------|-----------|
| Drawer instead of modal for batch comment | Keeps context visible; reviewer can still see selected items |
| Inline panel for single-dismiss | Avoids full-page context switch for a quick comment |
| "Just Dismiss" fast path | Not every dismissal needs a reason; must not slow down power users |
| Separate FeedbackCard from AnnotationLog | Different semantics, different lifecycle, different query needs |
| GuidelineLinkPicker as modal | Selecting from a list is inherently modal; backdrop focuses attention |
| Live markdown preview in editor | Reviewers write markdown; they need to see rendered output before saving |
| MailboxBadge only when non-empty | Most contexts have a single mailbox; badge adds noise when uniform |
| No optimistic feedback creation | Feedback includes server-generated ID and targets; too risky for v1 |
| Guideline runs viewed as list, not embedded | Run detail is a separate page; avoid deep nesting |
