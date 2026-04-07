import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import type {
  Annotation,
  AnnotationFilter,
  AnnotationListResponse,
  TargetGroup,
  GroupFilter,
  GroupListResponse,
  GroupDetail,
  AnnotationLog,
  LogFilter,
  LogListResponse,
  AgentRunSummary,
  AgentRunListResponse,
  AgentRunDetail,
  SenderRow,
  SenderFilter,
  SenderListResponse,
  SenderDetail,
  SavedQuery,
  SavedQueryListResponse,
  QueryResult,
  ExecuteQueryRequest,
  SaveQueryRequest,
} from "../types/annotations";
import type {
  ReviewFeedback,
  ReviewFeedbackListResponse,
  CreateFeedbackRequest,
  UpdateFeedbackRequest,
  FeedbackFilter,
  ReviewCommentDraft,
} from "../types/reviewFeedback";
import type {
  ReviewGuideline,
  ReviewGuidelineListResponse,
  CreateGuidelineRequest,
  UpdateGuidelineRequest,
  GuidelineFilter,
} from "../types/reviewGuideline";
import type {
  BatchReviewRequest,
  LinkRunGuidelineRequest,
  ReviewAnnotationRequest,
} from "../gen/smailnail/annotationui/v1/review";

export const annotationsApi = createApi({
  reducerPath: "annotationsApi",
  baseQuery: fetchBaseQuery({ baseUrl: "/api" }),
  tagTypes: [
    "Annotations",
    "Groups",
    "Logs",
    "Runs",
    "Senders",
    "Queries",
    "Feedback",
    "Guidelines",
  ],
  endpoints: (builder) => ({
    // ── Annotations ──────────────────────────────
    listAnnotations: builder.query<Annotation[], AnnotationFilter>({
      query: (filter) => ({ url: "annotations", params: filter }),
      transformResponse: (response: AnnotationListResponse) => response.items,
      providesTags: ["Annotations"],
    }),
    getAnnotation: builder.query<Annotation, string>({
      query: (id) => `annotations/${id}`,
      providesTags: ["Annotations"],
    }),
    reviewAnnotation: builder.mutation<
      Annotation,
      {
        id: string;
        reviewState: string;
        comment?: ReviewCommentDraft;
        guidelineIds?: string[];
        mailboxName?: string;
      }
    >({
      query: ({ id, reviewState, comment, guidelineIds, mailboxName }) => ({
        url: `annotations/${id}/review`,
        method: "PATCH",
        body: {
          reviewState,
          comment,
          guidelineIds: guidelineIds ?? [],
          mailboxName: mailboxName ?? "",
        } satisfies ReviewAnnotationRequest,
      }),
      invalidatesTags: ["Annotations", "Runs", "Feedback", "Senders"],
    }),
    batchReview: builder.mutation<
      void,
      {
        ids: string[];
        reviewState: string;
        comment?: ReviewCommentDraft;
        guidelineIds?: string[];
        agentRunId?: string;
        mailboxName?: string;
      }
    >({
      query: ({ ids, reviewState, comment, guidelineIds, agentRunId, mailboxName }) => ({
        url: "annotations/batch-review",
        method: "POST",
        body: {
          ids,
          reviewState,
          comment,
          guidelineIds: guidelineIds ?? [],
          agentRunId: agentRunId ?? "",
          mailboxName: mailboxName ?? "",
        } satisfies BatchReviewRequest,
      }),
      invalidatesTags: ["Annotations", "Runs", "Feedback", "Senders"],
    }),

    // ── Groups ───────────────────────────────────
    listGroups: builder.query<TargetGroup[], GroupFilter>({
      query: (filter) => ({ url: "annotation-groups", params: filter }),
      transformResponse: (response: GroupListResponse) => response.items,
      providesTags: ["Groups"],
    }),
    getGroup: builder.query<GroupDetail, string>({
      query: (id) => `annotation-groups/${id}`,
      providesTags: ["Groups"],
    }),

    // ── Logs ─────────────────────────────────────
    listLogs: builder.query<AnnotationLog[], LogFilter>({
      query: (filter) => ({ url: "annotation-logs", params: filter }),
      transformResponse: (response: LogListResponse) => response.items,
      providesTags: ["Logs"],
    }),
    getLog: builder.query<AnnotationLog, string>({
      query: (id) => `annotation-logs/${id}`,
      providesTags: ["Logs"],
    }),

    // ── Runs (aggregated) ────────────────────────
    listRuns: builder.query<AgentRunSummary[], void>({
      query: () => "annotation-runs",
      transformResponse: (response: AgentRunListResponse) => response.items,
      providesTags: ["Runs"],
    }),
    getRun: builder.query<AgentRunDetail, string>({
      query: (id) => `annotation-runs/${id}`,
      providesTags: ["Runs"],
    }),

    // ── Senders ──────────────────────────────────
    listSenders: builder.query<SenderRow[], SenderFilter>({
      query: (filter) => ({ url: "mirror/senders", params: filter }),
      transformResponse: (response: SenderListResponse) => response.items,
      providesTags: ["Senders"],
    }),
    getSender: builder.query<SenderDetail, string>({
      query: (email) => `mirror/senders/${encodeURIComponent(email)}`,
      providesTags: ["Senders"],
    }),

    // ── Query Editor ─────────────────────────────
    executeQuery: builder.mutation<QueryResult, ExecuteQueryRequest>({
      query: (body) => ({
        url: "query/execute",
        method: "POST",
        body: body satisfies ExecuteQueryRequest,
      }),
    }),
    getPresets: builder.query<SavedQuery[], void>({
      query: () => "query/presets",
      transformResponse: (response: SavedQueryListResponse) => response.items,
    }),
    getSavedQueries: builder.query<SavedQuery[], void>({
      query: () => "query/saved",
      transformResponse: (response: SavedQueryListResponse) => response.items,
      providesTags: ["Queries"],
    }),
    saveQuery: builder.mutation<SavedQuery, SaveQueryRequest>({
      query: (body) => ({
        url: "query/saved",
        method: "POST",
        body: body satisfies SaveQueryRequest,
      }),
      invalidatesTags: ["Queries"],
    }),

    // ── Review Feedback ─────────────────────────
    listReviewFeedback: builder.query<ReviewFeedback[], FeedbackFilter>({
      query: (filter) => ({ url: "review-feedback", params: filter }),
      transformResponse: (response: ReviewFeedbackListResponse) => response.items,
      providesTags: ["Feedback"],
    }),
    getReviewFeedback: builder.query<ReviewFeedback, string>({
      query: (id) => `review-feedback/${id}`,
      providesTags: ["Feedback"],
    }),
    createReviewFeedback: builder.mutation<ReviewFeedback, CreateFeedbackRequest>({
      query: (body) => ({ url: "review-feedback", method: "POST", body }),
      invalidatesTags: ["Feedback"],
    }),
    updateReviewFeedback: builder.mutation<
      ReviewFeedback,
      { id: string } & UpdateFeedbackRequest
    >({
      query: ({ id, ...body }) => ({
        url: `review-feedback/${id}`,
        method: "PATCH",
        body,
      }),
      invalidatesTags: ["Feedback"],
    }),

    // ── Review Guidelines ───────────────────────
    listGuidelines: builder.query<ReviewGuideline[], GuidelineFilter>({
      query: (filter) => ({ url: "review-guidelines", params: filter }),
      transformResponse: (response: ReviewGuidelineListResponse) => response.items,
      providesTags: ["Guidelines"],
    }),
    getGuideline: builder.query<ReviewGuideline, string>({
      query: (id) => `review-guidelines/${id}`,
      providesTags: ["Guidelines"],
    }),
    getGuidelineRuns: builder.query<AgentRunSummary[], string>({
      query: (guidelineId) => `review-guidelines/${guidelineId}/runs`,
      transformResponse: (response: AgentRunListResponse) => response.items,
      providesTags: ["Guidelines", "Runs"],
    }),
    createGuideline: builder.mutation<ReviewGuideline, CreateGuidelineRequest>({
      query: (body) => ({ url: "review-guidelines", method: "POST", body }),
      invalidatesTags: ["Guidelines"],
    }),
    updateGuideline: builder.mutation<
      ReviewGuideline,
      { id: string } & UpdateGuidelineRequest
    >({
      query: ({ id, ...body }) => ({
        url: `review-guidelines/${id}`,
        method: "PATCH",
        body,
      }),
      invalidatesTags: ["Guidelines"],
    }),

    // ── Run-Guideline Links ─────────────────────
    getRunGuidelines: builder.query<ReviewGuideline[], string>({
      query: (runId) => `annotation-runs/${runId}/guidelines`,
      transformResponse: (response: ReviewGuidelineListResponse) => response.items,
      providesTags: ["Guidelines", "Runs"],
    }),
    linkGuidelineToRun: builder.mutation<
      ReviewGuideline[],
      { runId: string; guidelineId: string }
    >({
      query: ({ runId, guidelineId }) => ({
        url: `annotation-runs/${runId}/guidelines`,
        method: "POST",
        body: { guidelineId } satisfies LinkRunGuidelineRequest,
      }),
      transformResponse: (response: ReviewGuidelineListResponse) => response.items,
      invalidatesTags: ["Guidelines", "Runs"],
    }),
    unlinkGuidelineFromRun: builder.mutation<
      void,
      { runId: string; guidelineId: string }
    >({
      query: ({ runId, guidelineId }) => ({
        url: `annotation-runs/${runId}/guidelines/${guidelineId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["Guidelines", "Runs"],
    }),
  }),
});

export const {
  useListAnnotationsQuery,
  useGetAnnotationQuery,
  useReviewAnnotationMutation,
  useBatchReviewMutation,
  useListGroupsQuery,
  useGetGroupQuery,
  useListLogsQuery,
  useGetLogQuery,
  useListRunsQuery,
  useGetRunQuery,
  useListSendersQuery,
  useGetSenderQuery,
  useExecuteQueryMutation,
  useGetPresetsQuery,
  useGetSavedQueriesQuery,
  useSaveQueryMutation,
  // ── Review Feedback hooks
  useListReviewFeedbackQuery,
  useGetReviewFeedbackQuery,
  useCreateReviewFeedbackMutation,
  useUpdateReviewFeedbackMutation,
  // ── Review Guidelines hooks
  useListGuidelinesQuery,
  useGetGuidelineQuery,
  useGetGuidelineRunsQuery,
  useCreateGuidelineMutation,
  useUpdateGuidelineMutation,
  // ── Run-Guideline Links hooks
  useGetRunGuidelinesQuery,
  useLinkGuidelineToRunMutation,
  useUnlinkGuidelineFromRunMutation,
} = annotationsApi;
