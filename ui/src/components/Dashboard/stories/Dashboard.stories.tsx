import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { DashboardStatGrid } from "../DashboardStatGrid";
import { LatestRunBanner } from "../LatestRunBanner";
import { RecentActivityList } from "../RecentActivityList";
import { mockRuns, mockLogs } from "../../../mocks/annotations";

const defaultStats = [
  { value: 247, label: "To Review", color: "#d29922" },
  { value: 189, label: "Reviewed", color: "#3fb950" },
  { value: 31, label: "Dismissed" },
  { value: 467, label: "Total" },
  { value: 5, label: "Agent Runs", color: "#58a6ff" },
  { value: 142, label: "Senders", color: "#a78bfa" },
];

const meta = {
  title: "Annotations/Dashboard",
  component: DashboardStatGrid,
  args: {
    stats: defaultStats,
  },
  decorators: [
    (Story) => (
      <Box sx={{ maxWidth: 900, bgcolor: "background.default", p: 2 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof DashboardStatGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

export const StatGrid: Story = {};

export const StatGridZeros: Story = {
  args: {
    stats: defaultStats.map((s) => ({ ...s, value: 0 })),
  },
};

export const RunBanner: Story = {
  render: () => (
    <LatestRunBanner
      run={mockRuns[0]!}
      onReview={fn()}
      onInspect={fn()}
    />
  ),
};

export const Activity: Story = {
  render: () => (
    <Stack spacing={2}>
      <Typography variant="overline">Recent Agent Activity</Typography>
      <RecentActivityList logs={mockLogs} />
    </Stack>
  ),
};

export const ActivityEmpty: Story = {
  render: () => (
    <Stack spacing={2}>
      <Typography variant="overline">Recent Agent Activity</Typography>
      <RecentActivityList logs={[]} />
    </Stack>
  ),
};

export const FullDashboard: Story = {
  render: () => (
    <Stack spacing={3}>
      <DashboardStatGrid stats={defaultStats} />
      <LatestRunBanner
        run={mockRuns[0]!}
        onReview={fn()}
        onInspect={fn()}
      />
      <Typography variant="overline">Recent Agent Activity</Typography>
      <RecentActivityList logs={mockLogs} />
    </Stack>
  ),
};
