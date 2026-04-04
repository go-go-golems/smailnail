/**
 * TypeScript types for review guidelines (reusable reviewer policy).
 * Matches the Go backend review_guidelines + run_guideline_links tables.
 */

// ── Enums ────────────────────────────────────────────────────

export type GuidelineScopeKind =
  | "global"
  | "mailbox"
  | "sender"
  | "domain"
  | "workflow";

export type GuidelineStatus =
  | "active"
  | "archived"
  | "draft";

// ── Core entity ──────────────────────────────────────────────

export interface ReviewGuideline {
  id: string;
  slug: string;
  title: string;
  scopeKind: GuidelineScopeKind;
  status: GuidelineStatus;
  priority: number;
  bodyMarkdown: string;
  createdBy: string;
  createdAt: string; // ISO 8601
  updatedAt: string;
}

// ── Request types ────────────────────────────────────────────

export interface CreateGuidelineRequest {
  slug: string;
  title: string;
  scopeKind: GuidelineScopeKind;
  bodyMarkdown: string;
}

export interface UpdateGuidelineRequest {
  title?: string;
  scopeKind?: GuidelineScopeKind;
  status?: GuidelineStatus;
  priority?: number;
  bodyMarkdown?: string;
}

// ── Filter types ─────────────────────────────────────────────

export interface GuidelineFilter {
  status?: GuidelineStatus;
  scopeKind?: GuidelineScopeKind;
  search?: string;
}
