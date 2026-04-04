import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { SenderProfileCard } from "../SenderProfileCard";
import { MessagePreviewTable } from "../MessagePreviewTable";
import { AgentReasoningPanel } from "../AgentReasoningPanel";
import {
  mockSenders,
  mockMessages,
  mockLogs,
  mockAnnotations,
} from "../../../mocks/annotations";
import type { SenderDetail } from "../../../types/annotations";

const senderDetail: SenderDetail = {
  ...mockSenders[0]!,
  firstSeen: "2025-01-15T00:00:00Z",
  lastSeen: "2026-04-01T08:00:00Z",
  annotations: mockAnnotations.filter(
    (a) => a.targetType === "sender" && a.targetId === mockSenders[0]!.email,
  ),
  logs: mockLogs.slice(0, 2),
  recentMessages: mockMessages,
};

const meta = {
  title: "Annotations/SenderProfile",
  component: SenderProfileCard,
  args: {
    sender: senderDetail,
  },
  decorators: [
    (Story) => (
      <Box sx={{ maxWidth: 800, bgcolor: "background.default", p: 2 }}>
        <Story />
      </Box>
    ),
  ],
} satisfies Meta<typeof SenderProfileCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ProfileCard: Story = {
  args: { sender: senderDetail },
};

export const ProfileCardNoTags: Story = {
  args: {
    sender: { ...senderDetail, tags: [] },
  },
};

export const Messages: Story = {
  render: () => (
    <Stack spacing={2}>
      <Typography variant="overline">Recent Messages</Typography>
      <MessagePreviewTable messages={mockMessages} />
    </Stack>
  ),
};

export const MessagesEmpty: Story = {
  render: () => (
    <Stack spacing={2}>
      <Typography variant="overline">Recent Messages</Typography>
      <MessagePreviewTable messages={[]} />
    </Stack>
  ),
};

export const AgentReasoning: Story = {
  render: () => (
    <Stack spacing={2}>
      <Typography variant="overline">Agent Reasoning</Typography>
      <AgentReasoningPanel logs={mockLogs.slice(0, 2)} />
    </Stack>
  ),
};

export const FullSenderDetail: Story = {
  render: () => (
    <Stack spacing={3}>
      <Typography variant="h3">
        {senderDetail.displayName} ({senderDetail.email})
      </Typography>
      <SenderProfileCard sender={senderDetail} />
      <Typography variant="overline">Agent Reasoning</Typography>
      <AgentReasoningPanel logs={senderDetail.logs} />
      <Typography variant="overline">Recent Messages</Typography>
      <MessagePreviewTable messages={senderDetail.recentMessages} />
    </Stack>
  ),
};
