import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import { BatchActionBar } from "../BatchActionBar";

const meta = {
  title: "Shared/BatchActionBar",
  component: BatchActionBar,
  args: {
    totalCount: 23,
    selectedCount: 0,
    allSelected: false,
    onToggleAll: fn(),
    onApprove: fn(),
    onDismiss: fn(),
    onReset: fn(),
  },
  decorators: [
    (Story) => (
      <Box sx={{ width: 600, border: 1, borderColor: "divider", borderRadius: 1 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof BatchActionBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NoSelection: Story = {};

export const SomeSelected: Story = {
  args: { selectedCount: 5 },
};

export const AllSelected: Story = {
  args: { selectedCount: 23, allSelected: true },
};

export const WithoutReset: Story = {
  args: { selectedCount: 3, onReset: undefined },
};
