import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { FeedbackKindBadge } from "../FeedbackKindBadge";

const meta = {
  title: "Shared/FeedbackKindBadge",
  component: FeedbackKindBadge,
  args: {
    kind: "reject_request",
  },
} satisfies Meta<typeof FeedbackKindBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const RejectRequest: Story = {};

export const Comment: Story = {
  args: { kind: "comment" },
};

export const GuidelineRequest: Story = {
  args: { kind: "guideline_request" },
};

export const Clarification: Story = {
  args: { kind: "clarification" },
};

export const AllKinds: Story = {
  render: () => (
    <Stack direction="row" spacing={1}>
      <FeedbackKindBadge kind="reject_request" />
      <FeedbackKindBadge kind="comment" />
      <FeedbackKindBadge kind="guideline_request" />
      <FeedbackKindBadge kind="clarification" />
    </Stack>
  ),
};
