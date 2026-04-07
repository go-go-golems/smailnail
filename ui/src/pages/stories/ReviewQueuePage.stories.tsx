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

export const StatefulMutationDemo: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "This story uses the shared mutable MSW annotation state. Approve, dismiss, or batch-review items and verify that pending-only queue queries shrink or update immediately instead of leaving stale rows behind.",
      },
    },
  },
};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", () => HttpResponse.json({ items: [] })),
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
        http.get("/api/annotations", ({ request }) => {
          const url = new URL(request.url);
          const reviewState = url.searchParams.get("reviewState");
          const reviewed = mockAnnotations.map((a) => ({
            ...a,
            reviewState: "reviewed",
          }));
          const items = reviewState === "to_review"
            ? []
            : reviewed;
          return HttpResponse.json({ items });
        }),
      ],
    },
  },
};

export const OnlyNewsletters: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", ({ request }) => {
          const url = new URL(request.url);
          let items = mockAnnotations.filter((a) => a.tag === "newsletter");
          const reviewState = url.searchParams.get("reviewState");
          if (reviewState) {
            items = items.filter((a) => a.reviewState === reviewState);
          }
          return HttpResponse.json({ items });
        }),
      ],
    },
  },
};
