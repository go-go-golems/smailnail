/**
 * Realistic mock data for Storybook stories and development.
 * Based on the JSX design sketch patterns.
 */
import type {
  Annotation,
  AnnotationLog,
  AgentRunSummary,
  TargetGroup,
  GroupMember,
  SenderRow,
  MessagePreview,
  SavedQuery,
  QueryResult,
} from "../types/annotations";

// ── Annotations ──────────────────────────────────────────────

export const mockAnnotations: Annotation[] = [
  {
    id: "ann-001",
    targetType: "sender",
    targetId: "news@techcrunch.com",
    tag: "newsletter",
    noteMarkdown: "Regular tech newsletter, automated delivery pattern",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "to_review",
    createdBy: "system",
    createdAt: "2026-04-01T10:30:00Z",
    updatedAt: "2026-04-01T10:30:00Z",
  },
  {
    id: "ann-002",
    targetType: "sender",
    targetId: "noreply@github.com",
    tag: "notification",
    noteMarkdown:
      "GitHub notifications — PR reviews, issue mentions. High volume but valuable. Consider filter rules instead of ignoring.",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "to_review",
    createdBy: "system",
    createdAt: "2026-04-01T10:31:00Z",
    updatedAt: "2026-04-01T10:31:00Z",
  },
  {
    id: "ann-003",
    targetType: "sender",
    targetId: "promotions@store.example.com",
    tag: "marketing",
    noteMarkdown: "Marketing emails from an online store, unsubscribe available",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "reviewed",
    createdBy: "system",
    createdAt: "2026-04-01T10:32:00Z",
    updatedAt: "2026-04-02T08:00:00Z",
  },
  {
    id: "ann-004",
    targetType: "sender",
    targetId: "alice@company.com",
    tag: "important",
    noteMarkdown: "Direct colleague, always important",
    sourceKind: "human",
    sourceLabel: "manuel",
    agentRunId: "",
    reviewState: "reviewed",
    createdBy: "manuel",
    createdAt: "2026-03-28T14:00:00Z",
    updatedAt: "2026-03-28T14:00:00Z",
  },
  {
    id: "ann-005",
    targetType: "sender",
    targetId: "updates@linkedin.com",
    tag: "bulk-sender",
    noteMarkdown: "LinkedIn update emails — low value, high volume",
    sourceKind: "heuristic",
    sourceLabel: "volume-detector",
    agentRunId: "",
    reviewState: "dismissed",
    createdBy: "system",
    createdAt: "2026-04-01T09:00:00Z",
    updatedAt: "2026-04-02T09:00:00Z",
  },
  {
    id: "ann-006",
    targetType: "sender",
    targetId: "billing@aws.amazon.com",
    tag: "transactional",
    noteMarkdown: "AWS billing notifications — keep for records",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-41",
    reviewState: "to_review",
    createdBy: "system",
    createdAt: "2026-03-30T15:00:00Z",
    updatedAt: "2026-03-30T15:00:00Z",
  },
  {
    id: "ann-007",
    targetType: "domain",
    targetId: "mailchimp.com",
    tag: "bulk-sender",
    noteMarkdown: "All senders from mailchimp.com domain are bulk/marketing",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "to_review",
    createdBy: "system",
    createdAt: "2026-04-01T10:35:00Z",
    updatedAt: "2026-04-01T10:35:00Z",
  },
];

// ── Annotation Logs ──────────────────────────────────────────

export const mockLogs: AnnotationLog[] = [
  {
    id: "log-001",
    logKind: "reasoning",
    title: "Sender classification: news@techcrunch.com",
    bodyMarkdown:
      "Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: `TechCrunch Daily - {date}`. Classified as **newsletter**.",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    createdBy: "system",
    createdAt: "2026-04-01T10:30:00Z",
  },
  {
    id: "log-002",
    logKind: "reasoning",
    title: "Sender classification: noreply@github.com",
    bodyMarkdown:
      "GitHub notifications sender. 312 messages in archive. Mix of PR reviews, issue updates, and CI notifications. Tagged as **notification** rather than bulk-sender because content is personalized.",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    createdBy: "system",
    createdAt: "2026-04-01T10:31:00Z",
  },
  {
    id: "log-003",
    logKind: "decision",
    title: "Run started — triage batch",
    bodyMarkdown:
      "Starting triage run for 23 unclassified senders. Using volume heuristics + content analysis.",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    createdBy: "system",
    createdAt: "2026-04-01T10:29:00Z",
  },
  {
    id: "log-004",
    logKind: "summary",
    title: "Run complete — 23 senders classified",
    bodyMarkdown:
      "Classified 23 senders:\n- 8 newsletters\n- 6 notifications\n- 5 bulk-sender\n- 3 transactional\n- 1 important\n\nAll annotations created with review_state=to_review.",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    createdBy: "system",
    createdAt: "2026-04-01T10:40:00Z",
  },
];

// ── Agent Runs ───────────────────────────────────────────────

