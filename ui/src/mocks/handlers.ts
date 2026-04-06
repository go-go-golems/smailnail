import { http, HttpResponse } from "msw";
import type {
  FeedbackKind,
  FeedbackScopeKind,
  FeedbackStatus,
} from "../types/reviewFeedback";
import type {
  GuidelineScopeKind,
  GuidelineStatus,
} from "../types/reviewGuideline";
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
  mockFeedback,
  mockGuidelines,
} from "./annotations";

/** Mutable mock state for run-guideline links (survives across requests in Storybook). */
const runGuidelineLinks = new Map<string, Set<string>>([
  ["run-42", new Set(["guideline-001", "guideline-002"])],
]);

const mutableFeedback = [...mockFeedback];
const mutableGuidelines = [...mockGuidelines];

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

    return HttpResponse.json({ items: result });
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
    return HttpResponse.json({ items: mockGroups });
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

    return HttpResponse.json({ items: result });
  }),

  http.get("/api/annotation-logs/:id", ({ params }) => {
    const log = mockLogs.find((l) => l.id === params["id"]);
    if (!log) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(log);
  }),

  // ── Runs ─────────────────────────────────────
  http.get("/api/annotation-runs", () => {
    return HttpResponse.json({ items: mockRuns });
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
    return HttpResponse.json({ items: mockSenders });
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
    return HttpResponse.json({ items: mockPresets });
  }),

  http.get("/api/query/saved", () => {
    return HttpResponse.json({ items: [] });
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

  // ── Review Feedback ──────────────────────────
  http.get("/api/review-feedback", ({ request }) => {
    const url = new URL(request.url);
    let result = [...mutableFeedback];

    const agentRunId = url.searchParams.get("agentRunId");
    if (agentRunId) result = result.filter((f) => f.agentRunId === agentRunId);

    const status = url.searchParams.get("status");
    if (status) result = result.filter((f) => f.status === status);

    const feedbackKind = url.searchParams.get("feedbackKind");
    if (feedbackKind) result = result.filter((f) => f.feedbackKind === feedbackKind);

    const mailboxName = url.searchParams.get("mailboxName");
    if (mailboxName) result = result.filter((f) => f.mailboxName === mailboxName);

    return HttpResponse.json({ items: result });
  }),

  http.post("/api/review-feedback", async ({ request }) => {
    const body = (await request.json()) as {
      scopeKind: string;
      feedbackKind: string;
      title: string;
      bodyMarkdown: string;
      agentRunId?: string;
      mailboxName?: string;
      targets?: Array<{ targetType: string; targetId: string }>;
    };
    const id = `fb-${Date.now()}`;
    const created = {
      id,
      scopeKind: body.scopeKind as FeedbackScopeKind,
      agentRunId: body.agentRunId ?? "",
      mailboxName: body.mailboxName ?? "",
      feedbackKind: body.feedbackKind as FeedbackKind,
      status: "open" as FeedbackStatus,
      title: body.title,
      bodyMarkdown: body.bodyMarkdown,
      createdBy: "manuel",
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      targets: body.targets ?? [],
    };
    mutableFeedback.unshift(created);
    return HttpResponse.json(created, { status: 201 });
  }),

  http.get("/api/review-feedback/:id", ({ params }) => {
    const fb = mutableFeedback.find((f) => f.id === params["id"]);
    if (!fb) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(fb);
  }),

  http.patch("/api/review-feedback/:id", async ({ params, request }) => {
    const index = mutableFeedback.findIndex((f) => f.id === params["id"]);
    if (index === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const body = (await request.json()) as { status?: string };
    const current = mutableFeedback[index]!;
    const updated = {
      ...current,
      status: (body.status ?? current.status) as FeedbackStatus,
      updatedAt: new Date().toISOString(),
    };
    mutableFeedback[index] = updated;
    return HttpResponse.json(updated);
  }),

  // ── Review Guidelines ────────────────────────
  http.get("/api/review-guidelines", ({ request }) => {
    const url = new URL(request.url);
    let result = [...mutableGuidelines];

    const status = url.searchParams.get("status");
    if (status) result = result.filter((g) => g.status === status);

    const scopeKind = url.searchParams.get("scopeKind");
    if (scopeKind) result = result.filter((g) => g.scopeKind === scopeKind);

    const search = url.searchParams.get("search");
    if (search) {
      const q = search.toLowerCase();
      result = result.filter(
        (g) =>
          g.title.toLowerCase().includes(q) ||
          g.slug.toLowerCase().includes(q) ||
          g.bodyMarkdown.toLowerCase().includes(q),
      );
    }

    return HttpResponse.json({ items: result });
  }),

  http.post("/api/review-guidelines", async ({ request }) => {
    const body = (await request.json()) as {
      slug: string;
      title: string;
      scopeKind: string;
      bodyMarkdown: string;
    };
    if (mutableGuidelines.some((g) => g.slug === body.slug)) {
      return HttpResponse.json(
        { error: `Guideline with slug '${body.slug}' already exists` },
        { status: 409 },
      );
    }
    const id = `guideline-${Date.now()}`;
    const created = {
      id,
      slug: body.slug,
      title: body.title,
      scopeKind: body.scopeKind as GuidelineScopeKind,
      status: "active" as GuidelineStatus,
      priority: 0,
      bodyMarkdown: body.bodyMarkdown,
      createdBy: "manuel",
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    mutableGuidelines.unshift(created);
    return HttpResponse.json(created, { status: 201 });
  }),

  http.get("/api/review-guidelines/:id", ({ params }) => {
    const g = mutableGuidelines.find((gl) => gl.id === params["id"]);
    if (!g) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(g);
  }),

  http.patch("/api/review-guidelines/:id", async ({ params, request }) => {
    const index = mutableGuidelines.findIndex((gl) => gl.id === params["id"]);
    if (index === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const body = (await request.json()) as Record<string, unknown>;
    const current = mutableGuidelines[index]!;
    const updated = {
      ...current,
      ...(body.title !== undefined ? { title: body.title as string } : {}),
      ...(body.scopeKind !== undefined ? { scopeKind: body.scopeKind as GuidelineScopeKind } : {}),
      ...(body.status !== undefined ? { status: body.status as GuidelineStatus } : {}),
      ...(body.priority !== undefined ? { priority: body.priority as number } : {}),
      ...(body.bodyMarkdown !== undefined ? { bodyMarkdown: body.bodyMarkdown as string } : {}),
      updatedAt: new Date().toISOString(),
    };
    mutableGuidelines[index] = updated;
    return HttpResponse.json(updated);
  }),

  // ── Run-Guideline Links ──────────────────────
  http.get("/api/annotation-runs/:id/guidelines", ({ params }) => {
    const runId = params["id"] as string;
    const linkedIds = runGuidelineLinks.get(runId);
    if (!linkedIds) return HttpResponse.json({ items: [] });
    const linked = mutableGuidelines.filter((g) => linkedIds.has(g.id));
    return HttpResponse.json({ items: linked });
  }),

  http.post("/api/annotation-runs/:id/guidelines", async ({ params, request }) => {
    const runId = params["id"] as string;
    const body = (await request.json()) as { guidelineId: string };
    if (!runGuidelineLinks.has(runId)) runGuidelineLinks.set(runId, new Set());
    runGuidelineLinks.get(runId)!.add(body.guidelineId);
    const linked = mutableGuidelines.filter((g) => runGuidelineLinks.get(runId)!.has(g.id));
    return HttpResponse.json({ items: linked });
  }),

  http.delete(
    "/api/annotation-runs/:id/guidelines/:guidelineId",
    ({ params }) => {
      const runId = params["id"] as string;
      const guidelineId = params["guidelineId"] as string;
      runGuidelineLinks.get(runId)?.delete(guidelineId);
      return HttpResponse.json(null, { status: 204 });
    },
  ),
];
