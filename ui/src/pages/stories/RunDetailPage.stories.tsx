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

const meta = {
  title: "Pages/RunDetailPage",
  component: RunDetailPage,
  decorators: [
    withAll("/annotations/runs/run-42", "/annotations/runs/:runId"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers,
    },
  },
} satisfies Meta<typeof RunDetailPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const StatefulMutationDemo: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "This story uses the shared mutable MSW annotation state. Approve or dismiss annotations and verify that the run counters, annotation rows, annotation-scoped feedback, and linked guideline sections stay coordinated.",
      },
    },
  },
};

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
    withAll("/annotations/runs/run-999", "/annotations/runs/:runId"),
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
        http.get("/api/annotation-runs/:id", () =>
          HttpResponse.json({
            runId: "run-42",
            sourceLabel: "triage-agent-v2",
            sourceKind: "agent",
            annotationCount: 4,
            pendingCount: 0,
            reviewedCount: 4,
            dismissedCount: 0,
            logCount: 4,
            groupCount: 2,
            startedAt: "2026-04-01T10:29:00Z",
            completedAt: "2026-04-01T10:40:00Z",
            annotations: mockAnnotations
              .filter((annotation) => annotation.agentRunId === "run-42")
              .map((annotation) => ({ ...annotation, reviewState: "reviewed" })),
            logs: mockLogs.filter((log) => log.agentRunId === "run-42"),
            groups: mockGroups.filter((group) => group.agentRunId === "run-42"),
          }),
        ),
      ],
    },
  },
};
