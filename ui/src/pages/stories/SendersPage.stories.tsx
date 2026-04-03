import type { Meta, StoryObj } from "@storybook/react";
import { http, HttpResponse } from "msw";
import Box from "@mui/material/Box";
import { SendersPage } from "../SendersPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/SendersPage",
  component: SendersPage,
  decorators: [
    withAll("/annotations/senders"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", minHeight: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof SendersPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get("/api/mirror/senders", () => HttpResponse.json([])),
      ],
    },
  },
};
