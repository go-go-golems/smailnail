import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { RunDetailPage } from "../RunDetailPage";
import { withAll } from "../../test-utils/storybook-decorators";
import {
  mockAnnotations,
  mockLogs,
  mockGroups,
} from "../../mocks/annotations";
import { handlers } from "../../mocks/handlers";

const runDetail = {
  runId: "run-42",
  sourceLabel: "triage-agent-v2",
  sourceKind: "agent" as const,
  annotationCount: 5,
  pendingCount: 3,
  reviewedCount: 1,
  dismissedCount: 1,
  logCount: 4,
  groupCount: 2,
  startedAt: "2026-04-01T10:29:00Z",
  completedAt: "2026-04-01T10:40:00Z",
  annotations: mockAnnotations.filter((a) => a.agentRunId === "run-42"),
  logs: mockLogs.filter((l) => l.agentRunId === "run-42"),
  groups: mockGroups.filter((g) => g.agentRunId === "run-42"),
};

const meta = {
  title: "Pages/RunDetailPage",
  component: RunDetailPage,
  decorators: [
    withAll("/annotations/runs/run-42"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/annotation-runs/run-42", () =>
          HttpResponse.json(runDetail),
        ),
      ],
    },
  },
} satisfies Meta<typeof RunDetailPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotation-runs/:id", async () => {
          await new Promise(() => {});
        }),
      ],
    },
  },
};

export const NotFound: Story = {
  decorators: [
    withAll("/annotations/runs/run-999"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotation-runs/:id", () =>
          HttpResponse.json({ error: "not found" }, { status: 404 }),
        ),
      ],
    },
  },
};

export const AllReviewed: Story = {
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/annotation-runs/run-42", () =>
          HttpResponse.json({
            ...runDetail,
            pendingCount: 0,
            reviewedCount: 5,
            annotations: runDetail.annotations.map((a) => ({
              ...a,
              reviewState: "reviewed",
            })),
          }),
        ),
      ],
    },
  },
};
