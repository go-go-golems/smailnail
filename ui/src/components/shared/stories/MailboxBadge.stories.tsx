import type { Meta, StoryObj } from "@storybook/react";
import Stack from "@mui/material/Stack";
import { MailboxBadge } from "../MailboxBadge";

const meta = {
  title: "Shared/MailboxBadge",
  component: MailboxBadge,
  args: {
    mailboxName: "INBOX",
  },
} satisfies Meta<typeof MailboxBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Inbox: Story = {};

export const Sent: Story = {
  args: { mailboxName: "Sent" },
};

export const Archive: Story = {
  args: { mailboxName: "Archive" },
};

export const Custom: Story = {
  args: { mailboxName: "Billing" },
};

export const Empty: Story = {
  args: { mailboxName: "" },
};

export const InlineVariant: Story = {
  args: { mailboxName: "INBOX", variant: "inline" },
};

export const AllVariants: Story = {
  render: () => (
    <Stack direction="row" spacing={1} alignItems="center">
      <MailboxBadge mailboxName="INBOX" />
      <MailboxBadge mailboxName="Sent" />
      <MailboxBadge mailboxName="Archive" />
      <MailboxBadge mailboxName="Billing" />
    </Stack>
  ),
};
