import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { StatBox } from "../StatBox";

const meta = {
  title: "Shared/StatBox",
  component: StatBox,
  args: {
    value: 247,
    label: "To Review",
  },
} satisfies Meta<typeof StatBox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithColor: Story = {
  args: { value: 42, label: "Approved", color: "success.main" },
};

export const DashboardRow: Story = {
  render: () => (
    <Stack direction="row" spacing={2}>
      <StatBox value={247} label="To Review" color="warning.main" />
      <StatBox value={189} label="Reviewed" color="success.main" />
      <StatBox value={31} label="Dismissed" color="text.secondary" />
      <StatBox value={467} label="Total" />
      <StatBox value={5} label="Agent Runs" color="info.main" />
      <StatBox value={142} label="Senders" color="secondary.main" />
    </Stack>
  ),
};

export const LargeNumber: Story = {
  args: { value: "12,345", label: "Messages Indexed" },
};

export const ZeroState: Story = {
  args: { value: 0, label: "Pending" },
};
