import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import { RunTimeline } from "../RunTimeline";
import { mockLogs } from "../../../mocks/annotations";

const meta = {
  title: "Annotations/RunTimeline",
  component: RunTimeline,
  args: {
    logs: mockLogs,
    onNavigateTarget: fn(),
  },
  decorators: [
    (Story) => (
      <Box sx={{ maxWidth: 800, bgcolor: "background.default", p: 2 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof RunTimeline>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  args: { logs: [] },
};

export const SingleEntry: Story = {
  args: { logs: [mockLogs[0]!] },
};

export const ManyEntries: Story = {
  args: {
    logs: [
      ...mockLogs,
      {
        id: "log-005",
        logKind: "error",
        title: "Failed to classify sender",
        bodyMarkdown:
          "Could not determine category for `unknown@mystery.io` — insufficient message history (only 1 message). Skipping.",
        sourceKind: "agent" as const,
        sourceLabel: "triage-agent-v2",
        agentRunId: "run-42",
        createdBy: "system",
        createdAt: "2026-04-01T10:33:00Z",
      },
    ],
  },
};
