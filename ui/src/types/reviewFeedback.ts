/**
 * TypeScript types for review feedback (reviewer-to-agent workflow).
 * Matches the Go backend review_feedback + review_feedback_targets tables.
 */

// ── Enums ────────────────────────────────────────────────────

export type FeedbackScopeKind =
  | "annotation"
  | "selection"
  | "run"
  | "guideline";

export type FeedbackKind =
  | "comment"
  | "reject_request"
  | "guideline_request"
  | "clarification";

export type FeedbackStatus =
  | "open"
  | "acknowledged"
  | "resolved"
  | "archived";

// ── Core entities ────────────────────────────────────────────

export interface FeedbackTarget {
  targetType: string;
  targetId: string;
}

export interface ReviewFeedback {
  id: string;
  scopeKind: FeedbackScopeKind;
  agentRunId: string;
  mailboxName: string;
  feedbackKind: FeedbackKind;
  status: FeedbackStatus;
  title: string;
  bodyMarkdown: string;
  createdBy: string;
  createdAt: string; // ISO 8601
  updatedAt: string;
  targets: FeedbackTarget[];
}

// ── Request types ────────────────────────────────────────────

export interface CreateFeedbackRequest {
  scopeKind: FeedbackScopeKind;
  agentRunId?: string;
  mailboxName?: string;
  feedbackKind: FeedbackKind;
  title: string;
  bodyMarkdown: string;
  targetIds?: string[]; // annotation IDs for scope_kind=selection
}

export interface UpdateFeedbackRequest {
  status?: FeedbackStatus;
  bodyMarkdown?: string;
}

/** Draft comment attached to a review action */
export interface ReviewCommentDraft {
  feedbackKind: FeedbackKind;
  title: string;
  bodyMarkdown: string;
}

// ── Filter types ─────────────────────────────────────────────

export interface FeedbackFilter {
  agentRunId?: string;
  status?: FeedbackStatus;
  feedbackKind?: FeedbackKind;
}
