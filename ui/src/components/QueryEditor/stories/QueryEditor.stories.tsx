import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Box from "@mui/material/Box";
import { QueryEditor } from "../QueryEditor";
import { mockPresets, mockQueryResult } from "../../../mocks/annotations";

const meta = {
  title: "QueryEditor/QueryEditor",
  component: QueryEditor,
  decorators: [
    (Story) => (
      <Box sx={{ height: "100vh", bgcolor: "background.default" }}>
        <Story />
      </Box>
    ),
  ],
  args: {
    sql: "SELECT tag, COUNT(*) as count\nFROM annotations\nGROUP BY tag\nORDER BY count DESC;",
    onSqlChange: fn(),
    onExecute: fn(),
    onSave: fn(),
    onSelectQuery: fn(),
    presets: mockPresets,
    savedQueries: [],
    result: null,
    error: null,
    isLoading: false,
  },
} satisfies Meta<typeof QueryEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptyState: Story = {};

export const WithResults: Story = {
  args: {
    result: mockQueryResult,
  },
};

export const WithError: Story = {
  args: {
    sql: "SELECT error_column FROM annotations;",
    error: {
      message:
        'Error: Referenced column "error_column" not found in FROM clause!\nCandidate bindings: "tag", "note_markdown", "source_label"',
    },
  },
};

export const Loading: Story = {
  args: {
    isLoading: true,
  },
};

export const NoSaveButton: Story = {
  args: {
    onSave: undefined,
    result: mockQueryResult,
  },
};
