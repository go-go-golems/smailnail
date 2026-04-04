import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { FeedbackStatusBadge } from "../FeedbackStatusBadge";

const meta = {
  title: "Shared/FeedbackStatusBadge",
  component: FeedbackStatusBadge,
  args: {
    status: "open",
  },
} satisfies Meta<typeof FeedbackStatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {};

export const Acknowledged: Story = {
  args: { status: "acknowledged" },
};

export const Resolved: Story = {
  args: { status: "resolved" },
};

export const Archived: Story = {
  args: { status: "archived" },
};

export const AllStatuses: Story = {
  render: () => (
    <Stack direction="row" spacing={1}>
      <FeedbackStatusBadge status="open" />
      <FeedbackStatusBadge status="acknowledged" />
      <FeedbackStatusBadge status="resolved" />
      <FeedbackStatusBadge status="archived" />
    </Stack>
  ),
};
