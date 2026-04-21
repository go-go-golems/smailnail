import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import { GuidelineCard } from "../GuidelineCard";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "RunGuideline/GuidelineCard",
  component: GuidelineCard,
  decorators: [
    (Story) => (
      <Box sx={{ bgcolor: "background.default", p: 2, width: 500 }}>
        <Story />
      </Box>
    ),
  ],
  args: {
    onUnlink: () => console.log("unlink"),
  },
} satisfies Meta<typeof GuidelineCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Active: Story = {
  args: { guideline: mockGuidelines[0]! },
};

export const Draft: Story = {
  args: { guideline: mockGuidelines[2]! },
};

export const Archived: Story = {
  args: { guideline: mockGuidelines[3]! },
};

export const Compact: Story = {
  args: { guideline: mockGuidelines[0]!, compact: true },
};

export const WithoutUnlink: Story = {
  args: { guideline: mockGuidelines[0]! },
};
