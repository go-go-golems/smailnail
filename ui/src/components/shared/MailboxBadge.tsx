import Chip from "@mui/material/Chip";
import MailOutlineIcon from "@mui/icons-material/MailOutline";
import SendIcon from "@mui/icons-material/Send";
import ArchiveIcon from "@mui/icons-material/Archive";
import FolderIcon from "@mui/icons-material/Folder";
import { parts } from "./parts";

export interface MailboxBadgeProps {
  mailboxName: string;
  variant?: "chip" | "inline";
  size?: "small" | "medium";
}

const knownIcons: Record<string, React.ReactElement> = {
  INBOX: <MailOutlineIcon fontSize="inherit" />,
  Sent: <SendIcon fontSize="inherit" />,
  Archive: <ArchiveIcon fontSize="inherit" />,
};

export function MailboxBadge({
  mailboxName,
  variant = "chip",
  size = "small",
}: MailboxBadgeProps) {
  if (!mailboxName) return null;

  const icon = knownIcons[mailboxName] ?? <FolderIcon fontSize="inherit" />;

  if (variant === "inline") {
    return (
      <span
        data-part={parts.mailboxBadge}
        style={{
          display: "inline-flex",
          alignItems: "center",
          gap: 4,
          fontSize: "0.75rem",
          fontFamily: "monospace",
          opacity: 0.7,
        }}
      >
        {icon} {mailboxName}
      </span>
    );
  }

  return (
    <Chip
      data-part={parts.mailboxBadge}
      icon={icon}
      label={mailboxName}
      size={size}
      variant="outlined"
      color="default"
      sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
    />
  );
}
