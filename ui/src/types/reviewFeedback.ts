import type {
  CreateFeedbackRequest as GeneratedCreateFeedbackRequest,
  FeedbackTarget as GeneratedFeedbackTarget,
  ReviewComment as GeneratedReviewComment,
  ReviewFeedback as GeneratedReviewFeedback,
  UpdateFeedbackRequest as GeneratedUpdateFeedbackRequest,
} from "../gen/smailnail/annotationui/v1/review";

export const FEEDBACK_SCOPE_KIND_VALUES = [
  "annotation",
  "selection",
  "run",
  "guideline",
] as const;

export type FeedbackScopeKind = (typeof FEEDBACK_SCOPE_KIND_VALUES)[number];

export const FEEDBACK_KIND_VALUES = [
  "comment",
  "reject_request",
  "guideline_request",
  "clarification",
] as const;

export type FeedbackKind = (typeof FEEDBACK_KIND_VALUES)[number];

export const FEEDBACK_STATUS_VALUES = [
  "open",
  "acknowledged",
  "resolved",
  "archived",
] as const;

export type FeedbackStatus = (typeof FEEDBACK_STATUS_VALUES)[number];

export type FeedbackTarget = GeneratedFeedbackTarget;

export type ReviewFeedback = Omit<
  GeneratedReviewFeedback,
  "scopeKind" | "feedbackKind" | "status"
> & {
  scopeKind: FeedbackScopeKind;
  feedbackKind: FeedbackKind;
  status: FeedbackStatus;
};

export type ReviewFeedbackListResponse = {
  items: ReviewFeedback[];
};

export type CreateFeedbackRequest = Omit<
  GeneratedCreateFeedbackRequest,
  "scopeKind" | "feedbackKind" | "mailboxName" | "targets"
> & {
  scopeKind: FeedbackScopeKind;
  feedbackKind: FeedbackKind;
  mailboxName?: string;
  targets?: FeedbackTarget[];
};

export type UpdateFeedbackRequest = Omit<GeneratedUpdateFeedbackRequest, "status"> & {
  status?: FeedbackStatus;
};

export type ReviewCommentDraft = Omit<GeneratedReviewComment, "feedbackKind"> & {
  feedbackKind: FeedbackKind;
};

export interface FeedbackFilter {
  scopeKind?: FeedbackScopeKind;
  agentRunId?: string;
  status?: FeedbackStatus;
  feedbackKind?: FeedbackKind;
  mailboxName?: string;
  targetType?: string;
  targetId?: string;
}
