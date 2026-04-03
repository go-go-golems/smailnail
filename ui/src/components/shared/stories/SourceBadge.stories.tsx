import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { SourceBadge } from "../SourceBadge";

const meta = {
  title: "Shared/SourceBadge",
  component: SourceBadge,
  args: {
    sourceKind: "agent",
    sourceLabel: "triage-agent-v2",
  },
} satisfies Meta<typeof SourceBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Agent: Story = {};

export const Human: Story = {
  args: { sourceKind: "human", sourceLabel: "manuel" },
};

export const Heuristic: Story = {
  args: { sourceKind: "heuristic", sourceLabel: "volume-detector" },
};

export const Import: Story = {
  args: { sourceKind: "import", sourceLabel: "csv-import-2026" },
};

export const AllKinds: Story = {
  render: () => (
    <Stack direction="row" spacing={1}>
      <SourceBadge sourceKind="agent" sourceLabel="triage-agent-v2" />
      <SourceBadge sourceKind="human" sourceLabel="manuel" />
      <SourceBadge sourceKind="heuristic" sourceLabel="volume-detector" />
      <SourceBadge sourceKind="import" sourceLabel="csv-import-2026" />
    </Stack>
  ),
};
