import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { ReviewStateBadge } from "../ReviewStateBadge";

const meta = {
  title: "Shared/ReviewStateBadge",
  component: ReviewStateBadge,
  args: {
    state: "to_review",
  },
} satisfies Meta<typeof ReviewStateBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ToReview: Story = {};

export const Reviewed: Story = {
  args: { state: "reviewed" },
};

export const Dismissed: Story = {
  args: { state: "dismissed" },
};

export const AllStates: Story = {
  render: () => (
    <Stack direction="row" spacing={1}>
      <ReviewStateBadge state="to_review" />
      <ReviewStateBadge state="reviewed" />
      <ReviewStateBadge state="dismissed" />
    </Stack>
  ),
};

export const MediumSize: Story = {
  args: { state: "to_review", size: "medium" },
};
