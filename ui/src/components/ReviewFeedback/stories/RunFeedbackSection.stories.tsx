import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { RunFeedbackSection } from "../RunFeedbackSection";
import { withStore } from "../../../test-utils/storybook-decorators";
import { mockFeedback } from "../../../mocks/annotations";

const meta = {
  title: "ReviewFeedback/RunFeedbackSection",
  component: RunFeedbackSection,
  decorators: [
    withStore,
    (Story) => (
      <Box sx={{ bgcolor: "background.default", p: 2, width: 600 }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: {
      handlers: [
        http.post("/api/review-feedback", async () =>
          HttpResponse.json(null, { status: 201 }),
        ),
        http.patch("/api/review-feedback/:id", async () =>
          HttpResponse.json(null, { status: 200 }),
        ),
      ],
    },
  },
  args: {
    runId: "run-42",
  },
} satisfies Meta<typeof RunFeedbackSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MultipleFeedback: Story = {
  args: {
    feedback: mockFeedback.filter((f) => f.agentRunId === "run-42"),
  },
};

export const Empty: Story = {
  args: {
    feedback: [],
  },
};
