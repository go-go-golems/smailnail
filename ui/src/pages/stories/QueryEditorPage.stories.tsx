import type { Meta, StoryObj } from "@storybook/react";
import Box from "@mui/material/Box";
import { QueryEditorPage } from "../QueryEditorPage";
import { withAll } from "../../test-utils/storybook-decorators";
import { handlers } from "../../mocks/handlers";

const meta = {
  title: "Pages/QueryEditorPage",
  component: QueryEditorPage,
  decorators: [
    withAll("/query"),
    (Story) => (
      <Box sx={{ bgcolor: "background.default", height: "100vh" }}>
        <Story />
      </Box>
    ),
  ],
  parameters: {
    msw: { handlers },
  },
} satisfies Meta<typeof QueryEditorPage>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
