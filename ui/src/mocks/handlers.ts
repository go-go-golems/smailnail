import { http, HttpResponse } from "msw";
import {
  mockAnnotations,
  mockLogs,
  mockRuns,
  mockGroups,
  mockGroupMembers,
  mockSenders,
  mockMessages,
  mockPresets,
  mockQueryResult,
} from "./annotations";

export const handlers = [
  // ── Annotations ──────────────────────────────
  http.get("/api/annotations", ({ request }) => {
    const url = new URL(request.url);
    let result = [...mockAnnotations];

    const tag = url.searchParams.get("tag");
    if (tag) result = result.filter((a) => a.tag === tag);

    const reviewState = url.searchParams.get("reviewState");
    if (reviewState) result = result.filter((a) => a.reviewState === reviewState);

    const sourceKind = url.searchParams.get("sourceKind");
    if (sourceKind) result = result.filter((a) => a.sourceKind === sourceKind);

    const agentRunId = url.searchParams.get("agentRunId");
    if (agentRunId) result = result.filter((a) => a.agentRunId === agentRunId);

    return HttpResponse.json(result);
  }),

  http.get("/api/annotations/:id", ({ params }) => {
    const ann = mockAnnotations.find((a) => a.id === params["id"]);
    if (!ann) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(ann);
  }),

  http.patch("/api/annotations/:id/review", async ({ params, request }) => {
    const body = (await request.json()) as { reviewState: string };
    const ann = mockAnnotations.find((a) => a.id === params["id"]);
    if (!ann) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json({ ...ann, reviewState: body.reviewState });
  }),

  http.post("/api/annotations/batch-review", async () => {
    return HttpResponse.json(null, { status: 204 });
  }),

  // ── Groups ───────────────────────────────────
  http.get("/api/annotation-groups", () => {
    return HttpResponse.json(mockGroups);
  }),

  http.get("/api/annotation-groups/:id", ({ params }) => {
    const group = mockGroups.find((g) => g.id === params["id"]);
    if (!group) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const members = mockGroupMembers.filter((m) => m.groupId === group.id);
    return HttpResponse.json({ ...group, members });
  }),

  // ── Logs ─────────────────────────────────────
  http.get("/api/annotation-logs", ({ request }) => {
    const url = new URL(request.url);
    let result = [...mockLogs];

    const agentRunId = url.searchParams.get("agentRunId");
    if (agentRunId) result = result.filter((l) => l.agentRunId === agentRunId);

    return HttpResponse.json(result);
  }),

  http.get("/api/annotation-logs/:id", ({ params }) => {
    const log = mockLogs.find((l) => l.id === params["id"]);
    if (!log) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(log);
  }),

  // ── Runs ─────────────────────────────────────
  http.get("/api/annotation-runs", () => {
    return HttpResponse.json(mockRuns);
  }),

  http.get("/api/annotation-runs/:id", ({ params }) => {
    const run = mockRuns.find((r) => r.runId === params["id"]);
    if (!run) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const annotations = mockAnnotations.filter((a) => a.agentRunId === run.runId);
    const logs = mockLogs.filter((l) => l.agentRunId === run.runId);
    const groups = mockGroups.filter((g) => g.agentRunId === run.runId);
    return HttpResponse.json({ ...run, annotations, logs, groups });
  }),

  // ── Senders ──────────────────────────────────
  http.get("/api/mirror/senders", () => {
    return HttpResponse.json(mockSenders);
  }),

  http.get("/api/mirror/senders/:email", ({ params }) => {
    const sender = mockSenders.find((s) => s.email === params["email"]);
    if (!sender) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const annotations = mockAnnotations.filter(
      (a) => a.targetType === "sender" && a.targetId === sender.email,
    );
    const logs = mockLogs.filter((l) =>
      annotations.some((a) => a.agentRunId === l.agentRunId),
    );
    return HttpResponse.json({
      ...sender,
      firstSeen: "2025-01-15T00:00:00Z",
      lastSeen: "2026-04-01T08:00:00Z",
      annotations,
      logs,
      recentMessages: mockMessages,
    });
  }),

  // ── Query Editor ─────────────────────────────
  http.post("/api/query/execute", async ({ request }) => {
    const body = (await request.json()) as { sql: string };
    if (body.sql.toLowerCase().includes("error")) {
      return HttpResponse.json(
        { message: 'Error: Referenced column "error" not found' },
        { status: 400 },
      );
    }
    return HttpResponse.json(mockQueryResult);
  }),

  http.get("/api/query/presets", () => {
    return HttpResponse.json(mockPresets);
  }),

  http.get("/api/query/saved", () => {
    return HttpResponse.json([]);
  }),

  http.post("/api/query/saved", async ({ request }) => {
    const body = (await request.json()) as {
      name: string;
      folder: string;
      description: string;
      sql: string;
    };
    return HttpResponse.json(body, { status: 201 });
  }),
];
