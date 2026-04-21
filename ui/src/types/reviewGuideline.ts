import type {
  CreateGuidelineRequest as GeneratedCreateGuidelineRequest,
  ReviewGuideline as GeneratedReviewGuideline,
  UpdateGuidelineRequest as GeneratedUpdateGuidelineRequest,
} from "../gen/smailnail/annotationui/v1/review";

export const GUIDELINE_SCOPE_KIND_VALUES = [
  "global",
  "mailbox",
  "sender",
  "domain",
  "workflow",
] as const;

export type GuidelineScopeKind = (typeof GUIDELINE_SCOPE_KIND_VALUES)[number];

export const GUIDELINE_STATUS_VALUES = [
  "active",
  "archived",
  "draft",
] as const;

export type GuidelineStatus = (typeof GUIDELINE_STATUS_VALUES)[number];

export type ReviewGuideline = Omit<
  GeneratedReviewGuideline,
  "scopeKind" | "status"
> & {
  scopeKind: GuidelineScopeKind;
  status: GuidelineStatus;
};

export type ReviewGuidelineListResponse = {
  items: ReviewGuideline[];
};

export type CreateGuidelineRequest = Omit<GeneratedCreateGuidelineRequest, "scopeKind"> & {
  scopeKind: GuidelineScopeKind;
};

export type UpdateGuidelineRequest = Omit<
  GeneratedUpdateGuidelineRequest,
  "scopeKind" | "status"
> & {
  scopeKind?: GuidelineScopeKind;
  status?: GuidelineStatus;
};

export interface GuidelineFilter {
  status?: GuidelineStatus;
  scopeKind?: GuidelineScopeKind;
  search?: string;
}
