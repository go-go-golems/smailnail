import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { ReviewProgressBar } from "../ReviewProgressBar";

const meta = {
  title: "Shared/ReviewProgressBar",
  component: ReviewProgressBar,
  args: {
    reviewed: 10,
    pending: 18,
    dismissed: 2,
  },
  decorators: [
    (Story) => (
      <Box sx={{ width: 300 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof ReviewProgressBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const AllReviewed: Story = {
  args: { reviewed: 30, pending: 0, dismissed: 0 },
};

export const AllPending: Story = {
  args: { reviewed: 0, pending: 30, dismissed: 0 },
};

export const MixedStates: Story = {
  render: () => (
    <Stack spacing={2}>
      {[
        { reviewed: 3, pending: 18, dismissed: 2, label: "Run 42 (mostly pending)" },
        { reviewed: 10, pending: 3, dismissed: 2, label: "Run 41 (mostly reviewed)" },
        { reviewed: 5, pending: 0, dismissed: 3, label: "Run 40 (complete)" },
        { reviewed: 0, pending: 0, dismissed: 0, label: "Empty run" },
      ].map((run) => (
        <Box key={run.label}>
          <Typography variant="caption" sx={{ mb: 0.5, display: "block" }}>
            {run.label}
          </Typography>
          <ReviewProgressBar
            reviewed={run.reviewed}
            pending={run.pending}
            dismissed={run.dismissed}
          />
        </Box>
      ))}
    </Stack>
  ),
};

export const TallerBar: Story = {
  args: { reviewed: 10, pending: 18, dismissed: 2, height: 12 },
};
