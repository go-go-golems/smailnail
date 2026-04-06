import type {
  AgentRunDetail as GeneratedAgentRunDetail,
  AgentRunSummary as GeneratedAgentRunSummary,
  Annotation as GeneratedAnnotation,
  AnnotationListRequest as GeneratedAnnotationListRequest,
  AnnotationLog as GeneratedAnnotationLog,
  ExecuteQueryRequest as GeneratedExecuteQueryRequest,
  GroupDetail as GeneratedGroupDetail,
  GroupListRequest as GeneratedGroupListRequest,
  GroupMember as GeneratedGroupMember,
  LogListRequest as GeneratedLogListRequest,
  MessagePreview as GeneratedMessagePreview,
  QueryResult as GeneratedQueryResult,
  SavedQuery as GeneratedSavedQuery,
  SaveQueryRequest as GeneratedSaveQueryRequest,
  SenderDetail as GeneratedSenderDetail,
  SenderListRequest as GeneratedSenderListRequest,
  SenderRow as GeneratedSenderRow,
  TargetGroup as GeneratedTargetGroup,
} from "../gen/smailnail/annotationui/v1/annotation";

export const SOURCE_KIND_VALUES = ["human", "agent", "heuristic", "import"] as const;
export type SourceKind = (typeof SOURCE_KIND_VALUES)[number];

export const REVIEW_STATE_VALUES = ["to_review", "reviewed", "dismissed"] as const;
export type ReviewState = (typeof REVIEW_STATE_VALUES)[number];

export type Annotation = Omit<GeneratedAnnotation, "sourceKind" | "reviewState"> & {
  sourceKind: SourceKind;
  reviewState: ReviewState;
};

export type AnnotationListResponse = {
  items: Annotation[];
};

export type TargetGroup = Omit<GeneratedTargetGroup, "sourceKind" | "reviewState"> & {
  sourceKind: SourceKind;
  reviewState: ReviewState;
};

export type GroupListResponse = {
  items: TargetGroup[];
};

export type GroupMember = GeneratedGroupMember;

export type GroupDetail = Omit<
  GeneratedGroupDetail,
  "sourceKind" | "reviewState" | "members"
> & {
  sourceKind: SourceKind;
  reviewState: ReviewState;
  members: GroupMember[];
};

export type AnnotationLog = Omit<GeneratedAnnotationLog, "sourceKind"> & {
  sourceKind: SourceKind;
};

export type LogListResponse = {
  items: AnnotationLog[];
};

export type AgentRunSummary = Omit<GeneratedAgentRunSummary, "sourceKind"> & {
  sourceKind: SourceKind;
};

export type AgentRunListResponse = {
  items: AgentRunSummary[];
};

export type AgentRunDetail = Omit<
  GeneratedAgentRunDetail,
  "sourceKind" | "annotations" | "logs" | "groups"
> & {
  sourceKind: SourceKind;
  annotations: Annotation[];
  logs: AnnotationLog[];
  groups: TargetGroup[];
};

export type SenderRow = GeneratedSenderRow;

export type SenderListResponse = {
  items: SenderRow[];
};

export type MessagePreview = GeneratedMessagePreview;

export type SenderDetail = Omit<GeneratedSenderDetail, "annotations" | "logs" | "recentMessages"> & {
  annotations: Annotation[];
  logs: AnnotationLog[];
  recentMessages: MessagePreview[];
};

export type AnnotationFilter = Partial<
  Omit<GeneratedAnnotationListRequest, "reviewState" | "sourceKind">
> & {
  reviewState?: ReviewState;
  sourceKind?: SourceKind;
  mailboxName?: string;
  feedbackStatus?: string;
};

export type GroupFilter = Partial<Omit<GeneratedGroupListRequest, "reviewState" | "sourceKind">> & {
  reviewState?: ReviewState;
  sourceKind?: SourceKind;
};

export type LogFilter = Partial<Omit<GeneratedLogListRequest, "sourceKind">> & {
  sourceKind?: SourceKind;
};

export type SenderFilter = Partial<GeneratedSenderListRequest>;

export type SavedQuery = GeneratedSavedQuery;

export type SavedQueryListResponse = {
  items: SavedQuery[];
};

export type ExecuteQueryRequest = GeneratedExecuteQueryRequest;

export type SaveQueryRequest = GeneratedSaveQueryRequest;

export type QueryResult = Omit<GeneratedQueryResult, "rows"> & {
  rows: Record<string, unknown>[];
};

export interface QueryError {
  message: string;
}
