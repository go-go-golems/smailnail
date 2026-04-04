import type { Meta, StoryObj } from "@storybook/react";
import { CountSummaryBar } from "../CountSummaryBar";

const meta = {
  title: "Shared/CountSummaryBar",
  component: CountSummaryBar,
  args: {
    items: [
      { label: "to review", value: 247, color: "#d29922" },
      { label: "agent", value: 189, color: "#58a6ff" },
      { label: "heuristic", value: 58 },
    ],
  },
} satisfies Meta<typeof CountSummaryBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const SingleItem: Story = {
  args: {
    items: [{ label: "pending", value: 42, color: "#d29922" }],
  },
};

export const RunStats: Story = {
  args: {
    items: [
      { label: "total", value: 23 },
      { label: "reviewed", value: 3, color: "#3fb950" },
      { label: "pending", value: 18, color: "#d29922" },
      { label: "dismissed", value: 2 },
    ],
  },
};
