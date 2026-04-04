import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import { FilterPillBar } from "../FilterPillBar";

const tagPills = [
  { key: "newsletter", label: "Newsletter", count: 8 },
  { key: "notification", label: "Notification", count: 6 },
  { key: "bulk-sender", label: "Bulk Sender", count: 5 },
  { key: "transactional", label: "Transactional", count: 3 },
  { key: "important", label: "Important", count: 1 },
];

const meta = {
  title: "Shared/FilterPillBar",
  component: FilterPillBar,
  args: {
    pills: tagPills,
    activeKey: null,
    onSelect: fn(),
  },
} satisfies Meta<typeof FilterPillBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AllSelected: Story = {};

export const OneActive: Story = {
  args: { activeKey: "newsletter" },
};

export const WithoutCounts: Story = {
  args: {
    pills: [
      { key: "agent", label: "Agent" },
      { key: "human", label: "Human" },
      { key: "heuristic", label: "Heuristic" },
    ],
    activeKey: "agent",
  },
};
