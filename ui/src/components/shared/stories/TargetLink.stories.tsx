import type { Meta, StoryObj } from "@storybook/react";
import { fn } from "@storybook/test";
import Stack from "@mui/material/Stack";
import { TargetLink } from "../TargetLink";

const meta = {
  title: "Shared/TargetLink",
  component: TargetLink,
  args: {
    targetType: "sender",
    targetId: "news@techcrunch.com",
  },
} satisfies Meta<typeof TargetLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Sender: Story = {};

export const Domain: Story = {
  args: { targetType: "domain", targetId: "github.com" },
};

export const Message: Story = {
  args: { targetType: "message", targetId: "msg-12345" },
};

export const Group: Story = {
  args: { targetType: "group", targetId: "grp-001" },
};

export const Clickable: Story = {
  args: { onClick: fn() },
};

export const AllTypes: Story = {
  render: () => (
    <Stack spacing={1}>
      <TargetLink targetType="sender" targetId="alice@example.com" onClick={fn()} />
      <TargetLink targetType="domain" targetId="example.com" onClick={fn()} />
      <TargetLink targetType="message" targetId="msg-42" onClick={fn()} />
      <TargetLink targetType="group" targetId="grp-001" onClick={fn()} />
    </Stack>
  ),
};
