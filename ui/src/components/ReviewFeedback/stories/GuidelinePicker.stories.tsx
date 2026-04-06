import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { GuidelinePicker } from "../GuidelinePicker";
import { withStore } from "../../../test-utils/storybook-decorators";
import { mockGuidelines } from "../../../mocks/annotations";

const meta = {
  title: "ReviewFeedback/GuidelinePicker",
  component: GuidelinePicker,
  decorators: [
    withStore,
    (Story) => (
      <Box sx={{ width: 350 }}>
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
    selectedIds: [],
    onToggle: (id: string) => console.log("toggle", id),
  },
} satisfies Meta<typeof GuidelinePicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const OneSelected: Story = {
  args: { selectedIds: ["guideline-001"] },
};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/review-guidelines", async () => {
          await new Promise(() => {});
        }),
      ],
    },
  },
};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/review-guidelines", () => HttpResponse.json([])),
      ],
    },
  },
};
