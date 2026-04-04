import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import { FeedbackCard } from "../FeedbackCard";
import { mockFeedback } from "../../../mocks/annotations";

const meta = {
  title: "ReviewFeedback/FeedbackCard",
  component: FeedbackCard,
  decorators: [
    (Story) => (
      <Box sx={{ bgcolor: "background.default", p: 2, width: 500 }}>
        <Story />
      </Box>
    ),
  ],
  args: {
    onAcknowledge: () => console.log("acknowledge"),
    onResolve: () => console.log("resolve"),
  },
} satisfies Meta<typeof FeedbackCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OpenRejectRequest: Story = {
  args: { feedback: mockFeedback[0]! },
};

export const ResolvedComment: Story = {
  args: { feedback: mockFeedback[1]! },
};

export const AcknowledgedClarification: Story = {
  args: { feedback: mockFeedback[2]! },
};

export const CompactMode: Story = {
  args: { feedback: mockFeedback[0]!, compact: true },
};

export const AllStates: Story = {
  args: { feedback: mockFeedback[0]! },
  render: function AllStatesRender(args) {
    return (
      <Stack spacing={1}>
        {mockFeedback.map((fb) => (
          <FeedbackCard
            key={fb.id}
            feedback={fb}
            onAcknowledge={args.onAcknowledge}
            onResolve={args.onResolve}
          />
        ))}
      </Stack>
    );
  },
};
