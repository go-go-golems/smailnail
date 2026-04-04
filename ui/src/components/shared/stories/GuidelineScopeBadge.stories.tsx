import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { GuidelineScopeBadge } from "../GuidelineScopeBadge";

const meta = {
  title: "Shared/GuidelineScopeBadge",
  component: GuidelineScopeBadge,
  args: {
    scopeKind: "global",
  },
} satisfies Meta<typeof GuidelineScopeBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Global: Story = {};

export const Mailbox: Story = {
  args: { scopeKind: "mailbox" },
};

export const Sender: Story = {
  args: { scopeKind: "sender" },
};

export const Domain: Story = {
  args: { scopeKind: "domain" },
};

export const Workflow: Story = {
  args: { scopeKind: "workflow" },
};

export const AllScopes: Story = {
  render: () => (
    <Stack direction="row" spacing={1}>
      <GuidelineScopeBadge scopeKind="global" />
      <GuidelineScopeBadge scopeKind="mailbox" />
      <GuidelineScopeBadge scopeKind="sender" />
      <GuidelineScopeBadge scopeKind="domain" />
      <GuidelineScopeBadge scopeKind="workflow" />
    </Stack>
  ),
};
