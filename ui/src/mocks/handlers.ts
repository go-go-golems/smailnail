import { http, HttpResponse } from "msw";
import type { Annotation } from "../types/annotations";
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

const mutableAnnotations = mockAnnotations.map((annotation) => ({ ...annotation }));
const mutableFeedback = mockFeedback.map((feedback) => ({
  ...feedback,
  targets: feedback.targets.map((target) => ({ ...target })),
}));
const mutableGuidelines = mockGuidelines.map((guideline) => ({ ...guideline }));

function nowISO() {
  return new Date().toISOString();
}

function nextID(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function listRunIDs() {
  return Array.from(
    new Set([
      ...mockRuns.map((run) => run.runId),
      ...mutableAnnotations
        .map((annotation) => annotation.agentRunId)
        .filter((runId) => runId.length > 0),
    ]),
  );
}

function buildRunSummary(runId: string) {
  const baseRun = mockRuns.find((run) => run.runId === runId);
  const runAnnotations = mutableAnnotations.filter((annotation) => annotation.agentRunId === runId);
  if (!baseRun && runAnnotations.length === 0) {
    return null;
  }

  const sourceLabel =
    baseRun?.sourceLabel ??
    runAnnotations[runAnnotations.length - 1]?.sourceLabel ??
    "";
  const sourceKind =
    baseRun?.sourceKind ??
    runAnnotations[runAnnotations.length - 1]?.sourceKind ??
    "";

  return {
    runId,
    sourceLabel,
    sourceKind,
    annotationCount: runAnnotations.length,
    pendingCount: runAnnotations.filter((annotation) => annotation.reviewState === "to_review").length,
    reviewedCount: runAnnotations.filter((annotation) => annotation.reviewState === "reviewed").length,
    dismissedCount: runAnnotations.filter((annotation) => annotation.reviewState === "dismissed").length,
    logCount: mockLogs.filter((log) => log.agentRunId === runId).length,
    groupCount: mockGroups.filter((group) => group.agentRunId === runId).length,
    startedAt: baseRun?.startedAt ?? runAnnotations[0]?.createdAt ?? "",
    completedAt: baseRun?.completedAt ?? runAnnotations[runAnnotations.length - 1]?.updatedAt ?? "",
  };
}

function inferSingleRunID(annotationIDs: string[]) {
  const runIDs = Array.from(
    new Set(
      mutableAnnotations
        .filter((annotation) => annotationIDs.includes(annotation.id))
        .map((annotation) => annotation.agentRunId)
        .filter((runId) => runId.length > 0),
    ),
  );
  return runIDs.length === 1 ? runIDs[0] : undefined;
}

export const handlers = [
  // ── Annotations ──────────────────────────────
  http.get("/api/annotations", ({ request }) => {
    const url = new URL(request.url);
    let result = [...mutableAnnotations];

    const tag = url.searchParams.get("tag");
    if (tag) result = result.filter((annotation) => annotation.tag === tag);

    const reviewState = url.searchParams.get("reviewState");
    if (reviewState) result = result.filter((annotation) => annotation.reviewState === reviewState);

    const sourceKind = url.searchParams.get("sourceKind");
    if (sourceKind) result = result.filter((annotation) => annotation.sourceKind === sourceKind);

    const agentRunId = url.searchParams.get("agentRunId");
    if (agentRunId) result = result.filter((annotation) => annotation.agentRunId === agentRunId);

    return HttpResponse.json({ items: result });
  }),

  http.get("/api/annotations/:id", ({ params }) => {
    const annotation = mutableAnnotations.find((candidate) => candidate.id === params["id"]);
    if (!annotation) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(annotation);
  }),

  http.patch("/api/annotations/:id/review", async ({ params, request }) => {
    const body = (await request.json()) as {
      reviewState: Annotation["reviewState"];
      comment?: { feedbackKind: FeedbackKind; title: string; bodyMarkdown: string };
      guidelineIds?: string[];
      mailboxName?: string;
    };
    const index = mutableAnnotations.findIndex((annotation) => annotation.id === params["id"]);
    if (index === -1) return HttpResponse.json({ error: "not found" }, { status: 404 });

    const current = mutableAnnotations[index]!;
    const updated = {
      ...current,
      reviewState: body.reviewState,
      updatedAt: nowISO(),
    };
    mutableAnnotations[index] = updated;

    if (body.comment?.bodyMarkdown) {
      mutableFeedback.unshift({
        id: nextID("fb"),
        scopeKind: "annotation",
        agentRunId: updated.agentRunId,
        mailboxName: body.mailboxName ?? "",
        feedbackKind: body.comment.feedbackKind,
        status: "open",
        title: body.comment.title,
        bodyMarkdown: body.comment.bodyMarkdown,
        createdBy: "storybook",
        createdAt: nowISO(),
        updatedAt: nowISO(),
        targets: [{ targetType: "annotation", targetId: updated.id }],
      });
    }

    if (body.guidelineIds?.length && updated.agentRunId) {
      if (!runGuidelineLinks.has(updated.agentRunId)) {
        runGuidelineLinks.set(updated.agentRunId, new Set());
      }
      for (const guidelineId of body.guidelineIds) {
        runGuidelineLinks.get(updated.agentRunId)!.add(guidelineId);
      }
    }

    return HttpResponse.json(updated);
  }),

  http.post("/api/annotations/batch-review", async ({ request }) => {
    const body = (await request.json()) as {
      ids: string[];
      reviewState: Annotation["reviewState"];
      comment?: { feedbackKind: FeedbackKind; title: string; bodyMarkdown: string };
      guidelineIds?: string[];
      agentRunId?: string;
      mailboxName?: string;
    };

    for (const id of body.ids) {
      const index = mutableAnnotations.findIndex((annotation) => annotation.id === id);
      if (index === -1) continue;
      mutableAnnotations[index] = {
        ...mutableAnnotations[index]!,
        reviewState: body.reviewState,
        updatedAt: nowISO(),
      };
    }

    const inferredRunID = body.agentRunId || inferSingleRunID(body.ids);

    if (body.comment?.bodyMarkdown) {
      mutableFeedback.unshift({
        id: nextID("fb"),
        scopeKind: "selection",
        agentRunId: inferredRunID ?? "",
        mailboxName: body.mailboxName ?? "",
        feedbackKind: body.comment.feedbackKind,
        status: "open",
        title: body.comment.title,
        bodyMarkdown: body.comment.bodyMarkdown,
        createdBy: "storybook",
        createdAt: nowISO(),
        updatedAt: nowISO(),
        targets: body.ids.map((id) => ({ targetType: "annotation", targetId: id })),
      });
    }

    if (body.guidelineIds?.length && inferredRunID) {
      if (!runGuidelineLinks.has(inferredRunID)) {
        runGuidelineLinks.set(inferredRunID, new Set());
      }
      for (const guidelineId of body.guidelineIds) {
        runGuidelineLinks.get(inferredRunID)!.add(guidelineId);
      }
    }

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
    return HttpResponse.json({
      items: listRunIDs()
        .map((runId) => buildRunSummary(runId))
        .filter((run): run is NonNullable<typeof run> => run !== null),
    });
  }),

  http.get("/api/annotation-runs/:id", ({ params }) => {
    const runId = params["id"] as string;
    const run = buildRunSummary(runId);
    if (!run) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const annotations = mutableAnnotations.filter((annotation) => annotation.agentRunId === runId);
    const logs = mockLogs.filter((log) => log.agentRunId === runId);
    const groups = mockGroups.filter((group) => group.agentRunId === runId);
    return HttpResponse.json({ ...run, annotations, logs, groups });
  }),

  // ── Senders ──────────────────────────────────
  http.get("/api/mirror/senders", () => {
    const items = mockSenders.map((sender) => {
      const annotations = mutableAnnotations.filter(
        (annotation) => annotation.targetType === "sender" && annotation.targetId === sender.email,
      );
      const tags = Array.from(new Set(annotations.map((annotation) => annotation.tag)));
      return {
        ...sender,
        annotationCount: annotations.length,
        tags,
      };
    });
    return HttpResponse.json({ items });
  }),

  http.get("/api/mirror/senders/:email", ({ params }) => {
    const sender = mockSenders.find((candidate) => candidate.email === params["email"]);
    if (!sender) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const annotations = mutableAnnotations.filter(
      (annotation) => annotation.targetType === "sender" && annotation.targetId === sender.email,
    );
    const logs = mockLogs.filter((log) =>
      annotations.some((annotation) => annotation.agentRunId === log.agentRunId),
    );
    return HttpResponse.json({
      ...sender,
      annotationCount: annotations.length,
      tags: Array.from(new Set(annotations.map((annotation) => annotation.tag))),
      firstSeen: "2025-01-15T00:00:00Z",
      lastSeen: "2026-04-01T08:00:00Z",
      annotations,
      logs,
      recentMessages: mockMessages,
    });
  }),

  http.get("/api/mirror/senders/:email/guidelines", ({ params }) => {
    const email = params["email"] as string;
    const senderAnnotations = mutableAnnotations.filter(
      (annotation) => annotation.targetType === "sender" && annotation.targetId === email,
    );
    const runIds = Array.from(
      new Set(
        senderAnnotations
          .map((annotation) => annotation.agentRunId)
          .filter((runId) => runId.length > 0),
      ),
    );

    const items = runIds
      .map((runId) => {
        const run = buildRunSummary(runId);
        const linkedIds = runGuidelineLinks.get(runId) ?? new Set<string>();
        const guidelines = mutableGuidelines.filter((guideline) => linkedIds.has(guideline.id));
        if (guidelines.length === 0) {
          return null;
        }
        return {
          runId,
          sourceLabel: run?.sourceLabel ?? "",
          sourceKind: run?.sourceKind ?? "",
          guidelines,
        };
      })
      .filter((item): item is NonNullable<typeof item> => item !== null);

    return HttpResponse.json({ items });
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

    const scopeKind = url.searchParams.get("scopeKind");
    if (scopeKind) result = result.filter((f) => f.scopeKind === scopeKind);

    const agentRunId = url.searchParams.get("agentRunId");
    if (agentRunId) result = result.filter((f) => f.agentRunId === agentRunId);

    const status = url.searchParams.get("status");
    if (status) result = result.filter((f) => f.status === status);

    const feedbackKind = url.searchParams.get("feedbackKind");
    if (feedbackKind) result = result.filter((f) => f.feedbackKind === feedbackKind);

    const mailboxName = url.searchParams.get("mailboxName");
    if (mailboxName) result = result.filter((f) => f.mailboxName === mailboxName);

    const targetType = url.searchParams.get("targetType");
    if (targetType) {
      result = result.filter((f) =>
        f.targets.some((target) => target.targetType === targetType),
      );
    }

    const targetId = url.searchParams.get("targetId");
    if (targetId) {
      result = result.filter((f) =>
        f.targets.some((target) => target.targetId === targetId),
      );
    }

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

  http.get("/api/review-guidelines/:id/runs", ({ params }) => {
    const guidelineId = params["id"] as string;
    const linkedRuns = listRunIDs()
      .filter((runId) => runGuidelineLinks.get(runId)?.has(guidelineId))
      .map((runId) => buildRunSummary(runId))
      .filter((run): run is NonNullable<typeof run> => run !== null);
    return HttpResponse.json({ items: linkedRuns });
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
