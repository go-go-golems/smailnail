import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { GuidelineDetailPage } from "../GuidelineDetailPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";
import { mockGuidelines } from "../../mocks/annotations";

const sampleGuideline = mockGuidelines[0]!;

const meta = {
  title: "Pages/GuidelineDetailPage",
  component: GuidelineDetailPage,
  decorators: [
    withAll(
      "/annotations/guidelines/guideline-001",
      "/annotations/guidelines/:guidelineId",
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
        ...handlers,
        http.get("/api/guidelines/guideline-001", () =>
          HttpResponse.json(sampleGuideline),
        ),
      ],
    },
  },
} satisfies Meta<typeof GuidelineDetailPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ViewMode: Story = {};

export const NotFound: Story = {
  decorators: [
    withAll(
      "/annotations/guidelines/guideline-999",
      "/annotations/guidelines/:guidelineId",
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
        ...handlers,
        http.get("/api/guidelines/guideline-999", () =>
          HttpResponse.json({ error: "not found" }, { status: 404 }),
        ),
      ],
    },
  },
};

export const CreateMode: Story = {
  decorators: [
    withAll("/annotations/guidelines/new", "/annotations/guidelines/:guidelineId"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
};
