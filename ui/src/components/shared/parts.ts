/**
 * Single source of truth for data-part attribute names
 * used by shared widgets. Keeps selectors stable for
 * theming and testing.
 */
export const parts = {
  tagChip: "tag-chip",
  reviewBadge: "review-badge",
  sourceBadge: "source-badge",
  targetLink: "target-link",
  statBox: "stat-box",
  reviewProgress: "review-progress",
  batchBar: "batch-bar",
  filterPills: "filter-pills",
  countSummary: "count-summary",
  markdownBody: "markdown-body",
  mailboxBadge: "mailbox-badge",
  feedbackStatusBadge: "feedback-status-badge",
  feedbackKindBadge: "feedback-kind-badge",
  guidelineScopeBadge: "guideline-scope-badge",
} as const;
