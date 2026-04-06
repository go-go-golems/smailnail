import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { GuidelinesListPage } from "../GuidelinesListPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";
import { mockGuidelines } from "../../mocks/annotations";

const meta = {
  title: "Pages/GuidelinesListPage",
  component: GuidelinesListPage,
  decorators: [
    withAll("/annotations/guidelines"),
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
        http.get("/api/review-guidelines", () =>
          HttpResponse.json({ items: mockGuidelines }),
        ),
      ],
    },
  },
} satisfies Meta<typeof GuidelinesListPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const OnlyActive: Story = {
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/review-guidelines", ({ request }) => {
          const url = new URL(request.url);
          const status = url.searchParams.get("status");
          const filtered = status
            ? mockGuidelines.filter((g) => g.status === status)
            : mockGuidelines;
          return HttpResponse.json({ items: filtered });
        }),
      ],
    },
  },
};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        ...handlers,
        http.get("/api/review-guidelines", () => HttpResponse.json({ items: [] })),
      ],
    },
  },
};
