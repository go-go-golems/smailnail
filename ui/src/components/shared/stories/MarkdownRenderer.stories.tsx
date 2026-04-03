import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import { MarkdownRenderer } from "../MarkdownRenderer";

const meta = {
  title: "Shared/MarkdownRenderer",
  component: MarkdownRenderer,
  args: {
    content:
      "Analyzed 47 messages from this sender. All have identical HTML structure with unsubscribe headers. Subject lines follow pattern: `TechCrunch Daily - {date}`. Classified as **newsletter**.",
  },
  decorators: [
    (Story) => (
      <Box sx={{ maxWidth: 600 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof MarkdownRenderer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithList: Story = {
  args: {
    content: `Classified 23 senders:
- 8 newsletters
- 6 notifications
- 5 bulk-sender
- 3 transactional
- 1 important

All annotations created with review_state=to_review.`,
  },
};

export const WithCode: Story = {
  args: {
    content:
      'Query used: `SELECT sender_email, COUNT(*) FROM messages GROUP BY sender_email HAVING COUNT(*) > 10`\n\nResults filtered by volume threshold.',
  },
};

export const Empty: Story = {
  args: { content: "" },
};