export const mockRuns: AgentRunSummary[] = [
  {
    runId: "run-42",
    sourceLabel: "triage-agent-v2",
    sourceKind: "agent",
    annotationCount: 23,
    pendingCount: 18,
    reviewedCount: 3,
    dismissedCount: 2,
    logCount: 4,
    groupCount: 2,
    startedAt: "2026-04-01T10:29:00Z",
    completedAt: "2026-04-01T10:40:00Z",
  },
  {
    runId: "run-41",
    sourceLabel: "triage-agent-v2",
    sourceKind: "agent",
    annotationCount: 15,
    pendingCount: 3,
    reviewedCount: 10,
    dismissedCount: 2,
    logCount: 3,
    groupCount: 1,
    startedAt: "2026-03-30T14:00:00Z",
    completedAt: "2026-03-30T14:15:00Z",
  },
  {
    runId: "run-40",
    sourceLabel: "volume-detector",
    sourceKind: "heuristic",
    annotationCount: 8,
    pendingCount: 0,
    reviewedCount: 5,
    dismissedCount: 3,
    logCount: 1,
    groupCount: 0,
    startedAt: "2026-03-28T09:00:00Z",
    completedAt: "2026-03-28T09:02:00Z",
  },
];

// ── Groups ───────────────────────────────────────────────────

export const mockGroups: TargetGroup[] = [
  {
    id: "grp-001",
    name: "Tech Newsletter Senders",
    description:
      "Senders identified as tech-focused newsletter publishers with regular automated delivery",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "to_review",
    createdBy: "system",
    createdAt: "2026-04-01T10:35:00Z",
    updatedAt: "2026-04-01T10:35:00Z",
  },
  {
    id: "grp-002",
    name: "CI/CD Notification Senders",
    description: "Automated CI/CD and DevOps notification senders",
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
    agentRunId: "run-42",
    reviewState: "reviewed",
    createdBy: "system",
    createdAt: "2026-04-01T10:36:00Z",
    updatedAt: "2026-04-02T08:00:00Z",
  },
];

export const mockGroupMembers: GroupMember[] = [
  {
    groupId: "grp-001",
    targetType: "sender",
    targetId: "news@techcrunch.com",
    addedAt: "2026-04-01T10:35:00Z",
  },
  {
    groupId: "grp-001",
    targetType: "sender",
    targetId: "newsletter@hackernewsletter.com",
    addedAt: "2026-04-01T10:35:00Z",
  },
  {
    groupId: "grp-002",
    targetType: "sender",
    targetId: "noreply@github.com",
    addedAt: "2026-04-01T10:36:00Z",
  },
];

// ── Senders ──────────────────────────────────────────────────

export const mockSenders: SenderRow[] = [
  {
    email: "news@techcrunch.com",
    displayName: "TechCrunch Daily",
    domain: "techcrunch.com",
    messageCount: 47,
    annotationCount: 1,
    tags: ["newsletter"],
    hasUnsubscribe: true,
  },
  {
    email: "noreply@github.com",
    displayName: "GitHub",
    domain: "github.com",
    messageCount: 312,
    annotationCount: 1,
    tags: ["notification"],
    hasUnsubscribe: true,
  },
  {
    email: "alice@company.com",
    displayName: "Alice Smith",
    domain: "company.com",
    messageCount: 24,
    annotationCount: 1,
    tags: ["important"],
    hasUnsubscribe: false,
  },
  {
    email: "updates@linkedin.com",
    displayName: "LinkedIn",
    domain: "linkedin.com",
    messageCount: 89,
    annotationCount: 1,
    tags: ["bulk-sender"],
    hasUnsubscribe: true,
  },
  {
    email: "billing@aws.amazon.com",
    displayName: "AWS Billing",
    domain: "aws.amazon.com",
    messageCount: 12,
    annotationCount: 1,
    tags: ["transactional"],
    hasUnsubscribe: false,
  },
];

// ── Messages ─────────────────────────────────────────────────

export const mockMessages: MessagePreview[] = [
  {
    uid: 1001,
    subject: "TechCrunch Daily - April 1, 2026",
    date: "2026-04-01T08:00:00Z",
    sizeBytes: 45320,
    mailboxName: "INBOX",
  },
  {
    uid: 1000,
    subject: "TechCrunch Daily - March 31, 2026",
    date: "2026-03-31T08:00:00Z",
    sizeBytes: 42100,
    mailboxName: "INBOX",
  },
  {
    uid: 999,
    subject: "TechCrunch Daily - March 30, 2026",
    date: "2026-03-30T08:00:00Z",
    sizeBytes: 38900,
    mailboxName: "INBOX",
  },
];

// ── Saved Queries ────────────────────────────────────────────

export const mockPresets: SavedQuery[] = [
  {
    name: "annotations-by-tag",
    folder: "annotations",
    description: "Count annotations grouped by tag",
    sql: "SELECT tag, COUNT(*) as count\nFROM annotations\nGROUP BY tag\nORDER BY count DESC;",
  },
  {
    name: "pending-review",
    folder: "annotations",
    description: "All annotations pending review",
    sql: "SELECT a.id, a.target_type, a.target_id, a.tag, a.note_markdown, a.source_label\nFROM annotations a\nWHERE a.review_state = 'to_review'\nORDER BY a.created_at DESC;",
  },
  {
    name: "sender-volume",
    folder: "mirror",
    description: "Top senders by message count",
    sql: "SELECT sender_email, sender_display_name, COUNT(*) as msg_count\nFROM messages\nGROUP BY sender_email\nORDER BY msg_count DESC\nLIMIT 50;",
  },
];

export const mockQueryResult: QueryResult = {
  columns: ["tag", "count"],
  rows: [
    { tag: "newsletter", count: 8 },
    { tag: "notification", count: 6 },
    { tag: "bulk-sender", count: 5 },
    { tag: "transactional", count: 3 },
    { tag: "important", count: 1 },
  ],
  durationMs: 12,
  rowCount: 5,
};
