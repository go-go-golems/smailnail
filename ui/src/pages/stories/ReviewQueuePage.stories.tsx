import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { ReviewQueuePage } from "../ReviewQueuePage";
import { withAll } from "../../test-utils/storybook-decorators";
import { mockAnnotations } from "../../mocks/annotations";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/ReviewQueuePage",
  component: ReviewQueuePage,
  decorators: [
    withAll("/annotations/review"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof ReviewQueuePage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", () => HttpResponse.json([])),
      ],
    },
  },
};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", async () => {
          await new Promise(() => {}); // never resolves
        }),
      ],
    },
  },
};

export const AllReviewed: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", () =>
          HttpResponse.json(
            mockAnnotations.map((a) => ({
              ...a,
              reviewState: "reviewed",
            })),
          ),
        ),
      ],
    },
  },
};

export const OnlyNewsletters: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", () =>
          HttpResponse.json(
            mockAnnotations.filter((a) => a.tag === "newsletter"),
          ),
        ),
      ],
    },
  },
};
