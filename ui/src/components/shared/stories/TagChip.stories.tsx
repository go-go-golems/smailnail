import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Stack from "@mui/material/Stack";
import { TagChip } from "../TagChip";

const meta = {
  title: "Shared/TagChip",
  component: TagChip,
  args: {
    tag: "newsletter",
  },
} satisfies Meta<typeof TagChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const AllTags: Story = {
  render: () => (
    <Stack direction="row" spacing={1} flexWrap="wrap">
      {[
        "newsletter",
        "notification",
        "bulk-sender",
        "important",
        "transactional",
        "ignore",
        "action-required",
        "personal",
        "marketing",
        "spam",
      ].map((tag) => (
        <TagChip key={tag} tag={tag} />
      ))}
    </Stack>
  ),
};

export const UnknownTag: Story = {
  args: { tag: "custom-tag-xyz" },
};

export const MediumSize: Story = {
  args: { tag: "important", size: "medium" },
};

export const Clickable: Story = {
  args: { tag: "newsletter", onClick: fn() },
};
