import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { ReviewCommentDrawer } from "../ReviewCommentDrawer";
import { withStore } from "../../../test-utils/storybook-decorators";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "ReviewFeedback/ReviewCommentDrawer",
  component: ReviewCommentDrawer,
  decorators: [
    withStore,
    (Story) => (
      <Box sx={{ bgcolor: "background.default", p: 2 }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers: [
        http.get("/api/review-guidelines", () =>
          HttpResponse.json(mockGuidelines.filter((g) => g.status === "active")),
        ),
      ],
    },
  },
  args: {
    onSubmit: (payload) => console.log("submit", payload),
    onCancel: () => console.log("cancel"),
  },
} satisfies Meta<typeof ReviewCommentDrawer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OpenBatchMode: Story = {
  args: {
    open: true,
    mode: "batch",
    targetCount: 3,
    mailboxName: "INBOX",
  },
};

export const OpenSingleMode: Story = {
  args: {
    open: true,
    mode: "single",
    targetCount: 1,
  },
};

export const OpenRunMode: Story = {
  args: {
    open: true,
    mode: "run",
    targetCount: 0,
    agentRunId: "run-42",
  },
};

export const Closed: Story = {
  args: {
    open: false,
    mode: "batch",
    targetCount: 3,
  },
};
