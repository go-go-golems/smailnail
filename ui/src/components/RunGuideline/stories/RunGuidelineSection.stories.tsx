import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { RunGuidelineSection } from "../RunGuidelineSection";
import { withStore } from "../../../test-utils/storybook-decorators";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "RunGuideline/RunGuidelineSection",
  component: RunGuidelineSection,
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
        http.get("/api/review-guidelines", () =>
          HttpResponse.json(mockGuidelines.filter((g) => g.status === "active")),
        ),
        http.post("/api/annotation-runs/:id/guidelines", () =>
          HttpResponse.json(null, { status: 204 }),
        ),
        http.delete("/api/annotation-runs/:id/guidelines/:guidelineId", () =>
          HttpResponse.json(null, { status: 204 }),
        ),
      ],
    },
  },
  args: {
    runId: "run-42",
    onCreateAndLink: () => console.log("create and link"),
  },
} satisfies Meta<typeof RunGuidelineSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MultipleGuidelines: Story = {
  args: {
    guidelines: mockGuidelines.filter((g) =>
      ["guideline-001", "guideline-002"].includes(g.id),
    ),
  },
};

export const OneGuideline: Story = {
  args: {
    guidelines: [mockGuidelines[0]!],
  },
};

export const Empty: Story = {
  args: {
    guidelines: [],
  },
};
