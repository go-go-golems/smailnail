import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import { GroupCard } from "../GroupCard";
import { mockGroups, mockGroupMembers } from "../../../mocks/annotations";

const meta = {
  title: "Annotations/GroupCard",
  component: GroupCard,
  args: {
    group: mockGroups[0]!,
    members: mockGroupMembers.filter((m) => m.groupId === mockGroups[0]!.id),
    onNavigateTarget: fn(),
  },
  decorators: [
    (Story) => (
      <Box sx={{ maxWidth: 600, bgcolor: "background.default", p: 2 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof GroupCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoMembers: Story = {
  args: { members: [] },
};

export const Reviewed: Story = {
  args: {
    group: mockGroups[1]!,
    members: mockGroupMembers.filter((m) => m.groupId === mockGroups[1]!.id),
  },
};

export const NoDescription: Story = {
  args: {
    group: { ...mockGroups[0]!, description: "" },
  },
};
