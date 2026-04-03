import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { DashboardPage } from "../DashboardPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/DashboardPage",
  component: DashboardPage,
  decorators: [
    withAll("/annotations"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof DashboardPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/annotations", () => HttpResponse.json([])),
        http.get("/api/annotation-runs", () => HttpResponse.json([])),
        http.get("/api/annotation-logs", () => HttpResponse.json([])),
        http.get("/api/mirror/senders", () => HttpResponse.json([])),
      ],
    },
  },
};
