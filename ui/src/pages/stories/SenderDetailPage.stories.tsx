import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { SenderDetailPage } from "../SenderDetailPage";
import { withAll } from "../../test-utils/storybook-decorators";
import {
  mockAnnotations,
  mockLogs,
  mockMessages,
  mockSenders,
} from "../../mocks/annotations";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/SenderDetailPage",
  component: SenderDetailPage,
  decorators: [
    withAll(
      "/annotations/senders/news@techcrunch.com",
      "/annotations/senders/:email",
    ),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof SenderDetailPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NotFound: Story = {
  decorators: [
    withAll(
      "/annotations/senders/unknown@nowhere.com",
      "/annotations/senders/:email",
    ),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers: [
        http.get("/api/mirror/senders/:email", () =>
          HttpResponse.json({ error: "not found" }, { status: 404 }),
        ),
      ],
    },
  },
};

export const NoAnnotations: Story = {
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/mirror/senders/:email", () =>
          HttpResponse.json({
            ...mockSenders[0],
            firstSeen: "2025-01-15T00:00:00Z",
            lastSeen: "2026-04-01T08:00:00Z",
            annotations: [],
            logs: [],
            recentMessages: mockMessages,
          }),
        ),
      ],
    },
  },
};

export const LotsOfData: Story = {
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/mirror/senders/:email", () =>
          HttpResponse.json({
            ...mockSenders[0],
            messageCount: 312,
            firstSeen: "2024-06-01T00:00:00Z",
            lastSeen: "2026-04-01T08:00:00Z",
            annotations: mockAnnotations.slice(0, 4),
            logs: mockLogs,
            recentMessages: mockMessages,
          }),
        ),
      ],
    },
  },
};
