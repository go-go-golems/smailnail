import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import type {
  Annotation,
  AnnotationFilter,
  TargetGroup,
  GroupFilter,
  GroupDetail,
  AnnotationLog,
  LogFilter,
  AgentRunSummary,
  AgentRunDetail,
  SenderRow,
  SenderFilter,
  SenderDetail,
  SavedQuery,
  QueryResult,
} from "../types/annotations";

export const annotationsApi = createApi({
  reducerPath: "annotationsApi",
  baseQuery: fetchBaseQuery({ baseUrl: "/api" }),
  tagTypes: ["Annotations", "Groups", "Logs", "Runs", "Senders", "Queries"],
  endpoints: (builder) => ({
    // ── Annotations ──────────────────────────────
    listAnnotations: builder.query<Annotation[], AnnotationFilter>({
      query: (filter) => ({ url: "annotations", params: filter }),
      providesTags: ["Annotations"],
    }),
    getAnnotation: builder.query<Annotation, string>({
      query: (id) => `annotations/${id}`,
    }),
    reviewAnnotation: builder.mutation<
      Annotation,
      { id: string; reviewState: string }
    >({
      query: ({ id, reviewState }) => ({
        url: `annotations/${id}/review`,
        method: "PATCH",
        body: { reviewState },
      }),
      invalidatesTags: ["Annotations", "Runs"],
    }),
    batchReview: builder.mutation<
      void,
      { ids: string[]; reviewState: string }
    >({
      query: (body) => ({
        url: "annotations/batch-review",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Annotations", "Runs"],
    }),

    // ── Groups ───────────────────────────────────
    listGroups: builder.query<TargetGroup[], GroupFilter>({
      query: (filter) => ({ url: "annotation-groups", params: filter }),
      providesTags: ["Groups"],
    }),
    getGroup: builder.query<GroupDetail, string>({
      query: (id) => `annotation-groups/${id}`,
    }),

    // ── Logs ─────────────────────────────────────
    listLogs: builder.query<AnnotationLog[], LogFilter>({
      query: (filter) => ({ url: "annotation-logs", params: filter }),
      providesTags: ["Logs"],
    }),
    getLog: builder.query<AnnotationLog, string>({
      query: (id) => `annotation-logs/${id}`,
    }),

    // ── Runs (aggregated) ────────────────────────
    listRuns: builder.query<AgentRunSummary[], void>({
      query: () => "annotation-runs",
      providesTags: ["Runs"],
    }),
    getRun: builder.query<AgentRunDetail, string>({
      query: (id) => `annotation-runs/${id}`,
    }),

    // ── Senders ──────────────────────────────────
    listSenders: builder.query<SenderRow[], SenderFilter>({
      query: (filter) => ({ url: "mirror/senders", params: filter }),
      providesTags: ["Senders"],
    }),
    getSender: builder.query<SenderDetail, string>({
      query: (email) => `mirror/senders/${encodeURIComponent(email)}`,
    }),

    // ── Query Editor ─────────────────────────────
    executeQuery: builder.mutation<QueryResult, { sql: string }>({
      query: (body) => ({ url: "query/execute", method: "POST", body }),
    }),
    getPresets: builder.query<SavedQuery[], void>({
      query: () => "query/presets",
    }),
    getSavedQueries: builder.query<SavedQuery[], void>({
      query: () => "query/saved",
      providesTags: ["Queries"],
    }),
    saveQuery: builder.mutation<
      SavedQuery,
      { name: string; folder: string; description: string; sql: string }
    >({
      query: (body) => ({ url: "query/saved", method: "POST", body }),
      invalidatesTags: ["Queries"],
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
} = annotationsApi;
