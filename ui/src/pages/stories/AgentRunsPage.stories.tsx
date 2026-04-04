import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { AgentRunsPage } from "../AgentRunsPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/AgentRunsPage",
  component: AgentRunsPage,
  decorators: [
    withAll("/annotations/runs"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof AgentRunsPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotation-runs", () => HttpResponse.json([])),
      ],
    },
  },
};

export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotation-runs", async () => {
          await new Promise(() => {});
        }),
      ],
    },
  },
};
