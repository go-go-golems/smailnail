/**
 * TypeScript types matching the Go backend annotation types.
 * Field names use camelCase (JSON tags from Go structs).
 */

// ── Source & Review constants ────────────────────────────────

export type SourceKind = "human" | "agent" | "heuristic" | "import";
export type ReviewState = "to_review" | "reviewed" | "dismissed";

// ── Core entities ────────────────────────────────────────────

export interface Annotation {
  id: string;
  targetType: string;
  targetId: string;
  tag: string;
  noteMarkdown: string;
  sourceKind: SourceKind;
  sourceLabel: string;
  agentRunId: string;
  reviewState: ReviewState;
  createdBy: string;
  createdAt: string; // ISO 8601
  updatedAt: string;
}

export interface TargetGroup {
  id: string;
  name: string;
  description: string;
  sourceKind: SourceKind;
  sourceLabel: string;
  agentRunId: string;
  reviewState: ReviewState;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface GroupMember {
  groupId: string;
  targetType: string;
  targetId: string;
  addedAt: string;
}

export interface AnnotationLog {
  id: string;
  logKind: string;
  title: string;
  bodyMarkdown: string;
  sourceKind: SourceKind;
  sourceLabel: string;
  agentRunId: string;
  createdBy: string;
  createdAt: string;
}

export interface LogTarget {
  logId: string;
  targetType: string;
  targetId: string;
}

// ── Aggregated / computed types (API responses) ──────────────

export interface AgentRunSummary {
  runId: string;
  sourceLabel: string;
  sourceKind: SourceKind;
  annotationCount: number;
  pendingCount: number;
  reviewedCount: number;
  dismissedCount: number;
  logCount: number;
  groupCount: number;
  startedAt: string;
  completedAt: string;
}

export interface AgentRunDetail extends AgentRunSummary {
  annotations: Annotation[];
  logs: AnnotationLog[];
  groups: TargetGroup[];
}

export interface GroupDetail extends TargetGroup {
  members: GroupMember[];
}

export interface SenderRow {
  email: string;
  displayName: string;
  domain: string;
  messageCount: number;
  annotationCount: number;
  tags: string[];
  hasUnsubscribe: boolean;
}

export interface SenderDetail extends SenderRow {
  firstSeen: string;
  lastSeen: string;
  annotations: Annotation[];
  logs: AnnotationLog[];
  recentMessages: MessagePreview[];
}

export interface MessagePreview {
  uid: number;
  subject: string;
  date: string;
  sizeBytes: number;
}

// ── Filter types ─────────────────────────────────────────────

export interface AnnotationFilter {
  targetType?: string;
  targetId?: string;
  tag?: string;
  reviewState?: ReviewState;
  sourceKind?: SourceKind;
  agentRunId?: string;
  limit?: number;
}

export interface GroupFilter {
  reviewState?: ReviewState;
  sourceKind?: SourceKind;
  limit?: number;
}

export interface LogFilter {
  sourceKind?: SourceKind;
  agentRunId?: string;
  limit?: number;
}

export interface SenderFilter {
  domain?: string;
  hasAnnotations?: boolean;
  tag?: string;
  limit?: number;
}

// ── Query editor types ───────────────────────────────────────

export interface SavedQuery {
  name: string;
  folder: string;
  description: string;
  sql: string;
}

export interface QueryResult {
  columns: string[];
  rows: Record<string, unknown>[];
  durationMs: number;
  rowCount: number;
}

export interface QueryError {
  message: string;
}
